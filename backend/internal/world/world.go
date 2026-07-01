// Package world is the server-authoritative location graph that replaced
// the original tile-grid overworld. A virtual tabletop session run by an
// automated Narrator doesn't need WASD movement between grid squares —
// the world is instead a graph of named locations you travel between.
// Keep this in sync with frontend/src/engine/locations.ts, which mirrors
// IDs/names/icons for the visual map (WorldMap.tsx).
package world

import "dnd5e-web/backend/internal/models"

type Kind string

const (
	KindHub       Kind = "hub"
	KindGuildHall Kind = "guild_hall"
	KindTavern    Kind = "tavern"
	KindNPC       Kind = "npc"
	KindQuestHook Kind = "quest_hook"
	KindOutdoor   Kind = "outdoor"
	KindWater     Kind = "water"
	KindBarrier   Kind = "barrier" // can be reached but blocks further travel
)

type Location struct {
	ID          models.LocationID   `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Kind        Kind                `json:"kind"`
	Connections []models.LocationID `json:"connections"`
}

// Location IDs
const (
	TheTown        models.LocationID = "the_town"
	GuildHall      models.LocationID = "guild_hall"
	Tavern         models.LocationID = "tavern"
	Market         models.LocationID = "market"
	NorthFields    models.LocationID = "north_fields"
	WestFields     models.LocationID = "west_fields"
	EastRiver      models.LocationID = "east_river"
	SouthMountains models.LocationID = "south_mountains"
)

// Locations is the world graph.
//
// Layout:
//
//	          North Fields
//	              │
//	West Fields ──┤The Town├── East River
//	              │
//	        South Mountains
//	         (impassable)
//
// Interior town services (Guild Hall, Tavern, Market) are sub-locations
// of The Town hub; the four cardinal directions lead outside.
var Locations = map[models.LocationID]Location{
	TheTown: {
		ID:          TheTown,
		Name:        "The Town",
		Description: "A modest town of stone and timber. The cobbled market square is flanked by the Guild Hall to one side and the Yawning Flask to the other. Four roads lead out: north across open fields, west toward ruined hills, east to the river, and south to mountains that have never been crossed.",
		Kind:        KindHub,
		Connections: []models.LocationID{GuildHall, Tavern, Market, NorthFields, WestFields, EastRiver, SouthMountains},
	},
	GuildHall: {
		ID:          GuildHall,
		Name:        "Adventurer's Guild Hall",
		Description: "A broad stone hall hung with bounty notices and crossed swords. The Guild Clerk eyes you over the counter. This is where adventurers register, recruit, and trade roster slots.",
		Kind:        KindGuildHall,
		Connections: []models.LocationID{TheTown},
	},
	Tavern: {
		ID:          Tavern,
		Name:        "The Yawning Flask",
		Description: "Lantern light and the smell of meat stew. A half-dozen regulars nurse their drinks, half-listening for rumours worth repeating.",
		Kind:        KindTavern,
		Connections: []models.LocationID{TheTown},
	},
	Market: {
		ID:          Market,
		Name:        "Market Square",
		Description: "Stalls and awnings crowd the square. A citizen catches your eye — the kind of look that says they've been waiting for someone reckless enough to ask the right question.",
		Kind:        KindNPC,
		Connections: []models.LocationID{TheTown},
	},
	NorthFields: {
		ID:          NorthFields,
		Name:        "Northern Fields",
		Description: "Flat farmland stretches north, broken by hedgerows and the occasional scarecrow. In the distance, a shepherd's hut sits on a gentle rise. Quiet out here — perhaps too quiet.",
		Kind:        KindOutdoor,
		Connections: []models.LocationID{TheTown},
	},
	WestFields: {
		ID:          WestFields,
		Name:        "Western Fields",
		Description: "Rolling grassland gives way to a low hill crowned with crumbling stone. The ruins of an old keep squat at the summit — its gate long since collapsed, its lower levels still intact and unaccounted for.",
		Kind:        KindQuestHook,
		Connections: []models.LocationID{TheTown},
	},
	EastRiver: {
		ID:          EastRiver,
		Name:        "Eastern River",
		Description: "The Silven runs fast here, grey-green and cold. Willows lean over the near bank; across the water the far shore is thickly wooded. A broken ferry platform creaks at its mooring. No way across today — but the fishing is good.",
		Kind:        KindWater,
		Connections: []models.LocationID{TheTown},
	},
	SouthMountains: {
		ID:          SouthMountains,
		Name:        "Southern Mountains",
		Description: "You walk south until the cobblestones run out and the road becomes a goat track, and then the goat track gives up entirely. A wall of sheer granite rises before you — the mountains have never let anyone through. The cold air tastes of snow. You turn back.",
		Kind:        KindBarrier,
		Connections: []models.LocationID{TheTown}, // only connection is back to town
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
const DefaultLocation = TheTown
