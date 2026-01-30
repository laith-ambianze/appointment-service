# Appointment-as-a-Service - Final Project Plan

## Project Overview

A microservice-based appointment management system designed to be integrated into multiple products, allowing users from different platforms to schedule and manage appointments with each other.

---

## 🎯 Core Objectives

- **Multi-tenant Service**: Support multiple products/applications
- **Production-Ready**: Dockerized deployment
- **Secure**: Token-based authentication for product integration
- **Scalable**: Own database, isolated from client applications
- **Easy Integration**: RESTful API with comprehensive documentation

---

## 🏗️ System Architecture

### Architecture Type

**Microservice Architecture** - Standalone service with REST API

### Components

```md
┌─────────────────────────────────────────────────┐
│              Product 1 (Client)                  │
│         (Frontend + Backend)                     │
└──────────────────┬──────────────────────────────┘
                   │ API Calls (Token Auth)
                   ▼
┌─────────────────────────────────────────────────┐
│      Appointment Service (Docker Container)      │
│                                                   │
│  ┌──────────────┐      ┌──────────────┐        │
│  │ API Gateway   │──────│ Auth Layer   │        │
│  │   (Go/Gin)    │      │ (JWT Token)  │        │
│  └──────────────┘      └──────────────┘        │
│          │                                       │
│  ┌──────────────────────────────┐              │
│  │    Business Logic Layer       │              │
│  │  - Appointment Management     │              │
│  │  - Product Management         │              │
│  │  - User Metadata Management   │              │
│  └──────────────┬───────────────┘              │
│                 │                                 │
│  ┌──────────────▼───────────────┐              │
│  │    PostgreSQL Database        │              │
│  │  - Products                   │              │
│  │  - Appointments               │              │
│  │  - User Metadata              │              │
│  └──────────────────────────────┘              │
└─────────────────────────────────────────────────┘
```

---

## 💻 Technology Stack

### Backend

- **Language**: Go (Golang) 1.21+
- **Framework**: Gin / Fiber
- **Router**: Chi (alternative)
- **Validation**: go-playground/validator
- **Documentation**: Swagger/OpenAPI (swag)

### Database

- **Primary DB**: PostgreSQL 15+
- **Database Driver**: pgx (PostgreSQL driver)
- **ORM/Query Builder**: GORM / sqlc
- **Migrations**: golang-migrate / goose

### Security

- **Authentication**: JWT (golang-jwt/jwt)
- **Encryption**: bcrypt from crypto/bcrypt
- **Rate Limiting**: golang.org/x/time/rate
- **CORS**: gin-contrib/cors or rs/cors

### DevOps

- **Containerization**: Docker + Docker Compose
- **Environment**: godotenv / viper
- **Logging**: zap / logrus
- **Testing**: Go testing package + testify

---

## 🗄️ Database Schema

### Tables

#### 1. **products**

Stores registered products/applications that use the service.

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    api_secret_hash VARCHAR(255) NOT NULL,
    callback_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. **appointments**

Core table for storing appointment data.

```sql
CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    
    -- User 1 (Initiator)
    user1_id VARCHAR(255) NOT NULL,
    user1_metadata JSONB NOT NULL,
    
    -- User 2 (Recipient)
    user2_id VARCHAR(255) NOT NULL,
    user2_metadata JSONB NOT NULL,
    
    -- Appointment Details
    title VARCHAR(500) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    location VARCHAR(500),
    meeting_type VARCHAR(50), -- 'online', 'in-person', 'phone'
    
    -- Status Management
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'confirmed', 'cancelled', 'completed'
    cancelled_by VARCHAR(255),
    cancellation_reason TEXT,
    cancelled_at TIMESTAMP,
    
    -- Additional Data
    additional_metadata JSONB,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_time_range CHECK (end_time > start_time)
);

CREATE INDEX idx_appointments_product ON appointments(product_id);
CREATE INDEX idx_appointments_user1 ON appointments(user1_id);
CREATE INDEX idx_appointments_user2 ON appointments(user2_id);
CREATE INDEX idx_appointments_status ON appointments(status);
CREATE INDEX idx_appointments_time ON appointments(start_time, end_time);
```

#### 3. **user_metadata_schema**

User metadata structure (stored as JSONB):

```json
{
    "userId": "string (required)",
    "firstName": "string (required)",
    "lastName": "string (required)",
    "email": "string (optional)",
    "phone": "string (optional)",
    "timezone": "string (optional)",
    "customFields": {
        "key": "value"
    }
}
```

#### 4. **appointment_history** (Audit Trail)

Track all changes to appointments.

```sql
CREATE TABLE appointment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id UUID NOT NULL REFERENCES appointments(id),
    action VARCHAR(50) NOT NULL, -- 'created', 'updated', 'cancelled', 'confirmed'
    changed_by VARCHAR(255),
    changes JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🔌 API Endpoints (MVP)

### Base URL

```md
https://api.appointment-service.com/v1
```

### Authentication

All endpoints (except product registration) require authentication:

```md
Headers:
  X-API-Key: {product_api_key}
  X-API-Secret: {product_api_secret}
```

---

### 1. Product Management

#### **POST /products/register**

Register a new product and receive API credentials.

**Request Body:**

```json
{
    "name": "Product Name",
    "description": "Product description",
    "callbackUrl": "https://product.com/webhook"
}
```

**Response:**

```json
{
    "success": true,
    "data": {
        "productId": "uuid",
        "apiKey": "prod_xxxxxxxxxxxxx",
        "apiSecret": "secret_xxxxxxxxxxxxx",
        "message": "Store these credentials securely. The secret will not be shown again."
    }
}
```

#### **GET /products/me**

Get current product information.

**Response:**

```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "name": "Product Name",
        "isActive": true,
        "createdAt": "2026-01-30T10:00:00Z"
    }
}
```

---

### 2. Appointment Management

#### **POST /appointments**

Create a new appointment.

**Request Body:**

```json
{
    "user1": {
        "userId": "user_123",
        "firstName": "John",
        "lastName": "Doe",
        "email": "john@example.com",
        "phone": "+1234567890",
        "customFields": {}
    },
    "user2": {
        "userId": "user_456",
        "firstName": "Jane",
        "lastName": "Smith",
        "email": "jane@example.com",
        "phone": "+0987654321",
        "customFields": {}
    },
    "title": "Business Meeting",
    "description": "Discuss Q1 goals",
    "startTime": "2026-02-15T10:00:00Z",
    "endTime": "2026-02-15T11:00:00Z",
    "location": "Conference Room A",
    "meetingType": "in-person",
    "additionalMetadata": {}
}
```

**Response:**

```json
{
    "success": true,
    "data": {
        "appointmentId": "uuid",
        "status": "pending",
        "createdAt": "2026-01-30T10:00:00Z"
    }
}
```

#### **GET /appointments**

Get all appointments for the authenticated product.

**Query Parameters:**

- `page` (default: 1)
- `limit` (default: 20, max: 100)
- `status` (optional: pending, confirmed, cancelled, completed)
- `startDate` (optional: ISO 8601)
- `endDate` (optional: ISO 8601)

**Response:**

```json
{
    "success": true,
    "data": {
        "appointments": [...],
        "pagination": {
            "page": 1,
            "limit": 20,
            "total": 100,
            "totalPages": 5
        }
    }
}
```

#### **GET /appointments/:appointmentId**

Get a specific appointment by ID.

**Response:**

```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "title": "Business Meeting",
        "user1": {...},
        "user2": {...},
        "startTime": "2026-02-15T10:00:00Z",
        "endTime": "2026-02-15T11:00:00Z",
        "status": "pending",
        "createdAt": "2026-01-30T10:00:00Z",
        "updatedAt": "2026-01-30T10:00:00Z"
    }
}
```

#### **GET /appointments/user/:userId**

Get all appointments for a specific user.

**Query Parameters:**

- Same as GET /appointments

**Response:** Same structure as GET /appointments

#### **PATCH /appointments/:appointmentId**

Update appointment details (except status).

**Request Body:**

```json
{
    "title": "Updated Meeting Title",
    "description": "Updated description",
    "startTime": "2026-02-15T11:00:00Z",
    "endTime": "2026-02-15T12:00:00Z",
    "location": "Updated location"
}
```

#### **PATCH /appointments/:appointmentId/status**

Update appointment status.

**Request Body:**

```json
{
    "status": "confirmed"
}
```

**Available Status Transitions:**

- pending → confirmed
- pending → cancelled
- confirmed → cancelled
- confirmed → completed

#### **PATCH /appointments/:appointmentId/cancel**

Cancel an appointment.

**Request Body:**

```json
{
    "cancelledBy": "user_123",
    "reason": "Schedule conflict"
}
```

**Response:**

```json
{
    "success": true,
    "data": {
        "appointmentId": "uuid",
        "status": "cancelled",
        "cancelledBy": "user_123",
        "cancelledAt": "2026-01-30T10:00:00Z"
    }
}
```

#### **DELETE /appointments/:appointmentId**

Permanently delete an appointment (admin only).

---

## 🔐 Security Implementation

### 1. API Authentication Flow

```md
1. Product registers → Receives API Key + Secret
2. Product stores credentials securely
3. For each request:
   - Include X-API-Key and X-API-Secret headers
   - Server validates credentials
   - Server generates short-lived JWT for session
   - Request processed if valid
```

### 2. Security Measures

- API secrets hashed with bcrypt (never stored in plain text)
- Rate limiting: 100 requests/minute per product
- CORS whitelist configuration
- Input validation on all endpoints
- SQL injection prevention (ORM parameterized queries)
- XSS protection
- Request size limits
- HTTPS only in production

### 3. Environment Variables

```env
NODE_ENV=production
PORT=3000
DATABASE_URL=postgresql://user:pass@db:5432/appointments
JWT_SECRET=your-secret-key
API_SECRET_SALT_ROUNDS=10
CORS_ORIGINS=https://product1.com,https://product2.com
RATE_LIMIT_WINDOW_MS=60000
RATE_LIMIT_MAX_REQUESTS=100
```

---

## 🐳 Docker Setup

### Project Structure

```md
appointment-service/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── appointment.go
│   │   └── product.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── ratelimit.go
│   ├── models/
│   │   ├── appointment.go
│   │   └── product.go
│   ├── repository/
│   │   ├── appointment_repo.go
│   │   └── product_repo.go
│   ├── service/
│   │   ├── appointment_service.go
│   │   └── product_service.go
│   ├── config/
│   │   └── config.go
│   └── routes/
│       └── routes.go
├── pkg/
│   ├── auth/
│   ├── logger/
│   └── validator/
├── migrations/
│   ├── 001_create_products.up.sql
│   ├── 001_create_products.down.sql
│   ├── 002_create_appointments.up.sql
│   └── 002_create_appointments.down.sql
├── tests/
│   ├── integration/
│   └── unit/
├── docs/
│   ├── swagger.yaml
│   └── API_INTEGRATION_GUIDE.md
├── .env.example
├── .dockerignore
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Copy .env if needed (or use environment variables)
COPY .env.example .env

EXPOSE 8080

CMD ["./main"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    build: .
    container_name: appointment-service
    ports:
      - "8080:8080"
    environment:
      GO_ENV: production
      DB_HOST: db
      DB_PORT: 5432
      DB_USER: appointments
      DB_PASSWORD: password
      DB_NAME: appointments
      DB_SSL_MODE: disable
      JWT_SECRET: ${JWT_SECRET}
      API_PORT: 8080
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - appointment-network
    command: ["/bin/sh", "-c", "sleep 5 && ./main"]

  db:
    image: postgres:15-alpine
    container_name: appointment-db
    environment:
      POSTGRES_USER: appointments
      POSTGRES_PASSWORD: password
      POSTGRES_DB: appointments
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U appointments"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - appointment-network

volumes:
  postgres_data:

networks:
  appointment-network:
    driver: bridge
```

### Deployment Commands

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down

# Rebuild after changes
docker-compose up -d --build

# Run migrations manually
docker-compose exec app ./main -migrate

# Build locally (without Docker)
make build

# Run locally
make run

# Run tests
make test
```

---

## 📚 Integration Documentation

### For Backend Developers (Client Products)

#### Step 1: Register Your Product

```bash
curl -X POST https://api.appointment-service.com/v1/products/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Product",
    "description": "Product description",
    "callbackUrl": "https://myproduct.com/webhook"
  }'
```

**Save the `apiKey` and `apiSecret` securely!**

#### Step 2: Store Credentials

```javascript
// Backend environment variables
APPOINTMENT_SERVICE_URL=https://api.appointment-service.com/v1
APPOINTMENT_API_KEY=prod_xxxxxxxxxxxxx
APPOINTMENT_API_SECRET=secret_xxxxxxxxxxxxx
```

#### Step 3: Create Appointment Helper (Node.js Example)

```javascript
const axios = require('axios');

class AppointmentService {
    constructor() {
        this.baseURL = process.env.APPOINTMENT_SERVICE_URL;
        this.apiKey = process.env.APPOINTMENT_API_KEY;
        this.apiSecret = process.env.APPOINTMENT_API_SECRET;
    }

    async createAppointment(data) {
        try {
            const response = await axios.post(
                `${this.baseURL}/appointments`,
                data,
                {
                    headers: {
                        'Content-Type': 'application/json',
                        'X-API-Key': this.apiKey,
                        'X-API-Secret': this.apiSecret
                    }
                }
            );
            return response.data;
        } catch (error) {
            console.error('Error creating appointment:', error.response?.data);
            throw error;
        }
    }

    async getUserAppointments(userId) {
        try {
            const response = await axios.get(
                `${this.baseURL}/appointments/user/${userId}`,
                {
                    headers: {
                        'X-API-Key': this.apiKey,
                        'X-API-Secret': this.apiSecret
                    }
                }
            );
            return response.data;
        } catch (error) {
            console.error('Error fetching appointments:', error.response?.data);
            throw error;
        }
    }

    async cancelAppointment(appointmentId, userId, reason) {
        try {
            const response = await axios.patch(
                `${this.baseURL}/appointments/${appointmentId}/cancel`,
                {
                    cancelledBy: userId,
                    reason: reason
                },
                {
                    headers: {
                        'Content-Type': 'application/json',
                        'X-API-Key': this.apiKey,
                        'X-API-Secret': this.apiSecret
                    }
                }
            );
            return response.data;
        } catch (error) {
            console.error('Error cancelling appointment:', error.response?.data);
            throw error;
        }
    }
}

module.exports = new AppointmentService();
```

#### Step 4: Usage in Your Backend

```javascript
// In your route handler
const appointmentService = require('./services/appointmentService');

app.post('/api/schedule-meeting', async (req, res) => {
    try {
        const { user1Id, user2Id, title, startTime, endTime } = req.body;
        
        // Fetch user details from your database
        const user1 = await getUserById(user1Id);
        const user2 = await getUserById(user2Id);
        
        // Create appointment
        const appointment = await appointmentService.createAppointment({
            user1: {
                userId: user1.id,
                firstName: user1.firstName,
                lastName: user1.lastName,
                email: user1.email,
                phone: user1.phone
            },
            user2: {
                userId: user2.id,
                firstName: user2.firstName,
                lastName: user2.lastName,
                email: user2.email,
                phone: user2.phone
            },
            title: title,
            startTime: startTime,
            endTime: endTime,
            meetingType: 'online'
        });
        
        res.json({ success: true, appointment });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});
```

---

## 📋 Development Roadmap

### Phase 1: MVP (Weeks 1-4)

- [ ] Project setup & configuration
- [ ] Database schema & migrations
- [ ] Product registration API
- [ ] Appointment CRUD APIs
- [ ] Authentication middleware
- [ ] Basic validation & error handling
- [ ] Docker setup
- [ ] API documentation
- [ ] Integration guide

### Phase 2: Testing & Refinement (Weeks 5-6)

- [ ] Unit tests (>80% coverage)
- [ ] Integration tests
- [ ] Load testing
- [ ] Security audit
- [ ] Bug fixes
- [ ] Performance optimization

### Phase 3: Post-MVP Features (Weeks 7-12)

- [ ] **User Availability Settings**
  - Set working hours per day
  - Block specific time slots
  - Recurring availability patterns
  - Timezone handling
- [ ] **Conflict Detection**
  - Check for overlapping appointments
  - Suggest alternative times
- [ ] **Notifications** (optional)
  - Webhook callbacks to products
  - Appointment reminders
- [ ] **Recurring Appointments**
  - Daily, weekly, monthly patterns
- [ ] **Advanced Search & Filters**
  - By date range, status, users
  - Full-text search
- [ ] **Analytics Dashboard**
  - Appointment statistics per product
  - Usage metrics

---

## 🧪 Testing Strategy

### Unit Tests

- Service layer logic
- Validation functions
- Utility functions
- Target: >80% coverage

### Integration Tests

- API endpoint testing
- Database operations
- Authentication flow
- Error handling

### Load Tests

- Concurrent requests handling
- Database connection pooling
- Rate limiting effectiveness

---

## 📊 Monitoring & Logging

### Logging Levels

```javascript
// Winston configuration
{
    error: Errors that need immediate attention
    warn: Warning messages
    info: General information (API calls, etc.)
    debug: Detailed debugging information
}
```

### Metrics to Track

- Request count per endpoint
- Response times
- Error rates
- Active products
- Total appointments
- Database query performance

### Recommended Tools

- **Logging**: Winston + Elasticsearch + Kibana
- **Monitoring**: Prometheus + Grafana
- **Error Tracking**: Sentry
- **APM**: New Relic / Datadog

---

## 🚀 Deployment Checklist

### Pre-Production

- [ ] Environment variables configured
- [ ] Database backups enabled
- [ ] SSL certificates installed
- [ ] CORS whitelist configured
- [ ] Rate limits tested
- [ ] API documentation published
- [ ] Load testing completed
- [ ] Security audit passed

### Production

- [ ] Docker containers running
- [ ] Database migrations applied
- [ ] Monitoring tools active
- [ ] Logging configured
- [ ] Backup strategy implemented
- [ ] Rollback plan documented
- [ ] Integration guide published

---

## 📖 Documentation Deliverables

1. **API_INTEGRATION_GUIDE.md**
   - Quick start guide
   - Authentication setup
   - Code examples in multiple languages
   - Error handling
   - Best practices

2. **API_REFERENCE.md**
   - Complete endpoint documentation
   - Request/response schemas
   - Error codes
   - Rate limits

3. **DEPLOYMENT_GUIDE.md**
   - Docker setup
   - Environment configuration
   - Scaling strategies
   - Backup procedures

4. **DEVELOPER_GUIDE.md**
   - Project structure
   - Development workflow
   - Testing guidelines
   - Contribution guidelines

---

## 🔮 Future Enhancements (Post-MVP)

### Availability Management System

```javascript
// User availability settings
{
    userId: "user_123",
    productId: "prod_456",
    timezone: "America/New_York",
    workingHours: {
        monday: [{ start: "09:00", end: "17:00" }],
        tuesday: [{ start: "09:00", end: "17:00" }],
        wednesday: [{ start: "09:00", end: "17:00" }],
        thursday: [{ start: "09:00", end: "17:00" }],
        friday: [{ start: "09:00", end: "17:00" }],
        saturday: [],
        sunday: []
    },
    blockedSlots: [
        {
            start: "2026-02-15T12:00:00Z",
            end: "2026-02-15T13:00:00Z",
            reason: "Lunch break"
        }
    ],
    minimumNotice: 24, // hours
    bufferTime: 15 // minutes between appointments
}
```

### Additional Features

- Calendar integration (Google, Outlook)
- Video conferencing links (Zoom, Meet)
- SMS/Email notifications
- Multi-language support
- Custom branding per product
- Advanced scheduling algorithms
- Resource booking (rooms, equipment)

---

## 💰 Cost Estimation (Monthly)

### Development Phase

- Developer time: 4-6 weeks for MVP
- Testing & QA: 1-2 weeks

### Infrastructure (Production)

- **Hosting**: $20-50/month
  - Docker container (2GB RAM, 1 CPU)
  - PostgreSQL instance
- **Domain & SSL**: $10-15/month
- **Monitoring Tools**: $0-50/month (free tiers available)
- **Backup Storage**: $5-10/month

**Total Estimated**: $35-125/month

---

## ✅ Success Metrics

### Technical Metrics

- API uptime: >99.9%
- Response time: <200ms (p95)
- Error rate: <0.1%
- Test coverage: >80%

### Business Metrics

- Number of integrated products
- Total appointments created
- API usage growth
- Product retention rate

---

## 🤝 Support & Maintenance

### Support Channels

- Documentation website
- Email support: <support@appointment-service.com>
- GitHub issues (for bug reports)
- Integration assistance for new products

### Maintenance Schedule

- Security patches: As needed
- Feature updates: Monthly
- Dependency updates: Quarterly
- Database optimization: Quarterly

---

## 📝 Notes & Considerations

1. **Data Privacy**: Ensure compliance with GDPR/CCPA if applicable
2. **Data Retention**: Define policy for old appointments (e.g., archive after 2 years)
3. **Scalability**: Database partitioning strategy if exceeding 1M appointments
4. **Multi-region**: Consider CDN and multiple database regions for global users
5. **Versioning**: Use API versioning (v1, v2) for backward compatibility
6. **Migration**: Provide data export functionality for products

---

## 📞 Contact & Resources

- **Project Repository**: <https://github.com/laith-ambianze/appointment-service>
- **Documentation**: <https://docs.appointment-service.com>
- **API Status**: <https://status.appointment-service.com>

---

**Document Version**: 1.0  
**Last Updated**: January 30, 2026  
**Author**: Project Planning Team  
**Status**: Final - Ready for Implementation
