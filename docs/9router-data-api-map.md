# 9router Data and API Mapping

## Scope

This document maps the 9router `v0.5.81` management entities to the current CLIProxyAPI-lite Go runtime. It is an implementation input, not a promise that every 9router feature is already supported.

Classification:

- **Reuse**: the Go runtime already owns equivalent data and behavior.
- **Extend**: an equivalent exists, but the management contract or stored fields are incomplete.
- **New**: no equivalent domain model exists in Go yet.
- **Unavailable**: keep the UI disabled until a safe Go implementation is designed and tested.

## Source-of-truth mapping

| Management area | 9router source and storage | CLIProxyAPI-lite source and storage | Classification | Migration direction |
| --- | --- | --- | --- | --- |
| Endpoint and API keys | `apiKeys` SQLite table; `/api/keys` and `/api/keys/:id` | `config.yaml` `api-keys`; `/v0/management/api-keys`; per-key counters at `/api-key-usage` | Extend | Keep Go configuration as the authority. Add stable key records and masked list responses without exposing full values after creation. |
| Provider connections | `providerConnections` SQLite table; `/api/providers`, `/api/providers/:id`, model/test routes | Provider-specific key arrays in `internal/config`, OAuth records under `auths/`, and `/v0/management/*-api-key`, `/auth-files`, OAuth routes | Extend | Normalize the existing config/auth records into one management read model. Mutations continue through the current config writer and auth manager. |
| Provider health and model discovery | Connection health fields plus `/api/providers/:id/test`, `/models`, `/test-models` | Auth status/model registry, `/auth-files/models`, `/model-definitions/:channel`, and authenticated `/api-call` | Extend | Expose safe health/model discovery DTOs backed by the registry and auth manager; never return credential material. |
| Model aliases and disabled models | `aliasRepo`, `disabledModelsRepo`; `/api/models/alias`, `/disabled`, `/custom` | `oauth-model-alias`, provider model mappings, `oauth-excluded-models` in `config.yaml` | Extend | Reuse current alias/exclusion semantics and add a normalized management view before adding UI editing. |
| Combos and vision adapter | `combos` SQLite table; `/api/combos`; request router resolves ordered model entries | No persisted combo entity or combo-aware request routing | New | Add a Go combo model, persistence, validation, registry exposure, ordered fallback, and capability metadata. Do not emulate combos with frontend state. |
| Usage history and request details | `usageHistory`, `usageDaily`, `requestDetails` SQLite tables; `/api/usage/*` | In-memory `internal/usageview`, Redis usage queue, `/usage-history`, SSE stream, API-key usage, and request log files | Extend | Define one normalized record. Preserve live Go collection, then add durable history/filter aggregation only where required by the UI. |
| Quota | Provider connection health/quota fetchers and `/api/usage/:connectionId` | `/quota/providers`, `/quota/fetch`, `/quota/reset`, plugin quota APIs, cooldown state | Reuse | Build the new quota page on the existing credential identity and quota providers. Add adapters only for providers not covered by Go. |
| Routing and fallback | Connection priority, active state, per-model locks, combo ordering | `routing.strategy`, retry controls, auth manager selection, quota/cooldown handling | Extend | Keep the Go auth manager and routing pipeline authoritative. Combo routing must compose with these controls instead of bypassing them. |
| CLI tools | Per-tool filesystem readers/writers under `/api/cli-tools/*` | `/v0/management/configure-tool` plus current endpoint/model configuration | Extend | Inventory supported tools around the existing safe writer. Add status and preview contracts before expanding tool-specific mutations. |
| Token Saver | Headroom and PXPIPE process APIs plus tool-specific integrations | No RTK, Headroom, Caveman, Ponytail, or PXPIPE service model | Unavailable | Do not expose toggles. Design each capability as an optional backend service with explicit installed/running/error state. |
| Proxy pools and provider nodes | `proxyPools` and `providerNodes` SQLite tables; `/api/proxy-pools`, `/provider-nodes` | One global `proxy-url`; no pool/node entity | New | Preserve the existing global proxy setting. Add pools/nodes only with explicit selection and request-pipeline integration. |
| Logs and translator | Console buffer and `/api/translator/*`; request detail records | Rotating logs, request/error log APIs, provider protocol translators under `internal/translator` | Extend | Reuse read-only log APIs. Treat interactive translation as a separate feature; do not couple the UI migration to translator internals. |
| Auth files and OAuth | Provider connection rows and `/api/oauth/*` | Auth store under `auths/`, `/auth-files`, provider OAuth URL/callback/status routes | Reuse | Keep the Go auth manager and configured storage backend authoritative. The UI consumes masked metadata and session state only. |
| Plugins and skills | MCP/skills pages and local plugin configuration | Dynamic plugin store/config/quota management APIs | Extend | Present Go plugins as the supported extension unit. 9router skill/MCP concepts require a separate compatibility design. |
| Media providers | Provider connections plus `/api/media-providers/*` and media `/api/v1/*` routes | Core OpenAI/Gemini/Claude/Codex-compatible routes; no matching management entity for all media kinds | New | Add only after runtime endpoints and provider executors exist. A navigation page alone does not count as support. |
| Settings and locale | `settings` SQLite row; `/api/settings`, `/api/locale` | `config.yaml`, management settings endpoints, no source-owned frontend locale store yet | Extend | Server settings remain in Go config. UI locale is a non-secret presentation preference with English default and Vietnamese translations. |
| Authentication to management | Dashboard session/login and optional OIDC/SAML | Management secret/local password middleware and remote-management rules | Reuse | Preserve current management authentication boundary. OIDC/SAML are out of scope until explicitly designed for the Go server. |
| MITM | CLI tool MITM routes and local process/config state | `/mitm/status`, certificate install/uninstall, start/stop, DNS and mappings | Reuse | Build UI directly on the authenticated Go APIs and retain platform-specific error reporting. |
| Tunnel and updater | Cloudflare/Tailscale process APIs and application updater | No equivalent management domain; binary/version update is external to the management UI | Unavailable | Exclude from the initial replacement. Add only with a separately reviewed lifecycle and security model. |

## Persistence ownership

The migration must not introduce a second database that competes with existing runtime state.

| Data | Authoritative target |
| --- | --- |
| Server and provider configuration | `config.yaml`, written through existing management/config helpers |
| OAuth and imported credentials | Configured auth store (`auths/`, Postgres, git, or object store) through the auth manager |
| Runtime models and availability | `internal/registry` plus auth-manager state |
| Live usage | Existing Go usage pipeline, `internal/usageview`, and Redis queue integration |
| Logs | Existing rotating/request log infrastructure |
| New combo definitions | A new versioned Go-owned store; storage design is required before implementation |
| Frontend locale/theme/sidebar state | Browser presentation storage only; never provider configuration or secrets |

## Security and compatibility constraints

1. Existing `/v1`, Gemini, Claude, Codex, WebSocket, and SDK behavior remains authoritative during the management migration.
2. New management endpoints remain under the existing authenticated `/v0/management` boundary.
3. List/read contracts return masked credential metadata. Secret values may be accepted on mutation but are never echoed, logged, embedded in assets, or persisted in browser storage.
4. The configured storage backend remains supported; new handlers must not assume local files when the auth store is Postgres, git, or object storage.
5. `internal/translator` is not a standalone migration target. Any translator change must be part of a broader runtime feature and follow repository contribution rules.
6. Unsupported 9router areas remain explicitly unavailable until backend state, mutations, errors, and tests exist.

## Phase dependencies established by this map

1. Define versioned management DTOs around the existing Go authorities before building the new frontend.
2. Add contract tests for masked keys, normalized providers, usage records, quota, CLI status, and unavailable capabilities.
3. Implement new persistence only for domains with no Go authority, beginning with combos after the management shell and core provider contracts are stable.
4. Remove the legacy management bundle only after every required route is backed by one of the authorities listed above.
