import { createFileRoute } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { BackButton } from '#/components/pastebin/PasteBackButton'
import { PasteItemControls } from '#/components/pastebin/PasteItemControls'
import { PasteMeta } from '#/components/pastebin/PasteMeta'
import { Textarea } from '#/components/ui/textarea'
import { getPastebinApiBaseUrl } from '#/env'
import type { Paste } from '#/components/pastebin/PastesList'

export const Route = createFileRoute('/pastebin/$id')({
  component: PasteDetailsPage,
})

const PASTEBIN_API_BASE_URL = getPastebinApiBaseUrl()
const PASTEBIN_INDEX_OUTER_MAX = 'max-w-[calc(640px+1.5rem+24rem)]'

function PasteDetailsPage() {
  const { id } = Route.useParams()
  const [paste, setPaste] = useState<Paste | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false

    async function loadPaste() {
      setLoading(true)
      try {
        const response = await fetch(`${PASTEBIN_API_BASE_URL}/api/pastebin/pastes/${id}`)
        const data = (await response.json()) as Partial<Paste> & { error?: string }

        if (!response.ok) {
          toast.error(data.error ?? 'Could not load paste.')
          return
        }

        if (!cancelled) {
          setPaste({
            id: data.id ?? id,
            title: data.title ?? '',
            content: data.content ?? '',
            created_at: data.created_at ?? '',
            updated_at: data.updated_at ?? '',
            is_public: data.is_public !== false,
            expires_at: data.expires_at ?? null,
          })
        }
      } catch {
        toast.error('Request failed. Make sure the pastebin API is reachable.')
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadPaste()
    return () => {
      cancelled = true
    }
  }, [id])

  return (
    <main className="page-wrap page-pastebin px-4 pb-8 pt-14">
      <div className={`mx-auto w-full space-y-3 ${PASTEBIN_INDEX_OUTER_MAX}`}>
        <BackButton to="/pastebin" />

        <section className="island-shell w-full rounded-2xl p-6 sm:p-8 sm:pl-0">
          <h1 className="display-title mb-5 text-2xl sm:text-3xl">{paste?.title || 'Untitled paste'}</h1>
          <PasteMeta
            content={paste?.content ?? ''}
            createdAt={paste?.created_at ?? ''}
            expiresAt={paste?.expires_at}
            className="mb-4 flex items-center justify-between gap-2 text-xs text-muted-foreground"
          />

          {loading ? (
            <p className="text-sm text-muted-foreground">Loading paste...</p>
          ) : (
            <div className="flex flex-col gap-2">
              {paste ? (
                <div className="flex justify-end">
                  <PasteItemControls
                    paste={paste}
                    mode="inline"
                    buttonSize="icon-sm"
                    iconClassName="size-5"
                  />
                </div>
              ) : null}

              <div className="min-w-0 w-full">
                <Textarea
                  readOnly
                  aria-readonly="true"
                  spellCheck={false}
                  value={paste?.content ?? ''}
                  rows={1}
                  className="field-sizing-content min-h-72 w-full min-w-0 max-h-[calc(100dvh-5rem)] resize-none outline-none cursor-text overflow-y-auto break-all font-mono text-sm leading-relaxed text-foreground shadow-none focus-visible:ring-0"
                />
              </div>
            </div>
          )}
        </section>
      </div>
    </main>
  )
}

