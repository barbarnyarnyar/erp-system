# System Architecture

Technical overview of the ERP system architecture — covering current implementation state, design decisions, and gaps between documented aspirations and deployed reality.

## Architecture Landscape

| Document | Coverage | Honest About Current State? |
|----------|----------|----------------------------|
| [System Overview](system-overview.md) | C4 models (Context → Container → Component → Code), ADRs, bounded contexts, scalability, resilience, observability, deployment, API strategy, testing | ✅ Yes — full current vs target comparison |
| [Services Overview](services-overview.md) | Per-service: all endpoints, domain models, subdomain services, events, known issues | ✅ Yes — port mismatches, dead code, stubs documented |
| [Event Architecture](event-architecture.md) | Full event catalog per service, cross-service flows, consumer actions, known issues | ✅ Yes — fire-and-forget, unused topics, logged-only consumers documented |
| [API Design](api-design.md) | Gateway routing, active vs inactive auth, route conventions, response formats, inconsistencies | ✅ Yes — active/inactive split, port mismatches documented |
| [Microservices Architecture](microservices-architecture.md) | Actual service structure, inter-service communication, API gateway, resilience, testing | ✅ Yes — aspirational patterns removed, real patterns documented |
| [Technology Stack](technology-stack.md) | Go dependencies, frameworks, infrastructure, deployment tools, shared utils | ✅ Yes — aspirational tools removed, unused deps noted |
| [Database Design](database-design.md) | In-memory storage architecture, migration files, consistency model | ✅ Yes — no database connected, migration path outlined |

## Cross-Service Architecture Docs

- [Deployment Architecture](deployment-architecture.md) — Docker Compose configuration and build details
- [Security Architecture](security-architecture.md) — Auth service, JWT, RBAC, current gaps
- [Performance Architecture](performance-architecture.md) — Concurrency, caching, bottlenecks

## Key Architecture Decisions

See [System Overview → Architecture Decision Records](system-overview.md#architecture-decision-records) for 12 ADRs covering Go/Gin, in-memory storage, Kafka, auth, CDD, ID generation, pagination, logging, testing, JWT secrets, event reliability, and consumer threading.

## Current State Summary

- **7 backend services** (Auth + 6 domain) running in Docker Compose
- **11 containers total** (PostgreSQL + Redis + ZK + Kafka + 7 services; API Gateway runs separately)
- **All services use** Gin, in-memory maps, Kafka-go, nanosecond-timestamp IDs
- **No database connected** despite 41 SQL migration files
- **No authentication** on deployed gateway (full JWT/RBAC exists but inactive)
- **~85 Kafka topics** defined, ~20+ never published
- **~325 API endpoints** across all services
- **1 test file** in the entire codebase
- **0 CI/CD pipelines, 0 Kubernetes configs, 0 metrics, 0 traces**

## Next Steps

After understanding the architecture:
- [Business Modules](../modules/README.md) — Per-module domain models, endpoints, events
- [Operations](../operations/README.md) — Deploy and maintain the system