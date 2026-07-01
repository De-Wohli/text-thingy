// Package models defines the core domain types shared across the gateway
// and worker services. These mirror the TypeScript types under
// frontend/src/engine/types.ts — when one changes, update the other.
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

// AbilityModifier applies the SRD formula (floor((score-10)/2)) — Go's
// integer division truncates toward zero, not floor, so this has to handle
// negative scores explicitly to match the rulebook for sub-10 scores.
func AbilityModifier(score int) int {
	diff := score - 10
	mod := diff / 2
	if diff%2 != 0 && diff < 0 {
		mod--
	}
	return mod
}

// ProficiencyBonusForLevel follows the SRD proficiency bonus progression.
func ProficiencyBonusForLevel(level int) int {
	return 2 + (level-1)/4
}

// ArmorClassFor is a simplified SRD-equipment assumption per class: Fighters
// are proficient with all armor and assumed to start in chain mail (AC 16,
// no Dex); Wizards have no armor proficiency, so their AC is unarmored
// (10 + Dex modifier).
func ArmorClassFor(classID ClassID, dexModifier int) int {
	if classID == ClassFighter {
		return 16
	}
	return 10 + dexModifier
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

// LocationID identifies a node in the world's location graph (see
// internal/world). The world is a small hub-and-spoke graph, not a tile
// grid — see outline.md's "Virtual tabletop" implementation note for why.
type LocationID string

type Account struct {
	ID                string     `json:"id"`
	DisplayName       string     `json:"displayName"`
	Honor             int        `json:"honor"`
	Gold              int        `json:"gold"`
	ActiveCharacterID *string    `json:"activeCharacterId"`
	LocationID        LocationID `json:"locationId"`
	PartyID           *string    `json:"partyId"`
}

// Party is a group of accounts who travel and adventure together — formed
// via an invite/accept handshake (see backend/cmd/gateway/party.go), not
// auto-assigned.
type Party struct {
	ID              string    `json:"id"`
	LeaderAccountID string    `json:"leaderAccountId"`
	CreatedAt       time.Time `json:"createdAt"`
}

type ChatChannel string

const (
	ChannelGlobal   ChatChannel = "global"
	ChannelGuild    ChatChannel = "guild"
	ChannelParty    ChatChannel = "party"
	ChannelRP       ChatChannel = "rp"
	ChannelNarrator ChatChannel = "narrator" // server-generated GM flavor text; not a client-writable channel
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

// Skill is one of the SRD skills relevant to the non-combat interactions
// this prototype supports (search for traps, investigate, etc.) — a subset
// of the full 5e skill list, not all eighteen.
type Skill string

const (
	SkillPerception    Skill = "perception"
	SkillInvestigation Skill = "investigation"
	SkillInsight       Skill = "insight"
	SkillStealth       Skill = "stealth"
	SkillArcana        Skill = "arcana"
	SkillAthletics     Skill = "athletics"
)

// SkillAbility maps each skill to the ability score that governs its check,
// per the SRD.
var SkillAbility = map[Skill]string{
	SkillPerception:    "wis",
	SkillInvestigation: "int",
	SkillInsight:       "wis",
	SkillStealth:       "dex",
	SkillArcana:        "int",
	SkillAthletics:     "str",
}

// ClassSkillProficiencies is a simplified fixed proficiency list per class
// (real SRD chargen lets a player choose 2 from a class list; this
// prototype's chargen has no such choice step yet, so each class gets a
// fixed, thematically appropriate pair).
var ClassSkillProficiencies = map[ClassID][]Skill{
	ClassFighter: {SkillAthletics, SkillPerception},
	ClassWizard:  {SkillArcana, SkillInvestigation},
}

func AbilityScoreFor(scores AbilityScores, ability string) int {
	switch ability {
	case "str":
		return scores.Str
	case "dex":
		return scores.Dex
	case "con":
		return scores.Con
	case "int":
		return scores.Int
	case "wis":
		return scores.Wis
	case "cha":
		return scores.Cha
	default:
		return 10
	}
}

func IsProficientInSkill(classID ClassID, skill Skill) bool {
	for _, s := range ClassSkillProficiencies[classID] {
		if s == skill {
			return true
		}
	}
	return false
}

type DungeonRoomType string

const (
	RoomStart    DungeonRoomType = "start"
	RoomHallway  DungeonRoomType = "hallway"
	RoomTreasure DungeonRoomType = "treasure"
	RoomBoss     DungeonRoomType = "boss"
)

// DungeonRoom no longer carries grid coordinates (X/Y/Width/Height) — the
// room-card UI doesn't render a literal grid, so they were vestigial.
// Label and Description carry the room's name and narrative flavor so
// dungeons can have themed rooms rather than generic "corridor/boss" labels.
type DungeonRoom struct {
	Type        DungeonRoomType `json:"type"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Icon        string          `json:"icon"`
	Cleared     bool            `json:"cleared"`
}

type Monster struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	ChallengeRating float64 `json:"challengeRating"`
	ArmorClass      int     `json:"armorClass"`
	HP              int     `json:"hp"`
	AttackBonus     int     `json:"attackBonus"`
	DamageDie       string  `json:"damageDie"`
}

type DungeonEncounter struct {
	RoomType DungeonRoomType `json:"roomType"`
	Monsters []Monster       `json:"monsters"`
}

// Dungeon no longer carries a rendered grid (Width/Height/Grid) — dropped
// along with DungeonRoom's coordinates, see above.
type Dungeon struct {
	ID         string             `json:"id"`
	PartyID    string             `json:"partyId"`
	Rooms      []DungeonRoom      `json:"rooms"`
	Encounters []DungeonEncounter `json:"encounters"`
	Resolved   bool               `json:"resolved"`
}
