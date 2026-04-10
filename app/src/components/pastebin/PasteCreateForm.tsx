import type { FormEvent } from 'react'
import { useMemo, useState } from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Checkbox } from '#/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { Textarea } from '#/components/ui/textarea'

const EXPIRY_PRESETS = [
  { value: '10m', label: '10 minutes' },
  { value: '1h', label: '1 hour' },
  { value: '1d', label: '1 day' },
  { value: '1w', label: '1 week' },
  { value: '1mo', label: '1 month' },
  { value: '1y', label: '1 year' },
  { value: 'never', label: 'Never' },
] as const

type ExpiryPreset = (typeof EXPIRY_PRESETS)[number]['value']

function expiresAtISOFromPreset(preset: ExpiryPreset): string | null {
  if (preset === 'never') return null
  const d = new Date()
  switch (preset) {
    case '10m':
      d.setMinutes(d.getMinutes() + 10)
      break
    case '1h':
      d.setHours(d.getHours() + 1)
      break
    case '1d':
      d.setDate(d.getDate() + 1)
      break
    case '1w':
      d.setDate(d.getDate() + 7)
      break
    case '1mo':
      d.setMonth(d.getMonth() + 1)
      break
    case '1y':
      d.setFullYear(d.getFullYear() + 1)
      break
    default:
      return null
  }
  return d.toISOString()
}

type PasteCreateFormProps = {
  onCreate: (params: {
    title: string
    content: string
    is_public: boolean
    expires_at: string | null
  }) => Promise<boolean>
}

export function PasteCreateForm({ onCreate }: PasteCreateFormProps) {
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [isPublic, setIsPublic] = useState(true)
  const [expiryPreset, setExpiryPreset] = useState<ExpiryPreset>('never')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const canSubmit = useMemo(
    () => content.trim().length > 0 && !isSubmitting,
    [content, isSubmitting],
  )

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setIsSubmitting(true)
    try {
      const expiresAt = expiresAtISOFromPreset(expiryPreset)
      const success = await onCreate({
        title: title.trim(),
        content,
        is_public: isPublic,
        expires_at: expiresAt,
      })
      if (success) {
        setTitle('')
        setContent('')
        setIsPublic(true)
        setExpiryPreset('never')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={onSubmit}>
      <div className="flex flex-col gap-2">
        <Label htmlFor="paste-title">Paste title</Label>
        <Input
          id="paste-title"
          name="title"
          autoComplete="off"
          placeholder="Optional title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="outline-none focus-visible:ring-0"
        />
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="paste-body">Paste</Label>
        <Textarea
          id="paste-body"
          name="content"
          required
          rows={8}
          placeholder="Your text…"
          className="min-h-40 outline-none focus-visible:ring-0"
          value={content}
          onChange={(e) => setContent(e.target.value)}
        />
      </div>
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-2 text-sm text-foreground">
          <Checkbox
            id="paste-is-public"
            checked={isPublic}
            onCheckedChange={(checked) => setIsPublic(Boolean(checked))}
          />
          <Label htmlFor="paste-is-public" className="cursor-pointer">
            Is public
          </Label>
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="paste-expires">Paste expiration</Label>
        <Select
          value={expiryPreset}
          onValueChange={(v) => {
            if (v != null) setExpiryPreset(v as ExpiryPreset)
          }}
        >
          <SelectTrigger id="paste-expires" className="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent alignItemWithTrigger={false} side="bottom">
            {EXPIRY_PRESETS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <Button type="submit" disabled={!canSubmit}>
        {isSubmitting ? 'Saving…' : 'Create paste'}
      </Button>
    </form>
  )
}

