# Shared Plane API client foundation

This document defines the reusable Plane transport contract for the future
resource slices. It intentionally contains no resource routes or resource
models. The route surface remains the responsibility of each later resource
package.

The cross-cutting contract is the checked-in [`spec/plane-api.yaml`](spec/plane-api.yaml),
the PSFT-5 implementation ticket, and the official
[Plane API introduction](https://developers.plane.so/api-reference/introduction).
The introduction describes REST/JSON requests, `200`/`201`/`204` responses,
cursor metadata, `fields`/`expand`, rate-limit headers, and the deprecation of
legacy `/issues/` routes in favor of `/work-items/`.

## Governance and boundary status

The implementation follows these Constitution sections:

- **Owned Domains** keeps command registration/execution in `cmd`, Viper and
  environment resolution in `cmd/root.go`, output dispatch in `config`,
  serialization in `objects`, and logging in `common`.
- **Forbidden Dependencies & Protected Zones** requires the transport to use
  the Go standard library and forbids generated/OpenAPI client code or new
  third-party dependencies.
- **Security & Resilience Hardlines** keeps credentials out of messages and
  output and preserves restricted auto-created configuration files at `0600`.
- **Architectural & Async Invariants** keeps commands below `cmd.RootCmd`,
  preserves the existing dependency direction, keeps configuration access in
  `cmd/root.go`, and uses one synchronous request with no retry worker.
- **Coding Conventions** preserves the `cmd`, `config`, `objects`, `common`,
  `help`, and `config.Outputtable` responsibilities.

The client/CLI package boundary is an uncharted decision and is **pending team
review**. The proposed minimal decision is one replaceable `planeshift/plane`
support package with an injectable synchronous `net/http` transport; `cmd/root.go`
resolves settings and supplies a lazy `plane.ClientFactory`; future resource
packages depend on the client interface rather than on `cmd` or Viper.

The consequence is that route-specific paths and models stay in later resource
slices, while authentication, request construction, response decoding,
pagination, upload handling, and safe errors remain shared. No new accepted ADR
has been added to `ARCHITECTURE.md`; if the team accepts this boundary, record it
as a reviewed ADR there before widening the integration.

## Configuration

The configuration model is `config.PlaneSettings`, embedded in
`config.Config`. The supported config-file keys are:

```yaml
plane:
  api_url: https://api.plane.so
  auth_mode: api-key
  api_key: plane_api_example
  bearer_token: ""
  timeout: 30s
```

The same settings can be supplied by these explicit environment variables:

| Setting | Environment variable | Default |
|---|---|---|
| API host/base URL | `PLANE_API_URL` | `https://api.plane.so` |
| credential mode | `PLANE_AUTH_MODE` | unset; required for resource calls |
| API key | `PLANE_API_KEY` | unset |
| bearer token | `PLANE_BEARER_TOKEN` | unset |
| request timeout | `PLANE_TIMEOUT` | `30s` |

The existing config-file discovery runs first. Explicit `PLANE_*` bindings are
then applied by `cmd/root.go`, so environment values take precedence over file
values. Timeout values use Go duration syntax and must be greater than zero and
no greater than ten minutes.

Exactly one mode must be selected for a resource request:

- `api-key` requires `PLANE_API_KEY` or `plane.api_key` and sends only
  `X-API-Key`.
- `bearer` requires `PLANE_BEARER_TOKEN` or `plane.bearer_token` and sends only
  `Authorization: Bearer ...`.

An invalid/missing mode, missing selected credential, ambiguous credentials, or
invalid timeout fails before a request is sent. The root command creates the
factory lazily, so `version` and `get version` do not require Plane settings or
construct a client. Secrets are not included in `String` output, JSON tags,
help text, logs, request errors, or response envelopes.

## URL and request contract

`plane.NormalizeBaseURL` accepts a host or base URL and produces exactly one
`/api/v1/` suffix:

- empty input becomes `https://api.plane.so/api/v1/`;
- `https://plane.example/` becomes
  `https://plane.example/api/v1/`;
- `https://plane.example/api/v1` remains
  `https://plane.example/api/v1/`.

Base URLs with credentials, fragments, queries, unsupported schemes, or invalid
syntax are rejected. Resource code supplies an explicit route path to
`plane.Client.Do`; it may use a route with or without the `/api/v1` prefix, and
the client preserves the route's trailing slash. Route queries and fragments
are rejected so query construction cannot be silently changed. Dynamic path
segments should be escaped with `plane.EscapePathSegment` before they are
joined, and query values use `url.Values`.

The minimal interface is:

```go
Do(ctx context.Context, method, route string, query url.Values,
   body any, headers http.Header, destination any) (plane.Response, error)
```

Non-nil normal bodies are JSON encoded. Requests advertise
`Accept: application/json`; `Content-Type: application/json` is added when a
JSON body is present. Caller headers may add safe request metadata, but auth,
cookie, and proxy-authentication headers are client-managed/rejected. The
factory accepts an `http.Client` or `RoundTripper` for deterministic tests and
performs one synchronous request only. It adds no automatic retries.

## Responses, errors, and pagination

`200` and `201` JSON bodies are decoded into the caller's destination. Empty
successful bodies are valid. `204 No Content` is successful and is never passed
to a JSON decoder. Every response body is closed.

`plane.Response` retains the status code and only safe metadata:

- `X-RateLimit-Remaining` and `X-RateLimit-Reset` are kept as raw strings and
  parsed into numeric values when possible; an unparseable value does not fail
  the request;
- a JSON cursor envelope exposes `next_cursor`, `prev_cursor`,
  `next_page_results`, `prev_page_results`, `count`, `total_pages`,
  `total_results`, and `extra_stats`;
- no arbitrary response headers are retained.

`plane.CursorPage[T]` is generic. Use `plane.RawCursorPage` (raw JSON results)
when a resource model is not yet available. Unknown top-level page fields are
retained in `Unknown`, and `json.RawMessage` preserves dynamic result and
nullable values.

`plane.APIError` retains the status code and only the documented `error`,
`detail`, and `message` fields. Non-JSON failures receive a bounded safe
message. Configured credentials and URL-shaped sensitive values are redacted;
request bodies, auth headers, form fields, upload content, and sensitive URLs
are never retained in errors. Context cancellation and deadline errors are
returned as their original context errors.

`plane.WithPagination`/`plane.PaginationQuery` copies an existing query and
adds `per_page`, `cursor`, `fields`, and `expand` without dropping unknown
resource-specific values. `per_page` is validated strictly to the documented
range of 1–100; it is never silently clamped.

The official introduction documents `X-API-Key`, while the checked-in Prism
contract and PSFT-5 ticket also require bearer mode. This implementation
supports exactly those two explicit modes and does not add an auth fallback.
The contract snapshot contains some legacy `/issues/` paths for validation
coverage; this foundation adds no aliases, and later resource slices must use
current `/work-items/` paths where documented.

## Upload transport

`plane.HTTPClient.Upload` accepts an `UploadRequest` containing the returned
method, absolute URL, headers, form fields, and raw/file reader. The upload URL
is used exactly as returned; it is not joined to the Plane base and Plane
authentication is never added. A raw body is streamed directly. When form
fields or multipart mode require framing, fields and the file part are emitted
as a streaming multipart reader while caller-provided content headers remain
authoritative.

Upload responses close their bodies and return status-only `UploadError`
values. Presigned URLs, fields, API keys, bearer tokens, and uploaded content do
not appear in default errors or output. Uploads honor caller context
cancellation/deadlines and do not retry.

## Output handling

`objects.RawJSON` implements the existing `config.Outputtable` shape for JSON,
YAML, GRON, text, and raw output. Its raw document preserves unknown fields and
nulls. `config` remains the single output-format dispatch point; resource
packages do not implement their own format switch. Existing root `version`
defaults to JSON and `get version` defaults to text.

## Deterministic verification

The reusable helpers in `internal/testsupport/http.go` wrap `httptest.Server`
and assert method/path/query/headers/body while emitting JSON, empty, malformed
JSON, non-JSON, and 204 responses. Foundation tests use only local servers or
injected transports:

```sh
CI=true go test -count=1 ./...
```

The documented build workflow remains the one in [`LOCAL_BUILD.md`](LOCAL_BUILD.md).
No live Plane call or committed credential is needed for this foundation.
