import { useNavigate } from 'react-router-dom'
import { ROUTES } from '../../constants/routes'
import { useTournaments } from '../../hooks/useTournaments'
import { TournamentCard } from '../../containers/tournament-list/TournamentCard'
import { TournamentFilters } from '../../containers/tournament-list/TournamentFilters'
import { Button } from '../../ui/Button/Button'

export function TournamentsPage() {
  const navigate = useNavigate()
  const { tournaments, filters, setFilters, reset } = useTournaments()

  return (
    <section>
      <div className="mb-6 flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
        <div>
          <h1 className="text-3xl font-extrabold">Турниры</h1>
          <p className="mt-1 text-slate-500">Найдите соревнование и подайте заявку за свою команду</p>
        </div>
        <Button onClick={() => navigate(ROUTES.createTournament)}>+ Создать турнир</Button>
      </div>
      <TournamentFilters filters={filters} onChange={setFilters} onReset={reset} />
      <div className="mt-6 grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        {tournaments.map((tournament) => <TournamentCard key={tournament.id} tournament={tournament} />)}
      </div>
      {tournaments.length === 0 && <p className="py-16 text-center text-slate-500">По выбранным фильтрам турниров нет.</p>}
    </section>
  )
}
