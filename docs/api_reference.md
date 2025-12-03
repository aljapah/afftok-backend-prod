# AffTok API Reference

Complete API reference for AffTok tracking endpoints.

## Base URL

```
https://api.afftok.com
```

## Authentication

All requests require authentication via API key:

```http
X-API-Key: your_api_key
```

Additionally, requests must include HMAC-SHA256 signature for verification.

---

## Endpoints

### POST /api/sdk/click

Track a click event.

**Headers:**
```http
Content-Type: application/json
X-API-Key: your_api_key
X-SDK-Version: 1.0.0
X-SDK-Platform: android|ios|flutter|react-native|web
```

**Request Body:**
```json
{
  "api_key": "your_api_key",
  "advertiser_id": "your_advertiser_id",
  "offer_id": "offer_123",
  "timestamp": 1234567890000,
  "nonce": "random_32_char_string",
  "signature": "hmac_sha256_signature",
  "tracking_code": "campaign_code",
  "sub_id_1": "source",
  "sub_id_2": "medium",
  "sub_id_3": "campaign",
  "user_id": "user_123",
  "device_info": {
    "device_id": "uuid",
    "fingerprint": "sha256_hash",
    "platform": "android",
    "os_version": "13",
    "model": "Pixel 7"
  },
  "custom_params": {
    "key": "value"
  }
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Click tracked successfully",
  "data": {
    "click_id": "click_abc123",
    "offer_id": "offer_123",
    "tracking_code": "campaign_code"
  }
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": "Invalid signature",
  "code": "INVALID_SIGNATURE"
}
```

---

### POST /api/sdk/conversion

Track a conversion event.

**Headers:**
```http
Content-Type: application/json
X-API-Key: your_api_key
X-SDK-Version: 1.0.0
X-SDK-Platform: android|ios|flutter|react-native|web
```

**Request Body:**
```json
{
  "api_key": "your_api_key",
  "advertiser_id": "your_advertiser_id",
  "offer_id": "offer_123",
  "transaction_id": "txn_abc123",
  "click_id": "click_xyz",
  "amount": 29.99,
  "currency": "USD",
  "status": "approved",
  "timestamp": 1234567890000,
  "nonce": "random_32_char_string",
  "signature": "hmac_sha256_signature",
  "user_id": "user_123",
  "device_info": {
    "device_id": "uuid",
    "fingerprint": "sha256_hash"
  },
  "custom_params": {
    "product_id": "prod_456"
  },
  "metadata": {
    "category": "electronics"
  }
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Conversion tracked successfully",
  "data": {
    "conversion_id": "conv_abc123",
    "offer_id": "offer_123",
    "transaction_id": "txn_abc123",
    "amount": 29.99,
    "currency": "USD",
    "status": "approved"
  }
}
```

---

### POST /api/postback

Server-to-server postback endpoint.

**Headers:**
```http
Content-Type: application/json
X-API-Key: your_api_key
```

**Request Body:**
```json
{
  "api_key": "your_api_key",
  "advertiser_id": "your_advertiser_id",
  "offer_id": "offer_123",
  "transaction_id": "txn_abc123",
  "click_id": "click_xyz",
  "amount": 29.99,
  "currency": "USD",
  "status": "approved",
  "timestamp": 1234567890000,
  "nonce": "random_32_char_string",
  "signature": "hmac_sha256_signature",
  "custom_params": {
    "product_id": "prod_456"
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Postback received",
  "data": {
    "conversion_id": "conv_abc123"
  }
}
```

---

### GET /api/c/:trackingCode

Redirect endpoint for tracking links.

**URL Parameters:**
- `trackingCode`: The signed tracking code

**Query Parameters:**
- `sub1`, `sub2`, `sub3`: Optional sub IDs

**Response:**
- 302 Redirect to destination URL

---

## Signature Generation

All requests must include an HMAC-SHA256 signature.

### Format

```
signature = HMAC-SHA256(api_key, data_to_sign)
data_to_sign = "{api_key}|{advertiser_id}|{timestamp}|{nonce}"
```

### Example (JavaScript)

```javascript
const crypto = require('crypto');

function generateSignature(apiKey, advertiserId, timestamp, nonce) {
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  return crypto.createHmac('sha256', apiKey).update(dataToSign).digest('hex');
}
```

### Example (Python)

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

---

## Request Parameters

### Required Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `api_key` | string | Your API key |
| `advertiser_id` | string | Your advertiser ID |
| `offer_id` | string | The offer ID |
| `timestamp` | integer | Unix timestamp in milliseconds |
| `nonce` | string | Random 32-character string |
| `signature` | string | HMAC-SHA256 signature |

### Optional Parameters (Click)

| Parameter | Type | Description |
|-----------|------|-------------|
| `tracking_code` | string | Campaign tracking code |
| `sub_id_1` | string | Sub ID 1 (source) |
| `sub_id_2` | string | Sub ID 2 (medium) |
| `sub_id_3` | string | Sub ID 3 (campaign) |
| `user_id` | string | User identifier |
| `device_info` | object | Device information |
| `custom_params` | object | Custom parameters |

### Optional Parameters (Conversion)

| Parameter | Type | Description |
|-----------|------|-------------|
| `transaction_id` | string | Unique transaction ID |
| `click_id` | string | Associated click ID |
| `amount` | number | Conversion amount |
| `currency` | string | Currency code (default: USD) |
| `status` | string | pending, approved, rejected |
| `user_id` | string | User identifier |
| `device_info` | object | Device information |
| `custom_params` | object | Custom parameters |
| `metadata` | object | Additional metadata |

---

## HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Bad Request - Invalid parameters |
| 401 | Unauthorized - Invalid API key |
| 403 | Forbidden - Invalid signature |
| 404 | Not Found - Offer not found |
| 409 | Conflict - Duplicate transaction |
| 429 | Too Many Requests - Rate limited |
| 500 | Internal Server Error |

---

## Error Codes

| Code | Description |
|------|-------------|
| `INVALID_API_KEY` | API key not found or revoked |
| `INVALID_SIGNATURE` | HMAC signature verification failed |
| `EXPIRED_REQUEST` | Request timestamp too old |
| `INVALID_OFFER` | Offer ID not found or inactive |
| `DUPLICATE_TRANSACTION` | Transaction ID already processed |
| `RATE_LIMITED` | Too many requests |
| `GEO_BLOCKED` | Country not allowed |
| `INVALID_PARAMS` | Missing or invalid parameters |

---

## Rate Limits

| Limit | Value |
|-------|-------|
| Requests per minute | 60 |
| Requests per hour | 1000 |
| Batch size | 100 |

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1234567890
```

---

## Webhooks

AffTok can send webhooks for conversion events.

### Webhook Payload

```json
{
  "event": "conversion.created",
  "timestamp": 1234567890000,
  "data": {
    "conversion_id": "conv_abc123",
    "offer_id": "offer_123",
    "transaction_id": "txn_abc123",
    "amount": 29.99,
    "currency": "USD",
    "status": "approved"
  },
  "signature": "hmac_sha256_signature"
}
```

### Verifying Webhooks

```javascript
function verifyWebhook(payload, signature, webhookSecret) {
  const expectedSignature = crypto
    .createHmac('sha256', webhookSecret)
    .update(JSON.stringify(payload))
    .digest('hex');
  return signature === expectedSignature;
}
```

---

## SDKs

Official SDKs are available for:

- [Android SDK](https://github.com/afftok/android-sdk)
- [iOS SDK](https://github.com/afftok/ios-sdk)
- [Flutter SDK](https://github.com/afftok/flutter-sdk)
- [React Native SDK](https://github.com/afftok/react-native-sdk)
- [Web SDK](https://github.com/afftok/web-sdk)

---

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- Status: https://status.afftok.com

