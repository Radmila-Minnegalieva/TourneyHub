export type TournamentStatus = 'registration' | 'running' | 'completed'
export type TournamentFormat = 'single-elimination' | 'round-robin' | 'groups-playoff'

export interface Tournament {
  id: string
  title: string
  discipline: string
  status: TournamentStatus
  format: TournamentFormat
  formatLabel: string
  organizer: string
  dateLabel: string
  participants: number
  capacity: number
}

export interface StandingRow {
  position: number
  teamId: string
  team: string
  shortName: string
  played: number
  wins: number
  draws: number
  losses: number
  goals: string
  points: number
}

export interface TournamentGroup {
  id: string
  title: string
  rows: StandingRow[]
}

export interface ScheduleMatch {
  id: string
  date: string
  group: string
  home: string
  homeShort: string
  away: string
  awayShort: string
  venue: string
  status: 'scheduled' | 'awaiting-confirmation' | 'completed'
  score?: string
}

export interface BracketMatch {
  id: string
  home: string
  homeShort: string
  away: string
  awayShort: string
  homeScore?: number
  awayScore?: number
  state?: 'awaiting-confirmation'
}
