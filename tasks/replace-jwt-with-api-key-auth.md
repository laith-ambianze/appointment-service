# Agent Task: Replace JWT Token Authentication with Per-Request API Key Authentication

**Status:** ✅ COMPLETED (March 31, 2026)

## Objective

Modify the authentication system to remove JWT Bearer tokens from API requests. Instead, authenticate every request using API key/secret headers while preserving the existing Role-Based Access Control (RBAC) system.

## Current State

- Authentication uses JWT Bearer tokens containing `product_id`, `external_user_id`, and `role`
- Users must first call `/v1/auth/token` with API credentials to get a JWT, then include it in subsequent requests
- RBAC middleware (`RequireAdmin`, `RequireProvider`, `RequireAdminOrProvider`) reads role from context set by JWT middleware
- Context keys are defined in `internal/middleware/jwt_auth.go`: `ContextKeyProductID`, `ContextKeyExternalUserID`, `ContextKeyRole`

## Target State

- Every authenticated request includes these headers:
  - `X-API-Key` — Product's API key
  - `X-API-Secret` — Product's API secret (validated against bcrypt hash)
  - `X-External-User-ID` — User identifier from the integrating system
  - `X-Role` — User's role (`admin`, `user`, or `provider`)
- No JWT tokens; no `/v1/auth/token` endpoint
- RBAC middleware continues to work unchanged

## Implementation Steps

### 1. Create new middleware file: `internal/middleware/api_key_auth.go`

Create `APIKeyAuth` middleware that:

- Extracts `X-API-Key`, `X-API-Secret`, `X-External-User-ID`, `X-Role` from request headers
- Validates API credentials using `ProductService.ValidateCredentials()`
- Validates that role is one of: `admin`, `user`, `provider`
- Sets context values using the existing keys: `ContextKeyProductID`, `ContextKeyExternalUserID`, `ContextKeyRole`
- Returns 401 for invalid/missing credentials
- Returns 400 for missing user ID or invalid role
- Supports skip paths for public endpoints

Use the existing `shouldSkipPath` function from `jwt_auth.go`.

### 2. Update `internal/routes/routes.go`

- Add `ProductService` to the `Config` struct
- Replace `middleware.JWTAuth(authConfig)` with `middleware.APIKeyAuth(apiKeyConfig)`
- Update skip paths to remove `/v1/auth/token`
- Remove the auth token route registration

### 3. Update `cmd/api/main.go`

- Pass `productService` to the routes config
- Remove `jwtManager` initialization and usage (or keep it but don't use for request auth)
- Remove `authHandler` if no longer needed

### 4. Remove or deprecate auth handler

- Delete `internal/handlers/auth.go` OR
- Keep it but remove the `GenerateToken` handler

### 5. Update CORS configuration

Add new headers to allowed headers in `.env.example` and `internal/middleware/cors.go`:

```md
X-API-Key, X-API-Secret, X-External-User-ID, X-Role
```

### 6. Update documentation

- Update `docs/EXTERNAL_USER_AUTHENTICATION.md` with new authentication flow
- Update `PRODUCT_BRIEF.md` to reflect header-based auth instead of JWT

## Files to Modify

1. `internal/middleware/api_key_auth.go` — **CREATE NEW**
2. `internal/middleware/jwt_auth.go` — Keep context key exports, can deprecate JWT functions
3. `internal/routes/routes.go` — Update middleware and config
4. `cmd/api/main.go` — Update initialization
5. `internal/handlers/auth.go` — Remove or deprecate
6. `internal/middleware/cors.go` — Add new headers
7. `.env.example` — Update CORS headers

## Constraints

- **Do NOT modify** `internal/middleware/role_auth.go` — RBAC must continue working
- **Do NOT modify** handlers, services, or repositories — they read from context which remains the same
- **Keep** `ProductService.ValidateCredentials()` — it already exists and validates api_key + api_secret
- **Keep** credential regeneration endpoint (`/v1/products/me/regenerate-credentials`)

## Testing Checklist

After implementation, verify:

1. `POST /v1/products/register` works without auth headers
2. `GET /health`, `/live`, `/ready` work without auth headers
3. `POST /v1/appointments` works with X-API-Key, X-API-Secret, X-External-User-ID, X-Role headers
4. `DELETE /v1/appointments/:id` returns 403 when X-Role is not `admin`
5. RBAC middleware (`RequireAdmin`, `RequireAdminOrProvider`) still enforces permissions correctly
6. Invalid API credentials return 401
7. Missing X-External-User-ID returns 400
8. Invalid X-Role returns 400

## Example Request After Implementation

```http
POST /v1/appointments HTTP/1.1
Host: localhost:8081
Content-Type: application/json
X-API-Key: apt_abc123
X-API-Secret: secret_xyz789
X-External-User-ID: user_456
X-Role: user

{
  "title": "Team Meeting",
  "start_time": "2026-04-01T10:00:00Z",
  "end_time": "2026-04-01T11:00:00Z",
  "timezone": "UTC",
  "participants": [
    {"external_user_id": "user_456", "role": "host"}
  ]
}
```

---

## Implementation Summary

### Files Created

- `internal/middleware/api_key_auth.go` — New API key authentication middleware with `ProductCredentialsValidator` interface

### Files Modified

1. **`internal/routes/routes.go`**
   - Removed `JWTManager` and `AuthHandler` from Config
   - Added `ProductValidator` (interface for API key validation)
   - Replaced `JWTAuth` middleware with `APIKeyAuth`
   - Removed `/v1/auth/token` route

2. **`cmd/api/main.go`**
   - Removed `jwtManager` and `authHandler` initialization
   - Updated `setupRoutes` to pass `productService` as `ProductValidator`
   - Removed `pkg/auth` import

3. **`internal/middleware/cors.go`**
   - Added `X-API-Secret`, `X-External-User-ID`, `X-Role` to allowed headers
   - Removed `Authorization` header

4. **`internal/config/config.go`**
   - Removed JWT_SECRET validation (no longer required)

5. **`.env.example`**
   - Commented out `JWT_SECRET`
   - Updated `CORS_ALLOWED_HEADERS`

6. **`PRODUCT_BRIEF.md`**
   - Updated authentication documentation to reflect API key auth

### RBAC Preserved

The existing role middleware (`RequireAdmin()`, `RequireProvider()`, `RequireAdminOrProvider()`) continues to work unchanged because the context keys (`ContextKeyProductID`, `ContextKeyExternalUserID`, `ContextKeyRole`) are set by the new `APIKeyAuth` middleware in the same way as the old JWT middleware.
