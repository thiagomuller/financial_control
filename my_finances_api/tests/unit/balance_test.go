package unit_test

import (
	"testing"
)

// transferAllowed checks whether a transfer of `amount` is permitted given `balance`.
func transferAllowed(balance, amount float64) bool {
	return balance >= amount
}

// applyOperation returns the new balance after applying a transaction.
func applyOperation(balance, value float64, op string) float64 {
	if op == "add" {
		return balance + value
	}
	return balance - value
}

func TestTransferAllowed(t *testing.T) {
	cases := []struct {
		balance float64
		amount  float64
		allowed bool
	}{
		{1000, 500, true},
		{500, 500, true},
		{499.99, 500, false},
		{0, 1, false},
	}

	for _, tc := range cases {
		got := transferAllowed(tc.balance, tc.amount)
		if got != tc.allowed {
			t.Errorf("transferAllowed(%.2f, %.2f) = %v, want %v",
				tc.balance, tc.amount, got, tc.allowed)
		}
	}
}

// transferOperation returns the operation for a feed entry given which account is being viewed.
func transferOperation(sourceAccountID, viewedAccountID string) string {
	if sourceAccountID == viewedAccountID {
		return "subtract"
	}
	return "add"
}

func TestTransferOperation(t *testing.T) {
	cases := []struct {
		sourceID string
		viewedID string
		want     string
	}{
		{"acc1", "acc1", "subtract"},
		{"acc1", "acc2", "add"},
		{"acc2", "acc2", "subtract"},
		{"acc2", "acc1", "add"},
	}

	for _, tc := range cases {
		got := transferOperation(tc.sourceID, tc.viewedID)
		if got != tc.want {
			t.Errorf("transferOperation(%q, %q) = %q, want %q",
				tc.sourceID, tc.viewedID, got, tc.want)
		}
	}
}

func TestApplyOperation(t *testing.T) {
	cases := []struct {
		balance float64
		value   float64
		op      string
		want    float64
	}{
		{1000, 200, "add", 1200},
		{1000, 200, "subtract", 800},
		{0, 100, "add", 100},
		{50, 50, "subtract", 0},
	}

	for _, tc := range cases {
		got := applyOperation(tc.balance, tc.value, tc.op)
		if got != tc.want {
			t.Errorf("applyOperation(%.2f, %.2f, %q) = %.2f, want %.2f",
				tc.balance, tc.value, tc.op, got, tc.want)
		}
	}
}
