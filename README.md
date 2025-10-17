# KBTG Backend API

A simple Go backend API using the Fiber framework.

## Features

-   ✅ Go + Fiber framework
-   ✅ Hello World API endpoint
-   ✅ CORS middleware
-   ✅ Request logging
-   ✅ Health check endpoint
-   ✅ JSON responses
-   ✅ Runs on port 3000

## Getting Started

### Prerequisites

-   Go 1.21 or higher
-   Git

### Installation

1. Clone or navigate to the project directory:

```bash
cd temp-kbtg-be
```

2. Initialize and download dependencies:

```bash
go mod tidy
```

3. Run the server:

```bash
go run main.go
```

The server will start on `http://localhost:3000`

### Available Endpoints

| Method | Endpoint         | Description                 |
| ------ | ---------------- | --------------------------- |
| GET    | `/`              | Hello World main endpoint   |
| GET    | `/api/v1/health` | Health check endpoint       |
| GET    | `/api/v1/hello`  | Hello endpoint under API v1 |

### Example Responses

**GET /**

```json
{
    "message": "Hello World!",
    "status": "success",
    "data": "Welcome to KBTG Backend API"
}
```

**GET /api/v1/health**

```json
{
    "status": "ok",
    "message": "Service is running"
}
```

**GET /api/v1/hello**

```json
{
    "message": "Hello from API v1!",
    "status": "success"
}
```

## Development

### Project Structure

```
temp-kbtg-be/
├── main.go          # Main application entry point
├── go.mod           # Go module definition
├── go.sum           # Go module checksums (generated)
└── README.md        # This file
```

### Running in Development Mode

For development with auto-reload, you can install `air`:

```bash
go install github.com/cosmtrek/air@latest
air
```

## Building for Production

```bash
go build -o bin/app main.go
./bin/app
```

## Testing the API

You can test the API using curl:

```bash
# Test Hello World endpoint
curl http://localhost:3000/

# Test health check
curl http://localhost:3000/api/v1/health

# Test API v1 hello endpoint
curl http://localhost:3000/api/v1/hello
```

## Tech Stack

-   **Language**: Go 1.21+
-   **Framework**: Fiber v2
-   **Middleware**: CORS, Logger
