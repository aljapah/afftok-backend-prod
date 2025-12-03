# Tenants API

Admin endpoints for managing tenants in the multi-tenant AffTok platform.

> **Note:** These endpoints require admin authentication.

---

## Overview

AffTok supports multi-tenant architecture where each tenant operates in complete isolation with their own:
- Users and teams
- Offers and networks
- Tracking links and statistics
- API keys and webhooks
- Geo rules and fraud settings
- Branding and settings

---

## List All Tenants

```
GET /api/admin/tenants
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `status` | string | Filter: `active`, `suspended`, `pending` |
| `plan` | string | Filter: `free`, `starter`, `pro`, `enterprise` |
| `search` | string | Search by name or slug |

### Response

```json
{
  "success": true,
  "data": {
    "tenants": [
      {
        "id": "tenant_abc123",
        "name": "Acme Corp",
        "slug": "acme-corp",
        "status": "active",
        "plan": "pro",
        "owner_email": "admin@acme.com",
        "domains": ["acme.afftok.com", "tracking.acme.com"],
        "stats": {
          "users": 25,
          "offers": 50,
          "clicks_today": 15000,
          "conversions_today": 300
        },
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150
    }
  }
}
```

---

## Create Tenant

```
POST /api/admin/tenants
```

### Request Body

```json
{
  "name": "Acme Corp",
  "slug": "acme-corp",
  "owner_email": "admin@acme.com",
  "plan": "pro",
  "settings": {
    "timezone": "America/New_York",
    "currency": "USD",
    "language": "en"
  },
  "limits": {
    "max_users": 50,
    "max_offers": 100,
    "max_clicks_per_day": 100000,
    "max_api_keys": 10
  },
  "features": {
    "geo_rules_enabled": true,
    "webhooks_enabled": true,
    "api_access_enabled": true,
    "smart_routing_enabled": true,
    "zero_drop_enabled": true
  },
  "branding": {
    "logo_url": "https://cdn.acme.com/logo.png",
    "primary_color": "#3B82F6",
    "favicon_url": "https://cdn.acme.com/favicon.ico"
  }
}
```

### Response

```json
{
  "success": true,
  "message": "Tenant created successfully",
  "data": {
    "id": "tenant_abc123",
    "name": "Acme Corp",
    "slug": "acme-corp",
    "status": "active",
    "plan": "pro",
    "api_key": "afftok_tenant_sk_xxxxx",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## Get Tenant Details

```
GET /api/admin/tenants/:id
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "tenant_abc123",
    "name": "Acme Corp",
    "slug": "acme-corp",
    "status": "active",
    "plan": "pro",
    "owner_email": "admin@acme.com",
    "settings": {
      "timezone": "America/New_York",
      "currency": "USD",
      "language": "en",
      "notification_email": "alerts@acme.com"
    },
    "limits": {
      "max_users": 50,
      "max_offers": 100,
      "max_clicks_per_day": 100000,
      "max_api_keys": 10
    },
    "usage": {
      "users": 25,
      "offers": 48,
      "clicks_today": 15000,
      "api_keys": 5
    },
    "features": {
      "geo_rules_enabled": true,
      "webhooks_enabled": true,
      "api_access_enabled": true,
      "smart_routing_enabled": true,
      "zero_drop_enabled": true,
      "custom_domain_enabled": true
    },
    "branding": {
      "logo_url": "https://cdn.acme.com/logo.png",
      "primary_color": "#3B82F6",
      "secondary_color": "#1E40AF",
      "favicon_url": "https://cdn.acme.com/favicon.ico"
    },
    "domains": [
      {
        "domain": "acme.afftok.com",
        "type": "subdomain",
        "verified": true,
        "primary": true
      },
      {
        "domain": "tracking.acme.com",
        "type": "custom",
        "verified": true,
        "primary": false
      }
    ],
    "stats": {
      "total_clicks": 1500000,
      "total_conversions": 30000,
      "total_earnings": 150000.00,
      "this_month_clicks": 150000,
      "this_month_conversions": 3000
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## Update Tenant

```
PUT /api/admin/tenants/:id
```

### Request Body

```json
{
  "name": "Acme Corporation",
  "settings": {
    "timezone": "Europe/London"
  },
  "limits": {
    "max_users": 100
  }
}
```

### Response

```json
{
  "success": true,
  "message": "Tenant updated successfully"
}
```

---

## Suspend Tenant

```
POST /api/admin/tenants/:id/suspend
```

### Request Body

```json
{
  "reason": "Payment overdue",
  "notify_owner": true
}
```

### Response

```json
{
  "success": true,
  "message": "Tenant suspended"
}
```

---

## Activate Tenant

```
POST /api/admin/tenants/:id/activate
```

### Response

```json
{
  "success": true,
  "message": "Tenant activated"
}
```

---

## Get Tenant Stats

```
GET /api/admin/tenants/:id/stats
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | `today`, `week`, `month`, `year`, `all` |

### Response

```json
{
  "success": true,
  "data": {
    "tenant_id": "tenant_abc123",
    "period": "month",
    "clicks": {
      "total": 150000,
      "unique": 125000,
      "by_day": [
        {"date": "2024-01-01", "count": 5000},
        {"date": "2024-01-02", "count": 5200}
      ]
    },
    "conversions": {
      "total": 3000,
      "approved": 2850,
      "pending": 100,
      "rejected": 50
    },
    "earnings": {
      "total": 15000.00,
      "pending": 500.00
    },
    "top_offers": [
      {"offer_id": "off_123", "title": "App Install", "conversions": 500}
    ],
    "top_users": [
      {"user_id": "usr_123", "name": "John", "conversions": 200}
    ]
  }
}
```

---

## Manage Tenant Domains

### Add Domain

```
POST /api/admin/tenants/:id/domains
```

```json
{
  "domain": "track.acme.com",
  "type": "custom"
}
```

### Remove Domain

```
DELETE /api/admin/tenants/:id/domains/:domain
```

### Verify Domain

```
POST /api/admin/tenants/:id/domains/:domain/verify
```

---

## Update Tenant Branding

```
PUT /api/admin/tenants/:id/branding
```

```json
{
  "logo_url": "https://cdn.acme.com/new-logo.png",
  "primary_color": "#10B981",
  "secondary_color": "#059669",
  "favicon_url": "https://cdn.acme.com/favicon.ico",
  "custom_css": ".header { background: #10B981; }"
}
```

---

## Change Tenant Plan

```
PUT /api/admin/tenants/:id/plan
```

```json
{
  "plan": "enterprise",
  "effective_date": "2024-02-01T00:00:00Z"
}
```

### Available Plans

| Plan | Max Users | Max Offers | Max Clicks/Day | Features |
|------|-----------|------------|----------------|----------|
| `free` | 5 | 10 | 10,000 | Basic |
| `starter` | 25 | 50 | 100,000 | + API Access |
| `pro` | 100 | 200 | 500,000 | + Smart Routing |
| `enterprise` | Unlimited | Unlimited | Unlimited | All Features |

---

## Update Tenant Settings

```
PUT /api/admin/tenants/:id/settings
```

```json
{
  "timezone": "America/Los_Angeles",
  "currency": "EUR",
  "language": "de",
  "notification_email": "ops@acme.com",
  "webhook_secret": "new_secret_key",
  "default_payout_method": "wire"
}
```

---

## Get Tenant Features

```
GET /api/admin/tenants/:id/features
```

### Response

```json
{
  "success": true,
  "data": {
    "features": {
      "geo_rules_enabled": true,
      "webhooks_enabled": true,
      "api_access_enabled": true,
      "smart_routing_enabled": true,
      "zero_drop_enabled": true,
      "custom_domain_enabled": true,
      "white_label_enabled": false,
      "advanced_analytics_enabled": true,
      "team_management_enabled": true,
      "sso_enabled": false
    }
  }
}
```

### Update Features

```
PUT /api/admin/tenants/:id/features
```

```json
{
  "sso_enabled": true,
  "white_label_enabled": true
}
```

---

## Get Tenant Audit Logs

```
GET /api/admin/tenants/:id/audit-logs
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `action` | string | Filter by action type |
| `user_id` | string | Filter by user |
| `start_date` | string | Start date |
| `end_date` | string | End date |

### Response

```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "log_abc123",
        "action": "offer.created",
        "user_id": "usr_123",
        "user_email": "john@acme.com",
        "resource_type": "offer",
        "resource_id": "off_xyz789",
        "details": {
          "offer_title": "New App Install"
        },
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 1500
    }
  }
}
```

---

## Get Tenants Report

```
GET /api/admin/tenants/report
```

### Response

```json
{
  "success": true,
  "data": {
    "summary": {
      "total_tenants": 150,
      "active_tenants": 140,
      "suspended_tenants": 8,
      "pending_tenants": 2
    },
    "by_plan": {
      "free": 50,
      "starter": 60,
      "pro": 30,
      "enterprise": 10
    },
    "growth": {
      "this_month": 12,
      "last_month": 10,
      "growth_rate": 20.0
    },
    "revenue": {
      "mrr": 15000.00,
      "arr": 180000.00
    },
    "top_tenants": [
      {
        "id": "tenant_abc123",
        "name": "Acme Corp",
        "clicks": 1500000,
        "conversions": 30000
      }
    ]
  }
}
```

---

## Delete Tenant

```
DELETE /api/admin/tenants/:id
```

> **Warning:** This action is irreversible and will delete all tenant data.

### Request Body

```json
{
  "confirm": true,
  "reason": "Customer requested deletion"
}
```

### Response

```json
{
  "success": true,
  "message": "Tenant scheduled for deletion"
}
```

---

## Error Responses

**Tenant Not Found (404):**

```json
{
  "success": false,
  "error": "Tenant not found",
  "code": "TENANT_NOT_FOUND"
}
```

**Slug Already Exists (409):**

```json
{
  "success": false,
  "error": "Tenant slug already exists",
  "code": "SLUG_EXISTS"
}
```

**Limit Exceeded (400):**

```json
{
  "success": false,
  "error": "Tenant has exceeded usage limits",
  "code": "LIMIT_EXCEEDED"
}
```

---

## Next Steps

- [Geo Rules API](geo-rules.md) - Manage tenant geo rules
- [API Keys API](api-keys.md) - Manage tenant API keys
- [Admin Dashboard](admin.md) - System administration

