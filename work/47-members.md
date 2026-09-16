# Work 47: Implement Plane API — Members

Use this file as the implementation prompt for the Members slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), and [`01-projects.md`](./01-projects.md) before editing.

Implement the public workspace and project member operations completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The scope is exactly 13 methods: both project member route families, their collection/detail CRUD, `project-members-lite`, and both workspace member list variants. The inventory contains no workspace-member removal path, so do not add one.

### Documentation to implement

- [Members overview](https://developers.plane.so/api-reference/members/overview.md)
- [Get all workspace members](https://developers.plane.so/api-reference/members/get-workspace-members.md)
- [List all project members](https://developers.plane.so/api-reference/members/get-project-members.md)
- [Create project member](https://developers.plane.so/api-reference/members/add-project-member.md)
- [Get project member](https://developers.plane.so/api-reference/members/get-project-member-detail.md)
- [Update project member](https://developers.plane.so/api-reference/members/update-project-member.md)
- [Delete project member](https://developers.plane.so/api-reference/members/delete-project-member.md)
- The `project-members-lite` and second project member route family are included because they are explicitly listed in [`PUBLIC_API.txt`](../PUBLIC_API.txt). The documentation-only workspace removal operation is not in scope.

### Public route matrix — 13 methods

#### `/members/` project route family — 5 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/members/` | `{slug}` workspace slug, `{project_id}` project UUID; collection |
| 2 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/members/` | `{slug}` workspace slug, `{project_id}` project UUID; create |
| 3 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/members/{member_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{member_id}` member UUID; detail |
| 4 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/members/{member_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{member_id}` member UUID; update |
| 5 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/members/{member_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{member_id}` member UUID; delete |

#### `/project-members/` project route family — 6 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 6 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/project-members/` | `{slug}` workspace slug, `{project_id}` project UUID; collection |
| 7 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/project-members/` | `{slug}` workspace slug, `{project_id}` project UUID; create |
| 8 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/project-members-lite/` | `{slug}` workspace slug, `{project_id}` project UUID; lightweight collection |
| 9 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/project-members/{member_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{member_id}` member UUID; detail |
| 10 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/project-members/{member_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{member_id}` member UUID; update |
| 11 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/project-members/{member_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{member_id}` member UUID; delete |

#### Workspace member list variants — 2 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 12 | GET | `/api/v1/workspaces/{slug}/members/` | `{slug}` workspace slug; full member collection |
| 13 | GET | `/api/v1/workspaces/{slug}/members-lite/` | `{slug}` workspace slug; lightweight member collection |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md) and project paths from [`01-projects.md`](./01-projects.md).
- Preserve role/access fields and distinguish project membership from workspace member listing. The public inventory contains no workspace-member deletion method, so keep this prompt limited to the matrix.

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing the route matrix.
- Add typed request/response models with presence-aware PATCH fields and lossless handling for nullable or dynamic JSON values.
- Expose each method through the established Cobra/config/output conventions. Collection methods must expose documented pagination/filter controls.
- Keep API keys and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 13 methods, including method/path/query/auth/body/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in the matrix has one client method, CLI surface, typed contract, and deterministic test.
- Both project route families, the lite route, workspace list variants, pagination, nullable fields, dynamic JSON, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
