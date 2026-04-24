package handlers

import (
	"net/http"
	"time"

	"github.com/thiago/my_finances_api/internal/middleware"
	"github.com/thiago/my_finances_api/internal/models"
)

func (h *Handler) getDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	// Load all bank accounts for the user
	accRows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, name, balance, icon_url, created_at, updated_at
		 FROM bank_accounts WHERE user_id=$1 ORDER BY name`, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer accRows.Close()

	var accounts []models.BankAccount
	for accRows.Next() {
		var a models.BankAccount
		accRows.Scan(&a.ID, &a.UserID, &a.Name, &a.Balance, &a.IconURL, &a.CreatedAt, &a.UpdatedAt)
		accounts = append(accounts, a)
	}

	summaries := make([]models.BankAccountSummary, 0, len(accounts))
	now := time.Now()
	horizon := now.AddDate(0, 1, 0) // next month for upcoming expenses

	for _, account := range accounts {
		// Latest 3 transactions
		txnRows, err := h.db.QueryContext(r.Context(),
			`SELECT id, user_id, bank_account_id, name, value, operation, date, created_at, updated_at
			 FROM transactions WHERE bank_account_id=$1 AND user_id=$2
			 ORDER BY date DESC LIMIT 3`,
			account.ID, userID,
		)
		var txns []models.Transaction
		if err == nil {
			defer txnRows.Close()
			for txnRows.Next() {
				var t models.Transaction
				txnRows.Scan(&t.ID, &t.UserID, &t.BankAccountID, &t.Name, &t.Value,
					&t.Operation, &t.Date, &t.CreatedAt, &t.UpdatedAt)
				t.Tags = h.fetchTransactionTags(r, t.ID)
				txns = append(txns, t)
			}
		}

		// Tag stats — count transactions per tag for this account
		tagRows, _ := h.db.QueryContext(r.Context(),
			`SELECT t.id, t.user_id, t.name, t.color, t.created_at, t.updated_at, COUNT(tt.tag_id) AS cnt
			 FROM tags t
			 JOIN transaction_tags tt ON tt.tag_id = t.id
			 JOIN transactions tx ON tx.id = tt.transaction_id
			 WHERE tx.bank_account_id=$1 AND tx.user_id=$2
			 GROUP BY t.id ORDER BY cnt DESC LIMIT 10`,
			account.ID, userID,
		)
		var tagStats []models.TagStat
		if tagRows != nil {
			defer tagRows.Close()
			for tagRows.Next() {
				var ts models.TagStat
				tagRows.Scan(&ts.Tag.ID, &ts.Tag.UserID, &ts.Tag.Name, &ts.Tag.Color,
					&ts.Tag.CreatedAt, &ts.Tag.UpdatedAt, &ts.Count)
				tagStats = append(tagStats, ts)
			}
		}

		// Upcoming expenses for this account (next 30 days)
		var upcoming []models.UpcomingItem
		expRows, _ := h.db.QueryContext(r.Context(),
			`SELECT name, value, repeatable_day FROM expenses
			 WHERE bank_account_id=$1 AND user_id=$2`, account.ID, userID,
		)
		if expRows != nil {
			defer expRows.Close()
			for expRows.Next() {
				var name string
				var value float64
				var day int
				expRows.Scan(&name, &value, &day)
				for _, d := range nextOccurrences(day, now, horizon) {
					upcoming = append(upcoming, models.UpcomingItem{
						Type: "expense", Name: name, Value: value,
						Operation: "subtract", Date: d, Source: name,
					})
				}
			}
		}

		summaries = append(summaries, models.BankAccountSummary{
			Account:            account,
			LatestTransactions: txns,
			TagStats:           tagStats,
			UpcomingExpenses:   upcoming,
		})
	}

	writeJSON(w, http.StatusOK, summaries)
}
