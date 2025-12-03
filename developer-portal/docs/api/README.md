# REST API Reference

Complete reference for the AffTok REST API.

## Base URL

| Environment | Base URL |
|-------------|----------|
| Production | `https://api.afftok.com` |
| Sandbox | `https://sandbox.api.afftok.com` |

## Authentication

AffTok supports two authentication methods:

### 1. JWT Token (User Authentication)

For user-specific endpoints (stats, offers, profile):

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 2. API Key (Server-to-Server)

For postbacks and server integrations:

```http
X-API-Key: afftok_live_sk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
```

## Request Format

All requests should:

- Use `Content-Type: application/json`
- Include appropriate authentication headers
- Send JSON-encoded request bodies for POST/PUT

## Response Format

All responses follow this structure:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": "Error description",
  "code": "ERROR_CODE",
  "details": {
    // Additional error context
  }
}
```

## Rate Limits

| Endpoint Type | Limit | Window |
|---------------|-------|--------|
| Standard | 60 requests | 1 minute |
| Postback | 120 requests | 1 minute |
| Admin | 30 requests | 1 minute |

Rate limit headers are included in all responses:

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1701234567
```

## API Sections

| Section | Description |
|---------|-------------|
| [Authentication](./authentication.md) | Login, register, token management |
| [Offers](./offers.md) | Offer listing and management |
| [Clicks](./clicks.md) | Click tracking endpoints |
| [Postbacks](./postbacks.md) | Conversion tracking |
| [Stats](./stats.md) | Analytics and reporting |
| [Geo Rules](./geo-rules.md) | Geographic restrictions |
| [Webhooks](./webhooks.md) | Webhook management |
| [Admin](./admin.md) | Administrative endpoints |

## Common Parameters

### Pagination

```http
GET /api/offers?page=1&limit=20
```

| Parameter | Type | Default | Max |
|-----------|------|---------|-----|
| `page` | integer | 1 | - |
| `limit` | integer | 20 | 100 |

### Filtering

```http
GET /api/clicks?start_date=2024-01-01&end_date=2024-01-31&status=approved
```

### Sorting

```http
GET /api/offers?sort=created_at&order=desc
```

## HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 429 | Rate Limited |
| 500 | Server Error |

## SDK Support

Official SDKs handle authentication and request formatting automatically:

- [JavaScript/Node.js](../sdks/web.md)
- [Python](../guides/python.md)
- [PHP](../guides/php.md)
- [Go](../guides/go.md)

---

## Quick Examples

### cURL

```bash
curl -X GET "https://api.afftok.com/api/offers" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### JavaScript

```javascript
const response = await fetch('https://api.afftok.com/api/offers', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});
const data = await response.json();
```

### Python

```python
import requests

response = requests.get(
    'https://api.afftok.com/api/offers',
    headers={'Authorization': f'Bearer {token}'}
)
data = response.json()
```

### PHP

```php
$ch = curl_init('https://api.afftok.com/api/offers');
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Authorization: Bearer ' . $token
]);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
$data = json_decode($response, true);
```

### Go

```go
req, _ := http.NewRequest("GET", "https://api.afftok.com/api/offers", nil)
req.Header.Set("Authorization", "Bearer "+token)
client := &http.Client{}
resp, _ := client.Do(req)
```

