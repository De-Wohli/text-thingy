// Package models defines the core domain types shared across the gateway
// and worker services. These mirror the TypeScript types under
// frontend/src/types.ts — when one changes, update the other.
package models

import "time"

type RaceID string

const (
	RaceHuman    RaceID = "human"
	RaceTiefling RaceID = "tiefling"
)

type ClassID string

const (
	ClassFighter ClassID = "fighter"
	ClassWizard  ClassID = "wizard"
)

type AbilityScores struct {
	Str int `json:"str"`
	Dex int `json:"dex"`
	Con int `json:"con"`
	Int int `json:"int"`
	Wis int `json:"wis"`
	Cha int `json:"cha"`
}

type Race struct {
	ID             RaceID         `json:"id"`
	Name           string         `json:"name"`
	AbilityBonuses map[string]int `json:"abilityBonuses"`
	Traits         []string       `json:"traits"`
}

type Class struct {
	ID                 ClassID     `json:"id"`
	Name               string      `json:"name"`
	HitDie             int         `json:"hitDie"`
	Proficiencies      []string    `json:"proficiencies"`
	Features           []string    `json:"features"`
	CantripsKnown      int         `json:"cantripsKnown,omitempty"`
	StartingSpellSlots map[int]int `json:"startingSpellSlots,omitempty"`
}

var Races = map[RaceID]Race{
	RaceHuman: {
		ID:             RaceHuman,
		Name:           "Human",
		AbilityBonuses: map[string]int{"str": 1, "dex": 1, "con": 1, "int": 1, "wis": 1, "cha": 1},
		Traits:         []string{"Versatile: +1 to all ability scores"},
	},
	RaceTiefling: {
		ID:             RaceTiefling,
		Name:           "Tiefling",
		AbilityBonuses: map[string]int{"cha": 2, "int": 1},
		Traits:         []string{"Darkvision (60ft)", "Hellish Resistance (fire damage resistance)"},
	},
}

var Classes = map[ClassID]Class{
	ClassFighter: {
		ID:            ClassFighter,
		Name:          "Fighter",
		HitDie:        10,
		Proficiencies: []string{"All armor", "Shields", "Martial weapons"},
		Features:      []string{"Second Wind"},
	},
	ClassWizard: {
		ID:                 ClassWizard,
		Name:               "Wizard",
		HitDie:             6,
		Proficiencies:      []string{"Daggers", "Darts", "Slings", "Quarterstaffs"},
		Features:           []string{"Arcane Recovery", "Spellcasting"},
		CantripsKnown:      3,
		StartingSpellSlots: map[int]int{1: 2},
	},
}

type CharacterStatus string

const (
	StatusIdle     CharacterStatus = "IDLE"
	StatusQuesting CharacterStatus = "QUESTING"
	StatusCrafting CharacterStatus = "CRAFTING"
)

type Character struct {
	ID            string          `json:"id"`
	AccountID     string          `json:"accountId"`
	Name          string          `json:"name"`
	RaceID        RaceID          `json:"raceId"`
	ClassID       ClassID         `json:"classId"`
	Level         int             `json:"level"`
	Status        CharacterStatus `json:"status"`
	HPCurrent     int             `json:"hpCurrent"`
	HPMax         int             `json:"hpMax"`
	AbilityScores AbilityScores   `json:"abilityScores"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Account struct {
	ID                string     `json:"id"`
	DisplayName       string     `json:"displayName"`
	Honor             int        `json:"honor"`
	Gold              int        `json:"gold"`
	ActiveCharacterID *string    `json:"activeCharacterId"`
	Coordinate        Coordinate `json:"coordinate"`
	PartyID           *string    `json:"partyId"`
}

type ChatChannel string

const (
	ChannelGlobal ChatChannel = "global"
	ChannelGuild  ChatChannel = "guild"
	ChannelParty  ChatChannel = "party"
	ChannelRP     ChatChannel = "rp"
)

type ChatMessage struct {
	Channel   ChatChannel `json:"channel"`
	AccountID string      `json:"accountId"`
	Name      string      `json:"name,omitempty"`
	Race      string      `json:"race,omitempty"`
	Class     string      `json:"class,omitempty"`
	Body      string      `json:"body"`
	Timestamp time.Time   `json:"timestamp"`
}

// ChoiceTypology drives the Honor & Alignment Impact Matrix: every choice or
// vote outcome is tagged with one of these three buckets.
type ChoiceTypology string

const (
	ChoiceMerciful  ChoiceTypology = "merciful"  // +10 Honor
	ChoicePragmatic ChoiceTypology = "pragmatic" // +0 Honor
	ChoiceRuthless  ChoiceTypology = "ruthless"  // -10 Honor

)

var HonorImpact = map[ChoiceTypology]int{
	ChoiceMerciful:  10,
	ChoicePragmatic: 0,
	ChoiceRuthless:  -10,
}

type ChoiceOption struct {
	ID       string         `json:"id"`
	Label    string         `json:"label"`
	Typology ChoiceTypology `json:"typology"`
}

type ChoiceMode string

const (
	ChoiceModeSolo  ChoiceMode = "solo"
	ChoiceModeParty ChoiceMode = "party"
)

type DungeonRoomType string

const (
	RoomStart    DungeonRoomType = "start"
	RoomHallway  DungeonRoomType = "hallway"
	RoomTreasure DungeonRoomType = "treasure"
	RoomBoss     DungeonRoomType = "boss"
)

type DungeonRoom struct {
	Type    DungeonRoomType `json:"type"`
	X       int             `json:"x"`
	Y       int             `json:"y"`
	Width   int             `json:"width"`
	Height  int             `json:"height"`
	Cleared bool            `json:"cleared"`
}

type Monster struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	ChallengeRating float64 `json:"challengeRating"`
	HP              int     `json:"hp"`
	AttackBonus     int     `json:"attackBonus"`
	DamageDie       string  `json:"damageDie"`
}

type DungeonEncounter struct {
	RoomType DungeonRoomType `json:"roomType"`
	Monsters []Monster       `json:"monsters"`
}

const DungeonSize = 15

type Dungeon struct {
	ID         string             `json:"id"`
	PartyID    string             `json:"partyId"`
	Width      int                `json:"width"`
	Height     int                `json:"height"`
	Grid       [][]string         `json:"grid"`
	Rooms      []DungeonRoom      `json:"rooms"`
	Encounters []DungeonEncounter `json:"encounters"`
	Resolved   bool               `json:"resolved"`
}
