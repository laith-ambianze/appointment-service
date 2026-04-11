# Task: Fix Day-of-Week Computed from UTC Instead of Provider's Local Timezone in BookAppointment

**Status:** 🔴 TODO  
**Priority:** Critical (Bug)  
**File:** `internal/service/availability_service.go`  
**Function:** `BookAppointment`

---

## Objective

Fix a timezone bug where the day-of-week for the availability rule lookup is derived from the UTC representation of `start_time`, ignoring the provider's local timezone. This causes incorrect or missing rule lookups for providers in non-UTC timezones.

---

## Current State (Bug)

```go
func (s *AvailabilityService) BookAppointment(ctx context.Context, ...) (*models.Appointment, error) {
    // req.StartTime is in UTC
    dayOfWeek := models.DayOfWeek(req.StartTime.Weekday()) // <- UTC weekday, WRONG

    rule, err := s.availabilityRepo.GetByProviderAndDay(ctx, productID, req.ProviderID, dayOfWeek)
```

**Concrete failure example:**

- Provider timezone: `Pacific/Auckland` (`UTC+12`)
- Client sends: `start_time = 2026-04-13T00:30:00Z` (Monday UTC)
- Provider's local time: `2026-04-13T12:30:00+12:00` → Tuesday locally
- Code looks up rule for `Monday` → no rule found → `ErrAvailabilityRuleNotFound`
- Correct behavior: look up rule for `Tuesday` using the provider's timezone

Note: `GetAvailableSlots` does **not** have this bug because it accepts a plain date string (`"2026-04-14"`) that is timezone-ambiguous by design.

---

## Target State

The day-of-week must be resolved after converting `start_time` into the provider's local timezone. Since the provider's timezone is stored on the availability rule, a two-step approach is needed: fetch the rule using the UTC day, and if not found try the adjacent day using the stored timezone.

The cleanest fix is to require the caller to supply the provider's local date explicitly.

---

## Implementation Steps

### Option A — Add `LocalDate` field to `BookAppointmentRequest` (recommended)

This is explicit, unambiguous, and avoids an extra DB round-trip.

#### 1. Add `LocalDate` to request struct in `internal/service/availability_service.go`

```go
type BookAppointmentRequest struct {
    ProviderID   string                 `json:"provider_id"  binding:"required"`
    StartTime    time.Time              `json:"start_time"   binding:"required"`
    LocalDate    string                 `json:"local_date"   binding:"required"` // YYYY-MM-DD in provider's timezone
    Title        string                 `json:"title"        binding:"required,min=1,max=255"`
    // ... existing fields unchanged
}
```

#### 2. Parse `LocalDate` and derive the weekday in `BookAppointment`

```go
// Parse the provider-local date supplied by the caller
localDate, err := time.Parse("2006-01-02", req.LocalDate)
if err != nil {
    return nil, fmt.Errorf("%w: invalid local_date format (expected YYYY-MM-DD): %v", ErrInvalidInput, err)
}
dayOfWeek := models.DayOfWeek(localDate.Weekday())

rule, err := s.availabilityRepo.GetByProviderAndDay(ctx, productID, req.ProviderID, dayOfWeek)
```

#### 3. Remove the old UTC-based weekday derivation

```go
// DELETE these two lines:
dayOfWeek := models.DayOfWeek(req.StartTime.Weekday())
```

### Option B — Derive weekday from StartTime using the rule's timezone (two-step lookup)

If adding a field to the request is undesirable:

```go
// Step 1: try UTC day
dayOfWeek := models.DayOfWeek(req.StartTime.UTC().Weekday())
rule, err := s.availabilityRepo.GetByProviderAndDay(ctx, productID, req.ProviderID, dayOfWeek)

if errors.Is(err, repository.ErrAvailabilityRuleNotFound) {
    // Step 2: the provider may be in a different day locally — try adjacent days
    // We don't know the timezone yet, so try both neighbors
    for _, offset := range []int{1, -1} {
        altDay := models.DayOfWeek((int(dayOfWeek) + offset + 7) % 7)
        rule, err = s.availabilityRepo.GetByProviderAndDay(ctx, productID, req.ProviderID, altDay)
        if err == nil {
            // Verify the start_time actually falls on this day in the rule's timezone
            loc, locErr := time.LoadLocation(rule.Timezone)
            if locErr == nil && models.DayOfWeek(req.StartTime.In(loc).Weekday()) == altDay {
                break
            }
            rule, err = nil, repository.ErrAvailabilityRuleNotFound
        }
    }
}
```

Option A is strongly preferred — Option B adds complexity and an extra DB query per booking.

---

## Acceptance Criteria

- [ ] A provider in `UTC+12` with a Tuesday rule correctly receives bookings when `start_time` is Monday UTC but Tuesday local.
- [ ] A provider in `UTC-8` with a Friday rule correctly receives bookings when `start_time` is Saturday UTC but Friday local.
- [ ] Invalid `local_date` (if Option A chosen) returns `400 Bad Request`.
- [ ] No change to the `GetAvailableSlots` flow (not affected by this bug).
- [ ] Unit test added for the UTC timezone boundary case.
