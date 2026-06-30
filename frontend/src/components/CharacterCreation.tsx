import { useState } from 'react'
import { useGame } from '../state/GameProvider'
import { listRaces } from '../engine/races'
import { listClasses } from '../engine/classes'
import type { ClassId, RaceId } from '../engine/types'

const panelClass =
  'fixed inset-0 bg-black/90 flex items-center justify-center p-4 z-20'
const cardClass =
  'bg-panel border border-accent rounded-md p-6 max-w-md w-full max-h-[85vh] overflow-y-auto'
const buttonClass = 'bg-accent text-ink rounded px-3 py-1.5 font-mono disabled:bg-[#4a3f2c] disabled:text-[#8a7e63]'
const inputClass = 'bg-[#0c0a08] text-parchment border border-accent rounded px-2 py-1 font-mono'

export function CharacterCreation() {
  const { state, actions } = useGame()
  const [name, setName] = useState('')
  const [raceId, setRaceId] = useState<RaceId>('human')
  const [classId, setClassId] = useState<ClassId>('fighter')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    actions.createCharacter(name, raceId, classId)
    setName('')
  }

  return (
    <div className={panelClass} role="dialog" aria-label="Character creation and roster">
      <div className={cardClass}>
        <header className="flex justify-between items-center mb-2">
          <h2 className="m-0">Adventurer&rsquo;s Guild Hall</h2>
          <button className={buttonClass} onClick={actions.closePanel}>
            Close
          </button>
        </header>

        <section className="mb-4">
          <h3>Roster</h3>
          {state.characters.length === 0 && <p className="text-sm">No characters yet — create your first below.</p>}
          <ul className="list-none p-0">
            {state.characters.map((character) => (
              <li key={character.id} className="flex justify-between items-center gap-2 py-1.5 border-b border-[#3a3120]">
                <span>
                  {character.name} (Lv{character.level} {character.raceId}/{character.classId})
                </span>
                <button
                  className={buttonClass}
                  disabled={state.account?.activeCharacterId === character.id}
                  onClick={() => actions.swapCharacter(character.id)}
                >
                  {state.account?.activeCharacterId === character.id ? 'Active' : 'Swap In'}
                </button>
              </li>
            ))}
          </ul>
        </section>

        <section>
          <h3>Recruit New Character</h3>
          <form onSubmit={handleSubmit} className="flex flex-col gap-2">
            <label className="flex flex-col text-sm gap-1">
              Name
              <input className={inputClass} value={name} onChange={(e) => setName(e.target.value)} required maxLength={24} />
            </label>
            <label className="flex flex-col text-sm gap-1">
              Race
              <select className={inputClass} value={raceId} onChange={(e) => setRaceId(e.target.value as RaceId)}>
                {listRaces().map((race) => (
                  <option key={race.id} value={race.id}>
                    {race.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col text-sm gap-1">
              Class
              <select className={inputClass} value={classId} onChange={(e) => setClassId(e.target.value as ClassId)}>
                {listClasses().map((klass) => (
                  <option key={klass.id} value={klass.id}>
                    {klass.name}
                  </option>
                ))}
              </select>
            </label>
            <button type="submit" className={buttonClass}>
              Create Character
            </button>
          </form>
        </section>

        {state.lastMessage && <p className="bg-[#2a2218] border-l-2 border-accent p-2 text-sm mt-3">{state.lastMessage}</p>}
      </div>
    </div>
  )
}
