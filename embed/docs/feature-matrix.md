# Feature Matrix: SurrealDB vs PostgreSQL

## Overview

| Feature | SurrealDB | PostgreSQL | Recommendation |
|---------|-----------|------------|----------------|
| Relational | Native SCHEMAFULL tables, SELECT/JOIN | Native SQL, strong typing, constraints, FK | **PostgreSQL** for complex relational models |
| Graph | Native `->` traversal, RELATE edges | Apache AGE extension, `cypher()` function | **SurrealDB** for ergonomic graph queries |
| Vector Search | Native HNSW (cosine/euclidean/manhattan) | pgvector extension (HNSW, IVFFlat) | **Both good**; pgvector more mature ecosystem |
| Timeseries | datetime fields + ORDER BY (no special engine) | TimescaleDB hypertables, `time_bucket()`, continuous aggregates | **PostgreSQL** significantly stronger |
| JSON/Flexible | Native SCHEMALESS mode | JSONB columns with GIN indexes | **SurrealDB** more natural for mixed schemas |
| Full-text Search | Built-in SEARCH analyzers, BM25 ranking | tsvector + GIN, `ts_rank` | **PostgreSQL** more mature ranking/config |

## Detailed Comparison

### Relational
- **SurrealDB**: SCHEMAFULL tables with typed fields. Supports basic JOINs but graph traversal is preferred for relationships.
- **PostgreSQL**: Gold standard for relational data. Full ACID, complex JOINs, constraints, triggers, stored procedures.

### Graph
- **SurrealDB**: First-class graph support. Create edges with `RELATE user:alice->knows->user:bob`. Traverse with `->knows->` arrow syntax. No JOINs needed.
- **PostgreSQL**: Apache AGE extension adds OpenCypher support. Use `SELECT * FROM cypher('graph', $$ MATCH (a)-[:KNOWS]->(b) RETURN a, b $$) as (a agtype, b agtype)`. Requires `LOAD 'age'` and `SET search_path`.

### Vector Search
- **SurrealDB**: Define HNSW index with `DEFINE INDEX idx ON table FIELDS embedding HNSW DIMENSION 384 DIST COSINE`. Query with `vector::similarity::cosine()`.
- **PostgreSQL**: pgvector with `VECTOR(384)` type. HNSW index with `USING hnsw (embedding vector_cosine_ops)`. Query with `ORDER BY embedding <=> query_vector LIMIT k`.

### Timeseries
- **SurrealDB**: Store timestamps as `datetime` type. Query with time ranges and ORDER BY. No special aggregation engine.
- **PostgreSQL**: TimescaleDB converts tables to hypertables with `create_hypertable()`. Use `time_bucket('1 hour', ts)` for aggregation. Supports continuous aggregates and retention policies.

### JSON/Flexible Objects
- **SurrealDB**: `DEFINE TABLE t SCHEMALESS` accepts any structure. Query nested fields directly: `SELECT custom_fields.favorite_category FROM user_preferences`.
- **PostgreSQL**: JSONB column with GIN index. Query with `->`, `->>`, `@>` operators. Update with `jsonb_set()`.

### Full-text Search
- **SurrealDB**: Define analyzer with `DEFINE ANALYZER ... TOKENIZERS blank,class FILTERS lowercase,snowball(english)`. Create search index with `FULLTEXT ANALYZER analyzer BM25`. Query with `search::score()`.
- **PostgreSQL**: Generated `tsvector` column with GIN index. Query with `to_tsquery('english', 'search terms')` and `@@` operator. Rank with `ts_rank()`.
