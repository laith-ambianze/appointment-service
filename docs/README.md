# Appointment Service

[![CI Pipeline](https://github.com/laith-ambianze/appointment-service/actions/workflows/ci.yml/badge.svg)](https://github.com/laith-ambianze/appointment-service/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/laith-ambianze/appointment-service)](https://goreportcard.com/report/github.com/laith-ambianze/appointment-service)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)

Multi-tenant appointment management microservice built with Go.

## Features

- 🏢 **Multi-tenant Architecture** - Complete product isolation with API key authentication
- 👥 **Flexible Participant Model** - Support for 1-on-1 and group appointments
- 🔐 **Secure Authentication** - API key/secret authentication for products
- 🌍 **Timezone Support** - Full timezone handling with UTC storage
- 📅 **Availability Management** - Configurable availability rules per user
- 🔔 **Webhook Notifications** - Event-driven notifications to products
- 📊 **Prometheus Metrics** - Built-in monitoring and observability
- ⚡ **High Performance** - Raw SQL with pgx for optimal database access

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Framework | Gin |
| Database | PostgreSQL 15+ |
| DB Driver | pgx v5 (raw SQL) |
| Cache | Redis 7+ |
| Migrations | golang-migrate |
| Logging | Zap |
| Testing | Go testing + testify |
| Deployment | Docker + Kubernetes |

## Project Structure

```
appointment-service/
├── cmd/api/                 # Application entry point
├── internal/                # Private application code
│   ├── config/             # Configuration management
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware (auth, CORS, logging)
│   ├── models/             # Domain models
│   ├── repository/         # Data access layer (pgx)
│   ├── routes/             # Route definitions
│   └── service/            # Business logic
├── pkg/                    # Public reusable packages
│   ├── auth/              # Authentication utilities
│   ├── database/          # Database connection utilities
│   ├── logger/            # Logging utilities
│   └── validator/         # Validation utilities
├── migrations/            # Database migrations (SQL files)
├── tests/                 # Test files
│   ├── unit/             # Unit tests
│   └── integration/      # Integration tests
├── docs/                  # Documentation
│   ├── adr/              # Architecture Decision Records
│   └── swagger/          # API documentation
├── deployments/          # Deployment configurations
│   ├── docker/           # Docker files
│   ├── kubernetes/       # K8s manifests
│   └── prometheus/       # Monitoring configs
└── scripts/              # Utility scripts
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)

### Installation

```bash
# Clone the repository
git clone https://github.com/laith-ambianze/appointment-service.git
cd appointment-service

# Install dependencies
make install

# Install development tools
make install-tools
```

### Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit configuration as needed
```

### Running Locally

```bash
# Start database and redis with Docker
make docker-up

# Run database migrations
make migrate-up

# Start the application
make run

# Or with hot reload
make dev
```

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

## API Endpoints

### Health Checks

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Service health status |
| GET | `/ready` | Readiness check (K8s) |
| GET | `/live` | Liveness check (K8s) |

### Products (Admin)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/products/register` | Register a new product |

### Appointments

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/appointments` | Create appointment |
| GET | `/v1/appointments/:id` | Get appointment by ID |
| GET | `/v1/appointments` | List appointments |
| PUT | `/v1/appointments/:id` | Update appointment |
| DELETE | `/v1/appointments/:id` | Cancel appointment |

### Availability

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/availability/:user_id` | Get user availability |
| PUT | `/v1/availability/:user_id` | Update availability settings |
| GET | `/v1/availability/:user_id/slots` | Get available time slots |

## Development

### Available Commands

```bash
make help              # Show all available commands
make install           # Install dependencies
make build             # Build binary
make run               # Run application
make dev               # Run with hot reload
make test              # Run tests
make test-coverage     # Run tests with coverage
make lint              # Run linter
make fmt               # Format code
make migrate-up        # Apply migrations
make migrate-down      # Rollback migrations
make migrate-create    # Create new migration
make docker-build      # Build Docker image
make docker-up         # Start Docker services
make docker-down       # Stop Docker services
make swagger           # Generate API docs
```

### Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `make fmt` before committing
- Ensure `make lint` passes
- Write tests for new features

### Testing

```bash
# Run all tests
make test

# Run with coverage report
make test-coverage

# Run specific test
go test -v ./internal/service/... -run TestCreateAppointment
```

## Architecture

This service is designed as a **general-purpose appointment microservice** that can serve multiple products. Key architectural decisions:

- **Multi-tenancy**: Product-level isolation with `product_id` on all queries
- **External Users**: No internal user management - uses external user IDs
- **Flexible Participants**: Support for any number of participants per appointment
- **Metadata-first**: JSONB fields for product-specific data
- **Event-driven**: Webhook notifications for all appointment events

See [docs/adr/](docs/adr/) for detailed Architecture Decision Records.

## Documentation

- [Appointment as a Service Design](APPOINTMENT_AS_A_SERVICE.md)
- [Architecture Comparison](ARCHITECTURE_COMPARISON.md)
- [Migration Strategy](MIGRATION_STRATEGY_AND_ARCHITECTURE.md)
- [Final Project Plan](FINAL_PROJECT_PLAN.md)
- [Design Decision: Participants](DESIGN_DECISION_PARTICIPANTS.md)
- [API Specification](docs/swagger/)
- [Architecture Decisions](docs/adr/)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

---

**Built with ❤️ for scalable appointment management**
