# API Testing Guide for Postman

This guide provides complete examples for testing all API endpoints using Postman.

---

## Setup: Get JWT Token First

**POST** `http://localhost:8081/v1/auth/token`

Body (JSON):

```json
{
  "api_key": "YOUR_API_KEY",
  "api_secret": "YOUR_API_SECRET",
  "user_id": "user-123",
  "role": "admin"
}
```

Copy the `token` from response and set it in Authorization header: `Bearer <token>`

---

## Appointments API

### 1. Create Appointment

**POST** `http://localhost:8081/v1/appointments`

```json
{
  "title": "Project Meeting",
  "description": "Discuss Q1 goals",
  "start_time": "2026-02-25T10:00:00Z",
  "end_time": "2026-02-25T11:00:00Z",
  "location": "Conference Room A",
  "metadata": {"priority": "high"}
}
```

### 2. List Appointments

**GET** `http://localhost:8081/v1/appointments`

Query params (optional): `?page=1&page_size=10&status=scheduled`

### 3. Get Single Appointment

**GET** `http://localhost:8081/v1/appointments/:id`

### 4. Update Appointment

**PATCH** `http://localhost:8081/v1/appointments/:id`

```json
{
  "title": "Updated Meeting Title",
  "location": "Room B"
}
```

### 5. Delete Appointment

**DELETE** `http://localhost:8081/v1/appointments/:id`

### 6. Cancel Appointment

**POST** `http://localhost:8081/v1/appointments/:id/cancel`

```json
{
  "reason": "Rescheduling to next week"
}
```

### 7. Respond to Appointment

**POST** `http://localhost:8081/v1/appointments/:id/response`

```json
{
  "response": "accepted"
}
```

Options: `accepted`, `declined`, `tentative`

---

## Participants API

### 8. Add Participant

**POST** `http://localhost:8081/v1/appointments/:id/participants`

```json
{
  "user_id": "user-456",
  "role": "attendee"
}
```

Roles: `organizer`, `attendee`, `optional`

### 9. Remove Participant

**DELETE** `http://localhost:8081/v1/appointments/:id/participants/:participantId`

### 10. Update Participant Response

**PATCH** `http://localhost:8081/v1/appointments/:id/participants/:participantId`

```json
{
  "response_status": "accepted"
}
```

---

## Products API

### 11. Register Product (Public - No Auth Required)

**POST** `http://localhost:8081/v1/products/register`

```json
{
  "name": "New Company",
  "description": "Company description"
}
```

### 12. Validate Credentials (Public - No Auth Required)

**POST** `http://localhost:8081/v1/products/validate`

```json
{
  "api_key": "apk_xxx",
  "api_secret": "aps_xxx"
}
```

### 13. Get Current Product (Authenticated)

**GET** `http://localhost:8081/v1/products/me`

### 14. Update Current Product

**PATCH** `http://localhost:8081/v1/products/me`

```json
{
  "name": "Updated Name"
}
```

### 15. Regenerate Credentials

**POST** `http://localhost:8081/v1/products/me/regenerate-credentials`

---

## Health Check

**GET** `http://localhost:8081/health`

---

## Authentication Flow

1. **Register a Product** (if you don't have one):
   - POST `/v1/products/register`
   - Save the `api_key` and `api_secret` from response

2. **Generate JWT Token**:
   - POST `/v1/auth/token` with your credentials
   - Copy the `token` from response

3. **Use Token in Requests**:
   - Add header: `Authorization: Bearer <your_token>`

---

## Role-Based Access

| Role | Permissions |
| ------ | ------------- |
| `admin` | Full access to all operations |
| `provider` | Create/manage appointments, manage participants |
| `user` | View appointments, respond to invitations |

---

## Response Status Codes

| Code | Description |
| ------ | ------------- |
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found |
| 500 | Internal Server Error |
