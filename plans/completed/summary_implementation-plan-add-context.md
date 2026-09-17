# Summary of Actions Taken

- Added the typed `config.Context` and `config.ProjectContext` models with
  explicit JSON, YAML, and mapstructure field names.
- Centralized context resolution and persistence in `cmd/root.go`, preserving
  unrelated Viper settings and updating in-memory state only after a successful
  write. Auto-created config files continue to use mode `0600`.
- Added the `context` command with `set` and `get` subcommands. Non-interactive
  updates support changed-only `--workspace` and `--project` flags; the
  no-flag flow uses survey/v2, lists existing Plane projects, and saves the
  selected project ID and name.
- Routed `context get` through the existing RawJSON and output dispatch path,
  with JSON default and coverage for all supported output formats.
- Added safe parent, subcommand, root, and manual-testing help plus executable
  help matrix coverage.
- Added focused model, root, persistence, command, interactive, and output
  tests using isolated Viper instances and injected fakes.
- Accepted ADR-003 documenting the local non-resource context command
  exception and its package/dependency boundaries.
- Added the mandated `github.com/AlecAivazis/survey/v2` dependency and
  refreshed module sums.
- Verification passed: gofmt, `CI=true go test -count=1 ./...`, `CI=true go
  vet ./...`, documented metadata build/version output, executable help and
  smoke checks, dependency inspection, and `git diff --check`.
