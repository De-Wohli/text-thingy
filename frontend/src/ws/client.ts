import type { InboundMessage, OutboundEnvelope } from './protocol'

export type ConnectionStatus = 'connecting' | 'open' | 'closed'

type Listeners = {
  onMessage: (msg: InboundMessage) => void
  onStatusChange: (status: ConnectionStatus) => void
}

const RECONNECT_DELAY_MS = 2000

export class GameSocket {
  private socket: WebSocket | null = null
  private accountId: string | null = null
  private listeners: Listeners
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private closedByCaller = false

  constructor(listeners: Listeners) {
    this.listeners = listeners
  }

  connect(accountId: string) {
    this.accountId = accountId
    this.closedByCaller = false
    this.open()
  }

  private open() {
    if (!this.accountId) return
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${protocol}://${window.location.host}/ws/${this.accountId}`
    this.listeners.onStatusChange('connecting')

    const socket = new WebSocket(url)
    this.socket = socket

    socket.onopen = () => this.listeners.onStatusChange('open')

    socket.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as InboundMessage
        this.listeners.onMessage(msg)
      } catch {
        // Ignore malformed frames rather than crashing the connection.
      }
    }

    socket.onclose = () => {
      this.listeners.onStatusChange('closed')
      if (!this.closedByCaller) {
        this.reconnectTimer = setTimeout(() => this.open(), RECONNECT_DELAY_MS)
      }
    }

    socket.onerror = () => socket.close()
  }

  send(envelope: OutboundEnvelope) {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(envelope))
    }
  }

  close() {
    this.closedByCaller = true
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.socket?.close()
  }
}
