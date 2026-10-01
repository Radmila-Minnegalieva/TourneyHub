import { useState, type FormEvent } from 'react'
import { useAuth } from '../../hooks/useAuth'
import { ru } from '../../locales/ru'
import { Badge } from '../../ui/Badge/Badge'
import { Button } from '../../ui/Button/Button'
import { Card } from '../../ui/Card/Card'
import { FormField } from '../../ui/FormField/FormField'
import { Input } from '../../ui/Input/Input'
import { Select } from '../../ui/Select/Select'

const roleLabels = { participant: 'Участник', organizer: 'Организатор', judge: 'Судья', admin: 'Администратор' } as const

export function ProfilePage() {
  const { user, updateProfile } = useAuth()
  const [saved, setSaved] = useState(false)
  if (!user) return null

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    updateProfile({
      ...user,
      nickname: String(form.get('nickname') ?? user.nickname),
      name: String(form.get('name') ?? user.name),
      timezone: String(form.get('timezone') ?? user.timezone),
      notificationSettings: {
        email: form.get('emailNotifications') === 'on',
        inApp: form.get('inAppNotifications') === 'on',
        matchChanges: form.get('matchChanges') === 'on',
        applicationUpdates: form.get('applicationUpdates') === 'on',
        resultConfirmation: form.get('resultConfirmation') === 'on',
      },
    })
    setSaved(true)
  }

  return (
    <section className="mx-auto max-w-4xl">
      <div><h1 className="text-3xl font-extrabold">{ru.profile.title}</h1><p className="mt-1 text-slate-500">{ru.profile.subtitle}</p></div>
      <form className="mt-6 grid gap-5" onSubmit={onSubmit}>
        <Card className="p-6">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
            <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-2xl bg-brand text-2xl font-extrabold text-white">{user.name.split(' ').map((part) => part[0]).join('').slice(0, 2)}</div>
            <div className="flex-1"><h2 className="text-xl font-extrabold">{user.name}</h2><p className="text-sm text-slate-500">{user.email}</p><div className="mt-2 flex flex-wrap gap-2">{user.roles.map((role) => <Badge key={role} tone="brand">{roleLabels[role]}</Badge>)}</div></div>
          </div>
          <div className="mt-6 grid gap-4 sm:grid-cols-2">
            <FormField label={ru.profile.nickname}><Input name="nickname" defaultValue={user.nickname} /></FormField>
            <FormField label={ru.profile.name}><Input name="name" defaultValue={user.name} /></FormField>
            <FormField label={ru.profile.timezone}><Select name="timezone" defaultValue={user.timezone}><option value="Europe/Moscow">Москва (UTC+3)</option><option value="Europe/Budapest">Будапешт</option><option value="UTC">UTC</option></Select></FormField>
            <FormField label={ru.profile.avatar} hint="PNG или JPG. Фактическая загрузка подключится к API."><Input type="file" accept="image/png,image/jpeg" /></FormField>
          </div>
        </Card>
        <Card className="p-6">
          <h2 className="text-xl font-extrabold">{ru.profile.notifications}</h2>
          <div className="mt-4 grid gap-3">
            {[
              ['emailNotifications', ru.profile.emailNotifications, user.notificationSettings.email],
              ['inAppNotifications', ru.profile.inAppNotifications, user.notificationSettings.inApp],
              ['matchChanges', ru.profile.matchChanges, user.notificationSettings.matchChanges],
              ['applicationUpdates', ru.profile.applicationUpdates, user.notificationSettings.applicationUpdates],
              ['resultConfirmation', ru.profile.resultConfirmation, user.notificationSettings.resultConfirmation],
            ].map(([name, label, checked]) => (
              <label key={String(name)} className="flex items-center gap-3 rounded-xl border border-slate-200 px-4 py-3 text-sm"><input name={String(name)} type="checkbox" defaultChecked={Boolean(checked)} /><span>{String(label)}</span></label>
            ))}
          </div>
        </Card>
        <div className="flex items-center justify-end gap-3">{saved ? <span className="text-sm font-medium text-emerald-700">{ru.profile.saved}</span> : null}<Button type="submit">{ru.common.save}</Button></div>
      </form>
    </section>
  )
}
