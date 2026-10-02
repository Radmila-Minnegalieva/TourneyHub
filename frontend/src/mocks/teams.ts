import type { Team } from '../types/team'

export const mockTeams: Team[] = [
  {
    id: 'northern-wolves',
    name: 'Северные волки',
    shortName: 'СВ',
    discipline: 'Мини-футбол',
    rating: 1587,
    members: [
      { id: 'u-ivan-petrov', nickname: 'iv.petrov', name: 'Иван Петров', role: 'captain' },
      { id: 'u-alex-smirnov', nickname: 'asmirnov', name: 'Алексей Смирнов', role: 'player' },
      { id: 'u-max-kim', nickname: 'max.kim', name: 'Максим Ким', role: 'player' },
      { id: 'u-pavel-orlov', nickname: 'p.orlov', name: 'Павел Орлов', role: 'player' },
    ],
    pendingInvites: ['neo.21'],
  },
  {
    id: 'night-owls',
    name: 'Ночные совы',
    shortName: 'НС',
    discipline: 'Интеллектуальные игры',
    rating: 1516,
    members: [
      { id: 'u-ivan-petrov', nickname: 'iv.petrov', name: 'Иван Петров', role: 'player' },
      { id: 'u-elena-fox', nickname: 'e.fox', name: 'Елена Фокс', role: 'captain' },
    ],
    pendingInvites: [],
  },
]
