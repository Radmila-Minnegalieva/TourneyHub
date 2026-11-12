import { useState } from "react";
import { allPages, type Schema } from "../api/client";
import {
  disciplines,
  formatMap,
  registrationOpen,
  tournamentView,
} from "../api/catalog";
import { useResource } from "./useResource";
import type { TournamentFormat, TournamentStatus } from "../types/tournament";
export interface TournamentFilters {
  query: string;
  discipline: string;
  status: "" | TournamentStatus;
  format: "" | TournamentFormat;
  startsFrom: string;
  startsTo: string;
}
const initialFilters: TournamentFilters = {
  query: "",
  discipline: "",
  status: "",
  format: "",
  startsFrom: "",
  startsTo: "",
};
export function useTournaments() {
  const [filters, setFilters] = useState(initialFilters);
  const ds = useResource(disciplines);
  const resource = useResource(async () => {
    const q = new URLSearchParams();
    if (filters.query) q.set("search", filters.query);
    if (filters.discipline) q.set("discipline_id", filters.discipline);
    if (
      filters.startsFrom &&
      filters.startsTo &&
      filters.startsFrom > filters.startsTo
    )
      throw new Error("Дата окончания диапазона раньше даты начала.");
    if (filters.startsFrom)
      q.set(
        "starts_from",
        new Date(`${filters.startsFrom}T00:00:00`).toISOString(),
      );
    if (filters.startsTo)
      q.set(
        "starts_to",
        new Date(`${filters.startsTo}T23:59:59.999`).toISOString(),
      );
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
  }, [
    filters.query,
    filters.discipline,
    filters.status,
    filters.format,
    filters.startsFrom,
    filters.startsTo,
  ]);
  return {
    tournaments: (resource.data || [])
      .filter((t) => filters.status !== "registration" || registrationOpen(t))
      .map((t) => tournamentView(t, ds.data || [])),
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
