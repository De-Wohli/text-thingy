Adding asynchronous message queues, real-time chat, and democratic party choices elevates this from a simple solo web game to a true multiplayer virtual tabletop. Golang handles high concurrency and WebSockets effortlessly, while RabbitMQ is perfect for managing decoupling, crafting timers, and heavy event routing.
Here is the ideal tech stack tailored for high modifiability, followed by the expanded AI Agent System Prompt containing the new RP, voting, and technical requirements.
Recommended Asynchronous Tech Stack
To keep this system highly modular, easy to modify, and lightning-fast, we want to split responsibilities cleanly.



[ Frontend: React / Tailwind ]
         │
         ▼  (WebSockets / REST API)
[ Gateway / Backend: Go (Fiber or Gin) ] ──(Pub/Sub)── [ Redis: Live Chat/Votes ]
         │
         ▼  (Publish Events)
   [ RabbitMQ ]
         │
         ▼  (Consume Tasks)
[ Asynchronous Workers (Go) ] ────► [ Database: PostgreSQL ]
   (Dungeon Gen, Crafting Timers)


1. Backend: Go (Golang) with Fiber or Gin
Why: Go's lightweight goroutines make handling thousands of concurrent WebSocket connections trivial.
Ease of Modification: Use Fiber (Express-like syntax, incredibly approachable) or Gin for the API gateway and WebSocket handlers.
2. Message Broker & Event Bus: RabbitMQ
Why: It excels at routing distinct event types using AMQP exchanges.
Queues to set up: * dungeon.generate: Offloads heavy procedural grid generation so the main API never lags.
crafting.timers: Handles the delayed background processing for real-time items.
quest.resolution: Processes the heavy database writes when a quest concludes or a vote closes.
3. Real-Time State Cache: Redis
Why: While RabbitMQ handles the tasks, Redis should handle transient, real-time states like active text RP rooms, live vote counts, and player coordinates. It keeps the database from being throttled.
4. Frontend: React or Svelte + Tailwind CSS
Why: Svelte or React allows you to build modular components for your tabletop UI. Tailwind CSS makes changing the visual aesthetic (like switching to a dark parchment style) a single-line update.
Expanded Agent System Prompt: 5e Web MMO (Phase 1 + RP & Voting)
Role & Objective
You are an expert Game Architect and Full-Stack Engineer specializing in event-driven architectures. Your objective is to generate the foundational backend architecture in Golang, utilizing RabbitMQ for asynchronous tasks, and a responsive Modern Web Framework (React/Svelte) frontend wrapper.
This prototype features an account-wide Roster, Honor/Alignment tracker, Text RP Engine, and a Party Choice/Voting System.
1. Scope & Ruleset Constraints (Core Data Models)
Implement the following strict data matrices for character creation and progression.
Races & Classes
Races:
Human: +1 to all Ability Scores.
Tiefling: Darkvision (60ft), Hellish Resistance (Fire damage resistance), +2 Charisma, +1 Intelligence.
Classes (Level 1):
Fighter: Hit Die: $1d10$, Proficiencies: All armor, Shields, Martial weapons. Features: Second Wind.
Wizard: Hit Die: $1d6$, Proficiencies: Daggers, Darts, Slings, Quarterstaffs. Features: Arcane Recovery, Spellcasting (3 Cantrips, 2 1st-level slots).
2. Text RP & Party Voting Systems
A. The Text RP Engine
The frontend must provide a dedicated tabletop chat log overlay. The Go backend routes messages over WebSockets to specific channels based on player coordinates:
/global - Server-wide out-of-character chat.
/guild - Read/write access inside Guild Halls or Outposts.
/party - Restricted to the player's current active party.
/rp - In-character localized chat. Appends the character's Name, Race, and Class to the message metadata.
B. Dynamic Choice & Voting Engine
When interacting with NPCs or triggering narrative forks in dungeons, the system shifts into a Choice State.
Solo Mode: The player clicks their choice directly.
Party Mode (Voting): A 30-second voting window triggers.
The backend broadcasts a JSON payload detailing the prompt and options via WebSockets.
Players cast their votes.
Tie-Breaker: If votes split evenly, the player with the highest Guild Renown (or a random coin flip rolled via the server) determines the outcome.
C. Honor & Alignment Impact Matrix
Decisions made via choices or votes alter the Account's Honor Score (-100 to +100), reshaping how NPCs respond.
Choice Typology
Honor Impact
Narrative / World Effect
Merciful / Noble
+10 Honor
Unlocks "Good" alignment paths, decreases shop prices by 10%.
Pragmatic / Selfish
0 Honor
Keeps alignment "Neutral", default world responses.
Ruthless / Malicious
-10 Honor
Unlocks "Evil" sub-quests, causes guards to track player coordinates.

3. UI Framework: Tabletop Aesthetic & ASCII Overworld
Render the interface with a modern web framework using a monospace canvas for the overworld grid alongside a clean, column-based tabletop character dashboard.
The ASCII Map Layout Canvas



Plaintext
############################################################
# . . . . . . . . . . . . . . . . . . . . . . . . . . . .  #
# . . . [A] . . . . . . . . . . . . . . . . . . . . . . .  #
# . . . . . . . . . . . . . . . . . . . . . . . . . . . .  #
# . . . . . . . . . . . . . . [T] . . . . . . . . . . . .  #
# . . . . . . . . . . . . . . . . . . . . . . . . . . . .  #
# . . . . . . . . . . @ . . . . . . . . . . . . . . . . .  #
# . . . . . . . . . . . . . . . . . . . . . . [N] . . . .  #
# ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ [?] . . #
############################################################


@ : Active Player Character | [A] : Guild Hall (Hotswap point) | [T] : Tavern | [N] : Quest Giver | [?] : Procedural Dungeon
4. Technical Architecture Requirements (Go & RabbitMQ)
The architecture must demonstrate a decoupled, event-driven pattern using standard Go idioms and an AMQP broker framework.
A. Go WebSocket / Gateway Server
Manage client lifecycle connections.
Parse incoming JSON actions: MOVE, RP_CHAT, CAST_VOTE, SWAP_CHARACTER.
Directly communicate with Redis to fetch active chat/vote rooms.
B. RabbitMQ Event Bus Integration
Define structured publishers and consumers for the following asynchronous operations:
dungeon_generation_queue: When an NPC quest is accepted or a player steps onto a [?] node, publish a payload ({ "party_id": "xyz", "cr_rating": 1 }). A background worker picks this up, generates the procedural array matrix, saves it to PostgreSQL, and notifies the party via WebSockets that the instance is ready.
vote_resolution_queue: When a 30-second timer expires, the voting window publishes its results to this queue. A worker aggregates the tallies, calculates the resulting Honor shift, updates the database, and clears the active room state.
Expected Output Deliverables
Provide the prototype codebase divided into:
Golang Backend Architecture: Code files implementing the WebSocket hub, a basic HTTP router, and the RabbitMQ Publisher/Consumer configuration for handling dungeon requests and voting inputs.
Frontend Interface (HTML/CSS/JS or Framework components): A functional dashboard showcasing the ASCII canvas layout alongside split windows for Text RP Chat and Active Voting Prompts. Ensure pressing keys updates coordinates and triggers WebSocket state sync.
Do you want to focus next on drafting the specific Go struct models for the WebSocket payloads (chat, movement, votes), or look into how RabbitMQ handles the asynchronous dungeon creation tasks when a quest is triggered?

---

## Implementation Note — UI direction (2026-06-30)

The backend, protocol, and underlying tile-grid movement/coordinate model described above were implemented as specified (MOVE/RP_CHAT/CAST_VOTE/SWAP_CHARACTER over WebSockets, RabbitMQ-backed dungeon generation and vote resolution, etc.). The literal *rendering* of section 3 was deliberately reinterpreted after the first pass felt closer to a terminal/ASCII dungeon crawler than the "tabletop" framing this section asks for:

- **The overworld map is a visual board, not monospace text.** Each grid cell renders as a colored/iconed tile (CSS grid, not a `<pre>` of `#`/`.`/`~` characters) — closer to a printed tabletop map with tokens on it than a terminal screen. The underlying data model (a `TileType[][]` grid, walkability rules, landmark coordinates) is unchanged; only the presentation layer changed.
- **Movement has an on-screen directional pad** (cardinal-direction buttons) in addition to WASD/arrow keys, so the game is playable without a keyboard.
- **Contextual actions are surfaced as buttons, not discovered by exploration.** A "Here" panel lists exactly the actions available at the player's current position (Talk to Citizen, Enter the Guild Hall, Investigate the point of interest), appearing only when in range rather than requiring the player to know the interaction exists.
- **The dungeon view is a room-card encounter track** (Entrance → Corridor → Treasure Vault → Boss's Den, each a card with its monster list and a "Resolve Encounter" button) instead of rendering the 15×15 grid as ASCII — this reads as a tabletop dungeon-crawl board rather than a video-game minimap, while the server-side procedural generation (CR budget, room layout) is unchanged.

See `frontend/src/components/OverworldCanvas.tsx`, `DirectionPad.tsx`, `LocationActions.tsx`, and `DungeonView.tsx`.

## Implementation Note — Game Master narrator + real 5e combat (2026-06-30)

Two more requests drove this pass: a "Game Master" narrator voice running through the experience, and combat that follows 5e rules more closely instead of "click a button to instantly clear the room."

- **Combat is now resolved with real dice**, not a flag flip: `backend/internal/combat` rolls a d20 + proficiency bonus + ability modifier against each monster's SRD Armor Class, applies SRD damage dice, and alternates the character's attack with every surviving monster's attack until one side is defeated. Character HP persists in Postgres across the fight.
- **A losing encounter isn't permadeath.** 5e doesn't define what "game over" means for a prototype like this, so a defeated character retreats and is healed back to full rather than being stuck at the SRD death-save threshold — see `internal/combat`'s package doc for the reasoning. This means an encounter has real stakes (you can lose a specific attempt) without a soft-lock or a "you lose, restart everything" failure state.
- **Boss-tier monsters are tuned below their real SRD CR.** The Bandit Captain and Cult Fanatic are CR 2 in the book, which assumes a 4-character party splitting the action economy. This prototype has no party-formation flow yet (every dungeon is fought solo), and a true CR 2 monster is mathematically unwinnable for one level-1 character. Tuned down to roughly CR 1/2 toughness instead — "dangerous but winnable solo fight." Revisit once party formation exists.
- **Character ability scores now follow the 5e standard array (15/14/13/12/10/8), assigned by class priority, with racial bonuses on top** — not flat 10s across the board. This was a quieter but necessary fix: the original flat-10 chargen made every class equally (and unrealistically) weak in combat, which is itself a combat-balance bug once real dice are involved.
- **A Game-Master-voiced Narrator** generates flavor text for dungeon entry, every attack roll, room victory/defeat, the final boss reward, and NPC/party choice resolutions (`backend/internal/narrator`, template-based — not an LLM call, to keep gateway latency and dependency surface small). Narration is both shown inline where the relevant action happened (the dungeon combat log, the choice resolution panel) and logged persistently to a new `/gm` chat channel so the table has a session record, not just a transient toast.

See `backend/internal/combat`, `backend/internal/narrator`, and `frontend/src/components/DungeonView.tsx` / `ChoicePanel.tsx` / `ChatPanel.tsx`.

## Implementation Note — VTT rework: location graph, party system, turn-based combat, hot-dropping (2026-07-01)

This pass completed the pivot to an actual virtual tabletop for online play. The key design decisions and their rationale:

- **The overworld moved from a tile grid to a location graph.** WASD movement between individual grid squares isn't how a tabletop session works — the DM says "you arrive at the tavern." The world is now a small hub-and-spoke graph of named locations (Town Square → Guild Hall, Tavern, Market Square, Old Mine Entrance), navigated by clicking. The underlying walkability/adjacency logic is gone; `backend/internal/world` just checks whether an edge exists in the graph. Locations carry their own SRD-style flavor descriptions, narrated by the Game Master on arrival. See `frontend/src/components/WorldMap.tsx`.

- **Party system added.** Any player can invite any other by display name; the target gets a PARTY_INVITE_RECEIVED prompt and accepts or declines; on accept both share a `party_id` (and a real `parties` table row) that scopes dungeon runs, chat, and voting. Requires a name on first login — `backend/cmd/gateway/party.go`, `frontend/src/components/PartyPanel.tsx`, `frontend/src/components/Onboarding.tsx`.

- **Combat is now turn-based.** A real initiative order (d20 + DEX modifier, highest first), one action per combatant per turn, monster turns auto-resolved by the server (the DM is automated, not a human), broadcast to the whole party in real time as `ENCOUNTER_STATE`. Players act via Attack (pick a target), Dodge (disadvantage on next incoming attack), or Flee (leave the encounter without ending the fight for everyone else). A 60-second per-turn timeout auto-submits a basic attack so one AFK player can't freeze the table. `backend/internal/combat` — `Resolve()` removed, replaced by the stateful `Encounter` engine. The gateway holds one `*combat.Encounter` per in-progress dungeon run in memory (mid-turn state is not persisted to Postgres; an acceptable simplification for a prototype — room-cleared progress IS durably persisted via `UpdateDungeonRooms`).

- **Hot-dropping works as designed.** A player can ENTER_DUNGEON at the Old Mine Entrance to join their party's existing in-progress run — they're added to `dungeonRun.PresentAccounts`, and if a fight is already underway they immediately receive `ENCOUNTER_STATE` to drop straight into the combat. The initial `DUNGEON_READY` notification is only sent to the player who actually initiated entry (not broadcast to the whole party preemptively) — a subtle but important detail that prevents other party members' dungeon view from opening before they've physically traveled there. `backend/cmd/gateway/combat.go` / `listener.go`.

- **Skill checks added for non-combat 5e interactions.** Before starting a room encounter, players can "Search for Traps" (Investigation) or "Listen for Danger" (Perception): d20 + ability modifier + proficiency vs. a DC, narrated outcome, one concrete mechanical consequence (successful check removes the room's weakest monster before the fight). Six skills modeled: Perception, Investigation, Insight, Stealth, Arcana, Athletics — with fixed class proficiencies (Fighter: Athletics/Perception; Wizard: Arcana/Investigation). `backend/internal/skillcheck`.

- **Narrator extensibility seam.** `narrator.Active` is a swappable `Backend` interface; the default `templateBackend` picks randomly from canned phrases. A future Ollama-backed implementation can be swapped in via a single line at startup without touching any call sites. Not built yet — just the seam.

- **Bugs found by actually running the new code**, all fixed with regression tests: nil-slice JSON serialization in `combat.Encounter.Log`/`Combatants` fields (the third time this exact pattern appeared in this codebase — documented with a comment in the fix); `DUNGEON_READY` being incorrectly broadcast to the whole party before party members had individually entered (premature dungeon UI opening on non-entered players); WASD movement still required until user explicitly asked for WASD removal (confirmed by user: "a room per room basis works too if that fits better").

See `backend/internal/world`, `backend/internal/combat`, `backend/internal/skillcheck`, `backend/cmd/gateway/{world,party,combat}.go`, and `frontend/src/components/{WorldMap,LocationScene,PartyPanel,CombatView,Onboarding}.tsx`.
