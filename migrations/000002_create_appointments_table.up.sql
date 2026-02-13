-- Create appointments table
CREATE TABLE IF NOT EXISTS appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    location VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    created_by VARCHAR(255) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT appointments_time_check CHECK (end_time > start_time),
    CONSTRAINT appointments_status_check CHECK (status IN ('scheduled', 'confirmed', 'cancelled', 'completed', 'no_show'))
);

-- Create indexes for performance
CREATE INDEX idx_appointments_product_id ON appointments(product_id);
CREATE INDEX idx_appointments_start_time ON appointments(start_time);
CREATE INDEX idx_appointments_end_time ON appointments(end_time);
CREATE INDEX idx_appointments_status ON appointments(status);
CREATE INDEX idx_appointments_created_by ON appointments(created_by);
CREATE INDEX idx_appointments_created_at ON appointments(created_at);
CREATE INDEX idx_appointments_deleted_at ON appointments(deleted_at) WHERE deleted_at IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX idx_appointments_product_time ON appointments(product_id, start_time, end_time);
CREATE INDEX idx_appointments_product_status ON appointments(product_id, status);
CREATE INDEX idx_appointments_time_range ON appointments(start_time, end_time) WHERE deleted_at IS NULL;

-- JSONB index for metadata queries
CREATE INDEX idx_appointments_metadata ON appointments USING gin(metadata);

-- Create trigger for updated_at
CREATE TRIGGER update_appointments_updated_at
    BEFORE UPDATE ON appointments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments
COMMENT ON TABLE appointments IS 'Core appointments with flexible participant support';
COMMENT ON COLUMN appointments.created_by IS 'External user ID who created the appointment';
COMMENT ON COLUMN appointments.timezone IS 'Timezone for the appointment (e.g., America/New_York)';
COMMENT ON COLUMN appointments.metadata IS 'Flexible JSON metadata for appointment-specific data';
COMMENT ON COLUMN appointments.status IS 'Appointment status: scheduled, confirmed, cancelled, completed, no_show';
