import { Input } from "../../ui/Input/Input";
import { Select } from "../../ui/Select/Select";
import { Button } from "../../ui/Button/Button";
import type { TournamentFilters as Filters } from "../../hooks/useTournaments";

interface Props {
  disciplines: { id: string; name: string }[];
  filters: Filters;
  onChange: (filters: Filters) => void;
  onReset: () => void;
}

export function TournamentFilters({
  filters,
  onChange,
  onReset,
  disciplines,
}: Props) {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <label className="grid gap-1.5 text-sm font-semibold">
        Поиск
        <Input
          value={filters.query}
          onChange={(event) =>
            onChange({ ...filters, query: event.target.value })
          }
          placeholder="Название турнира или организатор"
        />
      </label>
      <label className="grid gap-1.5 text-sm font-semibold">
        Начало турнира с
        <Input
          type="date"
          value={filters.startsFrom}
          max={filters.startsTo || undefined}
          onChange={(e) => onChange({ ...filters, startsFrom: e.target.value })}
        />
      </label>
      <label className="grid gap-1.5 text-sm font-semibold">
        Начало турнира по
        <Input
          type="date"
          value={filters.startsTo}
          min={filters.startsFrom || undefined}
          onChange={(e) => onChange({ ...filters, startsTo: e.target.value })}
        />
      </label>
      <label className="grid gap-1.5 text-sm font-semibold">
        Дисциплина
        <Select
          value={filters.discipline}
          onChange={(event) =>
            onChange({ ...filters, discipline: event.target.value })
          }
        >
          <option value="">Все дисциплины</option>
          {disciplines.map((d) => (
            <option key={d.id} value={d.id}>
              {d.name}
            </option>
          ))}
        </Select>
      </label>
      <label className="grid gap-1.5 text-sm font-semibold">
        Статус
        <Select
          value={filters.status}
          onChange={(event) =>
            onChange({
              ...filters,
              status: event.target.value as Filters["status"],
            })
          }
        >
          <option value="">Любой</option>
          <option value="registration">Регистрация открыта</option>
          <option value="running">Идёт</option>
          <option value="completed">Завершён</option>
        </Select>
      </label>
      <label className="grid gap-1.5 text-sm font-semibold">
        Формат
        <Select
          value={filters.format}
          onChange={(event) =>
            onChange({
              ...filters,
              format: event.target.value as Filters["format"],
            })
          }
        >
          <option value="">Любой</option>
          <option value="single-elimination">Олимпийская система</option>
          <option value="round-robin">Круговая система</option>
          <option value="groups-playoff">Группы + плей-офф</option>
        </Select>
      </label>
      <Button variant="secondary" className="self-end" onClick={onReset}>
        Сбросить
      </Button>
    </div>
  );
}
