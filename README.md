# BLADE Ingestion Service

A microservice for ingesting BLADE data from Databricks into a catalog system.

## Quick Start

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Generate proto files:
   ```bash
   make proto
   ```

4. Run the service:
   ```bash
   make run
   ```

## Development

See the implementation guide in `blade-ingestion-implementation-guide/` for detailed instructions.

## API Documentation

- gRPC: localhost:9090
- REST: localhost:9091
- Swagger UI: http://localhost:9091/swagger-ui/