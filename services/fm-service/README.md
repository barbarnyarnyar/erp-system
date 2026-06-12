# Financial Management Service (fm-service)

The Financial Management Service is a core microservice of the ERP system responsible for handling all financial operations including:

- Chart of Accounts management
- Transaction processing and journal entries
- Financial reporting (Balance Sheet, Income Statement, Cash Flow)
- Budget management and variance analysis
- Tax calculations and compliance

## Architecture

This service follows Domain-Driven Design (DDD) principles with clean architecture:

- **Domain Layer**: Core business entities and rules
- **Business Layer**: Services and application logic
- **API Layer**: HTTP handlers and routes
- **Data Layer**: Repositories and database models
- **Infrastructure**: External integrations and cross-cutting concerns

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- RabbitMQ (optional for messaging)

### Installation

1. Clone the repository and navigate to fm-service:
```bash
cd fm-service
```

2. Copy environment configuration:
```bash
cp .env.example .env
```

3. Install dependencies:
```bash
make deps
```

4. Run the service:
```bash
make run
```

The service will start on port 8001 by default.

### Docker

Build and run with Docker:

```bash
make docker-build
make docker-run
```

## API Endpoints

### Health Check
- `GET /health` - Service health status

### Legal Entities
- `GET /api/v1/legal-entities` - List legal entities
- `POST /api/v1/legal-entities` - Create legal entity
- `GET /api/v1/legal-entities/:id` - Get legal entity

### Accounts
- `GET /api/v1/accounts` - List all accounts
- `POST /api/v1/accounts` - Create new account
- `GET /api/v1/accounts/:id` - Get specific account
- `PUT /api/v1/accounts/:id` - Update account
- `DELETE /api/v1/accounts/:id` - Delete account
- `GET /api/v1/accounts/:id/balance` - Get account balance

### Journal Entries
- `GET /api/v1/journal-entries` - List journal entries
- `POST /api/v1/journal-entries` - Create journal entry
- `GET /api/v1/journal-entries/:id` - Get journal entry with lines
- `PUT /api/v1/journal-entries/:id` - Update journal entry
- `DELETE /api/v1/journal-entries/:id` - Delete journal entry

### Invoices (AR)
- `GET /api/v1/invoices` - List invoices
- `POST /api/v1/invoices` - Create invoice
- `GET /api/v1/invoices/:id` - Get invoice details
- `PUT /api/v1/invoices/:id` - Update invoice
- `DELETE /api/v1/invoices/:id` - Delete invoice
- `POST /api/v1/invoices/:id/send` - Send invoice
- `GET /api/v1/invoices/:id/lines` - Get invoice lines (flat model compatibility)

### Vendor Bills (AP)
- `GET /api/v1/vendor-bills` - List vendor bills
- `POST /api/v1/vendor-bills` - Create vendor bill
- `GET /api/v1/vendor-bills/:id/lines` - Get vendor bill lines

### Payments & Banking
- `GET /api/v1/payments` - List payments
- `POST /api/v1/payments` - Record payment
- `GET /api/v1/payments/:id` - Get payment details
- `GET /api/v1/bank-statements/:id/lines` - Get bank statement lines

### Fixed Assets
- `GET /api/v1/assets` - List assets
- `POST /api/v1/assets/capitalize` - Capitalize fixed asset
- `GET /api/v1/assets/:id` - Get asset details
- `POST /api/v1/assets/:id/depreciation-schedule` - Generate depreciation schedule
- `POST /api/v1/assets/depreciate` - Post monthly depreciation

### Reports
- `GET /api/v1/reports/balance-sheet` - Balance Sheet report
- `GET /api/v1/reports/income-statement` - Income Statement report
- `GET /api/v1/reports/cash-flow` - Cash Flow report

## Development

### Available Make Commands

- `make build` - Build the application
- `make run` - Run the application
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage
- `make lint` - Run linter
- `make dev` - Run in development mode with hot reload
- `make clean` - Clean build artifacts

### Database Migrations

- `make migrate-up` - Run migrations up
- `make migrate-down` - Run migrations down
- `make migrate-create name=migration_name` - Create new migration

## Configuration

The service uses environment variables for configuration. See `.env.example` for available options.

Key configuration areas:
- **Server**: Port and environment settings
- **Database**: PostgreSQL connection details
- **Cache**: Redis configuration for caching
- **Messaging**: RabbitMQ for event publishing

## Testing

Run unit tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

## Monitoring

The service exposes metrics and health endpoints for monitoring:

- Health check: `GET /health`
- Metrics: Prometheus-compatible metrics (when enabled)
- Logging: Structured JSON logs with correlation IDs

## Contributing

1. Follow the existing code structure and patterns
2. Add unit tests for new functionality
3. Update documentation as needed
4. Ensure all tests pass before submitting

## License

Part of the ERP System - Internal Use