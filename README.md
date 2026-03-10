# agenticdbs

```
                          "find friends-of-friends"
                          "build me a REST API"
                          "migrate users to postgres"
                                    |
                                    v
                            +--------------+
                You ------->| Claude Code  |
                            |   /db skill  |
                            +--------------+
                              |          |
                    query /   |          |   \ generate
                   migrate    |          |    app code
                              v          v
                    +------------+  +------------+
                    | SurrealDB  |  | PostgreSQL |    <-- containers
                    |------------|  |------------|
                    | relational |  | relational |
                    | graph      |  | pgvector   |
                    | vector     |  | AGE(graph) |
                    | timeseries |  | TimescaleDB|
                    | JSON       |  | JSONB      |
                    | full-text  |  | tsvector   |
                    +------------+  +------------+
                              ^          ^
                              |          |
                              +----++----+
                                   ||
                            your app connects
                              directly too
```

A multi-backend agentic database system that runs **SurrealDB** and **PostgreSQL** side by side — each loaded with relational, graph, vector, timeseries, JSON, and full-text search capabilities.

Talk to it in natural language. It queries, migrates, debugs, and builds apps for you — or gives your app direct access to the databases.

## Getting Started

### Requirements

- Go 1.25+
- Docker & Docker Compose
- [Claude Code](https://claude.ai/claude-code) (for agentic usage via the `/db` skill)

### Install

```bash
go install github.com/mumoshu/agenticdbs@latest
```

## Using with Claude Code

`agenticdbs setup` installs a `/db` skill into `.claude/skills/db.md`. Once set up, just talk to Claude Code in natural language from your project directory:

```
You: "Find all users who bought electronics and show their social connections"

Claude Code: I'll query both backends for this. Let me use the /db skill...
             [runs: agenticdbs query -b surreal
               "SELECT name, ->purchased->products[WHERE category = 'electronics'].name AS bought,
                       ->knows->users.name AS friends
                FROM users"]

⟹ Alice Johnson  | bought: [Headphones, Mouse]  | friends: [Bob, Carol]
   Bob Smith      | bought: [Keyboard]           | friends: [Dave]
```

```
You: "Migrate the users table from SurrealDB to PostgreSQL"

Claude Code: I'll spawn sub-agents for both backends in parallel...
             [SurrealDB agent exports schema + data]
             [PostgreSQL agent translates and imports]
             Done — 5 users migrated. Verified counts match on both sides.
```

The skill auto-detects intent (query, debug, migrate, app-dev) and picks the right backend. For complex tasks, it delegates to backend-specific sub-agents.

## Implementation

- [docs/design.md](docs/design.md) — Architecture and interaction flow (how humans and Claude Code use agenticdbs)
- [docs/backends.md](docs/backends.md) — Backend details, Dockerfiles, query guides, and skills
- [docs/agenticdbs_cli.md](docs/agenticdbs_cli.md) — CLI reference and query examples
