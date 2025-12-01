# AffTok Integration - Deployment Guide

## Overview
This guide covers deploying the new promoter landing page feature to production.

## Changes Summary

### Backend Changes
1. New endpoint: `GET /api/promoter/user/:username`
2. Static file serving for landing pages
3. Files modified:
   - `internal/handlers/promoter.go` - New handler function
   - `cmd/api/main.go` - New route and static serving

### Frontend Changes
1. Updated Flutter app profile button URL
2. File modified:
   - `lib/screens/profile_screen_enhanced.dart`

### New Files
1. `public/promoter_landing.html` - Main landing page
2. `public/privacy.html` - Privacy policy
3. `public/terms.html` - Terms of service

## Pre-Deployment Checklist

### Code Review
- [ ] Review all changes in promoter.go
- [ ] Review URL change in profile_screen_enhanced.dart
- [ ] Verify no breaking changes to existing endpoints
- [ ] Check for any hardcoded values that need updating

### Testing
- [ ] Run unit tests for promoter handler
- [ ] Test new endpoint with valid/invalid usernames
- [ ] Test backward compatibility with UUID endpoint
- [ ] Test static file serving
- [ ] Test on multiple devices (mobile, tablet, desktop)

### Database
- [ ] Verify database schema is up to date
- [ ] Verify test data exists (users with offers)
- [ ] Check database backups are current

## Deployment Steps

### 1. Backend Deployment

#### Option A: Using Railway (Recommended)
```bash
cd /home/ubuntu/backend

# Ensure all changes are committed
git status

# Add all changes
git add .

# Commit with descriptive message
git commit -m "Add promoter landing page with username endpoint

- Add GetPromoterPageByUsername handler
- Add /api/promoter/user/:username endpoint
- Add static file serving for landing pages
- Update profile screen URL in Flutter app"

# Push to Railway
git push origin main
```

Railway will automatically:
1. Detect changes
2. Build the Go application
3. Deploy to production
4. Restart the service

#### Option B: Manual Deployment
```bash
cd /home/ubuntu/backend

# Build the application
go build -o server ./cmd/api/main.go

# Test locally (if possible)
./server

# Deploy binary to production server
# (Copy to your hosting provider)
```

### 2. Flutter App Deployment

#### iOS
```bash
cd /home/ubuntu/mobile

# Update version in pubspec.yaml
# Build for iOS
flutter build ios --release

# Upload to App Store
# (Use Xcode or fastlane)
```

#### Android
```bash
cd /home/ubuntu/mobile

# Update version in pubspec.yaml
# Build for Android
flutter build apk --release
flutter build appbundle --release

# Upload to Google Play Store
# (Use Play Console)
```

### 3. Post-Deployment Verification

#### Check Backend
```bash
# Test health endpoint
curl https://afftok-backend-prod-production.up.railway.app/health

# Test new endpoint
curl https://afftok-backend-prod-production.up.railway.app/api/promoter/user/testuser

# Test static files
curl https://afftok-backend-prod-production.up.railway.app/public/promoter_landing.html
```

#### Check Logs
```bash
# Railway logs
# (Check in Railway dashboard)

# Look for any errors related to:
# - Database connections
# - File serving
# - Handler execution
```

#### Test in App
1. Install updated Flutter app
2. Login with test account
3. Navigate to Profile
4. Click "صفحتي" button
5. Verify landing page loads correctly

## Rollback Plan

If issues occur after deployment:

### Backend Rollback
```bash
cd /home/ubuntu/backend

# Revert to previous commit
git revert HEAD

# Push to trigger redeploy
git push origin main
```

### Flutter Rollback
1. Revert to previous version in App Store/Play Store
2. Users will be prompted to update when new version is available

## Monitoring

### Key Metrics to Monitor
- [ ] API response times for `/api/promoter/user/:username`
- [ ] Error rates (4xx, 5xx responses)
- [ ] Static file serving performance
- [ ] Database query performance
- [ ] User engagement with landing pages

### Error Tracking
- [ ] Monitor error logs for new endpoint
- [ ] Check for SQL errors
- [ ] Verify CORS errors don't occur
- [ ] Monitor for file serving errors

### User Feedback
- [ ] Collect feedback on landing page design
- [ ] Monitor click-through rates on offers
- [ ] Track conversion rates
- [ ] Gather performance feedback

## Maintenance

### Regular Tasks
- [ ] Monitor landing page performance
- [ ] Update offers regularly
- [ ] Check for broken images/links
- [ ] Review user feedback
- [ ] Update privacy/terms if needed

### Scheduled Updates
- [ ] Weekly: Review error logs
- [ ] Monthly: Analyze user engagement metrics
- [ ] Quarterly: Update content and offers

## Support

### Common Issues

#### Landing page returns 404
- Verify user exists in database
- Check username spelling (case-sensitive)
- Verify database connection

#### Static files not serving
- Check `/public` directory exists
- Verify file permissions
- Check static route in main.go

#### Page loads but no offers
- Verify offers exist in database
- Check offers are marked as "active"
- Verify user_offers relationship

#### Bilingual toggle not working
- Check JavaScript is enabled
- Verify HTML file is complete
- Check browser console for errors

## Contact

For deployment issues:
- Email: support@afftokapp.com
- Slack: #deployment channel
- On-call: [On-call schedule]

---

**Last Updated**: December 1, 2025
**Version**: 1.0
**Status**: Ready for Deployment
