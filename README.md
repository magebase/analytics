# Analytics Service

A Go-based analytics service that tracks API usage and integrates with billing systems for PaaS platforms.

## Features

- **Event Tracking**: Track analytics events with automatic billing integration
- **Usage Analytics**: Retrieve usage statistics with cost breakdown
- **Billing Integration**: Automatic billing event generation for tracked events
- **RESTful API**: Clean HTTP endpoints for event tracking and usage retrieval
- **OpenTelemetry**: Built-in tracing and observability
- **Validation**: Input validation and error handling

## Architecture

The service follows a clean architecture pattern with:

- **App Layer**: HTTP server and routing using Fiber
- **Service Layer**: Business logic for analytics and billing
- **Models**: Data structures for events and usage
- **Testing**: Comprehensive test coverage following TDD principles

## API Endpoints

### POST /api/v1/analytics/events
Track an analytics event.

**Headers:**
- `X-API-Key`: Required API key for authentication
- `X-User-ID`: Required user identifier
- `Content-Type`: application/json

**Request Body:**
```json
{
  "event_type": "page_view",
  "page": "/home",
  "properties": {
    "referrer": "google.com",
    "utm_source": "search"
  }
}
```

**Response:**
```json
{
  "status": "success",
  "event_id": "uuid",
  "tracked_at": "timestamp",
  "billing_event_id": "uuid"
}
```

### GET /api/v1/analytics/usage
Retrieve usage statistics for a user.

**Query Parameters:**
- `user_id`: Required user identifier
- `start_date`: Start date (YYYY-MM-DD format, defaults to 30 days ago)
- `end_date`: End date (YYYY-MM-DD format, defaults to today)

**Response:**
```json
{
  "total_events": 42,
  "events_by_type": {
    "page_view": 30,
    "click": 12
  },
  "billing_summary": {
    "total_cost": 0.054,
    "cost_breakdown": {
      "page_view": 0.03,
      "click": 0.024
    },
    "currency": "USD"
  },
  "total_cost": 0.054,
  "cost_breakdown": {
    "page_view": 0.03,
    "click": 0.024
  }
}
```

## Development Setup

### Prerequisites
- Go 1.21 or later
- Git

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd analytics
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run tests:
```bash
go test ./test/ -v
```

4. Start the service:
```bash
go run main/main.go
```

The service will start on port 8080 by default. You can change this by setting the `PORT` environment variable.

## Testing (TDD Workflow)

This project follows Test-Driven Development (TDD) principles with the Red-Green-Refactor cycle:

### 🟥 RED Phase
Write failing tests that define the expected behavior:
```bash
go test ./test/ -v
# Tests should fail initially
```

### 🟩 GREEN Phase
Implement the minimum code to make tests pass:
```bash
go test ./test/ -v
# All tests should now pass
```

### ✅ Commit
After tests pass, commit with descriptive messages:
```bash
git add .
git commit -m "feat: implement API usage tracking to pass test"
```

### 🛠 REFACTOR Phase
Improve code structure and readability while maintaining test coverage.

## Project Structure

```
analytics/
├── app/                    # Application logic
│   ├── app.go             # Main app structure and HTTP handlers
│   ├── models.go          # Data models and structures
│   └── service.go         # Business logic service layer
├── main/                  # Entry point
│   └── main.go           # Main function and server startup
├── test/                  # Test files
│   └── api_usage_tracking_test.go
├── go.mod                 # Go module dependencies
└── README.md             # This file
```

## Dependencies

- **Fiber**: Fast HTTP web framework
- **OpenTelemetry**: Observability and tracing
- **UUID**: Unique identifier generation
- **Testify**: Testing utilities and assertions

## Environment Variables

- `PORT`: Server port (default: 8080)

## Contributing

1. Follow TDD workflow: Red → Green → Commit → Refactor
2. Write tests first for any new functionality
3. Ensure all tests pass before committing
4. Use descriptive commit messages
5. Follow Go coding standards and conventions

## License

[Add your license information here]
