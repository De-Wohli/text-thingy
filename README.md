# 5e Web MMO — Prototype

A tabletop-styled prototype of a persistent 5e-compatible web MMO: an account-wide character roster, an Honor/Alignment system that reshapes how the world reacts to you, a Text RP chat engine, a party Choice/Voting system, and procedurally generated dungeons. See [outline.md](outline.md) for the original design brief, including a note on why the overworld renders as a visual tile board with buttons rather than literal ASCII text.

## Architecture

```
[ Frontend: React / Tailwind ]
         │
         ▼  (WebSocket /ws, REST /api)
[ Gateway: Go (Fiber) ] ──(Pub/Sub)── [ Redis: live chat, votes, coordinates ]
         │
         ▼  (publish jobs)
   [ RabbitMQ ]
         │
         ▼  (consume jobs)
[ Worker (Go) ] ────► [ Postgres: accounts, characters, honor log, dungeons ]
   (dungeon generation, vote-window honor writes)
```

- **gateway** (`backend/cmd/gateway`) — owns client WebSocket connections, parses inbound actions (`MOVE`, `RP_CHAT`, `CAST_VOTE`, `SWAP_CHARACTER`, ...), reads/writes Postgres for anything that needs to happen inline (movement, character CRUD, chat routing), and offloads the two heavy/async operations to RabbitMQ instead of blocking the socket.
- **worker** (`backend/cmd/worker`) — consumes `dungeon_generation_queue` (procedural 15×15 grid generation) and `vote_resolution_queue` (batched Honor writes once a party vote's 30-second window closes), then publishes the result back over Redis pub/sub so the gateway can relay it to the right client(s).
- **Redis** — transient, real-time state only: live chat fan-out between gateway replicas, in-flight vote tallies, and player coordinates. Nothing here is durable.
- **Postgres** — the durable system of record: accounts, characters, an append-only honor audit log, and dungeon instances.
- **frontend** (`frontend/`) — React + TypeScript + Tailwind. A thin client: it renders whatever state the gateway pushes over the WebSocket and sends user intents back as typed messages. See `frontend/src/ws/protocol.ts` and `backend/internal/wsproto/protocol.go` — **keep these two in sync by hand**, there's no shared codegen. The overworld is a visual tile board (CSS grid with colored/iconed cells, not ASCII text) navigated with on-screen direction buttons or keyboard; a "Here" panel surfaces only the actions actually available at the player's position (Talk to Citizen, Enter the Guild Hall, etc.) instead of disabled buttons the player has to discover; dungeons render as a room-card encounter track rather than a literal grid. See the "Implementation Note" at the bottom of `outline.md`.

### Why this split

The gateway is the only thing that needs to be fast and always-on for every connected player. Dungeon generation and vote resolution are bursty and involve heavier Postgres writes, so they're pushed onto a queue a worker can consume independently — the gateway never blocks a WebSocket request on either. Redis exists purely so a second gateway replica could be added later without re-architecting chat/vote fan-out (the prototype runs a single replica, but `internal/chat.Hub` + Redis pub/sub already assume there could be more than one).

### Known simplifications (Phase 1 prototype)

- Single gateway/worker replica in `docker-compose.yml` (the Redis pub/sub fan-out exists so adding replicas later doesn't require a rewrite, but it isn't load-tested here).
- `/guild` and `/rp` chat reach every locally-connected client rather than being filtered by zone/proximity — `/party` is properly scoped by `PartyID`.
- No inventory/spellbook persistence per character yet (the original Phase 1 spec calls for it; this rebuild prioritized the RP/voting/dungeon-async pipeline). `ability_*` columns and `honor_log` are there; inventory is a natural next migration.
- There's no party-formation flow yet — `account.partyId` exists in the schema and is honored everywhere (chat scoping, vote rooms, dungeon instances), but nothing currently sets it, so every account is effectively solo. Wiring up an invite/accept flow is the natural next step before party voting/dungeons can be exercised with more than one player.

## Local development

Requires Docker and Docker Compose.

```bash
cp .env.example .env   # defaults work as-is; edit if you want different credentials
docker compose up --build
```

This starts Postgres, Redis, RabbitMQ, the gateway (`:8080`), the worker, and the frontend (`:5173`, nginx-served, proxying `/ws` and `/api` to the gateway). The gateway runs Postgres migrations automatically on startup (`backend/internal/store/migrations/`), guarded by a Postgres advisory lock so the gateway and worker starting concurrently can't double-apply a migration.

RabbitMQ's management UI is at `:15672` (guest/guest) if you want to watch `dungeon_generation_queue` / `vote_resolution_queue` directly.

### Frontend only, against a local Go backend

```bash
cd backend && go run ./cmd/gateway   # needs Postgres/Redis/RabbitMQ reachable, see internal/config
cd backend && go run ./cmd/worker
cd frontend && npm install && npm run dev
```

`frontend/vite.config.ts` proxies `/ws` and `/api` to `localhost:8080` in dev mode.

### Running tests

```bash
cd frontend && npm run lint && npm run typecheck && npm run test
cd backend && go vet ./... && go test ./...
```

## CI/CD

- **`.github/workflows/ci.yml`** — on every push/PR: lints, type-checks, tests, and builds both the frontend (`frontend/`) and backend (`backend/`) independently.
- **`.github/workflows/docker-publish.yml`** — on push to `main`: builds and pushes three images to GitHub Container Registry — `ghcr.io/<owner>/<repo>/gateway`, `.../worker`, and `.../frontend` — each tagged `latest` and with the short commit SHA.

This repo doesn't deploy those images anywhere automatically — there's no live server target wired up. To run a published build:

```bash
docker pull ghcr.io/<owner>/<repo>/gateway:latest
docker pull ghcr.io/<owner>/<repo>/worker:latest
docker pull ghcr.io/<owner>/<repo>/frontend:latest
# then run them the same way docker-compose.yml does, against your own
# Postgres/Redis/RabbitMQ instances (or point `image:` at these instead of
# `build:` in a copy of docker-compose.yml and run it on your server).
```

Wiring up an actual auto-deploy step (Fly.io, Railway, a VPS over SSH, etc.) is a small addition to `docker-publish.yml` once you have a target — ask for it and bring the relevant secrets.

## Repository layout

```
backend/
  cmd/gateway/         WebSocket+REST entrypoint
  cmd/worker/           RabbitMQ consumer entrypoint
  internal/models/       shared domain types
  internal/honor/        Honor & Alignment Matrix
  internal/dungeon/      procedural dungeon generation + CR budget
  internal/voting/       solo/party Choice & Voting engine
  internal/chat/         in-process WebSocket connection hub
  internal/wsproto/      WebSocket JSON protocol (mirror of frontend/src/ws/protocol.ts)
  internal/redisstate/   Redis pub/sub + transient state
  internal/queue/        RabbitMQ publisher/consumer
  internal/store/        Postgres access + migrations
  internal/worldmap/     server-authoritative overworld layout (mirror of frontend/src/data/overworld.ts)
frontend/
  src/engine/             race/class static data + display-only honor bands
  src/data/overworld.ts    client-side overworld layout/rendering
  src/ws/                  WebSocket client + protocol types
  src/state/                React state synced from the gateway
  src/components/            UI
outline.md                original design brief
docker-compose.yml          local dev stack
```
