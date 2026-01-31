# Appointment Service Analysis and Documentation

**Analysis Date:** January 31, 2026  
**Service Version:** Current  
**Analyzed By:** GitHub Copilot

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Data Models](#data-models)
4. [Repository Layer](#repository-layer)
5. [API Endpoints](#api-endpoints)
6. [Business Logic Analysis](#business-logic-analysis)
7. [Key Features](#key-features)
8. [Timezone Handling](#timezone-handling)
9. [Email Notifications](#email-notifications)
10. [Potential Issues and Recommendations](#potential-issues-and-recommendations)

---

## Overview

The Appointment Service is a comprehensive booking and scheduling system built with Go and Gin framework. It enables users to schedule appointments with partners (service providers), manage availability settings, and handle appointment lifecycle including notifications and calendar integration.

### Core Responsibilities

- **Appointment Management:** CRUD operations for appointments
- **Availability Configuration:** Partners can define their availability schedules
- **Time Slot Calculation:** Dynamic availability calculation based on bookings and settings
- **Calendar Integration:** Provides calendar views with availability
- **Notification System:** Email notifications for appointment events
- **Multi-language Support:** Arabic and English email templates

---

## Architecture

### Layered Architecture Pattern

```md
┌─────────────────────────────────────────┐
│         API Layer (Handlers)            │
│  - appointment-api.go                   │
│  - appointment-settings-api.go          │
│  - appintment-availabilty-api.go        │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│       Service Layer (Business Logic)     │
│  - Availability Calculation              │
│  - Time Slot Generation                  │
│  - Email Service Integration             │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│      Repository Layer (Data Access)      │
│  - appointement-repo.go                  │
│  - appointment-settings-repo.go          │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│         Data Layer (Database)            │
│  - appointments.appointments             │
│  - appointments.appointment_settings     │
│  - appointments.time_ranges              │
│  - appointments.date_ranges              │
└─────────────────────────────────────────┘
```

---

## Data Models

### 1. Appointment Model

**Location:** `App/model/appointment.go`

```go
type Appointment struct {
    Model                          // Embedded base model (ID, CreatedAt, UpdatedAt, DeletedAt)
    UserID    uint                 // User booking the appointment
    PartnerID uint                 // Partner/service provider
    Date      string               // Format: "2006-01-02"
    Time      string               // Format: "15:04:05"
    EndTime   string               // Format: "15:04:05"
    Duration  int                  // Duration in minutes
    Reason    string               // Purpose of appointment
    User      User                 // Relationship to user
    Partner   User                 // Relationship to partner
    Status    appointmentStatus    // 1=active, 2=inactive
}
```

**Table:** `appointments.appointments`

**Status Constants:**

- `active = 1` - Appointment is confirmed and active
- `inactive = 2` - Appointment is cancelled or inactive

### 2. AppointmentSetting Model

**Location:** `App/model/appointment-setting.go`

```go
type AppointmentSetting struct {
    Model
    Enabled              bool                  // Enable/disable appointment bookings
    UserID               uint                  // Partner user ID
    Duration             int                   // Default appointment duration (minutes)
    DateRange            int                   // How many days/weeks ahead bookings allowed
    DateRangeDaysOrWeeks DateRangeDaysOrWeeks  // "days" or "weeks"
    DateRangeType        DateRangeType         // "select_days" or "select_calendar"
    BufferTime           int                   // Buffer between appointments (minutes)
    DailyLimit           int                   // Max appointments per day
    MinNotice            string                // Minimum advance booking time (e.g., "2h")
    StartTimeIncrement   string                // Time slot intervals (e.g., "30m")
    AvailableDays        AvailableDays         // Weekly schedule configuration (JSON)
    TimeRanges           []TimeRange           // Detailed time ranges per day
    DateRanges           []DateRange           // Specific date range availability
}
```

**Table:** `appointments.appointment_settings`

### 3. TimeRange Model

Defines specific time slots for each day of the week.

```go
type TimeRange struct {
    ID                   uuid.UUID
    AppointmentSettingID uint
    Day                  string    // "monday", "tuesday", etc.
    StartTime            string    // "09:00"
    EndTime              string    // "17:00"
}
```

**Table:** `appointments.time_ranges`

### 4. DateRange Model

Defines specific date periods when appointments can be booked.

```go
type DateRange struct {
    ID                   uuid.UUID
    AppointmentSettingID uint
    StartDate            string    // "2026-01-01"
    EndDate              string    // "2026-12-31"
}
```

**Table:** `appointments.date_ranges`

### 5. AvailableDays Structure

Complex JSON structure stored in `appointment_settings.available_days`:

```go
type AvailableDays struct {
    Monday    AvailableDay
    Tuesday   AvailableDay
    Wednesday AvailableDay
    Thursday  AvailableDay
    Friday    AvailableDay
    Saturday  AvailableDay
    Sunday    AvailableDay
}

type AvailableDay struct {
    Enabled            bool
    StartTime          string
    EndTime            string
    TimeRange          []TimeRangeAvailableDay
    EventsHours        []EventsHour
    FullOfAppointments bool
}
```

---

## Repository Layer

### AppointmentRepo Interface

**Location:** `App/repositories/appointement-repo.go`

#### AppointmentRepo Methods

1. **GetAllAppointments()** - Retrieve all appointments with user/partner details
2. **CreateAppointment(data)** - Create new appointment
3. **UpdateAppointment(data)** - Update existing appointment
4. **DeleteAppointment(id)** - Delete appointment by ID
5. **GetAppointmentByID(id)** - Get single appointment with relations
6. **GetAppointmentsByUserID(userId)** - All appointments for a user
7. **GetAppointmentsByDate(date)** - All appointments on a specific date
8. **GetAppointmentsByStatus(status)** - Filter by status
9. **GetAppointmentsByUserIDAndDate(userId, date)** - User appointments on date
10. **GetAppointmentsByUserIDAndStatus(userId, status)** - User appointments by status
11. **GetAppointmentsByPartnerID(partnerId)** - All partner appointments
12. **GetAppointmentsByPartnerIDAndDate(partnerId, date)** - Partner appointments on date
13. **GetAppointmentByUserIDOrPartnerIDAndDate(userId, partnerId, date)** - Complex query
14. **GetUpcomingAppointmentsByUserID(userId)** - Future appointments from today
15. **GetOutgoingAppointments(userId, startDate, endDate, statuses)** - User's booked appointments
16. **GetIncomingAppointments(userId, startDate, endDate, statuses)** - Appointments booked with user
17. **Count()** - Total appointment count

### AppointmentSettingRepo Interface

**Location:** `App/repositories/appointment-settings-repo.go`

#### Methods

1. **Create(data)** - Create appointment settings
2. **FindAll()** - Get all settings with preloaded relations
3. **FindByID(id)** - Get settings by ID
4. **FindByUserID(id)** - Get user's appointment settings
5. **UserHasAppointmentSetting(id)** - Check if user has settings
6. **Update(data)** - Update settings with transaction handling
7. **Delete(data)** - Delete appointment settings
8. **Count()** - Total settings count

#### Key Implementation Detail

The **Update** method uses **transactions** to ensure atomicity:

- Updates main appointment setting fields
- Deletes existing TimeRanges
- Deletes existing DateRanges
- Creates new TimeRanges with fresh UUIDs
- Creates new DateRanges with fresh UUIDs
- Reloads data with associations

This prevents orphaned records and ensures data consistency.

---

## API Endpoints

### Appointment Management APIs

**File:** `App/apis/appointment-api.go`

| Endpoint | Method | Description | Auth Required |
| ---------- | -------- | ------------- | --------------- |
| `/appointments` | GET | Get all appointments (paginated) | ✓ |
| `/appointment` | POST | Create new appointment | ✓ |
| `/appointment` | PUT | Update appointment | ✓ |
| `/appointment/:id` | DELETE | Delete appointment | ✓ |
| `/appointment/:id` | GET | Get appointment by ID | ✓ |
| `/appointments/user/:id` | GET | Get user's appointments | ✓ |
| `/appointments/date/:date` | GET | Get appointments by date | ✓ |
| `/appointments/status/:status` | GET | Get appointments by status | ✓ |
| `/appointments/user/:id/date/:date` | GET | Get user appointments on date | ✓ |
| `/appointments/user/:id/status/:status` | GET | Get user appointments by status | ✓ |
| `/appointment/partner/:id` | GET | Get partner appointments | ✓ |
| `/appointments/partner/:id/date/:date` | GET | Get partner appointments on date | ✓ |
| `/appointment/availability/partner/{id}/date/{date}` | GET | Get partner date availability | ✓ |
| `/appointments/my-calendar` | GET | Get calendar with availability | ✓ |

### Appointment Settings APIs

**File:** `App/apis/appointment-settings-api.go`

| Endpoint | Method | Description | Auth Required |
| ---------- | -------- | ------------- | --------------- |
| `/appointment-settings` | GET | Get all appointment settings | ✓ |
| `/appointment-settings/{id}` | GET | Get setting by ID | ✓ |
| `/appointment-settings` | POST | Create/Update settings | ✓ |
| `/appointment-settings/{id}` | PUT | Update settings | ✓ |
| `/appointment-settings/{id}` | DELETE | Delete settings | ✓ |
| `/appointment-settings/user` | GET | Get current user's settings | ✓ |
| `/appointment-settings/user/{id}` | GET | Get user settings by ID | ✓ |

### Availability APIs

**File:** `App/apis/appintment-availabilty-api.go`

| Endpoint | Method | Description | Auth Required |
| ---------- | -------- | ------------- | --------------- |
| `/appointment_availability/:user_id` | GET | Get available days for user | ✓ |
| `/appointment_availability/:user_id/date/:date` | GET | Get available time slots for date | ✓ |

---

## Business Logic Analysis

### 1. Appointment Creation Flow

**Function:** `CreateAppointment()`

**Process:**

1. Parse appointment data from request
2. Get authenticated user ID
3. Set appointment user ID
4. Create appointment in database
5. Retrieve partner details
6. Retrieve user details
7. Determine email language preference from device settings
8. Send email notification to partner (APP01-EN or APP01-AR)
9. Send email notification to user (APP03-EN or APP03-AR)
10. Create audit log entry
11. Return created appointment

**Email Notifications:**

- Partner receives notification about new booking
- User receives confirmation of their booking
- Language determined by last device language setting

### 2. Availability Calculation Algorithm

**Function:** `GetUserAvailableAppointmentTimes()`

**Complex Algorithm Steps:**

1. **Load Settings:** Retrieve user's appointment settings
2. **Get Day of Week:** Parse date and determine weekday
3. **Filter Time Ranges:** Get time ranges for specific day
4. **Load Existing Appointments:** Get all appointments for user and requester on date
5. **Calculate Time Slots:**
   - Start from time range start time + buffer
   - End at time range end time - buffer
   - Increment by start time increment (e.g., 30 minutes)
6. **Apply Constraints:**
   - Skip times in the past
   - Skip times within minimum notice period
   - Skip times conflicting with existing appointments
   - Apply buffer time around existing appointments
7. **Return Available Slots:** List of available time slots with duration

**Timezone Handling:**

- Uses **Asia/Amman** timezone by default
- Converts current time to Jordan timezone
- Adjusts by subtracting 1 day for comparison (Note: This seems like a bug)

### 3. Calendar Integration

**Function:** `GetMyCalendar()`

**Features:**

- Date range query (max 180 days)
- Combines incoming and outgoing appointments
- Calculates daily availability
- Filters time ranges by appointment bookings
- Returns structured calendar view

**Response Structure:**

```go
type MyCalendarAPIResponse struct {
    Week      []DayCalendar      // Daily breakdown
    StartDate string
    EndDate   string
    Summary   CalendarSummary    // Overall statistics
}

type DayCalendar struct {
    Date                string
    DayOfWeek           string
    AsUser              AppointmentDayInfo    // Appointments I booked
    AsPartner           PartnerDayInfo        // Appointments booked with me
    IsAvailable         bool
    StartTime           string
    EndTime             string
    AvailableSlots      []AvailableHour
    AvailableSlotsCount int
}
```

### 4. Date Range Processing

**Function:** `getDaysBetweenDates()`

**Logic:**

1. Parse start and end dates
2. Add one day to end date for inclusive range
3. Get current date in Asia/Amman timezone
4. Adjust current date by -1 day (timezone adjustment)
5. Iterate from start to end date
6. Skip dates in the past
7. Return list of available dates in "2006-01-02" format

---

## Key Features

### 1. Flexible Scheduling

- **Multiple Time Ranges:** Partners can define multiple time slots per day
- **Date-Specific Availability:** Override weekly schedule with specific dates
- **Buffer Time:** Configurable breaks between appointments
- **Minimum Notice:** Prevent last-minute bookings

### 2. Smart Conflict Resolution

- Checks existing appointments for both user and partner
- Applies buffer time to prevent back-to-back bookings
- Considers appointment duration from settings
- Prevents double-booking

### 3. Multi-Role Support

- **User Role:** Books appointments with partners
- **Partner Role:** Provides services and manages availability
- **Dual Tracking:** System tracks both outgoing (as user) and incoming (as partner) appointments

### 4. Calendar Views

- **My Calendar:** Unified view of all appointments
- **Daily Breakdown:** Shows availability and bookings per day
- **Slot-Level Details:** Available time slots with start/end times

### 5. Query Optimization

- Preloads user and partner relationships
- Uses database-level joins for performance
- Supports pagination for large datasets
- Efficient filtering by date, status, and user

---

## Timezone Handling

### Current Implementation

**Timezone:** Asia/Amman (UTC+2/+3 with DST)

**Issues Identified:**

1. **Hardcoded Timezone:**

   ```go
   location, err := time.LoadLocation("Asia/Amman")
   ```

   - Not configurable per user
   - Assumes all users are in Jordan timezone

2. **Day Offset Bug:**

   ```go
   currentTime = time.Date(currentTime.Year(), currentTime.Month(), 
                           currentTime.Day()-1, currentTime.Hour(), ...)
   ```

   - Subtracts 1 day from current time
   - Likely incorrect timezone handling
   - Should use proper UTC conversion

3. **String-Based Date Storage:**
   - Dates stored as strings ("2006-01-02")
   - Times stored as strings ("15:04:05")
   - Makes timezone calculations more complex
   - Risk of timezone-related bugs

### Recommended Approach

1. Store dates/times in UTC in database
2. Add timezone field to user/partner settings
3. Convert to user's timezone for display
4. Remove the `-1 day` adjustment
5. Use proper timezone conversion functions

---

## Email Notifications

### Email Event Keys

**Partner Notifications:**

- `APP01-EN` - New appointment (English)
- `APP01-AR` - New appointment (Arabic)

**User Notifications:**

- `APP03-EN` - Appointment confirmation (English)
- `APP03-AR` - Appointment confirmation (Arabic)

### Email Data

```go
dataEmail := map[string]interface{}{
    "PartnerFName":      partner.FirstName,
    "PartnerLName":      partner.LastName,
    "UserFName":         user.FirstName,
    "UserLName":         user.LastName,
    "AppointmentDate":   appointment.Date,
    "AppointmentTime":   appointment.Time,
    "AppointmentReason": appointment.Reason,
}
```

### Language Detection

Language preference determined by:

1. Query device repository for user's last device
2. Get device language setting
3. Default to English if not found or not Arabic

### Email Addresses

- Partners: Use `Email` if `RegisterType` is email, otherwise `SocialEmail`
- Users: Same logic applies

---

## Potential Issues and Recommendations

### Critical Issues

#### 1. Timezone Handling Bug 🔴

**Issue:**

```go
currentTime = time.Date(currentTime.Year(), currentTime.Month(), 
                        currentTime.Day()-1, ...)
```

**Impact:** Available time slots may be calculated incorrectly

**Recommendation:**

```go
// Instead of subtracting a day, properly convert to UTC
currentTime = time.Now().UTC()
// Or if using Asia/Amman as base:
location, _ := time.LoadLocation("Asia/Amman")
currentTime = time.Now().In(location)
```

#### 2. Missing Transaction in Create 🔴

**Issue:** CreateAppointment doesn't validate availability before creating

**Impact:** Potential double-bookings if concurrent requests

**Recommendation:**

```go
// Wrap in transaction
err := db.Transaction(func(tx *gorm.DB) error {
    // Lock partner's appointments for this time slot
    // Check availability
    // If available, create appointment
    return nil
})
```

#### 3. No Appointment Duration Validation 🟡

**Issue:** EndTime not calculated or validated during creation

**Impact:** Appointments may not respect actual duration

**Recommendation:**

- Calculate EndTime from Time + Duration
- Validate EndTime doesn't exceed partner's availability
- Check overlap with existing appointments including EndTime

### Performance Issues

#### 4. N+1 Query Problem 🟡

**Issue:** GetAllAppointments uses joins but may still cause N+1

**Recommendation:**

- Ensure all relationships are properly preloaded
- Consider using explicit select fields to reduce data transfer
- Add database indexes on foreign keys

#### 5. Missing Indexes 🟡

**Recommended Indexes:**

```sql
CREATE INDEX idx_appointments_user_id ON appointments.appointments(user_id);
CREATE INDEX idx_appointments_partner_id ON appointments.appointments(partner_id);
CREATE INDEX idx_appointments_date ON appointments.appointments(date);
CREATE INDEX idx_appointments_status ON appointments.appointments(status);
CREATE INDEX idx_appointments_user_date ON appointments.appointments(user_id, date);
CREATE INDEX idx_appointments_partner_date ON appointments.appointments(partner_id, date);
```

### Security Issues

#### 6. Missing Authorization Checks 🔴

**Issue:** Most endpoints don't verify user permissions

**Examples:**

- User can access any partner's settings
- User can delete any appointment
- No check if user owns the appointment

**Recommendation:**

```go
// Check ownership before operations
if appointment.UserID != currentUserID && appointment.PartnerID != currentUserID {
    return errors.New("unauthorized")
}
```

#### 7. No Rate Limiting 🟡

**Issue:** No protection against appointment spam

**Recommendation:**

- Add rate limiting middleware
- Limit appointments per user per day
- Implement daily limit check during creation

### Data Integrity Issues

#### 8. Inconsistent Date Formats 🟡

**Issue:** Multiple date parsing attempts with different formats

**Recommendation:**

- Standardize on single date format
- Use database date types instead of strings
- Add validation at model level

#### 9. No Soft Delete Verification 🟡

**Issue:** Deleted appointments may still block time slots

**Recommendation:**

```go
// Add to queries
Where("deleted_at IS NULL")
```

### Usability Issues

#### 10. Limited Error Messages 🟡

**Issue:** Generic error messages don't help users

**Current:**

```go
c.JSON(500, gin.H{"error": err.Error()})
```

**Recommendation:**

```go
c.JSON(400, gin.H{
    "error": "slot_unavailable",
    "message": "The selected time slot is no longer available",
    "available_slots": calculateNearbySlots(),
})
```

#### 11. No Appointment Reminders 🟠

**Missing Feature:** No reminder system

**Recommendation:**

- Add background job for reminders
- Send email/push notification 24h and 1h before appointment
- Add reminder preferences to appointment settings

#### 12. No Cancellation Policy 🟠

**Missing Feature:** No cancellation rules or penalties

**Recommendation:**

- Add minimum cancellation notice to settings
- Implement cancellation policy
- Track cancellation history

---

## API Response Examples

### Get Available Time Slots

**Request:**

```md
GET /appointment_availability/123/date/2026-02-15
Authorization: Bearer {token}
```

**Response:**

```json
{
  "available_times": [
    "09:00:00",
    "09:30:00",
    "10:00:00",
    "10:30:00",
    "14:00:00",
    "14:30:00"
  ],
  "duration": 30
}
```

### Get My Calendar

**Request:**

```md
GET /appointments/my-calendar?start_date=2026-02-01&end_date=2026-02-07
Authorization: Bearer {token}
```

**Response:**

```json
{
  "week": [
    {
      "date": "2026-02-01",
      "day_of_week": "Sunday",
      "as_user": {
        "count": 1,
        "appointments": [...]
      },
      "as_partner": {
        "count": 2,
        "appointments": [...]
      },
      "is_available": true,
      "start_time": "09:00",
      "end_time": "17:00",
      "available_slots": [...],
      "available_slots_count": 12
    }
  ],
  "start_date": "2026-02-01",
  "end_date": "2026-02-07",
  "summary": {
    "total_appointments": 3,
    "as_user": 1,
    "as_partner": 2
  }
}
```

---

## Testing Recommendations

### Unit Tests Needed

1. **Time Slot Calculation**
   - Test with various buffer times
   - Test with minimum notice constraints
   - Test with existing appointments
   - Test timezone edge cases

2. **Date Range Processing**
   - Test date parsing with different formats
   - Test past date filtering
   - Test leap years and month boundaries

3. **Availability Logic**
   - Test with multiple time ranges
   - Test with date overrides
   - Test with fully booked slots

### Integration Tests Needed

1. **Concurrent Booking Prevention**
   - Simulate concurrent requests for same slot
   - Verify only one booking succeeds

2. **Email Notification Flow**
   - Verify correct language selection
   - Verify email content with correct data
   - Test fallback to social email

3. **Calendar API**
   - Test with various date ranges
   - Test with mixed appointment types
   - Verify availability calculations

---

## Performance Metrics to Monitor

1. **Average Response Time:** Target < 200ms for availability checks
2. **Database Query Count:** Should be minimal with proper preloading
3. **Concurrent Users:** Test with 100+ simultaneous bookings
4. **Calendar Load Time:** Target < 500ms for 30-day range

---

## Future Enhancement Suggestions

### Short Term (1-2 weeks)

1. Fix timezone handling bug
2. Add authorization checks
3. Implement proper transaction handling
4. Add database indexes

### Medium Term (1-2 months)

1. Implement appointment reminders
2. Add cancellation policy
3. Implement waiting list feature
4. Add appointment notes/attachments

### Long Term (3-6 months)

1. Video call integration
2. Recurring appointments
3. Group appointments
4. Advanced analytics dashboard
5. Mobile app optimization
6. Multi-timezone support
7. Payment integration

---

## Conclusion

The Appointment Service is a well-structured booking system with good separation of concerns. The architecture follows best practices with clear layers and comprehensive API coverage.

**Strengths:**

- ✅ Clean architecture with separation of concerns
- ✅ Comprehensive API endpoints
- ✅ Good database relationship design
- ✅ Multi-language support
- ✅ Flexible availability configuration
- ✅ Detailed calendar integration

**Areas for Improvement:**

- ⚠️ Timezone handling needs correction
- ⚠️ Authorization and security checks missing
- ⚠️ Transaction handling for concurrent bookings
- ⚠️ Performance optimizations needed
- ⚠️ Error handling and user feedback
- ⚠️ Testing coverage

**Priority Actions:**

1. Fix the timezone bug (Day-1 adjustment)
2. Add authorization middleware
3. Implement proper transaction handling
4. Add comprehensive unit tests
5. Create database indexes

---

**Document Version:** 1.0  
**Last Updated:** January 31, 2026  
**Maintained By:** Development Team
