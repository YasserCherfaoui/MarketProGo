# Koyeb Deployment Guide

## Quick Setup

1. **Push your code** to GitHub
2. **Go to Koyeb Dashboard** and create a new app
3. **Connect your repository**
4. **Set environment variables** (see below)
5. **Deploy!**

## Environment Variables

Set these in Koyeb:

```env
# Application
PORT=8080
GIN_MODE=release

# Database
DB_HOST=your-database-host
DB_USER=your-database-user
DB_PASSWORD=your-database-password
DB_NAME=your-database-name
DB_PORT=5432

# Google Cloud Storage
GCS_BUCKET_NAME=your-bucket-name
GCS_PROJECT_ID=your-project-id
GCS_CREDENTIALS_FILE={"type":"service_account","project_id":"...","private_key":"...","client_email":"..."}

# JWT Secret
JWT_SECRET=your-secure-jwt-secret

# Appwrite (optional)
APPWRITE_ENDPOINT=https://cloud.appwrite.io/v1
APPWRITE_PROJECT=your-project-id
APPWRITE_KEY=your-api-key
APPWRITE_BUCKET_ID=your-bucket-id
```

## Important Notes

- The `Procfile` specifies `web: bin/MarketProGo` as the main process
- The `migrate: bin/migrate` process is available for database migrations
- Make sure to set `GCS_CREDENTIALS_FILE` with the **entire JSON content** (not file path)
- Set `GIN_MODE=release` for production

## Health Check

Your app includes a health check endpoint at `/ping` 