# Tracking Links

AffTok uses cryptographically signed tracking links to ensure security, prevent tampering, and enable replay protection.

## Link Formats

### Standard Tracking Link

```
https://track.afftok.com/c/{tracking_code}
```

Example:
```
https://track.afftok.com/c/abc123def456
```

### Signed Tracking Link (Recommended)

```
https://track.afftok.com/c/{tracking_code}.{timestamp}.{nonce}.{signature}
```

Example:
```
https://track.afftok.com/c/abc123.1699876543210.nonce123abc.sig456def
```

### Short Link

```
https://afftok.link/{short_code}
```

Example:
```
https://afftok.link/xYz123
```

## Signed Link Components

| Component | Description | Example |
|-----------|-------------|---------|
| `tracking_code` | Unique identifier for user-offer | `abc123def456` |
| `timestamp` | Unix timestamp in milliseconds | `1699876543210` |
| `nonce` | Random string (16-32 chars) | `nonce123abc456def` |
| `signature` | HMAC-SHA256 signature (first 16 chars) | `a1b2c3d4e5f6g7h8` |

## Signature Generation

The signature is generated using HMAC-SHA256:

```
data_to_sign = "{tracking_code}.{timestamp}.{nonce}"
signature = HMAC-SHA256(secret_key, data_to_sign)[0:16]
```

### JavaScript Example

```javascript
const crypto = require('crypto');

function generateSignedLink(trackingCode, secretKey) {
  const timestamp = Date.now();
  const nonce = crypto.randomBytes(16).toString('hex');
  const dataToSign = `${trackingCode}.${timestamp}.${nonce}`;
  
  const signature = crypto
    .createHmac('sha256', secretKey)
    .update(dataToSign)
    .digest('hex')
    .substring(0, 16);
  
  return `https://track.afftok.com/c/${trackingCode}.${timestamp}.${nonce}.${signature}`;
}

// Usage
const link = generateSignedLink('abc123', 'your_secret_key');
console.log(link);
// https://track.afftok.com/c/abc123.1699876543210.a1b2c3d4e5f6g7h8.sig123
```

### Python Example

```python
import hmac
import hashlib
import time
import secrets

def generate_signed_link(tracking_code: str, secret_key: str) -> str:
    timestamp = int(time.time() * 1000)
    nonce = secrets.token_hex(16)
    data_to_sign = f"{tracking_code}.{timestamp}.{nonce}"
    
    signature = hmac.new(
        secret_key.encode(),
        data_to_sign.encode(),
        hashlib.sha256
    ).hexdigest()[:16]
    
    return f"https://track.afftok.com/c/{tracking_code}.{timestamp}.{nonce}.{signature}"

# Usage
link = generate_signed_link('abc123', 'your_secret_key')
print(link)
```

### PHP Example

```php
function generateSignedLink($trackingCode, $secretKey) {
    $timestamp = round(microtime(true) * 1000);
    $nonce = bin2hex(random_bytes(16));
    $dataToSign = "{$trackingCode}.{$timestamp}.{$nonce}";
    
    $signature = substr(hash_hmac('sha256', $dataToSign, $secretKey), 0, 16);
    
    return "https://track.afftok.com/c/{$trackingCode}.{$timestamp}.{$nonce}.{$signature}";
}

// Usage
$link = generateSignedLink('abc123', 'your_secret_key');
echo $link;
```

### Go Example

```go
package main

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

func generateSignedLink(trackingCode, secretKey string) string {
    timestamp := time.Now().UnixMilli()
    
    nonceBytes := make([]byte, 16)
    rand.Read(nonceBytes)
    nonce := hex.EncodeToString(nonceBytes)
    
    dataToSign := fmt.Sprintf("%s.%d.%s", trackingCode, timestamp, nonce)
    
    h := hmac.New(sha256.New, []byte(secretKey))
    h.Write([]byte(dataToSign))
    signature := hex.EncodeToString(h.Sum(nil))[:16]
    
    return fmt.Sprintf("https://track.afftok.com/c/%s.%d.%s.%s",
        trackingCode, timestamp, nonce, signature)
}

func main() {
    link := generateSignedLink("abc123", "your_secret_key")
    fmt.Println(link)
}
```

## TTL (Time-To-Live)

Signed links have a configurable TTL to prevent abuse:

| Setting | Default | Range |
|---------|---------|-------|
| Link TTL | 5 minutes | 1 min - 24 hours |

### Validation

When a click arrives, AffTok validates:

1. **Signature** - HMAC signature matches
2. **TTL** - `current_time - timestamp < TTL`
3. **Replay** - Nonce hasn't been used before

### Expired Link Handling

```
┌─────────────────────────────────────────────────────────────┐
│                  Expired Link Flow                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  User clicks expired link                                   │
│         │                                                   │
│         ▼                                                   │
│  Check timestamp: now - link_timestamp > TTL                │
│         │                                                   │
│         ▼                                                   │
│  Link expired? ────Yes───▶ Log fraud event                  │
│         │                        │                          │
│         No                       ▼                          │
│         │               Skip click tracking                 │
│         ▼                        │                          │
│  Continue normal flow            ▼                          │
│                          Redirect to destination            │
│                          (user still sees landing page)     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

> **Note**: Even expired links redirect users to the destination. The click just isn't tracked.

## Replay Protection

Each nonce can only be used once:

```
┌─────────────────────────────────────────────────────────────┐
│                  Replay Protection                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Click arrives with nonce: "abc123xyz"                      │
│         │                                                   │
│         ▼                                                   │
│  Check Redis: EXISTS replay:nonce:abc123xyz                 │
│         │                                                   │
│         ├── Exists ──▶ REPLAY_ATTEMPT (reject)              │
│         │                                                   │
│         └── Not exists                                      │
│                │                                            │
│                ▼                                            │
│         Store nonce: SET replay:nonce:abc123xyz 1 EX 300    │
│                │                                            │
│                ▼                                            │
│         Continue processing                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Sub IDs

Add sub IDs to tracking links for additional tracking:

```
https://track.afftok.com/c/abc123.ts.nonce.sig?sub1=facebook&sub2=banner1&sub3=mobile
```

### Available Sub ID Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `sub1` | Primary source | `facebook`, `google`, `email` |
| `sub2` | Secondary source | `campaign_123`, `banner_1` |
| `sub3` | Tertiary source | `creative_456`, `mobile` |
| `sub4` | Custom | Any value |
| `sub5` | Custom | Any value |

### Sub ID Macros

Use macros for dynamic values:

| Macro | Description | Replaced With |
|-------|-------------|---------------|
| `{click_id}` | Unique click ID | `clk_abc123` |
| `{timestamp}` | Click timestamp | `1699876543` |
| `{ip}` | User IP | `203.0.113.45` |
| `{country}` | Country code | `US` |
| `{device}` | Device type | `mobile` |
| `{os}` | Operating system | `iOS` |
| `{browser}` | Browser | `Safari` |

Example with macros:
```
https://example.com/landing?click_id={click_id}&country={country}
```

## Backward Compatibility

AffTok supports legacy unsigned links with a configuration flag:

```bash
# Environment variable
ALLOW_LEGACY_TRACKING_CODES=true
```

When enabled:
- Unsigned links (`/c/abc123`) are accepted
- A warning is logged
- Signature/TTL checks are skipped
- Geo rules and deduplication still apply

> **Warning**: Legacy mode is less secure. Use signed links in production.

## API: Get Tracking Link

When you join an offer, you receive tracking links:

```bash
curl -X POST https://api.afftok.com/api/offers/off_123/join \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**

```json
{
  "success": true,
  "data": {
    "user_offer_id": "uo_xyz789",
    "tracking_code": "abc123def456",
    "tracking_url": "https://track.afftok.com/c/abc123def456",
    "signed_link": "https://track.afftok.com/c/abc123def456.1699876543210.nonce123.sig456",
    "short_link": "https://afftok.link/xYz",
    "destination_url": "https://example.com/landing"
  }
}
```

## API: Generate New Signed Link

Generate a fresh signed link for an existing tracking code:

```bash
curl -X POST https://api.afftok.com/api/links/sign \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tracking_code": "abc123def456",
    "ttl_minutes": 10
  }'
```

**Response:**

```json
{
  "success": true,
  "data": {
    "signed_link": "https://track.afftok.com/c/abc123def456.1699876600000.newnonce.newsig",
    "expires_at": "2024-01-15T10:10:00Z",
    "ttl_seconds": 600
  }
}
```

## Best Practices

1. **Always use signed links** - Better security and fraud protection
2. **Generate fresh links** - Don't reuse old signed links
3. **Set appropriate TTL** - Balance security vs. user experience
4. **Use sub IDs** - Track traffic sources for optimization
5. **Monitor expired links** - High rates may indicate issues
6. **Rotate secrets** - Periodically rotate signing secrets

## Next Steps

- [Conversions Guide](conversions.md) - Track conversions
- [Security Guide](../security/link-signing.md) - Deep dive into link security
- [API Reference](../api-reference/clicks.md) - Full click API documentation

