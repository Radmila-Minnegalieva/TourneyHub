import type { TournamentFormat, TournamentStatus } from '../types/tournament'

export const tournamentStatusLabel: Record<TournamentStatus, string> = {
  registration: 'Регистрация открыта',
  running: 'Идёт',
  completed: 'Завершён',
}

export const tournamentFormatLabel: Record<TournamentFormat, string> = {
  'single-elimination': 'Олимпийская система',
  'round-robin': 'Круговая система',
  'groups-playoff': 'Группы + плей-офф',
}

export function getFreePlaces(participants: number, capacity: number) {
  return Math.max(capacity - participants, 0)
}
