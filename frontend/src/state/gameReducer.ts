import type { ChatChannel } from '../engine/types'
import type { ConnectionStatus } from '../ws/client'
import type { InboundMessage } from '../ws/protocol'
import { appendChatMessage, type GameState, type View } from './gameState'

export type GameAction =
  | { type: 'CONNECTION_STATUS'; status: ConnectionStatus }
  | { type: 'INBOUND'; message: InboundMessage }
  | { type: 'SET_VIEW'; view: View }
  | { type: 'SET_CHAT_CHANNEL'; channel: ChatChannel }
  | { type: 'CLOSE_PANEL' }

export function gameReducer(state: GameState, action: GameAction): GameState {
  switch (action.type) {
    case 'CONNECTION_STATUS':
      return { ...state, connection: action.status }

    case 'SET_VIEW':
      return { ...state, view: action.view }

    case 'SET_CHAT_CHANNEL':
      return { ...state, activeChatChannel: action.channel }

    case 'CLOSE_PANEL':
      return { ...state, view: 'overworld', choice: null, voteTallies: null, lastMessage: null }

    case 'INBOUND':
      return applyInbound(state, action.message)

    default:
      return state
  }
}

function applyInbound(state: GameState, message: InboundMessage): GameState {
  switch (message.type) {
    case 'STATE_SYNC':
      return { ...state, account: message.account, characters: message.characters }

    case 'CHAT_MESSAGE':
      return { ...state, chatMessages: appendChatMessage(state.chatMessages, message.message) }

    case 'CHOICE_STATE':
      return {
        ...state,
        view: 'choice',
        choice: {
          promptId: message.promptId,
          prompt: message.prompt,
          mode: message.mode,
          options: message.options,
          deadline: message.deadline,
          narration: message.narration,
        },
        voteTallies: null,
        voteResolution: null,
      }

    case 'VOTE_UPDATE':
      if (state.choice?.promptId !== message.promptId) return state
      return { ...state, voteTallies: message.tallies }

    case 'VOTE_RESOLVED':
      if (state.choice?.promptId !== message.promptId) return state
      return {
        ...state,
        voteResolution: {
          optionId: message.optionId,
          honorDelta: message.honorDelta,
          newHonor: message.newHonor,
          tieBreak: message.tieBreak,
          narration: message.narration,
        },
        account: state.account ? { ...state.account, honor: message.newHonor } : state.account,
      }

    case 'DUNGEON_READY':
      return {
        ...state,
        view: 'dungeon',
        activeDungeon: message.dungeon,
        dungeonEntryNarration: message.narration ?? null,
        lastRoomResolution: null,
      }

    case 'ROOM_RESOLVED':
      return {
        ...state,
        activeDungeon: message.dungeon,
        lastRoomResolution: {
          roomType: message.roomType,
          victory: message.victory,
          combatLog: message.combatLog,
          narration: message.narration,
        },
      }

    case 'DUNGEON_RESOLVED':
      return {
        ...state,
        view: 'overworld',
        activeDungeon: null,
        lastRoomResolution: null,
        lastMessage: message.narration,
      }

    case 'ERROR':
      return { ...state, lastMessage: message.message }

    default:
      return state
  }
}
