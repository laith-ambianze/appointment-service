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
