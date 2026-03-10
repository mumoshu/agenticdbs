# SurrealDB Query Guide

## Connection Info
- **Host**: localhost:18000
- **Namespace**: agenticdbs
- **Database**: agenticdbs
- **Auth**: root / root
- **Protocol**: HTTP REST or WebSocket

## Query Execution
Use `agenticdbs query --backend surreal "YOUR QUERY HERE"`

All queries must be prefixed with namespace/database context (handled automatically by the tool).

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

### Join orders with users
```sql
SELECT *, user.name AS user_name, product.name AS product_name FROM orders;
```

### Aggregate
```sql
SELECT count() AS total, math::sum(total) AS revenue FROM orders GROUP ALL;
```

---

## Graph Queries

### Traverse friends (direct)
```sql
SELECT ->knows->users.name AS friends FROM users:alice;
```

### Friends of friends
```sql
SELECT ->knows->users->knows->users.name AS fof FROM users:alice;
```

### All reviews by a user
```sql
SELECT ->reviewed->products.name AS reviewed_products, ->reviewed.rating AS ratings FROM users:alice;
```

### Reverse traversal (who knows Bob?)
```sql
SELECT <-knows<-users.name AS known_by FROM users:bob;
```

### Find paths
```sql
SELECT ->knows->users WHERE name = 'Eve Davis' FROM users:alice;
```

---

## Vector Search

### Find similar products (cosine similarity)
```sql
SELECT name, vector::similarity::cosine(embedding, $query_vector) AS similarity
FROM products
WHERE embedding <|10|> $query_vector
ORDER BY similarity DESC;
```

### Using a product's own embedding as query
```sql
LET $ref = (SELECT embedding FROM products:headphones);
SELECT name, vector::similarity::cosine(embedding, $ref.embedding) AS similarity
FROM products
WHERE id != products:headphones
ORDER BY similarity DESC
LIMIT 3;
```

---

## Timeseries Queries

### Recent metrics
```sql
SELECT * FROM page_metrics ORDER BY ts DESC LIMIT 10;
```

### Metrics in time range
```sql
SELECT * FROM page_metrics
WHERE ts >= d'2024-06-01T10:00:00Z' AND ts <= d'2024-06-01T12:00:00Z'
ORDER BY ts;
```

### Aggregate by product
```sql
SELECT product, math::sum(views) AS total_views, math::sum(clicks) AS total_clicks
FROM page_metrics
GROUP BY product;
```

---

## JSON / Schema-less Queries

### Get all preferences
```sql
SELECT * FROM user_preferences;
```

### Query nested fields
```sql
SELECT user, theme, notifications.email AS email_notifications
FROM user_preferences
WHERE theme = 'dark';
```

### Access deeply nested fields
```sql
SELECT custom_fields.accessibility FROM user_preferences:dave;
```

### Update partial
```sql
UPDATE user_preferences:alice SET notifications.sms = true;
```

---

## Full-text Search

### Search product descriptions
```sql
SELECT name, description, search::score(1) AS relevance
FROM products
WHERE description @1@ 'wireless'
ORDER BY relevance DESC;
```

### Search with multiple terms
```sql
SELECT name, description, search::score(1) AS relevance
FROM products
WHERE description @1@ 'wireless bluetooth'
ORDER BY relevance DESC;
```

---

## Schema Inspection

### Table info
```sql
INFO FOR TABLE users;
```

### Database info
```sql
INFO FOR DB;
```

### Namespace info
```sql
INFO FOR NS;
```

---

## Common Pitfalls

1. **Record IDs**: SurrealDB uses `table:id` format (e.g., `users:alice`). Don't use numeric IDs without the table prefix.
2. **Graph traversal**: Use `->edge->target` for outgoing, `<-edge<-source` for incoming. The arrow direction matters.
3. **SCHEMALESS vs SCHEMAFULL**: Schema-less tables accept any field. Schema-full tables reject undefined fields.
4. **Datetime format**: Use `d'2024-01-01T00:00:00Z'` for datetime literals (note the `d` prefix).
5. **Vector queries**: The `<|k|>` operator limits vector search to k nearest neighbors. Always combine with `ORDER BY similarity DESC`.
6. **Namespace/Database**: Always ensure you're using the correct NS/DB context.
