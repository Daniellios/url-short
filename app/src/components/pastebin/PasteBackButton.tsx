import { Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'

type BackButtonProps = {
  to: '/' | '/pastebin'
  className?: string
}

export function BackButton({ to, className }: BackButtonProps) {
  return (
    <Link
      to={to}
      className={className ?? 'inline-flex items-center gap-1 text-sm text-muted-foreground hover:underline'}
    >
      <ArrowLeft className="size-4" />
      Back
    </Link>
  )
}

