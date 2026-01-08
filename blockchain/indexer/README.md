# Blockchain Indexer Service

A high-performance microservice for indexing and querying blockchain data, built with Go and designed for scalability and reliability in enterprise environments.

## Overview

The Blockchain Indexer Service is a critical component of the CSIC Platform that continuously monitors blockchain networks, extracts relevant transaction and state data, and makes it readily available through a performant API. This service is essential for maintaining real-time visibility into blockchain activities, supporting compliance operations, and enabling sophisticated analytics and reporting capabilities.

This service is designed to handle high-throughput blockchain data ingestion while maintaining low-latency query responses. It supports multiple blockchain networks and provides flexible indexing strategies that can be configured to suit various use cases, from simple transaction tracking to complex smart contract event analysis. The architecture follows clean design principles, ensuring maintainability and extensibility as blockchain ecosystems evolve.

The indexer operates by connecting to blockchain nodes, either through direct RPC connections or by subscribing to node event streams, depending on the network's capabilities. It processes blocks and transactions in real-time, extracting and transforming data according to predefined schemas before persisting it to the database. This data then becomes available for querying through the service's RESTful API, enabling downstream applications to build rich user experiences on top of blockchain activity.

## Features

The Blockchain Indexer Service offers a comprehensive set of features designed to meet the demanding requirements of enterprise blockchain applications. Real-time block processing ensures that data is available within seconds of on-chain activity, enabling applications to react immediately to important events. The service supports multiple blockchain networks simultaneously, allowing organizations to monitor diverse ecosystems from a single deployment.

Smart contract event indexing provides deep visibility into decentralized applications and token activities. The service can parse arbitrary contract ABIs and extract specific events, filtering noise to focus on the data that matters most to your application. This capability is particularly valuable for DeFi protocols, NFT platforms, and any application that relies on understanding complex smart contract interactions.

Transaction tracking and tracing capabilities allow for detailed analysis of fund flows and contract interactions. The service maintains comprehensive transaction metadata, including gas consumption, status, and linked events, enabling forensic analysis and compliance verification. Address monitoring features support watchlist functionality, alerting applications when monitored addresses interact with the blockchain.

A powerful query API provides flexible access to all indexed data. Complex filters, pagination, and aggregation capabilities enable efficient data retrieval without overwhelming network resources. The API supports both synchronous queries for interactive applications and webhook notifications for event-driven architectures.

## Architecture

The Blockchain Indexer Service is built using a layered architecture that separates concerns and enables independent scaling of components. This design choice facilitates maintenance and allows the service to evolve without requiring wholesale changes to the codebase.

At the foundation lies the configuration layer, which loads settings from `config.yaml` and environment variables. This approach supports containerized deployments and enables the same binary to be configured differently across environments. The configuration system supports hot reload for certain parameters, minimizing the need for restarts during operational adjustments.

The domain layer defines the core business entities and interfaces that govern the service's behavior. This includes block and transaction representations, event models, and repository interfaces. By isolating domain logic, the service remains focused on its core purpose and is protected from infrastructure concerns bleeding into business code.

The application layer contains the services and use cases that orchestrate business operations. Indexer services manage the blockchain data ingestion pipelines, processing blocks and events according to configured strategies. Query services handle data retrieval operations, applying caching and optimization techniques to ensure responsive performance.

The infrastructure layer implements the adapters and repositories that connect the service to external systems. Database repositories handle persistence using PostgreSQL with TimescaleDB extensions for time-series optimization. Blockchain adapters connect to node APIs using configurable providers that abstract network-specific details. Message publishers integrate with Kafka for event distribution to downstream systems.

## Getting Started

### Prerequisites

Before deploying the Blockchain Indexer Service, ensure that your environment meets the following requirements. The service is tested primarily on Linux systems, though it should run on any platform supported by the Go runtime. Docker and Docker Compose are required for containerized deployments, while manual deployments require Go 1.21 or later, PostgreSQL 14 or later, and access to a Kafka instance.

You will need network access to the blockchain nodes you intend to index. For production deployments, dedicated infrastructure is recommended, including sufficient storage for historical data and network bandwidth to maintain synchronization with network tip. The specific resource requirements depend heavily on the blockchains being indexed and the indexing depth required.

### Installation

To set up the Blockchain Indexer Service, begin by cloning the repository and navigating to the service directory. The service uses Go modules for dependency management, so no additional installation steps are required beyond having Go installed. Dependencies are automatically resolved when building the service.

For containerized deployment, the included Dockerfile builds an optimized image that can be deployed to any container runtime. The image is based on a minimal Go runtime image and includes all necessary components for operation. Build the image using Docker's standard build commands, tagging it appropriately for your container registry.

Manual deployment requires compiling the binary and ensuring all configuration files are in place. The service is a single binary with no external runtime dependencies beyond the database and message queue infrastructure. This simplicity makes it straightforward to deploy using configuration management tools or container orchestration systems.

### Configuration

The service is configured through the `config.yaml` file, which provides a centralized location for all operational parameters. Configuration can be overridden through environment variables, enabling container-friendly deployments where secrets and sensitive values are injected at runtime.

The configuration file is organized into logical sections covering database connections, blockchain network definitions, indexing behavior, and API server settings. Each blockchain network requires its own configuration block specifying connection details, network parameters, and indexing preferences. The service supports both HTTP and WebSocket connections to blockchain nodes, automatically selecting the appropriate protocol based on network capabilities.

Database configuration requires connection details for the PostgreSQL instance, including host, port, credentials, and database name. The schema is automatically applied on startup through migration files, ensuring the database is properly initialized. Connection pooling parameters can be tuned to match the expected load characteristics of your deployment.

### Running the Service

The service can be started using Docker Compose for development and testing environments. The included `docker-compose.yml` file defines a complete stack including PostgreSQL, Kafka, and the indexer service. This provides a quick way to get a functioning environment for evaluation purposes.

For production deployments, the service binary should be started directly or managed through your organization's standard service management approach. The service accepts command-line flags for controlling log verbosity and configuration file paths. Health check endpoints are exposed on the API port, enabling integration with container orchestrators and load balancers.

Monitor the service logs during initial startup to verify successful database migration and blockchain connection establishment. The indexer will begin processing blocks from its configured starting point, which can be a specific block height or the genesis block for full historical indexing. Progress is logged regularly, providing visibility into the synchronization status.

## API Reference

The Blockchain Indexer Service exposes a RESTful API for querying indexed data and managing indexing operations. All API endpoints require authentication via API key, which should be included in the `X-API-Key` header of requests. The base URL for API requests is determined by the server configuration, typically `http://localhost:8080` for local deployments.

### Blocks

The blocks endpoints provide access to indexed block data. The primary endpoint for retrieving blocks accepts a block height or hash parameter, returning comprehensive block information including transactions, gas consumption, and miner details. The response structure is normalized across supported blockchains, abstracting network-specific variations in block format.

Block listings support pagination and filtering by time range. This is particularly useful for generating reports or analyzing blockchain activity over specific periods. The response includes pagination metadata enabling efficient traversal of large result sets without memory overhead.

### Transactions

Transaction endpoints expose detailed transaction data including inputs, outputs, value transfers, and execution status. Transactions can be retrieved by their hash identifier, with responses including all available metadata such as gas consumption, nonce, and linked events. This granular data supports both display requirements and analytical use cases.

Transaction search capabilities enable filtering by address, time range, or value characteristics. Complex queries can be constructed using the query parameter syntax, supporting combinations of filters for precise data retrieval. Results are returned in a consistent format regardless of the source blockchain network.

### Events

Event endpoints provide access to smart contract event logs that have been indexed by the service. Events can be filtered by contract address, event signature, or block range. This capability is essential for applications that need to track specific on-chain activities such as token transfers, governance actions, or protocol interactions.

The event indexing system supports dynamic ABI loading, enabling the service to decode event parameters without requiring preconfiguration of contract schemas. Decoded events include both the raw log data and the human-readable parameter values, facilitating both programmatic processing and user interface display.

### Monitoring

Health and status endpoints provide operational visibility into the service's state. The health endpoint returns the service's current operational status and connectivity to dependencies. Metrics endpoints expose Prometheus-compatible metrics for integration with monitoring infrastructure.

Blockchain synchronization status is available through dedicated endpoints, showing current block height, indexed height, and synchronization progress. This information is valuable for determining data freshness and identifying synchronization issues before they impact downstream applications.

## Development

### Project Structure

The project follows Go project layout conventions with clear separation between layers. The `cmd` directory contains the entry point for the service, while `internal` houses all application code organized by package responsibility. This structure prevents accidental import cycles and clarifies the public API surface of each package.

Domain models reside in `internal/domain`, defining the data structures and interfaces that represent core business concepts. The repository interfaces defined here provide abstraction over database implementations, enabling different storage backends without changes to application code. Service implementations in `internal/service` contain the business logic for indexing and querying operations.

Infrastructure code lives in `internal/repository` and `internal/indexer`, implementing the adapters that connect to external systems. Database repositories use Go's standard database interface, supporting any SQL database with appropriate drivers. Indexer implementations encapsulate blockchain-specific logic, making it straightforward to add support for new networks.

### Building

The project uses Go modules for dependency management, with versioned dependencies declared in `go.mod`. To build the service, run `go build -o bin/indexer ./cmd` from the project root. This produces a statically linked binary suitable for deployment in containerized environments.

The build process supports several compile-time options through Go build tags. These include features that may require additional dependencies or are specific to certain deployment scenarios. Review the build command in the Dockerfile for the recommended configuration options.

Testing is integrated into the build process, with unit tests covering core functionality and integration tests verifying database and blockchain interactions. Run `go test ./...` to execute the test suite. The project maintains high test coverage for critical paths, with coverage reports generated during continuous integration builds.

### Extending

Adding support for new blockchain networks requires implementing the `BlockReader` and `EventSubscriber` interfaces defined in the domain package. Each blockchain implementation should be placed in its own package under `internal/blockchain`, following the pattern established by existing implementations. The configuration system supports registration of new blockchain types through package initialization.

Custom indexing strategies can be implemented by extending the `Indexer` service with specialized processing logic. This is useful for applications that need to extract application-specific data from transactions or events. The clean architecture makes it straightforward to add new processing stages without modifying core indexing infrastructure.

API extensions should follow the established patterns in the handler layer. New endpoints are added by implementing the `Handler` interface and registering routes in the application setup code. Authentication and validation middleware can be applied to protect sensitive endpoints or enforce rate limiting.

## Contributing

Contributions to the Blockchain Indexer Service are welcome and encouraged. Before beginning work on a significant feature or bug fix, please open an issue to discuss the proposed changes. This helps ensure that efforts align with the project's direction and prevents duplication of work.

Code contributions should follow the project's coding standards, which emphasize clarity, testability, and documentation. All contributions must pass the automated test suite and lint checks before being merged. The project uses conventional commit messages for changelog generation, following the Conventional Commits specification.

## License

The Blockchain Indexer Service is proprietary software. All rights are reserved. Use of this software is subject to the terms and conditions established in your licensing agreement. Contact the project maintainers for licensing inquiries.
