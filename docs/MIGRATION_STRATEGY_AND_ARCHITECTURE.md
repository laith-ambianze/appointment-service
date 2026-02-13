# Migration Strategy: From Monolithic to Multi-Tenant Appointment Microservice

**Document Type:** Strategic Architecture & Migration Plan  
**Date:** January 31, 2026  
**Purpose:** Transform existing appointment system into a general-purpose microservice

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current System Analysis](#current-system-analysis)
3. [New Architecture Vision](#new-architecture-vision)
4. [Key Differences: Old vs New](#key-differences-old-vs-new)
5. [Microservice Integration Patterns](#microservice-integration-patterns)
6. [Data Migration Strategy](#data-migration-strategy)
7. [Implementation Roadmap](#implementation-roadmap)
8. [API Gateway & Service Communication](#api-gateway--service-communication)
9. [Security & Multi-Tenancy](#security--multi-tenancy)
10. [Lessons Learned & Best Practices](#lessons-learned--best-practices)

---

## Executive Summary

### The Problem

Your current appointment service (analyzed in APPOINTMENT_SERVICE_ANALYSIS.md) is:

- ❌ **Tightly coupled** to a single product's user system
- ❌ **Limited flexibility** with hardcoded partner/user relationships
- ❌ **Product-specific** business logic mixed with appointment logic
- ❌ **Not reusable** across multiple products
- ❌ **Timezone issues** and security gaps

### The Solution

Build a **new general-purpose Appointment-as-a-Service** that:

- ✅ **Supports multiple products** with isolated data
- ✅ **Flexible participant model** (1-on-1, groups, observers)
- ✅ **External user references** - no user management
- ✅ **Microservice architecture** with clean APIs
- ✅ **SaaS-ready** for potential commercialization
- ✅ **Production-grade** security and performance

### Business Value

1. **Cost Reduction:** Build once, use everywhere
2. **Faster Time-to-Market:** New products get appointments instantly
3. **Consistency:** Same appointment logic across all products
4. **Scalability:** Independent scaling per product
5. **Revenue Opportunity:** Potential external SaaS offering

---

## Current System Analysis

### Architecture Issues

#### 1. Tight Coupling to User System

**Current Implementation:**

```go
type Appointment struct {
    UserID    uint   // Direct FK to users table
    PartnerID uint   // Direct FK to users table
    User      User   // GORM relationship
    Partner   User   // GORM relationship
}
```

**Problems:**

- Cannot work with other products' user systems
- Requires all products to use same user table
- User schema changes break appointments
- Cannot support external authentication

#### 2. Limited Participant Model

**Current:**

- Fixed two-person appointments (user + partner)
- No support for group appointments
- Roles hardcoded (user vs partner)
- Cannot add observers or optional attendees

**Impact:**

- Cannot handle team meetings
- Cannot support multi-provider appointments
- Limited business use cases

#### 3. Product-Specific Logic

**Current:**

```go
// Email logic embedded in appointment creation
emailService.SendEmail(partner.Email, "APP01-EN", dataEmail)
emailService.SendEmail(user.Email, "APP03-EN", dataEmail)
```

**Problems:**

- Email templates hardcoded in service
- Cannot customize per product
- Business logic mixed with appointment logic

#### 4. Timezone Hardcoding

**Current:**

```go
location, err := time.LoadLocation("Asia/Amman")
currentTime = time.Date(currentTime.Year(), currentTime.Month(), 
                        currentTime.Day()-1, ...)  // BUG!
```

**Problems:**

- Only works for Jordan timezone
- Day offset bug causes incorrect availability
- Cannot support international users

#### 5. No Multi-Tenancy

**Current:**

- Single product assumption
- No product isolation
- Cannot track which product owns appointments
- Security issues with cross-product access

---

## New Architecture Vision

### Core Principles

1. **Product Agnostic:** Works with any product's user system
2. **Metadata-First:** Store user details as JSONB, not relations
3. **Flexible Participants:** Support any number of participants with roles
4. **API-First:** Clean REST API with proper authentication
5. **Event-Driven:** Webhooks for product notifications
6. **Stateless:** Can scale horizontally

### Architecture Diagram

```md
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway (Optional)                    │
│                     Kong / Nginx / AWS ALB                       │
└────────────────────────────┬────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   Product A     │  │   Product B     │  │   Product C     │
│   Backend       │  │   Backend       │  │   Backend       │
│                 │  │                 │  │                 │
│  - User Mgmt    │  │  - User Mgmt    │  │  - User Mgmt    │
│  - Auth         │  │  - Auth         │  │  - Auth         │
│  - Business     │  │  - Business     │  │  - Business     │
└────────┬────────┘  └────────┬────────┘  └────────┬────────┘
         │                    │                     │
         │  Product API Key   │  Product API Key   │  Product API Key
         │  + External IDs    │  + External IDs    │  + External IDs
         │                    │                     │
         └────────────────────┼─────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────────┐
         │    Appointment Microservice (Go/Gin)       │
         │                                            │
         │  ┌──────────────────────────────────┐    │
         │  │   Authentication Middleware      │    │
         │  │   - Product API Key Validation   │    │
         │  │   - Rate Limiting per Product    │    │
         │  └──────────────────────────────────┘    │
         │                                            │
         │  ┌──────────────────────────────────┐    │
         │  │     Product Isolation Layer      │    │
         │  │   - All queries scoped by        │    │
         │  │     product_id                   │    │
         │  └──────────────────────────────────┘    │
         │                                            │
         │  ┌──────────────────────────────────┐    │
         │  │      Business Logic Layer        │    │
         │  │   - Appointment CRUD             │    │
         │  │   - Availability Calculation     │    │
         │  │   - Conflict Detection           │    │
         │  │   - Participant Management       │    │
         │  └──────────────────────────────────┘    │
         │                                            │
         │  ┌──────────────────────────────────┐    │
         │  │       Event System               │    │
         │  │   - Webhook Dispatcher           │    │
         │  │   - Event Queue (Optional)       │    │
         │  └──────────────────────────────────┘    │
         │                                            │
         └──────────────┬─────────────────────────────┘
                        │
                        ▼
         ┌────────────────────────────────────────────┐
         │      PostgreSQL Database (Isolated)         │
         │                                            │
         │  ┌──────────────────────────────────┐    │
         │  │  products (multi-tenant)         │    │
         │  │  - id, name, api_key, etc.       │    │
         │  └──────────────────────────────────┘    │
         │                                            │
         │  ┌──────────────────────────────────┐    │
         │  │  appointments                    │    │
         │  │  - product_id (isolation)        │    │
         │  │  - created_by (external ID)      │    │
         │  │  - metadata (JSONB)              │    │
         │  └──────────────────────────────────┘    │
         │                                            │
         │  ┌──────────────────────────────────┐    │
         │  │  appointment_participants        │    │
         │  │  - external_user_id              │    │
         │  │  - role (host/guest/attendee)    │    │
         │  │  - user_metadata (JSONB)         │    │
         │  └──────────────────────────────────┘    │
         │                                            │
         │  ┌──────────────────────────────────┐    │
         │  │  appointment_settings            │    │
         │  │  - product_id + external_user_id │    │
         │  │  - availability rules            │    │
         │  └──────────────────────────────────┘    │
         └────────────────────────────────────────────┘
```

---

## Key Differences: Old vs New

### Comparison Table

| Aspect | Old System | New System |
| -------- | ----------- | ------------ |
| **User Management** | Owns users in DB | References external user IDs |
| **Participant Model** | Fixed 2 people (user + partner) | Flexible N participants with roles |
| **Product Support** | Single product only | Multi-tenant (unlimited products) |
| **User Data** | Relational (User FK) | Metadata (JSONB) |
| **Authentication** | Shared with product | Product API keys |
| **Isolation** | None | Product-level data isolation |
| **Scalability** | Monolithic | Microservice (horizontal scaling) |
| **Timezone** | Hardcoded (Asia/Amman) | Per-product configurable |
| **Notifications** | Embedded email logic | Webhook events to products |
| **Group Meetings** | Not supported | Fully supported |
| **Availability** | Partner-specific | Per external user with product scope |
| **Integration** | Direct DB access | RESTful API |

### Code Comparison

#### Old Appointment Model

```go
type Appointment struct {
    Model
    UserID    uint        // FK to users table
    PartnerID uint        // FK to users table
    Date      string      // String date
    Time      string      // String time
    EndTime   string
    Duration  int
    Reason    string
    User      User        // GORM relation
    Partner   User        // GORM relation
    Status    appointmentStatus
}
```

#### New Appointment Model

```go
type Appointment struct {
    ID          uuid.UUID
    ProductID   uuid.UUID              // Multi-tenant isolation
    CreatedBy   string                  // External user ID
    Title       string
    Description string
    StartTime   time.Time               // Proper timestamp
    EndTime     time.Time
    Location    string
    MeetingType string                  // online/in-person/phone
    Status      string                  // pending/confirmed/cancelled
    Metadata    json.RawMessage         // Product-specific data
    Participants []Participant          // Flexible participants
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Participant struct {
    ID             uuid.UUID
    AppointmentID  uuid.UUID
    ExternalUserID string              // Product's user ID
    Role           string               // host/guest/attendee/observer
    UserMetadata   json.RawMessage     // User details from product
    Status         string               // pending/accepted/declined
}
```

#### Old API Call Pattern

```go
// Product makes direct DB call (tight coupling)
appointment := Appointment{
    UserID:    currentUser.ID,
    PartnerID: partner.ID,
    Date:      "2026-02-15",
    Time:      "10:00:00",
}
db.Create(&appointment)
```

#### New API Call Pattern

```go
// Product calls microservice API (loose coupling)
POST /v1/appointments
Headers:
  X-API-Key: prod_abc123
  X-API-Secret: secret_xyz789

Body:
{
  "title": "Consultation",
  "startTime": "2026-02-15T10:00:00Z",
  "endTime": "2026-02-15T11:00:00Z",
  "participants": [
    {
      "externalUserId": "user_123",
      "role": "host",
      "metadata": {
        "firstName": "Ahmed",
        "lastName": "Ali",
        "email": "ahmed@example.com",
        "phone": "+962...",
        "timezone": "Asia/Amman"
      }
    },
    {
      "externalUserId": "partner_456",
      "role": "guest",
      "metadata": {
        "firstName": "Sarah",
        "lastName": "Khan",
        "email": "sarah@example.com",
        "role": "consultant"
      }
    }
  ],
  "metadata": {
    "serviceType": "medical",
    "department": "cardiology",
    "insuranceId": "INS123"
  }
}
```

---

## Microservice Integration Patterns

### 1. Product Registration Flow

Step 1: **Product Registers with Appointment Service**

```bash
# One-time registration per product
POST /v1/products/register
Content-Type: application/json

{
  "name": "Healthcare Portal",
  "description": "Patient appointment system",
  "callbackUrl": "https://healthcare.example.com/webhooks/appointments",
  "settings": {
    "timezone": "Asia/Amman",
    "defaultDuration": 30,
    "language": "ar"
  }
}

Response:
{
  "success": true,
  "data": {
    "productId": "uuid",
    "apiKey": "prod_abc123xyz",
    "apiSecret": "secret_abc123xyz456",
    "message": "Store these credentials securely. The secret will not be shown again."
  }
}
```

Step 2: **Product Stores Credentials**

```go
// In Product A's backend
type AppointmentServiceConfig struct {
    BaseURL   string
    APIKey    string
    APISecret string
}

config := AppointmentServiceConfig{
    BaseURL:   "https://appointments.company.com/v1",
    APIKey:    os.Getenv("APPOINTMENT_API_KEY"),
    APISecret: os.Getenv("APPOINTMENT_API_SECRET"),
}
```

### 2. Authentication Pattern

**Every Request Requires Product Authentication:**

```go
// Product's HTTP client wrapper
type AppointmentClient struct {
    config AppointmentServiceConfig
    client *http.Client
}

func (ac *AppointmentClient) makeRequest(method, path string, body interface{}) (*http.Response, error) {
    req, _ := http.NewRequest(method, ac.config.BaseURL+path, jsonBody)
    
    // Add authentication headers
    req.Header.Set("X-API-Key", ac.config.APIKey)
    req.Header.Set("X-API-Secret", ac.config.APISecret)
    req.Header.Set("Content-Type", "application/json")
    
    return ac.client.Do(req)
}
```

### 3. Create Appointment Integration

**Product A's Backend API Handler:**

```go
// Product A's appointment creation endpoint
func (h *Handler) CreateAppointment(c *gin.Context) {
    var req CreateAppointmentRequest
    c.BindJSON(&req)
    
    // Get current user from product's auth
    currentUser := h.getCurrentUser(c)
    
    // Get partner details from product's DB
    partner := h.userRepo.GetByID(req.PartnerID)
    
    // Prepare appointment data for microservice
    appointmentReq := AppointmentServiceRequest{
        Title:       req.Title,
        StartTime:   req.StartTime,
        EndTime:     req.EndTime,
        Location:    req.Location,
        MeetingType: "in-person",
        Participants: []ParticipantRequest{
            {
                ExternalUserID: currentUser.ID,  // Your user ID
                Role:           "host",
                Metadata: map[string]interface{}{
                    "firstName": currentUser.FirstName,
                    "lastName":  currentUser.LastName,
                    "email":     currentUser.Email,
                    "phone":     currentUser.Phone,
                    "timezone":  "Asia/Amman",
                },
            },
            {
                ExternalUserID: partner.ID,
                Role:           "guest",
                Metadata: map[string]interface{}{
                    "firstName": partner.FirstName,
                    "lastName":  partner.LastName,
                    "email":     partner.Email,
                    "phone":     partner.Phone,
                },
            },
        },
        Metadata: map[string]interface{}{
            "serviceType":   req.ServiceType,
            "patientId":     currentUser.ID,
            "providerId":    partner.ID,
            "insuranceInfo": req.InsuranceInfo,
        },
    }
    
    // Call appointment microservice
    appointment, err := h.appointmentClient.CreateAppointment(appointmentReq)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to create appointment"})
        return
    }
    
    // Store appointment ID in your DB for reference (optional)
    h.repo.SaveAppointmentReference(currentUser.ID, appointment.ID)
    
    // Send your own notifications if needed
    h.emailService.SendAppointmentConfirmation(currentUser.Email, appointment)
    
    c.JSON(200, gin.H{"appointment": appointment})
}
```

### 4. Get User Appointments

**Product A Retrieves Appointments:**

```go
func (h *Handler) GetMyAppointments(c *gin.Context) {
    currentUser := h.getCurrentUser(c)
    
    // Call appointment microservice
    appointments, err := h.appointmentClient.GetUserAppointments(
        currentUser.ID,
        c.Query("startDate"),
        c.Query("endDate"),
        c.Query("status"),
    )
    
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to fetch appointments"})
        return
    }
    
    // Enrich with product-specific data if needed
    enriched := h.enrichAppointments(appointments)
    
    c.JSON(200, gin.H{"appointments": enriched})
}

func (h *Handler) enrichAppointments(appointments []Appointment) []EnrichedAppointment {
    enriched := make([]EnrichedAppointment, len(appointments))
    
    for i, apt := range appointments {
        // Get full user details from your DB
        participants := h.getUserDetails(apt.Participants)
        
        enriched[i] = EnrichedAppointment{
            Appointment:  apt,
            Participants: participants,
            // Add product-specific data
            ServiceDetails: h.getServiceDetails(apt.Metadata["serviceType"]),
        }
    }
    
    return enriched
}
```

### 5. Check Availability

**Product A Checks Partner Availability:**

```go
func (h *Handler) GetPartnerAvailability(c *gin.Context) {
    partnerID := c.Param("partnerId")
    date := c.Query("date")
    
    // Get partner's availability from appointment service
    availability, err := h.appointmentClient.GetUserAvailability(
        partnerID,
        date,
    )
    
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to check availability"})
        return
    }
    
    c.JSON(200, gin.H{
        "date":           date,
        "availableSlots": availability.TimeSlots,
        "duration":       availability.Duration,
    })
}
```

### 6. Webhook Integration (Event-Driven)

**Appointment Service Sends Events to Product:**

```go
// Appointment service sends webhook
POST https://healthcare.example.com/webhooks/appointments
Headers:
  X-Webhook-Signature: sha256(secret + body)
  Content-Type: application/json

Body:
{
  "event": "appointment.created",
  "timestamp": "2026-01-31T10:00:00Z",
  "data": {
    "appointmentId": "uuid",
    "productId": "uuid",
    "createdBy": "user_123",
    "participants": [
      {
        "externalUserId": "user_123",
        "role": "host"
      },
      {
        "externalUserId": "partner_456",
        "role": "guest"
      }
    ],
    "startTime": "2026-02-15T10:00:00Z",
    "status": "pending"
  }
}
```

**Product A's Webhook Handler:**

```go
func (h *Handler) HandleAppointmentWebhook(c *gin.Context) {
    // Verify webhook signature
    if !h.verifyWebhookSignature(c) {
        c.JSON(401, gin.H{"error": "Invalid signature"})
        return
    }
    
    var webhook AppointmentWebhook
    c.BindJSON(&webhook)
    
    switch webhook.Event {
    case "appointment.created":
        h.handleAppointmentCreated(webhook.Data)
    case "appointment.cancelled":
        h.handleAppointmentCancelled(webhook.Data)
    case "appointment.updated":
        h.handleAppointmentUpdated(webhook.Data)
    }
    
    c.JSON(200, gin.H{"received": true})
}

func (h *Handler) handleAppointmentCreated(data AppointmentData) {
    // Send custom notifications
    for _, participant := range data.Participants {
        user := h.userRepo.GetByID(participant.ExternalUserID)
        h.emailService.SendCustomNotification(user, data)
    }
    
    // Update local cache/records if needed
    h.cache.InvalidateUserAppointments(data.Participants)
    
    // Trigger product-specific workflows
    h.workflowService.OnAppointmentCreated(data)
}
```

### 7. SDK/Client Library Approach

**Create a Go SDK for Easy Integration:**

```go
// appointment-client/client.go
package appointmentclient

type Client struct {
    baseURL   string
    apiKey    string
    apiSecret string
    http      *http.Client
}

func NewClient(baseURL, apiKey, apiSecret string) *Client {
    return &Client{
        baseURL:   baseURL,
        apiKey:    apiKey,
        apiSecret: apiSecret,
        http:      &http.Client{Timeout: 10 * time.Second},
    }
}

// High-level methods
func (c *Client) CreateAppointment(req CreateAppointmentRequest) (*Appointment, error)
func (c *Client) GetAppointment(id string) (*Appointment, error)
func (c *Client) GetUserAppointments(userID string, opts QueryOptions) ([]Appointment, error)
func (c *Client) GetAvailability(userID, date string) (*Availability, error)
func (c *Client) CancelAppointment(id, reason string) error
func (c *Client) UpdateAppointmentSettings(userID string, settings Settings) error
```

**Product Uses SDK:**

```go
// In Product A's code
import "company.com/appointment-client"

client := appointmentclient.NewClient(
    os.Getenv("APPOINTMENT_SERVICE_URL"),
    os.Getenv("APPOINTMENT_API_KEY"),
    os.Getenv("APPOINTMENT_API_SECRET"),
)

appointment, err := client.CreateAppointment(appointmentclient.CreateAppointmentRequest{
    Title:     "Medical Consultation",
    StartTime: time.Now().Add(24 * time.Hour),
    Duration:  30,
    Participants: []appointmentclient.Participant{
        {UserID: "user_123", Role: "host", Metadata: userMeta},
        {UserID: "doc_456", Role: "guest", Metadata: docMeta},
    },
})
```

---

## Data Migration Strategy

### Phase 1: Preparation (Week 1-2)

#### 1.1 Analyze Current Data

```sql
-- Analyze existing appointments
SELECT 
    COUNT(*) as total_appointments,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT partner_id) as unique_partners,
    MIN(date) as earliest_appointment,
    MAX(date) as latest_appointment
FROM appointments.appointments
WHERE deleted_at IS NULL;

-- Analyze appointment settings
SELECT 
    COUNT(*) as total_settings,
    AVG(duration) as avg_duration,
    AVG(daily_limit) as avg_daily_limit
FROM appointments.appointment_settings;
```

#### 1.2 Create Product Record

```sql
-- Register existing product in new system
INSERT INTO products (id, name, description, api_key, api_secret_hash, is_active)
VALUES (
    gen_random_uuid(),
    'Healthcare Portal (Legacy)',
    'Migrated from old appointment system',
    'prod_legacy_healthcare',
    -- Generate and hash secret
    '$2a$10$...',
    true
);
```

### Phase 2: Schema Mapping (Week 2-3)

#### 2.1 Create Migration Script

```sql
-- Migration script: old_to_new_appointments.sql

-- Step 1: Migrate appointment_settings
INSERT INTO appointment_settings_new (
    id,
    product_id,
    external_user_id,
    enabled,
    duration,
    date_range,
    date_range_type,
    buffer_time,
    daily_limit,
    min_notice,
    start_time_increment,
    timezone,
    metadata,
    created_at,
    updated_at
)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM products WHERE name = 'Healthcare Portal (Legacy)'),
    CAST(user_id AS VARCHAR),  -- Convert to external user ID
    enabled,
    duration,
    date_range,
    date_range_days_or_weeks::VARCHAR,
    buffer_time,
    daily_limit,
    min_notice,
    start_time_increment,
    'Asia/Amman',  -- Set default timezone
    jsonb_build_object(
        'available_days', available_days,
        'legacy_id', id
    ),
    created_at,
    updated_at
FROM appointments.appointment_settings
WHERE deleted_at IS NULL;

-- Step 2: Migrate time_ranges
INSERT INTO time_ranges_new (
    id,
    appointment_setting_id,
    day,
    start_time,
    end_time,
    created_at,
    updated_at
)
SELECT 
    tr.id,
    asn.id,  -- New appointment_setting_id
    tr.day,
    tr.start_time,
    tr.end_time,
    tr.created_at,
    NOW()
FROM appointments.time_ranges tr
INNER JOIN appointments.appointment_settings as_old ON tr.appointment_setting_id = as_old.id
INNER JOIN appointment_settings_new asn ON CAST(as_old.user_id AS VARCHAR) = asn.external_user_id
WHERE as_old.deleted_at IS NULL;

-- Step 3: Migrate appointments
INSERT INTO appointments_new (
    id,
    product_id,
    created_by,
    title,
    description,
    start_time,
    end_time,
    location,
    meeting_type,
    status,
    metadata,
    created_at,
    updated_at
)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM products WHERE name = 'Healthcare Portal (Legacy)'),
    CAST(user_id AS VARCHAR),  -- User who created the appointment
    COALESCE(reason, 'Medical Appointment'),  -- Use reason as title
    reason,  -- Use reason as description
    -- Combine date and time into timestamp
    (date || ' ' || time)::TIMESTAMP,
    -- Combine date and end_time into end timestamp
    (date || ' ' || end_time)::TIMESTAMP,
    NULL,  -- No location in old system
    'in-person',  -- Default meeting type
    CASE 
        WHEN status = 1 THEN 'confirmed'
        WHEN status = 2 THEN 'cancelled'
        ELSE 'pending'
    END,
    jsonb_build_object(
        'legacy_appointment_id', id,
        'legacy_user_id', user_id,
        'legacy_partner_id', partner_id,
        'duration', duration
    ),
    created_at,
    updated_at
FROM appointments.appointments
WHERE deleted_at IS NULL;

-- Step 4: Create participants from appointments
-- Host participant (user who booked)
INSERT INTO appointment_participants_new (
    id,
    appointment_id,
    external_user_id,
    role,
    user_metadata,
    status,
    created_at,
    updated_at
)
SELECT 
    gen_random_uuid(),
    anew.id,
    CAST(aold.user_id AS VARCHAR),
    'host',
    jsonb_build_object(
        'userId', aold.user_id,
        'firstName', u.first_name,
        'lastName', u.last_name,
        'email', u.email,
        'phone', u.phone,
        'legacy_user_id', u.id
    ),
    'accepted',
    aold.created_at,
    NOW()
FROM appointments.appointments aold
INNER JOIN appointments_new anew ON (anew.metadata->>'legacy_appointment_id')::INT = aold.id
INNER JOIN users u ON u.id = aold.user_id
WHERE aold.deleted_at IS NULL;

-- Guest participant (partner/service provider)
INSERT INTO appointment_participants_new (
    id,
    appointment_id,
    external_user_id,
    role,
    user_metadata,
    status,
    created_at,
    updated_at
)
SELECT 
    gen_random_uuid(),
    anew.id,
    CAST(aold.partner_id AS VARCHAR),
    'guest',
    jsonb_build_object(
        'userId', aold.partner_id,
        'firstName', p.first_name,
        'lastName', p.last_name,
        'email', p.email,
        'phone', p.phone,
        'role', 'service_provider',
        'legacy_partner_id', p.id
    ),
    CASE 
        WHEN aold.status = 1 THEN 'accepted'
        ELSE 'pending'
    END,
    aold.created_at,
    NOW()
FROM appointments.appointments aold
INNER JOIN appointments_new anew ON (anew.metadata->>'legacy_appointment_id')::INT = aold.id
INNER JOIN users p ON p.id = aold.partner_id
WHERE aold.deleted_at IS NULL;
```

#### 2.2 Validation Queries

```sql
-- Verify migration counts match
SELECT 
    'Old System' as source,
    COUNT(*) as appointment_count
FROM appointments.appointments
WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'New System',
    COUNT(*)
FROM appointments_new;

-- Verify participants created correctly
SELECT 
    'Participants' as type,
    COUNT(*) as count
FROM appointment_participants_new;
-- Should be 2x appointments (host + guest for each)

-- Verify settings migrated
SELECT 
    'Settings Old' as source,
    COUNT(*)
FROM appointments.appointment_settings
WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'Settings New',
    COUNT(*)
FROM appointment_settings_new;
```

### Phase 3: Dual-Write Period (Week 3-6)

During transition, write to both old and new systems:

```go
func (s *AppointmentService) CreateAppointment(req CreateRequest) error {
    // Write to old system
    oldAppointment := convertToOldFormat(req)
    if err := s.oldRepo.Create(oldAppointment); err != nil {
        return err
    }
    
    // Write to new system (microservice)
    newAppointment := convertToNewFormat(req)
    if err := s.newClient.Create(newAppointment); err != nil {
        log.Error("Failed to sync to new system", err)
        // Don't fail the request, just log
    }
    
    return nil
}
```

### Phase 4: Read Migration (Week 6-8)

Gradually shift reads to new system:

```go
func (s *AppointmentService) GetUserAppointments(userID string) ([]Appointment, error) {
    // Try new system first
    appointments, err := s.newClient.GetUserAppointments(userID)
    if err != nil {
        log.Warn("New system unavailable, falling back to old system")
        return s.oldRepo.GetByUserID(userID)
    }
    
    return appointments, nil
}
```

### Phase 5: Cutover (Week 8)

1. **Stop writing to old system**
2. **Archive old data**
3. **Monitor new system closely**
4. **Keep old system as read-only backup for 30 days**

---

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)

#### Week 1-2: Core Microservice Setup

- ✅ Initialize Go project with proper structure
- ✅ Set up PostgreSQL with multi-tenant schema
- ✅ Implement product registration API
- ✅ Build authentication middleware
- ✅ Create basic CRUD for appointments
- ✅ Implement participant model

#### Week 3-4: Business Logic

- ✅ Availability calculation engine
- ✅ Conflict detection logic
- ✅ Time slot generation
- ✅ Appointment settings management
- ✅ Timezone support
- ✅ Buffer time and minimum notice

**Deliverable:** Working microservice with core APIs

### Phase 2: Integration (Weeks 5-8)

#### Week 5-6: Product Integration

- ✅ Build SDK/client library for products
- ✅ Create integration documentation
- ✅ Implement webhook system
- ✅ Build event dispatcher
- ✅ Create webhook signature verification

#### Week 7-8: First Product Integration

- ✅ Migrate Healthcare Portal (Product A)
- ✅ Test dual-write mode
- ✅ Validate data consistency
- ✅ Performance testing
- ✅ Load testing

**Deliverable:** One product successfully integrated

### Phase 3: Scale & Features (Weeks 9-12)

#### Week 9-10: Additional Products

- ✅ Migrate Product B
- ✅ Migrate Product C
- ✅ Optimize queries for scale
- ✅ Add caching layer (Redis)
- ✅ Implement rate limiting

#### Week 11-12: Advanced Features

- ✅ Recurring appointments
- ✅ Group appointments (3+ participants)
- ✅ Calendar sync (Google/Outlook)
- ✅ Reminder system
- ✅ Analytics dashboard

**Deliverable:** Production-ready multi-product service

### Phase 4: Optimization (Weeks 13-16)

#### Week 13-14: Performance

- ✅ Database indexing optimization
- ✅ Query performance tuning
- ✅ Connection pooling optimization
- ✅ API response caching
- ✅ Load balancing setup

#### Week 15-16: Monitoring & Ops

- ✅ Prometheus metrics
- ✅ Grafana dashboards
- ✅ Alert rules
- ✅ Log aggregation (ELK stack)
- ✅ Incident response procedures

**Deliverable:** Production-grade monitoring

---

## API Gateway & Service Communication

### Option 1: Direct Product-to-Microservice

**Simple approach for internal company use:**

```md
Product A ──────────┐
                    │
Product B ──────────┼───> Appointment Microservice
                    │      (Load Balanced)
Product C ──────────┘
```

**Pros:**

- ✅ Simple architecture
- ✅ Low latency
- ✅ Easy to debug

**Cons:**

- ❌ Products must handle authentication
- ❌ No centralized rate limiting
- ❌ No request routing

### Option 2: API Gateway (Recommended for SaaS)

**Enterprise approach with API Gateway:**

```md
Product A ──────────┐
                    │
Product B ──────────┼───> API Gateway ───> Appointment Microservice
                    │     (Kong/AWS ALB)      (Multiple Instances)
Product C ──────────┘
```

**API Gateway Features:**

1. **Authentication & Authorization**
   - Validates API keys
   - Manages product permissions
   - Handles OAuth if needed

2. **Rate Limiting**
   - Per-product limits
   - Prevent abuse
   - Fair usage policies

3. **Request Routing**
   - Route to different service versions
   - A/B testing
   - Canary deployments

4. **Monitoring & Analytics**
   - Request logging
   - Performance metrics
   - Usage analytics

5. **Transformation**
   - Request/response transformation
   - Protocol translation
   - Backward compatibility

**Kong Configuration Example:**

```yaml
services:
  - name: appointment-service
    url: http://appointment-service:8081
    routes:
      - name: appointments-route
        paths:
          - /v1/appointments
        strip_path: false
    plugins:
      - name: key-auth
        config:
          key_names:
            - X-API-Key
      - name: rate-limiting
        config:
          minute: 100
          hour: 5000
      - name: request-transformer
        config:
          add:
            headers:
              - X-Product-ID:$(api_key.product_id)
```

### Option 3: Service Mesh (Advanced)

**For complex microservice architectures:**

```md
┌─────────────────────────────────────────────┐
│            Service Mesh (Istio)             │
│                                             │
│  ┌─────────────┐    ┌──────────────────┐  │
│  │  Product A  │───▶│  Appointment Svc │  │
│  └─────────────┘    └──────────────────┘  │
│                              │              │
│  ┌─────────────┐    ┌──────────────────┐  │
│  │  Product B  │───▶│  Notification    │  │
│  └─────────────┘    │  Service         │  │
│                     └──────────────────┘  │
└─────────────────────────────────────────────┘
```

**Benefits:**

- Traffic management
- Service discovery
- Circuit breaking
- Distributed tracing

---

## Security & Multi-Tenancy

### 1. Product Isolation

**Database-Level Isolation:**

```go
// Every query must include product_id
type Repository interface {
    GetAppointments(ctx context.Context, productID uuid.UUID, filters Filters) ([]Appointment, error)
}

// Implementation enforces isolation
func (r *repository) GetAppointments(ctx context.Context, productID uuid.UUID, filters Filters) ([]Appointment, error) {
    query := `
        SELECT * FROM appointments 
        WHERE product_id = $1  -- ALWAYS filter by product
        AND deleted_at IS NULL
    `
    // ... additional filters
}
```

**Middleware Enforcement:**

```go
func ProductIsolationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get product from authenticated API key
        productID := c.GetString("product_id")
        
        // Inject into context
        ctx := context.WithValue(c.Request.Context(), "product_id", productID)
        c.Request = c.Request.WithContext(ctx)
        
        c.Next()
    }
}
```

### 2. API Key Management

**Secure Key Generation:**

```go
func GenerateAPIKey() (key, secret string, err error) {
    // Generate API key
    keyBytes := make([]byte, 32)
    rand.Read(keyBytes)
    key = fmt.Sprintf("prod_%s", base64.URLEncoding.EncodeToString(keyBytes))
    
    // Generate API secret
    secretBytes := make([]byte, 48)
    rand.Read(secretBytes)
    secret = base64.URLEncoding.EncodeToString(secretBytes)
    
    return key, secret, nil
}

func HashSecret(secret string) (string, error) {
    return bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
}
```

**Authentication Middleware:**

```go
func APIKeyAuthMiddleware(repo ProductRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        apiSecret := c.GetHeader("X-API-Secret")
        
        if apiKey == "" || apiSecret == "" {
            c.JSON(401, gin.H{"error": "Missing credentials"})
            c.Abort()
            return
        }
        
        // Get product by API key
        product, err := repo.GetByAPIKey(c.Request.Context(), apiKey)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid API key"})
            c.Abort()
            return
        }
        
        // Verify secret
        if !bcrypt.CompareHashAndPassword([]byte(product.APISecretHash), []byte(apiSecret)) {
            c.JSON(401, gin.H{"error": "Invalid API secret"})
            c.Abort()
            return
        }
        
        // Check if product is active
        if !product.IsActive {
            c.JSON(403, gin.H{"error": "Product is inactive"})
            c.Abort()
            return
        }
        
        // Store product info in context
        c.Set("product_id", product.ID)
        c.Set("product_name", product.Name)
        
        c.Next()
    }
}
```

### 3. Rate Limiting Per Product

```go
type RateLimiter struct {
    limiters map[uuid.UUID]*rate.Limiter
    mu       sync.RWMutex
}

func (rl *RateLimiter) GetLimiter(productID uuid.UUID) *rate.Limiter {
    rl.mu.RLock()
    limiter, exists := rl.limiters[productID]
    rl.mu.RUnlock()
    
    if !exists {
        rl.mu.Lock()
        limiter = rate.NewLimiter(100, 20) // 100 req/min, burst 20
        rl.limiters[productID] = limiter
        rl.mu.Unlock()
    }
    
    return limiter
}

func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        productID := c.GetString("product_id")
        
        limiter := rl.GetLimiter(uuid.MustParse(productID))
        
        if !limiter.Allow() {
            c.JSON(429, gin.H{
                "error": "rate_limit_exceeded",
                "retry_after": 60,
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 4. Data Encryption

**Encrypt Sensitive Metadata:**

```go
func EncryptMetadata(metadata map[string]interface{}, key []byte) ([]byte, error) {
    jsonData, _ := json.Marshal(metadata)
    
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    
    ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)
    return ciphertext, nil
}
```

---

## Lessons Learned & Best Practices

### From Old System Analysis

#### ❌ Anti-Patterns to Avoid

1. **Tight Coupling to User System**
   - Don't reference product's user tables directly
   - Use external user IDs as strings
   - Store user metadata as JSONB

2. **Hardcoded Business Logic**
   - Don't embed email templates in code
   - Use webhooks to let products handle notifications
   - Make rules configurable per product

3. **Timezone Assumptions**
   - Don't hardcode timezones
   - Store all times in UTC
   - Convert to user's timezone on display

4. **Missing Authorization**
   - Always verify ownership
   - Check product_id on every query
   - Implement role-based access

5. **String-Based Dates**
   - Use proper timestamp types
   - Store in UTC
   - Use Go time.Time, not strings

#### ✅ Best Practices to Follow

1. **API-First Design**
   - Design API before implementation
   - Document with OpenAPI/Swagger
   - Version your APIs (v1, v2)

2. **Metadata-First Approach**
   - Store product-specific data in JSONB
   - Keep core schema clean
   - Allow products to extend with custom fields

3. **Event-Driven Architecture**
   - Emit events for all important actions
   - Use webhooks for product notifications
   - Consider message queue for reliability

4. **Proper Transactions**
   - Use database transactions for multi-table operations
   - Implement optimistic locking for conflicts
   - Handle race conditions properly

5. **Comprehensive Testing**
   - Unit tests for business logic
   - Integration tests for APIs
   - Load tests for performance
   - Security tests for vulnerabilities

### Performance Optimization

#### Database Optimization

```sql
-- Essential indexes for multi-tenant queries
CREATE INDEX idx_appointments_product_id ON appointments(product_id);
CREATE INDEX idx_appointments_product_created_by ON appointments(product_id, created_by);
CREATE INDEX idx_appointments_product_status ON appointments(product_id, status);
CREATE INDEX idx_appointments_product_time ON appointments(product_id, start_time);
CREATE INDEX idx_appointments_time_range ON appointments(start_time, end_time);

-- Participant lookup optimization
CREATE INDEX idx_participants_appointment ON appointment_participants(appointment_id);
CREATE INDEX idx_participants_product_user ON appointment_participants(
    appointment_id, 
    external_user_id
) WHERE (SELECT product_id FROM appointments WHERE id = appointment_id) = ?;

-- Settings lookup
CREATE INDEX idx_settings_product_user ON appointment_settings(product_id, external_user_id);

-- JSONB indexing for metadata queries
CREATE INDEX idx_appointments_metadata ON appointments USING GIN (metadata);
CREATE INDEX idx_participants_metadata ON appointment_participants USING GIN (user_metadata);
```

#### Query Optimization

```go
// Bad: N+1 query problem
func GetAppointments(productID uuid.UUID) ([]Appointment, error) {
    appointments, _ := repo.Find(productID)
    for i := range appointments {
        appointments[i].Participants = repo.GetParticipants(appointments[i].ID)
    }
    return appointments
}

// Good: Preload relationships
func GetAppointments(productID uuid.UUID) ([]Appointment, error) {
    query := `
        SELECT 
            a.*,
            json_agg(
                json_build_object(
                    'id', p.id,
                    'external_user_id', p.external_user_id,
                    'role', p.role,
                    'user_metadata', p.user_metadata
                )
            ) as participants
        FROM appointments a
        LEFT JOIN appointment_participants p ON a.id = p.appointment_id
        WHERE a.product_id = $1
        GROUP BY a.id
    `
    return repo.Query(query, productID)
}
```

#### Caching Strategy

```go
// Cache frequently accessed data
type CacheLayer struct {
    redis *redis.Client
}

func (c *CacheLayer) GetUserAppointments(productID, userID string, date string) ([]Appointment, error) {
    cacheKey := fmt.Sprintf("appointments:%s:%s:%s", productID, userID, date)
    
    // Try cache first
    cached, err := c.redis.Get(cacheKey).Result()
    if err == nil {
        var appointments []Appointment
        json.Unmarshal([]byte(cached), &appointments)
        return appointments, nil
    }
    
    // Cache miss - query database
    appointments, err := c.repo.GetUserAppointments(productID, userID, date)
    if err != nil {
        return nil, err
    }
    
    // Store in cache with 5-minute TTL
    data, _ := json.Marshal(appointments)
    c.redis.Set(cacheKey, data, 5*time.Minute)
    
    return appointments, nil
}

// Invalidate cache on updates
func (c *CacheLayer) CreateAppointment(appointment *Appointment) error {
    err := c.repo.Create(appointment)
    if err != nil {
        return err
    }
    
    // Invalidate all affected cache keys
    for _, participant := range appointment.Participants {
        pattern := fmt.Sprintf("appointments:%s:%s:*", 
            appointment.ProductID, participant.ExternalUserID)
        c.redis.Del(pattern)
    }
    
    return nil
}
```

### Monitoring & Observability

#### Key Metrics to Track

```go
// Prometheus metrics
var (
    appointmentCreated = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "appointments_created_total",
            Help: "Total number of appointments created",
        },
        []string{"product_id", "status"},
    )
    
    appointmentDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "appointment_operation_duration_seconds",
            Help:    "Duration of appointment operations",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation", "product_id"},
    )
    
    availabilityCheckDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "availability_check_duration_seconds",
            Help:    "Time to check availability",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1},
        },
    )
)

// Usage in handlers
func (h *Handler) CreateAppointment(c *gin.Context) {
    timer := prometheus.NewTimer(appointmentDuration.WithLabelValues("create", productID))
    defer timer.ObserveDuration()
    
    // ... create appointment logic
    
    appointmentCreated.WithLabelValues(productID, "success").Inc()
}
```

#### Distributed Tracing

```go
import "go.opentelemetry.io/otel"

func (s *Service) CreateAppointment(ctx context.Context, req CreateRequest) error {
    ctx, span := otel.Tracer("appointment-service").Start(ctx, "CreateAppointment")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("product.id", req.ProductID),
        attribute.Int("participants.count", len(req.Participants)),
    )
    
    // ... business logic with child spans
    
    return nil
}
```

---

## Deployment Architecture

### Docker Compose (Development)

```yaml
version: '3.8'

services:
  appointment-service:
    build: .
    ports:
      - "8081:8081"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    networks:
      - appointment-network

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=appointments
      - POSTGRES_USER=admin
      - POSTGRES_PASSWORD=secret
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - appointment-network

  redis:
    image: redis:7-alpine
    networks:
      - appointment-network

  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"
    networks:
      - appointment-network

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    networks:
      - appointment-network

networks:
  appointment-network:

volumes:
  postgres-data:
```

### Kubernetes (Production)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: appointment-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: appointment-service
  template:
    metadata:
      labels:
        app: appointment-service
    spec:
      containers:
      - name: appointment-service
        image: company/appointment-service:v1.0.0
        ports:
        - containerPort: 8081
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: host
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: appointment-service
spec:
  selector:
    app: appointment-service
  ports:
  - port: 80
    targetPort: 8081
  type: LoadBalancer
```

---

## Success Criteria

### Technical Metrics

- ✅ **API Latency:** p95 < 100ms, p99 < 200ms
- ✅ **Availability:** 99.9% uptime
- ✅ **Throughput:** 1000+ req/sec per instance
- ✅ **Error Rate:** < 0.1% of requests
- ✅ **Database Connections:** Efficient pooling, no leaks

### Business Metrics

- ✅ **Product Adoption:** 3+ products integrated in first quarter
- ✅ **Data Accuracy:** 100% appointment sync between systems
- ✅ **User Satisfaction:** Reduced appointment booking time by 50%
- ✅ **Cost Savings:** 70% reduction in development time for new products
- ✅ **Scalability:** Support 10,000+ concurrent users

### Migration Success

- ✅ **Zero Data Loss:** All historical appointments migrated
- ✅ **Zero Downtime:** Seamless cutover with dual-write
- ✅ **Feature Parity:** All old features available in new system
- ✅ **Performance Improvement:** 3x faster availability checks
- ✅ **Team Confidence:** Successful rollback procedure tested

---

## Next Steps

### Immediate Actions (This Week)

1. ✅ Review this document with team
2. ✅ Finalize architecture decisions
3. ✅ Set up development environment
4. ✅ Create project repository
5. ✅ Initialize Go module structure

### Short Term (Next 2 Weeks)

1. ✅ Implement core database schema
2. ✅ Build authentication system
3. ✅ Create basic CRUD APIs
4. ✅ Write unit tests
5. ✅ Set up CI/CD pipeline

### Medium Term (Next 2 Months)

1. ✅ Complete business logic implementation
2. ✅ Build SDK/client library
3. ✅ Integrate first product
4. ✅ Perform load testing
5. ✅ Set up monitoring

### Long Term (Next 6 Months)

1. ✅ Migrate all products
2. ✅ Implement advanced features
3. ✅ Optimize performance
4. ✅ Prepare for SaaS launch
5. ✅ Build customer portal

---

## Conclusion

This migration from a monolithic appointment system to a general-purpose microservice represents a significant architectural improvement. By learning from the limitations of the old system and embracing modern microservice patterns, you'll create a flexible, scalable foundation that can serve multiple products and potentially become a commercial SaaS offering.

**Key Takeaways:**

1. **Decouple from user systems** - Use external IDs and metadata
2. **Design for multi-tenancy** - Product isolation from day one
3. **API-first approach** - Clean contracts between services
4. **Event-driven integration** - Webhooks for product notifications
5. **Learn from the past** - Address timezone, security, and performance issues
6. **Plan for scale** - Build with SaaS in mind

The roadmap is clear, the architecture is solid, and the benefits are substantial. Time to build! 🚀

---

**Document Status:** Final Draft  
**Next Review:** After Phase 1 completion  
**Maintained By:** Architecture Team
