# Task: Fix Double-Buffer Exclusion Zone in filterAvailableSlots

**Status:** 🔴 TODO  
**Priority:** Critical (Bug)  
**File:** `internal/service/availability_service.go`  
**Function:** `filterAvailableSlots`

---

## Objective

Fix a correctness bug where buffer minutes are applied to both the candidate slot and the existing appointment, making the effective exclusion zone twice the configured buffer value.

---

## Current State (Bug)

```go
func (s *AvailabilityService) filterAvailableSlots(slots []models.TimeSlot, appointments []models.Appointment, rule *models.AvailabilityRule) []models.TimeSlot {
    bufferBefore := time.Duration(rule.BufferBeforeMinutes) * time.Minute
    bufferAfter  := time.Duration(rule.BufferAfterMinutes) * time.Minute

    for _, slot := range slots {
        // Buffer applied to the SLOT
        effectiveSlotStart := slot.StartTime.Add(-bufferBefore)
        effectiveSlotEnd   := slot.EndTime.Add(bufferAfter)

        for _, appt := range appointments {
            // Buffer also applied to the APPOINTMENT — doubling the exclusion zone
            effectiveApptStart := appt.StartTime.Add(-bufferBefore)
            effectiveApptEnd   := appt.EndTime.Add(bufferAfter)

            if effectiveSlotStart.Before(effectiveApptEnd) && effectiveSlotEnd.After(effectiveApptStart) {
                hasConflict = true
            }
        }
    }
}
```

With `BufferBeforeMinutes=15` and `BufferAfterMinutes=15`, a 30-minute appointment at 10:00–10:30 blocks:

- Expected blocked zone: `09:45–10:45` (15 min on each side)
- Actual blocked zone: `09:30–11:00` (30 min on each side — **2× configured**)

---

## Target State

The buffer defines a **blocked zone around each existing appointment**. A candidate slot must not overlap that zone. The buffer is only applied once — to the appointment's boundaries.

```go
for _, appt := range appointments {
    blockedStart := appt.StartTime.Add(-bufferBefore)
    blockedEnd   := appt.EndTime.Add(bufferAfter)

    if slot.StartTime.Before(blockedEnd) && slot.EndTime.After(blockedStart) {
        hasConflict = true
        break
    }
}
```

---

## Implementation Steps

### 1. Update `filterAvailableSlots` in `internal/service/availability_service.go`

Replace the inner loop from:

```go
// Calculate effective slot boundaries with buffers
effectiveSlotStart := slot.StartTime.Add(-bufferBefore)
effectiveSlotEnd := slot.EndTime.Add(bufferAfter)

// Check for conflicts with existing appointments
hasConflict := false
for _, appt := range appointments {
    // Calculate effective appointment boundaries with buffers
    effectiveApptStart := appt.StartTime.Add(-bufferBefore)
    effectiveApptEnd := appt.EndTime.Add(bufferAfter)

    // Check for overlap
    if effectiveSlotStart.Before(effectiveApptEnd) && effectiveSlotEnd.After(effectiveApptStart) {
        hasConflict = true
        break
    }
}
```

To:

```go
// Check for conflicts: does this slot overlap the blocked zone around any existing appointment?
hasConflict := false
for _, appt := range appointments {
    blockedStart := appt.StartTime.Add(-bufferBefore)
    blockedEnd   := appt.EndTime.Add(bufferAfter)

    if slot.StartTime.Before(blockedEnd) && slot.EndTime.After(blockedStart) {
        hasConflict = true
        break
    }
}
```

### 2. Update `TestSlotGenerationWithBuffers` in `internal/service/availability_service_test.go`

Verify the test asserts the correct expected slot count after the fix. With the bug, the test may have been written to match the incorrect doubled-buffer behavior. Recalculate expected available slots manually and update assertions if needed.

---

## Acceptance Criteria

- [ ] A 15-minute `buffer_before` blocks exactly 15 minutes before each existing appointment.
- [ ] A 15-minute `buffer_after` blocks exactly 15 minutes after each existing appointment.
- [ ] Slots outside the buffer zone on either side remain available.
- [ ] `TestSlotGenerationWithBuffers` passes with corrected assertions.
- [ ] No regression in `TestSlotGeneration` (zero-buffer case is unaffected).
