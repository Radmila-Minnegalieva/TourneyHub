import { useNavigate } from 'react-router-dom'
import { ROUTES } from '../../constants/routes'
import type { BracketMatch } from '../../types/tournament'
import { Badge } from '../../ui/Badge/Badge'

export function BracketMatchCard({ match }: { match: BracketMatch }) {
  const navigate = useNavigate()
  const clickable = match.id === 'semifinal-1'
  return (
    <button
      type="button"
      onClick={() => clickable && navigate(ROUTES.match(match.id))}
      className="relative w-full rounded-xl border border-slate-200 bg-white text-left transition hover:border-slate-300"
    >
      {match.state && <span className="absolute -top-3 right-3"><Badge tone="warning">ожидает</Badge></span>}
      <div className="flex items-center justify-between border-b border-slate-200 px-3 py-2.5"><span className="font-semibold">{match.home}</span><strong>{match.homeScore ?? ''}</strong></div>
      <div className="flex items-center justify-between px-3 py-2.5"><span>{match.away}</span><strong>{match.awayScore ?? ''}</strong></div>
    </button>
  )
}
