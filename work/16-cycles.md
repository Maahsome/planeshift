# Work 16: Implement Plane API — Cycles

Use this file as the implementation prompt for the Cycles slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), [`01-projects.md`](./01-projects.md), and [`04-work-items.md`](./04-work-items.md) before editing.

Implement the public Cycles resource completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The scope is exactly 14 methods, including the lightweight collection and cycle-issue detail additions.

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
- The `cycles-lite` collection and cycle-issue detail method are included because they are explicitly listed in [`PUBLIC_API.txt`](../PUBLIC_API.txt).

### Public route matrix — 14 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/` | `{slug}` workspace slug, `{project_id}` project UUID; cycle collection |
| 2 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/` | `{slug}` workspace slug, `{project_id}` project UUID; create |
| 3 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles-lite/` | `{slug}` workspace slug, `{project_id}` project UUID; lightweight collection |
| 4 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/{cycle_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID; detail |
| 5 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/{cycle_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID; update |
| 6 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/{cycle_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID; delete |
| 7 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/{cycle_id}/cycle-issues/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID; membership collection |
| 8 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/{cycle_id}/cycle-issues/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID; add membership |
| 9 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/{cycle_id}/cycle-issues/{issue_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID, `{issue_id}` issue UUID; membership detail |
| 10 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/{cycle_id}/cycle-issues/{issue_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID, `{issue_id}` issue UUID; remove membership |
| 11 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/{cycle_id}/transfer-issues/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID; transfer |
| 12 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/cycles/{cycle_id}/archive/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID; archive |
| 13 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/archived-cycles/` | `{slug}` workspace slug, `{project_id}` project UUID; archived collection |
| 14 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/archived-cycles/{cycle_id}/unarchive/` | `{slug}` workspace slug, `{project_id}` project UUID, `{cycle_id}` cycle UUID; restore |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md), project paths from [`01-projects.md`](./01-projects.md), and work-item identity from [`04-work-items.md`](./04-work-items.md).
- Preserve cycle dates, state/count fields, and membership identifiers. Do not add routes outside this matrix, generated SDK code, or third-party dependencies.

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing the route matrix.
- Add typed request/response models with presence-aware PATCH fields and lossless handling for nullable or dynamic JSON values.
- Expose each method through the established Cobra/config/output conventions. Collection methods must expose documented pagination/filter controls.
- Keep API keys and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 14 methods, including method/path/query/auth/body/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in the matrix has one client method, CLI surface, typed contract, and deterministic test.
- Pagination, nullable fields, dynamic JSON, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
