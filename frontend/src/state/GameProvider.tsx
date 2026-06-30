import { createContext, useContext, useEffect, useMemo, useReducer, useRef, type ReactNode } from 'react'
import { gameReducer } from './gameReducer'
import { createInitialState, type GameState } from './gameState'
import { GameSocket } from '../ws/client'
import type { OutboundEnvelope } from '../ws/protocol'
import type { ChatChannel, ClassId, DungeonRoomType, RaceId } from '../engine/types'

const ACCOUNT_STORAGE_KEY = 'dnd5e-web:accountId'

type GameActions = {
  move: (dx: number, dy: number) => void
  createCharacter: (name: string, raceId: RaceId, classId: ClassId) => void
  swapCharacter: (characterId: string) => void
  sendChat: (channel: ChatChannel, body: string) => void
  talkToNpc: () => void
  makeChoice: (promptId: string, optionId: string) => void
  castVote: (promptId: string, optionId: string) => void
  enterPOI: () => void
  clearDungeonRoom: (roomType: DungeonRoomType) => void
  resolveDungeon: () => void
  setView: (view: GameState['view']) => void
  setChatChannel: (channel: ChatChannel) => void
  closePanel: () => void
}

type GameContextValue = {
  state: GameState
  actions: GameActions
}

const GameContext = createContext<GameContextValue | null>(null)

async function ensureAccountId(): Promise<string> {
  const existing = localStorage.getItem(ACCOUNT_STORAGE_KEY)
  if (existing) return existing

  const response = await fetch('/api/accounts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ displayName: 'Adventurer' }),
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
    let cancelled = false

    const socket = new GameSocket({
      onMessage: (message) => dispatch({ type: 'INBOUND', message }),
      onStatusChange: (status) => dispatch({ type: 'CONNECTION_STATUS', status }),
    })
    socketRef.current = socket

    ensureAccountId()
      .then((accountId) => {
        if (!cancelled) socket.connect(accountId)
      })
      .catch(() => {
        if (!cancelled) dispatch({ type: 'CONNECTION_STATUS', status: 'closed' })
      })

    return () => {
      cancelled = true
      socket.close()
    }
  }, [])

  const send = (envelope: OutboundEnvelope) => socketRef.current?.send(envelope)

  const actions = useMemo<GameActions>(
    () => ({
      move: (dx, dy) => send({ type: 'MOVE', payload: { dx, dy } }),
      createCharacter: (name, raceId, classId) =>
        send({ type: 'CREATE_CHARACTER', payload: { name, raceId, classId } }),
      swapCharacter: (characterId) => send({ type: 'SWAP_CHARACTER', payload: { characterId } }),
      sendChat: (channel, body) => send({ type: 'RP_CHAT', payload: { channel, body } }),
      talkToNpc: () => send({ type: 'TALK_TO_NPC', payload: {} }),
      makeChoice: (promptId, optionId) => send({ type: 'MAKE_CHOICE', payload: { promptId, optionId } }),
      castVote: (promptId, optionId) => send({ type: 'CAST_VOTE', payload: { promptId, optionId } }),
      enterPOI: () => send({ type: 'ENTER_POI', payload: {} }),
      clearDungeonRoom: (roomType) => send({ type: 'CLEAR_DUNGEON_ROOM', payload: { roomType } }),
      resolveDungeon: () => send({ type: 'RESOLVE_DUNGEON', payload: {} }),
      setView: (view) => dispatch({ type: 'SET_VIEW', view }),
      setChatChannel: (channel) => dispatch({ type: 'SET_CHAT_CHANNEL', channel }),
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
