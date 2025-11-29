# AffTok Configuration Guide

**Complete configuration reference for all components**

---

## 📋 Table of Contents

1. [Backend Configuration](#backend-configuration)
2. [Admin Panel Configuration](#admin-panel-configuration)
3. [Mobile App Configuration](#mobile-app-configuration)
4. [Database Configuration](#database-configuration)
5. [Environment Variables](#environment-variables)

---

## 🔧 Backend Configuration

### Environment Variables

Create `.env` file in backend directory:

```env
# Database
DATABASE_URL=postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require

# Redis
REDIS_URL=redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com:10232

# Server
PORT=8080
ENV=production

# JWT
JWT_SECRET=your_jwt_secret_key_here
JWT_EXPIRY=24h

# CORS
CORS_ORIGINS=https://afftok-admin-prod-production.up.railway.app

# Logging
LOG_LEVEL=info
```

### Railway Configuration

In Railway dashboard, set these environment variables:

```
DATABASE_URL=postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require
REDIS_URL=redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com:10232
PORT=8080
ENV=production
JWT_SECRET=your_jwt_secret_key_here
```

### Build Configuration

**Procfile:**
```
web: go run main.go
```

**Go Version:**
```
1.19+
```

---

## 🎨 Admin Panel Configuration

### Environment Variables

Create `.env.local` file in admin directory:

```env
# API Configuration
VITE_API_URL=https://afftok-backend-prod-production.up.railway.app
VITE_API_PREFIX=/api

# OAuth
OAUTH_SERVER_URL=https://afftok-backend-prod-production.up.railway.app
OAUTH_CLIENT_ID=your_oauth_client_id
OAUTH_CLIENT_SECRET=your_oauth_client_secret

# Database (for admin panel)
DATABASE_URL=postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require

# Server
PORT=3000
NODE_ENV=production
```

### Railway Configuration

```
VITE_API_URL=https://afftok-backend-prod-production.up.railway.app
OAUTH_SERVER_URL=https://afftok-backend-prod-production.up.railway.app
PORT=3000
NODE_ENV=production
```

### Build Configuration

**package.json scripts:**
```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "start": "node server.js"
  }
}
```

**Node Version:**
```
18+
```

---

## 📱 Mobile App Configuration

### API Configuration

Edit `lib/config/api_config.dart`:

```dart
class ApiConfig {
  // Production API
  static const String baseUrl = 'https://afftok-backend-prod-production.up.railway.app';
  
  // API prefix
  static const String apiPrefix = '/api';
  
  // Endpoints
  static const String healthEndpoint = '/health';
  static const String usersEndpoint = '/users';
  static const String offersEndpoint = '/offers';
  static const String authEndpoint = '/auth';
  
  // Timeouts
  static const Duration connectionTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);
}
```

### Build Configuration

**pubspec.yaml:**
```yaml
name: afftok
description: AffTok Mobile Application
version: 1.0.0+1

environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  flutter:
    sdk: flutter
  http: ^1.1.0
  provider: ^6.0.0
  shared_preferences: ^2.0.0

dev_dependencies:
  flutter_test:
    sdk: flutter
```

### Android Configuration

**android/app/build.gradle:**
```gradle
android {
    compileSdkVersion 33
    
    defaultConfig {
        applicationId "com.afftok.mobile"
        minSdkVersion 21
        targetSdkVersion 33
        versionCode 1
        versionName "1.0.0"
    }
}
```

### iOS Configuration

**ios/Podfile:**
```ruby
platform :ios, '11.0'

target 'Runner' do
  flutter_root = File.expand_path(File.join(packages_path, 'flutter'))
  load File.join(flutter_root, 'packages', 'flutter_tools', 'bin', 'podhelper')

  flutter_ios_podfile_setup
end
```

---

## 🗄️ Database Configuration

### PostgreSQL (Neon)

**Connection Details:**
```
Host: ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech
Port: 5432
Database: neondb
User: neondb_owner
Password: npg_fuzC0cUrBLA5
SSL Mode: require
```

**Connection String:**
```
postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require
```

**Connect via psql:**
```bash
psql postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require
```

### Redis (RedisLabs)

**Connection Details:**
```
Host: redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com
Port: 10232
```

**Connection String:**
```
redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com:10232
```

**Connect via redis-cli:**
```bash
redis-cli -h redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com -p 10232
```

---

## 🔐 Environment Variables

### Backend

| Variable | Value | Required |
|----------|-------|----------|
| DATABASE_URL | PostgreSQL connection string | ✅ Yes |
| REDIS_URL | Redis connection string | ✅ Yes |
| PORT | 8080 | ✅ Yes |
| ENV | production | ✅ Yes |
| JWT_SECRET | Secret key for JWT | ✅ Yes |
| CORS_ORIGINS | Admin panel URL | ✅ Yes |

### Admin Panel

| Variable | Value | Required |
|----------|-------|----------|
| VITE_API_URL | Backend API URL | ✅ Yes |
| PORT | 3000 | ✅ Yes |
| NODE_ENV | production | ✅ Yes |
| OAUTH_SERVER_URL | Backend OAuth URL | ⚠️ Optional |

### Mobile App

| Variable | Value | Required |
|----------|-------|----------|
| API_BASE_URL | Backend API URL | ✅ Yes |
| API_PREFIX | /api | ✅ Yes |

---

## ✅ Configuration Checklist

- [ ] Backend environment variables set
- [ ] Admin panel environment variables set
- [ ] Mobile app API configuration updated
- [ ] Database connection verified
- [ ] Redis connection verified
- [ ] CORS configured
- [ ] JWT secret configured
- [ ] All endpoints accessible

---

**Configuration Version:** 1.0.0  
**Last Updated:** November 28, 2025
