# CLIProxyAPI-lite — Roadmap

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

### ✅ Done — Security hardening (commit `470ed23d`)

| Area | What was done |
|---|---|
| Loopback binding | `DefaultHost = "127.0.0.1"` used in config loader, parser, all fallback paths, `NewServer` defense-in-depth, callback forwarder |
| Antigravity OAuth | `parseAntigravityCallbackPayload` helper rejects empty/mismatched state, missing code, provider errors |
| Tests | `host_test.go`, `server_host_test.go`, `antigravity_callback_test.go` |
| Example config | `config.example.yaml` updated to local-only default |

### ⏳ Pending — Build verification

Go is not installed on the current machine. Run the three verification commands above before continuing.

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

#### 3. Settings (`/settings`)
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
