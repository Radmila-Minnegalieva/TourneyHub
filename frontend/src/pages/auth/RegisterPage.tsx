import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { AuthShell } from "../../containers/auth/AuthShell";
import { useAuth } from "../../hooks/useAuth";
import { errorMessage } from "../../api/client";
import { Button } from "../../ui/Button/Button";
import { Input } from "../../ui/Input/Input";
import { FormField } from "../../ui/FormField/FormField";
import { Feedback } from "../../ui/Feedback";
export function RegisterPage() {
  const { register } = useAuth();
  const [error, setError] = useState("");
  const [sent, setSent] = useState(false);
  const [busy, setBusy] = useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const f = new FormData(event.currentTarget);
    setError("");
    if (f.get("password") !== f.get("repeat")) {
      setError("Пароли не совпадают");
      return;
    }
    setBusy(true);
    try {
      await register({
        email: String(f.get("email")),
        nickname: String(f.get("nickname")),
        name: String(f.get("name")),
        password: String(f.get("password")),
        personal_data_consent: true,
      });
      setSent(true);
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <AuthShell title="Регистрация" subtitle="Создайте аккаунт участника">
      {sent ? (
        <p role="status">
          Проверьте почту и подтвердите email по ссылке из письма. После
          подтверждения можно войти.
        </p>
      ) : (
        <form onSubmit={submit} className="grid gap-4">
          <FormField label="Email">
            <Input name="email" type="email" required />
          </FormField>
          <FormField label="Никнейм" hint="3–32 латинские буквы, цифры или _">
            <Input name="nickname" pattern="[A-Za-z0-9_]{3,32}" required />
          </FormField>
          <FormField label="Имя">
            <Input name="name" maxLength={120} required />
          </FormField>
          <FormField label="Пароль" hint="Не менее 12 символов">
            <Input name="password" type="password" minLength={12} required />
          </FormField>
          <FormField label="Повторите пароль">
            <Input name="repeat" type="password" minLength={12} required />
          </FormField>
          <label>
            <input type="checkbox" required /> Согласие на обработку
            персональных данных
          </label>
          <Feedback error={error} />
          <Button disabled={busy}>Зарегистрироваться</Button>
        </form>
      )}
      <Link className="mt-4 block" to="/auth/login">
        Войти
      </Link>
    </AuthShell>
  );
}
