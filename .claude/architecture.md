---
version: 1.0
status: draft
last-updated: 2026-04-21
owner: "[Tech Lead]"
---

# Architecture

> Living document for tech stack, system components, data flows, and design decisions.
> Update when architectural decisions change: bump `version` and add a row to the Design Decisions Log.

## Tech Stack

| Layer | Technology | Version | Notes |
|---|---|---|---|
| API coding language | [Go] | [1.25.8] | |
| UI coding language | [Angular] | [Latest] | |
| Database | [PostgreSQL] | [Latest] | |
| AI / LLM | [Claude code] | | |
| Hosting | [Running locally only, UI has its own container, API has its own container, DB has its own container] | | |

## Dependencies
Do not use any external dependencies outsite the stack core, default tools. I.E. Go default testing framework, Angular default testing framework.

## Folder structure
`My Finances API` project should be on a new folder called "my_finances_api"
`My Finances UI` project should be on a new folder called "my_finances_ui"

## System Components

### [My Finances DB]

**Responsibility:** [Persists the data coming from My Finances API component]
**Technology:** [Postgressql]

---

### [My Finances API]

**Responsibility:** [Exposes the API for persisting and operating the data models for My Finances UI]
**Technology:** [Go]
**Interfaces:** [Exposes HTTP REST endpoints to be called by My Finances UI, receives HTTP calls from My Finances UI]

---

### [My Finances UI]

**Responsibility:** [Provides user interface for My Finances application]
**Technology:** [Angular]
**Interfaces:** [Receives data from the user / Exposes an user interface capable of calling the running My Finances API]

---

## Data Flows

### [User to database]

```
[User] -> [My Finances UI] -> [My Finances API] -> [Postgres Database] 
```

User opens My Finances UI, then uses its features to create bank accounts, tags, transactions, transfers or goals in it.

### [Database to user]

```
[Postgres Database] -> [My Finances API] -> [My Finances UI] -> [User] 
```

User opens My Finances UI, then its features to retrieve their bank accounts, tags, transactions, transfers or goals the database.

## Deployment
These components won't be deployed anywhere, instead, they will run locally only.
Each component, `My Finances DB`, `My Finances API`, `My Finances UI` must have their own container.
`My Finances API` and `My Finances UI` must have their own Dockerfile, in their folders.
When building `My Finances API` and `My Finances UI` docker images from their Dockerfiles, do it so by prefixing the command with a: `distrobox-host-exec <command>`. The same applies to running the docker-compose.yaml file.
