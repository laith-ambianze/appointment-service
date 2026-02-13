# Week 01: Immediate Actions - Task Index

## Overview

Week 01 focuses on project foundation and infrastructure setup. All tasks are technical and designed for solo work with AI agent assistance.

**Total Estimated Time**: 12-16 hours  
**Goal**: Complete project setup with working HTTP server and CI/CD pipeline

---

## Task List

| Task | Name | Est. Time | Status |
| ------ | ------ | ----------- | -------- |
| [W01-01](TASK_W01_01_PROJECT_SETUP.md) | Project Setup & Repository Structure | 2-3 hours | ⏸️ Not Started |
| [W01-02](TASK_W01_02_CONFIGURATION_FILES.md) | Configuration Files | 2-3 hours | ⏸️ Not Started |
| [W01-03](TASK_W01_03_ADR_DOCUMENTS.md) | Architecture Decision Records | 1-2 hours | ⏸️ Not Started |
| [W01-04](TASK_W01_04_CICD_PIPELINE.md) | CI/CD Pipeline Setup | 2-3 hours | ⏸️ Not Started |
| [W01-05](TASK_W01_05_INITIAL_CODE.md) | Initial Code Skeleton | 2-3 hours | ⏸️ Not Started |

---

## Execution Order

These tasks must be completed in sequence:

1. **TASK_W01_01** - Creates Git repo and project structure
2. **TASK_W01_02** - Adds Makefile, Dockerfile, docker-compose.yml
3. **TASK_W01_03** - Documents architecture decisions
4. **TASK_W01_04** - Sets up automated testing and deployment
5. **TASK_W01_05** - Implements working HTTP server

---

## What You'll Have After Week 01

### ✅ Infrastructure

- Git repository with branch protection
- Complete Go project structure
- Development environment (Docker Compose)
- CI/CD pipeline (GitHub Actions)
- Makefile for common tasks

### ✅ Documentation

- Architecture Decision Records (5 ADRs)
- README with project overview
- Contributing guidelines
- Configuration documentation

### ✅ Working Code

- Configuration management
- Structured logging (Zap)
- HTTP server (Gin)
- Health check endpoints
- Graceful shutdown
- Request logging middleware

### ✅ Automation

- Automated linting
- Automated testing
- Automated builds
- Security scanning
- Dependency updates

---

## Success Criteria

Week 01 is complete when:

- [ ] Application runs successfully (`make run`)
- [ ] Health endpoint returns 200 OK (`curl http://localhost:8081/health`)
- [ ] Docker Compose starts all services (`make docker-up`)
- [ ] CI pipeline passes on GitHub Actions
- [ ] All code is committed and pushed
- [ ] Documentation is complete

---

## Quick Start

```bash
# Clone repository
git clone https://github.com/laith-ambianze/appointment-service.git
cd appointment-service

# Start with first task
open tasks/week-01/TASK_W01_01_PROJECT_SETUP.md

# Or jump directly to implementation
make install        # Install dependencies
make run           # Start development server
make test          # Run tests
make docker-up     # Start with Docker
```

---

## Commands Reference

### Development

```bash
make help          # Show all available commands
make install       # Install Go dependencies
make run           # Run application
make dev           # Run with hot reload (requires air)
make build         # Build binary
```

### Testing

```bash
make test          # Run all tests
make test-coverage # Generate coverage report
make lint          # Run linter
make fmt           # Format code
```

### Database

```bash
make migrate-up         # Apply migrations
make migrate-down       # Rollback migrations
make migrate-create name=xxx  # Create new migration
```

### Docker

```bash
make docker-build  # Build Docker image
make docker-up     # Start all services
make docker-down   # Stop all services
make docker-logs   # View logs
```

### Tools

```bash
make install-tools # Install development tools
make swagger       # Generate API documentation
make clean         # Clean build artifacts
```

---

## Technology Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL 15+ (pgx driver)
- **Cache**: Redis 7+
- **Logger**: Zap
- **Config**: godotenv
- **Testing**: testify
- **CI/CD**: GitHub Actions
- **Containerization**: Docker + Docker Compose

---

## File Structure After Week 01

```md
appointment-service/
├── .github/workflows/      # CI/CD pipelines
├── cmd/api/               # Application entry point
│   └── main.go           # ✅ Working HTTP server
├── internal/
│   ├── config/           # ✅ Configuration management
│   ├── handlers/         # ✅ Health check handler
│   ├── middleware/       # (empty, for Week 02)
│   ├── models/           # (empty, for Week 02)
│   ├── repository/       # (empty, for Week 02)
│   └── service/          # (empty, for Week 02)
├── pkg/
│   ├── logger/           # ✅ Zap logger wrapper
│   ├── auth/             # (empty, for Week 02)
│   ├── validator/        # (empty, for Week 02)
│   └── database/         # (empty, for Week 02)
├── docs/
│   └── adr/              # ✅ 5 Architecture Decision Records
├── deployments/
│   ├── docker/           # ✅ Dockerfile
│   └── kubernetes/       # (empty, for later)
├── .env.example          # ✅ Configuration template
├── .gitignore            # ✅ Git ignore rules
├── Makefile              # ✅ Development commands
├── docker-compose.yml    # ✅ Local development setup
├── README.md             # ✅ Project documentation
└── go.mod                # ✅ Go dependencies
```

---

## Notes for Solo Development

- **No Meetings**: All tasks are hands-on technical work
- **AI Agent**: Use AI for code generation, debugging, and documentation
- **Incremental**: Each task builds on the previous one
- **Testable**: Verify each task before moving to the next
- **Documented**: ADRs explain all major decisions

---

## Troubleshooting

### Application won't start

```bash
# Check configuration
cat .env

# Check if port is available
lsof -i :8081

# Check logs
tail -f logs/app.log
```

### Docker Compose fails

```bash
# Check Docker is running
docker ps

# Check compose configuration
docker-compose config

# Rebuild containers
docker-compose up --build
```

### CI Pipeline fails

```bash
# Run tests locally
make test

# Run linter
make lint

# Check GitHub Actions tab for detailed errors
```

---

## Next Steps

After completing Week 01, proceed to:

**Week 02**: Database Setup & Core Implementation

- Database migrations
- Repository layer (pgx)
- Service layer (business logic)
- API endpoints (CRUD operations)

---

**Document Status**: Ready for Implementation  
**Last Updated**: 2026-01-31  
**Maintained By**: Solo Developer + AI Agent
