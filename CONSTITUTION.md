# Constitution: planeshift

## Preamble: System Archetype

This is a single-binary Go 1.25 CLI for local/operator-facing plan-resource and version inspection. Predictable Cobra command behavior, correct version/build metadata, stable output formats, centralized configuration, and safe local configuration handling are primary quality attributes.

## Owned Domains

* The repository owns CLI command registration and execution.
* The repository owns local configuration-file and environment resolution.
* The repository owns version and build metadata.
* The repository owns JSON, YAML, GRON, text, table, and raw output formatting.
* The repository owns runtime logging and command help text.

## Forbidden Dependencies & Protected Zones

* New third-party dependencies MUST NOT be added without explicit team approval.
* The approved third-party dependency set is the dependency set declared in `go.mod`.
* No generated-code, `DO NOT EDIT`, OpenAPI, Swagger, protobuf, or other protected source zone currently exists.
* Existing package boundaries MUST NOT be bypassed by introducing unapproved dependency direction.

## Security & Resilience Hardlines

* The current CLI has no authentication or authorization boundary.
* Auto-created configuration files MUST be created through `createRestrictedConfigFile`.
* Auto-created configuration files MUST have permissions set to `0600`.

## Architectural & Async Invariants

* All CLI commands MUST be registered beneath the Cobra hierarchy rooted at `cmd.RootCmd`.
* The executable entry point MUST invoke `cmd.Execute()`.
* Environment and Viper configuration access MUST remain centralized in `cmd/root.go`.
* Package dependencies MUST preserve the observed direction from the executable and command packages toward supporting packages.
* No event-driven or asynchronous processing model is currently defined.

## Observability & Error Bounding

* Runtime logging MUST use Logrus through the existing standard logger setup.
* Version-command JSON unmarshalling errors MUST be wrapped with `pkg/errors` before being returned.
* Explicit startup and configuration failures MUST continue to use the existing fail-fast behavior through `cobra.CheckErr`, `logrus.Fatal`, or `os.Exit`.

## Coding Conventions

* `cmd` and `cmd/get` contain Cobra command orchestration.
* `config` contains shared configuration, output dispatch, and semantic-version parsing.
* `objects` contains version data and serialization implementations.
* `common` contains logger initialization.
* `help` contains command descriptions and examples.
* Output-producing objects implement the `config.Outputtable` interface.
* Output format names are normalized to lowercase before dispatch.
* `get` commands require at least one argument through Cobra argument validation.
* The root `version` command defaults to JSON output.
* The `get version` command defaults to text output.
* Semantic versions are parsed through `config.ParseSemver`.
* Serialization helpers log conversion failures and return empty output where the existing implementation does so.

## Governance

* Constitution supersedes all other project documentation.
* Amendments require documentation and team approval.
* All PRs/reviews MUST verify compliance with these invariants.

**Version**: 1.0.0 | **Ratified**: 2026-09-12
