import { useState, type FormEvent } from "react";
import { api, errorMessage, type Schema } from "../../api/client";
import { dateLabel, tournamentStatus } from "../../api/catalog";
import { useAuth } from "../../hooks/useAuth";
import { useResource } from "../../hooks/useResource";
import { Button } from "../../ui/Button/Button";
import { Card } from "../../ui/Card/Card";
import { Feedback } from "../../ui/Feedback";
import { Input } from "../../ui/Input/Input";

const sections = {
  users: "Пользователи",
  disciplines: "Дисциплины",
  tournaments: "Турниры",
  audit: "Журнал действий",
} as const;
type Section = keyof typeof sections;
const roles: Record<Schema["Role"], string> = {
  participant: "Участник",
  organizer: "Организатор",
  referee: "Судья",
  admin: "Администратор",
};
const limit = 20;

function disciplineData(fields: FormData): Schema["DisciplineRequest"] {
  return {
    name: String(fields.get("name")).trim(),
    score_unit: String(fields.get("unit")).trim(),
    min_roster_size: Number(fields.get("min")),
    max_roster_size: Number(fields.get("max")),
    allows_draw: fields.has("draw"),
  };
}

function DisciplineFields({ item }: { item?: Schema["Discipline"] }) {
  return (
    <>
      <label>
        Название
        <Input name="name" defaultValue={item?.name} required maxLength={120} />
      </label>
      <label>
        Единица счёта
        <Input
          name="unit"
          defaultValue={item?.score_unit}
          required
          maxLength={50}
        />
      </label>
      <label>
        Минимум в составе
        <Input
          name="min"
          type="number"
          defaultValue={item?.min_roster_size ?? 1}
          min={1}
          max={100}
          required
        />
      </label>
      <label>
        Максимум в составе
        <Input
          name="max"
          type="number"
          defaultValue={item?.max_roster_size ?? 1}
          min={1}
          max={100}
          required
        />
      </label>
      <label>
        <input name="draw" type="checkbox" defaultChecked={item?.allows_draw} />{" "}
        Допускается ничья
      </label>
      {item && (
        <label>
          <input name="active" type="checkbox" defaultChecked={item.active} />{" "}
          Активна
        </label>
      )}
    </>
  );
}

export function AdminPage() {
  const { user } = useAuth();
  const isAdmin = !!user?.roles.includes("admin");
  const [section, setSection] = useState<Section>("users");
  const [offset, setOffset] = useState(0);
  const [search, setSearch] = useState("");
  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState({ error: "", message: "" });
  const resource = useResource(async () => {
    if (!isAdmin) return undefined;
    const query = new URLSearchParams({
      limit: String(limit),
      offset: String(offset),
    });
    if (search && (section === "users" || section === "tournaments"))
      query.set("search", search);
    switch (section) {
      case "users":
        return {
          section,
          ...(await api<Schema["ProfilePage"]>(`/admin/users?${query}`)),
        };
      case "disciplines":
        query.set("include_inactive", "true");
        return {
          section,
          ...(await api<Schema["DisciplinePage"]>(`/disciplines?${query}`)),
        };
      case "tournaments":
        return {
          section,
          ...(await api<Schema["TournamentPage"]>(
            `/admin/tournaments?${query}`,
          )),
        };
      case "audit":
        return {
          section,
          ...(await api<Schema["AuditEntryPage"]>(`/admin/audit?${query}`)),
        };
    }
  }, [isAdmin, section, offset, search]);
  const result = resource.data?.section === section ? resource.data : undefined;

  function mutate(
    event: FormEvent<HTMLFormElement>,
    work: (fields: FormData) => Promise<unknown>,
    message: string,
  ) {
    event.preventDefault();
    const fields = new FormData(event.currentTarget);
    setBusy(true);
    setFeedback({ error: "", message: "" });
    void work(fields)
      .then(() => {
        setFeedback({ error: "", message });
        resource.reload();
      })
      .catch((error) =>
        setFeedback({ error: errorMessage(error), message: "" }),
      )
      .finally(() => setBusy(false));
  }

  if (!isAdmin) return <Feedback error="Раздел доступен администраторам." />;
  return (
    <section>
      <h1 className="mb-4 text-3xl font-extrabold">Администрирование</h1>
      <nav
        aria-label="Разделы администрирования"
        className="mb-4 flex flex-wrap gap-2"
      >
        {(Object.entries(sections) as [Section, string][]).map(
          ([id, label]) => (
            <Button
              key={id}
              disabled={busy}
              aria-pressed={section === id}
              variant={section === id ? "primary" : "secondary"}
              onClick={() => {
                setSection(id);
                setOffset(0);
                setSearch("");
                setFeedback({ error: "", message: "" });
              }}
            >
              {label}
            </Button>
          ),
        )}
      </nav>
      {(section === "users" || section === "tournaments") && (
        <form
          key={section}
          className="mb-4 flex flex-wrap items-end gap-2"
          onSubmit={(event) => {
            event.preventDefault();
            setOffset(0);
            setSearch(
              String(new FormData(event.currentTarget).get("search")).trim(),
            );
          }}
        >
          <label className="min-w-48 flex-1">
            Поиск
            <Input name="search" type="search" />
          </label>
          <Button>Найти</Button>
        </form>
      )}
      <Feedback
        loading={resource.loading}
        error={resource.error}
        retry={resource.reload}
      />
      <Feedback error={feedback.error} />
      {feedback.message && (
        <p className="my-3 text-emerald-700" role="status">
          {feedback.message}
        </p>
      )}
      {section === "disciplines" && (
        <Card className="my-4 p-4">
          <h2 className="mb-2 font-bold">Новая дисциплина</h2>
          <form
            className="grid gap-2 sm:grid-cols-2"
            onSubmit={(event) =>
              mutate(
                event,
                (fields) =>
                  api("/admin/disciplines", {
                    method: "POST",
                    body: disciplineData(fields),
                  }),
                "Дисциплина добавлена",
              )
            }
          >
            <DisciplineFields />
            <Button disabled={busy}>Добавить дисциплину</Button>
          </form>
        </Card>
      )}
      <div className="grid gap-3">
        {result?.section === "users" &&
          result.items.map((p) => (
            <Card key={`${p.id}/${p.version}`} className="p-4">
              <h2 className="font-bold">@{p.nickname}</h2>
              <p className="break-words">
                {p.email} ·{" "}
                {p.email_verified
                  ? "Почта подтверждена"
                  : "Почта не подтверждена"}{" "}
                · {p.blocked ? "Заблокирован" : "Активен"}
              </p>
              <form
                className="my-3 flex flex-wrap items-center gap-3"
                onSubmit={(event) =>
                  mutate(
                    event,
                    (fields) =>
                      api(`/admin/users/${p.id}/roles`, {
                        method: "PUT",
                        body: {
                          version: p.version,
                          roles: fields.getAll("role"),
                        },
                      }),
                    "Роли сохранены",
                  )
                }
              >
                {(Object.entries(roles) as [Schema["Role"], string][]).map(
                  ([role, label]) => (
                    <label key={role}>
                      <input
                        name="role"
                        type="checkbox"
                        value={role}
                        defaultChecked={p.roles.includes(role)}
                      />{" "}
                      {label}
                    </label>
                  ),
                )}
                <Button disabled={busy}>Сохранить роли</Button>
              </form>
              <form
                className="flex flex-wrap items-end gap-2"
                onSubmit={(event) =>
                  mutate(
                    event,
                    (fields) =>
                      api(`/admin/users/${p.id}/block`, {
                        method: "PUT",
                        body: {
                          version: p.version,
                          blocked: !p.blocked,
                          ...(!p.blocked
                            ? { reason: String(fields.get("reason")).trim() }
                            : {}),
                        },
                      }),
                    p.blocked
                      ? "Пользователь разблокирован"
                      : "Пользователь заблокирован",
                  )
                }
              >
                {!p.blocked && (
                  <label className="min-w-48 flex-1">
                    Причина блокировки
                    <Input name="reason" required maxLength={2000} />
                  </label>
                )}
                <Button
                  disabled={busy}
                  variant={p.blocked ? "secondary" : "danger"}
                >
                  {p.blocked ? "Разблокировать" : "Заблокировать"}
                </Button>
              </form>
            </Card>
          ))}
        {result?.section === "disciplines" &&
          result.items.map((d) => (
            <Card key={`${d.id}/${d.version}`} className="p-4">
              <form
                className="grid gap-2 sm:grid-cols-2"
                onSubmit={(event) =>
                  mutate(
                    event,
                    (fields) =>
                      api(`/admin/disciplines/${d.id}`, {
                        method: "PATCH",
                        body: {
                          ...disciplineData(fields),
                          active: fields.has("active"),
                          version: d.version,
                        },
                      }),
                    "Дисциплина сохранена",
                  )
                }
              >
                <DisciplineFields item={d} />
                <Button disabled={busy}>Сохранить дисциплину</Button>
              </form>
            </Card>
          ))}
        {result?.section === "tournaments" &&
          result.items.map((t) => (
            <Card key={`${t.id}/${t.version}`} className="p-4">
              <h2 className="font-bold">{t.name}</h2>
              <p>
                {tournamentStatus(t)} · @{t.organizer.nickname} ·{" "}
                {t.hidden ? "Скрыт" : "Виден"}
              </p>
              <form
                className="mt-3 flex flex-wrap items-end gap-2"
                onSubmit={(event) => {
                  const deleting =
                    (event.nativeEvent as SubmitEvent).submitter?.getAttribute(
                      "value",
                    ) === "delete";
                  if (
                    deleting &&
                    !window.confirm(`Удалить турнир «${t.name}»?`)
                  ) {
                    event.preventDefault();
                    return;
                  }
                  mutate(
                    event,
                    (fields) => {
                      const body = {
                        version: t.version,
                        reason: String(fields.get("reason")).trim(),
                      };
                      return deleting
                        ? api(`/admin/tournaments/${t.id}`, {
                            method: "DELETE",
                            body,
                          })
                        : api(`/admin/tournaments/${t.id}/visibility`, {
                            method: "PUT",
                            body: { ...body, hidden: !t.hidden },
                          });
                    },
                    deleting
                      ? "Турнир удалён"
                      : t.hidden
                        ? "Турнир открыт"
                        : "Турнир скрыт",
                  );
                }}
              >
                <label className="min-w-48 flex-1">
                  Причина модерации
                  <Input name="reason" required maxLength={2000} />
                </label>
                <Button disabled={busy} variant="secondary">
                  {t.hidden ? "Показать" : "Скрыть"}
                </Button>
                <Button value="delete" disabled={busy} variant="danger">
                  Удалить
                </Button>
              </form>
            </Card>
          ))}
        {result?.section === "audit" &&
          result.items.map((entry) => (
            <Card key={entry.id} className="p-4">
              <h2 className="font-bold">
                {entry.action} · {dateLabel(entry.created_at)}
              </h2>
              <p className="break-words">
                Объект {entry.entity_type} · {entry.entity_id}
              </p>
              <p className="break-words">Автор {entry.actor_id ?? "система"}</p>
              <details>
                <summary>Изменения</summary>
                <pre className="whitespace-pre-wrap break-all">
                  {JSON.stringify(
                    { до: entry.before, после: entry.after },
                    null,
                    2,
                  )}
                </pre>
              </details>
            </Card>
          ))}
      </div>
      {result?.total === 0 && <p>Записей нет.</p>}
      {result && result.total > limit && (
        <div className="mt-4 flex items-center gap-3">
          <Button
            variant="secondary"
            disabled={busy || offset === 0}
            onClick={() => setOffset(offset - limit)}
          >
            Назад
          </Button>
          <span>
            {offset + 1}–{Math.min(offset + limit, result.total)} из{" "}
            {result.total}
          </span>
          <Button
            variant="secondary"
            disabled={busy || offset + limit >= result.total}
            onClick={() => setOffset(offset + limit)}
          >
            Далее
          </Button>
        </div>
      )}
    </section>
  );
}
