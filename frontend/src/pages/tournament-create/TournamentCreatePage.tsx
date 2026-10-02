import { useState } from 'react'
import { Card } from '../../ui/Card/Card'
import { Button } from '../../ui/Button/Button'
import { Input } from '../../ui/Input/Input'
import { Select } from '../../ui/Select/Select'
import { cn } from '../../lib/cn'

const formats = [
  { id: 'single-elimination', title: 'Олимпийская система', description: 'Проигравший выбывает. Число участников дополняется до степени двойки.' },
  { id: 'round-robin', title: 'Круговая система', description: 'Каждый играет с каждым. Места — по очкам и дополнительным показателям.' },
  { id: 'groups-playoff', title: 'Группы + плей-офф', description: 'Круговой турнир в группах, лучшие выходят в сетку на выбывание.' },
] as const

export function TournamentCreatePage() {
  const [format, setFormat] = useState<(typeof formats)[number]['id']>('groups-playoff')

  return (
    <section>
      <h1 className="text-3xl font-extrabold">Новый турнир</h1>
      <p className="mt-1 text-slate-500">Шаг 2 из 4 — формат и регламент</p>
      <div className="mt-6 grid gap-6 lg:grid-cols-[280px_1fr]">
        <Card className="h-fit p-4">
          {['Основные сведения', 'Формат и регламент', 'Регистрация', 'Проверка и публикация'].map((step, index) => (
            <div key={step} className={cn('mb-2 flex items-center gap-3 rounded-xl px-3 py-3 text-sm', index === 1 && 'bg-brand-soft font-semibold text-brand')}>
              <span className={cn('flex h-7 w-7 items-center justify-center rounded-full border text-xs font-bold', index === 0 && 'border-success bg-success text-white', index === 1 && 'border-brand bg-brand text-white')}>
                {index === 0 ? '✓' : index + 1}
              </span>
              {step}
            </div>
          ))}
        </Card>
        <Card className="p-5 sm:p-6">
          <h2 className="text-xl font-extrabold">Формат турнира</h2>
          <div className="mt-4 grid gap-4 xl:grid-cols-3">
            {formats.map((item) => (
              <button
                key={item.id}
                onClick={() => setFormat(item.id)}
                className={cn('rounded-2xl border p-5 text-left transition', format === item.id ? 'border-brand bg-brand-soft' : 'border-slate-200 bg-white hover:border-slate-300')}
              >
                <div className="mb-6 text-4xl text-slate-400">▦</div>
                <div className="font-extrabold">{item.title}</div>
                <p className="mt-2 text-sm leading-5 text-slate-500">{item.description}</p>
              </button>
            ))}
          </div>

          <h2 className="mt-7 text-xl font-extrabold">Параметры группового этапа</h2>
          <div className="mt-4 grid gap-4 md:grid-cols-3">
            <label className="grid gap-1.5 text-sm font-semibold">Число групп<Input type="number" defaultValue={3} /></label>
            <label className="grid gap-1.5 text-sm font-semibold">Выходят из группы<Input type="number" defaultValue={2} /></label>
            <label className="grid gap-1.5 text-sm font-semibold">Лучшие команды с 3-х мест<Input type="number" defaultValue={2} /></label>
          </div>

          <h2 className="mt-7 text-xl font-extrabold">Регламент матча</h2>
          <div className="mt-4 grid gap-4 md:grid-cols-3">
            <label className="grid gap-1.5 text-sm font-semibold">Формат серии<Select defaultValue="bo1"><option value="bo1">Bo1 — один матч</option><option value="bo3">Bo3</option><option value="bo5">Bo5</option></Select></label>
            <label className="grid gap-1.5 text-sm font-semibold">Очки: победа / ничья / поражение<Input defaultValue="3 / 1 / 0" /></label>
            <label className="grid gap-1.5 text-sm font-semibold">Дополнительные показатели<Select><option>Личные встречи → разница мячей</option><option>Разница мячей → забитые мячи</option></Select></label>
            <label className="grid gap-1.5 text-sm font-semibold">Посев<Select><option>По рейтингу Эло</option><option>Случайный</option></Select></label>
            <label className="grid gap-1.5 text-sm font-semibold">Подтверждение результата<Select><option>Капитан соперника, 24 ч</option><option>Капитан соперника, 12 ч</option></Select></label>
            <label className="grid gap-1.5 text-sm font-semibold">Матч за 3-е место<Select><option>Да</option><option>Нет</option></Select></label>
          </div>

          <div className="mt-5 rounded-xl bg-brand-soft px-4 py-3 text-sm text-indigo-800">Будет создано: групповой этап — 18 матчей, плей-офф на 8 команд — 8 матчей с учётом матча за 3-е место.</div>
          <div className="mt-5 flex flex-col-reverse justify-between gap-3 sm:flex-row">
            <Button variant="secondary">← Назад</Button>
            <div className="flex flex-col gap-3 sm:flex-row"><Button variant="secondary">Сохранить черновик</Button><Button>Далее →</Button></div>
          </div>
        </Card>
      </div>
    </section>
  )
}
