package unit_test

import (
	"strings"
	"testing"
)

func projectedMonthlyNet(monthlyAdds, monthlySubtracts, transferIn, transferOut, newValue float64, newOp string) float64 {
	net := monthlyAdds - monthlySubtracts + transferIn - transferOut
	if newOp == "subtract" {
		net -= newValue
	} else {
		net += newValue
	}
	return net
}

func TestRepeatableTransactionProjectedBalance(t *testing.T) {
	cases := []struct {
		desc        string
		adds        float64
		subtracts   float64
		transferIn  float64
		transferOut float64
		newValue    float64
		newOp       string
		wantWarning bool
	}{
		{
			desc: "net positive, no warning",
			adds: 3000, subtracts: 1000, transferIn: 0, transferOut: 0,
			newValue: 500, newOp: "subtract", wantWarning: false,
		},
		{
			desc: "net negative, warning",
			adds: 1000, subtracts: 1000, transferIn: 0, transferOut: 0,
			newValue: 100, newOp: "subtract", wantWarning: true,
		},
		{
			desc: "exactly zero, no warning",
			adds: 1000, subtracts: 900, transferIn: 0, transferOut: 0,
			newValue: 100, newOp: "subtract", wantWarning: false,
		},
		{
			desc: "transfer out tips negative, warning",
			adds: 500, subtracts: 200, transferIn: 0, transferOut: 400,
			newValue: 50, newOp: "subtract", wantWarning: true,
		},
		{
			desc: "transfer in saves the day, no warning",
			adds: 500, subtracts: 800, transferIn: 500, transferOut: 0,
			newValue: 100, newOp: "subtract", wantWarning: false,
		},
		{
			desc: "repeatable add, net positive",
			adds: 500, subtracts: 800, transferIn: 0, transferOut: 0,
			newValue: 400, newOp: "add", wantWarning: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			net := projectedMonthlyNet(tc.adds, tc.subtracts, tc.transferIn, tc.transferOut, tc.newValue, tc.newOp)
			gotWarning := net < 0
			if gotWarning != tc.wantWarning {
				t.Errorf("net=%.2f, gotWarning=%v, wantWarning=%v", net, gotWarning, tc.wantWarning)
			}
		})
	}
}

func isReservedName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	return n == "income" || n == "expense"
}

func TestSystemTagNameValidation(t *testing.T) {
	reserved := []string{"income", "expense", "Income", "Expense", " Expense ", " income "}
	for _, name := range reserved {
		if !isReservedName(name) {
			t.Errorf("expected %q to be reserved", name)
		}
	}

	allowed := []string{"salary", "rent", "food", "MyIncome", "myExpense", "groceries"}
	for _, name := range allowed {
		if isReservedName(name) {
			t.Errorf("expected %q to be allowed", name)
		}
	}
}
