import type { PropsWithChildren } from 'react'
import { cn } from '../../lib/cn'

type Tone = 'neutral' | 'success' | 'warning' | 'brand'

const tones: Record<Tone, string> = {
  neutral: 'bg-slate-100 text-slate-700',
  success: 'bg-emerald-50 text-emerald-700',
  warning: 'bg-orange-50 text-orange-700',
  brand: 'bg-indigo-50 text-brand',
}

export function Badge({ children, tone = 'neutral' }: PropsWithChildren<{ tone?: Tone }>) {
  return <span className={cn('inline-flex rounded-full px-3 py-1 text-xs font-bold', tones[tone])}>{children}</span>
}
