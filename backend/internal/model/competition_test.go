package model

import "testing"

func TestRoundRobin(t *testing.T) {
	for _, n := range []int{2, 3, 4, 5, 16, 255, 256} {
		for _, legs := range []int{1, 2} {
			pairs, err := RoundRobin(n, legs)
			if err != nil {
				t.Fatal(err)
			}
			if len(pairs) != n*(n-1)/2*legs {
				t.Fatalf("n=%d legs=%d got %d", n, legs, len(pairs))
			}
			seen := map[[2]int]int{}
			rounds := map[[2]int]bool{}
			for _, p := range pairs {
				a, b := p.Home, p.Away
				if a > b {
					a, b = b, a
				}
				seen[[2]int{a, b}]++
				for _, player := range []int{p.Home, p.Away} {
					k := [2]int{p.Round, player}
					if rounds[k] {
						t.Fatal("participant plays twice per round")
					}
					rounds[k] = true
				}
			}
			for _, count := range seen {
				if count != legs {
					t.Fatal("pair repeated incorrectly")
				}
			}
		}
	}
}
func TestSeedOrder(t *testing.T) {
	for _, n := range []int{2, 3, 4, 5, 8, 9, 256} {
		seeds, err := SeedOrder(n)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[int]bool{}
		for _, s := range seeds {
			if seen[s] {
				t.Fatal("duplicate seed")
			}
			seen[s] = true
		}
		for i := 0; i < len(seeds); i += 2 {
			if seeds[i]+seeds[i+1] != len(seeds)+1 {
				t.Fatal("wrong seeded pairing")
			}
		}
		if seeds[0] != 1 || seeds[len(seeds)/2] != 2 {
			t.Fatal("strongest seeds are not separated")
		}
	}
}
func TestSnakeGroups(t *testing.T) {
	got, err := SnakeGroups(8, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{0, 3, 4, 7}, {1, 2, 5, 6}}
	for g := range want {
		for i := range want[g] {
			if got[g][i] != want[g][i] {
				t.Fatal(got)
			}
		}
	}
	if _, err = SnakeGroups(3, 2); err == nil {
		t.Fatal("accepted undersized groups")
	}
}
func TestValidateSeries(t *testing.T) {
	cases := []struct {
		name    string
		bo      int
		h       int
		a       int
		games   []Game
		draw    bool
		playoff bool
		valid   bool
	}{
		{"Bo1", 1, 21, 19, []Game{{1, 21, 19}}, false, false, true},
		{"draw", 1, 2, 2, []Game{{1, 2, 2}}, true, false, true},
		{"playoff draw", 1, 2, 2, []Game{{1, 2, 2}}, true, true, false},
		{"Bo3", 3, 2, 1, []Game{{1, 1, 0}, {2, 0, 1}, {3, 1, 0}}, false, true, true},
		{"incomplete", 3, 1, 1, []Game{{1, 1, 0}, {2, 0, 1}}, false, false, false},
		{"extra game", 3, 2, 1, []Game{{1, 1, 0}, {2, 1, 0}, {3, 0, 1}}, false, false, false},
		{"number", 1, 1, 0, []Game{{2, 1, 0}}, false, false, false},
		{"mismatched", 1, 1, 0, []Game{{1, 2, 0}}, false, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateSeries(c.bo, c.h, c.a, c.games, c.draw, c.playoff)
			if (err == nil) != c.valid {
				t.Fatal(err)
			}
		})
	}
}
