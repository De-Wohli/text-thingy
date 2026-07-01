// Package dungeon implements the Procedural Dungeon Generation Engine: a
// fixed start/hallway/treasure/boss room sequence (rendered by the client
// as a room-card track, not a literal grid — see frontend's DungeonView),
// populated with SRD-derived monsters sized to the active character's
// level via an XP budget (the "CR Budget Math" requirement).
package dungeon

import (
	"math/rand"

	"dnd5e-web/backend/internal/models"
)

func buildRooms() []models.DungeonRoom {
	return []models.DungeonRoom{
		{Type: models.RoomStart, Cleared: true},
		{Type: models.RoomHallway, Cleared: false},
		{Type: models.RoomTreasure, Cleared: false},
		{Type: models.RoomBoss, Cleared: false},
	}
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

// Generate produces a fresh dungeon instance for the given party and active
// character level. Phase 1 only ships a level-1 monster dataset; any other
// level still resolves against it.
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
