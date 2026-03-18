-- Add provider_id column to appointments table
-- This enables linking appointments to specific providers for availability management

ALTER TABLE appointments ADD COLUMN provider_id VARCHAR(255);

-- Create index for provider_id queries
CREATE INDEX idx_appointments_provider_id ON appointments(provider_id) WHERE provider_id IS NOT NULL;

-- Create composite index for provider + time range queries (used by availability checks)
CREATE INDEX idx_appointments_provider_time_range ON appointments(provider_id, start_time, end_time) 
    WHERE provider_id IS NOT NULL AND deleted_at IS NULL;

-- Add comment
COMMENT ON COLUMN appointments.provider_id IS 'External provider ID who will conduct the appointment';
