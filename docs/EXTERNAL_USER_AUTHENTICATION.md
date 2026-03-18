# External User Authentication Guide

## Overview

The Appointment Service uses a **stateless JWT-based authentication** model where external users are represented solely through JWT claims. The service **does not store user accounts internally** — all user identification comes from the integrating product's system.

## Key Concepts

### External User ID

- An `external_user_id` is the user identifier from your (the integrating product's) system
- This can be any string up to 255 characters (e.g., UUID, email, username, database ID)
- The Appointment Service treats this as an opaque identifier
- Same `external_user_id` in different products are completely isolated (multi-tenancy)

### Product

- A Product represents your application/company that integrates with the Appointment Service
- Each Product has unique `api_key` and `api_secret` credentials
- All data (appointments, participants) is isolated per Product

### Roles

Three roles are supported:

- `admin`: Full access to all operations within the product
- `provider`: Can manage appointments, update statuses, manage availability
- `user`: Can create/view appointments, manage own participation

## Authentication Flow

```md
┌──────────────────┐     ┌─────────────────────┐     ┌────────────────────┐
│   Your Product   │     │ Appointment Service │     │      Database      │
│    (Backend)     │     │        API          │     │                    │
└────────┬─────────┘     └──────────┬──────────┘     └─────────┬──────────┘
         │                          │                          │
         │  1. Register Product     │                          │
         │  POST /v1/products/register                         │
         │  {name: "My App"}        │                          │
         │ ─────────────────────────>                          │
         │                          │                          │
         │  {api_key, api_secret}   │   Store product          │
         │ <─────────────────────────────────────────────────> │
         │                          │                          │
         │  2. Your user logs in    │                          │
         │  (handled by your system)│                          │
         │                          │                          │
         │  3. Request JWT Token    │                          │
         │  POST /v1/auth/token     │                          │
         │  {api_key, api_secret,   │                          │
         │   external_user_id,      │   Validate credentials   │
         │   role}                  │ ─────────────────────────>│
         │ ─────────────────────────>                          │
         │                          │                          │
         │  {token, expires_in}     │                          │
         │ <─────────────────────────                          │
         │                          │                          │
         │  4. API Requests with JWT│                          │
         │  Authorization: Bearer <token>                      │
         │ ─────────────────────────>                          │
         │                          │  Extract product_id,     │
         │                          │  external_user_id, role  │
         │                          │  from JWT claims         │
         │                          │                          │
         │  Response (tenant-isolated)                         │
         │ <─────────────────────────                          │
```

## API Reference

### 1. Register Your Product

**Request:**

```http
POST /v1/products/register
Content-Type: application/json

{
  "name": "My Healthcare App",
  "description": "Appointment scheduling for clinics",
  "webhook_url": "https://myapp.com/webhooks/appointments"
}
```

**Response:**

```json
{
  "product": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "My Healthcare App",
    "status": "active",
    "created_at": "2026-03-18T10:00:00Z"
  },
  "api_key": "apt_a1b2c3d4e5f6...",
  "api_secret": "secret_xyz789..."
}
```

> ⚠️ **Important:** Save the `api_secret` securely. It is only shown once during registration.

### 2. Generate JWT Token for External User

When a user in your system needs to access the Appointment Service, generate a JWT for them:

**Request:**

```http
POST /v1/auth/token
Content-Type: application/json

{
  "api_key": "apt_a1b2c3d4e5f6...",
  "api_secret": "secret_xyz789...",
  "external_user_id": "user_12345",
  "role": "user"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "token_type": "Bearer"
}
```

### JWT Token Structure

The JWT contains these claims:

```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "external_user_id": "user_12345",
  "role": "user",
  "iss": "appointment-service",
  "sub": "user_12345",
  "exp": 1710849600,
  "iat": 1710763200,
  "nbf": 1710763200
}
```

### 3. Make Authenticated API Requests

Use the JWT token in the `Authorization` header:

**Create Appointment:**

```http
POST /v1/appointments
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "title": "Doctor Visit",
  "start_time": "2026-03-20T14:00:00Z",
  "end_time": "2026-03-20T14:30:00Z",
  "timezone": "America/New_York",
  "participants": [
    {
      "external_user_id": "user_12345",
      "role": "host",
      "user_metadata": {
        "name": "John Doe",
        "email": "john@example.com"
      }
    },
    {
      "external_user_id": "provider_789",
      "role": "attendee",
      "user_metadata": {
        "name": "Dr. Smith"
      }
    }
  ]
}
```

**Response:**

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Doctor Visit",
  "status": "scheduled",
  "created_by": "user_12345",
  "participants": [
    {
      "external_user_id": "user_12345",
      "role": "host",
      "status": "accepted"
    },
    {
      "external_user_id": "provider_789",
      "role": "attendee",
      "status": "pending"
    }
  ]
}
```

## Multi-Tenancy & Data Isolation

Each product's data is completely isolated:

```md
Product A (Healthcare App)          Product B (Salon App)
├── Appointments                     ├── Appointments
│   └── Created by product A users   │   └── Created by product B users
├── Availability Rules               ├── Availability Rules
│   └── Product A providers          │   └── Product B providers
└── Participants                     └── Participants
    └── Product A external users         └── Product B external users
```

The same `external_user_id` can exist in multiple products without conflict:

- Product A might have `external_user_id: "user_123"` (John Doe)
- Product B might have `external_user_id: "user_123"` (Different person)

These are completely separate entities.

## Role-Based Access Control (RBAC)

| Endpoint                                    | user | provider | admin |
|---------------------------------------------|------|----------|-------|
| GET /v1/appointments                        | ✓    | ✓        | ✓     |
| POST /v1/appointments                       | ✓    | ✓        | ✓     |
| PATCH /v1/appointments/:id                  | ✓*   | ✓        | ✓     |
| DELETE /v1/appointments/:id                 | ✗    | ✗        | ✓     |
| PATCH /v1/appointments/:id/response         | ✗    | ✓        | ✓     |
| GET /v1/availability                        | ✓    | ✓        | ✓     |
| POST /v1/providers/:id/availability         | ✗    | ✓        | ✓     |
| GET /v1/products/me                         | ✓    | ✓        | ✓     |
| GET /v1/products (list all)                 | ✗    | ✗        | ✓     |

*Users can only update appointments they created

## User Metadata

Since users are not stored internally, you can attach metadata to participants:

```json
{
  "external_user_id": "user_12345",
  "role": "host",
  "user_metadata": {
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+1-555-123-4567",
    "avatar_url": "https://example.com/avatars/john.jpg"
  }
}
```

This metadata is stored with the appointment participant and returned in API responses.

## Best Practices

1. **Token Management**
   - Generate tokens on your backend, never expose `api_secret` to clients
   - Tokens expire in 24 hours by default; regenerate as needed
   - Consider shorter expiry for sensitive operations

2. **External User IDs**
   - Use stable identifiers (database IDs, UUIDs)
   - Avoid using mutable data (emails that might change)
   - Keep under 255 characters

3. **Role Assignment**
   - Assign the minimum required role
   - Use `admin` sparingly
   - Map your application roles to our three roles

4. **Error Handling**
   - `401 Unauthorized`: Invalid or expired token
   - `403 Forbidden`: Valid token but insufficient role permissions
   - `404 Not Found`: Resource doesn't exist or belongs to different product

## Example Integration (Node.js)

```javascript
const axios = require('axios');

class AppointmentServiceClient {
  constructor(apiKey, apiSecret, baseUrl = 'https://appointments.example.com') {
    this.apiKey = apiKey;
    this.apiSecret = apiSecret;
    this.baseUrl = baseUrl;
  }

  // Get JWT for an external user
  async getToken(externalUserId, role = 'user') {
    const response = await axios.post(`${this.baseUrl}/v1/auth/token`, {
      api_key: this.apiKey,
      api_secret: this.apiSecret,
      external_user_id: externalUserId,
      role: role
    });
    return response.data.token;
  }

  // Create appointment for user
  async createAppointment(externalUserId, appointmentData) {
    const token = await this.getToken(externalUserId);
    
    const response = await axios.post(
      `${this.baseUrl}/v1/appointments`,
      appointmentData,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    return response.data;
  }

  // Get user's appointments
  async getAppointments(externalUserId) {
    const token = await this.getToken(externalUserId);
    
    const response = await axios.get(
      `${this.baseUrl}/v1/appointments`,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    return response.data;
  }
}

// Usage
const client = new AppointmentServiceClient(
  'apt_your_api_key',
  'your_api_secret'
);

// When user "john_123" wants to book an appointment:
const appointment = await client.createAppointment('john_123', {
  title: 'Consultation',
  start_time: '2026-03-20T10:00:00Z',
  end_time: '2026-03-20T10:30:00Z',
  timezone: 'America/New_York',
  participants: [
    { external_user_id: 'john_123', role: 'host' },
    { external_user_id: 'dr_smith', role: 'attendee' }
  ]
});
```

## Summary

The Appointment Service authentication flow:

1. **Register once**: Your product registers and receives `api_key` + `api_secret`
2. **No user storage**: Users live in your system, not ours
3. **JWT per request**: Generate JWTs containing `external_user_id` for API access
4. **Full isolation**: Each product's data is completely separate
5. **Simple RBAC**: Three roles cover all common use cases
