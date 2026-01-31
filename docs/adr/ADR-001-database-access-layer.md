# ADR-001: Database Access Layer

**Status**: Accepted  
**Date**: 2026-01-31  
**Deciders**: Solo Developer + AI Agent  
**Technical Story**: Multi-tenant microservice requires efficient database access

## Context

We need to choose a database access pattern for our Go microservice. The main options are:
1. ORM (GORM) - High-level abstraction
2. Raw SQL (pgx) - Direct database access
3. SQL Builder (sqlx) - Middle ground

Our requirements:
- Multi-tenant queries with product_id filtering on every query
- Complex joins for participants table
- JSONB metadata queries
- High performance for availability checks
- Fine-grained control over query optimization

## Decision

We will use **pgx v5 with raw SQL** queries, implementing the Repository pattern for abstraction.

## Rationale

1. **Performance**: pgx is 20-30% faster than GORM for our use cases
2. **JSONB Support**: Better control over JSONB queries and indexing
3. **Complex Queries**: Participants pattern requires optimized JOIN queries
4. **Transparency**: Full visibility into SQL execution and query plans
5. **Connection Pooling**: pgxpool provides excellent connection management
6. **Learning**: Forces better understanding of SQL and database design

## Consequences

### Positive
- Maximum query performance
- Full control over SQL optimization
- Better understanding of database operations
- Easier to add database indexes strategically
- No ORM magic or hidden N+1 queries

### Negative
- More boilerplate code to write
- Need to manually write SQL queries
- Query strings must be maintained carefully
- Steeper learning curve for developers unfamiliar with SQL

## Alternatives Considered

### Alternative 1: GORM
**Pros**: Easy to use, less boilerplate, automatic migrations  
**Cons**: Slower performance (20-30%), less control over JSONB, hidden N+1 queries  
**Why Not**: Performance is critical for availability calculations

### Alternative 2: sqlx
**Pros**: Middle ground between GORM and pgx  
**Cons**: Less features than pgx, not as fast  
**Why Not**: If we're writing SQL anyway, pgx provides more features

## Implementation

```go
// Repository interface example
type AppointmentRepository interface {
    Create(ctx context.Context, appointment *Appointment) error
    GetByID(ctx context.Context, productID uuid.UUID, id uuid.UUID) (*Appointment, error)
    GetByUserID(ctx context.Context, productID uuid.UUID, userID string) ([]Appointment, error)
}

// pgx implementation with connection pool
type pgxAppointmentRepository struct {
    pool *pgxpool.Pool
}
```

## References

- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [MIGRATION_STRATEGY_AND_ARCHITECTURE.md](../../docs/MIGRATION_STRATEGY_AND_ARCHITECTURE.md)
