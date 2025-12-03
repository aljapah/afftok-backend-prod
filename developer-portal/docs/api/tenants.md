# Tenants API

Manage multi-tenant operations for enterprise deployments.

## Overview

AffTok supports full multi-tenant architecture where each tenant has:

- Isolated data (offers, clicks, conversions, users)
- Custom branding and domains
- Separate rate limits and quotas
- Independent API keys and webhooks

---

## GET /api/admin/tenants

List all tenants.

### Request

```http
GET /api/admin/tenants
Authorization: Bearer <superadmin_token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status (active, suspended, pending) |
| `plan` | string | Filter by plan (free, starter, pro, enterprise) |
| `page` | integer | Page number |
| `limit` | integer | Items per page |

### Response

```json
{
  "success": true,
  "data": {
    "tenants": [
      {
        "id": "tenant_123",
        "name": "Acme Corporation",
        "slug": "acme",
        "status": "active",
        "plan": "pro",
        "domains": ["acme.afftok.com", "tracking.acme.com"],
        "stats": {
          "users": 150,
          "offers": 45,
          "clicks_today": 12500,
          "conversions_today": 234
        },
        "created_at": "2024-01-15T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 35
    }
  }
}
```

---

## GET /api/admin/tenants/:id

Get tenant details.

### Request

```http
GET /api/admin/tenants/tenant_123
Authorization: Bearer <superadmin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "tenant_123",
    "name": "Acme Corporation",
    "slug": "acme",
    "status": "active",
    "plan": "pro",
    "domains": [
      {
        "domain": "acme.afftok.com",
        "primary": true,
        "verified": true
      },
      {
        "domain": "tracking.acme.com",
        "primary": false,
        "verified": true
      }
    ],
    "branding": {
      "logo_url": "https://cdn.acme.com/logo.png",
      "primary_color": "#3B82F6",
      "secondary_color": "#1E40AF",
      "custom_css": null
    },
    "settings": {
      "default_currency": "USD",
      "timezone": "America/New_York",
      "language": "en",
      "email_notifications": true
    },
    "limits": {
      "max_users": 500,
      "max_offers": 100,
      "max_clicks_per_day": 100000,
      "max_api_keys": 20
    },
    "features": {
      "geo_rules": true,
      "webhooks": true,
      "api_keys": true,
      "custom_domains": true,
      "white_label": true,
      "advanced_analytics": true
    },
    "stats": {
      "users": 150,
      "offers": 45,
      "total_clicks": 1250000,
      "total_conversions": 23400,
      "total_earnings": 125000.00
    },
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-12-01T12:00:00Z"
  }
}
```

---

## POST /api/admin/tenants

Create a new tenant.

### Request

```http
POST /api/admin/tenants
Authorization: Bearer <superadmin_token>
Content-Type: application/json
```

```json
{
  "name": "New Company Inc",
  "slug": "newcompany",
  "plan": "starter",
  "admin_email": "admin@newcompany.com",
  "settings": {
    "default_currency": "EUR",
    "timezone": "Europe/London"
  }
}
```

### Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Tenant display name |
| `slug` | string | Yes | URL-safe identifier (unique) |
| `plan` | string | No | Plan tier (default: free) |
| `admin_email` | string | Yes | Initial admin email |
| `settings` | object | No | Tenant settings |

### Response

```json
{
  "success": true,
  "message": "Tenant created successfully",
  "data": {
    "id": "tenant_456",
    "name": "New Company Inc",
    "slug": "newcompany",
    "status": "pending",
    "plan": "starter",
    "default_domain": "newcompany.afftok.com",
    "admin_invite_sent": true,
    "created_at": "2024-12-01T12:00:00Z"
  }
}
```

---

## PUT /api/admin/tenants/:id

Update tenant details.

### Request

```http
PUT /api/admin/tenants/tenant_123
Authorization: Bearer <superadmin_token>
Content-Type: application/json
```

```json
{
  "name": "Acme Corp Updated",
  "settings": {
    "timezone": "America/Los_Angeles"
  }
}
```

### Response

```json
{
  "success": true,
  "message": "Tenant updated successfully",
  "data": {
    "id": "tenant_123",
    "name": "Acme Corp Updated",
    "updated_at": "2024-12-01T12:30:00Z"
  }
}
```

---

## POST /api/admin/tenants/:id/suspend

Suspend a tenant.

### Request

```http
POST /api/admin/tenants/tenant_123/suspend
Authorization: Bearer <superadmin_token>
Content-Type: application/json
```

```json
{
  "reason": "Payment overdue",
  "notify_admin": true
}
```

### Response

```json
{
  "success": true,
  "message": "Tenant suspended",
  "data": {
    "id": "tenant_123",
    "status": "suspended",
    "suspended_at": "2024-12-01T12:00:00Z",
    "suspension_reason": "Payment overdue"
  }
}
```

### Effects of Suspension

- All tracking links redirect to suspension page
- API keys are blocked
- Admin access is read-only
- Data is preserved

---

## POST /api/admin/tenants/:id/activate

Activate a suspended tenant.

### Request

```http
POST /api/admin/tenants/tenant_123/activate
Authorization: Bearer <superadmin_token>
```

### Response

```json
{
  "success": true,
  "message": "Tenant activated",
  "data": {
    "id": "tenant_123",
    "status": "active",
    "activated_at": "2024-12-01T12:00:00Z"
  }
}
```

---

## GET /api/admin/tenants/:id/stats

Get tenant statistics.

### Request

```http
GET /api/admin/tenants/tenant_123/stats?period=30d
Authorization: Bearer <superadmin_token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | Time period (24h, 7d, 30d, 90d, all) |

### Response

```json
{
  "success": true,
  "data": {
    "tenant_id": "tenant_123",
    "period": "30d",
    "summary": {
      "total_clicks": 125000,
      "unique_clicks": 98000,
      "conversions": 2340,
      "conversion_rate": 1.87,
      "earnings": 12500.00,
      "avg_earnings_per_conversion": 5.34
    },
    "usage": {
      "users": { "current": 150, "limit": 500, "percentage": 30 },
      "offers": { "current": 45, "limit": 100, "percentage": 45 },
      "clicks_today": { "current": 12500, "limit": 100000, "percentage": 12.5 },
      "api_keys": { "current": 8, "limit": 20, "percentage": 40 }
    },
    "daily_stats": [
      {
        "date": "2024-12-01",
        "clicks": 4200,
        "conversions": 78,
        "earnings": 420.00
      }
      // ... more days
    ]
  }
}
```

---

## Domains Management

### GET /api/admin/tenants/:id/domains

List tenant domains.

```http
GET /api/admin/tenants/tenant_123/domains
Authorization: Bearer <superadmin_token>
```

### POST /api/admin/tenants/:id/domains

Add a custom domain.

```http
POST /api/admin/tenants/tenant_123/domains
Authorization: Bearer <superadmin_token>
Content-Type: application/json
```

```json
{
  "domain": "track.acme.com",
  "make_primary": false
}
```

### Response

```json
{
  "success": true,
  "message": "Domain added",
  "data": {
    "domain": "track.acme.com",
    "verified": false,
    "verification_record": {
      "type": "CNAME",
      "name": "track",
      "value": "tenant_123.afftok-edge.com"
    }
  }
}
```

### DELETE /api/admin/tenants/:id/domains/:domain

Remove a domain.

```http
DELETE /api/admin/tenants/tenant_123/domains/track.acme.com
Authorization: Bearer <superadmin_token>
```

---

## Branding

### PUT /api/admin/tenants/:id/branding

Update tenant branding.

```http
PUT /api/admin/tenants/tenant_123/branding
Authorization: Bearer <superadmin_token>
Content-Type: application/json
```

```json
{
  "logo_url": "https://cdn.acme.com/new-logo.png",
  "primary_color": "#10B981",
  "secondary_color": "#059669",
  "favicon_url": "https://cdn.acme.com/favicon.ico",
  "custom_css": ".header { background: linear-gradient(...); }"
}
```

---

## Plan Management

### PUT /api/admin/tenants/:id/plan

Change tenant plan.

```http
PUT /api/admin/tenants/tenant_123/plan
Authorization: Bearer <superadmin_token>
Content-Type: application/json
```

```json
{
  "plan": "enterprise",
  "effective_date": "2024-12-01T00:00:00Z"
}
```

### Response

```json
{
  "success": true,
  "message": "Plan updated",
  "data": {
    "id": "tenant_123",
    "previous_plan": "pro",
    "new_plan": "enterprise",
    "new_limits": {
      "max_users": 10000,
      "max_offers": 1000,
      "max_clicks_per_day": 10000000,
      "max_api_keys": 100
    },
    "new_features": {
      "geo_rules": true,
      "webhooks": true,
      "api_keys": true,
      "custom_domains": true,
      "white_label": true,
      "advanced_analytics": true,
      "dedicated_support": true,
      "sla_guarantee": true
    }
  }
}
```

### Available Plans

| Plan | Max Users | Max Offers | Daily Clicks | API Keys |
|------|-----------|------------|--------------|----------|
| free | 10 | 5 | 1,000 | 2 |
| starter | 50 | 20 | 10,000 | 5 |
| pro | 500 | 100 | 100,000 | 20 |
| enterprise | 10,000 | 1,000 | 10,000,000 | 100 |

---

## Features

### GET /api/admin/tenants/:id/features

Get tenant features.

```http
GET /api/admin/tenants/tenant_123/features
Authorization: Bearer <superadmin_token>
```

### PUT /api/admin/tenants/:id/features

Update tenant features.

```http
PUT /api/admin/tenants/tenant_123/features
Authorization: Bearer <superadmin_token>
Content-Type: application/json
```

```json
{
  "geo_rules": true,
  "webhooks": true,
  "advanced_analytics": true,
  "white_label": false
}
```

---

## Audit Logs

### GET /api/admin/tenants/:id/audit-logs

Get tenant audit logs.

```http
GET /api/admin/tenants/tenant_123/audit-logs?limit=100
Authorization: Bearer <superadmin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "audit_123",
        "action": "plan_changed",
        "actor_id": "admin_456",
        "actor_email": "admin@afftok.com",
        "details": {
          "previous_plan": "pro",
          "new_plan": "enterprise"
        },
        "ip_address": "192.168.1.1",
        "created_at": "2024-12-01T12:00:00Z"
      }
    ]
  }
}
```

---

## Tenant Report

### GET /api/admin/tenants/report

Get aggregated tenant report.

```http
GET /api/admin/tenants/report
Authorization: Bearer <superadmin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "total_tenants": 35,
    "by_status": {
      "active": 30,
      "suspended": 3,
      "pending": 2
    },
    "by_plan": {
      "free": 10,
      "starter": 12,
      "pro": 10,
      "enterprise": 3
    },
    "total_users": 2500,
    "total_offers": 450,
    "total_clicks_today": 250000,
    "total_conversions_today": 4500,
    "total_earnings_today": 25000.00,
    "top_tenants": [
      {
        "id": "tenant_123",
        "name": "Acme Corporation",
        "clicks_today": 50000,
        "conversions_today": 1200
      }
    ]
  }
}
```

---

## Tenant Resolution

Tenants are resolved from requests using:

1. **X-Tenant-ID Header** - Direct tenant ID
2. **Subdomain** - e.g., `acme.afftok.com` → tenant `acme`
3. **Custom Domain** - Mapped domain lookup
4. **API Key** - Tenant bound to API key
5. **JWT Token** - Tenant ID in token claims

### Resolution Priority

```
X-Tenant-ID > Custom Domain > Subdomain > API Key > JWT
```

---

## Rate Limits

Each tenant has independent rate limits:

| Endpoint | Free | Starter | Pro | Enterprise |
|----------|------|---------|-----|------------|
| Clicks/min | 100 | 500 | 2000 | 10000 |
| API calls/min | 60 | 300 | 1000 | 5000 |
| Postbacks/min | 50 | 200 | 1000 | 5000 |

---

Next: [Admin API](./admin.md)

