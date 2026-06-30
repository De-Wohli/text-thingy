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

export type Coordinate = { x: number; y: number }

export type Account = {
  id: string
  displayName: string
  honor: number
  gold: number
  activeCharacterId: string | null
  coordinate: Coordinate
  partyId: string | null
}

export type TileType = 'floor' | 'wall' | 'water' | 'guild' | 'tavern' | 'npc' | 'poi'

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

export type DungeonRoomType = 'start' | 'hallway' | 'treasure' | 'boss'

export type DungeonRoom = {
  type: DungeonRoomType
  x: number
  y: number
  width: number
  height: number
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

export type DungeonTile = 'wall' | 'floor'

export type Dungeon = {
  id: string
  partyId: string
  width: number
  height: number
  grid: DungeonTile[][]
  rooms: DungeonRoom[]
  encounters: DungeonEncounter[]
  resolved: boolean
}
