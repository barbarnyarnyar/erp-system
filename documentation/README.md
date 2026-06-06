# ERP System — Documentation

Microservices-based ERP system built with Go, Gin, Kafka, and Docker Compose.

## Quick Start

```bash
# Clone, build, and run
git clone https://github.com/your-org/erp-system.git
cd erp-system
make run         # Start all services (docker compose)
cd api-gateway && go run cmd/main.go   # Start gateway (not in docker-compose)
```

See [Getting Started](getting-started/README.md) for prerequisites and detailed setup.

## Documentation Sections

### [Getting Started](getting-started/README.md)
Setup guides for new developers: prerequisites, installation, configuration, development environment, and workflow.

### [Architecture](architecture/README.md)
System design, C4 model, service patterns, technology stack, database design, API design, event architecture, and CDD contract reference.

### [Modules](modules/README.md)
Business domain documentation for all 7 services: Auth, Financial Management, Human Resources, Supply Chain, Manufacturing, CRM, and Project Management.

### [Operations](operations/README.md)
Deployment, monitoring, security, troubleshooting, and maintenance guides.

### [PRDs](PRDs/active/)
Active product requirements documents and phase breakdowns for ongoing documentation work.

## Project Status

This system is a development prototype with honest documentation:

- **All services use in-memory storage** — no database is connected at runtime
- **No authentication on API Gateway** — all endpoints are publicly accessible
- **Fire-and-forget Kafka** — event publishing errors are silently ignored
- **Single test file** — only FM service has tests (2 test cases)
- **Port inconsistencies** — CRM/HR/SCM have code-default ports differing from documented values
- **41 known issues** — documented in [Common Issues](getting-started/common-issues.md)

## Documentation Conventions

1. **Current-state first** — docs describe what the code actually does, not what it should do
2. **Gaps noted explicitly** — where target architecture differs from current implementation
3. **CDD-aligned** — event topics, entities, and service boundaries match `.cdd` contract files
4. **No fictional features** — if it's not in the codebase, it's not in the docs

## File Inventory

| Directory | Files | Purpose |
|-----------|-------|---------|
| `getting-started/` | 8 | Prerequisites, installation, config, dev setup, workflow, testing, issues |
| `architecture/` | 12 | C4 model, ADRs, tech stack, services, events, API, database, deployment, security, performance, CDD ref |
| `modules/` | 10 | Business module descriptions per service (7 modules + FM sub-docs) |
| `operations/` | 13 | Operations, deployment, monitoring, security guides |
| `PRDs/active/` | 3 | Active PRD + phase breakdowns |
