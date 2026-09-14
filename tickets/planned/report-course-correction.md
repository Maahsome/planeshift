# Codex Plan

# Work Ticket

ID: PSFT-7-20260913-1835-8E6381
Title: Course Correction

## Description
Before we add more features we need to establish a good pattern for the cobra command structure.

We ended up with the first addition being "project" commands, and they were added under the top level "get" command.

I think our top level commands should represent the objects. "Project", "Workspace", etc.

Each top level command should be a package, and each sub-command of the top level command should be a separate file in the top level command directory.

For instance, currently "get" is a top level command, and "version" is a sub-command of get.

I think we should make "projects" a top level command, renamed to "project" and "projects" as an alias.

We should use the non-plural as the primary top level command names, with the plural as the alias.

I would like to create an "aliases.go" file as part of the "config" package, so the aliases are referenced from there.

Implementation plan has been generated and saved to [plans/implementation-plan-course-correction.md](../../plans/implementation-plan-course-correction.md)
