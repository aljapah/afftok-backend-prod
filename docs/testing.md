# AffTok Testing Guide

Complete guide for testing AffTok integrations.

## Overview

This guide covers:
- Test environment setup
- SDK testing
- API testing
- End-to-end testing
- Debugging tools

---

## Test Environment

### Sandbox Mode

AffTok provides a sandbox environment for testing:

```
Base URL: https://sandbox.api.afftok.com
```

### Test Credentials

Generate test credentials in the dashboard:

1. Navigate to Settings → API Keys
2. Click "Generate Test Key"
3. Use prefix: `afftok_test_sk_`

### Test vs Production

| Feature | Test | Production |
|---------|------|------------|
| Base URL | sandbox.api.afftok.com | api.afftok.com |
| API Key Prefix | `afftok_test_sk_` | `afftok_live_sk_` |
| Data Retention | 24 hours | Permanent |
| Rate Limits | Higher | Standard |
| Webhooks | Test endpoint | Real endpoint |

---

## SDK Testing

### Debug Mode

Enable debug mode to see detailed logs:

```kotlin
// Android
Afftok.init(context, AfftokOptions(
    apiKey = "afftok_test_sk_xxx",
    advertiserId = "test_advertiser",
    debug = true  // Enable debug logging
))
```

```swift
// iOS
Afftok.shared.initialize(options: AfftokOptions(
    apiKey: "afftok_test_sk_xxx",
    advertiserId: "test_advertiser",
    debug: true
))
```

```dart
// Flutter
await Afftok.instance.initialize(AfftokOptions(
    apiKey: 'afftok_test_sk_xxx',
    advertiserId: 'test_advertiser',
    debug: true,
));
```

### Testing Clicks

```kotlin
// Test click tracking
val response = Afftok.trackClick(ClickParams(
    offerId = "test_offer",
    trackingCode = "test_campaign",
    subId1 = "test_source"
))

// Verify response
assert(response.success)
assert(response.data?.containsKey("click_id") == true)
```

### Testing Conversions

```kotlin
// Test conversion tracking
val response = Afftok.trackConversion(ConversionParams(
    offerId = "test_offer",
    transactionId = "test_txn_${System.currentTimeMillis()}",
    amount = 9.99,
    status = "approved"
))

// Verify response
assert(response.success)
assert(response.data?.containsKey("conversion_id") == true)
```

### Testing Offline Queue

```kotlin
// Disable network
// Track events while offline
Afftok.trackClick(ClickParams(offerId = "test_offer"))
Afftok.trackConversion(ConversionParams(
    offerId = "test_offer",
    transactionId = "offline_txn_1"
))

// Check queue
val pendingCount = Afftok.getPendingCount()
assert(pendingCount == 2)

// Enable network
// Manually flush
Afftok.flush {
    val newCount = Afftok.getPendingCount()
    assert(newCount == 0)
}
```

---

## API Testing

### Using cURL

```bash
# Test click endpoint
curl -X POST https://sandbox.api.afftok.com/api/sdk/click \
  -H "Content-Type: application/json" \
  -H "X-API-Key: afftok_test_sk_xxx" \
  -d '{
    "api_key": "afftok_test_sk_xxx",
    "advertiser_id": "test_advertiser",
    "offer_id": "test_offer",
    "timestamp": 1234567890000,
    "nonce": "test_nonce_12345678901234567890",
    "signature": "calculated_signature"
  }'
```

```bash
# Test conversion endpoint
curl -X POST https://sandbox.api.afftok.com/api/sdk/conversion \
  -H "Content-Type: application/json" \
  -H "X-API-Key: afftok_test_sk_xxx" \
  -d '{
    "api_key": "afftok_test_sk_xxx",
    "advertiser_id": "test_advertiser",
    "offer_id": "test_offer",
    "transaction_id": "test_txn_123",
    "amount": 29.99,
    "status": "approved",
    "timestamp": 1234567890000,
    "nonce": "test_nonce_12345678901234567890",
    "signature": "calculated_signature"
  }'
```

### Using Postman

Import the AffTok Postman collection:

1. Download from: https://docs.afftok.com/postman
2. Import into Postman
3. Set environment variables:
   - `api_key`: Your test API key
   - `advertiser_id`: Your advertiser ID
   - `base_url`: https://sandbox.api.afftok.com

---

## Test Utilities

### Signature Generator

```javascript
// test_signature_generator.js
const crypto = require('crypto');

function generateTestSignature(apiKey, advertiserId) {
  const timestamp = Date.now();
  const nonce = generateNonce();
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  const signature = crypto.createHmac('sha256', apiKey).update(dataToSign).digest('hex');
  
  return { timestamp, nonce, signature };
}

function generateNonce() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  return Array(32).fill(0).map(() => chars[Math.floor(Math.random() * chars.length)]).join('');
}

// Usage
const { timestamp, nonce, signature } = generateTestSignature(
  'afftok_test_sk_xxx',
  'test_advertiser'
);
console.log({ timestamp, nonce, signature });
```

### Click Simulator

```javascript
// test_click_generator.js
const axios = require('axios');
const crypto = require('crypto');

const config = {
  apiKey: process.env.AFFTOK_TEST_API_KEY,
  advertiserId: process.env.AFFTOK_TEST_ADVERTISER_ID,
  baseUrl: 'https://sandbox.api.afftok.com',
};

async function simulateClick(offerId, options = {}) {
  const timestamp = Date.now();
  const nonce = generateNonce();
  const signature = generateSignature(timestamp, nonce);
  
  const payload = {
    api_key: config.apiKey,
    advertiser_id: config.advertiserId,
    offer_id: offerId,
    timestamp,
    nonce,
    signature,
    ...options,
  };
  
  try {
    const response = await axios.post(
      `${config.baseUrl}/api/sdk/click`,
      payload,
      { headers: { 'X-API-Key': config.apiKey } }
    );
    console.log('Click tracked:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error:', error.response?.data || error.message);
    throw error;
  }
}

// Simulate multiple clicks
async function simulateClicks(count, offerId) {
  const results = [];
  for (let i = 0; i < count; i++) {
    const result = await simulateClick(offerId, {
      tracking_code: `test_campaign_${i}`,
      sub_id_1: `source_${i % 3}`,
    });
    results.push(result);
    await sleep(100); // Rate limit protection
  }
  return results;
}

// Usage
simulateClicks(10, 'test_offer').then(console.log);
```

### Conversion Simulator

```javascript
// test_conversion_simulator.js
const axios = require('axios');

async function simulateConversion(offerId, clickId, amount) {
  const timestamp = Date.now();
  const nonce = generateNonce();
  const signature = generateSignature(timestamp, nonce);
  
  const payload = {
    api_key: config.apiKey,
    advertiser_id: config.advertiserId,
    offer_id: offerId,
    transaction_id: `test_txn_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
    click_id: clickId,
    amount,
    currency: 'USD',
    status: 'approved',
    timestamp,
    nonce,
    signature,
  };
  
  try {
    const response = await axios.post(
      `${config.baseUrl}/api/sdk/conversion`,
      payload,
      { headers: { 'X-API-Key': config.apiKey } }
    );
    console.log('Conversion tracked:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error:', error.response?.data || error.message);
    throw error;
  }
}

// Usage
simulateConversion('test_offer', 'click_abc123', 29.99);
```

### Latency Tester

```javascript
// test_latency.js
async function measureLatency(endpoint, payload, iterations = 10) {
  const latencies = [];
  
  for (let i = 0; i < iterations; i++) {
    const start = Date.now();
    
    try {
      await axios.post(endpoint, payload, {
        headers: { 'X-API-Key': config.apiKey },
      });
      const latency = Date.now() - start;
      latencies.push(latency);
      console.log(`Request ${i + 1}: ${latency}ms`);
    } catch (error) {
      console.error(`Request ${i + 1} failed:`, error.message);
    }
    
    await sleep(100);
  }
  
  const avg = latencies.reduce((a, b) => a + b, 0) / latencies.length;
  const min = Math.min(...latencies);
  const max = Math.max(...latencies);
  const p95 = latencies.sort((a, b) => a - b)[Math.floor(latencies.length * 0.95)];
  
  console.log('\nLatency Results:');
  console.log(`  Average: ${avg.toFixed(2)}ms`);
  console.log(`  Min: ${min}ms`);
  console.log(`  Max: ${max}ms`);
  console.log(`  P95: ${p95}ms`);
  
  return { avg, min, max, p95, latencies };
}
```

---

## End-to-End Testing

### Test Scenario: Full Conversion Flow

```javascript
async function testFullConversionFlow() {
  console.log('=== Full Conversion Flow Test ===\n');
  
  // Step 1: Track click
  console.log('Step 1: Tracking click...');
  const clickResponse = await simulateClick('test_offer', {
    tracking_code: 'e2e_test',
    sub_id_1: 'test_source',
  });
  const clickId = clickResponse.data?.click_id;
  console.log(`Click ID: ${clickId}\n`);
  
  // Step 2: Simulate user journey (wait)
  console.log('Step 2: Simulating user journey...');
  await sleep(2000);
  
  // Step 3: Track conversion
  console.log('Step 3: Tracking conversion...');
  const conversionResponse = await simulateConversion(
    'test_offer',
    clickId,
    49.99
  );
  const conversionId = conversionResponse.data?.conversion_id;
  console.log(`Conversion ID: ${conversionId}\n`);
  
  // Step 4: Verify in dashboard
  console.log('Step 4: Verify in dashboard');
  console.log(`https://dashboard.afftok.com/conversions/${conversionId}\n`);
  
  console.log('=== Test Complete ===');
}

testFullConversionFlow();
```

### Test Scenario: Offline Recovery

```javascript
async function testOfflineRecovery() {
  console.log('=== Offline Recovery Test ===\n');
  
  // Step 1: Queue events while "offline"
  console.log('Step 1: Queueing events...');
  Afftok.enqueue('click', { offer_id: 'test_offer' });
  Afftok.enqueue('conversion', { 
    offer_id: 'test_offer',
    transaction_id: 'offline_txn_1'
  });
  
  const queueSize = Afftok.getPendingCount();
  console.log(`Queue size: ${queueSize}\n`);
  
  // Step 2: Flush queue
  console.log('Step 2: Flushing queue...');
  await Afftok.flush();
  
  const newQueueSize = Afftok.getPendingCount();
  console.log(`Queue size after flush: ${newQueueSize}\n`);
  
  // Step 3: Verify
  console.log('Step 3: Verifying...');
  assert(newQueueSize < queueSize, 'Queue should be smaller after flush');
  
  console.log('=== Test Complete ===');
}
```

---

## Debugging

### Common Issues

#### 1. Invalid Signature

```javascript
// Check signature generation
const testSignature = generateSignature(apiKey, advertiserId, timestamp, nonce);
console.log('Generated signature:', testSignature);

// Verify data to sign
const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
console.log('Data to sign:', dataToSign);
```

#### 2. Expired Requests

```javascript
// Check timestamp
const now = Date.now();
const requestTimestamp = 1234567890000;
const age = now - requestTimestamp;
console.log(`Request age: ${age}ms (max: 300000ms)`);
```

#### 3. Rate Limiting

```javascript
// Check rate limit headers
const response = await axios.post(url, payload);
console.log('Rate limit:', response.headers['x-ratelimit-limit']);
console.log('Remaining:', response.headers['x-ratelimit-remaining']);
console.log('Reset:', response.headers['x-ratelimit-reset']);
```

### Debug Logging

Enable verbose logging:

```kotlin
// Android
Afftok.init(context, AfftokOptions(
    apiKey = "xxx",
    advertiserId = "xxx",
    debug = true
))

// Check Logcat for [AffTok SDK] messages
```

---

## Continuous Integration

### GitHub Actions Example

```yaml
name: AffTok Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: npm install
      
      - name: Run integration tests
        env:
          AFFTOK_TEST_API_KEY: ${{ secrets.AFFTOK_TEST_API_KEY }}
          AFFTOK_TEST_ADVERTISER_ID: ${{ secrets.AFFTOK_TEST_ADVERTISER_ID }}
        run: npm test
```

---

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- Status: https://status.afftok.com

