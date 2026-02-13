-- Drop trigger
DROP TRIGGER IF EXISTS update_appointments_updated_at ON appointments;

-- Drop indexes
DROP INDEX IF EXISTS idx_appointments_metadata;
DROP INDEX IF EXISTS idx_appointments_time_range;
DROP INDEX IF EXISTS idx_appointments_product_status;
DROP INDEX IF EXISTS idx_appointments_product_time;
DROP INDEX IF EXISTS idx_appointments_deleted_at;
DROP INDEX IF EXISTS idx_appointments_created_at;
DROP INDEX IF EXISTS idx_appointments_created_by;
DROP INDEX IF EXISTS idx_appointments_status;
DROP INDEX IF EXISTS idx_appointments_end_time;
DROP INDEX IF EXISTS idx_appointments_start_time;
DROP INDEX IF EXISTS idx_appointments_product_id;

-- Drop table
DROP TABLE IF EXISTS appointments;
