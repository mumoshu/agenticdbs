# agenticdbs CLI Reference

## Quick Start

```bash
# Extract config files (docker-compose, seeds, docs, Claude skill)
agenticdbs setup

# Start SurrealDB + PostgreSQL containers
agenticdbs up

# Load demo data (users, products, orders, social graph, embeddings, metrics)
agenticdbs seed
```

## Commands

| Command | Description |
|---------|-------------|
| `agenticdbs setup` | Extract config files (docker-compose, seeds, docs, Claude skill) |
| `agenticdbs up` | Start SurrealDB + PostgreSQL containers |
| `agenticdbs seed` | Load demo data into both backends |
| `agenticdbs query -b <backend> "QUERY"` | Execute a query against surreal or postgres |
| `agenticdbs down` | Stop and remove containers and volumes |

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
