# Offers API

Endpoints for managing and retrieving offers in the AffTok platform.

---

## List Available Offers

Get all active offers available for promotion.

```
GET /api/offers
```

### Headers

```http
Authorization: Bearer <jwt_token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number (default: 1) |
| `limit` | integer | Items per page (default: 20, max: 100) |
| `category` | string | Filter by category |
| `network_id` | string | Filter by network |
| `status` | string | Filter by status: `active`, `paused`, `ended` |
| `sort` | string | Sort field: `created_at`, `payout`, `title` |
| `order` | string | Sort order: `asc`, `desc` |

### Response

```json
{
  "success": true,
  "data": {
    "offers": [
      {
        "id": "off_abc123def456",
        "title": "Premium App Install",
        "title_ar": "تثبيت التطبيق المميز",
        "description": "Promote our premium mobile application",
        "description_ar": "روّج لتطبيقنا المحمول المميز",
        "terms": "Valid for US traffic only. No incentivized traffic.",
        "terms_ar": "صالح لحركة المرور الأمريكية فقط. لا يُسمح بحركة المرور المحفزة.",
        "payout": 5.00,
        "currency": "USD",
        "payout_type": "CPA",
        "category": "mobile_apps",
        "network_id": "net_123",
        "network_name": "Example Network",
        "destination_url": "https://example.com/app",
        "preview_url": "https://example.com/preview",
        "thumbnail_url": "https://cdn.afftok.com/offers/off_abc123.jpg",
        "status": "active",
        "geo_targeting": ["US", "CA", "GB"],
        "device_targeting": ["mobile", "tablet"],
        "daily_cap": 1000,
        "total_cap": 50000,
        "remaining_cap": 45230,
        "conversion_flow": "install",
        "approval_required": false,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-10T12:00:00Z"
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

### Code Examples

**cURL:**

```bash
curl -X GET "https://api.afftok.com/api/offers?category=mobile_apps&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**JavaScript:**

```javascript
const response = await fetch('https://api.afftok.com/api/offers?limit=20', {
  headers: {
    'Authorization': `Bearer ${jwtToken}`,
  },
});
const data = await response.json();
console.log(data.data.offers);
```

**Python:**

```python
import httpx

response = httpx.get(
    "https://api.afftok.com/api/offers",
    headers={"Authorization": f"Bearer {jwt_token}"},
    params={"limit": 20, "category": "mobile_apps"},
)
offers = response.json()["data"]["offers"]
```

---

## Get Offer Details

Retrieve detailed information about a specific offer.

```
GET /api/offers/:id
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "off_abc123def456",
    "title": "Premium App Install",
    "title_ar": "تثبيت التطبيق المميز",
    "description": "Promote our premium mobile application...",
    "description_ar": "روّج لتطبيقنا المحمول المميز...",
    "terms": "Valid for US traffic only...",
    "terms_ar": "صالح لحركة المرور الأمريكية فقط...",
    "payout": 5.00,
    "currency": "USD",
    "payout_type": "CPA",
    "category": "mobile_apps",
    "network_id": "net_123",
    "network_name": "Example Network",
    "destination_url": "https://example.com/app",
    "preview_url": "https://example.com/preview",
    "thumbnail_url": "https://cdn.afftok.com/offers/off_abc123.jpg",
    "status": "active",
    "geo_targeting": ["US", "CA", "GB"],
    "device_targeting": ["mobile", "tablet"],
    "os_targeting": ["ios", "android"],
    "browser_targeting": [],
    "daily_cap": 1000,
    "total_cap": 50000,
    "remaining_cap": 45230,
    "conversion_flow": "install",
    "conversion_window_hours": 24,
    "approval_required": false,
    "creatives": [
      {
        "id": "crv_123",
        "type": "banner",
        "size": "300x250",
        "url": "https://cdn.afftok.com/creatives/banner_300x250.jpg"
      }
    ],
    "stats": {
      "total_clicks": 125000,
      "total_conversions": 2500,
      "conversion_rate": 2.0,
      "epc": 0.10
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-10T12:00:00Z"
  }
}
```

---

## Join Offer

Join an offer to receive your unique tracking link.

```
POST /api/offers/:id/join
```

### Headers

```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### Request Body (Optional)

```json
{
  "sub_id": "campaign_123",
  "source": "facebook"
}
```

### Response

```json
{
  "success": true,
  "message": "Successfully joined offer",
  "data": {
    "user_offer_id": "uo_xyz789",
    "offer_id": "off_abc123def456",
    "tracking_code": "tc_a1b2c3d4e5f6",
    "short_link": "https://trk.afftok.com/c/a1b2c3",
    "signed_link": "https://trk.afftok.com/c/a1b2c3.1699876543000.n1o2n3c4e5.sig123abc",
    "tracking_url": "https://api.afftok.com/api/c/tc_a1b2c3d4e5f6",
    "full_tracking_url": "https://api.afftok.com/api/c/tc_a1b2c3d4e5f6?sub={sub_id}",
    "joined_at": "2024-01-15T10:30:00Z"
  }
}
```

### Tracking Link Types

| Link Type | Description | Use Case |
|-----------|-------------|----------|
| `short_link` | Short, clean URL | Social media, SMS |
| `signed_link` | Cryptographically signed URL with TTL | Secure tracking |
| `tracking_url` | Standard tracking URL | General use |
| `full_tracking_url` | URL with sub-ID placeholder | Advanced tracking |

### Code Examples

**cURL:**

```bash
curl -X POST "https://api.afftok.com/api/offers/off_abc123def456/join" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"sub_id": "campaign_123"}'
```

**JavaScript:**

```javascript
const response = await fetch('https://api.afftok.com/api/offers/off_abc123def456/join', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${jwtToken}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ sub_id: 'campaign_123' }),
});
const data = await response.json();
console.log('Tracking Link:', data.data.signed_link);
```

---

## Get My Offers

List all offers you have joined.

```
GET /api/offers/my
```

### Response

```json
{
  "success": true,
  "data": {
    "offers": [
      {
        "user_offer_id": "uo_xyz789",
        "offer_id": "off_abc123def456",
        "offer_title": "Premium App Install",
        "offer_payout": 5.00,
        "tracking_code": "tc_a1b2c3d4e5f6",
        "short_link": "https://trk.afftok.com/c/a1b2c3",
        "signed_link": "https://trk.afftok.com/c/a1b2c3.1699876543000.n1o2n3c4e5.sig123abc",
        "total_clicks": 1250,
        "total_conversions": 25,
        "earnings": 125.00,
        "conversion_rate": 2.0,
        "status": "active",
        "joined_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-20T15:00:00Z"
      }
    ],
    "summary": {
      "total_offers": 5,
      "total_clicks": 5000,
      "total_conversions": 100,
      "total_earnings": 500.00
    }
  }
}
```

---

## Leave Offer

Leave an offer you have joined.

```
DELETE /api/offers/:id/leave
```

### Response

```json
{
  "success": true,
  "message": "Successfully left offer"
}
```

---

## Get Offer Stats

Get detailed statistics for a specific offer you've joined.

```
GET /api/offers/:id/stats
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | Time period: `today`, `yesterday`, `week`, `month`, `all` |
| `start_date` | string | Start date (ISO 8601) |
| `end_date` | string | End date (ISO 8601) |

### Response

```json
{
  "success": true,
  "data": {
    "offer_id": "off_abc123def456",
    "period": "week",
    "stats": {
      "total_clicks": 1250,
      "unique_clicks": 980,
      "total_conversions": 25,
      "approved_conversions": 23,
      "pending_conversions": 2,
      "rejected_conversions": 0,
      "earnings": 115.00,
      "pending_earnings": 10.00,
      "conversion_rate": 2.35,
      "epc": 0.092
    },
    "daily_breakdown": [
      {
        "date": "2024-01-15",
        "clicks": 180,
        "conversions": 4,
        "earnings": 20.00
      },
      {
        "date": "2024-01-16",
        "clicks": 195,
        "conversions": 3,
        "earnings": 15.00
      }
    ]
  }
}
```

---

## Offer Categories

| Category | Description |
|----------|-------------|
| `mobile_apps` | Mobile app installs |
| `ecommerce` | E-commerce/shopping |
| `gaming` | Games and gaming apps |
| `finance` | Financial services |
| `dating` | Dating apps and sites |
| `health` | Health and fitness |
| `education` | Educational content |
| `entertainment` | Entertainment and media |
| `utilities` | Utility apps |
| `sweepstakes` | Sweepstakes and contests |

## Payout Types

| Type | Description |
|------|-------------|
| `CPA` | Cost Per Action (fixed payout) |
| `CPS` | Cost Per Sale (percentage of sale) |
| `CPL` | Cost Per Lead |
| `CPI` | Cost Per Install |
| `RevShare` | Revenue sharing |

## Error Responses

**Offer Not Found (404):**

```json
{
  "success": false,
  "error": "Offer not found",
  "code": "OFFER_NOT_FOUND"
}
```

**Already Joined (409):**

```json
{
  "success": false,
  "error": "Already joined this offer",
  "code": "ALREADY_JOINED"
}
```

**Offer Paused (400):**

```json
{
  "success": false,
  "error": "Offer is currently paused",
  "code": "OFFER_PAUSED"
}
```

**Cap Reached (400):**

```json
{
  "success": false,
  "error": "Daily cap reached for this offer",
  "code": "CAP_REACHED"
}
```

---

## Next Steps

- [Tracking Links](../quick-start/tracking-links.md) - How to use your tracking links
- [Stats API](stats.md) - View detailed statistics
- [Webhooks](../webhooks/overview.md) - Receive conversion notifications

