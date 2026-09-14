# Codex Plan

# Work Ticket

ID: PSFT-9-20260914-1027-2DAECF
Title: project features

## Description

### Work 02: Implement Plane API — Project Features

Use this file as the implementation prompt for the Project Features slice of `planeshift` (Jira ticket PSFT-9).

#### Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, and the foundation prompt before editing. This slice depends on [`01-projects.md`](./01-projects.md).

Implement the Project Features resource completely. The scope is the 2 HTTP operations listed below. Do not implement another resource or invent behavior not present in the linked Plane documentation.

##### Documentation to implement

- [Project Features overview](https://developers.plane.so/api-reference/project-features/overview.md)
- [Get project features](https://developers.plane.so/api-reference/project-features/get-project-features.md)
- [Update project features](https://developers.plane.so/api-reference/project-features/update-project-features.md)

Treat feature flags as a partial-update surface: preserve fields the caller did not set and model the documented boolean/value types exactly.

##### Required implementation behavior

- Use the shared client, configuration, authentication, pagination, error, and output conventions established by `work/00-api-client-foundation.md`; do not create a second transport or configuration path.
- Read every linked operation page before coding and implement its exact HTTP method, path (including trailing slash), path parameters, query parameters, request body, OAuth scope, success status, response shape, and documented error behavior. The links are the source of truth when names and URL directory slugs differ.
- Add typed request/response models for this slice. Use pointers or equivalent presence-aware fields for nullable and PATCH fields, and preserve arbitrary JSON with a lossless representation instead of dropping unknown data.
- Expose each listed operation through the CLI’s established command hierarchy with predictable flags/arguments and the existing output formats. List operations must make cursor/per-page controls and documented `fields`/`expand` options available where supported.
- Keep API keys, bearer tokens, invitation data, presigned URLs, and upload form fields out of logs and accidental default output. Return useful structured errors, including non-JSON and 204 responses.
- Add deterministic `httptest` coverage for every operation (method, path, query, auth header, body, response decoding, status handling, and representative error cases). Do not require live Plane credentials for unit tests.
- Do not add third-party dependencies, generated SDK code, speculative endpoints, or deprecated `/issues/` aliases. Preserve existing commands and tests.

##### Definition of done

- Every operation listed in this prompt has a client method, CLI surface, typed contract, and test.
- Pagination, nullable fields, dynamic JSON, and 204 responses are verified.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes are limited to this resource slice and its focused shared test/model support.

Implementation plan has been generated and saved to [plans/implementation-plan-project-features.md](../../plans/implementation-plan-project-features.md)
