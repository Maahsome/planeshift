# Codex Plan

# Work Ticket

ID: PSFT-11-20260915-1923-C544B8
Title: regenerate work files

## Description
It is time to re-plan our work. It turns out that the online API Documentation includes endpoints that are only available internally. I have analyzed the source code for the open-source version from [github.com/makeplane/plane](http://github.com/makeplane/plane), and have the public API endpoints defined in the file PUBLIC_API.txt.

I would like to review the items in the `work` directory and compare that to the PUBLIC_API.txt and remove items that don't have a public API endpoint, and add any that are in the PUBLIC_API.txt file and were not already defined in the `work` items.

## Decision Protocol

Before making any architectural decision, you MUST state which section of CONSTITUTION.md governs that decision. If no section applies, flag it as uncharted area requiring team review.

## Context Anchors & Routing

* Before making architectural decisions, the Agent MUST read `CONSTITUTION.md` and `ARCHITECTURE.md`.
* Use `CONSTITUTION.md` for invariants, protected dependencies, package boundaries, security rules, and coding conventions.
* Use `ARCHITECTURE.md` for recorded decisions, consequences, and rejected approaches.
* Use `LOCAL_BUILD.md` for the documented local build workflow.

## Agent SOPs

### Jira/Ticket Tracking

* The Agent MUST include the Jira ticket ID in the branch name and commit messages, for example `feat/PROJ-123-add-widget`.

### Git Branch Conventions

* Branch names MUST use one of these prefixes: `feat/`, `fix/`, `chore/`, or `hotfix/`.
* Branches MUST be created from `main`.
* Agentic SDLC initialization MUST be performed on a new branch named `chore/init-agentic-sdlc` before the governance files are written.
* This repository uses direct commits; opening a PR/MR is not required unless explicitly requested.

### Commit Messages

* Commit messages MUST use conventional commit prefixes including `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, and `test:`.
* Commit messages MUST reference the Jira ticket where applicable.

### Cross-Cutting Process Rules

* The Agent MUST NOT introduce new third-party dependencies unless explicitly mandated.
* The Agent MUST NOT implement features or speculative abstractions not explicitly requested.
* The Agent MUST NOT silently remove existing working code or tests to resolve errors.
* The Agent MUST append non-interactive and run-once flags (e.g., `--watch=false`, `CI=true`, `--no-interaction`, `-y`) to all test and build commands to prevent hanging in watch mode or on interactive prompts.

### Operational Commands

#### Documented Build

The documented build workflow is:

```bash
SEMVER=v0.0.999; echo ${SEMVER}
BUILD_DATE=$(gdate --utc +%FT%T.%3NZ); echo ${BUILD_DATE}
GIT_COMMIT=$(git rev-parse HEAD); echo ${GIT_COMMIT}

go build -ldflags "-X planeshift/cmd.semVer=${SEMVER} -X planeshift/cmd.buildDate=${BUILD_DATE} -X planeshift/cmd.gitCommit=${GIT_COMMIT} -X planeshift/cmd.gitRef=/refs/tags/${SEMVER}" && \
./planeshift version | jq .
```

#### Tests

* No test files or repository-specific test command were detected.
* [INFERRED] Until repository-specific test instructions exist, use `go test -count=1 ./...` when tests are added.

#### Linting

* No lint command is currently documented.

#### Local Test Support

* [INFERRED] No mock servers, seed commands, or test credentials are currently required because none were found in the repository.

Implementation plan has been generated and saved to [plans/implementation-plan-regenerate-work-files.md](../../plans/implementation-plan-regenerate-work-files.md)
