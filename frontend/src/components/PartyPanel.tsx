import { useGame } from '../state/GameProvider'

export function PartyPanel() {
  const { state, actions } = useGame()

  return (
    <section className="bg-panel border border-accent rounded p-4 flex flex-col gap-2">
      <h3 className="m-0 text-sm uppercase tracking-wide text-accent">Party</h3>

      {state.pendingInvites.map((invite) => (
        <div
          key={invite.inviteId}
          className="bg-[#2a2218] border-l-2 border-accent p-2 text-sm flex justify-between items-center gap-2"
        >
          <span>{invite.fromDisplayName} invites you to party up</span>
          <div className="flex gap-1 shrink-0">
            <button
              className="bg-good text-ink rounded px-2 py-1 text-xs"
              onClick={() => {
                actions.acceptPartyInvite(invite.inviteId)
                actions.dismissInvite(invite.inviteId)
              }}
            >
              Accept
            </button>
            <button
              className="bg-evil text-parchment rounded px-2 py-1 text-xs"
              onClick={() => {
                actions.declinePartyInvite(invite.inviteId)
                actions.dismissInvite(invite.inviteId)
              }}
            >
              Decline
            </button>
          </div>
        </div>
      ))}

      {state.party.length === 0 ? (
        <p className="text-sm italic text-[#8a7e63] m-0">You're adventuring solo. Invite someone nearby to party up.</p>
      ) : (
        <>
          <ul className="list-none p-0 m-0 space-y-1">
            {state.party.map((m) => (
              <li key={m.accountId} className="flex justify-between items-center text-sm">
                <span>
                  {m.displayName}
                  {m.characterName && <span className="text-[#8a7e63]"> ({m.characterName})</span>}
                </span>
                {m.hpMax ? (
                  <span className="text-xs">
                    HP {m.hpCurrent}/{m.hpMax}
                  </span>
                ) : null}
              </li>
            ))}
          </ul>
          <button className="bg-[#4a3f2c] text-parchment rounded px-3 py-1 text-xs self-start" onClick={actions.leaveParty}>
            Leave Party
          </button>
        </>
      )}
    </section>
  )
}
