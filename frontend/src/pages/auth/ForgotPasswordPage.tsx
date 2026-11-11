import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { AuthShell } from "../../containers/auth/AuthShell";
import { api, errorMessage } from "../../api/client";
import { Button } from "../../ui/Button/Button";
import { Input } from "../../ui/Input/Input";
import { Feedback } from "../../ui/Feedback";
export function ForgotPasswordPage() {
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    setBusy(true);
    setError("");
    try {
      await api(
        "/auth/password-reset",
        { method: "POST", body: { email: String(f.get("email")) } },
        false,
      );
      setSent(true);
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <AuthShell title="Восстановление пароля" subtitle="Укажите email аккаунта">
      <form onSubmit={submit} className="grid gap-4">
        <Input aria-label="Email" name="email" type="email" required />
        <Feedback error={error} />
        {sent && (
          <p role="status">
            Если аккаунт существует, письмо со ссылкой отправлено.
          </p>
        )}
        <Button disabled={busy}>Отправить</Button>
        <Link to="/auth/login">Вернуться ко входу</Link>
      </form>
    </AuthShell>
  );
}
