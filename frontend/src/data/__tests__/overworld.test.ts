import { describe, expect, it } from 'vitest'
import { buildOverworldGrid, findStartCoordinate, findTilesOfType, isAdjacentOrEqual, isWalkable } from '../overworld'

describe('overworld map', () => {
  it('places exactly one guild hall, tavern, and POI', () => {
    const grid = buildOverworldGrid()
    expect(findTilesOfType(grid, 'guild')).toHaveLength(1)
    expect(findTilesOfType(grid, 'tavern')).toHaveLength(1)
    expect(findTilesOfType(grid, 'poi')).toHaveLength(1)
  })

  it('finds the player start coordinate on floor', () => {
    const grid = buildOverworldGrid()
    const start = findStartCoordinate()
    expect(isWalkable(grid, start)).toBe(true)
  })

  it('treats walls and water as non-walkable', () => {
    const grid = buildOverworldGrid()
    expect(isWalkable(grid, { x: 0, y: 0 })).toBe(false)
  })

  it('detects adjacency within one tile in any direction', () => {
    expect(isAdjacentOrEqual({ x: 4, y: 4 }, { x: 5, y: 5 })).toBe(true)
    expect(isAdjacentOrEqual({ x: 4, y: 4 }, { x: 4, y: 4 })).toBe(true)
    expect(isAdjacentOrEqual({ x: 4, y: 4 }, { x: 6, y: 4 })).toBe(false)
  })
})
