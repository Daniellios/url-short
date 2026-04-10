type PasteMetaProps = {
  content: string
  createdAt: string
  expiresAt?: string | null
  className?: string
}

function formatPasteSize(content: string) {
  const bytes = new TextEncoder().encode(content).length
  if (bytes < 1024) return `${bytes} B`
  const kb = bytes / 1024
  if (kb < 1024) return `${kb.toFixed(1)} KB`
  const mb = kb / 1024
  return `${mb.toFixed(1)} MB`
}

function formatCreatedAt(value: string) {
  if (!value) return ''
  const match = value.match(/^(\d{4}-\d{2}-\d{2})[T ](\d{2}:\d{2})/)
  if (match) return `${match[1]} ${match[2]}`
  return value
}

export function PasteMeta({ content, createdAt, expiresAt, className }: PasteMetaProps) {
  return (
    <div className={className ?? 'flex items-center justify-between gap-2 text-xs text-muted-foreground'}>
      <span>Size | {formatPasteSize(content)}</span>
      <span className="flex flex-col items-end gap-0.5 sm:flex-row sm:items-center sm:gap-3">
        {expiresAt ? <span>Expires | {formatCreatedAt(expiresAt)}</span> : null}
        <span>Created | {formatCreatedAt(createdAt)}</span>
      </span>
    </div>
  )
}

