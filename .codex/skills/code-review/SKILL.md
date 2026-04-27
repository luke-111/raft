---
name: code-review
description: Use when reviewing git diffs, pull requests, merge requests, or completed implementation work to detect bugs, security risks, performance issues, missing tests, and maintainability problems
---

# Code Review

Review code changes as a senior engineer and architecture reviewer.

The goal is to evaluate whether the change is correct, safe, maintainable, tested, and ready to merge.

Core principle: review the actual diff against the intended requirement. Do not review by taste alone.

Do not modify code. Do not rewrite files. Do not apply patches. This skill only produces review findings and recommendations.

## Required Output File

You MUST save the review result to a Markdown file.

Default output path:

```bash
docs/code-reviews/YYYY-MM-DD-HH-MM-code-review.md
```

If the user provides a different output path, use the user's path.

If the directory does not exist, create it.

After writing the file, tell the user the exact path of the saved report.

Do not only print the review in chat. The file is the source of record.

## When To Use

Use this skill when:

- Reviewing a git diff
- Reviewing a pull request or merge request
- Reviewing a completed implementation task
- Reviewing code before merge
- Reviewing a bug fix
- Reviewing a refactor
- Reviewing generated or agent-written code
- Checking whether implementation matches a plan, ticket, spec, or requirement

If the user asks for "code review", "review this diff", "check this PR", "detect bugs", "review my changes", or similar, use this skill.

## Required Review Mindset

Review with these principles:

- Be objective: base findings on code evidence, requirements, and project conventions.
- Be constructive: provide specific, practical recommendations.
- Be educational: explain why the issue matters when it is not obvious.
- Be complete: consider correctness, security, performance, maintainability, architecture, database behavior, and tests.
- Be precise: cite file and line references whenever possible.
- Be skeptical but fair: do not invent risks that are not plausible in this codebase.
- Be severity-aware: do not mark style issues as critical.

## Review Inputs

Before producing the review, gather as much of this as available:

- Git diff
- PR or MR description
- Ticket, plan, spec, or requirement
- Changed source files
- Nearby unchanged code around modified lines
- Direct callers of changed functions or methods
- Direct callees used by changed functions or methods
- Related tests
- Existing project conventions
- Database migrations, schemas, queries, and ORM mappings when relevant
- Configuration changes when relevant

If important context is missing, continue with the review using available evidence and explicitly list the missing context in the report.

## Useful Git Commands

For the latest commit:

```bash
git diff --stat HEAD~1..HEAD
git diff HEAD~1..HEAD
```

For changes against main:

```bash
git fetch origin
git diff --stat origin/main...HEAD
git diff origin/main...HEAD
```

For staged changes:

```bash
git diff --cached --stat
git diff --cached
```

For unstaged changes:

```bash
git diff --stat
git diff
```

For changed files:

```bash
git diff --name-only origin/main...HEAD
```

For call-site investigation, use repository search tools such as grep, rg, or language-aware search.

## Review Workflow

Follow this process in order.

### 1. Understand The Change

Determine:

- What changed
- Why it changed
- What business or technical goal it appears to serve
- Which modules, layers, APIs, jobs, data models, or configuration files are affected
- Whether the change matches the stated requirement
- Whether there is scope creep beyond the stated requirement
- Whether the change is a bug fix, new feature, refactor, test-only change, config change, migration, or mixed change

If no requirement is provided, infer the likely intent from the diff and label it as an inference.

### 2. Inspect The Impact Area

For each meaningful changed function, method, API, query, model, configuration, or migration:

- Inspect the surrounding code.
- Inspect direct callers.
- Inspect direct callees.
- Inspect related tests.
- Inspect error handling paths.
- Inspect data flow and state changes.
- Inspect transaction boundaries when applicable.
- Inspect concurrency behavior when applicable.
- Inspect resource ownership and cleanup when applicable.

Do not review changed lines in isolation when behavior depends on surrounding code.

### 3. Check Functional Correctness

Look for:

- Logic errors
- Missing branches
- Incorrect assumptions
- Broken existing behavior
- Null or empty value handling bugs
- Boundary condition bugs
- Off-by-one errors
- Incorrect exception handling
- Incorrect error propagation
- Incorrect return values
- State inconsistency
- Race conditions
- Transaction bugs
- Resource leaks
- Backward compatibility breaks
- Unexpected behavior changes

Prioritize real defects over style comments.

### 4. Check Security

Look for:

- Missing input validation
- Incorrect authorization
- Permission bypass
- SQL injection
- Command injection
- Path traversal
- XSS
- CSRF
- SSRF
- Unsafe deserialization
- Sensitive data in logs
- Secret leakage
- Weak cryptography
- Unsafe file handling
- Unsafe dependency or library usage

Only flag a security issue when there is a plausible failure mode. Explain the attack path or risk.

### 5. Check Performance

Look for:

- N+1 queries
- Missing indexes for new queries
- Unbounded loops
- Inefficient algorithms
- Excessive memory use
- Missing pagination
- Excessive network calls
- Blocking work in latency-sensitive paths
- Cache misuse
- Batch operation inefficiency
- Connection pool pressure
- Thread pool pressure
- Repeated expensive computation

Explain why the issue matters at realistic scale.

### 6. Check Maintainability

Look for:

- Unclear naming
- Large methods or classes
- Mixed responsibilities
- Duplicate logic
- Tight coupling
- Poor module boundaries
- Unnecessary abstraction
- Missing abstraction where duplication is real
- Hard-coded constants
- Inconsistent project conventions
- Poor error messages
- Comments that explain confusing code instead of improving the code
- Code that will be hard to test or extend

Do not over-polish urgent or very small fixes unless the maintainability risk is meaningful.

### 7. Check Architecture

Look for:

- Violations of existing layering
- Incorrect dependency direction
- Leaky abstractions
- Inconsistent API design
- Poor domain boundaries
- Breaking changes
- Missing migration strategy
- Incorrect configuration ownership
- Framework misuse
- Design that conflicts with nearby project patterns
- Scope that should be split into smaller changes

Prefer existing architecture and local conventions unless there is a clear reason to deviate.

### 8. Check Database Changes

If database code, schema, queries, migrations, repositories, or ORM mappings changed, inspect:

- Query correctness
- Index usage
- N+1 risk
- Transaction boundaries
- Locking behavior
- Concurrency behavior
- Data consistency
- Data integrity constraints
- Migration safety
- Rollback behavior
- Backfill requirements
- Pagination
- Nullability
- Default values
- Constraint correctness
- Compatibility with existing data

Flag data loss, unsafe migrations, and consistency bugs as Critical.

### 9. Check Tests

Evaluate whether tests:

- Cover changed behavior
- Cover core business logic
- Cover edge cases
- Cover failure paths
- Cover security-sensitive paths where applicable
- Cover integration behavior where unit tests are insufficient
- Use real behavior rather than only verifying mocks
- Avoid over-mocking
- Would fail without the implementation
- Protect against regression
- Match public behavior rather than implementation details

Missing tests are Important when the change affects behavior, correctness, security, data, or complex logic.

## Severity Levels

Use these exact severity levels.

### Critical

Must fix before proceeding.

Use for:

- Security vulnerability
- Data loss risk
- Data corruption
- Broken core functionality
- Incorrect authorization
- Serious concurrency bug
- Serious transaction bug
- Production crash risk
- Unsafe migration
- Severe performance regression on a critical path
- Null pointer or equivalent runtime failure in a likely path
- Thread safety issue with realistic production impact

### Important

Should fix before merge or before continuing to the next task.

Use for:

- Missing requirement
- Important edge case not handled
- Poor error handling
- Meaningful test gap
- Architecture problem
- Maintainability issue likely to cause defects
- Moderate performance risk
- Backward compatibility concern
- Inconsistent behavior with existing code
- Framework or design pattern misuse with real impact

### Minor

Nice to have. Should not block progress by itself.

Use for:

- Style issue
- Naming improvement
- Small refactor
- Documentation improvement
- Small optimization
- Non-blocking consistency improvement
- Test readability improvement

## Review Responsibilities

You should:

- Identify code problems and risks.
- Provide specific improvement suggestions.
- Explain technical reasoning.
- Inspect repository source when needed.
- Locate findings to file and line when possible.
- Inspect complete call chains for meaningful behavior changes.
- Recommend tools, tests, or follow-up checks where useful.
- Give a clear merge-readiness verdict.
- Save the final review to a Markdown file.

You should not:

- Directly modify code.
- Rewrite the implementation.
- Execute code changes.
- Make final technical decisions for the human.
- Replace human review approval.
- Provide large replacement code blocks unless the user explicitly asks.
- Review unrelated legacy code unless the change interacts with it.
- Claim tests pass unless you ran them or saw evidence.

## Special Case Handling

### Legacy Code

Focus on new and modified behavior. Do not require broad legacy cleanup unless the change makes legacy risk worse or depends on unsafe legacy behavior.

### Urgent Fix

Prioritize correctness, safety, and regression coverage. Treat style and broad refactoring as Minor unless they create real risk.

### Refactor

Focus on behavior preservation, API compatibility, test coverage, simplification, and accidental behavior changes.

### New Feature

Review requirements coverage, design fit, extensibility, tests, error handling, security, performance, and production readiness.

### Large PR

Start with architecture, data flow, public interfaces, database changes, security-sensitive paths, and critical logic. Then sample lower-risk files.

### Small PR

Review precisely. Small diffs should have clear intent, minimal scope, and appropriate tests.

### Test-Only Change

Check whether tests assert meaningful behavior, avoid over-mocking, cover edge cases, and would fail for the bug or missing behavior.

### Configuration Change

Check environment-specific behavior, defaults, secrets, compatibility, deployment impact, rollback, and whether the configuration is documented.

## Report File Format

Write the saved Markdown report using this exact structure.

```markdown
# Code Review Report

## Overview

- Change intent: ...
- Impact area: ...
- Requirement match: Matches / Partially matches / Does not match / Unknown
- Tests reviewed: Yes / No / Not found
- Overall rating: X/5

## Missing Context

- ...

## Strengths

- ...

## Critical Issues

### 1. Title

- Severity: Critical
- Location: path/to/file.ext:line
- Problem: ...
- Impact: ...
- Recommendation: ...
- Evidence: ...

## Important Issues

### 1. Title

- Severity: Important
- Location: path/to/file.ext:line
- Problem: ...
- Impact: ...
- Recommendation: ...
- Evidence: ...

## Minor Issues

### 1. Title

- Severity: Minor
- Location: path/to/file.ext:line
- Problem: ...
- Impact: ...
- Recommendation: ...
- Evidence: ...

## Recommendations

- ...

## Assessment

Ready to merge: Yes / No / With fixes

Reasoning: ...
```

If there are no findings in a section, write:

```markdown
None found.
```

## Required Chat Response After Saving The File

After saving the review report file, respond in chat with only:

```text
Code review saved to: <path>
Ready to merge: Yes / No / With fixes
Critical: <count>
Important: <count>
Minor: <count>
```

Do not duplicate the full report in chat unless the user asks.

## Rating Guide

Use the overall rating carefully.

- 5/5: Production-ready; no meaningful issues found.
- 4/5: Good implementation; only minor issues.
- 3/5: Mostly correct; important fixes needed.
- 2/5: Significant correctness, design, test, or maintainability problems.
- 1/5: Unsafe or not ready; critical issues present.

A review with any Critical issue should normally be 1/5 or 2/5.

A review with unresolved Important issues should normally not be higher than 3/5.

## Issue Quality Standard

Every issue must answer:

- Where is the problem?
- What is wrong?
- Why does it matter?
- How should it be fixed?
- Is this based on code evidence or an assumption?

Bad:

```text
Improve error handling.
```

Good:

```text
- Severity: Important
- Location: src/user/UserService.java:84
- Problem: createUser catches SQLException and returns null.
- Impact: Callers cannot distinguish duplicate email, database outage, and unexpected failure. This can cause incorrect success handling or NullPointerException downstream.
- Recommendation: Return a typed error result or throw a domain-specific exception consistent with the rest of this service.
- Evidence: getUserById in the same class throws UserRepositoryException for database failures.
```

## Final Checklist Before Responding

Before returning the review, verify:

- The review is based on the diff and available requirements.
- Every issue has a severity.
- Every issue has a location when possible.
- Every issue explains impact.
- Every issue has a practical recommendation.
- Critical issues are truly critical.
- Minor style comments are not blocking.
- Missing context is explicitly listed.
- The final merge-readiness verdict is clear.
- The report was saved to a Markdown file.
- The chat response includes the saved report path.

## Final Rule

Review the diff against the requirement.

Find real risks.

Be specific.

Do not edit code.

Save the report to a Markdown file.

Give a clear verdict.
