-- Create provider_availability_rules table
-- Defines working hours and slot configuration for providers

CREATE TABLE IF NOT EXISTS provider_availability_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    provider_id VARCHAR(255) NOT NULL,
    day_of_week INT NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 30,
    slot_interval_minutes INT NOT NULL DEFAULT 15,
    buffer_before_minutes INT NOT NULL DEFAULT 0,
    buffer_after_minutes INT NOT NULL DEFAULT 0,
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT par_day_of_week_check CHECK (day_of_week >= 0 AND day_of_week <= 6),
    CONSTRAINT par_time_check CHECK (end_time > start_time),
    CONSTRAINT par_duration_check CHECK (duration_minutes > 0 AND duration_minutes <= 480),
    CONSTRAINT par_interval_check CHECK (slot_interval_minutes > 0 AND slot_interval_minutes <= duration_minutes),
    CONSTRAINT par_buffer_before_check CHECK (buffer_before_minutes >= 0 AND buffer_before_minutes <= 120),
    CONSTRAINT par_buffer_after_check CHECK (buffer_after_minutes >= 0 AND buffer_after_minutes <= 120)
);

-- Create indexes for efficient queries
CREATE INDEX idx_par_provider_id ON provider_availability_rules(provider_id);
CREATE INDEX idx_par_product_id ON provider_availability_rules(product_id);
CREATE INDEX idx_par_provider_day ON provider_availability_rules(provider_id, day_of_week);
CREATE INDEX idx_par_product_provider ON provider_availability_rules(product_id, provider_id);
CREATE INDEX idx_par_active ON provider_availability_rules(provider_id, is_active) WHERE is_active = true AND deleted_at IS NULL;

-- Unique constraint: one rule per provider per day (within same product)
CREATE UNIQUE INDEX idx_par_unique_provider_day ON provider_availability_rules(product_id, provider_id, day_of_week) 
    WHERE deleted_at IS NULL;

-- Create trigger for updated_at
CREATE TRIGGER update_provider_availability_rules_updated_at
    BEFORE UPDATE ON provider_availability_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments
COMMENT ON TABLE provider_availability_rules IS 'Defines provider working hours and slot configuration for dynamic slot generation';
COMMENT ON COLUMN provider_availability_rules.provider_id IS 'External provider ID (matches appointments.provider_id)';
COMMENT ON COLUMN provider_availability_rules.day_of_week IS 'Day of week (0=Sunday, 1=Monday, ..., 6=Saturday)';
COMMENT ON COLUMN provider_availability_rules.start_time IS 'Start time of availability window in provider timezone';
COMMENT ON COLUMN provider_availability_rules.end_time IS 'End time of availability window in provider timezone';
COMMENT ON COLUMN provider_availability_rules.duration_minutes IS 'Duration of each appointment slot in minutes';
COMMENT ON COLUMN provider_availability_rules.slot_interval_minutes IS 'Interval between slot start times (e.g., 15 = slots at :00, :15, :30, :45)';
COMMENT ON COLUMN provider_availability_rules.buffer_before_minutes IS 'Buffer time required before each appointment';
COMMENT ON COLUMN provider_availability_rules.buffer_after_minutes IS 'Buffer time required after each appointment';
COMMENT ON COLUMN provider_availability_rules.timezone IS 'Timezone for interpreting start_time and end_time (e.g., America/New_York)';
COMMENT ON COLUMN provider_availability_rules.is_active IS 'Whether this availability rule is currently active';
