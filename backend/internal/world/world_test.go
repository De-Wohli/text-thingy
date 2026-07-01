package world

import (
	"testing"

	"dnd5e-web/backend/internal/models"
)

func TestHubConnectsToEverySpokeAndBack(t *testing.T) {
	hub := Locations[TownSquare]
	spokes := []models.LocationID{GuildHall, Tavern, Market, MineEntrance}
	if len(hub.Connections) != len(spokes) {
		t.Fatalf("expected hub to connect to %d spokes, got %d", len(spokes), len(hub.Connections))
	}
	for _, spoke := range spokes {
		if !CanTravel(TownSquare, spoke) {
			t.Errorf("expected to be able to travel from hub to %s", spoke)
		}
		if !CanTravel(spoke, TownSquare) {
			t.Errorf("expected to be able to travel from %s back to hub", spoke)
		}
	}
}

func TestSpokesAreNotDirectlyConnected(t *testing.T) {
	if CanTravel(GuildHall, Tavern) {
		t.Error("expected no direct travel between spokes (must pass through the hub)")
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
	if CanTravel("nowhere", TownSquare) {
		t.Error("expected CanTravel from an unknown location to be false")
	}
}
