import { useGame } from '../state/GameProvider'
import { RACES } from '../engine/races'
import { CLASSES } from '../engine/classes'

export function CharacterSheet() {
  const { state } = useGame()
  const activeId = state.account?.activeCharacterId
  const character = activeId ? state.characters.find((c) => c.id === activeId) : null

  if (!character) {
    return (
      <div className="border-t border-dashed border-[#4a3f2c] pt-3">
        <p className="text-sm text-[#b3a78c]">No active character. Visit the Guild Hall [A] to recruit one.</p>
      </div>
    )
  }

  const race = RACES[character.raceId]
  const klass = CLASSES[character.classId]

  return (
    <div className="border-t border-dashed border-[#4a3f2c] pt-3">
      <h3 className="m-0 flex items-center gap-2">
        {character.name}
        <span className="text-[0.7rem] bg-accent text-[#1a1a1a] rounded px-1.5 py-0.5">{character.status}</span>
      </h3>
      <p className="my-1 text-sm">
        Level {character.level} {race.name} {klass.name}
      </p>
      <p className="my-1 text-sm">
        HP: {character.hpCurrent} / {character.hpMax}
      </p>
      <dl className="grid grid-cols-3 gap-1 my-2">
        {Object.entries(character.abilityScores).map(([key, value]) => (
          <div key={key} className="bg-[#2a2218] rounded p-1 text-center">
            <dt className="text-[0.7rem] text-accent">{key.toUpperCase()}</dt>
            <dd className="m-0 font-bold">{value}</dd>
          </div>
        ))}
      </dl>
      <p className="text-sm text-[#b3a78c]">{klass.features.join(', ')}</p>
    </div>
  )
}
