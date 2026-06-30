import { useGame } from '../state/GameProvider'
import type { DungeonRoom } from '../engine/types'

const ROOM_ORDER: DungeonRoom['type'][] = ['start', 'hallway', 'treasure', 'boss']

const ROOM_ICON: Record<DungeonRoom['type'], string> = {
  start: '\u{1F6AA}', // door
  hallway: '\u{1F6B6}', // walking figure
  treasure: '\u{1F4B0}', // money bag
  boss: '\u{1F479}', // ogre
}

const ROOM_LABEL: Record<DungeonRoom['type'], string> = {
  start: 'Entrance',
  hallway: 'Corridor',
  treasure: 'Treasure Vault',
  boss: "Boss's Den",
}

export function DungeonView() {
  const { state, actions } = useGame()
  const dungeon = state.activeDungeon
  if (!dungeon) return null

  return (
    <div
      className="fixed inset-0 bg-black/90 flex items-center justify-center p-4 z-20"
      role="dialog"
      aria-label="Dungeon encounter track"
    >
      <div className="bg-panel border border-accent rounded-md p-6 max-w-2xl w-full max-h-[85vh] overflow-y-auto">
        <header className="mb-4">
          <h2 className="m-0">Dungeon Instance</h2>
          <p className="text-sm text-[#b3a78c] m-0">Resolve each encounter in order to reach the boss.</p>
        </header>

        <div className="flex flex-col md:flex-row items-stretch gap-2">
          {ROOM_ORDER.map((type, i) => {
            const room = dungeon.rooms.find((r) => r.type === type)
            const encounter = dungeon.encounters.find((e) => e.roomType === type)
            if (!room) return null
            const isBoss = type === 'boss'

            return (
              <div key={type} className="flex items-center md:flex-col flex-1 gap-2">
                <div
                  className={`flex-1 w-full rounded border p-3 flex flex-col items-center text-center gap-1 ${
                    room.cleared
                      ? 'border-good bg-[#1c2a1c]'
                      : isBoss
                        ? 'border-evil bg-[#2a1c1c]'
                        : 'border-accent bg-[#241f17]'
                  }`}
                >
                  <span className="text-3xl" aria-hidden>
                    {ROOM_ICON[type]}
                  </span>
                  <strong>{ROOM_LABEL[type]}</strong>

                  {encounter && encounter.monsters.length > 0 && (
                    <div className="flex flex-wrap justify-center gap-1 text-xs">
                      {encounter.monsters.map((m, idx) => (
                        <span key={idx} className="bg-[#0c0a08] rounded px-1.5 py-0.5">
                          {m.name}
                        </span>
                      ))}
                    </div>
                  )}

                  {room.cleared ? (
                    <span className="text-good font-bold text-sm">✓ Cleared</span>
                  ) : (
                    <button
                      className="bg-accent text-ink rounded px-2 py-1 text-sm mt-1"
                      onClick={() => actions.clearDungeonRoom(type)}
                    >
                      Resolve Encounter
                    </button>
                  )}
                </div>

                {i < ROOM_ORDER.length - 1 && (
                  <span className="text-accent text-xl md:rotate-0 rotate-90" aria-hidden>
                    →
                  </span>
                )}
              </div>
            )
          })}
        </div>

        <button
          className="bg-good text-ink w-full mt-4 rounded px-3 py-2 disabled:bg-[#4a3f2c] disabled:text-[#8a7e63]"
          disabled={!dungeon.resolved}
          onClick={actions.resolveDungeon}
        >
          Return to City Gates
        </button>
      </div>
    </div>
  )
}
