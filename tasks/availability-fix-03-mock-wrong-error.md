# Task: Fix Mock Returning Wrong Error Sentinel in Availability Service Tests

**Status:** 🔴 TODO  
**Priority:** Critical (Test correctness)  
**File:** `internal/service/availability_service_test.go`  
**Function:** `mockAvailabilityRepo` methods

---

## Objective

Fix the availability service test mock so it returns `repository.ErrAvailabilityRuleNotFound` instead of the service-level `ErrAvailabilityRuleNotFound`. As written, the error-translation path in the service (repo error → service error) is never exercised by any test.

---

## Current State (Bug)

The mock returns the service-layer error directly:

```go
// availability_service_test.go

func (m *mockAvailabilityRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.AvailabilityRule, error) {
    for _, rule := range m.rules {
        if rule.ID == id {
            return rule, nil
        }
    }
    return nil, ErrAvailabilityRuleNotFound  // <- service.ErrAvailabilityRuleNotFound (WRONG)
}

func (m *mockAvailabilityRepo) GetByProviderAndDay(...) (*models.AvailabilityRule, error) {
    // ...
    return nil, ErrAvailabilityRuleNotFound  // <- service.ErrAvailabilityRuleNotFound (WRONG)
}

func (m *mockAvailabilityRepo) Delete(ctx context.Context, id uuid.UUID) error {
    // ...
    return ErrAvailabilityRuleNotFound  // <- service.ErrAvailabilityRuleNotFound (WRONG)
}
```

The service code it is meant to test:

```go
// availability_service.go

rule, err := s.availabilityRepo.GetByID(ctx, ruleID)
if err != nil {
    if errors.Is(err, repository.ErrAvailabilityRuleNotFound) {  // <- checks REPO error
        return nil, ErrAvailabilityRuleNotFound                  // translates to SERVICE error
    }
    return nil, err  // passes through unknown errors
}
```

Because `errors.Is(service.ErrAvailabilityRuleNotFound, repository.ErrAvailabilityRuleNotFound)` is `false`, the `if errors.Is(...)` branch is **never taken in tests**. The mock error falls through to `return nil, err`, which accidentally returns the correct final value — but the translation logic is untested. A bug introduced in that branch would be invisible.

---

## Target State

All mock repository methods that represent "not found" must return `repository.ErrAvailabilityRuleNotFound`.

---

## Implementation Steps

### 1. Add the `repository` import to the test file

```go
import (
    // existing imports...
    "github.com/laith-ambianze/appointment-service/internal/repository"
)
```

### 2. Update all "not found" returns in the mock

Replace every `return nil, ErrAvailabilityRuleNotFound` (and `return ErrAvailabilityRuleNotFound`) in the mock struct methods:

```go
// GetByID
return nil, repository.ErrAvailabilityRuleNotFound

// GetByProviderAndDay
return nil, repository.ErrAvailabilityRuleNotFound

// Delete
return repository.ErrAvailabilityRuleNotFound

// DeleteByProviderAndDay
return repository.ErrAvailabilityRuleNotFound
```

### 3. Add a test that verifies error translation

Add a test case that confirms the service returns `service.ErrAvailabilityRuleNotFound` (not a raw repo error) when the rule does not exist:

```go
func TestGetAvailabilityRule_NotFound(t *testing.T) {
    svc, _, _ := newAvailabilityTestService()
    productID := uuid.New()

    _, err := svc.GetAvailabilityRule(context.Background(), productID, uuid.New())

    assert.ErrorIs(t, err, ErrAvailabilityRuleNotFound,
        "service should translate repo not-found to service-level ErrAvailabilityRuleNotFound")
}

func TestDeleteAvailabilityRule_NotFound(t *testing.T) {
    svc, _, _ := newAvailabilityTestService()
    productID := uuid.New()

    err := svc.DeleteAvailabilityRule(context.Background(), productID, uuid.New())

    assert.ErrorIs(t, err, ErrAvailabilityRuleNotFound)
}
```

---

## Acceptance Criteria

- [ ] All mock `AvailabilityRepository` methods return `repository.ErrAvailabilityRuleNotFound` for the not-found case.
- [ ] `TestGetAvailabilityRule_NotFound` passes and confirms the service-level error is returned.
- [ ] `TestDeleteAvailabilityRule_NotFound` passes.
- [ ] All existing passing tests continue to pass.
