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

## Entity Relationship Diagram

```md
┌─────────────────┐       ┌─────────────────────┐       ┌──────────────────────────┐
│    products     │       │    appointments     │       │ appointment_participants │
├─────────────────┤       ├─────────────────────┤       ├──────────────────────────┤
│ id (PK)         │◄──────│ product_id (FK)     │       │ id (PK)                  │
│ name            │       │ id (PK)             │◄──────│ appointment_id (FK)      │
│ description     │       │ title               │       │ external_user_id         │
│ api_key         │       │ description         │       │ role                     │
│ api_secret_hash │       │ start_time          │       │ status                   │
│ status          │       │ end_time            │       │ user_metadata            │
│ webhook_url     │       │ timezone            │       │ created_at               │
│ webhook_secret  │       │ location            │       │ updated_at               │
│ metadata        │       │ status              │       └──────────────────────────┘
│ created_at      │       │ created_by          │
│ updated_at      │       │ metadata            │
│ deleted_at      │       │ created_at          │
└─────────────────┘       │ updated_at          │
                          │ deleted_at          │
                          └─────────────────────┘
```

---

## Tables

### 1. products

Stores registered products that use the appointment service.

**Columns:**

| Column | Type | Constraints | Description |
| -------- | ------ | ------------- | ------------- |
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

**Status Values:**

| Status | Description |
| -------- | ------------- |
| scheduled | Initial state after creation |
| confirmed | Confirmed by all parties |
| cancelled | Appointment was cancelled |
| completed | Appointment has finished |
| no_show | Participant(s) did not attend |

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

**Role Values:**

| Role | Description |
| ------ | ------------- |
| host | The appointment organizer |
| guest | A primary participant |
| attendee | A regular attendee |
| observer | View-only participant |

**Status Values:**

| Status | Description |
| -------- | ------------- |
| pending | Awaiting response |
| accepted | Confirmed attendance |
| declined | Declined invitation |
| tentative | Tentatively accepted |

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

### Check for Scheduling Conflicts

```sql
SELECT a.id, a.title, a.start_time, a.end_time
FROM appointments a
INNER JOIN appointment_participants ap ON a.id = ap.appointment_id
WHERE ap.external_user_id = $1
  AND a.product_id = $2
  AND a.deleted_at IS NULL
  AND a.status NOT IN ('cancelled', 'completed')
  AND (
    ($3 >= a.start_time AND $3 < a.end_time)
    OR ($4 > a.start_time AND $4 <= a.end_time)
    OR ($3 <= a.start_time AND $4 >= a.end_time)
  );
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

# Rollback all migrations
make migrate-down-all

# Check migration status
make migrate-status

# Force specific version (emergency only)
make migrate-force version=3
```

---

## Database Commands

```bash
# Start database
make db-start

# Stop database
make db-stop

# Reset database (destroys all data)
make db-reset

# Open PostgreSQL console
make db-console

# View database logs
make db-logs

# Seed with test data
make db-seed
```

---

## Backup & Restore

```bash
# Backup database
docker-compose exec postgres pg_dump -U appointments appointments_dev > backup.sql

# Restore database
docker-compose exec -T postgres psql -U appointments appointments_dev < backup.sql
```

---

## Connection Information

| Setting | Development Value |
| -------- | ------------------ |
| Host | localhost |
| Port | 1998 |
| Database | appointments_dev |
| User | appointments |
| Password | secure_password |
| SSL Mode | disable |

**Connection String:**

```md
postgresql://appointments:secure_password@localhost:1998/appointments_dev?sslmode=disable
```
