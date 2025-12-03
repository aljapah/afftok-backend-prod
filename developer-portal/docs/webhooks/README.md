# Webhook Documentation

Complete guide to AffTok's webhook system for real-time event notifications.

## Overview

AffTok webhooks enable real-time notifications when events occur in your tracking pipeline:

- **Conversion Events** - When a conversion is recorded
- **Click Events** - When a click is tracked (high volume)
- **Postback Events** - When a postback is received
- **Fraud Events** - When fraud is detected

## Key Features

| Feature | Description |
|---------|-------------|
| Multi-Step Pipelines | Chain multiple webhook calls |
| Retry with Backoff | Automatic retry on failures |
| Payload Signing | HMAC-SHA256 or JWT signatures |
| Template Engine | Dynamic payload generation |
| Dead Letter Queue | Handle persistent failures |
| Failover URLs | Backup endpoints |

## Quick Start

### 1. Create a Webhook Pipeline

```bash
curl -X POST https://api.afftok.com/api/admin/webhooks/pipelines \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Conversion Notification",
    "trigger_type": "conversion",
    "status": "active",
    "steps": [
      {
        "order": 1,
        "name": "Notify Partner",
        "url": "https://partner.example.com/webhook",
        "method": "POST",
        "headers": {
          "Content-Type": "application/json"
        },
        "body_template": "{\"conversion_id\": \"{{conversion.id}}\", \"amount\": {{conversion.amount}}}",
        "signature_mode": "hmac",
        "signing_key": "your_secret_key"
      }
    ]
  }'
```

### 2. Receive Webhook

Your endpoint receives:

```http
POST /webhook HTTP/1.1
Host: partner.example.com
Content-Type: application/json
X-Afftok-Signature: sha256=abc123...
X-Afftok-Timestamp: 1701432000
X-Afftok-Delivery-ID: wh_del_123

{
  "conversion_id": "conv_456",
  "amount": 49.99
}
```

### 3. Verify Signature

```javascript
const crypto = require('crypto');

function verifyWebhook(body, signature, secret) {
  const expected = 'sha256=' + crypto
    .createHmac('sha256', secret)
    .update(body)
    .digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expected)
  );
}

// Express.js example
app.post('/webhook', express.raw({ type: 'application/json' }), (req, res) => {
  const signature = req.headers['x-afftok-signature'];
  const body = req.body.toString();
  
  if (!verifyWebhook(body, signature, process.env.WEBHOOK_SECRET)) {
    return res.status(401).send('Invalid signature');
  }
  
  const event = JSON.parse(body);
  console.log('Received:', event);
  
  res.status(200).send('OK');
});
```

---

## Trigger Types

### Conversion

Triggered when a new conversion is recorded.

```json
{
  "trigger_type": "conversion",
  "available_data": {
    "conversion": {
      "id": "conv_123",
      "click_id": "click_456",
      "offer_id": "off_789",
      "user_id": "user_012",
      "amount": 49.99,
      "currency": "USD",
      "status": "approved",
      "external_id": "ext_123",
      "timestamp": "2024-12-01T12:00:00Z"
    },
    "click": {
      "id": "click_456",
      "ip": "192.168.1.1",
      "country": "US",
      "device": "mobile",
      "browser": "Chrome",
      "os": "iOS"
    },
    "offer": {
      "id": "off_789",
      "title": "Premium VPN",
      "payout": 25.00
    }
  }
}
```

### Click

Triggered when a click is tracked. **Warning: High volume!**

```json
{
  "trigger_type": "click",
  "available_data": {
    "click": {
      "id": "click_456",
      "offer_id": "off_789",
      "user_id": "user_012",
      "ip": "192.168.1.1",
      "country": "US",
      "city": "New York",
      "device": "mobile",
      "browser": "Chrome",
      "os": "iOS",
      "user_agent": "...",
      "timestamp": "2024-12-01T12:00:00Z"
    }
  }
}
```

### Postback

Triggered when a postback is received from an advertiser.

```json
{
  "trigger_type": "postback",
  "available_data": {
    "postback": {
      "id": "pb_123",
      "conversion_id": "conv_456",
      "external_id": "ext_789",
      "status": "approved",
      "amount": 49.99,
      "raw_data": "{...}",
      "timestamp": "2024-12-01T12:00:00Z"
    }
  }
}
```

### Fraud

Triggered when fraud is detected.

```json
{
  "trigger_type": "fraud",
  "available_data": {
    "fraud": {
      "type": "bot_detection",
      "ip": "192.168.1.1",
      "risk_score": 95,
      "indicators": ["headless_browser", "high_frequency"],
      "blocked": true,
      "timestamp": "2024-12-01T12:00:00Z"
    }
  }
}
```

---

## Template Engine

Use placeholders in URL, headers, and body:

### Syntax

```
{{object.property}}
```

### Available Placeholders

#### Click Data

| Placeholder | Description |
|-------------|-------------|
| `{{click.id}}` | Click ID |
| `{{click.ip}}` | IP address |
| `{{click.country}}` | Country code |
| `{{click.city}}` | City name |
| `{{click.device}}` | Device type |
| `{{click.browser}}` | Browser name |
| `{{click.os}}` | Operating system |
| `{{click.user_agent}}` | Full user agent |
| `{{click.timestamp}}` | ISO timestamp |

#### Conversion Data

| Placeholder | Description |
|-------------|-------------|
| `{{conversion.id}}` | Conversion ID |
| `{{conversion.click_id}}` | Associated click ID |
| `{{conversion.offer_id}}` | Offer ID |
| `{{conversion.user_id}}` | User ID |
| `{{conversion.amount}}` | Conversion amount |
| `{{conversion.currency}}` | Currency code |
| `{{conversion.status}}` | Status (approved/pending/rejected) |
| `{{conversion.external_id}}` | External conversion ID |
| `{{conversion.timestamp}}` | ISO timestamp |

#### Offer Data

| Placeholder | Description |
|-------------|-------------|
| `{{offer.id}}` | Offer ID |
| `{{offer.title}}` | Offer title |
| `{{offer.payout}}` | Payout amount |
| `{{offer.network_id}}` | Network ID |

#### System Data

| Placeholder | Description |
|-------------|-------------|
| `{{timestamp}}` | Current ISO timestamp |
| `{{timestamp_unix}}` | Current Unix timestamp |
| `{{uuid}}` | Random UUID |
| `{{nonce}}` | Random nonce |
| `{{delivery_id}}` | Webhook delivery ID |

### Example Template

```json
{
  "body_template": "{\n  \"event\": \"conversion\",\n  \"id\": \"{{conversion.id}}\",\n  \"offer\": \"{{offer.title}}\",\n  \"amount\": {{conversion.amount}},\n  \"currency\": \"{{conversion.currency}}\",\n  \"source\": {\n    \"country\": \"{{click.country}}\",\n    \"device\": \"{{click.device}}\"\n  },\n  \"meta\": {\n    \"timestamp\": \"{{timestamp}}\",\n    \"nonce\": \"{{nonce}}\"\n  }\n}"
}
```

Renders to:

```json
{
  "event": "conversion",
  "id": "conv_123",
  "offer": "Premium VPN",
  "amount": 49.99,
  "currency": "USD",
  "source": {
    "country": "US",
    "device": "mobile"
  },
  "meta": {
    "timestamp": "2024-12-01T12:00:00Z",
    "nonce": "abc123xyz"
  }
}
```

---

## Signature Verification

### HMAC-SHA256

AffTok adds `X-Afftok-Signature` header:

```
X-Afftok-Signature: sha256=<hex_digest>
```

#### Verification Examples

**Node.js:**

```javascript
const crypto = require('crypto');

function verifyHMAC(body, signature, secret) {
  const [algo, hash] = signature.split('=');
  const expected = crypto
    .createHmac('sha256', secret)
    .update(body)
    .digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(hash),
    Buffer.from(expected)
  );
}
```

**Python:**

```python
import hmac
import hashlib

def verify_hmac(body: bytes, signature: str, secret: str) -> bool:
    algo, hash_value = signature.split('=')
    expected = hmac.new(
        secret.encode(),
        body,
        hashlib.sha256
    ).hexdigest()
    
    return hmac.compare_digest(hash_value, expected)
```

**PHP:**

```php
function verifyHMAC($body, $signature, $secret) {
    list($algo, $hash) = explode('=', $signature, 2);
    $expected = hash_hmac('sha256', $body, $secret);
    
    return hash_equals($hash, $expected);
}
```

**Go:**

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "strings"
)

func verifyHMAC(body []byte, signature, secret string) bool {
    parts := strings.SplitN(signature, "=", 2)
    if len(parts) != 2 {
        return false
    }
    
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := hex.EncodeToString(mac.Sum(nil))
    
    return hmac.Equal([]byte(parts[1]), []byte(expected))
}
```

### JWT

AffTok adds `Authorization: Bearer <jwt>` header.

JWT payload:

```json
{
  "delivery_id": "wh_del_123",
  "pipeline_id": "wh_pipe_456",
  "trigger_type": "conversion",
  "trigger_id": "conv_789",
  "iat": 1701432000,
  "exp": 1701432300
}
```

#### Verification Example (Node.js)

```javascript
const jwt = require('jsonwebtoken');

function verifyJWT(token, secret) {
  try {
    const decoded = jwt.verify(token, secret, {
      algorithms: ['HS256'],
    });
    return { valid: true, payload: decoded };
  } catch (error) {
    return { valid: false, error: error.message };
  }
}

// Express.js
app.post('/webhook', (req, res) => {
  const authHeader = req.headers['authorization'];
  if (!authHeader?.startsWith('Bearer ')) {
    return res.status(401).send('Missing token');
  }
  
  const token = authHeader.slice(7);
  const result = verifyJWT(token, process.env.WEBHOOK_SECRET);
  
  if (!result.valid) {
    return res.status(401).send('Invalid token');
  }
  
  // Process webhook
  res.status(200).send('OK');
});
```

---

## Retry Policy

Default retry schedule:

| Attempt | Delay |
|---------|-------|
| 1 | Immediate |
| 2 | 5 seconds |
| 3 | 10 seconds |
| 4 | 30 seconds |
| 5 | 1 minute |
| 6 | 5 minutes |

After max retries:
1. Moves to **failover queue**
2. Tries failover URL (if configured)
3. If failover fails, moves to **Dead Letter Queue (DLQ)**

### Custom Retry Policy

```json
{
  "retry_policy": {
    "max_attempts": 3,
    "initial_delay_seconds": 10,
    "backoff_multiplier": 3,
    "max_delay_seconds": 300
  }
}
```

---

## Response Requirements

Your endpoint must:

1. **Return 2xx status** within timeout (default: 30s)
2. **Respond quickly** - Don't process synchronously
3. **Be idempotent** - Handle duplicate deliveries

### Recommended Pattern

```javascript
app.post('/webhook', async (req, res) => {
  // 1. Verify signature
  if (!verifySignature(req)) {
    return res.status(401).send('Invalid signature');
  }
  
  // 2. Acknowledge immediately
  res.status(200).send('OK');
  
  // 3. Process asynchronously
  setImmediate(() => {
    processWebhook(req.body);
  });
});
```

---

## Dead Letter Queue

Failed webhooks are stored in the DLQ for manual review:

### View DLQ Items

```bash
curl https://api.afftok.com/api/admin/webhooks/dlq \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### Retry DLQ Item

```bash
curl -X POST https://api.afftok.com/api/admin/webhooks/dlq/dlq_123/retry \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### Delete DLQ Item

```bash
curl -X DELETE https://api.afftok.com/api/admin/webhooks/dlq/dlq_123 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

---

## Best Practices

### 1. Always Verify Signatures

Never trust webhooks without signature verification.

### 2. Respond Quickly

Return 200 immediately, process asynchronously.

### 3. Handle Duplicates

Use `X-Afftok-Delivery-ID` for idempotency:

```javascript
const processedDeliveries = new Set();

app.post('/webhook', (req, res) => {
  const deliveryId = req.headers['x-afftok-delivery-id'];
  
  if (processedDeliveries.has(deliveryId)) {
    return res.status(200).send('Already processed');
  }
  
  processedDeliveries.add(deliveryId);
  // Process webhook...
  res.status(200).send('OK');
});
```

### 4. Log Everything

Keep logs for debugging:

```javascript
app.post('/webhook', (req, res) => {
  console.log('Webhook received:', {
    deliveryId: req.headers['x-afftok-delivery-id'],
    timestamp: req.headers['x-afftok-timestamp'],
    body: req.body,
  });
  // ...
});
```

### 5. Use HTTPS

Always use HTTPS endpoints for production.

### 6. Set Appropriate Timeouts

Configure your server to handle the webhook timeout (default: 30s).

---

## Troubleshooting

### Webhook Not Received

1. Check pipeline status is `active`
2. Verify URL is accessible
3. Check firewall rules
4. Review logs in admin panel

### Signature Verification Failing

1. Ensure raw body is used (not parsed JSON)
2. Check signing key matches
3. Verify signature header name

### Timeouts

1. Respond faster (< 30s)
2. Process asynchronously
3. Increase timeout in pipeline config

### High Failure Rate

1. Check endpoint health
2. Review error responses
3. Consider adding failover URL

---

## Headers Reference

| Header | Description |
|--------|-------------|
| `X-Afftok-Signature` | HMAC signature (if enabled) |
| `X-Afftok-Timestamp` | Unix timestamp of delivery |
| `X-Afftok-Delivery-ID` | Unique delivery identifier |
| `X-Afftok-Retry-Count` | Current retry attempt (0-based) |
| `Authorization` | JWT token (if JWT signing enabled) |
| `Content-Type` | Always `application/json` |

---

Next: [Security Documentation](../security/README.md)

