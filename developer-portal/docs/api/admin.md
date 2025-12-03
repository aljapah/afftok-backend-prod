# Admin API

System administration and monitoring endpoints.

## Overview

Admin endpoints require superadmin authentication and provide:

- System health monitoring
- Performance metrics
- Log access
- Fraud intelligence
- Database diagnostics
- Stress testing

---

## Dashboard

### GET /api/admin/dashboard

Get comprehensive system dashboard.

```http
GET /api/admin/dashboard
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "system_health": {
      "status": "healthy",
      "uptime_seconds": 864000,
      "version": "2.1.0"
    },
    "database": {
      "status": "healthy",
      "latency_ms": 2.5,
      "connections": {
        "active": 15,
        "idle": 35,
        "max": 100
      }
    },
    "redis": {
      "status": "healthy",
      "latency_ms": 0.8,
      "memory_used_mb": 256,
      "keys_count": 125000
    },
    "metrics": {
      "clicks_today": 125000,
      "conversions_today": 2340,
      "postbacks_today": 1890,
      "earnings_today": 12500.00
    },
    "fraud": {
      "blocked_clicks_today": 1234,
      "bot_detections_today": 567,
      "geo_blocks_today": 890
    },
    "performance": {
      "avg_click_latency_ms": 15,
      "avg_postback_latency_ms": 45,
      "requests_per_second": 250
    }
  }
}
```

---

## Health Endpoints

### GET /api/admin/health

Full system health check.

```http
GET /api/admin/health
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "checks": {
      "database": {
        "status": "healthy",
        "latency_ms": 2.5,
        "details": "PostgreSQL 15.2 - Primary"
      },
      "database_replica": {
        "status": "healthy",
        "latency_ms": 3.1,
        "details": "PostgreSQL 15.2 - Replica"
      },
      "redis": {
        "status": "healthy",
        "latency_ms": 0.8,
        "details": "Redis 7.0.11"
      },
      "wal": {
        "status": "healthy",
        "pending_entries": 0
      },
      "streams": {
        "status": "healthy",
        "lag": 0
      }
    },
    "runtime": {
      "go_version": "1.21.5",
      "goroutines": 156,
      "memory_alloc_mb": 128,
      "memory_sys_mb": 256,
      "gc_pause_ms": 1.2
    }
  }
}
```

### GET /api/admin/connections

Get connection pool status.

```http
GET /api/admin/connections
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "database": {
      "max_open": 100,
      "open": 50,
      "in_use": 15,
      "idle": 35,
      "wait_count": 0,
      "wait_duration_ms": 0,
      "max_idle_closed": 12,
      "max_lifetime_closed": 5
    },
    "redis": {
      "pool_size": 50,
      "active": 12,
      "idle": 38,
      "stale": 0,
      "hits": 1250000,
      "misses": 12500,
      "timeouts": 0
    }
  }
}
```

---

## Metrics

### GET /api/admin/metrics

Get all performance metrics.

```http
GET /api/admin/metrics
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "clicks": {
      "total": 5000000,
      "today": 125000,
      "unique": 98000,
      "blocked": 12500,
      "from_bots": 5600,
      "rate_limited": 890
    },
    "conversions": {
      "total": 95000,
      "today": 2340,
      "approved": 2100,
      "pending": 200,
      "rejected": 40
    },
    "postbacks": {
      "total": 85000,
      "today": 1890,
      "successful": 1850,
      "failed": 40
    },
    "performance": {
      "avg_click_processing_ms": 15,
      "avg_postback_processing_ms": 45,
      "avg_db_query_ms": 2.5,
      "avg_redis_latency_ms": 0.8
    },
    "errors": {
      "total": 234,
      "today": 12,
      "by_category": {
        "database": 5,
        "redis": 2,
        "network": 3,
        "validation": 2
      }
    }
  }
}
```

### GET /api/admin/metrics/export

Export metrics as JSON file.

```http
GET /api/admin/metrics/export
Authorization: Bearer <admin_token>
```

Returns a downloadable JSON file with all metrics.

---

## Logs

### GET /api/admin/logs/recent

Get recent system logs.

```http
GET /api/admin/logs/recent?limit=100&category=click_event
Authorization: Bearer <admin_token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Number of logs (default: 100) |
| `category` | string | Filter by category |
| `level` | string | Filter by level (info, warn, error) |
| `since` | string | ISO timestamp |

### Response

```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "log_123",
        "timestamp": "2024-12-01T12:00:00.123Z",
        "level": "info",
        "category": "click_event",
        "message": "Click tracked successfully",
        "metadata": {
          "user_id": "user_123",
          "offer_id": "off_456",
          "ip": "192.168.1.1",
          "country": "US",
          "device": "mobile"
        },
        "correlation_id": "corr_abc123"
      }
    ],
    "total": 100
  }
}
```

### GET /api/admin/logs/errors

Get error logs.

```http
GET /api/admin/logs/errors?limit=50
Authorization: Bearer <admin_token>
```

### GET /api/admin/logs/fraud

Get fraud-related logs.

```http
GET /api/admin/logs/fraud?limit=50
Authorization: Bearer <admin_token>
```

### GET /api/admin/logs/categories

Get available log categories.

```http
GET /api/admin/logs/categories
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "categories": [
      "click_event",
      "conversion_event",
      "postback_event",
      "fraud_detection",
      "rate_limit_block",
      "admin_access",
      "auth_event",
      "api_key_event",
      "geo_rule_event",
      "webhook_event",
      "db_event",
      "redis_event",
      "error"
    ]
  }
}
```

### GET /api/admin/logs/ip/:ip

Get logs for a specific IP.

```http
GET /api/admin/logs/ip/192.168.1.1?limit=100
Authorization: Bearer <admin_token>
```

### GET /api/admin/logs/user/:id

Get logs for a specific user.

```http
GET /api/admin/logs/user/user_123?limit=100
Authorization: Bearer <admin_token>
```

---

## Fraud Intelligence

### GET /api/admin/fraud/insights

Get fraud intelligence dashboard.

```http
GET /api/admin/fraud/insights
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "summary": {
      "total_fraud_events_today": 1234,
      "blocked_clicks_today": 890,
      "bot_detections_today": 567,
      "geo_blocks_today": 345,
      "rate_limit_blocks_today": 123
    },
    "top_risky_ips": [
      {
        "ip": "192.168.1.1",
        "risk_score": 95,
        "fraud_events": 156,
        "last_seen": "2024-12-01T12:00:00Z",
        "indicators": ["bot_pattern", "high_frequency", "suspicious_ua"]
      }
    ],
    "recent_fraud_attempts": [
      {
        "timestamp": "2024-12-01T12:00:00Z",
        "type": "bot_detection",
        "ip": "192.168.1.1",
        "details": "Headless browser detected"
      }
    ],
    "hourly_histogram": {
      "00:00": 45,
      "01:00": 32,
      "02:00": 28
      // ... 24 hours
    },
    "by_type": {
      "bot_detection": 567,
      "geo_block": 345,
      "rate_limit": 123,
      "replay_attempt": 89,
      "invalid_signature": 56,
      "expired_link": 34
    }
  }
}
```

### POST /api/admin/fraud/block-ip

Block an IP address.

```http
POST /api/admin/fraud/block-ip
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "ip": "192.168.1.1",
  "reason": "Confirmed bot traffic",
  "duration_hours": 24
}
```

### DELETE /api/admin/fraud/block-ip/:ip

Unblock an IP address.

```http
DELETE /api/admin/fraud/block-ip/192.168.1.1
Authorization: Bearer <admin_token>
```

### GET /api/admin/fraud/blocked-ips

Get list of blocked IPs.

```http
GET /api/admin/fraud/blocked-ips
Authorization: Bearer <admin_token>
```

---

## Diagnostics

### GET /api/admin/diagnostics/redis

Get Redis diagnostics.

```http
GET /api/admin/diagnostics/redis
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "info": {
      "redis_version": "7.0.11",
      "uptime_seconds": 864000,
      "connected_clients": 45,
      "used_memory_human": "256.00M",
      "used_memory_peak_human": "512.00M"
    },
    "keyspace": {
      "db0": {
        "keys": 125000,
        "expires": 98000,
        "avg_ttl": 3600
      }
    },
    "stats": {
      "total_commands_processed": 50000000,
      "instantaneous_ops_per_sec": 1500,
      "keyspace_hits": 45000000,
      "keyspace_misses": 500000,
      "hit_rate": 98.9
    }
  }
}
```

### GET /api/admin/diagnostics/db

Get database diagnostics.

```http
GET /api/admin/diagnostics/db
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "version": "PostgreSQL 15.2",
    "size": {
      "total": "5.2 GB",
      "tables": {
        "clicks": "2.1 GB",
        "conversions": "450 MB",
        "users": "125 MB"
      }
    },
    "connections": {
      "active": 15,
      "idle": 35,
      "max": 100
    },
    "slow_queries": [
      {
        "query": "SELECT ... FROM clicks WHERE ...",
        "duration_ms": 1250,
        "calls": 5,
        "last_called": "2024-12-01T12:00:00Z"
      }
    ],
    "table_stats": [
      {
        "table": "clicks",
        "rows": 5000000,
        "dead_tuples": 12500,
        "last_vacuum": "2024-12-01T06:00:00Z",
        "last_analyze": "2024-12-01T06:00:00Z"
      }
    ]
  }
}
```

### GET /api/admin/diagnostics/system

Get system diagnostics.

```http
GET /api/admin/diagnostics/system
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "os": {
      "platform": "linux",
      "arch": "amd64",
      "cpus": 8,
      "hostname": "afftok-prod-1"
    },
    "runtime": {
      "go_version": "1.21.5",
      "goroutines": 156,
      "cgo_calls": 0
    },
    "memory": {
      "alloc_mb": 128,
      "total_alloc_mb": 5120,
      "sys_mb": 256,
      "heap_alloc_mb": 100,
      "heap_sys_mb": 200,
      "gc_cycles": 1250,
      "last_gc": "2024-12-01T12:00:00Z"
    },
    "worker_pools": {
      "click_workers": { "active": 5, "idle": 15, "queue": 12 },
      "postback_workers": { "active": 2, "idle": 8, "queue": 3 },
      "analytics_workers": { "active": 1, "idle": 4, "queue": 0 }
    }
  }
}
```

---

## Database Management

### GET /api/admin/db/stats

Get table statistics.

```http
GET /api/admin/db/stats
Authorization: Bearer <admin_token>
```

### GET /api/admin/db/indexes

Get index information.

```http
GET /api/admin/db/indexes
Authorization: Bearer <admin_token>
```

### GET /api/admin/db/vacuum-plan

Get vacuum recommendations.

```http
GET /api/admin/db/vacuum-plan
Authorization: Bearer <admin_token>
```

### GET /api/admin/db/partitions

Get partition status (for clicks table).

```http
GET /api/admin/db/partitions
Authorization: Bearer <admin_token>
```

### POST /api/admin/db/partition/create

Create a new partition.

```http
POST /api/admin/db/partition/create?year=2025&month=01
Authorization: Bearer <admin_token>
```

---

## Stress Testing

### POST /api/admin/stress/clicks

Simulate click load.

```http
POST /api/admin/stress/clicks
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "count": 1000,
  "duration_seconds": 60,
  "concurrent": 10
}
```

### Response

```json
{
  "success": true,
  "data": {
    "total_requests": 1000,
    "successful": 995,
    "failed": 5,
    "duration_seconds": 60,
    "requests_per_second": 16.67,
    "latency": {
      "min_ms": 5,
      "max_ms": 250,
      "avg_ms": 25,
      "p50_ms": 20,
      "p90_ms": 45,
      "p95_ms": 80,
      "p99_ms": 150
    }
  }
}
```

### POST /api/admin/stress/postbacks

Simulate postback load.

```http
POST /api/admin/stress/postbacks
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "count": 500,
  "duration_seconds": 60
}
```

### POST /api/admin/stress/full

Run full stress test.

```http
POST /api/admin/stress/full
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "scenarios": ["clicks", "postbacks", "webhooks"],
  "duration_seconds": 300,
  "ramp_up_seconds": 30
}
```

---

## Zero-Drop System

### GET /api/admin/zero-drop/status

Get Zero-Drop system status.

```http
GET /api/admin/zero-drop/status
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "enabled": true,
    "wal": {
      "status": "healthy",
      "pending_entries": 0,
      "last_write": "2024-12-01T12:00:00Z",
      "size_mb": 12.5
    },
    "streams": {
      "clicks": { "pending": 0, "lag": 0 },
      "conversions": { "pending": 0, "lag": 0 },
      "postbacks": { "pending": 0, "lag": 0 }
    },
    "failover_queue": {
      "size": 0,
      "last_flush": "2024-12-01T11:00:00Z"
    }
  }
}
```

### POST /api/admin/zero-drop/replay

Trigger WAL replay.

```http
POST /api/admin/zero-drop/replay
Authorization: Bearer <admin_token>
```

### POST /api/admin/zero-drop/wal/compact

Compact WAL file.

```http
POST /api/admin/zero-drop/wal/compact
Authorization: Bearer <admin_token>
```

---

## Link Signing

### GET /api/admin/link-signing/config

Get link signing configuration.

```http
GET /api/admin/link-signing/config
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "enabled": true,
    "ttl_seconds": 86400,
    "allow_legacy": false,
    "replay_protection": true,
    "nonce_cache_size": 100000
  }
}
```

### PUT /api/admin/link-signing/config

Update link signing configuration.

```http
PUT /api/admin/link-signing/config
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "ttl_seconds": 43200,
  "allow_legacy": true
}
```

### POST /api/admin/link-signing/rotate-secret

Rotate signing secret.

```http
POST /api/admin/link-signing/rotate-secret
Authorization: Bearer <admin_token>
```

### POST /api/admin/link-signing/test

Test a signed link.

```http
POST /api/admin/link-signing/test
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "link": "https://track.afftok.com/c/abc123.1701432000.xyz789.sig456"
}
```

---

## Launch Dashboard

### GET /api/admin/launch-dashboard

Get launch readiness dashboard.

```http
GET /api/admin/launch-dashboard
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "readiness": {
      "score": 98,
      "status": "ready",
      "checks": {
        "database": "pass",
        "redis": "pass",
        "wal": "pass",
        "streams": "pass",
        "edge": "pass",
        "security": "pass"
      }
    },
    "live_metrics": {
      "rps": 250,
      "error_rate": 0.1,
      "avg_latency_ms": 15
    },
    "alerts": {
      "active": 0,
      "acknowledged": 2
    }
  }
}
```

### GET /api/admin/live-metrics

Get real-time metrics stream.

```http
GET /api/admin/live-metrics
Authorization: Bearer <admin_token>
Accept: text/event-stream
```

Returns Server-Sent Events (SSE) with live metrics.

---

## Preflight Check

### GET /api/admin/preflight/check

Run pre-deployment checks.

```http
GET /api/admin/preflight/check
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "status": "pass",
    "checks": [
      { "name": "database_health", "status": "pass", "latency_ms": 2 },
      { "name": "redis_health", "status": "pass", "latency_ms": 1 },
      { "name": "wal_status", "status": "pass", "pending": 0 },
      { "name": "streams_status", "status": "pass", "lag": 0 },
      { "name": "consistency", "status": "pass", "issues": 0 },
      { "name": "security", "status": "pass", "findings": 0 },
      { "name": "performance", "status": "pass", "avg_latency_ms": 15 }
    ],
    "warnings": [],
    "ready_for_deploy": true
  }
}
```

---

Next: [Errors Reference](./errors.md)

