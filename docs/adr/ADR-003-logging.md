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
