# Architecture Decision Log: planeshift

## ADR Index

| ID | Title | Status |
|---|---|---|
| ADR-001 | Adoption of Agentic SDLC | Accepted |

## ADR-001: Adoption of Agentic SDLC

**Status**: Accepted

**Context**

`planeshift` is a single-binary Go 1.25 CLI organized into `cmd`, `cmd/get`, `common`, `config`, `help`, and `objects` packages. Cobra owns command registration and execution, Viper handles local configuration and environment values, Logrus provides runtime logging, and `objects.Version` provides JSON, YAML, GRON, text, table, and raw output.

The repository has no HTTP server, authentication layer, event-processing system, persistence layer, migration system, test suite, or distributed-tracing implementation. Important existing boundaries include the Cobra hierarchy rooted at `cmd.RootCmd`, centralized configuration access in `cmd/root.go`, and restricted permissions for auto-created configuration files.

**Decision**

Adopt the Agentic SDLC governance model through `GEMINI.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, and `AGENTS.md`.

Agents must use `CONSTITUTION.md` as the authoritative source for invariants and coding conventions, and `ARCHITECTURE.md` as the authoritative decision log before proposing architectural changes. New work must preserve the existing Cobra command funnel, configuration centralization, approved dependency set, and local configuration-file protection.

**Consequences**

Agents have an explicit routing and decision process for this CLI's package boundaries and operational workflow.

The repository will avoid speculative HTTP, persistence, event-driven, authentication, or tracing abstractions that are not requested. Changes to command behavior, output formats, configuration handling, logging, or version metadata must be evaluated against the Constitution.

The documented local build workflow remains the operational reference: [`LOCAL_BUILD.md`](LOCAL_BUILD.md).

## Legacy Documentation

- [`LOCAL_BUILD.md`](LOCAL_BUILD.md) is the only relevant existing operational documentation detected.
- No README, ADR collection, or legacy architecture document was detected.
