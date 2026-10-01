import { useState, type FormEvent } from 'react'
import { NavLink } from 'react-router-dom'
import { AuthShell } from '../../containers/auth/AuthShell'
import { ROUTES } from '../../constants/routes'
import { ru } from '../../locales/ru'
import { Button } from '../../ui/Button/Button'
import { FormField } from '../../ui/FormField/FormField'
import { Input } from '../../ui/Input/Input'

export function ForgotPasswordPage() {
  const [sent, setSent] = useState(false)
  const onSubmit = (event: FormEvent) => { event.preventDefault(); setSent(true) }

  return (
    <AuthShell title={ru.auth.forgotTitle} subtitle={ru.auth.forgotSubtitle}>
      <form className="grid gap-4" onSubmit={onSubmit}>
        <FormField label={ru.auth.email}><Input type="email" required /></FormField>
        {sent ? <div className="rounded-xl bg-emerald-50 p-3 text-sm text-emerald-800">{ru.auth.resetSuccess}</div> : null}
        <Button type="submit">{ru.common.send}</Button>
        <NavLink className="text-center text-sm font-semibold text-brand" to={ROUTES.login}>Вернуться ко входу</NavLink>
      </form>
    </AuthShell>
  )
}
