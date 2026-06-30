// Package worldmap is the server-authoritative mirror of the overworld
// layout in frontend/src/data/overworld.ts. The RAW_MAP string must stay
// identical between the two so landmark coordinates (guild hall, NPC, POI)
// never drift out of sync between client and server.
package worldmap

import "dnd5e-web/backend/internal/models"

// Every row must be exactly the same length (61 chars) — the frontend
// renders this as a CSS grid, where a ragged row silently shifts every tile
// after it by one column.
var rawMap = []string{
	"#############################################################",
	"#...........................................................#",
	"#...A.......................................................#",
	"#...........................................................#",
	"#..............T............................................#",
	"#...........................................................#",
	"#...........................................................#",
	"#..........@................................................#",
	"#...........................................................#",
	"#.....................................N.....................#",
	"#...........................................................#",
	"#...........................................................#",
	"#~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~?~~#",
	"#############################################################",
}

func findRune(target rune) models.Coordinate {
	for y, row := range rawMap {
		for x, ch := range row {
			if ch == target {
				return models.Coordinate{X: x, Y: y}
			}
		}
	}
	return models.Coordinate{X: 1, Y: 1}
}

var (
	GuildHall = findRune('A')
	Tavern    = findRune('T')
	NPC       = findRune('N')
	POI       = findRune('?')
	Start     = findRune('@')
)

func IsAdjacentOrEqual(a, b models.Coordinate) bool {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx <= 1 && dy <= 1
}

func tileAt(coord models.Coordinate) byte {
	if coord.Y < 0 || coord.Y >= len(rawMap) {
		return '#'
	}
	row := rawMap[coord.Y]
	if coord.X < 0 || coord.X >= len(row) {
		return '#'
	}
	return row[coord.X]
}

func IsWalkable(coord models.Coordinate) bool {
	tile := tileAt(coord)
	return tile != '#' && tile != '~'
}
