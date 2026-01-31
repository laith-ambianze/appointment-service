# Task W01-01: Project Setup & Repository Structure

**Status**: ✅ COMPLETED  
**Estimated Time**: 2-3 hours  
**Prerequisites**: None  
**Next Task**: TASK_W01_02_CONFIGURATION_FILES.md  
**Completed**: 2026-01-31

---

## Objective

Initialize the Git repository, create the Go project structure, and set up the foundational directory layout for the appointment microservice.

---

## Technical Requirements

- Go 1.21 or higher installed
- Git installed
- GitHub account with repository access

---

## Steps

### 1. Create GitHub Repository

```bash
# If repository doesn't exist, create it on GitHub:
# - Repository name: appointment-service
# - Visibility: Private (or Public)
# - Initialize with: None (we'll push our own structure)

# Then clone locally
git clone https://github.com/laith-ambianze/appointment-service.git
cd appointment-service
```

**Configure Branch Protection** (GitHub Settings → Branches):

- Protect `master` branch
- Require pull request reviews before merging
- Require status checks to pass before merging

### 2. Initialize Go Module

```bash
# Initialize Go module
go mod init github.com/laith-ambianze/appointment-service

# Verify
go version  # Should show 1.21+
```

### 3. Create Directory Structure

```bash
# Core application directories
mkdir -p cmd/api
mkdir -p internal/handlers
mkdir -p internal/middleware
mkdir -p internal/models
mkdir -p internal/repository
mkdir -p internal/service
mkdir -p internal/config
mkdir -p internal/routes

# Public packages
mkdir -p pkg/auth
mkdir -p pkg/logger
mkdir -p pkg/validator
mkdir -p pkg/database

# Database migrations
mkdir -p migrations

# Tests
mkdir -p tests/unit
mkdir -p tests/integration

# Documentation
mkdir -p docs/adr
mkdir -p docs/swagger

# Scripts
mkdir -p scripts

# Deployment configs
mkdir -p deployments/docker
mkdir -p deployments/kubernetes
mkdir -p deployments/prometheus

# Create initial placeholder files
touch cmd/api/main.go
touch internal/config/config.go
touch pkg/logger/logger.go
touch README.md
touch .gitignore
```

### 4. Create Initial README.md

```markdown
# Appointment Service

Multi-tenant appointment management microservice built with Go.

## Features

- 🏢 Multi-tenant architecture with product isolation
- 👥 Flexible participant model (1-on-1 and group appointments)
- 🔐 API key authentication
- 🌍 Timezone support
- 📅 Availability management
- 🔔 Webhook notifications
- 📊 Prometheus metrics

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL 15+ with pgx driver
- **Cache**: Redis 7+
- **Migrations**: golang-migrate
- **Testing**: Go testing + testify
- **Deployment**: Docker + Kubernetes

## Project Structure

```md
appointment-service/
├── cmd/api/                 # Application entry point
├── internal/                # Private application code
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Domain models
│   ├── repository/         # Data access layer
│   ├── service/            # Business logic
│   └── config/             # Configuration
├── pkg/                    # Public reusable packages
│   ├── auth/              # Authentication utilities
│   ├── logger/            # Logging utilities
│   └── database/          # Database utilities
├── migrations/            # Database migrations
├── tests/                 # Tests
├── docs/                  # Documentation
└── deployments/           # Deployment configs
```

## Quick Start

See [docs/DEVELOPMENT_SETUP.md](docs/DEVELOPMENT_SETUP.md) for setup instructions.

## Documentation

- [Architecture](../MIGRATION_STRATEGY_AND_ARCHITECTURE.md)
- [API Specification](docs/API_SPECIFICATION.yaml)
- [Database Schema](docs/DATABASE_SCHEMA.md)
- [ADR Documents](docs/adr/)

## Development

```bash
# Install dependencies
make install

# Run tests
make test

# Start development server
make dev

# Build binary
make build
```

## License

MIT

### 5. Create Initial .gitignore

```gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary
*.test

# Output of the go coverage tool
*.out
coverage.html

# Dependency directories
vendor/

# Go workspace file
go.work

# Environment variables
.env
.env.local
.env.*.local
.env.development
.env.production

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Database
*.db
*.sqlite

# Temporary files
tmp/
temp/
```

### 6. Create CONTRIBUTING.md

````markdown
# Contributing Guidelines

## Development Workflow

1. Create a feature branch from `master`
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following our coding standards

3. Run tests and linter

   ```bash
   make test
   make lint
   make fmt
   ```

4. Commit your changes with clear messages

   ```bash
   git commit -m "feat: add appointment creation endpoint"
   ```

5. Push and create a pull request

   ```bash
   git push origin feature/your-feature-name
   ```

## Commit Message Format

Follow conventional commits:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

Examples:

```
feat: add participant role validation
fix: resolve timezone conversion bug
docs: update API documentation
refactor: simplify availability calculation
test: add unit tests for appointment service
chore: update dependencies
```

## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `make fmt` before committing
- Ensure `make lint` passes
- Write tests for new features
- Keep functions small and focused
- Document exported functions

## Testing

- Unit tests: `internal/*/service/*_test.go`
- Integration tests: `tests/integration/*_test.go`
- Coverage target: > 80%

Run tests:

```bash
make test
make test-coverage  # Generate HTML coverage report
```
````

### 7. Initial Commit

```bash
# Add all files
git add .

# Commit
git commit -m "chore: initialize project structure

- Add Go module
- Create directory structure
- Add README and contributing guidelines
- Configure .gitignore"

# Push to GitHub
git push -u origin master
```

---

## Verification Checklist

- [ ] GitHub repository created and accessible
- [ ] Go module initialized (`go.mod` exists)
- [ ] Complete directory structure created (cmd, internal, pkg, migrations, tests, docs, deployments)
- [ ] README.md created with project overview
- [ ] CONTRIBUTING.md created with workflow guidelines
- [ ] .gitignore configured properly
- [ ] Initial commit pushed to master branch
- [ ] Repository structure matches planned architecture

---

## Expected Output

After completion, your repository should look like:

```md
appointment-service/
├── cmd/
│   └── api/
│       └── main.go (empty for now)
├── internal/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── repository/
│   ├── service/
│   ├── config/
│   │   └── config.go (empty for now)
│   └── routes/
├── pkg/
│   ├── auth/
│   ├── logger/
│   │   └── logger.go (empty for now)
│   ├── validator/
│   └── database/
├── migrations/
├── tests/
│   ├── unit/
│   └── integration/
├── docs/
│   ├── adr/
│   └── swagger/
├── scripts/
├── deployments/
│   ├── docker/
│   ├── kubernetes/
│   └── prometheus/
├── .gitignore
├── README.md
├── CONTRIBUTING.md
└── go.mod
```

---

## Commands Summary

```bash
# Setup
git clone https://github.com/laith-ambianze/appointment-service.git
cd appointment-service
go mod init github.com/laith-ambianze/appointment-service

# Create structure (run all mkdir commands above)

# Initial commit
git add .
git commit -m "chore: initialize project structure"
git push -u origin master
```

---

## Next Steps

Once this task is complete:

1. Proceed to **TASK_W01_02_CONFIGURATION_FILES.md**
2. Create Makefile, Dockerfile, docker-compose.yml, and .env.example

---

## Notes for AI Agent

- This is a foundational task - ensure all directories are created correctly
- The structure follows Go best practices (internal for private code, pkg for public)
- Empty files are intentional - they'll be populated in subsequent tasks
- Focus on getting the structure right; code comes later

---

**Status**: ⏸️ Ready to Start
