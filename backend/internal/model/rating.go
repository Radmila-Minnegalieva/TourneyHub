// rating.go реализует рейтинг Эло и его изменение после подтверждённого матча.
package model

import (
	"errors"
	"math"
)

// InitialRating — начальный рейтинг команды до первого матча.
const InitialRating = 1500.0

// RatingK — коэффициент изменения рейтинга после матча.
const RatingK = 32.0

// EloRating — рейтинг команды в одной дисциплине.
//
// Начальное значение равно InitialRating. Каждый подтверждённый матч меняет
// рейтинг с коэффициентом RatingK. В БД значение хранится без округления;
// интерфейс может округлить его при показе. Проход без соперника рейтинг не меняет.
type EloRating struct {
	// TeamID — идентификатор команды, для которой рассчитывается рейтинг.
	TeamID string
	// DisciplineID — идентификатор дисциплины, в которой учитывается рейтинг.
	DisciplineID string
	// Value — текущее значение Эло без округления.
	Value float64
}

// MatchRating описывает рейтинг после подтверждённого матча.
//
// Встраивание EloRating позволяет использовать идентификаторы команды,
// дисциплины и итоговое значение. В Go это композиция типов.
type MatchRating struct {
	// EloRating содержит итоговый рейтинг команды и её идентификаторы.
	EloRating
	// Before — значение рейтинга до матча.
	Before float64
	// Delta — изменение со знаком: Value минус Before.
	Delta float64
}

// NewEloRating возвращает начальный рейтинг команды в дисциплине.
//
// teamID и disciplineID — идентификаторы команды и дисциплины.
// Возвращённый EloRating имеет Value равное 1500. Функция не записывает данные в БД.
func NewEloRating(teamID, disciplineID string) EloRating {
	return EloRating{TeamID: teamID, DisciplineID: disciplineID, Value: InitialRating}
}

// AfterMatch рассчитывает новый рейтинг, сохраняя исходное значение.
//
// opponent — рейтинг соперника до матча без округления; оба рейтинга должны
// быть конечными числами. score равен 1 при победе, 0.5 при ничьей и 0 при поражении.
// Результат MatchRating содержит исходное значение, изменение со знаком и итог.
// При недопустимом score или бесконечном рейтинге возвращаются нулевой MatchRating
// и ошибка. NaN также считается недопустимым значением.
//
// Вызывающий код рассчитывает обоих соперников по исходным рейтингам и сохраняет
// результаты в одной транзакции с подтверждением матча. Изменения равны по модулю
// и противоположны по знаку. Допустимость ничьей проверяется регламентом турнира.
func (r EloRating) AfterMatch(opponent, score float64) (MatchRating, error) {
	if math.IsNaN(r.Value) || math.IsInf(r.Value, 0) || math.IsNaN(opponent) || math.IsInf(opponent, 0) {
		return MatchRating{}, errors.New("ratings must be finite")
	}
	if score != 0 && score != 0.5 && score != 1 {
		return MatchRating{}, errors.New("score must be 0, 0.5 or 1")
	}
	expected := 1 / (1 + math.Pow(10, (opponent-r.Value)/400))
	delta := RatingK * (score - expected)
	result := r
	result.Value += delta
	return MatchRating{EloRating: result, Before: r.Value, Delta: delta}, nil
}
