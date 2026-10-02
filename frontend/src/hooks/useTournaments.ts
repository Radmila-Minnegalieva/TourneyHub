import { useMemo, useState } from 'react'
import { tournaments as source } from '../mocks/tournaments'
import type { TournamentFormat, TournamentStatus } from '../types/tournament'

export interface TournamentFilters {
  query: string
  discipline: string
  status: '' | TournamentStatus
  format: '' | TournamentFormat
}

const initialFilters: TournamentFilters = { query: '', discipline: '', status: '', format: '' }

export function useTournaments() {
  const [filters, setFilters] = useState(initialFilters)

  const tournaments = useMemo(() => {
    const query = filters.query.trim().toLowerCase()
    return source.filter((item) => {
      const matchesQuery = !query || `${item.title} ${item.organizer}`.toLowerCase().includes(query)
      const matchesDiscipline = !filters.discipline || item.discipline === filters.discipline
      const matchesStatus = !filters.status || item.status === filters.status
      const matchesFormat = !filters.format || item.format === filters.format
      return matchesQuery && matchesDiscipline && matchesStatus && matchesFormat
    })
  }, [filters])

  const reset = () => setFilters(initialFilters)

  return { tournaments, filters, setFilters, reset }
}
