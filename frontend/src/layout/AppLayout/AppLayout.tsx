import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { ROUTES } from '../../constants/routes'
import { useAuth } from '../../hooks/useAuth'
import { cn } from '../../lib/cn'
import { ru } from '../../locales/ru'
import { Button } from '../../ui/Button/Button'

interface NavItem { to: string; label: string; end?: boolean }

export function AppLayout() {
  const { user, isAuthenticated, logout } = useAuth()
  const [mobileOpen, setMobileOpen] = useState(false)
  const isOrganizer = Boolean(user?.roles.includes('organizer'))

  const publicLinks: NavItem[] = [
    { to: ROUTES.tournaments, label: ru.nav.tournaments, end: true },
    { to: ROUTES.rating, label: ru.nav.rating },
  ]
  const privateLinks: NavItem[] = isAuthenticated ? [{ to: ROUTES.teams, label: ru.nav.teams }] : []
  const organizerLinks: NavItem[] = isOrganizer ? [{ to: ROUTES.createTournament, label: ru.nav.createTournament }] : []
  const links = [...publicLinks, ...privateLinks, ...organizerLinks]

  const linkClass = ({ isActive }: { isActive: boolean }) => cn(
    'rounded-xl px-3 py-2 text-sm font-semibold text-slate-300 transition hover:bg-slate-800 hover:text-white',
    isActive && 'bg-slate-800 text-white',
  )

  return (
    <div className="min-h-screen bg-page text-ink">
      <header className="sticky top-0 z-40 border-b border-slate-800 bg-slate-950 text-white shadow-sm">
        <div className="mx-auto flex min-h-16 max-w-[1440px] items-center gap-4 px-4 sm:px-6 lg:px-8">
          <NavLink to={ROUTES.tournaments} className="shrink-0 text-xl font-extrabold tracking-tight">Tourney<span className="text-indigo-300">Hub</span></NavLink>
          <nav className="hidden flex-1 items-center gap-1 md:flex" aria-label="Основная навигация">
            {links.map((link) => <NavLink key={link.to} to={link.to} end={link.end} className={linkClass}>{link.label}</NavLink>)}
          </nav>
          <div className="ml-auto hidden items-center gap-2 md:flex">
            {isAuthenticated && user ? (
              <>
                <NavLink to={ROUTES.profile} className={({ isActive }) => cn('flex items-center gap-2 rounded-xl px-2 py-1.5 transition hover:bg-slate-800', isActive && 'bg-slate-800')}>
                  <span className="flex h-9 w-9 items-center justify-center rounded-full bg-indigo-500 text-sm font-extrabold">{user.name.split(' ').map((part) => part[0]).join('').slice(0, 2)}</span>
                  <span className="hidden text-left lg:block"><span className="block text-sm font-semibold">{user.name}</span><span className="block text-xs text-slate-400">@{user.nickname}</span></span>
                </NavLink>
                <Button variant="secondary" className="border-slate-700 bg-slate-900 text-white hover:bg-slate-800" onClick={logout}>{ru.nav.signOut}</Button>
              </>
            ) : <NavLink to={ROUTES.login}><Button>{ru.nav.signIn}</Button></NavLink>}
          </div>
          <button type="button" className="ml-auto inline-flex h-10 w-10 items-center justify-center rounded-xl border border-slate-700 text-xl md:hidden" aria-label={ru.nav.menu} aria-expanded={mobileOpen} onClick={() => setMobileOpen((value) => !value)}>☰</button>
        </div>
        {mobileOpen ? (
          <div className="border-t border-slate-800 px-4 pb-4 md:hidden">
            <nav className="grid gap-1 pt-3" aria-label="Мобильная навигация">{links.map((link) => <NavLink key={link.to} to={link.to} end={link.end} onClick={() => setMobileOpen(false)} className={linkClass}>{link.label}</NavLink>)}</nav>
            <div className="mt-3 border-t border-slate-800 pt-3">{isAuthenticated && user ? <div className="grid gap-2"><NavLink to={ROUTES.profile} onClick={() => setMobileOpen(false)} className={linkClass}>{ru.nav.profile} · @{user.nickname}</NavLink><Button variant="secondary" className="w-full border-slate-700 bg-slate-900 text-white" onClick={() => { logout(); setMobileOpen(false) }}>{ru.nav.signOut}</Button></div> : <NavLink to={ROUTES.login} onClick={() => setMobileOpen(false)} className={linkClass}>{ru.nav.signIn}</NavLink>}</div>
          </div>
        ) : null}
      </header>
      <main className="mx-auto max-w-[1440px] px-4 py-8 sm:px-6 lg:px-8"><Outlet /></main>
    </div>
  )
}
