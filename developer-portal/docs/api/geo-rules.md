# Geo Rules API

Manage geographic restrictions for offers and advertisers.

## Overview

Geo Rules allow you to control which countries can access your offers. Rules can be applied at three levels:

1. **Offer Level** - Specific to a single offer
2. **Advertiser Level** - Applies to all offers from an advertiser
3. **Global Level** - System-wide rules

### Rule Resolution Order

```
1. Offer-specific rules (highest priority)
2. Advertiser-level rules
3. Global rules
4. Default: Allow all
```

---

## GET /api/admin/geo-rules

List all geo rules.

### Request

```http
GET /api/admin/geo-rules
Authorization: Bearer <admin_token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `scope_type` | string | Filter by scope (offer, advertiser, global) |
| `status` | string | Filter by status (active, inactive) |
| `page` | integer | Page number |
| `limit` | integer | Items per page |

### Response

```json
{
  "success": true,
  "data": {
    "rules": [
      {
        "id": "geo_rule_123",
        "name": "US Traffic Only",
        "description": "Allow only US traffic for this offer",
        "scope_type": "offer",
        "scope_id": "off_123456",
        "mode": "allow",
        "countries": ["US"],
        "priority": 100,
        "status": "active",
        "created_at": "2024-01-15T10:00:00Z",
        "updated_at": "2024-12-01T12:00:00Z"
      },
      {
        "id": "geo_rule_456",
        "name": "Block High-Fraud Countries",
        "description": "Block traffic from high-fraud regions",
        "scope_type": "global",
        "scope_id": null,
        "mode": "block",
        "countries": ["XX", "YY", "ZZ"],
        "priority": 10,
        "status": "active",
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

## GET /api/admin/geo-rules/:id

Get a specific geo rule.

### Request

```http
GET /api/admin/geo-rules/geo_rule_123
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "geo_rule_123",
    "name": "US Traffic Only",
    "description": "Allow only US traffic for this offer",
    "scope_type": "offer",
    "scope_id": "off_123456",
    "scope_name": "Premium VPN Service",
    "mode": "allow",
    "countries": ["US"],
    "priority": 100,
    "status": "active",
    "stats": {
      "total_checks": 15234,
      "allowed": 12456,
      "blocked": 2778
    },
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-12-01T12:00:00Z"
  }
}
```

---

## POST /api/admin/geo-rules

Create a new geo rule.

### Request

```http
POST /api/admin/geo-rules
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "name": "North America Only",
  "description": "Allow traffic from US and Canada",
  "scope_type": "offer",
  "scope_id": "off_123456",
  "mode": "allow",
  "countries": ["US", "CA"],
  "priority": 100,
  "status": "active"
}
```

### Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Rule name |
| `description` | string | No | Rule description |
| `scope_type` | string | Yes | `offer`, `advertiser`, or `global` |
| `scope_id` | string | No | ID of offer/advertiser (null for global) |
| `mode` | string | Yes | `allow` or `block` |
| `countries` | array | Yes | ISO 3166-1 alpha-2 country codes |
| `priority` | integer | No | Higher = checked first (default: 50) |
| `status` | string | No | `active` or `inactive` (default: active) |

### Response

```json
{
  "success": true,
  "message": "Geo rule created successfully",
  "data": {
    "id": "geo_rule_789",
    "name": "North America Only",
    "scope_type": "offer",
    "scope_id": "off_123456",
    "mode": "allow",
    "countries": ["US", "CA"],
    "priority": 100,
    "status": "active",
    "created_at": "2024-12-01T12:00:00Z"
  }
}
```

### Example

```bash
curl -X POST https://api.afftok.com/api/admin/geo-rules \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "EU Countries",
    "scope_type": "advertiser",
    "scope_id": "adv_123456",
    "mode": "allow",
    "countries": ["DE", "FR", "IT", "ES", "NL", "BE", "AT", "PL"],
    "priority": 80
  }'
```

---

## PUT /api/admin/geo-rules/:id

Update an existing geo rule.

### Request

```http
PUT /api/admin/geo-rules/geo_rule_123
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "name": "Updated Rule Name",
  "countries": ["US", "CA", "GB"],
  "priority": 90
}
```

### Response

```json
{
  "success": true,
  "message": "Geo rule updated successfully",
  "data": {
    "id": "geo_rule_123",
    "name": "Updated Rule Name",
    "countries": ["US", "CA", "GB"],
    "priority": 90,
    "updated_at": "2024-12-01T12:30:00Z"
  }
}
```

---

## DELETE /api/admin/geo-rules/:id

Delete a geo rule.

### Request

```http
DELETE /api/admin/geo-rules/geo_rule_123
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "message": "Geo rule deleted successfully"
}
```

---

## GET /api/admin/offers/:id/geo-rules

Get geo rules for a specific offer.

### Request

```http
GET /api/admin/offers/off_123456/geo-rules
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "offer_id": "off_123456",
    "offer_title": "Premium VPN Service",
    "rules": [
      {
        "id": "geo_rule_123",
        "name": "US Traffic Only",
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

## POST /api/admin/geo-rules/test

Test geo rule resolution for a country.

### Request

```http
POST /api/admin/geo-rules/test
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "offer_id": "off_123456",
  "country": "US"
}
```

### Response

```json
{
  "success": true,
  "data": {
    "country": "US",
    "country_name": "United States",
    "allowed": true,
    "matched_rule": {
      "id": "geo_rule_123",
      "name": "US Traffic Only",
      "scope_type": "offer",
      "mode": "allow"
    },
    "resolution_path": [
      { "level": "offer", "rule_id": "geo_rule_123", "result": "allow" }
    ]
  }
}
```

---

## GET /api/admin/geo-rules/countries

Get list of supported country codes.

### Request

```http
GET /api/admin/geo-rules/countries
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "countries": [
      { "code": "US", "name": "United States" },
      { "code": "CA", "name": "Canada" },
      { "code": "GB", "name": "United Kingdom" },
      { "code": "DE", "name": "Germany" },
      { "code": "FR", "name": "France" }
      // ... all 249 countries
    ]
  }
}
```

---

## GET /api/admin/geo-rules/stats

Get geo rule statistics.

### Request

```http
GET /api/admin/geo-rules/stats
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "total_rules": 45,
    "active_rules": 42,
    "by_scope": {
      "offer": 30,
      "advertiser": 10,
      "global": 5
    },
    "by_mode": {
      "allow": 35,
      "block": 10
    },
    "total_checks_today": 152340,
    "blocks_today": 12456,
    "top_blocked_countries": [
      { "country": "XX", "blocks": 5234 },
      { "country": "YY", "blocks": 3456 },
      { "country": "ZZ", "blocks": 2345 }
    ]
  }
}
```

---

## Rule Modes

### Allow Mode

Only specified countries are allowed. All others are blocked.

```json
{
  "mode": "allow",
  "countries": ["US", "CA", "GB"]
}
```

### Block Mode

Specified countries are blocked. All others are allowed.

```json
{
  "mode": "block",
  "countries": ["XX", "YY", "ZZ"]
}
```

---

## Priority System

Higher priority rules are evaluated first:

| Priority Range | Typical Use |
|----------------|-------------|
| 90-100 | Critical offer-specific rules |
| 70-89 | Standard offer rules |
| 50-69 | Advertiser-level rules |
| 30-49 | Network-level rules |
| 1-29 | Global fallback rules |

---

## Caching

Geo rules are cached for performance:

- **Cache TTL**: 10 minutes
- **Cache Key**: `geo_rule:{scope_type}:{scope_id}`

### Manual Cache Refresh

```http
POST /api/admin/geo-rules/refresh-cache
Authorization: Bearer <admin_token>
```

---

## Blocked Click Behavior

When a click is blocked by geo rules:

1. Click is NOT recorded in database
2. Fraud event is logged
3. `geo_blocks_total` metric incremented
4. User is redirected to offer anyway (to avoid revealing block)

---

Next: [Webhooks API](./webhooks.md)

