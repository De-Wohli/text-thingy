import { useEffect, useMemo } from 'react'
import { useGame } from '../state/GameProvider'
import { buildOverworldGrid } from '../data/overworld'
import type { TileType } from '../engine/types'

// Single-character glyphs keep every grid cell the same width so rows stay
// aligned in the monospace <pre>; the bracketed [A]/[T]/[N]/[?] notation from
// the design doc is shown in the legend instead (see MapLegend).
const TILE_GLYPH: Record<TileType, string> = {
  wall: '#',
  floor: '.',
  water: '~',
  guild: 'A',
  tavern: 'T',
  npc: 'N',
  poi: '?',
}

const KEY_TO_DELTA: Record<string, { dx: number; dy: number }> = {
  w: { dx: 0, dy: -1 },
  ArrowUp: { dx: 0, dy: -1 },
  s: { dx: 0, dy: 1 },
  ArrowDown: { dx: 0, dy: 1 },
  a: { dx: -1, dy: 0 },
  ArrowLeft: { dx: -1, dy: 0 },
  d: { dx: 1, dy: 0 },
  ArrowRight: { dx: 1, dy: 0 },
}

export function OverworldCanvas() {
  const { state, actions } = useGame()
  const grid = useMemo(() => buildOverworldGrid(), [])
  const coordinate = state.account?.coordinate

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      const delta = KEY_TO_DELTA[e.key]
      if (!delta) return
      e.preventDefault()
      actions.move(delta.dx, delta.dy)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [actions])

  return (
    <div
      id="overworld-screen"
      aria-label="Overworld map, use WASD or arrow keys to move"
      className="bg-[#0c0a08] border-2 border-accent rounded overflow-x-auto p-4"
    >
      <pre className="m-0 text-sm leading-tight tracking-wide text-[#8fbf8f]">
        {grid
          .map((row, y) =>
            row
              .map((tile, x) =>
                coordinate && x === coordinate.x && y === coordinate.y ? '@' : TILE_GLYPH[tile],
              )
              .join(''),
          )
          .join('\n')}
      </pre>
    </div>
  )
}
