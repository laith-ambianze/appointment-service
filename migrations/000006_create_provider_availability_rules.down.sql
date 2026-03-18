-- Drop provider_availability_rules table and related objects

DROP TRIGGER IF EXISTS update_provider_availability_rules_updated_at ON provider_availability_rules;
DROP TABLE IF EXISTS provider_availability_rules;
