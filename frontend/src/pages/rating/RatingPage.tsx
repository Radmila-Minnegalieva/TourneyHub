import { rating } from '../../mocks/rating'
import { Card } from '../../ui/Card/Card'
import { Select } from '../../ui/Select/Select'

export function RatingPage() {
  return (
    <section>
      <div className="flex flex-col justify-between gap-4 md:flex-row md:items-start">
        <div><h1 className="text-3xl font-extrabold">Рейтинг команд</h1><p className="mt-1 text-slate-500">Рейтинг Эло пересчитывается после каждого подтверждённого матча. Начальное значение — 1500.</p></div>
        <div className="grid grid-cols-2 gap-3"><Select defaultValue="Мини-футбол"><option>Мини-футбол</option><option>Волейбол</option></Select><Select defaultValue="all"><option value="all">За всё время</option><option value="month">За 30 дней</option></Select></div>
      </div>
      <div className="mt-6 grid gap-4 md:grid-cols-3">
        {rating.slice(0, 3).map((entry) => <Card key={entry.team} className="flex items-center gap-4 p-5"><div className="text-3xl font-extrabold text-amber-500">{entry.position}</div><div className="flex h-12 w-12 items-center justify-center rounded-xl bg-brand font-bold text-white">{entry.shortName}</div><div><div className="font-extrabold">{entry.team}</div><div className="text-slate-500">Эло <strong className="text-ink">{entry.rating}</strong> · {entry.matches} матчей</div></div></Card>)}
      </div>
      <Card className="mt-5 overflow-hidden p-4">
        <div className="overflow-x-auto"><table className="w-full min-w-[900px] text-sm"><thead className="text-left text-slate-500"><tr><th className="py-3">#</th><th>Команда</th><th>Рейтинг Эло</th><th>Изм. за 30 дней</th><th>Матчи</th><th>Победы</th><th>% побед</th></tr></thead><tbody>{rating.map((entry) => <tr key={entry.team} className="border-t border-slate-200"><td className="py-3">{entry.position}</td><td className="font-medium"><span className="mr-2 inline-flex h-6 w-6 items-center justify-center rounded-md bg-brand text-[10px] font-bold text-white">{entry.shortName}</span>{entry.team}</td><td className="font-extrabold">{entry.rating}</td><td className={entry.change >= 0 ? 'font-bold text-emerald-700' : 'font-bold text-red-700'}>{entry.change >= 0 ? '▲ +' : '▼ '}{entry.change}</td><td>{entry.matches}</td><td>{entry.wins}</td><td>{entry.winRate}%</td></tr>)}</tbody></table></div>
      </Card>
    </section>
  )
}
