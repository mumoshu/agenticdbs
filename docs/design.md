# Design

## How It Works

There are two actors — **you** (the human) and **Claude Code** — and they interact with agenticdbs at different levels.

### Setup & Lifecycle (you run these directly)

```
  You (terminal)
    |
    |  $ agenticdbs setup
    |  $ agenticdbs up
    |  $ agenticdbs seed
    |  $ agenticdbs down
    |
    v
  +---------------------------+
  |  agenticdbs CLI           |
  |---------------------------|
  |  setup   extract embedded |-------> .claude/skills/db.md
  |          files to disk    |-------> docker/, seed/, docs/
  |                           |-------> CLAUDE.md (for end users)
  |  up      docker compose   |-------> SurrealDB  :18000
  |          up + health wait |-------> PostgreSQL :15432
  |  seed    load demo data   |-------> (both backends)
  |  down    teardown         |-------> containers removed
  +---------------------------+
```

### Querying & Tasks (Claude Code drives these)

```
  You: "find friends-of-friends in the graph"
    |
    v
  +----------------------------------------------+
  | Claude Code                                  |
  |----------------------------------------------|
  |  1. /db skill activates                      |
  |  2. reads docs/feature-matrix.md             |
  |     --> picks best backend for the task       |
  |  3. reads docs/{surrealdb,postgres}-guide.md |
  |     --> learns query syntax                   |
  |  4. runs:                                    |
  |     $ agenticdbs query --backend surreal \   |
  |       "SELECT ->knows->users.name FROM ..."  |
  |  5. returns result to you                    |
  +----------------------------------------------+
        |                              |
        v                              v
  +------------+                +------------+
  | SurrealDB  |                | PostgreSQL |
  | :18000     |                | :15432     |
  +------------+                +------------+
```

### Complex Tasks (sub-agents)

For migrations, debugging, and app-dev, Claude Code spawns specialized sub-agents:

```
  You: "migrate users from SurrealDB to PostgreSQL"
    |
    v
  +----------------------------------------------+
  | Claude Code  (/db skill)                     |
  |  detects MIGRATE intent                      |
  |----------------------------------------------|
  |                                              |
  |  spawns two sub-agents in parallel:          |
  |                                              |
  |  +-------------------+  +------------------+ |
  |  | SurrealDB expert  |  | PostgreSQL expert| |
  |  |-------------------|  |------------------| |
  |  | export schema     |  | translate syntax | |
  |  | export data       |  | import data      | |
  |  | verify counts     |  | verify counts    | |
  |  +-------------------+  +------------------+ |
  |         |                       |            |
  |         v                       v            |
  |  agenticdbs query         agenticdbs query   |
  |    --backend surreal        --backend postgres|
  +----------------------------------------------+
```

### End-to-End Flow

```
  +-----------+          +-------------+         +-------------------+
  |   Human   |          | Claude Code |         |   agenticdbs CLI  |
  +-----------+          +-------------+         +-------------------+
       |                       |                          |
       |  agenticdbs setup     |                          |
       |---------------------------------------------->   |
       |                       |          extract files   |
       |                       |          install /db     |
       |                       |            skill         |
       |  agenticdbs up        |                          |
       |---------------------------------------------->   |
       |                       |     docker compose up    |
       |                       |     health-check loop    |
       |  agenticdbs seed      |                          |
       |---------------------------------------------->   |
       |                       |     load schema + data   |
       |                       |                          |
       |  "query all orders    |                          |
       |   over $100"          |                          |
       |---------------------> |                          |
       |                       |  read feature-matrix.md  |
       |                       |  read backend guide      |
       |                       |  agenticdbs query -----> |
       |                       |      --backend surreal   |
       |                       |      "SELECT ..."        |
       |                       |                          |
       |                       | <--- JSON result ------- |
       |  <-- pretty answer -- |                          |
       |                       |                          |
       |  agenticdbs down      |                          |
       |---------------------------------------------->   |
       |                       |     docker compose down  |
       |                       |                          |
```

### What Gets Embedded in the Binary

The `agenticdbs` binary embeds everything needed for zero-dependency distribution:

```
  agenticdbs (single binary)
    |
    +-- docker/
    |     docker-compose.yml
    |     postgres/Dockerfile     (pgvector + AGE + TimescaleDB)
    |     surrealdb/Dockerfile
    |
    +-- seed/
    |     surreal/  00-schema.surql, 01-data.surql
    |     postgres/ 00-extensions.sql, 01-schema.sql, 02-data.sql
    |
    +-- docs/
    |     feature-matrix.md       (backend comparison)
    |     surrealdb-guide.md      (SurrealQL reference)
    |     postgres-guide.md       (SQL + extensions reference)
    |
    +-- skills/
          db.md                   --> .claude/skills/db.md
```

`agenticdbs setup` extracts these to disk so Docker can build images and Claude Code can read the docs and skill.
