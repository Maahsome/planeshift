# Manual Testing of planeshift CLI

Automated CLI lifecycle coverage is available through the opt-in
[`mise function-test`](functional/README.md) workflow. This does not replace
the manual checklist below or its Business-license marker.

## Project Command

- [x] archive
- [x] create
- [-] create-template (This is a Business License Feature)
- [ ] delete
- [x] get
- [x] list
- [x] unarchive
- [x] update

## Work-Item Command

- [ ] create
- [ ] delete
- [ ] get
- [ ] get-by-identifier
- [ ] list
- [ ] relations-create
- [ ] relations-list
- [ ] search
- [ ] update

## Context Command

- [ ] `context set --workspace my-workspace`
- [ ] `context set --project project-uuid`
- [ ] `context set` prompts for workspace, then project name
- [ ] `context get` displays only the saved context
- [ ] `context prompt` with a configured workspace/project prints exactly `workspace | project-name` on one line and displays the project name rather than the project ID

## Resource Context Resolution

- [ ] With a configured context, run representative `project list`, `state list`, `work-item list`, `work-item relations-list`, and hidden `work-item legacy list` commands without route flags; confirm each uses the saved workspace/project.
- [ ] Repeat representative project, state, work-item, relation, and hidden legacy commands with different `--workspace` and `--project-id` values; confirm the request uses the overrides.
- [ ] Run `context get` after an override and confirm the saved workspace, project ID, and project name are unchanged.
- [ ] Confirm project commands no longer accept positional workspace/project route values; state get/update/delete retain only `state_id`, and work-item detail/update/delete/relation commands retain only `work_item_id`.
- [ ] Confirm `work-item search --project-id` remains a query filter and `get-by-identifier PROJECT_IDENTIFIER ISSUE_IDENTIFIER` retains both human identifier arguments while taking workspace from context or `--workspace`.
- [ ] Run a representative command with `--workspace ""` or `--project-id ""`; confirm it fails before a request. Remove the corresponding saved context value and confirm the required-context error occurs before client creation.
