import type { Coordinate, TileType } from '../engine/types'

// Legend: # wall, . floor, ~ water, A guild hall, T tavern, N npc, ? POI
// Every row must be exactly the same length (61 chars) — this is rendered as
// a CSS grid in OverworldCanvas, so a ragged row silently shifts every tile
// after it by one column. Keep this in sync with backend/internal/worldmap.
const RAW_MAP = [
  '#############################################################',
  '#...........................................................#',
  '#...A.......................................................#',
  '#...........................................................#',
  '#..............T............................................#',
  '#...........................................................#',
  '#...........................................................#',
  '#..........@................................................#',
  '#...........................................................#',
  '#.....................................N.....................#',
  '#...........................................................#',
  '#...........................................................#',
  '#~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~?~~#',
  '#############################################################',
]

const CHAR_TO_TILE: Record<string, TileType> = {
  '#': 'wall',
  '.': 'floor',
  '~': 'water',
  A: 'guild',
  T: 'tavern',
  N: 'npc',
  '?': 'poi',
  '@': 'floor',
}

export const MAP_WIDTH = RAW_MAP[0].length
export const MAP_HEIGHT = RAW_MAP.length

export function buildOverworldGrid(): TileType[][] {
  return RAW_MAP.map((row) => row.split('').map((ch) => CHAR_TO_TILE[ch] ?? 'floor'))
}

export function findStartCoordinate(): Coordinate {
  for (let y = 0; y < RAW_MAP.length; y++) {
    const x = RAW_MAP[y].indexOf('@')
    if (x !== -1) return { x, y }
  }
  return { x: 1, y: 1 }
}

export function findTilesOfType(grid: TileType[][], type: TileType): Coordinate[] {
  const coords: Coordinate[] = []
  for (let y = 0; y < grid.length; y++) {
    for (let x = 0; x < grid[y].length; x++) {
      if (grid[y][x] === type) coords.push({ x, y })
    }
  }
  return coords
}

export function isWalkable(grid: TileType[][], coord: Coordinate): boolean {
  const row = grid[coord.y]
  if (!row) return false
  const tile = row[coord.x]
  if (!tile) return false
  return tile !== 'wall' && tile !== 'water'
}

export function isAdjacentOrEqual(a: Coordinate, b: Coordinate): boolean {
  return Math.abs(a.x - b.x) <= 1 && Math.abs(a.y - b.y) <= 1
}

// Computed once and shared between the reducer and the UI so both agree on
// where the guild hall / NPC / point-of-interest sit on the map.
const landmarkGrid = buildOverworldGrid()
export const LANDMARKS = {
  guildHall: findTilesOfType(landmarkGrid, 'guild')[0],
  npc: findTilesOfType(landmarkGrid, 'npc')[0],
  poi: findTilesOfType(landmarkGrid, 'poi')[0],
}
