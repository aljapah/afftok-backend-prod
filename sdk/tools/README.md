# AffTok Testing Tools

Command-line tools for testing and validating AffTok integrations.

## Available Tools

| Tool | Description |
|------|-------------|
| `test_click_generator.js` | Generate test clicks |
| `test_conversion_simulator.js` | Simulate conversions |
| `signature_validator.js` | Generate and validate signatures |
| `latency_tester.js` | Test API latency and performance |

## Setup

### Environment Variables

```bash
export AFFTOK_API_KEY="your_api_key"
export AFFTOK_ADVERTISER_ID="your_advertiser_id"
export AFFTOK_BASE_URL="https://sandbox.api.afftok.com"  # Optional
```

### Make Executable

```bash
chmod +x *.js
```

---

## Click Generator

Generate test clicks for validating click tracking.

### Usage

```bash
node test_click_generator.js [options]
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--count, -c` | Number of clicks | 1 |
| `--offer, -o` | Offer ID | test_offer |
| `--delay, -d` | Delay between clicks (ms) | 100 |
| `--verbose, -v` | Verbose output | false |

### Examples

```bash
# Generate 10 clicks
node test_click_generator.js --count 10 --offer offer_123

# Verbose output
node test_click_generator.js -c 5 -o test_offer -v

# Custom delay
node test_click_generator.js --count 100 --delay 50
```

---

## Conversion Simulator

Simulate conversions for validating conversion tracking.

### Usage

```bash
node test_conversion_simulator.js [options]
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--count, -c` | Number of conversions | 1 |
| `--offer, -o` | Offer ID | test_offer |
| `--amount, -a` | Base amount | 29.99 |
| `--currency` | Currency code | USD |
| `--status, -s` | Status | approved |
| `--click-id` | Associated click ID | null |
| `--delay, -d` | Delay between conversions (ms) | 100 |
| `--verbose, -v` | Verbose output | false |

### Examples

```bash
# Simulate 10 conversions
node test_conversion_simulator.js --count 10 --offer offer_123

# With specific amount
node test_conversion_simulator.js -c 5 -a 99.99 -v

# With click attribution
node test_conversion_simulator.js --click-id click_abc123 --amount 49.99
```

---

## Signature Validator

Generate and validate HMAC-SHA256 signatures.

### Usage

```bash
node signature_validator.js [mode] [options]
```

### Modes

| Mode | Description |
|------|-------------|
| `--generate, -g` | Generate new signature |
| `--validate, -v` | Validate existing signature |

### Options

| Option | Description |
|--------|-------------|
| `--api-key, -k` | API key |
| `--advertiser-id, -a` | Advertiser ID |
| `--timestamp, -t` | Timestamp (ms) |
| `--nonce, -n` | Nonce string |
| `--signature, -s` | Signature to validate |

### Examples

```bash
# Generate signature
node signature_validator.js --generate \
  --api-key your_key \
  --advertiser-id your_id

# Validate signature
node signature_validator.js --validate \
  --api-key your_key \
  --advertiser-id your_id \
  --timestamp 1234567890000 \
  --nonce abc123... \
  --signature def456...

# Interactive mode
node signature_validator.js
```

---

## Latency Tester

Test API latency and performance metrics.

### Usage

```bash
node latency_tester.js [options]
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--endpoint, -e` | Endpoint: click, conversion | click |
| `--iterations, -i` | Number of requests | 10 |
| `--concurrent, -c` | Concurrent requests | 1 |
| `--warmup, -w` | Warmup requests | 3 |
| `--verbose, -v` | Verbose output | false |

### Examples

```bash
# Test click endpoint
node latency_tester.js --endpoint click --iterations 100

# Test conversion endpoint with concurrency
node latency_tester.js -e conversion -i 50 -c 5 -v

# Quick test
node latency_tester.js -i 10
```

### Output

```
=== Latency Statistics ===

Min: 45.23ms
Max: 234.56ms
Avg: 78.45ms
Median: 65.12ms
P90: 120.34ms
P95: 156.78ms
P99: 210.45ms
Std Dev: 32.45ms
```

---

## Integration Testing

### Full Flow Test

```bash
# 1. Generate clicks
node test_click_generator.js -c 10 -o test_offer -v

# 2. Note click IDs from output
# 3. Simulate conversions with click attribution
node test_conversion_simulator.js -c 10 -o test_offer --click-id click_xxx -v

# 4. Check latency
node latency_tester.js -e click -i 100 -v
node latency_tester.js -e conversion -i 100 -v
```

### Signature Debugging

```bash
# Generate signature for debugging
node signature_validator.js -g -k your_key -a your_id

# Validate a failed request's signature
node signature_validator.js -v \
  -k your_key \
  -a your_id \
  -t 1234567890000 \
  -n abc123... \
  -s def456...
```

---

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check base URL
   - Verify network connectivity

2. **Invalid Signature**
   - Use signature_validator.js to debug
   - Check timestamp format (milliseconds)

3. **Rate Limited**
   - Increase delay between requests
   - Reduce concurrent requests

4. **High Latency**
   - Check network conditions
   - Try different regions

---

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com

