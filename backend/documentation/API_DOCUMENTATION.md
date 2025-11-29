# AffTok Backend API Documentation

**Complete API Reference for AffTok Backend**

---

## 📋 Base URL

```
https://afftok-backend-prod-production.up.railway.app
```

---

## 🔐 Authentication

All endpoints require authentication via Bearer token.

**Header:**
```
Authorization: Bearer {token}
```

---

## 📚 API Endpoints

### Health Check

#### GET /health

Check if backend is running.

**Request:**
```bash
curl https://afftok-backend-prod-production.up.railway.app/health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-11-28T01:30:00Z"
}
```

---

### Users

#### GET /api/users

Get all users.

**Request:**
```bash
curl -H "Authorization: Bearer {token}" \
  https://afftok-backend-prod-production.up.railway.app/api/users
```

**Response:**
```json
{
  "data": [
    {
      "id": "user-001",
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2025-11-28T01:00:00Z"
    }
  ],
  "total": 1
}
```

#### POST /api/users

Create a new user.

**Request:**
```bash
curl -X POST \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "secure_password"
  }' \
  https://afftok-backend-prod-production.up.railway.app/api/users
```

**Response:**
```json
{
  "id": "user-002",
  "name": "Jane Doe",
  "email": "jane@example.com",
  "created_at": "2025-11-28T01:30:00Z"
}
```

#### GET /api/users/{id}

Get user by ID.

**Request:**
```bash
curl -H "Authorization: Bearer {token}" \
  https://afftok-backend-prod-production.up.railway.app/api/users/user-001
```

**Response:**
```json
{
  "id": "user-001",
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2025-11-28T01:00:00Z"
}
```

---

### Authentication

#### POST /api/auth/login

User login.

**Request:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "secure_password"
  }' \
  https://afftok-backend-prod-production.up.railway.app/api/auth/login
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user-001",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

#### POST /api/auth/logout

User logout.

**Request:**
```bash
curl -X POST \
  -H "Authorization: Bearer {token}" \
  https://afftok-backend-prod-production.up.railway.app/api/auth/logout
```

**Response:**
```json
{
  "message": "Logged out successfully"
}
```

---

### Content/Offers

#### GET /api/offers

Get all offers.

**Request:**
```bash
curl -H "Authorization: Bearer {token}" \
  https://afftok-backend-prod-production.up.railway.app/api/offers
```

**Response:**
```json
{
  "data": [
    {
      "id": "offer-001",
      "title": "Special Offer",
      "description": "Limited time offer",
      "price": 99.99,
      "created_at": "2025-11-28T01:00:00Z"
    }
  ],
  "total": 1
}
```

#### POST /api/offers

Create a new offer.

**Request:**
```bash
curl -X POST \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Offer",
    "description": "Amazing deal",
    "price": 149.99
  }' \
  https://afftok-backend-prod-production.up.railway.app/api/offers
```

**Response:**
```json
{
  "id": "offer-002",
  "title": "New Offer",
  "description": "Amazing deal",
  "price": 149.99,
  "created_at": "2025-11-28T01:30:00Z"
}
```

---

## 🔄 Response Format

All API responses follow this format:

**Success (200):**
```json
{
  "data": {...},
  "message": "Success"
}
```

**Error (400, 401, 500):**
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "timestamp": "2025-11-28T01:30:00Z"
}
```

---

## 📊 Status Codes

| Code | Meaning |
|------|---------|
| 200 | OK |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 500 | Internal Server Error |

---

## 🔒 Error Responses

### 401 Unauthorized

```json
{
  "error": "Invalid or missing token",
  "code": "UNAUTHORIZED"
}
```

### 404 Not Found

```json
{
  "error": "Resource not found",
  "code": "NOT_FOUND"
}
```

### 500 Internal Server Error

```json
{
  "error": "Internal server error",
  "code": "INTERNAL_ERROR"
}
```

---

## 📝 Rate Limiting

- **Limit:** 1000 requests per hour
- **Header:** `X-RateLimit-Remaining`

---

## 🧪 Testing

### Using Postman

1. Import collection from `./documentation/postman_collection.json`
2. Set environment variables:
   - `base_url`: https://afftok-backend-prod-production.up.railway.app
   - `token`: Your bearer token
3. Run requests

### Using cURL

```bash
# Set variables
BASE_URL="https://afftok-backend-prod-production.up.railway.app"
TOKEN="your_bearer_token"

# Test health
curl $BASE_URL/health

# Get users
curl -H "Authorization: Bearer $TOKEN" $BASE_URL/api/users
```

---

## 📞 Support

For API issues or questions:
- Check logs: `railway logs -s backend`
- Review documentation: `./documentation/`
- Contact: aljapah@gmail.com

---

**API Version:** 1.0.0  
**Last Updated:** November 28, 2025
