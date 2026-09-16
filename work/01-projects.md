# Work 01: Implement Plane API — Projects

Use this file as the implementation prompt for the Projects slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, and [`00-api-client-foundation.md`](./00-api-client-foundation.md) before editing. This slice is the foundation for project-scoped resources such as work items, states, labels, cycles, modules, intake, estimates, and members.

Implement the public Projects resource completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The scope is exactly 9 HTTP methods; the paths below preserve the inventory's method, parameter spelling, and trailing slash.

### Documentation to implement

- [Projects overview](https://developers.plane.so/api-reference/project/overview.md)
- [Create a project](https://developers.plane.so/api-reference/project/add-project.md)
- [List all projects](https://developers.plane.so/api-reference/project/list-projects.md)
- [Retrieve a project](https://developers.plane.so/api-reference/project/get-project-detail.md)
- [Update a project](https://developers.plane.so/api-reference/project/update-project-detail.md)
- [Archive project](https://developers.plane.so/api-reference/project/archive-project.md)
- [Unarchive project](https://developers.plane.so/api-reference/project/unarchive-project.md)
- [Delete a project](https://developers.plane.so/api-reference/project/delete-project.md)
- The `projects-lite` and `summary` additions are included because they are explicitly listed in [`PUBLIC_API.txt`](../PUBLIC_API.txt); implement only the methods in the matrix.

### Public route matrix — 9 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/` | `{slug}` workspace slug; project collection |
| 2 | POST | `/api/v1/workspaces/{slug}/projects/` | `{slug}` workspace slug; create project |
| 3 | GET | `/api/v1/workspaces/{slug}/projects-lite/` | `{slug}` workspace slug; lightweight collection |
| 4 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/` | `{slug}` workspace slug, `{project_id}` project UUID; detail |
| 5 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/` | `{slug}` workspace slug, `{project_id}` project UUID; update |
| 6 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/` | `{slug}` workspace slug, `{project_id}` project UUID; delete |
| 7 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/archive/` | `{slug}` workspace slug, `{project_id}` project UUID; archive |
| 8 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/archive/` | `{slug}` workspace slug, `{project_id}` project UUID; unarchive |
| 9 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/summary/` | `{slug}` workspace slug, `{project_id}` project UUID; summary |

### Dependencies and boundaries

- Reuse the shared client, configuration, authentication, pagination, error, and output contract from [`00-api-client-foundation.md`](./00-api-client-foundation.md).
- Keep project-scoped work-item labels in [`06-work-item-labels.md`](./06-work-item-labels.md) and do not recreate the excluded workspace-level `/project-labels/` route.
- Do not add a route outside this matrix, speculative endpoint, generated SDK, or third-party dependency.

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing the route matrix.
- Add typed request/response models with presence-aware PATCH fields and lossless handling for nullable or dynamic JSON values.
- Expose each method through the established Cobra/config/output conventions. List methods must expose their documented pagination and filter controls.
- Keep API keys, bearer tokens, invitation data, and any sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 9 methods, including method/path/query/auth/body/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in the matrix has one client method, CLI surface, typed contract, and deterministic test.
- Pagination, nullable fields, dynamic JSON, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
