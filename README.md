# agenticdbs

A multi-backend agentic database system that runs **SurrealDB** and **PostgreSQL** side by side — each loaded with relational, graph, vector, timeseries, JSON, and full-text search capabilities.

Built for AI agents (like Claude Code) to query, migrate, and build apps across both backends through a single CLI.

## Getting Started

### Install

```bash
go install github.com/mumoshu/agenticdbs@latest
```

### Run

```bash
# Extract config files (docker-compose, seeds, docs, Claude skill)
agenticdbs setup

# Start SurrealDB + PostgreSQL containers
agenticdbs up

# Load demo data (users, products, orders, social graph, embeddings, metrics)
agenticdbs seed
```

## Example: One Dataset, Two Query Languages

**Relational** — find Alice's orders:

```bash
$ agenticdbs query -b surreal "SELECT * FROM orders WHERE user = users:alice"
$ agenticdbs query -b postgres "SELECT * FROM orders WHERE user_id = 1"
```

**Graph traversal** — who does Alice know, and who do *they* know?

```bash
$ agenticdbs query -b surreal \
    "SELECT ->knows->users->knows->users.name AS friends_of_friends FROM users:alice"
# ⟹ [{ friends_of_friends: ["Dave Brown", "Eve Davis"] }]
```

```bash
$ agenticdbs query -b postgres \
    "SELECT * FROM cypher('social', \$\$ MATCH (a:User {name: 'Alice Johnson'})-[:KNOWS]->()-[:KNOWS]->(fof) RETURN fof.name \$\$) as (name agtype)"
# ⟹ "Dave Brown"
#    "Eve Davis"
```

**Full-text search** — find products mentioning "wireless":

```bash
$ agenticdbs query -b surreal \
    "SELECT name, search::score(1) AS score FROM products WHERE description @1@ 'wireless' ORDER BY score DESC"
# ⟹ Wireless Bluetooth Headphones  (score: 0.69)
#    Wireless Ergonomic Mouse        (score: 0.61)
#    The Great Adventure Novel       (score: 0.42)
```

**Teardown:**

```bash
agenticdbs down
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

## What's Inside

| Backend    | Port  | Capabilities |
|------------|-------|-------------|
| SurrealDB  | 18000 | Native relational, graph (RELATE), vector (HNSW), timeseries, schemaless JSON, full-text (BM25) |
| PostgreSQL | 15432 | pgvector, Apache AGE (graph/Cypher), TimescaleDB, JSONB, tsvector |

## Requirements

- Go 1.25+
- Docker & Docker Compose
