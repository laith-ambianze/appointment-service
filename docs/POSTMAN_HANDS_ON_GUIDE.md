# Appointment Service: Hands-On Guide & Postman Testing

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture & Key Concepts](#2-architecture--key-concepts)
3. [Workflow Explanation](#3-workflow-explanation)
4. [Postman Setup & Testing](#4-postman-setup--testing)
5. [Testing Scenarios](#5-testing-scenarios)
6. [Tips & Best Practices](#6-tips--best-practices)

---

## 1. Project Overview

### What is the Appointment Service?

The Appointment Service is a **multi-tenant, API-first scheduling platform** that allows external applications ("Products") to integrate appointment booking functionality without managing scheduling infrastructure themselves.

### Key Characteristics

| Feature | Description |
| --------- | ------------- |
| **Multi-Tenant** | Each registered Product has isolated data - appointments, users, and availability rules are completely separate |
| **Stateless Authentication** | Uses JWT tokens; the service does NOT store user accounts internally |
| **External User IDs** | Users exist in YOUR system, not in the Appointment Service |
| **RBAC** | Three roles (admin, provider, user) control access to different operations |

### Architecture Overview

```md
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Your Application (Product)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                         │
│  │   User A    │  │   User B    │  │  Provider   │                         │
│  │ (customer)  │  │ (customer)  │  │ (dr_smith)  │                         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘                         │
│         │                │                │                                  │
│         └────────────────┼────────────────┘                                  │
│                          │                                                   │
│                    ┌─────┴─────┐                                             │
│                    │  Backend  │  ← Stores api_key & api_secret securely    │
│                    └─────┬─────┘                                             │
└──────────────────────────┼───────────────────────────────────────────────────┘
                           │
                           │ JWT Token Requests
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Appointment Service API                               │
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Products   │  │     Auth     │  │ Appointments │  │ Availability │    │
│  │   /register  │  │    /token    │  │    CRUD      │  │    Rules     │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                        PostgreSQL Database                            │  │
│  │  [products] [appointments] [participants] [availability_rules]        │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Architecture & Key Concepts

### 2.1 Products (Tenants)

A **Product** represents your application/company that integrates with the Appointment Service.

- Each Product receives unique `api_key` and `api_secret` credentials upon registration
- All data is isolated per Product (multi-tenancy)
- Products can have multiple external users

```json
// Product Registration Response
{
  "product": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "My Healthcare App",
    "status": "active"
  },
  "api_key": "apk_a1b2c3d4e5f6...",
  "api_secret": "aps_xyz789..."  // ⚠️ Only shown once!
}
```

### 2.2 External Users

The Appointment Service does **NOT** store user accounts. Instead, it uses `external_user_id` - an identifier from YOUR system.

| Property | Description |
| ---------- | ------------- |
| `external_user_id` | Any string up to 255 characters (UUID, email, database ID, etc.) |
| Treated as opaque | The service doesn't interpret or validate the format |
| Product-scoped | Same `external_user_id` in different Products are completely separate |

**Example:** Your system has user `john@example.com` with internal ID `user_12345`. When calling the Appointment Service, you pass `external_user_id: "user_12345"` to represent this user.

### 2.3 JWT Token Structure

```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "external_user_id": "user_12345",
  "role": "user",
  "iss": "appointment-service",
  "sub": "user_12345",
  "exp": 1710849600,
  "iat": 1710763200
}
```

### 2.4 RBAC Roles

| Role | Capabilities |
| ------ | ------------- |
| **admin** | Full access: manage all appointments, delete appointments, manage all availability rules, list all products |
| **provider** | Manage appointments, respond to appointments (confirm/complete), create/manage own availability rules |
| **user** | Create appointments, view own appointments, manage own participation, view availability |

### Role-Based Access Matrix

| Endpoint | user | provider | admin |
| ---------- | ------ | ---------- | ------- |
| `POST /v1/appointments` | ✅ | ✅ | ✅ |
| `GET /v1/appointments` | ✅ | ✅ | ✅ |
| `PATCH /v1/appointments/:id` | ✅* | ✅ | ✅ |
| `DELETE /v1/appointments/:id` | ❌ | ❌ | ✅ |
| `PATCH /v1/appointments/:id/response` | ❌ | ✅ | ✅ |
| `POST /v1/providers/:id/availability` | ❌ | ✅ | ✅ |
| `GET /v1/availability` | ✅ | ✅ | ✅ |
| `GET /v1/products` (list all) | ❌ | ❌ | ✅ |

*Users can only update appointments they created

---

## 3. Workflow Explanation

### 3.1 Complete Integration Flow

```md
┌──────────────────────────────────────────────────────────────────────────────┐
│                        PHASE 1: PRODUCT SETUP (One-time)                     │
└──────────────────────────────────────────────────────────────────────────────┘

Step 1: Register Your Product
─────────────────────────────
[Your Backend] ──POST /v1/products/register──> [Appointment Service]
                {name: "My App"}
                                              
                <── {api_key, api_secret} ───

     ⚠️ Store api_secret securely - it's only shown once!

┌──────────────────────────────────────────────────────────────────────────────┐
│                   PHASE 2: PROVIDER SETUP (Per Provider)                     │
└──────────────────────────────────────────────────────────────────────────────┘

Step 2: Generate Admin/Provider JWT Token
─────────────────────────────────────────
[Your Backend] ──POST /v1/auth/token──> [Appointment Service]
                {api_key, api_secret,
                 external_user_id: "dr_smith",
                 role: "provider"}
                                              
                <── {token: "eyJhbG..."} ───

Step 3: Create Availability Rules
─────────────────────────────────
[Your Backend] ──POST /v1/providers/dr_smith/availability──> [Appointment Service]
                Authorization: Bearer <token>
                {day_of_week: 1, start_time: "09:00", 
                 end_time: "17:00", duration_minutes: 30}
                                              
                <── {rule_id, ...} ───

┌──────────────────────────────────────────────────────────────────────────────┐
│                    PHASE 3: BOOKING FLOW (Per User Request)                  │
└──────────────────────────────────────────────────────────────────────────────┘

Step 4: Generate User JWT Token
───────────────────────────────
[Your Backend] ──POST /v1/auth/token──> [Appointment Service]
                {api_key, api_secret,
                 external_user_id: "customer_123",
                 role: "user"}
                                              
                <── {token: "eyJhbG..."} ───

Step 5: Check Available Slots
─────────────────────────────
[Your Backend] ──GET /v1/availability?provider_id=dr_smith&date=2026-03-20──>
                Authorization: Bearer <token>
                                              
                <── {slots: [{start_time, end_time, available: true}, ...]} ───

Step 6: Book Appointment
────────────────────────
[Your Backend] ──POST /v1/appointments/book──> [Appointment Service]
                Authorization: Bearer <token>
                {provider_id: "dr_smith",
                 start_time: "2026-03-20T10:00:00Z",
                 title: "Consultation",
                 participants: [...]}
                                              
                <── {appointment_id, status: "scheduled", ...} ───

┌──────────────────────────────────────────────────────────────────────────────┐
│                  PHASE 4: APPOINTMENT MANAGEMENT (Ongoing)                   │
└──────────────────────────────────────────────────────────────────────────────┘

Step 7: Provider Confirms Appointment
─────────────────────────────────────
[Your Backend] ──PATCH /v1/appointments/:id/response──> [Appointment Service]
                Authorization: Bearer <provider_token>
                {status: "confirmed"}
                                              
                <── {appointment with status: "confirmed"} ───

Step 8: Complete/Cancel as Needed
─────────────────────────────────
[Your Backend] ──POST /v1/appointments/:id/cancel──> [Appointment Service]
                Authorization: Bearer <token>
                                              
                <── {appointment with status: "cancelled"} ───
```

### 3.2 When to Use Each API

| API Endpoint | When to Use | Who Should Call |
| -------------- | ------------- | ----------------- |
| `POST /v1/products/register` | Initial setup, once per application | Your backend during deployment |
| `POST /v1/auth/token` | Every time a user needs to interact with appointments | Your backend (on behalf of user) |
| `POST /v1/providers/:id/availability` | Setting up provider schedules | Admin or Provider |
| `GET /v1/availability` | When user wants to see available times | Any authenticated user |
| `POST /v1/appointments/book` | User selects a time slot to book | Any authenticated user |
| `POST /v1/appointments` | Creating appointments without slot validation | Any authenticated user |
| `PATCH /v1/appointments/:id/response` | Provider confirms/completes appointment | Provider or Admin |
| `POST /v1/appointments/:id/cancel` | Cancelling an appointment | Creator, Provider, or Admin |

### 3.3 External User ID Usage Examples

```json
// Scenario: Healthcare app with doctors and patients

// Token for a patient (role: user)
{
  "api_key": "apk_xxx",
  "api_secret": "aps_xxx",
  "external_user_id": "patient_john_doe_12345",  // Your patient ID
  "role": "user"
}

// Token for a doctor (role: provider)
{
  "api_key": "apk_xxx",
  "api_secret": "aps_xxx",
  "external_user_id": "dr_smith_67890",  // Your doctor ID
  "role": "provider"
}

// Token for admin staff (role: admin)
{
  "api_key": "apk_xxx",
  "api_secret": "aps_xxx",
  "external_user_id": "admin_jane_99999",  // Your admin ID
  "role": "admin"
}
```

---

## 4. Postman Setup & Testing

### 4.1 Environment Setup

Create a new Postman Environment with these variables:

| Variable | Initial Value | Description |
| ---------- | -------------- | ------------- |
| `base_url` | `http://localhost:8080` | API base URL |
| `api_key` | (empty) | Filled after product registration |
| `api_secret` | (empty) | Filled after product registration |
| `jwt_token` | (empty) | Current JWT token |
| `product_id` | (empty) | Your product's UUID |
| `provider_id` | `dr-smith-001` | Example provider identifier |
| `external_user_id` | `patient-john-doe` | Example user identifier |
| `appointment_id` | (empty) | For appointment operations |

### 4.2 Collection Variables Script

Add this pre-request script to auto-populate dates:

```javascript
// Set today's date
const today = new Date();
pm.collectionVariables.set("today_date", today.toISOString().split('T')[0]);

// Set next Monday
const daysUntilMonday = (1 - today.getDay() + 7) % 7 || 7;
const nextMonday = new Date(today);
nextMonday.setDate(today.getDate() + daysUntilMonday);
pm.collectionVariables.set("next_monday", nextMonday.toISOString().split('T')[0]);
```

### 4.3 API Requests

#### 4.3.1 Register Product (Public - No Auth)

**POST** `{{base_url}}/v1/products/register`

```json
{
  "name": "My Test Application",
  "description": "Testing the appointment booking system",
  "webhook_url": "https://myapp.com/webhooks"
}
```

**Expected Response (201 Created):**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "api_key": "apk_a1b2c3d4e5f6...",
    "api_secret": "aps_xyz789...",
    "name": "My Test Application",
    "status": "active"
  }
}
```

**Post-request Script:**

```javascript
var jsonData = pm.response.json();
pm.collectionVariables.set("product_id", jsonData.data.id);
pm.collectionVariables.set("api_key", jsonData.data.api_key);
pm.collectionVariables.set("api_secret", jsonData.data.api_secret);
```

---

#### 4.3.2 Generate JWT Token (Public - No Auth)

**POST** `{{base_url}}/v1/auth/token`

```json
{
  "api_key": "{{api_key}}",
  "api_secret": "{{api_secret}}",
  "external_user_id": "{{external_user_id}}",
  "role": "admin"
}
```

**Expected Response (200 OK):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "token_type": "Bearer"
}
```

**Post-request Script:**

```javascript
var jsonData = pm.response.json();
pm.collectionVariables.set("jwt_token", jsonData.token);
```

---

#### 4.3.3 Create Availability Rule (Requires Auth - Provider/Admin)

**POST** `{{base_url}}/v1/providers/{{provider_id}}/availability`

**Headers:**

```md
Authorization: Bearer {{jwt_token}}
Content-Type: application/json
```

**Body:**

```json
{
  "day_of_week": 1,
  "start_time": "09:00",
  "end_time": "17:00",
  "duration_minutes": 30,
  "slot_interval_minutes": 15,
  "buffer_before_minutes": 5,
  "buffer_after_minutes": 5,
  "timezone": "UTC",
  "is_active": true
}
```

**Expected Response (201 Created):**

```json
{
  "data": {
    "id": "rule-uuid-here",
    "product_id": "product-uuid",
    "provider_id": "dr-smith-001",
    "day_of_week": 1,
    "start_time": "09:00:00",
    "end_time": "17:00:00",
    "duration_minutes": 30,
    "slot_interval_minutes": 15,
    "buffer_before_minutes": 5,
    "buffer_after_minutes": 5,
    "timezone": "UTC",
    "is_active": true
  }
}
```

**Field Reference:**

| Field | Description | Example |
| ------- | ------------- | --------- |
| `day_of_week` | 0=Sunday, 1=Monday, ..., 6=Saturday | `1` (Monday) |
| `start_time` | Daily availability start (HH:MM) | `"09:00"` |
| `end_time` | Daily availability end (HH:MM) | `"17:00"` |
| `duration_minutes` | Appointment length | `30` |
| `slot_interval_minutes` | Gap between slot start times | `15` |
| `buffer_before_minutes` | Prep time before appointment | `5` |
| `buffer_after_minutes` | Buffer after appointment | `5` |

---

#### 4.3.4 Get Available Slots (Requires Auth)

**GET** `{{base_url}}/v1/availability?provider_id={{provider_id}}&date={{next_monday}}`

**Headers:**

```md
Authorization: Bearer {{jwt_token}}
```

**Expected Response (200 OK):**

```json
{
  "data": {
    "provider_id": "dr-smith-001",
    "date": "2026-03-23",
    "timezone": "UTC",
    "duration_minutes": 30,
    "slots": [
      {
        "start_time": "2026-03-23T09:00:00Z",
        "end_time": "2026-03-23T09:30:00Z",
        "available": true
      },
      {
        "start_time": "2026-03-23T09:15:00Z",
        "end_time": "2026-03-23T09:45:00Z",
        "available": true
      }
    ]
  }
}
```

---

#### 4.3.5 Book Appointment (Requires Auth)

**POST** `{{base_url}}/v1/appointments/book`

**Headers:**

```md
Authorization: Bearer {{jwt_token}}
Content-Type: application/json
```

**Body:**

```json
{
  "provider_id": "{{provider_id}}",
  "start_time": "2026-03-23T09:00:00Z",
  "title": "Initial Consultation",
  "description": "First meeting with the provider",
  "location": "Room 101",
  "timezone": "UTC",
  "participants": [
    {
      "external_user_id": "{{external_user_id}}",
      "role": "guest",
      "user_metadata": {
        "name": "John Doe",
        "email": "john@example.com",
        "phone": "+1-555-123-4567"
      }
    }
  ]
}
```

**Expected Response (201 Created):**

```json
{
  "data": {
    "id": "appointment-uuid",
    "product_id": "product-uuid",
    "provider_id": "dr-smith-001",
    "title": "Initial Consultation",
    "status": "scheduled",
    "start_time": "2026-03-23T09:00:00Z",
    "end_time": "2026-03-23T09:30:00Z",
    "created_by": "patient-john-doe",
    "participants": [
      {
        "external_user_id": "patient-john-doe",
        "role": "guest",
        "status": "accepted"
      }
    ]
  }
}
```

---

#### 4.3.6 Create Appointment (Direct - Without Slot Validation)

**POST** `{{base_url}}/v1/appointments`

**Headers:**

```md
Authorization: Bearer {{jwt_token}}
Content-Type: application/json
```

**Body:**

```json
{
  "title": "Team Meeting",
  "description": "Weekly sync-up",
  "start_time": "2026-03-25T14:00:00Z",
  "end_time": "2026-03-25T15:00:00Z",
  "timezone": "America/New_York",
  "location": "Virtual - Zoom",
  "metadata": {
    "meeting_link": "https://zoom.us/j/123456789",
    "priority": "high"
  },
  "participants": [
    {
      "external_user_id": "{{external_user_id}}",
      "role": "host"
    },
    {
      "external_user_id": "colleague_456",
      "role": "attendee",
      "user_metadata": {
        "name": "Jane Smith",
        "email": "jane@example.com"
      }
    }
  ]
}
```

---

#### 4.3.7 Add Participant to Appointment

**POST** `{{base_url}}/v1/appointments/{{appointment_id}}/participants`

**Headers:**

```md
Authorization: Bearer {{jwt_token}}
Content-Type: application/json
```

**Body:**

```json
{
  "external_user_id": "new_participant_789",
  "role": "attendee",
  "user_metadata": {
    "name": "Bob Wilson",
    "email": "bob@example.com"
  }
}
```

**Expected Response (201 Created):**

```json
{
  "data": {
    "id": "participant-uuid",
    "appointment_id": "appointment-uuid",
    "external_user_id": "new_participant_789",
    "role": "attendee",
    "status": "pending"
  }
}
```

---

#### 4.3.8 Update Participant Status

**PATCH** `{{base_url}}/v1/appointments/{{appointment_id}}/participants/{{external_user_id}}/status`

**Headers:**

```md
Authorization: Bearer {{jwt_token}}
Content-Type: application/json
```

**Body:**

```json
{
  "status": "accepted"
}
```

**Status Options:** `pending`, `accepted`, `declined`, `tentative`

---

#### 4.3.9 Provider Responds to Appointment (Provider/Admin Only)

**PATCH** `{{base_url}}/v1/appointments/{{appointment_id}}/response`

**Headers:**

```md
Authorization: Bearer {{jwt_token}}
Content-Type: application/json
```

**Body:**

```json
{
  "status": "confirmed"
}
```

**Status Options:** `confirmed`, `completed`, `no_show`

---

#### 4.3.10 Cancel Appointment

**POST** `{{base_url}}/v1/appointments/{{appointment_id}}/cancel`

**Headers:**

```md
Authorization: Bearer {{jwt_token}}
```

**Expected Response (200 OK):**

```json
{
  "data": {
    "id": "appointment-uuid",
    "status": "cancelled"
  }
}
```

---

## 5. Testing Scenarios

### Scenario 1: Complete Booking Flow

```md
1. Register Product          → Save api_key, api_secret
2. Generate Admin Token      → role: "admin"
3. Create Availability Rules → Monday-Friday, 9AM-5PM
4. Generate User Token       → role: "user" 
5. Get Available Slots       → Check Monday availability
6. Book Appointment          → Select first slot
7. Verify Slot Unavailable   → Same slot should be blocked
```

### Scenario 2: Double Booking Prevention

```javascript
// Test Script for Postman
pm.test("Double booking returns 409 Conflict", function () {
    pm.response.to.have.status(409);
    var jsonData = pm.response.json();
    pm.expect(jsonData.error.toLowerCase()).to.include("conflict");
});
```

**Steps:**

1. Book an appointment at 9:00 AM
2. Try to book the same slot again
3. Expect: `409 Conflict` error

### Scenario 3: Role-Based Access Control

```md
Test 1: User tries to delete appointment
─────────────────────────────────────────
Token: role="user"
Request: DELETE /v1/appointments/:id
Expected: 403 Forbidden

Test 2: User tries to respond to appointment
─────────────────────────────────────────────
Token: role="user"
Request: PATCH /v1/appointments/:id/response
Expected: 403 Forbidden

Test 3: Provider creates availability rule
────────────────────────────────────────────
Token: role="provider"
Request: POST /v1/providers/:id/availability
Expected: 201 Created
```

### Scenario 4: Invalid JWT Token

**Request with expired/invalid token:**

```md
Authorization: Bearer invalid_or_expired_token
```

**Expected Response (401 Unauthorized):**

```json
{
  "error": "Unauthorized",
  "message": "invalid or expired token"
}
```

### Scenario 5: Multi-Tenant Isolation

```md
Test: Product A cannot see Product B's appointments
───────────────────────────────────────────────────
1. Register Product A           → Get credentials A
2. Create appointment (Product A)
3. Register Product B           → Get credentials B
4. Generate token for Product B
5. Try to GET the appointment from Product A
Expected: 404 Not Found (not 403 - don't leak existence)
```

### Scenario 6: Error Handling

| Scenario | Expected Status | Error Message |
| ---------- | ----------------- | --------------- |
| Invalid api_key/api_secret | 401 | "invalid API credentials" |
| Missing required field | 400 | "field X is required" |
| Invalid appointment ID | 400 | "invalid appointment ID" |
| Appointment not found | 404 | "appointment not found" |
| Time slot not available | 409 | "booking conflict" |
| Past date booking | 400 | "cannot book in the past" |
| Invalid role in token request | 400 | "invalid role" |

---

## 6. Tips & Best Practices

### 6.1 Reusing JWT Tokens

Tokens expire in **24 hours**. To efficiently test:

```javascript
// Pre-request script to check token expiry
const tokenExpiry = pm.collectionVariables.get("token_expiry");
const now = Math.floor(Date.now() / 1000);

if (!tokenExpiry || now > tokenExpiry - 300) { // 5 min buffer
    // Token expired or about to expire - generate new one
    pm.sendRequest({
        url: pm.collectionVariables.get("base_url") + "/v1/auth/token",
        method: 'POST',
        header: { 'Content-Type': 'application/json' },
        body: {
            mode: 'raw',
            raw: JSON.stringify({
                api_key: pm.collectionVariables.get("api_key"),
                api_secret: pm.collectionVariables.get("api_secret"),
                external_user_id: pm.collectionVariables.get("external_user_id"),
                role: "admin"
            })
        }
    }, function (err, response) {
        const json = response.json();
        pm.collectionVariables.set("jwt_token", json.token);
        pm.collectionVariables.set("token_expiry", now + json.expires_in);
    });
}
```

### 6.2 Simulating Multiple Users

Create tokens for different users with different roles:

```javascript
// Function to get token for specific user/role
function getTokenFor(userId, role, callback) {
    pm.sendRequest({
        url: pm.collectionVariables.get("base_url") + "/v1/auth/token",
        method: 'POST',
        header: { 'Content-Type': 'application/json' },
        body: {
            mode: 'raw',
            raw: JSON.stringify({
                api_key: pm.collectionVariables.get("api_key"),
                api_secret: pm.collectionVariables.get("api_secret"),
                external_user_id: userId,
                role: role
            })
        }
    }, callback);
}

// Usage
getTokenFor("patient_001", "user", (err, res) => {
    pm.collectionVariables.set("patient_token", res.json().token);
});

getTokenFor("dr_smith", "provider", (err, res) => {
    pm.collectionVariables.set("provider_token", res.json().token);
});
```

### 6.3 Testing Multi-Tenant Isolation

```javascript
// Test that confirms multi-tenant isolation
pm.test("Cannot access other product's data", function() {
    // After switching to Product B's token and trying to access Product A's appointment
    pm.response.to.have.status(404);
});
```

### 6.4 Validating Availability Rules

```javascript
// Validate slot generation
pm.test("Slots respect business rules", function() {
    var slots = pm.response.json().data.slots;
    
    slots.forEach(function(slot) {
        var start = new Date(slot.start_time);
        var end = new Date(slot.end_time);
        var durationMinutes = (end - start) / (1000 * 60);
        
        // Verify duration matches rule
        pm.expect(durationMinutes).to.equal(30);
        
        // Verify within business hours (9 AM - 5 PM UTC)
        pm.expect(start.getUTCHours()).to.be.at.least(9);
        pm.expect(end.getUTCHours()).to.be.at.most(17);
    });
});
```

### 6.5 Environment Management

Create separate environments for different purposes:

| Environment | base_url | Use Case |
| ------------- | ---------- | ---------- |
| Local | `http://localhost:8080` | Development testing |
| Docker | `http://localhost:8080` | Container testing |
| Staging | `https://api-staging.example.com` | Pre-production |
| Production | `https://api.example.com` | Production (read-only tests) |

### 6.6 Common Mistakes to Avoid

| Mistake | Solution |
| --------- | ---------- |
| Exposing api_secret to frontend | Always call /v1/auth/token from your backend |
| Using expired tokens | Implement token refresh logic |
| Hardcoding external_user_id | Map from your user system dynamically |
| Ignoring timezone | Always specify timezone in requests |
| Not validation slot alignment | Use slots returned by /v1/availability |

---

## Quick Reference Card

### API Endpoints Summary

```md
PUBLIC (No Auth):
  POST /v1/products/register     - Register new product
  POST /v1/products/validate     - Validate credentials
  POST /v1/auth/token            - Generate JWT token

AUTHENTICATED (JWT Required):
  GET  /v1/appointments          - List appointments
  POST /v1/appointments          - Create appointment
  GET  /v1/appointments/:id      - Get appointment
  PATCH /v1/appointments/:id     - Update appointment
  POST /v1/appointments/:id/cancel - Cancel appointment
  
  POST /v1/appointments/book     - Book with slot validation
  GET  /v1/availability          - Get available slots
  
  POST   /v1/providers/:id/availability      - Create rule
  GET    /v1/providers/:id/availability      - List rules
  PATCH  /v1/providers/:id/availability/:rid - Update rule
  DELETE /v1/providers/:id/availability/:rid - Delete rule

ADMIN/PROVIDER ONLY:
  DELETE /v1/appointments/:id           - Delete appointment (admin)
  PATCH  /v1/appointments/:id/response  - Confirm/complete (admin/provider)
```

### HTTP Status Codes

| Code | Meaning |
| ------ | --------- |
| 200 | Success |
| 201 | Created |
| 204 | No Content (successful delete) |
| 400 | Bad Request (invalid input) |
| 401 | Unauthorized (invalid/missing token) |
| 403 | Forbidden (insufficient role permissions) |
| 404 | Not Found |
| 409 | Conflict (double booking) |
| 500 | Internal Server Error |

---

## Appendix: Sample Postman Collection Import

Save this as `appointment-service.postman_collection.json` and import into Postman:

```json
{
  "info": {
    "name": "Appointment Service - Hands-On Guide",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {"key": "base_url", "value": "http://localhost:8080"},
    {"key": "api_key", "value": ""},
    {"key": "api_secret", "value": ""},
    {"key": "jwt_token", "value": ""},
    {"key": "product_id", "value": ""},
    {"key": "provider_id", "value": "dr-smith-001"},
    {"key": "external_user_id", "value": "patient-john-doe"},
    {"key": "appointment_id", "value": ""}
  ]
}
```

---

## Last updated: March 2026
