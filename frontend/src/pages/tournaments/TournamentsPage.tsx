import { useResource } from "../../hooks/useResource";
import { allPages, type Schema } from "../../api/client";
import { Link } from "react-router-dom";
import { Feedback } from "../../ui/Feedback";
import { useAuth } from "../../hooks/useAuth";
import { useNavigate } from "react-router-dom";
import { ROUTES } from "../../constants/routes";
import { useTournaments } from "../../hooks/useTournaments";
import { TournamentCard } from "../../containers/tournament-list/TournamentCard";
import { TournamentFilters } from "../../containers/tournament-list/TournamentFilters";
import { Button } from "../../ui/Button/Button";

export function TournamentsPage() {
  const navigate = useNavigate();
  const {
    tournaments,
    filters,
    setFilters,
    reset,
    disciplines,
    loading,
    error,
    reload,
  } = useTournaments();

  const { user } = useAuth();
  const own = useResource(
    () =>
      user?.roles.some((r) => r === "organizer" || r === "admin")
        ? allPages<Schema["Tournament"]>("/users/me/tournaments")
        : Promise.resolve([]),
    [user?.id],
  );
  return (
    <section>
      <div className="mb-6 flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
        <div>
          <h1 className="text-3xl font-extrabold">Турниры</h1>
          <p className="mt-1 text-slate-500">
            Найдите соревнование и подайте заявку за свою команду
          </p>
        </div>
        {user?.roles.some((r) => r === "organizer" || r === "admin") && (
          <Button onClick={() => navigate(ROUTES.createTournament)}>
            + Создать турнир
          </Button>
        )}
      </div>
      <Feedback loading={loading} error={error} retry={reload} />
      <Feedback error={own.error} retry={own.reload} />
      {own.data?.some((t) => t.status === "draft") && (
        <div className="mb-5 rounded-xl bg-white p-4">
          <h2 className="font-bold">Мои черновики</h2>
          {own.data
            .filter((t) => t.status === "draft")
            .map((t) => (
              <Link
                className="mt-2 block text-brand"
                key={t.id}
                to={`/tournaments/${t.id}`}
              >
                {t.name}
              </Link>
            ))}
        </div>
      )}
      <TournamentFilters
        disciplines={disciplines}
        filters={filters}
        onChange={setFilters}
        onReset={reset}
      />
      <div className="mt-6 grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        {tournaments.map((tournament) => (
          <TournamentCard key={tournament.id} tournament={tournament} />
        ))}
      </div>
      {!loading && !error && tournaments.length === 0 && (
        <p className="py-16 text-center text-slate-500">
          По выбранным фильтрам турниров нет.
        </p>
      )}
    </section>
  );
}
