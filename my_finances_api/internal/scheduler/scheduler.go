package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

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

	processRepeatableTransactions(ctx, db, today, now)
	processRepeatableTransfers(ctx, db, today, now)
}

func processRepeatableTransactions(ctx context.Context, db *sql.DB, today int, now time.Time) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, user_id, bank_account_id, name, value, operation, repeatable_day, last_executed_at
		 FROM transactions WHERE is_repeatable=true AND repeatable_day <= $1`, today,
	)
	if err != nil {
		slog.Error("scheduler: querying repeatable transactions", "error", err)
		return
	}
	defer rows.Close()

	type repeatableTxn struct {
		id, userID, accountID, name, operation string
		value                                   float64
		day                                     int
		lastExec                                *time.Time
	}

	var templates []repeatableTxn
	for rows.Next() {
		var rt repeatableTxn
		if err := rows.Scan(&rt.id, &rt.userID, &rt.accountID, &rt.name, &rt.value,
			&rt.operation, &rt.day, &rt.lastExec); err != nil {
			continue
		}
		templates = append(templates, rt)
	}
	rows.Close()

	for _, rt := range templates {
		if rt.lastExec != nil {
			y1, m1, _ := rt.lastExec.Date()
			y2, m2, _ := now.Date()
			if y1 == y2 && m1 == m2 {
				continue
			}
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}

		var newTxnID string
		err = tx.QueryRowContext(ctx,
			`INSERT INTO transactions (user_id, bank_account_id, name, value, operation, date, is_repeatable)
			 VALUES ($1,$2,$3,$4,$5,$6,false)
			 RETURNING id`,
			rt.userID, rt.accountID, rt.name, rt.value, rt.operation, now,
		).Scan(&newTxnID)
		if err != nil {
			tx.Rollback()
			slog.Error("scheduler: inserting transaction", "template_id", rt.id, "error", err)
			continue
		}

		copyTemplateTags(ctx, tx, rt.id, newTxnID, "transaction")

		if rt.operation == "add" {
			tx.ExecContext(ctx,
				`UPDATE bank_accounts SET balance=balance+$1, updated_at=NOW() WHERE id=$2`,
				rt.value, rt.accountID)
		} else {
			tx.ExecContext(ctx,
				`UPDATE bank_accounts SET balance=balance-$1, updated_at=NOW() WHERE id=$2`,
				rt.value, rt.accountID)
		}

		tx.ExecContext(ctx,
			`UPDATE transactions SET last_executed_at=$1 WHERE id=$2`, now, rt.id)

		if err := tx.Commit(); err != nil {
			slog.Error("scheduler: committing repeatable transaction", "id", rt.id, "error", err)
		}
	}
}

func processRepeatableTransfers(ctx context.Context, db *sql.DB, today int, now time.Time) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, user_id, source_account_id, target_account_id, name, value, repeatable_day, last_executed_at
		 FROM transfers WHERE is_repeatable=true AND repeatable_day <= $1`, today,
	)
	if err != nil {
		slog.Error("scheduler: querying repeatable transfers", "error", err)
		return
	}
	defer rows.Close()

	type repeatableTransfer struct {
		id, userID, srcID, tgtID, name string
		value                           float64
		day                             int
		lastExec                        *time.Time
	}

	var templates []repeatableTransfer
	for rows.Next() {
		var rt repeatableTransfer
		if err := rows.Scan(&rt.id, &rt.userID, &rt.srcID, &rt.tgtID, &rt.name,
			&rt.value, &rt.day, &rt.lastExec); err != nil {
			continue
		}
		templates = append(templates, rt)
	}
	rows.Close()

	for _, rt := range templates {
		if rt.lastExec != nil {
			y1, m1, _ := rt.lastExec.Date()
			y2, m2, _ := now.Date()
			if y1 == y2 && m1 == m2 {
				continue
			}
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}

		var newTransferID string
		err = tx.QueryRowContext(ctx,
			`INSERT INTO transfers (user_id, source_account_id, target_account_id, name, value, date, is_repeatable)
			 VALUES ($1,$2,$3,$4,$5,$6,false)
			 RETURNING id`,
			rt.userID, rt.srcID, rt.tgtID, rt.name, rt.value, now,
		).Scan(&newTransferID)
		if err != nil {
			tx.Rollback()
			slog.Error("scheduler: inserting transfer", "template_id", rt.id, "error", err)
			continue
		}

		copyTemplateTags(ctx, tx, rt.id, newTransferID, "transfer")

		tx.ExecContext(ctx,
			`UPDATE bank_accounts SET balance=balance-$1, updated_at=NOW() WHERE id=$2`,
			rt.value, rt.srcID)
		tx.ExecContext(ctx,
			`UPDATE bank_accounts SET balance=balance+$1, updated_at=NOW() WHERE id=$2`,
			rt.value, rt.tgtID)
		tx.ExecContext(ctx,
			`UPDATE transfers SET last_executed_at=$1 WHERE id=$2`, now, rt.id)

		if err := tx.Commit(); err != nil {
			slog.Error("scheduler: committing repeatable transfer", "id", rt.id, "error", err)
		}
	}
}

func copyTemplateTags(ctx context.Context, tx *sql.Tx, templateID, newID, kind string) {
	var srcTable, srcCol, dstTable, dstCol string
	if kind == "transaction" {
		srcTable, srcCol = "transaction_tags", "transaction_id"
		dstTable, dstCol = "transaction_tags", "transaction_id"
	} else {
		srcTable, srcCol = "transfer_tags", "transfer_id"
		dstTable, dstCol = "transfer_tags", "transfer_id"
	}

	tagRows, err := tx.QueryContext(ctx,
		`SELECT tag_id FROM `+srcTable+` WHERE `+srcCol+`=$1`, templateID,
	)
	if err != nil {
		return
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var tagID string
		tagRows.Scan(&tagID)
		tx.ExecContext(ctx,
			`INSERT INTO `+dstTable+` (`+dstCol+`, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			newID, tagID,
		)
	}
}
