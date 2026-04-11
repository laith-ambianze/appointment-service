# Task: Auto-Add Creator as Host Participant in BookAppointment

**Status:** 🔴 TODO  
**Priority:** High (Logic inconsistency)  
**File:** `internal/service/availability_service.go`  
**Function:** `BookAppointment`

---

## Objective

Mirror the host auto-add logic from `AppointmentService.Create` into `AvailabilityService.BookAppointment`. Currently a booking can succeed with no host participant, which is inconsistent with the domain model and with how regular appointment creation works.

---

## Current State

`AppointmentService.Create` already handles this correctly:

```go
// appointment_service.go — Create
// Auto-add creator as host if not in the list
hasCreatorAsHost := false
for _, p := range input.Participants {
    if p.ExternalUserID == userID && p.Role == models.ParticipantRoleHost {
        hasCreatorAsHost = true
        break
    }
}
if !hasCreatorAsHost {
    hostParticipant := models.AppointmentParticipant{
        ExternalUserID: userID,
        Role:           models.ParticipantRoleHost,
        Status:         models.ParticipantStatusAccepted,
    }
    input.Participants = append([]models.AppointmentParticipant{hostParticipant}, input.Participants...)
}
```

`AvailabilityService.BookAppointment` only sets the creator's **status** to accepted, but does not enforce the **host role**:

```go
// availability_service.go — BookAppointment (current, incomplete)
for _, p := range req.Participants {
    participant := models.AppointmentParticipant{
        // ...
        Role:   p.Role,  // role comes from request — host is not guaranteed
        Status: models.ParticipantStatusPending,
    }
    if p.ExternalUserID == userID {
        participant.Status = models.ParticipantStatusAccepted  // status only, role not enforced
    }
    participants = append(participants, participant)
}
```

A caller can:

- Submit `participants` with no `host` role → appointment created without a host.
- Omit themselves entirely → their `accepted` status is never set.

---

## Target State

Before building the participants slice, check whether the calling user (`userID`) already appears in the list with `role = host`. If not, prepend them as an accepted host — identical behavior to `AppointmentService.Create`.

---

## Implementation Steps

### 1. Add host-auto-add logic in `BookAppointment` (`internal/service/availability_service.go`)

Insert the following block **before** the `for _, p := range req.Participants` loop:

```go
// Ensure the booking user is present as the host participant
hasCreatorAsHost := false
for _, p := range req.Participants {
    if p.ExternalUserID == userID && p.Role == models.ParticipantRoleHost {
        hasCreatorAsHost = true
        break
    }
}
if !hasCreatorAsHost {
    req.Participants = append([]ParticipantRequest{
        {
            ExternalUserID: userID,
            Role:           models.ParticipantRoleHost,
            UserMetadata:   nil,
        },
    }, req.Participants...)
}
```

Then in the participant-building loop, keep the existing `Status = Accepted` logic for `userID`.

### 2. Add unit tests in `internal/service/availability_service_test.go`

```go
func TestBookAppointment_CreatorAutoAddedAsHost(t *testing.T) {
    svc, availRepo, _ := newAvailabilityTestService()
    productID := uuid.New()
    providerID := "providerX"
    userID := "user123"

    // Setup rule for tomorrow
    tomorrow := time.Now().Add(24 * time.Hour)
    rule := &models.AvailabilityRule{ /* ... Monday, 9:00-17:00 */ }
    availRepo.rules[...] = rule

    req := BookAppointmentRequest{
        ProviderID:   providerID,
        StartTime:    tomorrow, // set to a valid aligned slot
        Title:        "Test booking",
        Participants: []ParticipantRequest{}, // empty — no explicit participants
    }

    appt, err := svc.BookAppointment(context.Background(), productID, userID, req)
    require.NoError(t, err)

    // Creator should be present as host
    hostFound := false
    for _, p := range appt.Participants {
        if p.ExternalUserID == userID && p.Role == models.ParticipantRoleHost {
            hostFound = true
            assert.Equal(t, models.ParticipantStatusAccepted, p.Status)
        }
    }
    assert.True(t, hostFound, "creator should be auto-added as host participant")
}

func TestBookAppointment_ExplicitHostNotDuplicated(t *testing.T) {
    // When caller explicitly includes themselves as host, they should not appear twice
}
```

---

## Acceptance Criteria

- [ ] `BookAppointment` always produces an appointment with at least one participant with `role = host`.
- [ ] If the caller is not in the `Participants` list, they are prepended as host with `status = accepted`.
- [ ] If the caller is already in the list with `role = host`, they are not duplicated.
- [ ] Unit tests cover both cases.
- [ ] Behavior is identical to `AppointmentService.Create`.
