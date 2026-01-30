# Task 03: Database Setup and Migrations

**Priority**: High  
**Estimated Time**: 3 hours  
**Dependencies**: TASK_02  
**Status**: Not Started

---

## Objective

Set up PostgreSQL database, create migration files, and implement database connection pooling.

---

## Prerequisites

- [ ] Task 02 completed
- [ ] Docker and Docker Compose installed
- [ ] golang-migrate tool installed

---

## Steps

### 1. Install Migration Tool

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 2. Create Database Package

**File**: `pkg/database/postgres.go`

```go
package database

import (
 "context"
 "fmt"
 "time"

 "appointment-service/internal/config"
 "github.com/jackc/pgx/v5/pgxpool"
 "go.uber.org/zap"
)

func NewPostgresPool(cfg *config.Config, logger *zap.Logger) (*pgxpool.Pool, error) {
 poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
 if err != nil {
  return nil, fmt.Errorf("failed to parse database URL: %w", err)
 }

 // Configure connection pool
 poolConfig.MaxConns = int32(cfg.Database.MaxConnections)
 poolConfig.MinConns = int32(cfg.Database.MaxIdleConnections)
 poolConfig.MaxConnLifetime = cfg.Database.MaxLifetime
 poolConfig.MaxConnIdleTime = 30 * time.Minute
 poolConfig.HealthCheckPeriod = 1 * time.Minute

 ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
 defer cancel()

 pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
 if err != nil {
  return nil, fmt.Errorf("failed to create connection pool: %w", err)
 }

 // Test connection
 if err := pool.Ping(ctx); err != nil {
  return nil, fmt.Errorf("failed to ping database: %w", err)
 }

 logger.Info("Database connection established",
  zap.String("host", cfg.Database.Host),
  zap.String("database", cfg.Database.Name),
  zap.Int("max_connections", cfg.Database.MaxConnections),
 )

 return pool, nil
}

func ClosePool(pool *pgxpool.Pool, logger *zap.Logger) {
 if pool != nil {
  pool.Close()
  logger.Info("Database connection closed")
 }
}
```

### 3. Create Migration Files

#### Migration 001: Products Table

**File**: `migrations/000001_create_products_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    api_secret_hash VARCHAR(255) NOT NULL,
    callback_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_api_key ON products(api_key);
CREATE INDEX idx_products_is_active ON products(is_active);

-- Add updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**File**: `migrations/000001_create_products_table.down.sql`

```sql
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS products;
```

#### Migration 002: Appointments Table

**File**: `migrations/000002_create_appointments_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    
    -- Creator/Host of the appointment
    created_by VARCHAR(255) NOT NULL,
    
    -- Appointment Details
    title VARCHAR(500) NOT NULL,
    description TEXT,
    -- Appointment Details
    title VARCHAR(500) NOT NULL,
    description TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location VARCHAR(500),
    meeting_type VARCHAR(50),
    
    -- Status Management
    status VARCHAR(50) DEFAULT 'pending' NOT NULL,
    cancelled_by VARCHAR(255),
    cancellation_reason TEXT,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    
    -- Additional Data
    metadata JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_time_range CHECK (end_time > start_time),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed'))
);

-- Indexes for performance
CREATE INDEX idx_appointments_product_id ON appointments(product_id);
CREATE INDEX idx_appointments_created_by ON appointments(created_by);
CREATE INDEX idx_appointments_status ON appointments(status);
CREATE INDEX idx_appointments_start_time ON appointments(start_time);
CREATE INDEX idx_appointments_end_time ON appointments(end_time);
CREATE INDEX idx_appointments_time_range ON appointments(start_time, end_time);
CREATE INDEX idx_appointments_product_creator ON appointments(product_id, created_by);

-- Add updated_at trigger
CREATE TRIGGER update_appointments_updated_at BEFORE UPDATE ON appointments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**File**: `migrations/000002_create_appointments_table.down.sql`

```sql
DROP TRIGGER IF EXISTS update_appointments_updated_at ON appointments;
DROP TABLE IF EXISTS appointments;
```

#### Migration 003: Appointment Participants Table

**File**: `migrations/000003_create_appointment_participants_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS appointment_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    
    -- External user reference (from client product)
    external_user_id VARCHAR(255) NOT NULL,
    
    -- Participant role
    role VARCHAR(50) NOT NULL,
    
    -- User metadata from client product
    user_metadata JSONB NOT NULL,
    
    -- Participant status
    status VARCHAR(50) DEFAULT 'pending' NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_role CHECK (role IN ('host', 'guest', 'attendee', 'observer')),
    CONSTRAINT valid_participant_status CHECK (status IN ('pending', 'accepted', 'declined')),
    CONSTRAINT unique_participant UNIQUE (appointment_id, external_user_id)
);

-- Indexes for performance
CREATE INDEX idx_participants_appointment_id ON appointment_participants(appointment_id);
CREATE INDEX idx_participants_external_user_id ON appointment_participants(external_user_id);
CREATE INDEX idx_participants_role ON appointment_participants(role);
CREATE INDEX idx_participants_status ON appointment_participants(status);

-- Add updated_at trigger
CREATE TRIGGER update_participants_updated_at BEFORE UPDATE ON appointment_participants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**File**: `migrations/000003_create_appointment_participants_table.down.sql`

```sql
DROP TRIGGER IF EXISTS update_participants_updated_at ON appointment_participants;
DROP TABLE IF EXISTS appointment_participants;
```

#### Migration 004: Appointment History Table

**File**: `migrations/000004_create_appointment_history_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS appointment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    changed_by VARCHAR(255),
    changes JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_action CHECK (action IN ('created', 'updated', 'cancelled', 'confirmed', 'completed'))
);

CREATE INDEX idx_appointment_history_appointment_id ON appointment_history(appointment_id);
CREATE INDEX idx_appointment_history_action ON appointment_history(action);
CREATE INDEX idx_appointment_history_created_at ON appointment_history(created_at);
```

**File**: `migrations/000004_create_appointment_history_table.down.sql`

```sql
DROP TABLE IF EXISTS appointment_history;
```

### 4. Create docker-compose.yml for Database

**File**: `docker-compose.dev.yml`

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: appointment-db-dev
    environment:
      POSTGRES_USER: appointments
      POSTGRES_PASSWORD: dev_password_123
      POSTGRES_DB: appointments
    ports:
      - "5432:5432"
    volumes:
      - postgres_data_dev:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U appointments"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data_dev:
```

### 5. Update Makefile

Add to existing Makefile:

```makefile
# Database tasks
.PHONY: db-start db-stop db-migrate-up db-migrate-down db-create-migration db-reset

db-start:
 @echo "Starting database..."
 docker-compose -f docker-compose.dev.yml up -d

db-stop:
 @echo "Stopping database..."
 docker-compose -f docker-compose.dev.yml down

db-migrate-up:
 @echo "Running migrations..."
 migrate -path migrations -database "$(DB_URL)" up

db-migrate-down:
 @echo "Rolling back migrations..."
 migrate -path migrations -database "$(DB_URL)" down 1

db-create-migration:
 @read -p "Enter migration name: " name; \
 migrate create -ext sql -dir migrations -seq $$name

db-reset:
 @echo "Resetting database..."
 docker-compose -f docker-compose.dev.yml down -v
 docker-compose -f docker-compose.dev.yml up -d
 @echo "Waiting for database to be ready..."
 @sleep 5
 make db-migrate-up
```

---

## Acceptance Criteria

- [ ] Migration tool installed
- [ ] Database package created with connection pooling
- [ ] All migration files created (up and down)
- [ ] docker-compose.dev.yml created
- [ ] Database starts successfully
- [ ] Migrations run successfully
- [ ] Database schema matches design
- [ ] Indexes created for performance

---

## Verification

```bash
# Start database
make db-start

# Wait for database to be ready
docker-compose -f docker-compose.dev.yml logs postgres

# Run migrations
make db-migrate-up

# Verify tables exist
docker exec appointment-db-dev psql -U appointments -d appointments -c "\dt"

# Check table structure
docker exec appointment-db-dev psql -U appointments -d appointments -c "\d products"
docker exec appointment-db-dev psql -U appointments -d appointments -c "\d appointments"

# Test rollback
make db-migrate-down
make db-migrate-up
```

---

## Next Task

[TASK_04_MODELS_AND_REPOSITORY.md](TASK_04_MODELS_AND_REPOSITORY.md)

---

## Notes

- Always test migration rollbacks
- Keep migrations atomic (one change per migration)
- Use transactions where appropriate
- Document any complex migrations
- Backup production database before running migrations
