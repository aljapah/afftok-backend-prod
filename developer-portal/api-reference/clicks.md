# Clicks API

Track click events through the AffTok API.

## Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/c/:signedCode` | Redirect endpoint for tracking links |
| POST | `/api/sdk/click` | SDK click tracking |
| GET | `/api/clicks/me` | Get user's clicks |
| GET | `/api/clicks/by-offer` | Get clicks by offer |

---

## Track Click (Redirect)

Redirect endpoint for tracking links. Used when users click on tracking links.

```
GET /api/c/:signedCode
```

### URL Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `signedCode` | string | Signed tracking code (format: `code.timestamp.nonce.signature`) |

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `sub1` | string | Sub ID 1 (source) |
| `sub2` | string | Sub ID 2 (medium) |
| `sub3` | string | Sub ID 3 (campaign) |
| `sub4` | string | Sub ID 4 (custom) |
| `sub5` | string | Sub ID 5 (custom) |

### Example Request

```
GET /api/c/abc123.1699876543210.nonce456.sig789?sub1=facebook&sub2=banner1
```

### Response

- **302 Redirect** to destination URL on success
- Click is tracked asynchronously

### Response Headers

```http
Location: https://example.com/landing?click_id=clk_abc123
X-AffTok-Click-ID: clk_abc123
```

---

## Track Click (SDK)

Track clicks via SDK or server-to-server integration.

```
POST /api/sdk/click
```

### Headers

```http
Content-Type: application/json
X-API-Key: afftok_live_sk_xxxxx
X-SDK-Version: 1.0.0
X-SDK-Platform: android|ios|flutter|react-native|web
```

### Request Body

```json
{
  "api_key": "afftok_live_sk_xxxxx",
  "advertiser_id": "adv_123456",
  "offer_id": "off_abc123",
  "timestamp": 1699876543210,
  "nonce": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "signature": "hmac_sha256_signature",
  "tracking_code": "campaign_summer",
  "sub_id_1": "facebook",
  "sub_id_2": "banner_1",
  "sub_id_3": "mobile",
  "user_id": "usr_optional",
  "device_info": {
    "device_id": "uuid-device-id",
    "fingerprint": "sha256_fingerprint",
    "platform": "android",
    "sdk_version": "1.0.0",
    "os_version": "13",
    "model": "Pixel 7",
    "manufacturer": "Google"
  },
  "custom_params": {
    "campaign_id": "camp_123",
    "creative_id": "cre_456"
  }
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `api_key` | string | Yes | Your API key |
| `advertiser_id` | string | Yes | Advertiser ID |
| `offer_id` | string | Yes | Offer ID |
| `timestamp` | integer | Yes | Unix timestamp (ms) |
| `nonce` | string | Yes | Random 32-char string |
| `signature` | string | Yes | HMAC-SHA256 signature |
| `tracking_code` | string | No | Campaign tracking code |
| `sub_id_1` | string | No | Sub ID 1 |
| `sub_id_2` | string | No | Sub ID 2 |
| `sub_id_3` | string | No | Sub ID 3 |
| `user_id` | string | No | Optional user identifier |
| `device_info` | object | No | Device information |
| `custom_params` | object | No | Custom parameters |

### Response

```json
{
  "success": true,
  "message": "Click tracked successfully",
  "data": {
    "click_id": "clk_abc123def456",
    "offer_id": "off_abc123",
    "tracking_code": "campaign_summer",
    "destination_url": "https://example.com/landing",
    "tracked_at": "2024-01-15T10:30:00Z"
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

**Offer Not Found (404):**

```json
{
  "success": false,
  "error": "Offer not found",
  "code": "INVALID_OFFER",
  "details": {
    "offer_id": "off_invalid"
  }
}
```

**Geo Blocked (403):**

```json
{
  "success": false,
  "error": "Country not allowed",
  "code": "GEO_BLOCKED",
  "details": {
    "country": "CN",
    "rule_id": "rule_123"
  }
}
```

---

## Get My Clicks

Retrieve clicks for the authenticated user.

```
GET /api/clicks/me
```

### Headers

```http
Authorization: Bearer JWT_TOKEN
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `limit` | integer | 20 | Items per page (max: 100) |
| `offer_id` | string | - | Filter by offer |
| `start_date` | string | - | Start date (ISO 8601) |
| `end_date` | string | - | End date (ISO 8601) |

### Example Request

```bash
curl -X GET "https://api.afftok.com/api/clicks/me?page=1&limit=20&offer_id=off_123" \
  -H "Authorization: Bearer JWT_TOKEN"
```

### Response

```json
{
  "success": true,
  "data": {
    "clicks": [
      {
        "id": "clk_abc123",
        "offer_id": "off_123",
        "offer_title": "Premium Subscription",
        "tracking_code": "campaign_summer",
        "ip_address": "203.0.113.45",
        "country": "US",
        "city": "New York",
        "device": "mobile",
        "browser": "Safari",
        "os": "iOS",
        "sub_ids": {
          "sub1": "facebook",
          "sub2": "banner_1"
        },
        "converted": true,
        "conversion_id": "conv_xyz789",
        "clicked_at": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "total_pages": 8
    }
  }
}
```

---

## Get Clicks by Offer

Retrieve click statistics grouped by offer.

```
GET /api/clicks/by-offer
```

### Headers

```http
Authorization: Bearer JWT_TOKEN
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `start_date` | string | 7 days ago | Start date |
| `end_date` | string | now | End date |

### Example Request

```bash
curl -X GET "https://api.afftok.com/api/clicks/by-offer?start_date=2024-01-01&end_date=2024-01-31" \
  -H "Authorization: Bearer JWT_TOKEN"
```

### Response

```json
{
  "success": true,
  "data": {
    "offers": [
      {
        "offer_id": "off_123",
        "offer_title": "Premium Subscription",
        "total_clicks": 1500,
        "unique_clicks": 1200,
        "conversions": 45,
        "conversion_rate": 3.75,
        "earnings": 225.00
      },
      {
        "offer_id": "off_456",
        "offer_title": "Free Trial",
        "total_clicks": 800,
        "unique_clicks": 650,
        "conversions": 32,
        "conversion_rate": 4.92,
        "earnings": 160.00
      }
    ],
    "summary": {
      "total_clicks": 2300,
      "total_conversions": 77,
      "total_earnings": 385.00,
      "avg_conversion_rate": 3.35
    }
  }
}
```

---

## Code Examples

### cURL

```bash
# Track click via SDK endpoint
curl -X POST https://api.afftok.com/api/sdk/click \
  -H "Content-Type: application/json" \
  -H "X-API-Key: afftok_live_sk_xxxxx" \
  -d '{
    "api_key": "afftok_live_sk_xxxxx",
    "advertiser_id": "adv_123456",
    "offer_id": "off_abc123",
    "timestamp": 1699876543210,
    "nonce": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
    "signature": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  }'
```

### JavaScript

```javascript
const crypto = require('crypto');
const axios = require('axios');

async function trackClick(offerId, subIds = {}) {
  const apiKey = process.env.AFFTOK_API_KEY;
  const advertiserId = process.env.AFFTOK_ADVERTISER_ID;
  const timestamp = Date.now();
  const nonce = crypto.randomBytes(16).toString('hex');
  
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  const signature = crypto.createHmac('sha256', apiKey).update(dataToSign).digest('hex');
  
  const response = await axios.post('https://api.afftok.com/api/sdk/click', {
    api_key: apiKey,
    advertiser_id: advertiserId,
    offer_id: offerId,
    timestamp,
    nonce,
    signature,
    sub_id_1: subIds.sub1,
    sub_id_2: subIds.sub2,
    sub_id_3: subIds.sub3,
  }, {
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': apiKey,
    },
  });
  
  return response.data;
}

// Usage
trackClick('off_abc123', { sub1: 'facebook', sub2: 'banner_1' })
  .then(console.log)
  .catch(console.error);
```

### Python

```python
import hmac
import hashlib
import time
import secrets
import httpx

async def track_click(offer_id: str, sub_ids: dict = None):
    api_key = os.environ['AFFTOK_API_KEY']
    advertiser_id = os.environ['AFFTOK_ADVERTISER_ID']
    timestamp = int(time.time() * 1000)
    nonce = secrets.token_hex(16)
    
    data_to_sign = f"{api_key}|{advertiser_id}|{timestamp}|{nonce}"
    signature = hmac.new(api_key.encode(), data_to_sign.encode(), hashlib.sha256).hexdigest()
    
    payload = {
        "api_key": api_key,
        "advertiser_id": advertiser_id,
        "offer_id": offer_id,
        "timestamp": timestamp,
        "nonce": nonce,
        "signature": signature,
    }
    
    if sub_ids:
        payload.update({
            "sub_id_1": sub_ids.get("sub1"),
            "sub_id_2": sub_ids.get("sub2"),
            "sub_id_3": sub_ids.get("sub3"),
        })
    
    async with httpx.AsyncClient() as client:
        response = await client.post(
            "https://api.afftok.com/api/sdk/click",
            json=payload,
            headers={"X-API-Key": api_key}
        )
        return response.json()

# Usage
import asyncio
result = asyncio.run(track_click("off_abc123", {"sub1": "facebook"}))
print(result)
```

### PHP

```php
<?php

function trackClick($offerId, $subIds = []) {
    $apiKey = getenv('AFFTOK_API_KEY');
    $advertiserId = getenv('AFFTOK_ADVERTISER_ID');
    $timestamp = round(microtime(true) * 1000);
    $nonce = bin2hex(random_bytes(16));
    
    $dataToSign = "{$apiKey}|{$advertiserId}|{$timestamp}|{$nonce}";
    $signature = hash_hmac('sha256', $dataToSign, $apiKey);
    
    $payload = [
        'api_key' => $apiKey,
        'advertiser_id' => $advertiserId,
        'offer_id' => $offerId,
        'timestamp' => $timestamp,
        'nonce' => $nonce,
        'signature' => $signature,
        'sub_id_1' => $subIds['sub1'] ?? null,
        'sub_id_2' => $subIds['sub2'] ?? null,
        'sub_id_3' => $subIds['sub3'] ?? null,
    ];
    
    $ch = curl_init('https://api.afftok.com/api/sdk/click');
    curl_setopt_array($ch, [
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_POST => true,
        CURLOPT_POSTFIELDS => json_encode($payload),
        CURLOPT_HTTPHEADER => [
            'Content-Type: application/json',
            'X-API-Key: ' . $apiKey,
        ],
    ]);
    
    $response = curl_exec($ch);
    curl_close($ch);
    
    return json_decode($response, true);
}

// Usage
$result = trackClick('off_abc123', ['sub1' => 'facebook']);
print_r($result);
```

### Go

```go
package main

import (
    "bytes"
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

func trackClick(offerID string, subIDs map[string]string) (map[string]interface{}, error) {
    apiKey := os.Getenv("AFFTOK_API_KEY")
    advertiserID := os.Getenv("AFFTOK_ADVERTISER_ID")
    timestamp := time.Now().UnixMilli()
    
    nonceBytes := make([]byte, 16)
    rand.Read(nonceBytes)
    nonce := hex.EncodeToString(nonceBytes)
    
    dataToSign := fmt.Sprintf("%s|%s|%d|%s", apiKey, advertiserID, timestamp, nonce)
    h := hmac.New(sha256.New, []byte(apiKey))
    h.Write([]byte(dataToSign))
    signature := hex.EncodeToString(h.Sum(nil))
    
    payload := map[string]interface{}{
        "api_key":       apiKey,
        "advertiser_id": advertiserID,
        "offer_id":      offerID,
        "timestamp":     timestamp,
        "nonce":         nonce,
        "signature":     signature,
    }
    
    if subIDs != nil {
        payload["sub_id_1"] = subIDs["sub1"]
        payload["sub_id_2"] = subIDs["sub2"]
        payload["sub_id_3"] = subIDs["sub3"]
    }
    
    jsonData, _ := json.Marshal(payload)
    
    req, _ := http.NewRequest("POST", "https://api.afftok.com/api/sdk/click", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", apiKey)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}

func main() {
    result, _ := trackClick("off_abc123", map[string]string{"sub1": "facebook"})
    fmt.Printf("%+v\n", result)
}
```

## Next Steps

- [Conversions API](conversions.md) - Track conversions
- [Postbacks API](postbacks.md) - Server-to-server postbacks
- [Stats API](stats.md) - Analytics and statistics

