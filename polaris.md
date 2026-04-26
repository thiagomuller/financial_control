# My Finances APP — Polaris (Project Tracker)

This file tracks deliverables across all milestones.

---

## Milestone 1

### Deliverable 1: Project Scaffolding
- [x] Go REST API with standard library only
- [x] Angular 21 SPA with standalone components
- [x] PostgreSQL database with migration runner
- [x] docker-compose.yml (db + api + ui)

### Deliverable 2: Authentication
- [x] POST /api/auth/register
- [x] POST /api/auth/login
- [x] JWT middleware protecting all private routes
- [x] Password hashing with SHA-256 + random salt

### Deliverable 3: Bank Accounts
- [x] CRUD endpoints
- [x] Balance tracking (updated atomically with transactions)
- [x] icon_url support

### Deliverable 4: Tags
- [x] CRUD endpoints
- [x] Color field
- [x] Many-to-many association with transactions and transfers

### Deliverable 5: Transactions
- [x] CRUD endpoints
- [x] add/subtract operations
- [x] Atomic balance update with DB transaction
- [x] Tag associations

### Deliverable 6: Transfers
- [x] CRUD endpoints
- [x] SELECT FOR UPDATE lock to prevent overdraft
- [x] Atomic balance update on both source and target accounts

### Deliverable 7: Dashboard, Goals, Incomes, Expenses, Scheduler
- [x] Dashboard summary endpoint
- [x] Goals CRUD
- [x] Incomes/Expenses CRUD
- [x] Recurring transaction scheduler (background goroutine)
- [x] Bank account statement endpoint with upcoming items

---

## Milestone 2

### Deliverable 8: README
- [x] AI disclaimer added
- [x] MIT license reference added
- [x] Running instructions (podman-compose up)
- [x] Test commands documented
- [x] Security scan commands documented

### Deliverable 9: E2E Tests (Playwright)
- [x] Playwright test suite created under e2e/
- [x] Tests cover registration, login, bank accounts, transactions, transfers

### Deliverable 10: Fix Bank Account Creation with Image
- [x] icon_url column type changed from VARCHAR(500) to TEXT
- [x] Update handler supports changing the icon on an existing account
- [x] Update handler supports removing the icon (clearing icon_url)

### Deliverable 11: Fix Bank Account Initial Balance Validation
- [x] Validation added: initial_balance must be >= 0
- [x] Returns 400 with descriptive error on negative value

### Deliverable 12: Contrast / Dark Theme
- [x] Dark/light theme toggle implemented with CSS variables
- [x] data-theme attribute on :root drives color swap
- [x] ThemeService persists selection to localStorage
- [x] Color palette picker added to navigation bar

### Deliverable 13: Transfers in Statement / Summary
- [x] Transfers appear in bank account statement alongside transactions
- [x] Transfers appear in dashboard summary recent-transactions
- [x] UNION query merges transactions and transfers ordered by date
- [x] type discriminator ("transaction" / "transfer") included in response

### Deliverable 14: Makefiles
- [x] my_finances_api/Makefile with lint, trivy, trivy-image, test, build, all targets
- [x] my_finances_ui/Makefile with lint, trivy, trivy-image, test, build, all targets

### Deliverable 15: Update Documentation
- [x] documentation.md Section 7 updated (icon_url is now TEXT)
- [x] documentation.md Section 8 updated (dark theme design documented)
- [x] documentation.md Section 12 added (Milestone 2 Changes)
- [x] Table of Contents updated in documentation.md
- [x] polaris.md updated with all Milestone 2 deliverables checked

---

## Milestone Table

| Milestone | Description                        | Status      |
|-----------|------------------------------------|-------------|
| M1        | Core features (auth, CRUD, UI)     | Done        |
| M2        | Quality, fixes, UX improvements    | Done        |
