# Summary of Actions Taken

- Implemented `planeshift context prompt` under `cmd/context` with exact fixed output: `<workspace> | <project name>` plus one newline.
- Kept the command read-only and limited to `config.Config.Context`; it does not use `config.OutputData`, Plane clients, Viper, survey, APIs, project IDs, or unrelated settings.
- Registered the command with `cobra.NoArgs` validation and preserved lazy Plane factory behavior.
- Added safe command, parent, and root help, including shell-prompt examples and fixed `--output` behavior.
- Added unit coverage for registration, no-args validation, exact output, omission of secrets/project ID, no factory invocation, and nil configuration.
- Added black-box help discovery/safety coverage and a manual testing checklist item.
- Verified with gofmt, targeted tests, `go vet`, the documented metadata build/version workflow, isolated-config smoke checks, functional help tests, and `git diff --check`.
- The repository-wide test command was also run; three pre-existing functional lifecycle tests require the unset `PLANESHIFT_BINARY` prerequisite. Elevated execution confirmed the remaining package tests pass.
