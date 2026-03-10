-- ============================================================
-- SurrealDB Schema: agenticdbs demo
-- Covers: relational, graph, vector, timeseries, JSON, full-text
-- ============================================================

-- Use the agenticdbs namespace and database
USE NS agenticdbs DB agenticdbs;

-- ============================================================
-- RELATIONAL: Users, Products, Orders
-- ============================================================

DEFINE TABLE OVERWRITE users SCHEMAFULL;
DEFINE FIELD OVERWRITE name ON users TYPE string;
DEFINE FIELD OVERWRITE email ON users TYPE string;
DEFINE FIELD OVERWRITE created_at ON users TYPE datetime DEFAULT time::now();
DEFINE INDEX OVERWRITE idx_users_email ON users FIELDS email UNIQUE;

DEFINE TABLE OVERWRITE products SCHEMAFULL;
DEFINE FIELD OVERWRITE name ON products TYPE string;
DEFINE FIELD OVERWRITE description ON products TYPE string;
DEFINE FIELD OVERWRITE price ON products TYPE float;
DEFINE FIELD OVERWRITE category ON products TYPE string;
DEFINE FIELD OVERWRITE created_at ON products TYPE datetime DEFAULT time::now();

DEFINE TABLE OVERWRITE orders SCHEMAFULL;
DEFINE FIELD OVERWRITE user ON orders TYPE record<users>;
DEFINE FIELD OVERWRITE product ON orders TYPE record<products>;
DEFINE FIELD OVERWRITE quantity ON orders TYPE int;
DEFINE FIELD OVERWRITE total ON orders TYPE float;
DEFINE FIELD OVERWRITE ordered_at ON orders TYPE datetime DEFAULT time::now();

-- ============================================================
-- GRAPH: Relationship edges
-- ============================================================

DEFINE TABLE OVERWRITE knows SCHEMAFULL TYPE RELATION FROM users TO users;
DEFINE FIELD OVERWRITE since ON knows TYPE datetime;

DEFINE TABLE OVERWRITE reviewed SCHEMAFULL TYPE RELATION FROM users TO products;
DEFINE FIELD OVERWRITE rating ON reviewed TYPE int;
DEFINE FIELD OVERWRITE text ON reviewed TYPE string;
DEFINE FIELD OVERWRITE reviewed_at ON reviewed TYPE datetime DEFAULT time::now();

DEFINE TABLE OVERWRITE purchased SCHEMAFULL TYPE RELATION FROM users TO products;
DEFINE FIELD OVERWRITE quantity ON purchased TYPE int;
DEFINE FIELD OVERWRITE purchased_at ON purchased TYPE datetime DEFAULT time::now();

-- ============================================================
-- VECTOR: Product embeddings with HNSW index
-- ============================================================

DEFINE FIELD OVERWRITE embedding ON products TYPE option<array<float>>;
DEFINE INDEX OVERWRITE idx_products_embedding ON products FIELDS embedding
  HNSW DIMENSION 384 DIST COSINE;

-- ============================================================
-- TIMESERIES: Page metrics
-- ============================================================

DEFINE TABLE OVERWRITE page_metrics SCHEMAFULL;
DEFINE FIELD OVERWRITE ts ON page_metrics TYPE datetime;
DEFINE FIELD OVERWRITE product ON page_metrics TYPE record<products>;
DEFINE FIELD OVERWRITE views ON page_metrics TYPE int;
DEFINE FIELD OVERWRITE clicks ON page_metrics TYPE int;
DEFINE INDEX OVERWRITE idx_metrics_ts ON page_metrics FIELDS ts;

-- ============================================================
-- JSON/FLEXIBLE: User preferences (schema-less)
-- ============================================================

DEFINE TABLE OVERWRITE user_preferences SCHEMALESS;

-- ============================================================
-- FULL-TEXT SEARCH: Product descriptions
-- ============================================================

DEFINE ANALYZER OVERWRITE product_analyzer TOKENIZERS blank,class FILTERS lowercase,snowball(english);
DEFINE INDEX OVERWRITE idx_products_ft ON products FIELDS description
  FULLTEXT ANALYZER product_analyzer BM25;
