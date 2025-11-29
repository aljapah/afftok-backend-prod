# AffTok Backend - Production Ready

This is a fully production-ready backend for the AffTok affiliate marketing platform, built with Go, Gin, GORM, PostgreSQL, and Redis.

## Quick Start

### Prerequisites
- Go 1.23+
- PostgreSQL 12+
- Redis 6+
- Docker (optional)

### Local Development

1. **Clone and Setup**
```bash
cd backend
go mod download
```

2. **Configure Environment**
```bash
cp .env.example .env
# Edit .env with your local database and Redis credentials
```

3. **Run Server**
```bash
go run ./cmd/api
```

Server will start on `http://localhost:8080`

### Docker Development

```bash
docker-compose up -d
```

This starts:
- PostgreSQL on port 5432
- Redis on port 6379
- API server on port 8080

## Project Structure

```
backend/
├── cmd/
│   ├── api/              # Main application entry point
│   └── worker/           # Background worker (optional)
├── internal/
│   ├── cache/            # Redis cache module
│   ├── config/           # Configuration loader
│   ├── database/         # Database connection
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Data models
│   └── services/         # Business logic
├── pkg/
│   ├── logger/           # Logging utilities
│   ├── utils/            # Helper functions
│   └── validator/        # Validation utilities
├── migrations/           # Database migrations
├── Dockerfile            # Production Docker image
├── docker-compose.yml    # Local development compose
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── .env.production      # Production environment template
└── PRODUCTION_DEPLOYMENT.md  # Deployment guide
```

## Key Features

### ✅ Production Ready
- Centralized configuration management
- Dependency injection pattern
- Error handling and logging
- Connection pooling
- Health check endpoint

### ✅ Database
- PostgreSQL with GORM ORM
- Automatic migrations
- UUID support
- Transaction support

### ✅ Caching
- Redis integration
- Cache helper functions
- Connection pooling

### ✅ Authentication
- JWT token-based auth
- Refresh token support
- Role-based access control

### ✅ API Features
- RESTful endpoints
- CORS middleware
- Request validation
- Error responses

### ✅ Containerization
- Multi-stage Docker build
- Docker Compose for development
- Production-optimized image

## Configuration

All configuration is managed through environment variables. See `.env.production` for all available options.

### Required Variables
```
POSTGRES_URL=postgresql://user:password@host:5432/db
REDIS_URL=redis://:password@host:6379
JWT_SECRET=your-secret-key
PORT=8080
ENV=production
```

## Building for Production

### Build Binary
```bash
go build -o server ./cmd/api
```

### Build Docker Image
```bash
docker build -t afftok-backend:latest .
```

### Run Docker Container
```bash
docker run -d \
  -p 8080:8080 \
  -e POSTGRES_URL="..." \
  -e REDIS_URL="..." \
  -e JWT_SECRET="..." \
  -e ENV="production" \
  afftok-backend:latest
```

## API Documentation

### Health Check
```
GET /health
```

### Authentication
```
POST /api/auth/register
POST /api/auth/login
POST /api/auth/refresh
POST /api/auth/logout
```

### Offers
```
GET /api/offers
GET /api/offers/:id
POST /api/offers/:id/join
```

### Users
```
GET /api/users
GET /api/users/:id
PUT /api/profile
```

### Teams
```
GET /api/teams
GET /api/teams/:id
POST /api/teams
POST /api/teams/:id/join
POST /api/teams/:id/leave
```

### Badges
```
GET /api/badges
GET /api/badges/my
```

### Clicks & Tracking
```
GET /api/c/:id
GET /api/clicks/my
GET /api/clicks/:id/stats
```

### Admin
```
POST /api/admin/offers
PUT /api/admin/offers/:id
DELETE /api/admin/offers/:id
```

## Database Models

### Users
- ID (UUID)
- Username, Email
- Password (hashed)
- Full Name, Bio, Avatar
- Role, Status
- Points, Level
- Statistics (clicks, conversions, earnings)

### Offers
- ID (UUID)
- Title, Description
- Image, Logo, Destination URL
- Category, Payout, Commission
- Status, Network

### Clicks & Conversions
- Click tracking with device/browser/OS info
- Conversion tracking with status
- Earnings calculation

### Teams
- Team management
- Member tracking
- Points aggregation

### Badges
- Achievement system
- User badge tracking
- Points rewards

## Development

### Run Tests
```bash
go test ./...
```

### Format Code
```bash
go fmt ./...
```

### Lint Code
```bash
golangci-lint run
```

### View Dependencies
```bash
go mod graph
```

## Deployment

For detailed deployment instructions, see [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md)

### Quick Deploy to Cloud

**AWS ECS:**
```bash
aws ecr get-login-password | docker login --username AWS --password-stdin $ECR_REGISTRY
docker tag afftok-backend:latest $ECR_REGISTRY/afftok-backend:latest
docker push $ECR_REGISTRY/afftok-backend:latest
```

**Google Cloud Run:**
```bash
gcloud run deploy afftok-backend \
  --image gcr.io/project/afftok-backend \
  --set-env-vars POSTGRES_URL="...",REDIS_URL="..."
```

## Monitoring

### Health Endpoint
```bash
curl http://localhost:8080/health
```

### Logs
```bash
docker logs afftok-api
```

### Database
```bash
psql -h localhost -U user -d afftok
```

### Redis
```bash
redis-cli -h localhost -p 6379
```

## Security

- ✅ Environment variables for secrets
- ✅ JWT token authentication
- ✅ Password hashing (bcrypt)
- ✅ CORS protection
- ✅ SQL injection prevention (GORM)
- ✅ SSL/TLS support

## Performance

- Connection pooling (10 idle, 100 max)
- Redis caching
- Efficient queries with GORM
- Request validation
- Error handling

## Troubleshooting

### Port Already in Use
```bash
lsof -i :8080
kill -9 <PID>
```

### Database Connection Failed
- Check POSTGRES_URL format
- Verify database is running
- Check credentials

### Redis Connection Failed
- Check REDIS_URL format
- Verify Redis is running
- Check password if required

## Support

For issues or questions, please refer to [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md) or contact the development team.

## License

Proprietary - AffTok Platform

## Version

- Backend Version: 1.0.0
- Go Version: 1.23+
- Last Updated: November 2025
