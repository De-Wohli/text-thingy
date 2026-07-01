// Wire-format types mirroring backend/internal/models/models.go. The Go
// backend is the source of truth for game state; keep these two in sync.

export type RaceId = 'human' | 'tiefling'
export type ClassId = 'fighter' | 'wizard'

export type AbilityScores = {
  str: number
  dex: number
  con: number
  int: number
  wis: number
  cha: number
}

export type Race = {
  id: RaceId
  name: string
  abilityBonuses: Partial<AbilityScores>
  traits: string[]
}

export type DamageDie = 6 | 10

export type SpellSlots = Record<number, number>

export type Class = {
  id: ClassId
  name: string
  hitDie: DamageDie
  proficiencies: string[]
  features: string[]
  cantripsKnown?: number
  startingSpellSlots?: SpellSlots
}

export type CharacterStatus = 'IDLE' | 'QUESTING' | 'CRAFTING'

export type Character = {
  id: string
  accountId: string
  name: string
  raceId: RaceId
  classId: ClassId
  level: number
  status: CharacterStatus
  hpCurrent: number
  hpMax: number
  abilityScores: AbilityScores
  createdAt: string
}

export type AlignmentBand = 'good' | 'neutral' | 'evil'

// A node in the world's location graph (see backend/internal/world) — the
// overworld is a small hub-and-spoke graph, not a tile grid.
export type LocationId = string

export type Account = {
  id: string
  displayName: string
  honor: number
  gold: number
  activeCharacterId: string | null
  locationId: LocationId
  partyId: string | null
}

export type LocationKind = 'hub' | 'guild_hall' | 'tavern' | 'npc' | 'quest_hook'

export type Location = {
  id: LocationId
  name: string
  description: string
  kind: LocationKind
  connections: LocationId[]
}

export type ChatChannel = 'global' | 'guild' | 'party' | 'rp' | 'narrator'

export type ChatMessage = {
  channel: ChatChannel
  accountId: string
  name?: string
  race?: string
  class?: string
  body: string
  timestamp: string
}

export type ChoiceTypology = 'merciful' | 'pragmatic' | 'ruthless'

export type ChoiceOption = {
  id: string
  label: string
  typology: ChoiceTypology
}

export type ChoiceMode = 'solo' | 'party'

export type Skill = 'perception' | 'investigation' | 'insight' | 'stealth' | 'arcana' | 'athletics'

export type DungeonRoomType = 'start' | 'hallway' | 'treasure' | 'boss'

export type DungeonRoom = {
  type: DungeonRoomType
  cleared: boolean
}

export type Monster = {
  id: string
  name: string
  challengeRating: number
  armorClass: number
  hp: number
  attackBonus: number
  damageDie: string
}

export type AttackRoll = {
  attacker: string
  target: string
  d20: number
  attackBonus: number
  total: number
  targetAc: number
  hit: boolean
  critical: boolean
  damage: number
}

export type DungeonEncounter = {
  roomType: DungeonRoomType
  monsters: Monster[]
}

export type Dungeon = {
  id: string
  partyId: string
  rooms: DungeonRoom[]
  encounters: DungeonEncounter[]
  resolved: boolean
}

export type CombatantKind = 'player' | 'monster'

export type Combatant = {
  id: string
  kind: CombatantKind
  accountId?: string
  name: string
  initiative: number
  hp: number
  maxHp: number
  ac: number
  attackBonus: number
  dodging: boolean
  fled: boolean
  defeated: boolean
}

export type CombatActionType = 'attack' | 'dodge' | 'flee'

export type SkillCheckResultData = {
  skill: Skill
  d20: number
  abilityModifier: number
  proficiencyBonus: number
  total: number
  dc: number
  proficient: boolean
  success: boolean
}
