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
