import { useEffect, useMemo } from 'react'
import { useGame } from '../state/GameProvider'
import { buildOverworldGrid, MAP_WIDTH } from '../data/overworld'
import type { TileType } from '../engine/types'
import { DirectionPad } from './DirectionPad'

// A small visual vocabulary per tile type — color + icon — so the overworld
// reads as a tabletop game board (think: a printed map with tokens on it)
// rather than a terminal/ASCII dungeon-crawler grid.
const TILE_STYLE: Record<TileType, { bg: string; icon?: string; title: string }> = {
  wall: { bg: 'bg-[#241f17]', title: 'Wall' },
  floor: { bg: 'bg-[#3a3122]', title: 'Ground' },
  water: { bg: 'bg-[#2b4a64]', icon: '~', title: 'River (impassable)' },
  guild: { bg: 'bg-accent', icon: '\u{1F3DB}\u{FE0F}', title: "Adventurer's Guild Hall" },
  tavern: { bg: 'bg-[#8a5a2b]', icon: '\u{1F37A}', title: 'The Yawning Flask Tavern' },
  npc: { bg: 'bg-[#6b3fa0]', icon: '\u{1F9D9}', title: 'Citizen' },
  poi: { bg: 'bg-evil', icon: '❓', title: 'Unexplored point of interest' },
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

const CELL_SIZE = 20

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
      aria-label="Overworld map"
      className="bg-[#0c0a08] border-2 border-accent rounded p-4 flex flex-col items-center gap-4"
    >
      <div className="overflow-x-auto max-w-full">
        <div
          className="grid gap-px"
          style={{ gridTemplateColumns: `repeat(${MAP_WIDTH}, ${CELL_SIZE}px)` }}
        >
          {grid.map((row, y) =>
            row.map((tile, x) => {
              const isPlayer = coordinate && x === coordinate.x && y === coordinate.y
              const style = TILE_STYLE[tile]
              return (
                <div
                  key={`${x}-${y}`}
                  title={isPlayer ? 'You' : style.title}
                  className={`flex items-center justify-center text-[11px] leading-none ${isPlayer ? 'bg-good ring-2 ring-parchment z-10' : style.bg}`}
                  style={{ width: CELL_SIZE, height: CELL_SIZE }}
                >
                  {isPlayer ? '⚔️' : style.icon}
                </div>
              )
            }),
          )}
        </div>
      </div>

      <DirectionPad />
    </div>
  )
}
