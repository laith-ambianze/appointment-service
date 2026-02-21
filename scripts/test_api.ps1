#!/usr/bin/env pwsh
# API Test Script for Appointment Service
# This script tests all major API endpoints

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$JwtSecret = "change-me-in-production"
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { param($msg) Write-Host "[PASS] $msg" -ForegroundColor Green }
function Write-Fail { param($msg) Write-Host "[FAIL] $msg" -ForegroundColor Red }
function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Step { param($msg) Write-Host "`n[STEP] $msg" -ForegroundColor Yellow }

Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  Appointment Service API Test Script" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

# Check if server is running
Write-Step "Checking if server is running..."
try {
    $health = Invoke-RestMethod -Uri "$BaseUrl/health" -Method GET -TimeoutSec 5
    Write-Success "Server is running: $($health.status)"
} catch {
    Write-Fail "Server is not running at $BaseUrl"
    Write-Host "Please start the server with: go run ./cmd/api/" -ForegroundColor Yellow
    exit 1
}

# Generate a test JWT token using Go
Write-Step "Generating test JWT token..."

$goCode = @'
package main

import (
	"fmt"
	"os"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func main() {
	secret := os.Args[1]
	productID := uuid.New()
	
	claims := jwt.MapClaims{
		"product_id": productID.String(),
		"user_id":    "test-user-123",
		"role":       os.Args[2],
		"iss":        "appointment-service",
		"sub":        "test-user-123",
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	fmt.Print(tokenString)
}
'@

# Save temp Go file
$tempDir = Join-Path $env:TEMP "jwt-gen-$(Get-Random)"
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
$tempFile = Join-Path $tempDir "main.go"
$goCode | Out-File -FilePath $tempFile -Encoding UTF8

# Generate tokens for different roles
Push-Location $tempDir
try {
    go mod init temp 2>$null
    go get github.com/golang-jwt/jwt/v5 2>$null
    go get github.com/google/uuid 2>$null
    
    $userToken = go run . $JwtSecret "user" 2>$null
    $adminToken = go run . $JwtSecret "admin" 2>$null
    $providerToken = go run . $JwtSecret "provider" 2>$null
    
    Write-Success "Generated tokens for user, admin, and provider roles"
} finally {
    Pop-Location
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}

$headers = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer $userToken"
}

$adminHeaders = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer $adminToken"
}

$providerHeaders = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer $providerToken"
}

# ===== Test Health Endpoints =====
Write-Step "Testing Health Endpoints..."

try {
    $live = Invoke-RestMethod -Uri "$BaseUrl/live" -Method GET
    Write-Success "GET /live - Status: $($live.status)"
} catch {
    Write-Fail "GET /live failed: $_"
}

try {
    $ready = Invoke-RestMethod -Uri "$BaseUrl/ready" -Method GET
    Write-Success "GET /ready - Status: $($ready.status)"
} catch {
    Write-Fail "GET /ready failed: $_"
}

# ===== Test CORS =====
Write-Step "Testing CORS Headers..."

try {
    $corsHeaders = @{
        "Origin" = "http://localhost:3000"
    }
    $response = Invoke-WebRequest -Uri "$BaseUrl/health" -Method OPTIONS -Headers $corsHeaders
    $allowOrigin = $response.Headers["Access-Control-Allow-Origin"]
    if ($allowOrigin) {
        Write-Success "CORS headers present: Access-Control-Allow-Origin = $allowOrigin"
    } else {
        Write-Fail "CORS headers missing"
    }
} catch {
    Write-Info "OPTIONS request returned: $($_.Exception.Response.StatusCode)"
}

# ===== Test Authentication =====
Write-Step "Testing Authentication..."

# Test without token
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments" -Method GET -ErrorAction Stop
    Write-Fail "GET /v1/appointments without token should fail"
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Success "Unauthenticated request correctly rejected (401)"
    } else {
        Write-Fail "Unexpected error: $_"
    }
}

# Test with valid token
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments" -Method GET -Headers $headers
    Write-Success "Authenticated request succeeded"
} catch {
    Write-Fail "Authenticated request failed: $_"
}

# ===== Test Create Appointment =====
Write-Step "Testing Create Appointment..."

$startTime = (Get-Date).AddDays(1).ToString("yyyy-MM-ddTHH:mm:ssZ")
$endTime = (Get-Date).AddDays(1).AddHours(1).ToString("yyyy-MM-ddTHH:mm:ssZ")

$appointmentData = @{
    title = "Test Appointment"
    description = "Created by API test script"
    start_time = $startTime
    end_time = $endTime
    timezone = "UTC"
    location = "Conference Room A"
    participants = @(
        @{
            external_user_id = "participant-1"
            role = "guest"
        }
    )
} | ConvertTo-Json -Depth 3

try {
    $appointment = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments" -Method POST -Headers $headers -Body $appointmentData
    Write-Success "Created appointment: $($appointment.id)"
    $appointmentId = $appointment.id
} catch {
    Write-Fail "Create appointment failed: $_"
    $appointmentId = $null
}

# ===== Test Get Appointment =====
if ($appointmentId) {
    Write-Step "Testing Get Appointment..."
    
    try {
        $fetched = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments/$appointmentId" -Method GET -Headers $headers
        Write-Success "Fetched appointment: $($fetched.title)"
    } catch {
        Write-Fail "Get appointment failed: $_"
    }
}

# ===== Test Update Appointment =====
if ($appointmentId) {
    Write-Step "Testing Update Appointment..."
    
    $updateData = @{
        title = "Updated Test Appointment"
        description = "Updated by API test script"
    } | ConvertTo-Json
    
    try {
        $updated = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments/$appointmentId" -Method PATCH -Headers $headers -Body $updateData
        Write-Success "Updated appointment title to: $($updated.title)"
    } catch {
        Write-Fail "Update appointment failed: $_"
    }
}

# ===== Test Respond to Appointment (Admin/Provider) =====
if ($appointmentId) {
    Write-Step "Testing Respond to Appointment..."
    
    # User should not be able to respond
    $respondData = @{ status = "confirmed" } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments/$appointmentId/response" -Method PATCH -Headers $headers -Body $respondData -ErrorAction Stop
        Write-Fail "User should not be able to respond to appointments"
    } catch {
        if ($_.Exception.Response.StatusCode -eq 403) {
            Write-Success "User correctly forbidden from responding (403)"
        }
    }
    
    # Admin should be able to respond
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments/$appointmentId/response" -Method PATCH -Headers $adminHeaders -Body $respondData
        Write-Success "Admin confirmed appointment: status = $($response.status)"
    } catch {
        Write-Fail "Admin respond failed: $_"
    }
}

# ===== Test Add Participant =====
if ($appointmentId) {
    Write-Step "Testing Add Participant..."
    
    $participantData = @{
        external_user_id = "new-participant-123"
        role = "attendee"
        user_metadata = @{
            name = "John Doe"
            email = "john@example.com"
        }
    } | ConvertTo-Json -Depth 3
    
    try {
        $participant = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments/$appointmentId/participants" -Method POST -Headers $headers -Body $participantData
        Write-Success "Added participant: $($participant.external_user_id)"
    } catch {
        Write-Fail "Add participant failed: $_"
    }
}

# ===== Test Update Participant Status =====
if ($appointmentId) {
    Write-Step "Testing Update Participant Status..."
    
    $statusData = @{ status = "accepted" } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments/$appointmentId/participants/new-participant-123/status" -Method PATCH -Headers $adminHeaders -Body $statusData
        Write-Success "Updated participant status: $($response.status)"
    } catch {
        Write-Fail "Update participant status failed: $_"
    }
}

# ===== Test Cancel Appointment =====
if ($appointmentId) {
    Write-Step "Testing Cancel Appointment..."
    
    try {
        $cancelled = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments/$appointmentId/cancel" -Method POST -Headers $headers
        Write-Success "Cancelled appointment: status = $($cancelled.status)"
    } catch {
        Write-Fail "Cancel appointment failed: $_"
    }
}

# ===== Test Delete Appointment (Admin Only) =====
Write-Step "Testing Delete Appointment..."

# Create a new appointment to delete
try {
    $newAppt = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments" -Method POST -Headers $adminHeaders -Body $appointmentData
    $deleteId = $newAppt.id
    
    # User should not be able to delete
    try {
        Invoke-RestMethod -Uri "$BaseUrl/v1/appointments/$deleteId" -Method DELETE -Headers $headers -ErrorAction Stop
        Write-Fail "User should not be able to delete appointments"
    } catch {
        if ($_.Exception.Response.StatusCode -eq 403) {
            Write-Success "User correctly forbidden from deleting (403)"
        }
    }
    
    # Admin should be able to delete
    try {
        Invoke-WebRequest -Uri "$BaseUrl/v1/appointments/$deleteId" -Method DELETE -Headers $adminHeaders | Out-Null
        Write-Success "Admin deleted appointment successfully"
    } catch {
        Write-Fail "Admin delete failed: $_"
    }
} catch {
    Write-Fail "Could not create appointment for delete test: $_"
}

# ===== Test List Appointments =====
Write-Step "Testing List Appointments..."

try {
    $list = Invoke-RestMethod -Uri "$BaseUrl/v1/appointments" -Method GET -Headers $headers
    Write-Success "Listed $($list.count) appointments"
} catch {
    Write-Fail "List appointments failed: $_"
}

# ===== Summary =====
Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "           Test Complete!" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta
