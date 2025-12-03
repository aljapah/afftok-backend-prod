# Webhooks Overview

AffTok webhooks allow you to receive real-time notifications when events occur in your account.

---

## How Webhooks Work

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│     Event       │────▶│    AffTok       │────▶│   Your Server   │
│   (Click,       │     │    Webhook      │     │   (Endpoint)    │
│   Conversion)   │     │    Engine       │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                        │
                                                        ▼
                                                ┌─────────────────┐
                                                │   Process &     │
                                                │   Respond 200   │
                                                └─────────────────┘
```

When an event occurs (click tracked, conversion recorded, etc.), AffTok sends an HTTP POST request to your configured endpoint with event data.

---

## Supported Events

| Event | Description |
|-------|-------------|
| `click.tracked` | A click was successfully tracked |
| `click.blocked` | A click was blocked (fraud/geo/rate) |
| `conversion.created` | A new conversion was created |
| `conversion.approved` | A conversion was approved |
| `conversion.rejected` | A conversion was rejected |
| `postback.received` | A postback was received |
| `postback.failed` | A postback delivery failed |
| `offer.joined` | A user joined an offer |
| `offer.left` | A user left an offer |
| `fraud.detected` | Fraud was detected |
| `api_key.used` | An API key was used |
| `api_key.rate_limited` | An API key was rate limited |

---

## Webhook Payload Structure

All webhooks follow this structure:

```json
{
  "id": "evt_abc123def456",
  "type": "conversion.created",
  "timestamp": "2024-01-15T10:30:45Z",
  "api_version": "2024-01-01",
  "data": {
    // Event-specific data
  },
  "metadata": {
    "correlation_id": "corr_xyz789",
    "tenant_id": "tenant_abc123"
  }
}
```

### Common Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique event ID |
| `type` | string | Event type |
| `timestamp` | string | ISO 8601 timestamp |
| `api_version` | string | API version |
| `data` | object | Event-specific data |
| `metadata` | object | Additional context |

---

## Event Payloads

### Click Tracked

```json
{
  "id": "evt_abc123",
  "type": "click.tracked",
  "timestamp": "2024-01-15T10:30:45Z",
  "data": {
    "click_id": "clk_xyz789",
    "tracking_code": "tc_abc123",
    "offer_id": "off_def456",
    "user_id": "usr_ghi789",
    "ip_address": "xxx.xxx.xxx.xxx",
    "country": "US",
    "city": "New York",
    "device": "mobile",
    "browser": "Chrome",
    "os": "iOS",
    "referer": "https://facebook.com",
    "user_agent": "Mozilla/5.0...",
    "sub_id": "campaign_123"
  }
}
```

### Conversion Created

```json
{
  "id": "evt_def456",
  "type": "conversion.created",
  "timestamp": "2024-01-15T10:35:00Z",
  "data": {
    "conversion_id": "conv_abc123",
    "click_id": "clk_xyz789",
    "offer_id": "off_def456",
    "user_id": "usr_ghi789",
    "transaction_id": "txn_123456",
    "amount": 49.99,
    "currency": "USD",
    "payout": 5.00,
    "status": "pending",
    "network_id": "net_abc",
    "network_name": "Example Network",
    "custom_params": {
      "product_id": "prod_456",
      "category": "electronics"
    }
  }
}
```

### Conversion Approved

```json
{
  "id": "evt_ghi789",
  "type": "conversion.approved",
  "timestamp": "2024-01-15T11:00:00Z",
  "data": {
    "conversion_id": "conv_abc123",
    "offer_id": "off_def456",
    "user_id": "usr_ghi789",
    "payout": 5.00,
    "approved_at": "2024-01-15T11:00:00Z",
    "approved_by": "auto"
  }
}
```

### Fraud Detected

```json
{
  "id": "evt_jkl012",
  "type": "fraud.detected",
  "timestamp": "2024-01-15T10:30:45Z",
  "data": {
    "event_type": "click",
    "reason": "bot_detection",
    "risk_score": 92,
    "ip_address": "xxx.xxx.xxx.xxx",
    "country": "CN",
    "indicators": ["bad_user_agent", "datacenter_ip"],
    "blocked": true,
    "offer_id": "off_def456",
    "tracking_code": "tc_abc123"
  }
}
```

---

## Delivery Guarantees

AffTok webhooks provide:

- **At-least-once delivery** - Events may be delivered multiple times
- **Ordered delivery** - Events are delivered in order per resource
- **Retry logic** - Failed deliveries are retried automatically

### Retry Policy

| Attempt | Delay |
|---------|-------|
| 1 | Immediate |
| 2 | 5 seconds |
| 3 | 10 seconds |
| 4 | 30 seconds |
| 5 | 1 minute |
| 6 | 5 minutes |
| 7 | 30 minutes |
| 8 | 1 hour |

After 8 failed attempts, the event is moved to the Dead Letter Queue (DLQ).

---

## Handling Webhooks

### Basic Handler (Node.js)

```javascript
const express = require('express');
const crypto = require('crypto');

const app = express();
app.use(express.json());

app.post('/webhooks/afftok', (req, res) => {
  // Verify signature
  const signature = req.headers['x-afftok-signature'];
  const payload = JSON.stringify(req.body);
  const expectedSignature = crypto
    .createHmac('sha256', process.env.WEBHOOK_SECRET)
    .update(payload)
    .digest('hex');

  if (signature !== expectedSignature) {
    return res.status(401).json({ error: 'Invalid signature' });
  }

  // Process event
  const event = req.body;
  
  switch (event.type) {
    case 'conversion.created':
      handleConversionCreated(event.data);
      break;
    case 'conversion.approved':
      handleConversionApproved(event.data);
      break;
    case 'fraud.detected':
      handleFraudDetected(event.data);
      break;
    default:
      console.log('Unhandled event type:', event.type);
  }

  // Respond quickly
  res.status(200).json({ received: true });
});

function handleConversionCreated(data) {
  console.log('New conversion:', data.conversion_id);
  // Update your database, send notifications, etc.
}

function handleConversionApproved(data) {
  console.log('Conversion approved:', data.conversion_id);
  // Credit user account, etc.
}

function handleFraudDetected(data) {
  console.log('Fraud detected:', data.reason);
  // Log fraud event, alert team, etc.
}

app.listen(3000);
```

### Python (Flask)

```python
from flask import Flask, request, jsonify
import hmac
import hashlib
import os

app = Flask(__name__)

@app.route('/webhooks/afftok', methods=['POST'])
def handle_webhook():
    # Verify signature
    signature = request.headers.get('X-Afftok-Signature')
    payload = request.get_data()
    expected_signature = hmac.new(
        os.environ['WEBHOOK_SECRET'].encode(),
        payload,
        hashlib.sha256
    ).hexdigest()

    if signature != expected_signature:
        return jsonify({'error': 'Invalid signature'}), 401

    # Process event
    event = request.json
    
    if event['type'] == 'conversion.created':
        handle_conversion_created(event['data'])
    elif event['type'] == 'conversion.approved':
        handle_conversion_approved(event['data'])
    elif event['type'] == 'fraud.detected':
        handle_fraud_detected(event['data'])
    
    return jsonify({'received': True}), 200

def handle_conversion_created(data):
    print(f"New conversion: {data['conversion_id']}")

def handle_conversion_approved(data):
    print(f"Conversion approved: {data['conversion_id']}")

def handle_fraud_detected(data):
    print(f"Fraud detected: {data['reason']}")

if __name__ == '__main__':
    app.run(port=3000)
```

### PHP

```php
<?php

// Verify signature
$signature = $_SERVER['HTTP_X_AFFTOK_SIGNATURE'] ?? '';
$payload = file_get_contents('php://input');
$expectedSignature = hash_hmac('sha256', $payload, getenv('WEBHOOK_SECRET'));

if (!hash_equals($expectedSignature, $signature)) {
    http_response_code(401);
    echo json_encode(['error' => 'Invalid signature']);
    exit;
}

// Process event
$event = json_decode($payload, true);

switch ($event['type']) {
    case 'conversion.created':
        handleConversionCreated($event['data']);
        break;
    case 'conversion.approved':
        handleConversionApproved($event['data']);
        break;
    case 'fraud.detected':
        handleFraudDetected($event['data']);
        break;
}

http_response_code(200);
echo json_encode(['received' => true]);

function handleConversionCreated($data) {
    error_log("New conversion: " . $data['conversion_id']);
}

function handleConversionApproved($data) {
    error_log("Conversion approved: " . $data['conversion_id']);
}

function handleFraudDetected($data) {
    error_log("Fraud detected: " . $data['reason']);
}
```

---

## Best Practices

### 1. Respond Quickly

Return a 2xx response within 5 seconds. Process events asynchronously.

```javascript
app.post('/webhooks/afftok', async (req, res) => {
  // Respond immediately
  res.status(200).json({ received: true });
  
  // Process asynchronously
  setImmediate(() => {
    processEvent(req.body);
  });
});
```

### 2. Handle Duplicates

Use the event `id` to deduplicate:

```javascript
const processedEvents = new Set();

function processEvent(event) {
  if (processedEvents.has(event.id)) {
    console.log('Duplicate event, skipping:', event.id);
    return;
  }
  
  processedEvents.add(event.id);
  // Process event...
}
```

### 3. Verify Signatures

Always verify webhook signatures to ensure authenticity.

### 4. Use HTTPS

Only use HTTPS endpoints for webhooks.

### 5. Log Everything

Log all webhook events for debugging and auditing.

---

## Next Steps

- [Webhook Security](security.md) - Signature verification
- [Webhook Management](management.md) - Create and manage webhooks
- [Webhook Testing](testing.md) - Test your webhook endpoint

