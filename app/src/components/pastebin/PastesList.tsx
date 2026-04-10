import { Link } from '@tanstack/react-router'
import { Item, ItemContent, ItemDescription, ItemGroup, ItemTitle } from '#/components/ui/item'
import { PasteItemControls } from './PasteItemControls'
import { PasteMeta } from './PasteMeta'

export type Paste = {
  id: string
  title: string
  content: string
  created_at: string
  updated_at: string
  is_public: boolean
  expires_at: string | null
}

type PastesListProps = {
  pastes: Paste[]
  isLoading?: boolean
}

function ListSkeleton() {
  return (
    <>
      {Array.from({ length: 5 }).map((_, i) => (
        <div
          key={i}
          className="h-20 animate-pulse rounded-xl border border-border/40 bg-muted/30"
        />
      ))}
    </>
  )
}

export function PastesList({ pastes, isLoading = false }: PastesListProps) {
  const visiblePastes = pastes.slice(0, 10)
  const showSkeleton = isLoading && visiblePastes.length === 0

  return (
    <section className="island-shell w-full max-w-sm rounded-2xl p-4 sm:p-5">
      <div className="mb-3 flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold text-foreground">Recent public pastes</h2>
      </div>

      <div
        className="max-h-[560px] min-h-[280px] overflow-y-auto overflow-x-hidden pr-1 [scrollbar-gutter:stable] lg:min-h-[560px]"
      >
        <ItemGroup className="gap-2">
          {showSkeleton ? (
            <ListSkeleton />
          ) : null}
          {visiblePastes.map((paste) => (
            <Item
              key={paste.id}
              variant="outline"
              size="xs"
              className="relative h-20 flex-col flex-nowrap items-stretch gap-1.5 overflow-hidden rounded-xl border-border/60 p-2.5 select-none"
            >
              <PasteItemControls paste={paste} />

              <ItemContent className="min-w-0 gap-1 pr-16">
                <ItemTitle className="w-full min-w-0 truncate line-clamp-1 text-ellipsis">
                  <Link
                    to="/pastebin/$id"
                    params={{ id: paste.id }}
                    className="inline-block max-w-full truncate hover:underline"
                  >
                    {paste.title || 'Untitled'}
                  </Link>
                </ItemTitle>
                <ItemDescription className="line-clamp-1 w-full min-w-0 overflow-hidden text-ellipsis font-mono text-xs leading-relaxed">
                  {paste.content}
                </ItemDescription>
              </ItemContent>

              <PasteMeta
                content={paste.content}
                createdAt={paste.created_at}
                className="absolute right-2 bottom-1 left-2 flex items-center justify-between gap-2 text-[10px] text-muted-foreground"
              />
            </Item>
          ))}
        </ItemGroup>
      </div>
    </section>
  )
}

