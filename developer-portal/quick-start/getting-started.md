# Getting Started with AffTok

Get up and running with AffTok in under 5 minutes.

## Prerequisites

- An AffTok account ([Sign up here](https://dashboard.afftok.com/signup))
- Basic understanding of HTTP APIs or SDKs

## Step 1: Generate API Key

### Via Dashboard

1. Log in to [AffTok Dashboard](https://dashboard.afftok.com)
2. Navigate to **Settings** → **API Keys**
3. Click **Generate New Key**
4. Copy and securely store your API key

> ⚠️ **Important**: The API key is shown only once. Store it securely!

### API Key Format

```
afftok_live_sk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
       │    │  └── 32-character random string
       │    └── Secret key indicator
       └── Environment (live/test)
```

## Step 2: Create Your First Offer

### Via Dashboard

1. Go to **Offers** → **Create Offer**
2. Fill in the details:
   - **Name**: My First Offer
   - **Destination URL**: https://example.com/landing
   - **Payout**: $5.00
   - **Status**: Active
3. Click **Create**

### Via API

```bash
curl -X POST https://api.afftok.com/api/offers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "My First Offer",
    "description": "Test offer for integration",
    "destination_url": "https://example.com/landing",
    "payout": 5.00,
    "currency": "USD",
    "status": "active"
  }'
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "off_abc123def456",
    "title": "My First Offer",
    "destination_url": "https://example.com/landing",
    "payout": 5.00,
    "status": "active",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

## Step 3: Join the Offer (Get Tracking Link)

As a publisher/affiliate, join the offer to get your tracking link:

```bash
curl -X POST https://api.afftok.com/api/offers/off_abc123def456/join \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**

```json
{
  "success": true,
  "data": {
    "user_offer_id": "uo_xyz789",
    "offer_id": "off_abc123def456",
    "tracking_code": "abc123",
    "tracking_url": "https://track.afftok.com/c/abc123",
    "signed_link": "https://track.afftok.com/c/abc123.1699876543210.nonce123.sig456",
    "short_link": "https://afftok.link/abc123"
  }
}
```

## Step 4: Track Clicks

### Option A: Use Tracking Link Directly

Simply redirect users to your tracking link:

```html
<a href="https://track.afftok.com/c/abc123.1699876543210.nonce123.sig456">
  Click Here
</a>
```

### Option B: Use JavaScript SDK

```html
<script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>
<script>
  Afftok.init({
    apiKey: 'afftok_live_sk_xxxxx',
    advertiserId: 'adv_123456',
    debug: true
  });

  // Track a click
  Afftok.trackClick({
    offerId: 'off_abc123def456',
    subId1: 'facebook',
    subId2: 'banner_1'
  }).then(response => {
    console.log('Click tracked:', response);
  });
</script>
```

### Option C: Use Mobile SDK

**Android (Kotlin):**

```kotlin
Afftok.init(context, AfftokOptions(
    apiKey = "afftok_live_sk_xxxxx",
    advertiserId = "adv_123456",
    debug = true
))

Afftok.trackClick(ClickParams(
    offerId = "off_abc123def456",
    subId1 = "app_banner"
))
```

**iOS (Swift):**

```swift
Afftok.shared.initialize(options: AfftokOptions(
    apiKey: "afftok_live_sk_xxxxx",
    advertiserId: "adv_123456",
    debug: true
))

await Afftok.shared.trackClick(params: ClickParams(
    offerId: "off_abc123def456",
    subId1: "app_banner"
))
```

## Step 5: Track Conversions

When a conversion occurs (purchase, signup, etc.), send a postback:

### Server-to-Server Postback

```bash
curl -X POST https://api.afftok.com/api/postback \
  -H "Content-Type: application/json" \
  -H "X-API-Key: afftok_live_sk_xxxxx" \
  -d '{
    "api_key": "afftok_live_sk_xxxxx",
    "advertiser_id": "adv_123456",
    "offer_id": "off_abc123def456",
    "transaction_id": "txn_unique_123",
    "click_id": "clk_abc123",
    "amount": 49.99,
    "currency": "USD",
    "status": "approved",
    "timestamp": 1699876600000,
    "nonce": "random32characterstring12345678",
    "signature": "HMAC_SHA256_SIGNATURE"
  }'
```

### Generate Signature

```javascript
const crypto = require('crypto');

function generateSignature(apiKey, advertiserId, timestamp, nonce) {
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  return crypto.createHmac('sha256', apiKey).update(dataToSign).digest('hex');
}

const timestamp = Date.now();
const nonce = crypto.randomBytes(16).toString('hex');
const signature = generateSignature(apiKey, advertiserId, timestamp, nonce);
```

### Using SDK

```javascript
// Web SDK
await Afftok.trackConversion({
  offerId: 'off_abc123def456',
  transactionId: 'order_12345',
  amount: 49.99,
  currency: 'USD',
  status: 'approved'
});
```

## Step 6: View Statistics

### Via Dashboard

Navigate to **Analytics** → **Overview** to see:
- Total clicks
- Total conversions
- Conversion rate
- Earnings
- Daily trends

### Via API

```bash
curl -X GET "https://api.afftok.com/api/stats/me" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**

```json
{
  "success": true,
  "data": {
    "total_clicks": 1250,
    "total_conversions": 45,
    "conversion_rate": 3.6,
    "total_earnings": 225.00,
    "today": {
      "clicks": 150,
      "conversions": 5,
      "earnings": 25.00
    },
    "this_week": {
      "clicks": 800,
      "conversions": 28,
      "earnings": 140.00
    }
  }
}
```

## Quick Reference

### Base URLs

| Environment | URL |
|-------------|-----|
| Production API | `https://api.afftok.com` |
| Sandbox API | `https://sandbox.api.afftok.com` |
| Tracking | `https://track.afftok.com` |
| Short Links | `https://afftok.link` |

### Key Endpoints

| Action | Method | Endpoint |
|--------|--------|----------|
| Track Click | GET | `/api/c/:trackingCode` |
| SDK Click | POST | `/api/sdk/click` |
| SDK Conversion | POST | `/api/sdk/conversion` |
| Postback | POST | `/api/postback` |
| Get Stats | GET | `/api/stats/me` |
| List Offers | GET | `/api/offers` |
| Join Offer | POST | `/api/offers/:id/join` |

### SDK Installation

| Platform | Installation |
|----------|--------------|
| Android | `implementation 'com.afftok:afftok-sdk:1.0.0'` |
| iOS | Swift Package: `https://github.com/afftok/ios-sdk` |
| Flutter | `afftok: ^1.0.0` in pubspec.yaml |
| React Native | `npm install @afftok/react-native-sdk` |
| Web | `<script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>` |

## What's Next?

- [First Integration Tutorial](first-integration.md) - Step-by-step integration guide
- [Tracking Links](tracking-links.md) - Learn about signed links and TTL
- [Conversions Guide](conversions.md) - Deep dive into conversion tracking
- [API Reference](../api-reference/authentication.md) - Full API documentation
- [SDK Guides](../sdks/overview.md) - Platform-specific SDK documentation

## Need Help?

- **Email**: support@afftok.com
- **Documentation**: https://docs.afftok.com
- **Status**: https://status.afftok.com
- **Community**: https://community.afftok.com

