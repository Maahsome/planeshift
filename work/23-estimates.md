# Work 23: Implement Plane API — Estimates

Use this file as the implementation prompt for the Estimates slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), and [`01-projects.md`](./01-projects.md) before editing.

Implement the public Estimates and Estimate Points operations completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The existing operation list is retained and must cover exactly 8 methods.

### Documentation to implement

- [Estimates overview](https://developers.plane.so/api-reference/estimate/overview.md)
- [Create an estimate](https://developers.plane.so/api-reference/estimate/add-estimate.md)
- [Get an estimate](https://developers.plane.so/api-reference/estimate/get-estimate.md)
- [Update an estimate](https://developers.plane.so/api-reference/estimate/update-estimate.md)
- [Delete an estimate](https://developers.plane.so/api-reference/estimate/delete-estimate.md)
- [List estimate points](https://developers.plane.so/api-reference/estimate/list-estimate-points.md)
- [Create estimate points](https://developers.plane.so/api-reference/estimate/add-estimate-points.md)
- [Update an estimate point](https://developers.plane.so/api-reference/estimate/update-estimate-point.md)
- [Delete an estimate point](https://developers.plane.so/api-reference/estimate/delete-estimate-point.md)

### Public route matrix — 8 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/estimates/` | `{slug}` workspace slug, `{project_id}` project UUID; estimate collection |
| 2 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/estimates/` | `{slug}` workspace slug, `{project_id}` project UUID; create estimate |
| 3 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/estimates/` | `{slug}` workspace slug, `{project_id}` project UUID; update estimate collection resource |
| 4 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/estimates/` | `{slug}` workspace slug, `{project_id}` project UUID; delete estimate collection resource |
| 5 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/estimates/{estimate_id}/estimate-points/` | `{slug}` workspace slug, `{project_id}` project UUID, `{estimate_id}` estimate UUID; points collection |
| 6 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/estimates/{estimate_id}/estimate-points/` | `{slug}` workspace slug, `{project_id}` project UUID, `{estimate_id}` estimate UUID; create point |
| 7 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/estimates/{estimate_id}/estimate-points/{estimate_point_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{estimate_id}` estimate UUID, `{estimate_point_id}` point UUID; update point |
| 8 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/estimates/{estimate_id}/estimate-points/{estimate_point_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{estimate_id}` estimate UUID, `{estimate_point_id}` point UUID; delete point |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md) and the project path established by [`01-projects.md`](./01-projects.md).
- Preserve estimate and estimate-point identifiers and collection-level methods exactly as listed. Do not add routes outside this matrix, generated SDK code, or third-party dependencies.

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing the route matrix.
- Add typed request/response models with presence-aware PATCH fields and lossless handling for nullable or dynamic JSON values.
- Expose each method through the established Cobra/config/output conventions and expose documented pagination/filter controls for collection methods.
- Keep API keys and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 8 methods, including method/path/query/auth/body/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in the matrix has one client method, CLI surface, typed contract, and deterministic test.
- Pagination, nullable fields, dynamic JSON, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
