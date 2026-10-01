import { useNavigate } from 'react-router-dom'
import { ROUTES } from '../../constants/routes'
import type { ScheduleMatch } from '../../types/tournament'
import { Badge } from '../../ui/Badge/Badge'
import { Card } from '../../ui/Card/Card'

export function ScheduleTable({ matches }: { matches: ScheduleMatch[] }) {
  const navigate = useNavigate()
  return (
    <Card className="mt-5 overflow-hidden p-4">
      <div className="mb-4 flex items-center justify-between gap-4"><h3 className="text-xl font-extrabold">Расписание · тур 3</h3></div>
      <div className="overflow-x-auto">
        <table className="w-full min-w-[900px] text-sm">
          <thead className="text-left text-slate-500"><tr><th className="py-3">Дата и время</th><th>Группа</th><th>Матч</th><th>Площадка</th><th>Статус</th><th>Счёт</th></tr></thead>
          <tbody>
            {matches.map((match) => (
              <tr key={match.id} className="cursor-pointer border-t border-slate-200 hover:bg-slate-50" onClick={() => match.id === 'semifinal-1' && navigate(ROUTES.match(match.id))}>
                <td className="py-3">{match.date}</td><td>{match.group}</td><td className="font-medium">{match.home} <span className="text-slate-400">vs</span> {match.away}</td><td>{match.venue}</td>
                <td>{match.status === 'awaiting-confirmation' ? <Badge tone="warning">Ожидает подтверждения</Badge> : <Badge>Запланирован</Badge>}</td>
                <td className="font-bold">{match.score ?? '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  )
}
