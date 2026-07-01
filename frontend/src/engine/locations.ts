import type { LocationId, LocationKind } from './types'

// Cosmetic-only mirror of backend/internal/world — the mechanical truth
// (which locations exist, what connects to what) is always server-driven
// via LOCATION_STATE; this is purely for rendering labels/icons before
// the player has physically traveled there, and for the compass-grid layout.
export const LOCATION_LABELS: Record<LocationId, { name: string; icon: string; kind: LocationKind }> = {
  the_town: { name: 'The Town', icon: '🏘️', kind: 'hub' },
  guild_hall: { name: "Adventurer's Guild Hall", icon: '🏛️', kind: 'guild_hall' },
  tavern: { name: 'The Yawning Flask', icon: '🍺', kind: 'tavern' },
  market: { name: 'Market Square', icon: '🧙', kind: 'npc' },
  north_fields: { name: 'Northern Fields', icon: '🌾', kind: 'outdoor' },
  west_fields: { name: 'Western Fields', icon: '❓', kind: 'quest_hook' },
  east_river: { name: 'Eastern River', icon: '🌊', kind: 'water' },
  south_mountains: { name: 'Southern Mountains', icon: '⛰️', kind: 'barrier' },
}

export const LOCATION_KIND_COLOR: Record<LocationKind, string> = {
  hub: 'bg-[#3a3122]',
  guild_hall: 'bg-accent',
  tavern: 'bg-[#8a5a2b]',
  npc: 'bg-[#6b3fa0]',
  quest_hook: 'bg-evil',
  outdoor: 'bg-[#2a4018]',
  water: 'bg-[#1a3a5a]',
  barrier: 'bg-[#3a2218]',
}

// COMPASS_GRID defines where each location node sits in a 3×3 compass grid
// (row 0-2, col 0-2; center is 1,1) so WorldMap.tsx can render the geography
// directionally instead of just listing connection buttons in a flat row.
// Interior town sub-locations don't get grid positions (they're rendered
// differently, inside the town node).
export const COMPASS_GRID: Partial<Record<LocationId, { row: number; col: number }>> = {
  north_fields:    { row: 0, col: 1 },
  west_fields:     { row: 1, col: 0 },
  the_town:        { row: 1, col: 1 },
  east_river:      { row: 1, col: 2 },
  south_mountains: { row: 2, col: 1 },
}
