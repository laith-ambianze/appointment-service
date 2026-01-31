# Design Decision: Flexible Participants Pattern (Option A)

## Decision

The appointment service uses a **flexible participants pattern** where appointments can have **multiple participants** tracked through a separate `appointment_participants` table.

## Why This Design?

### ✅ Benefits

1. **Scalability**: Supports 1-on-1 meetings AND group meetings (3+ participants)
2. **Flexibility**: Can add/remove participants without schema changes
3. **Clear Roles**: Each participant has a defined role (host, guest, attendee, observer)
4. **Future-Proof**: Easy to extend with new features like:
   - Group appointments (>2 people)
   - Multiple hosts
   - Optional attendees
   - Observer roles
5. **Industry Standard**: This is how major systems (Google Calendar, Microsoft Outlook) handle appointments

### ❌ What We Avoided

The alternative design (hardcoded two users) had limitations:

- Only supports 1-on-1 meetings
- Cannot add a third person to an existing meeting
- Requires schema changes for group features
- Less flexible for future requirements

---

## Database Schema

### Appointments Table

Stores the core appointment information:

```sql
CREATE TABLE appointments (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    
    -- Creator/Host identifier
    created_by VARCHAR(255) NOT NULL,
    
    -- Appointment details
    title VARCHAR(500) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    
    -- ... other fields
);
```

**Key Point**: The `created_by` field stores the external user ID of whoever created the appointment (usually the host).

### Participants Table

Tracks all users involved in the appointment:

```sql
CREATE TABLE appointment_participants (
    id UUID PRIMARY KEY,
    appointment_id UUID NOT NULL,
    
    -- External user reference (from your product)
    external_user_id VARCHAR(255) NOT NULL,
    
    -- Role in this appointment
    role VARCHAR(50) NOT NULL, -- 'host', 'guest', 'attendee', 'observer'
    
    -- User details from your product
    user_metadata JSONB NOT NULL,
    
    -- Participant's response status
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'accepted', 'declined'
    
    UNIQUE (appointment_id, external_user_id)
);
```

---

## How It Works

### Creating an Appointment

When you create an appointment with 2 people:

```json
{
  "title": "Business Meeting",
  "startTime": "2026-02-15T10:00:00Z",
  "endTime": "2026-02-15T11:00:00Z",
  "participants": [
    {
      "externalUserId": "user_123",
      "role": "host",
      "metadata": {
        "firstName": "John",
        "lastName": "Doe",
        "email": "john@example.com"
      }
    },
    {
      "externalUserId": "user_456",
      "role": "guest",
      "metadata": {
        "firstName": "Jane",
        "lastName": "Smith",
        "email": "jane@example.com"
      }
    }
  ]
}
```

**What happens:**

1. Creates 1 row in `appointments` table with `created_by = "user_123"`
2. Creates 2 rows in `appointment_participants` table:
   - One for John (host)
   - One for Jane (guest)

### Finding Appointments for a User

To get all appointments where user_123 is involved:

```sql
SELECT a.*
FROM appointments a
INNER JOIN appointment_participants ap ON a.id = ap.appointment_id
WHERE ap.external_user_id = 'user_123'
  AND a.product_id = $your_product_id
ORDER BY a.start_time DESC;
```

### Finding Who Created an Appointment

The `created_by` field tells you who initiated the appointment:

```go
// In the appointment object
appointment.CreatedBy // Returns "user_123"

// To get full creator details, find them in participants
for _, participant := range appointment.Participants {
    if participant.ExternalUserID == appointment.CreatedBy {
        // This is the creator
        creatorMetadata := participant.UserMetadata
    }
}
```

### Finding All Participants

```sql
SELECT external_user_id, role, user_metadata, status
FROM appointment_participants
WHERE appointment_id = $appointment_id
ORDER BY CASE role 
    WHEN 'host' THEN 1 
    WHEN 'guest' THEN 2 
    ELSE 3 
END;
```

---

## Common Queries

### 1. Get appointments I created (as host)

```sql
SELECT a.*
FROM appointments a
WHERE a.created_by = 'user_123'
  AND a.product_id = $product_id;
```

### 2. Get appointments where I'm a participant

```sql
SELECT DISTINCT a.*
FROM appointments a
INNER JOIN appointment_participants ap ON a.id = ap.appointment_id
WHERE ap.external_user_id = 'user_123'
  AND a.product_id = $product_id;
```

### 3. Get my role in a specific appointment

```sql
SELECT role, status
FROM appointment_participants
WHERE appointment_id = $appointment_id
  AND external_user_id = 'user_123';
```

### 4. Get all participants of an appointment

```sql
SELECT ap.external_user_id, ap.role, ap.user_metadata, ap.status
FROM appointment_participants ap
WHERE ap.appointment_id = $appointment_id
ORDER BY CASE ap.role 
    WHEN 'host' THEN 1 
    WHEN 'guest' THEN 2 
    ELSE 3 
END;
```

### 5. Find appointments with pending responses

```sql
SELECT a.*
FROM appointments a
INNER JOIN appointment_participants ap ON a.id = ap.appointment_id
WHERE ap.external_user_id = 'user_123'
  AND ap.status = 'pending'
  AND a.product_id = $product_id;
```

---

## Participant Roles

### Host

- Creates the appointment
- Primary organizer
- Usually listed first
- Can cancel the appointment

### Guest

- Invited participants
- Can accept/decline invitation
- Equal status to other guests

### Attendee

- General participant
- Useful for larger meetings

### Observer

- Can view but not required to attend
- Useful for optional participants or CC'd people

---

## Future Possibilities

With this design, you can easily add:

1. **Group Meetings**: Just add more participants

   ```json
   {
     "participants": [
       {"externalUserId": "user_1", "role": "host"},
       {"externalUserId": "user_2", "role": "guest"},
       {"externalUserId": "user_3", "role": "guest"},
       {"externalUserId": "user_4", "role": "observer"}
     ]
   }
   ```

2. **Participant Management**: Add/remove participants after creation

   ```http
   POST /appointments/:id/participants
   DELETE /appointments/:id/participants/:userId
   ```

3. **RSVP System**: Track who accepted/declined

   ```http
   PATCH /appointments/:id/participants/:userId/response
   ```

4. **Multiple Hosts**: Multiple people can co-host

   ```json
   {"externalUserId": "user_1", "role": "host"},
   {"externalUserId": "user_2", "role": "host"}
   ```

---

## Integration Example

### Step 1: Create an Appointment

```bash
curl -X POST https://api.appointment-service.com/v1/appointments \
  -H "X-API-Key: your_api_key" \
  -H "X-API-Secret: your_api_secret" \
  -d '{
    "title": "Project Kickoff",
    "startTime": "2026-02-15T10:00:00Z",
    "endTime": "2026-02-15T11:00:00Z",
    "participants": [
      {
        "externalUserId": "user_123",
        "role": "host",
        "metadata": {"firstName": "John", "lastName": "Doe"}
      },
      {
        "externalUserId": "user_456",
        "role": "guest",
        "metadata": {"firstName": "Jane", "lastName": "Smith"}
      }
    ]
  }'
```

### Step 2: Get User's Appointments

```bash
curl https://api.appointment-service.com/v1/appointments/user/user_123 \
  -H "X-API-Key: your_api_key" \
  -H "X-API-Secret: your_api_secret"
```

**Response:**

```json
{
  "success": true,
  "data": {
    "appointments": [
      {
        "id": "uuid",
        "title": "Project Kickoff",
        "createdBy": "user_123",
        "startTime": "2026-02-15T10:00:00Z",
        "participants": [
          {
            "externalUserId": "user_123",
            "role": "host",
            "metadata": {"firstName": "John", "lastName": "Doe"}
          },
          {
            "externalUserId": "user_456",
            "role": "guest",
            "metadata": {"firstName": "Jane", "lastName": "Smith"}
          }
        ]
      }
    ]
  }
}
```

---

## Migration from Two-User Pattern

If you started with the two-user pattern, migration is straightforward:

```sql
-- For each existing appointment, create two participants
INSERT INTO appointment_participants (id, appointment_id, external_user_id, role, user_metadata, status)
SELECT 
    gen_random_uuid(),
    id,
    user1_id,
    'host',
    user1_metadata,
    'accepted'
FROM appointments;

INSERT INTO appointment_participants (id, appointment_id, external_user_id, role, user_metadata, status)
SELECT 
    gen_random_uuid(),
    id,
    user2_id,
    'guest',
    user2_metadata,
    'pending'
FROM appointments;

-- Update appointments to use created_by
UPDATE appointments SET created_by = user1_id;
```

---

## Summary

✅ **Use This Design For:**

- Flexible participant management
- Scalable to group meetings
- Clear role definitions
- Future-proof architecture

📋 **Key Tables:**

- `appointments`: Core appointment data + `created_by`
- `appointment_participants`: All participants with roles

🔑 **Key Concepts:**

- `created_by`: Who initiated the appointment
- `role`: What role each participant plays
- `external_user_id`: Your system's user identifier
- `user_metadata`: User details stored as JSONB

This design gives you maximum flexibility while keeping the API simple and intuitive!
