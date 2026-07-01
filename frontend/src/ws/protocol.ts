// Mirrors backend/internal/wsproto/protocol.go — keep both in sync.
import type {
  Account,
  AttackRoll,
  ChatChannel,
  ChatMessage,
  ChoiceMode,
  ChoiceOption,
  ClassId,
  Combatant,
  CombatActionType,
  Dungeon,
  DungeonRoomType,
  Location,
  RaceId,
  Character,
  Skill,
  SkillCheckResultData,
} from '../engine/types'

// --- Outbound (client -> gateway) ---

export type OutboundEnvelope =
  | { type: 'TRAVEL'; payload: { toLocationId: string } }
  | { type: 'SWAP_CHARACTER'; payload: { characterId: string } }
  | { type: 'CREATE_CHARACTER'; payload: { name: string; raceId: RaceId; classId: ClassId } }
  | { type: 'RP_CHAT'; payload: { channel: ChatChannel; body: string } }
  | { type: 'INVITE_TO_PARTY'; payload: { targetDisplayName: string } }
  | { type: 'ACCEPT_PARTY_INVITE'; payload: { inviteId: string } }
  | { type: 'DECLINE_PARTY_INVITE'; payload: { inviteId: string } }
  | { type: 'LEAVE_PARTY'; payload: Record<string, never> }
  | { type: 'TALK_TO_NPC'; payload: Record<string, never> }
  | { type: 'MAKE_CHOICE'; payload: { promptId: string; optionId: string } }
  | { type: 'CAST_VOTE'; payload: { promptId: string; optionId: string } }
  | { type: 'ENTER_DUNGEON'; payload: Record<string, never> }
  | { type: 'START_ENCOUNTER'; payload: { roomType: DungeonRoomType; roomLabel?: string } }
  | { type: 'COMBAT_ACTION'; payload: { action: CombatActionType; targetId?: string } }
  | { type: 'SKILL_CHECK'; payload: { skill: Skill; context: string } }
  | { type: 'RESOLVE_DUNGEON'; payload: Record<string, never> }

// --- Inbound (gateway -> client) ---

export type StateSyncMessage = {
  type: 'STATE_SYNC'
  account: Account
  characters: Character[]
}

export type ChatBroadcastMessage = {
  type: 'CHAT_MESSAGE'
  message: ChatMessage
}

export type PresentAccount = {
  accountId: string
  displayName: string
}

export type LocationStateMessage = {
  type: 'LOCATION_STATE'
  location: Location
  present: PresentAccount[]
  narration?: string
}

export type PartyInviteReceivedMessage = {
  type: 'PARTY_INVITE_RECEIVED'
  inviteId: string
  fromAccountId: string
  fromDisplayName: string
}

export type PartyMemberData = {
  accountId: string
  displayName: string
  characterName?: string
  hpCurrent?: number
  hpMax?: number
}

export type PartyStateMessage = {
  type: 'PARTY_STATE'
  partyId?: string
  members: PartyMemberData[]
}

export type ChoiceStateMessage = {
  type: 'CHOICE_STATE'
  promptId: string
  prompt: string
  mode: ChoiceMode
  options: ChoiceOption[]
  deadline?: number // unix millis, party mode only
  narration?: string
}

export type VoteUpdateMessage = {
  type: 'VOTE_UPDATE'
  promptId: string
  tallies: Record<string, number>
}

export type VoteResolvedMessage = {
  type: 'VOTE_RESOLVED'
  promptId: string
  optionId: string
  honorDelta: number
  newHonor: number
  tieBreak: boolean
  narration?: string
}

export type DungeonReadyMessage = {
  type: 'DUNGEON_READY'
  dungeon: Dungeon
  narration?: string
}

// Broadcast to the whole party every time the turn order advances.
export type EncounterStateMessage = {
  type: 'ENCOUNTER_STATE'
  combatants: Combatant[]
  currentCombatantId?: string
  round: number
  log: AttackRoll[]
  roomType: DungeonRoomType
  roomLabel?: string
}

// Sent once a room's encounter ends — carries the final combat log so the
// UI can render the outcome.
export type RoomResolvedMessage = {
  type: 'ROOM_RESOLVED'
  roomType: DungeonRoomType
  victory: boolean
  combatLog: AttackRoll[]
  narration: string
  dungeon: Dungeon
}

// Tells the client the instance is fully cleared and it's safe to close
// the dungeon view and return to the world map.
export type DungeonResolvedMessage = {
  type: 'DUNGEON_RESOLVED'
  narration: string
  goldAwarded: number
}

export type SkillCheckResultMessage = {
  type: 'SKILL_CHECK_RESULT'
  result: SkillCheckResultData
  narration: string
}

export type ErrorMessage = {
  type: 'ERROR'
  message: string
}

export type InboundMessage =
  | StateSyncMessage
  | ChatBroadcastMessage
  | LocationStateMessage
  | PartyInviteReceivedMessage
  | PartyStateMessage
  | ChoiceStateMessage
  | VoteUpdateMessage
  | VoteResolvedMessage
  | DungeonReadyMessage
  | EncounterStateMessage
  | RoomResolvedMessage
  | DungeonResolvedMessage
  | SkillCheckResultMessage
  | ErrorMessage
