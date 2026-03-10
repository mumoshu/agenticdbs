-- ============================================================
-- PostgreSQL Extensions Setup
-- ============================================================

-- Vector similarity search
CREATE EXTENSION IF NOT EXISTS vector;

-- Graph database (Apache AGE)
CREATE EXTENSION IF NOT EXISTS age;

-- Timeseries (TimescaleDB)
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Load AGE and set search path
LOAD 'age';
SET search_path = ag_catalog, "$user", public;
