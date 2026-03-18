# Postman Testing Guide - Provider Availability & Booking System

## Quick Start

1. **Import Collection**: Import `Availability_System_Collection.postman_collection.json` into Postman
2. **Start Server**:

   ```powershell
   cd "c:\Users\SOKKER\Desktop\Appointment Project"
   go run ./cmd/api/
   ```

3. **Run Migrations** (if not done):

   ```powershell
   # Make sure PostgreSQL is running via Docker
   docker-compose up -d
   
   # Run migrations using migrate tool
   migrate -path migrations -database "postgres://postgres:postgres@localhost:5433/appointment_db?sslmode=disable" up
   ```

4. **Run Collection**: Execute requests in order (0 → 1 → 2 → 3 → 4 → 5 → 6)

## Collection Structure

### 0. Setup & Health

| Request | Description |
| --------- | ------------- |
| Health Check | Verify server is running |
| Register Product | Create product and get API credentials |
| Get JWT Token | Exchange credentials for auth token |

### 1. Availability Rules CRUD

| Request | Description |
| --------- | ------------- |
| Create Rule (Monday) | 9-17 UTC, 30min duration, 15min interval |
| Create Rule (Tuesday) | 10-18 New York timezone, 60min slots |
| Bulk Create (Wed-Fri) | Create multiple rules at once |
| List Rules | Get all provider rules |
| Get Single Rule | Get specific rule details |
| Update Rule | Modify Monday to 45min duration |

### 2. Slot Generation

| Request | Description |
| --------- | ------------- |
| Slots (Today) | Get today's availability |
| Slots (Next Monday) | Get slots with Monday rule |
| Slots (Past Date) | Should return 400 error |
| Slots (Sunday) | No rule - returns empty |

### 3. Booking Appointments

| Request | Expected Result |
| --------- | ----------------- |
| Book First Slot | 201 Created |
| Book Same Slot | 409 Conflict (double booking prevented) |
| Verify Slot Blocked | Booked slot not in list |
| Invalid Time | 400 - time not aligned with interval |
| Outside Window | 400 - before/after availability hours |

### 4. Buffer Time Verification

Verifies slots overlapping with buffer times are properly blocked.

### 5. Cancellation

| Request | Expected Result |
| --------- | ----------------- |
| Cancel Appointment | 200 - status changed to "cancelled" |
| Verify Slot Available | Slot appears in available list again |

### 6. Cleanup

Delete test rules to clean up.

## Variables Used

| Variable | Description | Auto-Set |
| ---------- | ------------- | ---------- |
| `base_url` | API base URL (default: `http://localhost:8080`) | No |
| `provider_id` | Provider identifier (default: `dr-smith-001`) | No |
| `user_id` | Test user ID | No |
| `jwt_token` | Auth token | Yes (from Get JWT Token) |
| `product_id` | Product UUID | Yes (from Register Product) |
| `api_key` | API Key | Yes (from Register Product) |
| `api_secret` | API Secret | Yes (from Register Product) |
| `rule_id_*` | Created rule IDs | Yes |
| `appointment_id` | Booked appointment ID | Yes |
| `slot_start_time` | First available slot | Yes |
| `next_monday` | Next Monday's date | Yes |

## Expected Test Results

When running the entire collection in order, you should see:

```md
✓ Health Check - Status 200
✓ Register Product - Status 201
✓ Get JWT Token - Status 200
✓ Create Rule (Monday) - Status 201
✓ Create Rule (Tuesday) - Status 201  
✓ Bulk Create - Status 201
✓ List Rules - Status 200 (5 rules)
✓ Get Single Rule - Status 200
✓ Update Rule - Status 200
✓ Get Slots (Today) - Status 200
✓ Get Slots (Monday) - Status 200 (has slots)
✓ Get Slots (Past) - Status 400
✓ Get Slots (Sunday) - Status 200 (empty)
✓ Book Appointment - Status 201
✓ Double Book - Status 409 (BLOCKED!)
✓ Verify Blocked - Slot not in list
✓ Invalid Time - Status 400
✓ Outside Window - Status 400
✓ Buffer Check - Overlapping slots blocked
✓ Cancel - Status 200
✓ Verify Available - Slot back in list
✓ Delete Rule - Status 200/204
```

## Key Features Tested

### 1. Dynamic Slot Generation

Slots are computed on-the-fly based on availability rules - never stored in the database.

### 2. Double Booking Prevention

PostgreSQL EXCLUSION CONSTRAINT ensures no overlapping appointments at the database level.

### 3. Buffer Times

5-minute buffers before/after appointments block adjacent slots.

### 4. Timezone Support

Rules can be set in any timezone (e.g., America/New_York), but all times returned are in UTC.

### 5. Concurrency Safety

REPEATABLE READ isolation with SELECT FOR UPDATE prevents race conditions.

## Troubleshooting

### "Server not running"

```powershell
go run ./cmd/api/
```

### "Database connection failed"

```powershell
docker-compose up -d
```

### "Migration errors"

```powershell
# Check if btree_gist extension is installed
docker exec -it appointment_postgres psql -U postgres -d appointment_db -c "CREATE EXTENSION IF NOT EXISTS btree_gist;"
```

### "JWT token invalid"

Re-run the "Get JWT Token" request to get a fresh token.
