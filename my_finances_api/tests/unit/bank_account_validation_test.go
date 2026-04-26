package unit_test

import "testing"

func initialBalanceValid(balance float64) bool {
	return balance >= 0
}

func TestInitialBalanceValidation(t *testing.T) {
	cases := []struct {
		balance float64
		valid   bool
	}{
		{0, true},
		{0.01, true},
		{1000, true},
		{-0.01, false},
		{-100, false},
	}

	for _, tc := range cases {
		got := initialBalanceValid(tc.balance)
		if got != tc.valid {
			t.Errorf("initialBalanceValid(%.2f) = %v, want %v", tc.balance, got, tc.valid)
		}
	}
}
