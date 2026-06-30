import { useGame } from '../state/GameProvider'

const buttonClass =
  'w-10 h-10 flex items-center justify-center rounded bg-accent text-ink text-lg font-bold active:scale-95 select-none'

export function DirectionPad() {
  const { actions } = useGame()

  return (
    <div
      className="grid grid-cols-3 grid-rows-3 gap-1"
      role="group"
      aria-label="Move (cardinal directions)"
    >
      <div />
      <button className={buttonClass} title="North" aria-label="Move north" onClick={() => actions.move(0, -1)}>
        ↑
      </button>
      <div />

      <button className={buttonClass} title="West" aria-label="Move west" onClick={() => actions.move(-1, 0)}>
        ←
      </button>
      <div className="w-10 h-10 flex items-center justify-center text-good">⚔️</div>
      <button className={buttonClass} title="East" aria-label="Move east" onClick={() => actions.move(1, 0)}>
        →
      </button>

      <div />
      <button className={buttonClass} title="South" aria-label="Move south" onClick={() => actions.move(0, 1)}>
        ↓
      </button>
      <div />
    </div>
  )
}
