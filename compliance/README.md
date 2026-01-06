# Compliance Service

A comprehensive compliance checking service for the CSIC Platform that evaluates transactions against regulatory rules, detects violations, and maintains an audit trail of all compliance activities.

## Overview

The Compliance Service serves as the central regulatory enforcement component of the CSIC Platform, implementing a sophisticated rules engine that evaluates all transactions against a configurable set of compliance rules. The service supports multiple rule types including Anti-Money Laundering (AML), Know Your Customer (KYC), sanctions screening, geographic restrictions, transaction limits, and frequency controls. By integrating with the platform's message bus, the service can process transactions in real-time as they occur, providing immediate compliance feedback.

The architecture follows clean architecture principles, separating the core business logic from infrastructure concerns. This design enables flexible deployment options and straightforward testing of compliance rules independent of external dependencies. The service maintains its own PostgreSQL database for rule storage and compliance history, while using Redis for high-performance caching and rate limiting operations.

## Key Features

The Compliance Service provides a comprehensive framework for regulatory compliance management. At its core, the service implements a flexible rules engine supporting multiple rule types, each with configurable parameters and severity levels. Rules can be enabled or disabled without service restarts, and the system supports rule versioning to track changes over time. Each rule is classified by type (AML, KYC, Sanctions, Transaction, Geographic, Amount, Frequency, Custom) and assigned a severity level (Info, Low, Medium, High, Critical), enabling prioritized enforcement.

The AML checking capabilities include real-time screening against internal blacklists and external watchlists. The service maintains a watchlist database that can be populated from various sources including government sanctions lists, politically exposed persons lists, and custom organizational watchlists. Transaction parties are automatically screened against these lists with configurable match score thresholds to balance false positive rates against compliance requirements.

Geographic compliance enforcement supports both allowlist and blocklist approaches. The service can restrict transactions based on the countries involved, supporting complex scenarios such as allowing transactions within specific regions while blocking high-risk jurisdictions. This capability is essential for maintaining compliance with international sanctions programs and internal risk policies.

Transaction monitoring includes amount-based controls with configurable minimum and maximum thresholds, as well as frequency controls that limit the number of transactions within a time window. These controls help detect structuring attempts, unusual trading patterns, and other potentially suspicious activities. The frequency tracking uses Redis for high-performance counting across distributed deployments.

## Architecture

The Compliance Service follows the hexagonal architecture pattern, organizing code into distinct layers with clear separation of concerns. The domain layer contains the core business entities including Rule, Transaction, ComplianceResult, Violation, and Entity models. These entities encapsulate all business logic and validation rules, remaining independent of any framework or infrastructure concerns.

The application layer implements the compliance checking workflow, orchestrating the evaluation of transactions against the rule set. The ComplianceService coordinates rule retrieval, result aggregation, and violation handling. This layer depends only on domain interfaces, enabling straightforward unit testing without external dependencies.

The adapter layer provides concrete implementations of domain interfaces for infrastructure concerns. HTTP handlers expose the REST API using the Gin framework. Database adapters handle persistence operations with PostgreSQL, while Redis adapters provide caching and rate limiting. Kafka adapters enable message-based processing of incoming transactions and publishing of compliance events.

The configuration layer centralizes all service configuration including database connections, Redis settings, Kafka topics, and rules engine parameters. Configuration is loaded from a YAML file with environment variable override support, following the twelve-factor methodology for cloud-native deployments.

## API Endpoints

The Compliance Service exposes a RESTful API for compliance checking and rule management. All endpoints are prefixed with the base path `/api/v1`.

### Compliance Checking

The compliance check endpoint accepts a transaction payload and returns a comprehensive compliance evaluation. The request includes transaction details such as source and target identifiers, amounts, countries, and metadata. The response includes the overall compliance status, risk score, detailed results for each rule evaluated, and any violations detected. This endpoint supports both synchronous and asynchronous processing modes depending on the configuration.

Cached compliance results can be retrieved using the transaction ID, enabling retrieval of historical check results without re-evaluation. This capability is useful for audit trails and dispute resolution scenarios where previous compliance decisions need to be verified.

### Rule Management

The rule management endpoints enable CRUD operations on compliance rules. Creating a rule requires specifying the rule name, type, severity, and parameters. Rule types correspond to different compliance domains, and parameters control the specific behavior of each rule type. Rules can be scheduled for expiration and automatically become inactive after their expiration date.

Rules can be retrieved individually by ID or listed by type. The list endpoint supports pagination for environments with large numbers of rules. Updates to rules automatically increment the version number and invalidate any cached rule sets, ensuring that all compliance checks use the most recent rule definitions.

### Entity Management

Entity management endpoints support the registration and screening of entities subject to compliance checks. Entities represent individuals, organizations, or accounts that participate in transactions. Each entity can be marked as blacklisted, watchlisted, or KYC verified, with these attributes influencing compliance check results.

### Violation Management

The violation management endpoints provide visibility into compliance violations and support their resolution. Open violations can be listed with pagination and filtering. Violations can be resolved by providing a resolution description and the identifier of the resolving party. Resolved violations are retained for audit purposes but excluded from open violation lists.

### Watchlist Management

Watchlist endpoints enable management of the screening database. Entries can be added from various sources including government sanctions lists and internal risk assessments. The search endpoint supports fuzzy matching to identify potential matches against watchlist entries, with configurable match score thresholds.

## Configuration

The service is configured through a YAML configuration file located at `config.yaml` in the service directory or `/etc/csic/compliance/config.yaml` in containerized deployments. Environment variables can override file-based configuration using the `CSIC_` prefix with keys converted to uppercase and periods replaced by underscores.

Database configuration specifies PostgreSQL connection parameters including host, port, credentials, and database name. Connection pool settings control the maximum number of concurrent connections. Redis configuration includes connection details, key prefix for namespace isolation, and pool size for concurrent operations.

Kafka configuration defines the broker addresses, consumer group, and topic names for transaction input and violation output. The rules engine configuration controls cache TTL, maximum rules per check, evaluation timeout, and parallel evaluation settings.

## Database Schema

The Compliance Service uses PostgreSQL for persistent storage with a schema optimized for compliance operations. The rules table stores all compliance rules with their parameters and metadata. The rulesets table groups rules into named collections that can be activated together. The entities table maintains information about parties subject to compliance checks.

Compliance results are stored in a dedicated table with JSONB columns for the check details and violation information. This schema design enables flexible storage of variable-length compliance data while maintaining query performance on indexed fields. The violations table tracks individual compliance violations with their resolution status.

The watchlist table maintains screening entries with their sources and match scores. All tables include standard audit columns for created and updated timestamps, supporting compliance requirements for maintaining audit trails.

## Getting Started

Development setup requires Go 1.21 or later, PostgreSQL, Redis, and Kafka. Begin by cloning the repository and navigating to the compliance service directory. Install dependencies using Go module commands, which will resolve all required packages.

Create a configuration file based on the provided example. Modify database connection parameters to match your local PostgreSQL instance. Ensure that the database user has sufficient privileges to create tables. Run database migrations either manually or using the embedded migration tool.

Start the service in development mode to verify configuration and connectivity. The service will apply migrations, connect to Redis and Kafka, and begin processing transactions. Use the health endpoint to verify that all dependencies are accessible.

## Running with Docker

Docker support enables containerized deployment for development and production environments. The Dockerfile builds a minimal image using multi-stage compilation, resulting in a small attack surface suitable for production use. The image includes only the compiled binary and necessary CA certificates.

Docker Compose provides a complete development environment including PostgreSQL, Redis, Kafka, and the compliance service. This configuration is suitable for local development, integration testing, and demonstration purposes. Execute docker-compose up to start all components with appropriate networking configuration.

For production deployments, customize the Docker Compose configuration to integrate with your infrastructure. Configure external database connections, logging destinations, and resource limits according to your operational requirements. Consider using Kubernetes for orchestration in large-scale deployments.

## Rule Types

The Compliance Service supports multiple rule types addressing different compliance domains. Each rule type has specific parameters that control its behavior and validation criteria.

AML rules screen transactions against blacklists and watchlists. The service compares transaction parties against the entity database and watchlist, flagging matches above configurable thresholds. High-severity AML violations trigger immediate blocking of the transaction and alert generation.

KYC rules verify that transaction parties have completed required identity verification procedures. Transactions from entities without verified KYC status can be flagged or blocked based on organizational policy. This capability supports regulatory requirements for customer due diligence.

Sanctions rules enforce government sanctions programs by blocking transactions involving sanctioned countries, entities, or individuals. The blocked countries parameter specifies jurisdictions subject to sanctions, with transactions involving these countries automatically failing compliance checks.

Transaction rules validate the structural integrity of transactions, ensuring that all required fields are present and properly formatted. This rule type catches malformed transactions early in the processing pipeline.

Geographic rules enforce geographic restrictions on transactions. The allowed countries parameter specifies permitted jurisdictions, while the blocked countries parameter specifies prohibited jurisdictions. Transactions involving non-allowed countries can be flagged or blocked based on configuration.

Amount rules enforce transaction value limits. Minimum and maximum thresholds flag or block transactions outside acceptable ranges. These rules help implement daily limits, single transaction limits, and other value-based controls.

Frequency rules limit transaction rates within configurable time windows. The service tracks transaction counts per entity using Redis sorted sets, enabling distributed rate limiting across multiple service instances. Transactions exceeding limits are flagged based on severity configuration.

Custom rules support organization-specific compliance logic through expression evaluation. Rules can evaluate transaction metadata against patterns or invoke external validation services through configurable expressions.

## Integration

The Compliance Service integrates with other CSIC Platform components through multiple mechanisms. Kafka message consumption enables real-time processing of transactions as they occur. The service subscribes to the transactions topic and processes each transaction through the compliance engine.

Violation events are published to a dedicated Kafka topic for consumption by alerting and reporting systems. Control Layer integration enables automatic policy updates when compliance rules change. Audit logging sends compliance events to the Audit Log Service for comprehensive audit trail maintenance.

## Contributing

Contributions to the Compliance Service are welcome and appreciated. Before starting work on significant changes, please open an issue to discuss the proposed modifications. This ensures that contributions align with project direction and architectural principles.

Code contributions should include appropriate tests and documentation updates. The project follows standard Go testing conventions with coverage reporting. Code style follows effective Go guidelines with gofmt formatting. Pull requests must pass all automated checks before merging.

## License

This service is part of the CSIC Platform and is licensed under the same terms as other platform components. See the platform license file for detailed licensing information.
