# Planeshift Plane API implementation work prompts

This directory contains the copy/paste-ready implementation prompts for the public Plane API surface in `planeshift` (Jira ticket PSFT-11). The source-controlled [`PUBLIC_API.txt`](../PUBLIC_API.txt), generated from the Plane source snapshot identified in that file, controls route inclusion. The broader Plane documentation index is useful for operation details but is not a route authorization source.

There are 19 retained numbered prompt files: one shared client foundation, 17 resource prompts, and one final integration/verification prompt. The public inventory has 13 route groups and 143 application methods. Work Items has 26 current methods and 24 explicitly listed deprecated `/issues/` compatibility aliases; the remaining methods cover projects, project labels, states, cycles, modules, intake, assets, estimates, stickies, workspace invitations, members, and the current user.

The workspace-level `/project-labels/` route is intentionally excluded because it is not present in the active inventory. Historical tickets, reports, completed plans, and the inventory itself are outside the active prompt set.

## Suggested execution order

1. Complete [Work 00](./00-api-client-foundation.md) and record any genuinely uncharted client/CLI boundary before resource work.
2. Implement the retained resource prompts in the order below. Follow each prompt's dependency references and keep `/work-items/` routes primary.
3. Complete [Work 50](./50-integration-and-verification.md) after all retained resource slices are present.
4. Keep each implementation slice focused. If the public inventory and an operation page disagree, record the exact method/path, source snapshot, response, and decision rather than inventing a route or silently omitting a listed method.

## Retained prompt index

| Order | File | Plane scope | Methods | Key dependencies |
|---:|---|---|---:|---|
| 00 | [00-api-client-foundation.md](./00-api-client-foundation.md) | Shared client foundation | — | Constitution, ADR-001/ADR-002 |
| 01 | [01-projects.md](./01-projects.md) | Projects | 9 | 00 |
| 04 | [04-work-items.md](./04-work-items.md) | Work items, relations, and listed issue aliases | 16 | 00, 01 |
| 05 | [05-work-item-states.md](./05-work-item-states.md) | States | 5 | 00, 01 |
| 06 | [06-work-item-labels.md](./06-work-item-labels.md) | Project-scoped labels | 5 | 00, 01 |
| 11 | [11-work-item-links.md](./11-work-item-links.md) | Current and deprecated work-item links | 10 | 00, 04 |
| 12 | [12-work-item-activity.md](./12-work-item-activity.md) | Current and deprecated work-item activity | 4 | 00, 04 |
| 13 | [13-work-item-comments.md](./13-work-item-comments.md) | Current and deprecated work-item comments | 10 | 00, 04 |
| 14 | [14-work-item-attachments.md](./14-work-item-attachments.md) | Current and deprecated work-item attachments | 10 | 00, 04 |
| 16 | [16-cycles.md](./16-cycles.md) | Cycles | 14 | 00, 01, 04 |
| 17 | [17-modules.md](./17-modules.md) | Modules | 12 | 00, 01, 04 |
| 20 | [20-intake.md](./20-intake.md) | Intake issues | 5 | 00, 01, 04 |
| 21 | [21-assets.md](./21-assets.md) | User and workspace assets | 9 | 00 |
| 23 | [23-estimates.md](./23-estimates.md) | Estimates and estimate points | 8 | 00, 01 |
| 44 | [44-stickies.md](./44-stickies.md) | Stickies | 6 | 00 |
| 46 | [46-workspace-invitations.md](./46-workspace-invitations.md) | Workspace invitations | 6 | 00, 47 |
| 47 | [47-members.md](./47-members.md) | Workspace and project members | 13 | 00, 01 |
| 48 | [48-user.md](./48-user.md) | Current user | 1 | 00 |
| 50 | [50-integration-and-verification.md](./50-integration-and-verification.md) | Final integration and verification | — | 00 and all retained resource prompts |

The 17 resource prompts sum to all 143 public methods. Each matrix lists the exact method and path, including path-parameter spelling and trailing slash. The `/issues/` entries are compatibility inventory only; no additional aliases are implied.

## Shared contract every prompt assumes

- Use the shared client/configuration, Cobra registration, output, logging, error, pagination, and restricted-config conventions from [Work 00](./00-api-client-foundation.md); do not create a second transport or configuration path.
- `PUBLIC_API.txt` is authoritative for whether a route belongs in the work plan. Preserve its exact HTTP method, path, parameter spelling, and trailing slash.
- Use the current `/work-items/` route family as primary. Implement only the deprecated `/issues/` aliases explicitly listed in the relevant matrices, and keep those aliases compatibility-only.
- Apply the operation pages for request bodies, query parameters, response shapes, OAuth scope, success status, and documented errors after route inclusion has been established by the inventory.
- Preserve nullable and dynamic JSON fields, cursor pagination metadata, documented `fields`/`expand` controls, rate-limit headers, and 204 behavior.
- Keep API keys, bearer tokens, invitation secrets, presigned URLs, and upload metadata out of logs and accidental default output.
- Do not add third-party dependencies, generated SDK code, speculative endpoints, or live Plane calls to deterministic tests. Use non-interactive commands such as `CI=true go test -count=1 ./...`.
