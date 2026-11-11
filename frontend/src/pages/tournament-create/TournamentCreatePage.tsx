import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { api, errorMessage, type Schema } from "../../api/client";
import { disciplines } from "../../api/catalog";
import { useResource } from "../../hooks/useResource";
import { useAuth } from "../../hooks/useAuth";
import { Button } from "../../ui/Button/Button";
import { Card } from "../../ui/Card/Card";
import { Input } from "../../ui/Input/Input";
import { Select } from "../../ui/Select/Select";
import { Feedback } from "../../ui/Feedback";
export function TournamentCreatePage() {
  const ds = useResource(disciplines);
  const { user } = useAuth();
  const navigate = useNavigate();
  const [format, setFormat] =
    useState<Schema["TournamentFormat"]>("single_elimination");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    setBusy(true);
    setError("");
    const body: Schema["CreateTournamentRequest"] = {
      name: String(f.get("name")),
      description: String(f.get("description")),
      discipline_id: String(f.get("discipline")),
      format,
      max_participants: Number(f.get("capacity")),
      timezone: String(f.get("timezone")),
      rules: {
        best_of: Number(f.get("bestOf")) as 1 | 3 | 5,
        round_robin_legs: 1,
        win_points: Number(f.get("win")),
        draw_points: Number(f.get("draw")),
        loss_points: Number(f.get("loss")),
        tie_breakers: [
          "head_to_head",
          "score_difference",
          "score_for",
          "elo",
          "draw",
        ],
        third_place_match: f.get("third") === "on",
        confirmation_timeout_hours: Number(f.get("timeout")),
        ...(format === "groups_playoffs"
          ? {
              group_settings: {
                group_count: Number(f.get("groups")),
                qualifiers_per_group: Number(f.get("qualifiers")),
                wildcard_count: 0,
              },
            }
          : {}),
      },
    };
    for (const key of [
      "registration_opens_at",
      "registration_closes_at",
      "starts_at",
      "ends_at",
    ] as const) {
      const value = String(f.get(key) || "");
      if (value) body[key] = new Date(value).toISOString();
    }
    try {
      const result = await api<Schema["Tournament"]>("/tournaments", {
        method: "POST",
        body,
      });
      navigate(`/tournaments/${result.id}`);
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  if (!user?.roles.some((r) => r === "organizer" || r === "admin"))
    return (
      <Feedback error="Создание турниров доступно организаторам. Обратитесь к администратору для назначения роли." />
    );
  return (
    <section className="mx-auto max-w-4xl">
      <h1 className="text-3xl font-extrabold">Новый турнир</h1>
      <p className="mt-2 text-slate-500">
        Сохраните черновик, затем проверьте сведения и опубликуйте турнир.
      </p>
      <Feedback loading={ds.loading} error={ds.error} retry={ds.reload} />
      <Card className="mt-6 p-6">
        <form className="grid gap-4 sm:grid-cols-2" onSubmit={submit}>
          <label>
            Название
            <Input name="name" required maxLength={120} />
          </label>
          <label>
            Дисциплина
            <Select name="discipline" required>
              <option value="">Выберите дисциплину</option>
              {ds.data
                ?.filter((d) => d.active)
                .map((d) => (
                  <option value={d.id} key={d.id}>
                    {d.name}
                  </option>
                ))}
            </Select>
          </label>
          <label className="sm:col-span-2">
            Описание
            <Input name="description" maxLength={10000} />
          </label>
          <label>
            Формат
            <Select
              value={format}
              onChange={(e) =>
                setFormat(e.target.value as Schema["TournamentFormat"])
              }
            >
              <option value="single_elimination">Олимпийская система</option>
              <option value="round_robin">Круговая система</option>
              <option value="groups_playoffs">Группы + плей-офф</option>
            </Select>
          </label>
          <label>
            Максимум участников
            <Input
              name="capacity"
              type="number"
              min={2}
              max={256}
              defaultValue={8}
              required
            />
          </label>
          {format === "groups_playoffs" && (
            <>
              <label>
                Число групп
                <Input
                  name="groups"
                  type="number"
                  min={2}
                  max={64}
                  defaultValue={2}
                  required
                />
              </label>
              <label>
                Выходят из группы
                <Input
                  name="qualifiers"
                  type="number"
                  min={1}
                  max={128}
                  defaultValue={2}
                  required
                />
              </label>
            </>
          )}
          <label>
            Серия
            <Select name="bestOf">
              <option value="1">Bo1</option>
              <option value="3">Bo3</option>
              <option value="5">Bo5</option>
            </Select>
          </label>
          <label>
            Подтверждение, часов
            <Input
              name="timeout"
              type="number"
              min={1}
              max={168}
              defaultValue={24}
              required
            />
          </label>
          {[
            ["win", "Очки за победу", 3],
            ["draw", "Очки за ничью", 1],
            ["loss", "Очки за поражение", 0],
          ].map(([name, label, value]) => (
            <label key={name}>
              {label}
              <Input
                name={String(name)}
                type="number"
                min={0}
                max={100}
                defaultValue={value}
                required
              />
            </label>
          ))}
          <label>
            Часовой пояс
            <Input name="timezone" defaultValue={user.timezone} required />
          </label>
          <label className="sm:col-span-2">
            <input name="third" type="checkbox" /> Матч за третье место
          </label>
          <p className="sm:col-span-2 text-sm text-slate-500">
            Даты вводятся в часовом поясе вашего устройства (
            {Intl.DateTimeFormat().resolvedOptions().timeZone}). Для публикации
            заполните все четыре даты.
          </p>
          {[
            ["registration_opens_at", "Открытие регистрации"],
            ["registration_closes_at", "Закрытие регистрации"],
            ["starts_at", "Начало турнира"],
            ["ends_at", "Окончание турнира"],
          ].map(([name, label]) => (
            <label key={name}>
              {label}
              <Input name={name} type="datetime-local" />
            </label>
          ))}
          <div className="sm:col-span-2">
            <Feedback error={error} />
            {!ds.loading && !ds.error && !ds.data?.length && (
              <p>Администратор должен добавить дисциплины через API.</p>
            )}
            <Button disabled={busy || ds.loading || !ds.data?.length}>
              Сохранить черновик
            </Button>
          </div>
        </form>
      </Card>
    </section>
  );
}
