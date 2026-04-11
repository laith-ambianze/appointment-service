# Task: Add Date Range Guard to GetAvailableSlotsForDateRange

**Status:** 🟡 TODO  
**Priority:** Medium (Safety / DoS prevention)  
**File:** `internal/service/availability_service.go`  
**Function:** `GetAvailableSlotsForDateRange`

---

## Objective

Add input validation to `GetAvailableSlotsForDateRange` to prevent unbounded DB queries. Currently no check is made that `endDate >= startDate` or that the range does not exceed a safe maximum. Each day in the range executes two database queries, so a caller with a large range can cause significant load.

---

## Current State

```go
func (s *AvailabilityService) GetAvailableSlotsForDateRange(
    ctx context.Context,
    productID uuid.UUID,
    providerID string,
    startDate, endDate time.Time,
    timezone string,
) (map[string][]models.TimeSlot, error) {
    result := make(map[string][]models.TimeSlot)

    // No validation: endDate may be before startDate; range may be arbitrarily large
    current := startDate
    for !current.After(endDate) {
        // 2 DB queries per day
        current = current.Add(24 * time.Hour)
    }

    return result, nil
}
```

This function is not currently wired to any route, but it is exported and could be called internally or wired in the future.

---

## Target State

Validate the date range before iterating. Reject inverted ranges and cap the maximum range at a configurable limit (recommended: 90 days).

---

## Implementation Steps

### 1. Add a maximum range constant near the top of the file

```go
// maxAvailabilityRangeDays is the maximum number of days allowed in a single
// GetAvailableSlotsForDateRange query to prevent excessive database load.
const maxAvailabilityRangeDays = 90
```

### 2. Add validation at the start of `GetAvailableSlotsForDateRange`

```go
func (s *AvailabilityService) GetAvailableSlotsForDateRange(
    ctx context.Context,
    productID uuid.UUID,
    providerID string,
    startDate, endDate time.Time,
    timezone string,
) (map[string][]models.TimeSlot, error) {
    // Validate date order
    if endDate.Before(startDate) {
        return nil, fmt.Errorf("%w: end_date must be on or after start_date", ErrInvalidInput)
    }

    // Cap range to prevent unbounded DB queries
    days := int(endDate.Sub(startDate).Hours()/24) + 1
    if days > maxAvailabilityRangeDays {
        return nil, fmt.Errorf("%w: date range exceeds maximum of %d days (got %d)",
            ErrInvalidInput, maxAvailabilityRangeDays, days)
    }

    result := make(map[string][]models.TimeSlot)
    // ... existing loop unchanged
```

### 3. If/when this function is wired to a route in `internal/handlers/availability.go`

The handler should parse `start_date` and `end_date` query params and pass them to the service. The service-level error `ErrInvalidInput` already maps to `HTTP 400` via `handleServiceError`.

### 4. Add unit tests

```go
func TestGetAvailableSlotsForDateRange_InvertedRange(t *testing.T) {
    svc, _, _ := newAvailabilityTestService()
    productID := uuid.New()

    start := time.Now().AddDate(0, 0, 5)
    end   := time.Now().AddDate(0, 0, 1) // before start

    _, err := svc.GetAvailableSlotsForDateRange(context.Background(), productID, "p1", start, end, "UTC")
    assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestGetAvailableSlotsForDateRange_TooLarge(t *testing.T) {
    svc, _, _ := newAvailabilityTestService()
    productID := uuid.New()

    start := time.Now()
    end   := start.AddDate(0, 4, 0) // > 90 days

    _, err := svc.GetAvailableSlotsForDateRange(context.Background(), productID, "p1", start, end, "UTC")
    assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestGetAvailableSlotsForDateRange_ValidRange(t *testing.T) {
    svc, _, _ := newAvailabilityTestService()
    productID := uuid.New()

    start := time.Now()
    end   := start.AddDate(0, 0, 6) // 7 days — within limit

    result, err := svc.GetAvailableSlotsForDateRange(context.Background(), productID, "p1", start, end, "UTC")
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

---

## Acceptance Criteria

- [ ] `endDate < startDate` returns `ErrInvalidInput`.
- [ ] Range exceeding `maxAvailabilityRangeDays` (90) returns `ErrInvalidInput`.
- [ ] Valid ranges within the limit return results without error.
- [ ] Unit tests for all three cases pass.
- [ ] If the function is wired to a route, the handler returns `HTTP 400` for invalid ranges.
