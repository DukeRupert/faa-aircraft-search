# FAA Aircraft Search API

A high-performance REST API for searching and retrieving FAA aircraft data. Built with Go, Echo web framework, PostgreSQL, and SQLC for type-safe database operations.

## Purpose

This application provides a searchable database of FAA aircraft data, allowing users to:
- Search aircraft by ICAO code, FAA designator, manufacturer, or model
- Retrieve detailed aircraft specifications and characteristics
- Access paginated results for large datasets
- Import and manage aircraft data from Excel files

## Architecture

- **Web Framework**: Echo v4 for high-performance HTTP routing
- **Database**: PostgreSQL with connection pooling
- **Query Builder**: SQLC for type-safe, generated database queries
- **Migration**: Custom Excel-to-database import tool
- **API Versioning**: RESTful API with v1 namespace

## Quick Start

### Prerequisites

- Go 1.24.3+
- Docker and Docker Compose
- Make (for build automation)

### Setup

1. **Clone and setup dependencies**:
   ```bash
   git clone <repository-url>
   cd faa-aircraft-search
   make deps
   ```

2. **Start the database**:
   ```bash
   make db-up
   ```

3. **Run database migrations**:
   ```bash
   make migrate-up
   ```

4. **Generate SQLC code**:
   ```bash
   make sqlc-generate
   ```

5. **Import aircraft data** (requires `aircraft_data.xlsx` file):
   ```bash
   make import-data
   ```

6. **Start the web server**:
   ```bash
   make web
   ```

7. **Test the API**:
   ```bash
   make test-api
   ```

### One-command development setup:
```bash
make dev
```

## Environment Configuration

Create a `.env` file in the project root:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=faa_aircraft
DB_HOST=localhost
DB_PORT=5432
DB_SSLMODE=disable
```

## API Endpoints

### Base URL: `http://localhost:8080`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check and database status |
| `/api/v1/aircraft/search` | GET | Search aircraft with pagination |
| `/api/v1/aircraft/:id` | GET | Get specific aircraft by ID |

### Search Parameters

- `q` (string): Search term (searches ICAO code, FAA designator, manufacturer, model)
- `page` (int): Page number (default: 1)
- `limit` (int): Results per page (default: 50, max: 100)

## Commands

For a complete list of available commands, run:

```bash
make help
```

Key commands include:
- `make dev` - Full development setup
- `make web` - Start web server
- `make db-up` - Start database
- `make import-data` - Import Excel data
- `make test-api` - Test API endpoints

## Troubleshooting

### Common Issues

1. **Database connection errors**: Ensure Docker is running and database is up
   ```bash
   make db-up
   make db-logs
   ```

2. **Environment variable errors**: Ensure `.env` file exists or use inline variables
   ```bash
   POSTGRES_USER=postgres POSTGRES_PASSWORD=postgres POSTGRES_DB=faa_aircraft make count-data
   ```

3. **SQLC compilation errors**: Regenerate after SQL changes
   ```bash
   make sqlc-generate
   ```

4. **Port conflicts**: Change port in `main.go` if 8080 is in use

## License

MIT