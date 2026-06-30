import { useGame } from '../state/GameProvider'
import { CharacterSheet } from './CharacterSheet'
import { HonorMeter } from './HonorMeter'
import { MapLegend } from './MapLegend'
import { LANDMARKS, isAdjacentOrEqual } from '../data/overworld'

const buttonClass = 'bg-accent text-ink rounded px-3 py-1.5 font-mono disabled:bg-[#4a3f2c] disabled:text-[#8a7e63]'

export function Dashboard() {
  const { state, actions } = useGame()
  const coordinate = state.account?.coordinate

  const atGuildHall = !!coordinate && !!LANDMARKS.guildHall && isAdjacentOrEqual(coordinate, LANDMARKS.guildHall)
  const atNpc = !!coordinate && !!LANDMARKS.npc && isAdjacentOrEqual(coordinate, LANDMARKS.npc)
  const atPoi = !!coordinate && !!LANDMARKS.poi && isAdjacentOrEqual(coordinate, LANDMARKS.poi)

  return (
    <aside className="flex flex-col gap-3 bg-panel border border-accent rounded p-4">
      <p className="text-xs text-[#8a7e63]">
        Connection: <span className={state.connection === 'open' ? 'text-good' : 'text-evil'}>{state.connection}</span>
      </p>
      <HonorMeter honor={state.account?.honor ?? 0} />
      <CharacterSheet />

      <section className="border-t border-dashed border-[#4a3f2c] pt-3 text-sm">
        <p className="my-1">Gold: {state.account?.gold ?? 0}</p>
        <p className="my-1">Party: {state.account?.partyId ?? 'solo'}</p>
      </section>

      <section className="flex flex-col gap-1.5">
        <h3 className="m-0">Actions</h3>
        <button className={buttonClass} disabled={!atGuildHall} onClick={() => actions.setView('character-creation')}>
          Guild Hall: Manage Roster
        </button>
        <button className={buttonClass} disabled={!atNpc} onClick={actions.talkToNpc}>
          Talk to Citizen
        </button>
        <button className={buttonClass} disabled={!atPoi} onClick={actions.enterPOI}>
          Explore Point of Interest
        </button>
      </section>

      {state.lastMessage && <p className="bg-[#2a2218] border-l-2 border-accent p-2 text-sm">{state.lastMessage}</p>}

      <MapLegend />
      <p className="text-xs text-[#8a7e63]">Move with WASD or arrow keys.</p>
    </aside>
  )
}
