# Webhook Security

Secure your webhook endpoints with signature verification and best practices.

---

## Signature Verification

Every webhook request includes a signature in the `X-Afftok-Signature` header. This signature is an HMAC-SHA256 hash of the request body using your webhook secret.

### Signature Format

```
X-Afftok-Signature: sha256=<hex_signature>
```

### Verification Process

1. Get the signature from the header
2. Compute HMAC-SHA256 of the raw request body
3. Compare signatures using constant-time comparison

---

## Implementation Examples

### Node.js

```javascript
const crypto = require('crypto');

function verifyWebhookSignature(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  // Constant-time comparison to prevent timing attacks
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}

// Express middleware
function webhookAuth(req, res, next) {
  const signature = req.headers['x-afftok-signature'];
  
  if (!signature) {
    return res.status(401).json({ error: 'Missing signature' });
  }
  
  // Remove 'sha256=' prefix if present
  const sig = signature.replace('sha256=', '');
  
  // Get raw body
  const payload = req.rawBody || JSON.stringify(req.body);
  
  if (!verifyWebhookSignature(payload, sig, process.env.WEBHOOK_SECRET)) {
    return res.status(401).json({ error: 'Invalid signature' });
  }
  
  next();
}

// Usage
app.post('/webhooks/afftok', webhookAuth, (req, res) => {
  // Signature verified, process event
  res.status(200).json({ received: true });
});
```

### Python

```python
import hmac
import hashlib

def verify_webhook_signature(payload: bytes, signature: str, secret: str) -> bool:
    expected_signature = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    
    # Constant-time comparison
    return hmac.compare_digest(signature, expected_signature)

# Flask example
from flask import Flask, request, jsonify
import os

app = Flask(__name__)

@app.route('/webhooks/afftok', methods=['POST'])
def handle_webhook():
    signature = request.headers.get('X-Afftok-Signature', '')
    
    # Remove 'sha256=' prefix if present
    sig = signature.replace('sha256=', '')
    
    if not verify_webhook_signature(
        request.get_data(),
        sig,
        os.environ['WEBHOOK_SECRET']
    ):
        return jsonify({'error': 'Invalid signature'}), 401
    
    # Process event
    return jsonify({'received': True}), 200
```

### PHP

```php
<?php

function verifyWebhookSignature(string $payload, string $signature, string $secret): bool {
    $expectedSignature = hash_hmac('sha256', $payload, $secret);
    
    // Constant-time comparison
    return hash_equals($expectedSignature, $signature);
}

// Get signature
$signature = $_SERVER['HTTP_X_AFFTOK_SIGNATURE'] ?? '';
$signature = str_replace('sha256=', '', $signature);

// Get raw body
$payload = file_get_contents('php://input');

if (!verifyWebhookSignature($payload, $signature, getenv('WEBHOOK_SECRET'))) {
    http_response_code(401);
    echo json_encode(['error' => 'Invalid signature']);
    exit;
}

// Process event
$event = json_decode($payload, true);
// ...
```

### Go

```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "io"
    "net/http"
    "os"
    "strings"
)

func verifyWebhookSignature(payload []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    signature := r.Header.Get("X-Afftok-Signature")
    signature = strings.TrimPrefix(signature, "sha256=")
    
    payload, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read body", http.StatusBadRequest)
        return
    }
    
    if !verifyWebhookSignature(payload, signature, os.Getenv("WEBHOOK_SECRET")) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Process event
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"received": true}`))
}
```

### Ruby

```ruby
require 'openssl'
require 'sinatra'
require 'json'

def verify_webhook_signature(payload, signature, secret)
  expected_signature = OpenSSL::HMAC.hexdigest('sha256', secret, payload)
  Rack::Utils.secure_compare(signature, expected_signature)
end

post '/webhooks/afftok' do
  signature = request.env['HTTP_X_AFFTOK_SIGNATURE'].to_s.sub('sha256=', '')
  payload = request.body.read
  
  unless verify_webhook_signature(payload, signature, ENV['WEBHOOK_SECRET'])
    halt 401, { error: 'Invalid signature' }.to_json
  end
  
  event = JSON.parse(payload)
  # Process event
  
  { received: true }.to_json
end
```

---

## JWT Signed Webhooks

For enhanced security, webhooks can also be signed with JWT:

### JWT Header

```
Authorization: Bearer <jwt_token>
```

### JWT Payload

```json
{
  "task_id": "task_abc123",
  "advertiser_id": "adv_xyz789",
  "timestamp": 1705320645,
  "iat": 1705320645,
  "exp": 1705320945
}
```

### Verification

```javascript
const jwt = require('jsonwebtoken');

function verifyJwtSignature(token, secret) {
  try {
    const decoded = jwt.verify(token, secret, {
      algorithms: ['HS256'],
      maxAge: '5m', // Token must be less than 5 minutes old
    });
    return { valid: true, payload: decoded };
  } catch (error) {
    return { valid: false, error: error.message };
  }
}

// Usage
const authHeader = req.headers['authorization'];
const token = authHeader?.replace('Bearer ', '');

if (token) {
  const result = verifyJwtSignature(token, process.env.WEBHOOK_SECRET);
  if (!result.valid) {
    return res.status(401).json({ error: 'Invalid JWT' });
  }
}
```

---

## Timestamp Validation

Prevent replay attacks by validating the timestamp:

```javascript
const MAX_AGE_SECONDS = 300; // 5 minutes

function validateTimestamp(timestamp) {
  const eventTime = new Date(timestamp).getTime();
  const now = Date.now();
  const age = (now - eventTime) / 1000;
  
  return age >= 0 && age <= MAX_AGE_SECONDS;
}

// In webhook handler
if (!validateTimestamp(event.timestamp)) {
  return res.status(400).json({ error: 'Event too old' });
}
```

---

## IP Allowlisting

Restrict webhook requests to AffTok IP addresses:

### AffTok IP Ranges

```
# Production
52.89.214.238/32
34.212.75.30/32
54.218.53.128/32

# Staging (optional)
34.211.200.85/32
```

### Implementation

```javascript
const allowedIPs = [
  '52.89.214.238',
  '34.212.75.30',
  '54.218.53.128',
];

function ipAllowlist(req, res, next) {
  const clientIP = req.ip || req.connection.remoteAddress;
  
  if (!allowedIPs.includes(clientIP)) {
    return res.status(403).json({ error: 'IP not allowed' });
  }
  
  next();
}

app.post('/webhooks/afftok', ipAllowlist, webhookAuth, handler);
```

---

## Idempotency

Handle duplicate deliveries safely:

```javascript
const Redis = require('ioredis');
const redis = new Redis();

async function processWebhook(event) {
  const eventId = event.id;
  
  // Check if already processed
  const processed = await redis.get(`webhook:${eventId}`);
  if (processed) {
    console.log('Duplicate event, skipping:', eventId);
    return { duplicate: true };
  }
  
  // Mark as processing
  await redis.set(`webhook:${eventId}`, 'processing', 'EX', 86400);
  
  try {
    // Process event
    await handleEvent(event);
    
    // Mark as completed
    await redis.set(`webhook:${eventId}`, 'completed', 'EX', 86400);
    return { success: true };
  } catch (error) {
    // Remove processing flag on error
    await redis.del(`webhook:${eventId}`);
    throw error;
  }
}
```

---

## Security Checklist

- [ ] **HTTPS Only** - Never accept webhooks over HTTP
- [ ] **Signature Verification** - Always verify HMAC signature
- [ ] **Constant-Time Comparison** - Use timing-safe comparison
- [ ] **Timestamp Validation** - Reject old events
- [ ] **IP Allowlisting** - Optionally restrict source IPs
- [ ] **Idempotency** - Handle duplicate deliveries
- [ ] **Rate Limiting** - Protect against flood attacks
- [ ] **Logging** - Log all webhook events
- [ ] **Error Handling** - Don't leak sensitive info in errors
- [ ] **Secret Rotation** - Rotate secrets periodically

---

## Secret Rotation

When rotating webhook secrets:

1. Generate new secret in AffTok dashboard
2. Update your server to accept both old and new secrets
3. Wait for all in-flight webhooks to complete
4. Remove old secret from your server
5. Delete old secret from AffTok

```javascript
const secrets = [
  process.env.WEBHOOK_SECRET,
  process.env.WEBHOOK_SECRET_OLD, // During rotation
].filter(Boolean);

function verifyWithMultipleSecrets(payload, signature) {
  return secrets.some(secret => 
    verifyWebhookSignature(payload, signature, secret)
  );
}
```

---

## Next Steps

- [Webhook Management](management.md) - Create and manage webhooks
- [Webhook Testing](testing.md) - Test your endpoint
- [Error Handling](../errors/error-codes.md) - Handle errors

