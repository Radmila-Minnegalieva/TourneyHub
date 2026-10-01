import { Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from '../layout/AppLayout/AppLayout'
import { ForgotPasswordPage } from '../pages/auth/ForgotPasswordPage'
import { LoginPage } from '../pages/auth/LoginPage'
import { RegisterPage } from '../pages/auth/RegisterPage'
import { MatchPage } from '../pages/match/MatchPage'
import { NotFoundPage } from '../pages/not-found/NotFoundPage'
import { ProfilePage } from '../pages/profile/ProfilePage'
import { RatingPage } from '../pages/rating/RatingPage'
import { TeamPage } from '../pages/team/TeamPage'
import { TeamsPage } from '../pages/teams/TeamsPage'
import { TournamentCreatePage } from '../pages/tournament-create/TournamentCreatePage'
import { TournamentPage } from '../pages/tournament/TournamentPage'
import { TournamentsPage } from '../pages/tournaments/TournamentsPage'
import { RequireAuth } from './RequireAuth'

function App() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route index element={<Navigate to="/tournaments" replace />} />
        <Route path="/tournaments" element={<TournamentsPage />} />
        <Route path="/tournaments/new" element={<RequireAuth><TournamentCreatePage /></RequireAuth>} />
        <Route path="/tournaments/:tournamentId/:tab?" element={<TournamentPage />} />
        <Route path="/matches/:matchId" element={<MatchPage />} />
        <Route path="/rating" element={<RatingPage />} />
        <Route path="/teams" element={<RequireAuth><TeamsPage /></RequireAuth>} />
        <Route path="/teams/:teamId" element={<RequireAuth><TeamPage /></RequireAuth>} />
        <Route path="/profile" element={<RequireAuth><ProfilePage /></RequireAuth>} />
        <Route path="/auth/login" element={<LoginPage />} />
        <Route path="/auth/register" element={<RegisterPage />} />
        <Route path="/auth/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}

export default App
