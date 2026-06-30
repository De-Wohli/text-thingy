import { useState } from 'react'
import { useGame } from '../state/GameProvider'
import type { ChatChannel } from '../engine/types'

const CHANNELS: ChatChannel[] = ['global', 'guild', 'party', 'rp', 'narrator']

export function ChatPanel() {
  const { state, actions } = useGame()
  const [draft, setDraft] = useState('')
  const isNarratorTab = state.activeChatChannel === 'narrator'

  const visible = state.chatMessages.filter((m) => m.channel === state.activeChatChannel).slice(-50)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!draft.trim() || isNarratorTab) return
    actions.sendChat(state.activeChatChannel, draft.trim())
    setDraft('')
  }

  return (
    <section className="bg-panel border border-accent rounded p-4 flex flex-col gap-2 h-72">
      <div className="flex gap-1">
        {CHANNELS.map((channel) => (
          <button
            key={channel}
            onClick={() => actions.setChatChannel(channel)}
            className={`px-2 py-1 text-xs rounded font-mono ${
              channel === state.activeChatChannel ? 'bg-accent text-ink' : 'bg-[#2a2218] text-parchment'
            }`}
          >
            /{channel === 'narrator' ? 'gm' : channel}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto text-sm space-y-1 bg-[#0c0a08] rounded p-2">
        {visible.length === 0 && (
          <p className="text-[#8a7e63] italic">
            {isNarratorTab
              ? 'The Game Master has nothing to say yet — go explore.'
              : `No messages yet in /${state.activeChatChannel}.`}
          </p>
        )}
        {visible.map((msg, i) =>
          isNarratorTab ? (
            <p key={i} className="m-0 italic text-accent">
              {msg.body}
            </p>
          ) : (
            <p key={i} className="m-0">
              {msg.channel === 'rp' && msg.name ? (
                <span className="text-accent">
                  {msg.name} ({msg.race}/{msg.class}):{' '}
                </span>
              ) : (
                <span className="text-accent">{msg.accountId.slice(0, 8)}: </span>
              )}
              {msg.body}
            </p>
          ),
        )}
      </div>

      <form onSubmit={handleSubmit} className="flex gap-2">
        <input
          className="flex-1 bg-[#0c0a08] text-parchment border border-accent rounded px-2 py-1 font-mono text-sm disabled:opacity-50"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder={isNarratorTab ? 'The Game Master speaks here — you cannot reply' : `Message /${state.activeChatChannel}...`}
          maxLength={280}
          disabled={isNarratorTab}
        />
        <button type="submit" className="bg-accent text-ink rounded px-3 py-1 text-sm disabled:opacity-50" disabled={isNarratorTab}>
          Send
        </button>
      </form>
    </section>
  )
}
