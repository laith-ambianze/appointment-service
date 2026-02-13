# Task W02-02: Database Migrations Setup

**Status**: ⏸️ Not Started  
**Estimated Time**: 3-4 hours  
**Prerequisites**: TASK_W02_01_DATABASE_PACKAGE.md  
**Next Task**: TASK_W02_03_DOMAIN_MODELS.md

---

## Objective

Set up the database migration system using golang-migrate and create the initial database schema for products, appointments, and participants tables.

---

## Steps

### 1. Install golang-migrate CLI

**Windows (PowerShell):**

```powershell
# Using Scoop (recommended)
scoop install migrate

# Or download from GitHub releases
# https://github.com/golang-migrate/migrate/releases/latest
```

**Verify installation:**

```bash
migrate -version
# Expected: 4.17.0 or higher
```

### 2. Update Makefile with Migration Commands

Location: `appointment-service/Makefile`

Add these migration commands:

```makefile
# Migration commands
MIGRATIONS_PATH=./migrations
DB_URL=postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: migrate-create migrate-up migrate-down migrate-status migrate-force migrate-version

migrate-create: ## Create new migration (usage: make migrate-create name=create_users)
 @echo "$(CYAN)Creating migration: $(name)$(NC)"
 @if [ -z "$(name)" ]; then \
  echo "$(RED)Error: name parameter is required$(NC)"; \
  echo "Usage: make migrate-create name=your_migration_name"; \
  exit 1; \
 fi
 migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)
 @echo "$(GREEN)Migration files created in $(MIGRATIONS_PATH)$(NC)"

migrate-up: ## Run all pending migrations
 @echo "$(CYAN)Running migrations up...$(NC)"
 migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
 @echo "$(GREEN)Migrations applied successfully$(NC)"

migrate-down: ## Rollback last migration
 @echo "$(CYAN)Rolling back last migration...$(NC)"
 migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1
 @echo "$(GREEN)Migration rolled back$(NC)"

migrate-status: ## Show migration status
 @echo "$(CYAN)Migration status:$(NC)"
 migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

migrate-force: ## Force migration version (usage: make migrate-force version=1)
 @echo "$(CYAN)Forcing migration version to $(version)$(NC)"
 @if [ -z "$(version)" ]; then \
  echo "$(RED)Error: version parameter is required$(NC)"; \
  exit 1; \
 fi
 migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(version)
 @echo "$(GREEN)Migration version forced$(NC)"

migrate-version: ## Show current migration version
 @migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version
```

Also update the DB_DSN variable at the top:

```makefile
# Database
DB_USER ?= appointments
DB_PASSWORD ?= password
DB_HOST ?= localhost
DB_PORT ?= 1998
DB_NAME ?= appointments_dev
DB_DSN=postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
```

### 3. Create Migration: Products Table

Run this command:

```bash
make migrate-create name=create_products_table
```

This creates two files:

- `migrations/000001_create_products_table.up.sql`
- `migrations/000001_create_products_table.down.sql`

**File: `migrations/000001_create_products_table.up.sql`**

```sql
-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    api_secret_hash VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    webhook_url TEXT,
    webhook_secret VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_products_api_key ON products(api_key);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_created_at ON products(created_at);
CREATE INDEX idx_products_deleted_at ON products(deleted_at) WHERE deleted_at IS NOT NULL;

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments
COMMENT ON TABLE products IS 'Multi-tenant products that integrate with the appointment service';
COMMENT ON COLUMN products.api_key IS 'Public API key for authentication';
COMMENT ON COLUMN products.api_secret_hash IS 'Hashed API secret (bcrypt)';
COMMENT ON COLUMN products.metadata IS 'Flexible JSON metadata for product-specific data';
COMMENT ON COLUMN products.deleted_at IS 'Soft delete timestamp';
```

**File: `migrations/000001_create_products_table.down.sql`**

```sql
-- Drop trigger
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_products_deleted_at;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_status;
DROP INDEX IF EXISTS idx_products_api_key;

-- Drop table
DROP TABLE IF EXISTS products;
```

### 4. Create Migration: Appointments Table

Run this command:

```bash
make migrate-create name=create_appointments_table
```

**File: `migrations/000002_create_appointments_table.up.sql`**

```sql
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
```

**File: `migrations/000002_create_appointments_table.down.sql`**

```sql
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
```

### 5. Create Migration: Appointment Participants Table

Run this command:

```bash
make migrate-create name=create_appointment_participants_table
```

**File: `migrations/000003_create_appointment_participants_table.up.sql`**

```sql
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
```

**File: `migrations/000003_create_appointment_participants_table.down.sql`**

```sql
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
```

### 6. Create Database Schema Documentation

Location: `appointment-service/docs/DATABASE_SCHEMA.md`

```markdown
# Database Schema Documentation

## Overview

The appointment service uses PostgreSQL 15+ with three core tables implementing a flexible, multi-tenant architecture.

## Database Design Principles

1. **Multi-Tenancy**: All appointments are scoped to a `product_id`
2. **Soft Deletes**: All tables support soft deletion via `deleted_at`
3. **Metadata-First**: JSONB columns for flexible data storage
4. **External Users**: No user table; users referenced by external IDs
5. **Flexible Participants**: Support for 1-on-1 and group appointments

---

## Tables

### 1. products

Stores registered products that use the appointment service.

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK, Default: gen_random_uuid() | Primary key |
| name | VARCHAR(255) | NOT NULL | Product display name |
| description | TEXT | | Product description |
| api_key | VARCHAR(255) | UNIQUE, NOT NULL | Public API key |
| api_secret_hash | VARCHAR(255) | NOT NULL | Bcrypt hash of API secret |
| status | VARCHAR(20) | NOT NULL, Default: 'active' | Product status |
| webhook_url | TEXT | | Webhook endpoint for notifications |
| webhook_secret | VARCHAR(255) | | Webhook signing secret |
| metadata | JSONB | Default: '{}' | Flexible metadata |
| created_at | TIMESTAMP | NOT NULL, Default: NOW() | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL, Default: NOW() | Last update timestamp |
| deleted_at | TIMESTAMP | | Soft delete timestamp |

**Indexes:**
- `idx_products_api_key` on `api_key`
- `idx_products_status` on `status`
- `idx_products_created_at` on `created_at`
- `idx_products_deleted_at` on `deleted_at` (partial)

**Example:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "CRM System",
  "api_key": "prod_abc123xyz",
  "status": "active",
  "metadata": {
    "plan": "enterprise",
    "features": ["webhooks", "analytics"]
  }
}
```

---

### 2. appointments

Core appointment records with multi-tenant support.

**Columns:**

| Column | Type | Constraints | Description |
| -------- | ------ | ------------- | ------------- |
| id | UUID | PK, Default: gen_random_uuid() | Primary key |
| product_id | UUID | FK(products), NOT NULL | Owning product |
| title | VARCHAR(255) | NOT NULL | Appointment title |
| description | TEXT | | Appointment description |
| start_time | TIMESTAMP | NOT NULL | Start time |
| end_time | TIMESTAMP | NOT NULL | End time |
| timezone | VARCHAR(50) | NOT NULL, Default: 'UTC' | Timezone |
| location | VARCHAR(500) | | Meeting location/URL |
| status | VARCHAR(20) | NOT NULL, Default: 'scheduled' | Appointment status |
| created_by | VARCHAR(255) | NOT NULL | External user ID of creator |
| metadata | JSONB | Default: '{}' | Flexible metadata |
| created_at | TIMESTAMP | NOT NULL, Default: NOW() | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL, Default: NOW() | Last update timestamp |
| deleted_at | TIMESTAMP | | Soft delete timestamp |

**Constraints:**

- `CHECK (end_time > start_time)` - End must be after start
- `CHECK (status IN ('scheduled', 'confirmed', 'cancelled', 'completed', 'no_show'))`

**Indexes:**

- Single column: `product_id`, `start_time`, `end_time`, `status`, `created_by`, `created_at`, `deleted_at`
- Composite: `(product_id, start_time, end_time)`, `(product_id, status)`, `(start_time, end_time)`
- JSONB: GIN index on `metadata`

**Example:**

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Product Demo",
  "start_time": "2026-02-15T14:00:00Z",
  "end_time": "2026-02-15T15:00:00Z",
  "timezone": "America/New_York",
  "status": "scheduled",
  "created_by": "user_123",
  "metadata": {
    "meeting_type": "demo",
    "video_url": "https://zoom.us/j/123456"
  }
}
```

---

### 3. appointment_participants

Tracks all participants in appointments with roles.

**Columns:**

| Column | Type | Constraints | Description |
| -------- | ------ | ------------- | ------------- |
| id | UUID | PK, Default: gen_random_uuid() | Primary key |
| appointment_id | UUID | FK(appointments), NOT NULL | Appointment reference |
| external_user_id | VARCHAR(255) | NOT NULL | User ID from external system |
| role | VARCHAR(20) | NOT NULL, Default: 'guest' | Participant role |
| status | VARCHAR(20) | NOT NULL, Default: 'pending' | Participation status |
| user_metadata | JSONB | Default: '{}' | User information |
| created_at | TIMESTAMP | NOT NULL, Default: NOW() | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL, Default: NOW() | Last update timestamp |

**Constraints:**

- `UNIQUE(appointment_id, external_user_id)` - One entry per user per appointment
- `CHECK (role IN ('host', 'guest', 'attendee', 'observer'))`
- `CHECK (status IN ('pending', 'accepted', 'declined', 'tentative'))`

**Indexes:**

- Single column: `appointment_id`, `external_user_id`, `role`, `status`
- Composite: `(appointment_id, role)`, `(external_user_id, status)`
- JSONB: GIN index on `user_metadata`

**Example:**

```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "appointment_id": "660e8400-e29b-41d4-a716-446655440000",
  "external_user_id": "user_123",
  "role": "host",
  "status": "accepted",
  "user_metadata": {
    "name": "John Doe",
    "email": "john@example.com",
    "avatar": "https://example.com/avatar.jpg"
  }
}
```

---

## Common Queries

### Get All Appointments for a User

```sql
SELECT DISTINCT a.*
FROM appointments a
INNER JOIN appointment_participants ap ON a.id = ap.appointment_id
WHERE ap.external_user_id = $1
  AND a.product_id = $2
  AND a.deleted_at IS NULL
ORDER BY a.start_time DESC;
```

### Get Appointments in Date Range

```sql
SELECT a.*
FROM appointments a
WHERE a.product_id = $1
  AND a.start_time >= $2
  AND a.end_time <= $3
  AND a.deleted_at IS NULL
ORDER BY a.start_time ASC;
```

### Get Participants for Appointment

```sql
SELECT *
FROM appointment_participants
WHERE appointment_id = $1
ORDER BY CASE role 
    WHEN 'host' THEN 1
    WHEN 'guest' THEN 2
    WHEN 'attendee' THEN 3
    ELSE 4
END, created_at ASC;
```

---

## Migration Commands

```bash
# Create new migration
make migrate-create name=your_migration_name

# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status

# Force specific version (emergency only)
make migrate-force version=3
```

---

## Backup & Restore

```bash
# Backup database
docker-compose exec postgres pg_dump -U appointments appointments_dev > backup.sql

# Restore database
docker-compose exec -T postgres psql -U appointments appointments_dev < backup.sql
```

```md
```

### 7. Run Migrations

```bash
# Start database if not running
make db-start

# Wait for PostgreSQL to be ready
sleep 5

# Run all migrations
make migrate-up

# Verify migrations
make migrate-status
```

### 8. Create Test Data Script (Optional)

Location: `appointment-service/scripts/seed_database.sql`

```sql
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

-- Insert test appointment
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

SELECT 'Database seeded successfully' AS result;
```

Add to Makefile:

```makefile
db-seed: ## Seed database with test data
 @echo "$(CYAN)Seeding database...$(NC)"
 docker-compose exec -T postgres psql -U appointments -d appointments_dev < scripts/seed_database.sql
 @echo "$(GREEN)Database seeded$(NC)"
```

---

## Verification Checklist

- [ ] golang-migrate CLI installed
- [ ] Makefile updated with migration commands
- [ ] Three migration files created (products, appointments, participants)
- [ ] Database schema documentation created
- [ ] Migrations run successfully: `make migrate-up`
- [ ] Migration status shows version 3: `make migrate-status`
- [ ] Can connect to database and see tables
- [ ] Seed script created (optional)
- [ ] Test data inserted (optional)

---

## Testing

```bash
# Start database
make db-start

# Create migrations (already done in steps above)
# Run migrations
make migrate-up

# Verify migration status
make migrate-status
# Expected: 3

# Connect to database and verify tables
make db-console

# In psql console:
\dt
# Should show: products, appointments, appointment_participants, schema_migrations

# Check products table
\d products

# Check appointments table
\d appointments

# Check participants table
\d appointment_participants

# Exit psql
\q

# Optional: Seed test data
make db-seed

# Verify test data
make db-console
SELECT COUNT(*) FROM products;
SELECT COUNT(*) FROM appointments;
SELECT COUNT(*) FROM appointment_participants;
\q
```

---

## Expected Output

After running `make migrate-up`:

```md
Running migrations up...
3/u create_products_table (15.234567ms)
3/u create_appointments_table (12.345678ms)
3/u create_appointment_participants_table (10.123456ms)
Migrations applied successfully
```

After `make migrate-status`:

```md
Migration status:
3
```

In database console (`\dt`):

```md
                    List of relations
 Schema |             Name              | Type  |    Owner     
--------+-------------------------------+-------+--------------
 public | appointments                  | table | appointments
 public | appointment_participants      | table | appointments
 public | products                      | table | appointments
 public | schema_migrations             | table | appointments
```

---

## Troubleshooting

**Issue**: `migrate: command not found`  
**Solution**: Install golang-migrate CLI (see Step 1)

**Issue**: Migration fails with "relation already exists"  
**Solution**: Either drop tables manually or use `make db-reset`

**Issue**: Cannot connect to database during migration  
**Solution**:

- Ensure database is running: `make db-start`
- Check .env file has correct database credentials
- Wait a few seconds for PostgreSQL to initialize

**Issue**: Migration version conflict  
**Solution**: Check current version with `make migrate-status`, then force correct version

---

## Next Task

After completing this task, proceed to [TASK_W02_03_DOMAIN_MODELS.md](TASK_W02_03_DOMAIN_MODELS.md) to create the Go domain models that map to these database tables.
