import { useState } from 'react'
import { useGame } from '../state/GameProvider'
import type { Combatant } from '../engine/types'

function HpBar({ c }: { c: Combatant }) {
  const pct = c.maxHp > 0 ? Math.max(0, Math.round((c.hp / c.maxHp) * 100)) : 0
  return (
    <div className="w-full h-1.5 bg-[#2a2218] rounded overflow-hidden">
      <div className={`h-full ${pct > 50 ? 'bg-good' : pct > 20 ? 'bg-accent' : 'bg-evil'}`} style={{ width: `${pct}%` }} />
    </div>
  )
}

export function CombatView() {
  const { state, actions } = useGame()
  const [selectedTarget, setSelectedTarget] = useState<string | null>(null)
  const encounter = state.activeEncounter
  if (!encounter) return null

  const combatants = encounter.combatants ?? []
  const log = encounter.log ?? []
  const myCharacterId = state.characters.find((c) => c.id === state.account?.activeCharacterId)?.id
  const isMyTurn = !!myCharacterId && encounter.currentCombatantId === myCharacterId
  const monsters = combatants.filter((c) => c.kind === 'monster')
  const players = combatants.filter((c) => c.kind === 'player')
  const currentName = combatants.find((c) => c.id === encounter.currentCombatantId)?.name

  function attack() {
    const targetId = selectedTarget ?? monsters.find((m) => !m.defeated)?.id
    if (targetId) actions.combatAction('attack', targetId)
    setSelectedTarget(null)
  }

  return (
    <div className="bg-[#0c0a08] border border-evil rounded p-3 mt-3">
      <p className="m-0 mb-2 text-sm text-accent">
        {encounter.roomLabel ? `${encounter.roomLabel} — ` : ''}Round {encounter.round}
      </p>

      <div className="grid grid-cols-2 gap-3 mb-3">
        <div>
          <p className="text-xs uppercase text-accent m-0 mb-1">Party</p>
          {players.map((p) => (
            <div key={p.id} className={`mb-1.5 ${encounter.currentCombatantId === p.id ? 'ring-1 ring-accent rounded p-1' : ''}`}>
              <div className="flex justify-between text-xs">
                <span>
                  {p.name}
                  {p.defeated ? ' (down)' : p.fled ? ' (fled)' : p.dodging ? ' (dodging)' : ''}
                </span>
                <span>
                  {p.hp}/{p.maxHp}
                </span>
              </div>
              <HpBar c={p} />
            </div>
          ))}
        </div>
        <div>
          <p className="text-xs uppercase text-evil m-0 mb-1">Foes</p>
          {monsters.map((m) => (
            <button
              key={m.id}
              onClick={() => !m.defeated && setSelectedTarget(m.id)}
              disabled={m.defeated}
              className={`block w-full text-left mb-1.5 disabled:opacity-50 ${
                encounter.currentCombatantId === m.id ? 'ring-1 ring-evil rounded p-1' : ''
              } ${selectedTarget === m.id ? 'bg-[#2a1c1c] rounded p-1' : ''}`}
            >
              <div className="flex justify-between text-xs">
                <span>
                  {m.name}
                  {m.defeated ? ' (defeated)' : ''}
                </span>
                <span>
                  {m.hp}/{m.maxHp}
                </span>
              </div>
              <HpBar c={m} />
            </button>
          ))}
        </div>
      </div>

      {isMyTurn ? (
        <div className="flex gap-2">
          <button className="bg-evil text-parchment rounded px-3 py-1.5 text-sm" onClick={attack}>
            Attack{selectedTarget ? ` (${monsters.find((m) => m.id === selectedTarget)?.name})` : ''}
          </button>
          <button className="bg-accent text-ink rounded px-3 py-1.5 text-sm" onClick={() => actions.combatAction('dodge')}>
            Dodge
          </button>
          <button className="bg-[#4a3f2c] text-parchment rounded px-3 py-1.5 text-sm" onClick={() => actions.combatAction('flee')}>
            Flee
          </button>
        </div>
      ) : (
        <p className="text-sm italic text-[#8a7e63] m-0">Waiting on {currentName ?? 'someone'}...</p>
      )}

      <div className="mt-3 bg-black/40 rounded p-2 text-xs space-y-1 max-h-32 overflow-y-auto font-mono">
        {log.map((roll, i) => (
          <p key={i} className="m-0">
            {roll.attacker} → {roll.target}: d20={roll.d20}+{roll.attackBonus}={roll.total} vs AC{roll.targetAc} —{' '}
            {roll.hit ? `hit for ${roll.damage}` : 'miss'}
          </p>
        ))}
      </div>
    </div>
  )
}
