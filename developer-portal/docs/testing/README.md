# Testing & QA Documentation

Complete guide to testing your AffTok integration.

## Testing Tools

AffTok provides several tools for testing:

| Tool | Purpose |
|------|---------|
| Test API Keys | Sandbox environment |
| Click Generator | Simulate clicks |
| Conversion Simulator | Test conversions |
| Signature Validator | Verify signatures |
| Latency Tester | Measure performance |

---

## Test Environment

### Test API Keys

Use test keys for development:

```javascript
Afftok.init({
    apiKey: 'afftok_test_sk_...',  // Test key
    advertiserId: 'adv_test_123',
    debugMode: true,
});
```

Test keys:
- Don't affect production data
- Have relaxed rate limits
- Return mock responses
- Log verbose debugging info

### Test Endpoints

| Environment | Base URL |
|-------------|----------|
| Production | `https://api.afftok.com` |
| Sandbox | `https://sandbox.afftok.com` |

---

## Click Testing

### Generate Test Click

```javascript
// Using SDK
await Afftok.generateTestClick({
    offerId: 'off_test_123',
    metadata: { test: true }
});

// Using API
curl -X POST https://sandbox.afftok.com/api/test/click \
  -H "Authorization: Bearer TEST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "offer_id": "off_test_123",
    "country": "US",
    "device": "mobile"
  }'
```

### Click Generator Script

```javascript
// tools/test_click_generator.js
const { Afftok } = require('@afftok/web-sdk');

async function generateClicks(count, options = {}) {
    Afftok.init({
        apiKey: process.env.AFFTOK_TEST_API_KEY,
        advertiserId: process.env.AFFTOK_ADVERTISER_ID,
        debugMode: true,
    });

    const results = [];
    
    for (let i = 0; i < count; i++) {
        try {
            const result = await Afftok.trackClick({
                offerId: options.offerId || 'off_test_123',
                trackingCode: options.trackingCode || `test_${Date.now()}_${i}`,
                metadata: {
                    test: true,
                    batch: i,
                    ...options.metadata,
                },
            });
            results.push({ success: true, ...result });
        } catch (error) {
            results.push({ success: false, error: error.message });
        }
        
        // Rate limit protection
        if (options.delay) {
            await new Promise(r => setTimeout(r, options.delay));
        }
    }
    
    return results;
}

// Usage
generateClicks(10, {
    offerId: 'off_test_123',
    delay: 100,
}).then(results => {
    console.log('Generated:', results.filter(r => r.success).length);
    console.log('Failed:', results.filter(r => !r.success).length);
});
```

---

## Conversion Testing

### Generate Test Conversion

```javascript
// Using SDK
await Afftok.generateTestConversion({
    offerId: 'off_test_123',
    amount: 49.99,
    currency: 'USD',
});

// Using API
curl -X POST https://sandbox.afftok.com/api/test/conversion \
  -H "Authorization: Bearer TEST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "offer_id": "off_test_123",
    "click_id": "click_test_456",
    "amount": 49.99,
    "currency": "USD"
  }'
```

### Conversion Simulator Script

```javascript
// tools/test_conversion_simulator.js
const crypto = require('crypto');
const axios = require('axios');

class ConversionSimulator {
    constructor(apiKey, advertiserId) {
        this.apiKey = apiKey;
        this.advertiserId = advertiserId;
        this.baseUrl = 'https://sandbox.afftok.com';
    }

    generateSignature(payload, timestamp, nonce) {
        const message = `${JSON.stringify(payload)}${timestamp}${nonce}`;
        return crypto
            .createHmac('sha256', this.apiKey)
            .update(message)
            .digest('hex');
    }

    async simulateConversion(data) {
        const timestamp = Math.floor(Date.now() / 1000).toString();
        const nonce = crypto.randomBytes(16).toString('hex');
        
        const payload = {
            click_id: data.clickId || `click_test_${Date.now()}`,
            external_id: data.orderId || `order_test_${Date.now()}`,
            amount: data.amount || 49.99,
            currency: data.currency || 'USD',
            status: data.status || 'approved',
            advertiser_id: this.advertiserId,
        };

        const signature = this.generateSignature(payload, timestamp, nonce);

        const response = await axios.post(
            `${this.baseUrl}/api/postback`,
            payload,
            {
                headers: {
                    'Content-Type': 'application/json',
                    'X-API-Key': this.apiKey,
                    'X-Afftok-Timestamp': timestamp,
                    'X-Afftok-Nonce': nonce,
                    'X-Afftok-Signature': `sha256=${signature}`,
                },
            }
        );

        return response.data;
    }

    async simulateBatch(count, options = {}) {
        const results = [];
        
        for (let i = 0; i < count; i++) {
            try {
                const result = await this.simulateConversion({
                    clickId: `click_test_${Date.now()}_${i}`,
                    orderId: `order_test_${Date.now()}_${i}`,
                    amount: options.amount || Math.random() * 100,
                    currency: options.currency || 'USD',
                    status: options.status || 'approved',
                });
                results.push({ success: true, ...result });
            } catch (error) {
                results.push({ success: false, error: error.message });
            }
            
            if (options.delay) {
                await new Promise(r => setTimeout(r, options.delay));
            }
        }
        
        return results;
    }
}

// Usage
const simulator = new ConversionSimulator(
    process.env.AFFTOK_TEST_API_KEY,
    process.env.AFFTOK_ADVERTISER_ID
);

simulator.simulateBatch(5, {
    amount: 49.99,
    delay: 200,
}).then(results => {
    console.log('Results:', results);
});
```

---

## Signature Validation

### Signature Validator Script

```javascript
// tools/signature_validator.js
const crypto = require('crypto');

class SignatureValidator {
    constructor(apiKey) {
        this.apiKey = apiKey;
    }

    // Validate request signature
    validateRequest(body, timestamp, nonce, signature) {
        const message = `${body}${timestamp}${nonce}`;
        const expected = crypto
            .createHmac('sha256', this.apiKey)
            .update(message)
            .digest('hex');

        const provided = signature.replace('sha256=', '');
        
        return {
            valid: crypto.timingSafeEqual(
                Buffer.from(expected),
                Buffer.from(provided)
            ),
            expected: `sha256=${expected}`,
            provided: signature,
        };
    }

    // Validate signed link
    validateSignedLink(signedCode) {
        const parts = signedCode.split('.');
        
        if (parts.length !== 4) {
            return { valid: false, error: 'Invalid format' };
        }

        const [trackingCode, timestamp, nonce, signature] = parts;
        
        // Check expiration (24 hours default)
        const now = Math.floor(Date.now() / 1000);
        const linkTime = parseInt(timestamp);
        const ttl = 86400;
        
        if (now - linkTime > ttl) {
            return { valid: false, error: 'Link expired' };
        }

        // Verify signature
        const message = `${trackingCode}.${timestamp}.${nonce}`;
        const expected = crypto
            .createHmac('sha256', this.apiKey)
            .update(message)
            .digest('hex')
            .substring(0, 12);

        return {
            valid: expected === signature,
            trackingCode,
            timestamp: new Date(linkTime * 1000).toISOString(),
            expiresAt: new Date((linkTime + ttl) * 1000).toISOString(),
            nonce,
        };
    }

    // Generate a new signed link for testing
    generateSignedLink(trackingCode) {
        const timestamp = Math.floor(Date.now() / 1000);
        const nonce = crypto.randomBytes(6).toString('hex');
        
        const message = `${trackingCode}.${timestamp}.${nonce}`;
        const signature = crypto
            .createHmac('sha256', this.apiKey)
            .update(message)
            .digest('hex')
            .substring(0, 12);

        return `${trackingCode}.${timestamp}.${nonce}.${signature}`;
    }
}

// Usage
const validator = new SignatureValidator(process.env.AFFTOK_API_KEY);

// Validate a signed link
const result = validator.validateSignedLink('abc123.1701432000.xyz789.a1b2c3d4e5f6');
console.log('Validation:', result);

// Generate a new signed link
const signedLink = validator.generateSignedLink('test_code_123');
console.log('Generated:', signedLink);
```

---

## Latency Testing

### Latency Tester Script

```javascript
// tools/latency_tester.js
const axios = require('axios');

class LatencyTester {
    constructor(baseUrl, apiKey) {
        this.baseUrl = baseUrl;
        this.apiKey = apiKey;
    }

    async measureEndpoint(endpoint, method = 'GET', data = null) {
        const start = process.hrtime.bigint();
        
        try {
            const config = {
                method,
                url: `${this.baseUrl}${endpoint}`,
                headers: {
                    'Authorization': `Bearer ${this.apiKey}`,
                    'Content-Type': 'application/json',
                },
            };
            
            if (data) {
                config.data = data;
            }
            
            await axios(config);
            
            const end = process.hrtime.bigint();
            const latencyMs = Number(end - start) / 1000000;
            
            return { success: true, latencyMs };
        } catch (error) {
            const end = process.hrtime.bigint();
            const latencyMs = Number(end - start) / 1000000;
            
            return { 
                success: false, 
                latencyMs, 
                error: error.message 
            };
        }
    }

    async runBenchmark(endpoint, iterations = 100) {
        const results = [];
        
        for (let i = 0; i < iterations; i++) {
            const result = await this.measureEndpoint(endpoint);
            results.push(result);
            
            // Small delay to avoid rate limiting
            await new Promise(r => setTimeout(r, 10));
        }
        
        const successful = results.filter(r => r.success);
        const latencies = successful.map(r => r.latencyMs).sort((a, b) => a - b);
        
        return {
            total: iterations,
            successful: successful.length,
            failed: iterations - successful.length,
            latency: {
                min: latencies[0],
                max: latencies[latencies.length - 1],
                avg: latencies.reduce((a, b) => a + b, 0) / latencies.length,
                p50: latencies[Math.floor(latencies.length * 0.5)],
                p90: latencies[Math.floor(latencies.length * 0.9)],
                p95: latencies[Math.floor(latencies.length * 0.95)],
                p99: latencies[Math.floor(latencies.length * 0.99)],
            },
        };
    }
}

// Usage
const tester = new LatencyTester(
    'https://api.afftok.com',
    process.env.AFFTOK_API_KEY
);

async function runTests() {
    console.log('Testing /api/health...');
    const healthResult = await tester.runBenchmark('/api/health', 50);
    console.log('Health endpoint:', healthResult);
    
    console.log('\nTesting /api/offers...');
    const offersResult = await tester.runBenchmark('/api/offers', 50);
    console.log('Offers endpoint:', offersResult);
}

runTests();
```

---

## E2E Test Scenarios

### Admin API: Run E2E Tests

```bash
# Run a specific scenario
curl -X POST https://api.afftok.com/api/admin/e2e-tests/run \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"scenario": "click_then_conversion"}'

# List available scenarios
curl https://api.afftok.com/api/admin/e2e-tests/scenarios \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Get test history
curl https://api.afftok.com/api/admin/e2e-tests/history \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Available Scenarios

| Scenario | Description |
|----------|-------------|
| `click_only` | Track a click, verify recorded |
| `click_then_conversion` | Track click, then conversion |
| `signed_link_flow` | Full signed link validation |
| `geo_rule_block` | Verify geo blocking works |
| `rate_limit_test` | Test rate limiting |
| `offline_queue` | Test offline queue sync |
| `webhook_delivery` | Test webhook delivery |
| `zero_drop_recovery` | Test crash recovery |

---

## Pre-Launch Checklist

### Run Preflight Check

```bash
curl https://api.afftok.com/api/admin/preflight/check \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Response

```json
{
  "success": true,
  "data": {
    "status": "pass",
    "checks": [
      { "name": "database_health", "status": "pass" },
      { "name": "redis_health", "status": "pass" },
      { "name": "wal_status", "status": "pass" },
      { "name": "streams_status", "status": "pass" },
      { "name": "consistency", "status": "pass" },
      { "name": "security", "status": "pass" },
      { "name": "performance", "status": "pass" }
    ],
    "warnings": [],
    "ready_for_deploy": true
  }
}
```

### Manual Checklist

#### Integration

- [ ] API keys configured correctly
- [ ] Test environment verified working
- [ ] Click tracking tested
- [ ] Conversion tracking tested
- [ ] Webhook delivery confirmed
- [ ] Signature verification working
- [ ] Error handling implemented

#### Security

- [ ] Using HTTPS everywhere
- [ ] API keys in environment variables
- [ ] IP allowlisting configured
- [ ] Rate limits tested
- [ ] Geo rules configured

#### Performance

- [ ] Latency within acceptable range
- [ ] No timeout errors
- [ ] Offline queue working
- [ ] Retry logic verified

#### Production

- [ ] Test keys replaced with live keys
- [ ] Debug mode disabled
- [ ] Monitoring configured
- [ ] Alerts set up

---

## Debugging

### Enable Debug Mode

```javascript
Afftok.init({
    apiKey: 'afftok_live_sk_...',
    advertiserId: 'adv_123456',
    debugMode: true,  // Enable verbose logging
});
```

### Check SDK Logs

```javascript
// Get internal state
console.log('Queue size:', Afftok.getQueueSize());
console.log('User ID:', Afftok.getUserId());
console.log('Fingerprint:', Afftok.getDeviceFingerprint());
```

### API Response Debugging

```javascript
try {
    const result = await Afftok.trackClick({ offerId: 'off_123' });
    console.log('Success:', result);
} catch (error) {
    console.error('Error code:', error.code);
    console.error('Error message:', error.message);
    console.error('Details:', error.details);
    console.error('Correlation ID:', error.correlationId);
}
```

---

## Common Issues

### "INVALID_API_KEY" Error

1. Check key format: `afftok_live_sk_...` or `afftok_test_sk_...`
2. Verify no whitespace
3. Confirm correct environment
4. Check key hasn't been revoked

### "RATE_LIMITED" Error

1. Implement exponential backoff
2. Check `Retry-After` header
3. Reduce request frequency
4. Consider caching

### Signature Verification Failing

1. Use raw request body (not parsed JSON)
2. Verify signing key matches
3. Check timestamp is current
4. Ensure nonce is unique

### Webhook Not Received

1. Verify endpoint is accessible
2. Check for firewall blocks
3. Confirm HTTPS is working
4. Review webhook logs in admin

---

Next: [Operational Guidelines](../operations/README.md)

