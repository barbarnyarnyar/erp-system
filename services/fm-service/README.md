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

### Accounts
- `GET /api/v1/accounts` - List all accounts
- `POST /api/v1/accounts` - Create new account
- `GET /api/v1/accounts/:id` - Get specific account
- `PUT /api/v1/accounts/:id` - Update account
- `DELETE /api/v1/accounts/:id` - Delete account
- `GET /api/v1/accounts/:id/balance` - Get account balance

### Transactions
- `GET /api/v1/transactions` - List all transactions
- `POST /api/v1/transactions` - Create new transaction
- `GET /api/v1/transactions/:id` - Get specific transaction
- `POST /api/v1/transactions/:id/post` - Post transaction
- `POST /api/v1/transactions/:id/reverse` - Reverse transaction

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