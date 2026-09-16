# Work 21: Implement Plane API — Assets

Use this file as the implementation prompt for the Assets slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, and [`00-api-client-foundation.md`](./00-api-client-foundation.md) before editing.

Implement the public user- and workspace-scoped Assets operations completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The scope is exactly 9 methods, including the server upload variants. Preserve returned upload metadata and keep credentials, URLs, and form fields out of logs where the operation returns them.

### Documentation to implement

- [Assets overview](https://developers.plane.so/api-reference/assets/overview.md)
- [Create user asset upload](https://developers.plane.so/api-reference/assets/create-user-asset-upload.md)
- [Update user asset](https://developers.plane.so/api-reference/assets/update-user-asset.md)
- [Delete user asset](https://developers.plane.so/api-reference/assets/delete-user-asset.md)
- [Create workspace asset upload](https://developers.plane.so/api-reference/assets/create-workspace-asset-upload.md)
- [Get workspace asset](https://developers.plane.so/api-reference/assets/get-workspace-asset.md)
- [Update workspace asset](https://developers.plane.so/api-reference/assets/update-workspace-asset.md)
- The `/server/` user-asset methods are included because they are explicitly listed in [`PUBLIC_API.txt`](../PUBLIC_API.txt).

### Public route matrix — 9 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | POST | `/api/v1/assets/user-assets/` | no path parameters; create user asset |
| 2 | PATCH | `/api/v1/assets/user-assets/{asset_id}/` | `{asset_id}` asset UUID; update user asset |
| 3 | DELETE | `/api/v1/assets/user-assets/{asset_id}/` | `{asset_id}` asset UUID; delete user asset |
| 4 | POST | `/api/v1/assets/user-assets/server/` | no path parameters; create server user asset |
| 5 | PATCH | `/api/v1/assets/user-assets/{asset_id}/server/` | `{asset_id}` asset UUID; update server user asset |
| 6 | DELETE | `/api/v1/assets/user-assets/{asset_id}/server/` | `{asset_id}` asset UUID; delete server user asset |
| 7 | POST | `/api/v1/workspaces/{slug}/assets/` | `{slug}` workspace slug; create workspace asset |
| 8 | GET | `/api/v1/workspaces/{slug}/assets/{asset_id}/` | `{slug}` workspace slug, `{asset_id}` asset UUID; get workspace asset |
| 9 | PATCH | `/api/v1/workspaces/{slug}/assets/{asset_id}/` | `{slug}` workspace slug, `{asset_id}` asset UUID; update workspace asset |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md) for configuration, authentication, transport, output, error, and safe sensitive-data handling.
- Keep the three user-asset routes, three server user-asset routes, and three workspace-asset routes distinct. Do not add a separate upload-credential or completion route absent from the matrix.
- Work-item attachment routes are owned by [`14-work-item-attachments.md`](./14-work-item-attachments.md).

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing the route matrix.
- Add typed request/response models with presence-aware PATCH fields and lossless handling for nullable or dynamic JSON values. Preserve all returned upload metadata required by a listed method.
- Expose each method through the established Cobra/config/output conventions. Keep auth headers, presigned URLs, and upload fields out of logs and accidental default output.
- Add deterministic `httptest` coverage for all 9 methods, including method/path/query/auth/body/decoding/status behavior, sensitive-data redaction, and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in the matrix has one client method, CLI surface, typed contract, and deterministic test.
- Nullable fields, dynamic JSON, upload metadata handling, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
