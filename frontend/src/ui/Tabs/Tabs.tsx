import { NavLink } from 'react-router-dom'
import { cn } from '../../lib/cn'

export interface TabItem {
  label: string
  to: string
}

export function Tabs({ items }: { items: TabItem[] }) {
  return (
    <nav className="flex gap-1 overflow-x-auto border-b border-slate-200" aria-label="Разделы турнира">
      {items.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          className={({ isActive }) =>
            cn(
              'whitespace-nowrap border-b-2 px-4 py-3 text-sm font-semibold transition',
              isActive ? 'border-brand text-brand' : 'border-transparent text-slate-500 hover:text-ink',
            )
          }
        >
          {item.label}
        </NavLink>
      ))}
    </nav>
  )
}
