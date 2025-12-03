# Postback API

Send conversion data to AffTok when users complete actions on your platform.

## Overview

Postbacks are server-to-server notifications that inform AffTok when a conversion occurs. This enables accurate attribution and commission calculation.

---

## POST /api/postback

Record a conversion/postback.

### Request

```http
POST /api/postback
Content-Type: application/json
X-API-Key: afftok_live_sk_xxxxx
```

### Basic Request

```json
{
  "click_id": "clk_a1b2c3d4e5f6",
  "transaction_id": "txn_your_unique_id_12345",
  "amount": 49.99,
  "currency": "USD",
  "status": "approved"
}
```

### Signed Request (Recommended)

```json
{
  "api_key": "afftok_live_sk_xxxxx",
  "advertiser_id": "adv_123456",
  "click_id": "clk_a1b2c3d4e5f6",
  "transaction_id": "txn_your_unique_id_12345",
  "amount": 49.99,
  "currency": "USD",
  "status": "approved",
  "timestamp": 1701234567890,
  "nonce": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "signature": "9f8e7d6c5b4a3210fedcba9876543210"
}
```

### Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `click_id` | string | Yes* | Original click ID |
| `transaction_id` | string | Yes | Your unique transaction ID |
| `amount` | number | No | Conversion amount |
| `currency` | string | No | ISO 4217 currency code (default: USD) |
| `status` | string | No | `pending`, `approved`, `rejected` (default: approved) |
| `api_key` | string | Yes** | Your API key |
| `advertiser_id` | string | Yes** | Your advertiser ID |
| `timestamp` | integer | Yes** | Unix timestamp in milliseconds |
| `nonce` | string | Yes** | Random 32-character string |
| `signature` | string | Yes** | HMAC-SHA256 signature |
| `custom_params` | object | No | Additional custom data |

*Either `click_id` or tracking parameters required
**Required for signed requests

### Response

```json
{
  "success": true,
  "message": "Conversion recorded successfully",
  "data": {
    "conversion_id": "conv_xyz789abc",
    "click_id": "clk_a1b2c3d4e5f6",
    "offer_id": "off_123456",
    "transaction_id": "txn_your_unique_id_12345",
    "amount": 49.99,
    "currency": "USD",
    "payout": 9.99,
    "status": "approved",
    "created_at": "2024-12-01T12:30:00Z"
  }
}
```

### Errors

| Code | Status | Description |
|------|--------|-------------|
| `INVALID_API_KEY` | 401 | API key not found or revoked |
| `INVALID_SIGNATURE` | 403 | Signature verification failed |
| `CLICK_NOT_FOUND` | 404 | Click ID doesn't exist |
| `DUPLICATE_TRANSACTION` | 409 | Transaction ID already processed |
| `EXPIRED_CLICK` | 400 | Click is outside attribution window |
| `GEO_BLOCKED` | 403 | Country blocked by geo rules |
| `RATE_LIMITED` | 429 | Too many requests |

---

## Signature Generation

### Algorithm

```
data_to_sign = "{api_key}|{advertiser_id}|{timestamp}|{nonce}"
signature = HMAC-SHA256(api_key, data_to_sign)
```

### JavaScript Example

```javascript
const crypto = require('crypto');

function generateSignature(apiKey, advertiserId) {
  const timestamp = Date.now();
  const nonce = crypto.randomBytes(16).toString('hex');
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  const signature = crypto
    .createHmac('sha256', apiKey)
    .update(dataToSign)
    .digest('hex');
  
  return { timestamp, nonce, signature };
}
```

### Python Example

```python
import hmac
import hashlib
import time
import secrets

def generate_signature(api_key, advertiser_id):
    timestamp = int(time.time() * 1000)
    nonce = secrets.token_hex(16)
    data_to_sign = f"{api_key}|{advertiser_id}|{timestamp}|{nonce}"
    signature = hmac.new(
        api_key.encode(),
        data_to_sign.encode(),
        hashlib.sha256
    ).hexdigest()
    
    return timestamp, nonce, signature
```

### PHP Example

```php
function generateSignature($apiKey, $advertiserId) {
    $timestamp = round(microtime(true) * 1000);
    $nonce = bin2hex(random_bytes(16));
    $dataToSign = "{$apiKey}|{$advertiserId}|{$timestamp}|{$nonce}";
    $signature = hash_hmac('sha256', $dataToSign, $apiKey);
    
    return [
        'timestamp' => $timestamp,
        'nonce' => $nonce,
        'signature' => $signature
    ];
}
```

### Go Example

```go
import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

func generateSignature(apiKey, advertiserId string) (int64, string, string) {
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

---

## Complete Examples

### cURL

```bash
#!/bin/bash

API_KEY="afftok_live_sk_xxxxx"
ADVERTISER_ID="adv_123456"
TIMESTAMP=$(date +%s000)
NONCE=$(openssl rand -hex 16)
DATA_TO_SIGN="${API_KEY}|${ADVERTISER_ID}|${TIMESTAMP}|${NONCE}"
SIGNATURE=$(echo -n "$DATA_TO_SIGN" | openssl dgst -sha256 -hmac "$API_KEY" | cut -d' ' -f2)

curl -X POST https://api.afftok.com/api/postback \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "api_key": "'$API_KEY'",
    "advertiser_id": "'$ADVERTISER_ID'",
    "click_id": "clk_a1b2c3d4e5f6",
    "transaction_id": "txn_'$(date +%s)'",
    "amount": 49.99,
    "currency": "USD",
    "status": "approved",
    "timestamp": '$TIMESTAMP',
    "nonce": "'$NONCE'",
    "signature": "'$SIGNATURE'"
  }'
```

### Node.js

```javascript
const crypto = require('crypto');
const axios = require('axios');

async function sendPostback(clickId, transactionId, amount) {
  const apiKey = process.env.AFFTOK_API_KEY;
  const advertiserId = process.env.AFFTOK_ADVERTISER_ID;
  const timestamp = Date.now();
  const nonce = crypto.randomBytes(16).toString('hex');
  
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  const signature = crypto
    .createHmac('sha256', apiKey)
    .update(dataToSign)
    .digest('hex');

  const payload = {
    api_key: apiKey,
    advertiser_id: advertiserId,
    click_id: clickId,
    transaction_id: transactionId,
    amount: amount,
    currency: 'USD',
    status: 'approved',
    timestamp: timestamp,
    nonce: nonce,
    signature: signature
  };

  try {
    const response = await axios.post(
      'https://api.afftok.com/api/postback',
      payload,
      {
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': apiKey
        }
      }
    );
    console.log('Postback sent:', response.data);
    return response.data;
  } catch (error) {
    console.error('Postback failed:', error.response?.data);
    throw error;
  }
}

// Usage
sendPostback('clk_a1b2c3d4e5f6', 'txn_12345', 49.99);
```

### Python

```python
import hmac
import hashlib
import time
import secrets
import requests

def send_postback(click_id, transaction_id, amount):
    api_key = os.environ['AFFTOK_API_KEY']
    advertiser_id = os.environ['AFFTOK_ADVERTISER_ID']
    timestamp = int(time.time() * 1000)
    nonce = secrets.token_hex(16)
    
    data_to_sign = f"{api_key}|{advertiser_id}|{timestamp}|{nonce}"
    signature = hmac.new(
        api_key.encode(),
        data_to_sign.encode(),
        hashlib.sha256
    ).hexdigest()
    
    payload = {
        'api_key': api_key,
        'advertiser_id': advertiser_id,
        'click_id': click_id,
        'transaction_id': transaction_id,
        'amount': amount,
        'currency': 'USD',
        'status': 'approved',
        'timestamp': timestamp,
        'nonce': nonce,
        'signature': signature
    }
    
    response = requests.post(
        'https://api.afftok.com/api/postback',
        json=payload,
        headers={
            'Content-Type': 'application/json',
            'X-API-Key': api_key
        }
    )
    
    return response.json()

# Usage
result = send_postback('clk_a1b2c3d4e5f6', 'txn_12345', 49.99)
print(result)
```

### PHP

```php
<?php

function sendPostback($clickId, $transactionId, $amount) {
    $apiKey = getenv('AFFTOK_API_KEY');
    $advertiserId = getenv('AFFTOK_ADVERTISER_ID');
    $timestamp = round(microtime(true) * 1000);
    $nonce = bin2hex(random_bytes(16));
    
    $dataToSign = "{$apiKey}|{$advertiserId}|{$timestamp}|{$nonce}";
    $signature = hash_hmac('sha256', $dataToSign, $apiKey);
    
    $payload = [
        'api_key' => $apiKey,
        'advertiser_id' => $advertiserId,
        'click_id' => $clickId,
        'transaction_id' => $transactionId,
        'amount' => $amount,
        'currency' => 'USD',
        'status' => 'approved',
        'timestamp' => $timestamp,
        'nonce' => $nonce,
        'signature' => $signature
    ];
    
    $ch = curl_init('https://api.afftok.com/api/postback');
    curl_setopt_array($ch, [
        CURLOPT_POST => true,
        CURLOPT_POSTFIELDS => json_encode($payload),
        CURLOPT_HTTPHEADER => [
            'Content-Type: application/json',
            'X-API-Key: ' . $apiKey
        ],
        CURLOPT_RETURNTRANSFER => true
    ]);
    
    $response = curl_exec($ch);
    curl_close($ch);
    
    return json_decode($response, true);
}

// Usage
$result = sendPostback('clk_a1b2c3d4e5f6', 'txn_12345', 49.99);
print_r($result);
```

---

## Postback URL Templates

For advertisers who prefer URL-based postbacks:

### Template Format

```
https://api.afftok.com/api/postback/url
  ?click_id={click_id}
  &transaction_id={transaction_id}
  &amount={amount}
  &currency={currency}
  &status={status}
  &api_key={api_key}
```

### Placeholders

| Placeholder | Description |
|-------------|-------------|
| `{click_id}` | AffTok click ID |
| `{transaction_id}` | Your transaction ID |
| `{amount}` | Conversion amount |
| `{currency}` | Currency code |
| `{status}` | Conversion status |

---

## Conversion Statuses

| Status | Description | Commission |
|--------|-------------|------------|
| `pending` | Awaiting verification | Held |
| `approved` | Verified and approved | Paid |
| `rejected` | Rejected/refunded | Not paid |

### Updating Status

```http
PUT /api/postback/{conversion_id}/status
Content-Type: application/json
X-API-Key: afftok_live_sk_xxxxx

{
  "status": "rejected",
  "reason": "Refund requested"
}
```

---

## Best Practices

1. **Use unique transaction IDs** - Prevents duplicate conversions
2. **Implement retry logic** - Handle temporary failures
3. **Sign all requests** - Enhanced security
4. **Send postbacks promptly** - Within attribution window
5. **Include all available data** - Better analytics
6. **Handle errors gracefully** - Log and alert on failures

---

## Retry Strategy

Implement exponential backoff for failed postbacks:

```javascript
async function sendWithRetry(payload, maxRetries = 5) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await sendPostback(payload);
    } catch (error) {
      if (error.response?.status === 409) {
        // Duplicate - don't retry
        return { duplicate: true };
      }
      
      const delay = Math.min(1000 * Math.pow(2, attempt), 300000);
      await sleep(delay);
    }
  }
  throw new Error('Max retries exceeded');
}
```

---

Next: [Stats API](./stats.md)

