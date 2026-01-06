# Regulatory Reports Service

A comprehensive microservice for generating regulatory compliance reports, built with Go and designed for enterprise blockchain compliance operations.

## Overview

The Regulatory Reports Service is a critical component of the CSIC Platform that provides automated generation, scheduling, and management of regulatory compliance reports for blockchain-based financial operations. This service addresses the demanding requirements of modern regulatory frameworks by enabling organizations to generate comprehensive reports for Anti-Money Laundering (AML), Counter Terrorist Financing (CTF), Know Your Customer (KYC), Suspicious Activity Reports (SARs), and other regulatory filings required by financial authorities worldwide.

Blockchain transactions present unique challenges for regulatory compliance due to their pseudonymous nature, cross-border reach, and technical complexity. This service provides the tools and automation necessary to transform raw blockchain data into meaningful, actionable reports that satisfy regulatory requirements while minimizing operational overhead. By integrating with other CSIC Platform components, the service can access comprehensive transaction data, compliance flags, risk assessments, and audit trails needed for accurate reporting.

The service supports multiple regulatory frameworks including the Financial Action Task Force (FATF) recommendations, the European Union's Markets in Crypto-Assets (MiCA) regulation, the US Bank Secrecy Act (BSA), and the General Data Protection Regulation (GDPR). Each report type is designed to meet the specific formatting and content requirements of its target regulatory framework, ensuring that reports are accepted without requiring manual intervention or reformatting.

Report generation can be triggered on-demand for immediate needs or scheduled for regular intervals such as daily, weekly, monthly, or quarterly periods. The scheduling system supports cron expressions for flexible timing configurations and can automatically distribute reports to designated recipients via email or other delivery channels. Reports are generated in multiple formats including PDF for formal filings, Excel for further analysis, CSV for data processing, and JSON for system integration.

## Features

The Regulatory Reports Service offers a comprehensive set of features designed to meet the diverse reporting requirements of blockchain financial operations. Multi-framework support enables generation of reports compliant with FATF, MiCA, BSA, GDPR, and other regulatory frameworks. Each report type includes appropriate data fields, calculations, and formatting specific to its target framework, ensuring regulatory acceptance without manual intervention.

The service provides an extensive library of pre-configured report types covering the most common regulatory requirements. AML Monitoring Reports provide ongoing transaction surveillance and suspicious activity detection. CTF Reports specifically address counter-terrorist financing compliance. KYC Reports document customer identification and verification activities. SARs enable rapid filing of suspicious activity reports when potential violations are detected. CTR and CMIR reports support currency transaction and international monetary report filings. Compliance Summary and Risk Assessment reports provide executive-level visibility into the organization's compliance posture.

Scheduled report generation allows organizations to automate regular reporting without manual intervention. The scheduling system supports cron expressions for flexible timing and can distribute reports automatically to designated recipients. Scheduled reports can be configured with specific parameters, filters, and formatting options to meet recurring business needs. The system maintains scheduling history and provides visibility into execution status and results.

Template management enables customization of report appearance and content. Organizations can create and manage templates for each report type, defining headers, footers, logos, and custom sections. Templates support variables that are automatically populated with report-specific data, enabling consistent branding and formatting across all generated reports. Version control tracks template changes and maintains audit trails of template modifications.

Flexible export formats support various downstream use cases. PDF generation produces professional, formally formatted documents suitable for regulatory filing. Excel exports provide detailed data with multi-sheet support and formatting. CSV exports enable data processing and integration with other systems. JSON exports support API integration and programmatic access to report data.

## Architecture

The Regulatory Reports Service follows a clean architecture pattern with clear separation between domain logic, application services, and infrastructure adapters. This design ensures maintainability, testability, and extensibility while keeping the codebase organized and understandable. The architecture enables the service to evolve independently of infrastructure concerns, supporting future enhancements without requiring fundamental changes to the core logic.

The domain layer defines the core business entities including reports, schedules, templates, and report types. These entities capture the essential properties and behaviors of regulatory reporting, providing a stable foundation for the application logic. The domain models are designed to be technology-agnostic, representing concepts that are universally applicable to regulatory reporting regardless of the specific regulatory frameworks being addressed.

The application layer contains the service implementations that orchestrate business operations. The Report Service handles all report-related operations including creation, generation, and lifecycle management. The Schedule Service manages scheduled report execution and cron-based triggering. The Template Service handles template creation and management. Each service is designed to be independently testable and can operate without knowledge of the specific infrastructure implementations being used.

The infrastructure layer implements the adapters that connect the application to external systems. Repository implementations provide database access using PostgreSQL for persistent storage of reports, schedules, and templates. The report generator creates reports in various formats including PDF, Excel, CSV, and JSON. The messaging adapter integrates with Kafka for event streaming to downstream systems. HTTP handlers implement the REST API using the Gin framework.

## Getting Started

### Prerequisites

Before deploying the Regulatory Reports Service, ensure your environment meets the following requirements. The service is tested primarily on Linux systems but should run on any platform supported by Go 1.21 or later. Docker and Docker Compose are required for containerized deployments, while manual deployments require Go 1.21 or later, PostgreSQL 14 or later, Redis 7 or later, and access to a Kafka instance.

You will need sufficient storage for report archives, with the specific requirements depending on your retention policies and report generation frequency. The service stores generated reports locally by default but can be configured to use S3-compatible storage for larger deployments. Network connectivity is required to access the CSIC Platform components that provide the underlying data for report generation.

### Installation

Clone the repository and navigate to the service directory. The service uses Go modules for dependency management, so no additional installation steps are required beyond having Go installed. Dependencies are automatically resolved when building the service.

For containerized deployment, the included Dockerfile builds an optimized image that can be deployed to any container runtime. Build the image using standard Docker commands and tag it appropriately for your container registry. The image is based on a minimal Alpine Linux image with the Go runtime, ensuring a small attack surface and fast startup times.

### Configuration

The service is configured through the `config.yaml` file, which provides a centralized location for all operational parameters. Configuration can be overridden through environment variables for container-friendly deployments where secrets and sensitive values are injected at runtime.

The configuration file is organized into logical sections covering application settings, database connections, Redis configuration, Kafka settings, report generation behavior, schedule definitions, and regulatory framework configurations. Each section includes sensible defaults that can be customized for your specific deployment requirements. The report storage section supports both local filesystem and S3-compatible storage backends.

Database configuration requires connection details for the PostgreSQL instance. The schema is automatically applied on startup through migration files, ensuring the database is properly initialized. Connection pooling parameters can be tuned to match the expected load characteristics of your deployment.

### Running the Service

The service can be started using Docker Compose for development and testing environments. The included `docker-compose.yml` file defines a complete stack including PostgreSQL, Redis, Kafka, and the Regulatory Reports Service. This provides a quick way to get a functioning environment for evaluation purposes.

For production deployments, the service binary should be started directly or managed through your organization's standard service management approach. The service accepts command-line flags for controlling log verbosity and configuration file paths. Health check endpoints are exposed on the API port, enabling integration with container orchestrators and load balancers.

## API Reference

The Regulatory Reports Service exposes a RESTful API for report management, scheduling, and template operations. All API endpoints require authentication via API key, which should be included in the `X-API-Key` header of requests. The base URL for API requests is determined by the server configuration, typically `http://localhost:8082` for local deployments.

### Reports

The reports endpoints provide access to report operations. Create a new report by sending a POST request with the report configuration including name, type, format, and optional filters and parameters. List all reports with optional filtering by type, status, or creation date. Retrieve, generate, or delete individual reports using their unique identifiers.

### Report Types

The report types endpoints provide information about available report configurations. List all available report types to see descriptions, supported formats, regulatory frameworks, and retention requirements for each type. Get detailed information about specific report types including their parameters and default configurations.

### Schedules

The schedules endpoints provide access to scheduled report configurations. Create a new schedule by specifying the report type, generation frequency using cron syntax, recipients, and optional parameters. List all schedules with filtering by report type or enabled status. Update, trigger, or delete schedules as needed.

### Templates

The templates endpoints provide access to report template management. Create custom templates for each report type, defining headers, footers, and custom sections with variables for dynamic content. List, update, or delete templates as needed. Templates support versioning to track changes over time.

## Development

### Project Structure

The project follows Go project layout conventions with clear separation between layers. The `cmd` directory contains the entry point for the service. The `internal` directory houses all application code organized by package responsibility. The `migrations` directory contains database migration files. The `templates` directory stores report template files.

Domain models reside in `internal/domain`, defining the data structures and interfaces that represent core business concepts. Repository implementations in `internal/repository` handle database operations. Service implementations in `internal/service` contain the business logic for report generation and scheduling. HTTP handlers in `internal/handler` implement the REST API. The generator package in `internal/generator` handles report format generation. The messaging package in `internal/messaging` handles Kafka integration.

### Building

The project uses Go modules for dependency management. Build the service by running `go build -o bin/reports-service ./cmd` from the project root. This produces a statically linked binary suitable for deployment in containerized environments. The build process supports several compile-time options for features that may require additional dependencies.

Run tests using `go test ./...` to execute the test suite. The project maintains test coverage for critical paths including business logic, repository operations, and API handlers. Integration tests verify database and messaging interactions in test environments.

### Extending

Adding support for new regulatory frameworks requires adding the framework configuration to `config.yaml` and potentially creating new report type definitions. The existing report generation infrastructure can be extended to support new frameworks by implementing the appropriate report type handler and template.

Custom report formats can be added by implementing new format handlers in the generator package. This is useful for organizations that require specific output formats beyond the standard PDF, Excel, CSV, and JSON options.

API extensions should follow the established patterns in the handler layer. New endpoints are added by implementing handler methods and registering routes in the application setup code. Authentication and validation middleware can be applied to protect sensitive endpoints.

## Contributing

Contributions to the Regulatory Reports Service are welcome and encouraged. Before beginning work on a significant feature or bug fix, please open an issue to discuss the proposed changes. This helps ensure that efforts align with the project's direction and prevents duplication of work.

Code contributions should follow the project's coding standards emphasizing clarity, testability, and documentation. All contributions must pass the automated test suite and lint checks before being merged. The project uses conventional commit messages for changelog generation.

## License

The Regulatory Reports Service is proprietary software. All rights are reserved. Use of this software is subject to the terms and conditions established in your licensing agreement. Contact the project maintainers for licensing inquiries.
