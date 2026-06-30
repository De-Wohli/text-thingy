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
- GitHub Actions (`ci.yml`, `docker-publish.yml`) have not been confirmed green on the actual repo — check https://github.com/De-Wohli/text-thingy/actions after pushing the latest commits.
- Party-mode voting (the 30-second timer, multi-voter tally, tie-break) was unit-tested (`internal/voting`) but not exercised live — would need a second account in the same party, and nothing sets `partyId` yet (see Known gaps).
- `/guild` and `/rp` chat channels not manually tested live (only `/global` was exercised in the live run).
- The UI redesign (see below) was screenshot-verified at one viewport size (1280×900) — no responsive/mobile pass.

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
- Combat is simplified 5e (one attack/round, no multiattack/spell-slots/conditions/saving throws); a lost encounter fully heals on retreat instead of implementing death saves; boss monsters are tuned well below their real SRD CR because there's no party to split the action economy with yet (see README "Combat and the Narrator").

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

### 2026-06-30 — Tabletop UI redesign
- User feedback: the overworld read as an ASCII dungeon-crawler, not a tabletop game — wanted a visual map, button-based cardinal-direction navigation, contextual action buttons (Talk to NPC, Enter Guild Hall) instead of disabled buttons the player has to discover, and a less literal dungeon rendering.
- Rewrote `OverworldCanvas.tsx` as a CSS-grid visual board (colored/iconed tiles instead of a `<pre>` of ASCII characters); added `DirectionPad.tsx` (cardinal-direction buttons); added `LocationActions.tsx` (a "Here" panel that only renders the actions actually available at the player's position); rewrote `DungeonView.tsx` as a room-card encounter track (Entrance → Corridor → Treasure Vault → Boss's Den) instead of rendering the 15×15 grid as ASCII; restyled `MapLegend.tsx` to match. Appended an "Implementation Note" to `outline.md` documenting this as a deliberate presentation-layer deviation (the underlying tile-grid/coordinate data model is unchanged).
- **Found and fixed a real data bug while building this**: `frontend/src/data/overworld.ts`'s `RAW_MAP` rows were inconsistent lengths (59–62 chars) — invisible in the old `<pre>` rendering (each row is its own text line) but would have silently shifted every tile after a short row by one column in the new CSS grid (which flattens all rows into one sequential list of grid items). Rebuilt the map at a uniform 61 chars/row and mirrored the exact same string array into `backend/internal/worldmap/worldmap.go` (verified byte-identical with a script diff) so server-side adjacency/walkability checks didn't drift from the client's rendering.
- Verified frontend lint/typecheck/test (9 tests)/build all pass.
- Visually verified in an actual browser: no project skill existed for this, so installed Playwright + a headless Chromium build into the scratchpad (no sudo needed — `npx playwright install chromium`, skipping `--with-deps` which wanted sudo) and drove the running `npm run dev` server. Confirmed via screenshots: the visual board renders correctly, direction-pad clicks move the player token, the "Here" panel correctly shows/hides "Enter the Guild Hall" based on proximity, character creation works end-to-end through the actual UI (not just the wire protocol), and the dungeon room-card track renders with the right monster/cleared state. Zero browser console errors throughout. (One false alarm along the way: an unscoped Playwright `input` selector matched the wrong element on a page with two inputs and silently mis-filled a field — confirmed it was a test-script bug, not an app bug, by re-running with a properly-scoped locator.)
- Rebuilt the `frontend` Docker image so the running `docker compose` stack reflects the new UI. Then realized the map-data fix lives in `backend/internal/worldmap`, not `frontend/` — separately rebuilt `gateway` (worker didn't need it; `cmd/worker` doesn't import `worldmap`, so its build output was byte-identical and Docker correctly skipped recreating it) and confirmed with a fresh WebSocket smoke test that NPC and POI adjacency checks still pass server-side against the corrected map.

### 2026-06-30 — Real 5e combat + GM narrator, "Return to City Gates" bug fix
- User reported the "Return to City Gates" button did nothing, and asked for a Game-Master-style narrator and a virtual-tabletop experience that follows 5e rules closely.
- Root-caused the button bug: the gateway processed `RESOLVE_DUNGEON` correctly (awarded gold, persisted) but only replied with a generic `STATE_SYNC`, which the frontend reducer never used to close the dungeon panel. Fixed by adding a dedicated `DUNGEON_RESOLVED` message the reducer handles explicitly.
- Built real combat (`backend/internal/combat`): d20 + proficiency bonus + ability modifier vs. each monster's SRD Armor Class, SRD damage dice, alternating character/monster attacks until one side falls, character HP persisted to Postgres. "Resolve Encounter" now actually fights instead of flipping a flag.
- Built a Game-Master narrator (`backend/internal/narrator`, template-based, no LLM call) generating flavor text for dungeon entry, every attack, room victory/defeat, the boss reward, and choice/vote resolutions — shown inline and logged to a new `/gm` chat channel.
- **Found three more real bugs by actually running the new combat system live** (not just unit tests):
  1. Defeat clamped character HP to 1 with no way to recover — a planned "heal on retreat" never got implemented, creating a soft-lock after the first loss. Fixed: a lost encounter now fully heals the character on retreat.
  2. The boss monsters used their real SRD CR 2 stats, which are mathematically unwinnable for a solo level-1 character (5e's CR math assumes a 4-person party) — confirmed via a live test where the same boss fight failed ~10 times in a row. Retuned boss stats down to roughly CR 1/2 toughness with a doc comment explaining why.
  3. `pickEncounterForLevel1` (pre-existing logic, not written this session) could pick a single random monster and immediately bail if it didn't fit the XP budget, leaving hallway/treasure rooms with *zero* monsters ~1/3 of the time. Harmless when "clearing a room" was a no-op flag flip; with real combat, a zero-monster encounter produced a nil `[]AttackRoll` slice that serializes to JSON `null`, crashing the frontend's `.map()` over it (same bug class as the `ListCharacters` nil-slice issue from the previous session). Fixed the generator to always include the room's first monster pick from ones that fit, and initialized the slice non-nil.
  4. Also fixed along the way: real characters were using flat ability scores (10 + racial bonus only, no class-based allocation), which made every class combat-equivalent and, combined with the original boss numbers, made fights brutally unwinnable. Switched to the 5e standard array (15/14/13/12/10/8) assigned by class priority before racial bonuses — both more authentic 5e chargen and the actual fix that made combat viable.
- Added regression tests for all of the above (`combat_test.go`, `dungeon_test.go`, `narrator_test.go`) — `go build`/`go vet`/`go test` all pass, `gofmt` clean.
- Verified live: 5 consecutive full dungeon runs via a WebSocket test script all completed successfully (account → character → dungeon entry → combat → boss defeat → `DUNGEON_RESOLVED` received), zero server errors in gateway/worker logs.
- Visually verified in an actual browser via Playwright (same approach as the prior session): screenshotted the combat log rendering (attack-by-attack d20 math, monster AC tags, victory/defeat narration) and confirmed the "Return to City Gates" button now actually closes the dungeon panel and returns to the overworld with updated gold/HP. Zero console errors.
- Rebuilt and verified all three changed Docker images (gateway, worker, frontend).

## Next steps (suggested, not started)
1. Confirm GitHub Actions are green on the latest pushed commits.
2. Decide on a party-formation flow (invite/accept) so `/party` chat and party voting can actually be exercised with 2+ players — this is the biggest remaining gap between "implemented" and "matches the full outline.md design." It would also let boss monsters be tuned back up toward their real SRD CR.
3. Decide whether to pursue inventory/spellbook persistence (deferred Phase 1 scope).
4. If/when a real hosting target is chosen, extend `docker-publish.yml` with an actual deploy step.
5. A responsive/mobile pass on the new UI — only verified at a 1280×900 desktop viewport so far.
6. Combat is currently melee-only with no spell-slot economy for the Wizard (cantrip reuses the same per-round attack loop as the Fighter's weapon). A real spellcasting system (slots, save-based spells, ranged positioning) is a natural next combat upgrade.
