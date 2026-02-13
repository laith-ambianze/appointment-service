-- Drop trigger
DROP TRIGGER IF EXISTS update_appointment_participants_updated_at ON appointment_participants;

-- Drop indexes
DROP INDEX IF EXISTS idx_participants_user_metadata;
DROP INDEX IF EXISTS idx_participants_user_status;
DROP INDEX IF EXISTS idx_participants_appointment_role;
DROP INDEX IF EXISTS idx_participants_status;
DROP INDEX IF EXISTS idx_participants_role;
DROP INDEX IF EXISTS idx_participants_external_user_id;
DROP INDEX IF EXISTS idx_participants_appointment_id;

-- Drop table
DROP TABLE IF EXISTS appointment_participants;
