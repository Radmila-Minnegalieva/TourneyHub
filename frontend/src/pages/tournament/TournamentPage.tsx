import { useParams } from 'react-router-dom'
import { ROUTES } from '../../constants/routes'
import { groups, quarterFinals, schedule, semiFinals } from '../../mocks/tournamentDetails'
import { Badge } from '../../ui/Badge/Badge'
import { Button } from '../../ui/Button/Button'
import { Card } from '../../ui/Card/Card'
import { Tabs } from '../../ui/Tabs/Tabs'
import { GroupTable } from '../../containers/tournament-stage/GroupTable'
import { ScheduleTable } from '../../containers/tournament-stage/ScheduleTable'
import { BracketMatchCard } from '../../containers/playoff-bracket/BracketMatchCard'

export function TournamentPage() {
  const { tournamentId = 'autumn-cup', tab = 'groups' } = useParams()
  const tabs = [
    { label: 'Групповой этап', to: ROUTES.tournament(tournamentId, 'groups') },
    { label: 'Плей-офф', to: ROUTES.tournament(tournamentId, 'playoff') },
    { label: 'Расписание', to: ROUTES.tournament(tournamentId, 'schedule') },
    { label: 'Участники', to: ROUTES.tournament(tournamentId, 'participants') },
    { label: 'Итоги', to: ROUTES.tournament(tournamentId, 'results') },
  ]

  return (
    <section>
      <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
        <div>
          <div className="mb-3 flex flex-wrap gap-2"><Badge tone="warning">Идёт</Badge><Badge tone="brand">Группы + плей-офф</Badge><Badge>Мини-футбол · Bo1</Badge></div>
          <h1 className="text-3xl font-extrabold">Осенний кубок по мини-футболу 2026</h1>
          <p className="mt-2 text-slate-500">Организатор: Спортклуб МИЭМ · 12 команд · 1 октября – 30 ноября 2026</p>
        </div>
        <div className="flex gap-3"><Button variant="secondary">Регламент</Button><Button variant="secondary">Поделиться</Button></div>
      </div>
      <div className="mt-5"><Tabs items={tabs} /></div>

      {tab === 'playoff' ? (
        <Card className="mt-5 overflow-x-auto p-5">
          <div className="grid min-w-[980px] grid-cols-4 gap-8">
            <div><h3 className="mb-5 text-sm font-bold uppercase text-slate-500">1/4 финала</h3><div className="space-y-8">{quarterFinals.map((match) => <BracketMatchCard key={match.id} match={match} />)}</div></div>
            <div><h3 className="mb-5 text-sm font-bold uppercase text-slate-500">1/2 финала · 14 ноя</h3><div className="space-y-32 pt-16">{semiFinals.map((match) => <BracketMatchCard key={match.id} match={match} />)}</div></div>
            <div><h3 className="mb-5 text-sm font-bold uppercase text-slate-500">Финал · 30 ноя</h3><div className="pt-52"><BracketMatchCard match={{ id: 'final', home: 'Победитель 1/2 №1', homeShort: '', away: 'Победитель 1/2 №2', awayShort: '' }} /></div></div>
            <div><h3 className="mb-5 text-sm font-bold uppercase text-slate-500">Матч за 3-е место · 29 ноя</h3><div className="pt-52"><BracketMatchCard match={{ id: 'third', home: 'Проигравший 1/2 №1', homeShort: '', away: 'Проигравший 1/2 №2', awayShort: '' }} /></div></div>
          </div>
        </Card>
      ) : tab === 'participants' ? (
        <Card className="mt-5 p-6"><h2 className="text-xl font-extrabold">Участники</h2><p className="mt-2 text-slate-500">12 команд зарегистрированы в турнире. Детальная работа с заявками подключится к API.</p></Card>
      ) : tab === 'results' ? (
        <Card className="mt-5 p-6"><h2 className="text-xl font-extrabold">Итоги</h2><p className="mt-2 text-slate-500">Турнир ещё идёт. Итоговые места появятся после завершения всех матчей.</p></Card>
      ) : (
        <>
          <div className="mt-5 grid gap-5 xl:grid-cols-3">{groups.map((group) => <GroupTable key={group.id} group={group} />)}</div>
          <ScheduleTable matches={schedule} />
        </>
      )}
    </section>
  )
}
