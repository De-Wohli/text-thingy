package world

import (
	"testing"

	"dnd5e-web/backend/internal/models"
)

func TestHubConnectsToAllLocations(t *testing.T) {
	hub := Locations[TheTown]
	// hub connects to 3 interior sub-locations + 4 cardinal exterior locations
	if len(hub.Connections) < 7 {
		t.Fatalf("expected hub to connect to at least 7 locations, got %d", len(hub.Connections))
	}
	for _, conn := range hub.Connections {
		if !CanTravel(TheTown, conn) {
			t.Errorf("expected to be able to travel from hub to %s", conn)
		}
		if !CanTravel(conn, TheTown) {
			t.Errorf("expected to be able to travel from %s back to hub", conn)
		}
	}
}

func TestSpokesAreNotDirectlyConnected(t *testing.T) {
	if CanTravel(GuildHall, Tavern) {
		t.Error("expected no direct travel between spokes (must pass through the hub)")
	}
	if CanTravel(NorthFields, WestFields) {
		t.Error("expected no direct travel between exterior locations (must pass through town)")
	}
}

func TestExteriorLocationsExist(t *testing.T) {
	for _, id := range []models.LocationID{NorthFields, WestFields, EastRiver, SouthMountains} {
		if !IsValid(id) {
			t.Errorf("exterior location %s should be valid", id)
		}
	}
}

func TestSouthMountainsHasNorthwardConnectionOnly(t *testing.T) {
	mountains := Locations[SouthMountains]
	if len(mountains.Connections) == 0 {
		t.Fatal("SouthMountains should at least connect back to the town")
	}
	for _, conn := range mountains.Connections {
		if conn != TheTown {
			t.Errorf("SouthMountains should only connect back to the town, found connection to %s", conn)
		}
	}
}

func TestWestFieldsIsDungeonTrigger(t *testing.T) {
	loc := Locations[WestFields]
	if loc.Kind != KindQuestHook {
		t.Errorf("WestFields should be a quest hook (dungeon trigger), got kind %s", loc.Kind)
	}
}

func TestEveryLocationIsValidAndReachable(t *testing.T) {
	for id, loc := range Locations {
		if !IsValid(id) {
			t.Errorf("location %s should be valid", id)
		}
		for _, conn := range loc.Connections {
			if !IsValid(conn) {
				t.Errorf("location %s connects to unknown location %s", id, conn)
			}
		}
	}
}

func TestDefaultLocationIsValid(t *testing.T) {
	if !IsValid(DefaultLocation) {
		t.Fatal("DefaultLocation must be a valid location")
	}
}

func TestCanTravelFromUnknownLocationIsFalse(t *testing.T) {
	if CanTravel("nowhere", TheTown) {
		t.Error("expected CanTravel from an unknown location to be false")
	}
}
