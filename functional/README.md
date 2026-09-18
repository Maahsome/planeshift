# CLI functional tests

The functional suite runs the built `planeshift` executable as a subprocess.
It exercises the real Cobra command tree, Viper/environment configuration,
Plane client, authentication, response decoding, output, and cleanup path.
It is intentionally opt-in because it creates, updates, archives, and deletes
projects, states, labels, and work items.

## Run

Use a disposable non-production Plane workspace or a Prism validation proxy:

```sh
PLANE_FUNCTIONAL_RUN=true \
PLANE_API_URL=http://localhost:4010 \
PLANE_AUTH_MODE=api-key \
PLANE_API_KEY=prism-test-key \
PLANE_FUNCTIONAL_WORKSPACE_SLUG=functional-workspace \
mise function-test
```

`PLANE_API_URL` is required even though the CLI has a Plane Cloud default. A
mutating functional run never silently selects that default. Prism is useful
for validating the request contract; a stateful disposable non-production
Plane workspace or forwarding proxy is required for the complete state and
project-label create/read/update/delete lifecycles.

## Settings

Required settings:

| Variable | Purpose |
| --- | --- |
| `PLANE_FUNCTIONAL_RUN=true` | Explicitly enables mutating functional tests. |
| `PLANE_API_URL` | Explicit `http` or `https` target without credentials, query, or fragment. |
| `PLANE_AUTH_MODE` | Exactly `api-key` or `bearer`. |
| `PLANE_API_KEY` or `PLANE_BEARER_TOKEN` | The credential selected by `PLANE_AUTH_MODE`; configure only one. |
| `PLANE_FUNCTIONAL_WORKSPACE_SLUG` | Disposable workspace used by generated resources. |

The existing `PLANE_TIMEOUT` duration is validated and passed to the CLI. It
must be greater than zero and no greater than ten minutes. API-key mode sends
`X-API-Key`; bearer mode sends `Authorization: Bearer`. The functional suite
does not accept an implicit auth fallback.

Optional settings:

| Variable | Purpose |
| --- | --- |
| `PLANE_FUNCTIONAL_INCLUDE_LEGACY=true` | Runs the seven hidden `work-item legacy` compatibility commands against `/issues/`. Default: skipped. |
| `PLANE_FUNCTIONAL_PROJECT_TEMPLATE_ID` | Runs `project create-template` and cleans up the generated project. Default: skipped. |
| `PLANE_FUNCTIONAL_ALLOW_PRODUCTION=true` | Overrides the Plane Cloud host safety gate only after explicit team review. Prefer Prism or a non-production host. |

The template command is a Business-license feature. It is always registered
and help-tested, but it is not executed without the template ID. Deprecated
`/issues/` routes are also never called without the legacy opt-in. If an opted-in
target returns an explicit unsupported/404/405 response for those routes, the
test records a visible skip naming the `/issues/` limitation; it never
substitutes a current `/work-items/` call.

## Isolation, cleanup, and output

`mise function-test` builds a temporary test binary as
`./planeshift-functional` and passes its absolute path through
`PLANESHIFT_BINARY`. Each subprocess receives a temporary `XDG_CONFIG_HOME`,
so it cannot read or modify the operator's normal Planeshift config. Only safe
process settings and the explicitly selected Plane settings are inherited.

Every generated project, state, label, and work-item ID is tracked immediately.
Cleanup handlers use only those IDs, delete child states, labels, and work
items before their project, delete labels before their generated parent
project, unarchive an archived project before deleting it, and log
already-deleted or failed cleanup resources visibly. Names and identifiers
include timestamp and process entropy to avoid existing resources. The suite
never accepts an arbitrary existing resource ID as a destructive target.

Object commands must return valid JSON. Archive, unarchive, and delete must
succeed with quiet output, matching the documented 204 behavior. Version and
help checks run without Plane credentials when a built binary is available.
Without `PLANE_FUNCTIONAL_RUN=true`, lifecycle tests show an explicit coverage
skip; missing core settings fail the `mise` gate with a concise diagnostic.

Credentials are never command arguments. The runner redacts configured API
keys and bearer tokens from captured stdout, stderr, command failures, and
cleanup logs. Help and output assertions reject credential markers, presigned
data, and authentication material. Do not commit credentials or place them in
this document, Compose configuration, or repository files.
