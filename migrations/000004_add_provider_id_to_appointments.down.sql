-- Remove provider_id column from appointments table

DROP INDEX IF EXISTS idx_appointments_provider_time_range;
DROP INDEX IF EXISTS idx_appointments_provider_id;
ALTER TABLE appointments DROP COLUMN IF EXISTS provider_id;
