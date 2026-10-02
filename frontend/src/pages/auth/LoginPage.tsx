import { useState, type FormEvent } from 'react'
import { NavLink, useNavigate } from 'react-router-dom'
import { AuthShell } from '../../containers/auth/AuthShell'
import { ROUTES } from '../../constants/routes'
import { useAuth } from '../../hooks/useAuth'
import { ru } from '../../locales/ru'
import { Button } from '../../ui/Button/Button'
import { FormField } from '../../ui/FormField/FormField'
import { Input } from '../../ui/Input/Input'

export function LoginPage() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [email, setEmail] = useState('ivan.petrov@example.com')
  const [password, setPassword] = useState('password')

  const onSubmit = (event: FormEvent) => {
    event.preventDefault()
    login(email, password)
    navigate(ROUTES.tournaments)
  }

  return (
    <AuthShell title={ru.auth.loginTitle} subtitle={ru.auth.loginSubtitle}>
      <form className="grid gap-4" onSubmit={onSubmit}>
        <FormField label={ru.auth.email} htmlFor="login-email"><Input id="login-email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required /></FormField>
        <FormField label={ru.auth.password} htmlFor="login-password"><Input id="login-password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required /></FormField>
        <div className="flex justify-end"><NavLink className="text-sm font-semibold text-brand" to={ROUTES.forgotPassword}>{ru.auth.forgot}</NavLink></div>
        <Button type="submit" className="w-full">{ru.auth.login}</Button>
        <p className="text-center text-sm text-slate-500">{ru.auth.noAccount} <NavLink className="font-semibold text-brand" to={ROUTES.register}>{ru.auth.register}</NavLink></p>
        <p className="rounded-xl bg-slate-50 p-3 text-xs leading-5 text-slate-500">{ru.auth.mockNotice}</p>
      </form>
    </AuthShell>
  )
}
