// Package dungeon implements the Procedural Dungeon Generation Engine: a
// 15x15 grid split into start/hallway/treasure/boss rooms, populated with
// SRD-derived monsters sized to the active character's level via an XP
// budget (the "CR Budget Math" requirement).
package dungeon

import (
	"math/rand"

	"dnd5e-web/backend/internal/models"
)

const Size = models.DungeonSize

// buildRooms splits the grid into four quadrant-rooms. A simple
// deterministic split rather than full cellular automata — sufficient for
// the prototype while leaving room for a richer generator later.
func buildRooms() []models.DungeonRoom {
	half := Size / 2
	return []models.DungeonRoom{
		{Type: models.RoomStart, X: 1, Y: 1, Width: half - 2, Height: half - 2, Cleared: true},
		{Type: models.RoomHallway, X: half - 1, Y: 1, Width: 2, Height: Size - 2, Cleared: false},
		{Type: models.RoomTreasure, X: 1, Y: half + 1, Width: half - 2, Height: half - 2, Cleared: false},
		{Type: models.RoomBoss, X: half + 1, Y: half + 1, Width: half - 2, Height: half - 2, Cleared: false},
	}
}

func carveGrid(rooms []models.DungeonRoom) [][]string {
	grid := make([][]string, Size)
	for y := range grid {
		grid[y] = make([]string, Size)
		for x := range grid[y] {
			grid[y][x] = "wall"
		}
	}
	for _, room := range rooms {
		for y := room.Y; y < room.Y+room.Height && y < Size; y++ {
			for x := room.X; x < room.X+room.Width && x < Size; x++ {
				grid[y][x] = "floor"
			}
		}
	}
	return grid
}

func pickEncounterForLevel1(roomType models.DungeonRoomType) []models.Monster {
	if roomType == models.RoomStart {
		return []models.Monster{}
	}
	budget := mediumXPBudget
	if roomType == models.RoomHallway {
		budget = easyXPBudget
	}

	encounter := []models.Monster{}
	spent := 0
	for spent < budget {
		monster := SRDLowCRMonsters[rand.Intn(len(SRDLowCRMonsters))]
		cost := xpByCR[monster.ChallengeRating]
		if cost == 0 {
			cost = 25
		}
		if spent+cost > budget {
			break
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
	grid := carveGrid(rooms)

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
		Width:      Size,
		Height:     Size,
		Grid:       grid,
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
