# CLI & IDE Connection Guide

How to connect popular AI coding tools to CLIProxyAPI-lite running at `http://127.0.0.1:8317`.

> **Prerequisites:** the server is running with at least one API key configured in `config.yaml` under `api-keys`.
> All examples below use `your-api-key` — replace it with one of your actual keys.

---

## Claude Code CLI

Claude Code connects via the Anthropic Messages API.

### Setup

```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:8317
export ANTHROPIC_API_KEY=your-api-key
```

On Windows (PowerShell):
```powershell
$env:ANTHROPIC_BASE_URL = "http://127.0.0.1:8317"
$env:ANTHROPIC_API_KEY = "your-api-key"
```

To make it permanent, add to your shell profile (`~/.bashrc`, `~/.zshrc`, or System Environment Variables on Windows).

### How it works

Claude Code appends `/v1/messages` to `ANTHROPIC_BASE_URL`, so your requests hit:
- `POST http://127.0.0.1:8317/v1/messages` — chat (streaming)
- `POST http://127.0.0.1:8317/v1/messages/count_tokens` — token counting
- `GET http://127.0.0.1:8317/v1/models` — model listing

### Verify

```bash
claude --version
claude "say hello"
```

---

## Claude Code in VS Code / Cursor / Windsurf (as Anthropic extension)

The Claude VS Code extension and Claude-compatible IDE extensions use the same env vars.

### Setup

Add to your VS Code `settings.json`:
```json
{
  "claude-code.apiBaseUrl": "http://127.0.0.1:8317"
}
```

Or set the environment variables globally before launching the IDE:
```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:8317
export ANTHROPIC_API_KEY=your-api-key
```

---

## Codex CLI (OpenAI)

Codex CLI connects via the OpenAI Responses API.

### Setup

```bash
export OPENAI_BASE_URL=http://127.0.0.1:8317/v1
export OPENAI_API_KEY=your-api-key
```

On Windows (PowerShell):
```powershell
$env:OPENAI_BASE_URL = "http://127.0.0.1:8317/v1"
$env:OPENAI_API_KEY = "your-api-key"
```

### How it works

Codex CLI uses:
- `POST http://127.0.0.1:8317/v1/responses` — main code generation
- `GET http://127.0.0.1:8317/v1/responses` — WebSocket streaming
- `GET http://127.0.0.1:8317/v1/models` — model listing

### Verify

```bash
codex "list files in current directory"
```

---

## Cursor

Cursor uses OpenAI-compatible endpoints.

### Setup

1. Open Cursor Settings → Models
2. Set **OpenAI API Base** to: `http://127.0.0.1:8317/v1`
3. Set **OpenAI API Key** to: `your-api-key`
4. Available models will auto-populate from `/v1/models`

### How it works

Cursor sends requests to:
- `POST http://127.0.0.1:8317/v1/chat/completions`
- `GET http://127.0.0.1:8317/v1/models`

---

## Windsurf (Codeium)

### Setup

1. Open Windsurf Settings → AI Provider
2. Select **OpenAI Compatible** as provider
3. Set **Base URL** to: `http://127.0.0.1:8317/v1`
4. Set **API Key** to: `your-api-key`
5. Select a model from the list

---

## Continue.dev (VS Code / JetBrains)

### Setup

Edit `~/.continue/config.json`:

```json
{
  "models": [
    {
      "title": "CLIProxyAPI-lite (OpenAI)",
      "provider": "openai",
      "model": "gpt-4o",
      "apiBase": "http://127.0.0.1:8317/v1",
      "apiKey": "your-api-key"
    },
    {
      "title": "CLIProxyAPI-lite (Anthropic)",
      "provider": "anthropic",
      "model": "claude-sonnet-4-20250514",
      "apiBase": "http://127.0.0.1:8317",
      "apiKey": "your-api-key"
    }
  ]
}
```

---

## Cline (VS Code)

### Setup

1. Open Cline settings in VS Code
2. Set **API Provider** to **OpenAI Compatible**
3. Set **Base URL** to: `http://127.0.0.1:8317/v1`
4. Set **API Key** to: `your-api-key`
5. Set **Model ID** to any available model (e.g. `gpt-4o` or `claude-sonnet-4-20250514`)

---

## aider

### Setup

```bash
export OPENAI_API_BASE=http://127.0.0.1:8317/v1
export OPENAI_API_KEY=your-api-key
aider --model gpt-4o
```

Or for Anthropic models:
```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:8317
export ANTHROPIC_API_KEY=your-api-key
aider --model claude-sonnet-4-20250514
```

---

## Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:8317/v1",
    api_key="your-api-key",
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(response.choices[0].message.content)
```

---

## Python (Anthropic SDK)

```python
import anthropic

client = anthropic.Anthropic(
    base_url="http://127.0.0.1:8317",
    api_key="your-api-key",
)

message = client.messages.create(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}],
)
print(message.content[0].text)
```

---

## Node.js / TypeScript (OpenAI SDK)

```typescript
import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "http://127.0.0.1:8317/v1",
  apiKey: "your-api-key",
});

const response = await client.chat.completions.create({
  model: "gpt-4o",
  messages: [{ role: "user", content: "Hello!" }],
});
console.log(response.choices[0].message.content);
```

---

## cURL

### OpenAI-compatible

```bash
curl http://127.0.0.1:8317/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Anthropic-compatible

```bash
curl http://127.0.0.1:8317/v1/messages \
  -H "x-api-key: your-api-key" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

## Authentication methods

The server accepts API keys from any of these sources (checked in order):

| Method | Example |
|---|---|
| `Authorization: Bearer <key>` | Most OpenAI-compatible tools |
| `x-api-key: <key>` | Anthropic SDK, Claude Code |
| `X-Goog-Api-Key: <key>` | Gemini-compatible tools |
| `?key=<key>` query parameter | Browser/quick tests |

All methods use the same key pool from `config.yaml` → `api-keys`.

If `api-keys` is empty or not set, authentication is disabled and all requests are accepted.

---

## Endpoint summary

| Protocol | Base URL | Endpoint | Used by |
|---|---|---|---|
| OpenAI Chat Completions | `http://127.0.0.1:8317/v1` | `POST /v1/chat/completions` | Cursor, Windsurf, Continue, Cline, aider, any OpenAI-compatible tool |
| OpenAI Responses | `http://127.0.0.1:8317/v1` | `POST /v1/responses` | Codex CLI |
| OpenAI Models | `http://127.0.0.1:8317/v1` | `GET /v1/models` | All tools (model discovery) |
| Anthropic Messages | `http://127.0.0.1:8317` | `POST /v1/messages` | Claude Code CLI, Claude VS Code extension |
| Anthropic Count Tokens | `http://127.0.0.1:8317` | `POST /v1/messages/count_tokens` | Claude Code CLI |

> **Note:** Anthropic SDK uses `ANTHROPIC_BASE_URL` without `/v1` suffix (the SDK appends it).
> OpenAI SDK uses `OPENAI_BASE_URL` with `/v1` suffix (the SDK does NOT append it).

---

## Troubleshooting

### "Connection refused"
- Is the server running? → `go run ./cmd/server --config config.yaml`
- Is it listening on the right port? → Check `port` in `config.yaml` (default: 8317)

### "401 Unauthorized"
- Is your API key in `config.yaml` under `api-keys`?
- Does the key in your env var exactly match one of the configured keys?

### "No models available"
- Have you logged in to at least one provider account? → Open `http://127.0.0.1:8317/` in your browser
- Check account status via management API: `curl http://127.0.0.1:8317/v0/management/auth-files -H "Authorization: Bearer <management-key>"`

### Models are not showing up in the IDE
- Some IDEs cache the model list. Try reloading or restarting the IDE.
- Verify models are available: `curl http://127.0.0.1:8317/v1/models -H "Authorization: Bearer your-api-key"`
