# Work 50: Integrate and verify the public Plane API surface

Use this file as the final integration and verification prompt for `planeshift` (Jira ticket PSFT-11). Read `AGENTS.md`, `CONSTITUTION.md`, `ARCHITECTURE.md`, `LOCAL_BUILD.md`, [`README.md`](./README.md), [`PUBLIC_API.txt`](../PUBLIC_API.txt), and every retained prompt before editing.

## Prompt

The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt) is the route source of truth. Verify the 19 retained numbered prompts (Work 00, 17 resource prompts, and Work 50) against its 13 route groups and 143 application methods. The verification must assert exact method/path coverage, including parameter spelling and trailing slash, rather than the removed documentation catalog.

### Verification checklist

- Confirm that every one of the 143 inventory method/path pairs appears exactly once across the route matrices in the 17 resource prompts, with no omitted, duplicated, or extra pair.
- Confirm that the 19 retained numbered prompt files are the only active prompt files referenced by `work/README.md`; no deleted resource prompt is linked from the active index or retained prompt dependencies.
- Check that all commands are registered below `cmd.RootCmd`, configuration remains centralized in `cmd/root.go`, existing output formats still work, and package dependency direction follows the Constitution.
- Exercise shared behavior through representative resources and contract tests: API-key and bearer authentication, self-hosted/custom base URL, URL joining/trailing slashes, JSON and 204 responses, malformed/non-JSON errors, cursor pagination, `fields`/`expand`, rate-limit headers, context cancellation, and safe logging.
- Check current `/work-items/` routes as the preferred family and verify that only the explicitly listed deprecated `/issues/` compatibility paths occur in the retained matrices or implementation. Do not remove a listed alias merely because it is deprecated, and do not add an unlisted alias.
- Review listed asset upload methods for exact returned metadata handling and redaction. Do not describe an attachment upload flow or page slice absent from the public inventory.
- Run `git diff --check`, validate Markdown links, and confirm that the final diff is documentation-only: no Go implementation files, tests, dependencies, or `PUBLIC_API.txt` changes.
- The Go build/test workflow in `LOCAL_BUILD.md` is not required for this documentation-only reconciliation. If implementation files are accidentally changed, restore the scope or run the documented non-interactive workflow before completion.

### Definition of done

- The 143 public method/path pairs have exact one-to-one coverage in the retained matrices.
- The 19 active numbered prompts, 13 route groups, operation counts, dependencies, primary routes, and bounded compatibility aliases are accurately indexed.
- No deleted-file links, obsolete operation totals, unsupported operations, or contradictory alias policy remains in active work prompts.
- `git diff --check` and Markdown-link validation pass, and the final diff preserves `PUBLIC_API.txt` and contains no Go code changes.
