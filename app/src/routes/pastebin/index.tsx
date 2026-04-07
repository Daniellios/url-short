import { createFileRoute } from '@tanstack/react-router'
import { useEffect, useState } from 'react'

export const Route = createFileRoute('/pastebin/')({
  component: RouteComponent,
})

const PASTEBIN_API_BASE_URL = import.meta.env.VITE_PASTEBIN_API_BASE_URL

function RouteComponent() {
  const [pasteContent, setPasteContent] = useState<any[]>([])


  useEffect(() => {
    fetch(`${PASTEBIN_API_BASE_URL}/api/pastebin/pastes`)
      .then(response => response.json())
      .then(data => {
        setPasteContent(data)
      })
      .catch(error => {
        console.error('Error fetching pastes:', error)
      })
  }, [])


  return <main className="page-wrap px-4 pb-8 pt-14">
    <section className="island-shell rise-in mx-auto max-w-2xl rounded-2xl p-6 sm:p-8">
      <h1 className="display-title mb-2 text-3xl sm:text-4xl">Paste a text snippet</h1>
      <p className="mb-6 text-sm opacity-80">Share text snippets with others.</p>
    </section>

    {pasteContent.map((paste: any) => (
      <div key={paste.id}>
        <h2>{paste.content}</h2>
      </div>
    ))}
  </main>
}
