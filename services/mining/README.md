# Mining Control Platform

A regulatory compliance system for registering mining operations, monitoring real-time hashrate, enforcing quotas, and issuing remote shutdown commands. This service is part of the CSIC (Crypto State Infrastructure Contractor) platform.

## Features

- **National Mining Registry**: Register and manage mining operations with wallet address tracking
- **Hashrate Monitoring**: Real-time telemetry ingestion with automatic unit normalization (TH/s, PH/s, GH/s)
- **Quota Enforcement**: Dynamic and fixed quotas with automatic violation detection
- **Remote Shutdown**: Graceful, immediate, and force-kill shutdown command capabilities
- **Compliance Reporting**: Track violations, pending commands, and registry statistics

## Architecture

This service follows the **Hexagonal Architecture** (Ports & Adapters) pattern:

```
csic-platform/services/mining/
├── cmd/
│   └── main.go                 # Application entry point
├── config/
│   └── config.yaml             # Configuration file
├── internal/
│   ├── adapters/
│   │   ├── handler/http/       # HTTP API adapters
│   │   │   ├── handlers.go     # HTTP handlers
│   │   │   ├── router.go       # Gin router configuration
│   │   │   └── middleware/     # Request logging, CORS, etc.
│   │   └── repository/         # Database adapters
│   │       └── postgres/       # PostgreSQL implementation
│   │           ├── connection.go
│   │           ├── config.go
│   │           └── repository.go
│   └── core/
│       ├── domain/             # Business entities
│       │   └── models.go
│       ├── ports/              # Interface definitions
│       │   ├── repository.go   # Output ports
│       │   └── service.go      # Input ports
│       └── services/           # Business logic
│           ├── mining_service.go
│           └── mining_service_test.go
├── migrations/                 # Database migrations
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker (optional)

### Configuration

Copy the example configuration and adjust as needed:

```bash
cp config.yaml.example config.yaml
```

Key configuration options:

```yaml
app:
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "csic_platform"

compliance:
  violation_threshold: 3  # Consecutive violations before auto-shutdown
  grace_period_seconds: 300
```

### Running with Docker

```bash
docker-compose up -d
```

### Running Locally

1. Set up the database:

```bash
psql -U postgres -f migrations/001_initial_schema.up.sql
```

2. Run the application:

```bash
go run cmd/main.go
```

## API Documentation

### Registry Management

#### Register Operation
```http
POST /api/v1/operations
Content-Type: application/json

{
  "operator_name": "Mining Corp Inc",
  "wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f7547b",
  "location": "Data Center A, Nevada",
  "region": "US-WEST",
  "machine_type": "Antminer S19 Pro",
  "initial_hashrate": 100.0
}
```

Response:
```json
{
  "message": "Operation registered successfully",
  "operation": {
    "id": "uuid",
    "status": "ACTIVE",
    ...
  }
}
```

#### Get Operation
```http
GET /api/v1/operations/:id
```

#### List Operations
```http
GET /api/v1/operations?status=ACTIVE&page=1&page_size=20
```

### Hashrate Monitoring

#### Report Telemetry
```http
POST /api/v1/operations/:id/telemetry
Content-Type: application/json

{
  "hashrate": 150.5,
  "unit": "TH/s",
  "block_height": 1000000,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Get Metrics History
```http
GET /api/v1/operations/:id/metrics?limit=1000
```

### Quota Management

#### Assign Quota
```http
POST /api/v1/quotas
Content-Type: application/json

{
  "operation_id": "uuid",
  "max_hashrate": 200.0,
  "quota_type": "FIXED",
  "valid_from": "2024-01-15T00:00:00Z",
  "valid_to": "2024-12-31T23:59:59Z",
  "region": "US-WEST",
  "priority": 1
}
```

#### Get Current Quota
```http
GET /api/v1/operations/:id/quota
```

### Remote Shutdown

#### Issue Shutdown Command
```http
POST /api/v1/operations/:id/shutdown
Content-Type: application/json

{
  "command_type": "GRACEFUL",
  "reason": "Grid emergency - load balancing",
  "issued_by": "Grid Operator",
  "expires_in": 300
}
```

#### Acknowledge Command (for mining client)
```http
POST /api/v1/commands/:id/ack
```

#### Confirm Shutdown Execution
```http
POST /api/v1/commands/:id/confirm
Content-Type: application/json

{
  "success": true,
  "result": "Graceful shutdown completed successfully"
}
```

#### Get Pending Commands (for mining client polling)
```http
GET /api/v1/operations/:id/commands
```

### Statistics

#### Get Registry Statistics
```http
GET /api/v1/stats
```

Response:
```json
{
  "total_operations": 150,
  "active_operations": 120,
  "suspended_operations": 20,
  "shutdown_operations": 10,
  "total_hashrate": 15000.5,
  "quota_violations": 5,
  "pending_commands": 2
}
```

## Domain Models

### Operation Status
- `ACTIVE`: Operation is running normally
- `SUSPENDED`: Operation is temporarily suspended
- `SHUTDOWN_ORDERED`: Shutdown command issued
- `NON_COMPLIANT`: Quota violations detected
- `SHUTDOWN_EXECUTED`: Shutdown completed
- `PENDING_REGISTRATION`: Awaiting registration approval

### Command Types
- `GRACEFUL`: Graceful shutdown with timeout
- `IMMEDIATE`: Immediate shutdown
- `FORCE_KILL`: Force kill without cleanup

### Quota Types
- `FIXED`: Fixed maximum hashrate
- `DYNAMIC_GRID`: Grid-dependent dynamic quotas

## Compliance Rules

### Quota Enforcement
When a mining operation reports hashrate exceeding their quota:

1. Log the violation
2. Increment violation counter on the operation
3. After 3 consecutive violations (configurable):
   - Create automatic shutdown command
   - Mark operation as `NON_COMPLIANT`
   - Issue `GRACEFUL` shutdown command

### Shutdown Workflow
1. **ISSUED**: Command created, operation notified
2. **ACKNOWLEDGED**: Mining client acknowledged the command
3. **EXECUTED**: Mining client confirmed shutdown completion
4. **FAILED**: Shutdown failed (operation remains in violation)

## Testing

Run unit tests:

```bash
go test ./internal/core/services/... -v
```

Run integration tests (requires PostgreSQL):

```bash
go test ./internal/adapters/... -v
```

## Monitoring

Health check endpoint:
```http
GET /health
```

## License

Part of the CSIC Platform - All rights reserved.
