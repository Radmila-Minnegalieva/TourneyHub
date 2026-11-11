import { useState, type FormEvent } from "react";
import { Link, useParams } from "react-router-dom";
import { api, errorMessage, type Schema } from "../../api/client";
import { useAuth } from "../../hooks/useAuth";
import { useTeam } from "../../hooks/useTeams";
import { Button } from "../../ui/Button/Button";
import { Card } from "../../ui/Card/Card";
import { Input } from "../../ui/Input/Input";
import { Feedback } from "../../ui/Feedback";
export function TeamPage() {
  const { teamId = "" } = useParams();
  const { user } = useAuth();
  const r = useTeam(teamId);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [link, setLink] = useState("");
  const [busy, setBusy] = useState(false);
  async function action(work: () => Promise<unknown>, message: string) {
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await work();
      setMessage(message);
      r.reload();
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  function invite(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    void action(
      () =>
        api(`/teams/${teamId}/invitations`, {
          method: "POST",
          body: { nickname: String(f.get("nickname")) },
        }),
      "Приглашение отправлено",
    );
  }
  const t = r.data;
  return (
    <section>
      <Link to="/teams">← Мои команды</Link>
      <Feedback loading={r.loading} error={r.error} retry={r.reload} />
      <Feedback error={error} />
      {message && <p role="status">{message}</p>}
      {t && (
        <>
          <h1 className="mt-4 text-3xl font-extrabold">{t.name}</h1>
          <p>{t.description}</p>
          <Card className="mt-5 p-5">
            <h2 className="text-xl font-bold">Участники</h2>
            {t.members.map((m) => (
              <div
                className="flex flex-wrap items-center gap-3 border-b py-4"
                key={m.user.id}
              >
                <span className="flex-1">
                  {m.user.name || m.user.nickname} · @{m.user.nickname}{" "}
                  {m.captain && "· Капитан"}
                </span>
                {t.captain_id === user?.id && !m.captain && (
                  <>
                    <Button
                      disabled={busy}
                      variant="secondary"
                      onClick={() =>
                        void action(
                          () =>
                            api(`/teams/${teamId}/captain`, {
                              method: "PUT",
                              body: { version: t.version, user_id: m.user.id },
                            }),
                          "Капитан изменён",
                        )
                      }
                    >
                      Передать капитанство
                    </Button>
                    <Button
                      disabled={busy}
                      variant="danger"
                      onClick={() =>
                        void action(
                          () =>
                            api(`/teams/${teamId}/members/${m.user.id}`, {
                              method: "DELETE",
                              body: { version: t.version },
                            }),
                          "Участник удалён",
                        )
                      }
                    >
                      Удалить
                    </Button>
                  </>
                )}
              </div>
            ))}
          </Card>
          {t.captain_id === user?.id && (
            <Card className="mt-5 grid gap-4 p-5">
              <h2 className="text-xl font-bold">Пригласить игрока</h2>
              <form onSubmit={invite} className="flex gap-3">
                <Input
                  name="nickname"
                  aria-label="Никнейм игрока"
                  pattern="[A-Za-z0-9_]{3,32}"
                  placeholder="nickname"
                  required
                />
                <Button disabled={busy}>Пригласить</Button>
              </form>
              <Button
                disabled={busy}
                variant="secondary"
                onClick={() =>
                  void action(async () => {
                    const result = await api<Schema["InvitationLink"]>(
                      `/teams/${teamId}/invitations`,
                      { method: "POST", body: {} },
                    );
                    setLink(result.url);
                  }, "Ссылка создана")
                }
              >
                Создать одноразовую ссылку
              </Button>
              {link && (
                <Input
                  aria-label="Ссылка приглашения"
                  readOnly
                  value={link}
                  onFocus={(e) => e.target.select()}
                />
              )}
            </Card>
          )}
        </>
      )}
    </section>
  );
}
