# Task W01-03: Architecture Decision Records (ADR)

**Status**: Not Started  
**Estimated Time**: 1-2 hours  
**Prerequisites**: TASK_W01_02_CONFIGURATION_FILES.md  
**Next Task**: TASK_W01_04_CICD_PIPELINE.md

---

## Objective

Document key architectural decisions using Architecture Decision Records (ADR) to provide context for future development and maintain consistency.

---

## What is an ADR?

An ADR documents a significant architectural decision along with its context and consequences. It helps future developers understand why certain choices were made.

---

## Steps

### 1. Create ADR Template

Location: `appointment-service/docs/adr/TEMPLATE.md`

```markdown
# ADR-XXX: [Title]

**Status**: [Proposed | Accepted | Deprecated | Superseded]  
**Date**: YYYY-MM-DD  
**Deciders**: [List of people involved]  
**Technical Story**: [Link to issue/story if applicable]

## Context

[Describe the context and problem statement. What is the issue we're trying to solve?]

## Decision

[Describe the decision that was made. What are we going to do?]

## Rationale

[Explain why this decision was made. What are the reasons?]

1. Reason 1
2. Reason 2
3. Reason 3

## Consequences

### Positive
- Benefit 1
- Benefit 2

### Negative
- Trade-off 1
- Trade-off 2

## Alternatives Considered

### Alternative 1: [Name]
[Description and why it wasn't chosen]

### Alternative 2: [Name]
[Description and why it wasn't chosen]

## References

- [Link to documentation]
- [Link to discussion]
```

### 2. ADR-001: Database Access Layer

Location: `appointment-service/docs/adr/ADR-001-database-access-layer.md`

```markdown
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
- [MIGRATION_STRATEGY_AND_ARCHITECTURE.md](../../MIGRATION_STRATEGY_AND_ARCHITECTURE.md)

```md

### 3. ADR-002: API Framework

Location: `appointment-service/docs/adr/ADR-002-api-framework.md`

```markdown
# ADR-002: API Framework

**Status**: Accepted  
**Date**: 2026-01-31  
**Deciders**: Solo Developer + AI Agent

## Context

We need to choose an HTTP framework for building REST APIs in Go. Main options:
1. Gin - Popular, feature-rich, good performance
2. Fiber - Fastest, Express.js-like API
3. Echo - Lightweight, minimalist
4. net/http - Standard library only

## Decision

We will use **Gin** for our HTTP API framework.

## Rationale

1. **Mature & Popular**: Large community, extensive documentation
2. **Performance**: Very fast (after Fiber), good enough for our needs
3. **Middleware Ecosystem**: Rich middleware support (CORS, auth, logging)
4. **JSON Handling**: Excellent JSON serialization performance
5. **Request Validation**: Built-in binding and validation
6. **Swagger Integration**: Works well with swag for API documentation

## Consequences

### Positive
- Fast development with clean API
- Extensive middleware available
- Good documentation and examples
- Easy testing with httptest
- Active community support

### Negative
- Slightly slower than Fiber
- More dependencies than standard library
- Not part of Go standard library

## Alternatives Considered

### Alternative 1: Fiber
**Pros**: Fastest framework, Express.js-like  
**Cons**: Non-standard context, different patterns  
**Why Not**: Gin provides sufficient performance with better ecosystem

### Alternative 2: net/http (Standard Library)
**Pros**: No dependencies, standard  
**Cons**: Too low-level, more boilerplate  
**Why Not**: Would require building too much infrastructure

## Implementation

```go
// Basic server setup
router := gin.Default()

// Middleware
router.Use(middleware.Logger())
router.Use(middleware.Auth())

// Routes
v1 := router.Group("/v1")
{
    v1.POST("/appointments", handlers.CreateAppointment)
    v1.GET("/appointments/:id", handlers.GetAppointment)
}

// Start server
router.Run(":8080")
```

## References

- [Gin Documentation](https://gin-gonic.com/docs/)
- [Go Web Framework Benchmarks](https://github.com/smallnest/go-web-framework-benchmark)

```

### 4. ADR-003: Logging Strategy

Location: `appointment-service/docs/adr/ADR-003-logging.md`

```markdown
# ADR-003: Logging Strategy

**Status**: Accepted  
**Date**: 2026-01-31  
**Deciders**: Solo Developer + AI Agent

## Context

We need a logging strategy that supports:
- Structured logging for easy parsing
- Multiple log levels (debug, info, warn, error)
- Production-ready performance
- Integration with monitoring tools
- Context propagation

## Decision

We will use **Zap** for structured logging with JSON output format.

## Rationale

1. **Performance**: Fastest Go logging library (zero allocation in production)
2. **Structured**: JSON output for easy parsing by log aggregators
3. **Type-Safe**: Strongly-typed fields prevent errors
4. **Flexible**: Development (console) and production (JSON) modes
5. **Context Support**: Easy to add contextual fields (request ID, user ID)

## Consequences

### Positive
- High performance logging
- Easy integration with ELK/Loki
- Structured data for better analysis
- Type-safe logging reduces bugs

### Negative
- More verbose than simple logging
- Requires learning Zap API
- JSON logs less readable in development

## Implementation

```go
// Initialize logger
logger, _ := zap.NewProduction() // JSON output
// or
logger, _ := zap.NewDevelopment() // Console output

// Usage
logger.Info("appointment created",
    zap.String("appointment_id", id),
    zap.String("product_id", productID),
    zap.String("user_id", userID),
)

// With context
logger.With(
    zap.String("request_id", reqID),
).Error("failed to create appointment", zap.Error(err))
```

## Alternatives Considered

### Alternative 1: logrus
**Pros**: Popular, easy to use  
**Cons**: Slower than Zap, uses reflection  
**Why Not**: Performance matters for high-throughput service

### Alternative 2: zerolog
**Pros**: Fast, zero allocation  
**Cons**: Less mature, smaller ecosystem  
**Why Not**: Zap has better community support

## References

- [Zap Documentation](https://pkg.go.dev/go.uber.org/zap)
- [Logging Best Practices](https://dave.cheney.net/2015/11/05/lets-talk-about-logging)

```md

### 5. ADR-004: Configuration Management

Location: `appointment-service/docs/adr/ADR-004-configuration.md`

```markdown
# ADR-004: Configuration Management

**Status**: Accepted  
**Date**: 2026-01-31  
**Deciders**: Solo Developer + AI Agent

## Context

We need a way to manage configuration across different environments:
- Local development
- Docker containers
- Kubernetes production

Requirements:
- Environment variables support
- Default values
- Type-safe configuration
- Easy to test

## Decision

We will use **godotenv** for loading .env files and a custom Config struct for type-safe access.

## Rationale

1. **Simple**: godotenv is straightforward and works well
2. **12-Factor App**: Follows environment variable pattern
3. **Docker-Friendly**: Environment variables work naturally in containers
4. **No Dependencies**: Minimal external dependencies
5. **Type-Safe**: Custom struct provides compile-time safety

## Implementation

```go
// config/config.go
type Config struct {
    Env      string
    APIPort  string
    APIHost  string
    
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    
    RedisHost string
    RedisPort string
    
    LogLevel  string
    LogFormat string
}

func Load() (*Config, error) {
    // Load .env file if it exists
    godotenv.Load()
    
    return &Config{
        Env:        getEnv("GO_ENV", "development"),
        APIPort:    getEnv("API_PORT", "8080"),
        DBHost:     getEnv("DB_HOST", "localhost"),
        // ... more fields
    }, nil
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

## Consequences

### Positive

- Simple and maintainable
- Works in all environments
- Type-safe configuration access
- Easy to mock in tests

### Negative

- No validation out of the box
- Need to manually parse complex types
- No hot-reload support

## Alternatives Considered

### Alternative 1: Viper
**Pros**: Feature-rich, supports many formats  
**Cons**: Overkill for our needs, more dependencies  
**Why Not**: godotenv + struct is simpler

### Alternative 2: envconfig
**Pros**: Struct tag-based config  
**Cons**: Less control, magic through reflection  
**Why Not**: Custom solution is more explicit

## References

- [12-Factor App Config](https://12factor.net/config)
- [godotenv](https://github.com/joho/godotenv)

```md

### 6. ADR-005: Testing Strategy

Location: `appointment-service/docs/adr/ADR-005-testing.md`

```markdown
# ADR-005: Testing Strategy

**Status**: Accepted  
**Date**: 2026-01-31  
**Deciders**: Solo Developer + AI Agent

## Context

We need a comprehensive testing strategy covering:
- Unit tests for business logic
- Integration tests for database operations
- API endpoint tests
- Mocking dependencies

## Decision

We will use **Go's built-in testing package** with **testify** for assertions and mocking.

## Rationale

1. **Standard**: Built-in testing is idiomatic Go
2. **Simple**: No framework complexity
3. **Testify**: Provides assertions and mocks without being a full framework
4. **Fast**: Tests run quickly with go test
5. **Coverage**: Easy to measure with go tool cover

## Testing Layers

### 1. Unit Tests
- Test business logic in service layer
- Mock repository dependencies
- Fast execution (<1s total)

### 2. Integration Tests
- Test repository with real database (testcontainers)
- Test full HTTP handlers
- Run in CI/CD pipeline

### 3. API Tests
- Test full request/response cycle
- Use httptest for HTTP testing
- Validate JSON responses

## Implementation

```go
// Unit test with mock
func TestCreateAppointment(t *testing.T) {
    // Arrange
    mockRepo := new(MockAppointmentRepository)
    service := NewAppointmentService(mockRepo)
    
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    
    // Act
    err := service.CreateAppointment(ctx, req)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}

// Integration test
func TestRepository_Create(t *testing.T) {
    // Setup test database
    pool := setupTestDB(t)
    defer pool.Close()
    
    repo := NewPostgresRepository(pool)
    
    // Test
    err := repo.Create(ctx, appointment)
    assert.NoError(t, err)
}
```

## Consequences

### Positive

- Standard Go testing practices
- Fast test execution
- Easy to run in CI/CD
- Good IDE integration

### Negative

- More boilerplate than some frameworks
- Need to manually setup test fixtures
- No built-in HTTP testing helpers

## Coverage Target

- Overall: > 80%
- Business Logic (service layer): > 90%
- Handlers: > 70%
- Repository: > 85%

## References

- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [testify](https://github.com/stretchr/testify)

```

### 7. Create ADR Index

Location: `appointment-service/docs/adr/README.md`

```markdown
# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the Appointment Service project.

## What is an ADR?

An ADR is a document that captures an important architectural decision made along with its context and consequences.

## ADR List

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [001](ADR-001-database-access-layer.md) | Database Access Layer | Accepted | 2026-01-31 |
| [002](ADR-002-api-framework.md) | API Framework | Accepted | 2026-01-31 |
| [003](ADR-003-logging.md) | Logging Strategy | Accepted | 2026-01-31 |
| [004](ADR-004-configuration.md) | Configuration Management | Accepted | 2026-01-31 |
| [005](ADR-005-testing.md) | Testing Strategy | Accepted | 2026-01-31 |

## Status Definitions

- **Proposed**: Decision is proposed and under review
- **Accepted**: Decision has been approved and is in effect
- **Deprecated**: Decision is no longer recommended
- **Superseded**: Decision has been replaced by a newer ADR

## Creating New ADRs

1. Copy `TEMPLATE.md`
2. Name it `ADR-XXX-title.md` (XXX = next number)
3. Fill in all sections
4. Add to this index
5. Commit with message: `docs: add ADR-XXX for [title]`
```

### 8. Commit ADRs

```bash
# Add all ADR files
git add docs/adr/

# Commit
git commit -m "docs: add Architecture Decision Records

- ADR-001: Database Access Layer (pgx)
- ADR-002: API Framework (Gin)
- ADR-003: Logging Strategy (Zap)
- ADR-004: Configuration Management (godotenv)
- ADR-005: Testing Strategy (testify)
- Add ADR template and index"

# Push
git push origin master
```

---

## Verification Checklist

- [ ] TEMPLATE.md created for future ADRs
- [ ] ADR-001 (Database Access Layer) created
- [ ] ADR-002 (API Framework) created
- [ ] ADR-003 (Logging) created
- [ ] ADR-004 (Configuration) created
- [ ] ADR-005 (Testing) created
- [ ] README.md index created
- [ ] All ADRs committed and pushed

---

## Next Steps

Proceed to **TASK_W01_04_CICD_PIPELINE.md** to set up GitHub Actions and CI/CD.

---

**Status**: ⏸️ Ready to Start
