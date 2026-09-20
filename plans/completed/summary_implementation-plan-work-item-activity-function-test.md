# Summary of Actions Taken

- Applied the repository Constitution and ADR-001/ADR-002 boundary to the test-only functional extension.
- Added executable help coverage for `activity`, `activities`, canonical list/get operations, and hidden legacy list/get operations, including context/query flags and credential-safety checks.
- Added page-based `activityID` extraction and opt-in canonical and legacy activity lifecycles using generated project/work-item prerequisites and existing cleanup handlers.
- Documented activity read-only coverage, observed-ID retention, parent-resource isolation, and legacy `/issues/` skip behavior in `functional/README.md`.
- Formatted and verified the suite with `CI=true PLANE_FUNCTIONAL_RUN=false go test -count=1 ./...`, the documented v0.0.999 build/version workflow, and `git diff --check`.
- Ran the explicitly configured non-production `CI=true mise function-test` successfully; detail assertions preserve the target's returned activity identity, and no credentials were exposed.
- No repository lint task was defined, no production activity code was changed, and no new dependencies were introduced.
