# Geo Rules API

Endpoints for managing geographic targeting rules to block or allow traffic from specific countries.

---

## Overview

Geo Rules allow you to control which countries can access your offers. Rules can be applied at three levels:

1. **Offer-level** - Applies to a specific offer
2. **Advertiser-level** - Applies to all offers from an advertiser
3. **Global** - Applies to all traffic

### Rule Resolution Order

When a click occurs, AffTok checks rules in this order:
1. Offer-specific rule (highest priority)
2. Advertiser-specific rule
3. Global rule
4. Default: Allow all

---

## List Geo Rules

```
GET /api/admin/geo-rules
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `scope_type` | string | Filter: `offer`, `advertiser`, `global` |
| `mode` | string | Filter: `allow`, `block` |
| `status` | string | Filter: `active`, `inactive` |

### Response

```json
{
  "success": true,
  "data": {
    "rules": [
      {
        "id": "geo_abc123",
        "name": "US Only Traffic",
        "scope_type": "offer",
        "scope_id": "off_xyz789",
        "scope_name": "Premium App Install",
        "mode": "allow",
        "countries": ["US"],
        "priority": 100,
        "status": "active",
        "stats": {
          "allowed": 15000,
          "blocked": 2500
        },
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-10T12:00:00Z"
      },
      {
        "id": "geo_def456",
        "name": "Block High-Fraud Countries",
        "scope_type": "global",
        "scope_id": null,
        "mode": "block",
        "countries": ["CN", "RU", "NG", "PK"],
        "priority": 10,
        "status": "active",
        "stats": {
          "allowed": 0,
          "blocked": 8500
        },
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45
    }
  }
}
```

---

## Create Geo Rule

```
POST /api/admin/geo-rules
```

### Request Body

```json
{
  "name": "US and Canada Only",
  "description": "Allow traffic only from US and Canada",
  "scope_type": "offer",
  "scope_id": "off_abc123",
  "mode": "allow",
  "countries": ["US", "CA"],
  "priority": 100,
  "status": "active"
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Rule name |
| `description` | string | No | Rule description |
| `scope_type` | string | Yes | `offer`, `advertiser`, or `global` |
| `scope_id` | string | No* | ID of offer/advertiser (required unless global) |
| `mode` | string | Yes | `allow` or `block` |
| `countries` | array | Yes | Array of ISO 3166-1 alpha-2 country codes |
| `priority` | integer | No | Higher = checked first (default: 50) |
| `status` | string | No | `active` or `inactive` (default: active) |

### Response

```json
{
  "success": true,
  "message": "Geo rule created successfully",
  "data": {
    "id": "geo_xyz789",
    "name": "US and Canada Only",
    "scope_type": "offer",
    "scope_id": "off_abc123",
    "mode": "allow",
    "countries": ["US", "CA"],
    "priority": 100,
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### Code Examples

**cURL:**

```bash
curl -X POST "https://api.afftok.com/api/admin/geo-rules" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Block Fraud Countries",
    "scope_type": "global",
    "mode": "block",
    "countries": ["CN", "RU", "NG"],
    "priority": 10
  }'
```

**JavaScript:**

```javascript
const response = await fetch('https://api.afftok.com/api/admin/geo-rules', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${adminToken}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    name: 'US Only for Premium Offer',
    scope_type: 'offer',
    scope_id: 'off_abc123',
    mode: 'allow',
    countries: ['US'],
    priority: 100,
  }),
});
const data = await response.json();
```

---

## Get Geo Rule

```
GET /api/admin/geo-rules/:id
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "geo_abc123",
    "name": "US Only Traffic",
    "description": "Allow only US traffic for this offer",
    "scope_type": "offer",
    "scope_id": "off_xyz789",
    "scope_name": "Premium App Install",
    "mode": "allow",
    "countries": ["US"],
    "priority": 100,
    "status": "active",
    "stats": {
      "total_checks": 17500,
      "allowed": 15000,
      "blocked": 2500,
      "last_blocked_at": "2024-01-15T10:25:00Z",
      "top_blocked_countries": [
        {"country": "IN", "count": 800},
        {"country": "PH", "count": 600},
        {"country": "BR", "count": 400}
      ]
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-10T12:00:00Z"
  }
}
```

---

## Update Geo Rule

```
PUT /api/admin/geo-rules/:id
```

### Request Body

```json
{
  "name": "US and UK Only",
  "countries": ["US", "GB"],
  "priority": 110
}
```

### Response

```json
{
  "success": true,
  "message": "Geo rule updated successfully"
}
```

---

## Delete Geo Rule

```
DELETE /api/admin/geo-rules/:id
```

### Response

```json
{
  "success": true,
  "message": "Geo rule deleted successfully"
}
```

---

## Get Geo Rules by Offer

```
GET /api/admin/offers/:id/geo-rules
```

### Response

```json
{
  "success": true,
  "data": {
    "offer_id": "off_abc123",
    "offer_title": "Premium App Install",
    "rules": [
      {
        "id": "geo_xyz789",
        "name": "US Only",
        "mode": "allow",
        "countries": ["US"],
        "priority": 100,
        "status": "active"
      }
    ],
    "effective_rule": {
      "mode": "allow",
      "countries": ["US"],
      "source": "offer"
    }
  }
}
```

---

## Get Geo Rules by Advertiser

```
GET /api/admin/advertisers/:id/geo-rules
```

### Response

```json
{
  "success": true,
  "data": {
    "advertiser_id": "adv_abc123",
    "advertiser_name": "Acme Corp",
    "rules": [
      {
        "id": "geo_def456",
        "name": "Block Asia",
        "mode": "block",
        "countries": ["CN", "IN", "PH"],
        "priority": 50,
        "status": "active"
      }
    ]
  }
}
```

---

## Test Geo Rule

Test how a geo rule would affect a specific country.

```
POST /api/admin/geo-rules/test
```

### Request Body

```json
{
  "offer_id": "off_abc123",
  "advertiser_id": "adv_xyz789",
  "country": "DE"
}
```

### Response

```json
{
  "success": true,
  "data": {
    "country": "DE",
    "allowed": true,
    "applied_rule": {
      "id": "geo_abc123",
      "name": "EU Countries Allowed",
      "scope_type": "advertiser",
      "mode": "allow"
    },
    "resolution_path": [
      {"scope": "offer", "rule": null, "result": "no_rule"},
      {"scope": "advertiser", "rule": "geo_abc123", "result": "allowed"},
      {"scope": "global", "rule": null, "result": "not_checked"}
    ]
  }
}
```

---

## Get Geo Rule Stats

```
GET /api/admin/geo-rules/:id/stats
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | `today`, `week`, `month`, `all` |

### Response

```json
{
  "success": true,
  "data": {
    "rule_id": "geo_abc123",
    "period": "month",
    "total_checks": 175000,
    "allowed": 150000,
    "blocked": 25000,
    "block_rate": 14.3,
    "by_country": [
      {"country": "US", "checks": 150000, "allowed": 150000, "blocked": 0},
      {"country": "IN", "checks": 8000, "allowed": 0, "blocked": 8000},
      {"country": "PH", "checks": 6000, "allowed": 0, "blocked": 6000}
    ],
    "by_day": [
      {"date": "2024-01-15", "allowed": 5000, "blocked": 850},
      {"date": "2024-01-14", "allowed": 4800, "blocked": 820}
    ]
  }
}
```

---

## Get Country Codes

Get the list of valid country codes.

```
GET /api/admin/geo-rules/countries
```

### Response

```json
{
  "success": true,
  "data": {
    "countries": [
      {"code": "US", "name": "United States"},
      {"code": "GB", "name": "United Kingdom"},
      {"code": "CA", "name": "Canada"},
      {"code": "AU", "name": "Australia"},
      {"code": "DE", "name": "Germany"},
      {"code": "FR", "name": "France"}
    ]
  }
}
```

---

## Common Geo Rule Patterns

### Allow Only Tier-1 Countries

```json
{
  "name": "Tier-1 Only",
  "scope_type": "offer",
  "scope_id": "off_abc123",
  "mode": "allow",
  "countries": ["US", "GB", "CA", "AU", "DE", "FR", "NL", "SE", "NO", "DK"],
  "priority": 100
}
```

### Block High-Fraud Countries (Global)

```json
{
  "name": "Block Fraud Countries",
  "scope_type": "global",
  "mode": "block",
  "countries": ["CN", "RU", "NG", "PK", "BD", "VN"],
  "priority": 10
}
```

### GDPR-Compliant (EU Only)

```json
{
  "name": "EU Only",
  "scope_type": "advertiser",
  "scope_id": "adv_xyz789",
  "mode": "allow",
  "countries": ["AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR", "DE", "GR", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL", "PL", "PT", "RO", "SK", "SI", "ES", "SE"],
  "priority": 80
}
```

### US-Only Offer

```json
{
  "name": "US Only",
  "scope_type": "offer",
  "scope_id": "off_abc123",
  "mode": "allow",
  "countries": ["US"],
  "priority": 100
}
```

---

## Error Responses

**Rule Not Found (404):**

```json
{
  "success": false,
  "error": "Geo rule not found",
  "code": "GEO_RULE_NOT_FOUND"
}
```

**Invalid Country Code (400):**

```json
{
  "success": false,
  "error": "Invalid country code",
  "code": "INVALID_COUNTRY_CODE",
  "details": {
    "invalid_codes": ["XX", "YY"]
  }
}
```

**Conflicting Rule (409):**

```json
{
  "success": false,
  "error": "Conflicting geo rule exists",
  "code": "CONFLICTING_RULE",
  "details": {
    "existing_rule_id": "geo_abc123"
  }
}
```

---

## Fraud Detection Integration

Geo rule violations are automatically logged as fraud events:

```json
{
  "category": "geo_rule_event",
  "event": "geo_block",
  "data": {
    "rule_id": "geo_abc123",
    "country": "CN",
    "offer_id": "off_xyz789",
    "ip_address": "xxx.xxx.xxx.xxx",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

View geo blocks in fraud insights:

```
GET /api/admin/fraud/insights
```

```json
{
  "geo_blocks": {
    "total": 25000,
    "today": 850,
    "top_blocked_countries": [
      {"country": "CN", "count": 8000},
      {"country": "IN", "count": 6000}
    ]
  }
}
```

---

## Next Steps

- [Fraud API](fraud.md) - View fraud detection insights
- [Offers API](offers.md) - Manage offers
- [Admin Dashboard](admin.md) - System administration

