// Package dungeon generates dungeon instances: a sequence of themed rooms
// rendered as a room-card track in the client (see DungeonView.tsx),
// populated with SRD-derived monsters sized by an XP budget.
package dungeon

import (
	"math/rand"

	"dnd5e-web/backend/internal/models"
)

// roomSpec defines the fixed properties of one room in the keep dungeon.
type roomSpec struct {
	rType       models.DungeonRoomType
	label       string
	description string
	icon        string
}

// keepRooms is the fixed room sequence for the ruined keep on the western
// fields. The functional type (start/hallway/treasure/boss) drives the
// monster XP budget; the label/description/icon drive the UI presentation.
var keepRooms = []roomSpec{
	{
		rType:       models.RoomStart,
		label:       "Broken Gateway",
		description: "The keep's gate lies in rubble. Cold air drifts out of the dark beyond. Whatever happened here, it was a long time ago.",
		icon:        "🚪",
	},
	{
		rType:       models.RoomHallway,
		label:       "Outer Court",
		description: "A courtyard choked with weeds and debris. Something moves in the shadows near the far wall.",
		icon:        "🌳",
	},
	{
		rType:       models.RoomHallway,
		label:       "Guard Barracks",
		description: "Overturned bunks and rusted armour. Whoever occupied this place is long dead — or maybe not.",
		icon:        "🛏",
	},
	{
		rType:       models.RoomTreasure,
		label:       "The Chapel",
		description: "Cracked stone pews face a defaced altar. The air smells of old incense and something fouler underneath.",
		icon:        "⛪",
	},
	{
		rType:       models.RoomTreasure,
		label:       "Undercroft",
		description: "Low arches, piled crates, the glint of coin in the torchlight. Whatever was stored here is still guarded.",
		icon:        "🪙",
	},
	{
		rType:       models.RoomBoss,
		label:       "The Warlord's Chamber",
		description: "A high-ceilinged hall with a cracked throne at one end. Whatever rules this keep meets you here.",
		icon:        "👑",
	},
}

func buildRooms() []models.DungeonRoom {
	rooms := make([]models.DungeonRoom, 0, len(keepRooms))
	for i, spec := range keepRooms {
		rooms = append(rooms, models.DungeonRoom{
			Type:        spec.rType,
			Label:       spec.label,
			Description: spec.description,
			Icon:        spec.icon,
			Cleared:     i == 0, // entrance pre-cleared
		})
	}
	return rooms
}

// pickEncounterForLevel1 always returns at least one monster for
// hallway/treasure rooms — picking a random monster and immediately
// breaking the moment it doesn't fit the remaining budget (the original
// approach) could land on an over-budget monster on the very first draw
// and silently return an empty encounter. An "empty fight" used to be
// harmless when clearing a room was just a flag flip; now that
// internal/combat actually resolves it, an empty monster list produces a
// nil combat log, which serializes to JSON `null` and crashes the
// frontend's .map() over it.
func pickEncounterForLevel1(roomType models.DungeonRoomType) []models.Monster {
	if roomType == models.RoomStart {
		return []models.Monster{}
	}
	budget := mediumXPBudget
	if roomType == models.RoomHallway {
		budget = easyXPBudget
	}

	affordable := make([]models.Monster, 0, len(SRDLowCRMonsters))
	for _, m := range SRDLowCRMonsters {
		if xpByCR[m.ChallengeRating] <= budget {
			affordable = append(affordable, m)
		}
	}
	if len(affordable) == 0 {
		affordable = SRDLowCRMonsters
	}

	first := affordable[rand.Intn(len(affordable))]
	encounter := []models.Monster{first}
	spent := xpByCR[first.ChallengeRating]

	for attempt := 0; attempt < 20 && spent < budget; attempt++ {
		monster := SRDLowCRMonsters[rand.Intn(len(SRDLowCRMonsters))]
		cost := xpByCR[monster.ChallengeRating]
		if cost == 0 {
			cost = 25
		}
		if spent+cost > budget {
			continue
		}
		encounter = append(encounter, monster)
		spent += cost
	}
	return encounter
}

func pickBossEncounter() []models.Monster {
	return []models.Monster{BossMonsters[rand.Intn(len(BossMonsters))]}
}

// Generate produces a fresh dungeon instance for the given party and
// character level (level-1 monster dataset used throughout Phase 1).
func Generate(id, partyID string, characterLevel int) models.Dungeon {
	rooms := buildRooms()

	encounters := make([]models.DungeonEncounter, 0, len(rooms))
	for _, room := range rooms {
		var monsters []models.Monster
		if room.Type == models.RoomBoss {
			monsters = pickBossEncounter()
		} else {
			monsters = pickEncounterForLevel1(room.Type)
		}
		encounters = append(encounters, models.DungeonEncounter{RoomType: room.Type, Monsters: monsters})
	}

	return models.Dungeon{
		ID:         id,
		PartyID:    partyID,
		Rooms:      rooms,
		Encounters: encounters,
		Resolved:   false,
	}
}

// ClearRoom marks a room cleared and recomputes whether the whole dungeon
// is resolved (only true once the boss room falls).
func ClearRoom(d models.Dungeon, roomType models.DungeonRoomType) models.Dungeon {
	rooms := make([]models.DungeonRoom, len(d.Rooms))
	resolved := false
	for i, room := range d.Rooms {
		if room.Type == roomType {
			room.Cleared = true
		}
		rooms[i] = room
		if room.Type == models.RoomBoss && room.Cleared {
			resolved = true
		}
	}
	d.Rooms = rooms
	d.Resolved = resolved
	return d
}
