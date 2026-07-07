# Logging Standard

Peapod expects every managed service to write structured JSON logs to stdout/stderr. This keeps the lightweight Dozzle path useful without requiring Loki.

## Required Fields

- `ts`: RFC3339 timestamp.
- `level`: `debug`, `info`, `warn`, `error`, or `fatal`.
- `service`: stable service name, for example `peapod`, `xiezuomao-api`, `9router`.
- `env`: `production`, `staging`, `development`, or another explicit environment name.
- `event`: stable machine-readable event name, for example `http_request`, `task_failed`, `upstream_error`.
- `message`: short human-readable summary.

## Recommended Fields

- `request_id`: request correlation id. Use `X-Request-ID` when provided, otherwise generate one.
- `trace_id`: OpenTelemetry trace id when available.
- `method`, `path`, `route`, `status`, `status_class`, `latency_ms`: HTTP request diagnostics.
- `user_id`, `project_id`, `task_id`, `pipeline`, `repo`, `branch`, `commit`: business correlation fields.
- `error`: non-empty error summary. Do not emit `error: ""` for successful events.

## Level Rules

- `debug`: local diagnostics only. Usually disabled in production.
- `info`: meaningful lifecycle or successful business events. Avoid high-volume health checks.
- `warn`: client errors, degraded fallback, retries, slow requests, near-limit resources.
- `error`: failed business action, upstream failure, 5xx, data loss risk.
- `fatal`: process cannot continue.

## Access Logs

Default mode is `attention`:

- Suppress healthy `/health`, `/healthz`, `/ready`, `/readyz`, `/metrics`.
- Suppress normal fast 2xx/3xx requests.
- Write `warn` for 4xx or slow requests.
- Write `error` for 5xx.

Use these common environment variables:

```env
SERVICE_NAME=my-service
APP_ENV=production
LOG_LEVEL=info
ACCESS_LOG_MODE=attention
ACCESS_LOG_SLOW_THRESHOLD_SECONDS=3
```

`ACCESS_LOG_MODE=all` is useful for temporary debugging. `ACCESS_LOG_MODE=off` disables access logs.

## Security

Do not log API keys, access tokens, refresh tokens, cookies, passwords, private keys, or full Authorization headers. Peapod masks obvious secrets, but services should avoid writing them in the first place.
