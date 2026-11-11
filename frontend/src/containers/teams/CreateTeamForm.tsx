import { useState, type FormEvent } from 'react'
import { errorMessage } from '../../api/client'
import { Button } from '../../ui/Button/Button'
import { Input } from '../../ui/Input/Input'
import { Feedback } from '../../ui/Feedback'
export function CreateTeamForm({ onCreate }: { onCreate: (name: string) => Promise<void> }) {
  const [open, setOpen] = useState(false); const [error, setError] = useState(''); const [busy, setBusy] = useState(false)
  async function submit(e: FormEvent<HTMLFormElement>) { e.preventDefault(); const f = new FormData(e.currentTarget); setBusy(true); setError(''); try { await onCreate(String(f.get('name'))); setOpen(false) } catch(e) { setError(errorMessage(e)) } finally { setBusy(false) } }
  return open ? <form className="grid gap-3" onSubmit={submit}><label>Название команды<Input name="name" required maxLength={120} /></label><Feedback error={error} /><div className="flex gap-3"><Button type="button" variant="secondary" onClick={() => setOpen(false)}>Отмена</Button><Button disabled={busy}>Создать команду</Button></div></form> : <Button onClick={() => setOpen(true)}>Создать команду</Button>
}
