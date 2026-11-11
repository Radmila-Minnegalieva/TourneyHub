import { useState } from "react";
import { Link } from "react-router-dom";
import { allPages, type Schema } from "../../api/client";
import { disciplines } from "../../api/catalog";
import { useResource } from "../../hooks/useResource";
import { Card } from "../../ui/Card/Card";
import { Select } from "../../ui/Select/Select";
import { Feedback } from "../../ui/Feedback";
export function RatingPage() {
  const ds = useResource(disciplines);
  const [selected, setSelected] = useState("");
  const id = selected || ds.data?.[0]?.id;
  const r = useResource(
    () =>
      id
        ? allPages<Schema["Rating"]>(`/ratings?discipline_id=${id}`, false)
        : Promise.resolve([]),
    [id],
  );
  return (
    <section>
      <h1 className="text-3xl font-extrabold">Рейтинг команд</h1>
      <p className="my-3 text-slate-500">
        Рейтинг Эло по подтверждённым результатам матчей.
      </p>
      <Select
        aria-label="Дисциплина"
        value={id || ""}
        onChange={(e) => setSelected(e.target.value)}
      >
        {ds.data?.map((d) => (
          <option value={d.id} key={d.id}>
            {d.name}
          </option>
        ))}
      </Select>
      <Feedback
        loading={ds.loading || r.loading}
        error={ds.error || r.error}
        retry={() => {
          ds.reload();
          r.reload();
        }}
      />
      <Card className="mt-5 overflow-x-auto p-5">
        <table className="w-full text-left">
          <thead>
            <tr>
              <th>Место</th>
              <th>Команда</th>
              <th>Эло</th>
              <th>Матчи</th>
            </tr>
          </thead>
          <tbody>
            {r.data?.map((entry) => (
              <tr key={entry.team.id} className="border-t">
                <td className="py-3">{entry.position}</td>
                <td>
                  <Link to={`/teams/${entry.team.id}`}>{entry.team.name}</Link>
                </td>
                <td>{Math.round(entry.value)}</td>
                <td>{entry.played}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {!r.loading && !r.error && !r.data?.length && (
          <p className="mt-4">Рейтинг пока пуст.</p>
        )}
      </Card>
    </section>
  );
}
