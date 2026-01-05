# Audit Log Service

A tamper-evident audit logging service for the CSIC Platform, built with Go following Clean Architecture principles.

## Features

- **Tamper-Evident Logging**: Hash-chained audit entries ensure data integrity
- **Event-Driven Architecture**: Publishes events to Kafka for downstream processing
- **Full-Text Search**: Search and filter audit entries by multiple criteria
- **Chain Verification**: Verify the integrity of the audit chain
- **Compliance Reporting**: Generate compliance and activity reports
- **Real-time Processing**: Process audit events from Kafka in real-time

## Architecture

```
services/audit-log/
├── cmd/
│   └── main.go                     # Application entry point
├── internal/
│   ├── adapter/
│   │   ├── messaging/              # Kafka producer/consumer
│   │   └── repository/             # PostgreSQL data access
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── core/
│   │   ├── domain/
│   │   │   └── models.go           # Domain entities
│   │   ├── ports/
│   │   │   └── interfaces.go       # Interface definitions
│   │   └── service/
│   │       └── audit_service.go    # Business logic
│   └── handler/
│       └── http_handler.go         # HTTP API handlers
├── migrations/
│   └── 001_init.sql               # Database schema
├── config.yaml                     # Configuration file
├── Dockerfile                      # Container definition
├── docker-compose.yml              # Local development setup
└── go.mod                          # Go module definition
```

## API Endpoints

### Entry Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/audit/entries` | Create new audit entry |
| GET | `/api/v1/audit/entries/:id` | Get entry by ID |
| POST | `/api/v1/audit/entries/search` | Search entries |
| GET | `/api/v1/audit/entries/search` | Quick search (GET) |
| GET | `/api/v1/audit/trace/:trace_id` | Get entries by trace ID |

### Verification

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/audit/entries/:id/verify` | Verify single entry |
| GET | `/api/v1/audit/verify` | Verify chain (with start_id) |

### Chain Information

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/audit/chain/summary` | Get chain statistics |

### Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/audit/reports/compliance` | Generate compliance report |
| POST | `/api/v1/audit/reports/activity` | Generate activity report |

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Apache Kafka
- Docker (optional)

### Local Development

1. **Start dependencies with Docker Compose**:
   ```bash
   docker-compose up -d
   ```

2. **Build and run the service**:
   ```bash
   go build -o main ./cmd/main.go
   ./main
   ```

### Docker Deployment

```bash
# Build the image
docker build -t audit-log .

# Run the container
docker run -p 8080:8080 \
  -v $(pwd)/config.yaml:/config/config.yaml \
  audit-log
```

### Configuration

Configuration is managed via `config.yaml`:

```yaml
app:
  name: "audit-log"
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 5432
  username: "audit"
  password: "audit_password"
  name: "audit_db"

messaging:
  type: "kafka"
  brokers:
    - "localhost:9092"
  topic_prefix: "audit"
```

Environment variables can override config values:
- `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`
- `KAFKA_BROKERS`
- `APP_PORT`, `APP_HOST`, `APP_DEBUG`
- `CONFIG_PATH`

## Hash Chaining

The audit log uses hash chaining to ensure tamper-evidence:

1. Each entry contains a `previous_hash` field referencing the previous entry
2. The `current_hash` is calculated as: `SHA256(previous_hash + entry_data)`
3. If any entry is modified, the hash chain breaks
4. The verification endpoint detects any tampering

Example hash calculation:
```
entry_data = {
  "id": "entry002",
  "trace_id": "trace-001",
  "actor_id": "user-admin",
  "action": "CREATE",
  "timestamp": "2024-01-15T10:30:00Z",
  "previous_hash": "abc123..."
}

current_hash = SHA256(previous_hash + JSON.stringify(entry_data))
```

## Example Usage

### Create Audit Entry

```bash
curl -X POST http://localhost:8080/api/v1/audit/entries \
  -H "Content-Type: application/json" \
  -d '{
    "trace_id": "trace-001",
    "actor_id": "user-admin",
    "actor_type": "user",
    "action": "CREATE",
    "resource": "evidence",
    "resource_id": "evidence-123",
    "operation": "evidence_management",
    "outcome": "success",
    "severity": "info",
    "payload": {"name": "Suspicious File"},
    "metadata": {"source": "email"}
  }'
```

### Search Entries

```bash
curl -X POST http://localhost:8080/api/v1/audit/entries/search \
  -H "Content-Type: application/json" \
  -d '{
    "actor_id": "user-admin",
    "action": "CREATE",
    "start_time": "2024-01-15T00:00:00Z",
    "end_time": "2024-01-15T23:59:59Z",
    "page": 1,
    "page_size": 20
  }'
```

### Verify Chain

```bash
curl "http://localhost:8080/api/v1/audit/entries/entry001/verify"
```

### Generate Compliance Report

```bash
curl -X POST http://localhost:8080/api/v1/audit/reports/compliance \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z"
  }'
```

## Kafka Events

The service publishes and consumes the following Kafka events:

### Published Events

| Topic | Event Type | Description |
|-------|------------|-------------|
| `audit.entries` | `AUDIT_ENTRY_CREATED` | New audit entry created |
| `audit.verifications` | `AUDIT_VERIFICATION` | Chain verification result |

### Consumed Events

The service listens for audit events from various services to automatically log them.

## Supported Audit Actions

### Authentication
- `LOGIN`, `LOGOUT`, `LOGIN_FAILED`
- `PASSWORD_CHANGE`, `MFA_ENABLE`, `MFA_DISABLE`
- `SESSION_CREATE`, `SESSION_REVOKE`

### Data Operations
- `CREATE`, `READ`, `UPDATE`, `DELETE`
- `EXPORT`, `IMPORT`, `SEARCH`
- `DOWNLOAD`, `UPLOAD`

### System Operations
- `CONFIG_CHANGE`, `PERMISSION_GRANT`, `PERMISSION_REVOKE`
- `ROLE_ASSIGN`, `ROLE_REMOVE`

### Security Events
- `THREAT_DETECTED`, `INTRUSION_ATTEMPT`
- `ACCESS_DENIED`, `POLICY_VIOLATION`

### API Gateway Events
- `API_REQUEST`, `RATE_LIMIT_HIT`, `AUTH_REQUIRED`

## Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "service": "audit-log",
  "time": "2024-01-15T10:30:00Z"
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests (`go test ./...`)
5. Submit a pull request

## License

This project is licensed under the MIT License.
