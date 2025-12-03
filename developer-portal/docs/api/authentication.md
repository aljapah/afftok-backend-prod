# Authentication API

Endpoints for user authentication and token management.

## Overview

AffTok uses JWT (JSON Web Tokens) for user authentication and API keys for server-to-server communication.

---

## POST /api/auth/register

Register a new user account.

### Request

```http
POST /api/auth/register
Content-Type: application/json
```

```json
{
  "email": "user@example.com",
  "password": "securePassword123!",
  "name": "John Doe",
  "phone": "+1234567890"
}
```

### Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | Valid email address |
| `password` | string | Yes | Min 8 characters |
| `name` | string | Yes | Full name |
| `phone` | string | No | Phone number |

### Response

```json
{
  "success": true,
  "message": "Registration successful",
  "data": {
    "user": {
      "id": "usr_a1b2c3d4e5f6",
      "email": "user@example.com",
      "name": "John Doe",
      "created_at": "2024-12-01T12:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### Errors

| Code | Description |
|------|-------------|
| `EMAIL_EXISTS` | Email already registered |
| `INVALID_EMAIL` | Invalid email format |
| `WEAK_PASSWORD` | Password doesn't meet requirements |

### Example

```bash
curl -X POST https://api.afftok.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123!",
    "name": "John Doe"
  }'
```

---

## POST /api/auth/login

Authenticate and receive JWT tokens.

### Request

```http
POST /api/auth/login
Content-Type: application/json
```

```json
{
  "email": "user@example.com",
  "password": "securePassword123!"
}
```

### Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | Registered email |
| `password` | string | Yes | Account password |

### Response

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "usr_a1b2c3d4e5f6",
      "email": "user@example.com",
      "name": "John Doe",
      "avatar": "https://cdn.afftok.com/avatars/usr_a1b2c3d4e5f6.jpg",
      "stats": {
        "total_clicks": 15234,
        "total_conversions": 456,
        "total_earnings": 2280.00,
        "monthly_clicks": 1523,
        "monthly_conversions": 45
      }
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  }
}
```

### Errors

| Code | Description |
|------|-------------|
| `INVALID_CREDENTIALS` | Email or password incorrect |
| `ACCOUNT_SUSPENDED` | Account has been suspended |
| `EMAIL_NOT_VERIFIED` | Email verification required |

### Example

```bash
curl -X POST https://api.afftok.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123!"
  }'
```

---

## POST /api/auth/refresh

Refresh an expired access token.

### Request

```http
POST /api/auth/refresh
Content-Type: application/json
```

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Response

```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  }
}
```

### Errors

| Code | Description |
|------|-------------|
| `INVALID_REFRESH_TOKEN` | Token is invalid or expired |
| `TOKEN_REVOKED` | Token has been revoked |

---

## GET /api/auth/me

Get current authenticated user profile.

### Request

```http
GET /api/auth/me
Authorization: Bearer <token>
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "usr_a1b2c3d4e5f6",
    "email": "user@example.com",
    "name": "John Doe",
    "phone": "+1234567890",
    "avatar": "https://cdn.afftok.com/avatars/usr_a1b2c3d4e5f6.jpg",
    "bio": "Digital marketer specializing in performance marketing",
    "social_links": {
      "twitter": "https://twitter.com/johndoe",
      "linkedin": "https://linkedin.com/in/johndoe"
    },
    "stats": {
      "total_clicks": 15234,
      "unique_clicks": 12456,
      "total_conversions": 456,
      "total_earnings": 2280.00,
      "monthly_clicks": 1523,
      "monthly_conversions": 45,
      "conversion_rate": 2.99
    },
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-12-01T12:00:00Z"
  }
}
```

### Example

```bash
curl -X GET https://api.afftok.com/api/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

---

## PUT /api/auth/profile

Update user profile.

### Request

```http
PUT /api/auth/profile
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "name": "John Doe Updated",
  "phone": "+1987654321",
  "bio": "Updated bio",
  "social_links": {
    "twitter": "https://twitter.com/johndoe",
    "instagram": "https://instagram.com/johndoe"
  }
}
```

### Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Full name |
| `phone` | string | No | Phone number |
| `bio` | string | No | User biography |
| `social_links` | object | No | Social media links |

### Response

```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "id": "usr_a1b2c3d4e5f6",
    "name": "John Doe Updated",
    "phone": "+1987654321",
    "bio": "Updated bio",
    "updated_at": "2024-12-01T12:30:00Z"
  }
}
```

---

## POST /api/auth/change-password

Change account password.

### Request

```http
POST /api/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "current_password": "oldPassword123!",
  "new_password": "newSecurePassword456!"
}
```

### Response

```json
{
  "success": true,
  "message": "Password changed successfully"
}
```

### Errors

| Code | Description |
|------|-------------|
| `INVALID_PASSWORD` | Current password incorrect |
| `WEAK_PASSWORD` | New password doesn't meet requirements |
| `SAME_PASSWORD` | New password same as current |

---

## POST /api/auth/logout

Invalidate current token.

### Request

```http
POST /api/auth/logout
Authorization: Bearer <token>
```

### Response

```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

---

## Token Structure

### Access Token (JWT)

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "usr_a1b2c3d4e5f6",
    "email": "user@example.com",
    "iat": 1701234567,
    "exp": 1701238167,
    "iss": "afftok.com"
  }
}
```

### Token Lifetimes

| Token Type | Lifetime |
|------------|----------|
| Access Token | 1 hour |
| Refresh Token | 30 days |

---

## Security Best Practices

1. **Store tokens securely** - Use secure storage (Keychain, encrypted SharedPreferences)
2. **Never expose tokens** - Don't log or transmit tokens insecurely
3. **Implement token refresh** - Refresh before expiration
4. **Handle logout properly** - Clear all stored tokens
5. **Use HTTPS only** - Never send tokens over HTTP

