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
const steps = [
  "Основные сведения",
  "Формат и регламент",
  "Регистрация",
  "Проверка",
];
const tieLabels: Record<Schema["TieBreaker"], string> = {
  head_to_head: "Личные встречи",
  score_difference: "Разница счёта",
  score_for: "Забито",
  elo: "Рейтинг Эло",
  draw: "Жребий",
};
const formats: Record<Schema["TournamentFormat"], string> = {
  single_elimination: "Олимпийская система",
  round_robin: "Круговая система",
  groups_playoffs: "Группы + плей-офф",
};
export function TournamentCreatePage() {
  const ds = useResource(disciplines);
  const { user } = useAuth();
  const navigate = useNavigate();
  const [format, setFormat] =
    useState<Schema["TournamentFormat"]>("single_elimination");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [step, setStep] = useState(0);
  const [summary, setSummary] = useState<[string, string][]>([]);
  const [wildcards, setWildcards] = useState(0);
  const [tieBreakers, setTieBreakers] = useState<Schema["TieBreaker"][]>(
    Object.keys(tieLabels) as Schema["TieBreaker"][],
  );
  function next(form: HTMLFormElement) {
    const fields = form.querySelectorAll<HTMLInputElement | HTMLSelectElement>(
      `[data-step="${step}"] input, [data-step="${step}"] select`,
    );
    for (const field of fields) if (!field.reportValidity()) return;
    if (step === 2) {
      const f = new FormData(form);
      const value = (key: string) => String(f.get(key) || "");
      const dates: [string, string][] = [
        [
          "Открытие регистрации",
          value("registration_opens_at").replace("T", " "),
        ],
        [
          "Закрытие регистрации",
          value("registration_closes_at").replace("T", " "),
        ],
        ["Начало турнира", value("starts_at").replace("T", " ")],
        ["Окончание турнира", value("ends_at").replace("T", " ")],
      ];
      setSummary([
        ["Название", value("name")],
        ["Описание", value("description")],
        [
          "Дисциплина",
          ds.data?.find((d) => d.id === value("discipline"))?.name ?? "",
        ],
        ["Формат", formats[format]],
        ["Серия", `Bo${value("bestOf")}`],
        ["Максимум участников", value("capacity")],
        [
          "Дополнительные показатели",
          tieBreakers.map((key) => tieLabels[key]).join(", "),
        ],
        ...dates,
      ]);
    }
    setStep(step + 1);
  }
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (step < 3) {
      next(e.currentTarget);
      return;
    }
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
        round_robin_legs: f.get("legs") === "2" ? 2 : 1,
        win_points: Number(f.get("win")),
        draw_points: Number(f.get("draw")),
        loss_points: Number(f.get("loss")),
        tie_breakers: tieBreakers,
        third_place_match: f.get("third") === "on",
        confirmation_timeout_hours: Number(f.get("timeout")),
        ...(format === "groups_playoffs"
          ? {
              group_settings: {
                group_count: Number(f.get("groups")),
                qualifiers_per_group: Number(f.get("qualifiers")),
                wildcard_count: wildcards,
                ...(wildcards > 0
                  ? { wildcard_place: Number(f.get("wildcardPlace")) }
                  : {}),
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
        <h2 className="mb-4 font-bold">
          Шаг {step + 1} из {steps.length}. {steps[step]}
        </h2>
        <form className="grid gap-4" onSubmit={submit} noValidate>
          <fieldset
            data-step={0}
            className={step === 0 ? "grid gap-4 sm:grid-cols-2" : "hidden"}
          >
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
          </fieldset>
          <fieldset
            data-step={1}
            className={step === 1 ? "grid gap-4 sm:grid-cols-2" : "hidden"}
          >
            <label>
              Формат
              <Select
                value={format}
                onChange={(e) =>
                  setFormat(e.target.value as Schema["TournamentFormat"])
                }
              >
                {Object.entries(formats).map(([key, label]) => (
                  <option key={key} value={key}>
                    {label}
                  </option>
                ))}
              </Select>
            </label>
            {format !== "single_elimination" && (
              <label>
                Число кругов
                <Select name="legs">
                  <option value="1">Один</option>
                  <option value="2">Два</option>
                </Select>
              </label>
            )}
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
                  Дополнительные места в плей-офф
                  <Input
                    name="wildcards"
                    type="number"
                    min={0}
                    max={128}
                    required
                    value={wildcards}
                    onChange={(e) => setWildcards(Number(e.target.value))}
                  />
                </label>
                {wildcards > 0 && (
                  <label>
                    Место для отбора лучших команд
                    <Input
                      name="wildcardPlace"
                      type="number"
                      min={2}
                      max={128}
                      defaultValue={3}
                      required
                    />
                  </label>
                )}
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
            {format !== "round_robin" && (
              <label className="sm:col-span-2">
                <input name="third" type="checkbox" /> Матч за третье место
              </label>
            )}
            <div className="sm:col-span-2">
              <p className="font-semibold">
                Дополнительные показатели в порядке применения
              </p>
              <ol className="list-decimal space-y-2 pl-5">
                {tieBreakers.map((key, i) => (
                  <li key={key}>
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="flex-1">{tieLabels[key]}</span>
                      <Button
                        type="button"
                        variant="secondary"
                        disabled={i === 0}
                        aria-label={`Поднять: ${tieLabels[key]}`}
                        onClick={() =>
                          setTieBreakers((current) => {
                            const ordered = [...current];
                            [ordered[i - 1], ordered[i]] = [
                              ordered[i],
                              ordered[i - 1],
                            ];
                            return ordered;
                          })
                        }
                      >
                        Выше
                      </Button>
                      <Button
                        type="button"
                        variant="secondary"
                        disabled={tieBreakers.length === 1}
                        aria-label={`Убрать: ${tieLabels[key]}`}
                        onClick={() =>
                          setTieBreakers((current) =>
                            current.filter((item) => item !== key),
                          )
                        }
                      >
                        Убрать
                      </Button>
                    </div>
                  </li>
                ))}
              </ol>
              {Object.entries(tieLabels)
                .filter(
                  ([key]) => !tieBreakers.includes(key as Schema["TieBreaker"]),
                )
                .map(([key, label]) => (
                  <Button
                    key={key}
                    type="button"
                    variant="secondary"
                    onClick={() =>
                      setTieBreakers((current) => [
                        ...current,
                        key as Schema["TieBreaker"],
                      ])
                    }
                  >
                    Добавить: {label}
                  </Button>
                ))}
            </div>
          </fieldset>
          <fieldset
            data-step={2}
            className={step === 2 ? "grid gap-4 sm:grid-cols-2" : "hidden"}
          >
            <label>
              Часовой пояс
              <Input name="timezone" defaultValue={user.timezone} required />
            </label>
            <p className="sm:col-span-2 text-sm text-slate-500">
              Даты вводятся в часовом поясе вашего устройства (
              {Intl.DateTimeFormat().resolvedOptions().timeZone}). Для
              публикации заполните все четыре даты.
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
          </fieldset>
          {step === 3 && (
            <dl className="grid gap-2">
              {summary.map(([label, value]) => (
                <div key={label}>
                  <dt className="font-semibold">{label}</dt>
                  <dd>{value || "Не указано"}</dd>
                </div>
              ))}
            </dl>
          )}
          <div className="flex flex-wrap gap-3">
            <Feedback error={error} />
            {!ds.loading && !ds.error && !ds.data?.length && (
              <p>Администратор должен добавить дисциплины в справочник.</p>
            )}
            {step > 0 && (
              <Button
                type="button"
                variant="secondary"
                onClick={() => setStep(step - 1)}
              >
                Назад
              </Button>
            )}
            <Button
              type="submit"
              disabled={busy || ds.loading || !ds.data?.length}
            >
              {step === 3 ? "Сохранить черновик" : "Далее"}
            </Button>
          </div>
        </form>
      </Card>
    </section>
  );
}
