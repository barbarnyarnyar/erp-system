# ERP Specialist Skill

## Overview

This skill provides **comprehensive expertise** for working with the ERP microservices system. An agent equipped with this skill becomes an expert in:

- **System Architecture**: Understanding the 10-service microservices design (API Gateway + 9 domain services)
- **Domain Services**: Deep knowledge of each service's responsibilities, databases, and APIs
- **Business Logic**: How services collaborate to execute business workflows
- **Event-Driven Patterns**: Kafka-based asynchronous communication, Outbox/Inbox patterns, and DLQ handling
- **Distributed Observability**: Distributed tracing using OpenTelemetry across HTTP and Kafka boundaries
- **Security**: Credential management, JWT authentication (active/inactive states), and RBAC controls
- **Development**: Building, testing, and deploying services using GORM and Clean Architecture Lite

## When to Use

Invoke this skill when working on:

### Feature Development
```bash
# Add new functionality to a service
@erp-specialist Add a new endpoint to the HR service for employee performance reviews
```

### Service Integration
```bash
# Coordinate changes across multiple services
@erp-specialist Design the event flow when an order is created in CRM that triggers SCM inventory check
```

### Troubleshooting
```bash
# Debug complex service interactions
@erp-specialist Why is the payment processing failing? Trace the event flow from CRM → FM → PM
```

### Observability & Resilience
```bash
# Debug tracing or event propagation
@erp-specialist Check if tracecontext headers are being propagated to Kafka message headers.
```

---

## Skill Contents

### SKILL.md
The main skill definition containing:

1. **Architecture & Infrastructure** - System design and components
2. **Domain Services** - 9 business services + Auth service
3. **Technical Stack** - Go, Gin, GORM, PostgreSQL, Kafka, OpenTelemetry
4. **Day 2 Operations Infrastructure** - Distributed Tracing and Dead Letter Queue details
5. **Contract-Driven Development** - API contracts and OpenAPI spec generation
6. **Development Workflows & Troubleshooting** - Commands and testing patterns

---

## Quick Reference

### Service Ports & Folders
```
API Gateway:      8080   (api-gateway/)
Auth Service:     8000   (services/auth-service/)
FM Service:       8001   (services/fm-service/)
CRM Service:      8002   (services/crm-service/)
HR Service:       8003   (services/hr-service/)
M Service:        8004   (services/mfg-service/)
PM/PRJ Service:   8005   (services/prj-service/)
SCM Service:      8006   (services/scm-service/)
EAM Service:      8007   (services/eam-service/)
PLM Service:      8008   (services/plm-service/)
QMS Service:      8009   (services/qms-service/)
```

### Essential Commands
```bash
./scripts/setup-secrets.sh --auto    # Generate credentials
make build                            # Build all services
make run                              # Start all services
make health                           # Check service health
make test                             # Run all tests
make logs                             # View logs
make stop                             # Stop all services
make clean                            # Full cleanup
```

### Database Per Service
```
auth_db, fm_db, crm_db, hr_db, m_db, pm_db, scm_db, eam_db, plm_db, qms_db
```

---

## Example Interactions

### Example 1: Add a New Feature
```
User: @erp-specialist How do I add a new "Commission" feature to the FM service?

Agent Response:
1. Define the endpoint in services/fm-service/contracts/fm.cdd
2. Implement handler in services/fm-service/internal/api/handlers/
3. Add business logic in services/fm-service/internal/business/service/
4. Update the database models and migrations
5. Test with make test in the service directory
6. Document in services/fm-service/README.md
```

### Example 2: Event-Driven Integration
```
User: @erp-specialist How do I ensure SCM is notified when CRM creates an order?

Agent Response:
1. In CRM service: Publish event 'crm.order.created' to Kafka
2. In SCM service: Subscribe to this topic in event consumer
3. Update inventory based on order details
4. Publish 'scm.inventory.updated' back
5. Both use the standard Kafka publisher: publisher.Publish(ctx, topic, event)
```

---

**Status**: Active  
**Last Updated**: 2026-06-19  
**Scope**: ERP Microservices System  
**Target Audience**: Developers, Architects, DevOps Engineers
