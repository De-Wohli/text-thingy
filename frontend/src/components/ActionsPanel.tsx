import { useEffect, useState } from 'react'
import { useGame } from '../state/GameProvider'
import { LOCATION_LABELS } from '../engine/locations'
import type { Skill } from '../engine/types'

const SKILL_CHECKS: { skill: Skill; label: string; context: string; icon: string; outcome: string }[] = [
  { skill: 'perception', label: 'Perception', context: 'listen', icon: '👁️', outcome: 'Players act first' },
  { skill: 'investigation', label: 'Investigation', context: 'search-for-traps', icon: '🔍', outcome: 'Remove weakest foe' },
  { skill: 'stealth', label: 'Stealth', context: 'listen', icon: '🌑', outcome: 'Free attack before initiative' },
  { skill: 'insight', label: 'Insight', context: 'read-the-room', icon: '🧠', outcome: '+2 attack this fight' },
  { skill: 'arcana', label: 'Arcana', context: 'investigate', icon: '✨', outcome: '+1 damage this fight' },
  { skill: 'athletics', label: 'Athletics', context: 'investigate', icon: '💪', outcome: '+3 HP before fight' },
]

const SKILL_FAILURE: Partial<Record<Skill, string>> = {
  investigation: 'Trap springs (1d4 damage)',
  perception: 'Monsters alert (+2 ATK)',
  stealth: 'Monsters alert (+2 ATK)',
}

const btnBase = 'w-full rounded px-3 py-2 text-sm flex items-center justify-between gap-2'
const btnPrimary = `${btnBase} bg-accent text-ink font-bold hover:brightness-110`
const btnSecondary = `${btnBase} bg-[#2a2218] border border-accent text-parchment hover:brightness-110`
const btnCooldown = `${btnBase} bg-[#2a2218] border border-[#4a3f2c] text-[#8a7e63] cursor-not-allowed`
const btnDanger = `${btnBase} bg-[#4a3f2c] text-parchment hover:brightness-110`

// Tick every second to update cooldown countdowns.
function useNow(active: boolean): number {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    if (!active) return
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [active])
  return now
}

export function ActionsPanel() {
  const { state, actions } = useGame()
  const [open, setOpen] = useState(false)

  const hasCooldowns = Object.keys(state.skillCooldowns).length > 0
  const now = useNow(open && hasCooldowns)

  const location = state.location
  const dungeon = state.activeDungeon
  const fighting = !!state.activeEncounter
  const currentRoom = dungeon?.rooms.find((r) => !r.cleared)
  const inDungeon = state.view === 'dungeon'

  const contextLabel = inDungeon ? (currentRoom?.label ?? 'Dungeon') : (location?.name ?? 'The World')

  function cooldownSecsFor(skill: Skill): number {
    const expiry = state.skillCooldowns[skill]
    if (!expiry || now >= expiry) return 0
    return Math.ceil((expiry - now) / 1000)
  }

  function renderSkillButton(check: (typeof SKILL_CHECKS)[number], context: string) {
    const secs = cooldownSecsFor(check.skill)
    if (secs > 0) {
      return (
        <button key={check.skill} className={btnCooldown} disabled>
          <span>
            {check.icon} {check.label}
          </span>
          <span className="text-xs shrink-0">⏳ {secs}s</span>
        </button>
      )
    }
    const failureNote = SKILL_FAILURE[check.skill]
    return (
      <button
        key={check.skill}
        className={btnDanger}
        onClick={() => actions.skillCheck(check.skill, context)}
        title={`Success: ${check.outcome}${failureNote ? ` · Failure: ${failureNote}` : ''}`}
      >
        <span>
          {check.icon} {check.label}
        </span>
        <span className="text-[0.6rem] text-[#8a7e63] text-right leading-tight shrink-0">
          ✓ {check.outcome}
          {failureNote && (
            <>
              <br />✗ {failureNote}
            </>
          )}
        </span>
      </button>
    )
  }

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
              <button className={btnPrimary} onClick={() => actions.startEncounter(currentRoom.type, currentRoom.label)}>
                ⚔️ Start Encounter
              </button>

              <p className="text-xs uppercase text-accent m-0 mt-1">Skill Checks</p>
              {renderSkillButton(SKILL_CHECKS.find((c) => c.skill === 'investigation')!, 'search-for-traps')}
              {renderSkillButton(SKILL_CHECKS.find((c) => c.skill === 'perception')!, 'listen')}
              {renderSkillButton(SKILL_CHECKS.find((c) => c.skill === 'stealth')!, 'listen')}
              {renderSkillButton(SKILL_CHECKS.find((c) => c.skill === 'insight')!, 'read-the-room')}
              {renderSkillButton(SKILL_CHECKS.find((c) => c.skill === 'arcana')!, 'investigate')}
              {renderSkillButton(SKILL_CHECKS.find((c) => c.skill === 'athletics')!, 'investigate')}
            </>
          )}

          {inDungeon && fighting && (
            <p className="text-xs text-[#8a7e63] italic m-0">Combat actions are shown in the encounter panel.</p>
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

              <p className="text-xs uppercase text-accent m-0 mt-1">Ability Checks</p>
              {SKILL_CHECKS.map((check) => renderSkillButton(check, check.context))}

              {location.connections.length > 0 && (
                <>
                  <p className="text-xs uppercase text-accent m-0 mt-1">Travel</p>
                  {location.connections.map((connId) => {
                    const label = LOCATION_LABELS[connId]
                    return (
                      <button key={connId} className={btnSecondary} onClick={() => actions.travel(connId)}>
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
