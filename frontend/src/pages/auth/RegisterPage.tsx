import { useState, type FormEvent } from 'react'
import { NavLink, useNavigate } from 'react-router-dom'
import { AuthShell } from '../../containers/auth/AuthShell'
import { ROUTES } from '../../constants/routes'
import { useAuth } from '../../hooks/useAuth'
import { ru } from '../../locales/ru'
import { Button } from '../../ui/Button/Button'
import { FormField } from '../../ui/FormField/FormField'
import { Input } from '../../ui/Input/Input'

export function RegisterPage() {
  const navigate = useNavigate()
  const { register } = useAuth()
  const [message, setMessage] = useState('')

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const password = String(form.get('password') ?? '')
    const repeat = String(form.get('repeat') ?? '')
    if (password !== repeat) {
      setMessage('Пароли не совпадают.')
      return
    }
    register({
      email: String(form.get('email') ?? ''),
      nickname: String(form.get('nickname') ?? ''),
      name: String(form.get('name') ?? ''),
    })
    setMessage(ru.auth.registrationSuccess)
    setTimeout(() => navigate(ROUTES.profile), 500)
  }

  return (
    <AuthShell title={ru.auth.registerTitle} subtitle={ru.auth.registerSubtitle}>
      <form className="grid gap-4" onSubmit={onSubmit}>
        <FormField label={ru.auth.email}><Input name="email" type="email" required /></FormField>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label={ru.auth.nickname}><Input name="nickname" required /></FormField>
          <FormField label={ru.auth.name}><Input name="name" required /></FormField>
        </div>
        <FormField label={ru.auth.password}><Input name="password" type="password" minLength={8} required /></FormField>
        <FormField label={ru.auth.passwordRepeat}><Input name="repeat" type="password" minLength={8} required /></FormField>
        <label className="flex items-start gap-3 text-sm leading-6 text-slate-600"><input className="mt-1" type="checkbox" required /><span>{ru.auth.consent}</span></label>
        {message ? <div className="rounded-xl bg-brand-soft p-3 text-sm text-brand">{message}</div> : null}
        <Button type="submit" className="w-full">{ru.auth.register}</Button>
        <p className="text-center text-sm text-slate-500">{ru.auth.hasAccount} <NavLink className="font-semibold text-brand" to={ROUTES.login}>{ru.auth.login}</NavLink></p>
      </form>
    </AuthShell>
  )
}
