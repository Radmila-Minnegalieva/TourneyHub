import { NavLink } from 'react-router-dom'
import { ROUTES } from '../../constants/routes'

export function NotFoundPage() {
  return <div className="py-24 text-center"><h1 className="text-4xl font-extrabold">Страница не найдена</h1><NavLink className="mt-4 inline-block font-semibold text-brand" to={ROUTES.tournaments}>Вернуться к турнирам</NavLink></div>
}
