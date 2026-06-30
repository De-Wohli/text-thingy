// Package wsproto defines the JSON message protocol exchanged between the
// gateway and connected clients over WebSockets. Keep this in sync with
// frontend/src/ws/protocol.ts.
package wsproto

import (
	"encoding/json"

	"dnd5e-web/backend/internal/combat"
	"dnd5e-web/backend/internal/models"
)

// Envelope is the generic shape every inbound client message arrives in;
// Payload is decoded into a concrete *Payload struct once Type is known.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// --- Inbound (client -> gateway) payloads ---

type MovePayload struct {
	DX int `json:"dx"`
	DY int `json:"dy"`
}

type SwapCharacterPayload struct {
	CharacterID string `json:"characterId"`
}

type CreateCharacterPayload struct {
	Name    string         `json:"name"`
	RaceID  models.RaceID  `json:"raceId"`
	ClassID models.ClassID `json:"classId"`
}

type ChatPayload struct {
	Channel models.ChatChannel `json:"channel"`
	Body    string             `json:"body"`
}

type MakeChoicePayload struct {
	PromptID string `json:"promptId"`
	OptionID string `json:"optionId"`
}

type CastVotePayload struct {
	PromptID string `json:"promptId"`
	OptionID string `json:"optionId"`
}

type ClearDungeonRoomPayload struct {
	RoomType models.DungeonRoomType `json:"roomType"`
}

// --- Outbound (gateway -> client) messages ---
// Each has its own Type so the frontend can discriminate on receipt.

type StateSync struct {
	Type       string             `json:"type"`
	Account    models.Account     `json:"account"`
	Characters []models.Character `json:"characters"`
}

func NewStateSync(account models.Account, characters []models.Character) StateSync {
	return StateSync{Type: "STATE_SYNC", Account: account, Characters: characters}
}

type ChatBroadcast struct {
	Type    string             `json:"type"`
	Message models.ChatMessage `json:"message"`
}

func NewChatBroadcast(msg models.ChatMessage) ChatBroadcast {
	return ChatBroadcast{Type: "CHAT_MESSAGE", Message: msg}
}

type ChoiceState struct {
	Type      string                `json:"type"`
	PromptID  string                `json:"promptId"`
	Prompt    string                `json:"prompt"`
	Mode      models.ChoiceMode     `json:"mode"`
	Options   []models.ChoiceOption `json:"options"`
	Deadline  *int64                `json:"deadline,omitempty"` // unix millis, party mode only
	Narration string                `json:"narration,omitempty"`
}

type VoteUpdate struct {
	Type     string         `json:"type"`
	PromptID string         `json:"promptId"`
	Tallies  map[string]int `json:"tallies"`
}

type VoteResolved struct {
	Type       string `json:"type"`
	PromptID   string `json:"promptId"`
	OptionID   string `json:"optionId"`
	HonorDelta int    `json:"honorDelta"`
	NewHonor   int    `json:"newHonor"`
	TieBreak   bool   `json:"tieBreak"`
	Narration  string `json:"narration,omitempty"`
}

type DungeonReady struct {
	Type      string         `json:"type"`
	Dungeon   models.Dungeon `json:"dungeon"`
	Narration string         `json:"narration,omitempty"`
}

func NewDungeonReady(d models.Dungeon, narration string) DungeonReady {
	return DungeonReady{Type: "DUNGEON_READY", Dungeon: d, Narration: narration}
}

// RoomResolved is sent after CLEAR_DUNGEON_ROOM actually fights the
// encounter (see internal/combat) — it carries the full attack-by-attack
// log so the client can render a combat log, not just a before/after state.
type RoomResolved struct {
	Type      string                 `json:"type"`
	RoomType  models.DungeonRoomType `json:"roomType"`
	Victory   bool                   `json:"victory"`
	CombatLog []combat.AttackRoll    `json:"combatLog"`
	Narration string                 `json:"narration"`
	Dungeon   models.Dungeon         `json:"dungeon"`
}

// DungeonResolved tells the client the instance is fully cleared and it's
// safe to close the dungeon view and return to the overworld.
type DungeonResolved struct {
	Type        string `json:"type"`
	Narration   string `json:"narration"`
	GoldAwarded int    `json:"goldAwarded"`
}

type ErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewError(message string) ErrorMessage {
	return ErrorMessage{Type: "ERROR", Message: message}
}

// --- Worker -> Gateway events (published over Redis, not sent to clients
// directly) ---
//
// These are self-contained: the worker includes everything the gateway
// needs to route the result (which account(s)/party to notify) so the
// gateway doesn't have to keep its own bookkeeping of in-flight jobs.

type DungeonReadyEvent struct {
	JobID     string         `json:"jobId"`
	AccountID string         `json:"accountId"`
	PartyID   string         `json:"partyId,omitempty"` // empty when solo
	Dungeon   models.Dungeon `json:"dungeon"`
}

type VoteResolvedResult struct {
	AccountID  string `json:"accountId"`
	HonorDelta int    `json:"honorDelta"`
	NewHonor   int    `json:"newHonor"`
}

type VoteResolvedEvent struct {
	PromptID string               `json:"promptId"`
	OptionID string               `json:"optionId"`
	TieBreak bool                 `json:"tieBreak"`
	Results  []VoteResolvedResult `json:"results"`
}
