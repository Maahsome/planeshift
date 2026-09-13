# Plane API testing

This repository uses [Prism](https://stoplight.io/open-source/prism) as a local mock and
validation proxy for the Plane REST API.

## Files

- [spec/plane-api.yaml](spec/plane-api.yaml) is the checked-in contract snapshot.
- [docker-compose.yml](docker-compose.yml) starts the Prism mock on port 4010.

The Plane API documentation does not currently publish a downloadable OpenAPI document.
The contract therefore keeps the documented method/path surface explicit while using
permissive JSON payload schemas. Tighten a resource schema when a client test needs
field-level request or response validation.

## Run the mock

Requires Docker Compose:

```sh
docker compose up --detach
```

Prism serves the documented `/api/v1` paths at `http://localhost:4010`. Use a
non-sensitive placeholder credential when exercising the mock:

```sh
curl --fail --silent --show-error \
  --header 'X-API-Key: prism-test-key' \
  http://localhost:4010/api/v1/users/me/
```

Stop the service when finished:

```sh
docker compose down
```

Set `PRISM_PORT` to use a different host port.

## Recommended workflow

1. Run client/unit tests against the mock with the base URL set to
   `http://localhost:4010`.
2. Assert request method, `/api/v1` path, query parameters, selected authentication
   header, JSON body, and expected success status.
3. Add invalid-input cases and verify Prism rejects requests that violate the contract.
4. Run a validation proxy against a non-production Plane instance before live integration
   testing. Keep the API key in the client environment; never commit it or place it in
   Compose configuration:

```sh
PLANE_API_URL=https://api.plane.so
docker compose run --rm --service-ports prism proxy -h 0.0.0.0 /spec/plane-api.yaml "${PLANE_API_URL}"
```

The proxy command forwards the client request to the upstream and validates the
request/response against the same contract. The direct presigned S3 upload step is
outside the Plane API host and should be tested using the URL and form fields returned
by Plane's upload-credential operation.

## Contract maintenance

Update the snapshot from the official [Plane API reference](https://developers.plane.so/api-reference/introduction)
when documented routes change. Keep `/api/v1` in the path definitions; only replace
payload objects with specific schemas when the test suite depends on them. Review the
documented status codes, pagination metadata, rate-limit headers, and authentication
requirements when adding a new operation.
