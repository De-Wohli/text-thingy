package main

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"dnd5e-web/backend/internal/chat"
	"dnd5e-web/backend/internal/combat"
	"dnd5e-web/backend/internal/models"
	"dnd5e-web/backend/internal/queue"
	"dnd5e-web/backend/internal/redisstate"
	"dnd5e-web/backend/internal/store"
	"dnd5e-web/backend/internal/voting"
	"dnd5e-web/backend/internal/wsproto"
)

// roomCombatModifier carries deferred bonuses from pre-combat skill checks
// that are applied when the room's encounter is built.
type roomCombatModifier struct {
	PlayerInitiativeBonus int  // Perception success: party always goes first
	PlayerAttackBonus     int  // Insight success: +N to all party attack rolls
	PlayerDamageBonus     int  // Arcana success: +N to all party damage rolls
	PlayerTempHP          int  // Athletics success: +N HP before the fight
	SneakAttack           bool // Stealth success: first player gets a free attack
	MonsterAlertBonus     int  // Perception/Stealth failure: monsters get +N to first attack
}

// dungeonRun is the gateway's in-memory view of one party's active dungeon
// instance: which accounts have actually traveled in (the hot-drop
// roster — see handleEnterDungeon) and the currently-active turn-based
// encounter, if a fight is in progress. Room-cleared progress is also
// persisted to Postgres (UpdateDungeonRooms) so it survives a restart;
// mid-fight turn state is in-memory only — see outline.md's
// implementation note for why that's an acceptable simplification here.
type dungeonRun struct {
	Dungeon         models.Dungeon
	PresentAccounts map[string]bool
	ActiveEncounter *combat.Encounter
	ActiveRoomType  models.DungeonRoomType
	ActiveRoomLabel string // label disambiguates rooms sharing the same functional type

	// SkillCooldowns tracks when a failed skill check's retry window expires.
	// key = accountID + ":" + string(skill) + ":" + roomLabel
	SkillCooldowns map[string]time.Time

	// RoomModifiers holds deferred skill-check bonuses, consumed by
	// handleStartEncounter when the encounter is built.
	// key = roomLabel
	RoomModifiers map[string]*roomCombatModifier
}

type partyInvite struct {
	ID              string
	FromAccountID   string
	FromDisplayName string
	ToAccountID     string
}

type server struct {
	store *store.Store
	redis *redisstate.Client
	queue *queue.Client
	hub   *chat.Hub

	voteRoomsMu sync.Mutex
	voteRooms   map[string]*voting.VoteRoom // promptID -> room

	dungeonsMu sync.Mutex
	dungeons   map[string]*dungeonRun // partyKey -> in-flight dungeon run

	partyInvitesMu sync.Mutex
	partyInvites   map[string]*partyInvite // inviteID -> invite

	// presence is gateway-in-memory only (rebuilt on connect/travel) —
	// purely for "who else is here to invite" UI, not durable state.
	presenceMu sync.Mutex
	presence   map[models.LocationID]map[string]bool
}

func newServer(st *store.Store, rs *redisstate.Client, q *queue.Client) *server {
	return &server{
		store:        st,
		redis:        rs,
		queue:        q,
		hub:          chat.NewHub(),
		voteRooms:    make(map[string]*voting.VoteRoom),
		dungeons:     make(map[string]*dungeonRun),
		partyInvites: make(map[string]*partyInvite),
		presence:     make(map[models.LocationID]map[string]bool),
	}
}

// partyKey returns the account's party ID, or the account ID itself when
// solo — this lets dungeon/voting code treat "solo" as a party of one
// instead of branching everywhere.
func partyKey(account models.Account) string {
	if account.PartyID != nil && *account.PartyID != "" {
		return *account.PartyID
	}
	return account.ID
}

// sendNarration delivers a Game-Master-voiced line to whoever should see it
// — the triggering account when solo, the whole party otherwise — as a
// regular chat message on the "narrator" channel, so it's both shown
// inline by the caller (most handlers also embed the same text in their
// response payload) and kept in the persistent chat log.
func (s *server) sendNarration(accountID, partyID, body string) {
	msg := models.ChatMessage{
		Channel:   models.ChannelNarrator,
		AccountID: "narrator",
		Name:      "Game Master",
		Body:      body,
		Timestamp: time.Now(),
	}
	broadcast := wsproto.NewChatBroadcast(msg)
	if partyID != "" {
		s.hub.BroadcastToParty(partyID, broadcast)
	} else {
		s.hub.SendTo(accountID, broadcast)
	}
}

// sendNarrationToAccounts is sendNarration's counterpart for a roster
// that isn't necessarily a formal party — e.g. everyone present in a
// dungeon run, used throughout combat (see combat.go).
func (s *server) sendNarrationToAccounts(accountIDs []string, body string) {
	msg := models.ChatMessage{
		Channel:   models.ChannelNarrator,
		AccountID: "narrator",
		Name:      "Game Master",
		Body:      body,
		Timestamp: time.Now(),
	}
	s.hub.BroadcastToAccounts(accountIDs, wsproto.NewChatBroadcast(msg))
}

// activeCharacterName looks up the account's active character for
// narration purposes, falling back to a generic label if none is set.
func (s *server) activeCharacterName(ctx context.Context, account models.Account) string {
	if account.ActiveCharacterID == nil {
		return "The adventurer"
	}
	c, err := s.store.GetCharacter(ctx, *account.ActiveCharacterID)
	if err != nil {
		return "The adventurer"
	}
	return c.Name
}

func (s *server) stateSync(ctx context.Context, accountID string) (wsproto.StateSync, error) {
	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		return wsproto.StateSync{}, err
	}
	characters, err := s.store.ListCharacters(ctx, accountID)
	if err != nil {
		return wsproto.StateSync{}, err
	}
	return wsproto.NewStateSync(account, characters), nil
}

// --- presence registry ---
// Tracks which accounts are standing at which location, purely so the
// client can show "who's here" and offer to invite them to a party.

func (s *server) setPresence(accountID string, locationID models.LocationID) {
	s.presenceMu.Lock()
	defer s.presenceMu.Unlock()
	for _, set := range s.presence {
		delete(set, accountID)
	}
	if s.presence[locationID] == nil {
		s.presence[locationID] = make(map[string]bool)
	}
	s.presence[locationID][accountID] = true
}

func (s *server) removePresence(accountID string) {
	s.presenceMu.Lock()
	defer s.presenceMu.Unlock()
	for _, set := range s.presence {
		delete(set, accountID)
	}
}

func (s *server) presentAccountIDs(locationID models.LocationID) []string {
	s.presenceMu.Lock()
	defer s.presenceMu.Unlock()
	ids := make([]string, 0, len(s.presence[locationID]))
	for id := range s.presence[locationID] {
		ids = append(ids, id)
	}
	return ids
}

// buildPresentAccounts resolves account IDs into display names for the
// client, excluding the requesting account itself.
func (s *server) buildPresentAccounts(ctx context.Context, accountIDs []string, excludeAccountID string) []wsproto.PresentAccount {
	present := []wsproto.PresentAccount{}
	for _, id := range accountIDs {
		if id == excludeAccountID {
			continue
		}
		account, err := s.store.GetAccount(ctx, id)
		if err != nil {
			continue
		}
		present = append(present, wsproto.PresentAccount{AccountID: account.ID, DisplayName: account.DisplayName})
	}
	return present
}

// --- REST handlers ---

type createAccountRequest struct {
	DisplayName string `json:"displayName"`
}

func (s *server) handleCreateAccount(c *fiber.Ctx) error {
	var req createAccountRequest
	if err := c.BodyParser(&req); err != nil || req.DisplayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "displayName is required"})
	}
	account, err := s.store.CreateAccount(c.Context(), req.DisplayName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(account)
}

func (s *server) handleGetAccount(c *fiber.Ctx) error {
	sync, err := s.stateSync(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "account not found"})
	}
	return c.JSON(sync)
}
