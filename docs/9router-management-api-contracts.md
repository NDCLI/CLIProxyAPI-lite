# 9router Management API Contracts

## Versioning

The authenticated management boundary remains `/v0/management` for compatibility. New normalized resources use `schema_version: 1` in their response envelope. A breaking response change requires a new schema version and a compatibility period; adding optional fields does not.

Existing management endpoints remain available while the new source-owned UI is built. The normalized contracts below become the only API surface used by that UI.

## Common response rules

Successful collection response:

```json
{
  "schema_version": 1,
  "items": [],
  "next_cursor": null
}
```

Successful single-resource response:

```json
{
  "schema_version": 1,
  "item": {}
}
```

Error response:

```json
{
  "error": {
    "code": "stable_machine_code",
    "message": "English fallback message",
    "params": {}
  }
}
```

The frontend translates `error.code` and interpolation parameters. It uses `message` only when the installed locale bundle does not recognize the code.

Mutation rules:

- `POST` creates a resource and returns HTTP 201.
- `PATCH` changes only supplied fields and returns the updated masked resource.
- `DELETE` requires an explicit stable resource ID and returns HTTP 204.
- Conflicting updates return HTTP 409; validation failures return HTTP 400; missing resources return HTTP 404.
- Destructive operations that can invalidate credentials or routing accept a revision or confirmation token to prevent stale UI writes.
- Pagination uses opaque `cursor` and bounded `limit` query parameters. Filters are explicit query parameters, not encoded SQL-like expressions.

## Secret contract

Fields containing API keys, access tokens, refresh tokens, client secrets, cookies, passwords, or credential files are write-only. Read responses contain only:

- a stable resource ID;
- a display label;
- a non-reversible mask such as the last four characters when safe;
- provider/auth type;
- enabled and health state;
- timestamps and non-secret routing metadata.

No response returns a recoverable secret. A create response may include a newly generated endpoint key exactly once; the server must not log it and subsequent reads return only its mask.

## Resource contracts

| Area | Normalized routes | Required contract behavior | Backend authority | Initial state |
| --- | --- | --- | --- | --- |
| Capability manifest | `GET /capabilities` | Returns all top-level feature IDs, `schema_version`, state, and translatable `reason_code` | Compiled Go capability inventory | Implemented |
| Endpoint keys | `GET, POST /endpoint-keys`; `GET, PATCH, DELETE /endpoint-keys/:id` | Stable IDs, one-time generated value, masked reads, active state, usage summary, optimistic revision | `config.yaml` API keys plus usage counters | Partial |
| Providers | `GET, POST /providers`; `GET, PATCH, DELETE /providers/:id`; `POST /providers/:id/test`; `GET /providers/:id/models` | Normalizes API-key and OAuth/file-backed credentials without exposing secret fields; reports health and supported operations | Config provider arrays, auth manager/store, model registry | Partial |
| Combos | `GET, POST /combos`; `GET, PATCH, DELETE /combos/:id`; `POST /combos/:id/validate` | Ordered model targets, kind/capabilities, enabled state, validation errors, routing revision | New Go-owned combo store and request routing | Planned |
| Usage | `GET /usage/records`; `GET /usage/summary`; `GET /usage/stream` | Normalized provider/model/key IDs, status, token categories, latency, time range and cursor filters; stream uses the same record shape | Go usage pipeline and durable extension | Partial |
| Quota | `GET /quota/providers`; `POST /quota/fetch`; `POST /quota/reset` | Uses stable credential IDs, typed capacity windows, observed time, provider error code, and reset support flag | Existing quota providers and cooldown state | Ready |
| Token Saver | `GET, PATCH /token-saver`; `POST /token-saver/headroom/test` | Controls opt-in RTK and Headroom compression, reports live byte savings and Headroom reachability, validates the service URL and timeout, and preserves requests on compression failure | `config.yaml` plus the Go request pipeline and configured Headroom service | Ready for RTK + Headroom |
| CLI tools | `GET /cli-tools`; `GET /cli-tools/:id`; `POST /cli-tools/:id/apply`; `POST /cli-tools/:id/reset` | Status and preview are read-only; apply/reset use allow-listed tools and paths; generated secrets remain masked | Existing safe tool configurator plus new status adapters | Partial |
| Logs | `GET /logs`; `GET /logs/:id`; `DELETE /logs` | Cursor, level/source/request filters, safe text payload, bounded downloads, explicit retention behavior | Existing rotating and request logs | Ready |
| System settings | `GET, PATCH /system-settings` | Allow-listed non-secret fields, validation, revision, and restart-required flags; never exposes management secret | `config.yaml` through existing management writer | Ready |

## Capability states

The capability manifest is the runtime authority for navigation and actions:

- `ready`: normalized backend behavior is safe for the new UI.
- `partial`: underlying behavior exists, but the normalized contract or required operations are incomplete.
- `planned`: a domain and intended contract are defined, but no backend exists.
- `unavailable`: no safe implementation is currently approved; the UI shows an unavailable explanation and no mutation control.

The frontend must not promote `partial`, `planned`, or `unavailable` to a working state based on local storage or build-time flags.

## Contract test gate

Before a resource state becomes `ready`, tests must prove:

1. the exact JSON field names and schema version;
2. authentication through the management middleware;
3. stable IDs and deterministic status codes;
4. secret masking and absence of credential material from responses;
5. validation, not-found, conflict, unavailable, and backend-error responses;
6. persistence through the configured authority and reload where applicable;
7. stream and list records use the same normalized shape where both exist.
