# Work 05: Implement Plane API — Work Item States

Use this file as the implementation prompt for the Work Item States slice of `planeshift` (Jira ticket PSFT-14).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), and [`01-projects.md`](./01-projects.md) before editing.

Implement the public project-scoped Work Item States resource completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The existing operation list is retained and must cover exactly 5 methods.

### Documentation to implement

- [Work Item States overview](https://developers.plane.so/api-reference/state/overview.md)
- [Create a state](https://developers.plane.so/api-reference/state/add-state.md)
- [List all states](https://developers.plane.so/api-reference/state/list-states.md)
- [Retrieve a state](https://developers.plane.so/api-reference/state/get-state-detail.md)
- [Update a state](https://developers.plane.so/api-reference/state/update-state-detail.md)
- [Delete a state](https://developers.plane.so/api-reference/state/delete-state.md)

### Public route matrix — 5 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/states/` | `{slug}` workspace slug, `{project_id}` project UUID; collection |
| 2 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/states/` | `{slug}` workspace slug, `{project_id}` project UUID; create |
| 3 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/states/{state_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{state_id}` state UUID; detail |
| 4 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/states/{state_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{state_id}` state UUID; update |
| 5 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/states/{state_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{state_id}` state UUID; delete |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md) and the project path established by [`01-projects.md`](./01-projects.md).
- This prompt owns only the five state routes above; do not add state-type, workflow, or other routes.

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing the route matrix.
- Add typed request/response models with presence-aware PATCH fields and lossless handling for nullable or dynamic JSON values.
- Expose each method through the established Cobra/config/output conventions and expose documented pagination/filter controls for the collection.
- Keep API keys and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 5 methods, including method/path/query/auth/body/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in the matrix has one client method, CLI surface, typed contract, and deterministic test.
- Pagination, nullable fields, dynamic JSON, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
