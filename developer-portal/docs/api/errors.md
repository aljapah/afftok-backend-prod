# Error Codes Reference

Complete list of AffTok API error codes and their meanings.

## Error Response Format

All errors follow this structure:

```json
{
  "success": false,
  "error": {
    "code": "INVALID_API_KEY",
    "message": "The provided API key is invalid or expired",
    "details": {
      "key_prefix": "afftok_live_sk_abc..."
    },
    "correlation_id": "corr_abc123def456"
  }
}
```

| Field | Description |
|-------|-------------|
| `code` | Machine-readable error code |
| `message` | Human-readable description |
| `details` | Additional context (optional) |
| `correlation_id` | Unique ID for support requests |

---

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Resource already exists |
| 422 | Unprocessable Entity - Validation failed |
| 429 | Too Many Requests - Rate limited |
| 500 | Internal Server Error |
| 502 | Bad Gateway - Upstream error |
| 503 | Service Unavailable - Maintenance |

---

## Authentication Errors (1xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `AUTH_REQUIRED` | 401 | Authentication required |
| `INVALID_TOKEN` | 401 | JWT token is invalid |
| `TOKEN_EXPIRED` | 401 | JWT token has expired |
| `TOKEN_MALFORMED` | 401 | JWT token format is invalid |
| `INVALID_SIGNATURE` | 401 | Token signature verification failed |
| `REFRESH_TOKEN_INVALID` | 401 | Refresh token is invalid |
| `REFRESH_TOKEN_EXPIRED` | 401 | Refresh token has expired |
| `USER_NOT_FOUND` | 401 | User account not found |
| `USER_SUSPENDED` | 403 | User account is suspended |
| `USER_INACTIVE` | 403 | User account is inactive |
| `INVALID_CREDENTIALS` | 401 | Email or password is incorrect |
| `EMAIL_NOT_VERIFIED` | 403 | Email verification required |

### Example

```json
{
  "success": false,
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "Your session has expired. Please log in again.",
    "details": {
      "expired_at": "2024-12-01T12:00:00Z"
    }
  }
}
```

---

## API Key Errors (2xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `INVALID_API_KEY` | 401 | API key is invalid |
| `API_KEY_EXPIRED` | 401 | API key has expired |
| `API_KEY_REVOKED` | 401 | API key has been revoked |
| `API_KEY_SUSPENDED` | 403 | API key is suspended |
| `API_KEY_IP_BLOCKED` | 403 | IP not in allowed list |
| `API_KEY_RATE_LIMITED` | 429 | API key rate limit exceeded |
| `API_KEY_PERMISSION_DENIED` | 403 | Key lacks required permission |
| `API_KEY_NOT_FOUND` | 404 | API key doesn't exist |

### Example

```json
{
  "success": false,
  "error": {
    "code": "API_KEY_RATE_LIMITED",
    "message": "Rate limit exceeded for this API key",
    "details": {
      "limit": 60,
      "window": "1 minute",
      "retry_after": 45
    }
  }
}
```

---

## Validation Errors (3xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `VALIDATION_FAILED` | 422 | Request validation failed |
| `MISSING_FIELD` | 400 | Required field is missing |
| `INVALID_FIELD` | 400 | Field value is invalid |
| `INVALID_FORMAT` | 400 | Field format is incorrect |
| `FIELD_TOO_LONG` | 400 | Field exceeds max length |
| `FIELD_TOO_SHORT` | 400 | Field below min length |
| `INVALID_EMAIL` | 400 | Email format is invalid |
| `INVALID_URL` | 400 | URL format is invalid |
| `INVALID_UUID` | 400 | UUID format is invalid |
| `INVALID_COUNTRY_CODE` | 400 | Country code not recognized |
| `INVALID_CURRENCY` | 400 | Currency code not recognized |
| `INVALID_DATE` | 400 | Date format is invalid |
| `INVALID_JSON` | 400 | JSON parsing failed |

### Example

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Request validation failed",
    "details": {
      "fields": [
        {
          "field": "email",
          "code": "INVALID_EMAIL",
          "message": "Invalid email format"
        },
        {
          "field": "amount",
          "code": "INVALID_FIELD",
          "message": "Amount must be positive"
        }
      ]
    }
  }
}
```

---

## Resource Errors (4xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `RESOURCE_NOT_FOUND` | 404 | Resource doesn't exist |
| `OFFER_NOT_FOUND` | 404 | Offer doesn't exist |
| `USER_OFFER_NOT_FOUND` | 404 | User-offer relationship not found |
| `CLICK_NOT_FOUND` | 404 | Click record not found |
| `CONVERSION_NOT_FOUND` | 404 | Conversion not found |
| `TENANT_NOT_FOUND` | 404 | Tenant doesn't exist |
| `WEBHOOK_NOT_FOUND` | 404 | Webhook pipeline not found |
| `GEO_RULE_NOT_FOUND` | 404 | Geo rule not found |
| `ALREADY_EXISTS` | 409 | Resource already exists |
| `OFFER_ALREADY_JOINED` | 409 | User already joined offer |
| `DUPLICATE_CONVERSION` | 409 | Conversion already recorded |
| `SLUG_TAKEN` | 409 | Slug is already in use |

### Example

```json
{
  "success": false,
  "error": {
    "code": "OFFER_NOT_FOUND",
    "message": "The requested offer does not exist",
    "details": {
      "offer_id": "off_123456"
    }
  }
}
```

---

## Tracking Errors (5xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `INVALID_TRACKING_CODE` | 400 | Tracking code is invalid |
| `TRACKING_CODE_EXPIRED` | 400 | Tracking link has expired |
| `INVALID_LINK_SIGNATURE` | 400 | Link signature verification failed |
| `LINK_REPLAY_DETECTED` | 400 | Replay attack detected |
| `MALFORMED_LINK` | 400 | Link format is incorrect |
| `OFFER_INACTIVE` | 400 | Offer is not active |
| `OFFER_EXPIRED` | 400 | Offer has expired |
| `OFFER_CAP_REACHED` | 400 | Offer daily cap reached |
| `USER_NOT_IN_OFFER` | 400 | User hasn't joined this offer |

### Example

```json
{
  "success": false,
  "error": {
    "code": "TRACKING_CODE_EXPIRED",
    "message": "This tracking link has expired",
    "details": {
      "expired_at": "2024-12-01T12:00:00Z",
      "ttl_seconds": 86400
    }
  }
}
```

---

## Geo & Fraud Errors (6xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `GEO_BLOCKED` | 403 | Country is blocked |
| `BOT_DETECTED` | 403 | Bot traffic detected |
| `FRAUD_DETECTED` | 403 | Fraudulent activity detected |
| `IP_BLOCKED` | 403 | IP address is blocked |
| `RATE_LIMITED` | 429 | Request rate limit exceeded |
| `SUSPICIOUS_ACTIVITY` | 403 | Suspicious pattern detected |
| `VPN_DETECTED` | 403 | VPN/Proxy detected |
| `DATACENTER_IP` | 403 | Datacenter IP blocked |

### Example

```json
{
  "success": false,
  "error": {
    "code": "GEO_BLOCKED",
    "message": "Traffic from your country is not allowed for this offer",
    "details": {
      "country": "XX",
      "rule_id": "geo_rule_123"
    }
  }
}
```

---

## Postback Errors (7xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `POSTBACK_INVALID` | 400 | Postback data is invalid |
| `POSTBACK_DUPLICATE` | 409 | Duplicate postback |
| `POSTBACK_CLICK_NOT_FOUND` | 400 | Referenced click not found |
| `POSTBACK_SIGNATURE_INVALID` | 401 | Postback signature invalid |
| `POSTBACK_EXPIRED` | 400 | Postback timestamp too old |
| `ADVERTISER_MISMATCH` | 400 | Advertiser doesn't match |

### Example

```json
{
  "success": false,
  "error": {
    "code": "POSTBACK_DUPLICATE",
    "message": "This conversion has already been recorded",
    "details": {
      "external_id": "ext_conv_123",
      "original_conversion_id": "conv_456"
    }
  }
}
```

---

## Tenant Errors (8xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `TENANT_NOT_FOUND` | 404 | Tenant doesn't exist |
| `TENANT_SUSPENDED` | 403 | Tenant is suspended |
| `TENANT_LIMIT_REACHED` | 403 | Tenant limit exceeded |
| `TENANT_FEATURE_DISABLED` | 403 | Feature not enabled for tenant |
| `INVALID_TENANT_DOMAIN` | 400 | Domain not verified |
| `TENANT_PLAN_REQUIRED` | 403 | Higher plan required |

### Example

```json
{
  "success": false,
  "error": {
    "code": "TENANT_LIMIT_REACHED",
    "message": "You have reached your plan's limit",
    "details": {
      "limit_type": "max_offers",
      "current": 100,
      "limit": 100,
      "plan": "pro",
      "upgrade_to": "enterprise"
    }
  }
}
```

---

## Webhook Errors (9xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `WEBHOOK_DELIVERY_FAILED` | 500 | Webhook delivery failed |
| `WEBHOOK_TIMEOUT` | 504 | Webhook request timed out |
| `WEBHOOK_INVALID_RESPONSE` | 502 | Invalid response from endpoint |
| `WEBHOOK_MAX_RETRIES` | 500 | Max retry attempts reached |
| `WEBHOOK_PIPELINE_INACTIVE` | 400 | Pipeline is not active |

### Example

```json
{
  "success": false,
  "error": {
    "code": "WEBHOOK_DELIVERY_FAILED",
    "message": "Failed to deliver webhook after 5 attempts",
    "details": {
      "pipeline_id": "wh_pipe_123",
      "last_error": "Connection refused",
      "attempts": 5,
      "moved_to_dlq": true
    }
  }
}
```

---

## System Errors (10xxx)

| Code | HTTP | Description |
|------|------|-------------|
| `INTERNAL_ERROR` | 500 | Unexpected server error |
| `DATABASE_ERROR` | 500 | Database operation failed |
| `REDIS_ERROR` | 500 | Redis operation failed |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |
| `MAINTENANCE_MODE` | 503 | System under maintenance |
| `UPSTREAM_ERROR` | 502 | Upstream service error |
| `TIMEOUT` | 504 | Request timed out |

### Example

```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred. Please try again.",
    "correlation_id": "corr_abc123def456"
  }
}
```

---

## Troubleshooting

### Common Issues

#### "INVALID_API_KEY" when key looks correct

1. Check for whitespace in the key
2. Ensure you're using the correct environment (live vs test)
3. Verify the key hasn't been revoked
4. Check if IP restrictions apply

#### "TOKEN_EXPIRED" after recent login

1. Check client clock synchronization
2. Token TTL is 1 hour by default
3. Use refresh token to get new access token

#### "RATE_LIMITED" errors

1. Implement exponential backoff
2. Check `Retry-After` header
3. Consider caching responses
4. Request rate limit increase if needed

#### "GEO_BLOCKED" for valid traffic

1. Verify geo rules configuration
2. Check if VPN detection is enabled
3. Review offer-specific geo settings

### Getting Help

When contacting support, include:

1. **correlation_id** from the error response
2. **Timestamp** of the error
3. **Request details** (endpoint, method, headers)
4. **Response body** (full error)

---

## Error Code Ranges

| Range | Category |
|-------|----------|
| 1xxx | Authentication |
| 2xxx | API Keys |
| 3xxx | Validation |
| 4xxx | Resources |
| 5xxx | Tracking |
| 6xxx | Geo & Fraud |
| 7xxx | Postbacks |
| 8xxx | Tenants |
| 9xxx | Webhooks |
| 10xxx | System |

---

Next: [SDK Documentation](../sdk/README.md)

