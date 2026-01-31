# Task W01-04: CI/CD Pipeline Setup

**Status**: Not Started  
**Estimated Time**: 2-3 hours  
**Prerequisites**: TASK_W01_03_ADR_DOCUMENTS.md  
**Next Task**: TASK_W01_05_INITIAL_CODE.md

---

## Objective

Set up automated CI/CD pipeline using GitHub Actions for testing, linting, and building the application on every push and pull request.

---

## Steps

### 1. Create GitHub Actions Directory

```bash
mkdir -p .github/workflows
```

### 2. Create Main CI Workflow

Location: `appointment-service/.github/workflows/ci.yml`

```yaml
name: CI Pipeline

on:
  push:
    branches: [ master, develop ]
  pull_request:
    branches: [ master, develop ]

jobs:
  lint:
    name: Lint Code
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true
      
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

  test:
    name: Run Tests
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: appointments
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: appointments_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true
      
      - name: Install dependencies
        run: go mod download
      
      - name: Install golang-migrate
        run: |
          curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
          sudo mv migrate /usr/local/bin/
          which migrate
      
      - name: Run database migrations
        run: migrate -path migrations -database "postgresql://appointments:test_password@localhost:5432/appointments_test?sslmode=disable" up
        continue-on-error: true  # Continue if no migrations exist yet
      
      - name: Run tests
        env:
          GO_ENV: test
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: appointments
          DB_PASSWORD: test_password
          DB_NAME: appointments_test
          DB_SSL_MODE: disable
          REDIS_HOST: localhost
          REDIS_PORT: 6379
        run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
          flags: unittests
          name: codecov-umbrella
          fail_ci_if_error: false

  build:
    name: Build Application
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true
      
      - name: Build binary
        run: |
          mkdir -p bin
          go build -o bin/appointment-service ./cmd/api
      
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: appointment-service-${{ github.sha }}
          path: bin/appointment-service
          retention-days: 7

  docker:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: [lint, test]
    if: github.ref == 'refs/heads/master' || github.ref == 'refs/heads/develop'
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Log in to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
        continue-on-error: true  # Don't fail if secrets not set
      
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: laithambianze/appointment-service
          tags: |
            type=ref,event=branch
            type=sha,prefix={{branch}}-
            type=raw,value=latest,enable={{is_default_branch}}
      
      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### 3. Create Dependency Update Workflow

Location: `appointment-service/.github/workflows/dependencies.yml`

```yaml
name: Update Dependencies

on:
  schedule:
    # Run every Monday at 9 AM UTC
    - cron: '0 9 * * 1'
  workflow_dispatch:  # Allow manual trigger

jobs:
  update:
    name: Update Go Dependencies
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Update dependencies
        run: |
          go get -u ./...
          go mod tidy
      
      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v6
        with:
          commit-message: 'chore: update go dependencies'
          title: 'chore: update go dependencies'
          body: |
            Automated dependency update
            
            - Updated all Go dependencies to latest versions
            - Ran `go mod tidy`
            
            Please review changes and ensure tests pass.
          branch: chore/update-dependencies
          delete-branch: true
```

### 4. Create Security Scan Workflow

Location: `appointment-service/.github/workflows/security.yml`

```yaml
name: Security Scan

on:
  push:
    branches: [ master ]
  pull_request:
    branches: [ master ]
  schedule:
    # Run every day at 2 AM UTC
    - cron: '0 2 * * *'

jobs:
  gosec:
    name: Security Scan (gosec)
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out results.sarif ./...'
      
      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif

  trivy:
    name: Container Scan (Trivy)
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Build Docker image
        run: docker build -t appointment-service:${{ github.sha }} .
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: 'appointment-service:${{ github.sha }}'
          format: 'sarif'
          output: 'trivy-results.sarif'
      
      - name: Upload Trivy results to GitHub Security tab
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: 'trivy-results.sarif'
```

### 5. Create Release Workflow

Location: `appointment-service/.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Build binaries
        run: |
          mkdir -p dist
          
          # Linux amd64
          GOOS=linux GOARCH=amd64 go build -o dist/appointment-service-linux-amd64 ./cmd/api
          
          # Linux arm64
          GOOS=linux GOARCH=arm64 go build -o dist/appointment-service-linux-arm64 ./cmd/api
          
          # macOS amd64
          GOOS=darwin GOARCH=amd64 go build -o dist/appointment-service-darwin-amd64 ./cmd/api
          
          # macOS arm64
          GOOS=darwin GOARCH=arm64 go build -o dist/appointment-service-darwin-arm64 ./cmd/api
          
          # Windows amd64
          GOOS=windows GOARCH=amd64 go build -o dist/appointment-service-windows-amd64.exe ./cmd/api
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
          generate_release_notes: true
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 6. Create Code Coverage Badge Workflow

Location: `appointment-service/.github/workflows/coverage.yml`

```yaml
name: Coverage

on:
  push:
    branches: [ master ]

jobs:
  coverage:
    name: Update Coverage Badge
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: appointments
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: appointments_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Run tests with coverage
        run: go test -v -coverprofile=coverage.out -covermode=atomic ./...
      
      - name: Calculate coverage
        id: coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "coverage=$COVERAGE" >> $GITHUB_OUTPUT
      
      - name: Update README with coverage
        run: |
          echo "Coverage: ${{ steps.coverage.outputs.coverage }}%"
          # Could update README.md badge here if needed
```

### 7. Add GitHub Actions Badge to README

Update `appointment-service/README.md`:

```markdown
# Appointment Service

[![CI Pipeline](https://github.com/laith-ambianze/appointment-service/actions/workflows/ci.yml/badge.svg)](https://github.com/laith-ambianze/appointment-service/actions/workflows/ci.yml)
[![Security Scan](https://github.com/laith-ambianze/appointment-service/actions/workflows/security.yml/badge.svg)](https://github.com/laith-ambianze/appointment-service/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/laith-ambianze/appointment-service)](https://goreportcard.com/report/github.com/laith-ambianze/appointment-service)
[![codecov](https://codecov.io/gh/laith-ambianze/appointment-service/branch/master/graph/badge.svg)](https://codecov.io/gh/laith-ambianze/appointment-service)

Multi-tenant appointment management microservice built with Go.

... (rest of README)
```

### 8. Configure GitHub Repository Settings

**Go to GitHub repository settings and configure:**

#### Branch Protection (Settings → Branches)
1. Add rule for `master` branch
2. Enable:
   - ✅ Require pull request reviews before merging
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
3. Required status checks:
   - `lint`
   - `test`
   - `build`

#### Secrets (Settings → Secrets and variables → Actions)

Add these secrets if you want Docker Hub integration:
- `DOCKER_USERNAME`: Your Docker Hub username
- `DOCKER_PASSWORD`: Your Docker Hub token

### 9. Commit GitHub Actions

```bash
# Add all workflow files
git add .github/workflows/

# Update README
git add README.md

# Commit
git commit -m "ci: add GitHub Actions workflows

- Add CI pipeline (lint, test, build, docker)
- Add dependency update workflow
- Add security scanning (gosec, trivy)
- Add release workflow for tags
- Add coverage reporting
- Update README with badges"

# Push
git push origin master
```

---

## Verification Checklist

- [ ] `.github/workflows/ci.yml` created
- [ ] `.github/workflows/dependencies.yml` created
- [ ] `.github/workflows/security.yml` created
- [ ] `.github/workflows/release.yml` created
- [ ] `.github/workflows/coverage.yml` created
- [ ] README.md updated with badges
- [ ] Branch protection rules configured on GitHub
- [ ] GitHub Actions workflows pushed
- [ ] First workflow run triggered automatically

---

## Testing the CI Pipeline

After pushing, verify:

1. **Go to GitHub Actions tab**
   - Should see "CI Pipeline" workflow running
   - Check that all jobs pass (lint, test, build)

2. **View workflow details**
   - Click on the running workflow
   - Verify each step completes successfully
   - Check test output

3. **Test Pull Request workflow**
   ```bash
   # Create a test branch
   git checkout -b test/ci-pipeline
   
   # Make a small change
   echo "# Test" >> test.txt
   git add test.txt
   git commit -m "test: verify CI pipeline"
   git push origin test/ci-pipeline
   
   # Create PR on GitHub
   # Verify CI runs on the PR
   ```

---

## Expected CI Pipeline Flow

```mermaid
graph LR
    A[Push/PR] --> B[Lint]
    A --> C[Test]
    B --> D[Build]
    C --> D
    D --> E{Is Master?}
    E -->|Yes| F[Docker Build]
    E -->|No| G[Complete]
```

---

## Next Steps

Proceed to **TASK_W01_05_INITIAL_CODE.md** to create the initial application code skeleton.

---

## Notes for AI Agent

- Workflows will initially show "skipped" for test/build until there's actual code
- Don't worry if Docker push fails without credentials - it's optional
- Security scans might report false positives initially - that's normal
- The CI pipeline is ready to validate all future code changes

---

**Status**: ⏸️ Ready to Start
