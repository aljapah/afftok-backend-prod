# Signed Tracking Links

Signed links provide enhanced security through cryptographic verification, TTL (time-to-live), and replay protection.

## Overview

AffTok signed links contain four components:

```
https://track.afftok.com/c/{tracking_code}.{timestamp}.{nonce}.{signature}
                           └──────┬──────┘ └────┬────┘ └──┬──┘ └────┬────┘
                             Tracking ID    Unix ms    Random   HMAC-SHA256
```

## Link Components

| Component | Description | Example |
|-----------|-------------|---------|
| `tracking_code` | Unique identifier for user+offer | `abc123xyz` |
| `timestamp` | Unix timestamp in milliseconds | `1701234567890` |
| `nonce` | Random 32-character string | `a1b2c3d4e5f6g7h8` |
| `signature` | HMAC-SHA256 signature (first 16 chars) | `9f8e7d6c5b4a3210` |

## How Signing Works

### Signature Generation

```
data_to_sign = "{tracking_code}.{timestamp}.{nonce}"
signature = HMAC-SHA256(secret_key, data_to_sign).substring(0, 16)
signed_link = "{base_url}/c/{tracking_code}.{timestamp}.{nonce}.{signature}"
```

### Example

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
const signedLink = generateSignedLink('abc123xyz', 'your_secret_key');
// Output: https://track.afftok.com/c/abc123xyz.1701234567890.a1b2c3d4e5f6g7h8.9f8e7d6c5b4a3210
```

## TTL (Time-to-Live)

Signed links expire after a configurable TTL:

| Setting | Default | Range |
|---------|---------|-------|
| Link TTL | 5 minutes | 1 min - 24 hours |

### Checking Expiration

```javascript
function isLinkExpired(timestamp, ttlMinutes = 5) {
  const now = Date.now();
  const age = now - timestamp;
  const ttlMs = ttlMinutes * 60 * 1000;
  return age > ttlMs;
}
```

### Expired Link Behavior

When a link expires:

1. Click is logged with `expired_link` flag
2. User is still redirected (to avoid UX issues)
3. Fraud event is recorded
4. Stats may be affected based on configuration

## Replay Protection

Each nonce can only be used once within the TTL window.

### How It Works

```
1. User clicks signed link
2. Edge extracts nonce from link
3. Check Redis: EXISTS nonce:{nonce}
4. If exists → REPLAY_ATTEMPT (reject)
5. If not exists → SETEX nonce:{nonce} {ttl} 1 (allow)
```

### Replay Detection Response

```json
{
  "success": false,
  "error": "Replay attempt detected",
  "code": "REPLAY_ATTEMPT"
}
```

## Validation Process

```
┌─────────────────────────────────────────────────────────────┐
│                 SIGNED LINK VALIDATION                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Parse Link                                              │
│     /c/abc123.1701234567890.a1b2c3d4.9f8e7d6c               │
│                                                             │
│  2. Validate Format                                         │
│     ├── Has 4 components? ✓                                 │
│     ├── Timestamp is numeric? ✓                             │
│     └── Nonce is 32 chars? ✓                                │
│                                                             │
│  3. Check TTL                                               │
│     now - timestamp < 5 minutes? ✓                          │
│                                                             │
│  4. Verify Signature                                        │
│     expected = HMAC(secret, "abc123.1701234567890.a1b2c3d4")│
│     expected.substring(0,16) == "9f8e7d6c"? ✓               │
│                                                             │
│  5. Check Replay                                            │
│     Redis GET nonce:a1b2c3d4 == null? ✓                     │
│                                                             │
│  6. Store Nonce                                             │
│     Redis SETEX nonce:a1b2c3d4 300 1                        │
│                                                             │
│  RESULT: VALID ✓                                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Configuration

### Admin API

```bash
# Get current config
GET /api/admin/link-signing/config

# Update TTL
PUT /api/admin/link-signing/config
{
  "ttl_minutes": 10,
  "allow_legacy": false
}

# Rotate signing secret
POST /api/admin/link-signing/rotate-secret

# Clear replay cache
POST /api/admin/link-signing/replay/clear
```

### Environment Variables

```bash
# Signing secret (required)
SIGNING_SECRET_KEY=your_32_char_secret_key_here

# TTL in minutes (default: 5)
LINK_TTL_MINUTES=5

# Allow legacy unsigned links (default: false)
ALLOW_LEGACY_TRACKING_CODES=false
```

## Legacy Link Support

For backward compatibility, unsigned links can be allowed:

```
# Legacy format
https://track.afftok.com/c/abc123xyz

# Enable legacy support
ALLOW_LEGACY_TRACKING_CODES=true
```

> ⚠️ **Warning**: Legacy links are less secure. Migrate to signed links as soon as possible.

## Testing Signed Links

### Generate Test Link

```bash
curl -X POST "https://api.afftok.com/api/admin/link-signing/test" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{"tracking_code": "test123"}'
```

### Validate Link

```bash
curl -X GET "https://api.afftok.com/api/admin/link-signing/test?link=https://track.afftok.com/c/test123.1701234567890.abc123.def456" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Response:**

```json
{
  "valid": true,
  "tracking_code": "test123",
  "timestamp": 1701234567890,
  "nonce": "abc123",
  "signature": "def456",
  "ttl_remaining_seconds": 245,
  "is_replay": false
}
```

## SDK Support

All AffTok SDKs support signed links:

### Web SDK

```javascript
await Afftok.trackSignedClick(
  'https://track.afftok.com/c/abc123.1701234567890.xyz.sig',
  { offerId: 'offer_123' }
);
```

### Mobile SDKs

```kotlin
// Android
Afftok.trackSignedClick(
    signedLink = "https://track.afftok.com/c/abc123.1701234567890.xyz.sig",
    params = ClickParams(offerId = "offer_123")
)
```

```swift
// iOS
await Afftok.shared.trackSignedClick(
    signedLink: "https://track.afftok.com/c/abc123.1701234567890.xyz.sig",
    params: ClickParams(offerId: "offer_123")
)
```

## Best Practices

1. **Always use signed links** in production
2. **Keep secrets secure** - never expose in client code
3. **Monitor expired links** - high rates may indicate issues
4. **Rotate secrets periodically** - at least every 90 days
5. **Set appropriate TTL** - balance security vs. usability

## Troubleshooting

### Invalid Signature

- Verify secret key matches
- Check timestamp format (milliseconds)
- Ensure nonce is exactly as generated

### Expired Links

- Check system clock synchronization
- Consider increasing TTL for slow networks
- Pre-generate links closer to use time

### Replay Detected

- Each link should be used only once
- Don't cache or reuse signed links
- Generate fresh links for each click

---

Next: [API Reference](../api/README.md)

