# Logs API

Endpoints for viewing system logs, error logs, and event history.

---

## Get Recent Logs

Retrieve recent system logs.

```
GET /api/admin/logs/recent
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page (max: 100) |
| `level` | string | Filter: `debug`, `info`, `warn`, `error` |
| `category` | string | Filter by category |
| `start_time` | string | Start timestamp (ISO 8601) |
| `end_time` | string | End timestamp (ISO 8601) |

### Response

```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "log_abc123",
        "timestamp": "2024-01-15T10:30:45.123Z",
        "level": "info",
        "category": "click_event",
        "event": "click_tracked",
        "message": "Click tracked successfully",
        "correlation_id": "corr_xyz789",
        "metadata": {
          "user_id": "usr_123",
          "offer_id": "off_456",
          "tracking_code": "tc_abc",
          "ip": "xxx.xxx.xxx.xxx",
          "country": "US",
          "device": "mobile",
          "processing_time_ms": 12
        }
      },
      {
        "id": "log_def456",
        "timestamp": "2024-01-15T10:30:44.890Z",
        "level": "warn",
        "category": "rate_limit",
        "event": "rate_limit_exceeded",
        "message": "Rate limit exceeded for IP",
        "correlation_id": "corr_abc123",
        "metadata": {
          "ip": "yyy.yyy.yyy.yyy",
          "limit": 60,
          "current": 65,
          "window": "1m"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 150000
    }
  }
}
```

---

## Get Error Logs

Retrieve error-level logs only.

```
GET /api/admin/logs/errors
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `category` | string | Filter by category |
| `error_code` | string | Filter by error code |

### Response

```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "log_err123",
        "timestamp": "2024-01-15T10:30:45.123Z",
        "level": "error",
        "category": "database",
        "event": "query_failed",
        "message": "Database query timeout",
        "error_code": "DB_TIMEOUT",
        "correlation_id": "corr_xyz789",
        "stack_trace": "at database.Query() line 123...",
        "metadata": {
          "query": "SELECT * FROM clicks WHERE...",
          "duration_ms": 30000,
          "retries": 3
        }
      }
    ],
    "summary": {
      "total_errors_today": 45,
      "by_category": {
        "database": 15,
        "redis": 10,
        "external_api": 12,
        "validation": 8
      }
    },
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 450
    }
  }
}
```

---

## Get Log Categories

List all available log categories.

```
GET /api/admin/logs/categories
```

### Response

```json
{
  "success": true,
  "data": {
    "categories": [
      {
        "name": "click_event",
        "description": "Click tracking events",
        "count_today": 150000
      },
      {
        "name": "postback_event",
        "description": "Postback/conversion events",
        "count_today": 3000
      },
      {
        "name": "fraud_detection",
        "description": "Fraud detection events",
        "count_today": 5000
      },
      {
        "name": "rate_limit",
        "description": "Rate limiting events",
        "count_today": 2500
      },
      {
        "name": "auth_event",
        "description": "Authentication events",
        "count_today": 8000
      },
      {
        "name": "api_key_event",
        "description": "API key usage events",
        "count_today": 12000
      },
      {
        "name": "webhook_event",
        "description": "Webhook delivery events",
        "count_today": 3500
      },
      {
        "name": "geo_rule_event",
        "description": "Geo rule enforcement events",
        "count_today": 4500
      },
      {
        "name": "database",
        "description": "Database operations",
        "count_today": 500000
      },
      {
        "name": "redis",
        "description": "Redis operations",
        "count_today": 2000000
      },
      {
        "name": "system",
        "description": "System events",
        "count_today": 1500
      }
    ]
  }
}
```

---

## Get Logs by Category

```
GET /api/admin/logs/category/:category
```

### Response

```json
{
  "success": true,
  "data": {
    "category": "click_event",
    "logs": [
      {
        "id": "log_abc123",
        "timestamp": "2024-01-15T10:30:45.123Z",
        "level": "info",
        "event": "click_tracked",
        "message": "Click tracked successfully",
        "metadata": {
          "user_id": "usr_123",
          "offer_id": "off_456",
          "tracking_code": "tc_abc"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 150000
    }
  }
}
```

---

## Get Logs by IP

```
GET /api/admin/logs/ip/:ip
```

### Response

```json
{
  "success": true,
  "data": {
    "ip_address": "xxx.xxx.xxx.xxx",
    "summary": {
      "total_events": 250,
      "first_seen": "2024-01-10T08:00:00Z",
      "last_seen": "2024-01-15T10:30:45Z",
      "countries": ["US"],
      "user_agents": ["Mozilla/5.0...", "Chrome/120..."],
      "risk_score": 15
    },
    "logs": [
      {
        "id": "log_abc123",
        "timestamp": "2024-01-15T10:30:45.123Z",
        "level": "info",
        "category": "click_event",
        "event": "click_tracked",
        "metadata": {
          "offer_id": "off_456",
          "country": "US"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 250
    }
  }
}
```

---

## Get Logs by User

```
GET /api/admin/logs/user/:id
```

### Response

```json
{
  "success": true,
  "data": {
    "user_id": "usr_abc123",
    "user_email": "user@example.com",
    "summary": {
      "total_events": 5000,
      "clicks": 4500,
      "conversions": 90,
      "errors": 15,
      "last_activity": "2024-01-15T10:30:45Z"
    },
    "logs": [
      {
        "id": "log_abc123",
        "timestamp": "2024-01-15T10:30:45.123Z",
        "level": "info",
        "category": "click_event",
        "event": "click_tracked",
        "metadata": {
          "offer_id": "off_456",
          "ip": "xxx.xxx.xxx.xxx"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 5000
    }
  }
}
```

---

## Search Logs

Full-text search across logs.

```
GET /api/admin/logs/search
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `q` | string | Search query |
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `level` | string | Filter by level |
| `category` | string | Filter by category |
| `start_time` | string | Start timestamp |
| `end_time` | string | End timestamp |

### Response

```json
{
  "success": true,
  "data": {
    "query": "timeout",
    "logs": [
      {
        "id": "log_err123",
        "timestamp": "2024-01-15T10:30:45.123Z",
        "level": "error",
        "category": "database",
        "event": "query_failed",
        "message": "Database query timeout after 30s",
        "highlights": ["Database query <em>timeout</em> after 30s"]
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 45
    }
  }
}
```

---

## Get Log by Correlation ID

Trace a request across all services.

```
GET /api/admin/logs/correlation/:id
```

### Response

```json
{
  "success": true,
  "data": {
    "correlation_id": "corr_xyz789",
    "request": {
      "method": "GET",
      "path": "/api/c/tc_abc123",
      "ip": "xxx.xxx.xxx.xxx",
      "user_agent": "Mozilla/5.0...",
      "started_at": "2024-01-15T10:30:45.100Z",
      "completed_at": "2024-01-15T10:30:45.150Z",
      "duration_ms": 50
    },
    "trace": [
      {
        "timestamp": "2024-01-15T10:30:45.100Z",
        "service": "api_gateway",
        "event": "request_received",
        "duration_ms": 0
      },
      {
        "timestamp": "2024-01-15T10:30:45.105Z",
        "service": "link_service",
        "event": "link_validated",
        "duration_ms": 5
      },
      {
        "timestamp": "2024-01-15T10:30:45.115Z",
        "service": "geo_service",
        "event": "geo_check_passed",
        "duration_ms": 10
      },
      {
        "timestamp": "2024-01-15T10:30:45.125Z",
        "service": "click_service",
        "event": "click_tracked",
        "duration_ms": 10
      },
      {
        "timestamp": "2024-01-15T10:30:45.135Z",
        "service": "redis",
        "event": "counter_incremented",
        "duration_ms": 10
      },
      {
        "timestamp": "2024-01-15T10:30:45.145Z",
        "service": "database",
        "event": "click_saved",
        "duration_ms": 10
      },
      {
        "timestamp": "2024-01-15T10:30:45.150Z",
        "service": "api_gateway",
        "event": "response_sent",
        "duration_ms": 5
      }
    ]
  }
}
```

---

## Export Logs

Export logs to file.

```
GET /api/admin/logs/export
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `format` | string | `json`, `csv`, `ndjson` |
| `start_time` | string | Start timestamp |
| `end_time` | string | End timestamp |
| `category` | string | Filter by category |
| `level` | string | Filter by level |

### Response

Returns a downloadable file with the specified format.

---

## Log Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| `debug` | Detailed debugging info | Development only |
| `info` | General information | Normal operations |
| `warn` | Warning conditions | Potential issues |
| `error` | Error conditions | Failures |

---

## Log Event Types

### Click Events

| Event | Description |
|-------|-------------|
| `click_tracked` | Click successfully tracked |
| `click_blocked` | Click blocked (fraud/geo/rate) |
| `click_deduplicated` | Duplicate click detected |

### Postback Events

| Event | Description |
|-------|-------------|
| `postback_received` | Postback received |
| `postback_processed` | Postback processed successfully |
| `postback_rejected` | Postback rejected |
| `postback_duplicate` | Duplicate postback detected |

### Auth Events

| Event | Description |
|-------|-------------|
| `login_success` | User logged in |
| `login_failed` | Login attempt failed |
| `token_refreshed` | JWT token refreshed |
| `token_expired` | JWT token expired |

### API Key Events

| Event | Description |
|-------|-------------|
| `api_key_used` | API key used for request |
| `api_key_invalid` | Invalid API key |
| `api_key_rate_limited` | API key rate limited |
| `api_key_revoked` | Revoked API key used |

### Webhook Events

| Event | Description |
|-------|-------------|
| `webhook_triggered` | Webhook triggered |
| `webhook_delivered` | Webhook delivered successfully |
| `webhook_failed` | Webhook delivery failed |
| `webhook_retrying` | Webhook retry scheduled |

---

## Error Responses

**Invalid Time Range (400):**

```json
{
  "success": false,
  "error": "Invalid time range",
  "code": "INVALID_TIME_RANGE"
}
```

**Log Not Found (404):**

```json
{
  "success": false,
  "error": "Log entry not found",
  "code": "LOG_NOT_FOUND"
}
```

---

## Next Steps

- [Fraud API](fraud.md) - View fraud detection logs
- [Metrics API](admin.md) - System metrics
- [Webhooks](../webhooks/overview.md) - Configure log webhooks

