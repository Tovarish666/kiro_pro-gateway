# kiro_pro-gateway

Anthropic Messages API-compatible proxy that routes requests through a Kiro PRO account.
Designed to connect AI assistants (OpenClaw, etc.) to Claude Opus 4.7 / Sonnet 4.6 via a Kiro PRO subscription.

## Architecture

```
OpenClaw / any Anthropic client
        │  HTTP POST /v1/messages
        ▼
  kiro-proxy  (:8080)
        │  subprocess (v1) → kiro-cli chat --no-interactive
        │  native API (v2, WIP) → runtime.us-east-1.kiro.dev
        ▼
  Kiro PRO backend → Claude Opus 4.7 / Sonnet 4.6
```

## Backends

| Version | How | Status |
|---------|-----|--------|
| v1 (current) | Wraps `kiro-cli` as subprocess | ✅ Working |
| v2 (planned) | Native HTTPS to `runtime.us-east-1.kiro.dev` | 🚧 WIP |

## Quick start

### Requirements
- [kiro-cli](https://kiro.dev) installed
- `ksk_...` API key from Kiro PRO account

### Linux

```bash
export KIRO_API_KEY=ksk_...
export KIRO_CLI_PATH=/usr/local/bin/kiro-cli
./bin/kiro-proxy-linux-amd64
```

Or with systemd — edit `deploy/kiro-proxy.service` with your key, then:
```bash
cp deploy/kiro-proxy.service /etc/systemd/system/
systemctl enable --now kiro-proxy
```

### Windows

Edit `scripts/start-windows.cmd` with your key, run it or add to Task Scheduler.

## Build

```bash
make build-linux    # → bin/kiro-proxy-linux-amd64
make build-windows  # → bin/kiro-proxy.exe
make build-all
```

## API

| Endpoint | Description |
|----------|-------------|
| `POST /v1/messages` | Anthropic Messages API (streaming + non-streaming) |
| `GET /v1/models` | Model list |
| `GET /healthz` | Health check |

Auth: `Authorization: Bearer ksk_...` or `x-api-key: ksk_...`

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `KIRO_API_KEY` | — | Kiro PRO API key (`ksk_...`) |
| `KIRO_CLI_PATH` | `/usr/local/bin/kiro-cli` | Path to kiro-cli binary |
| `LISTEN_ADDR` | `:8080` | Listen address |

## OpenClaw integration

Copy `configs/openclaw.json.example` → `~/.openclaw/openclaw.json`, fill in your values, then:
```bash
openclaw gateway --port 18789 --allow-unconfigured --bind lan
```

## Roadmap

- [x] v1: kiro-cli subprocess proxy
- [x] SSE streaming support
- [x] system-as-array support (OpenClaw compatibility)
- [ ] v2: native Kiro API auth (reverse-engineer ksk_ JWT flow)
- [ ] OpenRouter fallback provider
- [ ] Debian/systemd deployment on Proxmox
