import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { BackButton } from '#/components/pastebin/PasteBackButton'
import { PasteCreateForm } from '#/components/pastebin/PasteCreateForm'
import { PastesList, type Paste } from '#/components/pastebin/PastesList'
import { getPastebinApiBaseUrl } from '#/env'

export const Route = createFileRoute('/pastebin/')({
  component: RouteComponent,
})

const PASTEBIN_API_BASE_URL = getPastebinApiBaseUrl()

function normalizePaste(p: Partial<Paste>): Paste {
  return {
    id: p.id ?? '',
    title: p.title ?? '',
    content: p.content ?? '',
    created_at: p.created_at ?? '',
    updated_at: p.updated_at ?? '',
    is_public: p.is_public !== false,
    expires_at: p.expires_at ?? null,
  }
}

function RouteComponent() {
  const navigate = useNavigate()
  const [pastes, setPastes] = useState<Paste[]>([])
  const [listLoading, setListLoading] = useState(true)

  const loadPastes = useCallback(async (options?: { showLoading?: boolean }) => {
    const showLoading = options?.showLoading ?? false
    if (showLoading) {
      setListLoading(true)
    }
    try {
      const response = await fetch(`${PASTEBIN_API_BASE_URL}/api/pastebin/pastes`)
      if (!response.ok) {
        toast.error('Could not load pastes.')
        return
      }
      const data = (await response.json()) as Partial<Paste>[]
      setPastes(Array.isArray(data) ? data.map(normalizePaste) : [])
    } catch {
      toast.error('Could not load pastes. Is the pastebin API running?')
    } finally {
      if (showLoading) {
        setListLoading(false)
      }
    }
  }, [])

  useEffect(() => {
    void loadPastes({ showLoading: true })
  }, [loadPastes])

  const createPaste = useCallback(
    async ({
      title,
      content,
      is_public,
      expires_at,
    }: {
      title: string
      content: string
      is_public: boolean
      expires_at: string | null
    }) => {
    if (!content.trim()) {
      toast.error('Paste content is required.')
      return false
    }

    try {
      const response = await fetch(`${PASTEBIN_API_BASE_URL}/api/pastebin/pastes`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title,
          content,
          is_public,
          ...(expires_at != null ? { expires_at } : {}),
        }),
      })
      const data = (await response.json()) as Partial<Paste> & { error?: string }

      if (!response.ok) {
        const message = data.error ?? 'Could not create paste.'
        toast.error(message)
        return false
      }

      if (data.id && data.content !== undefined) {
        navigate({ to: `/pastebin/${data.id}` })
        toast.success('Paste created successfully.')
        return true
      } else {
        await loadPastes()
        toast.success('Paste created successfully.')
        return true
      }
    } catch {
      const message = 'Request failed. Make sure the pastebin API is reachable.'
      toast.error(message)
      return false
    }
  },
  [loadPastes, navigate],
)

  return (
    <main className="page-wrap page-pastebin px-4 pb-8 pt-14">
      <div className="mx-auto grid w-full max-w-6xl gap-6 lg:w-fit lg:grid-cols-[640px_minmax(0,24rem)]">
        <div className="min-w-0 space-y-3">
          <BackButton to="/" />

          <section className="island-shell w-full rounded-2xl p-6 sm:p-8 sm:pl-0">
            <h1 className="display-title mb-2 text-3xl sm:text-4xl">Paste a text snippet</h1>
            <p className="mb-6 text-sm opacity-80">Share text snippets with others.</p>

            <PasteCreateForm onCreate={createPaste} />
          </section>
        </div>

        <div className="min-w-0 lg:sticky lg:top-6 lg:self-start">
          <PastesList pastes={pastes} isLoading={listLoading} />
        </div>
      </div>
    </main>
  )
}
