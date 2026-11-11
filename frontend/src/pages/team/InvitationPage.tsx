import { useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { api, errorMessage, type Schema } from "../../api/client";
import { Button } from "../../ui/Button/Button";
import { Feedback } from "../../ui/Feedback";
export function InvitationPage() {
  const { invitationId } = useParams();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function accept() {
    setBusy(true);
    setError("");
    try {
      const result = await api<Schema["Invitation"]>(
        `/invitations/${invitationId}/decision`,
        {
          method: "POST",
          body: { decision: "accept", token: params.get("token") || undefined },
        },
      );
      navigate(`/teams/${result.team_id}`);
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <section>
      <h1 className="mb-5 text-3xl font-bold">Приглашение в команду</h1>
      <Feedback error={error} />
      <Button disabled={busy} onClick={() => void accept()}>
        Принять приглашение
      </Button>
    </section>
  );
}
