export type UserRole = 'participant' | 'organizer' | 'judge' | 'admin'

export interface UserNotificationSettings {
  email: boolean
  inApp: boolean
  matchChanges: boolean
  applicationUpdates: boolean
  resultConfirmation: boolean
}

export interface UserProfile {
  id: string
  email: string
  nickname: string
  name: string
  avatarUrl?: string
  timezone: string
  roles: UserRole[]
  notificationSettings: UserNotificationSettings
}
