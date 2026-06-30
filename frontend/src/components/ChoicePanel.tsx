import { useEffect, useState } from 'react'
import { useGame } from '../state/GameProvider'
import type { ChoiceState } from '../state/gameState'

const panelClass = 'fixed inset-0 bg-black/90 flex items-center justify-center p-4 z-20'
const cardClass = 'bg-panel border border-accent rounded-md p-6 max-w-md w-full max-h-[85vh] overflow-y-auto'
const buttonClass = 'bg-accent text-ink rounded px-3 py-1.5 font-mono w-full text-left mt-2'

function useCountdown(deadline?: number): number {
  const [remaining, setRemaining] = useState(() => (deadline ? deadline - Date.now() : 0))
  useEffect(() => {
    if (!deadline) return
    const interval = setInterval(() => setRemaining(deadline - Date.now()), 250)
    return () => clearInterval(interval)
  }, [deadline])
  return Math.max(0, Math.ceil(remaining / 1000))
}

export function ChoicePanel() {
  const { state, actions } = useGame()
  const { choice, voteTallies, voteResolution } = state
  const secondsLeft = useCountdown(choice?.mode === 'party' ? choice.deadline : undefined)

  if (!choice) return null

  const votingOpen = choice.mode === 'party' && !voteResolution && secondsLeft > 0

  return (
    <div className={panelClass} role="dialog" aria-label="Narrative choice">
      <div className={cardClass}>
        <header className="flex justify-between items-center mb-2">
          <h2 className="m-0">Citizen</h2>
          <button className="bg-accent text-ink rounded px-3 py-1.5" onClick={actions.closePanel}>
            Close
          </button>
        </header>

        <p className="italic">&ldquo;{choice.prompt}&rdquo;</p>

        {choice.mode === 'party' && (
          <p className="text-sm text-[#b3a78c]">
            {voteResolution ? 'Vote resolved.' : `Party vote — ${secondsLeft}s remaining`}
          </p>
        )}

        {voteResolution ? (
          <div className="bg-[#2a2218] border-l-2 border-good p-2 text-sm mt-2 space-y-1">
            <p className="m-0 text-xs uppercase tracking-wide text-accent">Game Master</p>
            <p className="m-0">
              Chosen: <strong>{optionLabel(choice, voteResolution.optionId)}</strong>
            </p>
            {voteResolution.narration && <p className="m-0 italic">{voteResolution.narration}</p>}
            <p className="m-0">
              Honor {voteResolution.honorDelta >= 0 ? '+' : ''}
              {voteResolution.honorDelta} (now {voteResolution.newHonor})
              {voteResolution.tieBreak && ' — tie-break invoked'}
            </p>
          </div>
        ) : (
          <ul className="list-none p-0">
            {choice.options.map((option) => (
              <li key={option.id}>
                <button
                  className={buttonClass}
                  disabled={choice.mode === 'party' && !votingOpen}
                  onClick={() =>
                    choice.mode === 'solo'
                      ? actions.makeChoice(choice.promptId, option.id)
                      : actions.castVote(choice.promptId, option.id)
                  }
                >
                  {option.label}
                  {voteTallies && <span className="float-right">{voteTallies[option.id] ?? 0} votes</span>}
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

function optionLabel(choice: ChoiceState, optionId: string): string {
  return choice.options.find((o) => o.id === optionId)?.label ?? optionId
}
