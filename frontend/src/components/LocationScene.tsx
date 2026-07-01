import { useState } from 'react'
import { useGame } from '../state/GameProvider'

const actionButtonClass = 'bg-accent text-ink rounded px-4 py-2 text-left font-mono hover:brightness-110'

export function LocationScene() {
  const { state, actions } = useGame()
  const [inviteName, setInviteName] = useState('')
  const location = state.location
  if (!location) return null

  const contextualActions: { key: string; label: string; icon: string; onClick: () => void }[] = []
  if (location.kind === 'guild_hall') {
    contextualActions.push({
      key: 'guild',
      icon: '\u{1F3DB}\u{FE0F}',
      label: 'Manage your roster',
      onClick: () => actions.setView('character-creation'),
    })
  }
  if (location.kind === 'npc') {
    contextualActions.push({ key: 'npc', icon: '\u{1F9D9}', label: 'Talk to the citizen', onClick: actions.talkToNpc })
  }
  if (location.kind === 'quest_hook') {
    contextualActions.push({ key: 'dungeon', icon: '\u{2753}', label: 'Explore the depths', onClick: actions.enterDungeon })
  }

  function handleInvite(e: React.FormEvent) {
    e.preventDefault()
    if (!inviteName.trim()) return
    actions.inviteToParty(inviteName.trim())
    setInviteName('')
  }

  return (
    <section className="bg-panel border border-accent rounded p-4 flex flex-col gap-3">
      <h3 className="m-0 text-sm uppercase tracking-wide text-accent">Here</h3>

      {contextualActions.length === 0 ? (
        <p className="text-sm italic text-[#8a7e63] m-0">Nothing to do here but talk and travel on.</p>
      ) : (
        <div className="flex flex-col gap-2">
          {contextualActions.map((a) => (
            <button key={a.key} className={actionButtonClass} onClick={a.onClick}>
              <span className="mr-2">{a.icon}</span>
              {a.label}
            </button>
          ))}
        </div>
      )}

      <div className="border-t border-dashed border-[#4a3f2c] pt-3">
        <p className="text-xs uppercase tracking-wide text-accent mb-1">Who's here</p>
        {state.presentAtLocation.length === 0 ? (
          <p className="text-sm italic text-[#8a7e63] m-0">No one else is around.</p>
        ) : (
          <ul className="list-none p-0 m-0 space-y-1">
            {state.presentAtLocation.map((p) => (
              <li key={p.accountId} className="flex justify-between items-center text-sm">
                <span>{p.displayName}</span>
                <button
                  className="bg-accent text-ink rounded px-2 py-0.5 text-xs"
                  onClick={() => actions.inviteToParty(p.displayName)}
                >
                  Invite
                </button>
              </li>
            ))}
          </ul>
        )}
        <form onSubmit={handleInvite} className="flex gap-2 mt-2">
          <input
            className="flex-1 bg-[#0c0a08] text-parchment border border-accent rounded px-2 py-1 text-sm"
            placeholder="Invite by name..."
            value={inviteName}
            onChange={(e) => setInviteName(e.target.value)}
          />
          <button type="submit" className="bg-accent text-ink rounded px-3 py-1 text-sm">
            Invite
          </button>
        </form>
      </div>
    </section>
  )
}
