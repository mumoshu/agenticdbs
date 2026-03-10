# PostgreSQL Query Guide

## Connection Info
- **Host**: localhost:15432
- **Database**: agenticdbs
- **User**: agenticdbs
- **Password**: agenticdbs

## Query Execution
Use `agenticdbs query --backend postgres "YOUR QUERY HERE"`

For AGE graph queries, the tool automatically runs `LOAD 'age'` and sets the search path.

---

## Relational Queries

### Select all users
```sql
SELECT * FROM users;
```

### Select with WHERE
```sql
SELECT * FROM users WHERE email = 'alice@example.com';
```

### Join orders with users and products
```sql
SELECT o.id, u.name AS user_name, p.name AS product_name, o.quantity, o.total
FROM orders o
JOIN users u ON o.user_id = u.id
JOIN products p ON o.product_id = p.id;
```

### Aggregate
```sql
SELECT COUNT(*) AS total_orders, SUM(total) AS total_revenue FROM orders;
```

---

## Graph Queries (Apache AGE)

**Important**: Every AGE query session needs:
```sql
LOAD 'age';
SET search_path = ag_catalog, "$user", public;
```

### Find friends of Alice
```sql
SELECT * FROM cypher('social', $$
  MATCH (a:User {name: 'Alice Johnson'})-[:KNOWS]->(friend:User)
  RETURN friend.name
$$) as (friend_name agtype);
```

### Friends of friends
```sql
SELECT * FROM cypher('social', $$
  MATCH (a:User {name: 'Alice Johnson'})-[:KNOWS]->()-[:KNOWS]->(fof:User)
  WHERE fof.name <> 'Alice Johnson'
  RETURN DISTINCT fof.name
$$) as (fof_name agtype);
```

### Find reviews by a user
```sql
SELECT * FROM cypher('social', $$
  MATCH (u:User {name: 'Alice Johnson'})-[r:REVIEWED]->(p:Product)
  RETURN p.name, r.rating, r.text
$$) as (product agtype, rating agtype, review_text agtype);
```

### Reverse traversal (who knows Bob?)
```sql
SELECT * FROM cypher('social', $$
  MATCH (knower:User)-[:KNOWS]->(b:User {name: 'Bob Smith'})
  RETURN knower.name
$$) as (knower_name agtype);
```

---

## Vector Search (pgvector)

### Find similar products (cosine distance)
```sql
SELECT name, 1 - (embedding <=> (SELECT embedding FROM products WHERE id = 1)) AS similarity
FROM products
WHERE id != 1
ORDER BY embedding <=> (SELECT embedding FROM products WHERE id = 1)
LIMIT 3;
```

### Vector search with explicit query vector
```sql
SELECT name, 1 - (embedding <=> '[0.12, -0.34, ...]'::vector) AS similarity
FROM products
ORDER BY embedding <=> '[0.12, -0.34, ...]'::vector
LIMIT 5;
```

---

## Timeseries Queries (TimescaleDB)

### Recent metrics
```sql
SELECT * FROM page_metrics ORDER BY ts DESC LIMIT 10;
```

### Hourly aggregation with time_bucket
```sql
SELECT
  time_bucket('1 hour', ts) AS hour,
  product_id,
  SUM(views) AS total_views,
  SUM(clicks) AS total_clicks
FROM page_metrics
GROUP BY hour, product_id
ORDER BY hour DESC;
```

### Daily aggregation
```sql
SELECT
  time_bucket('1 day', ts) AS day,
  SUM(views) AS total_views,
  SUM(clicks) AS total_clicks,
  ROUND(SUM(clicks)::numeric / NULLIF(SUM(views), 0) * 100, 2) AS ctr_percent
FROM page_metrics
GROUP BY day
ORDER BY day;
```

### Time range filter
```sql
SELECT * FROM page_metrics
WHERE ts >= '2024-06-01T10:00:00Z' AND ts <= '2024-06-01T12:00:00Z'
ORDER BY ts;
```

---

## JSON / JSONB Queries

### Get all preferences
```sql
SELECT u.name, up.preferences
FROM user_preferences up
JOIN users u ON up.user_id = u.id;
```

### Extract specific fields
```sql
SELECT
  u.name,
  up.preferences->>'theme' AS theme,
  up.preferences->'notifications'->>'email' AS email_notifications
FROM user_preferences up
JOIN users u ON up.user_id = u.id;
```

### Containment query
```sql
SELECT u.name, up.preferences
FROM user_preferences up
JOIN users u ON up.user_id = u.id
WHERE up.preferences @> '{"theme": "dark"}'::jsonb;
```

### Update JSONB field
```sql
UPDATE user_preferences
SET preferences = jsonb_set(preferences, '{notifications,sms}', 'true')
WHERE user_id = 1;
```

### Query deeply nested fields
```sql
SELECT preferences->'custom_fields'->'accessibility' AS accessibility
FROM user_preferences
WHERE user_id = 4;
```

---

## Full-text Search

### Search product descriptions
```sql
SELECT name, description,
  ts_rank(description_tsv, to_tsquery('english', 'wireless')) AS relevance
FROM products
WHERE description_tsv @@ to_tsquery('english', 'wireless')
ORDER BY relevance DESC;
```

### Search with multiple terms (AND)
```sql
SELECT name, description,
  ts_rank(description_tsv, to_tsquery('english', 'wireless & bluetooth')) AS relevance
FROM products
WHERE description_tsv @@ to_tsquery('english', 'wireless & bluetooth')
ORDER BY relevance DESC;
```

### Search with OR
```sql
SELECT name, description,
  ts_rank(description_tsv, to_tsquery('english', 'wireless | keyboard')) AS relevance
FROM products
WHERE description_tsv @@ to_tsquery('english', 'wireless | keyboard')
ORDER BY relevance DESC;
```

---

## Schema Inspection

### List tables
```sql
SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';
```

### Describe table
```sql
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'users';
```

### List indexes
```sql
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'products';
```

### Check extensions
```sql
SELECT extname, extversion FROM pg_extension;
```

---

## Common Pitfalls

1. **AGE search_path**: Every session using AGE needs `LOAD 'age'; SET search_path = ag_catalog, "$user", public;` BEFORE any cypher() call.
2. **AGE return types**: All `cypher()` results must be cast with `as (col agtype)`. The number of columns must match the RETURN clause.
3. **pgvector operators**: `<=>` = cosine distance, `<->` = L2 distance, `<#>` = inner product (negative). Lower distance = more similar.
4. **TimescaleDB**: Tables must be converted to hypertables with `create_hypertable()` BEFORE inserting data.
5. **JSONB operators**: `->` returns JSONB, `->>` returns TEXT. Use `@>` for containment queries with GIN index.
6. **tsvector**: The `description_tsv` column is auto-generated. Use `to_tsquery('english', 'term')` with `@@` operator.
