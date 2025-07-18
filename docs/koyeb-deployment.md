# Koyeb Deployment Guide

This guide will help you deploy your Algeria Market API service to Koyeb.

## Prerequisites

1. A Koyeb account
2. Your Go application code pushed to a Git repository (GitHub, GitLab, etc.)
3. Database credentials (PostgreSQL recommended for production)
4. Google Cloud Storage credentials (if using GCS)

## Deployment Steps

### 1. Build and Test Locally

First, test your Docker build locally:

```bash
# Build the Docker image
docker build -t algeria-market-api .

# Test the container locally
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e GCS_BUCKET_NAME=your-bucket-name \
  -e DB_HOST=your-db-host \
  -e DB_USER=your-db-user \
  -e DB_PASSWORD=your-db-password \
  -e DB_NAME=your-db-name \
  -e DB_PORT=5432 \
  algeria-market-api
```

### 2. Deploy to Koyeb

#### Option A: Deploy via Koyeb Dashboard (Dockerfile Method - Recommended)

1. Go to [Koyeb Dashboard](https://app.koyeb.com/)
2. Click "Create App"
3. Choose "GitHub" or your preferred Git provider
4. Select your repository
5. Configure the deployment:

**Build Settings:**
- Build Command: Leave empty (Dockerfile will handle this)
- Run Command: Leave empty (CMD in Dockerfile)
- Dockerfile: `Dockerfile`

#### Option B: Deploy via Koyeb Dashboard (Procfile Method - Alternative)

If the Dockerfile method doesn't work:

1. Go to [Koyeb Dashboard](https://app.koyeb.com/)
2. Click "Create App"
3. Choose "GitHub" or your preferred Git provider
4. Select your repository
5. Configure the deployment:

**Build Settings:**
- Build Command: `go build -o algeria-market-api-service main.go`
- Run Command: `./algeria-market-api-service`
- Procfile: `Procfile`

1. Go to [Koyeb Dashboard](https://app.koyeb.com/)
2. Click "Create App"
3. Choose "GitHub" or your preferred Git provider
4. Select your repository
5. Configure the deployment:

**Environment Variables:**
Set the following environment variables in Koyeb:

```env
# Application
PORT=8080
GIN_MODE=release  # This suppresses debug log messages in production

# Database Configuration
DB_HOST=your-database-host
DB_USER=your-database-user
DB_PASSWORD=your-database-password
DB_NAME=your-database-name
DB_PORT=5432

# Google Cloud Storage (if using GCS)
GCS_BUCKET_NAME=your-gcs-bucket-name
GCS_PROJECT_ID=your-gcs-project-id
GCS_CREDENTIALS_FILE={"type":"service_account","project_id":"your-project","private_key_id":"...","private_key":"-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n","client_email":"...","client_id":"...","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_x509_cert_url":"..."}

# Appwrite Configuration (if using Appwrite)
APPWRITE_ENDPOINT=https://cloud.appwrite.io/v1
APPWRITE_PROJECT=your-appwrite-project-id
APPWRITE_KEY=your-appwrite-api-key
APPWRITE_BUCKET_ID=your-appwrite-bucket-id

# JWT Secret (generate a secure random string)
JWT_SECRET=your-secure-jwt-secret-key
```

#### Option C: Deploy via Koyeb CLI

1. Install Koyeb CLI:
```bash
# macOS
brew install koyeb/tap/cli

# Linux
curl -fsSL https://cli.koyeb.com/install.sh | bash
```

2. Login to Koyeb:
```bash
koyeb login
```

3. Deploy your app:
```bash
koyeb app init algeria-market-api \
  --docker . \
  --ports 8080:http \
  --env PORT=8080 \
  --env GIN_MODE=release \
  --env DB_HOST=your-db-host \
  --env DB_USER=your-db-user \
  --env DB_PASSWORD=your-db-password \
  --env DB_NAME=your-db-name \
  --env DB_PORT=5432 \
  --env GCS_BUCKET_NAME=your-bucket-name \
  --env JWT_SECRET=your-jwt-secret
```

### 3. Database Setup

#### Option A: Use Koyeb Database

1. Create a PostgreSQL database in Koyeb
2. Use the provided connection details in your environment variables

#### Option B: Use External Database

- **Supabase**: Free PostgreSQL hosting
- **PlanetScale**: MySQL hosting
- **Railway**: PostgreSQL hosting
- **Neon**: Serverless PostgreSQL

### 4. Google Cloud Storage Setup (Optional)

If you're using GCS for file storage:

1. **Create a GCS bucket** in Google Cloud Console
2. **Create a service account** with Storage Admin permissions
3. **Download the JSON credentials file**
4. **Set the credentials as an environment variable in Koyeb:**
   - Key: `GCS_CREDENTIALS_FILE`
   - Value: The **entire JSON content** of your service account key (not the file path)

**Important:** Copy the entire JSON content from your service account key file, not just the filename. The JSON should look like:
```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "your-service-account@your-project.iam.gserviceaccount.com",
  "client_id": "...",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "..."
}
```

### 5. Environment Variables Reference

| Variable | Required | Description | Default |
|----------|----------|-------------|---------|
| `PORT` | Yes | Port to run the server on | `8080` |
| `GIN_MODE` | No | Gin framework mode | `debug` |
| `DB_HOST` | Yes | Database host | `localhost` |
| `DB_USER` | Yes | Database username | `admin` |
| `DB_PASSWORD` | Yes | Database password | `securepass` |
| `DB_NAME` | Yes | Database name | `main` |
| `DB_PORT` | Yes | Database port | `5434` |
| `GCS_BUCKET_NAME` | Yes | GCS bucket name | - |
| `GCS_PROJECT_ID` | No | GCS project ID | - |
| `GCS_CREDENTIALS_FILE` | No | GCS credentials JSON | - |
| `APPWRITE_ENDPOINT` | No | Appwrite endpoint | `https://cloud.appwrite.io/v1` |
| `APPWRITE_PROJECT` | No | Appwrite project ID | - |
| `APPWRITE_KEY` | No | Appwrite API key | - |
| `APPWRITE_BUCKET_ID` | No | Appwrite bucket ID | - |
| `JWT_SECRET` | Yes | JWT signing secret | - |

### 6. Health Check

The application includes a health check endpoint at `/ping` that returns:

```json
{
  "message": "pong"
}
```

### 7. Monitoring and Logs

- **Logs**: View application logs in the Koyeb dashboard
- **Metrics**: Monitor CPU, memory, and network usage
- **Health Checks**: Automatic health checks every 30 seconds

### 8. Custom Domain (Optional)

1. In Koyeb dashboard, go to your app
2. Click "Settings" → "Domains"
3. Add your custom domain
4. Configure DNS records as instructed

### 9. Scaling

Koyeb automatically scales your application based on traffic. You can also manually configure:

- **Min instances**: Minimum number of running instances
- **Max instances**: Maximum number of instances
- **CPU/Memory limits**: Resource allocation per instance

### 10. Troubleshooting

#### Common Issues:

1. **"no command to run your application"**: 
   - **Solution 1**: Use the Dockerfile approach (recommended)
   - **Solution 2**: Use the Procfile approach with build/run commands
   - **Solution 3**: Set explicit run command: `./algeria-market-api-service`

2. **Build fails**: Check if all dependencies are in `go.mod`
3. **Database connection fails**: Verify database credentials and network access
4. **GCS errors**: 
   - **Error**: `cannot read credentials file: open account_keys.json: no such file or directory`
   - **Solution**: Set `GCS_CREDENTIALS_FILE` environment variable with the **entire JSON content** (not file path)
   - **Alternative**: Leave `GCS_CREDENTIALS_FILE` empty to use Application Default Credentials
5. **Port binding**: Make sure `PORT` environment variable is set

#### Debug Commands:

```bash
# Check container logs
koyeb app logs algeria-market-api

# Check app status
koyeb app get algeria-market-api

# Restart app
koyeb app restart algeria-market-api
```

### 11. Security Best Practices

1. **Environment Variables**: Never commit secrets to your repository
2. **Database**: Use strong passwords and enable SSL
3. **JWT Secret**: Use a long, random string for JWT signing
4. **CORS**: Configure CORS origins properly for production
5. **HTTPS**: Koyeb provides automatic HTTPS

### 12. Performance Optimization

1. **Database Indexing**: Ensure proper database indexes
2. **Connection Pooling**: Configure database connection pooling
3. **Caching**: Implement Redis caching if needed
4. **CDN**: Use Koyeb's CDN for static assets

## Support

If you encounter issues:

1. Check the application logs in Koyeb dashboard
2. Verify all environment variables are set correctly
3. Test the application locally with the same environment variables
4. Contact Koyeb support if the issue persists 