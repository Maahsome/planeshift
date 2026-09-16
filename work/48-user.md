# Work 48: Implement Plane API — User

Use this file as the implementation prompt for the User slice of `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, and [`00-api-client-foundation.md`](./00-api-client-foundation.md) before editing.

Implement the public current-user operation completely. The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is authoritative for route inclusion. The existing operation list is retained and must cover exactly 1 method.

### Documentation to implement

- [User overview](https://developers.plane.so/api-reference/user/overview.md)
- [Retrieve current user](https://developers.plane.so/api-reference/user/get-current-user.md)

### Public route matrix — 1 method

| # | Method | Exact path | Path parameters / role |
|---:|---|---|---|
| 1 | GET | `/api/v1/users/me/` | no path parameters; current authenticated user |

### Dependencies and boundaries

- Reuse [`00-api-client-foundation.md`](./00-api-client-foundation.md) for configuration, authentication, transport, error, and output behavior.
- Preserve the authenticated user response losslessly. Do not add user administration, workspace membership, or unlisted routes.

### Required implementation behavior

- Read the linked operation page and implement its exact query parameters, OAuth scope, success status, response shape, and documented errors without changing the route matrix.
- Add a typed response model that preserves nullable and dynamic JSON values.
- Expose the method through the established Cobra/config/output conventions.
- Keep credentials and sensitive response fields out of logs and accidental default output. Return bounded structured errors for JSON, empty, malformed, and non-JSON responses.
- Add deterministic `httptest` coverage for the method, including method/path/query/auth/decoding/status behavior and representative errors. Do not require live Plane credentials.

### Definition of done

- The method in the matrix has one client method, CLI surface, typed contract, and deterministic test.
- Nullable fields, dynamic JSON, and applicable empty responses are verified.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
- Changes stay within this resource slice and focused shared test/model support.
