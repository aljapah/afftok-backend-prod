# Security Documentation

Complete security guide for AffTok integration.

## Overview

AffTok implements multiple layers of security:

- **API Key Authentication** - Secure access control
- **Request Signing** - HMAC-SHA256 signatures
- **Link Signing** - Tamper-proof tracking links
- **Replay Protection** - Nonce-based deduplication
- **Rate Limiting** - Abuse prevention
- **Geo Rules** - Geographic restrictions
- **Tenant Isolation** - Multi-tenant security

---

## API Key Security

### Key Format

```
afftok_live_sk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
       │    │  │
       │    │  └── 32-character random string
       │    └───── Key type (sk = secret key)
       └────────── Environment (live/test)
```

### Best Practices

1. **Never expose keys in client code**

   ```javascript
   // ❌ BAD - Key in frontend
   const apiKey = 'afftok_live_sk_...';
   
   // ✅ GOOD - Key on server only
   const apiKey = process.env.AFFTOK_API_KEY;
   ```

2. **Use environment variables**

   ```bash
   # .env
   AFFTOK_API_KEY=afftok_live_sk_...
   ```

3. **Rotate keys regularly**

   ```bash
   curl -X POST https://api.afftok.com/api/admin/api-keys/key_123/rotate \
     -H "Authorization: Bearer ADMIN_TOKEN"
   ```

4. **Use separate keys per environment**

   - `afftok_test_sk_...` for development
   - `afftok_live_sk_...` for production

5. **Restrict IP addresses**

   ```json
   {
     "allowed_ips": ["203.0.113.0/24", "198.51.100.50"]
   }
   ```

### Key Permissions

| Permission | Description |
|------------|-------------|
| `tracking:write` | Record clicks and conversions |
| `tracking:read` | Read tracking data |
| `stats:read` | Access statistics |
| `offers:read` | View offers |
| `postback:write` | Send postbacks |

---

## Request Signing

All SDK requests are signed with HMAC-SHA256.

### Signature Format

```
X-Afftok-Signature: sha256=<hex_digest>
X-Afftok-Timestamp: <unix_timestamp>
X-Afftok-Nonce: <random_string>
```

### Signature Calculation

```javascript
const crypto = require('crypto');

function signRequest(payload, timestamp, nonce, apiKey) {
  const message = `${payload}${timestamp}${nonce}`;
  return crypto
    .createHmac('sha256', apiKey)
    .update(message)
    .digest('hex');
}
```

### Verification

```javascript
function verifyRequest(req, apiKey) {
  const signature = req.headers['x-afftok-signature'];
  const timestamp = req.headers['x-afftok-timestamp'];
  const nonce = req.headers['x-afftok-nonce'];
  const body = JSON.stringify(req.body);
  
  // Check timestamp (5 minute window)
  const now = Math.floor(Date.now() / 1000);
  if (Math.abs(now - parseInt(timestamp)) > 300) {
    return false;
  }
  
  // Verify signature
  const expected = signRequest(body, timestamp, nonce, apiKey);
  return signature === `sha256=${expected}`;
}
```

---

## Link Signing

Tracking links are cryptographically signed to prevent tampering.

### Signed Link Format

```
https://track.afftok.com/c/<tracking_code>.<timestamp>.<nonce>.<signature>
```

Example:
```
https://track.afftok.com/c/abc123.1701432000.xyz789.a1b2c3d4e5f6
```

### Components

| Component | Description |
|-----------|-------------|
| `tracking_code` | User-offer binding code |
| `timestamp` | Unix timestamp (link creation) |
| `nonce` | Random string (replay protection) |
| `signature` | HMAC-SHA256 signature |

### TTL (Time-To-Live)

Links expire after a configurable TTL (default: 24 hours).

```json
{
  "ttl_seconds": 86400
}
```

### Replay Protection

Each nonce is stored in Redis to prevent replay attacks:

```
Key: replay:nonce:<nonce>
TTL: Same as link TTL
```

---

## Rate Limiting

### Default Limits

| Endpoint | Limit |
|----------|-------|
| Click tracking | 1000/min per IP |
| Postback | 100/min per API key |
| Stats API | 60/min per API key |
| Admin API | 30/min per admin |

### Per-Key Limits

API keys can have custom rate limits:

```json
{
  "rate_limit_per_minute": 120
}
```

### Rate Limit Headers

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1701432060
Retry-After: 30
```

### Handling Rate Limits

```javascript
async function makeRequest(url, options, retries = 3) {
  const response = await fetch(url, options);
  
  if (response.status === 429 && retries > 0) {
    const retryAfter = response.headers.get('Retry-After') || 60;
    await sleep(retryAfter * 1000);
    return makeRequest(url, options, retries - 1);
  }
  
  return response;
}
```

---

## Geo Rules

Block or allow traffic by country.

### Rule Types

| Mode | Description |
|------|-------------|
| `allow` | Only listed countries allowed |
| `block` | Listed countries blocked |

### Scope Levels

| Scope | Priority |
|-------|----------|
| Offer | Highest |
| Advertiser | Medium |
| Global | Lowest |

### Example

```json
{
  "scope_type": "offer",
  "scope_id": "off_123",
  "mode": "allow",
  "countries": ["US", "CA", "GB"]
}
```

---

## Tenant Isolation

### Data Isolation

Each tenant has completely isolated:

- Users and accounts
- Offers and networks
- Clicks and conversions
- API keys and webhooks
- Statistics and reports

### Redis Namespacing

```
tenant:{tenant_id}:stats:clicks
tenant:{tenant_id}:offers:cache
tenant:{tenant_id}:fraud:ips
```

### Database Scoping

All queries automatically include tenant filter:

```sql
SELECT * FROM clicks WHERE tenant_id = 'tenant_123' AND ...
```

---

## IP Allowlisting

Restrict API key access by IP:

```bash
curl -X POST https://api.afftok.com/api/admin/api-keys/key_123/allow-ip \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ip": "203.0.113.50"}'
```

### CIDR Support

```json
{
  "allowed_ips": [
    "203.0.113.0/24",
    "198.51.100.50"
  ]
}
```

---

## Fraud Prevention

### Bot Detection

Automatic detection of:

- Headless browsers
- Known bot user agents
- Datacenter IPs
- Tor exit nodes
- Impossible browser combinations

### Click Fingerprinting

Unique fingerprint per click:

```
SHA256(IP + UserAgent + Timestamp + Salt)
```

### Duplicate Prevention

- Click deduplication within time window
- Conversion deduplication by external ID
- Postback replay protection

---

## Data Encryption

### In Transit

- TLS 1.3 required
- HSTS enabled
- Certificate pinning (SDKs)

### At Rest

- Database encryption (AES-256)
- Redis encryption (TLS)
- Backup encryption

### Sensitive Fields

Hashed with Argon2id:

- API key secrets
- User passwords
- Webhook signing keys

---

## Security Headers

All responses include:

```http
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
```

---

## Audit Logging

All security events are logged:

| Event | Details Logged |
|-------|---------------|
| API key created | Key ID, creator, permissions |
| API key revoked | Key ID, revoker, reason |
| Login attempt | User, IP, success/failure |
| Rate limit hit | Key/IP, endpoint, count |
| Fraud detected | Type, IP, indicators |
| Geo block | Country, rule, offer |

### Viewing Audit Logs

```bash
curl https://api.afftok.com/api/admin/tenants/tenant_123/audit-logs \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

---

## Incident Response

### Compromised API Key

1. **Revoke immediately**

   ```bash
   curl -X POST https://api.afftok.com/api/admin/api-keys/key_123/revoke \
     -H "Authorization: Bearer ADMIN_TOKEN"
   ```

2. **Generate new key**

3. **Review audit logs**

4. **Update integrations**

### Suspicious Activity

1. **Block IP**

   ```bash
   curl -X POST https://api.afftok.com/api/admin/fraud/block-ip \
     -H "Authorization: Bearer ADMIN_TOKEN" \
     -d '{"ip": "192.168.1.1", "reason": "Suspicious activity"}'
   ```

2. **Review fraud logs**

3. **Adjust rate limits**

---

## Compliance

### Data Retention

| Data Type | Retention |
|-----------|-----------|
| Clicks | 90 days (configurable) |
| Conversions | 2 years |
| Audit logs | 1 year |
| API logs | 30 days |

### Data Export

Request data export via admin panel or API.

### Data Deletion

GDPR-compliant deletion available per user/tenant.

---

## Security Checklist

### Integration

- [ ] API keys stored in environment variables
- [ ] HTTPS used for all requests
- [ ] Request signatures verified
- [ ] Rate limit handling implemented
- [ ] Error handling doesn't expose secrets

### Production

- [ ] Test keys replaced with live keys
- [ ] IP allowlisting configured
- [ ] Geo rules set up
- [ ] Webhook signatures verified
- [ ] Audit logging reviewed

### Ongoing

- [ ] API keys rotated quarterly
- [ ] Audit logs reviewed weekly
- [ ] Rate limits monitored
- [ ] Fraud alerts configured

---

## Reporting Vulnerabilities

Report security issues to: security@afftok.com

Include:
- Description of vulnerability
- Steps to reproduce
- Potential impact
- Your contact information

We respond within 24 hours and follow responsible disclosure.

---

Next: [Integration Guides](../guides/README.md)

