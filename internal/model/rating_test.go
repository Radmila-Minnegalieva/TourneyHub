package model_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/Radmila-Minnegalieva/TourneyHub/internal/model"
)

func ExampleEloRating_AfterMatch() {
	rating := model.NewEloRating("team-a", "chess")
	result, err := rating.AfterMatch(1500, 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("before=%.0f delta=%+.0f after=%.0f\n", result.Before, result.Delta, result.Value)
	// Output: before=1500 delta=+16 after=1516
}

func TestEloConservesTotal(t *testing.T) {
	for _, tc := range []struct{ a, b, score float64 }{{1500, 1500, 1}, {1500, 1500, 0.5}, {1800, 1200, 0}, {1000, 2000, 1}} {
		a := model.EloRating{Value: tc.a}
		b := model.EloRating{Value: tc.b}
		first, err := a.AfterMatch(tc.b, tc.score)
		if err != nil {
			t.Fatal(err)
		}
		second, err := b.AfterMatch(tc.a, 1-tc.score)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(first.Delta+second.Delta) > 1e-10 {
			t.Fatalf("nonzero total delta: %+v %+v", first, second)
		}
		if a.Value != tc.a || b.Value != tc.b {
			t.Fatal("calculation mutated input")
		}
	}
}

func TestEloRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct{ own, opponent, score float64 }{{1500, 1500, 0.3}, {1500, 1500, math.NaN()}, {math.Inf(1), 1500, 1}, {1500, math.NaN(), 1}} {
		if _, err := (model.EloRating{Value: tc.own}).AfterMatch(tc.opponent, tc.score); err == nil {
			t.Fatalf("accepted invalid input: %+v", tc)
		}
	}
}
