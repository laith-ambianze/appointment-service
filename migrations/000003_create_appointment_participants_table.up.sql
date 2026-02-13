-- Create appointment_participants table
CREATE TABLE IF NOT EXISTS appointment_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    external_user_id VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'guest',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    user_metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT appointment_participants_unique_user UNIQUE(appointment_id, external_user_id),
    CONSTRAINT appointment_participants_role_check CHECK (role IN ('host', 'guest', 'attendee', 'observer')),
    CONSTRAINT appointment_participants_status_check CHECK (status IN ('pending', 'accepted', 'declined', 'tentative'))
);

-- Create indexes
CREATE INDEX idx_participants_appointment_id ON appointment_participants(appointment_id);
CREATE INDEX idx_participants_external_user_id ON appointment_participants(external_user_id);
CREATE INDEX idx_participants_role ON appointment_participants(role);
CREATE INDEX idx_participants_status ON appointment_participants(status);

-- Composite indexes for common queries
CREATE INDEX idx_participants_appointment_role ON appointment_participants(appointment_id, role);
CREATE INDEX idx_participants_user_status ON appointment_participants(external_user_id, status);

-- JSONB index for user metadata queries
CREATE INDEX idx_participants_user_metadata ON appointment_participants USING gin(user_metadata);

-- Create trigger for updated_at
CREATE TRIGGER update_appointment_participants_updated_at
    BEFORE UPDATE ON appointment_participants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments
COMMENT ON TABLE appointment_participants IS 'Tracks all participants in appointments with flexible roles';
COMMENT ON COLUMN appointment_participants.external_user_id IS 'User ID from the external product system';
COMMENT ON COLUMN appointment_participants.role IS 'Participant role: host, guest, attendee, observer';
COMMENT ON COLUMN appointment_participants.status IS 'Participation status: pending, accepted, declined, tentative';
COMMENT ON COLUMN appointment_participants.user_metadata IS 'Flexible JSON metadata for user-specific data (name, email, etc.)';
