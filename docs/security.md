# AffTok Security Guide

Best practices for secure integration with AffTok.

## Overview

AffTok implements multiple layers of security:

- **HMAC-SHA256 Signatures** - Request authentication
- **API Key Management** - Access control
- **Rate Limiting** - Abuse prevention
- **Link Signing** - Tamper protection
- **Replay Protection** - Nonce validation
- **Geo Rules** - Geographic restrictions
- **Bot Detection** - Fraud prevention

---

## API Key Security

### Generating API Keys

1. Log in to AffTok Dashboard
2. Navigate to Settings → API Keys
3. Click "Generate New Key"
4. Store the key securely (shown only once)

### Key Best Practices

✅ **DO:**
- Store keys in environment variables
- Use different keys for dev/staging/production
- Rotate keys periodically
- Monitor key usage

❌ **DON'T:**
- Hardcode keys in source code
- Commit keys to version control
- Share keys between applications
- Use production keys in development

### Environment Variables

```bash
# .env file (never commit!)
AFFTOK_API_KEY=afftok_live_sk_xxxxxxxxxxxxx
AFFTOK_ADVERTISER_ID=adv_xxxxxxxxxxxxx
```

```javascript
// Load from environment
const apiKey = process.env.AFFTOK_API_KEY;
```

### Key Rotation

Rotate API keys regularly:

1. Generate new key in dashboard
2. Update application configuration
3. Deploy changes
4. Revoke old key after verification

---

## Signature Generation

All requests must include HMAC-SHA256 signature.

### Signature Format

```
signature = HMAC-SHA256(api_key, data_to_sign)
data_to_sign = "{api_key}|{advertiser_id}|{timestamp}|{nonce}"
```

### Implementation Examples

#### JavaScript

```javascript
const crypto = require('crypto');

function generateSignature(apiKey, advertiserId, timestamp, nonce) {
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  return crypto
    .createHmac('sha256', apiKey)
    .update(dataToSign)
    .digest('hex');
}
```

#### Python

```python
import hmac
import hashlib

def generate_signature(api_key, advertiser_id, timestamp, nonce):
    data_to_sign = f"{api_key}|{advertiser_id}|{timestamp}|{nonce}"
    return hmac.new(
        api_key.encode(),
        data_to_sign.encode(),
        hashlib.sha256
    ).hexdigest()
```

#### PHP

```php
function generateSignature($apiKey, $advertiserId, $timestamp, $nonce) {
    $dataToSign = "{$apiKey}|{$advertiserId}|{$timestamp}|{$nonce}";
    return hash_hmac('sha256', $dataToSign, $apiKey);
}
```

#### Go

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
)

func generateSignature(apiKey, advertiserId string, timestamp int64, nonce string) string {
    dataToSign := fmt.Sprintf("%s|%s|%d|%s", apiKey, advertiserId, timestamp, nonce)
    h := hmac.New(sha256.New, []byte(apiKey))
    h.Write([]byte(dataToSign))
    return hex.EncodeToString(h.Sum(nil))
}
```

### Nonce Generation

Generate unique 32-character nonces:

```javascript
function generateNonce(length = 32) {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  const randomValues = new Uint8Array(length);
  crypto.getRandomValues(randomValues);
  for (let i = 0; i < length; i++) {
    result += chars[randomValues[i] % chars.length];
  }
  return result;
}
```

---

## Link Signing

Tracking links are cryptographically signed to prevent tampering.

### Signed Link Format

```
https://track.afftok.com/c/{tracking_code}.{timestamp}.{nonce}.{signature}
```

### Validating Signed Links

```javascript
function validateSignedLink(link, secretKey) {
  const parts = link.split('.');
  if (parts.length !== 4) return false;
  
  const [trackingCode, timestamp, nonce, signature] = parts;
  
  // Check expiration (5 minutes)
  const now = Date.now();
  if (now - parseInt(timestamp) > 5 * 60 * 1000) {
    return false;
  }
  
  // Verify signature
  const dataToSign = `${trackingCode}.${timestamp}.${nonce}`;
  const expectedSignature = crypto
    .createHmac('sha256', secretKey)
    .update(dataToSign)
    .digest('hex')
    .substring(0, 16);
  
  return signature === expectedSignature;
}
```

---

## Rate Limiting

### Default Limits

| Limit | Value |
|-------|-------|
| Requests per minute | 60 |
| Requests per hour | 1000 |
| Requests per day | 10000 |

### Rate Limit Headers

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1234567890
```

### Handling Rate Limits

```javascript
async function makeRequest(url, data) {
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  
  if (response.status === 429) {
    const retryAfter = response.headers.get('X-RateLimit-Reset');
    const waitTime = (retryAfter - Date.now() / 1000) * 1000;
    await sleep(waitTime);
    return makeRequest(url, data); // Retry
  }
  
  return response.json();
}
```

---

## Replay Protection

AffTok prevents replay attacks using nonces.

### How It Works

1. Each request includes a unique nonce
2. Server stores used nonces (TTL: 5 minutes)
3. Duplicate nonces are rejected

### Best Practices

- Generate new nonce for each request
- Never reuse nonces
- Use cryptographically secure random generator

---

## Bot Detection

AffTok detects and blocks bot traffic.

### Detection Methods

- User agent analysis
- Behavioral patterns
- IP reputation
- Device fingerprinting
- Request timing

### Avoiding False Positives

- Use real browser user agents
- Include device information
- Don't make requests too quickly
- Use consistent fingerprints

---

## Geo Rules

Restrict traffic by geographic location.

### Configuration

Set geo rules in the dashboard:

1. Navigate to Offers → Geo Rules
2. Select Allow or Block mode
3. Choose countries
4. Set priority

### Rule Resolution Order

1. Offer-specific rules
2. Advertiser rules
3. Global rules
4. Default: Allow

---

## Webhook Security

### Verifying Webhooks

```javascript
const crypto = require('crypto');

function verifyWebhook(payload, signature, webhookSecret) {
  const expectedSignature = crypto
    .createHmac('sha256', webhookSecret)
    .update(JSON.stringify(payload))
    .digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}

// Usage
app.post('/webhook', (req, res) => {
  const signature = req.headers['x-afftok-signature'];
  
  if (!verifyWebhook(req.body, signature, WEBHOOK_SECRET)) {
    return res.status(401).send('Invalid signature');
  }
  
  // Process webhook
  res.status(200).send('OK');
});
```

---

## Data Protection

### Sensitive Data

Never log or store:
- Full API keys
- User passwords
- Payment card numbers
- Personal identifiable information (PII)

### Data Transmission

- Always use HTTPS
- Verify SSL certificates
- Use TLS 1.2 or higher

### Data Storage

- Encrypt sensitive data at rest
- Use secure key management
- Implement access controls

---

## Security Checklist

### Before Launch

- [ ] API keys stored in environment variables
- [ ] Signature generation verified
- [ ] Rate limit handling implemented
- [ ] Error handling for security errors
- [ ] Webhook verification implemented
- [ ] HTTPS enforced
- [ ] Logging doesn't include sensitive data

### Ongoing

- [ ] Monitor API key usage
- [ ] Review rate limit metrics
- [ ] Check for suspicious activity
- [ ] Rotate keys periodically
- [ ] Update SDKs regularly
- [ ] Review security logs

---

## Reporting Security Issues

If you discover a security vulnerability:

1. **Do NOT** disclose publicly
2. Email: security@afftok.com
3. Include detailed description
4. Provide steps to reproduce
5. Wait for acknowledgment

We respond within 24 hours and offer bug bounties for valid reports.

---

## Support

- Security Email: security@afftok.com
- Documentation: https://docs.afftok.com
- Status: https://status.afftok.com

