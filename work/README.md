# Planeshift Plane API implementation work prompts

This directory breaks the Plane REST API integration into copy/paste-ready implementation prompts. The source snapshot is the official [Plane API introduction](https://developers.plane.so/api-reference/introduction.md) and its [machine-readable documentation index](https://developers.plane.so/llms.txt), reviewed on 2026-09-12.

There are 51 prompt files: one shared foundation, one file for each of the 49 documentation resource sections, and one final integration/verification slice. Together they cover 291 non-overview documentation entries: 290 HTTP operations plus the Pages `page-content-html` content guide.

## Suggested execution order

1. Complete [Work 00](./00-api-client-foundation.md) and settle its explicitly uncharted client/CLI boundary.
2. Implement the numbered resource prompts. They are organized in the same order as the Plane API documentation; prompts with related-resource dependencies should follow those dependencies in the table.
3. Complete [Work 50](./50-integration-and-verification.md) after all resource slices are present.
4. Keep each implementation slice focused. If the live API or documentation contradicts a prompt, record the exact endpoint, version/host, response, and decision rather than inventing a fallback or omitting the operation.

## Shared contract every prompt assumes

- Plane Cloud base: `https://api.plane.so/api/v1/`; self-hosted base: the configured instance host plus `/api/v1/`.
- Authenticate with one of `X-API-Key` or `Authorization: Bearer`; secrets and presigned upload data must not appear in logs.
- Use JSON and standard REST status handling; treat 204 as an empty response.
- Use cursor pagination (`per_page` up to the documented server maximum of 100 and `cursor`) and preserve pagination metadata. Expose documented `fields` and `expand` values.
- Honor the documented 60-requests-per-minute rate-limit headers and return bounded, useful errors. Do not introduce unbounded automatic retries.
- Prefer current `/work-items/` routes. The documentation uses legacy directory names for some pages and warns about deprecated `/issues/` routes; do not add deprecated aliases.
- Do not add third-party dependencies or generated SDK code. Preserve the existing Cobra/config/output conventions, and run tests non-interactively with `CI=true go test -count=1 ./...`.

## Prompt index

| File | Plane section | HTTP operations |
|---|---|---:|
| [00-api-client-foundation.md](./00-api-client-foundation.md) | Shared client foundation | — |
| [01-projects.md](./01-projects.md) | Projects | 8 |
| [02-project-features.md](./02-project-features.md) | Project Features | 2 |
| [03-project-labels.md](./03-project-labels.md) | Project Labels | 5 |
| [04-work-items.md](./04-work-items.md) | Work Items | 8 |
| [05-work-item-states.md](./05-work-item-states.md) | Work Item States | 5 |
| [06-work-item-labels.md](./06-work-item-labels.md) | Work Item Labels | 5 |
| [07-work-item-types.md](./07-work-item-types.md) | Work Item Types | 6 |
| [08-custom-properties.md](./08-custom-properties.md) | Custom Properties | 5 |
| [09-custom-property-values.md](./09-custom-property-values.md) | Custom Property Values | 5 |
| [10-custom-property-options.md](./10-custom-property-options.md) | Custom Property Options | 5 |
| [11-work-item-links.md](./11-work-item-links.md) | Work Item Links | 5 |
| [12-work-item-activity.md](./12-work-item-activity.md) | Work Item Activity | 2 |
| [13-work-item-comments.md](./13-work-item-comments.md) | Work Item Comments | 5 |
| [14-work-item-attachments.md](./14-work-item-attachments.md) | Work Item Attachments | 7 |
| [15-work-item-page-links.md](./15-work-item-page-links.md) | Work Item Page Links | 4 |
| [16-cycles.md](./16-cycles.md) | Cycles | 12 |
| [17-modules.md](./17-modules.md) | Modules | 11 |
| [18-pages.md](./18-pages.md) | Pages | 18 (+ 1 guide) |
| [19-collections.md](./19-collections.md) | Collections | 14 |
| [20-intake.md](./20-intake.md) | Intake | 5 |
| [21-assets.md](./21-assets.md) | Assets | 6 |
| [22-milestones.md](./22-milestones.md) | Milestones | 6 |
| [23-estimates.md](./23-estimates.md) | Estimates | 8 |
| [24-time-tracking.md](./24-time-tracking.md) | Time Tracking | 5 |
| [25-epics.md](./25-epics.md) | Epics | 7 |
| [26-initiatives.md](./26-initiatives.md) | Initiatives | 5 |
| [27-initiative-labels.md](./27-initiative-labels.md) | Initiative Labels | 8 |
| [28-initiative-projects.md](./28-initiative-projects.md) | Initiative Projects | 3 |
| [29-initiative-epics.md](./29-initiative-epics.md) | Initiative Epics | 3 |
| [30-releases.md](./30-releases.md) | Releases | 5 |
| [31-project-releases.md](./31-project-releases.md) | Project Releases | 5 |
| [32-release-work-items.md](./32-release-work-items.md) | Release Work Items | 3 |
| [33-release-labels.md](./33-release-labels.md) | Release Labels | 8 |
| [34-release-tags.md](./34-release-tags.md) | Release Tags | 5 |
| [35-release-comments.md](./35-release-comments.md) | Release Comments | 5 |
| [36-release-links.md](./36-release-links.md) | Release Links | 5 |
| [37-release-changelog.md](./37-release-changelog.md) | Release Changelog | 2 |
| [38-customers.md](./38-customers.md) | Customers | 8 |
| [39-customer-properties.md](./39-customer-properties.md) | Customer Properties | 8 |
| [40-customer-requests.md](./40-customer-requests.md) | Customer Requests | 5 |
| [41-teamspaces.md](./41-teamspaces.md) | Teamspaces | 5 |
| [42-teamspace-members.md](./42-teamspace-members.md) | Teamspace Members | 3 |
| [43-teamspace-projects.md](./43-teamspace-projects.md) | Teamspace Projects | 3 |
| [44-stickies.md](./44-stickies.md) | Stickies | 5 |
| [45-workspace-features.md](./45-workspace-features.md) | Workspace Features | 2 |
| [46-workspace-invitations.md](./46-workspace-invitations.md) | Workspace Invitations | 5 |
| [47-members.md](./47-members.md) | Members | 7 |
| [48-user.md](./48-user.md) | User | 1 |
| [49-idp-group-sync.md](./49-idp-group-sync.md) | IDP Group Sync | 12 |
| [50-integration-and-verification.md](./50-integration-and-verification.md) | Final integration and verification | — |

