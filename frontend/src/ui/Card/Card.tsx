import type { HTMLAttributes, PropsWithChildren } from 'react'
import { cn } from '../../lib/cn'

export function Card({ children, className, ...props }: PropsWithChildren<HTMLAttributes<HTMLDivElement>>) {
  return (
    <div className={cn('rounded-2xl border border-slate-200 bg-white shadow-card', className)} {...props}>
      {children}
    </div>
  )
}
