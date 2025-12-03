# AffTok Error Codes Reference

Complete reference for all AffTok API error codes.

## Error Response Format

```json
{
  "success": false,
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context"
  }
}
```

---

## Authentication Errors

### INVALID_API_KEY

**HTTP Status:** 401

**Description:** The provided API key is invalid, not found, or has been revoked.

**Causes:**
- API key doesn't exist
- API key has been revoked
- API key is expired

**Solution:**
- Verify your API key in the AffTok dashboard
- Generate a new API key if needed
- Check for typos in the API key

```json
{
  "success": false,
  "error": "Invalid API key",
  "code": "INVALID_API_KEY"
}
```

---

### INVALID_SIGNATURE

**HTTP Status:** 403

**Description:** The HMAC-SHA256 signature verification failed.

**Causes:**
- Incorrect signature generation
- Wrong API key used for signing
- Timestamp/nonce mismatch

**Solution:**
- Verify signature generation logic
- Ensure timestamp is in milliseconds
- Check nonce is 32 characters

```json
{
  "success": false,
  "error": "Invalid signature",
  "code": "INVALID_SIGNATURE"
}
```

---

### EXPIRED_REQUEST

**HTTP Status:** 403

**Description:** The request timestamp is too old (> 5 minutes).

**Causes:**
- Clock skew between client and server
- Request was delayed in transit
- Timestamp is incorrect

**Solution:**
- Synchronize system clock
- Use current timestamp
- Reduce network latency

```json
{
  "success": false,
  "error": "Request expired",
  "code": "EXPIRED_REQUEST",
  "details": {
    "max_age_seconds": 300
  }
}
```

---

### REVOKED_API_KEY

**HTTP Status:** 401

**Description:** The API key has been explicitly revoked.

**Causes:**
- Key was manually revoked
- Security breach detected
- Account suspended

**Solution:**
- Contact support
- Generate a new API key
- Review account status

```json
{
  "success": false,
  "error": "API key has been revoked",
  "code": "REVOKED_API_KEY"
}
```

---

## Validation Errors

### INVALID_PARAMS

**HTTP Status:** 400

**Description:** One or more required parameters are missing or invalid.

**Causes:**
- Missing required fields
- Invalid data types
- Malformed JSON

**Solution:**
- Check all required parameters
- Verify data types
- Validate JSON format

```json
{
  "success": false,
  "error": "Invalid parameters",
  "code": "INVALID_PARAMS",
  "details": {
    "missing": ["offer_id", "timestamp"],
    "invalid": ["amount"]
  }
}
```

---

### INVALID_OFFER

**HTTP Status:** 404

**Description:** The specified offer ID was not found or is inactive.

**Causes:**
- Offer doesn't exist
- Offer is paused or expired
- Wrong offer ID

**Solution:**
- Verify offer ID in dashboard
- Check offer status
- Use correct offer ID

```json
{
  "success": false,
  "error": "Offer not found",
  "code": "INVALID_OFFER",
  "details": {
    "offer_id": "offer_123"
  }
}
```

---

### INVALID_CURRENCY

**HTTP Status:** 400

**Description:** The specified currency code is not supported.

**Causes:**
- Invalid ISO 4217 code
- Currency not enabled
- Typo in currency code

**Solution:**
- Use valid ISO 4217 code (USD, EUR, etc.)
- Check supported currencies
- Default to USD if unsure

```json
{
  "success": false,
  "error": "Invalid currency code",
  "code": "INVALID_CURRENCY",
  "details": {
    "currency": "XYZ",
    "supported": ["USD", "EUR", "GBP", "JPY"]
  }
}
```

---

### INVALID_STATUS

**HTTP Status:** 400

**Description:** The conversion status is not valid.

**Causes:**
- Invalid status value
- Typo in status

**Solution:**
- Use: pending, approved, or rejected

```json
{
  "success": false,
  "error": "Invalid status",
  "code": "INVALID_STATUS",
  "details": {
    "status": "unknown",
    "allowed": ["pending", "approved", "rejected"]
  }
}
```

---

## Duplicate Errors

### DUPLICATE_TRANSACTION

**HTTP Status:** 409

**Description:** A conversion with this transaction ID already exists.

**Causes:**
- Transaction already processed
- Retry of successful request
- Duplicate submission

**Solution:**
- Use unique transaction IDs
- Check if already processed
- Implement idempotency

```json
{
  "success": false,
  "error": "Duplicate transaction",
  "code": "DUPLICATE_TRANSACTION",
  "details": {
    "transaction_id": "txn_abc123",
    "existing_conversion_id": "conv_xyz789"
  }
}
```

---

### DUPLICATE_CLICK

**HTTP Status:** 409

**Description:** A duplicate click was detected (same fingerprint within time window).

**Causes:**
- User clicked multiple times
- Bot activity
- Page refresh

**Solution:**
- Implement client-side debouncing
- Check response before retrying

```json
{
  "success": false,
  "error": "Duplicate click detected",
  "code": "DUPLICATE_CLICK",
  "details": {
    "fingerprint": "abc123...",
    "window_seconds": 60
  }
}
```

---

## Rate Limiting Errors

### RATE_LIMITED

**HTTP Status:** 429

**Description:** Too many requests in the time window.

**Causes:**
- Exceeded request limit
- Burst traffic
- Missing rate limit handling

**Solution:**
- Implement exponential backoff
- Reduce request frequency
- Use batch endpoints

```json
{
  "success": false,
  "error": "Rate limit exceeded",
  "code": "RATE_LIMITED",
  "details": {
    "limit": 60,
    "window": "minute",
    "retry_after": 45
  }
}
```

---

### API_KEY_RATE_LIMITED

**HTTP Status:** 429

**Description:** This specific API key has exceeded its rate limit.

**Causes:**
- Key-specific limit reached
- Shared key usage
- Attack detected

**Solution:**
- Wait for rate limit reset
- Use separate keys for different services
- Contact support for limit increase

```json
{
  "success": false,
  "error": "API key rate limited",
  "code": "API_KEY_RATE_LIMITED",
  "details": {
    "limit": 60,
    "reset_at": 1234567890
  }
}
```

---

## Geo Errors

### GEO_BLOCKED

**HTTP Status:** 403

**Description:** The request originated from a blocked country.

**Causes:**
- Country not in allow list
- Country in block list
- Geo rule violation

**Solution:**
- Check geo rules in dashboard
- Verify user's country
- Update geo rules if needed

```json
{
  "success": false,
  "error": "Country not allowed",
  "code": "GEO_BLOCKED",
  "details": {
    "country": "XX",
    "rule_id": "rule_123"
  }
}
```

---

## Security Errors

### INVALID_LINK

**HTTP Status:** 403

**Description:** The tracking link is invalid or tampered.

**Causes:**
- Link signature invalid
- Link expired
- Link modified

**Solution:**
- Use fresh tracking links
- Don't modify link parameters
- Check link TTL

```json
{
  "success": false,
  "error": "Invalid tracking link",
  "code": "INVALID_LINK"
}
```

---

### EXPIRED_LINK

**HTTP Status:** 403

**Description:** The tracking link has expired.

**Causes:**
- Link TTL exceeded
- Old link used

**Solution:**
- Generate new tracking link
- Use links within TTL

```json
{
  "success": false,
  "error": "Tracking link expired",
  "code": "EXPIRED_LINK",
  "details": {
    "expired_at": 1234567890
  }
}
```

---

### REPLAY_ATTEMPT

**HTTP Status:** 403

**Description:** This request appears to be a replay attack.

**Causes:**
- Nonce already used
- Request replayed
- Duplicate nonce

**Solution:**
- Generate unique nonce per request
- Don't reuse nonces

```json
{
  "success": false,
  "error": "Replay attempt detected",
  "code": "REPLAY_ATTEMPT"
}
```

---

### BOT_DETECTED

**HTTP Status:** 403

**Description:** The request appears to be from a bot.

**Causes:**
- Bot user agent
- Suspicious patterns
- Data center IP

**Solution:**
- Use real browser/device
- Contact support if false positive

```json
{
  "success": false,
  "error": "Bot detected",
  "code": "BOT_DETECTED",
  "details": {
    "reason": "suspicious_user_agent"
  }
}
```

---

## Server Errors

### INTERNAL_ERROR

**HTTP Status:** 500

**Description:** An unexpected server error occurred.

**Causes:**
- Server issue
- Database error
- Temporary outage

**Solution:**
- Retry with exponential backoff
- Check status page
- Contact support if persistent

```json
{
  "success": false,
  "error": "Internal server error",
  "code": "INTERNAL_ERROR",
  "details": {
    "request_id": "req_abc123"
  }
}
```

---

### SERVICE_UNAVAILABLE

**HTTP Status:** 503

**Description:** The service is temporarily unavailable.

**Causes:**
- Maintenance
- High load
- Outage

**Solution:**
- Retry later
- Check status page
- Queue requests offline

```json
{
  "success": false,
  "error": "Service temporarily unavailable",
  "code": "SERVICE_UNAVAILABLE",
  "details": {
    "retry_after": 60
  }
}
```

---

## Best Practices

### Error Handling

```javascript
try {
  const response = await Afftok.trackClick(params);
  
  if (!response.success) {
    switch (response.code) {
      case 'RATE_LIMITED':
        // Wait and retry
        await sleep(response.details.retry_after * 1000);
        return retry();
        
      case 'DUPLICATE_CLICK':
        // Ignore, already tracked
        return;
        
      case 'INVALID_SIGNATURE':
        // Fix signature generation
        console.error('Signature error:', response);
        break;
        
      default:
        // Log and handle
        console.error('Error:', response);
    }
  }
} catch (error) {
  // Network error, queue for retry
  Afftok.enqueue('click', params);
}
```

---

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- Status: https://status.afftok.com

