import { useNavigate } from 'react-router-dom'
import { getFreePlaces, tournamentStatusLabel } from '../../lib/tournament'
import { ROUTES } from '../../constants/routes'
import type { Tournament } from '../../types/tournament'
import { Badge } from '../../ui/Badge/Badge'
import { Button } from '../../ui/Button/Button'
import { Card } from '../../ui/Card/Card'

const statusView = {
  draft: { tone: 'neutral', variant: 'secondary', action: 'Открыть' },
  cancelled: { tone: 'neutral', variant: 'secondary', action: 'Открыть' },
  registration: { tone: 'success', variant: 'primary', action: 'Подать заявку' },
  running: { tone: 'warning', variant: 'secondary', action: 'Открыть' },
  completed: { tone: 'neutral', variant: 'secondary', action: 'Открыть' },
} as const

export function TournamentCard({ tournament }: { tournament: Tournament }) {
  const navigate = useNavigate()
  const freePlaces = getFreePlaces(tournament.participants, tournament.capacity)
  const view = statusView[tournament.status]
  const hints = {
    draft: 'Ещё не опубликован',
    cancelled: 'Турнир отменён',
    registration: `Осталось мест: ${freePlaces}`,
    running: 'Смотреть сетку и результаты',
    completed: 'Итоги опубликованы',
  }
  const progress = Math.round((tournament.participants / tournament.capacity) * 100)

  return (
    <Card className="flex min-h-[285px] flex-col p-5">
      <div className="mb-4 flex items-center justify-between gap-3">
        <Badge tone={view.tone}>{tournamentStatusLabel[tournament.status]}</Badge>
        <Badge>{tournament.discipline}</Badge>
      </div>
      <h2 className="mb-4 text-xl font-extrabold leading-tight">{tournament.title}</h2>
      <dl className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
        <div><dt className="text-slate-500">Формат</dt><dd className="font-semibold">{tournament.formatLabel}</dd></div>
        <div><dt className="text-slate-500">Даты</dt><dd className="font-semibold">{tournament.dateLabel}</dd></div>
        <div><dt className="text-slate-500">Организатор</dt><dd className="font-semibold">{tournament.organizer}</dd></div>
        <div><dt className="text-slate-500">Участники</dt><dd className="font-semibold">{tournament.participants} / {tournament.capacity}</dd></div>
      </dl>
      <div className="mt-4 h-1.5 overflow-hidden rounded-full bg-slate-100">
        <div className="h-full bg-brand" style={{ width: `${progress}%` }} />
      </div>
      <div className="mt-auto flex items-center justify-between gap-3 pt-4 text-sm text-slate-500">
        <span>{hints[tournament.status]}</span>
        <Button
          variant={view.variant}
          onClick={() => navigate(ROUTES.tournament(tournament.id, 'groups'))}
        >
          {view.action}
        </Button>
      </div>
    </Card>
  )
}
