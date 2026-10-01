import { CreateTeamForm } from '../../containers/teams/CreateTeamForm'
import { TeamCard } from '../../containers/teams/TeamCard'
import { useTeams } from '../../hooks/useTeams'
import { ru } from '../../locales/ru'

export function TeamsPage() {
  const { teams, createTeam } = useTeams()
  return (
    <section>
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end"><div><h1 className="text-3xl font-extrabold">{ru.teams.title}</h1><p className="mt-1 text-slate-500">{ru.teams.subtitle}</p></div><CreateTeamForm onCreate={createTeam} /></div>
      <div className="mt-6 grid gap-5 md:grid-cols-2 xl:grid-cols-3">{teams.map((team) => <TeamCard key={team.id} team={team} />)}</div>
    </section>
  )
}
