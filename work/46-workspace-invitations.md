# Work 46: Implement Plane API — Workspace Invitations

Use this file as the implementation prompt for the Workspace Invitations slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, [`00-api-client-foundation.md`](./00-api-client-foundation.md), and [`47-members.md`](./47-members.md) before editing.

Implement the public Workspace Invitations resource completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The scope is exactly 6 methods; both `PUT` and `PATCH` on the detail route are distinct inventory methods and must remain distinct in the client and CLI contract.

### Documentation to implement

- [Workspace Invitations overview](https://developers.plane.so/api-reference/workspace-invitations/overview.md)
- [Create workspace invitation](https://developers.plane.so/api-reference/workspace-invitations/add-workspace-invitation.md)
- [List workspace invitations](https://developers.plane.so/api-reference/workspace-invitations/list-workspace-invitations.md)
- [Get workspace invitation](https://developers.plane.so/api-reference/workspace-invitations/get-workspace-invitation-detail.md)
- [Update workspace invitation](https://developers.plane.so/api-reference/workspace-invitations/update-workspace-invitation.md)
- [Delete workspace invitation](https://developers.plane.so/api-reference/workspace-invitations/delete-workspace-invitation.md)

### Public route matrix — 6 methods

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/workspaces/{slug}/invitations/` | `{slug}` workspace slug; collection |
| 2 | POST | `/api/v1/workspaces/{slug}/invitations/` | `{slug}` workspace slug; create |
| 3 | GET | `/api/v1/workspaces/{slug}/invitations/{invitation_id}/` | `{slug}` workspace slug, `{invitation_id}` invitation UUID; detail |
| 4 | PUT | `/api/v1/workspaces/{slug}/invitations/{invitation_id}/` | `{slug}` workspace slug, `{invitation_id}` invitation UUID; full update |
| 5 | PATCH | `/api/v1/workspaces/{slug}/invitations/{invitation_id}/` | `{slug}` workspace slug, `{invitation_id}` invitation UUID; partial update |
| 6 | DELETE | `/api/v1/workspaces/{slug}/invitations/{invitation_id}/` | `{slug}` workspace slug, `{invitation_id}` invitation UUID; delete |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md) and member identity/role conventions from [`47-members.md`](./47-members.md).
- Preserve email, role, status, token, and expiry behavior as documented, but redact invitation secrets. Do not collapse the `PUT` and `PATCH` methods or add a workspace-member-removal route here.

### Required implementation behavior

- Read the linked operation pages and implement their exact request body, query parameters, OAuth scope, success status, response shape, and documented errors without changing the route matrix.
- Add typed request/response models with presence-aware update fields and lossless handling for nullable or dynamic JSON values.
- Expose each method through the established Cobra/config/output conventions and expose documented pagination/filter controls for the collection.
- Keep API keys, invitation secrets, and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for all 6 methods, including method/path/query/auth/body/decoding/status behavior, secret redaction, and representative errors. Do not require live Plane credentials.

### Definition of done

- Every method in the matrix has one client method, CLI surface, typed contract, and deterministic test.
- `PUT` and `PATCH` semantics, pagination, nullable fields, dynamic JSON, and 204 responses are verified where applicable.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
