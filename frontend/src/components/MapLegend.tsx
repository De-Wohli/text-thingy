const ENTRIES: [string, string][] = [
  ['@', 'Active Player Character'],
  ['A', "Adventurer's Guild Hall — Safe Zone / Character Swapping Point"],
  ['T', 'The Yawning Flask Tavern — Gather Rumors / Rent Crafting Spaces'],
  ['N', 'Citizen NPCs — Quest Givers'],
  ['~', 'Impassable River'],
  ['?', 'Unexplored Point of Interest — Discovers a Procedural Dungeon'],
]

export function MapLegend() {
  return (
    <ul className="list-none p-0 m-0 text-xs text-[#b3a78c] space-y-0.5">
      {ENTRIES.map(([glyph, label]) => (
        <li key={glyph}>
          <span className="inline-block w-4 font-bold text-accent">{glyph}</span> {label}
        </li>
      ))}
    </ul>
  )
}
