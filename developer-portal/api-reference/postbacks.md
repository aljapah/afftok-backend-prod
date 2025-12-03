# Postbacks API

Server-to-server postback endpoint for tracking conversions from advertiser systems.

## Overview

Postbacks are the primary method for advertisers to report conversions to AffTok. When a conversion occurs (purchase, signup, install), the advertiser's server sends a postback to AffTok.

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   User Action   │────▶│   Advertiser    │────▶│    AffTok       │
│   (Purchase)    │     │   Server        │     │    Postback     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

---

## Send Postback

Report a conversion to AffTok.

```
POST /api/postback
```

### Headers

```http
Content-Type: application/json
X-API-Key: afftok_live_sk_xxxxx
```

### Request Body

```json
{
  "api_key": "afftok_live_sk_xxxxx",
  "advertiser_id": "adv_123456",
  "offer_id": "off_abc123",
  "transaction_id": "txn_unique_123456",
  "click_id": "clk_abc123def456",
  "amount": 49.99,
  "currency": "USD",
  "payout": 5.00,
  "status": "approved",
  "timestamp": 1699876600000,
  "nonce": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "signature": "hmac_sha256_signature",
  "external_conversion_id": "ext_conv_789",
  "network_id": "net_123",
  "network_name": "Example Network",
  "custom_params": {
    "product_id": "prod_456",
    "category": "electronics",
    "user_type": "new"
  }
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `api_key` | string | Yes | Your API key |
| `advertiser_id` | string | Yes | Advertiser ID |
| `offer_id` | string | Yes | Offer ID |
| `transaction_id` | string | Yes | Unique transaction ID (for deduplication) |
| `click_id` | string | No* | Associated click ID |
| `amount` | number | No | Conversion amount |
| `currency` | string | No | Currency code (default: USD) |
| `payout` | number | No | Affiliate payout amount |
| `status` | string | Yes | Status: `pending`, `approved`, `rejected` |
| `timestamp` | integer | Yes | Unix timestamp (ms) |
| `nonce` | string | Yes | Random 32-char string |
| `signature` | string | Yes | HMAC-SHA256 signature |
| `external_conversion_id` | string | No | External system conversion ID |
| `network_id` | string | No | Network identifier |
| `network_name` | string | No | Network name |
| `custom_params` | object | No | Additional custom data |

> *`click_id` is highly recommended for accurate attribution.

### Response (Success)

```json
{
  "success": true,
  "message": "Conversion tracked successfully",
  "data": {
    "conversion_id": "conv_xyz789abc123",
    "offer_id": "off_abc123",
    "transaction_id": "txn_unique_123456",
    "click_id": "clk_abc123def456",
    "amount": 49.99,
    "payout": 5.00,
    "status": "approved",
    "user_id": "usr_456",
    "converted_at": "2024-01-15T10:35:00Z"
  }
}
```

### Error Responses

**Invalid Signature (403):**

```json
{
  "success": false,
  "error": "Invalid signature",
  "code": "INVALID_SIGNATURE"
}
```

**Duplicate Transaction (409):**

```json
{
  "success": false,
  "error": "Duplicate transaction",
  "code": "DUPLICATE_TRANSACTION",
  "details": {
    "transaction_id": "txn_unique_123456",
    "existing_conversion_id": "conv_existing123"
  }
}
```

**Click Not Found (404):**

```json
{
  "success": false,
  "error": "Click not found",
  "code": "CLICK_NOT_FOUND",
  "details": {
    "click_id": "clk_invalid"
  }
}
```

**Geo Blocked (403):**

```json
{
  "success": false,
  "error": "Country not allowed for postback",
  "code": "GEO_BLOCKED",
  "details": {
    "country": "CN"
  }
}
```

---

## Postback URL Template

When setting up postbacks on advertiser platforms, use this URL template:

```
https://api.afftok.com/api/postback?api_key={api_key}&advertiser_id={advertiser_id}&offer_id={offer_id}&transaction_id={transaction_id}&click_id={click_id}&amount={amount}&status=approved&timestamp={timestamp}&nonce={nonce}&signature={signature}
```

### Common Macros

Map these macros to your advertiser platform's variables:

| AffTok Parameter | Common Macros |
|------------------|---------------|
| `click_id` | `{clickid}`, `{click_id}`, `{aff_sub}`, `{subid}` |
| `transaction_id` | `{order_id}`, `{txn_id}`, `{conversion_id}` |
| `amount` | `{amount}`, `{revenue}`, `{sale_amount}` |
| `payout` | `{payout}`, `{commission}` |

### Example Postback URLs

**HasOffers/TUNE:**
```
https://api.afftok.com/api/postback?click_id={aff_sub}&transaction_id={order_id}&amount={sale_amount}&status=approved&...
```

**Everflow:**
```
https://api.afftok.com/api/postback?click_id={sub1}&transaction_id={transaction_id}&amount={revenue}&status=approved&...
```

**Voluum:**
```
https://api.afftok.com/api/postback?click_id={clickid}&transaction_id={txid}&amount={payout}&status=approved&...
```

---

## Conversion Statuses

| Status | Description | Earnings Impact |
|--------|-------------|-----------------|
| `pending` | Awaiting approval | Not counted |
| `approved` | Confirmed conversion | Added to earnings |
| `rejected` | Declined/reversed | Not counted |

### Update Conversion Status

Update an existing conversion's status:

```
PUT /api/conversions/:id/status
```

```json
{
  "status": "rejected",
  "reason": "Refund requested"
}
```

---

## Signature Generation

All postbacks must include a valid HMAC-SHA256 signature.

### Signature Format

```
data_to_sign = "{api_key}|{advertiser_id}|{timestamp}|{nonce}"
signature = HMAC-SHA256(api_key, data_to_sign)
```

### Implementation Examples

**JavaScript:**

```javascript
const crypto = require('crypto');

function generatePostbackSignature(apiKey, advertiserId) {
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

def generate_postback_signature(api_key: str, advertiser_id: str):
    timestamp = int(time.time() * 1000)
    nonce = secrets.token_hex(16)
    data_to_sign = f"{api_key}|{advertiser_id}|{timestamp}|{nonce}"
    signature = hmac.new(api_key.encode(), data_to_sign.encode(), hashlib.sha256).hexdigest()
    
    return {"timestamp": timestamp, "nonce": nonce, "signature": signature}
```

**PHP:**

```php
function generatePostbackSignature($apiKey, $advertiserId) {
    $timestamp = round(microtime(true) * 1000);
    $nonce = bin2hex(random_bytes(16));
    $dataToSign = "{$apiKey}|{$advertiserId}|{$timestamp}|{$nonce}";
    $signature = hash_hmac('sha256', $dataToSign, $apiKey);
    
    return ['timestamp' => $timestamp, 'nonce' => $nonce, 'signature' => $signature];
}
```

---

## Code Examples

### cURL

```bash
curl -X POST https://api.afftok.com/api/postback \
  -H "Content-Type: application/json" \
  -H "X-API-Key: afftok_live_sk_xxxxx" \
  -d '{
    "api_key": "afftok_live_sk_xxxxx",
    "advertiser_id": "adv_123456",
    "offer_id": "off_abc123",
    "transaction_id": "txn_'$(date +%s)'",
    "click_id": "clk_abc123def456",
    "amount": 49.99,
    "status": "approved",
    "timestamp": '$(date +%s%3N)',
    "nonce": "'$(openssl rand -hex 16)'",
    "signature": "CALCULATED_SIGNATURE"
  }'
```

### JavaScript (Node.js)

```javascript
const crypto = require('crypto');
const axios = require('axios');

async function sendPostback(conversionData) {
  const apiKey = process.env.AFFTOK_API_KEY;
  const advertiserId = process.env.AFFTOK_ADVERTISER_ID;
  const timestamp = Date.now();
  const nonce = crypto.randomBytes(16).toString('hex');
  
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  const signature = crypto.createHmac('sha256', apiKey).update(dataToSign).digest('hex');
  
  const payload = {
    api_key: apiKey,
    advertiser_id: advertiserId,
    offer_id: conversionData.offerId,
    transaction_id: conversionData.transactionId,
    click_id: conversionData.clickId,
    amount: conversionData.amount,
    currency: conversionData.currency || 'USD',
    status: conversionData.status || 'approved',
    timestamp,
    nonce,
    signature,
    custom_params: conversionData.customParams,
  };
  
  try {
    const response = await axios.post('https://api.afftok.com/api/postback', payload, {
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey,
      },
      timeout: 30000,
    });
    
    console.log('Postback sent successfully:', response.data);
    return { success: true, data: response.data };
  } catch (error) {
    console.error('Postback failed:', error.response?.data || error.message);
    return { success: false, error: error.response?.data || error.message };
  }
}

// Usage
sendPostback({
  offerId: 'off_abc123',
  transactionId: `txn_${Date.now()}`,
  clickId: 'clk_abc123def456',
  amount: 49.99,
  status: 'approved',
  customParams: {
    product_id: 'prod_456',
  },
});
```

### Python

```python
import hmac
import hashlib
import time
import secrets
import httpx
import os

async def send_postback(conversion_data: dict):
    api_key = os.environ['AFFTOK_API_KEY']
    advertiser_id = os.environ['AFFTOK_ADVERTISER_ID']
    timestamp = int(time.time() * 1000)
    nonce = secrets.token_hex(16)
    
    data_to_sign = f"{api_key}|{advertiser_id}|{timestamp}|{nonce}"
    signature = hmac.new(api_key.encode(), data_to_sign.encode(), hashlib.sha256).hexdigest()
    
    payload = {
        "api_key": api_key,
        "advertiser_id": advertiser_id,
        "offer_id": conversion_data["offer_id"],
        "transaction_id": conversion_data["transaction_id"],
        "click_id": conversion_data.get("click_id"),
        "amount": conversion_data.get("amount"),
        "currency": conversion_data.get("currency", "USD"),
        "status": conversion_data.get("status", "approved"),
        "timestamp": timestamp,
        "nonce": nonce,
        "signature": signature,
        "custom_params": conversion_data.get("custom_params"),
    }
    
    async with httpx.AsyncClient() as client:
        response = await client.post(
            "https://api.afftok.com/api/postback",
            json=payload,
            headers={"X-API-Key": api_key},
            timeout=30.0,
        )
        return response.json()

# Usage
import asyncio
result = asyncio.run(send_postback({
    "offer_id": "off_abc123",
    "transaction_id": f"txn_{int(time.time())}",
    "click_id": "clk_abc123def456",
    "amount": 49.99,
    "status": "approved",
}))
print(result)
```

### PHP

```php
<?php

function sendPostback($conversionData) {
    $apiKey = getenv('AFFTOK_API_KEY');
    $advertiserId = getenv('AFFTOK_ADVERTISER_ID');
    $timestamp = round(microtime(true) * 1000);
    $nonce = bin2hex(random_bytes(16));
    
    $dataToSign = "{$apiKey}|{$advertiserId}|{$timestamp}|{$nonce}";
    $signature = hash_hmac('sha256', $dataToSign, $apiKey);
    
    $payload = [
        'api_key' => $apiKey,
        'advertiser_id' => $advertiserId,
        'offer_id' => $conversionData['offer_id'],
        'transaction_id' => $conversionData['transaction_id'],
        'click_id' => $conversionData['click_id'] ?? null,
        'amount' => $conversionData['amount'] ?? null,
        'currency' => $conversionData['currency'] ?? 'USD',
        'status' => $conversionData['status'] ?? 'approved',
        'timestamp' => $timestamp,
        'nonce' => $nonce,
        'signature' => $signature,
        'custom_params' => $conversionData['custom_params'] ?? null,
    ];
    
    $ch = curl_init('https://api.afftok.com/api/postback');
    curl_setopt_array($ch, [
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_POST => true,
        CURLOPT_POSTFIELDS => json_encode($payload),
        CURLOPT_HTTPHEADER => [
            'Content-Type: application/json',
            'X-API-Key: ' . $apiKey,
        ],
        CURLOPT_TIMEOUT => 30,
    ]);
    
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    return json_decode($response, true);
}

// Usage
$result = sendPostback([
    'offer_id' => 'off_abc123',
    'transaction_id' => 'txn_' . time(),
    'click_id' => 'clk_abc123def456',
    'amount' => 49.99,
    'status' => 'approved',
]);
print_r($result);
```

---

## Best Practices

1. **Always include click_id** - Ensures accurate attribution
2. **Use unique transaction_ids** - Prevents duplicate conversions
3. **Implement retry logic** - Handle network failures gracefully
4. **Validate response** - Check for success before marking conversion as sent
5. **Log postback attempts** - For debugging and auditing
6. **Set reasonable timeouts** - 30 seconds is recommended

## Next Steps

- [Webhooks](../webhooks/overview.md) - Receive conversion notifications
- [Stats API](stats.md) - View conversion statistics
- [Error Codes](../errors/error-codes.md) - Full error reference

