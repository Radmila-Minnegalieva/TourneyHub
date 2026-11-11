import { useState, type FormEvent } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { AuthShell } from "../../containers/auth/AuthShell";
import { useAuth } from "../../hooks/useAuth";
import { errorMessage } from "../../api/client";
import { Button } from "../../ui/Button/Button";
import { Input } from "../../ui/Input/Input";
import { FormField } from "../../ui/FormField/FormField";
import { Feedback } from "../../ui/Feedback";
export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setBusy(true);
    setError("");
    try {
      await login(String(form.get("email")), String(form.get("password")));
      const from = location.state?.from;
      navigate(
        typeof from === "string" &&
          from.startsWith("/") &&
          !from.startsWith("//")
          ? from
          : "/tournaments",
        { replace: true },
      );
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <AuthShell title="Вход" subtitle="Войдите в свой аккаунт TourneyHub">
      <form onSubmit={submit} className="grid gap-4">
        <FormField label="Email">
          <Input name="email" type="email" autoComplete="email" required />
        </FormField>
        <FormField label="Пароль">
          <Input
            name="password"
            type="password"
            autoComplete="current-password"
            required
          />
        </FormField>
        <Feedback error={error} />
        <Button disabled={busy}>Войти</Button>
        <Link to="/auth/forgot-password">Забыли пароль?</Link>
        <Link to="/auth/register">Создать аккаунт</Link>
      </form>
    </AuthShell>
  );
}
