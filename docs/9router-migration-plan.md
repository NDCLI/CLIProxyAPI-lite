# 9router Feature and UI Migration Plan

## Purpose

Replace the legacy management surface in CLIProxyAPI-lite with the feature set and visual structure used by 9router. The end state is a single, source-owned management application with real Go APIs behind every enabled control, English as the default language, and complete Vietnamese translations.

This is an implementation plan for the isolated `test` branch. The `main` branch is unchanged until the migration is reviewed and approved.

## Baseline

- Reference implementation: the local 9router checkout, version `v0.5.81`.
- Target branch: `test`.
- Current status: the existing management bundle and experimental navigation still exist; the full replacement is not complete.
- The first experiment removed fake router panels and preserved the existing management routes. That experiment is not considered feature completion.

## Rules for the migration

1. Use one source of truth for navigation, page state, and API state. Do not layer a second fake menu over the old application.
2. Every enabled control must call a real backend capability and show loading, success, error, and empty states.
3. Do not use local-only switches, placeholder cards, fake counts, or `localStorage` as a substitute for persistence.
4. Keep credentials, OAuth tokens, API keys, and generated auth files backend-only. Never package, render, or log their contents.
5. English is the persisted default. Vietnamese is a complete runtime translation, including validation, empty, loading, and error messages.
6. Translate at render time from source data. Do not mutate the rendered application with a `MutationObserver`.
7. Preserve existing compatible API behavior while adding the 9router capabilities. Document any intentional breaking change before making it.
8. When an item is completed, wrap its plan text in Markdown strikethrough (`~~...~~`) and add evidence to the completion ledger. Completed items are never silently deleted from this document.

## Work plan

### Phase 0 — Preparation

- ~~Create and push the isolated `test` branch.~~
- ~~Inspect the 9router architecture, sidebar, routes, and management API inventory.~~
- ~~Record the migration rules and completion evidence format in this document.~~

### Phase 1 — Architecture and data mapping

- ~~Map each 9router entity and API to the existing Go configuration, auth store, usage store, management assets, and request pipeline.~~
- Define versioned management API contracts for providers, endpoint keys, combos, usage, quota, token saver, CLI tools, logs, and system settings.
- Decide which 9router capabilities are supported directly, which require a Go equivalent, and which remain unavailable until a real backend exists.
- Add contract tests before replacing the UI so the migration does not depend on visual behavior.

### Phase 2 — New management shell

- Build a source-owned frontend shell and router instead of DOM injection into the legacy bundle.
- Implement the 9router navigation groups and responsive sidebar: Endpoint & Key, Providers, Combo & Vision Adapter, Usage, Quota Tracker, Token Saver, and CLI Tools.
- Add shared layout primitives for cards, tables, forms, dialogs, toasts, loading states, empty states, and error states.
- Make deep links, refresh, browser back/forward, collapsed navigation, and mobile layout work consistently.

### Phase 3 — English and Vietnamese localization

- Make English the default and persisted locale.
- Add Vietnamese translations for every visible string, including API errors and form validation.
- Add a language selector and persistence that survives reload and deep links.
- Add a source check that reports missing translation keys in either language.

### Phase 4 — Endpoint & Key and Providers

- Implement real endpoint/key CRUD, masked secret display, validation, test connection, and safe save/delete flows.
- Implement provider listing, provider configuration, OAuth/API-key status, model discovery, aliases, enable/disable, and health/error state.
- Add confirmation and rollback behavior for destructive provider and credential changes.

### Phase 5 — Combo & Vision Adapter

- Add persisted combo definitions with ordered providers, model matching, fallback rules, and vision capability metadata.
- Implement create, edit, duplicate, validate, enable/disable, and delete flows.
- Route requests through the selected combo and expose the actual fallback result in logs and usage data.

### Phase 6 — Usage and Quota Tracker

- Replace placeholder metrics with live usage history, request status, token accounting, latency, provider/model filters, and time ranges.
- Add quota limits, reset windows, remaining capacity, and provider-level error states.
- Ensure all totals use one normalized backend representation and remain correct after refresh.

### Phase 7 — Token Saver

- Implement the real token-saver capabilities represented by 9router (RTK, Headroom, Caveman, Ponytail, and PXPIPE where supported by the backend).
- Expose status, configuration, start/stop/restart, and diagnostics through authenticated management APIs.
- Show an explicit unavailable state for a capability that has no safe Go implementation; do not expose a working-looking toggle.

### Phase 8 — CLI Tools

- Implement backend status and configuration APIs for the supported CLI tools.
- Generate real, copyable commands/configuration for Claude, Codex, OpenCode, and other supported clients.
- Keep generated secrets masked and ensure copied configuration uses the selected endpoint and model settings.

### Phase 9 — Remaining 9router management areas

- Add the remaining applicable pages: auth files, OAuth, logs/translator, proxy pools, skills, media providers, system information, and quick start.
- Reuse the shared shell, localization, permissions, and error handling from earlier phases.
- Mark a page unavailable until its backend behavior is complete and tested.

### Phase 10 — Remove the legacy surface

- Switch the management entry point to the new application only after all required pages have real behavior.
- Remove the old navigation injection, fake panels, obsolete DOM translators, and unused management assets.
- Migrate or explicitly version existing configuration and auth data; do not silently discard user settings.
- Search the final bundle and source for stale labels, fake counters, legacy routes, and test-only placeholders.

### Phase 11 — Verification and release gate

- Run Go formatting, unit/integration tests, and `go build -o test-output ./cmd/server`.
- Build and lint the frontend, then exercise every management API against a clean test configuration.
- Verify English and Vietnamese at desktop and narrow/mobile widths, including deep links and reloads.
- Check that credentials do not appear in source, generated assets, logs, screenshots, or Git history introduced by this migration.
- Record branch, commit, build, test, and HTTP evidence before proposing a merge to `main`.

## Completion ledger

| Date | Item | Commit | Evidence |
| --- | --- | --- | --- |
| 2026-09-22 | Isolated `test` branch created and pushed | `6db4e2ba` | `origin/test` exists and tracks the branch |
| 2026-09-22 | 9router architecture/sidebar/API inventory captured | `6db4e2ba` | Local 9router `v0.5.81` source inspection |
| 2026-09-22 | Migration rules and checklist recorded | `4fa0061a` | This document |
| 2026-09-22 | Entity, API, and persistence ownership mapped | `a391e856` | `docs/9router-data-api-map.md`; verified against 9router schema/routes and Go management routes/config/auth/usage stores |

## Update convention

Each implementation commit must update this document when a checklist item is genuinely complete. Strike through the exact completed item, add the commit hash and verification evidence to the ledger, and leave unfinished work visible. A green build alone does not complete a feature if the corresponding UI and backend behavior are not verified.
