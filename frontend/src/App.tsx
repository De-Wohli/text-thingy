import { useGame } from './state/GameProvider'
import { WorldMap } from './components/WorldMap'
import { Dashboard } from './components/Dashboard'
import { ChatPanel } from './components/ChatPanel'
import { LocationScene } from './components/LocationScene'
import { CharacterCreation } from './components/CharacterCreation'
import { ChoicePanel } from './components/ChoicePanel'
import { DungeonView } from './components/DungeonView'
import { Onboarding } from './components/Onboarding'

export function App() {
  const { state } = useGame()

  if (state.needsOnboarding && !state.account) {
    return <Onboarding />
  }

  return (
    <div className="max-w-6xl mx-auto p-4">
      <header>
        <h1 className="mb-0 tracking-wide">5e Virtual Tabletop — Prototype</h1>
        <p className="mt-1 text-accent">Account: {state.account?.displayName ?? 'connecting...'}</p>
      </header>

      <main className="grid grid-cols-1 md:grid-cols-[1fr_320px] gap-4 items-start mt-4">
        <div className="flex flex-col gap-4">
          <WorldMap />
          <LocationScene />
          <ChatPanel />
        </div>
        <Dashboard />
      </main>

      {state.view === 'character-creation' && <CharacterCreation />}
      {state.view === 'choice' && <ChoicePanel />}
      {state.view === 'dungeon' && <DungeonView />}
    </div>
  )
}
