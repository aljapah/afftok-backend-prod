# AffTok Troubleshooting Guide

**Common issues and solutions**

---

## 🔧 Backend Issues

### Issue: Backend won't start

**Symptoms:**
- Railway deployment fails
- Error: "Cannot connect to database"

**Solutions:**
1. Check DATABASE_URL environment variable
2. Verify PostgreSQL credentials
3. Check database is accessible
4. Review logs: `railway logs -s backend`

**Commands:**
```bash
# Test database connection
psql $DATABASE_URL

# Check environment variables
railway variables -s backend

# Restart service
railway restart -s backend
```

---

### Issue: Redis connection failed

**Symptoms:**
- Cache not working
- Error: "ECONNREFUSED"

**Solutions:**
1. Verify REDIS_URL environment variable
2. Check Redis host and port
3. Test connection manually

**Commands:**
```bash
# Test Redis connection
redis-cli -h redis-10232.crce214.us-east-1-3.ec2.cloud.redislabs.com -p 10232 ping

# Check Redis logs
railway logs -s backend | grep redis
```

---

## 🎨 Admin Panel Issues

### Issue: Blank page on load

**Symptoms:**
- White screen
- Console errors

**Solutions:**
1. Check VITE_API_URL is correct
2. Verify backend is running
3. Check browser console for errors
4. Clear browser cache

**Commands:**
```bash
# Test API connection
curl https://afftok-backend-prod-production.up.railway.app/health

# Check admin logs
railway logs -s admin

# Rebuild
railway redeploy -s admin
```

---

### Issue: API requests failing

**Symptoms:**
- 401 Unauthorized
- CORS errors

**Solutions:**
1. Check authentication token
2. Verify CORS settings in backend
3. Check API endpoint URL

**Commands:**
```bash
# Test API with curl
curl -H "Authorization: Bearer TOKEN" \
  https://afftok-backend-prod-production.up.railway.app/api/users
```

---

## 📱 Mobile App Issues

### Issue: Cannot connect to backend

**Symptoms:**
- Network error
- Timeout

**Solutions:**
1. Verify API endpoint in `api_config.dart`
2. Check backend is running
3. Test with curl
4. Check device network

**Commands:**
```bash
# Test backend
curl https://afftok-backend-prod-production.up.railway.app/health

# Check Flutter logs
flutter logs
```

---

### Issue: Build fails

**Symptoms:**
- Compilation errors
- Missing dependencies

**Solutions:**
1. Clean build
2. Get dependencies
3. Check Flutter version

**Commands:**
```bash
flutter clean
flutter pub get
flutter doctor
flutter build apk --release
```

---

## 🗄️ Database Issues

### Issue: Migration failed

**Symptoms:**
- Database schema mismatch
- Query errors

**Solutions:**
1. Check migration files
2. Run migrations manually
3. Verify database permissions

**Commands:**
```bash
# Run migrations
cd backend
go run main.go migrate

# Check database
psql $DATABASE_URL -c "\dt"
```

---

### Issue: Connection pool exhausted

**Symptoms:**
- Slow queries
- Timeout errors

**Solutions:**
1. Increase connection pool size
2. Check for connection leaks
3. Monitor active connections

**Commands:**
```bash
# Check active connections
psql $DATABASE_URL -c "SELECT count(*) FROM pg_stat_activity;"
```

---

## 🚀 Deployment Issues

### Issue: Auto-deployment not working

**Symptoms:**
- Push to GitHub doesn't trigger deployment
- Webhook not firing

**Solutions:**
1. Check GitHub webhook settings
2. Verify Railway integration
3. Check webhook logs

**Steps:**
1. Go to GitHub repository settings
2. Click "Webhooks"
3. Verify Railway webhook is active
4. Check recent deliveries

---

### Issue: Build timeout

**Symptoms:**
- Deployment fails after 10 minutes
- Build process hangs

**Solutions:**
1. Optimize build process
2. Remove unnecessary files
3. Use .dockerignore

**Commands:**
```bash
# Add .dockerignore
echo "node_modules" >> .dockerignore
echo ".git" >> .dockerignore
echo "*.log" >> .dockerignore
```

---

## 📊 Performance Issues

### Issue: Slow API responses

**Symptoms:**
- High latency
- Timeout errors

**Solutions:**
1. Check database query performance
2. Verify Redis cache is working
3. Monitor server resources

**Commands:**
```bash
# Check slow queries
psql $DATABASE_URL -c "SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"

# Test cache
redis-cli -h HOST -p PORT GET test_key
```

---

### Issue: High memory usage

**Symptoms:**
- Out of memory errors
- Crashes

**Solutions:**
1. Check for memory leaks
2. Optimize queries
3. Increase server resources

**Commands:**
```bash
# Check memory usage
railway logs -s backend | grep memory

# Monitor resources
railway metrics -s backend
```

---

## 🔐 Security Issues

### Issue: Unauthorized access

**Symptoms:**
- 401 errors
- Token expired

**Solutions:**
1. Refresh authentication token
2. Check JWT expiry
3. Verify credentials

**Commands:**
```bash
# Test authentication
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  https://afftok-backend-prod-production.up.railway.app/api/auth/login
```

---

## 📞 Getting Help

If issues persist:

1. Check documentation in `./documentation/`
2. Review logs: `railway logs`
3. Check GitHub issues
4. Contact: aljapah@gmail.com

---

**Last Updated:** November 28, 2025
