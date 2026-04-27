//
// raft.go
//

package raft

//
// API
// ===
// This is an outline of the API that your raft implementation should
// expose.
//
// rf = NewPeer(...)
//   Create a new Raft server(node).
//
// rf.PutCommand(command interface{}) (index, term, isleader)
//   PutCommand agreement on a new log entry
//
// rf.GetState() (me, term, isLeader)
//   Ask a Raft peer for "me", its current term, and whether it thinks it
//   is a leader
//
// ApplyCommand
//   Each time a new entry is committed to the log, each Raft peer
//   should send an ApplyCommand to the service (e.g. tester) on the
//   same server, via the applyCh channel passed to NewPeer()
//

import (
	"fmt"
	"log"
	"math/rand"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/cmu440/rpc"
)

// Set to false to disable debug logs completely
// Make sure to set kEnableDebugLogs to false before submitting
const kEnableDebugLogs = false

// Set to true to log to stdout instead of file
const kLogToStdout = true

// Change this to output logs to a different directory
const kLogOutputDir = "./raftlogs/"

// ApplyCommand
// ========
//
// As each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyCommand to the service (or
// tester) on the same server, via the applyCh passed to NewPeer()
type ApplyCommand struct {
	Index   int         // log index for this entry
	Command interface{} // user command to apply
}

// Raft struct
// ===========
//
// A Go object implementing a single Raft peer
type Raft struct {
	mux   sync.Mutex       // guards this peer's state
	peers []*rpc.ClientEnd // RPC endpoints of all peers
	me    int              // index of this peer in peers[]

	logger *log.Logger // per-peer logger

	state       int        // follower, candidate, or leader
	currentTerm int        // latest term seen
	votedFor    int        // candidate id we voted for this term
	log         []LogEntry // replicated log entries

	commitIndex int // highest log index known committed
	lastApplied int // highest log index applied to state machine

	applyCh   chan ApplyCommand // delivers committed entries to service
	applyCond *sync.Cond        // signals when commit index moves

	electionTimeoutMin time.Duration // min randomized election timeout
	electionTimeoutMax time.Duration // max randomized election timeout
	heartbeatInterval  time.Duration // interval between heartbeats

	electionResetCh chan struct{} // tells ticker to reset election timer
	killCh          chan struct{} // closed to shut down helpers
	dead            bool          // set after Stop()

	nextIndex  []int // next log index to send per follower
	matchIndex []int // highest index known replicated per follower

	rand    *rand.Rand // RNG for timeout jitter
	randMux sync.Mutex // protects rand
}



// NewPeer
// ====
//
//
// The port numbers of all the Raft servers (including this one)
// are in peers[]
//
// This server's port is peers[me]
//
// All the servers' peers[] arrays have the same order
//
// applyCh
// =======
//
// applyCh is a channel on which the tester or service expects
// Raft to send ApplyCommand messages
//
// NewPeer() must return quickly, so it should start Goroutines
// for any long-running work
// NewPeer creates and returns a new Raft server instance.
func NewPeer(peers []*rpc.ClientEnd, me int, applyCh chan ApplyCommand) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.me = me

	if kEnableDebugLogs {
		peerName := peers[me].String()
		logPrefix := fmt.Sprintf("%s ", peerName)
		if kLogToStdout {
			rf.logger = log.New(os.Stdout, peerName, log.Lmicroseconds|log.Lshortfile)
		} else {
			err := os.MkdirAll(kLogOutputDir, os.ModePerm)
			if err != nil {
				panic(err.Error())
			}
			logOutputFile, err := os.OpenFile(fmt.Sprintf("%s/%s.txt", kLogOutputDir, logPrefix), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				panic(err.Error())
			}
			rf.logger = log.New(logOutputFile, logPrefix, log.Lmicroseconds|log.Lshortfile)
		}
		rf.logger.Println("logger initialized")
	} else {
		rf.logger = log.New(ioutil.Discard, "", 0)
	}

	rf.state = stateFollower
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.log = []LogEntry{{Term: 0}}
	rf.applyCh = applyCh
	rf.applyCond = sync.NewCond(&rf.mux)
	rf.electionTimeoutMin = 400 * time.Millisecond
	rf.electionTimeoutMax = 600 * time.Millisecond
	rf.heartbeatInterval = 120 * time.Millisecond
	rf.electionResetCh = make(chan struct{}, 1)
	rf.killCh = make(chan struct{})
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))
	rf.rand = rand.New(rand.NewSource(time.Now().UnixNano() + int64(me)))

	go rf.runApplier()
	go rf.ticker() // background goroutine drives elections

	return rf
}

// ticker waits for election timeouts and triggers new elections.
func (rf *Raft) ticker() {
	timeout := rf.randomElectionTimeout()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			rf.mux.Lock()
			if rf.state != stateLeader && !rf.killed() {
				// timeout, follower should try to become leader
				rf.startElection()
				timeout = rf.randomElectionTimeout()
				rf.mux.Unlock()
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(timeout)
				continue
			}
			timeout = rf.randomElectionTimeout()
			rf.mux.Unlock()
			timer.Reset(timeout)
		case <-rf.electionResetCh:
			// heartbeat or vote reset: restart timer with a fresh interval
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timeout = rf.randomElectionTimeout()
			timer.Reset(timeout)
		case <-rf.killCh:
			// Stop() closed the channel: exit goroutine
			return
		}
	}
}



// LogEntry represents a single replicated state machine command.
type LogEntry struct {
	Term    int         // term when leader received the entry
	Command interface{} // client command to run
}

const (
	stateFollower  = iota // passively replicates log entries and grants votes
	stateCandidate        // running for leadership and requesting votes
	stateLeader           // elected leader
)

// GetState
// ==========
//
// Return "me", current term and whether this peer
// believes it is the leader
func (rf *Raft) GetState() (int, int, bool) {
	rf.mux.Lock()
	defer rf.mux.Unlock()

	me := rf.me
	term := rf.currentTerm
	isleader := rf.state == stateLeader
	return me, term, isleader
}

// RequestVoteArgs
type RequestVoteArgs struct {
	Term         int // candidate's current term
	CandidateID  int // who is requesting the vote
	LastLogIndex int // index of candidate's last log entry
	LastLogTerm  int // term of candidate's last log entry
}

// RequestVoteReply
type RequestVoteReply struct {
	Term        int  // responder's current term
	VoteGranted bool // true if vote is granted
}

// RequestVote
// RequestVote processes an incoming vote request and decides whether to grant it.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	shouldReset := false
	rf.mux.Lock()
	defer func() {
		rf.mux.Unlock()
		if shouldReset {
			rf.resetElectionTimer()
		}
	}()

	if args.Term < rf.currentTerm {
		// old term: reject but hint the latest term back to the caller
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}

	if args.Term > rf.currentTerm {
		// saw a newer term back to follower
		rf.becomeFollower(args.Term)
	}

	reply.Term = rf.currentTerm

	alreadyVoted := rf.votedFor != -1 && rf.votedFor != args.CandidateID

	// check if candidate's log is at least as up-to-date as ours
	myIndex, myTerm := rf.lastLogIndexTerm()
	upToDate := false
	if args.LastLogTerm != myTerm {
		upToDate = args.LastLogTerm > myTerm
	} else {
		upToDate = args.LastLogIndex >= myIndex
	}

	if alreadyVoted || !upToDate {
		// vote to candidate log is old
		reply.VoteGranted = false
		return
	}

	// vote for this peer
	rf.votedFor = args.CandidateID
	reply.VoteGranted = true
	shouldReset = true
}

// sendRequestVote
// ===============
//
// Example code to send a RequestVote RPC to a server.
//
// server int -- index of the target server in rf.peers[]
//
// args *RequestVoteArgs -- RPC arguments in args
//
// reply *RequestVoteReply -- RPC reply
//
// The types of args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers)
//
// The rpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost
//
// Call() sends a request and waits for a reply.
//
// If a reply arrives within a timeout interval, Call() returns true;
// otherwise Call() returns false
//
// Thus Call() may not return for a while.
//
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply
//
// Call() is guaranteed to return (perhaps after a delay)
// *except* if the handler function on the server side does not return
//
// Thus there
// is no need to implement your own timeouts around Call()
//
// Please look at the comments and documentation in ../rpc/rpc.go
// for more details
//
// If you are having trouble getting RPC to work, check that you have
// capitalized all field names in the struct passed over RPC, and
// that the caller passes the address of the reply struct with "&",
// not the struct itself
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// PutCommand
// =====
//
// The service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log
//
// If this server is not the leader, return false.
//
// Otherwise start the agreement and return immediately.
//
// # There is no guarantee that this command will ever be committed to the Raft log, since the leader may fail or lose an election
//
// # The first return value is the index that the command will appear at if it is ever committed
//
// The second return value is the current term.
//
// The third return value is true if this server believes it is the leader
func (rf *Raft) PutCommand(command interface{}) (int, int, bool) {
	rf.mux.Lock()
	if rf.state != stateLeader {
		term := rf.currentTerm
		rf.mux.Unlock()
		return -1, term, false
	}
	term := rf.currentTerm
	index := len(rf.log)
	rf.log = append(rf.log, LogEntry{Term: term, Command: command})
	rf.matchIndex[rf.me] = index
	rf.nextIndex[rf.me] = index + 1
	rf.mux.Unlock()

	go rf.broadcastAppendEntries()
	return index, term, true
}

func (rf *Raft) broadcastAppendEntries() {
	rf.mux.Lock()
	if rf.state != stateLeader {
		rf.mux.Unlock()
		return
	}
	term := rf.currentTerm
	commit := rf.commitIndex
	peers := len(rf.peers)
	rf.mux.Unlock()

	for i := 0; i < peers; i++ {
		if i == rf.me {
			continue
		}
		go rf.replicatePeer(i, term, commit)
	}
}

// replicatePeer continuously sends AppendEntries until follower accepts or leader changes.
func (rf *Raft) replicatePeer(server int, term int, commit int) {
	for {
		rf.mux.Lock()
		// leader may have changed
		if rf.state != stateLeader || rf.currentTerm != term {
			rf.mux.Unlock()
			return
		}

		args := rf.buildAppendArgs(server, commit)
		rf.mux.Unlock()

		reply := &AppendEntriesReply{}
		if !rf.sendAppendEntries(server, &args, reply) {
			return // network fail, give up for this round
		}

		rf.mux.Lock()

		// outdated term → revert to follower
		if reply.Term > rf.currentTerm {
			rf.becomeFollower(reply.Term)
			rf.mux.Unlock()
			rf.resetElectionTimer()
			return
		}

		// leadership lost → stop retrying
		if rf.state != stateLeader || args.Term != rf.currentTerm {
			rf.mux.Unlock()
			return
		}

		// success → advance indexes & try committing
		if reply.Success {
			match := args.PrevLogIndex + len(args.Entries)
			rf.matchIndex[server] = match
			rf.nextIndex[server] = match + 1
			rf.updateCommitIndex()
			rf.mux.Unlock()
			return
		}

		// otherwise backtrack nextIndex
		if rf.nextIndex[server] > 1 {
			rf.nextIndex[server]--
		}
		rf.mux.Unlock()
	}
}

// buildAppendArgs builds AppendEntries RPC args to a follower.
func (rf *Raft) buildAppendArgs(server int, commit int) AppendEntriesArgs {
	nextIdx := rf.nextIndex[server]
	if nextIdx < 1 {
		nextIdx = 1
	}
	if nextIdx > len(rf.log) {
		nextIdx = len(rf.log)
	}

	prevIdx := nextIdx - 1

	var entries []LogEntry
	if nextIdx < len(rf.log) {
		entries = append([]LogEntry(nil), rf.log[nextIdx:]...)
	}

	return AppendEntriesArgs{
		Term:         rf.currentTerm,
		LeaderID:     rf.me,
		PrevLogIndex: prevIdx,
		PrevLogTerm:  rf.log[prevIdx].Term,
		Entries:      entries,
		LeaderCommit: commit,
	}
}

// updateCommitIndex advances commit index when a majority replicated new entries.
func (rf *Raft) updateCommitIndex() {
	lastIdx := len(rf.log) - 1
	majority := len(rf.peers)/2 + 1
	for idx := lastIdx; idx > rf.commitIndex; idx-- {
		if rf.log[idx].Term != rf.currentTerm {
			continue
		}
		count := 1 // include leader
		for i := range rf.peers {
			if i == rf.me {
				continue
			}
			if rf.matchIndex[i] >= idx {
				count++
			}
		}
		if count >= majority {
			rf.commitIndex = idx
			if rf.applyCond != nil {
				rf.applyCond.Broadcast()
			}
			break
		}
	}
}

// Stop
// ====
//
// The tester calls Stop() when a Raft instance will not
// be needed again
//
// You are not required to do anything
// in Stop(), but it might be convenient to (for example)
// turn off debug output from this instance
func (rf *Raft) Stop() {
	rf.mux.Lock()
	if rf.dead {
		rf.mux.Unlock()
		return
	}
	rf.dead = true // mark as stopped so future calls no-op
	if rf.applyCond != nil {
		rf.applyCond.Broadcast()
	}
	close(rf.killCh)
	rf.mux.Unlock()
	rf.resetElectionTimer() // nudge timer goroutine so it notices killCh closure
}

// AppendEntriesArgs defines the arguments for AppendEntries RPCs.
type AppendEntriesArgs struct {
	Term         int        // leader's current term
	LeaderID     int        // id of the leader sending this
	PrevLogIndex int        // index before the new entries
	PrevLogTerm  int        // term at PrevLogIndex
	Entries      []LogEntry // entries to store (empty = heartbeat)
	LeaderCommit int        // leader's commit index
}

// AppendEntriesReply defines the reply for AppendEntries RPCs.
type AppendEntriesReply struct {
	Term    int  // responder's current term
	Success bool // true if follower accepted the log
}

// AppendEntries applies the leader's heartbeat or log replication request.
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	shouldReset := false

	rf.mux.Lock()
	defer func() {
		rf.mux.Unlock()
		if shouldReset {
			rf.resetElectionTimer()
		}
	}()

	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	if args.Term > rf.currentTerm {
		rf.becomeFollower(args.Term)
	}

	reply.Term = rf.currentTerm
	rf.state = stateFollower

	// Record contact from current leader to avoid starting elections.
	shouldReset = true

	if args.PrevLogIndex >= len(rf.log) {
		reply.Success = false
		return
	}

	if rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		reply.Success = false
		return
	}

	// Append any new entries, removing conflicts first.
	insertIndex := args.PrevLogIndex + 1
	for i, entry := range args.Entries {
		if insertIndex < len(rf.log) {
			if rf.log[insertIndex].Term != entry.Term {
				rf.log = rf.log[:insertIndex]
				rf.log = append(rf.log, args.Entries[i:]...)
				goto appended
			}
			insertIndex++
			continue
		}
		rf.log = append(rf.log, args.Entries[i:]...)
		goto appended
	}

appended:
	reply.Success = true

	if args.LeaderCommit > rf.commitIndex {
		lastNew := len(rf.log) - 1
		if args.LeaderCommit < lastNew {
			rf.commitIndex = args.LeaderCommit
		} else {
			rf.commitIndex = lastNew
		}
		if rf.applyCond != nil {
			rf.applyCond.Broadcast()
		}
	}
}

// sendAppendEntries issues an AppendEntries RPC to the given peer.
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}


// startElection increase the term and sends vote requests to peers.
func (rf *Raft) startElection() {
	rf.currentTerm++
	termStarted := rf.currentTerm
	rf.state = stateCandidate
	rf.votedFor = rf.me // vote for self

	lastIndex, lastTerm := rf.lastLogIndexTerm()
	majority := len(rf.peers)/2 + 1

	votes := 1 // self-vote already counted

	for server := range rf.peers {
		if server == rf.me {
			continue
		}

		args := &RequestVoteArgs{
			Term:         termStarted,
			CandidateID:  rf.me,
			LastLogIndex: lastIndex,
			LastLogTerm:  lastTerm,
		}

		go func(server int, args *RequestVoteArgs) {
			reply := &RequestVoteReply{}
			ok := rf.sendRequestVote(server, args, reply)
			if !ok {
				return
			}

			resetTimer := false
			startHeartbeat := false

			rf.mux.Lock()

			if reply.Term > rf.currentTerm {
				rf.becomeFollower(reply.Term)
				resetTimer = true
			} else if termStarted == rf.currentTerm && rf.state == stateCandidate {
				if reply.VoteGranted {
					votes++
					if votes >= majority {
						// enough votes: we are the new leader
						rf.becomeLeader()
						startHeartbeat = true
					}
				}
			}

			rf.mux.Unlock()

			if resetTimer {
				rf.resetElectionTimer()
				return
			}
			if startHeartbeat {
				// start periodic heartbeats once leadership established
				rf.startLeaderHeartbeats()
			}
		}(server, args)
	}
}

// startLeaderHeartbeats kicks off periodic AppendEntries RPCs while leader.
func (rf *Raft) startLeaderHeartbeats() {
	rf.mux.Lock()
	if rf.state != stateLeader {
		rf.mux.Unlock()
		return
	}
	term := rf.currentTerm // capture term so we can notice if we lose leadership
	rf.mux.Unlock()

	rf.broadcastAppendEntries()

	go func(term int) {
		ticker := time.NewTicker(rf.heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rf.mux.Lock()
				if rf.state != stateLeader || rf.currentTerm != term || rf.killed() {
					// leadership changed or server stopped, stop sending heartbeats
					rf.mux.Unlock()
					return
				}
				rf.mux.Unlock()
				rf.broadcastAppendEntries()
			case <-rf.killCh:
				return
			}
		}
	}(term)
}

// randomElectionTimeout picks a randomized timeout to avoid split votes.
func (rf *Raft) randomElectionTimeout() time.Duration {
	rf.randMux.Lock()
	defer rf.randMux.Unlock()

	min := rf.electionTimeoutMin
	max := rf.electionTimeoutMax
	if max <= min {
		return min
	}
	delta := max - min
	return min + time.Duration(rf.rand.Int63n(int64(delta)))
}

// becomeLeader switches state to leader and initializes replication indexes.
func (rf *Raft) becomeLeader() {
	rf.state = stateLeader
	lastIndex := len(rf.log) - 1
	for i := range rf.nextIndex {
		rf.nextIndex[i] = len(rf.log)
		rf.matchIndex[i] = 0
	}
	rf.matchIndex[rf.me] = lastIndex
}

// killed reports whether Stop has been called.
func (rf *Raft) killed() bool {
	return rf.dead
}

// resetElectionTimer notifies the ticker to restart its timeout.
func (rf *Raft) resetElectionTimer() {
	select {
	case rf.electionResetCh <- struct{}{}:
	default:
	}
}

// becomeFollower updates state to follower for the given term.
func (rf *Raft) becomeFollower(term int) {
	rf.currentTerm = term
	rf.state = stateFollower
	rf.votedFor = -1
}

// lastLogIndexTerm returns the index and term of the latest log entry.
func (rf *Raft) lastLogIndexTerm() (int, int) {
	if len(rf.log) == 0 {
		return 0, 0
	}
	lastIndex := len(rf.log) - 1
	return lastIndex, rf.log[lastIndex].Term
}

// runApplier delivers committed log entries to the service in order.
func (rf *Raft) runApplier() {
	for {
		rf.mux.Lock()
		for !rf.dead && rf.lastApplied >= rf.commitIndex {
			rf.applyCond.Wait()
		}
		if rf.dead && rf.lastApplied >= rf.commitIndex {
			rf.mux.Unlock()
			return
		}
		index := rf.lastApplied + 1
		entry := rf.log[index]
		rf.lastApplied = index
		rf.mux.Unlock()

		rf.applyCh <- ApplyCommand{Index: index, Command: entry.Command}
	}
}
