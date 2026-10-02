import type { PropsWithChildren } from 'react'
import { NavLink } from 'react-router-dom'
import { ROUTES } from '../../constants/routes'
import { Card } from '../../ui/Card/Card'

interface Props {
  title: string
  subtitle: string
}

export function AuthShell({ title, subtitle, children }: PropsWithChildren<Props>) {
  return (
    <div className="mx-auto grid min-h-[calc(100vh-10rem)] max-w-5xl items-center gap-8 py-8 lg:grid-cols-[0.9fr_1.1fr]">
      <div className="hidden lg:block">
        <NavLink to={ROUTES.tournaments} className="text-3xl font-extrabold tracking-tight text-slate-950">
          Tourney<span className="text-brand">Hub</span>
        </NavLink>
        <h1 className="mt-8 max-w-md text-4xl font-extrabold leading-tight">Турниры, команды, результаты и рейтинг — в одном месте.</h1>
        <p className="mt-4 max-w-md leading-7 text-slate-500">Публичные страницы доступны без входа. Авторизация нужна для управления командами, заявками и турнирами.</p>
      </div>
      <Card className="p-6 sm:p-8">
        <h1 className="text-2xl font-extrabold">{title}</h1>
        <p className="mt-2 text-sm leading-6 text-slate-500">{subtitle}</p>
        <div className="mt-6">{children}</div>
      </Card>
    </div>
  )
}
