# Stats API

Retrieve analytics and performance data.

---

## GET /api/stats/me

Get your overall statistics.

### Request

```http
GET /api/stats/me
Authorization: Bearer <token>
```

### Response

```json
{
  "success": true,
  "data": {
    "summary": {
      "total_clicks": 152340,
      "unique_clicks": 124560,
      "total_conversions": 4567,
      "total_earnings": 22835.00,
      "conversion_rate": 3.67,
      "epc": 0.18
    },
    "today": {
      "clicks": 1523,
      "unique_clicks": 1245,
      "conversions": 45,
      "earnings": 225.00
    },
    "this_week": {
      "clicks": 10234,
      "conversions": 312,
      "earnings": 1560.00
    },
    "this_month": {
      "clicks": 45678,
      "conversions": 1456,
      "earnings": 7280.00
    },
    "pending_earnings": 450.00,
    "available_balance": 22385.00
  }
}
```

### Example

```bash
curl -X GET https://api.afftok.com/api/stats/me \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## GET /api/stats/daily

Get daily statistics.

### Request

```http
GET /api/stats/daily
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `start_date` | string | Start date (YYYY-MM-DD) |
| `end_date` | string | End date (YYYY-MM-DD) |
| `offer_id` | string | Filter by offer |

### Response

```json
{
  "success": true,
  "data": {
    "daily_stats": [
      {
        "date": "2024-12-01",
        "clicks": 1523,
        "unique_clicks": 1245,
        "conversions": 45,
        "earnings": 225.00,
        "conversion_rate": 3.61,
        "epc": 0.18
      },
      {
        "date": "2024-11-30",
        "clicks": 1456,
        "unique_clicks": 1189,
        "conversions": 42,
        "earnings": 210.00,
        "conversion_rate": 3.53,
        "epc": 0.18
      }
    ],
    "totals": {
      "clicks": 45678,
      "conversions": 1456,
      "earnings": 7280.00
    }
  }
}
```

### Example

```bash
curl -X GET "https://api.afftok.com/api/stats/daily?start_date=2024-11-01&end_date=2024-11-30" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## GET /api/stats/by-offer

Get statistics grouped by offer.

### Request

```http
GET /api/stats/by-offer
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `start_date` | string | Start date |
| `end_date` | string | End date |
| `sort` | string | Sort by (clicks, conversions, earnings) |
| `order` | string | Sort order (asc, desc) |
| `limit` | integer | Number of results |

### Response

```json
{
  "success": true,
  "data": {
    "offers": [
      {
        "offer_id": "off_123456",
        "offer_title": "Premium VPN Service",
        "clicks": 15234,
        "unique_clicks": 12456,
        "conversions": 456,
        "earnings": 6840.00,
        "conversion_rate": 3.66,
        "epc": 0.55
      },
      {
        "offer_id": "off_789012",
        "offer_title": "Fitness App",
        "clicks": 8567,
        "unique_clicks": 7234,
        "conversions": 234,
        "earnings": 2340.00,
        "conversion_rate": 3.23,
        "epc": 0.32
      }
    ]
  }
}
```

---

## GET /api/stats/by-country

Get statistics grouped by country.

### Request

```http
GET /api/stats/by-country
Authorization: Bearer <token>
```

### Response

```json
{
  "success": true,
  "data": {
    "countries": [
      {
        "country": "US",
        "country_name": "United States",
        "clicks": 85000,
        "conversions": 2800,
        "earnings": 14000.00,
        "conversion_rate": 3.29
      },
      {
        "country": "CA",
        "country_name": "Canada",
        "clicks": 32000,
        "conversions": 950,
        "earnings": 4750.00,
        "conversion_rate": 2.97
      },
      {
        "country": "GB",
        "country_name": "United Kingdom",
        "clicks": 21000,
        "conversions": 550,
        "earnings": 2750.00,
        "conversion_rate": 2.62
      }
    ]
  }
}
```

---

## GET /api/stats/by-device

Get statistics grouped by device type.

### Request

```http
GET /api/stats/by-device
Authorization: Bearer <token>
```

### Response

```json
{
  "success": true,
  "data": {
    "devices": [
      {
        "device": "mobile",
        "clicks": 95000,
        "conversions": 2850,
        "earnings": 14250.00,
        "conversion_rate": 3.00,
        "percentage": 62.3
      },
      {
        "device": "desktop",
        "clicks": 52000,
        "conversions": 1550,
        "earnings": 7750.00,
        "conversion_rate": 2.98,
        "percentage": 34.1
      },
      {
        "device": "tablet",
        "clicks": 5340,
        "conversions": 167,
        "earnings": 835.00,
        "conversion_rate": 3.13,
        "percentage": 3.5
      }
    ]
  }
}
```

---

## GET /api/stats/hourly

Get hourly statistics for today.

### Request

```http
GET /api/stats/hourly
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `date` | string | Specific date (default: today) |
| `timezone` | string | Timezone (default: UTC) |

### Response

```json
{
  "success": true,
  "data": {
    "date": "2024-12-01",
    "timezone": "UTC",
    "hourly": [
      { "hour": 0, "clicks": 45, "conversions": 1, "earnings": 5.00 },
      { "hour": 1, "clicks": 32, "conversions": 0, "earnings": 0.00 },
      { "hour": 2, "clicks": 28, "conversions": 1, "earnings": 5.00 },
      // ... hours 3-23
      { "hour": 23, "clicks": 89, "conversions": 3, "earnings": 15.00 }
    ],
    "peak_hour": 14,
    "total_clicks": 1523,
    "total_conversions": 45
  }
}
```

---

## GET /api/clicks/my

Get your click history.

### Request

```http
GET /api/clicks/my
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `offer_id` | string | Filter by offer |
| `start_date` | string | Start date |
| `end_date` | string | End date |
| `country` | string | Filter by country |
| `device` | string | Filter by device |
| `converted` | boolean | Filter converted clicks |

### Response

```json
{
  "success": true,
  "data": {
    "clicks": [
      {
        "click_id": "clk_a1b2c3d4e5f6",
        "offer_id": "off_123456",
        "offer_title": "Premium VPN Service",
        "ip_address": "203.0.113.xxx",
        "country": "US",
        "city": "New York",
        "device": "mobile",
        "browser": "Chrome",
        "os": "Android",
        "converted": true,
        "conversion_id": "conv_xyz789",
        "earnings": 15.00,
        "clicked_at": "2024-12-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 1523,
      "total_pages": 77
    }
  }
}
```

---

## GET /api/conversions/my

Get your conversion history.

### Request

```http
GET /api/conversions/my
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `offer_id` | string | Filter by offer |
| `status` | string | Filter by status |
| `start_date` | string | Start date |
| `end_date` | string | End date |

### Response

```json
{
  "success": true,
  "data": {
    "conversions": [
      {
        "conversion_id": "conv_xyz789abc",
        "click_id": "clk_a1b2c3d4e5f6",
        "offer_id": "off_123456",
        "offer_title": "Premium VPN Service",
        "transaction_id": "txn_advertiser_12345",
        "amount": 49.99,
        "payout": 15.00,
        "currency": "USD",
        "status": "approved",
        "country": "US",
        "device": "mobile",
        "converted_at": "2024-12-01T12:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 456,
      "total_pages": 23
    },
    "summary": {
      "total_conversions": 456,
      "approved": 420,
      "pending": 30,
      "rejected": 6,
      "total_earnings": 6300.00
    }
  }
}
```

---

## GET /api/stats/realtime

Get real-time statistics (last 5 minutes).

### Request

```http
GET /api/stats/realtime
Authorization: Bearer <token>
```

### Response

```json
{
  "success": true,
  "data": {
    "last_5_minutes": {
      "clicks": 23,
      "conversions": 1,
      "earnings": 15.00
    },
    "active_visitors": 156,
    "clicks_per_minute": 4.6,
    "top_countries": [
      { "country": "US", "clicks": 12 },
      { "country": "CA", "clicks": 5 },
      { "country": "GB", "clicks": 3 }
    ],
    "recent_clicks": [
      {
        "click_id": "clk_latest1",
        "country": "US",
        "device": "mobile",
        "clicked_at": "2024-12-01T12:04:30Z"
      }
    ],
    "timestamp": "2024-12-01T12:05:00Z"
  }
}
```

---

## Export Statistics

### GET /api/stats/export

Export statistics as CSV or JSON.

### Request

```http
GET /api/stats/export?format=csv&start_date=2024-11-01&end_date=2024-11-30
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `format` | string | Export format (csv, json) |
| `type` | string | Data type (clicks, conversions, daily) |
| `start_date` | string | Start date |
| `end_date` | string | End date |

### Response (CSV)

```csv
date,clicks,unique_clicks,conversions,earnings,conversion_rate
2024-11-01,1523,1245,45,225.00,3.61
2024-11-02,1456,1189,42,210.00,3.53
...
```

---

Next: [Geo Rules API](./geo-rules.md)

