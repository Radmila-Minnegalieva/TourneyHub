import { useState } from "react";
import { allPages, type Schema } from "../api/client";
import { disciplines, formatMap, tournamentView } from "../api/catalog";
import { useResource } from "./useResource";
import type { TournamentFormat, TournamentStatus } from "../types/tournament";
export interface TournamentFilters {
  query: string;
  discipline: string;
  status: "" | TournamentStatus;
  format: "" | TournamentFormat;
}
const initialFilters: TournamentFilters = {
  query: "",
  discipline: "",
  status: "",
  format: "",
};
export function useTournaments() {
  const [filters, setFilters] = useState(initialFilters);
  const ds = useResource(disciplines);
  const resource = useResource(async () => {
    const q = new URLSearchParams();
    if (filters.query) q.set("search", filters.query);
    if (filters.discipline) q.set("discipline_id", filters.discipline);
    if (filters.status)
      q.set(
        "status",
        filters.status === "running" ? "in_progress" : filters.status,
      );
    if (filters.format)
      q.set(
        "format",
        Object.entries(formatMap).find(([, v]) => v === filters.format)![0],
      );
    return allPages<Schema["Tournament"]>(`/tournaments?${q}`, false);
  }, [filters.query, filters.discipline, filters.status, filters.format]);
  return {
    tournaments: (resource.data || []).map((t) =>
      tournamentView(t, ds.data || []),
    ),
    disciplines: ds.data || [],
    filters,
    setFilters,
    reset: () => setFilters(initialFilters),
    loading: resource.loading || ds.loading,
    error: resource.error || ds.error,
    reload: () => {
      resource.reload();
      ds.reload();
    },
  };
}
