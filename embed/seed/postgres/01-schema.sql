-- ============================================================
-- PostgreSQL Schema: agenticdbs demo
-- Covers: relational, graph, vector, timeseries, JSON, full-text
-- ============================================================

-- ============================================================
-- RELATIONAL: Users, Products, Orders
-- ============================================================

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price FLOAT NOT NULL,
    category TEXT NOT NULL,
    embedding VECTOR(384),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    product_id INT REFERENCES products(id),
    quantity INT NOT NULL,
    total FLOAT NOT NULL,
    ordered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- VECTOR: HNSW index on product embeddings
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_products_embedding ON products
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- ============================================================
-- TIMESERIES: Page metrics with TimescaleDB hypertable
-- ============================================================

CREATE TABLE IF NOT EXISTS page_metrics (
    ts TIMESTAMPTZ NOT NULL,
    product_id INT REFERENCES products(id),
    views INT NOT NULL,
    clicks INT NOT NULL
);

SELECT create_hypertable('page_metrics', by_range('ts'), if_not_exists => TRUE);

-- ============================================================
-- JSON/FLEXIBLE: User preferences with JSONB
-- ============================================================

CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) UNIQUE,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_user_prefs_gin ON user_preferences USING GIN (preferences);

-- ============================================================
-- FULL-TEXT SEARCH: On product descriptions
-- ============================================================

ALTER TABLE products ADD COLUMN IF NOT EXISTS description_tsv TSVECTOR
    GENERATED ALWAYS AS (to_tsvector('english', description)) STORED;

CREATE INDEX IF NOT EXISTS idx_products_ft ON products USING GIN (description_tsv);

-- ============================================================
-- GRAPH: Social graph using Apache AGE
-- ============================================================

-- Load AGE for graph operations
LOAD 'age';
SET search_path = ag_catalog, "$user", public;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM ag_graph WHERE name = 'social') THEN
        PERFORM create_graph('social');
    END IF;
END $$;
