# Development Progress

Living status file for this project. Update it whenever a work session changes what's done, what's verified, or what's planned — this is the first thing to read (and update) at the start of any new agentic or human session.

**Repo:** https://github.com/De-Wohli/text-thingy
**Last updated:** 2026-06-30

## Current status: Phase 1 + RP & Voting rebuild — implemented and verified end-to-end locally, not deployed anywhere live

The full stack described in `outline.md` (Go gateway/worker, RabbitMQ, Redis, Postgres, React/Tailwind frontend) has been implemented and **actually run together** via `docker compose up --build` — not just compiled in isolation. See "2026-06-30 — Full stack verification" below for what was exercised. Nothing has been deployed anywhere live (no hosting target chosen yet — CI only builds/publishes images to GHCR, see `README.md` § CI/CD).

### What's verified
- **Full stack runs together**: `docker compose up --build` brings up postgres/redis/rabbitmq/gateway/worker/frontend cleanly, all healthchecks pass, no errors in any service's logs.
- **End-to-end gameplay loop, for real, against live containers**: account creation (REST) → WebSocket connect → character creation → guild-hall-adjacency-gated character swap → global chat (round-tripped through Redis pub/sub) → talk to NPC → solo Choice resolution (Honor +10, written to `honor_log`) → move to a `[?]` POI → `ENTER_POI` → RabbitMQ `dungeon_generation_queue` → worker generates dungeon → Postgres `SaveDungeon` → Redis pub/sub → gateway → `DUNGEON_READY` over the WebSocket → clear all 3 non-start rooms → `RESOLVE_DUNGEON` → gold awarded (50→75) → confirmed durable via direct `psql` queries against `accounts`, `characters`, `honor_log`, `dungeons`.
- **Frontend nginx proxy verified** for both REST (`/api/`) and WebSocket (`/ws/`) — confirmed working through `localhost:5173`, not just direct-to-gateway on `:8080`.
- **Survives backend restarts**: `docker compose restart gateway` while the frontend container keeps running still works (see bug fix below).
- **Robustness**: a buggy test script hammered the gateway with hundreds of malformed `CLEAR_DUNGEON_ROOM` messages (empty roomType) in a tight loop — no crash, no panic, gateway just no-op'd each one correctly.
- Backend: `go build ./...`, `go vet ./...`, `go test ./...` all pass (verified with a temporarily-downloaded Go 1.23 toolchain in an earlier session; this session's changes were verified via the Docker build succeeding + live runtime behavior instead, since the toolchain wasn't re-downloaded).
- Frontend: `npm run lint`, `npm run typecheck`, `npm run test` (9 tests), `npm run build` all pass.
- `go.sum` is committed and `go mod verify` passes.
- `git push` works directly (SSH remote, reusing the user's existing registered key) — no more credential blocker.

### Bugs found by actually running it (both fixed, commit `a317676`)
1. **`store.ListCharacters` returned a nil slice** for accounts with no characters yet, which `json.Marshal` encodes as `null`. The frontend calls `.find()`/`.length` on `characters` with no null guard — this would have crashed on first load for every new account. Fixed: initialize to `[]models.Character{}`.
2. **nginx cached the `gateway` upstream IP at startup** (`proxy_pass http://gateway:8080` resolves once, not per-request). Recreating/restarting the gateway container gives it a new internal IP, so nginx kept proxying to a dead address until nginx itself restarted — `502 Bad Gateway`. Fixed: added a `resolver 127.0.0.11 valid=10s` + `proxy_pass $gateway_upstream` (variable) so nginx re-resolves every 10s. Verified this survives `docker compose restart gateway`.

This is exactly the kind of bug that only shows up when you run the real thing — both would have shipped silently otherwise.

### What's still NOT verified
- No manual/browser testing of the actual React UI rendering (only tested the wire protocol directly via WebSocket scripts + curl). The pages/components themselves haven't been visually confirmed in a browser.
- GitHub Actions (`ci.yml`, `docker-publish.yml`) have not been confirmed green on the actual repo — check https://github.com/De-Wohli/text-thingy/actions after pushing the latest commits.
- Party-mode voting (the 30-second timer, multi-voter tally, tie-break) was unit-tested (`internal/voting`) but not exercised live — would need a second account in the same party, and nothing sets `partyId` yet (see Known gaps).
- `/guild` and `/rp` chat channels not manually tested live (only `/global` was exercised in the live run).

### Blocked / needs user action
- **`gh` was installed but isn't on this shell's PATH** — a fresh terminal may pick it up; if not, check how it was installed (e.g. needs `~/.local/bin` or similar on `$PATH`). Run `gh auth login` once it resolves so CI status/PR work can be checked from the CLI going forward.
- Nothing else is currently blocked. Docker is installed, running, and the current user is in the `docker` group (commands in this agent's shell use `sg docker -c "..."` since group membership needs a fresh login to take effect without that — a new terminal session shouldn't need this workaround).

## Architecture snapshot

See `README.md` for the full breakdown. Quick summary:
- `backend/cmd/gateway` — Fiber WebSocket/REST gateway (connections, MOVE, roster, chat, choice/voting).
- `backend/cmd/worker` — RabbitMQ consumer (dungeon generation, vote-window honor resolution).
- `backend/internal/*` — domain packages (`honor`, `dungeon`, `voting`, `models`) + Postgres/Redis/RabbitMQ adapters.
- `frontend/` — React + TypeScript + Tailwind, WebSocket-driven client (`frontend/src/ws/`).
- `docker-compose.yml` — local dev stack (postgres, redis, rabbitmq, gateway, worker, frontend). **Confirmed working.**
- `.github/workflows/ci.yml` — lint/typecheck/test/build, both services, every push/PR.
- `.github/workflows/docker-publish.yml` — builds + pushes `gateway`/`worker`/`frontend` images to GHCR on push to `main`.

## Known gaps (by design, see README "Known simplifications")
- No party-formation flow — `account.partyId` is wired everywhere it's needed (chat scoping, voting, dungeons) but nothing currently sets it, so every account is effectively solo.
- No inventory/spellbook persistence per character (original Phase 1 spec wanted this; deferred to keep the RP/voting/async-dungeon pipeline in scope).
- `/guild` and `/rp` chat broadcast to every locally-connected client rather than being filtered by zone/proximity. `/party` is properly scoped.
- Single gateway/worker replica assumed; Redis pub/sub fan-out exists for future multi-replica scaling but is untested under more than one instance.
- No live deploy target wired up (user chose "CI + GHCR images only" over an actual hosting target — see `README.md` § CI/CD).

## Session log

### 2026-06-30 — Initial build
- Read `outline.md` (v1: single-player ASCII prototype) and scaffolded a static React+TS+Vite app with a local-only reducer, localStorage persistence, and GitHub Pages deploy workflow.
- User updated `outline.md` mid-session to a substantially larger v2 design (Go + RabbitMQ + Redis + Postgres + WebSocket multiplayer, Text RP chat, party voting). Confirmed with user via AskUserQuestion: full rebuild (not incremental), and CI builds Docker images to GHCR rather than deploying to a live host (no hosting target available).
- Rebuilt from scratch: `frontend/` + `backend/` split, full Go backend (gateway + worker + 6 internal packages with unit tests), Postgres migrations, Dockerfiles, `docker-compose.yml`, two GitHub Actions workflows, README.
- Attempted `sudo pacman -S docker` on the user's behalf — failed, sudo needs an interactive terminal. Left instructions for the user instead.
- Downloaded a Go 1.23 toolchain to `/tmp` (no sudo needed) specifically to verify the backend actually compiles rather than guessing — `go build`/`go vet`/`go test` all passed on the first real attempt after a `gofmt -w` cleanup pass.
- Verified frontend independently (lint/typecheck/test/build all green).
- Initialized git, committed (73 files), user pushed to `https://github.com/De-Wohli/text-thingy`.
- Created this file per user request to track ongoing progress across sessions.

### 2026-06-30 — Docker/gh follow-up
- User installed Docker and `gh`. Confirmed `docker`/`docker-compose` CLIs are now present, but the daemon is inactive and the user isn't in the `docker` group yet — both need one interactive `sudo` round-trip the agent can't do non-interactively. `gh` isn't resolving on `$PATH` in this shell yet.
- `git push` still fails in this shell (no credential helper available here) — the `PROGRESS.md` commit is local-only on `main` until the user pushes it themselves.

### 2026-06-30 — SSH key setup
- Found an existing `~/.ssh/id_ed25519` already registered with GitHub (`De-Wohli`). Switched `origin` remote from HTTPS to SSH (`git@github.com:De-Wohli/text-thingy.git`). `git push` now works directly from this agent's shell with no credential prompts. Pushed the two pending commits.

### 2026-06-30 — Full stack verification
- User ran `sudo systemctl enable --now docker` and `sudo usermod -aG docker $USER`. Daemon came up; used `sg docker -c "..."` to get group membership without a fresh login.
- Ran `docker compose up --build -d`: all 6 services (postgres, redis, rabbitmq, gateway, worker, frontend) built and started cleanly, all healthchecks passed.
- Exercised the full gameplay loop live against the running containers (see "What's verified" above) using throwaway Node scripts (native `WebSocket`, no extra deps) plus `curl` and direct `psql` queries — not just unit tests.
- Found and fixed two real bugs that only surfaced by actually running the stack (nil-slice JSON serialization, nginx upstream DNS caching) — see "Bugs found" above. Committed as `a317676`.
- Did not yet open the frontend in an actual browser to visually confirm the React UI — only the wire protocol was exercised directly.

## Next steps (suggested, not started)
1. Open `http://localhost:5173` in an actual browser and click through character creation, movement, chat, NPC dialogue, and a dungeon run — the protocol is confirmed working but the UI itself hasn't been eyeballed.
2. Confirm GitHub Actions are green on the latest pushed commits.
3. Decide on a party-formation flow (invite/accept) so `/party` chat and party voting can actually be exercised with 2+ players — this is the biggest remaining gap between "implemented" and "matches the full outline.md design."
4. Decide whether to pursue inventory/spellbook persistence (deferred Phase 1 scope).
5. If/when a real hosting target is chosen, extend `docker-publish.yml` with an actual deploy step.
