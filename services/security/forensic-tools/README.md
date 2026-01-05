# Forensic Tools Service

A comprehensive forensic evidence management and analysis service built with Go, following Clean Architecture principles.

## Features

- **Evidence Collection**: Collect and catalog digital evidence with full metadata tracking
- **Chain of Custody**: Maintain tamper-evident chain of custody records
- **Analysis Pipeline**: Perform various forensic analyses on collected evidence
- **Search & Discovery**: Search and filter evidence by multiple criteria
- **Secure Storage**: Store evidence files with integrity verification
- **Event Streaming**: Publish events to Kafka for downstream processing

## Architecture

```
services/security/forensic-tools/
├── cmd/
│   └── main.go                     # Application entry point
├── internal/
│   ├── adapter/
│   │   ├── messaging/              # Kafka event publisher
│   │   ├── repository/             # PostgreSQL data access
│   │   └── storage/                # File storage backend
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── core/
│   │   ├── domain/
│   │   │   └── models.go           # Domain entities
│   │   ├── ports/
│   │   │   └── interfaces.go       # Interface definitions
│   │   └── service/
│   │       └── forensic_service.go # Business logic
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

### Evidence Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/forensic/evidence` | Collect new evidence |
| POST | `/api/v1/forensic/evidence/batch` | Batch collect evidence |
| GET | `/api/v1/forensic/evidence/:id` | Get evidence by ID |
| GET | `/api/v1/forensic/evidence/:id/download` | Download evidence file |
| DELETE | `/api/v1/forensic/evidence/:id` | Delete evidence |

### Chain of Custody

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/forensic/evidence/:id/custody` | Get chain of custody |
| POST | `/api/v1/forensic/evidence/:id/custody` | Add custody record |
| POST | `/api/v1/forensic/evidence/:id/custody/verify` | Verify chain integrity |

### Analysis

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/forensic/analysis` | Start analysis |
| GET | `/api/v1/forensic/analysis/:id` | Get analysis status |
| GET | `/api/v1/forensic/analysis/:id/results` | Get analysis results |
| GET | `/api/v1/forensic/analysis` | List analyses |

### Search

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/forensic/search` | Search evidence |
| GET | `/api/v1/forensic/search?q=...` | Quick search |

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

3. **Or run with live reload** (requires air):
   ```bash
   air
   ```

### Docker Deployment

```bash
# Build the image
docker build -t forensic-tools .

# Run the container
docker run -p 8080:8080 \
  -v $(pwd)/config.yaml:/config/config.yaml \
  -v forensic-data:/data/forensic/evidence \
  forensic-tools
```

### Configuration

Configuration is managed via `config.yaml`:

```yaml
app:
  name: "forensic-tools"
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 5432
  username: "forensic"
  password: "forensic_password"
  name: "forensic_db"

storage:
  type: "local"
  base_path: "/data/forensic/evidence"

messaging:
  type: "kafka"
  brokers:
    - "localhost:9092"
  topic_prefix: "forensic"
```

Environment variables can override config values:
- `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`
- `STORAGE_PATH`
- `KAFKA_BROKERS`
- `APP_PORT`, `APP_HOST`, `APP_DEBUG`
- `CONFIG_PATH`

## Evidence Types

- `disk_image` - Disk image files (E01, VMDK, etc.)
- `memory_dump` - Memory dumps (LiME, WinPmem, etc.)
- `network_dump` - Network captures (PCAP)
- `log_file` - System and application logs
- `registry` - Windows registry hives
- `file` - Individual files
- `database` - Database exports
- `email` - Email archives
- `mobile` - Mobile device data
- `cloud` - Cloud service exports
- `other` - Other evidence types

## Analysis Types

- `hash_verification` - Calculate and verify file hashes
- `file_carving` - Recover deleted files
- `timeline_analysis` - Build filesystem timeline
- `malware_analysis` - Malware detection and analysis
- `network_analysis` - Network traffic analysis
- `memory_analysis` - Memory dump analysis
- `registry_analysis` - Windows registry analysis
- `string_extraction` - Extract strings from files
- `metadata_analysis` - File metadata analysis
- `hash_lookup` - Look up hashes in threat databases
- `yara_scan` - YARA rule scanning
- `custom` - Custom analysis plugins

## Chain of Custody

The chain of custody is maintained through:

1. **Immutable Records**: Each custody transfer creates a hash-chained record
2. **Digital Signatures**: Records can be signed by handlers
3. **Timestamp Verification**: Timestamps are validated in order
4. **Audit Trail**: All actions are logged to the audit table

## Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "service": "forensic-tools",
  "time": "2024-01-15T10:30:00Z"
}
```

## Example Usage

### Collect Evidence

```bash
curl -X POST http://localhost:8080/api/v1/forensic/evidence \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Suspicious PDF",
    "type": "file",
    "source": "email-attachment",
    "description": "Suspicious PDF from unknown sender",
    "tags": ["malware", "email"],
    "metadata": {"sender": "unknown@example.com"}
  }'
```

### Start Analysis

```bash
curl -X POST http://localhost:8080/api/v1/forensic/analysis \
  -H "Content-Type: application/json" \
  -d '{
    "evidence_id": "abc123...",
    "analysis_type": "hash_lookup",
    "parameters": {"databases": ["virustotal", "malwarebazaar"]}
  }'
```

### Search Evidence

```bash
curl -X POST http://localhost:8080/api/v1/forensic/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "malware",
    "evidence_types": ["file", "disk_image"],
    "tags": ["suspicious"],
    "page": 1,
    "page_size": 20
  }'
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests (`go test ./...`)
5. Submit a pull request

## License

This project is licensed under the MIT License.
