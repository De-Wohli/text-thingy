import { useState } from 'react'
import { useGame } from '../state/GameProvider'
import { LOCATION_LABELS } from '../engine/locations'
import type { Skill } from '../engine/types'

const SKILL_CHECKS: { skill: Skill; label: string; context: string; icon: string }[] = [
  { skill: 'perception', label: 'Perception', context: 'listen', icon: '👁️' },
  { skill: 'investigation', label: 'Investigation', context: 'search-for-traps', icon: '🔍' },
  { skill: 'insight', label: 'Insight', context: 'read-the-room', icon: '🧠' },
  { skill: 'stealth', label: 'Stealth', context: 'listen', icon: '🌑' },
  { skill: 'arcana', label: 'Arcana', context: 'investigate', icon: '✨' },
  { skill: 'athletics', label: 'Athletics', context: 'investigate', icon: '💪' },
]

const btnBase = 'w-full text-left rounded px-3 py-2 text-sm flex items-center gap-2 hover:brightness-110'
const btnPrimary = `${btnBase} bg-accent text-ink font-bold`
const btnSecondary = `${btnBase} bg-[#2a2218] border border-accent text-parchment`
const btnDanger = `${btnBase} bg-[#4a3f2c] text-parchment`

export function ActionsPanel() {
  const { state, actions } = useGame()
  const [open, setOpen] = useState(false)
  const location = state.location
  const dungeon = state.activeDungeon
  const fighting = !!state.activeEncounter
  const currentRoom = dungeon?.rooms.find((r) => !r.cleared)
  const inDungeon = state.view === 'dungeon'

  const contextLabel = inDungeon
    ? (currentRoom?.label ?? 'Dungeon')
    : (location?.name ?? 'The World')

  return (
    <section className="border-t border-dashed border-[#4a3f2c] pt-3">
      <button
        onClick={() => setOpen((v) => !v)}
        className="w-full flex justify-between items-center text-sm uppercase tracking-wide text-accent mb-2"
      >
        <span>🎲 Actions</span>
        <span>{open ? '▲' : '▼'}</span>
      </button>

      {open && (
        <div className="flex flex-col gap-1.5">
          <p className="text-xs text-[#8a7e63] m-0 mb-1">
            {inDungeon ? `Room: ${contextLabel}` : `Location: ${contextLabel}`}
          </p>

          {/* ── DUNGEON ACTIONS ─────────────────────────────── */}
          {inDungeon && !fighting && currentRoom && (
            <>
              <p className="text-xs uppercase text-accent m-0">Interactions</p>
              <button
                className={btnPrimary}
                onClick={() => actions.startEncounter(currentRoom.type, currentRoom.label)}
              >
                ⚔️ Start Encounter
              </button>
              <p className="text-xs uppercase text-accent m-0 mt-1">Skill Checks</p>
              <button
                className={btnSecondary}
                onClick={() => actions.skillCheck('investigation', 'search-for-traps')}
              >
                🔍 Search for Traps <span className="text-[#8a7e63] text-xs">(Investigation)</span>
              </button>
              <button
                className={btnSecondary}
                onClick={() => actions.skillCheck('perception', 'listen')}
              >
                👂 Listen for Danger <span className="text-[#8a7e63] text-xs">(Perception)</span>
              </button>
            </>
          )}

          {inDungeon && fighting && (
            <p className="text-xs text-[#8a7e63] italic m-0">
              Combat actions are shown in the encounter panel.
            </p>
          )}

          {inDungeon && dungeon?.resolved && (
            <button className={btnPrimary} onClick={actions.resolveDungeon}>
              🏡 Return to Town
            </button>
          )}

          {/* ── OVERWORLD ACTIONS ───────────────────────────── */}
          {!inDungeon && location && (
            <>
              {location.kind === 'guild_hall' && (
                <>
                  <p className="text-xs uppercase text-accent m-0">Interactions</p>
                  <button className={btnPrimary} onClick={() => actions.setView('character-creation')}>
                    📋 Manage Roster
                  </button>
                </>
              )}
              {location.kind === 'npc' && (
                <>
                  <p className="text-xs uppercase text-accent m-0">Interactions</p>
                  <button className={btnPrimary} onClick={actions.talkToNpc}>
                    💬 Talk to Citizen
                  </button>
                </>
              )}
              {location.kind === 'quest_hook' && (
                <>
                  <p className="text-xs uppercase text-accent m-0">Interactions</p>
                  <button className={btnPrimary} onClick={actions.enterDungeon}>
                    🗡️ Explore the Ruins
                  </button>
                </>
              )}

              {/* Skill checks for the overworld */}
              <p className="text-xs uppercase text-accent m-0 mt-1">Ability Checks</p>
              {SKILL_CHECKS.map((check) => (
                <button
                  key={check.skill}
                  className={btnDanger}
                  onClick={() => actions.skillCheck(check.skill, check.context)}
                >
                  {check.icon} {check.label}
                </button>
              ))}

              {/* Quick travel */}
              {location.connections.length > 0 && (
                <>
                  <p className="text-xs uppercase text-accent m-0 mt-1">Travel</p>
                  {location.connections.map((connId) => {
                    const label = LOCATION_LABELS[connId]
                    return (
                      <button
                        key={connId}
                        className={btnSecondary}
                        onClick={() => actions.travel(connId)}
                      >
                        {label?.icon ?? '📍'} {label?.name ?? connId}
                      </button>
                    )
                  })}
                </>
              )}
            </>
          )}
        </div>
      )}
    </section>
  )
}
