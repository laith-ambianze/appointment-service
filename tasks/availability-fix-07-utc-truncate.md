# Task: Fix Timezone-Unsafe time.Now().Truncate in GetAvailableSlots

**Status:** 🟡 TODO  
**Priority:** Medium (Bug — server-timezone-dependent)  
**File:** `internal/service/availability_service.go`  
**Function:** `GetAvailableSlots`

---

## Objective

Replace `time.Now().Truncate(24 * time.Hour)` with the timezone-safe equivalent `time.Now().UTC().Truncate(24 * time.Hour)`. The current code produces the wrong result when the service is deployed on a host whose local timezone is not UTC.

---

## Current State (Bug)

```go
func (s *AvailabilityService) GetAvailableSlots(ctx context.Context, ...) (*AvailableSlotsResponse, error) {
    // Parse date — always returns UTC midnight regardless of server timezone
    date, err := time.Parse("2006-01-02", req.Date)

    // Check if date is in the past
    today := time.Now().Truncate(24 * time.Hour) // <- affected by server's local timezone
    if date.Before(today) {
        return nil, ErrDateInPast
    }
```

`time.Parse("2006-01-02", "2026-04-09")` returns `2026-04-09 00:00:00 +0000 UTC`.  
`time.Now().Truncate(24 * time.Hour)` on a server in `UTC+9` at 23:00 local returns `2026-04-09 14:00:00 +0000 UTC` (14:00 UTC = start of the local day).

This means `date.Before(today)` evaluates as `00:00 UTC < 14:00 UTC → true`, incorrectly rejecting today's date as "in the past" for the entire 14-hour window before UTC midnight.

On a server in `UTC-8`, the opposite occurs: the "today" midnight is 08:00 UTC, so dates from 00:00–07:59 UTC on the current UTC day are accepted even though they are technically "yesterday" in UTC.

---

## Target State

Both `date` (from `time.Parse`) and `today` must be computed in the same timezone — UTC — so the comparison is consistent regardless of the server's local timezone setting.

---

## Implementation Steps

### 1. Change the `today` calculation in `GetAvailableSlots`

```go
// Before (wrong)
today := time.Now().Truncate(24 * time.Hour)

// After (correct)
today := time.Now().UTC().Truncate(24 * time.Hour)
```

That is the **entire change** — one method call inserted.

### 2. Search for other instances of the same pattern

Run a search across the service layer to ensure no other "in the past" date checks use the same unsafe pattern:

```bash
grep -rn "time.Now().Truncate" internal/ pkg/
```

Apply the same fix to any other occurrences found.

### 3. Add a unit test for the boundary condition

```go
func TestGetAvailableSlots_TodayNotRejectedAsInPast(t *testing.T) {
    svc, availRepo, _ := newAvailabilityTestService()
    productID := uuid.New()
    providerID := "testProvider"

    // Request today's date in YYYY-MM-DD (UTC)
    todayStr := time.Now().UTC().Format("2006-01-02")

    // Even with no rule, should return empty slots — NOT ErrDateInPast
    resp, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
        ProviderID: providerID,
        Date:       todayStr,
    })

    // No rule exists, so we expect an empty-slots response, not a date-in-past error.
    // If it returns ErrDateInPast, the UTC truncation is broken.
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Empty(t, resp.Slots)
    _ = availRepo
}

func TestGetAvailableSlots_YesterdayRejected(t *testing.T) {
    svc, _, _ := newAvailabilityTestService()
    productID := uuid.New()

    yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

    _, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
        ProviderID: "any",
        Date:       yesterday,
    })

    assert.ErrorIs(t, err, ErrDateInPast)
}
```

---

## Acceptance Criteria

- [ ] `time.Now().UTC().Truncate(24 * time.Hour)` is used instead of `time.Now().Truncate(...)`.
- [ ] Today's date (UTC) is never rejected as being in the past.
- [ ] Yesterday's date (UTC) is correctly rejected.
- [ ] Both tests pass on machines in any timezone.
- [ ] No other unsafe `time.Now().Truncate` patterns remain in the service layer.
