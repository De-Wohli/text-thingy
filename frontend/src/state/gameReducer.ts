import type { ChatChannel } from '../engine/types'
import type { ConnectionStatus } from '../ws/client'
import type { InboundMessage } from '../ws/protocol'
import { appendChatMessage, type GameState, type View } from './gameState'

export type GameAction =
  | { type: 'CONNECTION_STATUS'; status: ConnectionStatus }
  | { type: 'NEEDS_ONBOARDING' }
  | { type: 'INBOUND'; message: InboundMessage }
  | { type: 'SET_VIEW'; view: View }
  | { type: 'SET_CHAT_CHANNEL'; channel: ChatChannel }
  | { type: 'DISMISS_INVITE'; inviteId: string }
  | { type: 'CLOSE_PANEL' }

export function gameReducer(state: GameState, action: GameAction): GameState {
  switch (action.type) {
    case 'CONNECTION_STATUS':
      return { ...state, connection: action.status }

    case 'NEEDS_ONBOARDING':
      return { ...state, needsOnboarding: true }

    case 'SET_VIEW':
      return { ...state, view: action.view }

    case 'SET_CHAT_CHANNEL':
      return { ...state, activeChatChannel: action.channel }

    case 'DISMISS_INVITE':
      return { ...state, pendingInvites: state.pendingInvites.filter((i) => i.inviteId !== action.inviteId) }

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

    case 'LOCATION_STATE':
      return { ...state, location: message.location, presentAtLocation: message.present }

    case 'PARTY_INVITE_RECEIVED':
      return {
        ...state,
        pendingInvites: [
          ...state.pendingInvites,
          { inviteId: message.inviteId, fromAccountId: message.fromAccountId, fromDisplayName: message.fromDisplayName },
        ],
      }

    case 'PARTY_STATE':
      return { ...state, party: message.members }

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

    case 'ENCOUNTER_STATE':
      return {
        ...state,
        view: 'dungeon',
        activeEncounter: {
          combatants: message.combatants,
          currentCombatantId: message.currentCombatantId,
          round: message.round,
          log: message.log,
          roomType: message.roomType,
          roomLabel: message.roomLabel,
        },
      }

    case 'ROOM_RESOLVED': {
      // Extract the label from the most recently cleared room of this type
      // in the updated dungeon state (multiple rooms may share the same
      // functional type, e.g. two hallway rooms, so we take the last cleared).
      const clearedRooms = message.dungeon.rooms.filter(
        (r) => r.type === message.roomType && r.cleared,
      )
      const resolvedRoom = clearedRooms[clearedRooms.length - 1]
      return {
        ...state,
        activeDungeon: message.dungeon,
        activeEncounter: null,
        lastRoomResolution: {
          roomType: message.roomType,
          label: resolvedRoom?.label ?? message.roomType,
          victory: message.victory,
          combatLog: message.combatLog,
          narration: message.narration,
        },
      }
    }

    case 'DUNGEON_RESOLVED':
      return {
        ...state,
        view: 'overworld',
        activeDungeon: null,
        activeEncounter: null,
        lastRoomResolution: null,
        lastMessage: message.narration,
      }

    case 'SKILL_CHECK_RESULT': {
      const updatedCooldowns = { ...state.skillCooldowns }
      if (message.result.cooldownSeconds > 0) {
        // Key on the skill name so all contexts for that skill are locked
        // together — consistent with the backend's per-skill-per-room logic.
        updatedCooldowns[String(message.result.skill)] = Date.now() + message.result.cooldownSeconds * 1000
      } else {
        // Success clears any existing cooldown for this skill.
        delete updatedCooldowns[String(message.result.skill)]
      }
      return { ...state, lastSkillCheck: { result: message.result, narration: message.narration }, skillCooldowns: updatedCooldowns }
    }

    case 'ERROR':
      return { ...state, lastMessage: message.message }

    default:
      return state
  }
}
