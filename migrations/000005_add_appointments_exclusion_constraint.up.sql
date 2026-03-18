-- Add indexes and extension for availability system support
-- While exclusion constraints with tstzrange have PostgreSQL IMMUTABLE requirements,
-- we rely on application-level concurrency control using:
-- 1. REPEATABLE READ isolation level
-- 2. SELECT FOR UPDATE locking
-- 3. Explicit conflict checking before inserts

-- First, enable the btree_gist extension (useful for GiST indexes on common types)
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Add GiST index on provider_id and time range for efficient overlap queries
-- This index accelerates conflict detection queries
CREATE INDEX IF NOT EXISTS idx_appointments_provider_time_range 
    ON appointments USING gist (provider_id, tsrange(start_time::timestamp, end_time::timestamp));

-- Add B-tree index for provider_id lookups
CREATE INDEX IF NOT EXISTS idx_appointments_provider_id 
    ON appointments (provider_id) 
    WHERE provider_id IS NOT NULL;

-- Add composite index for common query pattern: provider + status + time range
CREATE INDEX IF NOT EXISTS idx_appointments_provider_status_time 
    ON appointments (provider_id, status, start_time, end_time) 
    WHERE provider_id IS NOT NULL AND deleted_at IS NULL;
