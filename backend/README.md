# AffTok - Complete Production Deployment

**Project Status:** ✅ Production Ready  
**Last Updated:** November 28, 2025  
**Version:** 1.0.0

---

## 📋 Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Components](#components)
4. [Deployment URLs](#deployment-urls)
5. [Database & Cache](#database--cache)
6. [GitHub Repositories](#github-repositories)
7. [Getting Started](#getting-started)
8. [Configuration](#configuration)
9. [Troubleshooting](#troubleshooting)
10. [Support](#support)

---

## 🎯 Project Overview

**AffTok** is a complete production-ready application consisting of three main components:

- **Backend API** (Go/Golang) - REST API with PostgreSQL and Redis
- **Admin Panel** (Web) - React/Vue-based admin dashboard
- **Mobile App** (Flutter) - Cross-platform mobile application

All components are deployed on **Railway** and version-controlled on **GitHub**.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    AffTok Production                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Mobile App  │  │ Admin Panel  │  │   Backend    │      │
│  │  (Flutter)   │  │   (Web)      │  │   (Go)       │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                 │                 │               │
│         └─────────────────┼─────────────────┘               │
│                           │                                 │
│                  ┌────────▼────────┐                        │
│                  │  Railway        │                        │
│                  │  (Deployment)   │                        │
│                  └────────┬────────┘                        │
│                           │                                 │
│         ┌─────────────────┼─────────────────┐               │
│         │                 │                 │               │
│    ┌────▼────┐      ┌─────▼─────┐    ┌────▼────┐          │
│    │PostgreSQL│      │   Redis   │    │  Neon   │          │
│    │  (Neon)  │      │  (Cloud)  │    │Database │          │
│    └──────────┘      └───────────┘    └─────────┘          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 📦 Components

### 1. Backend (Go)

**Location:** `./backend/`  
**Status:** ✅ Deployed on Railway  
**URL:** `https://afftok-backend-prod-production.up.railway.app`  
**Repository:** `https://github.com/aljapah/afftok-backend-prod`

**Features:**
- REST API endpoints
- PostgreSQL database integration
- Redis caching
- Authentication & Authorization
- Auto-deployment from GitHub

**Technology Stack:**
- Language: Go (Golang)
- Framework: Gin/Echo
- Database: PostgreSQL (Neon)
- Cache: Redis (RedisLabs)
- Deployment: Railway

---

### 2. Admin Panel (Web)

**Location:** `./admin/`  
**Status:** ✅ Deployed on Railway  
**URL:** `https://afftok-admin-prod-production.up.railway.app`  
**Repository:** `https://github.com/aljapah/afftok-admin-prod`

**Features:**
- Dashboard for content management
- User management
- Analytics & reporting
- OAuth integration
- Connected to Backend API

**Technology Stack:**
- Framework: React/Vue.js
- Build Tool: Vite
- Styling: Tailwind CSS
- Deployment: Railway (Node.js)

---

### 3. Mobile App (Flutter)

**Location:** `./mobile/`  
**Status:** ✅ Ready for Build  
**Repository:** `https://github.com/aljapah/afftok-mobile-prod`

**Features:**
- Cross-platform (iOS & Android)
- Connected to production Backend API
- User authentication
- Real-time data sync
- Offline support

**Technology Stack:**
- Framework: Flutter
- Language: Dart
- API Connection: REST
- Backend URL: `https://afftok-backend-prod-production.up.railway.app`

---

## 🌐 Deployment URLs

| Component | URL | Status |
|-----------|-----|--------|
| Backend API | `https://afftok-backend-prod-production.up.railway.app` | ✅ Live |
| Admin Panel | `https://afftok-admin-prod-production.up.railway.app` | ✅ Live |
| Mobile App | GitHub Repository | ✅ Ready |

---

## 🗄️ Database & Cache

### PostgreSQL (Neon)

```
Connection String:
postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require

Host: ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech
Port: 5432
Database: neondb
User: neondb_owner
Password: npg_fuzC0cUrBLA5
SSL Mode: require
```

### Redis (RedisLabs)

```
Connection String:
redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com:10232

Host: redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com
Port: 10232
```

---

## 🔗 GitHub Repositories

All repositories are version-controlled and auto-deploy from GitHub to Railway.

| Repository | URL | Branch |
|------------|-----|--------|
| Backend | `https://github.com/aljapah/afftok-backend-prod` | main |
| Admin Panel | `https://github.com/aljapah/afftok-admin-prod` | main |
| Mobile App | `https://github.com/aljapah/afftok-mobile-prod` | main |

---

## 🚀 Getting Started

### Prerequisites

- Git
- GitHub account (aljapah)
- Railway account
- Flutter SDK (for mobile development)
- Go 1.19+ (for backend development)
- Node.js 18+ (for admin panel development)

### Quick Start

#### 1. Clone Repositories

```bash
# Backend
git clone https://github.com/aljapah/afftok-backend-prod.git
cd afftok-backend-prod

# Admin Panel
git clone https://github.com/aljapah/afftok-admin-prod.git
cd afftok-admin-prod

# Mobile App
git clone https://github.com/aljapah/afftok-mobile-prod.git
cd afftok-mobile-prod
```

#### 2. Backend Setup

```bash
cd backend
go mod download
go run main.go
```

#### 3. Admin Panel Setup

```bash
cd admin
npm install
npm run dev
```

#### 4. Mobile App Setup

```bash
cd mobile
flutter pub get
flutter run
```

---

## ⚙️ Configuration

### Backend Configuration

**Environment Variables:**
```
DATABASE_URL=postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require
REDIS_URL=redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com:10232
PORT=8080
```

### Admin Panel Configuration

**Environment Variables:**
```
VITE_API_URL=https://afftok-backend-prod-production.up.railway.app
OAUTH_SERVER_URL=https://afftok-backend-prod-production.up.railway.app
DATABASE_URL=postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require
PORT=3000
```

### Mobile App Configuration

**API Configuration:**
```dart
// lib/config/api_config.dart
class ApiConfig {
  static const String baseUrl = 'https://afftok-backend-prod-production.up.railway.app';
  static const String apiPrefix = '/api';
}
```

---

## 🔧 Troubleshooting

### Backend Issues

**Problem:** Cannot connect to database
```
Solution: Check DATABASE_URL environment variable
Verify PostgreSQL credentials and connection string
Ensure SSL mode is set to 'require'
```

**Problem:** Redis connection failed
```
Solution: Verify REDIS_URL environment variable
Check Redis host and port
Ensure firewall allows connection
```

### Admin Panel Issues

**Problem:** API not responding
```
Solution: Verify VITE_API_URL is correct
Check backend is running
Verify CORS settings in backend
```

### Mobile App Issues

**Problem:** Cannot connect to backend
```
Solution: Verify API endpoint in lib/config/api_config.dart
Check backend is running and accessible
Verify network connectivity
```

---

## 📱 Building Mobile App

### Android (APK)

```bash
cd mobile
flutter build apk --release
# Output: build/app/outputs/flutter-app.apk
```

### iOS (IPA)

```bash
cd mobile
flutter build ios --release
# Output: build/ios/iphoneos/Runner.app
```

---

## 📊 Monitoring & Logs

### Railway Dashboard

- Backend: https://railway.app/project/...
- Admin Panel: https://railway.app/project/...

### View Logs

```bash
# Backend logs
railway logs -s backend

# Admin logs
railway logs -s admin
```

---

## 🔐 Security Notes

⚠️ **Important:**
- Never commit `.env` files
- Rotate credentials regularly
- Use strong passwords
- Enable 2FA on GitHub
- Keep dependencies updated

---

## 📞 Support

**Issues & Questions:**
- GitHub Issues: Use repository issue tracker
- Email: aljapah@gmail.com
- Documentation: See `./documentation/` folder

---

## 📝 License

All rights reserved. AffTok © 2025

---

## ✅ Deployment Checklist

- [x] Backend deployed on Railway
- [x] Admin Panel deployed on Railway
- [x] Mobile App code pushed to GitHub
- [x] Database migrations completed
- [x] Redis cache configured
- [x] Environment variables set
- [x] Auto-deployment configured
- [x] All repositories on GitHub
- [x] Documentation complete

---

**Last Updated:** November 28, 2025  
**Status:** Production Ready ✅
