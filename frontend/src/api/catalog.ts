import { allPages, api, type Schema } from "./client";
import type { Tournament } from "../types/tournament";
import { tournamentFormatLabel } from "../lib/tournament";
export const disciplines = () =>
  allPages<Schema["Discipline"]>("/disciplines", false);
export const formatMap = {
  single_elimination: "single-elimination",
  round_robin: "round-robin",
  groups_playoffs: "groups-playoff",
} as const;
export const statusLabels: Record<Schema["TournamentStatus"], string> = {
  draft: "Черновик",
  registration: "Регистрация открыта",
  in_progress: "Идёт",
  completed: "Завершён",
  cancelled: "Отменён",
};
export const matchLabels: Record<Schema["MatchStatus"], string> = {
  scheduled: "Запланирован",
  awaiting_confirmation: "Ожидает подтверждения",
  disputed: "Спор",
  completed: "Завершён",
  bye: "Проход без игры",
};
export const dateLabel = (s?: string) =>
  s ? new Date(s).toLocaleString("ru-RU") : "Не назначено";
export function tournamentView(
  t: Schema["Tournament"],
  ds: Schema["Discipline"][],
): Tournament {
  const format = formatMap[t.format];
  return {
    id: t.id,
    title: t.name,
    discipline:
      ds.find((d) => d.id === t.discipline_id)?.name || "Дисциплина недоступна",
    status: t.status === "in_progress" ? "running" : t.status,
    format,
    formatLabel: tournamentFormatLabel[format],
    organizer: t.organizer.name || t.organizer.nickname,
    dateLabel: dateLabel(t.starts_at),
    participants: t.accepted_count,
    capacity: t.max_participants,
  };
}
export async function teamDetails(id: string) {
  const [team, members] = await Promise.all([
    api<Schema["Team"]>(`/teams/${id}`),
    api<Schema["TeamMember"][]>(`/teams/${id}/members`),
  ]);
  return { ...team, members };
}
