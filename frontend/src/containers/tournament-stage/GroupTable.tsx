import type { TournamentGroup } from '../../types/tournament'
import { Card } from '../../ui/Card/Card'

export function GroupTable({ group }: { group: TournamentGroup }) {
  return (
    <Card className="overflow-hidden p-4">
      <h3 className="mb-4 text-xl font-extrabold">{group.title}</h3>
      <div className="overflow-x-auto">
        <table className="w-full min-w-[420px] text-sm">
          <thead className="text-left text-slate-500"><tr><th className="py-2">#</th><th>Команда</th><th>И</th><th>В</th><th>Н</th><th>П</th><th>М</th><th>О</th></tr></thead>
          <tbody>
            {group.rows.map((row) => (
              <tr key={row.teamId} className="border-t border-slate-200">
                <td className="py-3">{row.position}</td>
                <td className="font-medium"><span className="mr-2 inline-flex h-6 w-6 items-center justify-center rounded-md bg-brand text-[10px] font-bold text-white">{row.shortName}</span>{row.team}</td>
                <td>{row.played}</td><td>{row.wins}</td><td>{row.draws}</td><td>{row.losses}</td><td>{row.goals}</td><td className="font-bold">{row.points}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  )
}
