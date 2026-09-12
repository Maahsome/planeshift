# Work 04: Implement Plane API — Work Items

Use this file as the implementation prompt for the Work Items slice of `planeshift` (Jira ticket PSFT-2).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, and the foundation prompt before editing. This slice depends on [`01-projects.md`](./01-projects.md), [`05-work-item-states.md`](./05-work-item-states.md), [`06-work-item-labels.md`](./06-work-item-labels.md), [`07-work-item-types.md`](./07-work-item-types.md).

Implement the Work Items resource completely. The scope is the 8 HTTP operations listed below. Do not implement another resource or invent behavior not present in the linked Plane documentation.

### Documentation to implement

- [Work Items overview](https://developers.plane.so/api-reference/issue/overview.md)
- [Create a work item](https://developers.plane.so/api-reference/issue/add-issue.md)
- [List all work items](https://developers.plane.so/api-reference/issue/list-issues.md)
- [Retrieve a work item by ID](https://developers.plane.so/api-reference/issue/get-issue-detail.md)
- [Retrieve a work item by identifier](https://developers.plane.so/api-reference/issue/get-issue-sequence-id.md)
- [Search work items](https://developers.plane.so/api-reference/issue/search-issues.md)
- [Advanced search work items](https://developers.plane.so/api-reference/issue/advanced-search-work-items.md)
- [Update a work item](https://developers.plane.so/api-reference/issue/update-issue-detail.md)
- [Delete a work item](https://developers.plane.so/api-reference/issue/delete-issue.md)

This is the core work-item surface: support UUID retrieval, human-facing identifier retrieval, normal search, advanced search, CRUD, relationships, assignees, labels, state, and custom fields as documented. Use `/work-items/` API routes; do not add deprecated `/issues/` aliases.

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
