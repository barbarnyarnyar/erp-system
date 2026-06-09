# ERP Specialist Skill

## Overview

This skill provides **comprehensive expertise** for working with the ERP microservices system. An agent equipped with this skill becomes an expert in:

- **System Architecture**: Understanding the 8-service microservices design
- **Domain Services**: Deep knowledge of each service's responsibilities and APIs
- **Business Logic**: How services collaborate to execute business workflows
- **Event-Driven Patterns**: Kafka-based asynchronous communication
- **Security**: Credential management and best practices
- **Development**: Building, testing, and deploying services

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

### Architecture & Design
```bash
# Make architecture decisions
@erp-specialist How should we implement a new supplier onboarding workflow across SCM and FM services?
```

### Security & Operations
```bash
# Manage security and deployment
@erp-specialist How do I set up credentials for production deployment?
```

### Performance Optimization
```bash
# Improve system performance
@erp-specialist Analyze and optimize the database query for listing all active projects with resource allocation
```

## Skill Contents

### SKILL.md
The main skill definition containing:

1. **Architecture & Infrastructure** - System design and components
2. **Domain Services** - 7 business services + Auth service
3. **Technical Stack** - Go, Gin, PostgreSQL, Redis, Kafka
4. **Development Workflows** - Build, deploy, and debug commands
5. **API Design & Conventions** - Standards across all services
6. **Contract-Driven Development** - API contracts and code generation
7. **Event-Driven Architecture** - Kafka patterns and topics
8. **Security Best Practices** - Credential management and hardening
9. **Common Tasks & Solutions** - How to accomplish typical development tasks
10. **Troubleshooting Guide** - Solutions to common issues

## Quick Reference

### Service Ports
```
API Gateway:      8080
Auth Service:     8000
FM Service:       8001
CRM Service:      8002
HR Service:       8003
M Service:        8004
PM Service:       8005
SCM Service:      8006
```

### Essential Commands
```bash
./scripts/setup-secrets.sh --auto    # Generate credentials
make build                            # Build all services
make run                              # Start all services
make health                           # Check service health
make test                             # Run all tests
make logs                             # View logs
```

### Service Entry Points
```
All services use: cmd/main.go
```

### Database Per Service
```
fm_db, hr_db, scm_db, m_db, crm_db, pm_db
```

## Example Interactions

### Example 1: Add a New Feature
```
User: @erp-specialist How do I add a new "Commission" feature to the FM service?

Agent Response:
1. Define the endpoint in services/fm-service/contracts/fm.cdd
2. Implement handler in services/fm-service/internal/api/handlers/
3. Add business logic in services/fm-service/internal/business/domain/
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

### Example 3: Troubleshooting
```
User: @erp-specialist Services won't start after I pulled latest changes

Agent Response:
1. Generate credentials: ./scripts/setup-secrets.sh --auto
2. Check config validity: docker-compose config > /dev/null
3. View logs: docker-compose logs {service-name}
4. Verify .env file exists and has all required vars
5. Check port conflicts: netstat -an | grep 808
```

## Knowledge Structure

The skill organizes ERP expertise into distinct domains:

- **Infrastructure**: How services communicate and persist data
- **Architecture**: How the system is organized and scaled
- **Development**: How to build and test features
- **Operations**: How to deploy and monitor
- **Security**: How to protect credentials and data
- **Troubleshooting**: How to diagnose and fix issues

## Integration with Other Skills

This skill works well with:

- **create-prd**: Use with ERP domain knowledge to create accurate PRDs
- **phase-prd**: Break down ERP features into implementation phases
- **pencil-design**: Design UI for ERP services with domain context

## Limitations

This skill does NOT:

- Execute code or deploy services (use make commands directly)
- Access live databases or services (use curl/psql for that)
- Replace the need for understanding Go syntax
- Make architectural decisions without your input

It DOES:

- Explain how the system works
- Guide implementation patterns
- Suggest troubleshooting steps
- Document best practices
- Provide code examples

## Maintenance

This skill should be updated when:

- New services are added to the ERP system
- Architecture patterns change
- New security procedures are implemented
- Development workflows evolve

---

**Status**: Active  
**Last Updated**: 2026-06-09  
**Scope**: ERP Microservices System  
**Target Audience**: Developers, Architects, DevOps Engineers
