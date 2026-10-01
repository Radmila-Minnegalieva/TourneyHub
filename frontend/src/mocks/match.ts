import type { MatchDetails } from '../types/match'

export const matchDetails: MatchDetails = {
  id: 'semifinal-1',
  tournamentId: 'autumn-cup',
  tournamentTitle: 'Осенний кубок по мини-футболу 2026',
  stageLabel: 'Плей-офф / Полуфинал 1',
  home: 'Орион',
  homeShort: 'ОР',
  homeRating: 1612,
  homeSeed: 2,
  away: 'Северные волки',
  awayShort: 'СВ',
  awayRating: 1587,
  awaySeed: 3,
  score: '3 : 1',
  date: '14 ноября, 19:00',
  venue: 'Зал №2',
  statusLabel: 'Ожидает подтверждения',
  periods: [
    { label: '1-й тайм', home: 1, away: 1 },
    { label: '2-й тайм', home: 2, away: 0 },
  ],
  attachment: 'protocol_sf1.pdf · 212 КБ',
  history: [
    { title: 'Матч назначен на 14.11, 19:00', meta: 'организатор · 02.11', tone: 'neutral' },
    { title: 'Назначен судья: Смирнов А.', meta: 'организатор · 05.11', tone: 'neutral' },
    { title: 'Результат 3 : 1 сообщён', meta: 'капитан «Орион» · 14.11, 21:07', tone: 'warning' },
  ],
  ratingDelta: { home: 14, away: -14 },
}
