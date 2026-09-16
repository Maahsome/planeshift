# Work 17: Implement Plane API — Modules

Use this file as the implementation prompt for the Modules slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), [`01-projects.md`](./01-projects.md), and [`04-work-items.md`](./04-work-items.md) before editing.

Implement the public Modules resource completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The scope is exactly 12 methods, including the lightweight collection addition.

### Documentation to implement

- [Modules overview](https://developers.plane.so/api-reference/module/overview.md)
- [Create a module](https://developers.plane.so/api-reference/module/add-module.md)
- [List all modules](https://developers.plane.so/api-reference/module/list-modules.md)
- [Retrieve a module](https://developers.plane.so/api-reference/module/get-module-detail.md)
- [Update a module](https://developers.plane.so/api-reference/module/update-module-detail.md)
- [Delete a module](https://developers.plane.so/api-reference/module/delete-module.md)
- [Add work items to module](https://developers.plane.so/api-reference/module/add-module-work-items.md)
- [List all work items in a module](https://developers.plane.so/api-reference/module/list-module-work-items.md)
- [Archive a module](https://developers.plane.so/api-reference/module/archive-module.md)
- [List all archived modules](https://developers.plane.so/api-reference/module/list-archived-modules.md)
- [Restore a module](https://developers.plane.so/api-reference/module/unarchive-module.md)
- The `modules-lite` collection is included because it is explicitly listed in [`PUBLIC_API.txt`](../PUBLIC_API.txt).

### Public route matrix — 12 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/modules/` | `{slug}` workspace slug, `{project_id}` project UUID; module collection |
| 2 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/modules/` | `{slug}` workspace slug, `{project_id}` project UUID; create |
| 3 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/modules-lite/` | `{slug}` workspace slug, `{project_id}` project UUID; lightweight collection |
| 4 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/modules/{module_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{module_id}` module UUID; detail |
| 5 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/modules/{module_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{module_id}` module UUID; update |
| 6 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/modules/{module_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{module_id}` module UUID; delete |
| 7 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/modules/{module_id}/module-issues/` | `{slug}` workspace slug, `{project_id}` project UUID, `{module_id}` module UUID; membership collection |
| 8 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/modules/{module_id}/module-issues/` | `{slug}` workspace slug, `{project_id}` project UUID, `{module_id}` module UUID; add membership |
| 9 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/modules/{module_id}/module-issues/{issue_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{module_id}` module UUID, `{issue_id}` issue UUID; remove membership |
| 10 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/modules/{module_id}/archive/` | `{slug}` workspace slug, `{project_id}` project UUID, `{module_id}` module UUID; archive |
| 11 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/archived-modules/` | `{slug}` workspace slug, `{project_id}` project UUID; archived collection |
| 12 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/archived-modules/{module_id}/unarchive/` | `{slug}` workspace slug, `{project_id}` project UUID, `{module_id}` module UUID; restore |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md), project paths from [`01-projects.md`](./01-projects.md), and work-item identity from [`04-work-items.md`](./04-work-items.md).
- Preserve module dates, status/count fields, and membership identifiers. Do not add routes outside this matrix, generated SDK code, or third-party dependencies.

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing the route matrix.
- Add typed request/response models with presence-aware PATCH fields and lossless handling for nullable or dynamic JSON values.
- Expose each method through the established Cobra/config/output conventions. Collection methods must expose documented pagination/filter controls.
- Keep API keys and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 12 methods, including method/path/query/auth/body/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in the matrix has one client method, CLI surface, typed contract, and deterministic test.
- Pagination, nullable fields, dynamic JSON, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
