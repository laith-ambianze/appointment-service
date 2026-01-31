# Appointment as a Service (AaaS)

A centralized appointment scheduling system built in Go that integrates with multiple products via APIs.

## 📋 Project Overview

This is a standalone backend service that:

- ✅ Owns appointments
- ✅ Integrates with multiple products via APIs
- ✅ Uses one shared database
- ❌ Does not own users (products maintain their own users)

## 🏗️ Architecture

```md
[ Product A ] ─┐
[ Product B ] ─┼──> Appointment API (Go)
[ Product C ] ─┘         |
                       Database
```

## 📚 Documentation

- **[APPOINTMENT_AS_A_SERVICE.md](APPOINTMENT_AS_A_SERVICE.md)** - Primary architecture document with AaaS approach
- **[PROJECT_PLANNING.md](PROJECT_PLANNING.md)** - Comprehensive planning with detailed infrastructure choices
- **[ARCHITECTURE_COMPARISON.md](ARCHITECTURE_COMPARISON.md)** - Detailed comparison of different architectural approaches

## 🎯 Recommended Approach

**Start with the AaaS (Appointment as a Service) model:**

- 40-50% faster MVP (4-6 weeks vs 8 weeks)
- 60% lower infrastructure costs
- Simpler architecture (no sync service)
- Better horizontal scalability
- No GDPR compliance burden
- Easier to evolve over time

## 🚀 Key Features (Planned)

### Phase 1 (MVP)

- Product authentication
- Create / list / cancel appointments
- Metadata support
- Time conflict validation
- Dockerized Go service

### Phase 2

- Availability management
- Participants support
- Webhooks

## 🛠️ Technology Stack

- **Language:** Go
- **Database:** PostgreSQL (JSONB for metadata)
- **API:** RESTful
- **Authentication:** Product-level API keys / JWT
- **Container:** Docker

## 📊 Core Data Model

```sql
products
- id (uuid)
- name
- status
- created_at

appointments
- id (uuid)
- product_id (fk)
- external_user_id
- start_time
- end_time
- status
- metadata (jsonb)
- created_at
```

## 🔑 Key Design Decisions

1. **No User Storage** - Store only `external_user_id` references
2. **Metadata-First** - Use JSONB for all product-specific data
3. **Product-Scoped Auth** - Every request requires product identification
4. **Stateless API** - No session storage, JWT-based auth

## 📈 Success Metrics

- API Latency: < 100ms p95
- Uptime: > 99.9%
- Concurrent Products: Support 10+ products initially
- Throughput: 1000+ appointments/day per product

## 🏁 Getting Started

Coming soon - Implementation in progress

## 📝 License

To be determined

## 👥 Contributing

Guidelines to be added

---

**Status:** Design Phase - Ready for Implementation  
**Created:** January 24, 2026  
**Next Action:** Initialize Go project and database schema
