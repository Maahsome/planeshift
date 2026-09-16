# Work 14: Implement Plane API — Work Item Attachments

Use this file as the implementation prompt for the Work Item Attachments slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), and [`04-work-items.md`](./04-work-items.md) before editing.

Implement current and explicitly listed compatibility Work Item Attachment operations. The current `/work-items/` family is primary, and the matching deprecated `/issues/` family is compatibility-only. The scope is exactly 10 methods. The public inventory does not list upload-credential, storage-upload, or upload-completion methods, so do not recreate that non-inventory flow.

### Documentation to implement

- [Work Item Attachments overview](https://developers.plane.so/api-reference/issue-attachments/overview.md)
- [List all attachments](https://developers.plane.so/api-reference/issue-attachments/get-attachments.md)
- [Retrieve an attachment](https://developers.plane.so/api-reference/issue-attachments/get-attachment-detail.md)
- [Update an attachment](https://developers.plane.so/api-reference/issue-attachments/update-attachment.md)
- [Delete an attachment](https://developers.plane.so/api-reference/issue-attachments/delete-attachment.md)
- The deprecated `/issues/` attachment family below is included only because it is explicitly listed in [`PUBLIC_API.txt`](../PUBLIC_API.txt).

### Public route matrix — 10 methods

#### Primary `/work-items/` routes — 5 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/attachments/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID; collection |
| 2 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/attachments/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID; create |
| 3 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/attachments/{attachment_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID, `{attachment_id}` attachment UUID; detail |
| 4 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/attachments/{attachment_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID, `{attachment_id}` attachment UUID; update |
| 5 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/work-items/{work_item_id}/attachments/{attachment_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{work_item_id}` work-item UUID, `{attachment_id}` attachment UUID; delete |

#### Deprecated `/issues/` compatibility routes — 5 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 6 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/issue-attachments/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID; compatibility collection |
| 7 | POST | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/issue-attachments/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID; compatibility create |
| 8 | GET | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/issue-attachments/{attachment_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID, `{attachment_id}` attachment UUID; compatibility detail |
| 9 | PATCH | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/issue-attachments/{attachment_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID, `{attachment_id}` attachment UUID; compatibility update |
| 10 | DELETE | `/api/v1/workspaces/{slug}/projects/{project_id}/issues/{issue_id}/issue-attachments/{attachment_id}/` | `{slug}` workspace slug, `{project_id}` project UUID, `{issue_id}` issue UUID, `{attachment_id}` attachment UUID; compatibility delete |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md) and the work-item identity rules in [`04-work-items.md`](./04-work-items.md).
- Keep attachment metadata and any returned URLs or opaque fields lossless and redact credentials or sensitive values from logs. Do not invent a separate presigned upload protocol for routes absent from the inventory.
- The matching asset upload routes are owned by [`21-assets.md`](./21-assets.md), not this prompt.

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing either route family.
- Add typed request/response models with presence-aware PATCH fields and lossless handling for nullable or dynamic JSON values.
- Expose primary routes through the established Cobra/config/output conventions; compatibility routes remain explicitly bounded and non-preferred.
- Keep API keys and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 10 methods, including method/path/query/auth/body/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in both matrices has one client method, CLI surface, typed contract, and deterministic test.
- Primary and compatibility route behavior is tested separately; pagination, nullable fields, dynamic JSON, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
