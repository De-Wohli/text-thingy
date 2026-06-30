const ENTRIES: { bg: string; icon: string; label: string }[] = [
  { bg: 'bg-good', icon: '⚔️', label: 'You' },
  { bg: 'bg-accent', icon: '\u{1F3DB}\u{FE0F}', label: "Adventurer's Guild Hall — recruit & swap characters" },
  { bg: 'bg-[#8a5a2b]', icon: '\u{1F37A}', label: 'The Yawning Flask Tavern — rumors & crafting' },
  { bg: 'bg-[#6b3fa0]', icon: '\u{1F9D9}', label: 'Citizen — talk for quests and choices' },
  { bg: 'bg-[#2b4a64]', icon: '~', label: 'Impassable river' },
  { bg: 'bg-evil', icon: '❓', label: 'Unexplored point of interest — triggers a dungeon' },
]

export function MapLegend() {
  return (
    <ul className="list-none p-0 m-0 text-xs text-[#b3a78c] space-y-1.5">
      {ENTRIES.map((entry) => (
        <li key={entry.label} className="flex items-center gap-2">
          <span className={`inline-flex items-center justify-center w-5 h-5 rounded ${entry.bg} text-[11px] shrink-0`}>
            {entry.icon}
          </span>
          {entry.label}
        </li>
      ))}
    </ul>
  )
}
