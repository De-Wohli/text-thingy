import type { AlignmentBand } from './types'

// Display-only mirror of backend/internal/honor: the gateway is the
// authority on Honor mutations (ApplyChoice / ApplyBetrayal write to
// Postgres), this module just maps a score to UI copy/coloring.
export const HONOR_MIN = -100
export const HONOR_MAX = 100

export type ReactivityProfile = {
  alignment: AlignmentBand
  greeting: string
  shopPriceModifier: number
  guardBehavior: string
  questTypes: string[]
}

const BANDS: { min: number; max: number; profile: ReactivityProfile }[] = [
  {
    min: 60,
    max: 100,
    profile: {
      alignment: 'good',
      greeting: 'Cheerful greetings',
      shopPriceModifier: -0.1,
      guardBehavior: 'Guards assist in fights',
      questTypes: ['Rescue civilians', 'Defend caravans', 'Holy relic recovery'],
    },
  },
  {
    min: -59,
    max: 59,
    profile: {
      alignment: 'neutral',
      greeting: 'Standard, professional dialogue',
      shopPriceModifier: 0,
      guardBehavior: 'Default market prices',
      questTypes: ['Monster cull', 'Bounty hunting', 'Material gathering'],
    },
  },
  {
    min: -100,
    max: -60,
    profile: {
      alignment: 'evil',
      greeting: 'Suspicious/hostile dialogue',
      shopPriceModifier: 0.2,
      guardBehavior: 'Guards follow closely',
      questTypes: ['Smuggling', 'Infiltrate rival hideouts', 'Assassination'],
    },
  },
]

export function clampHonor(value: number): number {
  return Math.min(HONOR_MAX, Math.max(HONOR_MIN, value))
}

export function reactivityForHonor(honor: number): ReactivityProfile {
  const band = BANDS.find((b) => honor >= b.min && honor <= b.max)
  if (!band) throw new Error(`No alignment band covers honor score ${honor}`)
  return band.profile
}
