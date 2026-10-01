export interface MatchHistoryItem {
  title: string
  meta: string
  tone: 'neutral' | 'warning'
}

export interface MatchDetails {
  id: string
  tournamentId: string
  tournamentTitle: string
  stageLabel: string
  home: string
  homeShort: string
  homeRating: number
  homeSeed: number
  away: string
  awayShort: string
  awayRating: number
  awaySeed: number
  score: string
  date: string
  venue: string
  statusLabel: string
  periods: Array<{ label: string; home: number; away: number }>
  attachment: string
  history: MatchHistoryItem[]
  ratingDelta: { home: number; away: number }
}
