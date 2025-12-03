# Stats API

Endpoints for retrieving analytics and performance statistics.

---

## Get User Stats

Retrieve comprehensive statistics for the authenticated user.

```
GET /api/stats/me
```

### Headers

```http
Authorization: Bearer <jwt_token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | Time period: `today`, `yesterday`, `week`, `month`, `all` |

### Response

```json
{
  "success": true,
  "data": {
    "user_id": "usr_abc123",
    "period": "month",
    "overview": {
      "total_clicks": 15000,
      "unique_clicks": 12500,
      "total_conversions": 300,
      "approved_conversions": 285,
      "pending_conversions": 10,
      "rejected_conversions": 5,
      "total_earnings": 1425.00,
      "pending_earnings": 50.00,
      "conversion_rate": 2.28,
      "epc": 0.095
    },
    "today": {
      "clicks": 520,
      "conversions": 12,
      "earnings": 60.00
    },
    "this_week": {
      "clicks": 3200,
      "conversions": 68,
      "earnings": 340.00
    },
    "this_month": {
      "clicks": 15000,
      "conversions": 300,
      "earnings": 1425.00
    },
    "top_offers": [
      {
        "offer_id": "off_abc123",
        "title": "Premium App",
        "clicks": 5000,
        "conversions": 120,
        "earnings": 600.00
      }
    ],
    "top_countries": [
      {
        "country": "US",
        "clicks": 8000,
        "conversions": 180,
        "percentage": 53.3
      },
      {
        "country": "GB",
        "clicks": 3500,
        "conversions": 70,
        "percentage": 23.3
      }
    ]
  }
}
```

### Code Examples

**cURL:**

```bash
curl -X GET "https://api.afftok.com/api/stats/me?period=month" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**JavaScript:**

```javascript
const response = await fetch('https://api.afftok.com/api/stats/me?period=week', {
  headers: {
    'Authorization': `Bearer ${jwtToken}`,
  },
});
const stats = await response.json();
console.log('Total Earnings:', stats.data.overview.total_earnings);
```

**Python:**

```python
import httpx

response = httpx.get(
    "https://api.afftok.com/api/stats/me",
    headers={"Authorization": f"Bearer {jwt_token}"},
    params={"period": "month"},
)
stats = response.json()["data"]
print(f"Total Clicks: {stats['overview']['total_clicks']}")
```

---

## Get Daily Stats

Retrieve daily breakdown of statistics for charts and reports.

```
GET /api/stats/daily
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `start_date` | string | Start date (YYYY-MM-DD) |
| `end_date` | string | End date (YYYY-MM-DD) |
| `offer_id` | string | Filter by specific offer |
| `metric` | string | Metric to return: `clicks`, `conversions`, `earnings`, `all` |

### Response

```json
{
  "success": true,
  "data": {
    "period": {
      "start": "2024-01-01",
      "end": "2024-01-31"
    },
    "daily": [
      {
        "date": "2024-01-01",
        "clicks": 450,
        "unique_clicks": 380,
        "conversions": 9,
        "earnings": 45.00,
        "conversion_rate": 2.0,
        "epc": 0.10
      },
      {
        "date": "2024-01-02",
        "clicks": 520,
        "unique_clicks": 440,
        "conversions": 11,
        "earnings": 55.00,
        "conversion_rate": 2.12,
        "epc": 0.106
      }
    ],
    "totals": {
      "clicks": 15000,
      "conversions": 300,
      "earnings": 1500.00
    }
  }
}
```

### Code Examples

**cURL:**

```bash
curl -X GET "https://api.afftok.com/api/stats/daily?start_date=2024-01-01&end_date=2024-01-31" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**JavaScript:**

```javascript
const startDate = '2024-01-01';
const endDate = '2024-01-31';
const response = await fetch(
  `https://api.afftok.com/api/stats/daily?start_date=${startDate}&end_date=${endDate}`,
  {
    headers: {
      'Authorization': `Bearer ${jwtToken}`,
    },
  }
);
const data = await response.json();

// Use for charts
const chartData = data.data.daily.map(day => ({
  x: day.date,
  y: day.clicks,
}));
```

---

## Get Clicks by Offer

Retrieve click statistics grouped by offer.

```
GET /api/clicks/by-offer
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | Time period: `today`, `yesterday`, `week`, `month`, `all` |
| `start_date` | string | Start date (YYYY-MM-DD) |
| `end_date` | string | End date (YYYY-MM-DD) |

### Response

```json
{
  "success": true,
  "data": {
    "offers": [
      {
        "offer_id": "off_abc123",
        "offer_title": "Premium App Install",
        "total_clicks": 5000,
        "unique_clicks": 4200,
        "conversions": 120,
        "earnings": 600.00,
        "conversion_rate": 2.86,
        "epc": 0.12
      },
      {
        "offer_id": "off_def456",
        "offer_title": "E-commerce Signup",
        "total_clicks": 3500,
        "unique_clicks": 2900,
        "conversions": 85,
        "earnings": 425.00,
        "conversion_rate": 2.93,
        "epc": 0.121
      }
    ],
    "totals": {
      "clicks": 8500,
      "conversions": 205,
      "earnings": 1025.00
    }
  }
}
```

---

## Get My Clicks

Retrieve detailed click history.

```
GET /api/clicks/my
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number (default: 1) |
| `limit` | integer | Items per page (default: 50, max: 100) |
| `offer_id` | string | Filter by offer |
| `country` | string | Filter by country code |
| `device` | string | Filter by device: `mobile`, `desktop`, `tablet` |
| `start_date` | string | Start date |
| `end_date` | string | End date |
| `converted` | boolean | Filter by conversion status |

### Response

```json
{
  "success": true,
  "data": {
    "clicks": [
      {
        "id": "clk_abc123def456",
        "offer_id": "off_abc123",
        "offer_title": "Premium App Install",
        "tracking_code": "tc_xyz789",
        "clicked_at": "2024-01-15T10:30:45Z",
        "ip_address": "192.168.x.x",
        "country": "US",
        "city": "New York",
        "device": "mobile",
        "os": "iOS",
        "browser": "Safari",
        "referer": "https://facebook.com",
        "converted": true,
        "conversion_id": "conv_xyz789",
        "conversion_amount": 5.00,
        "sub_id": "campaign_123"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 1250,
      "total_pages": 25
    }
  }
}
```

### Code Examples

**cURL:**

```bash
curl -X GET "https://api.afftok.com/api/clicks/my?limit=20&converted=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**JavaScript:**

```javascript
const response = await fetch('https://api.afftok.com/api/clicks/my?limit=50&country=US', {
  headers: {
    'Authorization': `Bearer ${jwtToken}`,
  },
});
const data = await response.json();
console.log('Clicks:', data.data.clicks);
```

---

## Get Offer Click Stats

Retrieve click statistics for a specific offer.

```
GET /api/offers/:id/clicks/stats
```

### Response

```json
{
  "success": true,
  "data": {
    "offer_id": "off_abc123",
    "period": "month",
    "clicks": {
      "total": 5000,
      "unique": 4200,
      "today": 180,
      "yesterday": 165,
      "this_week": 1200,
      "this_month": 5000
    },
    "conversions": {
      "total": 120,
      "approved": 115,
      "pending": 3,
      "rejected": 2,
      "today": 5,
      "yesterday": 4,
      "this_week": 28,
      "this_month": 120
    },
    "earnings": {
      "total": 575.00,
      "pending": 15.00,
      "today": 25.00,
      "yesterday": 20.00,
      "this_week": 140.00,
      "this_month": 575.00
    },
    "breakdown": {
      "by_country": [
        {"country": "US", "clicks": 3000, "conversions": 75, "percentage": 60},
        {"country": "GB", "clicks": 1000, "conversions": 25, "percentage": 20},
        {"country": "CA", "clicks": 500, "conversions": 12, "percentage": 10}
      ],
      "by_device": [
        {"device": "mobile", "clicks": 3500, "percentage": 70},
        {"device": "desktop", "clicks": 1200, "percentage": 24},
        {"device": "tablet", "clicks": 300, "percentage": 6}
      ],
      "by_os": [
        {"os": "iOS", "clicks": 2000, "percentage": 40},
        {"os": "Android", "clicks": 1500, "percentage": 30},
        {"os": "Windows", "clicks": 1000, "percentage": 20}
      ]
    }
  }
}
```

---

## Get Conversion Stats

Retrieve detailed conversion statistics.

```
GET /api/conversions/stats
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | Time period |
| `offer_id` | string | Filter by offer |
| `status` | string | Filter by status: `pending`, `approved`, `rejected` |

### Response

```json
{
  "success": true,
  "data": {
    "overview": {
      "total": 300,
      "approved": 285,
      "pending": 10,
      "rejected": 5,
      "approval_rate": 95.0,
      "total_value": 14250.00,
      "total_earnings": 1425.00
    },
    "by_status": {
      "approved": {
        "count": 285,
        "earnings": 1425.00,
        "avg_payout": 5.00
      },
      "pending": {
        "count": 10,
        "potential_earnings": 50.00
      },
      "rejected": {
        "count": 5,
        "lost_earnings": 25.00
      }
    },
    "recent": [
      {
        "id": "conv_xyz789",
        "offer_id": "off_abc123",
        "offer_title": "Premium App",
        "amount": 49.99,
        "payout": 5.00,
        "status": "approved",
        "converted_at": "2024-01-15T10:35:00Z"
      }
    ]
  }
}
```

---

## Get Earnings Summary

Retrieve earnings summary and payout information.

```
GET /api/earnings/summary
```

### Response

```json
{
  "success": true,
  "data": {
    "balance": {
      "available": 1425.00,
      "pending": 50.00,
      "lifetime": 5000.00
    },
    "this_month": {
      "earnings": 1425.00,
      "conversions": 285,
      "clicks": 15000
    },
    "last_month": {
      "earnings": 1200.00,
      "conversions": 240,
      "clicks": 12000
    },
    "payouts": {
      "last_payout": {
        "amount": 500.00,
        "date": "2024-01-01T00:00:00Z",
        "method": "paypal"
      },
      "next_payout_date": "2024-02-01T00:00:00Z",
      "minimum_payout": 50.00
    }
  }
}
```

---

## Real-time Stats (WebSocket)

Connect to real-time stats updates via WebSocket.

```
wss://api.afftok.com/ws/stats
```

### Connection

```javascript
const ws = new WebSocket('wss://api.afftok.com/ws/stats');

ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'subscribe',
    token: 'YOUR_JWT_TOKEN',
    channels: ['clicks', 'conversions', 'earnings'],
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Real-time update:', data);
};
```

### Message Types

**Click Event:**

```json
{
  "type": "click",
  "data": {
    "offer_id": "off_abc123",
    "country": "US",
    "device": "mobile",
    "timestamp": "2024-01-15T10:30:45Z"
  }
}
```

**Conversion Event:**

```json
{
  "type": "conversion",
  "data": {
    "offer_id": "off_abc123",
    "amount": 49.99,
    "payout": 5.00,
    "status": "approved",
    "timestamp": "2024-01-15T10:35:00Z"
  }
}
```

---

## Error Responses

**Invalid Date Range (400):**

```json
{
  "success": false,
  "error": "Invalid date range",
  "code": "INVALID_DATE_RANGE"
}
```

**No Data (404):**

```json
{
  "success": false,
  "error": "No statistics found for the specified period",
  "code": "NO_DATA"
}
```

---

## Next Steps

- [Offers API](offers.md) - Manage your offers
- [Clicks API](clicks.md) - Detailed click tracking
- [Webhooks](../webhooks/overview.md) - Real-time notifications

