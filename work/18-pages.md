# Work 18: Implement Plane API — Pages

Use this file as the implementation prompt for the Pages slice of `planeshift` (Jira ticket PSFT-2).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, and the foundation prompt before editing. This slice depends on [`01-projects.md`](./01-projects.md), [`21-assets.md`](./21-assets.md).

Implement the Pages resource completely. The scope is the 18 HTTP operations listed below plus the linked content guide. Do not implement another resource or invent behavior not present in the linked Plane documentation.

### Documentation to implement

- [Pages overview](https://developers.plane.so/api-reference/page/overview.md)
- [Page content HTML](https://developers.plane.so/api-reference/page/page-content-html.md) — content/formatting guide, not a separate HTTP operation
- [List workspace wiki pages](https://developers.plane.so/api-reference/page/list-workspace-pages.md)
- [Create a wiki page](https://developers.plane.so/api-reference/page/add-workspace-page.md)
- [Retrieve a wiki page](https://developers.plane.so/api-reference/page/get-workspace-page.md)
- [Update a workspace page](https://developers.plane.so/api-reference/page/update-workspace-page.md)
- [Archive a workspace page](https://developers.plane.so/api-reference/page/archive-workspace-page.md)
- [Restore a workspace page](https://developers.plane.so/api-reference/page/restore-workspace-page.md)
- [Delete a workspace page](https://developers.plane.so/api-reference/page/delete-workspace-page.md)
- [Retrieve workspace page attachment metadata](https://developers.plane.so/api-reference/page/get-workspace-page-attachment.md)
- [Confirm a workspace page attachment upload](https://developers.plane.so/api-reference/page/confirm-workspace-page-attachment-upload.md)
- [Download a workspace page attachment](https://developers.plane.so/api-reference/page/download-workspace-page-attachment.md)
- [Delete a workspace page attachment](https://developers.plane.so/api-reference/page/delete-workspace-page-attachment.md)
- [List project pages](https://developers.plane.so/api-reference/page/list-project-pages.md)
- [Create a project page](https://developers.plane.so/api-reference/page/add-project-page.md)
- [Retrieve a project page](https://developers.plane.so/api-reference/page/get-project-page.md)
- [Update a project page](https://developers.plane.so/api-reference/page/update-project-page.md)
- [Archive a project page](https://developers.plane.so/api-reference/page/archive-project-page.md)
- [Restore a project page](https://developers.plane.so/api-reference/page/restore-project-page.md)
- [Delete a project page](https://developers.plane.so/api-reference/page/delete-project-page.md)

Implement workspace- and project-scoped page CRUD/lifecycle plus the documented page-attachment operations. Apply the linked HTML-content guide: description_html is replaced as a complete body, content is sanitized, the documented 10 MB limit applies, and entity-backed components use UUIDs. Reconcile the guide’s workspace-asset upload flow with the attachment endpoint pages and record any API-version discrepancy rather than silently omitting an operation.

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
