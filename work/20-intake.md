# Work 20: Implement Plane API — Intake

Use this file as the implementation prompt for the Intake slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), [`01-projects.md`](./01-projects.md), and [`04-work-items.md`](./04-work-items.md) before editing.

Implement the public project-scoped Intake resource completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The existing operation list is retained and must cover exactly 5 methods. Keep intake identifiers distinct from normal project work-item identifiers.

### Documentation to implement

- [Intake overview](https://developers.plane.so/api-reference/intake-issue/overview.md)
- [Create an intake work item](https://developers.plane.so/api-reference/intake-issue/add-intake-issue.md)
- [List all intake work items](https://developers.plane.so/api-reference/intake-issue/list-intake-issues.md)
- [Retrieve an intake work item](https://developers.plane.so/api-reference/intake-issue/get-intake-issue-detail.md)
- [Update an intake work item](https://developers.plane.so/api-reference/intake-issue/update-intake-issue-detail.md)
- [Delete an intake work item](https://developers.plane.so/api-reference/intake-issue/delete-intake-issue.md)

### Public route matrix — 5 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/intake-issues/` | `{slug}` workspace slug, `{project_id}` project UUID; collection |
| 2 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/intake-issues/` | `{slug}` workspace slug, `{project_id}` project UUID; create |
| 3 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/intake-issues/{issue_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` intake issue UUID; detail |
| 4 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/intake-issues/{issue_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` intake issue UUID; update |
| 5 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/intake-issues/{issue_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` intake issue UUID; delete |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md), project paths from [`01-projects.md`](./01-projects.md), and shared work-item conventions from [`04-work-items.md`](./04-work-items.md).
- Preserve intake review/accept/reject fields as documented. Do not add normal work-item, alias, or unlisted intake routes.

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
