import { useMemo, useState, type FormEvent } from 'react'
import { NavLink, useParams } from 'react-router-dom'
import { ROUTES } from '../../constants/routes'
import { useAuth } from '../../hooks/useAuth'
import { useTeam } from '../../hooks/useTeams'
import { ru } from '../../locales/ru'
import { Badge } from '../../ui/Badge/Badge'
import { Button } from '../../ui/Button/Button'
import { Card } from '../../ui/Card/Card'
import { FormField } from '../../ui/FormField/FormField'
import { Input } from '../../ui/Input/Input'

export function TeamPage() {
  const { teamId } = useParams()
  const { user } = useAuth()
  const { team, inviteByNickname, removeMember, transferCaptain } = useTeam(teamId)
  const [copied, setCopied] = useState(false)
  const [message, setMessage] = useState('')
  const captain = team.members.find((member) => member.role === 'captain')
  const isCaptain = captain?.id === user?.id
  const inviteLink = useMemo(() => `${window.location.origin}/invite/${team.id}/demo-token`, [team.id])

  const onInvite = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    inviteByNickname(String(form.get('nickname') ?? ''))
    event.currentTarget.reset()
    setMessage(ru.teams.invited)
  }

  const copyInvite = async () => {
    try { await navigator.clipboard.writeText(inviteLink) } catch { /* clipboard may be unavailable in local preview */ }
    setCopied(true)
  }

  return (
    <section>
      <NavLink className="text-sm font-semibold text-brand" to={ROUTES.teams}>← {ru.common.back}</NavLink>
      <div className="mt-4 flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
        <div className="flex items-start gap-4"><div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-brand text-lg font-extrabold text-white">{team.shortName}</div><div><div className="flex flex-wrap items-center gap-2"><h1 className="text-3xl font-extrabold">{team.name}</h1><Badge>{team.discipline}</Badge></div><p className="mt-1 text-slate-500">Капитан: {captain?.name} · Эло {team.rating}</p></div></div>
      </div>
      <div className="mt-6 grid gap-5 lg:grid-cols-[1.4fr_0.9fr]">
        <Card className="overflow-hidden">
          <div className="border-b border-slate-200 p-5"><h2 className="text-xl font-extrabold">{ru.teams.members}</h2></div>
          <div>{team.members.map((member) => <div key={member.id} className="flex flex-col gap-3 border-b border-slate-100 px-5 py-4 last:border-b-0 sm:flex-row sm:items-center"><div className="flex-1"><div className="font-semibold">{member.name}</div><div className="text-sm text-slate-500">@{member.nickname}</div></div><Badge tone={member.role === 'captain' ? 'brand' : 'neutral'}>{member.role === 'captain' ? ru.teams.captain : ru.teams.player}</Badge>{isCaptain && member.id !== user?.id ? <div className="flex flex-wrap gap-2"><Button variant="secondary" onClick={() => transferCaptain(member.id)}>{ru.teams.transferCaptain}</Button><Button variant="danger" onClick={() => removeMember(member.id)}>{ru.teams.removeMember}</Button></div> : null}</div>)}</div>
        </Card>
        <div className="grid content-start gap-5">
          {isCaptain ? <Card className="p-5"><h2 className="text-xl font-extrabold">{ru.teams.invite}</h2><form className="mt-4 grid gap-3" onSubmit={onInvite}><FormField label={ru.teams.inviteNickname}><Input name="nickname" placeholder="nickname" required /></FormField><Button type="submit">{ru.common.send}</Button></form>{message ? <p className="mt-3 text-sm text-emerald-700">{message}</p> : null}<div className="my-5 border-t border-slate-200" /><FormField label={ru.teams.oneTimeLink}><Input readOnly value={inviteLink} /></FormField><Button className="mt-3 w-full" variant="secondary" onClick={copyInvite}>{copied ? 'Скопировано' : ru.common.copy}</Button></Card> : null}
          <Card className="p-5"><h2 className="text-lg font-extrabold">{ru.teams.pending}</h2>{team.pendingInvites.length ? <div className="mt-3 space-y-2">{team.pendingInvites.map((nickname) => <div key={nickname} className="rounded-xl bg-slate-50 px-3 py-2 text-sm">@{nickname}</div>)}</div> : <p className="mt-2 text-sm text-slate-500">Нет активных приглашений.</p>}</Card>
        </div>
      </div>
    </section>
  )
}
