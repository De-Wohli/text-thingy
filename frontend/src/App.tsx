import { useGame } from './state/GameProvider'
import { OverworldCanvas } from './components/OverworldCanvas'
import { Dashboard } from './components/Dashboard'
import { ChatPanel } from './components/ChatPanel'
import { LocationActions } from './components/LocationActions'
import { CharacterCreation } from './components/CharacterCreation'
import { ChoicePanel } from './components/ChoicePanel'
import { DungeonView } from './components/DungeonView'

export function App() {
  const { state } = useGame()

  return (
    <div className="max-w-6xl mx-auto p-4">
      <header>
        <h1 className="mb-0 tracking-wide">5e Web MMO — Prototype</h1>
        <p className="mt-1 text-accent">Account: {state.account?.displayName ?? 'connecting...'}</p>
      </header>

      <main className="grid grid-cols-1 md:grid-cols-[1fr_320px] gap-4 items-start mt-4">
        <div className="flex flex-col gap-4">
          <OverworldCanvas />
          <LocationActions />
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
