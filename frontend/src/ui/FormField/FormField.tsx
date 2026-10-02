import type { PropsWithChildren, ReactNode } from 'react'

interface FormFieldProps {
  label: string
  hint?: ReactNode
  htmlFor?: string
}

export function FormField({ label, hint, htmlFor, children }: PropsWithChildren<FormFieldProps>) {
  return (
    <label className="grid gap-1.5 text-sm font-semibold" htmlFor={htmlFor}>
      <span>{label}</span>
      {children}
      {hint ? <span className="text-xs font-normal text-slate-500">{hint}</span> : null}
    </label>
  )
}
