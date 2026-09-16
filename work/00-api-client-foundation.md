# Work 00: Implement the shared Plane API client foundation

Use this file as the shared implementation prompt for the public Plane API work in `planeshift` (Jira ticket PSFT-11).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, and `LOCAL_BUILD.md` before editing. This foundation is shared by the 17 retained resource prompts listed in [`README.md`](./README.md) and by the final verification prompt.

The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is the route source of truth for this work. It contains 13 public route groups and 143 application methods. Current `/work-items/` paths are primary; the inventory's explicitly listed `/issues/` paths are bounded compatibility aliases owned by the work-item, link, activity, comment, and attachment prompts. Do not infer any route from the broader documentation catalog.

### Governing architectural decision protocol

Apply the **Owned Domains**, **Forbidden Dependencies & Protected Zones**, **Security & Resilience Hardlines**, **Architectural & Async Invariants**, and **Coding Conventions** sections of `CONSTITUTION.md`. ADR-001 and ADR-002 in `ARCHITECTURE.md` preserve the shared client/configuration foundation, the existing package dependency direction, and resource-oriented Cobra boundaries. If a requested foundation change would require a new package boundary, authentication model, transport, or async model, stop and record it as an uncharted area requiring team review.

### Plane contract

- Plane Cloud uses `https://api.plane.so/api/v1/`; self-hosted installations use the configured host with `/api/v1/`. Avoid double-prefixing when users provide a base URL.
- Support exactly one configured credential mode per request: `X-API-Key` or `Authorization: Bearer ...`. Never print credentials, invitation secrets, presigned URLs, or sensitive upload metadata.
- Use JSON for normal request/response bodies, standard REST verbs, context cancellation, bounded timeouts, and the documented 200/201/204 handling. Never attempt to decode a 204 body.
- Implement cursor pagination with `per_page` (server maximum 100), `cursor`, all documented response metadata, and optional `fields`/`expand` query values without losing unknown query parameters needed by a retained resource slice.
- Preserve response headers relevant to rate limiting (`X-RateLimit-Remaining` and `X-RateLimit-Reset`) in a safe structured form. Do not add unbounded automatic retries; if retry behavior is proposed, record it as a separate reviewed decision.
- Use current `/work-items/` paths where the public inventory lists them. The explicitly listed `/issues/` aliases are compatibility coverage only and must never be silently expanded or preferred.
- Support streaming/multipart or presigned-storage requests only when a retained public operation requires them, and honor returned method, URL, headers, and form fields exactly.

### Implementation scope

1. Define minimal client/configuration interfaces according to the repository's existing package conventions. Centralize Viper/environment resolution in `cmd/root.go`; make base URL, credential mode, timeout, and output settings testable.
2. Implement a standard-library HTTP transport using `net/http` and `encoding/json` unless a reviewed architectural decision requires otherwise. Build URLs safely, encode query values correctly, set `Accept`/content headers appropriately, and attach only the selected auth header.
3. Add typed shared representations for API errors, pagination metadata, rate-limit metadata, and raw/dynamic JSON fields. Make errors retain status code and a bounded safe response message while avoiding secret leakage.
4. Define the resource-client and CLI registration seam reused by the 17 resource prompts. Resource slices must not duplicate request construction, auth, pagination, output conversion, or error handling.
5. Add test helpers around `httptest.Server` for asserting method/path/query/header/body and returning JSON, empty, malformed, and non-JSON responses. Include tests for both credential modes, invalid configuration, URL joining, cancellation/timeout, pagination, rate-limit headers, and 204 handling.
6. Add only the minimal command/help/config wiring required to exercise the foundation and make the retained resource prompts straightforward. Do not add speculative endpoints or compatibility aliases absent from `PUBLIC_API.txt`.

### Definition of done

- The foundation contract is documented in code comments or repository documentation, including any uncharted-area decision note.
- No new third-party dependency or generated SDK is introduced.
- Credential handling, restricted config-file permissions, command registration, output dispatch, error/status handling, pagination, listed uploads, and safe logging are covered by deterministic tests.
- Existing version commands retain their behavior.
- `gofmt` is clean and `CI=true go test -count=1 ./...` passes.
- The documented build workflow in `LOCAL_BUILD.md` remains valid.
