# Appointment as a Service (AaaS)

**Date:** January 24, 2026  
**Approach:** Centralized appointment service in Go that can be shared by multiple products

---

## 1. High-level Concept

You are building: **Appointment as a Service (AaaS)**

A standalone backend that:

- ✅ Owns appointments
- ❌ Does not own users
- ✅ Integrates with multiple products via APIs
- ✅ Uses one shared database

Each product:

- Has its own users
- Sends user metadata to the appointment service
- Uses the appointment API to create / manage bookings

---

## 2. Core Architecture

### System View

```md
[ Product A ] ─┐
[ Product B ] ─┼──> Appointment API (Go)
[ Product C ] ─┘         |
                       Database
```

### Key Principles

- **Single source of truth for appointments**
- **Products are clients, not tenants inside your DB**
- **Appointment service is stateless**
- **All cross-product context comes via metadata**

---

## 3. Multi-Product Strategy (Very Important)

You need product isolation without multiple databases.

### Product Identification

Every request must include:

- `product_id`
- `product_api_key` or JWT

**Example header:**

```http
X-Product-ID: crm
Authorization: Bearer <token>
```

This enables:

- ✅ Data isolation
- ✅ Rate limiting per product
- ✅ Future billing / analytics

---

## 4. Data Model (Core Tables)

### 1️⃣ products

```sql
products
- id (uuid)
- name
- status
- created_at
```

### 2️⃣ appointments

```sql
appointments
- id (uuid)
- product_id (fk)
- external_user_id
- start_time
- end_time
- status (pending, confirmed, cancelled)
- metadata (jsonb)
- created_at
```

**Why `external_user_id`?**

- You never store users
- You reference them by ID from the product

### 3️⃣ appointment_participants (optional)

```sql
appointment_participants
- id
- appointment_id
- role (host, guest)
- external_user_id
- metadata (jsonb)
```

### 4️⃣ availability (optional / advanced)

```sql
availability
- id
- product_id
- external_user_id
- day_of_week
- start_time
- end_time
```

---

## 5. Metadata-First Design (Smart Choice)

Since users come from other products:

```json
{
  "external_user_id": "user_123",
  "metadata": {
    "name": "Ahmed",
    "phone": "+9627...",
    "service": "consultation",
    "language": "ar"
  }
}
```

### Benefits

- ✅ No coupling to product schemas
- ✅ No migrations when products evolve
- ✅ Works for CRM, ERP, medical, education, etc.

---

## 6. API Design (REST – Clean & Stable)

### Create Appointment

```http
POST /api/v1/appointments
```

```json
{
  "external_user_id": "user_123",
  "start_time": "2026-02-01T14:00:00Z",
  "end_time": "2026-02-01T15:00:00Z",
  "metadata": {
    "service": "demo",
    "notes": "first visit"
  }
}
```

### List Appointments (per product)

```http
GET /api/v1/appointments?from=&to=&external_user_id=
```

### Cancel Appointment

```http
PATCH /api/v1/appointments/{id}/cancel
```

---

## 7. Go Project Structure (Production-Grade)

```md
/appointment-service
│
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── routes.go
│   │
│   ├── domain/
│   │   ├── appointment.go
│   │   ├── product.go
│   │
│   ├── service/
│   │   ├── appointment_service.go
│   │
│   ├── repository/
│   │   ├── appointment_repo.go
│   │   └── product_repo.go
│
│   ├── auth/
│   │   └── product_auth.go
│
│   └── infrastructure/
│       ├── db/
│       └── logger/
│
├── migrations/
├── go.mod
└── docker-compose.yml
```

---

## 8. Authentication Model (Critical)

### Recommended

#### Product-level authentication

Each product gets:

- `product_id`
- `api_key` or JWT secret

### Middleware

1. Validate token
2. Inject `product_id` into context
3. Every DB query scoped by `product_id`

---

## 9. Database Choice

### ✅ PostgreSQL

**Why?**

- JSONB for metadata
- Strong constraints
- Time-based queries
- Indexing on `(product_id, start_time)`

---

## 10. Future-Proof Features (Plan Early)

You don't implement now, but design for them:

- ⏰ **Conflict detection**
- 🔁 **Recurring appointments**
- 🔔 **Webhooks per product**
- 🌍 **Timezone support**
- 📊 **Product-level analytics**
- 💳 **Paid bookings** (Stripe later)

---

## 11. MVP Scope (What to Build First)

### Phase 1 (Strong MVP)

- ✅ Product auth
- ✅ Create / list / cancel appointments
- ✅ Metadata support
- ✅ Time conflict validation
- ✅ Dockerized Go service

### Phase 2

- ✅ Availability
- ✅ Participants
- ✅ Webhooks

---

## 12. Why This Fits You (Based on Your Background)

### Go → perfect for API + concurrency

You already built:

- Appointment systems
- Query builders
- ERP-style platforms

This can become:

- ✅ A reusable internal service
- ✅ Or a standalone SaaS later

---

## Implementation Checklist

### Setup Phase

- [ ] Initialize Go module
- [ ] Set up PostgreSQL database
- [ ] Create Docker Compose configuration
- [ ] Set up migration tool (golang-migrate)

### Core Features

- [ ] Implement product authentication middleware
- [ ] Create appointment domain model
- [ ] Build appointment repository (CRUD)
- [ ] Implement appointment service layer
- [ ] Create REST API handlers
- [ ] Add metadata validation
- [ ] Implement conflict detection

### API Endpoints

- [ ] `POST /api/v1/appointments` - Create
- [ ] `GET /api/v1/appointments` - List
- [ ] `GET /api/v1/appointments/{id}` - Get one
- [ ] `PATCH /api/v1/appointments/{id}` - Update
- [ ] `PATCH /api/v1/appointments/{id}/cancel` - Cancel
- [ ] `DELETE /api/v1/appointments/{id}` - Delete

### Security & Operations

- [ ] Product-level API key management
- [ ] Request validation
- [ ] Error handling
- [ ] Structured logging
- [ ] Health check endpoint
- [ ] Database connection pooling

### Testing

- [ ] Unit tests for services
- [ ] Integration tests for repositories
- [ ] API endpoint tests
- [ ] Load testing

### Documentation

- [ ] API documentation (Swagger/OpenAPI)
- [ ] Setup instructions
- [ ] Product integration guide
- [ ] Authentication guide

---

## Key Design Decisions

### 1. No User Storage

- **Decision:** Store only `external_user_id` references
- **Rationale:** Products own their users; we just track appointments
- **Impact:** Simpler data model, no GDPR compliance burden on our side

### 2. Metadata-First Approach

- **Decision:** Use JSONB for all product-specific data
- **Rationale:** Maximum flexibility across different product types
- **Impact:** No schema changes needed when products evolve

### 3. Product-Scoped Authentication

- **Decision:** Every request requires product identification
- **Rationale:** Multi-tenancy without database sharding
- **Impact:** Built-in isolation, easier to scale per product

### 4. Stateless API

- **Decision:** No session storage, JWT-based auth
- **Rationale:** Horizontal scaling without shared state
- **Impact:** Can deploy multiple instances behind load balancer

---

## Success Metrics

- **API Latency:** < 100ms p95
- **Uptime:** > 99.9%
- **Concurrent Products:** Support 10+ products initially
- **Throughput:** 1000+ appointments/day per product
- **Conflict Detection:** < 1% false positives

---

## Next Steps

1. **Initialize Project**
   - Create Go module
   - Set up basic project structure
   - Configure database connection

2. **Implement Core Domain**
   - Product model
   - Appointment model
   - Repository interfaces

3. **Build MVP API**
   - Authentication middleware
   - Create appointment endpoint
   - List appointments endpoint
   - Cancel appointment endpoint

4. **Testing & Documentation**
   - Write tests
   - Generate API docs
   - Create integration guide

5. **Deploy & Iterate**
   - Dockerize application
   - Deploy to staging
   - Onboard first product
   - Gather feedback

---

**Status:** Design Document - Ready for Implementation  
**Next Action:** Initialize Go project and database schema
