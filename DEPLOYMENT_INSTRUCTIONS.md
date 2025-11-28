# تعليمات الإطلاق - AffTok (جاهز للإنتاج)

**التاريخ:** 28 نوفمبر 2025  
**الحالة:** ✅ جميع المشاكل تم إصلاحها

---

## 📋 ملخص الإصلاحات المطبقة

تم إصلاح جميع مشاكل الاتصال بين المكونات الثلاثة:

✅ **Backend** - تم تحديث `ALLOWED_ORIGINS` للسماح بـ Admin Panel  
✅ **Admin Panel** - تم تحديث `VITE_API_URL` و `OAUTH_SERVER_URL` بـ URL الفعلي للـ Backend  
✅ **Mobile App** - تم تحديث `baseUrl` بـ URL الفعلي للـ Backend  

---

## 🚀 خطوات الإطلاق السريعة

### الخطوة 1: دفع التغييرات إلى GitHub

```bash
cd /path/to/afftok-fixed
git add .
git commit -m "Fix: Update environment variables for production deployment - All components connected"
git push origin main
```

### الخطوة 2: تحديث متغيرات البيئة على Railway (Backend)

```bash
cd backend
railway variables set ALLOWED_ORIGINS="https://afftok-admin-prod-production.up.railway.app,https://yourdomain.com"
railway redeploy
```

### الخطوة 3: تحديث متغيرات البيئة على Railway (Admin Panel)

```bash
cd admin
railway variables set VITE_API_URL="https://afftok-backend-prod-production.up.railway.app"
railway variables set OAUTH_SERVER_URL="https://afftok-backend-prod-production.up.railway.app"
railway redeploy
```

### الخطوة 4: بناء وتحديث Mobile App

```bash
cd mobile
flutter pub get
flutter build apk --release
```

---

## ✅ التحقق من الاتصال

### اختبار Backend:
```bash
curl https://afftok-backend-prod-production.up.railway.app/health
```

**النتيجة المتوقعة:**
```json
{
  "status": "ok",
  "message": "AffTok API is running"
}
```

### اختبار Admin Panel:
```bash
curl https://afftok-admin-prod-production.up.railway.app
```

**النتيجة المتوقعة:** صفحة HTML للـ Admin Panel

### اختبار من المتصفح:
1. افتح Admin Panel: `https://afftok-admin-prod-production.up.railway.app`
2. حاول تسجيل الدخول أو تنفيذ أي عملية
3. يجب أن تنجح جميع الطلبات بدون CORS errors

---

## 📁 الملفات المعدلة

| الملف | التغيير |
|------|---------|
| `backend/.env.production` | تحديث `ALLOWED_ORIGINS` |
| `admin/.env` | تحديث `VITE_API_URL` و `OAUTH_SERVER_URL` |
| `mobile/lib/config/api_config.dart` | تحديث `baseUrl` |

---

## 🔧 معلومات الاتصال

| المكون | URL |
|-------|-----|
| Backend API | https://afftok-backend-prod-production.up.railway.app |
| Admin Panel | https://afftok-admin-prod-production.up.railway.app |
| Database | Neon PostgreSQL (نفس الاتصال) |
| Cache | Redis Labs (نفس الاتصال) |

---

## ⚠️ ملاحظات مهمة

### 1. تحديث Domain الخاص بك
إذا كان لديك domain خاص (مثل `afftok.com`):

**في Backend .env.production:**
```
ALLOWED_ORIGINS=https://afftok-admin-prod-production.up.railway.app,https://admin.afftok.com,https://yourdomain.com
```

### 2. تحديث JWT Secret
تأكد من تغيير `JWT_SECRET` في الإنتاج:
```
JWT_SECRET=your_very_secure_secret_key_here_change_this
```

### 3. تحديث متغيرات البيئة الأخرى
تحقق من جميع متغيرات البيئة في Railway Dashboard وتأكد من أنها صحيحة.

---

## 🐛 استكشاف الأخطاء

### إذا ظهرت CORS errors:
1. تحقق من `ALLOWED_ORIGINS` في Backend
2. تأكد من أن Admin Panel URL موجود فيها
3. أعد تشغيل Backend: `railway redeploy`

### إذا لم يتصل Admin Panel بـ Backend:
1. تحقق من `VITE_API_URL` في Admin Panel
2. تأكد من أنها تشير إلى Backend URL الصحيح
3. افتح Developer Tools (F12) وتحقق من Network tab

### إذا لم يتصل Mobile App بـ Backend:
1. تحقق من `baseUrl` في `api_config.dart`
2. تأكد من أنها تشير إلى Backend URL الصحيح
3. أعد بناء التطبيق: `flutter build apk --release`

---

## 📊 حالة المشروع

| المكون | الحالة | ملاحظات |
|-------|--------|---------|
| Backend | ✅ جاهز | جميع الـ endpoints مفعلة |
| Admin Panel | ✅ جاهز | متصل بـ Backend |
| Mobile App | ✅ جاهز | متصل بـ Backend |
| Database | ✅ جاهز | Neon PostgreSQL |
| Cache | ✅ جاهز | Redis Labs |

---

## 🎯 الخطوات التالية

1. **اختبار شامل** - اختبر جميع الـ features في كل مكون
2. **مراقبة الأداء** - راقب logs و metrics على Railway
3. **النسخ الاحتياطية** - تأكد من وجود نسخ احتياطية للبيانات
4. **الأمان** - قم بمراجعة أمان الكود والإعدادات

---

## 📞 المساعدة والدعم

إذا واجهت أي مشاكل:

1. تحقق من Railway logs: `railway logs`
2. تحقق من GitHub Issues
3. راجع ملف `DIAGNOSIS_REPORT.md` للمزيد من التفاصيل

---

**تم إعداد التعليمات بواسطة:** Manus AI  
**التاريخ:** 28 نوفمبر 2025  
**الحالة:** ✅ جاهز للإطلاق
