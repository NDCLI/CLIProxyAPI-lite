# Lumina — Roadmap

Lightweight, local-only AI gateway forked from [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI).
One binary, one machine, web UI, no cloud infrastructure.

---

## Prerequisites

Install Go 1.26+ and add it to PATH before any code work.

```bash
# Windows
winget install GoLang.Go
# then restart terminal so `go` is on PATH

# verify
go version          # ≥ 1.26
gofmt -w .          # format
go build -o test-output ./cmd/server && rm test-output   # compile check
go test ./...        # run all tests
```

---

## Current status

### ✅ Done — Completed milestones

| Commit | Area | What was done |
|---|---|---|
| `470ed23d` | Security hardening | `DefaultHost = "127.0.0.1"` loopback binding across config, parser, server, and forwarder; strict Antigravity OAuth callback state validation |
| `93d8fe10` | Roadmap & Architecture | Created `ROADMAP.md` detailing 6 phases to transform upstream into a streamlined local-only gateway |
| `a426525d` | Documentation | Created `docs/cli-ide-setup.md` covering setup instructions for 10+ AI tools |
| `02396353` | Tool auto-configuration v1 | Added backend handler `POST /v0/management/configure-tool` with tests, registered route in management API, created `web/connect.html` with model picker |

### ⏳ In progress — Separate CLI Tools vs IDEs & MITM Proxy for IDEs

The embedded Vietnamese setup UI now serves `/` and `/connect`, with separate CLI, direct IDE, MITM, 9router account-import, and Endpoint/API key tabs. The Endpoint tab follows 9router's local endpoint plus masked key-list pattern and reuses the existing management API for create, reveal, copy, and delete actions. Per-tool model configuration and the Root CA backend are implemented. The Root CA is generated locally and can be downloaded, installed, or removed through the management API on Windows.

9router JSON import accepts a single account, an account array, or a full backup containing `providerConnections`. It converts Codex, Claude, and Antigravity accounts into native auth files while skipping unsupported providers without exposing tokens in the response.

The loopback MITM listener now runs on `127.0.0.1:443`, generates SNI leaf certificates, supports per-tool DNS and model mappings for GitHub Copilot and Antigravity, passes unmapped traffic through, and reuses the existing Gemini/OpenAI translators for mapped Antigravity traffic. Raw request/response logging is deliberately disabled to avoid persisting prompts or credentials.

Claude Code and Codex CLI configuration now follows each tool's native files. Apply merges only the gateway-owned settings; Reset removes those settings while preserving unrelated user configuration.

### ✅ Done — Toolchain & build verification

Verified on Windows with Go 1.27.0: `gofmt`, server build, focused management/OAuth tests, and `go test ./...` pass.

---

## Phase 1 — Verify & stabilize the fork

**Goal:** confirm the fork compiles, tests pass, and nothing is broken by the security changes.

1. Install Go 1.26+
2. Run `gofmt -w .` to format all files
3. Run `go build -o test-output ./cmd/server && rm test-output`
4. Run `go test ./...`
5. Fix any compilation or test failures

---

## Phase 2 — Prune providers & protocols outside MVP scope

**Goal:** remove dead weight so the binary is smaller and the codebase is easier to navigate.

### Providers to KEEP
- **Codex** (OpenAI) — `internal/auth/codex/`, executor, translator
- **Antigravity** (Anthropic/Claude) — `internal/auth/antigravity/`, executor, translator
- **Claude** (Anthropic direct API keys) — `internal/auth/claude/`, executor, translator

### Providers to REMOVE from product surface
These should be excluded from the binary and route registration. Do not delete the source files yet — move them to a `_disabled/` directory or use build tags so they can be restored later if needed.

- Gemini / AI Studio / Vertex — `internal/auth/vertex/`, gemini handlers, `v1beta/` routes
- Kimi — `internal/auth/kimi/`
- xAI / Grok — `internal/auth/xai/`, grokbuild, xAI video routes
- Meta — `internal/auth/meta/`
- Devin — `internal/auth/devin/`
- OpenAI Compatibility (third-party) — `openai-compatibility` config and handlers
- Vertex Compat API keys — `vertex-api-key` config and handlers

### Routes to REMOVE
- `/v1beta/*` (Gemini)
- `/v1/images/*`, `/v1/videos/*` (image/video generation)
- `/v1/realtime/*`, `/v1/live` (Codex Live/WebSocket relay — complex, not MVP)
- `/openai/v1/videos/*`
- `/backend-api/codex/*` (Codex CLI direct aliases)
- `/v1/alpha/search` (Codex Alpha Search)
- `/callback`, `/devin/callback` (Devin OAuth)
- Management routes for removed providers: `gemini-api-key`, `interactions-api-key`, `xai-api-key`, `meta-api-key`, `vertex-api-key`, `openai-compatibility`, `kimi-*-auth-url`, `xai-auth-url`, `devin-auth-url`, `meta-auth-url`
- Plugin management routes (`/plugins/*`, `/plugin-store/*`)
- Usage queue, request logging, log file management routes

### Routes to KEEP
- `/healthz`
- `/v1/models`
- `/v1/chat/completions`
- `/v1/completions`
- `/v1/responses` (OpenAI Responses API)
- `/v1/messages` (Anthropic Messages API)
- `/v1/messages/count_tokens`
- `/anthropic/callback`, `/codex/callback`, `/antigravity/callback`
- Management: `oauth-callback`, `config`, `api-keys`, `auth-files` (CRUD/refresh/status), `anthropic-auth-url`, `codex-auth-url`, `antigravity-auth-url`, `get-auth-status`, `oauth-session` (cancel), `quota/providers`, `quota/fetch`, `debug`, `codex-api-key`, `claude-api-key`

### Management routes to KEEP (for the web UI)
```
GET    /v0/management/config
GET    /v0/management/api-keys
PUT    /v0/management/api-keys
PATCH  /v0/management/api-keys
DELETE /v0/management/api-keys
GET    /v0/management/auth-files
GET    /v0/management/auth-files/models
DELETE /v0/management/auth-files
PATCH  /v0/management/auth-files/status
POST   /v0/management/auth-files/refresh
GET    /v0/management/anthropic-auth-url
GET    /v0/management/codex-auth-url
GET    /v0/management/antigravity-auth-url
GET    /v0/management/get-auth-status
DELETE /v0/management/oauth-session
POST   /v0/management/oauth-callback
GET    /v0/management/oauth-callback
GET    /v0/management/quota/providers
POST   /v0/management/quota/fetch
GET    /v0/management/claude-api-key
PUT    /v0/management/claude-api-key
GET    /v0/management/codex-api-key
PUT    /v0/management/codex-api-key
POST   /v0/management/configure-tool
```

### Packages to REMOVE or disable
- `internal/tui/` — Bubbletea TUI (replaced by web UI)
- `internal/home/` — CLIProxyAPIHome control plane
- `internal/store/` — Postgres, Git, Object storage backends (keep file-based only)
- `internal/watcher/` — evaluate; may keep config hot-reload
- `internal/wsrelay/` — WebSocket relay sessions
- `internal/usage/` — heavy usage analytics (keep basic quota check)
- `internal/managementasset/` — management panel GitHub auto-update
- `internal/client/grokbuild/` — Grok build/keepalive
- `internal/client/codex/live/` — Codex Live media relay
- `internal/client/codex/optimize-multi-agent-v2/` — Codex multi-agent optimization
- `internal/api/modules/amp/` — Amp integration

### How to prune safely

1. Build the import graph: `go list -json ./... | jq '.ImportPath, .Imports[]'`
2. Remove route registrations first (in `server_routes.go` and `server_management.go`)
3. Remove handler references, then executor/translator registrations
4. Remove auth provider registrations
5. Remove config fields for disabled providers
6. Compile after each removal — `go build -o test-output ./cmd/server && rm test-output`
7. Run `go test ./...` after the full removal pass

---

## Phase 3 — Web UI

**Goal:** a local web interface served by the Go binary for managing accounts and viewing status.

### Tech stack
- Embedded static files via Go `embed` package
- Vanilla HTML + CSS + JS (no build step, no npm, no framework)
- Fetch calls to `/v0/management/*` endpoints
- Served at `GET /` and `GET /static/*` by the Go server

### Pages

#### 1. Dashboard (`/`)
- Server status (listening address, uptime)
- Provider status cards: Codex, Antigravity, Claude
- Quick stats: number of active accounts per provider, cooldown status

#### 2. Provider detail (`/provider/:name`)
- Account list with status (active, cooling, disabled, expired)
- Login button → calls `GET /v0/management/{provider}-auth-url`, opens OAuth in new tab
- OAuth progress indicator (polls `GET /v0/management/get-auth-status?state=...`)
- Per-account actions: refresh, disable, delete
- Quota info (calls `POST /v0/management/quota/fetch`)

#### 3. Connect (`/connect`)
- **CLI/IDE quick setup cards** — one card per tool, each with:
  - Tool logo/icon and name
  - The exact env vars or config to set (pre-filled with the current server address and a configured API key)
  - Copy-to-clipboard button for each command/snippet
  - "Test connection" button that pings `/v1/models` with the key
- Supported tools (cards):
  - Claude Code CLI — `ANTHROPIC_BASE_URL` + `ANTHROPIC_API_KEY`
  - Codex CLI — `OPENAI_BASE_URL` + `OPENAI_API_KEY`
  - Cursor — OpenAI API Base URL + Key
  - Windsurf — OpenAI Compatible provider settings
  - VS Code (Continue.dev) — `config.json` snippet
  - VS Code (Cline) — extension settings
  - aider — env vars for both OpenAI and Anthropic modes
  - Python SDK — code snippet (OpenAI + Anthropic)
  - Node.js SDK — code snippet
  - cURL — ready-to-run commands
- Auto-detect server address from `window.location` for the snippets
- Link to full docs: `docs/cli-ide-setup.md`

#### 4. Settings (`/settings`)
- API keys: show current proxy API keys, add/remove
- Endpoint reference: copy-paste examples for OpenAI-compatible and Anthropic-compatible
- Model list: calls `GET /v1/models` and displays available models

### UI behavior
- Auto-refresh account status every 10 seconds
- OAuth login opens provider page in a new browser tab; UI polls for completion
- Toast notifications for success/error on actions
- Responsive layout, works on a single machine's browser
- No authentication on the UI itself (management API is localhost-only)

### File structure
```
web/
  connect.html        # Auto-configure IDE/CLI page with model picker
  index.html          # SPA shell with client-side routing
  static/
    style.css         # All styles
    app.js            # All JS — fetch wrappers, routing, rendering
    favicon.svg       # Simple icon
```

### Embedding
```go
//go:embed web/*
var webFS embed.FS

// In setupRoutes():
s.engine.StaticFS("/static", http.FS(sub))
s.engine.GET("/", func(c *gin.Context) { /* serve index.html */ })
```

---

## Phase 3.1 — Separate CLI Tools vs IDEs & MITM Proxy Integration

**Goal:** Separate configuration workflows for CLI tools vs IDEs, and support IDEs/extensions that cannot configure custom API base URLs directly (e.g. VS Code Copilot, Google Antigravity IDE) via a local MITM interception proxy (similar to 9router).

### 1. Separation of Concerns & UX

- **CLI Tools (`[CLI Tools]` tab)**:
  - **Targets**: Claude Code CLI, Codex CLI, Aider, OpenCode, Shell scripts / curl.
  - **Connection mechanism**: Direct local environment variables (`ANTHROPIC_BASE_URL`, `OPENAI_BASE_URL`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`) and tool-specific config files (`~/.claude/settings.json`, Windows user env via `setx`, Linux `~/.bashrc`).
  - **UX**: 1-click write configuration files or copy-paste shell export blocks.

- **IDEs & Code Editors (`[IDEs & Editors]` tab)**:
  - **Category A: Direct Base URL IDEs**:
    - Cursor, Windsurf, VS Code with Continue.dev, VS Code with Cline.
    - Mechanism: Direct configuration file writes (`%APPDATA%/Cursor/User/settings.json`, `~/.continue/config.json`, etc.).
  - **Category B: Intercepted / MITM Proxy IDEs**:
    - VS Code with official GitHub Copilot extension, Google Antigravity IDE.
    - Mechanism: Local transparent / MITM HTTPS proxy intercepting outbound AI requests without requiring hardcoded endpoint changes in the IDE.

### 2. Local MITM Proxy Architecture for IDEs

For tools that connect only to hardcoded vendor endpoints:

1. **Local MITM Proxy Engine (`internal/proxy/mitm/`)**:
   - Loopback HTTP/HTTPS proxy listener (e.g., `127.0.0.1:8318` or unified port).
   - Handles `CONNECT` tunneling for HTTPS traffic.
   - Intercepts requests targeting known AI endpoints (Copilot completions, Antigravity API domains).
   - Passthrough: non-AI domain traffic is forwarded untouched or blocked according to security policy.

2. **Root Certificate Authority (CA) & 1-Click Auto-Installation**:
   - Automatically generate local Root CA certificate (`ca.crt`) and private key (`ca.key`) on first launch under `certs/` (excluded from git).
   - Dynamic on-the-fly certificate generation for intercepted domains signed by local CA.
   - Endpoint `GET /v0/management/mitm/ca.crt` to download the public CA certificate.
   - **Backend 1-Click Auto-Install (`POST /v0/management/mitm/install-cert`)**:
     - Automatically installs the certificate directly from the server without manual terminal commands:
       - **Windows**: executes `certutil -addstore -user Root certs\ca.crt` (installs into current user's trusted root store — **no Administrator UAC elevation required**).
       - **macOS**: executes `security add-trusted-cert -d -r trustRoot -k ~/Library/Keychains/login.keychain-db certs/ca.crt`.
       - **Linux**: installs into user/system trust store or sets `NODE_EXTRA_CA_CERTS`.
     - Endpoint `POST /v0/management/mitm/uninstall-cert`: cleanly uninstalls the root certificate (`certutil -delstore -user Root "CLIProxyAPI Root CA"`).

3. **IDE Proxy Configuration**:
   - Backend endpoint `POST /v0/management/configure-ide`:
     - **VS Code**: updates `settings.json` with:
       ```json
       {
         "http.proxy": "http://127.0.0.1:8318",
         "http.proxyStrictSSL": false
       }
       ```
       or injects `NODE_EXTRA_CA_CERTS` so strict SSL remains valid.
     - **Antigravity IDE**: configures IDE proxy environment variables (`HTTPS_PROXY`, `NODE_EXTRA_CA_CERTS`) or sidecar hook.

4. **Payload Translation & Upstream Execution**:
   - Intercepted Copilot / Antigravity requests are parsed into canonical internal formats.
   - Dispatched through the existing account pool (Codex, Antigravity, Claude).
   - Formatted and streamed back to the IDE matching the vendor's protocol response schema.

### 3. Web UI Updates (`web/connect.html` & `web/index.html`)

- **Vertical Sidebar Tab Layout (Left sidebar navigation + Right content viewport)**:
  - Replaces the long vertically scrolling page with a modern 2-column dashboard layout.
  - **Left Sidebar Tabs**:
    - `⚡ All Tools` — Combined overview with quick action buttons.
    - `💻 CLI Tools` — Claude Code CLI, Codex CLI, Shell Environment Variables.
    - `🖥️ IDEs (Direct API)` — Cursor, Windsurf, VS Code (Continue.dev), VS Code (Cline).
    - `🛡️ IDEs (MITM Proxy)` — VS Code (Copilot), Google Antigravity IDE, Root CA Cert setup.
    - Server status badge (`127.0.0.1:8317`) and active port indicator.
- **Per-Tool Model Configuration & Reset Lifecycle (Tool → Model → Apply → Reset)**:
  - Replaces global single-model assignment with independent per-tool model selection:
    - **Claude Code CLI**: configurable model (e.g. `claude-3-7-sonnet`, `claude-3-5-sonnet`) saved into `~/.claude/settings.json`.
    - **OpenAI Codex CLI**: configurable model (e.g. `gpt-4o`, `o3-mini`, `codex-v2`) exported or saved into user env.
    - **IDEs (Cursor, Continue, Cline, VS Code)**: model persisted into respective extension/IDE settings.
  - **4-Step Tool Configuration UX**:
    1. **Select Tool**: Click any CLI or IDE card from the vertical sidebar view.
    2. **Select Model**: Dedicated model picker displaying active models for that specific tool.
    3. **Apply Configuration**: 1-click apply sending `{tool, api_key, model}` to `POST /v0/management/configure-tool`.
    4. **Reset Tool**: Explicit post-apply action prompting tool restart:
       - Terminal/CLI: prompt to reopen terminal or reload shell with test command snippet.
       - IDEs: prompt to reload window (`Developer: Reload Window`) or restart IDE to load new config.
- **9router-Style Visual Design & Provider Icons**:
  - Crisp, authentic brand SVG icons for every tool and provider (matching 9router visual language):
    - **Claude Code**: Anthropic orange sunburst icon.
    - **Antigravity**: Google Antigravity rainbow prism/loop gradient icon.
    - **OpenAI Codex**: OpenAI swirl icon.
    - **GitHub Copilot**: Copilot robot visor icon.
    - **Cursor IDE**: Cursor 3D dark cube icon.
    - **Cline**: Cline purple robot icon.
    - **Continue.dev**: Continue cyan forward-play icon.
    - **Terminal / Env**: CLI terminal prompt icon.
  - 9router card components:
    - High-contrast dark cards with rounded borders.
    - Left brand icon badge + title + live status pill (`● 1 Connected` / `● Sẵn sàng`).
    - 1-click **"Tự động cài Root CA"** button in MITM Proxy section with animated progress & feedback.

---

## Phase 4 — Integration testing

**Goal:** end-to-end verification that the trimmed server works correctly.

### Test categories

1. **Config loading**
   - Missing config file → defaults to loopback
   - `host: ""` → normalized to 127.0.0.1
   - Explicit host → preserved

2. **OAuth flows** (against test fixtures, not real providers)
   - Start login → state registered → callback file written → session completed
   - Callback with empty state → rejected
   - Callback with mismatched state → rejected
   - Callback with error → session marked failed
   - Session timeout → session marked timed out
   - Session cancellation → session cleaned up

3. **Account CRUD**
   - Upload auth file → appears in list
   - Delete auth file → removed from list
   - Disable/enable auth file → status updated
   - Refresh → token refreshed (mock exchange)

4. **API proxy** (mock upstream)
   - `POST /v1/chat/completions` streaming and non-streaming
   - `POST /v1/messages` streaming and non-streaming
   - `GET /v1/models` returns models from active accounts
   - Request with invalid API key → 401
   - Request with no available accounts → appropriate error
   - Account failover: first account fails → second account used

5. **Security**
   - Management API reachable from 127.0.0.1
   - Auth files not in git (check `.gitignore`)
   - No tokens/keys in HTTP responses to client API
   - No tokens/keys in server logs

6. **Web UI smoke test**
   - `GET /` returns HTML
   - `GET /static/app.js` returns JS
   - UI can list accounts via fetch (mock server test)

---

## Phase 5 — Package & distribute

**Goal:** a single binary + config file that can be run on any Windows/Mac/Linux machine.

1. Build for current platform: `go build -o cli-proxy-api-lite ./cmd/server`
2. Cross-compile if needed:
   ```bash
   GOOS=windows GOARCH=amd64 go build -o cli-proxy-api-lite.exe ./cmd/server
   GOOS=darwin GOARCH=arm64 go build -o cli-proxy-api-lite-darwin ./cmd/server
   GOOS=linux GOARCH=amd64 go build -o cli-proxy-api-lite-linux ./cmd/server
   ```
3. Verify binary does not contain secrets: `strings cli-proxy-api-lite | grep -i "token\|secret\|password"` (should show only field names, not values)
4. Create a minimal `config.example.yaml` (already done)
5. Write a `README.md` with quickstart instructions

---

## Phase 6 — Add more free providers (post-MVP)

Each new provider is added through an isolated adapter pattern:

1. Create `internal/auth/{provider}/` with auth, token, filename, constants
2. Create executor in `internal/runtime/executor/`
3. Create translator if protocol differs
4. Register the provider in the model registry
5. Add OAuth routes: `{provider}-auth-url`, `{provider}/callback`
6. Add management CRUD routes for the provider's accounts
7. Add provider card to the web UI
8. Add integration tests for the new provider
9. Update `config.example.yaml` if new config fields are needed

### Candidate free providers
- Google AI Studio (Gemini) — free tier available
- Groq — free tier
- Together AI — free tier
- Others as they become available

---

## Security invariants (must hold at every phase)

- [ ] `host` defaults to `127.0.0.1`; explicit `""` is treated as loopback
- [ ] Management API binds localhost only
- [ ] OAuth callback forwarder binds `127.0.0.1`, not `0.0.0.0`
- [ ] Auth files stored outside source tree, excluded by `.gitignore`
- [ ] Access tokens, refresh tokens, API keys never appear in logs
- [ ] Access tokens, refresh tokens never appear in HTTP responses to client API
- [ ] OAuth state must be non-empty and matched exactly
- [ ] OAuth sessions are cancellable and time-bounded (5 min default)
- [ ] Client API keys are separate from provider OAuth credentials
- [ ] Final binary does not contain embedded secrets

---

## Quick reference: startup

```bash
# After Go is installed and the binary is built:
cp config.example.yaml config.yaml
# Edit config.yaml: set port, API keys
go run ./cmd/server --config config.yaml

# Open browser:
# http://127.0.0.1:8317/          → Web UI
# http://127.0.0.1:8317/v1/models → Model list

# Client usage (OpenAI-compatible):
curl http://127.0.0.1:8317/v1/chat/completions \
  -H "Authorization: Bearer YOUR-LOCAL-API-KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'

# Client usage (Anthropic-compatible):
curl http://127.0.0.1:8317/v1/messages \
  -H "x-api-key: YOUR-LOCAL-API-KEY" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"claude-sonnet-4-20250514","max_tokens":1024,"messages":[{"role":"user","content":"Hello"}]}'
```

---

## Quick reference: connect CLI & IDE tools

After the server is running, connect your favorite tools. Full guide: [`docs/cli-ide-setup.md`](docs/cli-ide-setup.md)

```bash
# Claude Code CLI
export ANTHROPIC_BASE_URL=http://127.0.0.1:8317
export ANTHROPIC_API_KEY=YOUR-LOCAL-API-KEY

# Codex CLI
export OPENAI_BASE_URL=http://127.0.0.1:8317/v1
export OPENAI_API_KEY=YOUR-LOCAL-API-KEY

# Cursor / Windsurf / other OpenAI-compatible IDEs
# Set API Base URL to: http://127.0.0.1:8317/v1
# Set API Key to: YOUR-LOCAL-API-KEY
```
