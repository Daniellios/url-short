import { useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({ component: App })

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL?.replace(/\/+$/, '') ?? 'http://localhost:8080'

function App() {
  const [longUrl, setLongUrl] = useState('')
  const [shortUrl, setShortUrl] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [copied, setCopied] = useState(false)

  const canSubmit = useMemo(() => longUrl.trim().length > 0 && !isSubmitting, [longUrl, isSubmitting])

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setShortUrl('')
    setCopied(false)

    let parsedURL: URL
    try {
      parsedURL = new URL(longUrl.trim())
    } catch {
      setError('Please enter a valid URL, including http:// or https://')
      return
    }

    setIsSubmitting(true)
    try {
      const response = await fetch(`${API_BASE_URL}/api/urls`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ original_url: parsedURL.toString() }),
      })

      const data = (await response.json()) as { short_url?: string; error?: string }

      if (!response.ok) {
        setError(data.error ?? 'Could not shorten this URL')
        return
      }

      if (!data.short_url) {
        setError('Server did not return a short URL')
        return
      }

      setShortUrl(data.short_url)
    } catch {
      setError('Request failed. Make sure the API server is running.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="page-wrap px-4 pb-8 pt-14">
      <section className="island-shell rise-in mx-auto max-w-2xl rounded-2xl p-6 sm:p-8">
        <h1 className="display-title mb-2 text-3xl sm:text-4xl">Paste a long URL</h1>
        <p className="mb-6 text-sm opacity-80">Generate a short link in one click.</p>

        <form className="flex flex-col gap-3" onSubmit={onSubmit}>
          <input
            id="long-url-input"
            type="url"
            placeholder="https://example.com/really/long/path"
            value={longUrl}
            onChange={(event) => {
              const nextValue = event.target.value

              setLongUrl(nextValue)
              if (nextValue.trim() === '') {
                setShortUrl('')
                setCopied(false)
              }
            }}
            className="w-full rounded-xl border border-(--line) bg-(--chip-bg) px-4 py-3 text-(--sea-ink) outline-none transition focus:border-(--lagoon)"
          />
          <button
            type="submit"
            disabled={!canSubmit}
            className="mt-1 inline-flex  items-center cursor-pointer justify-center rounded-xl border border-(--lagoon-deep) bg-(--lagoon-deep) px-4 py-2.5 font-semibold text-black
             disabled:cursor-not-allowed disabled:opacity-55 hover:text-black hover:bg-(--lagoon-deep)  "
          >
            {isSubmitting ? 'Shortening...' : 'Shorten URL'}
          </button>
        </form>

        {error ? (
          <p className="mt-4 rounded-lg border border-red-400/40 bg-red-500/10 px-3 py-2 text-sm text-red-200">
            {error}
          </p>
        ) : null}

        {shortUrl ? (
          <div className="mt-5 rounded-xl border border-(--chip-line) bg-(--chip-bg) p-4">
            <div className="mb-2 flex items-center justify-between gap-3">
              <p className="text-sm font-semibold">Short link</p>
              <button
                type="button"
                onClick={async () => {
                  try {
                    await navigator.clipboard.writeText(shortUrl)
                    setCopied(true)
                    window.setTimeout(() => setCopied(false), 1500)
                  } catch {
                    setError('Copy failed. Please copy manually.')
                  }
                }}
                className="rounded-lg border border-(--lagoon) cursor-pointer px-3 py-1 text-xs font-semibold text-(--lagoon) hover:bg-(--link-bg-hover)"
              >
                {copied ? 'Copied!' : 'Copy'}
              </button>
            </div>
            <a className="break-all text-base font-medium" href={shortUrl} target="_blank" rel="noreferrer">
              {shortUrl}
            </a>
          </div>
        ) : null}
      </section>
    </main>
  )
}
