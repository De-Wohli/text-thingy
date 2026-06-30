# Development Progress

Living status file for this project. Update it whenever a work session changes what's done, what's verified, or what's planned — this is the first thing to read (and update) at the start of any new agentic or human session.

**Repo:** https://github.com/De-Wohli/text-thingy
**Last updated:** 2026-06-30

## Current status: Phase 1 + RP & Voting rebuild — initial implementation complete, unreleased

The full stack described in `outline.md` (Go gateway/worker, RabbitMQ, Redis, Postgres, React/Tailwind frontend) has been implemented end-to-end and pushed to `main`. Nothing has been deployed anywhere live yet — this has only run via local `go build`/`go test`/`npm run build` and has **not** been exercised against real Postgres/Redis/RabbitMQ instances (e.g. via `docker compose up`).

### What's verified
- Backend: `go build ./...`, `go vet ./...`, `go test ./...` all pass (verified locally with a temporarily-downloaded Go 1.23 toolchain — Go is not installed system-wide in this dev environment as of this writing).
- Frontend: `npm run lint`, `npm run typecheck`, `npm run test` (9 tests), `npm run build` all pass.
- `go.sum` is committed and `go mod verify` passes.

### What's NOT verified yet
- The full stack has never actually been run together (`docker compose up --build`). Docker is not installed in this dev environment — see "Blocked" below.
- No manual/browser testing of the UI against a live gateway.
- GitHub Actions (`ci.yml`, `docker-publish.yml`) have not been confirmed green on the actual repo — check https://github.com/De-Wohli/text-thingy/actions after each push.
- RabbitMQ/Redis pub/sub wiring (dungeon generation round-trip, vote resolution round-trip) is implemented per the design but only reasoned through, not exercised against running brokers.

### Blocked / needs user action
- **Docker CLI is installed but the daemon isn't running** (`systemctl is-active docker` → `inactive`), and the current user isn't in the `docker` group yet. `sudo` requires an interactive password prompt this agent can't supply. Run in a real terminal:
  ```
  sudo systemctl enable --now docker
  sudo usermod -aG docker $USER   # then log out/in or `newgrp docker`
  ```
  Once that's done, `docker compose up --build` from the repo root will actually exercise the full stack for the first time.
- **`git push` has no credentials in this agent's shell** — `PROGRESS.md`'s own addition is committed locally on `main` but unpushed as of this writing. Run `git push` from a terminal with your GitHub auth available.
- **`gh` was installed but isn't on this shell's PATH** — a fresh terminal may pick it up; if not, check how it was installed (e.g. needs `~/.local/bin` or similar on `$PATH`). Run `gh auth login` once it resolves so CI status/PR work can be checked from the CLI going forward.

## Architecture snapshot

See `README.md` for the full breakdown. Quick summary:
- `backend/cmd/gateway` — Fiber WebSocket/REST gateway (connections, MOVE, roster, chat, choice/voting).
- `backend/cmd/worker` — RabbitMQ consumer (dungeon generation, vote-window honor resolution).
- `backend/internal/*` — domain packages (`honor`, `dungeon`, `voting`, `models`) + Postgres/Redis/RabbitMQ adapters.
- `frontend/` — React + TypeScript + Tailwind, WebSocket-driven client (`frontend/src/ws/`).
- `docker-compose.yml` — local dev stack (postgres, redis, rabbitmq, gateway, worker, frontend).
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

## Next steps (suggested, not started)
1. User installs Docker, runs `docker compose up --build`, confirms the stack actually comes up and the frontend can create an account / move around.
2. Confirm GitHub Actions are green on the pushed commit.
3. Decide on a party-formation flow (invite/accept) so `/party` chat and party voting can actually be exercised with 2+ players.
4. Decide whether to pursue inventory/spellbook persistence (deferred Phase 1 scope).
5. If/when a real hosting target is chosen, extend `docker-publish.yml` with an actual deploy step.
