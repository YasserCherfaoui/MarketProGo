# Technology Stack

## Core Technologies
- **Language**: Go 1.24.2
- **Web Framework**: Gin (HTTP router and middleware)
- **ORM**: GORM with PostgreSQL driver
- **Database**: PostgreSQL
- **Authentication**: JWT tokens using golang-jwt/jwt
- **Configuration**: Environment variables with godotenv support

## Key Dependencies
- **Cloud Storage**: Google Cloud Storage SDK
- **File Management**: Appwrite SDK for Go
- **Password Hashing**: bcrypt (golang.org/x/crypto)
- **CORS**: gin-contrib/cors
- **Validation**: go-playground/validator

## Build & Development Commands

### Local Development
```bash
# Install dependencies
go mod download

# Run the application
go run main.go

# Build binary
go build -o algeria-market-api-service main.go

# Run with hot reload (if using air)
air
```

### Docker
```bash
# Build Docker image
docker build -t algeria-market-api .

# Run container
docker run -p 8080:8080 algeria-market-api
```

### Environment Setup
- Copy `.env.example` to `.env` and configure required variables
- Required environment variables: `GCS_BUCKET_NAME`, database credentials, JWT secrets
- Default port: 8080 (configurable via `PORT` env var)

## Configuration
- All configuration loaded through `cfg/config.go`
- Supports both environment variables and `.env` files
- Graceful fallbacks to default values where appropriate