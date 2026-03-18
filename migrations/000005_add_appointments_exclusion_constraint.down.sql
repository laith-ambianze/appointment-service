-- Remove the indexes
DROP INDEX IF EXISTS idx_appointments_provider_status_time;
DROP INDEX IF EXISTS idx_appointments_provider_id;
DROP INDEX IF EXISTS idx_appointments_provider_time_range;

-- Note: We don't drop the btree_gist extension as other objects might depend on it
