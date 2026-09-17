# Codex Plan

# Work Ticket

ID: PSFT-16-20260917-1322-520771
Title: switch to context

## Description
I would like to go through all of the comands/sub-commands we have so far, that expect `workspace_slug` and/or `project_id` as positional arguments, and replace them with named parameters, `—workspace` and `—projectd-id`, which will be used as an override, with the default values coming from the `context` block in the config file.

Implementation plan has been generated and saved to [plans/implementation-plan-switch-to-context.md](../../plans/implementation-plan-switch-to-context.md)
