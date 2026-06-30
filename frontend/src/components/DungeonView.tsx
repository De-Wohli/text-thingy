import { useGame } from '../state/GameProvider'
import type { DungeonRoom } from '../engine/types'

const ROOM_ORDER: DungeonRoom['type'][] = ['start', 'hallway', 'treasure', 'boss']

export function DungeonView() {
  const { state, actions } = useGame()
  const dungeon = state.activeDungeon
  if (!dungeon) return null

  return (
    <div className="fixed inset-0 bg-black/90 flex items-center justify-center p-4 z-20" role="dialog" aria-label="Procedural dungeon instance">
      <div className="bg-panel border border-accent rounded-md p-6 max-w-md w-full max-h-[85vh] overflow-y-auto">
        <header className="mb-2">
          <h2 className="m-0">Dungeon Instance</h2>
        </header>

        <pre className="bg-[#0c0a08] p-2 text-[10px] leading-tight text-[#8fbf8f] overflow-x-auto">
          {dungeon.grid.map((row) => row.map((tile) => (tile === 'floor' ? '.' : '#')).join('')).join('\n')}
        </pre>

        <ul className="list-none p-0">
          {ROOM_ORDER.map((type) => {
            const room = dungeon.rooms.find((r) => r.type === type)
            const encounter = dungeon.encounters.find((e) => e.roomType === type)
            if (!room) return null
            return (
              <li
                key={type}
                className={`flex justify-between items-center gap-2 py-1.5 border-b border-[#3a3120] ${room.cleared ? 'opacity-60' : ''}`}
              >
                <span>
                  <strong className="capitalize">{type}</strong>
                  {encounter && encounter.monsters.length > 0 && (
                    <span> — {encounter.monsters.map((m) => m.name).join(', ')}</span>
                  )}
                </span>
                {room.cleared ? (
                  <span className="text-good font-bold">Cleared</span>
                ) : (
                  <button
                    className="bg-accent text-ink rounded px-2 py-1"
                    onClick={() => actions.clearDungeonRoom(type)}
                  >
                    Resolve Encounter
                  </button>
                )}
              </li>
            )
          })}
        </ul>

        <button
          className="bg-good text-ink w-full mt-3 rounded px-3 py-1.5 disabled:bg-[#4a3f2c] disabled:text-[#8a7e63]"
          disabled={!dungeon.resolved}
          onClick={actions.resolveDungeon}
        >
          Return to City Gates
        </button>
      </div>
    </div>
  )
}
