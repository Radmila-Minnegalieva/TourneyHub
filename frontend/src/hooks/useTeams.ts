import { useMemo, useState } from 'react'
import { mockTeams } from '../mocks/teams'
import type { Team } from '../types/team'

export function useTeams() {
  const [teams, setTeams] = useState<Team[]>(mockTeams)

  const createTeam = (payload: Pick<Team, 'name' | 'shortName' | 'discipline'>) => {
    const team: Team = {
      id: `${payload.shortName.toLowerCase()}-${Date.now()}`,
      ...payload,
      rating: 1500,
      members: [{ id: 'u-ivan-petrov', nickname: 'iv.petrov', name: 'Иван Петров', role: 'captain' }],
      pendingInvites: [],
    }
    setTeams((current) => [team, ...current])
    return team
  }

  return { teams, createTeam }
}

export function useTeam(teamId?: string) {
  const initial = useMemo(() => mockTeams.find((team) => team.id === teamId) ?? mockTeams[0], [teamId])
  const [team, setTeam] = useState<Team>(initial)

  const inviteByNickname = (nickname: string) => {
    if (!nickname.trim()) return
    setTeam((current) => ({ ...current, pendingInvites: [...current.pendingInvites, nickname.trim()] }))
  }

  const removeMember = (memberId: string) => {
    setTeam((current) => ({ ...current, members: current.members.filter((member) => member.id !== memberId) }))
  }

  const transferCaptain = (memberId: string) => {
    setTeam((current) => ({
      ...current,
      members: current.members.map((member) => ({
        ...member,
        role: member.id === memberId ? 'captain' : member.role === 'captain' ? 'player' : member.role,
      })),
    }))
  }

  return { team, inviteByNickname, removeMember, transferCaptain }
}
