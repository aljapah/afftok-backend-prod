# AffTok Backend - Production Deployment Guide

## Overview

This guide covers the deployment of the AffTok backend to production infrastructure. The project has been fully refactored for production readiness with centralized configuration, dependency injection, and containerization.

## What's New

### 1. Configuration Management
- **File**: `internal/config/config.go`
- Centralized configuration loader that reads from environment variables
- Supports multiple environments (development, production)
- Validates all required environment variables at startup

### 2. Database Connection Module
- **File**: `internal/database/database.go`
- Updated to use the config loader
- Connection pooling configured for production
- Automatic migrations on startup

### 3. Redis Cache Module
- **File**: `internal/cache/redis.go`
- Redis client initialization with connection pooling
- Helper functions for common cache operations
- Graceful shutdown support

### 4. Dependency Injection
- All handlers now receive database connection via constructor
- Eliminates global state dependencies
- Enables better testing and modularity

### 5. Docker Support
- **Dockerfile**: Multi-stage production build
- **docker-compose.yml**: Local development environment with PostgreSQL and Redis

## Environment Variables

Create a `.env.production` file with the following variables:

```
POSTGRES_URL=postgresql://user:password@host:5432/dbname?sslmode=require&channel_binding=require
REDIS_URL=redis://:password@host:6379
JWT_SECRET=your-secret-key-here
PORT=8080
ENV=production
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h
ALLOWED_ORIGINS=https://yourdomain.com
LOG_LEVEL=error
```

### Required Variables Explanation

| Variable | Description | Example |
|----------|-------------|---------|
| `POSTGRES_URL` | PostgreSQL connection string | `postgresql://user:pass@neon.tech/db?sslmode=require` |
| `REDIS_URL` | Redis connection URL | `redis://:password@redis.cloud:10232` |
| `JWT_SECRET` | Secret key for JWT signing | Generate a strong random string |
| `PORT` | Server port | `8080` |
| `ENV` | Environment mode | `production` or `development` |
| `JWT_EXPIRATION` | Access token expiration | `24h` |
| `JWT_REFRESH_EXPIRATION` | Refresh token expiration | `168h` (7 days) |
| `ALLOWED_ORIGINS` | CORS allowed origins | `https://yourdomain.com` |
| `LOG_LEVEL` | Logging level | `error`, `info`, `debug` |

## Deployment Methods

### Method 1: Docker (Recommended)

#### Build Image
```bash
docker build -t afftok-backend:latest .
```

#### Run Container
```bash
docker run -d \
  --name afftok-api \
  -p 8080:8080 \
  -e POSTGRES_URL="postgresql://..." \
  -e REDIS_URL="redis://..." \
  -e JWT_SECRET="your-secret" \
  -e ENV="production" \
  afftok-backend:latest
```

#### Using Docker Compose (Local Development)
```bash
docker-compose up -d
```

### Method 2: Direct Binary Deployment

#### Build Binary
```bash
go build -o server ./cmd/api
```

#### Run Server
```bash
export POSTGRES_URL="postgresql://..."
export REDIS_URL="redis://..."
export JWT_SECRET="your-secret"
export ENV="production"
export PORT="8080"

./server
```

### Method 3: Cloud Deployment (AWS, GCP, Azure)

#### AWS ECS/Fargate
1. Push Docker image to ECR
2. Create ECS task definition
3. Configure environment variables in task definition
4. Deploy service

#### Google Cloud Run
```bash
gcloud run deploy afftok-backend \
  --image gcr.io/project/afftok-backend \
  --set-env-vars POSTGRES_URL="...",REDIS_URL="..." \
  --memory 512Mi \
  --cpu 1
```

#### Azure Container Instances
```bash
az container create \
  --resource-group mygroup \
  --name afftok-backend \
  --image myregistry.azurecr.io/afftok-backend:latest \
  --environment-variables POSTGRES_URL="..." REDIS_URL="..."
```

## Database Setup

### PostgreSQL Requirements
- Version: 12+
- Extensions: uuid-ossp (auto-created on first run)
- SSL mode: Recommended for production

### Initial Setup
The application automatically runs migrations on startup. Ensure the database user has:
- CREATE TABLE permissions
- CREATE EXTENSION permissions
- Full CRUD permissions on all tables

## Redis Setup

### Redis Requirements
- Version: 6+
- Memory: Minimum 256MB
- Persistence: Enabled (RDB or AOF)

### Connection
Redis connection is established on startup. If Redis is unavailable, the server will fail to start. Ensure Redis is accessible before deploying.

## Health Checks

### Endpoint
```
GET /health
```

### Response
```json
{
  "status": "ok",
  "message": "AffTok API is running"
}
```

## Monitoring and Logging

### Logs
- Development: INFO level logs to stdout
- Production: ERROR level logs to stdout

### Key Metrics to Monitor
- Database connection pool usage
- Redis connection status
- Request latency
- Error rates
- JWT token validation failures

## Security Checklist

- [ ] Change default JWT_SECRET to a strong random value
- [ ] Use HTTPS/TLS for all connections
- [ ] Enable SSL mode for PostgreSQL
- [ ] Use strong Redis password
- [ ] Restrict CORS origins to your domain
- [ ] Enable rate limiting (if available)
- [ ] Use environment variables for all secrets (never hardcode)
- [ ] Keep dependencies updated
- [ ] Enable database backups
- [ ] Monitor error logs for suspicious activity

## Troubleshooting

### Database Connection Errors
```
Error: failed to connect to PostgreSQL
```
- Verify POSTGRES_URL format
- Check database server is running
- Verify network connectivity
- Check database credentials

### Redis Connection Errors
```
Error: failed to connect to Redis
```
- Verify REDIS_URL format
- Check Redis server is running
- Verify Redis password if required
- Check network connectivity

### Port Already in Use
```
Error: listen tcp :8080: bind: address already in use
```
- Change PORT environment variable
- Kill process using port: `lsof -i :8080`

## Performance Tuning

### Database Connection Pool
- MaxIdleConns: 10 (configurable in database.go)
- MaxOpenConns: 100 (configurable in database.go)
- ConnMaxLifetime: 1 hour

### Recommendations
- Increase MaxOpenConns for high-traffic scenarios
- Use read replicas for read-heavy workloads
- Enable query caching in Redis
- Use CDN for static assets

## Rollback Procedure

1. Keep previous Docker image tagged
2. Rollback to previous image:
```bash
docker run -d --name afftok-api-old -p 8080:8080 afftok-backend:previous-tag
```

3. Update load balancer/reverse proxy to point to old container
4. Investigate issues with new version
5. Redeploy after fixes

## Support and Maintenance

### Regular Tasks
- Monitor logs daily
- Review error rates
- Check database performance
- Verify Redis cache hit rates
- Update dependencies monthly

### Backup Strategy
- Daily database backups
- Weekly full backups
- Test restore procedures monthly
- Store backups in separate region

## API Endpoints

### Public Endpoints
- `GET /health` - Health check
- `GET /api/offers` - List all offers
- `GET /api/offers/:id` - Get specific offer
- `GET /api/c/:id` - Click tracking
- `GET /api/promoter/:id` - Promoter page
- `POST /api/rate-promoter` - Rate promoter
- `POST /api/postback` - Conversion postback
- `POST /api/auth/register` - Register user
- `POST /api/auth/login` - Login user

### Protected Endpoints (Require JWT Token)
- `GET /api/auth/me` - Get current user
- `PUT /api/profile` - Update profile
- `GET /api/users` - List users
- `POST /api/offers/:id/join` - Join offer
- And more...

### Admin Endpoints (Require Admin Role)
- `POST /api/admin/offers` - Create offer
- `PUT /api/admin/offers/:id` - Update offer
- `DELETE /api/admin/offers/:id` - Delete offer
- And more...

## Version Information

- Go Version: 1.23+
- GORM: v1.25.5
- Gin: v1.9.1
- Redis Client: v9.0.0+
- PostgreSQL Driver: v1.5.4

## Contact and Support

For issues or questions, contact the development team or submit an issue in the project repository.
