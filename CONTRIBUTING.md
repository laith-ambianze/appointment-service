# Contributing Guidelines

Thank you for your interest in contributing to the Appointment Service!

## Development Workflow

### 1. Create a Feature Branch

```bash
# Always branch from master
git checkout master
git pull origin master

# Create your feature branch
git checkout -b feature/your-feature-name

# For bug fixes
git checkout -b bugfix/issue-description

# For hotfixes
git checkout -b hotfix/critical-fix
```

### 2. Make Your Changes

Follow our coding standards (see below) and ensure:

- Code is properly formatted
- Tests are written for new features
- Documentation is updated if needed

### 3. Run Quality Checks

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Check test coverage
make test-coverage
```

### 4. Commit Your Changes

Follow the conventional commits format:

```bash
git commit -m "feat: add appointment creation endpoint"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

---

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

### Format

```md
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
| ------ | ------------- |
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `style` | Code style changes (formatting, semicolons) |
| `refactor` | Code refactoring (no feature or fix) |
| `perf` | Performance improvements |
| `test` | Adding or updating tests |
| `build` | Build system or dependencies |
| `ci` | CI/CD configuration |
| `chore` | Maintenance tasks |
| `revert` | Revert a previous commit |

### Examples

```md
# Feature
feat(appointments): add participant role validation

# Bug fix
fix(availability): resolve timezone conversion bug

# Documentation
docs(api): update appointment endpoint examples

# Refactor
refactor(service): simplify availability calculation

# Tests
test(repository): add unit tests for appointment queries

# Chore
chore(deps): update go dependencies
```

---

## Code Style

### General Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Keep functions small and focused (< 50 lines preferred)
- Use meaningful variable and function names
- Comment exported functions and types
- Handle all errors explicitly

### Naming Conventions

```go
// Package naming: lowercase, single word
package appointment

// Constants: PascalCase
const (
    StatusPending   = "pending"
    StatusConfirmed = "confirmed"
    StatusCancelled = "cancelled"
)

// Functions: PascalCase for exported, camelCase for private
func CreateAppointment() {} // Exported
func validateRequest() {}   // Private

// Interfaces: -er suffix when appropriate
type Repository interface {
    Create(ctx context.Context, appointment *Appointment) error
    GetByID(ctx context.Context, id uuid.UUID) (*Appointment, error)
}

// Structs: PascalCase
type Appointment struct {
    ID        uuid.UUID
    Title     string
    StartTime time.Time
}
```

### Error Handling

```go
// Always wrap errors with context
appointment, err := service.Create(req)
if err != nil {
    return fmt.Errorf("failed to create appointment: %w", err)
}

// Use custom errors for business logic
var (
    ErrAppointmentNotFound = errors.New("appointment not found")
    ErrTimeSlotUnavailable = errors.New("time slot is unavailable")
)
```

### Comments

```go
// CreateAppointment creates a new appointment with the given participants.
// It validates the request, checks availability for all participants,
// and returns the created appointment or an error.
//
// The caller must ensure that the context contains a valid product ID.
func (s *AppointmentService) CreateAppointment(ctx context.Context, req CreateRequest) (*Appointment, error) {
    // Implementation
}
```

### Imports

```go
import (
    // Standard library first
    "context"
    "fmt"
    "time"

    // Third-party packages
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    // Internal packages
    "github.com/laith-ambianze/appointment-service/internal/models"
    "github.com/laith-ambianze/appointment-service/pkg/logger"
)
```

---

## Testing

### Test File Location

- Unit tests: Same package as code, `*_test.go`
- Integration tests: `tests/integration/`

### Test Naming

```go
// Function: Test<FunctionName>_<Scenario>_<ExpectedBehavior>
func TestCreateAppointment_ValidRequest_ReturnsAppointment(t *testing.T) {}
func TestCreateAppointment_InvalidTime_ReturnsError(t *testing.T) {}
func TestCreateAppointment_ConflictingSlot_ReturnsConflictError(t *testing.T) {}
```

### Test Structure (AAA Pattern)

```go
func TestCreateAppointment_ValidRequest_ReturnsAppointment(t *testing.T) {
    // Arrange
    mockRepo := new(MockAppointmentRepository)
    service := NewAppointmentService(mockRepo)
    req := CreateRequest{
        Title:     "Test Appointment",
        StartTime: time.Now().Add(24 * time.Hour),
    }
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    // Act
    appointment, err := service.Create(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, appointment)
    assert.Equal(t, req.Title, appointment.Title)
    mockRepo.AssertExpectations(t)
}
```

### Running Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific package
go test -v ./internal/service/...

# Run specific test
go test -v ./internal/service/... -run TestCreateAppointment

# Run with coverage
make test-coverage
```

### Coverage Target

- Overall: > 80%
- Service layer: > 90%
- Handlers: > 70%
- Repository: > 85%

---

## Pull Request Process

### Before Submitting

- [ ] Code follows our style guidelines
- [ ] `make fmt` has been run
- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] New code has tests
- [ ] Documentation is updated
- [ ] Commit messages follow conventions

### PR Title Format

```md
<type>(<scope>): <description>
```

Examples:

- `feat(appointments): add bulk creation endpoint`
- `fix(availability): handle DST transitions correctly`
- `docs(readme): update quick start guide`

### PR Description Template

```markdown
## Summary
Brief description of changes.

## Changes
- Change 1
- Change 2

## Testing
How was this tested?

## Related Issues
Closes #123
```

### Review Process

1. Automated checks must pass (lint, test, build)
2. At least one approval required
3. All comments must be resolved
4. Branch must be up to date with master

---

## Questions?

If you have questions about contributing, feel free to:

- Open an issue for discussion
- Check existing issues and PRs
- Review the documentation in `docs/`

Thank you for contributing! 🎉
