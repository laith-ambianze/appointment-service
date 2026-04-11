-- Create availability_breaks table
-- Defines break periods within a provider's availability rule where no slots are generated

CREATE TABLE IF NOT EXISTS availability_breaks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES provider_availability_rules(id) ON DELETE CASCADE,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ab_time_check CHECK (end_time > start_time)
);

-- Index for efficient lookup by rule
CREATE INDEX idx_ab_rule_id ON availability_breaks(rule_id);

-- Add comments
COMMENT ON TABLE availability_breaks IS 'Break periods within an availability rule where no appointment slots are generated';
COMMENT ON COLUMN availability_breaks.rule_id IS 'References the parent availability rule';
COMMENT ON COLUMN availability_breaks.start_time IS 'Start of the break period (in rule timezone)';
COMMENT ON COLUMN availability_breaks.end_time IS 'End of the break period (in rule timezone)';
