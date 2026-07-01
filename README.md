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

- **gateway** (`backend/cmd/gateway`) — owns client WebSocket connections, parses inbound actions (`TRAVEL`, `RP_CHAT`, `CAST_VOTE`, `SWAP_CHARACTER`, `INVITE_TO_PARTY`, `START_ENCOUNTER`, `COMBAT_ACTION`, `SKILL_CHECK`, ...), reads/writes Postgres for anything that needs to happen inline, and offloads the two heavy/async operations to RabbitMQ instead of blocking the socket.
- **worker** (`backend/cmd/worker`) — consumes `dungeon_generation_queue` (procedural room generation) and `vote_resolution_queue` (batched Honor writes once a party vote's 30-second window closes), then publishes the result back over Redis pub/sub so the gateway can relay it to the right client(s).
- **Redis** — transient, real-time state only: live chat fan-out between gateway replicas, in-flight vote tallies. Nothing here is durable (player location presence is gateway in-memory, not Redis-cached).
- **Postgres** — the durable system of record: accounts, characters, an append-only honor audit log, parties, and dungeon instances (including room-cleared progress, which persists across gateway restarts so in-progress dungeon runs survive).
- **frontend** (`frontend/`) — React + TypeScript + Tailwind. A thin client: it renders whatever state the gateway pushes over the WebSocket and sends user intents back as typed messages. See `frontend/src/ws/protocol.ts` and `backend/internal/wsproto/protocol.go` — **keep these two in sync by hand**, there's no shared codegen. Navigation is location-graph based (hub-and-spoke map, not a tile grid); dungeons are room-card encounter tracks with real turn-based combat (`CombatView.tsx`); a party panel, skill-check buttons, and the GM narrator channel round out the VTT feel. See the "Implementation Notes" at the bottom of `outline.md` for full design history.

### Why this split

The gateway is the only thing that needs to be fast and always-on for every connected player. Dungeon generation and vote resolution are bursty and involve heavier Postgres writes, so they're pushed onto a queue a worker can consume independently — the gateway never blocks a WebSocket request on either. Redis exists purely so a second gateway replica could be added later without re-architecting chat/vote fan-out (the prototype runs a single replica, but `internal/chat.Hub` + Redis pub/sub already assume there could be more than one).

### Combat, narrator, and the VTT model

Combat is turn-based: real initiative order (d20 + DEX modifier), each party member acts on their own turn (Attack / Dodge / Flee), monster turns are auto-resolved by the server (the DM is automated, not a human player). Multiple party members fight the same monsters together with shared HP tracking. A 60-second per-turn timeout auto-submits a basic attack to prevent AFK players from freezing the table.

The `backend/internal/narrator` package (template-based, not an LLM call — swappable interface for future LLM integration, see the package doc) generates GM-voiced flavor text for arrival at locations, dungeon entry, every attack, room outcomes, skill checks, and NPC/party choice resolutions. All narration is also logged persistently to a `/gm` chat channel. Six 5e skills are modeled for non-combat interactions (Perception, Investigation, Insight, Stealth, Arcana, Athletics) — a successful pre-combat check removes the room's weakest monster from the upcoming fight.

Players join each other's in-progress dungeons by forming a party (invite by display name, accept on the PARTY panel) then traveling to the Mine Entrance and entering — the "hot-drop" path (`handleEnterDungeon`) adds them to the run's present-accounts roster and delivers current dungeon/combat state immediately.

See `outline.md`'s "Implementation Notes" for the full design history and all documented simplifications.

### Known simplifications

- Single gateway/worker replica in `docker-compose.yml` (the Redis pub/sub fan-out exists so adding replicas later doesn't require a rewrite, but it isn't load-tested here).
- `/guild` and `/rp` chat reach every locally-connected client rather than being filtered by location — `/party` is properly scoped.
- No inventory/spellbook persistence per character yet.
- Combat is one attack per turn (no multiattack, no spell slot tracking — Wizards roll a different damage die instead of managing slots), no conditions or saving throws. A losing encounter heals the defeated party on retreat instead of implementing death saves. Boss monsters are tuned below their real SRD CR because party size is variable and a true CR 2 monster is unwinnable solo.
- Mid-fight turn state (`*combat.Encounter`) lives in gateway memory only — a gateway restart mid-fight loses the current turn order, but room-cleared progress (and dungeon existence) is persisted to Postgres so a restarted party can re-enter and replay that specific room.

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
  internal/combat/       d20 attack rolls, SRD damage dice, encounter resolution
  internal/narrator/     Game-Master-voiced flavor text generation
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
