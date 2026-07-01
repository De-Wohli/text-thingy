import { useGame } from '../state/GameProvider'
import { CombatView } from './CombatView'
import type { AttackRoll } from '../engine/types'

const SKILL_CHECKS: { key: string; label: string; skill: 'investigation' | 'perception'; context: string }[] = [
  { key: 'traps', label: 'Search for Traps', skill: 'investigation', context: 'search-for-traps' },
  { key: 'listen', label: 'Listen for Danger', skill: 'perception', context: 'listen' },
]

function AttackLine({ roll }: { roll: AttackRoll }) {
  const verb = roll.hit ? (roll.critical ? 'CRIT' : 'hits') : 'misses'
  return (
    <p className={`m-0 ${roll.hit ? (roll.critical ? 'text-evil font-bold' : '') : 'text-[#8a7e63] italic'}`}>
      {roll.attacker} attacks {roll.target}: d20={roll.d20}+{roll.attackBonus}={roll.total} vs AC {roll.targetAc} —{' '}
      {verb}
      {roll.hit ? ` for ${roll.damage} damage` : ''}
    </p>
  )
}

export function DungeonView() {
  const { state, actions } = useGame()
  const dungeon = state.activeDungeon
  if (!dungeon) return null

  const resolution = state.lastRoomResolution
  const currentRoom = dungeon.rooms.find((r) => !r.cleared)
  const fighting = !!state.activeEncounter

  return (
    <div
      className="fixed inset-0 bg-black/90 flex items-center justify-center p-4 z-20"
      role="dialog"
      aria-label="Dungeon encounter track"
    >
      <div className="bg-panel border border-accent rounded-md p-6 max-w-3xl w-full max-h-[85vh] overflow-y-auto">
        <header className="mb-4">
          <h2 className="m-0">Ruined Keep</h2>
          {state.dungeonEntryNarration && (
            <p className="text-sm italic text-accent mt-1">{state.dungeonEntryNarration}</p>
          )}
        </header>

        {/* Room progression track — iterate server-ordered rooms directly */}
        <div className="flex flex-col sm:flex-row flex-wrap gap-2">
          {dungeon.rooms.map((room, i) => {
            // Match by index — rooms and encounters are generated in the same
            // order, and multiple rooms may share the same functional type
            // (e.g. two hallway rooms), so find-by-type picks the wrong one.
            const encounter = dungeon.encounters[i]
            const isCurrent = currentRoom?.label === room.label
            const isBoss = room.type === 'boss'

            return (
              <div key={`${room.type}-${i}`} className="flex items-center gap-2">
                <div
                  className={`flex-1 min-w-[120px] rounded border p-3 flex flex-col items-center text-center gap-1 ${
                    room.cleared
                      ? 'border-good bg-[#1c2a1c]'
                      : isBoss
                        ? 'border-evil bg-[#2a1c1c]'
                        : 'border-accent bg-[#241f17]'
                  }`}
                >
                  <span className="text-2xl" aria-hidden>
                    {room.icon || (isBoss ? '👑' : '🚪')}
                  </span>
                  <strong className="text-sm">{room.label || room.type}</strong>

                  {encounter && encounter.monsters.length > 0 && !room.cleared && (
                    <div className="flex flex-wrap justify-center gap-1 text-xs">
                      {encounter.monsters.map((m, idx) => (
                        <span key={idx} className="bg-[#0c0a08] rounded px-1.5 py-0.5">
                          {m.name} (AC {m.armorClass})
                        </span>
                      ))}
                    </div>
                  )}

                  {room.cleared ? (
                    <span className="text-good font-bold text-sm">✓ Cleared</span>
                  ) : isCurrent && !fighting ? (
                    <div className="flex flex-col gap-1 w-full mt-1">
                      {SKILL_CHECKS.map((check) => (
                        <button
                          key={check.key}
                          className="bg-[#4a3f2c] text-parchment rounded px-2 py-1 text-xs"
                          onClick={() => actions.skillCheck(check.skill, check.context)}
                        >
                          {check.label}
                        </button>
                      ))}
                      <button
                        className="bg-accent text-ink rounded px-2 py-1 text-sm"
                        onClick={() => actions.startEncounter(room.type, room.label)}
                      >
                        Start Encounter
                      </button>
                    </div>
                  ) : !room.cleared ? (
                    <span className="text-xs text-[#8a7e63]">Not yet reached</span>
                  ) : null}
                </div>

                {i < dungeon.rooms.length - 1 && (
                  <span className="text-accent" aria-hidden>
                    →
                  </span>
                )}
              </div>
            )
          })}
        </div>

        {state.lastSkillCheck && (
          <div
            className={`mt-3 rounded border p-2 text-sm ${state.lastSkillCheck.result.success ? 'border-good bg-[#1c2a1c]' : 'border-[#4a3f2c] bg-[#241f17]'}`}
          >
            <p className="m-0 italic">{state.lastSkillCheck.narration}</p>
            <p className="m-0 text-xs text-[#8a7e63]">
              d20={state.lastSkillCheck.result.d20} + {state.lastSkillCheck.result.abilityModifier} ability
              {state.lastSkillCheck.result.proficient ? ` + ${state.lastSkillCheck.result.proficiencyBonus} proficiency` : ''} ={' '}
              {state.lastSkillCheck.result.total} vs DC {state.lastSkillCheck.result.dc}
            </p>
          </div>
        )}

        {fighting && <CombatView />}

        {resolution && !fighting && (
          <div
            className={`mt-4 rounded border p-3 ${resolution.victory ? 'border-good bg-[#1c2a1c]' : 'border-evil bg-[#2a1c1c]'}`}
          >
            <p className={`m-0 mb-2 font-bold ${resolution.victory ? 'text-good' : 'text-evil'}`}>
              {resolution.victory ? 'Victory' : 'Defeat'} — {resolution.label || resolution.roomType}
            </p>
            <p className="italic text-sm mb-2">{resolution.narration}</p>
            <div className="bg-[#0c0a08] rounded p-2 text-xs space-y-1 max-h-40 overflow-y-auto font-mono">
              {(resolution.combatLog ?? []).length === 0 ? (
                <p className="m-0 italic text-[#8a7e63]">No blows exchanged — the room was already clear.</p>
              ) : (
                resolution.combatLog.map((roll, i) => <AttackLine key={i} roll={roll} />)
              )}
            </div>
          </div>
        )}

        <button
          className="bg-good text-ink w-full mt-4 rounded px-3 py-2 disabled:bg-[#4a3f2c] disabled:text-[#8a7e63]"
          disabled={!dungeon.resolved}
          onClick={actions.resolveDungeon}
        >
          Return to Town
        </button>
      </div>
    </div>
  )
}
