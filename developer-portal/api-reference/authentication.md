# Authentication

AffTok supports multiple authentication methods depending on your use case.

## Authentication Methods

| Method | Use Case | Header |
|--------|----------|--------|
| JWT Token | Dashboard, user-specific operations | `Authorization: Bearer <token>` |
| API Key | Server-to-server, SDK | `X-API-Key: <key>` |
| Signed Request | Postbacks, high-security operations | HMAC signature in body |

## JWT Authentication

Used for user-authenticated requests (dashboard, mobile apps).

### Obtain Token

```bash
curl -X POST https://api.afftok.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "your_password"
  }'
```

**Response:**

```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "user": {
      "id": "usr_abc123",
      "email": "user@example.com",
      "name": "John Doe"
    }
  }
}
```

### Use Token

```bash
curl -X GET https://api.afftok.com/api/offers \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Refresh Token

```bash
curl -X POST https://api.afftok.com/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### Token Expiration

| Token Type | Expiration |
|------------|------------|
| Access Token | 1 hour |
| Refresh Token | 7 days |

## API Key Authentication

Used for server-to-server communication and SDKs.

### API Key Format

```
afftok_{environment}_{type}_{random}

Examples:
afftok_live_sk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6  (Live secret key)
afftok_test_sk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6  (Test secret key)
afftok_live_pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6  (Live public key)
```

### Use API Key

**Header Method (Recommended):**

```bash
curl -X POST https://api.afftok.com/api/sdk/click \
  -H "Content-Type: application/json" \
  -H "X-API-Key: afftok_live_sk_xxxxx" \
  -d '{...}'
```

**Body Method:**

```bash
curl -X POST https://api.afftok.com/api/postback \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "afftok_live_sk_xxxxx",
    ...
  }'
```

### API Key Permissions

| Permission | Description |
|------------|-------------|
| `clicks:write` | Track clicks |
| `conversions:write` | Track conversions |
| `stats:read` | Read statistics |
| `offers:read` | List offers |
| `webhooks:manage` | Manage webhooks |

### Generate API Key (Admin)

```bash
curl -X POST https://api.afftok.com/api/admin/advertisers/adv_123/api-keys \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Server",
    "permissions": ["clicks:write", "conversions:write", "stats:read"],
    "allowed_ips": ["203.0.113.0/24"],
    "rate_limit_per_minute": 120
  }'
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "key_abc123",
    "name": "Production Server",
    "api_key": "afftok_live_sk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
    "key_hint": "...o5p6",
    "permissions": ["clicks:write", "conversions:write", "stats:read"],
    "created_at": "2024-01-15T10:00:00Z"
  },
  "warning": "This is the only time the full API key will be shown. Store it securely."
}
```

## Signed Requests

For high-security operations, requests must include an HMAC-SHA256 signature.

### Signature Generation

```
data_to_sign = "{api_key}|{advertiser_id}|{timestamp}|{nonce}"
signature = HMAC-SHA256(api_key, data_to_sign)
```

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `api_key` | string | Your API key |
| `advertiser_id` | string | Your advertiser ID |
| `timestamp` | integer | Unix timestamp in milliseconds |
| `nonce` | string | Random 32-character string |
| `signature` | string | HMAC-SHA256 signature |

### Example Request

```bash
curl -X POST https://api.afftok.com/api/postback \
  -H "Content-Type: application/json" \
  -H "X-API-Key: afftok_live_sk_xxxxx" \
  -d '{
    "api_key": "afftok_live_sk_xxxxx",
    "advertiser_id": "adv_123456",
    "offer_id": "off_abc123",
    "transaction_id": "txn_xyz789",
    "amount": 49.99,
    "status": "approved",
    "timestamp": 1699876543210,
    "nonce": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
    "signature": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  }'
```

### Signature Implementation

**JavaScript:**

```javascript
const crypto = require('crypto');

function signRequest(apiKey, advertiserId) {
  const timestamp = Date.now();
  const nonce = crypto.randomBytes(16).toString('hex');
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  const signature = crypto.createHmac('sha256', apiKey).update(dataToSign).digest('hex');
  
  return { timestamp, nonce, signature };
}
```

**Python:**

```python
import hmac
import hashlib
import time
import secrets

def sign_request(api_key: str, advertiser_id: str) -> dict:
    timestamp = int(time.time() * 1000)
    nonce = secrets.token_hex(16)
    data_to_sign = f"{api_key}|{advertiser_id}|{timestamp}|{nonce}"
    signature = hmac.new(api_key.encode(), data_to_sign.encode(), hashlib.sha256).hexdigest()
    
    return {"timestamp": timestamp, "nonce": nonce, "signature": signature}
```

**PHP:**

```php
function signRequest($apiKey, $advertiserId) {
    $timestamp = round(microtime(true) * 1000);
    $nonce = bin2hex(random_bytes(16));
    $dataToSign = "{$apiKey}|{$advertiserId}|{$timestamp}|{$nonce}";
    $signature = hash_hmac('sha256', $dataToSign, $apiKey);
    
    return ['timestamp' => $timestamp, 'nonce' => $nonce, 'signature' => $signature];
}
```

**Go:**

```go
import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

func signRequest(apiKey, advertiserId string) (int64, string, string) {
    timestamp := time.Now().UnixMilli()
    
    nonceBytes := make([]byte, 16)
    rand.Read(nonceBytes)
    nonce := hex.EncodeToString(nonceBytes)
    
    dataToSign := fmt.Sprintf("%s|%s|%d|%s", apiKey, advertiserId, timestamp, nonce)
    h := hmac.New(sha256.New, []byte(apiKey))
    h.Write([]byte(dataToSign))
    signature := hex.EncodeToString(h.Sum(nil))
    
    return timestamp, nonce, signature
}
```

## Multi-Tenant Authentication

For multi-tenant setups, include tenant context:

### Via Header

```bash
curl -X GET https://api.afftok.com/api/offers \
  -H "Authorization: Bearer JWT_TOKEN" \
  -H "X-Tenant-ID: tenant_abc123"
```

### Via Subdomain

```bash
curl -X GET https://tenant-slug.api.afftok.com/api/offers \
  -H "Authorization: Bearer JWT_TOKEN"
```

### Via API Key

API keys are automatically bound to a tenant:

```bash
curl -X POST https://api.afftok.com/api/sdk/click \
  -H "X-API-Key: afftok_live_sk_xxxxx"  # Tenant resolved from key
```

## Rate Limits

| Authentication | Limit | Window |
|----------------|-------|--------|
| JWT Token | 1000 req | per minute |
| API Key (default) | 60 req | per minute |
| API Key (custom) | Configurable | per minute |
| Per IP | 100 req | per minute |

### Rate Limit Headers

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1699876600
```

### Rate Limit Response

```json
{
  "success": false,
  "error": "Rate limit exceeded",
  "code": "RATE_LIMITED",
  "details": {
    "limit": 60,
    "window": "minute",
    "retry_after": 45
  }
}
```

## Security Best Practices

1. **Never expose secret keys** - Keep API keys server-side only
2. **Use HTTPS** - All requests must use HTTPS
3. **Rotate keys** - Periodically rotate API keys
4. **Limit permissions** - Grant only necessary permissions
5. **IP allowlisting** - Restrict API keys to known IPs
6. **Monitor usage** - Watch for unusual activity

## Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 401 | `INVALID_API_KEY` | API key not found or invalid |
| 401 | `EXPIRED_TOKEN` | JWT token has expired |
| 401 | `REVOKED_API_KEY` | API key has been revoked |
| 403 | `INVALID_SIGNATURE` | HMAC signature verification failed |
| 403 | `EXPIRED_REQUEST` | Request timestamp too old |
| 403 | `IP_NOT_ALLOWED` | IP not in allowlist |
| 429 | `RATE_LIMITED` | Too many requests |

## Next Steps

- [Clicks API](clicks.md) - Track click events
- [Postbacks API](postbacks.md) - Send conversion postbacks
- [Security Guide](../security/api-keys.md) - API key best practices

