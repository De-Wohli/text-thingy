import { useGame } from '../state/GameProvider'
import { CharacterSheet } from './CharacterSheet'
import { HonorMeter } from './HonorMeter'
import { PartyPanel } from './PartyPanel'
import { ActionsPanel } from './ActionsPanel'

export function Dashboard() {
  const { state } = useGame()

  return (
    <aside className="flex flex-col gap-3 bg-panel border border-accent rounded p-4">
      <p className="text-xs text-[#8a7e63]">
        Connection: <span className={state.connection === 'open' ? 'text-good' : 'text-evil'}>{state.connection}</span>
      </p>
      <HonorMeter honor={state.account?.honor ?? 0} />
      <CharacterSheet />

      <section className="border-t border-dashed border-[#4a3f2c] pt-3 text-sm">
        <p className="my-1">Gold: {state.account?.gold ?? 0}</p>
      </section>

      <ActionsPanel />

      {state.lastMessage && <p className="bg-[#2a2218] border-l-2 border-accent p-2 text-sm">{state.lastMessage}</p>}

      <PartyPanel />
    </aside>
  )
}
