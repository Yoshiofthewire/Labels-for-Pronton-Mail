# Lumo Lab: Proton Mail Auto-Labeler

Lumo Lab is a Dockerized application that scans unread Proton Mail inbox messages, classifies them with Lumo API V2, and applies Proton labels using deterministic rules.

Classification prompts are composed in this order:

1. `GARDRAIL.md`
2. `TUNING.md`
3. Message-specific classification prompt (sender, subject, body, allowlist)

This repository currently includes:

- Go backend daemon + API server
- React web interface
- Single-container runtime with `supervisord`
- Local Lumo API V2 process support inside the same container

## Overview

The app runs as a long-lived daemon and polls Proton Mail every 5 minutes by default.

For each eligible message:

1. Fetch unread Inbox messages only
2. Convert body to plain text and redact configured sensitive patterns
3. Send classification prompt to Lumo
4. Parse Lumo output using configured Proton allowlist
5. Apply matching label to Proton message
6. Persist state/checkpoint to avoid re-processing

## Architecture

- Backend (`backend/`)
: Go service with daemon and HTTP API modes
- Frontend (`frontend/`)
: React + Vite web UI
- Runtime
: Docker container with three supervised processes:
  - API server (`lumo-lab --mode server`)
  - Poll daemon (`lumo-lab --mode daemon`)
  - Local Lumo API V2 (`node lumo.js`)

## Key Features Implemented

- Local auth with session cookie middleware
- First-run admin bootstrap (`admin.env`)
- Config/state persisted in mounted volumes
- 30-day rolling cleanup for processed IDs and decision history
- Console + rotating log files (16 MB, keep 8)
- Health endpoint and repair endpoint
- Automatic unhealthy restart escalation support
- Lumo settings configurable through web config API/UI
- Lumo connectivity test endpoint + UI test button
- Editable `TUNING.md` from the web UI
- Proton label discovery exposed to the UI for tuning order/definitions

## Project Structure

- `backend/cmd/main.go`
- `backend/internal/app/`
- `backend/internal/api/`
- `backend/internal/adapters/proton/`
- `backend/internal/adapters/lumo/`
- `backend/internal/processor/`
- `backend/internal/redaction/`
- `backend/internal/state/`
- `frontend/src/`
- `Dockerfile`
- `docker-compose.yml`
- `supervisord.conf`
- `scripts/bootstrap.sh`
- `scripts/start-lumo.sh`

## Requirements

- Docker + Docker Compose

Optional for local non-Docker development:

- Go 1.22+
- Node.js 20+
- npm

## Quick Start (Docker)

1. Create environment file:

```bash
cp .env.example .env
```

2. Edit `.env` values:

- `PROTON_SECRET_KEY`
- `PROTON_USERNAME` / `PROTON_PASSWORD` (or token-based Proton vars)
- `LUMO_API_KEY` if your Lumo route requires it
- `LUMO_BASE_URL` (defaults to local in-container Lumo)

3. Build and run:

```bash
docker compose up --build -d
```

4. Open the app:

- Web UI: http://localhost:5866

## Volumes and Persistence

The container persists data via named volumes mapped to:

- `/lumo_lab/config`
- `/lumo_lab/logs`
- `/lumo_lab/state`

Important files:

- `/lumo_lab/config/config.yaml`
- `/lumo_lab/config/admin.env`
- `/lumo_lab/config/lumo-auth.json` (for local Lumo API V2 session)
- `/lumo_lab/config/TUNING.md` (runtime tuning instructions; created/updated by API)
- `/lumo_lab/state/state.json`
- `/lumo_lab/state/decisions.json`

## Local Lumo API V2 in Container

The image installs `carlostkd/Lumo-Api-V2` and runs it under `supervisord`.

### Auth bootstrap for Lumo API V2

Lumo API V2 requires an auth session file (`auth.json`) produced by its login flow.

For this project, place that file at:

- `/lumo_lab/config/lumo-auth.json`

`start-lumo.sh` copies it into the Lumo runtime directory before launching `node lumo.js`.

If missing, the Lumo process logs a warning and idles instead of crashing the whole container.

### Disable local Lumo process

Set:

```env
LUMO_LOCAL_ENABLED=false
```

Then point `LUMO_BASE_URL` to an external Lumo service.

## Web Configuration

`/api/config` (and the Config page) supports:

- `lumo.baseUrl`
- `lumo.apiKey`
- `lumo.classifyPath`
- `labels.allowlist`
- timezone/log/scan/rate-limit settings

The Config page also supports:

- Loading and editing `TUNING.md`
- Syncing labels discovered from Proton (`GET /api/labels`)
- Reordering labels for priority guidance
- Managing per-label definition notes
- Generating/resetting tuning templates

## Tuning and Guardrails

The backend prepends `GARDRAIL.md` and `TUNING.md` before each classify request.

Default tuning file resolution order:

1. `TUNING_FILE` environment path (if set)
2. `/lumo_lab/config/TUNING.md`
3. `TUNING.md` (repo root)
4. `/opt/lumo-lab/TUNING.md`

You can read and update runtime tuning through:

- `GET /api/tuning`
- `PUT /api/tuning`

## Lumo Test Button

The Config page includes **Run Lumo Test**.

It calls backend endpoint:

- `POST /api/lumo/test`

Payload:

```json
{ "prompt": "Return only the label Questionable" }
```

Response includes connection target and returned text.

## Authentication and Security

- Local username/password login (`/api/auth/login`)
- Session cookie (`lumo_session`) with sliding expiry
- Auth middleware protects operational routes
- `/api/health`, `/api/auth/login`, `/api/auth/me`, `/api/setup` are public by design
- API key and sensitive values can be configured in YAML and/or environment overrides

## API Endpoints (Current)

Public:

- `GET /api/health`
- `POST /api/auth/login`
- `GET /api/auth/me`
- `GET /api/setup`

Protected (session required):

- `POST /api/auth/logout`
- `POST /api/auth/password`
- `GET /api/status`
- `GET|PUT /api/config`
- `GET /api/labels`
- `GET|PUT /api/tuning`
- `GET /api/decisions`
- `GET /api/logs`
- `POST /api/lumo/test`
- `POST /api/health/repair`

`GET /api/labels` returns both configured labels and Proton-discovered labels:

```json
{
  "configured": ["Questionable", "Primary"],
  "proton": ["Important", "Primary", "Promotions", "Social", "Updates"]
}
```

## Proton Configuration

Supported credential modes:

1. Username/password mode:
- `PROTON_USERNAME`
- `PROTON_PASSWORD`
- optional `PROTON_TOTP`

2. Token mode:
- `PROTON_UID`
- `PROTON_ACCESS_TOKEN`
- `PROTON_REFRESH_TOKEN`

## Development Commands

Backend:

```bash
cd backend
go mod tidy
go build ./...
```

Frontend:

```bash
cd frontend
npm install
npm run build
npm run dev
```

## Logging

- Console logs enabled
- Rotating file logs at `/lumo_lab/logs/app.log`
- Rotation policy: 16 MB, up to 8 rotated files

## Known Limitations

- Local session-based auth is in-memory and not yet distributed/multi-instance aware
- Lumo API V2 itself requires manual session bootstrap (`auth.json`) from upstream flow
- Frontend still has placeholder pages for some advanced operational views
- Tuning editor parsing is best-effort for common markdown patterns; unusual custom formats may need manual adjustment

## Troubleshooting

### Lumo test fails

- Verify `lumo.baseUrl` and `lumo.classifyPath`
- Check `/api/logs?lines=200`
- Confirm `/lumo_lab/config/lumo-auth.json` exists for local Lumo mode
- Confirm local Lumo process logs in `/lumo_lab/logs/lumo.log`

### Unauthorized API responses

- Login first via web UI
- Check `/api/auth/me`

### No labels applied

- Ensure `labels.allowlist` is not empty
- Verify Proton credentials and unread inbox state
- Check daemon logs and decisions endpoint

## License

This repository is licensed under AGPL V3.0.
Respect upstream licenses for dependencies. 
The Lumo API V2 is licensed under AGPL v3.0
The Proton API SDK is licensed by Proton AG under The MIT License.
