# Appointment Service — Product Brief

## 1. Product Overview

**Appointment Service** is a multi-tenant, microservice-based appointment scheduling backend built in Go. It operates as an **Appointment-as-a-Service (AaaS)** platform, enabling external products (SaaS applications, CRMs, healthcare systems, etc.) to integrate appointment booking functionality without building it from scratch.

**Core Problem Solved:** Appointment scheduling is a repetitive concern across many domains. This service provides a centralized, API-first scheduling backend that any product can embed, eliminating the need to build and maintain appointment systems individually.

**Key Differentiator:** The service **does not own users**—it relies entirely on external user identifiers (`external_user_id`) passed via JWT claims. This stateless design allows seamless integration without user data synchronization.

---

## 2. Key Features

- **Multi-Tenant Architecture** — Complete data isolation via `product_id`; each integrating product has isolated appointments, participants, and availability rules
- **JWT-Based Authentication** — Stateless tokens containing `product_id`, `external_user_id`, and `role` claims
- **Flexible Participant Model** — Support for 1-on-1 and group appointments with configurable participant roles (host, guest, attendee, observer)
- **Provider Availability Management** — Define working hours, appointment durations, slot intervals, and buffers per provider per day
- **Dynamic Slot Generation** — Calculate available booking slots based on provider rules and existing appointments
- **Concurrency-Safe Booking** — Uses PostgreSQL `REPEATABLE READ` isolation and row locking to prevent double-booking
- **Status Workflow** — Appointments transition through `scheduled → confirmed → completed` with cancel and no-show paths
- **Webhook Support** — Products can configure webhook URLs for event notifications
- **Kubernetes-Ready** — Health (`/health`), liveness (`/live`), and readiness (`/ready`) endpoints; graceful shutdown handling
- **Prometheus Metrics** — Built-in observability support

---

## 3. Architecture Summary

### Layers

| Layer | Location | Responsibility |
| ------- | ---------- | ---------------- |
| **HTTP Handlers** | `internal/handlers/` | Request parsing, response formatting, error mapping |
| **Middleware** | `internal/middleware/` | JWT validation, CORS, role-based authorization |
| **Service** | `internal/service/` | Business logic, validation, authorization rules |
| **Repository** | `internal/repository/` | Data access via raw SQL (pgx v5) |
| **Models** | `internal/models/` | Domain entities and validation |
| **Packages** | `pkg/` | Reusable utilities (auth, database, logger) |

### Technology Stack

- **Language:** Go 1.24+
- **HTTP Framework:** Gin
- **Database:** PostgreSQL 15+ with pgx v5 (raw SQL, no ORM)
- **Logging:** Zap (structured logging)
- **Authentication:** JWT (HS256) via `golang-jwt/jwt/v5`
- **Password Hashing:** bcrypt
- **Containerization:** Docker, Docker Compose
- **Monitoring:** Prometheus-ready

---

## 4. API Highlights

### Public Endpoints

| Method | Endpoint | Description |
| ----- ---|----------|-------------|
| GET | `/health` | Service health status |
| GET | `/live` | Kubernetes liveness probe |
| GET | `/ready` | Readiness probe with DB health |
| POST | `/v1/products/register` | Register a new product (returns API credentials) |
| POST | `/v1/auth/token` | Generate JWT using API key/secret |

### Appointments (JWT Required)

| Method | Endpoint | Description |
| -------- | ---------- | ------------- |
| POST | `/v1/appointments` | Create appointment with participants |
| GET | `/v1/appointments` | List user's appointments |
| GET | `/v1/appointments/:id` | Get appointment details |
| PATCH | `/v1/appointments/:id` | Update appointment (creator/admin/provider) |
| DELETE | `/v1/appointments/:id` | Soft-delete (admin only) |
| POST | `/v1/appointments/:id/cancel` | Cancel appointment |
| PATCH | `/v1/appointments/:id/response` | Confirm/complete (admin/provider) |
| POST | `/v1/appointments/:id/participants` | Add participant |
| DELETE | `/v1/appointments/:id/participants/:user_id` | Remove participant |
| PATCH | `/v1/appointments/:id/participants/:user_id/status` | Update participant RSVP |

### Availability (JWT Required)

| Method | Endpoint | Description |
| -------- | ---------- | ------------- |
| GET | `/v1/availability` | Get available slots for a provider on a date |
| POST | `/v1/appointments/book` | Book appointment (concurrency safe) |
| POST | `/v1/providers/:provider_id/availability` | Create availability rule |
| GET | `/v1/providers/:provider_id/availability` | List provider's rules |
| POST | `/v1/providers/:provider_id/availability/bulk` | Bulk create rules |
| PATCH | `/v1/providers/:provider_id/availability/:rule_id` | Update rule |
| DELETE | `/v1/providers/:provider_id/availability/:rule_id` | Delete rule |

### Products (JWT Required)

| Method | Endpoint | Description |
| -------- | ---------- | ------------- |
| GET | `/v1/products/me` | Get current product info |
| PATCH | `/v1/products/me` | Update current product |
| GET | `/v1/products` | List all products (admin only) |

---

## 5. Data Models

### Products

Represents an integrating application with API credentials (`api_key`, `api_secret_hash`) and optional webhook configuration. Supports `active`, `inactive`, `suspended` statuses.

### Appointments

Core scheduling entity scoped to a product. Contains title, description, time range, timezone, location, status, creator reference, provider reference, and flexible JSONB metadata.

### Appointment Participants

Links external users to appointments with roles (`host`, `guest`, `attendee`, `observer`) and RSVP status (`pending`, `accepted`, `declined`, `tentative`). Supports per-participant metadata.

### Provider Availability Rules

Defines provider working hours per day of week, including appointment duration, slot intervals, and buffer times. Enables dynamic slot calculation.

### Relationships

```md
Products (1) ──< Appointments (N) ──< Appointment Participants (N)
Products (1) ──< Provider Availability Rules (N)
```

---

## 6. Roles & Permissions

| Role | Capabilities |
| ------ | -------------- |
| **admin** | Full access: delete appointments, view all products, manage any appointment |
| **provider** | Confirm/complete appointments, manage own availability rules, update appointment statuses |
| **user** | Create appointments, view/update own appointments, manage own participation status |

Authorization is enforced via middleware (`RequireAdmin()`, `RequireProvider()`, `RequireAdminOrProvider()`) and service-layer checks.

---

## 7. Key Considerations

### Security

- **API Credentials:** Products authenticate via `api_key`/`api_secret`; secrets are bcrypt-hashed and never stored in plaintext
- **JWT Tokens:** 24-hour expiration; signed with HS256
- **Multi-Tenancy:** All queries filter by `product_id` to ensure data isolation
- **CORS:** Configurable allowed origins, methods, and headers

### Error Handling

- Consistent JSON error responses with `error`, `message`, and optional `details` fields
- Service errors mapped to appropriate HTTP status codes (400, 401, 403, 404, 409, 500)
- Soft deletes preserve audit trails

### Business Rules

- Appointments must have `end_time > start_time`
- Start time must be in the future (5-minute grace period)
- Creators are automatically added as participants with `accepted` status
- Booking uses database-level locking to prevent race conditions
- Provider availability rules enforce one rule per provider per day per product

### Database

- PostgreSQL exclusion constraints prevent overlapping appointments for the same provider
- JSONB columns for flexible metadata storage
- Comprehensive indexing for query performance
- Automatic `updated_at` timestamps via triggers

---

## *Generated from codebase analysis — March 2026*