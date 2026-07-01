import type { LocationId, LocationKind } from './types'

// Cosmetic-only mirror of backend/internal/world's location names/icons —
// the mechanical truth (which locations exist, what connects to what) is
// always server-driven via LOCATION_STATE; this is purely for rendering
// labels/icons on connection buttons before the player has traveled there.
export const LOCATION_LABELS: Record<LocationId, { name: string; icon: string; kind: LocationKind }> = {
  town_square: { name: 'Town Square', icon: '\u{26F2}', kind: 'hub' },
  guild_hall: { name: "Adventurer's Guild Hall", icon: '\u{1F3DB}\u{FE0F}', kind: 'guild_hall' },
  tavern: { name: 'The Yawning Flask', icon: '\u{1F37A}', kind: 'tavern' },
  market: { name: 'Market Square', icon: '\u{1F9D9}', kind: 'npc' },
  mine_entrance: { name: 'Old Mine Entrance', icon: '\u{2753}', kind: 'quest_hook' },
}

export const LOCATION_KIND_COLOR: Record<LocationKind, string> = {
  hub: 'bg-[#3a3122]',
  guild_hall: 'bg-accent',
  tavern: 'bg-[#8a5a2b]',
  npc: 'bg-[#6b3fa0]',
  quest_hook: 'bg-evil',
}
