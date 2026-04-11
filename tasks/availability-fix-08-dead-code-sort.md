# Task: Remove Dead SortSlotsByStartTime Exported Function

**Status:** 🟢 TODO  
**Priority:** Low (Cleanup)  
**File:** `internal/service/availability_service.go`  
**Function:** `SortSlotsByStartTime`

---

## Objective

Delete the exported `SortSlotsByStartTime` function from the service layer. It is never called internally, lives in the wrong package, and the slots it would sort are already generated in ascending time order by `generateSlots`. Keeping it as dead exported code adds noise and creates a misleading API surface.

---

## Current State

```go
// availability_service.go — bottom of file
// SortSlotsByStartTime sorts time slots by start time
func SortSlotsByStartTime(slots []models.TimeSlot) {
    sort.Slice(slots, func(i, j int) bool {
        return slots[i].StartTime.Before(slots[j].StartTime)
    })
}
```

Issues:

- Not called anywhere in the codebase (`grep -rn "SortSlotsByStartTime" .` returns zero results beyond the definition).
- `generateSlots` already iterates slots in ascending order — no sorting is needed.
- It is exported from the `service` package, suggesting callers should use it. No caller does.
- The `sort` import is only present because of this function; removing it cleans up the import block.

---

## Implementation Steps

### 1. Verify no callers exist

```bash
grep -rn "SortSlotsByStartTime" .
```

Expected output: only the definition in `availability_service.go`. If any caller is found, assess whether they can be removed or whether the function should be moved to `models` instead.

### 2. Delete the function from `internal/service/availability_service.go`

Remove:

```go
// SortSlotsByStartTime sorts time slots by start time
func SortSlotsByStartTime(slots []models.TimeSlot) {
    sort.Slice(slots, func(i, j int) bool {
        return slots[i].StartTime.Before(slots[j].StartTime)
    })
}
```

### 3. Remove the `sort` import if it is now unused

Check the import block:

```go
import (
    // ...
    "sort"   // <- remove if SortSlotsByStartTime was the only user
    // ...
)
```

Run `go build ./...` to confirm no unused import error.

### 4. Run all tests to confirm no regression

```bash
go test ./internal/service/... -v
```

---

## Acceptance Criteria

- [ ] `SortSlotsByStartTime` is deleted from `availability_service.go`.
- [ ] The `sort` import is removed if unused.
- [ ] `go build ./...` passes with no errors.
- [ ] All existing tests continue to pass.
- [ ] No other file references `SortSlotsByStartTime`.
