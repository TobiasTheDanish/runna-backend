# Runna Backend API

Go backend API for tracking running sessions.

## Prerequisites

- Go 1.23 or higher
- Docker and Docker Compose
- Turso database account and connection URL

## Setup

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Update `.env` with your Turso database URL:
```
DATABASE_URL=libsql://your-database-url.turso.io?authToken=your-auth-token
```

## Running Locally

To run the backend in watch mode (hot reload), use [air](https://github.com/air-verse/air):

```bash
# Install air if you haven't already
go install github.com/air-verse/air@latest

# Run the project
air
```

Or run normally:

```bash
go run cmd/api/main.go
```

## Running with Docker

```bash
docker-compose up --build
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Health Check
```
GET /health
```

### Create Session
```
POST /api/sessions
Content-Type: application/json

{
  "date": "2024-01-15T10:00:00Z",
  "distance": 5.5,
  "duration": 1800,
  "notes": "Morning run"
}
```

Response: `201 Created`

### Get Sessions
```
GET /api/sessions?start_date=2024-01-01&end_date=2024-01-31
```

Query Parameters:
- `start_date` (optional): Start date in YYYY-MM-DD format. Default: 1 month ago
- `end_date` (optional): End date in YYYY-MM-DD format. Default: today

Response: `200 OK`

## Database Schema

### sessions table
- `id`: INTEGER PRIMARY KEY
- `date`: DATETIME - Date and time of the run
- `distance`: REAL - Distance in kilometers
- `duration`: INTEGER - Duration in seconds
- `notes`: TEXT - Optional notes
- `created_at`: DATETIME
- `updated_at`: DATETIME
