import { useGame } from '../state/GameProvider'
import { COMPASS_GRID, LOCATION_KIND_COLOR, LOCATION_LABELS } from '../engine/locations'
import type { LocationId } from '../engine/types'

// Locations with compass grid positions are rendered in a 3×3 grid so
// north/south/east/west geography is visually obvious.
// Interior sub-locations (guild hall, tavern, market) have no compass
// position and are listed separately as "places in town."
function CompassMap({ currentId, connections, onTravel }: {
  currentId: LocationId
  connections: LocationId[]
  onTravel: (id: LocationId) => void
}) {
  const cells: Record<string, LocationId | null> = {}
  // Seed the grid with all known compass locations
  for (const [id, pos] of Object.entries(COMPASS_GRID)) {
    if (!pos) continue
    cells[`${pos.row}-${pos.col}`] = id as LocationId
  }
  const rows = [0, 1, 2]
  const cols = [0, 1, 2]

  return (
    <div className="grid gap-1" style={{ gridTemplateColumns: 'repeat(3,1fr)', gridTemplateRows: 'repeat(3,auto)' }}>
      {rows.map((row) =>
        cols.map((col) => {
          const cellId = cells[`${row}-${col}`]
          if (!cellId) return <div key={`${row}-${col}`} />
          const label = LOCATION_LABELS[cellId]
          const isCurrent = cellId === currentId
          const isConnection = connections.includes(cellId)

          if (isCurrent) {
            return (
              <div
                key={cellId}
                className={`flex flex-col items-center gap-0.5 rounded-lg border-2 border-good px-2 py-2 text-center ${LOCATION_KIND_COLOR[label?.kind ?? 'hub']}`}
              >
                <span className="text-xl" aria-hidden>{label?.icon ?? '📍'}</span>
                <span className="text-xs font-bold leading-tight">{label?.name ?? cellId}</span>
                <span className="text-[0.6rem] text-good">Here</span>
              </div>
            )
          }
          if (isConnection) {
            return (
              <button
                key={cellId}
                onClick={() => onTravel(cellId)}
                className={`flex flex-col items-center gap-0.5 rounded-lg border border-accent px-2 py-2 text-center hover:brightness-110 ${LOCATION_KIND_COLOR[label?.kind ?? 'hub']}`}
              >
                <span className="text-xl" aria-hidden>{label?.icon ?? '📍'}</span>
                <span className="text-xs leading-tight">{label?.name ?? cellId}</span>
              </button>
            )
          }
          // Visible on the map but not directly reachable from current location
          return (
            <div
              key={cellId}
              className={`flex flex-col items-center gap-0.5 rounded-lg border border-dashed border-[#4a3f2c] px-2 py-2 text-center opacity-40 ${LOCATION_KIND_COLOR[label?.kind ?? 'hub']}`}
            >
              <span className="text-xl" aria-hidden>{label?.icon ?? '📍'}</span>
              <span className="text-xs leading-tight">{label?.name ?? cellId}</span>
            </div>
          )
        }),
      )}
    </div>
  )
}

export function WorldMap() {
  const { state, actions } = useGame()
  const location = state.location

  if (!location) {
    return (
      <div className="bg-panel border border-accent rounded p-4 text-sm text-[#8a7e63]">
        Finding your place in the world...
      </div>
    )
  }

  const hasCompassPosition = !!COMPASS_GRID[location.id]

  // Interior sub-locations (guild hall, tavern, market) that have no compass
  // position — show as a simple list of town services to return to.
  const interiorConnections = location.connections.filter((id) => !COMPASS_GRID[id])

  return (
    <div className="bg-panel border border-accent rounded p-4">
      <h3 className="m-0 mb-3 text-sm uppercase tracking-wide text-accent">World Map</h3>

      {hasCompassPosition ? (
        <CompassMap
          currentId={location.id}
          connections={location.connections}
          onTravel={actions.travel}
        />
      ) : (
        /* Interior location: show a compact current-location card + back-to-town */
        <div className="flex flex-col items-center gap-2">
          <div className={`flex flex-col items-center gap-1 rounded-lg border-2 border-good px-4 py-3 ${LOCATION_KIND_COLOR[location.kind]}`}>
            <span className="text-2xl" aria-hidden>
              {LOCATION_LABELS[location.id]?.icon ?? '📍'}
            </span>
            <strong>{location.name}</strong>
            <span className="text-xs text-[#b3a78c]">You are here</span>
          </div>
          {location.connections.map((connId) => {
            const label = LOCATION_LABELS[connId]
            return (
              <button
                key={connId}
                onClick={() => actions.travel(connId)}
                className={`flex flex-col items-center gap-1 rounded-lg border border-accent px-3 py-2 hover:brightness-110 ${LOCATION_KIND_COLOR[label?.kind ?? 'hub']}`}
              >
                <span className="text-xl" aria-hidden>
                  {label?.icon ?? '📍'}
                </span>
                <span className="text-xs">{label?.name ?? connId}</span>
              </button>
            )
          })}
        </div>
      )}

      {/* Interior services when in the town hub */}
      {interiorConnections.length > 0 && (
        <div className="mt-3">
          <p className="text-xs uppercase tracking-wide text-accent mb-1">Places in town</p>
          <div className="flex flex-wrap gap-2">
            {interiorConnections.map((connId) => {
              const label = LOCATION_LABELS[connId]
              return (
                <button
                  key={connId}
                  onClick={() => actions.travel(connId)}
                  className={`flex items-center gap-1.5 rounded border border-accent px-2 py-1 text-xs hover:brightness-110 ${LOCATION_KIND_COLOR[label?.kind ?? 'hub']}`}
                >
                  <span aria-hidden>{label?.icon}</span>
                  {label?.name ?? connId}
                </button>
              )
            })}
          </div>
        </div>
      )}

      <p className="text-sm italic text-[#b3a78c] mt-3 mb-0">{location.description}</p>
    </div>
  )
}
