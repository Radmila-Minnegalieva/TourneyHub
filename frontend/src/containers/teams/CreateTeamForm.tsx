import { useState, type FormEvent } from 'react'
import { ru } from '../../locales/ru'
import type { Team } from '../../types/team'
import { Button } from '../../ui/Button/Button'
import { Card } from '../../ui/Card/Card'
import { FormField } from '../../ui/FormField/FormField'
import { Input } from '../../ui/Input/Input'
import { Select } from '../../ui/Select/Select'

interface Props { onCreate: (payload: Pick<Team, 'name' | 'shortName' | 'discipline'>) => void }

export function CreateTeamForm({ onCreate }: Props) {
  const [open, setOpen] = useState(false)
  const [message, setMessage] = useState('')
  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onCreate({ name: String(form.get('name') ?? ''), shortName: String(form.get('shortName') ?? ''), discipline: String(form.get('discipline') ?? '') })
    event.currentTarget.reset()
    setMessage(ru.teams.created)
    setOpen(false)
  }
  if (!open) return <div className="flex items-center gap-3"><Button onClick={() => setOpen(true)}>{ru.teams.createTitle}</Button>{message ? <span className="text-sm text-emerald-700">{message}</span> : null}</div>
  return (
    <Card className="p-5">
      <form className="grid gap-4 sm:grid-cols-3" onSubmit={onSubmit}>
        <FormField label={ru.teams.name}><Input name="name" required /></FormField>
        <FormField label={ru.teams.shortName}><Input name="shortName" maxLength={4} required /></FormField>
        <FormField label={ru.teams.discipline}><Select name="discipline" defaultValue="Мини-футбол"><option>Мини-футбол</option><option>Волейбол</option><option>Шахматы</option><option>Баскетбол 3×3</option><option>Настольный теннис</option><option>Интеллектуальные игры</option><option>Киберспорт</option></Select></FormField>
        <div className="flex gap-3 sm:col-span-3 sm:justify-end"><Button type="button" variant="secondary" onClick={() => setOpen(false)}>{ru.common.cancel}</Button><Button type="submit">{ru.common.create}</Button></div>
      </form>
    </Card>
  )
}
