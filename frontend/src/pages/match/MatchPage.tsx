import { useState, type FormEvent } from "react";
import { Link, useParams } from "react-router-dom";
import { api, allPages, errorMessage, type Schema } from "../../api/client";
import { dateLabel, matchLabels } from "../../api/catalog";
import { useResource } from "../../hooks/useResource";
import { useAuth } from "../../hooks/useAuth";
import { Button } from "../../ui/Button/Button";
import { Card } from "../../ui/Card/Card";
import { Input } from "../../ui/Input/Input";
import { Feedback } from "../../ui/Feedback";
export function MatchPage() {
  const { matchId = "" } = useParams();
  const { user } = useAuth();
  const base = `/matches/${matchId}`;
  const r = useResource(async () => {
    const m = await api<Schema["Match"]>(base);
    const [t, participants] = await Promise.all([
      api<Schema["Tournament"]>(`/tournaments/${m.tournament_id}`),
      allPages<Schema["Application"]>(
        `/tournaments/${m.tournament_id}/participants`,
      ),
    ]);
    const sides = participants.filter(
      (a) => a.id === m.home_application_id || a.id === m.away_application_id,
    );
    const owner =
      t.organizer.id === user?.id || !!user?.roles.includes("admin");
    const referee =
      !!user &&
      (user.roles.includes("admin") ||
        (user.roles.includes("judge") &&
          user.id === (m.referee_id || t.referee_id)));
    const captain = sides.some((a) => a.team.captain_id === user?.id);
    const canRead =
      !!user &&
      (owner ||
        referee ||
        sides.some((a) => a.players.some((p) => p.id === user.id)));
    const reports = canRead
      ? await api<Schema["ResultReport"][]>(`${base}/reports`)
      : [];
    return { m, t, participants, owner, referee, captain, reports };
  }, [matchId, user?.id]);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  async function action(work: () => Promise<unknown>, success: string) {
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await work();
      setMessage(success);
      r.reload();
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  const d = r.data;
  function report(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!d) return;
    const f = new FormData(e.currentTarget);
    const games: Schema["GameScore"][] = [];
    for (let i = 1; i <= d.t.rules.best_of; i++) {
      const h = f.get(`home${i}`),
        a = f.get(`away${i}`);
      if (h !== "" && a !== "" && h !== null && a !== null)
        games.push({ number: i, home_score: Number(h), away_score: Number(a) });
    }
    const home_score =
      d.t.rules.best_of === 1
        ? games[0]?.home_score
        : games.filter((g) => g.home_score > g.away_score).length;
    const away_score =
      d.t.rules.best_of === 1
        ? games[0]?.away_score
        : games.filter((g) => g.away_score > g.home_score).length;
    void action(async () => {
      const file = f.get("evidence") as File;
      const evidence_file_ids: string[] = [];
      if (file?.size) {
        const body = new FormData();
        body.set("purpose", "match_protocol");
        body.set("file", file);
        evidence_file_ids.push(
          (await api<Schema["File"]>("/files", { method: "POST", body })).id,
        );
      }
      await api(`${base}/${d.referee ? "referee-result" : "reports"}`, {
        method: d.referee ? "PUT" : "POST",
        body: {
          version: d.m.version,
          home_score,
          away_score,
          games,
          evidence_file_ids,
        },
      });
    }, "Результат отправлен");
  }
  const name = (id?: string) =>
    d?.participants.find((a) => a.id === id)?.team.name ||
    "Участник определится позже";
  return (
    <section>
      <Feedback loading={r.loading} error={r.error} retry={r.reload} />
      <Feedback error={error} />
      {message && (
        <p role="status" className="my-3 text-emerald-700">
          {message}
        </p>
      )}
      {d && (
        <>
          <Link to={`/tournaments/${d.t.id}/schedule`}>← {d.t.name}</Link>
          <Card className="mt-5 p-6 text-center">
            <h1 className="text-2xl font-extrabold">
              {name(d.m.home_application_id)} — {name(d.m.away_application_id)}
            </h1>
            <p className="my-4 text-4xl font-bold">
              {d.m.home_score ?? "—"} : {d.m.away_score ?? "—"}
            </p>
            <p>
              {matchLabels[d.m.status]} · Раунд {d.m.round}
            </p>
            <p>
              {dateLabel(d.m.starts_at)} · {d.m.venue}
            </p>
          </Card>
          {d.owner && d.m.status === "scheduled" && (
            <Card className="mt-5 p-5">
              <h2 className="text-xl font-bold">Назначить время и площадку</h2>
              <form
                className="mt-3 grid gap-3"
                onSubmit={(e) => {
                  e.preventDefault();
                  const f = new FormData(e.currentTarget);
                  void action(
                    () =>
                      api(`${base}/schedule`, {
                        method: "PUT",
                        body: {
                          version: d.m.version,
                          starts_at: new Date(
                            String(f.get("starts_at")),
                          ).toISOString(),
                          venue: String(f.get("venue")),
                        },
                      }),
                    "Расписание сохранено",
                  );
                }}
              >
                <label>
                  Начало (часовой пояс устройства)
                  <Input name="starts_at" type="datetime-local" required />
                </label>
                <label>
                  Площадка
                  <Input name="venue" required maxLength={200} />
                </label>
                <Button disabled={busy}>Назначить матч</Button>
              </form>
            </Card>
          )}
          {(d.captain || d.referee) && d.m.status === "scheduled" && (
            <Card className="mt-5 p-5">
              <h2 className="text-xl font-bold">
                Сообщить результат · Bo{d.t.rules.best_of}
              </h2>
              <form className="mt-3 grid gap-3" onSubmit={report}>
                {Array.from({ length: d.t.rules.best_of }, (_, i) => (
                  <div className="grid grid-cols-2 gap-3" key={i}>
                    <label>
                      Партия {i + 1}: {name(d.m.home_application_id)}
                      <Input
                        name={`home${i + 1}`}
                        type="number"
                        min={0}
                        max={100000}
                        required={i < (d.t.rules.best_of + 1) / 2}
                      />
                    </label>
                    <label>
                      {name(d.m.away_application_id)}
                      <Input
                        name={`away${i + 1}`}
                        type="number"
                        min={0}
                        max={100000}
                        required={i < (d.t.rules.best_of + 1) / 2}
                      />
                    </label>
                  </div>
                ))}
                <label>
                  Протокол PDF/PNG/JPEG, до 5 МиБ
                  <Input
                    name="evidence"
                    type="file"
                    accept="application/pdf,image/png,image/jpeg"
                  />
                </label>
                <Button disabled={busy}>
                  {d.referee
                    ? "Зафиксировать результат судьёй"
                    : "Отправить результат"}
                </Button>
              </form>
            </Card>
          )}
          {d.reports.length > 0 && (
            <Card className="mt-5 p-5">
              <h2 className="text-xl font-bold">Протоколы результатов</h2>
              {d.reports.map((report) => (
                <div key={report.id} className="border-b py-4">
                  <p className="font-bold">
                    {report.home_score} : {report.away_score} · {report.status}
                  </p>
                  <p>{dateLabel(report.created_at)}</p>
                  {report.games.map((g) => (
                    <p key={g.number}>
                      Партия {g.number}: {g.home_score} : {g.away_score}
                    </p>
                  ))}
                  {report.status === "pending" && (
                    <p>
                      Подтверждение до {dateLabel(report.confirmation_deadline)}
                    </p>
                  )}
                  {d.captain &&
                    report.status === "pending" &&
                    report.author_id !== user?.id && (
                      <div className="mt-3 grid gap-3">
                        <Button
                          disabled={busy}
                          onClick={() =>
                            void action(
                              () =>
                                api(`${base}/reports/${report.id}/confirm`, {
                                  method: "POST",
                                  body: { version: d.m.version },
                                }),
                              "Результат подтверждён",
                            )
                          }
                        >
                          Подтвердить результат
                        </Button>
                        <form
                          className="flex gap-3"
                          onSubmit={(e) => {
                            e.preventDefault();
                            const f = new FormData(e.currentTarget);
                            void action(
                              () =>
                                api(`${base}/reports/${report.id}/dispute`, {
                                  method: "POST",
                                  body: {
                                    version: d.m.version,
                                    reason: String(f.get("reason")),
                                  },
                                }),
                              "Спор открыт",
                            );
                          }}
                        >
                          <Input
                            name="reason"
                            aria-label="Причина спора"
                            placeholder="Причина спора"
                            maxLength={2000}
                            required
                          />
                          <Button disabled={busy} variant="danger">
                            Оспорить
                          </Button>
                        </form>
                      </div>
                    )}
                </div>
              ))}
            </Card>
          )}
        </>
      )}
    </section>
  );
}
