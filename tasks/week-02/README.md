# Week 02: Database & Core Models - Task Index

## Overview

Week 02 focuses on database setup, schema implementation, and core domain models. This builds on the foundation from Week 01.

**Total Estimated Time**: 14-18 hours  
**Goal**: Complete database setup with working migrations and core domain models

---

## Week 01 Completion Analysis

### ✅ What Was Completed

1. **Project Structure** (W01-01)
   - Git repository initialized
   - Complete Go project structure
   - Directory structure created
   - README.md with badges
   - .gitignore configured

2. **Configuration Files** (W01-02)
   - Makefile with all commands
   - Dockerfile for containerization
   - docker-compose.yml for local development
   - .env.example template
   - .golangci.yml for linting
   - .air.toml for hot reload

3. **Architecture Documentation** (W01-03)
   - 5 ADR documents created:
     - ADR-001: Database Access Layer (pgx chosen)
     - ADR-002: API Framework (Gin chosen)
     - ADR-003: Logging (Zap chosen)
     - ADR-004: Configuration (godotenv chosen)
     - ADR-005: Testing (testify + Go testing)
   - ADR template created

4. **CI/CD Pipeline** (W01-04)
   - GitHub Actions workflows:
     - ci.yml (lint, test, build, docker)
     - security.yml (gosec, trivy)
     - dependencies.yml (auto-updates)
     - release.yml (version tagging)
     - coverage.yml (codecov)
   - Branch protection configured

5. **Initial Code** (W01-05)
   - Configuration package (internal/config)
   - Logger package (pkg/logger)
   - Health handler (internal/handlers/health.go)
   - Working HTTP server (cmd/api/main.go)
   - Graceful shutdown
   - Request logging middleware

### ⚠️ What's Missing (Not Critical)

- Database connection not yet implemented
- No migrations created yet
- No models/repositories yet
- No authentication/authorization yet

### 🎯 Week 01 Success Metrics

- ✅ Application starts successfully
- ✅ Health endpoint responds (<http://localhost:8081/health>)
- ✅ Docker Compose configuration ready
- ✅ CI pipeline passing
- ✅ Documentation complete

---

## Week 02 Focus Areas

### Database Layer

- PostgreSQL connection setup
- Migration system implementation
- Core schema creation (products, appointments, participants)

### Domain Models

- Product model with validation
- Appointment model with participants
- Repository pattern implementation

### Database Package

- Connection pool management
- Transaction support
- Query helper utilities

---

## Task List

| Task | Name | Est. Time | Status |
| ------ | ------ | ----------- | -------- |
| [W02-01](TASK_W02_01_DATABASE_PACKAGE.md) | Database Connection Package | 3-4 hours | ⏸️ Not Started |
| [W02-02](TASK_W02_02_MIGRATIONS.md) | Database Migrations Setup | 3-4 hours | ⏸️ Not Started |
| [W02-03](TASK_W02_03_DOMAIN_MODELS.md) | Domain Models & Entities | 2-3 hours | ⏸️ Not Started |
| [W02-04](TASK_W02_04_PRODUCT_REPOSITORY.md) | Product Repository Implementation | 3-4 hours | ⏸️ Not Started |
| [W02-05](TASK_W02_05_APPOINTMENT_REPOSITORY.md) | Appointment Repository Implementation | 3-4 hours | ⏸️ Not Started |

---

## Execution Order

These tasks must be completed in sequence:

1. **TASK_W02_01** - Creates database connection package with pgx pool
2. **TASK_W02_02** - Sets up migration system and creates initial schema
3. **TASK_W02_03** - Defines domain models and entities
4. **TASK_W02_04** - Implements product repository with CRUD operations
5. **TASK_W02_05** - Implements appointment repository with complex queries

---

## What You'll Have After Week 02

### ✅ Database Infrastructure

- PostgreSQL connection pool with pgx
- Migration system with golang-migrate
- Database initialization in main.go
- Connection health checks
- Transaction support utilities

### ✅ Database Schema

- `products` table - Multi-tenant product registry
- `appointments` table - Core appointment storage
- `appointment_participants` table - Flexible participant tracking
- Proper indexes for performance
- Foreign key constraints
- JSONB metadata support

### ✅ Domain Models

- Product model with validation
- Appointment model with status enum
- AppointmentParticipant model with roles
- Proper UUID primary keys
- Timestamp tracking (created_at, updated_at)

### ✅ Repository Layer

- ProductRepository interface & implementation
  - Create, Update, Delete products
  - Find by ID, Find by API key
  - List all products
- AppointmentRepository interface & implementation
  - Create appointments with participants
  - Find appointments by various criteria
  - Update appointment status
  - Complex queries (user appointments, date range, etc.)

### ✅ Testing

- Database integration tests
- Repository unit tests with mocks
- Test fixtures and helpers

---

## Success Criteria

Week 02 is complete when:

- [ ] Database connection works (`make migrate-up` succeeds)
- [ ] All migrations apply successfully
- [ ] Models compile without errors
- [ ] Repository tests pass (`make test`)
- [ ] Can create a product via repository
- [ ] Can create an appointment with participants
- [ ] Docker Compose includes PostgreSQL
- [ ] CI pipeline includes database tests

---

## Quick Start

```bash
# Start from where Week 01 left off
cd appointment-service

# Start database
make docker-up

# Create first migration
make migrate-create name=create_initial_schema

# Run migrations
make migrate-up

# Run tests with database
make test

# Check migration status
make migrate-status
```

---

## Database Schema Preview

### Products Table

```sql
CREATE TABLE products (
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
```

### Appointments Table

```sql
CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
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
    deleted_at TIMESTAMP
);
```

### Appointment Participants Table

```sql
CREATE TABLE appointment_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    external_user_id VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    user_metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(appointment_id, external_user_id)
);
```

---

## Commands Reference

### Database

```bash
make migrate-create name=xxx   # Create new migration
make migrate-up                 # Apply migrations
make migrate-down               # Rollback one migration
make migrate-status             # Show migration status
make migrate-force version=1    # Force version (use with caution)
```

### Testing

```bash
make test                       # Run all tests
make test-integration          # Run integration tests only
make test-unit                 # Run unit tests only
make test-coverage             # Generate coverage report
```

### Database Management

```bash
make db-reset                  # Drop and recreate database
make db-seed                   # Seed with test data
make db-console                # Open psql console
```

---

## Technology Stack (Same as Week 01)

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15+
- **Database Driver**: pgx v5
- **Migrations**: golang-migrate
- **Testing**: testify + Go testing
- **UUID**: google/uuid

---

## File Structure After Week 02

```md
appointment-service/
├── cmd/api/
│   └── main.go                # ✅ Now with DB connection
├── internal/
│   ├── config/
│   │   └── config.go         # ✅ Already exists
│   ├── models/               # 🆕 Domain models
│   │   ├── product.go
│   │   ├── appointment.go
│   │   └── participant.go
│   ├── repository/           # 🆕 Data access layer
│   │   ├── product_repo.go
│   │   ├── appointment_repo.go
│   │   └── repository.go     # Common interfaces
│   └── handlers/
│       └── health.go         # ✅ Already exists
├── pkg/
│   ├── database/             # 🆕 DB connection package
│   │   ├── postgres.go
│   │   ├── transaction.go
│   │   └── health.go
│   └── logger/
│       └── logger.go         # ✅ Already exists
├── migrations/               # 🆕 SQL migrations
│   ├── 000001_create_products.up.sql
│   ├── 000001_create_products.down.sql
│   ├── 000002_create_appointments.up.sql
│   ├── 000002_create_appointments.down.sql
│   ├── 000003_create_participants.up.sql
│   └── 000003_create_participants.down.sql
├── tests/
│   ├── integration/          # 🆕 DB integration tests
│   │   ├── product_repo_test.go
│   │   └── appointment_repo_test.go
│   └── fixtures/             # 🆕 Test data
│       └── test_data.go
├── docker-compose.yml        # ✅ Updated with PostgreSQL
└── Makefile                  # ✅ Updated with DB commands
```

---

## Prerequisites for Week 02

Before starting Week 02 tasks, ensure:

1. ✅ Week 01 tasks are complete
2. ✅ Application runs: `make run`
3. ✅ Health endpoint works: `curl http://localhost:8081/health`
4. ✅ Docker Compose file exists
5. ✅ Makefile has basic commands
6. 🔧 PostgreSQL 15+ installed (local) OR Docker available
7. 🔧 `golang-migrate` CLI tool installed

### Install golang-migrate

**Windows (PowerShell):**

```powershell
# Using Scoop
scoop install migrate

# Or download binary from GitHub releases
# https://github.com/golang-migrate/migrate/releases
```

**macOS:**

```bash
brew install golang-migrate
```

**Linux:**

```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
```

**Verify installation:**

```bash
migrate -version
```

---

## Next Steps After Week 02

After completing Week 02, you'll be ready for:

- **Week 03**: Product Management API
  - Product registration endpoints
  - API key authentication
  - Product CRUD operations
  
- **Week 04**: Appointment API (Core)
  - Create appointment endpoint
  - List appointments endpoint
  - Update/Cancel appointment endpoints

- **Week 05**: Advanced Queries & Features
  - Availability checking
  - Conflict detection
  - Timezone handling
  - Webhook notifications

---

## Support & Troubleshooting

### Common Issues

**Issue**: `migrate: command not found`  
**Solution**: Install golang-migrate CLI (see Prerequisites)

**Issue**: Database connection refused  
**Solution**: Check PostgreSQL is running: `docker-compose ps`

**Issue**: Migration fails with syntax error  
**Solution**: Check SQL syntax in migration file, ensure PostgreSQL 15+ features

**Issue**: Tests fail with "database does not exist"  
**Solution**: Run `make migrate-up` before running tests

---

## Learning Resources

- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [golang-migrate Guide](https://github.com/golang-migrate/migrate)
- [PostgreSQL JSONB](https://www.postgresql.org/docs/current/datatype-json.html)
- [Repository Pattern in Go](https://threedots.tech/post/repository-pattern-in-go/)

---

**Ready to start?** Begin with [TASK_W02_01_DATABASE_PACKAGE.md](TASK_W02_01_DATABASE_PACKAGE.md)
