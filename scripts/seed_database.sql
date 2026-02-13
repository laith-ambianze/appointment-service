-- Seed test data for development

-- Insert test product
INSERT INTO products (id, name, description, api_key, api_secret_hash, status, metadata)
VALUES 
(
    '550e8400-e29b-41d4-a716-446655440000',
    'Test Product',
    'A test product for development',
    'test_api_key_123',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- bcrypt hash of "secret"
    'active',
    '{"plan": "free", "test": true}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

-- Insert second test product
INSERT INTO products (id, name, description, api_key, api_secret_hash, status, metadata)
VALUES 
(
    '550e8400-e29b-41d4-a716-446655440001',
    'Demo CRM',
    'A demo CRM product for testing',
    'demo_api_key_456',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'active',
    '{"plan": "enterprise", "features": ["webhooks", "analytics"]}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

-- Insert test appointments
INSERT INTO appointments (id, product_id, title, description, start_time, end_time, timezone, status, created_by, metadata)
VALUES
(
    '660e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440000',
    'Test Appointment',
    'A test appointment for development',
    NOW() + INTERVAL '1 day',
    NOW() + INTERVAL '1 day' + INTERVAL '1 hour',
    'UTC',
    'scheduled',
    'user_test_123',
    '{"meeting_type": "test"}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO appointments (id, product_id, title, description, start_time, end_time, timezone, location, status, created_by, metadata)
VALUES
(
    '660e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440000',
    'Product Demo Call',
    'Demo of our new product features',
    NOW() + INTERVAL '2 days',
    NOW() + INTERVAL '2 days' + INTERVAL '30 minutes',
    'America/New_York',
    'https://zoom.us/j/123456789',
    'confirmed',
    'user_test_123',
    '{"meeting_type": "demo", "priority": "high"}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO appointments (id, product_id, title, description, start_time, end_time, timezone, status, created_by, metadata)
VALUES
(
    '660e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440001',
    'Weekly Team Sync',
    'Regular weekly team sync meeting',
    NOW() + INTERVAL '3 days',
    NOW() + INTERVAL '3 days' + INTERVAL '45 minutes',
    'Europe/London',
    'scheduled',
    'user_manager_001',
    '{"meeting_type": "recurring", "recurrence": "weekly"}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

-- Insert test participants
INSERT INTO appointment_participants (appointment_id, external_user_id, role, status, user_metadata)
VALUES
(
    '660e8400-e29b-41d4-a716-446655440000',
    'user_test_123',
    'host',
    'accepted',
    '{"name": "Test Host", "email": "host@test.com"}'::jsonb
),
(
    '660e8400-e29b-41d4-a716-446655440000',
    'user_test_456',
    'guest',
    'pending',
    '{"name": "Test Guest", "email": "guest@test.com"}'::jsonb
)
ON CONFLICT (appointment_id, external_user_id) DO NOTHING;

-- Participants for Product Demo Call
INSERT INTO appointment_participants (appointment_id, external_user_id, role, status, user_metadata)
VALUES
(
    '660e8400-e29b-41d4-a716-446655440001',
    'user_test_123',
    'host',
    'accepted',
    '{"name": "John Smith", "email": "john@company.com", "title": "Sales Manager"}'::jsonb
),
(
    '660e8400-e29b-41d4-a716-446655440001',
    'user_client_001',
    'guest',
    'accepted',
    '{"name": "Jane Doe", "email": "jane@client.com", "company": "Client Corp"}'::jsonb
),
(
    '660e8400-e29b-41d4-a716-446655440001',
    'user_tech_001',
    'attendee',
    'tentative',
    '{"name": "Bob Tech", "email": "bob@company.com", "title": "Tech Lead"}'::jsonb
)
ON CONFLICT (appointment_id, external_user_id) DO NOTHING;

-- Participants for Weekly Team Sync
INSERT INTO appointment_participants (appointment_id, external_user_id, role, status, user_metadata)
VALUES
(
    '660e8400-e29b-41d4-a716-446655440002',
    'user_manager_001',
    'host',
    'accepted',
    '{"name": "Manager One", "email": "manager@company.com"}'::jsonb
),
(
    '660e8400-e29b-41d4-a716-446655440002',
    'user_dev_001',
    'attendee',
    'accepted',
    '{"name": "Developer One", "email": "dev1@company.com"}'::jsonb
),
(
    '660e8400-e29b-41d4-a716-446655440002',
    'user_dev_002',
    'attendee',
    'pending',
    '{"name": "Developer Two", "email": "dev2@company.com"}'::jsonb
)
ON CONFLICT (appointment_id, external_user_id) DO NOTHING;

SELECT 'Database seeded successfully!' AS result;
SELECT 'Products: ' || COUNT(*) FROM products;
SELECT 'Appointments: ' || COUNT(*) FROM appointments;
SELECT 'Participants: ' || COUNT(*) FROM appointment_participants;
