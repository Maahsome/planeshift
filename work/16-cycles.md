# Work 16: Implement Plane API — Cycles

Use this file as the implementation prompt for the Cycles slice of `planeshift` (Jira ticket PSFT-2).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, and the foundation prompt before editing. This slice depends on [`01-projects.md`](./01-projects.md), [`04-work-items.md`](./04-work-items.md).

Implement the Cycles resource completely. The scope is the 12 HTTP operations listed below. Do not implement another resource or invent behavior not present in the linked Plane documentation.

### Documentation to implement

- [Cycles overview](https://developers.plane.so/api-reference/cycle/overview.md)
- [Create a cycle](https://developers.plane.so/api-reference/cycle/add-cycle.md)
- [Add work items to cycle](https://developers.plane.so/api-reference/cycle/add-cycle-work-items.md)
- [Transfer cycle work items](https://developers.plane.so/api-reference/cycle/transfer-cycle-work-items.md)
- [Archive a cycle](https://developers.plane.so/api-reference/cycle/archive-cycle.md)
- [List all cycles](https://developers.plane.so/api-reference/cycle/list-cycles.md)
- [Retrieve a cycle](https://developers.plane.so/api-reference/cycle/get-cycle-detail.md)
- [List all work items in a cycle](https://developers.plane.so/api-reference/cycle/list-cycle-work-items.md)
- [List all archived cycles](https://developers.plane.so/api-reference/cycle/list-archived-cycles.md)
- [Update a cycle](https://developers.plane.so/api-reference/cycle/update-cycle-detail.md)
- [Restore a cycle](https://developers.plane.so/api-reference/cycle/unarchive-cycle.md)
- [Remove work item from cycle](https://developers.plane.so/api-reference/cycle/remove-cycle-work-item.md)
- [Delete a cycle](https://developers.plane.so/api-reference/cycle/delete-cycle.md)

Implement the full cycle lifecycle and membership operations, including current versus archived listing, adding/removing work items, and transferring items between cycles. Preserve dates and status/count fields.

### Required implementation behavior

- Use the shared client, configuration, authentication, pagination, error, and output conventions established by `work/00-api-client-foundation.md`; do not create a second transport or configuration path.
- Read every linked operation page before coding and implement its exact HTTP method, path (including trailing slash), path parameters, query parameters, request body, OAuth scope, success status, response shape, and documented error behavior. The links are the source of truth when names and URL directory slugs differ.
- Add typed request/response models for this slice. Use pointers or equivalent presence-aware fields for nullable and PATCH fields, and preserve arbitrary JSON with a lossless representation instead of dropping unknown data.
- Expose each listed operation through the CLI’s established command hierarchy with predictable flags/arguments and the existing output formats. List operations must make cursor/per-page controls and documented `fields`/`expand` options available where supported.
- Keep API keys, bearer tokens, invitation data, presigned URLs, and upload form fields out of logs and accidental default output. Return useful structured errors, including non-JSON and 204 responses.
- Add deterministic `httptest` coverage for every operation (method, path, query, auth header, body, response decoding, status handling, and representative error cases). Do not require live Plane credentials for unit tests.
- Do not add third-party dependencies, generated SDK code, speculative endpoints, or deprecated `/issues/` aliases. Preserve existing commands and tests.

### Definition of done

- Every operation listed in this prompt has a client method, CLI surface, typed contract, and test.
- Pagination, nullable fields, dynamic JSON, and 204 responses are verified.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes are limited to this resource slice and its focused shared test/model support.
