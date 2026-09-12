# Work 50: Integrate and verify the complete Plane API surface

Use this file as the implementation prompt for final integration and contract verification of `planeshift` (Jira ticket PSFT-2).

## Prompt

You are an AI coding agent working in the `planeshift` repository. Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, `work/README.md`, and all completed work prompts before editing.

Verify that the foundation and all 49 resource slices together implement the complete documented Plane API surface. This work item is for integration, consistency, and tests; do not add a new resource or silently remove an existing operation.

### Verification checklist

- Confirm every operation link in `work/01-projects.md` through `work/49-idp-group-sync.md` has exactly one client method, CLI command, typed request/response contract, and deterministic test. The inventory is 290 HTTP operations plus the Pages content guide.
- Check that all commands are registered below `cmd.RootCmd`, configuration remains centralized in `cmd/root.go`, existing output formats still work, and package dependency direction follows the Constitution.
- Exercise shared behavior through representative resources and contract tests: API-key and OAuth authentication, self-hosted/custom base URL, URL joining/trailing slashes, JSON and 204 responses, malformed/non-JSON errors, cursor pagination, `fields`/`expand`, rate-limit headers, context cancellation, and safe logging.
- Check that every list operation exposes the documented pagination/filter/query controls and that no response decoder discards nullable or dynamic JSON fields.
- Check the three-step upload flows for work-item attachments and assets, plus the Pages content/attachment guidance. Ensure presigned requests use returned upload data and do not leak credentials.
- Search for deprecated `/issues/` request paths and remove only newly introduced aliases; do not delete working unrelated code. Compare route choices with the linked operation pages when documentation slugs are legacy.
- Run the repository’s documented build and test workflow non-interactively, including `CI=true go test -count=1 ./...` and the build command from `LOCAL_BUILD.md`. If live smoke tests are possible, use explicitly supplied credentials only and never commit them.
- Record documentation/API-version mismatches (especially Pages availability or attachment-flow contradictions) as clear notes or focused tests; do not hide them behind fallback behavior.

### Definition of done

The complete CLI builds, existing commands remain compatible, all 290 API operations are covered, tests are deterministic without live credentials, no unapproved dependency was added, and the final diff is limited to the requested Plane API implementation and its verification support.

