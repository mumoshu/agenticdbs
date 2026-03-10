# Backends

| Backend    | Port  | Capabilities |
|------------|-------|-------------|
| SurrealDB  | 18000 | Native relational, graph (RELATE), vector (HNSW), timeseries, schemaless JSON, full-text (BM25) |
| PostgreSQL | 15432 | pgvector, Apache AGE (graph/Cypher), TimescaleDB, JSONB, tsvector |

## Dockerfiles

- [embed/docker/surrealdb/Dockerfile](../embed/docker/surrealdb/Dockerfile)
- [embed/docker/postgres/Dockerfile](../embed/docker/postgres/Dockerfile)
- [embed/docker/docker-compose.yml](../embed/docker/docker-compose.yml)

## Guides

- [docs/surrealdb-guide.md](surrealdb-guide.md) — SurrealQL syntax reference
- [docs/postgres-guide.md](postgres-guide.md) — SQL + extensions syntax reference
- [docs/feature-matrix.md](feature-matrix.md) — Backend comparison and recommendations

## Skills

- [embed/skills/db.md](../embed/skills/db.md) — Claude Code `/db` skill (installed by `agenticdbs setup`)
