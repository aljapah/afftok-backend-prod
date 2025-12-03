# Offers API

Manage offers, join campaigns, and retrieve tracking links.

---

## GET /api/offers

List all available offers.

### Request

```http
GET /api/offers
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number (default: 1) |
| `limit` | integer | Items per page (default: 20, max: 100) |
| `category` | string | Filter by category |
| `country` | string | Filter by target country |
| `status` | string | Filter by status (active, paused) |
| `search` | string | Search by title/description |
| `sort` | string | Sort field (created_at, payout, title) |
| `order` | string | Sort order (asc, desc) |

### Response

```json
{
  "success": true,
  "data": {
    "offers": [
      {
        "id": "off_123456",
        "title": "Premium VPN Service",
        "title_ar": "خدمة VPN المميزة",
        "description": "High-converting VPN offer with great payouts",
        "description_ar": "عرض VPN عالي التحويل مع عوائد ممتازة",
        "category": "software",
        "payout": 15.00,
        "payout_type": "cpa",
        "currency": "USD",
        "preview_url": "https://example.com/offer-preview",
        "thumbnail": "https://cdn.afftok.com/offers/off_123456.jpg",
        "countries": ["US", "CA", "GB", "AU"],
        "devices": ["desktop", "mobile", "tablet"],
        "status": "active",
        "conversion_rate": 3.5,
        "epc": 0.52,
        "terms": "No incentivized traffic. Email allowed.",
        "terms_ar": "لا يُسمح بحركة المرور المحفزة. البريد الإلكتروني مسموح.",
        "network": {
          "id": "net_abc123",
          "name": "Premium Network"
        },
        "created_at": "2024-01-15T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 156,
      "total_pages": 8
    }
  }
}
```

### Example

```bash
curl -X GET "https://api.afftok.com/api/offers?category=software&country=US&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## GET /api/offers/:id

Get details of a specific offer.

### Request

```http
GET /api/offers/off_123456
Authorization: Bearer <token>
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "off_123456",
    "title": "Premium VPN Service",
    "title_ar": "خدمة VPN المميزة",
    "description": "High-converting VPN offer with great payouts",
    "description_ar": "عرض VPN عالي التحويل مع عوائد ممتازة",
    "category": "software",
    "payout": 15.00,
    "payout_type": "cpa",
    "currency": "USD",
    "preview_url": "https://example.com/offer-preview",
    "landing_url": "https://example.com/landing",
    "thumbnail": "https://cdn.afftok.com/offers/off_123456.jpg",
    "countries": ["US", "CA", "GB", "AU"],
    "devices": ["desktop", "mobile", "tablet"],
    "status": "active",
    "terms": "No incentivized traffic. Email allowed.",
    "terms_ar": "لا يُسمح بحركة المرور المحفزة. البريد الإلكتروني مسموح.",
    "requirements": {
      "min_daily_clicks": 100,
      "approval_required": false
    },
    "caps": {
      "daily_conversions": 500,
      "total_conversions": null
    },
    "stats": {
      "total_clicks": 125000,
      "total_conversions": 4375,
      "conversion_rate": 3.5,
      "epc": 0.52
    },
    "network": {
      "id": "net_abc123",
      "name": "Premium Network"
    },
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-12-01T12:00:00Z"
  }
}
```

---

## POST /api/offers/join

Join an offer to receive your tracking link.

### Request

```http
POST /api/offers/join
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "offer_id": "off_123456"
}
```

### Response

```json
{
  "success": true,
  "message": "Successfully joined offer",
  "data": {
    "user_offer_id": "uo_abc123xyz",
    "offer_id": "off_123456",
    "offer_title": "Premium VPN Service",
    "tracking_url": "https://track.afftok.com/c/abc123xyz",
    "signed_link": "https://track.afftok.com/c/abc123xyz.1701234567890.nonce123.sig456",
    "short_link": "https://afftok.link/abc123xyz",
    "payout": 15.00,
    "joined_at": "2024-12-01T12:00:00Z"
  }
}
```

### Errors

| Code | Description |
|------|-------------|
| `ALREADY_JOINED` | You've already joined this offer |
| `OFFER_PAUSED` | Offer is currently paused |
| `OFFER_NOT_FOUND` | Offer doesn't exist |
| `APPROVAL_REQUIRED` | Offer requires manual approval |
| `GEO_RESTRICTED` | Your country is not allowed |

### Example

```bash
curl -X POST https://api.afftok.com/api/offers/join \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"offer_id": "off_123456"}'
```

---

## GET /api/offers/my

List offers you've joined.

### Request

```http
GET /api/offers/my
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `status` | string | Filter by status |

### Response

```json
{
  "success": true,
  "data": {
    "user_offers": [
      {
        "user_offer_id": "uo_abc123xyz",
        "offer": {
          "id": "off_123456",
          "title": "Premium VPN Service",
          "thumbnail": "https://cdn.afftok.com/offers/off_123456.jpg",
          "payout": 15.00,
          "status": "active"
        },
        "tracking_url": "https://track.afftok.com/c/abc123xyz",
        "signed_link": "https://track.afftok.com/c/abc123xyz.1701234567890.nonce123.sig456",
        "short_link": "https://afftok.link/abc123xyz",
        "stats": {
          "total_clicks": 1523,
          "unique_clicks": 1245,
          "total_conversions": 45,
          "earnings": 675.00,
          "conversion_rate": 3.61
        },
        "joined_at": "2024-06-15T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 12,
      "total_pages": 1
    }
  }
}
```

---

## DELETE /api/offers/my/:user_offer_id

Leave an offer.

### Request

```http
DELETE /api/offers/my/uo_abc123xyz
Authorization: Bearer <token>
```

### Response

```json
{
  "success": true,
  "message": "Successfully left offer"
}
```

---

## GET /api/offers/:id/stats

Get detailed statistics for a specific offer.

### Request

```http
GET /api/offers/off_123456/stats
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `start_date` | string | Start date (YYYY-MM-DD) |
| `end_date` | string | End date (YYYY-MM-DD) |
| `group_by` | string | Grouping (day, week, month) |

### Response

```json
{
  "success": true,
  "data": {
    "offer_id": "off_123456",
    "summary": {
      "total_clicks": 15234,
      "unique_clicks": 12456,
      "total_conversions": 456,
      "earnings": 6840.00,
      "conversion_rate": 3.66,
      "epc": 0.55
    },
    "daily": [
      {
        "date": "2024-12-01",
        "clicks": 523,
        "conversions": 18,
        "earnings": 270.00
      },
      {
        "date": "2024-11-30",
        "clicks": 498,
        "conversions": 15,
        "earnings": 225.00
      }
    ],
    "by_country": [
      { "country": "US", "clicks": 8500, "conversions": 280 },
      { "country": "CA", "clicks": 3200, "conversions": 95 },
      { "country": "GB", "clicks": 2100, "conversions": 55 }
    ],
    "by_device": [
      { "device": "mobile", "clicks": 9500, "conversions": 285 },
      { "device": "desktop", "clicks": 5200, "conversions": 155 },
      { "device": "tablet", "clicks": 534, "conversions": 16 }
    ]
  }
}
```

---

## POST /api/offers/:id/refresh-link

Generate a new signed tracking link.

### Request

```http
POST /api/offers/off_123456/refresh-link
Authorization: Bearer <token>
```

### Response

```json
{
  "success": true,
  "data": {
    "tracking_url": "https://track.afftok.com/c/abc123xyz",
    "signed_link": "https://track.afftok.com/c/abc123xyz.1701234567890.newnonce.newsig",
    "expires_at": "2024-12-01T12:05:00Z"
  }
}
```

---

## Offer Categories

| Category | Description |
|----------|-------------|
| `software` | Software & Apps |
| `ecommerce` | E-commerce & Retail |
| `finance` | Finance & Banking |
| `gaming` | Gaming & Entertainment |
| `health` | Health & Wellness |
| `education` | Education & Courses |
| `travel` | Travel & Hospitality |
| `utilities` | Utilities & Services |

---

## Payout Types

| Type | Description |
|------|-------------|
| `cpa` | Cost Per Action (fixed payout) |
| `cps` | Cost Per Sale (percentage) |
| `cpl` | Cost Per Lead |
| `cpi` | Cost Per Install |
| `revshare` | Revenue Share |

---

Next: [Clicks API](./clicks.md)

