import { useState, type FormEvent } from "react";
import { useAuth } from "../../hooks/useAuth";
import { errorMessage } from "../../api/client";
import { Button } from "../../ui/Button/Button";
import { Card } from "../../ui/Card/Card";
import { Input } from "../../ui/Input/Input";
import { FormField } from "../../ui/FormField/FormField";
import { Feedback } from "../../ui/Feedback";
const labels = {
  participant: "Участник",
  organizer: "Организатор",
  judge: "Судья",
  admin: "Администратор",
};
export function ProfilePage() {
  const { user, updateProfile } = useAuth();
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  if (!user) return null;
  const current = user;
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    setBusy(true);
    setError("");
    setSaved(false);
    try {
      await updateProfile(
        {
          ...current,
          name: String(f.get("name")),
          nickname: String(f.get("nickname")),
          timezone: String(f.get("timezone")),
          notificationSettings: {
            ...current.notificationSettings,
            email: f.get("email") === "on",
          },
        },
        f.get("avatar") as File,
      );
      setSaved(true);
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="mx-auto max-w-3xl">
      <h1 className="text-3xl font-extrabold">Профиль</h1>
      <Card className="mt-6 p-6">
        <p>
          {user.email} · {user.roles.map((r) => labels[r]).join(", ")}
        </p>
        {user.avatarUrl && (
          <img
            alt="Аватар"
            src={user.avatarUrl}
            className="mt-4 h-20 w-20 rounded-full object-cover"
          />
        )}
        <form key={user.id} onSubmit={submit} className="mt-5 grid gap-4">
          <FormField label="Имя">
            <Input
              name="name"
              defaultValue={user.name}
              maxLength={120}
              required
            />
          </FormField>
          <FormField label="Никнейм">
            <Input
              name="nickname"
              defaultValue={user.nickname}
              pattern="[A-Za-z0-9_]{3,32}"
              required
            />
          </FormField>
          <FormField label="Часовой пояс IANA">
            <Input name="timezone" defaultValue={user.timezone} required />
          </FormField>
          <FormField label="Аватар PNG/JPEG, до 5 МиБ">
            <Input name="avatar" type="file" accept="image/png,image/jpeg" />
          </FormField>
          <label>
            <input
              name="email"
              type="checkbox"
              defaultChecked={user.notificationSettings.email}
            />{" "}
            Уведомления по email
          </label>
          <Feedback error={error} />
          {saved && <p role="status">Изменения сохранены</p>}
          <Button disabled={busy}>Сохранить</Button>
        </form>
      </Card>
    </section>
  );
}
