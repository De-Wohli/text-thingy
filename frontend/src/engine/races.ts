import type { Race, RaceId } from './types'

export const RACES: Record<RaceId, Race> = {
  human: {
    id: 'human',
    name: 'Human',
    abilityBonuses: { str: 1, dex: 1, con: 1, int: 1, wis: 1, cha: 1 },
    traits: ['Versatile: +1 to all ability scores'],
  },
  tiefling: {
    id: 'tiefling',
    name: 'Tiefling',
    abilityBonuses: { cha: 2, int: 1 },
    traits: ['Darkvision (60ft)', 'Hellish Resistance (fire damage resistance)'],
  },
}

export function listRaces(): Race[] {
  return Object.values(RACES)
}
