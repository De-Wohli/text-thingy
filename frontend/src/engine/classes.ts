import type { Class, ClassId } from './types'

export const CLASSES: Record<ClassId, Class> = {
  fighter: {
    id: 'fighter',
    name: 'Fighter',
    hitDie: 10,
    proficiencies: ['All armor', 'Shields', 'Martial weapons'],
    features: ['Second Wind'],
  },
  wizard: {
    id: 'wizard',
    name: 'Wizard',
    hitDie: 6,
    proficiencies: ['Daggers', 'Darts', 'Slings', 'Quarterstaffs'],
    features: ['Arcane Recovery', 'Spellcasting'],
    cantripsKnown: 3,
    startingSpellSlots: { 1: 2 },
  },
}

export function listClasses(): Class[] {
  return Object.values(CLASSES)
}

// Preview-only: the gateway computes the authoritative max HP when a
// character is actually created (backend/cmd/gateway/character.go).
export function previewMaxHpAtLevel1(classId: ClassId, conModifier: number): number {
  return CLASSES[classId].hitDie + conModifier
}
