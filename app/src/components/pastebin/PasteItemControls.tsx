import { Button } from "../ui/button"
import { Check, Copy, Download } from 'lucide-react'
import { ItemActions } from '#/components/ui/item'
import type React from "react"
import type { Paste } from "./PastesList"
import { useState } from "react"
import { toast } from "sonner"

interface PasteItemControlsProps {
    paste: Paste
    buttonSize?: 'icon-xs' | 'icon-sm'
    iconClassName?: string
    /** Floating on the card (list). Inline row for toolbars above content (detail page). */
    mode?: 'floating' | 'inline'
}

function buildFileName(paste: Paste) {
    const base = (paste.title || 'paste')
        .trim()
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '')
    return `${base || 'paste'}-${paste.id}.txt`
}

function onDownload(paste: Paste) {
    const blob = new Blob([paste.content], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = buildFileName(paste)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
}

export const PasteItemControls: React.FC<PasteItemControlsProps> = ({
    paste,
    buttonSize = 'icon-xs',
    iconClassName,
    mode = 'floating',
}) => {
    const [copiedPasteId, setCopiedPasteId] = useState<string | null>(null)

    async function onCopy(paste: Paste) {
        try {
            await navigator.clipboard.writeText(paste.content)
            setCopiedPasteId(paste.id)
            window.setTimeout(() => setCopiedPasteId((current) => (current === paste.id ? null : current)), 1200)
        } catch {
            toast.error('Could not copy to clipboard.')
        }
    }

    const actionsClassName =
        mode === 'floating'
            ? 'absolute top-2 right-2 z-10'
            : 'relative flex shrink-0 items-center justify-end gap-0.5'

    return (
        <ItemActions className={actionsClassName}>
            <Button
                type="button"
                variant="ghost"
                title="Download paste as text file"
                size={buttonSize}
                aria-label="Download paste as text file"
                onClick={() => onDownload(paste)}
            >
                <Download className={iconClassName} />
            </Button>
            <Button
                type="button"
                variant="ghost"
                title='Copy paste'
                size={buttonSize}
                className={
                    copiedPasteId === paste.id
                        ? 'bg-emerald-600 text-white hover:bg-emerald-600/90 dark:bg-emerald-500 dark:hover:bg-emerald-500/90'
                        : undefined
                }
                aria-label={copiedPasteId === paste.id ? 'Copied' : 'Copy paste'}
                onClick={() => onCopy(paste)}
            >
                {copiedPasteId === paste.id ? <Check className={iconClassName} /> : <Copy className={iconClassName} />}
            </Button>
        </ItemActions>
    )
}