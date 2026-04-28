package database

import "database/sql"

const schema = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    username    VARCHAR(100) NOT NULL UNIQUE,
    email       VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bank_accounts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    balance    NUMERIC(18,2) NOT NULL DEFAULT 0,
    icon_url   TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tags (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    color      VARCHAR(7),
    is_system  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_system_name ON tags(name) WHERE is_system=true;

CREATE TABLE IF NOT EXISTS transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    value           NUMERIC(18,2) NOT NULL,
    operation       VARCHAR(10) NOT NULL CHECK (operation IN ('add', 'subtract')),
    date            TIMESTAMPTZ NOT NULL,
    is_repeatable   BOOLEAN NOT NULL DEFAULT FALSE,
    repeatable_day  INT CHECK (repeatable_day >= 1 AND repeatable_day <= 31),
    last_executed_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transaction_tags (
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag_id         UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
);

CREATE TABLE IF NOT EXISTS transfers (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_account_id UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    target_account_id UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    name              VARCHAR(255) NOT NULL,
    value             NUMERIC(18,2) NOT NULL,
    date              TIMESTAMPTZ NOT NULL,
    is_repeatable     BOOLEAN NOT NULL DEFAULT FALSE,
    repeatable_day    INT CHECK (repeatable_day >= 1 AND repeatable_day <= 31),
    last_executed_at  TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transfer_tags (
    transfer_id UUID NOT NULL REFERENCES transfers(id) ON DELETE CASCADE,
    tag_id      UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (transfer_id, tag_id)
);

CREATE TABLE IF NOT EXISTS goals (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_account_id UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    target_account_id UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    name              VARCHAR(255) NOT NULL,
    start_date        TIMESTAMPTZ NOT NULL,
    end_date          TIMESTAMPTZ NOT NULL,
    interval_days     INT NOT NULL DEFAULT 30,
    target_value      NUMERIC(18,2) NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const alterations = `
ALTER TABLE bank_accounts ALTER COLUMN icon_url TYPE TEXT;
DO $$ BEGIN ALTER TABLE tags ALTER COLUMN user_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
ALTER TABLE tags ADD COLUMN IF NOT EXISTS is_system BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS is_repeatable BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS repeatable_day INT CHECK (repeatable_day >= 1 AND repeatable_day <= 31);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS last_executed_at TIMESTAMPTZ;
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS is_repeatable BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS repeatable_day INT CHECK (repeatable_day >= 1 AND repeatable_day <= 31);
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS last_executed_at TIMESTAMPTZ;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS incomes;
`

func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	if _, err := db.Exec(alterations); err != nil {
		return err
	}
	return nil
}
