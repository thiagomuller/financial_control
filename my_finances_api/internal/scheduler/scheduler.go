package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// Run starts a background goroutine that processes recurring incomes and expenses daily.
func Run(db *sql.DB) {
	go func() {
		tick := time.NewTicker(24 * time.Hour)
		defer tick.Stop()

		processAll(db)

		for range tick.C {
			processAll(db)
		}
	}()
}

func processAll(db *sql.DB) {
	ctx := context.Background()
	now := time.Now()
	today := now.Day()

	processIncomes(ctx, db, today, now)
	processExpenses(ctx, db, today, now)
}

func processIncomes(ctx context.Context, db *sql.DB, today int, now time.Time) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, user_id, bank_account_id, name, value, repeatable_day, last_executed_at
		 FROM incomes WHERE repeatable_day <= $1`, today,
	)
	if err != nil {
		slog.Error("scheduler: querying incomes", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID, accountID, name string
		var value float64
		var day int
		var lastExec *time.Time

		if err := rows.Scan(&id, &userID, &accountID, &name, &value, &day, &lastExec); err != nil {
			continue
		}

		// Only execute once per month
		if lastExec != nil {
			y1, m1, _ := lastExec.Date()
			y2, m2, _ := now.Date()
			if y1 == y2 && m1 == m2 {
				continue
			}
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}

		tx.ExecContext(ctx,
			`INSERT INTO transactions (user_id, bank_account_id, name, value, operation, date)
			 VALUES ($1,$2,$3,$4,'add',$5)`,
			userID, accountID, name, value, now,
		)
		tx.ExecContext(ctx,
			`UPDATE bank_accounts SET balance=balance+$1, updated_at=NOW() WHERE id=$2`,
			value, accountID,
		)
		tx.ExecContext(ctx,
			`UPDATE incomes SET last_executed_at=$1 WHERE id=$2`, now, id,
		)

		if err := tx.Commit(); err != nil {
			slog.Error("scheduler: committing income transaction", "id", id, "error", err)
		}
	}
}

func processExpenses(ctx context.Context, db *sql.DB, today int, now time.Time) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, user_id, bank_account_id, name, value, repeatable_day, last_executed_at
		 FROM expenses WHERE repeatable_day <= $1`, today,
	)
	if err != nil {
		slog.Error("scheduler: querying expenses", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID, accountID, name string
		var value float64
		var day int
		var lastExec *time.Time

		if err := rows.Scan(&id, &userID, &accountID, &name, &value, &day, &lastExec); err != nil {
			continue
		}

		if lastExec != nil {
			y1, m1, _ := lastExec.Date()
			y2, m2, _ := now.Date()
			if y1 == y2 && m1 == m2 {
				continue
			}
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}

		tx.ExecContext(ctx,
			`INSERT INTO transactions (user_id, bank_account_id, name, value, operation, date)
			 VALUES ($1,$2,$3,$4,'subtract',$5)`,
			userID, accountID, name, value, now,
		)
		tx.ExecContext(ctx,
			`UPDATE bank_accounts SET balance=balance-$1, updated_at=NOW() WHERE id=$2`,
			value, accountID,
		)
		tx.ExecContext(ctx,
			`UPDATE expenses SET last_executed_at=$1 WHERE id=$2`, now, id,
		)

		if err := tx.Commit(); err != nil {
			slog.Error("scheduler: committing expense transaction", "id", id, "error", err)
		}
	}
}
