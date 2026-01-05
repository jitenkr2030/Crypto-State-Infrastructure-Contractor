# Control Layer Service

A policy decision point (PDP) service for the CSIC Platform, providing centralized access control and policy enforcement.

## Features

- **Policy Management**: Create, update, delete, and version policies
- **Access Control**: Real-time access decision checking
- **Policy Versioning**: Full history and rollback capabilities
- **Policy Templates**: Reusable policy patterns
- **Event-Driven**: Publishes events to Kafka for auditing
- **Caching**: High-performance access decisions with caching

## Architecture

```
services/control-layer/
├── cmd/
│   └── main.go                     # Application entry point
├── internal/
│   ├── adapter/
│   │   ├── messaging/              # Kafka producer
│   │   └── repository/             # PostgreSQL data access
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── core/
│   │   ├── domain/
│   │   │   └── models.go           # Domain entities
│   │   ├── ports/
│   │   │   └── interfaces.go       # Interface definitions
│   │   └── service/
│   │       └── policy_service.go   # Business logic
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

### Policy Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/control/policies` | Create new policy |
| GET | `/api/v1/control/policies` | List all policies |
| GET | `/api/v1/control/policies/:id` | Get policy by ID |
| PUT | `/api/v1/control/policies/:id` | Update policy |
| DELETE | `/api/v1/control/policies/:id` | Delete policy |

### Version History

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/control/policies/:id/history` | Get version history |
| POST | `/api/v1/control/policies/:id/restore` | Restore to version |

### Templates

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/control/templates` | Get available templates |
| POST | `/api/v1/control/templates/:id/apply` | Apply template |

### Access Control (Main Enforcement Point)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/control/check` | Check access permission |
| POST | `/api/v1/control/check/bulk` | Bulk check permissions |

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
docker build -t control-layer .

# Run the container
docker run -p 8081:8081 \
  -v $(pwd)/config.yaml:/config/config.yaml \
  control-layer
```

### Configuration

Configuration is managed via `config.yaml`:

```yaml
app:
  name: "control-layer"
  host: "0.0.0.0"
  port: 8081

database:
  host: "localhost"
  port: 5432
  username: "control"
  password: "control_password"
  name: "control_db"

messaging:
  type: "kafka"
  brokers:
    - "localhost:9092"
  topic_prefix: "control"
```

Environment variables can override config values:
- `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`
- `KAFKA_BROKERS`
- `APP_PORT`, `APP_HOST`, `APP_DEBUG`
- `CONFIG_PATH`

## Policy Structure

```json
{
  "id": "pol-example",
  "name": "Example Policy",
  "description": "Example policy description",
  "effect": "allow",
  "resources": ["api/*", "documents/*"],
  "actions": ["read", "list"],
  "subjects": ["user", "admin"],
  "conditions": {
    "time_start": "09:00",
    "time_end": "17:00"
  },
  "priority": 100,
  "version": 1,
  "is_active": true
}
```

### Policy Effects

- **allow**: Grants access when the policy matches
- **deny**: Denies access when the policy matches

### Conditions

Supported condition types:
- `time_start` / `time_end`: Restrict by time of day (HH:MM format)
- `ip_whitelist`: Restrict by IP address
- `environment`: Restrict by environment (production, staging, etc.)
- `user_agent_pattern`: Regex pattern for user agent

## Example Usage

### Create Policy

```bash
curl -X POST http://localhost:8081/api/v1/control/policies \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Read Access Policy",
    "description": "Grants read access to documents",
    "effect": "allow",
    "resources": ["documents/*"],
    "actions": ["read", "list"],
    "priority": 100
  }'
```

### Check Access

```bash
curl -X POST http://localhost:8081/api/v1/control/check \
  -H "Content-Type: application/json" \
  -d '{
    "subject": {
      "id": "user-123",
      "type": "user",
      "roles": ["user"]
    },
    "action": "read",
    "resource": {
      "type": "documents",
      "id": "doc-456"
    },
    "context": {
      "ip_address": "192.168.1.100"
    }
  }'
```

Response:
```json
{
  "allowed": true,
  "reason": "Matched policy: Read Access Policy (version 1)",
  "policy_id": "pol-example",
  "policy_name": "Read Access Policy",
  "checked_at": "2024-01-15T10:30:00Z"
}
```

### Restore Policy Version

```bash
curl -X POST http://localhost:8081/api/v1/control/policies/pol-123/restore \
  -H "Content-Type: application/json" \
  -d '{
    "version": 2,
    "reason": "Rolling back due to issues"
  }'
```

## Policy Templates

The service includes several built-in templates:

| ID | Name | Description |
|----|------|-------------|
| `read-only` | Read-Only Access | Grants read-only access to resources |
| `admin-access` | Full Admin Access | Grants full admin access |
| `restricted-delete` | Restricted Delete | Allows delete only during business hours |

Apply a template:
```bash
curl -X POST http://localhost:8081/api/v1/control/templates/read-only/apply \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Read-Only Policy"
  }'
```

## Health Check

```bash
curl http://localhost:8081/health
```

Response:
```json
{
  "status": "healthy",
  "service": "control-layer",
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
