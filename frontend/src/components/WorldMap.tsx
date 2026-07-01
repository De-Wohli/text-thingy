import { useGame } from '../state/GameProvider'
import { LOCATION_LABELS, LOCATION_KIND_COLOR } from '../engine/locations'

// The map is mostly a visual aid, not a tactical grid: it shows where you
// are and what's directly reachable, and clicking a connected location is
// how you travel there. See outline.md's "Virtual tabletop" note.
export function WorldMap() {
  const { state, actions } = useGame()
  const location = state.location

  if (!location) {
    return (
      <div className="bg-panel border border-accent rounded p-4 text-sm text-[#8a7e63]">Finding your place in the world...</div>
    )
  }

  const here = LOCATION_LABELS[location.id]

  return (
    <div className="bg-panel border border-accent rounded p-4">
      <h3 className="m-0 mb-3 text-sm uppercase tracking-wide text-accent">World Map</h3>
      <div className="flex flex-col items-center gap-2">
        <div className={`flex flex-col items-center gap-1 rounded-lg border-2 border-good px-4 py-3 ${LOCATION_KIND_COLOR[location.kind]}`}>
          <span className="text-2xl" aria-hidden>
            {here?.icon ?? '\u{1F4CD}'}
          </span>
          <strong>{location.name}</strong>
          <span className="text-xs text-[#b3a78c]">You are here</span>
        </div>

        {location.connections.length > 0 && (
          <>
            <span className="text-accent text-xl" aria-hidden>
              ↓
            </span>
            <div className="flex flex-wrap justify-center gap-3">
              {location.connections.map((connId) => {
                const label = LOCATION_LABELS[connId]
                return (
                  <button
                    key={connId}
                    onClick={() => actions.travel(connId)}
                    className={`flex flex-col items-center gap-1 rounded-lg border border-accent px-3 py-2 hover:brightness-110 ${LOCATION_KIND_COLOR[label?.kind ?? 'hub']}`}
                  >
                    <span className="text-xl" aria-hidden>
                      {label?.icon ?? '\u{1F4CD}'}
                    </span>
                    <span className="text-xs">{label?.name ?? connId}</span>
                  </button>
                )
              })}
            </div>
          </>
        )}
      </div>
      <p className="text-sm italic text-[#b3a78c] mt-3 mb-0">{location.description}</p>
    </div>
  )
}
