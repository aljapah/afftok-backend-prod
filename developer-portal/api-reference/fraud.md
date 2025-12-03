# Fraud Detection API

Endpoints for viewing fraud detection insights, managing blocked IPs, and monitoring suspicious activity.

---

## Overview

AffTok's fraud detection system monitors:

- **Bot Detection** - Automated traffic patterns
- **Click Fraud** - Duplicate clicks, click injection
- **Geo Violations** - Traffic from blocked countries
- **Rate Limit Violations** - Excessive request rates
- **Link Tampering** - Invalid signatures, expired links
- **Replay Attacks** - Reused tracking links
- **API Key Abuse** - Unauthorized or excessive API usage

---

## Get Fraud Insights

Comprehensive fraud detection dashboard.

```
GET /api/admin/fraud/insights
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | `today`, `week`, `month`, `all` |
| `tenant_id` | string | Filter by tenant (admin only) |

### Response

```json
{
  "success": true,
  "data": {
    "period": "month",
    "summary": {
      "total_events": 15000,
      "blocked_clicks": 12500,
      "blocked_postbacks": 500,
      "blocked_api_requests": 2000,
      "risk_score_avg": 35.5
    },
    "by_category": {
      "bot_detection": 5000,
      "geo_block": 4500,
      "rate_limit": 2500,
      "invalid_signature": 1500,
      "replay_attempt": 800,
      "expired_link": 500,
      "api_key_violation": 200
    },
    "top_risky_ips": [
      {
        "ip": "xxx.xxx.xxx.xxx",
        "events": 2500,
        "risk_score": 95,
        "countries": ["CN"],
        "last_seen": "2024-01-15T10:30:00Z",
        "blocked": true
      },
      {
        "ip": "yyy.yyy.yyy.yyy",
        "events": 1800,
        "risk_score": 88,
        "countries": ["RU"],
        "last_seen": "2024-01-15T10:25:00Z",
        "blocked": false
      }
    ],
    "hourly_histogram": [
      {"hour": "00:00", "events": 450},
      {"hour": "01:00", "events": 380},
      {"hour": "02:00", "events": 320}
    ],
    "geo_blocks": {
      "total": 4500,
      "top_countries": [
        {"country": "CN", "count": 2000},
        {"country": "RU", "count": 1200},
        {"country": "NG", "count": 800}
      ]
    },
    "bot_detection": {
      "total": 5000,
      "by_indicator": {
        "bad_user_agent": 2000,
        "headless_browser": 1500,
        "datacenter_ip": 1000,
        "missing_headers": 500
      }
    },
    "link_security": {
      "invalid_signatures": 1500,
      "expired_links": 500,
      "replay_attempts": 800,
      "legacy_links_used": 200
    }
  }
}
```

---

## Get Recent Fraud Events

```
GET /api/admin/logs/fraud
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page (max: 100) |
| `category` | string | Filter by category |
| `ip` | string | Filter by IP address |
| `offer_id` | string | Filter by offer |
| `severity` | string | `low`, `medium`, `high`, `critical` |

### Response

```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": "fraud_abc123",
        "timestamp": "2024-01-15T10:30:45Z",
        "category": "bot_detection",
        "event": "bot_blocked",
        "severity": "high",
        "ip_address": "xxx.xxx.xxx.xxx",
        "user_agent": "python-requests/2.28.0",
        "country": "CN",
        "offer_id": "off_xyz789",
        "tracking_code": "tc_abc123",
        "indicators": ["bad_user_agent", "datacenter_ip"],
        "risk_score": 92,
        "blocked": true,
        "metadata": {
          "asn": "AS12345",
          "org": "Cloud Provider Inc"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 15000
    }
  }
}
```

---

## Block IP Address

```
POST /api/admin/fraud/block-ip
```

### Request Body

```json
{
  "ip_address": "xxx.xxx.xxx.xxx",
  "reason": "Automated bot traffic",
  "duration": "permanent",
  "scope": "global"
}
```

### Duration Options

| Value | Description |
|-------|-------------|
| `1h` | 1 hour |
| `24h` | 24 hours |
| `7d` | 7 days |
| `30d` | 30 days |
| `permanent` | Permanent block |

### Response

```json
{
  "success": true,
  "message": "IP blocked successfully",
  "data": {
    "ip_address": "xxx.xxx.xxx.xxx",
    "blocked_at": "2024-01-15T10:30:00Z",
    "expires_at": null,
    "reason": "Automated bot traffic"
  }
}
```

---

## Unblock IP Address

```
POST /api/admin/fraud/unblock-ip
```

### Request Body

```json
{
  "ip_address": "xxx.xxx.xxx.xxx",
  "reason": "False positive"
}
```

### Response

```json
{
  "success": true,
  "message": "IP unblocked successfully"
}
```

---

## Get Blocked IPs

```
GET /api/admin/fraud/blocked-ips
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `scope` | string | `global`, `tenant`, `offer` |

### Response

```json
{
  "success": true,
  "data": {
    "blocked_ips": [
      {
        "ip_address": "xxx.xxx.xxx.xxx",
        "reason": "Automated bot traffic",
        "blocked_at": "2024-01-15T10:30:00Z",
        "expires_at": null,
        "blocked_by": "admin@example.com",
        "scope": "global",
        "events_before_block": 2500,
        "country": "CN"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 150
    }
  }
}
```

---

## Get Fraud Event Details

```
GET /api/admin/fraud/events/:id
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "fraud_abc123",
    "timestamp": "2024-01-15T10:30:45Z",
    "category": "bot_detection",
    "event": "bot_blocked",
    "severity": "high",
    "ip_address": "xxx.xxx.xxx.xxx",
    "user_agent": "python-requests/2.28.0",
    "country": "CN",
    "city": "Shanghai",
    "offer_id": "off_xyz789",
    "offer_title": "Premium App Install",
    "tracking_code": "tc_abc123",
    "request_url": "/api/c/tc_abc123",
    "request_method": "GET",
    "request_headers": {
      "accept": "*/*",
      "accept-encoding": "gzip, deflate"
    },
    "indicators": [
      {
        "type": "bad_user_agent",
        "description": "Known bot user agent pattern",
        "score": 40
      },
      {
        "type": "datacenter_ip",
        "description": "IP belongs to cloud provider",
        "score": 35
      },
      {
        "type": "missing_headers",
        "description": "Missing standard browser headers",
        "score": 17
      }
    ],
    "risk_score": 92,
    "risk_breakdown": {
      "base_score": 0,
      "user_agent_score": 40,
      "ip_score": 35,
      "header_score": 17
    },
    "blocked": true,
    "action_taken": "click_blocked",
    "related_events": [
      {"id": "fraud_xyz789", "timestamp": "2024-01-15T10:30:40Z"}
    ],
    "metadata": {
      "asn": "AS12345",
      "org": "Cloud Provider Inc",
      "isp": "Cloud Provider",
      "connection_type": "datacenter"
    }
  }
}
```

---

## Get Fraud Statistics

```
GET /api/admin/fraud/stats
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | `today`, `week`, `month`, `year` |
| `group_by` | string | `category`, `country`, `offer`, `hour` |

### Response

```json
{
  "success": true,
  "data": {
    "period": "month",
    "total_events": 15000,
    "total_blocked": 14500,
    "false_positive_rate": 3.2,
    "trends": {
      "vs_last_period": "+12%",
      "direction": "up"
    },
    "by_day": [
      {"date": "2024-01-15", "events": 520, "blocked": 505},
      {"date": "2024-01-14", "events": 480, "blocked": 465}
    ],
    "effectiveness": {
      "clicks_protected": 14500,
      "estimated_savings": 7250.00,
      "conversion_quality_improvement": 15.5
    }
  }
}
```

---

## Configure Fraud Rules

```
PUT /api/admin/fraud/config
```

### Request Body

```json
{
  "bot_detection": {
    "enabled": true,
    "block_headless_browsers": true,
    "block_datacenter_ips": true,
    "block_known_bots": true,
    "min_risk_score_to_block": 70
  },
  "rate_limiting": {
    "enabled": true,
    "max_clicks_per_ip_per_minute": 10,
    "max_clicks_per_ip_per_hour": 100,
    "max_api_requests_per_key_per_minute": 60
  },
  "geo_blocking": {
    "enabled": true,
    "enforce_on_postback": false
  },
  "link_security": {
    "enabled": true,
    "link_ttl_minutes": 60,
    "replay_protection": true,
    "allow_legacy_links": false
  },
  "alerting": {
    "enabled": true,
    "alert_threshold": 1000,
    "alert_channels": ["slack", "email"]
  }
}
```

### Response

```json
{
  "success": true,
  "message": "Fraud configuration updated"
}
```

---

## Get Fraud Configuration

```
GET /api/admin/fraud/config
```

### Response

```json
{
  "success": true,
  "data": {
    "bot_detection": {
      "enabled": true,
      "block_headless_browsers": true,
      "block_datacenter_ips": true,
      "block_known_bots": true,
      "min_risk_score_to_block": 70
    },
    "rate_limiting": {
      "enabled": true,
      "max_clicks_per_ip_per_minute": 10,
      "max_clicks_per_ip_per_hour": 100
    },
    "geo_blocking": {
      "enabled": true,
      "enforce_on_postback": false
    },
    "link_security": {
      "enabled": true,
      "link_ttl_minutes": 60,
      "replay_protection": true
    }
  }
}
```

---

## Fraud Indicators

| Indicator | Description | Risk Score |
|-----------|-------------|------------|
| `bad_user_agent` | Known bot user agent | 30-50 |
| `headless_browser` | Headless browser detected | 40-60 |
| `datacenter_ip` | IP from cloud/datacenter | 30-40 |
| `tor_exit_node` | Tor network detected | 50-70 |
| `missing_headers` | Missing standard headers | 15-25 |
| `impossible_browser` | Invalid browser/OS combo | 40-50 |
| `suspicious_referer` | Suspicious or missing referer | 10-20 |
| `rapid_clicks` | Too many clicks too fast | 30-50 |
| `duplicate_fingerprint` | Same device fingerprint | 20-40 |
| `geo_mismatch` | IP country doesn't match | 20-30 |
| `invalid_signature` | Tampered tracking link | 80-100 |
| `replay_attempt` | Reused tracking link | 70-90 |
| `expired_link` | Expired tracking link | 50-60 |

---

## Error Responses

**IP Already Blocked (409):**

```json
{
  "success": false,
  "error": "IP is already blocked",
  "code": "IP_ALREADY_BLOCKED"
}
```

**Invalid IP Format (400):**

```json
{
  "success": false,
  "error": "Invalid IP address format",
  "code": "INVALID_IP"
}
```

**Event Not Found (404):**

```json
{
  "success": false,
  "error": "Fraud event not found",
  "code": "EVENT_NOT_FOUND"
}
```

---

## Webhooks for Fraud Events

Configure webhooks to receive real-time fraud alerts:

```json
{
  "trigger_type": "fraud_event",
  "url": "https://your-server.com/webhooks/fraud",
  "events": ["bot_blocked", "geo_blocked", "rate_limited"],
  "min_severity": "high"
}
```

Example webhook payload:

```json
{
  "event": "fraud.detected",
  "timestamp": "2024-01-15T10:30:45Z",
  "data": {
    "event_id": "fraud_abc123",
    "category": "bot_detection",
    "severity": "high",
    "ip_address": "xxx.xxx.xxx.xxx",
    "risk_score": 92,
    "blocked": true
  }
}
```

---

## Next Steps

- [Geo Rules API](geo-rules.md) - Configure geographic blocking
- [API Keys API](api-keys.md) - Manage API key security
- [Logs API](logs.md) - View detailed logs

