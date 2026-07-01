import type {
  Account,
  AttackRoll,
  ChatChannel,
  ChatMessage,
  ChoiceMode,
  ChoiceOption,
  Combatant,
  Dungeon,
  DungeonRoomType,
  Location,
  SkillCheckResultData,
} from '../engine/types'
import type { Character } from '../engine/types'
import type { PartyMemberData, PresentAccount } from '../ws/protocol'
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
  label: string
  victory: boolean
  combatLog: AttackRoll[]
  narration: string
}

export type EncounterState = {
  combatants: Combatant[]
  currentCombatantId?: string
  round: number
  log: AttackRoll[]
  roomType: DungeonRoomType
  roomLabel?: string
}

export type PartyInvite = {
  inviteId: string
  fromAccountId: string
  fromDisplayName: string
}

export type GameState = {
  connection: ConnectionStatus
  needsOnboarding: boolean
  account: Account | null
  characters: Character[]
  view: View
  chatMessages: ChatMessage[]
  activeChatChannel: ChatChannel
  location: Location | null
  presentAtLocation: PresentAccount[]
  party: PartyMemberData[]
  pendingInvites: PartyInvite[]
  choice: ChoiceState | null
  voteTallies: Record<string, number> | null
  voteResolution: VoteResolution | null
  activeDungeon: Dungeon | null
  dungeonEntryNarration: string | null
  activeEncounter: EncounterState | null
  lastRoomResolution: RoomResolution | null
  lastSkillCheck: { result: SkillCheckResultData; narration: string } | null
  lastMessage: string | null
}

const MAX_CHAT_HISTORY = 200

export function createInitialState(): GameState {
  return {
    connection: 'connecting',
    needsOnboarding: false,
    account: null,
    characters: [],
    view: 'overworld',
    chatMessages: [],
    activeChatChannel: 'global',
    location: null,
    presentAtLocation: [],
    party: [],
    pendingInvites: [],
    choice: null,
    voteTallies: null,
    voteResolution: null,
    activeDungeon: null,
    dungeonEntryNarration: null,
    activeEncounter: null,
    lastRoomResolution: null,
    lastSkillCheck: null,
    lastMessage: null,
  }
}

export function appendChatMessage(messages: ChatMessage[], message: ChatMessage): ChatMessage[] {
  const next = [...messages, message]
  return next.length > MAX_CHAT_HISTORY ? next.slice(next.length - MAX_CHAT_HISTORY) : next
}
