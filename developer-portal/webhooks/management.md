# Webhook Management

Create, configure, and manage webhooks through the AffTok API.

---

## Create Webhook Pipeline

```
POST /api/admin/webhooks/pipelines
```

### Request Body

```json
{
  "name": "Conversion Notifications",
  "description": "Notify external system of conversions",
  "trigger_type": "conversion",
  "status": "active",
  "steps": [
    {
      "name": "Send to CRM",
      "order": 1,
      "url": "https://api.yourcrm.com/webhooks/afftok",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json",
        "X-API-Key": "{{env.CRM_API_KEY}}"
      },
      "body_template": {
        "event": "{{trigger_type}}",
        "conversion_id": "{{conversion.id}}",
        "amount": "{{conversion.amount}}",
        "user_id": "{{conversion.user_id}}"
      },
      "timeout_seconds": 30,
      "retry_policy": {
        "max_attempts": 5,
        "backoff_type": "exponential",
        "initial_delay_seconds": 5
      },
      "signature_mode": "hmac",
      "signing_key": "{{env.WEBHOOK_SECRET}}"
    }
  ],
  "failover_url": "https://backup.yourcrm.com/webhooks/afftok"
}
```

### Trigger Types

| Type | Description |
|------|-------------|
| `click` | Click tracked |
| `conversion` | Conversion created |
| `conversion_approved` | Conversion approved |
| `conversion_rejected` | Conversion rejected |
| `postback` | Postback received |
| `fraud` | Fraud detected |
| `offer_joined` | User joined offer |
| `api_key_event` | API key event |

### Response

```json
{
  "success": true,
  "data": {
    "id": "pipe_abc123",
    "name": "Conversion Notifications",
    "trigger_type": "conversion",
    "status": "active",
    "steps": [...],
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## List Webhooks

```
GET /api/admin/webhooks/pipelines
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `limit` | integer | Items per page |
| `status` | string | Filter: `active`, `paused`, `disabled` |
| `trigger_type` | string | Filter by trigger type |

### Response

```json
{
  "success": true,
  "data": {
    "pipelines": [
      {
        "id": "pipe_abc123",
        "name": "Conversion Notifications",
        "trigger_type": "conversion",
        "status": "active",
        "steps_count": 1,
        "last_triggered": "2024-01-15T10:30:00Z",
        "stats": {
          "total_executions": 1500,
          "successful": 1480,
          "failed": 20,
          "success_rate": 98.67
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 5
    }
  }
}
```

---

## Get Webhook Details

```
GET /api/admin/webhooks/pipelines/:id
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "pipe_abc123",
    "name": "Conversion Notifications",
    "description": "Notify external system of conversions",
    "trigger_type": "conversion",
    "status": "active",
    "steps": [
      {
        "id": "step_xyz789",
        "name": "Send to CRM",
        "order": 1,
        "url": "https://api.yourcrm.com/webhooks/afftok",
        "method": "POST",
        "headers": {
          "Content-Type": "application/json"
        },
        "body_template": {...},
        "timeout_seconds": 30,
        "retry_policy": {...},
        "signature_mode": "hmac",
        "stats": {
          "executions": 1500,
          "successful": 1480,
          "avg_latency_ms": 250
        }
      }
    ],
    "failover_url": "https://backup.yourcrm.com/webhooks/afftok",
    "stats": {
      "total_executions": 1500,
      "successful": 1480,
      "failed": 20,
      "in_retry": 5,
      "in_dlq": 2
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## Update Webhook

```
PUT /api/admin/webhooks/pipelines/:id
```

### Request Body

```json
{
  "name": "Updated Conversion Notifications",
  "status": "paused",
  "steps": [...]
}
```

---

## Delete Webhook

```
DELETE /api/admin/webhooks/pipelines/:id
```

---

## Webhook Steps Configuration

### Step Structure

```json
{
  "name": "Step Name",
  "order": 1,
  "url": "https://example.com/webhook",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer {{env.API_TOKEN}}"
  },
  "body_template": {
    "event": "{{trigger_type}}",
    "data": "{{data}}"
  },
  "timeout_seconds": 30,
  "retry_policy": {
    "max_attempts": 5,
    "backoff_type": "exponential",
    "initial_delay_seconds": 5,
    "max_delay_seconds": 300
  },
  "stop_on_failure": false,
  "signature_mode": "hmac",
  "signing_key": "{{env.SIGNING_KEY}}"
}
```

### Template Variables

| Variable | Description |
|----------|-------------|
| `{{trigger_type}}` | Event type |
| `{{timestamp}}` | Event timestamp |
| `{{click.id}}` | Click ID |
| `{{click.ip}}` | Click IP address |
| `{{click.country}}` | Click country |
| `{{conversion.id}}` | Conversion ID |
| `{{conversion.amount}}` | Conversion amount |
| `{{conversion.user_id}}` | User ID |
| `{{offer.id}}` | Offer ID |
| `{{offer.title}}` | Offer title |
| `{{env.VAR_NAME}}` | Environment variable |

### Signature Modes

| Mode | Description |
|------|-------------|
| `none` | No signature |
| `hmac` | HMAC-SHA256 in `X-Afftok-Signature` header |
| `jwt` | JWT in `Authorization: Bearer` header |

---

## Retry Policy

### Backoff Types

| Type | Description |
|------|-------------|
| `fixed` | Fixed delay between retries |
| `exponential` | Exponential backoff (5s, 10s, 20s, ...) |
| `exponential_jitter` | Exponential with random jitter |

### Example

```json
{
  "retry_policy": {
    "max_attempts": 8,
    "backoff_type": "exponential_jitter",
    "initial_delay_seconds": 5,
    "max_delay_seconds": 3600,
    "jitter_factor": 0.2
  }
}
```

---

## Execution Logs

### Get Recent Executions

```
GET /api/admin/webhooks/logs/recent
```

### Response

```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "exec_abc123",
        "pipeline_id": "pipe_xyz789",
        "pipeline_name": "Conversion Notifications",
        "trigger_type": "conversion",
        "trigger_id": "conv_def456",
        "status": "completed",
        "started_at": "2024-01-15T10:30:00Z",
        "completed_at": "2024-01-15T10:30:01Z",
        "duration_ms": 850,
        "steps": [
          {
            "step_id": "step_abc123",
            "step_name": "Send to CRM",
            "status": "success",
            "http_status": 200,
            "latency_ms": 250,
            "response_body": "{\"received\": true}"
          }
        ]
      }
    ]
  }
}
```

### Get Failed Executions

```
GET /api/admin/webhooks/logs/failures
```

### Get DLQ Items

```
GET /api/admin/webhooks/logs/dlq
```

### Response

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "dlq_abc123",
        "pipeline_id": "pipe_xyz789",
        "trigger_type": "conversion",
        "trigger_id": "conv_def456",
        "failed_at": "2024-01-15T10:30:00Z",
        "attempts": 8,
        "last_error": "Connection timeout",
        "payload": {...}
      }
    ]
  }
}
```

### Retry DLQ Item

```
POST /api/admin/webhooks/logs/dlq/:id/retry
```

### Delete DLQ Item

```
DELETE /api/admin/webhooks/logs/dlq/:id
```

---

## Testing Webhooks

### Test Pipeline

```
POST /api/admin/webhooks/test/pipeline/:id
```

### Request Body

```json
{
  "test_data": {
    "conversion": {
      "id": "test_conv_123",
      "amount": 49.99,
      "user_id": "test_user_456"
    }
  }
}
```

### Response

```json
{
  "success": true,
  "data": {
    "execution_id": "exec_test_abc123",
    "status": "completed",
    "steps": [
      {
        "step_name": "Send to CRM",
        "status": "success",
        "http_status": 200,
        "latency_ms": 250,
        "request": {
          "url": "https://api.yourcrm.com/webhooks/afftok",
          "method": "POST",
          "headers": {...},
          "body": {...}
        },
        "response": {
          "status": 200,
          "body": "{\"received\": true}"
        }
      }
    ]
  }
}
```

### Test Single Step

```
POST /api/admin/webhooks/test/step
```

### Request Body

```json
{
  "url": "https://api.example.com/webhook",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "test": true,
    "message": "Hello from AffTok"
  },
  "timeout_seconds": 10
}
```

---

## Webhook Statistics

```
GET /api/admin/webhooks/stats
```

### Response

```json
{
  "success": true,
  "data": {
    "summary": {
      "total_pipelines": 5,
      "active_pipelines": 4,
      "total_executions_today": 1500,
      "success_rate": 98.5,
      "avg_latency_ms": 250
    },
    "by_trigger_type": {
      "conversion": {
        "executions": 1000,
        "success_rate": 99.0
      },
      "click": {
        "executions": 500,
        "success_rate": 97.5
      }
    },
    "queues": {
      "primary": 25,
      "failover": 5,
      "dlq": 2
    },
    "top_errors": [
      {
        "error": "Connection timeout",
        "count": 15
      },
      {
        "error": "HTTP 500",
        "count": 5
      }
    ]
  }
}
```

---

## Webhook Events Reference

### Conversion Created

```json
{
  "type": "conversion.created",
  "data": {
    "id": "conv_abc123",
    "click_id": "clk_xyz789",
    "offer_id": "off_def456",
    "user_id": "usr_ghi789",
    "transaction_id": "txn_123456",
    "amount": 49.99,
    "currency": "USD",
    "payout": 5.00,
    "status": "pending"
  }
}
```

### Click Tracked

```json
{
  "type": "click.tracked",
  "data": {
    "id": "clk_abc123",
    "tracking_code": "tc_xyz789",
    "offer_id": "off_def456",
    "user_id": "usr_ghi789",
    "ip_address": "xxx.xxx.xxx.xxx",
    "country": "US",
    "device": "mobile"
  }
}
```

### Fraud Detected

```json
{
  "type": "fraud.detected",
  "data": {
    "event_type": "click",
    "reason": "bot_detection",
    "risk_score": 92,
    "ip_address": "xxx.xxx.xxx.xxx",
    "blocked": true
  }
}
```

---

## Next Steps

- [Webhook Security](security.md) - Secure your webhooks
- [Webhook Testing](testing.md) - Test your endpoint
- [API Reference](../api-reference/admin.md) - Full API documentation

