import { useNavigate } from 'react-router-dom'
import { ROUTES } from '../../constants/routes'
import type { Team } from '../../types/team'
import { Badge } from '../../ui/Badge/Badge'
import { Button } from '../../ui/Button/Button'
import { Card } from '../../ui/Card/Card'

export function TeamCard({ team }: { team: Team }) {
  const navigate = useNavigate()
  const captain = team.members.find((member) => member.role === 'captain')
  return (
    <Card className="flex flex-col p-5">
      <div className="flex items-start justify-between gap-4"><div className="flex h-12 w-12 items-center justify-center rounded-xl bg-brand font-extrabold text-white">{team.shortName}</div><Badge>{team.discipline}</Badge></div>
      <h2 className="mt-4 text-xl font-extrabold">{team.name}</h2>
      <dl className="mt-4 grid grid-cols-2 gap-4 text-sm"><div><dt className="text-slate-500">Капитан</dt><dd className="font-semibold">{captain?.name ?? '—'}</dd></div><div><dt className="text-slate-500">Рейтинг Эло</dt><dd className="font-semibold">{team.rating}</dd></div><div><dt className="text-slate-500">Игроков</dt><dd className="font-semibold">{team.members.length}</dd></div><div><dt className="text-slate-500">Приглашения</dt><dd className="font-semibold">{team.pendingInvites.length}</dd></div></dl>
      <Button variant="secondary" className="mt-5" onClick={() => navigate(ROUTES.team(team.id))}>Открыть команду</Button>
    </Card>
  )
}
