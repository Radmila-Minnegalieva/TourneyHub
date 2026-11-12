import { useEffect } from "react";
import { API_BASE } from "../api/client";

export function useTournamentEvents(
  tournamentId: string | undefined,
  reload: () => void,
) {
  useEffect(() => {
    if (!tournamentId) return;
    const stream = new EventSource(
      `${API_BASE}/tournaments/${encodeURIComponent(tournamentId)}/events`,
    );
    let pending: number | undefined;
    const refresh = () => {
      if (pending !== undefined) return;
      pending = window.setTimeout(() => {
        pending = undefined;
        reload();
      }, 50);
    };
    stream.onopen = refresh;
    for (const event of [
      "tournament_updated",
      "match_updated",
      "standings_updated",
      "bracket_updated",
      "resync_required",
    ]) {
      stream.addEventListener(event, refresh);
    }
    return () => {
      stream.close();
      window.clearTimeout(pending);
    };
  }, [tournamentId, reload]);
}
