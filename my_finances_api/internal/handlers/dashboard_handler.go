package handlers

import (
	"net/http"
	"time"

	"github.com/thiago/my_finances_api/internal/middleware"
	"github.com/thiago/my_finances_api/internal/models"
)

func (h *Handler) getDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

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
	horizon := now.AddDate(0, 1, 0)

	for _, account := range accounts {
		feedRows, err := h.db.QueryContext(r.Context(),
			`SELECT id, name, value, operation, date, 'transaction' AS kind
			   FROM transactions WHERE bank_account_id=$1 AND user_id=$2
			 UNION ALL
			 SELECT id, name, value,
			   CASE WHEN source_account_id=$1 THEN 'subtract' ELSE 'add' END,
			   date, 'transfer' AS kind
			   FROM transfers WHERE (source_account_id=$1 OR target_account_id=$1) AND user_id=$2
			 ORDER BY date DESC LIMIT 3`,
			account.ID, userID,
		)
		var feed []models.FeedEntry
		if err == nil {
			defer feedRows.Close()
			for feedRows.Next() {
				var f models.FeedEntry
				feedRows.Scan(&f.ID, &f.Name, &f.Value, &f.Operation, &f.Date, &f.Kind)
				if f.Kind == "transaction" {
					f.Tags = h.fetchTransactionTags(r, f.ID)
				}
				feed = append(feed, f)
			}
		}

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
			LatestTransactions: feed,
			TagStats:           tagStats,
			UpcomingExpenses:   upcoming,
		})
	}

	writeJSON(w, http.StatusOK, summaries)
}
