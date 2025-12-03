# Quick Start Guide

Get up and running with AffTok in under 5 minutes.

## Prerequisites

Before you begin, you'll need:

- An AffTok account ([Sign up here](https://dashboard.afftok.com/signup))
- Access to the AffTok Dashboard
- Basic understanding of HTTP APIs

## Step 1: Generate Your API Key

### Via Dashboard

1. Log in to [AffTok Dashboard](https://dashboard.afftok.com)
2. Navigate to **Settings** → **API Keys**
3. Click **Generate New Key**
4. Copy and securely store your API key

> ⚠️ **Important**: Your API key is shown only once. Store it securely.

### API Key Format

```
afftok_live_sk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
└─────┬─────┘└┬┘└─────────────┬──────────────┘
   Prefix   Type          Secret
```

- `afftok_` - Platform identifier
- `live_` or `test_` - Environment
- `sk_` - Secret key type
- Remaining - Unique identifier

## Step 2: Join an Offer

### Via Dashboard

1. Go to **Offers** → **Available Offers**
2. Find an offer you want to promote
3. Click **Join Offer**
4. Copy your unique tracking link

### Via API

```bash
curl -X POST https://api.afftok.com/api/offers/join \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "offer_id": "off_123456"
  }'
```

**Response:**

```json
{
  "success": true,
  "data": {
    "user_offer_id": "uo_abc123",
    "offer_id": "off_123456",
    "tracking_url": "https://track.afftok.com/c/abc123",
    "signed_link": "https://track.afftok.com/c/abc123.1701234567890.xyz789.a1b2c3d4",
    "short_link": "https://afftok.link/abc123"
  }
}
```

## Step 3: Track Your First Click

### Using the Tracking Link

Simply share your tracking link. When a user clicks it:

1. AffTok records the click
2. User is redirected to the offer
3. Click appears in your dashboard

### Testing the Link

```bash
# Follow redirects to see the full flow
curl -L -v "https://track.afftok.com/c/abc123.1701234567890.xyz789.a1b2c3d4"
```

### Using the SDK (Optional)

```javascript
// Web SDK
await Afftok.init({
  apiKey: 'your_api_key',
  advertiserId: 'your_advertiser_id'
});

await Afftok.trackClick({
  offerId: 'off_123456',
  trackingCode: 'campaign_summer'
});
```

## Step 4: Send Your First Postback

When a conversion occurs, send a postback to AffTok:

### Basic Postback

```bash
curl -X POST https://api.afftok.com/api/postback \
  -H "Content-Type: application/json" \
  -H "X-API-Key: afftok_live_sk_xxxxx" \
  -d '{
    "click_id": "clk_a1b2c3d4e5f6",
    "transaction_id": "txn_your_unique_id",
    "amount": 29.99,
    "status": "approved"
  }'
```

**Response:**

```json
{
  "success": true,
  "data": {
    "conversion_id": "conv_xyz789",
    "click_id": "clk_a1b2c3d4e5f6",
    "amount": 29.99,
    "payout": 5.99,
    "status": "approved"
  }
}
```

### With Signature (Recommended)

```bash
# Generate signature
TIMESTAMP=$(date +%s000)
NONCE=$(openssl rand -hex 16)
API_KEY="afftok_live_sk_xxxxx"
ADVERTISER_ID="adv_123456"
DATA_TO_SIGN="${API_KEY}|${ADVERTISER_ID}|${TIMESTAMP}|${NONCE}"
SIGNATURE=$(echo -n "$DATA_TO_SIGN" | openssl dgst -sha256 -hmac "$API_KEY" | cut -d' ' -f2)

curl -X POST https://api.afftok.com/api/postback \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "api_key": "'$API_KEY'",
    "advertiser_id": "'$ADVERTISER_ID'",
    "click_id": "clk_a1b2c3d4e5f6",
    "transaction_id": "txn_your_unique_id",
    "amount": 29.99,
    "status": "approved",
    "timestamp": '$TIMESTAMP',
    "nonce": "'$NONCE'",
    "signature": "'$SIGNATURE'"
  }'
```

## Step 5: View Your Stats

### Via Dashboard

Visit [dashboard.afftok.com/stats](https://dashboard.afftok.com/stats) to see:

- Total clicks
- Unique clicks
- Conversions
- Earnings
- Conversion rate

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
    "total_clicks": 1523,
    "unique_clicks": 1245,
    "conversions": 45,
    "earnings": 225.00,
    "conversion_rate": 3.61,
    "today": {
      "clicks": 156,
      "conversions": 5,
      "earnings": 25.00
    }
  }
}
```

## Quick Reference

### Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://api.afftok.com` |
| Sandbox | `https://sandbox.api.afftok.com` |

### Key Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/offers` | GET | List available offers |
| `/api/offers/join` | POST | Join an offer |
| `/api/offers/my` | GET | Your joined offers |
| `/api/postback` | POST | Send conversion |
| `/api/stats/me` | GET | Your statistics |

### Authentication

```http
# JWT Token (for user endpoints)
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

# API Key (for postbacks)
X-API-Key: afftok_live_sk_xxxxx
```

## Next Steps

- [API Reference](../api/README.md) - Complete API documentation
- [SDK Installation](../sdks/installation.md) - Install mobile/web SDKs
- [Webhook Setup](../webhooks/README.md) - Real-time notifications
- [Security Best Practices](../security/README.md) - Secure your integration

## Need Help?

- **Documentation**: [docs.afftok.com](https://docs.afftok.com)
- **Email**: support@afftok.com
- **Developer Support**: developers@afftok.com

