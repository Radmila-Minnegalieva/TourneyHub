import type { UserProfile } from '../types/user'

export const mockUser: UserProfile = {
  id: 'u-ivan-petrov',
  email: 'ivan.petrov@example.com',
  nickname: 'iv.petrov',
  name: 'Иван Петров',
  timezone: 'Europe/Moscow',
  roles: ['participant', 'organizer'],
  notificationSettings: {
    email: true,
    inApp: true,
    matchChanges: true,
    applicationUpdates: true,
    resultConfirmation: true,
  },
}
