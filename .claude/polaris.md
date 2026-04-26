---
version: 1.0
status: draft
last-updated: 2026-04-21
owner: "[Thiago]"
---

# Polaris Project My Finances

> The single source of truth for **what the team is building and why**.
> Populated during inception from the validated Intent Document provided by product.
> Update when client intent evolves: bump `version` in frontmatter and add a row to the Change Log below.

## Project Overview

| Field | Value |
|---|---|
| **Project Name** | [My Finances] |
| **Client** | [Myself] |
| **Engagement Start** | 2026-04-21 |
| **Target Delivery** | 2026-04-21 |
| **Team Lead** | [Thiago] |

## Vision Statement

> This will be my personal financial control application, containing a backend written in golang, and a UI written in angular.

## Intended Deliverables and Acceptance Criteria

### Deliverable 1: My Finances API -> Bank Accounts

**Description:** Bank accounts hold the balance for that account, and the bank statement for that account.

**Acceptance Criteria:**
- [x] Bank accounts must have a name, an initial balance that defaults to zero, an optional icon to be uploaded by the user.
- [x] The API must be able to create, update, list and delete bank accounts.

**Priority:** High
**Target Milestone:** M1

### Deliverable 2: My Finances API -> Tags

**Description:** Tags are just labels to mark `transactions` with a certain, user defined, type.

**Acceptance Criteria:**
- [x] Tags must have a name.
- [x] Tags can optionally have a color.
- [x] The API must be able to create, update, list and delete tags.

**Target Milestone:** M1

---

### Deliverable 3: My Finances API -> Transactions

**Description:** Transactions are balance addition or substraction to a given bank account, at a given date and time.

**Acceptance Criteria:**
- [x] Transactions must have a name, a value, an operation between add or subtract, a target bank account, a date.
- [x] Transactions can optionally have one or more tags.
- [x] The API must be able to create, update, list and delete transactions.

**Target Milestone:** M1

---

### Deliverable 4: My Finances API -> Transfers

**Description:** Transfers are balance movement from one source bank account, to a target bank account.

**Acceptance Criteria:**
- [x] Transfers must have a name, a value, a source bank account, a target bank account, and a date.
- [x] Transfers may only occur if the balance in the source account is greater or equal the transfer's value
- [x] A transfer can optionally have one or more tags

**Target Milestone:** M1

---

### Deliverable 5: My Finances API -> Goals

**Description:** Goals are target value that the user wants to accumulate over a given time interval.

**Acceptance Criteria:**
- [x] Goals must have a start date, an end date, an interval, a source bank account, a target bank account and a target value.
- [x] Automatic transfers will be automatically created for a goal, based off the source bank account, the income, and the time interval.

**Target Milestone:** M1

---

### Deliverable 6: My Finances API -> Income and expenses

**Description:** Incomes and expenses are abstraction entities that track whatever transactions add or remove fixed values from their bank accounts on a periodically manner.

**Acceptance Criteria:**
- [x] Incomes must have a name, a repetable date, a value, and a target bank account.
- [x] Everytime the income repeatable date arrives, the income value is then added to the target bank account as a transaction.
- [x] Expenses must have a name, a repetable date, a value, and a target bank account.
- [x] Everytime the expense repeatable date arrives, the expense value is then subtracted from the target bank account as a transaction.

**Target Milestone:** M1

---

### Deliverable 6: My Finances API -> User account

**Description:** User accounts are used to login into the system, enabling a user to create their bank accounts, tags, transactions, etc.

**Acceptance Criteria:**
- [x] User accounts must have a name, an username, an email.
- [x] A user account might optionally have one or more bank accounts.
- [x] A user account might optionally have one or more incomes.
- [x] A user account might optionally have one or more expenses.

**Target Milestone:** M1

---

### Deliverable 7: My Finances UI -> Landpage

**Description:** The page that the user will land on after loging. 

**Acceptance Criteria:**
- [x] It should have a list of their bank account summaries.
- [x] Each bank account summary needs to have the latest three transactions.
- [x] Each bank account summary needs to have a graph showing the tags with most transactions.
- [x] Each bank account summary should show incoming expenses.
- [x] Each bank account summary should have a link to the detailed bank account statement.

**Target Milestone:** M1

---

### Deliverable 7: My Finances UI -> Bank account details page

**Description:** A page that shows all transactions of a bank account in details, paginated, in a grid. 

**Acceptance Criteria:**
- [x] Should show all transactions of that bank account on a grid, paginated by the ammount chosen by the user, limited to a maximum of a hundred sites.
- [x] Should show incoming incomes and expenses as future transactions.
- [x] Should show incoming goals calculated transactions.

**Target Milestone:** M1

___

### Deliverable 8: Create a README file for the project

**Description:** Fill the README file explaining what is this project and how it works.

**Acceptance Criteria:**
- [x] Should make very clear that this project is almost entirely AI generated, and it should not be used in production.
- [x] Should choose a license that makes this project completely public, free to be used by whoever like to do so.

---

### Deliverable 9: Create E2E tests

**Description:** Create automated E2E tests that go through the entire stack

**Acceptance Criteria:**
- [x] Should write E2E test scenarios for the existing functionallity by pulling the entire stack, API, UI, and DB, and using a test user to do the tests.


**Target Milestone:** M2

---

### Deliverable 10: Fix bank account creation with image Issue

**Description:** Trying to create a bank account with a image gets a 500 error from the backend

**Acceptance Criteria:**
- [x] Should be able to create a bank account with the image sucessfully.
- [x] Should be able to, once created, change that image to another one at any given time.
- [x] Should be able to just remove the image, if the user wants to.
- [x] Must add unit, integration and E2E test scenarios for this to make sure they validate the feature and avoid this bug in the future.
- [x] The unit, integration and E2E tests must be passing before considering this done.
- [x] Unit, integration and E2E test code must follow clean code practices.


**Target Milestone:** M2

---

### Deliverable 11: Fix the bank account creation with initial balance Issue

**Description:**: Unable to create a bank account with initial balance right now.

**Acceptance Criteria:**
- [x] Should be able to create a bank account with initial balance.
- [x] Initial balance on bank account creation must be a positive number, never negative.
- [x] Initial balance on bank account creation can be zero.
- [x] Must add unit, integration and E2E test scenarios for this to make sure they validate the feature and avoid this bug in the future.
- [x] The unit, integration and E2E tests must be passing before considering this done.
- [x] Unit, integration and E2E test code must follow clean code practices.

**Target Milestone:** M2

---

### Deliverable 12: Fix the contrast Issue with the UI

**Description:**: The cards for the different resources have the same color, or colors too close to the background, making it difficult to see their boundaries.

**Acceptance Criteria:**
- [x] Implement a dark theme on the app.
- [x] Dark theme must be toggleable by the user.
- [x] Both dark and default themes must have a color pallete picker for the user to customize their visual experience.
- [x] If in dark theme, the resources cards need to be at least slightly lighter than the background.
- [x] If in default/white theme, the resources cards need to be slightly darker than the background.
- [x] Add unit, integration and E2E tests to validate this functionality and make it testable before future commits changing this.
- [x] Unit, integration and E2E test code must follow clean code practices.

**Target Milestone:** M2

---

### Deliverable 13: Make it so it shows the transfers on the bank account summary and the bank account statement

**Description:**: Transfers should be considered part of the list on the bank account summary and bank account statement.

**Acceptance Criteria:**
- [x] Implement a feature that makes transfers part of the bank account summary, like a transaction
- [x] Implement a feature that makes transfer part of the bank account statement, like a transaction.
- [x] Add unit, integration and E2E tests to validate this functionality and make it testable before future commits changing this.
- [x] Unit, integration and E2E test code must follow clean code practices.

**Target Milestone:** M2

---

---

### Deliverable 14: Create Makefiles for both the UI and the API

**Description:**: The makefiles should make it possible to run the linter and the trivy scan for both projects.

**Acceptance Criteria:**
- [x] Make trivy runnable by using a Makefile command on each: the API, and the UI
- [x] Make trivy docker image scan runnable by using a Makefile command on each: the API, and the UI
- [x] Make linter runnable by using a Makefile command on each: the API, and the UI
- [x] Both trivy, trivy docker image scan and linter must pass on both projects before the milestone is considered done.

**Target Milestone:** M2

---

### Deliverable 15: Update the README.md and the documentation.md accordingly.

**Description:**: Update documentation.md and README.md with the relevant information after the second milestone.

**Acceptance Criteria:**
- [x] README.md must describe operational stuff, such as: how to run the project, how to run the security scans on the UI and the API, how to run the linter, etc.
- [x] documentation.md must describe, in depth, all the features present in the project, and must be up to date with the second milestone features.


**Target Milestone:** M2

---

## Milestones

| Milestone | Description | Target Date | Status |
|---|---|---|---|
| M1 | Basic functionalities | 2026-04-25 | Done |
| M2 | Bug solving and little improvements | 2026-04-26 | Not Started |