import { useState } from 'react'
import { useGame } from '../state/GameProvider'

export function Onboarding() {
  const { actions } = useGame()
  const [name, setName] = useState('')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    actions.beginAdventure(name.trim())
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="bg-panel border border-accent rounded-md p-8 max-w-sm w-full text-center">
        <h1 className="mb-1 tracking-wide">5e Virtual Tabletop</h1>
        <p className="text-sm text-[#b3a78c] mb-4">
          What should the table call you? Friends will invite you to their party by this name.
        </p>
        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <input
            autoFocus
            className="bg-[#0c0a08] text-parchment border border-accent rounded px-3 py-2 text-center"
            placeholder="Your name..."
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={24}
          />
          <button type="submit" className="bg-accent text-ink rounded px-4 py-2 font-bold">
            Begin Adventure
          </button>
        </form>
      </div>
    </div>
  )
}
