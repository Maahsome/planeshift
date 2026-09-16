# Work 12: Implement Plane API — Work Item Activity

Use this file as the implementation prompt for the Work Item Activity slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), and [`04-work-items.md`](./04-work-items.md) before editing.

Implement current and explicitly listed compatibility Work Item Activity operations. This is read-only history: preserve verb, field, old/new values, actor, timestamps, and pagination without synthesizing mutations. The current `/work-items/` family is primary, and the matching `/issues/` family is compatibility-only. The scope is exactly 4 methods.

### Documentation to implement

- [Work Item Activity overview](https://developers.plane.so/api-reference/issue-activity/overview.md)
- [List all work item activity](https://developers.plane.so/api-reference/issue-activity/list-issue-activities.md)
- [Retrieve a work item activity](https://developers.plane.so/api-reference/issue-activity/get-issue-activity-detail.md)
- The deprecated `/issues/` activity family below is included only because it is explicitly listed in [`PUBLIC_API.txt`](../PUBLIC_API.txt).

### Public route matrix — 4 methods

#### Primary `/work-items/` routes — 2 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/activities/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID; collection |
| 2 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/activities/{activity_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID, `{activity_id}` activity UUID; detail |

#### Deprecated `/issues/` compatibility routes — 2 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 3 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/activities/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID; compatibility collection |
| 4 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/activities/{activity_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID, `{activity_id}` activity UUID; compatibility detail |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md) and the work-item identity rules in [`04-work-items.md`](./04-work-items.md).
- This prompt is read-only and owns only the four activity routes above. It must not create an activity or add unlisted aliases.

### Required implementation behavior

- Read the linked operation pages and implement their exact query parameters, OAuth scope, success status, response shape, and documented errors without changing either route family.
- Add typed response models that preserve nullable and dynamic old/new values without loss.
- Expose primary routes through the established Cobra/config/output conventions; compatibility routes remain explicitly bounded and non-preferred.
- Keep credentials and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 4 methods, including method/path/query/auth/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in both matrices has one client method, CLI surface, typed contract, and deterministic test.
- Primary and compatibility route behavior is tested separately; pagination and dynamic JSON are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
