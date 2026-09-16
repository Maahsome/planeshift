# Work 11: Implement Plane API — Work Item Links

Use this file as the implementation prompt for the Work Item Links slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), and [`04-work-items.md`](./04-work-items.md) before editing.

Implement current and explicitly listed compatibility Work Item Links operations. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The current `/work-items/` family is primary, and the matching `/issues/` family is compatibility-only. The scope is exactly 10 methods.

### Documentation to implement

- [Work Item Links overview](https://developers.plane.so/api-reference/link/overview.md)
- [Create a link](https://developers.plane.so/api-reference/link/add-link.md)
- [List all links](https://developers.plane.so/api-reference/link/list-links.md)
- [Retrieve a link](https://developers.plane.so/api-reference/link/get-link-detail.md)
- [Update a link](https://developers.plane.so/api-reference/link/update-link-detail.md)
- [Delete a link](https://developers.plane.so/api-reference/link/delete-link.md)
- The deprecated `/issues/` link family below is included only because it is explicitly listed in [`PUBLIC_API.txt`](../PUBLIC_API.txt); do not infer additional aliases.

### Public route matrix — 10 methods

#### Primary `/work-items/` routes — 5 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/links/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID; collection |
| 2 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/links/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID; create |
| 3 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/links/{link_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID, `{link_id}` link UUID; detail |
| 4 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/links/{link_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID, `{link_id}` link UUID; update |
| 5 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/links/{link_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID, `{link_id}` link UUID; delete |

#### Deprecated `/issues/` compatibility routes — 5 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 6 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/links/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID; compatibility collection |
| 7 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/links/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID; compatibility create |
| 8 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/links/{link_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID, `{link_id}` link UUID; compatibility detail |
| 9 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/links/{link_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID, `{link_id}` link UUID; compatibility update |
| 10 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/links/{link_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID, `{link_id}` link UUID; compatibility delete |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md) and the work-item identity rules in [`04-work-items.md`](./04-work-items.md).
- Preserve relationship direction, link type, and source/target identifiers. Do not move relation collection methods from Work 04 into this prompt or invent link aliases.

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing either route family.
- Add typed request/response models with presence-aware PATCH fields and lossless handling for nullable or dynamic JSON values.
- Expose primary routes through the established Cobra/config/output conventions; compatibility routes remain explicitly bounded and non-preferred.
- Keep credentials and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 10 methods, including method/path/query/auth/body/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in both matrices has one client method, CLI surface, typed contract, and deterministic test.
- Primary and compatibility route behavior is tested separately; pagination, nullable fields, dynamic JSON, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
