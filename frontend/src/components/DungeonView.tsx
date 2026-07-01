import { useGame } from '../state/GameProvider'
import { CombatView } from './CombatView'
import type { AttackRoll } from '../engine/types'

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
  const fighting = !!state.activeEncounter

  return (
    <div className="bg-panel border border-accent rounded p-4 flex flex-col gap-4">
      <header>
        <h2 className="m-0 text-lg">Ruined Keep</h2>
        {state.dungeonEntryNarration && (
          <p className="text-sm italic text-accent mt-1 mb-0">{state.dungeonEntryNarration}</p>
        )}
      </header>

      {/* Room progression track — iterate server-ordered rooms directly */}
      <div className="flex flex-wrap gap-2">
        {dungeon.rooms.map((room, i) => {
          const encounter = dungeon.encounters[i]
          const isBoss = room.type === 'boss'

          return (
            <div key={`${room.type}-${i}`} className="flex items-center gap-2">
              <div
                className={`min-w-[110px] rounded border p-2 flex flex-col items-center text-center gap-1 ${
                  room.cleared
                    ? 'border-good bg-[#1c2a1c]'
                    : isBoss
                      ? 'border-evil bg-[#2a1c1c]'
                      : 'border-accent bg-[#241f17]'
                }`}
              >
                <span className="text-xl" aria-hidden>
                  {room.icon || (isBoss ? '👑' : '🚪')}
                </span>
                <strong className="text-xs leading-tight">{room.label || room.type}</strong>

                {encounter && encounter.monsters.length > 0 && !room.cleared && (
                  <div className="flex flex-wrap justify-center gap-1" style={{ fontSize: '0.6rem' }}>
                    {encounter.monsters.map((m, idx) => (
                      <span key={idx} className="bg-[#0c0a08] rounded px-1 py-0.5">
                        {m.name}
                      </span>
                    ))}
                  </div>
                )}

                {room.cleared ? (
                  <span className="text-good font-bold" style={{ fontSize: '0.65rem' }}>✓ Cleared</span>
                ) : (
                  <span style={{ fontSize: '0.65rem' }} className="text-[#8a7e63]">
                    {room.cleared ? '' : 'Not yet reached'}
                  </span>
                )}
              </div>

              {i < dungeon.rooms.length - 1 && (
                <span className="text-accent text-sm" aria-hidden>→</span>
              )}
            </div>
          )
        })}
      </div>

      {state.lastSkillCheck && (
        <div
          className={`rounded border p-2 text-sm ${state.lastSkillCheck.result.success ? 'border-good bg-[#1c2a1c]' : 'border-[#4a3f2c] bg-[#241f17]'}`}
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
        <div className={`rounded border p-3 ${resolution.victory ? 'border-good bg-[#1c2a1c]' : 'border-evil bg-[#2a1c1c]'}`}>
          <p className={`m-0 mb-1 font-bold text-sm ${resolution.victory ? 'text-good' : 'text-evil'}`}>
            {resolution.victory ? 'Victory' : 'Defeat'} — {resolution.label || resolution.roomType}
          </p>
          <p className="italic text-sm mb-2">{resolution.narration}</p>
          <div className="bg-[#0c0a08] rounded p-2 text-xs space-y-1 max-h-32 overflow-y-auto font-mono">
            {(resolution.combatLog ?? []).length === 0 ? (
              <p className="m-0 italic text-[#8a7e63]">No blows exchanged.</p>
            ) : (
              resolution.combatLog.map((roll, i) => <AttackLine key={i} roll={roll} />)
            )}
          </div>
        </div>
      )}

      <button
        className="bg-good text-ink w-full rounded px-3 py-2 disabled:bg-[#4a3f2c] disabled:text-[#8a7e63] text-sm"
        disabled={!dungeon.resolved}
        onClick={actions.resolveDungeon}
      >
        Return to Town
      </button>
    </div>
  )
}
