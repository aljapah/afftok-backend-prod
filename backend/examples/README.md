# AffTok Server-to-Server Integration Examples

This directory contains examples for integrating with AffTok's Server-to-Server API in various programming languages.

## Available Examples

| Language | File | Description |
|----------|------|-------------|
| Node.js | `nodejs/index.js` | Express/Node.js example with axios |
| PHP | `php/tracker.php` | PHP example with cURL |
| Python | `python/main.py` | FastAPI example with httpx |
| Go | `go/main.go` | Go example with net/http |

## Quick Start

### 1. Set Environment Variables

```bash
export AFFTOK_API_KEY="your_api_key"
export AFFTOK_ADVERTISER_ID="your_advertiser_id"
export AFFTOK_BASE_URL="https://api.afftok.com"  # Optional
```

### 2. Run Examples

#### Node.js

```bash
cd nodejs
npm install axios
node index.js
```

#### PHP

```bash
cd php
php tracker.php
```

#### Python

```bash
cd python
pip install httpx fastapi pydantic
python main.py

# Or run as FastAPI server:
uvicorn main:app --reload
```

#### Go

```bash
cd go
go run main.go
```

## API Endpoints

### Send Postback/Conversion

```
POST /api/postback
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

### Track Click (Server-Side)

```
POST /api/sdk/click
```

**Request Body:**
```json
{
  "api_key": "your_api_key",
  "advertiser_id": "your_advertiser_id",
  "offer_id": "offer_123",
  "tracking_code": "campaign_code",
  "sub_id_1": "source",
  "sub_id_2": "medium",
  "sub_id_3": "campaign",
  "ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "timestamp": 1234567890000,
  "nonce": "random_32_char_string",
  "signature": "hmac_sha256_signature"
}
```

## Signature Generation

All requests must include an HMAC-SHA256 signature for authentication.

### Signature Format

```
signature = HMAC-SHA256(api_key, data_to_sign)
data_to_sign = "{api_key}|{advertiser_id}|{timestamp}|{nonce}"
```

### Example (Node.js)

```javascript
const crypto = require('crypto');

function generateSignature(apiKey, advertiserId, timestamp, nonce) {
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  return crypto.createHmac('sha256', apiKey).update(dataToSign).digest('hex');
}
```

### Example (PHP)

```php
function generateSignature($apiKey, $advertiserId, $timestamp, $nonce) {
    $dataToSign = "{$apiKey}|{$advertiserId}|{$timestamp}|{$nonce}";
    return hash_hmac('sha256', $dataToSign, $apiKey);
}
```

### Example (Python)

```python
import hmac
import hashlib

def generate_signature(api_key, advertiser_id, timestamp, nonce):
    data_to_sign = f"{api_key}|{advertiser_id}|{timestamp}|{nonce}"
    return hmac.new(api_key.encode(), data_to_sign.encode(), hashlib.sha256).hexdigest()
```

### Example (Go)

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

## Response Format

### Success Response

```json
{
  "success": true,
  "message": "Conversion tracked successfully",
  "data": {
    "conversion_id": "conv_abc123",
    "offer_id": "offer_123",
    "transaction_id": "txn_abc123"
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": "Invalid signature",
  "code": "INVALID_SIGNATURE"
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `INVALID_SIGNATURE` | HMAC signature verification failed |
| `EXPIRED_REQUEST` | Request timestamp too old (> 5 minutes) |
| `DUPLICATE_TRANSACTION` | Transaction ID already processed |
| `INVALID_OFFER` | Offer ID not found or inactive |
| `RATE_LIMITED` | Too many requests, try again later |
| `INVALID_API_KEY` | API key not found or revoked |

## Rate Limits

- 60 requests per minute per API key
- 1000 requests per hour per API key
- Batch requests count as multiple requests

## Best Practices

1. **Always validate responses** - Check the `success` field
2. **Implement retry logic** - Use exponential backoff for failures
3. **Use unique transaction IDs** - Prevent duplicate conversions
4. **Store API keys securely** - Use environment variables
5. **Log all requests** - For debugging and auditing
6. **Handle rate limits** - Implement proper backoff

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- API Status: https://status.afftok.com

