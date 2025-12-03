# Admin API

Administrative endpoints for system management, monitoring, and configuration.

> **Note:** All admin endpoints require admin authentication.

---

## Dashboard

Get comprehensive system dashboard data.

```
GET /api/admin/dashboard
```

### Response

```json
{
  "success": true,
  "data": {
    "timestamp": "2024-01-15T10:30:00Z",
    "uptime_seconds": 864000,
    "health": {
      "status": "healthy",
      "database": {
        "status": "healthy",
        "latency_ms": 2,
        "connections": {
          "active": 25,
          "idle": 75,
          "max": 100
        }
      },
      "redis": {
        "status": "healthy",
        "latency_ms": 1,
        "memory_used_mb": 256,
        "memory_max_mb": 1024
      }
    },
    "metrics": {
      "clicks": {
        "total": 15000000,
        "today": 150000,
        "per_second": 52.3,
        "blocked": 12500
      },
      "conversions": {
        "total": 300000,
        "today": 3000,
        "approved": 285000,
        "pending": 10000,
        "rejected": 5000
      },
      "postbacks": {
        "total": 350000,
        "today": 3500,
        "success_rate": 98.5
      },
      "earnings": {
        "total": 1500000.00,
        "today": 15000.00,
        "pending": 50000.00
      }
    },
    "fraud": {
      "blocked_today": 12500,
      "top_reason": "bot_detection",
      "risk_score_avg": 35.5
    },
    "system": {
      "goroutines": 150,
      "memory_mb": 512,
      "cpu_percent": 25.5
    },
    "tenants": {
      "total": 150,
      "active": 140
    }
  }
}
```

---

## System Health

```
GET /api/admin/health
```

### Response

```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "components": {
      "database": {
        "status": "healthy",
        "latency_ms": 2,
        "version": "PostgreSQL 15.2"
      },
      "redis": {
        "status": "healthy",
        "latency_ms": 1,
        "version": "7.0.5"
      },
      "wal": {
        "status": "healthy",
        "pending_entries": 0
      },
      "streams": {
        "status": "healthy",
        "lag": 0
      },
      "workers": {
        "status": "healthy",
        "active": 10,
        "queue_depth": 25
      }
    },
    "runtime": {
      "go_version": "1.21.5",
      "goroutines": 150,
      "memory_alloc_mb": 256,
      "memory_sys_mb": 512,
      "gc_pause_ms": 0.5
    }
  }
}
```

---

## System Metrics

```
GET /api/admin/metrics
```

### Response

```json
{
  "success": true,
  "data": {
    "timestamp": "2024-01-15T10:30:00Z",
    "counters": {
      "clicks_total": 15000000,
      "clicks_unique": 12500000,
      "clicks_blocked": 500000,
      "conversions_total": 300000,
      "postbacks_total": 350000,
      "postbacks_failed": 5250,
      "api_requests_total": 50000000,
      "auth_attempts": 100000,
      "auth_failures": 500
    },
    "gauges": {
      "active_connections_db": 25,
      "active_connections_redis": 50,
      "queue_depth_clicks": 100,
      "queue_depth_webhooks": 25,
      "memory_mb": 512,
      "goroutines": 150
    },
    "latencies": {
      "click_processing_ms": {
        "p50": 5,
        "p90": 15,
        "p95": 25,
        "p99": 50,
        "avg": 8
      },
      "postback_processing_ms": {
        "p50": 10,
        "p90": 30,
        "p95": 50,
        "p99": 100,
        "avg": 15
      },
      "db_query_ms": {
        "p50": 2,
        "p90": 5,
        "p95": 10,
        "p99": 25,
        "avg": 3
      },
      "redis_operation_ms": {
        "p50": 0.5,
        "p90": 1,
        "p95": 2,
        "p99": 5,
        "avg": 0.8
      }
    },
    "rates": {
      "clicks_per_second": 52.3,
      "postbacks_per_second": 1.2,
      "api_requests_per_second": 173.6
    }
  }
}
```

---

## Export Metrics

```
GET /api/admin/metrics/export
```

Returns metrics as downloadable JSON file.

---

## Connection Status

```
GET /api/admin/connections
```

### Response

```json
{
  "success": true,
  "data": {
    "database": {
      "primary": {
        "status": "connected",
        "host": "db-primary.example.com",
        "active": 25,
        "idle": 75,
        "max": 100,
        "wait_count": 0
      },
      "replica": {
        "status": "connected",
        "host": "db-replica.example.com",
        "active": 15,
        "idle": 35,
        "max": 50
      }
    },
    "redis": {
      "status": "connected",
      "host": "redis.example.com",
      "pool_size": 100,
      "active": 50,
      "idle": 50
    }
  }
}
```

---

## Database Management

### Get Database Stats

```
GET /api/admin/db/stats
```

### Response

```json
{
  "success": true,
  "data": {
    "tables": [
      {
        "name": "clicks",
        "rows": 15000000,
        "size_mb": 2500,
        "index_size_mb": 500,
        "dead_tuples": 1500,
        "last_vacuum": "2024-01-15T06:00:00Z",
        "last_analyze": "2024-01-15T06:00:00Z"
      },
      {
        "name": "conversions",
        "rows": 300000,
        "size_mb": 150,
        "index_size_mb": 30
      }
    ],
    "total_size_mb": 5000,
    "connections": {
      "active": 25,
      "idle": 75,
      "max": 100
    }
  }
}
```

### Get Index Status

```
GET /api/admin/db/indexes
```

### Get Partition Status

```
GET /api/admin/db/partitions
```

### Create Partition

```
POST /api/admin/db/partition/create
```

```json
{
  "year": 2024,
  "month": 2
}
```

---

## Redis Diagnostics

```
GET /api/admin/diagnostics/redis
```

### Response

```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "7.0.5",
    "uptime_seconds": 864000,
    "memory": {
      "used_mb": 256,
      "peak_mb": 300,
      "max_mb": 1024,
      "fragmentation_ratio": 1.05
    },
    "clients": {
      "connected": 50,
      "blocked": 0,
      "max": 10000
    },
    "stats": {
      "total_commands": 500000000,
      "ops_per_second": 5800,
      "hits": 450000000,
      "misses": 50000000,
      "hit_rate": 90.0
    },
    "keys": {
      "total": 150000,
      "expires": 100000,
      "avg_ttl_seconds": 3600
    }
  }
}
```

---

## System Diagnostics

```
GET /api/admin/diagnostics/system
```

### Response

```json
{
  "success": true,
  "data": {
    "os": "linux",
    "arch": "amd64",
    "hostname": "afftok-api-1",
    "cpu": {
      "cores": 8,
      "usage_percent": 25.5
    },
    "memory": {
      "total_mb": 16384,
      "used_mb": 8192,
      "free_mb": 8192,
      "usage_percent": 50.0
    },
    "disk": {
      "total_gb": 500,
      "used_gb": 250,
      "free_gb": 250,
      "usage_percent": 50.0
    },
    "network": {
      "interfaces": [
        {
          "name": "eth0",
          "ip": "10.0.0.1",
          "rx_bytes": 1000000000,
          "tx_bytes": 500000000
        }
      ]
    },
    "process": {
      "pid": 12345,
      "uptime_seconds": 864000,
      "goroutines": 150,
      "threads": 20
    }
  }
}
```

---

## Stress Testing

### Run Click Stress Test

```
POST /api/admin/stress/clicks
```

```json
{
  "count": 10000,
  "concurrency": 100,
  "duration_seconds": 60
}
```

### Run Full Stress Test

```
POST /api/admin/stress/full
```

### Get Worker Pool Stats

```
GET /api/admin/stress/pools
```

### Get Cache Stats

```
GET /api/admin/stress/cache
```

---

## Zero-Drop System

### Get Zero-Drop Status

```
GET /api/admin/zero-drop/status
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
      "last_write": "2024-01-15T10:30:00Z",
      "size_mb": 50
    },
    "streams": {
      "clicks": {
        "pending": 0,
        "lag": 0
      },
      "conversions": {
        "pending": 0,
        "lag": 0
      },
      "postbacks": {
        "pending": 0,
        "lag": 0
      }
    },
    "failover_queue": {
      "size": 0,
      "last_flush": "2024-01-15T10:00:00Z"
    },
    "metrics": {
      "dropped_clicks": 0,
      "dropped_postbacks": 0,
      "wal_replayed": 1500,
      "recovered_events": 1500
    }
  }
}
```

### Replay WAL

```
POST /api/admin/zero-drop/replay
```

### Compact WAL

```
POST /api/admin/zero-drop/wal/compact
```

---

## Edge CDN Management

### Get Edge Status

```
GET /api/admin/edge/status
```

### Response

```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "regions": [
      {
        "name": "US-East",
        "status": "healthy",
        "requests_per_second": 1500,
        "latency_ms": 5
      },
      {
        "name": "EU-West",
        "status": "healthy",
        "requests_per_second": 800,
        "latency_ms": 8
      }
    ],
    "total_requests": 50000000,
    "cache_hit_rate": 95.5,
    "bot_blocks": 125000
  }
}
```

### Get Edge Stats

```
GET /api/admin/edge/stats
```

### Flush Edge Queue

```
POST /api/admin/edge/queue/flush
```

---

## Launch Dashboard

```
GET /api/admin/launch-dashboard
```

### Response

```json
{
  "success": true,
  "data": {
    "status": "production",
    "health_score": 98,
    "metrics": {
      "global_rps": 5000,
      "edge_rps": 4500,
      "backend_rps": 500,
      "avg_latency_ms": 15,
      "error_rate": 0.01
    },
    "alerts": {
      "active": 0,
      "acknowledged": 2,
      "resolved_today": 5
    },
    "capacity": {
      "cpu_usage": 25,
      "memory_usage": 50,
      "db_connections": 25,
      "redis_memory": 25
    },
    "trends": {
      "traffic": "stable",
      "errors": "decreasing",
      "latency": "stable"
    }
  }
}
```

---

## Live Metrics

```
GET /api/admin/live-metrics
```

Real-time metrics stream (Server-Sent Events).

```javascript
const eventSource = new EventSource('/api/admin/live-metrics');

eventSource.onmessage = (event) => {
  const metrics = JSON.parse(event.data);
  console.log('Live metrics:', metrics);
};
```

---

## Logging Modes

### Get Current Mode

```
GET /api/admin/logging/mode
```

### Set Logging Mode

```
PUT /api/admin/logging/mode
```

```json
{
  "mode": "verbose",
  "duration_minutes": 10
}
```

### Available Modes

| Mode | Description |
|------|-------------|
| `normal` | Standard logging |
| `verbose` | Detailed logging |
| `critical` | Maximum logging (auto-resets) |

---

## Alerts

### Get Active Alerts

```
GET /api/admin/alerts/active
```

### Get Alert History

```
GET /api/admin/alerts/history
```

### Acknowledge Alert

```
POST /api/admin/alerts/:id/acknowledge
```

### Configure Alert Thresholds

```
PUT /api/admin/alerts/thresholds
```

```json
{
  "cpu_percent": 80,
  "memory_percent": 85,
  "error_rate_percent": 1,
  "latency_ms": 100,
  "queue_depth": 1000
}
```

---

## Load Testing

### Run Load Test

```
POST /api/admin/loadtest/run
```

```json
{
  "scenario": "clicks",
  "requests": 100000,
  "concurrency": 500,
  "duration_seconds": 60
}
```

### Get Load Test Report

```
GET /api/admin/loadtest/report/:id
```

---

## Error Responses

**Unauthorized (401):**

```json
{
  "success": false,
  "error": "Admin authentication required",
  "code": "ADMIN_AUTH_REQUIRED"
}
```

**Forbidden (403):**

```json
{
  "success": false,
  "error": "Insufficient permissions",
  "code": "INSUFFICIENT_PERMISSIONS"
}
```

---

## Next Steps

- [Tenants API](tenants.md) - Manage tenants
- [Fraud API](fraud.md) - Fraud detection
- [Logs API](logs.md) - System logs

