# Work 19: Implement Plane API — Collections

Use this file as the implementation prompt for the Collections slice of `planeshift` (Jira ticket PSFT-2).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, and the foundation prompt before editing. This slice depends on [`18-pages.md`](./18-pages.md), [`47-members.md`](./47-members.md).

Implement the Collections resource completely. The scope is the 14 HTTP operations listed below. Do not implement another resource or invent behavior not present in the linked Plane documentation.

### Documentation to implement

- [Collections overview](https://developers.plane.so/api-reference/collection/overview.md)
- [List collections](https://developers.plane.so/api-reference/collection/list-collections.md)
- [Create a collection](https://developers.plane.so/api-reference/collection/create-collection.md)
- [Retrieve a collection](https://developers.plane.so/api-reference/collection/retrieve-collection.md)
- [Update a collection](https://developers.plane.so/api-reference/collection/update-collection.md)
- [Delete a collection](https://developers.plane.so/api-reference/collection/delete-collection.md)
- [List collection members](https://developers.plane.so/api-reference/collection/list-collection-members.md)
- [Add a collection member](https://developers.plane.so/api-reference/collection/add-collection-member.md)
- [Update a collection member](https://developers.plane.so/api-reference/collection/update-collection-member.md)
- [Remove a collection member](https://developers.plane.so/api-reference/collection/remove-collection-member.md)
- [List collection pages](https://developers.plane.so/api-reference/collection/list-collection-pages.md)
- [Search addable collection pages](https://developers.plane.so/api-reference/collection/search-collection-pages.md)
- [Add pages to a collection](https://developers.plane.so/api-reference/collection/add-collection-pages.md)
- [Move or reorder a collection page](https://developers.plane.so/api-reference/collection/move-or-reorder-collection-page.md)
- [Remove a page from a collection](https://developers.plane.so/api-reference/collection/remove-collection-page.md)

Implement collection CRUD, public/private visibility, explicit member access, page-tree listing/search, add/remove, and move/reorder semantics. Preserve ordering and branch/tree identifiers exactly.

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
