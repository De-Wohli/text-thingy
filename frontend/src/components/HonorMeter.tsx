import { reactivityForHonor } from '../engine/honor'

const FILL_COLOR: Record<string, string> = {
  good: 'bg-good',
  neutral: 'bg-accent',
  evil: 'bg-evil',
}

export function HonorMeter({ honor }: { honor: number }) {
  const reactivity = reactivityForHonor(honor)
  const fillPercent = ((honor + 100) / 200) * 100

  return (
    <div className="border-t border-dashed border-[#4a3f2c] pt-3">
      <div className="flex justify-between font-bold text-sm">
        <span>Honor</span>
        <span>
          {honor} ({reactivity.alignment})
        </span>
      </div>
      <div className="my-1 h-2.5 rounded bg-[#2a2218] overflow-hidden">
        <div className={`h-full ${FILL_COLOR[reactivity.alignment]}`} style={{ width: `${fillPercent}%` }} />
      </div>
      <p className="text-xs text-[#b3a78c]">{reactivity.greeting}</p>
    </div>
  )
}
