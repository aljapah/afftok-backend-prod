# Webhooks API

Manage webhook pipelines for real-time event notifications.

## Overview

AffTok's webhook system supports:

- **Multi-step pipelines** - Chain multiple webhook calls
- **Retry with exponential backoff** - Automatic retry on failures
- **Payload signing** - HMAC-SHA256 or JWT signatures
- **Template engine** - Dynamic payload generation
- **Dead Letter Queue** - Handle persistent failures

---

## GET /api/admin/webhooks/pipelines

List all webhook pipelines.

### Request

```http
GET /api/admin/webhooks/pipelines
Authorization: Bearer <admin_token>
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `trigger_type` | string | Filter by trigger (conversion, click, postback) |
| `status` | string | Filter by status (active, inactive, paused) |
| `page` | integer | Page number |
| `limit` | integer | Items per page |

### Response

```json
{
  "success": true,
  "data": {
    "pipelines": [
      {
        "id": "wh_pipe_123",
        "name": "Conversion Notification",
        "description": "Notify partner system on conversions",
        "trigger_type": "conversion",
        "status": "active",
        "steps_count": 2,
        "last_triggered": "2024-12-01T12:00:00Z",
        "success_rate": 98.5,
        "created_at": "2024-01-15T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 15
    }
  }
}
```

---

## GET /api/admin/webhooks/pipelines/:id

Get a specific pipeline with all steps.

### Request

```http
GET /api/admin/webhooks/pipelines/wh_pipe_123
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "id": "wh_pipe_123",
    "name": "Conversion Notification",
    "description": "Notify partner system on conversions",
    "trigger_type": "conversion",
    "status": "active",
    "failover_url": "https://backup.example.com/webhook",
    "steps": [
      {
        "id": "wh_step_1",
        "order": 1,
        "name": "Primary Notification",
        "url": "https://partner.example.com/api/conversion",
        "method": "POST",
        "headers": {
          "Content-Type": "application/json",
          "X-Partner-ID": "partner_123"
        },
        "body_template": "{\"conversion_id\": \"{{conversion.id}}\", \"amount\": {{conversion.amount}}, \"user_id\": \"{{conversion.user_id}}\"}",
        "timeout_seconds": 30,
        "retry_policy": {
          "max_attempts": 5,
          "backoff_multiplier": 2,
          "initial_delay_seconds": 5
        },
        "signature_mode": "hmac",
        "stop_on_failure": false
      },
      {
        "id": "wh_step_2",
        "order": 2,
        "name": "Analytics Update",
        "url": "https://analytics.example.com/event",
        "method": "POST",
        "signature_mode": "jwt",
        "stop_on_failure": true
      }
    ],
    "stats": {
      "total_executions": 5234,
      "successful": 5156,
      "failed": 78,
      "in_dlq": 12,
      "avg_latency_ms": 245
    },
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-12-01T12:00:00Z"
  }
}
```

---

## POST /api/admin/webhooks/pipelines

Create a new webhook pipeline.

### Request

```http
POST /api/admin/webhooks/pipelines
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "name": "Partner Conversion Webhook",
  "description": "Send conversion data to partner system",
  "trigger_type": "conversion",
  "status": "active",
  "failover_url": "https://backup.partner.com/webhook",
  "steps": [
    {
      "order": 1,
      "name": "Primary Endpoint",
      "url": "https://partner.com/api/conversions",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json",
        "X-API-Version": "2.0"
      },
      "body_template": "{\n  \"event\": \"conversion\",\n  \"data\": {\n    \"id\": \"{{conversion.id}}\",\n    \"offer_id\": \"{{conversion.offer_id}}\",\n    \"amount\": {{conversion.amount}},\n    \"currency\": \"{{conversion.currency}}\",\n    \"timestamp\": \"{{timestamp}}\"\n  }\n}",
      "timeout_seconds": 30,
      "retry_policy": {
        "max_attempts": 5,
        "backoff_multiplier": 2,
        "initial_delay_seconds": 5
      },
      "signature_mode": "hmac",
      "signing_key": "your_hmac_secret_key",
      "stop_on_failure": false
    }
  ]
}
```

### Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Pipeline name |
| `description` | string | No | Pipeline description |
| `trigger_type` | string | Yes | `conversion`, `click`, `postback`, `fraud` |
| `status` | string | No | `active`, `inactive`, `paused` |
| `failover_url` | string | No | Fallback URL if all steps fail |
| `steps` | array | Yes | Array of webhook steps |

### Step Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `order` | integer | Yes | Execution order (1-based) |
| `name` | string | Yes | Step name |
| `url` | string | Yes | Webhook URL (supports templates) |
| `method` | string | No | HTTP method (default: POST) |
| `headers` | object | No | HTTP headers (supports templates) |
| `body_template` | string | No | Request body template |
| `timeout_seconds` | integer | No | Request timeout (default: 30) |
| `retry_policy` | object | No | Retry configuration |
| `signature_mode` | string | No | `none`, `hmac`, `jwt` |
| `signing_key` | string | No | Key for signing (required if signing) |
| `stop_on_failure` | boolean | No | Stop pipeline on step failure |

### Response

```json
{
  "success": true,
  "message": "Webhook pipeline created successfully",
  "data": {
    "id": "wh_pipe_456",
    "name": "Partner Conversion Webhook",
    "trigger_type": "conversion",
    "steps_count": 1,
    "created_at": "2024-12-01T12:00:00Z"
  }
}
```

---

## PUT /api/admin/webhooks/pipelines/:id

Update an existing pipeline.

### Request

```http
PUT /api/admin/webhooks/pipelines/wh_pipe_123
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "name": "Updated Pipeline Name",
  "status": "paused"
}
```

### Response

```json
{
  "success": true,
  "message": "Webhook pipeline updated successfully",
  "data": {
    "id": "wh_pipe_123",
    "name": "Updated Pipeline Name",
    "status": "paused",
    "updated_at": "2024-12-01T12:30:00Z"
  }
}
```

---

## DELETE /api/admin/webhooks/pipelines/:id

Delete a webhook pipeline.

### Request

```http
DELETE /api/admin/webhooks/pipelines/wh_pipe_123
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "message": "Webhook pipeline deleted successfully"
}
```

---

## POST /api/admin/webhooks/test/pipeline

Test a webhook pipeline without saving.

### Request

```http
POST /api/admin/webhooks/test/pipeline
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "pipeline_id": "wh_pipe_123",
  "test_data": {
    "conversion": {
      "id": "test_conv_123",
      "offer_id": "off_456",
      "amount": 25.00,
      "currency": "USD"
    }
  }
}
```

### Response

```json
{
  "success": true,
  "data": {
    "pipeline_id": "wh_pipe_123",
    "test_mode": true,
    "results": [
      {
        "step_order": 1,
        "step_name": "Primary Endpoint",
        "status": "success",
        "status_code": 200,
        "latency_ms": 156,
        "request": {
          "url": "https://partner.com/api/conversions",
          "method": "POST",
          "headers": { "Content-Type": "application/json" },
          "body": "{\"event\": \"conversion\", ...}"
        },
        "response": {
          "status": 200,
          "body": "{\"received\": true}"
        }
      }
    ],
    "total_latency_ms": 156,
    "all_steps_passed": true
  }
}
```

---

## GET /api/admin/webhooks/logs/recent

Get recent webhook execution logs.

### Request

```http
GET /api/admin/webhooks/logs/recent?limit=50
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "wh_exec_123",
        "pipeline_id": "wh_pipe_123",
        "pipeline_name": "Conversion Notification",
        "trigger_type": "conversion",
        "trigger_id": "conv_456",
        "status": "success",
        "steps_executed": 2,
        "steps_succeeded": 2,
        "total_latency_ms": 312,
        "created_at": "2024-12-01T12:00:00Z"
      }
    ]
  }
}
```

---

## GET /api/admin/webhooks/logs/failures

Get failed webhook executions.

### Request

```http
GET /api/admin/webhooks/logs/failures?limit=50
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "failures": [
      {
        "id": "wh_exec_789",
        "pipeline_id": "wh_pipe_123",
        "pipeline_name": "Conversion Notification",
        "trigger_id": "conv_999",
        "status": "failed",
        "failed_step": 1,
        "error": "Connection timeout after 30s",
        "retry_count": 5,
        "in_dlq": true,
        "created_at": "2024-12-01T11:00:00Z",
        "last_retry_at": "2024-12-01T11:25:00Z"
      }
    ]
  }
}
```

---

## GET /api/admin/webhooks/dlq

Get Dead Letter Queue items.

### Request

```http
GET /api/admin/webhooks/dlq
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "dlq_123",
        "execution_id": "wh_exec_789",
        "pipeline_id": "wh_pipe_123",
        "pipeline_name": "Conversion Notification",
        "trigger_type": "conversion",
        "trigger_id": "conv_999",
        "original_payload": "{...}",
        "error": "Max retries exceeded",
        "retry_count": 5,
        "created_at": "2024-12-01T11:25:00Z"
      }
    ],
    "total": 12
  }
}
```

---

## POST /api/admin/webhooks/dlq/:id/retry

Retry a DLQ item.

### Request

```http
POST /api/admin/webhooks/dlq/dlq_123/retry
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "message": "DLQ item queued for retry",
  "data": {
    "id": "dlq_123",
    "new_execution_id": "wh_exec_1001",
    "status": "queued"
  }
}
```

---

## DELETE /api/admin/webhooks/dlq/:id

Delete a DLQ item.

### Request

```http
DELETE /api/admin/webhooks/dlq/dlq_123
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "message": "DLQ item deleted"
}
```

---

## GET /api/admin/webhooks/stats

Get webhook system statistics.

### Request

```http
GET /api/admin/webhooks/stats
Authorization: Bearer <admin_token>
```

### Response

```json
{
  "success": true,
  "data": {
    "total_pipelines": 15,
    "active_pipelines": 12,
    "total_executions_today": 5234,
    "success_rate": 98.5,
    "avg_latency_ms": 245,
    "by_trigger_type": {
      "conversion": { "count": 3500, "success_rate": 99.1 },
      "click": { "count": 1200, "success_rate": 97.5 },
      "postback": { "count": 534, "success_rate": 98.0 }
    },
    "queue_status": {
      "primary_queue": 12,
      "failover_queue": 3,
      "dlq": 5
    },
    "retry_stats": {
      "total_retries_today": 156,
      "successful_retries": 134,
      "failed_retries": 22
    }
  }
}
```

---

## Template Placeholders

Use double curly braces for dynamic values:

### Click Event Placeholders

| Placeholder | Description |
|-------------|-------------|
| `{{click.id}}` | Click ID |
| `{{click.ip}}` | IP address |
| `{{click.country}}` | Country code |
| `{{click.device}}` | Device type |
| `{{click.browser}}` | Browser name |
| `{{click.os}}` | Operating system |
| `{{click.user_agent}}` | Full user agent |
| `{{click.timestamp}}` | Click timestamp |

### Conversion Event Placeholders

| Placeholder | Description |
|-------------|-------------|
| `{{conversion.id}}` | Conversion ID |
| `{{conversion.click_id}}` | Associated click ID |
| `{{conversion.offer_id}}` | Offer ID |
| `{{conversion.user_id}}` | User ID |
| `{{conversion.amount}}` | Conversion amount |
| `{{conversion.currency}}` | Currency code |
| `{{conversion.status}}` | Conversion status |
| `{{conversion.timestamp}}` | Conversion timestamp |

### System Placeholders

| Placeholder | Description |
|-------------|-------------|
| `{{timestamp}}` | Current ISO timestamp |
| `{{timestamp_unix}}` | Current Unix timestamp |
| `{{uuid}}` | New random UUID |
| `{{nonce}}` | Random nonce for signing |

### Example Template

```json
{
  "event_type": "conversion",
  "event_id": "{{conversion.id}}",
  "data": {
    "offer": "{{conversion.offer_id}}",
    "amount": {{conversion.amount}},
    "currency": "{{conversion.currency}}",
    "source_click": {
      "id": "{{click.id}}",
      "country": "{{click.country}}",
      "device": "{{click.device}}"
    }
  },
  "meta": {
    "timestamp": "{{timestamp}}",
    "nonce": "{{nonce}}"
  }
}
```

---

## Signature Modes

### HMAC-SHA256

Adds `X-Afftok-Signature` header:

```
X-Afftok-Signature: sha256=abc123...
```

Signature is computed over the request body:

```
HMAC-SHA256(body, signing_key)
```

### Verification Example (Node.js)

```javascript
const crypto = require('crypto');

function verifySignature(body, signature, secret) {
  const expected = 'sha256=' + crypto
    .createHmac('sha256', secret)
    .update(body)
    .digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expected)
  );
}
```

### JWT

Adds `Authorization: Bearer <jwt>` header.

JWT payload includes:

```json
{
  "task_id": "wh_exec_123",
  "pipeline_id": "wh_pipe_456",
  "trigger_type": "conversion",
  "trigger_id": "conv_789",
  "iat": 1701432000,
  "exp": 1701432300
}
```

---

## Retry Policy

Default retry policy:

| Attempt | Delay |
|---------|-------|
| 1 | 5 seconds |
| 2 | 10 seconds |
| 3 | 30 seconds |
| 4 | 1 minute |
| 5 | 5 minutes |

After max attempts, item moves to failover queue, then DLQ.

### Custom Retry Policy

```json
{
  "retry_policy": {
    "max_attempts": 3,
    "initial_delay_seconds": 10,
    "backoff_multiplier": 3,
    "max_delay_seconds": 300
  }
}
```

---

## Trigger Types

| Type | Triggered When |
|------|---------------|
| `conversion` | New conversion recorded |
| `click` | Click tracked (use carefully - high volume) |
| `postback` | Postback received from network |
| `fraud` | Fraud event detected |

---

Next: [Tenants API](./tenants.md)

