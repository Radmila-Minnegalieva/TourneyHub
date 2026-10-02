export const ROUTES = {
  tournaments: '/tournaments',
  createTournament: '/tournaments/new',
  rating: '/rating',
  tournament: (id: string, tab = 'groups') => `/tournaments/${id}/${tab}`,
  match: (id: string) => `/matches/${id}`,
  login: '/auth/login',
  register: '/auth/register',
  forgotPassword: '/auth/forgot-password',
  profile: '/profile',
  teams: '/teams',
  team: (id: string) => `/teams/${id}`,
} as const
