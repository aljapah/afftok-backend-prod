# AffTok Deployment Guide

**Complete step-by-step guide for deploying AffTok to production**

---

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Backend Deployment](#backend-deployment)
3. [Admin Panel Deployment](#admin-panel-deployment)
4. [Mobile App Deployment](#mobile-app-deployment)
5. [Database Setup](#database-setup)
6. [Post-Deployment](#post-deployment)

---

## ✅ Prerequisites

Before deploying, ensure you have:

- [ ] GitHub account (aljapah)
- [ ] Railway account
- [ ] Neon PostgreSQL account
- [ ] RedisLabs account
- [ ] Git installed
- [ ] Go 1.19+ installed
- [ ] Node.js 18+ installed
- [ ] Flutter SDK installed

---

## 🔧 Backend Deployment

### Step 1: Push to GitHub

```bash
cd backend
git init
git config user.email "aljapah@gmail.com"
git config user.name "aljapah"
git add .
git commit -m "Initial backend commit"
git remote add origin https://github.com/aljapah/afftok-backend-prod.git
git branch -M main
git push -u origin main
```

### Step 2: Create Railway Project

1. Go to https://railway.app
2. Click "New Project"
3. Select "Deploy from GitHub"
4. Select repository: `afftok-backend-prod`
5. Click "Deploy"

### Step 3: Configure Environment Variables

In Railway dashboard, set:

```
DATABASE_URL=postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require
REDIS_URL=redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com:10232
PORT=8080
```

### Step 4: Run Migrations

```bash
railway run go run main.go migrate
```

### Step 5: Verify Deployment

```bash
curl https://afftok-backend-prod-production.up.railway.app/health
```

**Expected Response:** `{"status":"ok"}`

---

## 🎨 Admin Panel Deployment

### Step 1: Push to GitHub

```bash
cd admin
git init
git config user.email "aljapah@gmail.com"
git config user.name "aljapah"
git add .
git commit -m "Initial admin panel commit"
git remote add origin https://github.com/aljapah/afftok-admin-prod.git
git branch -M main
git push -u origin main
```

### Step 2: Create Railway Project

1. Go to https://railway.app
2. Click "New Project"
3. Select "Deploy from GitHub"
4. Select repository: `afftok-admin-prod`
5. Click "Deploy"

### Step 3: Configure Environment Variables

In Railway dashboard, set:

```
VITE_API_URL=https://afftok-backend-prod-production.up.railway.app
OAUTH_SERVER_URL=https://afftok-backend-prod-production.up.railway.app
DATABASE_URL=postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require
PORT=3000
```

### Step 4: Configure Build Settings

In Railway dashboard:

- **Build Command:** `npm run build`
- **Start Command:** `npm start`
- **Node Version:** 18

### Step 5: Verify Deployment

```bash
curl https://afftok-admin-prod-production.up.railway.app
```

---

## 📱 Mobile App Deployment

### Step 1: Push to GitHub

```bash
cd mobile
git init
git config user.email "aljapah@gmail.com"
git config user.name "aljapah"
git add .
git commit -m "Initial mobile app commit - connected to production backend"
git remote add origin https://github.com/aljapah/afftok-mobile-prod.git
git branch -M main
git push -u origin main
```

### Step 2: Update Configuration

Edit `lib/config/api_config.dart`:

```dart
class ApiConfig {
  static const String baseUrl = 'https://afftok-backend-prod-production.up.railway.app';
  static const String apiPrefix = '/api';
}
```

### Step 3: Build APK (Android)

```bash
cd mobile
flutter clean
flutter pub get
flutter build apk --release
```

**Output:** `build/app/outputs/flutter-app.apk`

### Step 4: Build IPA (iOS)

```bash
cd mobile
flutter clean
flutter pub get
flutter build ios --release
```

**Output:** `build/ios/iphoneos/Runner.app`

### Step 5: Upload to App Stores

- **Google Play:** Upload APK via Google Play Console
- **App Store:** Upload IPA via App Store Connect

---

## 🗄️ Database Setup

### PostgreSQL (Neon)

**Connection Details:**
```
Host: ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech
Port: 5432
Database: neondb
User: neondb_owner
Password: npg_fuzC0cUrBLA5
```

**Connect via psql:**
```bash
psql postgresql://neondb_owner:npg_fuzC0cUrBLA5@ep-divine-pond-ahcjjnmh-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require
```

**Run Migrations:**
```bash
cd backend
go run main.go migrate
```

### Redis (RedisLabs)

**Connection Details:**
```
Host: redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com
Port: 10232
```

**Test Connection:**
```bash
redis-cli -h redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com -p 10232 ping
```

---

## ✅ Post-Deployment

### 1. Verify All Components

```bash
# Backend
curl https://afftok-backend-prod-production.up.railway.app/health

# Admin Panel
curl https://afftok-admin-prod-production.up.railway.app

# Mobile App
# Test on device or emulator
```

### 2. Test API Endpoints

```bash
# Get health status
curl https://afftok-backend-prod-production.up.railway.app/api/health

# Get users (example)
curl https://afftok-backend-prod-production.up.railway.app/api/users
```

### 3. Check Logs

```bash
# Backend logs
railway logs -s afftok-backend-prod

# Admin logs
railway logs -s afftok-admin-prod
```

### 4. Monitor Performance

- Railway Dashboard: https://railway.app
- Check CPU, Memory, Network usage
- Set up alerts for errors

### 5. Enable Auto-Deployment

In GitHub repository settings:
- Go to "Webhooks"
- Verify Railway webhook is configured
- Auto-deploy on push to main branch

---

## 🔄 Continuous Deployment

### GitHub to Railway

1. Push to GitHub main branch
2. Railway automatically detects changes
3. Runs build process
4. Deploys to production
5. Monitors for errors

### Manual Deployment

```bash
# If needed, redeploy manually
railway deploy
```

---

## 🚨 Troubleshooting

### Backend Won't Start

```
Check logs: railway logs -s backend
Verify DATABASE_URL is correct
Verify REDIS_URL is correct
Check Go version compatibility
```

### Admin Panel Shows Blank Page

```
Check browser console for errors
Verify VITE_API_URL is correct
Check backend is running
Verify CORS settings
```

### Mobile App Can't Connect

```
Verify API endpoint in api_config.dart
Check backend is running
Test with curl: curl https://afftok-backend-prod-production.up.railway.app
Check network connectivity
```

---

## 📊 Deployment Checklist

- [ ] Backend deployed on Railway
- [ ] Admin Panel deployed on Railway
- [ ] Environment variables configured
- [ ] Database migrations completed
- [ ] Redis cache working
- [ ] All endpoints responding
- [ ] Mobile app built and tested
- [ ] Auto-deployment configured
- [ ] Monitoring set up
- [ ] Documentation updated

---

## 🔐 Security Checklist

- [ ] All credentials stored in environment variables
- [ ] No secrets in code
- [ ] HTTPS enabled
- [ ] Database backups configured
- [ ] Access logs enabled
- [ ] Rate limiting configured
- [ ] CORS properly configured
- [ ] Authentication working

---

**Deployment Status:** ✅ Complete  
**Last Updated:** November 28, 2025
