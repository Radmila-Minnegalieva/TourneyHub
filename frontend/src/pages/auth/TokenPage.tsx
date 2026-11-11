import { useState, type FormEvent } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { AuthShell } from "../../containers/auth/AuthShell";
import { api, errorMessage, setTokens } from "../../api/client";
import { Button } from "../../ui/Button/Button";
import { Input } from "../../ui/Input/Input";
import { Feedback } from "../../ui/Feedback";
export function TokenPage({ reset = false }: { reset?: boolean }) {
  const [params] = useSearchParams();
  const token = params.get("token");
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    setError("");
    if (reset && f.get("password") !== f.get("repeat")) {
      setError("Пароли не совпадают");
      return;
    }
    setBusy(true);
    try {
      await api(
        reset ? "/auth/password-reset/confirm" : "/auth/verify-email",
        {
          method: "POST",
          body: reset
            ? { token, password: String(f.get("password")) }
            : { token },
        },
        false,
      );
      if (reset) setTokens(null);
      setDone(true);
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <AuthShell
      title={reset ? "Новый пароль" : "Подтверждение email"}
      subtitle="TourneyHub"
    >
      {done ? (
        <p role="status">
          {reset ? "Пароль изменён." : "Email подтверждён."} Теперь можно войти.
        </p>
      ) : !token ? (
        <Feedback error="В ссылке отсутствует токен. Откройте ссылку из письма." />
      ) : (
        <form onSubmit={submit} className="grid gap-4">
          {reset && (
            <>
              <Input
                aria-label="Новый пароль"
                name="password"
                type="password"
                minLength={12}
                required
              />
              <Input
                aria-label="Повторите пароль"
                name="repeat"
                type="password"
                minLength={12}
                required
              />
            </>
          )}
          <Feedback error={error} />
          <Button disabled={busy}>
            {reset ? "Сохранить пароль" : "Подтвердить email"}
          </Button>
        </form>
      )}
      <Link className="mt-4 block" to="/auth/login">
        Войти
      </Link>
    </AuthShell>
  );
}
