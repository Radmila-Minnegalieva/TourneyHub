package model

import "fmt"

// Pairing задаёт номера участников (от нуля), тур и позицию матча.
type Pairing struct {
	Home     int
	Away     int
	Round    int
	Position int
}

// RoundRobin строит круговое расписание методом вращения Бергера.
// Для нечётного n свободный участник пропускает тур; второй круг меняет стороны.
func RoundRobin(n, legs int) ([]Pairing, error) {
	if n < 2 || n > 256 || (legs != 1 && legs != 2) {
		return nil, fmt.Errorf("invalid number of participants or legs")
	}
	size := n + n%2
	players := make([]int, size)
	for i := range players {
		players[i] = i
	}
	pairs := make([]Pairing, 0, n*(n-1)/2*legs)
	for round := 1; round < size; round++ {
		pos := 0
		for i := 0; i < size/2; i++ {
			home, away := players[i], players[size-1-i]
			if home == n || away == n {
				continue
			}
			if (round+i)%2 == 0 {
				home, away = away, home
			}
			pos++
			pairs = append(pairs, Pairing{home, away, round, pos})
		}
		last := players[size-1]
		copy(players[2:], players[1:size-1])
		players[1] = last
	}
	if legs == 2 {
		count := len(pairs)
		for i := 0; i < count; i++ {
			p := pairs[i]
			p.Home, p.Away = p.Away, p.Home
			p.Round += size - 1
			pairs = append(pairs, p)
		}
	}
	return pairs, nil
}

// SeedOrder возвращает классические позиции посева для ближайшей степени двойки.
// Пустые позиции (seed > n) становятся bye, сильнейшие встречаются в финале.
func SeedOrder(n int) ([]int, error) {
	if n < 2 || n > 256 {
		return nil, fmt.Errorf("participants must be between 2 and 256")
	}
	size := 2
	for size < n {
		size *= 2
	}
	seeds := []int{1, 2}
	for len(seeds) < size {
		next := make([]int, 0, len(seeds)*2)
		for _, seed := range seeds {
			next = append(next, seed, len(seeds)*2+1-seed)
		}
		seeds = next
	}
	return seeds, nil
}

// SnakeGroups распределяет отсортированных участников по группам змейкой.
func SnakeGroups(n, count int) ([][]int, error) {
	if count < 2 || count > 64 || n < count*2 || n > 256 {
		return nil, fmt.Errorf("each group needs at least two participants")
	}
	groups := make([][]int, count)
	for i := 0; i < n; i++ {
		g := i % count
		if i/count%2 == 1 {
			g = count - 1 - g
		}
		groups[g] = append(groups[g], i)
	}
	return groups, nil
}

// Game хранит счёт отдельной партии.
type Game struct {
	Number int
	Home   int
	Away   int
}

// ValidateSeries проверяет итог и порядок партий, включая досрочное окончание Bo3/Bo5.
func ValidateSeries(bestOf, home, away int, games []Game, allowsDraw, playoffs bool) error {
	if home < 0 || away < 0 || (bestOf != 1 && bestOf != 3 && bestOf != 5) {
		return fmt.Errorf("invalid score")
	}
	if len(games) == 0 || len(games) > bestOf {
		return fmt.Errorf("invalid number of games")
	}
	for i, g := range games {
		if g.Number != i+1 || g.Home < 0 || g.Away < 0 {
			return fmt.Errorf("game numbers must be consecutive and scores nonnegative")
		}
	}
	if bestOf == 1 {
		if len(games) != 1 || home != games[0].Home || away != games[0].Away {
			return fmt.Errorf("Bo1 score must equal its only game")
		}
		if home == away && (!allowsDraw || playoffs) {
			return fmt.Errorf("draw is not allowed")
		}
		return nil
	}
	return validateBestOf(bestOf, home, away, games)
}

func validateBestOf(bestOf, home, away int, games []Game) error {
	winsHome, winsAway, target := 0, 0, (bestOf+1)/2
	for _, g := range games {
		if winsHome == target || winsAway == target {
			return fmt.Errorf("series continued after a winner was decided")
		}
		if g.Home == g.Away {
			return fmt.Errorf("a series game cannot end in a draw")
		}
		if g.Home > g.Away {
			winsHome++
		} else {
			winsAway++
		}
	}
	if winsHome != home || winsAway != away || (winsHome != target && winsAway != target) {
		return fmt.Errorf("series score must match games and winning threshold")
	}
	return nil
}
