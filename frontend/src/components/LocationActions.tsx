import { useGame } from '../state/GameProvider'
import { LANDMARKS, isAdjacentOrEqual } from '../data/overworld'

const actionButtonClass =
  'bg-accent text-ink rounded px-4 py-2 text-left font-mono hover:brightness-110'

export function LocationActions() {
  const { state, actions } = useGame()
  const coordinate = state.account?.coordinate

  const atGuildHall = !!coordinate && !!LANDMARKS.guildHall && isAdjacentOrEqual(coordinate, LANDMARKS.guildHall)
  const atNpc = !!coordinate && !!LANDMARKS.npc && isAdjacentOrEqual(coordinate, LANDMARKS.npc)
  const atPoi = !!coordinate && !!LANDMARKS.poi && isAdjacentOrEqual(coordinate, LANDMARKS.poi)

  const available: { key: string; label: string; icon: string; onClick: () => void }[] = []
  if (atGuildHall) {
    available.push({
      key: 'guild',
      icon: '\u{1F3DB}\u{FE0F}',
      label: 'Enter the Guild Hall — manage your roster',
      onClick: () => actions.setView('character-creation'),
    })
  }
  if (atNpc) {
    available.push({ key: 'npc', icon: '\u{1F9D9}', label: 'Talk to the citizen', onClick: actions.talkToNpc })
  }
  if (atPoi) {
    available.push({ key: 'poi', icon: '❓', label: 'Investigate the point of interest', onClick: actions.enterPOI })
  }

  return (
    <section className="bg-panel border border-accent rounded p-4">
      <h3 className="m-0 mb-2 text-sm uppercase tracking-wide text-accent">Here</h3>
      {available.length === 0 ? (
        <p className="text-sm italic text-[#8a7e63] m-0">
          Nothing to do here. Walk toward a landmark on the map to see what&rsquo;s nearby.
        </p>
      ) : (
        <div className="flex flex-col gap-2">
          {available.map((a) => (
            <button key={a.key} className={actionButtonClass} onClick={a.onClick}>
              <span className="mr-2">{a.icon}</span>
              {a.label}
            </button>
          ))}
        </div>
      )}
    </section>
  )
}
