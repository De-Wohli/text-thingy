// Mirrors backend/internal/wsproto/protocol.go — keep both in sync.
import type {
  Account,
  AttackRoll,
  ChatChannel,
  ChatMessage,
  ChoiceMode,
  ChoiceOption,
  ClassId,
  Dungeon,
  DungeonRoomType,
  RaceId,
  Character,
} from '../engine/types'

// --- Outbound (client -> gateway) ---

export type OutboundEnvelope =
  | { type: 'MOVE'; payload: { dx: number; dy: number } }
  | { type: 'SWAP_CHARACTER'; payload: { characterId: string } }
  | { type: 'CREATE_CHARACTER'; payload: { name: string; raceId: RaceId; classId: ClassId } }
  | { type: 'RP_CHAT'; payload: { channel: ChatChannel; body: string } }
  | { type: 'TALK_TO_NPC'; payload: Record<string, never> }
  | { type: 'MAKE_CHOICE'; payload: { promptId: string; optionId: string } }
  | { type: 'CAST_VOTE'; payload: { promptId: string; optionId: string } }
  | { type: 'ENTER_POI'; payload: Record<string, never> }
  | { type: 'CLEAR_DUNGEON_ROOM'; payload: { roomType: DungeonRoomType } }
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

// Sent after CLEAR_DUNGEON_ROOM actually fights the encounter — carries the
// full attack-by-attack log so the UI can render a combat log.
export type RoomResolvedMessage = {
  type: 'ROOM_RESOLVED'
  roomType: DungeonRoomType
  victory: boolean
  combatLog: AttackRoll[]
  narration: string
  dungeon: Dungeon
}

// Tells the client the instance is fully cleared and it's safe to close the
// dungeon view and return to the overworld.
export type DungeonResolvedMessage = {
  type: 'DUNGEON_RESOLVED'
  narration: string
  goldAwarded: number
}

export type ErrorMessage = {
  type: 'ERROR'
  message: string
}

export type InboundMessage =
  | StateSyncMessage
  | ChatBroadcastMessage
  | ChoiceStateMessage
  | VoteUpdateMessage
  | VoteResolvedMessage
  | DungeonReadyMessage
  | RoomResolvedMessage
  | DungeonResolvedMessage
  | ErrorMessage
