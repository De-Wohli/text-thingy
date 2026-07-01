// Package world is the server-authoritative location graph that replaced
// the original tile-grid overworld. A virtual tabletop session run by an
// automated Narrator doesn't need WASD movement between grid squares — a
// real DM just says "you arrive at the tavern." The world is instead a
// small hub-and-spoke graph of named locations; "moving" is choosing a
// connected location to travel to. Keep this in sync with
// frontend/src/data/locations.ts, which mirrors these IDs/names for the
// visual map (see frontend/src/components/WorldMap.tsx).
package world

import "dnd5e-web/backend/internal/models"

type Kind string

const (
	KindHub       Kind = "hub"
	KindGuildHall Kind = "guild_hall"
	KindTavern    Kind = "tavern"
	KindNPC       Kind = "npc"
	KindQuestHook Kind = "quest_hook"
)

type Location struct {
	ID          models.LocationID   `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Kind        Kind                `json:"kind"`
	Connections []models.LocationID `json:"connections"`
}

const (
	TownSquare   models.LocationID = "town_square"
	GuildHall    models.LocationID = "guild_hall"
	Tavern       models.LocationID = "tavern"
	Market       models.LocationID = "market"
	MineEntrance models.LocationID = "mine_entrance"
)

// Locations is a hub-and-spoke graph: every spoke connects back to the
// Town Square hub, and the hub connects to every spoke. There is no
// direct spoke-to-spoke travel, keeping the graph trivial to reason about
// and to render as a map.
var Locations = map[models.LocationID]Location{
	TownSquare: {
		ID:          TownSquare,
		Name:        "Town Square",
		Description: "Cobblestones radiate out from a mossy fountain at the heart of town. Roads lead off toward the Guild Hall, the tavern, the market, and the old road out past the mine.",
		Kind:        KindHub,
		Connections: []models.LocationID{GuildHall, Tavern, Market, MineEntrance},
	},
	GuildHall: {
		ID:          GuildHall,
		Name:        "Adventurer's Guild Hall",
		Description: "A broad stone hall hung with banners and bounty notices. This is where adventurers register, recruit, and swap who's leading the party.",
		Kind:        KindGuildHall,
		Connections: []models.LocationID{TownSquare},
	},
	Tavern: {
		ID:          Tavern,
		Name:        "The Yawning Flask",
		Description: "Lantern light and the smell of stew. A few regulars nurse their drinks in the corner, half-listening for rumors worth repeating.",
		Kind:        KindTavern,
		Connections: []models.LocationID{TownSquare},
	},
	Market: {
		ID:          Market,
		Name:        "Market Square",
		Description: "Stalls and awnings crowd the square. A citizen catches your eye, like they've been waiting for someone to ask the right question.",
		Kind:        KindNPC,
		Connections: []models.LocationID{TownSquare},
	},
	MineEntrance: {
		ID:          MineEntrance,
		Name:        "Old Mine Entrance",
		Description: "A collapsed timber frame marks a shaft sunk into the hillside. Cold air drifts up from the dark — something below hasn't been disturbed in a long time.",
		Kind:        KindQuestHook,
		Connections: []models.LocationID{TownSquare},
	},
}

// CanTravel reports whether `to` is directly reachable from `from`.
func CanTravel(from, to models.LocationID) bool {
	loc, ok := Locations[from]
	if !ok {
		return false
	}
	for _, c := range loc.Connections {
		if c == to {
			return true
		}
	}
	return false
}

// IsValid reports whether id names a real location.
func IsValid(id models.LocationID) bool {
	_, ok := Locations[id]
	return ok
}

// DefaultLocation is where every new account starts.
const DefaultLocation = TownSquare
