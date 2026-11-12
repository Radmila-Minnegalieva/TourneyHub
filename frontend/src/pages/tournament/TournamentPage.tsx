import { useState, type FormEvent } from "react";
import { Link, useParams } from "react-router-dom";
import {
  api,
  allPages,
  errorMessage,
  type Schema,
  API_BASE,
  ApiError,
} from "../../api/client";
import {
  disciplines,
  dateLabel,
  registrationOpen,
  tournamentStatus,
  matchLabels,
} from "../../api/catalog";
import { useResource } from "../../hooks/useResource";
import { useAuth } from "../../hooks/useAuth";
import { useTournamentEvents } from "../../hooks/useTournamentEvents";
import { Button } from "../../ui/Button/Button";
import { Card } from "../../ui/Card/Card";
import { Input } from "../../ui/Input/Input";
import { Select } from "../../ui/Select/Select";
import { Feedback } from "../../ui/Feedback";
import { Tabs } from "../../ui/Tabs/Tabs";
export function TournamentPage() {
  const { tournamentId = "", tab = "groups" } = useParams();
  const { user } = useAuth();
  const base = `/tournaments/${tournamentId}`;
  const r = useResource(async () => {
    let t: Schema["Tournament"];
    try {
      t = await api<Schema["Tournament"]>(base);
    } catch (e) {
      if (!(e instanceof ApiError) || e.status !== 404 || !user) throw e;
      const own = await allPages<Schema["Tournament"]>(
        user.roles.includes("admin")
          ? "/admin/tournaments"
          : "/users/me/tournaments",
      );
      const draft = own.find((item) => item.id === tournamentId);
      if (!draft) throw e;
      t = draft;
    }
    const owner = t.organizer.id === user?.id || user?.roles.includes("admin");
    const ds = await disciplines();
    if (t.status === "draft")
      return {
        t,
        ds,
        owner,
        applications: [] as Schema["Application"][],
        bracket: undefined as Schema["Bracket"] | undefined,
        standings: [] as Schema["Standings"][],
      };
    const [applications, bracket, standings] = await Promise.all([
      allPages<Schema["Application"]>(
        `${base}/${owner ? "applications" : "participants"}`,
      ),
      api<Schema["Bracket"]>(`${base}/bracket`),
      api<Schema["Standings"][]>(`${base}/standings`),
    ]);
    return { t, ds, owner, applications, bracket, standings };
  }, [tournamentId, user?.id]);
  const visible = r.data && !r.data.t.hidden && r.data.t.status !== "draft";
  useTournamentEvents(visible ? tournamentId : undefined, r.refresh);
  const teams = useResource(
    () =>
      user ? allPages<Schema["Team"]>("/users/me/teams") : Promise.resolve([]),
    [user?.id],
  );
  const [teamId, setTeamId] = useState("");
  const members = useResource(
    () =>
      teamId
        ? api<Schema["TeamMember"][]>(`/teams/${teamId}/members`)
        : Promise.resolve([]),
    [teamId],
  );
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [preview, setPreview] = useState<Schema["DrawPreview"]>();
  const [manualOrder, setManualOrder] = useState<string[]>();
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
  function apply(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    void action(
      () =>
        api(`${base}/applications`, {
          method: "POST",
          body: {
            team_id: teamId || undefined,
            player_ids: teamId ? f.getAll("players") : [user!.id],
          },
        }),
      "Заявка отправлена",
    );
  }
  const data = r.data;
  const accepted =
    data?.applications.filter((a) => a.status === "accepted") ?? [];
  const canDraw =
    data?.t.status === "registration" ||
    (data?.t.status === "in_progress" &&
      data.bracket?.matches.every(
        (m) =>
          m.status === "bye" ||
          (m.status === "scheduled" &&
            (!m.starts_at || Date.parse(m.starts_at) > Date.now())),
      ));
  const rounds = new Map<string, Schema["Match"]>();
  const unavailableRounds = new Set<string>();
  for (const match of data?.bracket?.matches ?? []) {
    const key = `${match.stage_id}/${match.round}`;
    if (match.status === "scheduled") rounds.set(key, match);
    else if (match.status !== "bye") unavailableRounds.add(key);
  }
  for (const key of unavailableRounds) rounds.delete(key);
  function previewDraw(
    method: Schema["SeedingRequest"]["method"],
    applicationIds?: string[],
  ) {
    if (!data) return;
    setPreview(undefined);
    void action(async () => {
      setPreview(
        await api<Schema["DrawPreview"]>(`${base}/draw-previews`, {
          method: "POST",
          body: {
            version: data.t.version,
            method,
            ...(applicationIds ? { application_ids: applicationIds } : {}),
          },
        }),
      );
    }, "Предпросмотр жеребьёвки готов");
  }
  const tabs = [
    ["groups", "Таблицы"],
    ["playoff", "Сетка"],
    ["schedule", "Расписание"],
    ["participants", "Участники и заявки"],
    ["results", "Итоги"],
  ].map(([id, label]) => ({ label, to: `${base}/${id}` }));
  const name = (id?: string) =>
    data?.applications.find((a) => a.id === id)?.team.name ||
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
      {data && (
        <>
          <h1 className="text-3xl font-extrabold">{data.t.name}</h1>
          <p className="mt-2 text-slate-500">
            {tournamentStatus(data.t)} ·{" "}
            {data.ds.find((d) => d.id === data.t.discipline_id)?.name} ·
            Организатор: {data.t.organizer.name || data.t.organizer.nickname}
          </p>
          <p className="mt-2">{data.t.description}</p>
          <p className="my-3">
            {dateLabel(data.t.starts_at)} — {dateLabel(data.t.ends_at)} ·{" "}
            {data.t.accepted_count} / {data.t.max_participants} участников · Bo
            {data.t.rules.best_of}
          </p>
          {data.owner && (
            <Card className="my-5 grid gap-4 p-5">
              <h2 className="text-xl font-bold">Управление турниром</h2>
              {data.t.status === "draft" && (
                <>
                  <p>
                    Для публикации нужны даты: открытие регистрации &lt;
                    закрытие ≤ начало &lt; окончание.
                  </p>
                  <form
                    className="grid gap-3 sm:grid-cols-2"
                    onSubmit={(e) => {
                      e.preventDefault();
                      const f = new FormData(e.currentTarget);
                      const body: Schema["UpdateTournamentRequest"] = {
                        version: data.t.version,
                      };
                      for (const k of [
                        "registration_opens_at",
                        "registration_closes_at",
                        "starts_at",
                        "ends_at",
                      ] as const) {
                        const v = String(f.get(k) || "");
                        if (v) body[k] = new Date(v).toISOString();
                      }
                      void action(
                        () => api(base, { method: "PATCH", body }),
                        "Даты сохранены",
                      );
                    }}
                  >
                    {[
                      ["registration_opens_at", "Открытие регистрации"],
                      ["registration_closes_at", "Закрытие регистрации"],
                      ["starts_at", "Начало"],
                      ["ends_at", "Окончание"],
                    ].map(([key, label]) => (
                      <label key={key}>
                        {label}
                        <Input name={key} type="datetime-local" required />
                      </label>
                    ))}
                    <Button disabled={busy}>Сохранить даты</Button>
                  </form>
                  <Button
                    disabled={busy}
                    onClick={() =>
                      void action(
                        () =>
                          api(`${base}/publish`, {
                            method: "POST",
                            body: { version: data.t.version },
                          }),
                        "Турнир опубликован",
                      )
                    }
                  >
                    Опубликовать
                  </Button>
                </>
              )}
              {canDraw && (
                <>
                  <Button disabled={busy} onClick={() => previewDraw("elo")}>
                    Предпросмотр жеребьёвки по Эло
                  </Button>
                  <Button
                    disabled={busy}
                    variant="secondary"
                    onClick={() => previewDraw("random")}
                  >
                    Предпросмотр случайной жеребьёвки
                  </Button>
                  <Button
                    disabled={busy || accepted.length < 2}
                    variant="secondary"
                    onClick={() =>
                      setManualOrder(
                        preview?.seeds.map((s) => s.application_id) ??
                          accepted.map((a) => a.id),
                      )
                    }
                  >
                    Настроить посев вручную
                  </Button>
                  {manualOrder && (
                    <div>
                      <ol className="list-decimal space-y-2 pl-5">
                        {manualOrder.map((id, index) => (
                          <li key={id}>
                            <div className="flex items-center gap-2">
                              <span className="flex-1">{name(id)}</span>
                              <Button
                                type="button"
                                variant="secondary"
                                disabled={index === 0}
                                aria-label={`Поднять ${name(id)}`}
                                onClick={() => {
                                  const next = [...manualOrder];
                                  [next[index - 1], next[index]] = [
                                    next[index],
                                    next[index - 1],
                                  ];
                                  setManualOrder(next);
                                  setPreview(undefined);
                                }}
                              >
                                Выше
                              </Button>
                            </div>
                          </li>
                        ))}
                      </ol>
                      <Button
                        disabled={busy}
                        onClick={() => previewDraw("manual", manualOrder)}
                      >
                        Предпросмотр ручного посева
                      </Button>
                    </div>
                  )}
                  {preview && (
                    <div>
                      <p>
                        {preview.seeds.length} участников,{" "}
                        {preview.matches.length} матчей
                      </p>
                      <ol>
                        {preview.seeds.map((s) => (
                          <li key={s.application_id}>
                            {s.position}. {name(s.application_id)}
                          </li>
                        ))}
                      </ol>
                      {preview.groups.map((g) => (
                        <p key={g.id}>
                          {g.name} ·{" "}
                          {
                            preview.matches.filter((m) => m.group_id === g.id)
                              .length
                          }{" "}
                          матчей
                        </p>
                      ))}
                      <ul className="max-h-48 overflow-auto">
                        {preview.matches.map((m) => (
                          <li key={m.id}>
                            Раунд {m.round}: {name(m.home_application_id)} —{" "}
                            {name(m.away_application_id)}
                          </li>
                        ))}
                      </ul>
                      <Button
                        disabled={busy}
                        onClick={() => {
                          if (
                            data.t.status === "in_progress" &&
                            !window.confirm(
                              "Пересобрать сетку и заменить текущее расписание?",
                            )
                          )
                            return;
                          void action(async () => {
                            await api(`${base}/draw`, {
                              method: "PUT",
                              body: {
                                version: preview.tournament_version,
                                preview_id: preview.id,
                              },
                            });
                            setPreview(undefined);
                            setManualOrder(undefined);
                          }, "Жеребьёвка подтверждена");
                        }}
                      >
                        Подтвердить жеребьёвку
                      </Button>
                    </div>
                  )}
                </>
              )}
              {data.t.status === "in_progress" && (
                <Button
                  disabled={busy}
                  onClick={() =>
                    void action(
                      () =>
                        api(`${base}/complete`, {
                          method: "POST",
                          body: { version: data.t.version },
                        }),
                      "Турнир завершён",
                    )
                  }
                >
                  Завершить турнир
                </Button>
              )}
            </Card>
          )}
          {registrationOpen(data.t) && user && (
            <Card className="my-5 p-5">
              <h2 className="text-xl font-bold">Подать заявку</h2>
              <Feedback error={teams.error || members.error} />
              <form onSubmit={apply} className="mt-3 grid gap-3">
                <Select
                  aria-label="Команда для заявки"
                  value={teamId}
                  onChange={(e) => setTeamId(e.target.value)}
                >
                  <option value="">
                    Личное участие (индивидуальная дисциплина)
                  </option>
                  {teams.data
                    ?.filter((t) => t.captain_id === user.id)
                    .map((t) => (
                      <option value={t.id} key={t.id}>
                        {t.name}
                      </option>
                    ))}
                </Select>
                {members.data?.map((m) => (
                  <label key={m.user.id}>
                    <input name="players" type="checkbox" value={m.user.id} />{" "}
                    {m.user.name || m.user.nickname}
                  </label>
                ))}
                <Button disabled={busy || members.loading}>
                  Подать заявку
                </Button>
              </form>
            </Card>
          )}
          <Tabs items={tabs} />
          {tab === "schedule" &&
            data.owner &&
            data.t.status === "in_progress" &&
            rounds.size > 0 && (
              <Card className="mt-5 p-5">
                <h2 className="mb-3 font-bold">Назначить расписание тура</h2>
                <form
                  className="grid gap-3 sm:grid-cols-2"
                  onSubmit={(event) => {
                    event.preventDefault();
                    const fields = new FormData(event.currentTarget);
                    const selected = rounds.get(String(fields.get("round")));
                    if (!selected) return;
                    void action(
                      () =>
                        api(`${base}/schedule-round`, {
                          method: "POST",
                          body: {
                            version: data.t.version,
                            stage_id: selected.stage_id,
                            round: selected.round,
                            starts_at: new Date(
                              String(fields.get("roundStartsAt")),
                            ).toISOString(),
                            venue: String(fields.get("roundVenue")),
                            interval_minutes: Number(
                              fields.get("intervalMinutes"),
                            ),
                          },
                        }),
                      "Расписание тура сохранено",
                    );
                  }}
                >
                  <label>
                    Тур
                    <Select name="round" required>
                      {[...rounds].map(([key, m]) => (
                        <option key={key} value={key}>
                          Этап{" "}
                          {
                            data.bracket?.stages.find(
                              (s) => s.id === m.stage_id,
                            )?.position
                          }
                          , тур {m.round}
                        </option>
                      ))}
                    </Select>
                  </label>
                  <label>
                    Первый матч
                    <Input
                      name="roundStartsAt"
                      type="datetime-local"
                      required
                    />
                  </label>
                  <label>
                    Площадка
                    <Input name="roundVenue" maxLength={200} required />
                  </label>
                  <label>
                    Интервал между матчами, минут
                    <Input
                      name="intervalMinutes"
                      type="number"
                      min={0}
                      max={1440}
                      defaultValue={60}
                      required
                    />
                  </label>
                  <Button disabled={busy}>Назначить тур</Button>
                </form>
              </Card>
            )}
          {tab === "participants" || tab === "results" ? (
            <Card className="mt-5 p-5">
              {!data.applications.length && <p>Участников пока нет.</p>}
              {data.applications
                .filter((a) => tab !== "results" || a.status === "accepted")
                .sort((a, b) =>
                  tab === "results"
                    ? (a.final_place ?? Infinity) - (b.final_place ?? Infinity)
                    : 0,
                )
                .map((a) => (
                  <div
                    key={a.id}
                    className="flex flex-wrap items-center gap-3 border-b py-3"
                  >
                    <Link className="flex-1" to={`/teams/${a.team.id}`}>
                      {a.final_place ? `${a.final_place}. ` : ""}
                      {a.team.name} ·{" "}
                      {a.players.map((p) => p.nickname).join(", ")}
                    </Link>
                    <span>
                      {
                        {
                          pending: "На рассмотрении",
                          accepted: "Принята",
                          rejected: "Отклонена",
                          withdrawn: "Отозвана",
                        }[a.status]
                      }
                    </span>
                    {data.owner && a.status === "pending" && (
                      <>
                        <Button
                          disabled={busy}
                          onClick={() =>
                            void action(
                              () =>
                                api(`${base}/applications/${a.id}/review`, {
                                  method: "POST",
                                  body: {
                                    version: a.version,
                                    decision: "accept",
                                  },
                                }),
                              "Заявка принята",
                            )
                          }
                        >
                          Принять
                        </Button>
                        <form
                          className="flex gap-2"
                          onSubmit={(e) => {
                            e.preventDefault();
                            const f = new FormData(e.currentTarget);
                            void action(
                              () =>
                                api(`${base}/applications/${a.id}/review`, {
                                  method: "POST",
                                  body: {
                                    version: a.version,
                                    decision: "reject",
                                    reason: String(f.get("reason")),
                                  },
                                }),
                              "Заявка отклонена",
                            );
                          }}
                        >
                          <Input
                            name="reason"
                            aria-label="Причина отклонения"
                            required
                            maxLength={2000}
                          />
                          <Button variant="danger" disabled={busy}>
                            Отклонить
                          </Button>
                        </form>
                      </>
                    )}
                  </div>
                ))}
              {tab === "results" && data.t.status === "completed" && (
                <a
                  className="mt-4 block text-brand"
                  href={`${API_BASE}${base}/export`}
                >
                  Скачать итоги CSV
                </a>
              )}
            </Card>
          ) : tab === "groups" ? (
            <div className="mt-5 grid gap-4">
              {!data.standings.length && (
                <p>
                  Таблицы появятся после жеребьёвки кругового или группового
                  этапа.
                </p>
              )}
              {data.standings.map((s, i) => (
                <Card
                  className="overflow-x-auto p-5"
                  key={s.group_id || s.stage_id}
                >
                  <h2 className="font-bold">
                    {data.bracket?.groups.find((g) => g.id === s.group_id)
                      ?.name || `Таблица ${i + 1}`}
                  </h2>
                  <table className="mt-3 w-full text-left">
                    <thead>
                      <tr>
                        <th>Место</th>
                        <th>Команда</th>
                        <th>И</th>
                        <th>В</th>
                        <th>Н</th>
                        <th>П</th>
                        <th>Забито</th>
                        <th>Пропущено</th>
                        <th>Разница</th>
                        <th>Очки</th>
                      </tr>
                    </thead>
                    <tbody>
                      {s.rows.map((row) => (
                        <tr className="border-t" key={row.application_id}>
                          <td>{row.position}</td>
                          <td>{name(row.application_id)}</td>
                          <td>{row.played}</td>
                          <td>{row.wins}</td>
                          <td>{row.draws}</td>
                          <td>{row.losses}</td>
                          <td>{row.score_for}</td>
                          <td>{row.score_against}</td>
                          <td>{row.score_difference}</td>
                          <td>{row.points}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </Card>
              ))}
            </div>
          ) : (
            <div className="mt-5 grid gap-3">
              {!data.bracket?.matches.length && (
                <p>Матчи появятся после жеребьёвки.</p>
              )}
              {data.bracket?.matches
                .filter(
                  (m) =>
                    tab !== "playoff" ||
                    data.bracket?.stages.find((s) => s.id === m.stage_id)
                      ?.kind === "playoffs",
                )
                .map((m) => (
                  <Card key={m.id} className="p-4">
                    <Link
                      className="font-bold text-brand"
                      to={`/matches/${m.id}`}
                    >
                      Раунд {m.round}: {name(m.home_application_id)} —{" "}
                      {name(m.away_application_id)}
                    </Link>
                    <p>
                      {m.home_score ?? "—"} : {m.away_score ?? "—"} ·{" "}
                      {matchLabels[m.status]} · {dateLabel(m.starts_at)} ·{" "}
                      {m.venue}
                    </p>
                  </Card>
                ))}
            </div>
          )}
        </>
      )}
    </section>
  );
}
