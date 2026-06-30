package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"

	"dnd5e-web/backend/internal/chat"
	"dnd5e-web/backend/internal/combat"
	"dnd5e-web/backend/internal/models"
	"dnd5e-web/backend/internal/narrator"
	"dnd5e-web/backend/internal/queue"
	"dnd5e-web/backend/internal/redisstate"
	"dnd5e-web/backend/internal/voting"
	"dnd5e-web/backend/internal/worldmap"
	"dnd5e-web/backend/internal/wsproto"
)

func (s *server) wsUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("accountId", c.Params("accountId"))
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (s *server) wsHandler(conn *websocket.Conn) {
	accountID := conn.Params("accountId")
	ctx := context.Background()

	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		_ = conn.WriteJSON(wsproto.NewError("unknown account"))
		_ = conn.Close()
		return
	}

	client := &chat.Client{Conn: conn, AccountID: accountID}
	if account.PartyID != nil {
		client.PartyID = *account.PartyID
	}
	s.hub.Register(client)
	defer s.hub.Unregister(accountID)

	if sync, err := s.stateSync(ctx, accountID); err == nil {
		_ = client.WriteJSON(sync)
	}

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var env wsproto.Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			_ = client.WriteJSON(wsproto.NewError("malformed message"))
			continue
		}
		s.dispatch(ctx, client, env)
	}
}

func (s *server) dispatch(ctx context.Context, client *chat.Client, env wsproto.Envelope) {
	var err error
	switch env.Type {
	case "MOVE":
		err = s.handleMove(ctx, client, env.Payload)
	case "SWAP_CHARACTER":
		err = s.handleSwapCharacter(ctx, client, env.Payload)
	case "CREATE_CHARACTER":
		err = s.handleCreateCharacter(ctx, client, env.Payload)
	case "RP_CHAT":
		err = s.handleChat(ctx, client, env.Payload)
	case "TALK_TO_NPC":
		err = s.handleTalkToNPC(ctx, client)
	case "MAKE_CHOICE":
		err = s.handleMakeChoice(ctx, client, env.Payload)
	case "CAST_VOTE":
		err = s.handleCastVote(ctx, client, env.Payload)
	case "ENTER_POI":
		err = s.handleEnterPOI(ctx, client)
	case "CLEAR_DUNGEON_ROOM":
		err = s.handleClearDungeonRoom(ctx, client, env.Payload)
	case "RESOLVE_DUNGEON":
		err = s.handleResolveDungeon(ctx, client)
	default:
		err = errUnknownMessageType
	}
	if err != nil {
		_ = client.WriteJSON(wsproto.NewError(err.Error()))
	}
}

var errUnknownMessageType = wsError("unknown message type")

type wsError string

func (e wsError) Error() string { return string(e) }

func (s *server) handleMove(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.MovePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	next := models.Coordinate{X: account.Coordinate.X + p.DX, Y: account.Coordinate.Y + p.DY}
	if !worldmap.IsWalkable(next) {
		return nil // silently ignore illegal moves, same as the original client-only prototype
	}
	if err := s.store.UpdateCoordinate(ctx, client.AccountID, next); err != nil {
		return err
	}
	_ = s.redis.SetCoordinate(ctx, client.AccountID, next)

	sync, err := s.stateSync(ctx, client.AccountID)
	if err != nil {
		return err
	}
	return client.WriteJSON(sync)
}

func (s *server) handleSwapCharacter(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.SwapCharacterPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if !worldmap.IsAdjacentOrEqual(account.Coordinate, worldmap.GuildHall) {
		return wsError("must be at the Adventurer's Guild Hall to swap characters")
	}
	if err := s.store.SetActiveCharacter(ctx, client.AccountID, p.CharacterID); err != nil {
		return err
	}
	sync, err := s.stateSync(ctx, client.AccountID)
	if err != nil {
		return err
	}
	return client.WriteJSON(sync)
}

func (s *server) handleCreateCharacter(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.CreateCharacterPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	character, err := buildCharacter(client.AccountID, p.Name, p.RaceID, p.ClassID)
	if err != nil {
		return err
	}
	if err := s.store.CreateCharacter(ctx, character); err != nil {
		return err
	}

	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if account.ActiveCharacterID == nil {
		if err := s.store.SetActiveCharacter(ctx, client.AccountID, character.ID); err != nil {
			return err
		}
	}

	sync, err := s.stateSync(ctx, client.AccountID)
	if err != nil {
		return err
	}
	return client.WriteJSON(sync)
}

func (s *server) handleChat(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.ChatPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}

	msg := models.ChatMessage{
		Channel:   p.Channel,
		AccountID: client.AccountID,
		Body:      p.Body,
		Timestamp: time.Now(),
	}

	switch p.Channel {
	case models.ChannelGuild:
		if !worldmap.IsAdjacentOrEqual(account.Coordinate, worldmap.GuildHall) {
			return wsError("you must be inside the Guild Hall to use /guild chat")
		}
	case models.ChannelParty:
		if client.PartyID == "" {
			return wsError("you are not in a party")
		}
	case models.ChannelRP:
		if account.ActiveCharacterID != nil {
			characters, err := s.store.ListCharacters(ctx, client.AccountID)
			if err == nil {
				for _, ch := range characters {
					if ch.ID == *account.ActiveCharacterID {
						msg.Name = ch.Name
						msg.Race = string(ch.RaceID)
						msg.Class = string(ch.ClassID)
					}
				}
			}
		}
	}

	broadcast := wsproto.NewChatBroadcast(msg)
	switch p.Channel {
	case models.ChannelGlobal:
		_ = s.redis.Publish(ctx, redisstate.ChannelChatGlobal, broadcast)
	case models.ChannelGuild:
		_ = s.redis.Publish(ctx, redisstate.ChannelChatGuild, broadcast)
	case models.ChannelRP:
		_ = s.redis.Publish(ctx, redisstate.ChannelChatRP, broadcast)
	case models.ChannelParty:
		_ = s.redis.Publish(ctx, redisstate.ChannelChatParty(client.PartyID), broadcast)
	default:
		return wsError("unknown chat channel")
	}
	return nil
}

func (s *server) handleTalkToNPC(ctx context.Context, client *chat.Client) error {
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if !worldmap.IsAdjacentOrEqual(account.Coordinate, worldmap.NPC) {
		return wsError("there is no one to talk to here")
	}

	prompt := defaultPrompt()

	if client.PartyID == "" {
		return client.WriteJSON(wsproto.ChoiceState{
			Type:     "CHOICE_STATE",
			PromptID: prompt.ID,
			Prompt:   prompt.Prompt,
			Mode:     models.ChoiceModeSolo,
			Options:  prompt.Options,
		})
	}

	prompt.Mode = models.ChoiceModeParty
	room := voting.NewVoteRoom(prompt)
	s.voteRoomsMu.Lock()
	s.voteRooms[prompt.ID] = room
	s.voteRoomsMu.Unlock()

	deadline := room.Deadline.UnixMilli()
	s.hub.BroadcastToParty(client.PartyID, wsproto.ChoiceState{
		Type:     "CHOICE_STATE",
		PromptID: prompt.ID,
		Prompt:   prompt.Prompt,
		Mode:     models.ChoiceModeParty,
		Options:  prompt.Options,
		Deadline: &deadline,
	})

	go s.resolveVoteAfterDeadline(prompt.ID, client.PartyID, room)
	return nil
}

func (s *server) handleMakeChoice(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.MakeChoicePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	prompt, ok := npcPrompts[p.PromptID]
	if !ok {
		return wsError("unknown prompt")
	}
	var typology models.ChoiceTypology
	found := false
	for _, o := range prompt.Options {
		if o.ID == p.OptionID {
			typology = o.Typology
			found = true
		}
	}
	if !found {
		return wsError("unknown option")
	}

	newHonor, err := s.store.ApplyHonorDelta(ctx, client.AccountID, models.HonorImpact[typology], typology, "npc choice: "+prompt.ID)
	if err != nil {
		return err
	}

	optionLabel := p.OptionID
	for _, o := range prompt.Options {
		if o.ID == p.OptionID {
			optionLabel = o.Label
		}
	}
	line := narrator.ChoiceResolution(optionLabel, string(typology))
	s.sendNarration(client.AccountID, "", line)

	return client.WriteJSON(wsproto.VoteResolved{
		Type:       "VOTE_RESOLVED",
		PromptID:   prompt.ID,
		OptionID:   p.OptionID,
		HonorDelta: models.HonorImpact[typology],
		NewHonor:   newHonor,
		Narration:  line,
	})
}

func (s *server) handleCastVote(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.CastVotePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	s.voteRoomsMu.Lock()
	room, ok := s.voteRooms[p.PromptID]
	s.voteRoomsMu.Unlock()
	if !ok {
		return wsError("voting window is not open")
	}
	if err := room.CastVote(client.AccountID, p.OptionID); err != nil {
		return err
	}

	s.hub.BroadcastToParty(client.PartyID, wsproto.VoteUpdate{
		Type:     "VOTE_UPDATE",
		PromptID: p.PromptID,
		Tallies:  room.Tally(),
	})
	return nil
}

// resolveVoteAfterDeadline waits out the 30-second window, resolves the
// vote, and hands the (potentially many) Honor updates to the worker via
// RabbitMQ rather than writing them inline on the gateway.
func (s *server) resolveVoteAfterDeadline(promptID, partyID string, room *voting.VoteRoom) {
	time.Sleep(time.Until(room.Deadline))

	s.voteRoomsMu.Lock()
	delete(s.voteRooms, promptID)
	s.voteRoomsMu.Unlock()

	renown := make(map[string]int)
	accountIDs := make([]string, 0, len(room.Votes))
	seen := make(map[string]bool)
	ctx := context.Background()
	for accountID := range room.Votes {
		if seen[accountID] {
			continue
		}
		seen[accountID] = true
		accountIDs = append(accountIDs, accountID)
		if account, err := s.store.GetAccount(ctx, accountID); err == nil {
			renown[accountID] = account.Honor
		}
	}

	res := voting.Resolve(room, renown)

	var typology models.ChoiceTypology
	for _, o := range room.Prompt.Options {
		if o.ID == res.OptionID {
			typology = o.Typology
		}
	}

	if err := s.queue.Publish(ctx, queue.QueueVoteResolution, queue.VoteResolutionJob{
		PromptID:        promptID,
		WinningOptionID: res.OptionID,
		Typology:        typology,
		AccountIDs:      accountIDs,
		TieBreak:        res.TieBreakUsed,
	}); err != nil {
		log.Printf("publish vote resolution job: %v", err)
	}
}

func (s *server) handleEnterPOI(ctx context.Context, client *chat.Client) error {
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if !worldmap.IsAdjacentOrEqual(account.Coordinate, worldmap.POI) {
		return wsError("there is nothing to explore here yet")
	}
	level := 1
	if account.ActiveCharacterID != nil {
		characters, err := s.store.ListCharacters(ctx, client.AccountID)
		if err == nil {
			for _, ch := range characters {
				if ch.ID == *account.ActiveCharacterID {
					level = ch.Level
				}
			}
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

var roomLabels = map[models.DungeonRoomType]string{
	models.RoomStart:    "entrance",
	models.RoomHallway:  "corridor",
	models.RoomTreasure: "treasure vault",
	models.RoomBoss:     "boss's den",
}

// handleClearDungeonRoom actually fights the room's encounter (see
// internal/combat) instead of instantly flipping a flag: a real d20 attack
// roll against the SRD Armor Class for each monster, real damage dice, and
// the character's own HP at risk. Losing an encounter doesn't end the
// dungeon run — see internal/combat's package docs for why this prototype
// has no permadeath.
func (s *server) handleClearDungeonRoom(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.ClearDungeonRoomPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if account.ActiveCharacterID == nil {
		return wsError("recruit a character at the Guild Hall before adventuring")
	}
	character, err := s.store.GetCharacter(ctx, *account.ActiveCharacterID)
	if err != nil {
		return err
	}
	key := partyKey(account)

	s.dungeonsMu.Lock()
	d, ok := s.dungeons[key]
	if !ok {
		s.dungeonsMu.Unlock()
		return wsError("no active dungeon instance")
	}
	var room *models.DungeonRoom
	for i := range d.Rooms {
		if d.Rooms[i].Type == p.RoomType {
			room = &d.Rooms[i]
		}
	}
	if room == nil {
		s.dungeonsMu.Unlock()
		return wsError("unknown room")
	}
	if room.Cleared {
		s.dungeonsMu.Unlock()
		return wsError("that room is already cleared")
	}
	monsters := []models.Monster{}
	for _, e := range d.Encounters {
		if e.RoomType == p.RoomType {
			monsters = e.Monsters
		}
	}
	s.dungeonsMu.Unlock()

	result := combat.Resolve(character, monsters)

	// A defeat isn't a permadeath (see internal/combat docs) — the
	// character retreats and is fully healed rather than persisting at the
	// narrative "barely standing" 1 HP, so a lost encounter is a setback to
	// retry, not a permanent soft-lock.
	persistedHP := result.CharacterHPAfter
	if !result.Victory {
		persistedHP = character.HPMax
	}
	if err := s.store.UpdateCharacterHP(ctx, character.ID, persistedHP); err != nil {
		return err
	}

	roomLabel := roomLabels[p.RoomType]
	var line string
	if result.Victory {
		line = narrator.RoomVictory(character.Name, roomLabel, result.MonstersDefeated)
	} else {
		line = narrator.RoomDefeat(character.Name, roomLabel)
	}

	s.dungeonsMu.Lock()
	if result.Victory {
		room.Cleared = true
		for i := range d.Rooms {
			if d.Rooms[i].Type == models.RoomBoss {
				d.Resolved = d.Rooms[i].Cleared
			}
		}
	}
	dungeonCopy := *d
	s.dungeonsMu.Unlock()

	s.sendNarration(client.AccountID, client.PartyID, line)

	if err := client.WriteJSON(wsproto.RoomResolved{
		Type:      "ROOM_RESOLVED",
		RoomType:  p.RoomType,
		Victory:   result.Victory,
		CombatLog: result.Rounds,
		Narration: line,
		Dungeon:   dungeonCopy,
	}); err != nil {
		return err
	}

	sync, err := s.stateSync(ctx, client.AccountID)
	if err != nil {
		return err
	}
	return client.WriteJSON(sync)
}

func (s *server) handleResolveDungeon(ctx context.Context, client *chat.Client) error {
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	key := partyKey(account)

	s.dungeonsMu.Lock()
	d, ok := s.dungeons[key]
	if ok && d.Resolved {
		delete(s.dungeons, key)
	}
	s.dungeonsMu.Unlock()

	if !ok || !d.Resolved {
		return wsError("the boss room has not been cleared yet")
	}

	if err := s.store.ResolveDungeon(ctx, d.ID); err != nil {
		return err
	}
	const goldReward = 25
	if err := s.store.AddGold(ctx, client.AccountID, goldReward); err != nil {
		return err
	}

	characterName := s.activeCharacterName(ctx, account)
	line := narrator.DungeonResolved(characterName, goldReward)
	s.sendNarration(client.AccountID, client.PartyID, line)

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
