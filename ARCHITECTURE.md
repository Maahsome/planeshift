# Architecture Decision Log: planeshift

## ADR Index

| ID | Title | Status |
|---|---|---|
| ADR-001 | Adoption of Agentic SDLC | Accepted |
| ADR-002 | Resource-Oriented Cobra Command Hierarchy | Accepted |

## ADR-001: Adoption of Agentic SDLC

**Status**: Accepted

**Context**

`planeshift` is a single-binary Go 1.25 CLI organized into `cmd`, resource
command packages such as `cmd/project`, the canonical `cmd/version` command,
`common`, `config`, `help`, and `objects` packages. Cobra owns command
registration and execution, Viper handles local configuration and environment
values, Logrus provides runtime logging, and `objects.Version` provides JSON,
YAML, GRON, text, table, and raw output.

The repository has no HTTP server, authentication layer, event-processing system, persistence layer, migration system, test suite, or distributed-tracing implementation. Important existing boundaries include the Cobra hierarchy rooted at `cmd.RootCmd`, centralized configuration access in `cmd/root.go`, and restricted permissions for auto-created configuration files.

**Decision**

Adopt the Agentic SDLC governance model through `GEMINI.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, and `AGENTS.md`.

Agents must use `CONSTITUTION.md` as the authoritative source for invariants and coding conventions, and `ARCHITECTURE.md` as the authoritative decision log before proposing architectural changes. New work must preserve the existing Cobra command funnel, configuration centralization, approved dependency set, and local configuration-file protection.

**Consequences**

Agents have an explicit routing and decision process for this CLI's package boundaries and operational workflow.

The repository will avoid speculative HTTP, persistence, event-driven, authentication, or tracing abstractions that are not requested. Changes to command behavior, output formats, configuration handling, logging, or version metadata must be evaluated against the Constitution.

The documented local build workflow remains the operational reference: [`LOCAL_BUILD.md`](LOCAL_BUILD.md).

## ADR-002: Resource-Oriented Cobra Command Hierarchy

**Status**: Accepted (PSFT-7)

**Context**

The initial project resource was registered under a generic `get` hierarchy,
and its operations were concentrated in one command file. That structure does
not establish a repeatable package boundary for additional Plane resources and
also made the generic hierarchy compete with the canonical root version
command.

**Decision**

Top-level Cobra commands represent Plane resources and use the singular name
as the primary invocation with the plural name as an alias. The Project
command is owned by `cmd/project`, with the parent command and each operation
in separate files. Command names are centralized in `config/aliases.go`.
The canonical version command remains `planeshift version` and is owned by
`cmd/version`; the obsolete generic `get` hierarchy is not retained.

The root package remains responsible for shared configuration, logging,
version metadata, and lazy `plane.ClientFactory` creation. Resource command
packages receive `*config.Config` and the factory explicitly, and route-specific
API behavior remains in the corresponding resource package.

**Consequences**

New resource commands have a predictable package/file layout and can be
registered directly below `cmd.RootCmd` without introducing a compatibility
`get` command. Existing version output remains available at the root with its
JSON default, while the Plane client and API contracts are unchanged.

## Legacy Documentation

- [`LOCAL_BUILD.md`](LOCAL_BUILD.md) is the only relevant existing operational documentation detected.
- No README, ADR collection, or legacy architecture document was detected.
