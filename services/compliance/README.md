# Compliance Module

A comprehensive regulatory compliance system for managing licensing workflows, compliance scoring, obligation tracking, and audit trails. This service is part of the CSIC (Crypto State Infrastructure Contractor) platform.

## Features

### 1. Licensing Workflow Engine
- Complete license application lifecycle (Draft → Submitted → Under Review → Approved/Rejected)
- License issuance, suspension, and revocation
- Automatic license number generation
- Expiration tracking and renewal reminders
- Support for multiple license types (Exchange, Custody, Mining, Trading, etc.)

### 2. Compliance Scoring System
- Automated compliance score calculation (0-100 scale)
- Tiered classification (Gold, Silver, Bronze, At-Risk, Critical)
- Score breakdown by category
- Historical score tracking and trends
- Deductions for overdue obligations and suspended licenses

### 3. Obligation Management
- Create and track regulatory obligations
- Deadline monitoring and overdue detection
- Evidence submission and fulfillment tracking
- Priority-based obligation management
- Automatic status updates for overdue items

### 4. Audit Support Tools
- Complete audit trail for all compliance activities
- Filter by entity, actor, resource, action type, and date range
- Immutable audit records with timestamps
- IP address and user agent tracking
- Support for regulatory audits and investigations

## Architecture

This service follows the **Hexagonal Architecture** (Ports & Adapters) pattern:

```
csic-platform/services/compliance/
├── cmd/
│   └── main.go                     # Application entry point
├── config/
│   └── config.yaml                 # Configuration file
├── internal/
│   ├── adapters/
│   │   ├── handler/http/           # HTTP API adapters
│   │   │   ├── handlers.go         # HTTP handlers
│   │   │   ├── router.go           # Gin router configuration
│   │   │   └── middleware/         # Request logging, CORS, Request-ID
│   │   └── repository/             # Database adapters
│   │       └── postgres/           # PostgreSQL implementation
│   │           ├── connection.go
│   │           ├── config.go
│   │           └── repository.go
│   └── core/
│       ├── domain/                 # Business entities
│       │   └── models.go
│       ├── ports/                  # Interface definitions
│       │   ├── repository.go       # Output ports
│       │   └── service.go          # Input ports
│       └── services/               # Business logic
│           ├── license_service.go
│           ├── compliance_service.go
│           ├── obligation_service.go
│           ├── audit_service.go
│           └── license_service_test.go
├── migrations/                     # Database migrations
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
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
  port: 8081

database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "csic_platform"

licensing:
  default_validity_days: 365
  renewal_warning_days: 30

scoring:
  base_score: 100.0
  overdue_deduction: 10.0
  suspended_license_deduction: 50.0
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

### License Application Endpoints

#### Submit Application
```http
POST /api/v1/licenses/applications
Content-Type: application/json

{
  "entity_id": "uuid",
  "type": "NEW",
  "license_type": "EXCHANGE",
  "requested_terms": "Exchange license for trading",
  "documents": ["doc1.pdf"]
}
```

#### Review Application
```http
PATCH /api/v1/licenses/applications/:id/review
Content-Type: application/json

{
  "approved": true,
  "reviewer_id": "uuid",
  "notes": "All requirements met",
  "granted_terms": "Full exchange license",
  "conditions": "Standard conditions"
}
```

#### List Applications
```http
GET /api/v1/licenses/applications?status=SUBMITTED&page=1&page_size=20
```

### License Endpoints

#### Issue License
```http
POST /api/v1/licenses
Content-Type: application/json

{
  "entity_id": "uuid",
  "license_type": "EXCHANGE",
  "license_number": "LCC-2024-000001",
  "expiry_days": 365,
  "jurisdiction": "US",
  "issued_by": "Admin"
}
```

#### Suspend License
```http
POST /api/v1/licenses/:id/suspend
Content-Type: application/json

{
  "reason": "Compliance investigation in progress"
}
```

#### Revoke License
```http
POST /api/v1/licenses/:id/revoke
Content-Type: application/json

{
  "reason": "Regulatory violation"
}
```

#### Get Expiring Licenses
```http
GET /api/v1/licenses/expiring?days=30
```

### Entity Endpoints

#### Register Entity
```http
POST /api/v1/entities
Content-Type: application/json

{
  "name": "Crypto Exchange Pro",
  "legal_name": "Crypto Exchange Pro LLC",
  "registration_num": "CRYPTO-2024-001",
  "jurisdiction": "US",
  "entity_type": "EXCHANGE",
  "address": "123 Blockchain Ave, NY",
  "contact_email": "compliance@cryptoexchange.com",
  "risk_level": "MEDIUM"
}
```

### Compliance Scoring Endpoints

#### Get Compliance Score
```http
GET /api/v1/entities/:id/compliance/score
```

Response:
```json
{
  "id": "uuid",
  "entity_id": "uuid",
  "total_score": 85.5,
  "tier": "SILVER",
  "breakdown": "{\"base_score\": 100, \"overdue_deductions\": -10, ...}",
  "calculated_at": "2024-01-15T10:30:00Z"
}
```

#### Recalculate Score
```http
POST /api/v1/entities/:id/compliance/score/recalculate
```

#### Get Score History
```http
GET /api/v1/entities/:id/compliance/score/history?limit=10
```

#### Get Compliance Statistics
```http
GET /api/v1/compliance/stats
```

### Obligation Endpoints

#### Create Obligation
```http
POST /api/v1/obligations
Content-Type: application/json

{
  "entity_id": "uuid",
  "regulation_id": "uuid",
  "description": "Submit quarterly audit report",
  "due_date": "2024-03-31T23:59:59Z",
  "priority": 1,
  "evidence_refs": "report.pdf"
}
```

#### Get Entity Obligations
```http
GET /api/v1/entities/:id/obligations
```

#### Get Overdue Obligations
```http
GET /api/v1/obligations/overdue
```

#### Fulfill Obligation
```http
POST /api/v1/obligations/:id/fulfill
Content-Type: application/json

{
  "evidence": "Quarterly report submitted on time"
}
```

#### Check Overdue Obligations
```http
POST /api/v1/obligations/check-overdue
```

### Audit Endpoints

#### Get Audit Logs
```http
GET /api/v1/audit-logs?entity_id=uuid&action_type=CREATE&from=2024-01-01&to=2024-12-31
```

#### Get Entity Audit Trail
```http
GET /api/v1/entities/:id/audit-trail?limit=100&offset=0
```

## Domain Models

### License Status
- `ACTIVE`: License is valid and in force
- `SUSPENDED`: License temporarily suspended
- `REVOKED`: License permanently revoked
- `EXPIRED`: License has expired

### Application Status
- `DRAFT`: Application is being prepared
- `SUBMITTED`: Application submitted for review
- `UNDER_REVIEW`: Application under active review
- `APPROVED`: Application approved
- `REJECTED`: Application rejected
- `WITHDRAWN`: Application withdrawn by applicant

### Compliance Tier
- `GOLD`: Score ≥ 90
- `SILVER`: Score ≥ 75
- `BRONZE`: Score ≥ 60
- `AT_RISK`: Score ≥ 40
- `CRITICAL`: Score < 40

### Obligation Status
- `PENDING`: Obligation not yet fulfilled
- `IN_PROGRESS`: Obligation is being worked on
- `FULFILLED`: Obligation completed
- `OVERdue`: Obligation past due date
- `WAIVED`: Obligation waived by regulator

## Scoring Algorithm

The compliance score is calculated as follows:

```
Base Score: 100 points

Deductions:
- -10 points per overdue obligation
- -50 points per suspended license

Bonuses:
- +5 points for early fulfillment (optional)

Final Score = Base Score - Overdue Deductions - Suspended License Deductions + Bonuses
```

The score is clamped between 0 and 100.

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

Response:
```json
{
  "status": "healthy",
  "service": "compliance-module"
}
```

## License

Part of the CSIC Platform - All rights reserved.
