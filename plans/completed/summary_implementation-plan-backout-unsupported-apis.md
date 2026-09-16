# Summary of Actions Taken

- Applied the repository governance from `CONSTITUTION.md` and ADR-002 in `ARCHITECTURE.md` to keep the rollback within the existing Cobra/resource boundaries and shared configuration/transport behavior.
- Removed the Project Features CLI commands, typed client/models, tests, focused help, registration, and Project Features-only helpers while retaining existing Project operations and `optionalBoolFlag`.
- Removed the workspace-level Project Labels commands, typed client/models, tests, help, root registration, aliases, and related test setup.
- Removed the unsupported Project Features and workspace `/project-labels/` routes, six dedicated schemas, and the unused `Project Labels` tag from `spec/plane-api.yaml`.
- Preserved the workspace `/features/` contract, project-scoped `/projects/{project_id}/labels/` contract, generic shared schemas, and the intentional exclusion/history references in `PUBLIC_API.txt`, `work/`, completed plans, and planned tickets.
- Removed stale Project Features and Project Labels sections from `MANUAL_TESTING.md`.
- Added regression coverage for the removed `features`, `project-label`, and `project-labels` registrations while retaining Project operation/flag, version, lazy-factory, and surviving alias coverage.
- Updated and completed `plans/implementation-plan-backout-unsupported-apis.md` incrementally.

## Verification

- Formatted changed Go files with `gofmt`.
- Passed `env -u PLANE_API_KEY CI=true go test -count=1 ./...`.
- Passed the documented `LOCAL_BUILD.md` build/version workflow for `v0.0.999`.
- Passed YAML parsing and `git diff --check`.
- Confirmed `go.mod` and `go.sum` are unchanged and active unsupported API references are absent except for intentional negative regression assertions.
- Confirmed `planeshift project --help` retains the supported lifecycle and `planeshift version` retains JSON version output.
