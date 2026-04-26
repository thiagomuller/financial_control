---
version: 1.0
status: draft
last-updated: YYYY-MM-DD
owner: "[QA Lead / Tech Lead]"
---

# Test Strategy

> Defines the testing approach, coverage targets, test types, and CI integration for this project.
> Update when new test suites are added or coverage targets change: bump `version` in frontmatter.

## Testing Principles

1. Tests are written before or alongside production code (TDD where applicable)
2. Every public interface has automated tests
3. Every test, weather unit, integration or E2E, must follow clean code practices, such as KISS, don't repeat yourself, etc.

## Test Types and Ownership

| Type | Tooling | Ownership | Run On |
|---|---|---|---|
| Unit | Go native/ Angular native | Claude | Every commit |
| Integration | Go Native + test containers | Claude | Every commit |
| End-to-end | Playwright, Go native, podman-compose to pull up the stack | Claude | Every commit |

## Coverage Targets

| Layer | Target | Current |
|---|---|---|
| Unit | 80% | TBD |
| Integration | 60% | TBD |
| E2E critical paths | 100% | TBD |

## Critical Test Paths

The following paths must have end-to-end test coverage at all times:

1. User authentication and session management
2. Core AI inference request and response flow
3. Data persistence and retrieval
4. Pulling the entire stack, then using a test user to test the workflow going from the UI all the way to the database


## Test Environments

| Environment | Purpose | Who Has Access |
|---|---|---|
| Local (dev) | Unit + integration + E2E | All developers |

## Test Data Strategy

[Describe how test data is managed â€” e.g. factories, fixtures, anonymized production data, mocked external services]

## Change Log

| Version | Date | Author | Summary |
|---|---|---|---|
| 1.0 | 2026-04-26 | Thiago | Test strategy |