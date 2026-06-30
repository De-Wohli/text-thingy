import type { AttackRoll, Account, ChatChannel, ChatMessage, ChoiceMode, ChoiceOption, Dungeon, DungeonRoomType } from '../engine/types'
import type { Character } from '../engine/types'
import type { ConnectionStatus } from '../ws/client'

export type View = 'overworld' | 'character-creation' | 'choice' | 'dungeon'

export type ChoiceState = {
  promptId: string
  prompt: string
  mode: ChoiceMode
  options: ChoiceOption[]
  deadline?: number
  narration?: string
}

export type VoteResolution = {
  optionId: string
  honorDelta: number
  newHonor: number
  tieBreak: boolean
  narration?: string
}

export type RoomResolution = {
  roomType: DungeonRoomType
  victory: boolean
  combatLog: AttackRoll[]
  narration: string
}

export type GameState = {
  connection: ConnectionStatus
  account: Account | null
  characters: Character[]
  view: View
  chatMessages: ChatMessage[]
  activeChatChannel: ChatChannel
  choice: ChoiceState | null
  voteTallies: Record<string, number> | null
  voteResolution: VoteResolution | null
  activeDungeon: Dungeon | null
  dungeonEntryNarration: string | null
  lastRoomResolution: RoomResolution | null
  lastMessage: string | null
}

const MAX_CHAT_HISTORY = 200

export function createInitialState(): GameState {
  return {
    connection: 'connecting',
    account: null,
    characters: [],
    view: 'overworld',
    chatMessages: [],
    activeChatChannel: 'global',
    choice: null,
    voteTallies: null,
    voteResolution: null,
    activeDungeon: null,
    dungeonEntryNarration: null,
    lastRoomResolution: null,
    lastMessage: null,
  }
}

export function appendChatMessage(messages: ChatMessage[], message: ChatMessage): ChatMessage[] {
  const next = [...messages, message]
  return next.length > MAX_CHAT_HISTORY ? next.slice(next.length - MAX_CHAT_HISTORY) : next
}
