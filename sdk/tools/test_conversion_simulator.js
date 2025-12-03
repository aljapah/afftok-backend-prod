#!/usr/bin/env node

/**
 * AffTok Conversion Simulator Test Tool
 * 
 * Simulates conversions for validating AffTok integration.
 * 
 * Usage:
 *   node test_conversion_simulator.js --count 10 --offer offer_123 --amount 29.99
 *   node test_conversion_simulator.js --help
 */

const crypto = require('crypto');
const https = require('https');
const http = require('http');

// Configuration
const config = {
  apiKey: process.env.AFFTOK_API_KEY || 'afftok_test_sk_xxx',
  advertiserId: process.env.AFFTOK_ADVERTISER_ID || 'test_advertiser',
  baseUrl: process.env.AFFTOK_BASE_URL || 'https://sandbox.api.afftok.com',
};

// Parse command line arguments
function parseArgs() {
  const args = process.argv.slice(2);
  const options = {
    count: 1,
    offer: 'test_offer',
    amount: 29.99,
    currency: 'USD',
    status: 'approved',
    delay: 100,
    verbose: false,
    help: false,
    clickId: null,
  };

  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '--count':
      case '-c':
        options.count = parseInt(args[++i], 10);
        break;
      case '--offer':
      case '-o':
        options.offer = args[++i];
        break;
      case '--amount':
      case '-a':
        options.amount = parseFloat(args[++i]);
        break;
      case '--currency':
        options.currency = args[++i];
        break;
      case '--status':
      case '-s':
        options.status = args[++i];
        break;
      case '--click-id':
        options.clickId = args[++i];
        break;
      case '--delay':
      case '-d':
        options.delay = parseInt(args[++i], 10);
        break;
      case '--verbose':
      case '-v':
        options.verbose = true;
        break;
      case '--help':
      case '-h':
        options.help = true;
        break;
      case '--api-key':
        config.apiKey = args[++i];
        break;
      case '--advertiser-id':
        config.advertiserId = args[++i];
        break;
      case '--base-url':
        config.baseUrl = args[++i];
        break;
    }
  }

  return options;
}

// Generate HMAC-SHA256 signature
function generateSignature(timestamp, nonce) {
  const dataToSign = `${config.apiKey}|${config.advertiserId}|${timestamp}|${nonce}`;
  return crypto.createHmac('sha256', config.apiKey).update(dataToSign).digest('hex');
}

// Generate random nonce
function generateNonce(length = 32) {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

// Generate unique transaction ID
function generateTransactionId() {
  const timestamp = Date.now().toString(36);
  const random = crypto.randomBytes(4).toString('hex');
  return `txn_${timestamp}_${random}`;
}

// Send HTTP request
function sendRequest(payload) {
  return new Promise((resolve, reject) => {
    const url = new URL(`${config.baseUrl}/api/sdk/conversion`);
    const isHttps = url.protocol === 'https:';
    const lib = isHttps ? https : http;

    const options = {
      hostname: url.hostname,
      port: url.port || (isHttps ? 443 : 80),
      path: url.pathname,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': config.apiKey,
        'X-SDK-Version': '1.0.0',
        'X-SDK-Platform': 'test-tool',
      },
    };

    const req = lib.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        try {
          resolve({
            statusCode: res.statusCode,
            data: JSON.parse(data),
          });
        } catch (e) {
          resolve({
            statusCode: res.statusCode,
            data: data,
          });
        }
      });
    });

    req.on('error', reject);
    req.write(JSON.stringify(payload));
    req.end();
  });
}

// Simulate a single conversion
async function simulateConversion(offerId, options, index) {
  const timestamp = Date.now();
  const nonce = generateNonce();
  const signature = generateSignature(timestamp, nonce);

  const payload = {
    api_key: config.apiKey,
    advertiser_id: config.advertiserId,
    offer_id: offerId,
    transaction_id: generateTransactionId(),
    amount: options.amount + (Math.random() * 10 - 5), // Vary amount slightly
    currency: options.currency,
    status: options.status,
    timestamp,
    nonce,
    signature,
    device_info: {
      device_id: `test_device_${index}`,
      fingerprint: crypto.randomBytes(16).toString('hex'),
      platform: 'test-tool',
      sdk_version: '1.0.0',
    },
    custom_params: {
      test_index: index,
      source: 'conversion_simulator',
    },
  };

  if (options.clickId) {
    payload.click_id = options.clickId;
  }

  const start = Date.now();
  const response = await sendRequest(payload);
  const latency = Date.now() - start;

  return {
    index,
    transactionId: payload.transaction_id,
    amount: payload.amount.toFixed(2),
    success: response.statusCode >= 200 && response.statusCode < 300,
    statusCode: response.statusCode,
    latency,
    conversionId: response.data?.data?.conversion_id,
    error: response.data?.error,
  };
}

// Sleep helper
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

// Print help
function printHelp() {
  console.log(`
AffTok Conversion Simulator Test Tool

Usage:
  node test_conversion_simulator.js [options]

Options:
  --count, -c      Number of conversions to simulate (default: 1)
  --offer, -o      Offer ID (default: test_offer)
  --amount, -a     Base conversion amount (default: 29.99)
  --currency       Currency code (default: USD)
  --status, -s     Conversion status: pending, approved, rejected (default: approved)
  --click-id       Associated click ID (optional)
  --delay, -d      Delay between conversions in ms (default: 100)
  --verbose, -v    Enable verbose output
  --api-key        API key (or set AFFTOK_API_KEY env var)
  --advertiser-id  Advertiser ID (or set AFFTOK_ADVERTISER_ID env var)
  --base-url       Base URL (or set AFFTOK_BASE_URL env var)
  --help, -h       Show this help

Examples:
  node test_conversion_simulator.js --count 10 --offer offer_123 --amount 49.99
  node test_conversion_simulator.js -c 5 -o test_offer -s approved -v
  node test_conversion_simulator.js --click-id click_abc123 --amount 99.99

Environment Variables:
  AFFTOK_API_KEY        Your API key
  AFFTOK_ADVERTISER_ID  Your advertiser ID
  AFFTOK_BASE_URL       API base URL (default: https://sandbox.api.afftok.com)
`);
}

// Main function
async function main() {
  const options = parseArgs();

  if (options.help) {
    printHelp();
    return;
  }

  console.log('=== AffTok Conversion Simulator ===\n');
  console.log(`Base URL: ${config.baseUrl}`);
  console.log(`Offer ID: ${options.offer}`);
  console.log(`Count: ${options.count}`);
  console.log(`Base Amount: ${options.amount} ${options.currency}`);
  console.log(`Status: ${options.status}`);
  console.log(`Delay: ${options.delay}ms\n`);

  const results = [];
  const startTime = Date.now();
  let totalAmount = 0;

  for (let i = 0; i < options.count; i++) {
    try {
      const result = await simulateConversion(options.offer, options, i);
      results.push(result);

      if (result.success) {
        totalAmount += parseFloat(result.amount);
      }

      if (options.verbose) {
        const status = result.success ? '✓' : '✗';
        console.log(`${status} Conversion ${i + 1}: ${result.transactionId} - $${result.amount} (${result.latency}ms) ${result.conversionId || result.error || ''}`);
      } else {
        process.stdout.write(result.success ? '.' : 'x');
      }

      if (i < options.count - 1) {
        await sleep(options.delay);
      }
    } catch (error) {
      results.push({
        index: i,
        success: false,
        error: error.message,
      });
      process.stdout.write('E');
    }
  }

  const totalTime = Date.now() - startTime;
  const successful = results.filter(r => r.success).length;
  const failed = results.length - successful;
  const latencies = results.filter(r => r.latency).map(r => r.latency);
  const avgLatency = latencies.length > 0 
    ? (latencies.reduce((a, b) => a + b, 0) / latencies.length).toFixed(2)
    : 0;
  const minLatency = latencies.length > 0 ? Math.min(...latencies) : 0;
  const maxLatency = latencies.length > 0 ? Math.max(...latencies) : 0;

  console.log('\n\n=== Results ===\n');
  console.log(`Total: ${results.length}`);
  console.log(`Successful: ${successful}`);
  console.log(`Failed: ${failed}`);
  console.log(`Total Amount: $${totalAmount.toFixed(2)} ${options.currency}`);
  console.log(`Total Time: ${totalTime}ms`);
  console.log(`Avg Latency: ${avgLatency}ms`);
  console.log(`Min Latency: ${minLatency}ms`);
  console.log(`Max Latency: ${maxLatency}ms`);
  console.log(`Throughput: ${((results.length / totalTime) * 1000).toFixed(2)} req/s`);

  if (failed > 0) {
    console.log('\n=== Errors ===\n');
    results.filter(r => !r.success).forEach(r => {
      console.log(`Conversion ${r.index}: ${r.error || `HTTP ${r.statusCode}`}`);
    });
  }

  if (options.verbose && successful > 0) {
    console.log('\n=== Successful Conversions ===\n');
    results.filter(r => r.success).forEach(r => {
      console.log(`${r.transactionId}: $${r.amount} -> ${r.conversionId}`);
    });
  }
}

main().catch(console.error);

