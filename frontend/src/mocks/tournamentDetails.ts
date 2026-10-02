import type { BracketMatch, ScheduleMatch, TournamentGroup } from '../types/tournament'

export const groups: TournamentGroup[] = [
  {
    id: 'A',
    title: 'Группа A',
    rows: [
      { position: 1, teamId: 'orion', team: 'Орион', shortName: 'ОР', played: 2, wins: 2, draws: 0, losses: 0, goals: '7:2', points: 6 },
      { position: 2, teamId: 'groza', team: 'Гроза', shortName: 'ГР', played: 2, wins: 1, draws: 0, losses: 1, goals: '4:4', points: 3 },
      { position: 3, teamId: 'meteor', team: 'Метеор', shortName: 'МТ', played: 2, wins: 0, draws: 1, losses: 1, goals: '3:5', points: 1 },
      { position: 4, teamId: 'bastion', team: 'Бастион', shortName: 'БС', played: 2, wins: 0, draws: 1, losses: 1, goals: '2:5', points: 1 },
    ],
  },
  {
    id: 'B',
    title: 'Группа B',
    rows: [
      { position: 1, teamId: 'wolves', team: 'Северные волки', shortName: 'СВ', played: 2, wins: 1, draws: 1, losses: 0, goals: '5:3', points: 4 },
      { position: 2, teamId: 'phoenix', team: 'Феникс', shortName: 'ФН', played: 2, wins: 1, draws: 1, losses: 0, goals: '4:3', points: 4 },
      { position: 3, teamId: 'legion', team: 'Легион', shortName: 'ЛГ', played: 2, wins: 1, draws: 0, losses: 1, goals: '3:3', points: 3 },
      { position: 4, teamId: 'vector', team: 'Вектор', shortName: 'ВК', played: 2, wins: 0, draws: 0, losses: 2, goals: '1:4', points: 0 },
    ],
  },
  {
    id: 'C',
    title: 'Группа C',
    rows: [
      { position: 1, teamId: 'foxes', team: 'Кибер лисы', shortName: 'КЛ', played: 2, wins: 2, draws: 0, losses: 0, goals: '5:1', points: 6 },
      { position: 2, teamId: 'alpha', team: 'Альфа', shortName: 'АЛ', played: 2, wins: 1, draws: 0, losses: 1, goals: '3:3', points: 3 },
      { position: 3, teamId: 'titans', team: 'Титаны', shortName: 'ТИ', played: 2, wins: 1, draws: 0, losses: 1, goals: '2:2', points: 3 },
      { position: 4, teamId: 'nightwatch', team: 'Ночной дозор', shortName: 'НД', played: 2, wins: 0, draws: 0, losses: 2, goals: '1:5', points: 0 },
    ],
  },
]

export const schedule: ScheduleMatch[] = [
  { id: 'm1', date: '14 окт, 18:00', group: 'A', home: 'Орион', homeShort: 'ОР', away: 'Бастион', awayShort: 'БС', venue: 'Зал №1', status: 'scheduled' },
  { id: 'm2', date: '14 окт, 19:30', group: 'A', home: 'Гроза', homeShort: 'ГР', away: 'Метеор', awayShort: 'МТ', venue: 'Зал №1', status: 'scheduled' },
  { id: 'semifinal-1', date: '15 окт, 18:00', group: 'B', home: 'Северные волки', homeShort: 'СВ', away: 'Вектор', awayShort: 'ВК', venue: 'Зал №2', status: 'awaiting-confirmation', score: '2 : 0' },
  { id: 'm4', date: '15 окт, 19:30', group: 'B', home: 'Феникс', homeShort: 'ФН', away: 'Легион', awayShort: 'ЛГ', venue: 'Зал №2', status: 'scheduled' },
  { id: 'm5', date: '16 окт, 18:00', group: 'C', home: 'Титаны', homeShort: 'ТИ', away: 'Альфа', awayShort: 'АЛ', venue: 'Зал №1', status: 'scheduled' },
]

export const quarterFinals: BracketMatch[] = [
  { id: 'q1', home: 'Орион', homeShort: 'ОР', away: 'Метеор', awayShort: 'МТ', homeScore: 2, awayScore: 0 },
  { id: 'q2', home: 'Северные волки', homeShort: 'СВ', away: 'Альфа', awayShort: 'АЛ', homeScore: 3, awayScore: 2 },
  { id: 'q3', home: 'Гроза', homeShort: 'ГР', away: 'Феникс', awayShort: 'ФН', homeScore: 1, awayScore: 2 },
  { id: 'q4', home: 'Титаны', homeShort: 'ТИ', away: 'Кибер лисы', awayShort: 'КЛ', homeScore: 0, awayScore: 1 },
]

export const semiFinals: BracketMatch[] = [
  { id: 'semifinal-1', home: 'Орион', homeShort: 'ОР', away: 'Северные волки', awayShort: 'СВ', homeScore: 3, awayScore: 1, state: 'awaiting-confirmation' },
  { id: 'semifinal-2', home: 'Феникс', homeShort: 'ФН', away: 'Кибер лисы', awayShort: 'КЛ' },
]
