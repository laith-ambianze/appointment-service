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

// HTTP handler test
func TestGetAppointment_Handler(t *testing.T) {
    router := setupTestRouter()
    
    req, _ := http.NewRequest("GET", "/v1/appointments/123", nil)
    resp := httptest.NewRecorder()
    
    router.ServeHTTP(resp, req)
    
    assert.Equal(t, http.StatusOK, resp.Code)
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
