export type TeamMemberRole = 'captain' | 'player'

export interface TeamMember {
  id: string
  nickname: string
  name: string
  role: TeamMemberRole
}

export interface Team {
  id: string
  name: string
  shortName: string
  discipline: string
  rating: number
  members: TeamMember[]
  pendingInvites: string[]
}
