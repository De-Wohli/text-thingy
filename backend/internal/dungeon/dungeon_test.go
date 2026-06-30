package dungeon

import (
	"testing"

	"dnd5e-web/backend/internal/models"
)

func TestGenerateProducesAllRoomTypes(t *testing.T) {
	d := Generate("d1", "party-1", 1)
	if d.Width != Size || d.Height != Size {
		t.Fatalf("expected %dx%d grid, got %dx%d", Size, Size, d.Width, d.Height)
	}
	want := map[models.DungeonRoomType]bool{
		models.RoomStart: true, models.RoomHallway: true, models.RoomTreasure: true, models.RoomBoss: true,
	}
	for _, room := range d.Rooms {
		delete(want, room.Type)
	}
	if len(want) != 0 {
		t.Fatalf("missing room types: %v", want)
	}
}

func TestStartRoomHasNoEncounter(t *testing.T) {
	d := Generate("d2", "party-1", 1)
	for _, e := range d.Encounters {
		if e.RoomType == models.RoomStart && len(e.Monsters) != 0 {
			t.Fatalf("expected start room to be encounter-free, got %d monsters", len(e.Monsters))
		}
	}
}

func TestBossRoomUsesBossTierMonsters(t *testing.T) {
	d := Generate("d3", "party-1", 1)
	for _, e := range d.Encounters {
		if e.RoomType != models.RoomBoss {
			continue
		}
		if len(e.Monsters) == 0 {
			t.Fatal("expected at least one boss monster")
		}
		for _, m := range e.Monsters {
			// Boss-tier monsters are tuned below their real SRD CR for solo
			// play (see BossMonsters' doc comment) but should still be
			// tougher than the regular low-CR encounter pool.
			if m.ChallengeRating < 0.5 {
				t.Fatalf("boss monster %s has CR %.2f, expected >= 0.5", m.Name, m.ChallengeRating)
			}
		}
	}
}

func TestHallwayAndTreasureRoomsNeverEmpty(t *testing.T) {
	// Regression: the original generator picked one random monster and
	// immediately bailed if it didn't fit the budget, which could leave
	// hallway/treasure with zero monsters ~1/3 of the time. Run many
	// generations to catch that probabilistically.
	for i := 0; i < 200; i++ {
		d := Generate("d5", "party-1", 1)
		for _, e := range d.Encounters {
			if e.RoomType == models.RoomHallway || e.RoomType == models.RoomTreasure {
				if len(e.Monsters) == 0 {
					t.Fatalf("%s room had zero monsters on iteration %d", e.RoomType, i)
				}
			}
		}
	}
}

func TestClearRoomResolvesOnlyAfterBoss(t *testing.T) {
	d := Generate("d4", "party-1", 1)
	d = ClearRoom(d, models.RoomTreasure)
	if d.Resolved {
		t.Fatal("expected dungeon unresolved before boss room is cleared")
	}
	d = ClearRoom(d, models.RoomBoss)
	if !d.Resolved {
		t.Fatal("expected dungeon resolved after boss room is cleared")
	}
}
