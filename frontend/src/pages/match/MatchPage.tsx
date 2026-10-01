import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { matchDetails as match } from '../../mocks/match'
import { ROUTES } from '../../constants/routes'
import { Badge } from '../../ui/Badge/Badge'
import { Button } from '../../ui/Button/Button'
import { Card } from '../../ui/Card/Card'

export function MatchPage() {
  const navigate = useNavigate()
  const [message, setMessage] = useState<string | null>(null)

  return (
    <section>
      <button onClick={() => navigate(ROUTES.tournament(match.tournamentId, 'playoff'))} className="mb-4 text-sm text-slate-500 hover:text-ink">
        {match.tournamentTitle} / {match.stageLabel}
      </button>
      <Card className="grid gap-6 p-6 text-center md:grid-cols-[1fr_auto_1fr] md:items-center">
        <div><div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-violet-600 text-xl font-extrabold text-white">{match.homeShort}</div><h2 className="mt-3 text-xl font-extrabold">{match.home}</h2><p className="text-slate-500">Эло {match.homeRating} · посев {match.homeSeed}</p></div>
        <div><div className="text-5xl font-extrabold">{match.score}</div><div className="mt-2"><Badge tone="warning">{match.statusLabel}</Badge></div><p className="mt-2 text-slate-500">{match.date} · {match.venue}</p></div>
        <div><div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-cyan-600 text-xl font-extrabold text-white">{match.awayShort}</div><h2 className="mt-3 text-xl font-extrabold">{match.away}</h2><p className="text-slate-500">Эло {match.awayRating} · посев {match.awaySeed}</p></div>
      </Card>

      <div className="mt-5 grid gap-5 xl:grid-cols-[1.4fr_1fr]">
        <Card className="p-5">
          <h2 className="text-xl font-extrabold">Результат матча</h2>
          <div className="mt-4 rounded-xl bg-brand-soft p-4 text-sm leading-6 text-indigo-800">Результат сообщил капитан команды «Орион» 14.11 в 21:07. Вы — капитан «Северных волков»: подтвердите счёт или оспорьте его до 15.11, 21:07. Если ответа не будет, результат подтвердится автоматически.</div>
          <div className="mt-5 overflow-x-auto"><table className="w-full min-w-[520px] text-sm"><thead className="text-left text-slate-500"><tr><th className="py-3">Период</th><th>{match.home}</th><th>{match.away}</th></tr></thead><tbody>{match.periods.map((period) => <tr key={period.label} className="border-t border-slate-200"><td className="py-3 font-medium">{period.label}</td><td>{period.home}</td><td>{period.away}</td></tr>)}<tr className="border-t border-slate-200 font-extrabold"><td className="py-3">Итог</td><td>3</td><td>1</td></tr></tbody></table></div>
          <div className="mt-4"><div className="mb-2 text-sm font-semibold">Протокол / доказательства</div><div className="rounded-xl border border-dashed border-slate-300 p-4 text-center text-slate-500">📎 {match.attachment}</div></div>
          {message && <div className="mt-4 rounded-xl bg-slate-100 p-3 text-sm">{message}</div>}
          <div className="mt-5 flex flex-col justify-end gap-3 sm:flex-row"><Button variant="danger" onClick={() => setMessage('Открыта форма оспаривания результата (демо без backend).')}>Оспорить результат</Button><Button variant="success" onClick={() => setMessage('Счёт подтверждён локально. После подключения API действие будет отправляться на сервер.')}>✓ Подтвердить счёт 3 : 1</Button></div>
        </Card>
        <div className="grid gap-5">
          <Card className="p-5"><h2 className="text-xl font-extrabold">История</h2><div className="mt-4 space-y-5">{match.history.map((item) => <div key={item.title} className="border-l-2 border-slate-200 pl-4"><div className="font-medium">{item.title}</div><div className="text-sm text-slate-500">{item.meta}</div></div>)}</div></Card>
          <Card className="p-5"><h2 className="text-xl font-extrabold">После подтверждения</h2><p className="mt-3 leading-6 text-slate-500">Победитель перейдёт в <strong className="text-ink">Финал</strong>, проигравший — в матч за 3-е место. Рейтинг Эло: «Орион» +{match.ratingDelta.home}, «Северные волки» {match.ratingDelta.away}.</p></Card>
        </div>
      </div>
    </section>
  )
}
