package models

import "time"

// ── User ─────────────────────────────────────────────────────────────────────

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ── BankAccount ───────────────────────────────────────────────────────────────

type BankAccount struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Balance   float64   `json:"balance"`
	IconURL   *string   `json:"icon_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateBankAccountRequest struct {
	Name           string  `json:"name"`
	InitialBalance float64 `json:"initial_balance"`
	IconURL        *string `json:"icon_url"`
}

type UpdateBankAccountRequest struct {
	Name    string  `json:"name"`
	IconURL *string `json:"icon_url"`
}

// ── Tag ───────────────────────────────────────────────────────────────────────

type Tag struct {
	ID        string    `json:"id"`
	UserID    *string   `json:"user_id"`
	Name      string    `json:"name"`
	Color     *string   `json:"color"`
	IsSystem  bool      `json:"is_system"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTagRequest struct {
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

type UpdateTagRequest struct {
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

// ── Transaction ───────────────────────────────────────────────────────────────

type Transaction struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	BankAccountID  string     `json:"bank_account_id"`
	Name           string     `json:"name"`
	Value          float64    `json:"value"`
	Operation      string     `json:"operation"` // "add" | "subtract"
	Date           time.Time  `json:"date"`
	IsRepeatable   bool       `json:"is_repeatable"`
	RepeatableDay  *int       `json:"repeatable_day"`
	LastExecutedAt *time.Time `json:"last_executed_at,omitempty"`
	Tags           []Tag      `json:"tags,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateTransactionRequest struct {
	Name          string    `json:"name"`
	Value         float64   `json:"value"`
	Operation     string    `json:"operation"`
	BankAccountID string    `json:"bank_account_id"`
	Date          time.Time `json:"date"`
	TagIDs        []string  `json:"tag_ids"`
	IsRepeatable  bool      `json:"is_repeatable"`
	RepeatableDay *int      `json:"repeatable_day"`
}

type CreateTransactionResponse struct {
	Transaction
	ProjectedBalanceWarning string `json:"projected_balance_warning,omitempty"`
}

type UpdateTransactionRequest struct {
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Operation string    `json:"operation"`
	Date      time.Time `json:"date"`
	TagIDs    []string  `json:"tag_ids"`
}

type PaginatedResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Pages int `json:"pages"`
}

// ── Transfer ──────────────────────────────────────────────────────────────────

type Transfer struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	SourceAccountID string     `json:"source_account_id"`
	TargetAccountID string     `json:"target_account_id"`
	Name            string     `json:"name"`
	Value           float64    `json:"value"`
	Date            time.Time  `json:"date"`
	IsRepeatable    bool       `json:"is_repeatable"`
	RepeatableDay   *int       `json:"repeatable_day"`
	LastExecutedAt  *time.Time `json:"last_executed_at,omitempty"`
	Tags            []Tag      `json:"tags,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateTransferRequest struct {
	Name            string    `json:"name"`
	Value           float64   `json:"value"`
	SourceAccountID string    `json:"source_account_id"`
	TargetAccountID string    `json:"target_account_id"`
	Date            time.Time `json:"date"`
	TagIDs          []string  `json:"tag_ids"`
	IsRepeatable    bool      `json:"is_repeatable"`
	RepeatableDay   *int      `json:"repeatable_day"`
}

type CreateTransferResponse struct {
	Transfer
	ProjectedBalanceWarning string `json:"projected_balance_warning,omitempty"`
}

type UpdateTransferRequest struct {
	Name   string    `json:"name"`
	Date   time.Time `json:"date"`
	TagIDs []string  `json:"tag_ids"`
}

// ── Goal ──────────────────────────────────────────────────────────────────────

type Goal struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	SourceAccountID string    `json:"source_account_id"`
	TargetAccountID string    `json:"target_account_id"`
	Name            string    `json:"name"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	IntervalDays    int       `json:"interval_days"`
	TargetValue     float64   `json:"target_value"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateGoalRequest struct {
	Name            string    `json:"name"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	IntervalDays    int       `json:"interval_days"`
	SourceAccountID string    `json:"source_account_id"`
	TargetAccountID string    `json:"target_account_id"`
	TargetValue     float64   `json:"target_value"`
}

type UpdateGoalRequest struct {
	Name         string    `json:"name"`
	EndDate      time.Time `json:"end_date"`
	IntervalDays int       `json:"interval_days"`
	TargetValue  float64   `json:"target_value"`
}

// ── Dashboard / Statement ────────────────────────────────────────────────────

type FeedEntry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Operation string    `json:"operation"`
	Date      time.Time `json:"date"`
	Kind      string    `json:"kind"` // "transaction" | "transfer"
	Tags      []Tag     `json:"tags,omitempty"`
}

type TagStat struct {
	Tag   Tag `json:"tag"`
	Count int `json:"count"`
}

type UpcomingItem struct {
	Type      string    `json:"type"`      // "income" | "expense" | "goal_transfer"
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Operation string    `json:"operation"` // "add" | "subtract"
	Date      time.Time `json:"date"`
	Source    string    `json:"source"`
}

type BankAccountSummary struct {
	Account            BankAccount    `json:"account"`
	LatestTransactions []FeedEntry    `json:"latest_transactions"`
	TagStats           []TagStat      `json:"tag_stats"`
	UpcomingExpenses   []UpcomingItem `json:"upcoming_expenses"`
}

type StatementResponse struct {
	Account      BankAccount                  `json:"account"`
	Transactions PaginatedResponse[FeedEntry] `json:"transactions"`
	Upcoming     []UpcomingItem               `json:"upcoming"`
}
