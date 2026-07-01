package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"dnd5e-web/backend/internal/chat"
	"dnd5e-web/backend/internal/combat"
	"dnd5e-web/backend/internal/models"
	"dnd5e-web/backend/internal/narrator"
	"dnd5e-web/backend/internal/queue"
	"dnd5e-web/backend/internal/skillcheck"
	"dnd5e-web/backend/internal/world"
	"dnd5e-web/backend/internal/wsproto"
)

// roomInfoFor retrieves the label and description for a specific room from
// the dungeon run, falling back to generic names if the room isn't found.
// Rooms now carry their own Label and Description (set by dungeon.Generate)
// so narration can use each room's authored voice rather than a generic
// lookup table.
func roomInfoFor(dungeon models.Dungeon, roomType models.DungeonRoomType) (label, description string) {
	for _, r := range dungeon.Rooms {
		if r.Type == roomType {
			if r.Label != "" {
				label = r.Label
			}
			if r.Description != "" {
				description = r.Description
			}
			return
		}
	}
	// Fallback for runs generated before the label fields were added.
	switch roomType {
	case models.RoomStart:
		label = "entrance"
	case models.RoomHallway:
		label = "corridor"
	case models.RoomTreasure:
		label = "treasure vault"
	case models.RoomBoss:
		label = "boss's den"
	default:
		label = string(roomType)
	}
	return
}

const combatTurnTimeout = 60 * time.Second

func presentIDsOf(run *dungeonRun) []string {
	ids := make([]string, 0, len(run.PresentAccounts))
	for id := range run.PresentAccounts {
		ids = append(ids, id)
	}
	return ids
}

func encounterStateMessage(run *dungeonRun) wsproto.EncounterState {
	e := run.ActiveEncounter
	currentID := ""
	if current := e.Current(); current != nil {
		currentID = current.ID
	}
	return wsproto.EncounterState{
		Type:               "ENCOUNTER_STATE",
		Combatants:         e.Combatants,
		CurrentCombatantID: currentID,
		Round:              e.Round,
		Log:                e.Log,
		RoomType:           run.ActiveRoomType,
		RoomLabel:          run.ActiveRoomLabel,
	}
}

// partyCharacters resolves each present account's active character —
// accounts with no active character simply don't get a combatant.
func (s *server) partyCharacters(ctx context.Context, accountIDs []string) []models.Character {
	characters := make([]models.Character, 0, len(accountIDs))
	for _, id := range accountIDs {
		account, err := s.store.GetAccount(ctx, id)
		if err != nil || account.ActiveCharacterID == nil {
			continue
		}
		c, err := s.store.GetCharacter(ctx, *account.ActiveCharacterID)
		if err != nil {
			continue
		}
		characters = append(characters, c)
	}
	return characters
}

// handleEnterDungeon gates on standing at a quest-hook location. If the
// party already has an unresolved dungeon run, this is a hot-drop: the
// account joins the existing run (and any fight already in progress)
// instead of generating a new instance.
func (s *server) handleEnterDungeon(ctx context.Context, client *chat.Client) error {
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	loc, ok := world.Locations[account.LocationID]
	if !ok || loc.Kind != world.KindQuestHook {
		return wsError("there is nothing to explore here")
	}
	key := partyKey(account)

	s.dungeonsMu.Lock()
	run, exists := s.dungeons[key]
	if exists && !run.Dungeon.Resolved {
		run.PresentAccounts[client.AccountID] = true
		dungeonCopy := run.Dungeon
		var state *wsproto.EncounterState
		if run.ActiveEncounter != nil {
			msg := encounterStateMessage(run)
			state = &msg
		}
		s.dungeonsMu.Unlock()

		line := narrator.DungeonEntry(s.activeCharacterName(ctx, account))
		if err := client.WriteJSON(wsproto.NewDungeonReady(dungeonCopy, line)); err != nil {
			return err
		}
		if state != nil {
			return client.WriteJSON(*state)
		}
		return nil
	}
	s.dungeonsMu.Unlock()

	level := 1
	if account.ActiveCharacterID != nil {
		if c, err := s.store.GetCharacter(ctx, *account.ActiveCharacterID); err == nil {
			level = c.Level
		}
	}
	partyID := ""
	if account.PartyID != nil {
		partyID = *account.PartyID
	}
	job := queue.DungeonGenerationJob{
		JobID:          uuid.NewString(),
		PartyID:        partyID,
		AccountID:      client.AccountID,
		CharacterLevel: level,
	}
	return s.queue.Publish(ctx, queue.QueueDungeonGeneration, job)
}

// handleStartEncounter builds a turn-based fight from every account
// currently present in the dungeon run and the room's monsters.
func (s *server) handleStartEncounter(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.StartEncounterPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	key := partyKey(account)

	s.dungeonsMu.Lock()
	run, ok := s.dungeons[key]
	if !ok {
		s.dungeonsMu.Unlock()
		return wsError("no active dungeon instance")
	}
	if run.ActiveEncounter != nil {
		s.dungeonsMu.Unlock()
		return wsError("a fight is already underway")
	}
	// Match by label when provided (multiple rooms may share the same
	// functional type, e.g. two hallway rooms); fall back to first-uncleared
	// room of that type if no label is available.
	var room *models.DungeonRoom
	var encounterIdx int = -1
	for i := range run.Dungeon.Rooms {
		r := &run.Dungeon.Rooms[i]
		if r.Type != p.RoomType {
			continue
		}
		if p.RoomLabel != "" {
			if r.Label == p.RoomLabel && !r.Cleared {
				room = r
				encounterIdx = i
				break
			}
		} else if !r.Cleared {
			room = r
			encounterIdx = i
		}
	}
	if room == nil {
		s.dungeonsMu.Unlock()
		return wsError("that room has nothing left to fight")
	}
	var monsters []models.Monster
	if encounterIdx >= 0 && encounterIdx < len(run.Dungeon.Encounters) {
		monsters = run.Dungeon.Encounters[encounterIdx].Monsters
	}
	presentIDs := presentIDsOf(run)
	s.dungeonsMu.Unlock()

	characters := s.partyCharacters(ctx, presentIDs)

	s.dungeonsMu.Lock()
	run.ActiveEncounter = combat.NewEncounter(characters, monsters)
	run.ActiveRoomType = p.RoomType
	run.ActiveRoomLabel = p.RoomLabel
	state := encounterStateMessage(run)
	s.dungeonsMu.Unlock()

	monsterNames := make([]string, len(monsters))
	for i, m := range monsters {
		monsterNames[i] = m.Name
	}
	_, roomDesc := roomInfoFor(run.Dungeon, p.RoomType)
	line := narrator.RoomEntry(roomDesc, monsterNames)
	s.sendNarrationToAccounts(presentIDs, line)
	s.hub.BroadcastToAccounts(presentIDs, state)

	s.scheduleCombatTimeout(key)
	return nil
}

// handleCombatAction is the player-driven entry point for COMBAT_ACTION;
// it delegates to applyCombatAction with the sender's own account/character.
func (s *server) handleCombatAction(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.CombatActionPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if account.ActiveCharacterID == nil {
		return wsError("you have no active character")
	}
	return s.applyCombatAction(ctx, partyKey(account), client.AccountID, *account.ActiveCharacterID, p.Action, p.TargetID)
}

// applyCombatAction is the shared core for both player-submitted
// COMBAT_ACTION messages and the turn-timeout auto-action.
func (s *server) applyCombatAction(ctx context.Context, key, actorAccountID, actorCharacterID, action, targetID string) error {
	s.dungeonsMu.Lock()
	run, ok := s.dungeons[key]
	if !ok || run.ActiveEncounter == nil {
		s.dungeonsMu.Unlock()
		return wsError("no fight is underway")
	}
	encounter := run.ActiveEncounter
	current := encounter.Current()
	if current == nil || current.AccountID != actorAccountID {
		s.dungeonsMu.Unlock()
		return wsError("it is not your turn")
	}
	actorName := current.Name

	var line string
	var actionErr error
	switch action {
	case "attack":
		if targetID == "" {
			for _, c := range encounter.Combatants {
				if c.Kind == combat.KindMonster && c.Alive() {
					targetID = c.ID
					break
				}
			}
		}
		var roll combat.AttackRoll
		roll, actionErr = encounter.Attack(actorCharacterID, targetID)
		if actionErr == nil {
			line = narrator.AttackSwing(roll.Attacker, roll.Target, roll.Hit, roll.Critical, roll.Damage)
		}
	case "dodge":
		actionErr = encounter.Dodge(actorCharacterID)
		if actionErr == nil {
			line = narrator.Dodge(actorName)
		}
	case "flee":
		actionErr = encounter.Flee(actorCharacterID)
		if actionErr == nil {
			line = narrator.Flee(actorName)
		}
	default:
		actionErr = wsError("unknown combat action")
	}
	if actionErr != nil {
		s.dungeonsMu.Unlock()
		return actionErr
	}

	encounter.AdvanceTurn()
	over, victory := encounter.Outcome()
	presentIDs := presentIDsOf(run)

	hpUpdates := make(map[string]int, len(encounter.Combatants))
	for _, c := range encounter.Combatants {
		if c.Kind == combat.KindPlayer {
			hpUpdates[c.ID] = c.HP
		}
	}

	roomType := run.ActiveRoomType
	roomLabel := run.ActiveRoomLabel
	var combatLog []combat.AttackRoll
	var defeatedMonsters []string
	var dungeonCopy models.Dungeon
	if over {
		combatLog = append([]combat.AttackRoll{}, encounter.Log...)
		for _, c := range encounter.Combatants {
			if c.Kind == combat.KindMonster && c.Defeated {
				defeatedMonsters = append(defeatedMonsters, c.Name)
			}
		}
		run.ActiveEncounter = nil
		if victory {
			// Clear the SPECIFIC room that was being fought, not ALL rooms
			// of that type — multiple rooms can share a functional type.
			cleared := false
			for i := range run.Dungeon.Rooms {
				r := &run.Dungeon.Rooms[i]
				if r.Type != roomType || r.Cleared {
					continue
				}
				if roomLabel == "" || r.Label == roomLabel {
					r.Cleared = true
					cleared = true
					break
				}
			}
			if !cleared {
				// Fallback: clear the first uncleared room of that type.
				for i := range run.Dungeon.Rooms {
					if run.Dungeon.Rooms[i].Type == roomType && !run.Dungeon.Rooms[i].Cleared {
						run.Dungeon.Rooms[i].Cleared = true
						break
					}
				}
			}
			for i := range run.Dungeon.Rooms {
				if run.Dungeon.Rooms[i].Type == models.RoomBoss {
					run.Dungeon.Resolved = run.Dungeon.Rooms[i].Cleared
				}
			}
		}
		dungeonCopy = run.Dungeon
	}
	var state wsproto.EncounterState
	if !over {
		state = encounterStateMessage(run)
	}
	s.dungeonsMu.Unlock()

	for charID, hp := range hpUpdates {
		_ = s.store.UpdateCharacterHP(ctx, charID, hp)
	}

	s.sendNarrationToAccounts(presentIDs, line)

	if !over {
		s.hub.BroadcastToAccounts(presentIDs, state)
		s.scheduleCombatTimeout(key)
		return nil
	}

	// roomLabel is already set from run.ActiveRoomLabel above; fall back
	// to roomInfoFor if it was empty (pre-label runs or label not sent).
	if roomLabel == "" {
		roomLabel, _ = roomInfoFor(dungeonCopy, roomType)
	}
	var resultLine string
	if victory {
		resultLine = narrator.RoomVictory("The party", roomLabel, defeatedMonsters)
		_ = s.store.UpdateDungeonRooms(ctx, dungeonCopy.ID, dungeonCopy.Rooms)
	} else {
		resultLine = narrator.RoomDefeat("The party", roomLabel)
		// No permadeath: a defeated party retreats and heals rather than
		// the game ending — see internal/combat's package doc.
		for _, accountID := range presentIDs {
			account, err := s.store.GetAccount(ctx, accountID)
			if err != nil || account.ActiveCharacterID == nil {
				continue
			}
			if c, err := s.store.GetCharacter(ctx, *account.ActiveCharacterID); err == nil {
				_ = s.store.UpdateCharacterHP(ctx, c.ID, c.HPMax)
			}
		}
	}
	s.sendNarrationToAccounts(presentIDs, resultLine)
	s.hub.BroadcastToAccounts(presentIDs, wsproto.RoomResolved{
		Type:      "ROOM_RESOLVED",
		RoomType:  roomType,
		Victory:   victory,
		CombatLog: combatLog,
		Narration: resultLine,
		Dungeon:   dungeonCopy,
	})
	return nil
}

// scheduleCombatTimeout gives the active player combatTurnTimeout to act;
// if they don't, a basic attack is submitted on their behalf so one AFK
// friend can't freeze the table for everyone else.
func (s *server) scheduleCombatTimeout(key string) {
	s.dungeonsMu.Lock()
	run, ok := s.dungeons[key]
	if !ok || run.ActiveEncounter == nil {
		s.dungeonsMu.Unlock()
		return
	}
	current := run.ActiveEncounter.Current()
	if current == nil || current.Kind != combat.KindPlayer {
		s.dungeonsMu.Unlock()
		return
	}
	encounter := run.ActiveEncounter
	roundAtSchedule := encounter.Round
	combatantID := current.ID
	accountID := current.AccountID
	s.dungeonsMu.Unlock()

	time.AfterFunc(combatTurnTimeout, func() {
		s.dungeonsMu.Lock()
		run, ok := s.dungeons[key]
		stillWaiting := ok && run.ActiveEncounter == encounter &&
			encounter.Current() != nil && encounter.Current().ID == combatantID &&
			encounter.Round == roundAtSchedule
		s.dungeonsMu.Unlock()
		if !stillWaiting {
			return
		}
		_ = s.applyCombatAction(context.Background(), key, accountID, combatantID, "attack", "")
	})
}

// handleSkillCheck rolls a non-combat ability check for the sender's
// active character. A successful pre-combat check removes the room's
// weakest monster from the fight about to start — a clear, easy-to-narrate
// consequence rather than extra state bookkeeping.
func (s *server) handleSkillCheck(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.SkillCheckPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if account.ActiveCharacterID == nil {
		return wsError("you have no active character")
	}
	character, err := s.store.GetCharacter(ctx, *account.ActiveCharacterID)
	if err != nil {
		return err
	}
	key := partyKey(account)

	s.dungeonsMu.Lock()
	run, ok := s.dungeons[key]
	if !ok {
		s.dungeonsMu.Unlock()
		return wsError("there is nothing here to check")
	}
	var activeRoomType models.DungeonRoomType
	for _, room := range run.Dungeon.Rooms {
		if !room.Cleared {
			activeRoomType = room.Type
			break
		}
	}
	encounterIdx := -1
	for i, e := range run.Dungeon.Encounters {
		if e.RoomType == activeRoomType {
			encounterIdx = i
		}
	}
	s.dungeonsMu.Unlock()

	dc := skillcheck.DCFor(activeRoomType, p.Context)
	result := skillcheck.Roll(character, p.Skill, dc)
	line := narrator.SkillCheckOutcome(character.Name, p.Context, result.Success)

	if result.Success && encounterIdx != -1 {
		s.dungeonsMu.Lock()
		if run, ok := s.dungeons[key]; ok && encounterIdx < len(run.Dungeon.Encounters) {
			monsters := run.Dungeon.Encounters[encounterIdx].Monsters
			if len(monsters) > 0 {
				weakest := 0
				for i, m := range monsters {
					if m.HP < monsters[weakest].HP {
						weakest = i
					}
				}
				run.Dungeon.Encounters[encounterIdx].Monsters = append(monsters[:weakest], monsters[weakest+1:]...)
			}
		}
		s.dungeonsMu.Unlock()
	}

	s.sendNarration(client.AccountID, client.PartyID, line)
	return client.WriteJSON(wsproto.SkillCheckResult{Type: "SKILL_CHECK_RESULT", Result: result, Narration: line})
}

func (s *server) handleResolveDungeon(ctx context.Context, client *chat.Client) error {
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	key := partyKey(account)

	s.dungeonsMu.Lock()
	run, ok := s.dungeons[key]
	if ok && run.Dungeon.Resolved {
		delete(s.dungeons, key)
	}
	s.dungeonsMu.Unlock()

	if !ok || !run.Dungeon.Resolved {
		return wsError("the boss room has not been cleared yet")
	}

	if err := s.store.ResolveDungeon(ctx, run.Dungeon.ID); err != nil {
		return err
	}
	const goldReward = 25
	if err := s.store.AddGold(ctx, client.AccountID, goldReward); err != nil {
		return err
	}

	characterName := s.activeCharacterName(ctx, account)
	line := narrator.DungeonResolved(characterName, goldReward)
	s.sendNarrationToAccounts(presentIDsOf(run), line)

	if err := client.WriteJSON(wsproto.DungeonResolved{
		Type:        "DUNGEON_RESOLVED",
		Narration:   line,
		GoldAwarded: goldReward,
	}); err != nil {
		return err
	}

	sync, err := s.stateSync(ctx, client.AccountID)
	if err != nil {
		return err
	}
	return client.WriteJSON(sync)
}
