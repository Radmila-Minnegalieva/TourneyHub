import { Link } from "react-router-dom";
import { CreateTeamForm } from "../../containers/teams/CreateTeamForm";
import { useTeams } from "../../hooks/useTeams";
import { useResource } from "../../hooks/useResource";
import { api, allPages, type Schema } from "../../api/client";
import { Card } from "../../ui/Card/Card";
import { Feedback } from "../../ui/Feedback";
import { Button } from "../../ui/Button/Button";
import { useState } from "react";
import { errorMessage } from "../../api/client";
export function TeamsPage() {
  const { teams, createTeam, loading, error, reload } = useTeams();
  const invites = useResource(() =>
    allPages<Schema["Invitation"]>("/users/me/invitations"),
  );
  const [actionError, setActionError] = useState("");
  const [busy, setBusy] = useState(false);
  async function decide(id: string, decision: "accept" | "decline") {
    setBusy(true);
    setActionError("");
    try {
      await api(`/invitations/${id}/decision`, {
        method: "POST",
        body: { decision },
      });
      invites.reload();
      reload();
    } catch (e) {
      setActionError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <section>
      <h1 className="mb-5 text-3xl font-extrabold">Мои команды</h1>
      <CreateTeamForm onCreate={createTeam} />
      <Feedback loading={loading} error={error} retry={reload} />
      <div className="mt-6 grid gap-5 md:grid-cols-3">
        {teams.map((t) => (
          <Card key={t.id} className="p-5">
            <h2 className="text-xl font-bold">{t.name}</h2>
            <p>{t.description}</p>
            <Link className="mt-4 block text-brand" to={`/teams/${t.id}`}>
              Открыть команду
            </Link>
          </Card>
        ))}
      </div>
      {!loading && !error && !teams.length && (
        <p className="py-5">
          Команд пока нет. Создайте свою или примите приглашение.
        </p>
      )}
      <h2 className="mt-8 text-xl font-bold">Приглашения</h2>
      <Feedback
        loading={invites.loading}
        error={invites.error || actionError}
        retry={invites.reload}
      />
      {invites.data
        ?.filter((i) => i.status === "pending")
        .map((i) => (
          <Card
            key={i.id}
            className="mt-3 flex flex-wrap items-center gap-3 p-4"
          >
            <Link to={`/teams/${i.team_id}`}>Открыть команду</Link>
            <Button disabled={busy} onClick={() => void decide(i.id, "accept")}>
              Принять
            </Button>
            <Button
              disabled={busy}
              variant="secondary"
              onClick={() => void decide(i.id, "decline")}
            >
              Отклонить
            </Button>
          </Card>
        ))}
    </section>
  );
}
