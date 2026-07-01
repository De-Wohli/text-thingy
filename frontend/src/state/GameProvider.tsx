import { createContext, useContext, useEffect, useMemo, useReducer, useRef, type ReactNode } from 'react'
import { gameReducer } from './gameReducer'
import { createInitialState, type GameState } from './gameState'
import { GameSocket } from '../ws/client'
import type { OutboundEnvelope } from '../ws/protocol'
import type { ChatChannel, ClassId, CombatActionType, DungeonRoomType, LocationId, RaceId, Skill } from '../engine/types'

const ACCOUNT_STORAGE_KEY = 'dnd5e-web:accountId'

type GameActions = {
  beginAdventure: (displayName: string) => void
  travel: (toLocationId: LocationId) => void
  createCharacter: (name: string, raceId: RaceId, classId: ClassId) => void
  swapCharacter: (characterId: string) => void
  sendChat: (channel: ChatChannel, body: string) => void
  inviteToParty: (targetDisplayName: string) => void
  acceptPartyInvite: (inviteId: string) => void
  declinePartyInvite: (inviteId: string) => void
  leaveParty: () => void
  talkToNpc: () => void
  makeChoice: (promptId: string, optionId: string) => void
  castVote: (promptId: string, optionId: string) => void
  enterDungeon: () => void
  startEncounter: (roomType: DungeonRoomType, roomLabel?: string) => void
  combatAction: (action: CombatActionType, targetId?: string) => void
  skillCheck: (skill: Skill, context: string) => void
  resolveDungeon: () => void
  setView: (view: GameState['view']) => void
  setChatChannel: (channel: ChatChannel) => void
  dismissInvite: (inviteId: string) => void
  closePanel: () => void
}

type GameContextValue = {
  state: GameState
  actions: GameActions
}

const GameContext = createContext<GameContextValue | null>(null)

async function createAccount(displayName: string): Promise<string> {
  const response = await fetch('/api/accounts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ displayName }),
  })
  if (!response.ok) throw new Error('Failed to create account')
  const account = (await response.json()) as { id: string }
  localStorage.setItem(ACCOUNT_STORAGE_KEY, account.id)
  return account.id
}

export function GameProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(gameReducer, undefined, createInitialState)
  const socketRef = useRef<GameSocket | null>(null)

  useEffect(() => {
    const socket = new GameSocket({
      onMessage: (message) => dispatch({ type: 'INBOUND', message }),
      onStatusChange: (status) => dispatch({ type: 'CONNECTION_STATUS', status }),
    })
    socketRef.current = socket

    const existingAccountId = localStorage.getItem(ACCOUNT_STORAGE_KEY)
    if (existingAccountId) {
      socket.connect(existingAccountId)
    } else {
      // First visit on this browser/profile — ask who's playing instead of
      // silently creating an account named "Adventurer". Party invites are
      // by display name, so every friend needs a name worth typing.
      dispatch({ type: 'NEEDS_ONBOARDING' })
    }

    return () => socket.close()
  }, [])

  const send = (envelope: OutboundEnvelope) => socketRef.current?.send(envelope)

  const actions = useMemo<GameActions>(
    () => ({
      beginAdventure: (displayName) => {
        createAccount(displayName)
          .then((accountId) => {
            dispatch({ type: 'CONNECTION_STATUS', status: 'connecting' })
            socketRef.current?.connect(accountId)
          })
          .catch(() => dispatch({ type: 'CONNECTION_STATUS', status: 'closed' }))
      },
      travel: (toLocationId) => send({ type: 'TRAVEL', payload: { toLocationId } }),
      createCharacter: (name, raceId, classId) =>
        send({ type: 'CREATE_CHARACTER', payload: { name, raceId, classId } }),
      swapCharacter: (characterId) => send({ type: 'SWAP_CHARACTER', payload: { characterId } }),
      sendChat: (channel, body) => send({ type: 'RP_CHAT', payload: { channel, body } }),
      inviteToParty: (targetDisplayName) => send({ type: 'INVITE_TO_PARTY', payload: { targetDisplayName } }),
      acceptPartyInvite: (inviteId) => send({ type: 'ACCEPT_PARTY_INVITE', payload: { inviteId } }),
      declinePartyInvite: (inviteId) => send({ type: 'DECLINE_PARTY_INVITE', payload: { inviteId } }),
      leaveParty: () => send({ type: 'LEAVE_PARTY', payload: {} }),
      talkToNpc: () => send({ type: 'TALK_TO_NPC', payload: {} }),
      makeChoice: (promptId, optionId) => send({ type: 'MAKE_CHOICE', payload: { promptId, optionId } }),
      castVote: (promptId, optionId) => send({ type: 'CAST_VOTE', payload: { promptId, optionId } }),
      enterDungeon: () => send({ type: 'ENTER_DUNGEON', payload: {} }),
      startEncounter: (roomType, roomLabel) => send({ type: 'START_ENCOUNTER', payload: { roomType, roomLabel } }),
      combatAction: (action, targetId) => send({ type: 'COMBAT_ACTION', payload: { action, targetId } }),
      skillCheck: (skill, context) => send({ type: 'SKILL_CHECK', payload: { skill, context } }),
      resolveDungeon: () => send({ type: 'RESOLVE_DUNGEON', payload: {} }),
      setView: (view) => dispatch({ type: 'SET_VIEW', view }),
      setChatChannel: (channel) => dispatch({ type: 'SET_CHAT_CHANNEL', channel }),
      dismissInvite: (inviteId) => dispatch({ type: 'DISMISS_INVITE', inviteId }),
      closePanel: () => dispatch({ type: 'CLOSE_PANEL' }),
    }),
    [],
  )

  return <GameContext.Provider value={{ state, actions }}>{children}</GameContext.Provider>
}

export function useGame(): GameContextValue {
  const ctx = useContext(GameContext)
  if (!ctx) throw new Error('useGame must be used within a GameProvider')
  return ctx
}
