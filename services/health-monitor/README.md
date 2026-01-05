# Health Monitor Service

A comprehensive health monitoring service for the CSIC Platform that aggregates system health status, tracks component availability, and provides real-time operational metrics through a RESTful API.

## Overview

The Health Monitor Service serves as the central nervous system for platform observability, continuously collecting health signals from all registered components, aggregating them into meaningful metrics, and exposing this information through a clean API interface. The service implements a polling-based health check mechanism that periodically queries registered components and stores the results for analysis and alerting purposes.

This service is built with a focus on reliability and performance, designed to handle high-frequency health check operations while maintaining minimal resource footprint. The architecture follows clean architecture principles, ensuring that the core business logic remains independent of delivery mechanisms and external dependencies.

## Key Features

The Health Monitor Service provides a comprehensive set of features designed to give platform operators complete visibility into system health. At its core, the service maintains a registry of all components that need to be monitored, allowing operators to add, remove, and configure health checks for any service or infrastructure component within the platform. Each component can be configured with custom check intervals, timeout values, and expected response criteria, providing flexibility to handle diverse monitoring requirements.

The service implements multiple health check types to accommodate different monitoring scenarios. HTTP endpoint checks verify that web services are responding correctly and returning expected status codes. TCP socket checks ensure that database connections and other network services are reachable. Custom health check handlers allow operators to implement business-specific health validation logic that goes beyond simple connectivity tests.

All health check results are persisted to PostgreSQL, creating a historical record of component availability and response times. This data enables trend analysis, outage documentation, and SLA compliance reporting. The service calculates aggregate statistics including uptime percentages, average response times, and failure rates for configurable time windows.

Real-time status aggregation provides an instant view of overall platform health. The service maintains a rolling window of recent check results and computes overall status based on configurable thresholds. This aggregation supports parent-child relationships between components, allowing hierarchical health views where the status of a composite component depends on the health of its dependencies.

## Architecture

The Health Monitor Service follows the clean architecture pattern, organizing code into distinct layers with clear separation of concerns. The domain layer contains the core business entities and interfaces that define the health monitoring domain model. This layer has no dependencies on external frameworks and expresses the fundamental concepts of components, health checks, and status results.

The application layer implements the business logic for health monitoring operations. This layer contains services that orchestrate health check execution, status aggregation, and metrics calculation. The service layer depends only on domain interfaces, making it testable and independent of specific implementations.

The adapter layer provides concrete implementations of interfaces defined in the domain layer. HTTP handlers expose the REST API using the Gin framework. Database adapters handle persistence operations with PostgreSQL. A scheduler adapter manages periodic health check execution. These adapters can be replaced without affecting the core business logic.

The configuration layer centralizes all service configuration including database connection parameters, check scheduling intervals, logging preferences, and service discovery settings. Configuration is loaded from a YAML file with environment variable override support, following the twelve-factor app methodology.

## API Endpoints

The Health Monitor Service exposes a RESTful API for managing health checks and retrieving health status information. All endpoints are prefixed with the base path `/api/v1`.

### Component Management

The component management endpoints allow operators to register and configure components for monitoring. Creating a new component registration requires providing a unique name, description, target address or URL, check type, and check interval. The service validates that required fields are present and that check configurations are sensible before persisting the registration.

Components can be updated to modify monitoring parameters or temporarily disabled without removal. The update endpoint supports partial updates, allowing modification of specific fields while preserving others. Disabling a component stops health check execution for that component while maintaining its registration and historical data.

Removing a component permanently deletes its registration and all associated historical data. This action cannot be undone, so the API requires explicit confirmation through a query parameter. Operators should consider disabling components instead of removing them if historical records need to be preserved.

### Health Status

The health status endpoints provide access to current and historical health information. The primary status endpoint returns an aggregated view of all registered components, computing overall platform health based on configured thresholds. This response includes summary statistics and a list of components sorted by their current status severity.

Individual component status endpoints return detailed information about a specific component, including its current status, last check timestamp, and historical availability metrics. The response includes response time statistics and failure counts for configurable time windows.

Historical status endpoints support outage analysis and SLA compliance reporting. These endpoints accept parameters for time range selection and aggregation granularity. Responses include time-series data showing status changes and metrics over the selected period.

### Metrics

The metrics endpoints provide programmatic access to aggregated health statistics. Uptime calculations show the percentage of time each component has been healthy over configurable periods. Response time distributions show percentiles including median, 95th percentile, and 99th percentile values.

## Configuration

The service is configured through a YAML configuration file with support for environment variable overrides. The configuration file is located at `config.yaml` in the service root directory. All configuration sections are documented below with their default values and allowed ranges.

Database configuration requires connection details for the PostgreSQL instance. The service uses the `pgx` driver for PostgreSQL connectivity, supporting connection pooling for concurrent health check execution. Database migrations are applied automatically on service startup using the migrations directory.

Scheduler configuration controls health check execution timing. The default scheduler uses a fixed-interval approach where all components are checked at their configured intervals. The scheduler supports concurrent execution with configurable worker pool size to handle high numbers of components efficiently.

Logging configuration determines output format and verbosity. JSON format is recommended for production deployments while console format with colored output is more convenient during development. Log levels range from debug through warning, with info level providing a good balance for normal operation.

## Database Schema

The health monitor service uses PostgreSQL for persistent storage with a schema designed for efficient querying of health status history and component registrations.

The components table stores registration information for all monitored components. Each record includes the component name as a unique identifier, descriptive text, the target address for health checks, the check type identifier, and configuration parameters stored as JSON. The table includes timestamps for record creation and last update.

The health_checks table stores individual health check execution results. Each record references the component that was checked, includes the check timestamp, execution duration in milliseconds, the result status, and optional response details. The table is partitioned by time range in production deployments to maintain query performance as historical data accumulates.

The aggregated_metrics table stores pre-computed statistics for performance optimization. Records are created at configurable intervals for each component, storing uptime percentages, average response times, and failure counts. This table enables efficient retrieval of historical trends without scanning individual check results.

## Getting Started

Setting up the Health Monitor Service for development requires a PostgreSQL database and Go runtime environment. Begin by cloning the repository and navigating to the service directory. Install dependencies using the Go module system which will resolve all required packages automatically.

Create a configuration file based on the provided example configuration. Modify database connection parameters to match your local PostgreSQL instance. Ensure that the database user has sufficient privileges to create tables and insert records. Run database migrations using the embedded migration tool or apply the SQL files manually.

Start the service in development mode to verify configuration and database connectivity. The service will apply migrations, connect to the database, and begin executing health checks for any registered components. Use the API to register components and verify that health checks are executing correctly.

## Running with Docker

The service includes Docker support for containerized deployments. The Dockerfile builds a minimal image containing the compiled binary and configuration files. Multi-stage compilation ensures that the runtime image contains only necessary dependencies.

The docker-compose configuration provides a complete development environment including PostgreSQL and the health monitor service. This configuration is suitable for local development and integration testing. Execute docker-compose up to start all components with proper networking configuration.

For production deployments, customize the docker-compose configuration to match your infrastructure requirements. Configure external database connections, logging destinations, and resource limits according to your operational standards.

## Health Check Types

The Health Monitor Service supports multiple health check types to accommodate diverse monitoring requirements. Each check type has specific configuration requirements and validation rules.

HTTP checks verify that web services are responding correctly. The check sends a GET request to the specified URL and validates that the response status code falls within expected ranges. Additional validation can check for specific response body content or header values. HTTP checks support authentication through bearer tokens, basic auth, and custom headers.

TCP checks verify network connectivity to services that do not expose HTTP endpoints. The check establishes a TCP connection to the specified address and port, optionally sending and validating response data. TCP checks are useful for databases, message queues, and other infrastructure components.

Custom checks execute external scripts or programs to perform application-specific health validation. The check executes the specified command with optional arguments and validates the exit code and output. Custom checks enable integration with existing monitoring tools and support complex validation logic.

## Contributing

Contributions to the Health Monitor Service are welcome and appreciated. Before starting work on significant changes, please open an issue to discuss the proposed modifications. This helps ensure that contributions align with the project direction and architectural principles.

All code contributions should include appropriate tests and documentation updates. The project uses standard Go testing conventions with coverage reporting. Code style follows the effective Go guidelines with gofmt formatting. Pull requests must pass all automated checks before merging.

## License

This service is part of the CSIC Platform and is licensed under the same terms as other platform components. See the platform license file for detailed licensing information.
