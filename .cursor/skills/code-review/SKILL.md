---
name: code-review
description: Use when reviewing a git diff, PR, MR, completed implementation task, or an entire project. Supports change review and whole-project review to detect bugs, security risks, performance issues, missing tests, architecture problems, engineering-standard violations, and maintainability issues.
---

# Code Review

Review code as a senior engineer, architecture reviewer, and quality owner.

This skill supports two review modes:

1. Change Review: review a git diff, PR, MR, commit, staged changes, unstaged changes, or the current implementation change.
2. Whole-Project Review: review the architecture, code quality, security, tests, dependencies, configuration, engineering standards, and maintainability of an entire project.

Both modes require careful and thorough review. The review target is different, but the review rigor is not.

The goal is to find real risks, not to provide a vague assessment. Reviews must be specific, evidence-based, and verifiable.

Core principles:

- Select the review mode based on the user's request.
- Review based on code evidence, project standards, business requirements, and the existing architecture.
- Do not review based on personal taste alone.
- Do not only check functional bugs. Also check engineering standards, code style, architectural boundaries, test quality, dependency safety, configuration safety, and maintainability.
- Use English for the entire report and chat summary.
- Severity names must be exactly: Critical, Important, Minor.

Do not modify code. Do not rewrite files. Do not apply patches. This skill only produces review findings and recommendations.

## Required Review File

You must save the code review result to a Markdown file.

Default output path:

```bash
docs/code-reviews/YYYY-MM-DD-HH-MM-code-review.md
```

If the user provides a different output path, use the user's path. If the directory does not exist, create it first.

After writing the file, tell the user the exact path of the saved review report.

Do not only output the review in chat. The saved review file is the source of record.

## Language Requirement

The review report must be written in English.

The following must also be written in English:

- Headings
- Section names
- Field names
- Problem descriptions
- Impact analysis
- Recommendations
- Evidence explanations
- Final assessment
- Chat summary after saving the file

Allowed non-English content:

- Original code comments or strings when quoting code
- Original business terms if they are part of the codebase
- User-provided identifiers or requirement names

Do not mix Chinese template fields into the English report. Use "Change intent", "Impact area", and "Ready to merge or proceed" rather than Chinese field names.

## Review Mode Selection

Select the review mode based on the user's request.

### Use Change Review when the request means:

- Review a git diff, PR, MR, current change, recent commit, staged changes, or unstaged changes
- Review the implementation of a feature, bug fix, refactor, or task
- Check whether a change is ready to merge
- The user explicitly says "diff only", "changes only", or "review only this change"

If the user does not explicitly ask for a whole-project review and a diff exists, prefer Change Review.

### Use Whole-Project Review when the request means:

- Review the entire project
- Review the whole codebase
- Perform a project health check
- Check project architecture, code quality, security risk, maintainability, test coverage, or overall project health
- The user explicitly says "entire project", "whole project", "whole repository", "full repo", "full codebase", or "project-level review"

Whole-Project Review is not a line-by-line review of every file. It is a systematic sampled review based on project structure, critical paths, core modules, configuration, dependencies, tests, and risk areas.

### If the mode is unclear:

1. If the user mentions diff, PR, MR, commit, changes, or modifications, use Change Review.
2. If the user mentions project, repository, architecture, overall quality, or health check, use Whole-Project Review.
3. If it is still unclear, prefer Change Review and note the inferred scope in the "Missing Context" section.

Do not stop just because the mode is unclear. Choose the most reasonable mode based on available information and proceed.

---

## Efficiency Rule: Read Files On Demand, Not All At Once

This is the main rule for controlling review time. Do not read every potentially relevant file before reviewing. Load files on demand in this order.

### Step 1: Read only targeted files

For Change Review: first read `git diff --stat` to understand the size of the change, then read the diff body and directly related test files.

For Whole-Project Review: first read the README, project directory structure using `find . -maxdepth 3`, and build files.

After Step 1, decide whether more files need to be read based on risk signals.

### Step 2: Read additional files based on risk signals

Only read additional files when Step 1 reveals one of these signals:

| Risk signal | Additional files to read |
| --- | --- |
| Database migration, schema, SQL, or ORM changes | Migration files, related entities, repositories |
| Authentication, authorization, or permission logic changes | Auth-related code, permission configuration |
| External HTTP calls, command execution, file upload, deserialization | Security-sensitive path context |
| Function signature or return type changes | Direct callers, limited to 2 levels |
| Configuration changes | Related environment config and deployment files |
| Missing tests or test changes | Test directory structure and test config |
| New dependency introduced | Dependency file and version list |
| Unclear project standards | Standard config files such as ESLint, Checkstyle, Ruff, etc. |

### Step 3: Additional reading for Whole-Project Review

After building the project map, read files gradually in this priority order. Stop reading a category once you have enough evidence to assess its risk.

1. Entry files and startup configuration
2. Authentication and authorization code
3. Core service modules, sampled
4. Data access layer, sampled
5. Configuration files and environment examples
6. Shared utilities and middleware, sampled
7. Logging and global error handling
8. Test directory and test configuration
9. Dockerfile, CI configuration, deployment scripts

You do not need to read every file before starting the review. Read enough to make evidence-based findings, then clearly state uncovered areas in the report.

### Call chain depth limit

When tracing call chains, trace at most 2 levels: direct caller plus one caller above it.

If the function is a public API entry point, a database write operation, or a security-sensitive path, you may trace up to 3 levels, but you must state the reason in the report.

---

## Pre-Review Preparation

### Common preparation for all modes

If project standard files exist, review against them first. Common standard files include:

```text
.editorconfig / checkstyle.xml / spotless.gradle / pom.xml / build.gradle
eslint.config.js / .eslintrc / .prettierrc / ruff.toml / pyproject.toml
mypy.ini / golangci.yml / CONTRIBUTING.md / AGENTS.md / CLAUDE.md / README.md
```

Only read standard files during Step 2 additional reading. Do not collect all of them during Step 1 unless they are directly needed.

### Step 1 input for Change Review

```bash
git diff --stat HEAD~1..HEAD
git diff HEAD~1..HEAD
```

Or adjust the range based on the user's requested scope:

```bash
git diff --stat origin/main...HEAD && git diff origin/main...HEAD
git diff --cached --stat && git diff --cached
git diff --stat && git diff
```

### Step 1 input for Whole-Project Review

```bash
find . -maxdepth 4 -type f \
  -not -path "*/.git/*" \
  -not -path "*/node_modules/*" \
  -not -path "*/target/*" \
  -not -path "*/build/*" \
  -not -path "*/dist/*" \
  -not -path "*/.gradle/*" \
  -not -path "*/.venv/*" \
  | sort
```

### Search commands to use on demand

Only run these after risk signals are identified. Do not run all of them at the start.

```bash
# Security-sensitive patterns
rg -n "password|passwd|secret|token|apiKey|privateKey|eval\(|exec\(|Runtime\.getRuntime|ProcessBuilder|deserialize|ObjectInputStream|SELECT .*\+" .

# Exception and resource handling patterns
rg -n "return null|catch \(.*Exception|catch \(Throwable|throws Exception|printStackTrace|TODO|FIXME" .

# Specific call-site search
grep -rn "functionName" src/
```

---

## Change Review Workflow

Follow these steps in order. Do not skip steps, and do not read ahead unnecessarily.

### 1. Read Step 1 files and understand the change

Read the diff and determine what changed, why it changed, which modules are affected, whether it matches the requirement, the change type, and which risk signals require Step 2 additional reading.

If no requirement is provided, infer the likely intent from the diff and mark it as an inference.

### 2. Build a change inventory

List:

- Changed files
- Change type for each file
- Key functions or methods
- Identified risk signals
- Files that need Step 2 additional reading

### 3. Read additional files based on risk signals

Read additional files according to the risk signals identified in the change inventory. After reading each category of files, update the corresponding findings.

### 4. Check call chains

For meaningful changed functions, inspect surrounding code, error paths, data flow, transaction boundaries if applicable, concurrency behavior if applicable, and resource cleanup if applicable.

Do not exceed 2 levels unless an exception applies: public API entry point, database write operation, or security-sensitive path.

### 5. Execute the mandatory review matrix

Check every dimension in the review matrix. Do not skip any dimension. If a dimension does not apply, mark it as "Not applicable" with a reason.

---

## Whole-Project Review Workflow

### 1. Read Step 1 files and build the project map

Identify:

- Technology stack and languages
- Build system
- Project entry points
- Main modules
- Data storage approach
- External dependencies
- Deployment approach
- Test framework
- Configuration management approach
- Code standard tooling

State in the report which directories and file types were actually inspected.

### 2. Identify high-risk areas

Prioritize:

- Authentication and authorization
- User input handling
- Database access
- File upload and download
- External HTTP requests
- Command execution
- Serialization and deserialization
- Caching and concurrency
- Scheduled and asynchronous jobs
- Critical business paths such as payment, orders, permissions, accounts, import, and export
- Global error handling
- Logging and sensitive data handling
- Configuration and secret management
- Dependency version risk
- Test gaps
- Deployment configuration

### 3. Read additional files by priority and sample core modules

Use the Step 3 priority list from the Efficiency Rule. Review the most important modules deeply instead of spreading attention evenly.

For each sampled module, inspect call chains, error paths, boundary conditions, standard compliance, and tests. The call chain depth limit still applies.

### 4. Check overall engineering quality

Assess:

- Whether architectural layering is clear
- Whether module boundaries are clear
- Whether dependency direction is reasonable
- Whether error handling is consistent
- Whether logs help debugging without leaking sensitive data
- Whether configuration is safe and maintainable
- Whether tests cover core risks
- Whether CI is effective
- Whether code standards are enforced by tools

### 5. Produce the whole-project risk conclusion

Must include:

- Overall project health
- Top 3 to 5 highest-risk areas
- Standards compliance status
- Priority repair path
- Whether any issue blocks release or further development
- Areas that need deeper specialized review

---

## Mandatory Review Matrix

Use this review matrix for both Change Review and Whole-Project Review.

The report must include a "Review Matrix Results" section explaining whether each dimension was checked, whether issues were found, and the evidence location or reason for non-coverage.

### 1. Functional Correctness

Check for logic errors, missing branches, wrong assumptions, broken existing behavior, null or empty value handling issues, boundary condition bugs, off-by-one errors, incorrect exception handling, incorrect error propagation, wrong return values, state inconsistency, race conditions, transaction bugs, resource leaks, backward compatibility breaks, and unexpected behavior changes.

### 2. Security

Check for missing input validation, incorrect authorization, permission bypass, SQL injection, command injection, path traversal, XSS, CSRF, SSRF, unsafe deserialization, sensitive data in logs, secret leakage, weak cryptography, unsafe file handling, and unsafe dependency or library usage.

Only report a security issue when there is a plausible failure path in the current codebase. Explain the attack path or risk.

### 3. Performance

Check for N+1 queries, missing indexes for new queries, unbounded loops, inefficient algorithms, excessive memory use, missing pagination, excessive network calls, blocking work in latency-sensitive paths, cache misuse, inefficient batch operations, connection pool pressure, thread pool pressure, and repeated expensive computation.

### 4. Maintainability

Check for unclear naming, oversized methods or classes, mixed responsibilities, duplicated logic, tight coupling, unclear module boundaries, unnecessary abstraction, missing abstraction when duplication is real, hard-coded constants, inconsistency with project conventions, unclear error messages, comments that explain confusing code instead of making code clearer, and code that is hard to test or extend.

### 5. Architecture and Layering

Check for violations of existing layers, wrong dependency direction, leaky abstractions, inconsistent API design, unclear domain boundaries, breaking changes, missing migration strategy, unclear configuration ownership, framework misuse, conflict with nearby project patterns, and changes that should be split into smaller changes.

### 6. Database and Transactions

If database-related code exists, check query correctness, index usage, N+1 risk, transaction boundaries, locking behavior, concurrency behavior, data consistency, integrity constraints, migration safety, rollback behavior, backfill requirements, pagination, nullability, default values, constraint correctness, and compatibility with existing data.

Data loss, unsafe migrations, and consistency bugs must be marked Critical.

### 7. Testing

Check whether tests cover changed behavior, core business logic, boundary cases, failure paths, security-sensitive paths, whether integration tests are needed, whether mocks are overused, whether tests only verify implementation details, whether bug fixes include regression tests, whether test names express behavior, and whether test data is stable and maintainable.

When a change affects behavior, correctness, security, data, or complex logic, missing tests should be marked Important.

### 8. Engineering Standards

Check whether the code follows project naming conventions, directory and package structure conventions, formatting rules, lint rules, exception handling standards, logging standards, comment standards, API response standards, error code or error response standards, dependency management standards, and commit or versioning conventions when relevant.

If the project has no explicit standards, judge based on existing local style and mainstream practice, and state that basis.

### 9. Java or JVM-Specific Checks

If the project uses Java or a JVM stack, additionally check null misuse, equals and hashCode consistency, BigDecimal precision, date and time zone handling, concurrent collection modification, Stream readability, overly broad exception handling such as catch Exception, swallowed exceptions or printStackTrace-only handling, transaction annotation placement, Spring bean lifecycle and injection, business logic in controllers, oversized services, repository query clarity, DTO/Entity/VO boundaries, Lombok risks, and parameterized logging.

### 10. Python-Specific Checks

If the project uses Python, additionally check type annotation sufficiency, mutable default arguments, overly broad exception handling, safe path handling, pinned dependency versions, synchronous blocking in async paths, unsafe deserialization, dynamic execution risk, and test coverage of key branches.

### 11. Frontend or Node.js-Specific Checks

If the project uses JavaScript, TypeScript, or a frontend framework, additionally check any misuse, state management clarity, oversized component responsibilities, effect cleanup, async request error and cancellation handling, XSS risk, dependency security and bundle size risk, frontend environment variable leakage, build configuration reasonableness, and test coverage for interactions and boundary states.

---

## Severity Levels

### Critical

Must be fixed before proceeding. Applies to security vulnerabilities, data loss risk, data corruption, broken core functionality, incorrect authorization, serious concurrency bugs, serious transaction bugs, production crash risk, unsafe migration, severe performance regression on a critical path, likely null pointer or similar runtime failure, and thread-safety issues with realistic production impact.

### Important

Should be fixed before merge, release, or the next task. Applies to missing requirements, important unhandled edge cases, poor error handling, meaningful test gaps, architecture problems, standard violations that may cause maintenance cost or defects, moderate performance risk, backward compatibility concerns, and behavior inconsistent with existing code.

### Minor

Nice to have. Should not block progress by itself. Applies to minor style issues, naming improvements, small refactors, documentation improvements, small optimizations, non-blocking consistency improvements, and test readability improvements.

### Severity for standards issues

- If a standards issue may cause security, data, concurrency, transaction, or runtime failures, mark it Critical.
- If a standards issue may cause maintenance difficulty, testing difficulty, misuse risk, or significant team collaboration cost, mark it Important.
- If a standards issue is only minor style inconsistency and does not affect understanding or maintainability, mark it Minor.

---

## Review Responsibilities

You should: choose Change Review or Whole-Project Review based on the user's request; identify code issues, standards issues, and systemic risks; provide specific recommendations; explain technical reasoning; inspect repository source when needed; locate findings to file and line when possible; inspect call chains for important behavior changes within the depth limit; check existing project standards and review against them; provide a clear merge, release, or proceed recommendation; save the final review as a Markdown file.

You should not: directly modify code; rewrite the implementation; execute code changes; claim a line-by-line review of the whole project unless actually done; use Chinese template fields in an English report; read every potentially relevant file at the beginning of the review.

---

## Special Cases

Legacy code: focus on new and modified behavior. Do not require broad cleanup unless this change increases legacy risk.

Urgent fix: prioritize correctness, security, and regression coverage. Non-critical style issues should be Minor.

Refactor: focus on behavior preservation, API compatibility, test coverage, and accidental behavior changes.

New feature: review requirement coverage, extensibility, testing, error handling, security, performance, and production readiness.

Large PR, more than 500 changed lines: review architecture, data flow, public interfaces, database changes, and security-sensitive paths first, then sample lower-risk files. Clearly state uncovered areas. Do not skip the review matrix because the PR is large.

Small PR, fewer than 50 changed lines: review precisely. Focus on whether intent and tests match the change.

Test-only change: check whether tests assert meaningful behavior, fail when the bug exists, and avoid over-mocking.

Configuration change: check environment differences, defaults, secrets, compatibility, deployment impact, rollback, and documentation.

---

## Report File Format

The saved Markdown report must use the following structure.

Sections with no content may be omitted, but do not omit "Review Matrix Results", "Review Scope", or "Final Assessment".

```markdown
# Code Review Report

## Review Overview

- Review mode: Change Review / Whole-Project Review
- Review target: ...
- Change intent or project goal: ...
- Impact area or review scope: ...
- Requirement match: Matches / Partially matches / Does not match / Unknown / Not applicable
- Tests checked: Yes / No / Not found
- Project standards checked: Yes / No / Not found
- Overall rating: X/5

## Review Scope

- Files or directories checked: ...
- Key modules checked: ...
- Call chains checked and depth: ...
- Standard files checked: ...
- Uncovered scope: ...

## Review Matrix Results

| Review dimension | Checked | Issues found | Evidence or notes |
| --- | --- | --- | --- |
| Functional correctness | Yes / No | Yes / No | ... |
| Security | Yes / No | Yes / No | ... |
| Performance | Yes / No | Yes / No | ... |
| Maintainability | Yes / No | Yes / No | ... |
| Architecture and layering | Yes / No | Yes / No | ... |
| Database and transactions | Yes / No / Not applicable | Yes / No / Not applicable | ... |
| Testing | Yes / No | Yes / No | ... |
| Engineering standards | Yes / No | Yes / No | ... |
| Language or framework-specific checks | Yes / No / Not applicable | Yes / No / Not applicable | ... |

## Missing Context

(Omit this section if none.)

## Strengths

(Omit this section if none.)

## Critical Issues

(Write "None found." if none.)

### 1. Title

- Severity: Critical
- Category: Functional correctness / Security / Performance / Maintainability / Architecture / Database / Testing / Engineering standards / Other
- Location: path/to/file.ext:line
- Problem: ...
- Impact: ...
- Recommendation: ...
- Evidence: ...

## Important Issues

(Write "None found." if none.)

## Minor Issues

(Write "None found." if none.)

## Standards Issues Summary

(Omit this section if none.)

- Naming standards: ...
- Directory and package structure: ...
- Exception handling standards: ...
- Logging standards: ...
- Testing standards: ...
- Configuration standards: ...
- Dependency management standards: ...

## Overall Recommendations

- ...

## Follow-Up Recommendations

(Omit this section if none.)

- Highest-priority fixes: ...
- Recommended specialized review: ...
- Recommended additional tests: ...
- Recommended automated checks: ...

## Final Assessment

Ready to merge or proceed: Yes / No / With fixes / Needs further review

Reasoning: ...
```

## Chat Response After Saving The File

```text
Code review saved to: <path>
Review mode: Change Review / Whole-Project Review
Ready to merge or proceed: Yes / No / With fixes / Needs further review
Critical: <count>
Important: <count>
Minor: <count>
```

Do not repeat the full report in chat unless the user asks.

## Rating Guide

- 5/5: Production-ready, no meaningful issues found, and review coverage was sufficient.
- 4/5: Overall good, only Minor issues.
- 3/5: Mostly usable, but Important issues need fixing.
- 2/5: Clear correctness, design, testing, security, standards, or maintainability problems.
- 1/5: Unsafe or not ready; Critical issues are present.

If any Critical issue exists, the rating should usually be 1/5 or 2/5. If unresolved Important issues exist, the rating should usually not exceed 3/5. For Whole-Project Review with limited coverage, the rating must reflect uncertainty.

## Issue Quality Standard

Every issue must answer: Where is the problem? What is wrong? Why does it matter? How should it be fixed? Is this based on code evidence, project standards, or a reasonable assumption?

Bad example: `Improve error handling.`

Good example:

```text
- Severity: Important
- Category: Engineering standards
- Location: src/user/UserService.java:84
- Problem: createUser catches SQLException and returns null, which is inconsistent with the exception handling style used in the same class.
- Impact: Callers cannot distinguish duplicate email, database outage, and unexpected failure. This can cause incorrect success handling or a NullPointerException downstream.
- Recommendation: Return a typed error result or throw a domain-specific exception consistent with the current service style.
- Evidence: getUserById in the same class throws UserRepositoryException for database failures.
```

## Final Checklist Before Responding

Before returning the review, confirm:

- You selected Change Review or Whole-Project Review based on the user's request.
- You read files on demand and did not read every file up front.
- Call chain tracing did not exceed the depth limit: 2 levels, or the exception reason was documented.
- Review scope is clear and uncovered scope is listed.
- Project standards were checked, or "Not found" was stated.
- Review Matrix Results were included with all 9 dimensions.
- The review is based on code evidence and available requirements.
- Every issue has severity, category, location when possible, impact explanation, and actionable recommendation.
- Critical issues are truly Critical.
- Important issues were not downgraded to Minor.
- Issue groups with no findings say "None found."
- The final assessment is clear and includes a merge or proceed recommendation.
- The report was saved as a Markdown file.
- The chat response includes the saved report path.
- The report uses English template fields.

## Final Rule

Select review mode based on the user's request: Change Review for diffs and Whole-Project Review for the entire project. Review carefully in both modes.

Read files on demand. Do not read every file at the beginning. First read targeted files, then read more based on risk signals.

Trace call chains at most 2 levels. If deeper tracing is needed, state that in the report.

Find real risks. Also find standards issues. Be specific. Output English. Do not edit code. Save the report to a Markdown file. Give a clear conclusion.
