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
