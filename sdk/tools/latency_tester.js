#!/usr/bin/env node

/**
 * AffTok Latency Tester Tool
 * 
 * Tests API latency and performance metrics.
 * 
 * Usage:
 *   node latency_tester.js --endpoint click --iterations 100
 *   node latency_tester.js --help
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
    endpoint: 'click',
    iterations: 10,
    concurrent: 1,
    warmup: 3,
    verbose: false,
    help: false,
  };

  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '--endpoint':
      case '-e':
        options.endpoint = args[++i];
        break;
      case '--iterations':
      case '-i':
        options.iterations = parseInt(args[++i], 10);
        break;
      case '--concurrent':
      case '-c':
        options.concurrent = parseInt(args[++i], 10);
        break;
      case '--warmup':
      case '-w':
        options.warmup = parseInt(args[++i], 10);
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

// Build payload based on endpoint
function buildPayload(endpoint) {
  const timestamp = Date.now();
  const nonce = generateNonce();
  const signature = generateSignature(timestamp, nonce);

  const basePayload = {
    api_key: config.apiKey,
    advertiser_id: config.advertiserId,
    offer_id: 'test_offer',
    timestamp,
    nonce,
    signature,
  };

  if (endpoint === 'conversion') {
    return {
      ...basePayload,
      transaction_id: generateTransactionId(),
      amount: 29.99,
      currency: 'USD',
      status: 'approved',
    };
  }

  return basePayload;
}

// Send HTTP request and measure latency
function sendRequest(endpoint, payload) {
  return new Promise((resolve, reject) => {
    const path = endpoint === 'click' ? '/api/sdk/click' : '/api/sdk/conversion';
    const url = new URL(`${config.baseUrl}${path}`);
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
      },
    };

    const startTime = process.hrtime.bigint();

    const req = lib.request(options, (res) => {
      let data = '';
      
      res.on('data', (chunk) => data += chunk);
      
      res.on('end', () => {
        const endTime = process.hrtime.bigint();
        const latencyNs = Number(endTime - startTime);
        const latencyMs = latencyNs / 1000000;

        resolve({
          statusCode: res.statusCode,
          latencyMs,
          success: res.statusCode >= 200 && res.statusCode < 300,
        });
      });
    });

    req.on('error', (error) => {
      const endTime = process.hrtime.bigint();
      const latencyNs = Number(endTime - startTime);
      const latencyMs = latencyNs / 1000000;

      resolve({
        statusCode: 0,
        latencyMs,
        success: false,
        error: error.message,
      });
    });

    req.write(JSON.stringify(payload));
    req.end();
  });
}

// Calculate statistics
function calculateStats(latencies) {
  if (latencies.length === 0) return null;

  const sorted = [...latencies].sort((a, b) => a - b);
  const sum = sorted.reduce((a, b) => a + b, 0);
  const avg = sum / sorted.length;
  
  const squaredDiffs = sorted.map(x => Math.pow(x - avg, 2));
  const avgSquaredDiff = squaredDiffs.reduce((a, b) => a + b, 0) / sorted.length;
  const stdDev = Math.sqrt(avgSquaredDiff);

  return {
    count: sorted.length,
    min: sorted[0],
    max: sorted[sorted.length - 1],
    avg,
    median: sorted[Math.floor(sorted.length / 2)],
    p90: sorted[Math.floor(sorted.length * 0.90)],
    p95: sorted[Math.floor(sorted.length * 0.95)],
    p99: sorted[Math.floor(sorted.length * 0.99)],
    stdDev,
  };
}

// Print help
function printHelp() {
  console.log(`
AffTok Latency Tester Tool

Usage:
  node latency_tester.js [options]

Options:
  --endpoint, -e   Endpoint to test: click, conversion (default: click)
  --iterations, -i Number of requests (default: 10)
  --concurrent, -c Concurrent requests (default: 1)
  --warmup, -w     Warmup requests (default: 3)
  --verbose, -v    Enable verbose output
  --api-key        API key (or set AFFTOK_API_KEY env var)
  --advertiser-id  Advertiser ID (or set AFFTOK_ADVERTISER_ID env var)
  --base-url       Base URL (or set AFFTOK_BASE_URL env var)
  --help, -h       Show this help

Examples:
  node latency_tester.js --endpoint click --iterations 100
  node latency_tester.js -e conversion -i 50 -c 5 -v

Environment Variables:
  AFFTOK_API_KEY        Your API key
  AFFTOK_ADVERTISER_ID  Your advertiser ID
  AFFTOK_BASE_URL       API base URL (default: https://sandbox.api.afftok.com)
`);
}

// Run test batch
async function runBatch(endpoint, count, concurrent, verbose) {
  const results = [];
  const batches = Math.ceil(count / concurrent);

  for (let b = 0; b < batches; b++) {
    const batchSize = Math.min(concurrent, count - b * concurrent);
    const promises = [];

    for (let i = 0; i < batchSize; i++) {
      const payload = buildPayload(endpoint);
      promises.push(sendRequest(endpoint, payload));
    }

    const batchResults = await Promise.all(promises);
    results.push(...batchResults);

    if (verbose) {
      const batchNum = b + 1;
      const successCount = batchResults.filter(r => r.success).length;
      const avgLatency = (batchResults.reduce((a, r) => a + r.latencyMs, 0) / batchResults.length).toFixed(2);
      console.log(`Batch ${batchNum}/${batches}: ${successCount}/${batchSize} success, avg ${avgLatency}ms`);
    } else {
      batchResults.forEach(r => process.stdout.write(r.success ? '.' : 'x'));
    }
  }

  return results;
}

// Main function
async function main() {
  const options = parseArgs();

  if (options.help) {
    printHelp();
    return;
  }

  console.log('=== AffTok Latency Tester ===\n');
  console.log(`Base URL: ${config.baseUrl}`);
  console.log(`Endpoint: ${options.endpoint}`);
  console.log(`Iterations: ${options.iterations}`);
  console.log(`Concurrent: ${options.concurrent}`);
  console.log(`Warmup: ${options.warmup}\n`);

  // Warmup
  if (options.warmup > 0) {
    console.log('Running warmup...');
    await runBatch(options.endpoint, options.warmup, 1, false);
    console.log(' done\n');
  }

  // Main test
  console.log('Running test...');
  const startTime = Date.now();
  const results = await runBatch(options.endpoint, options.iterations, options.concurrent, options.verbose);
  const totalTime = Date.now() - startTime;

  console.log('\n');

  // Calculate results
  const successful = results.filter(r => r.success);
  const failed = results.filter(r => !r.success);
  const latencies = successful.map(r => r.latencyMs);
  const stats = calculateStats(latencies);

  console.log('=== Results ===\n');
  console.log(`Total Requests: ${results.length}`);
  console.log(`Successful: ${successful.length} (${((successful.length / results.length) * 100).toFixed(1)}%)`);
  console.log(`Failed: ${failed.length}`);
  console.log(`Total Time: ${totalTime}ms`);
  console.log(`Throughput: ${((results.length / totalTime) * 1000).toFixed(2)} req/s`);

  if (stats) {
    console.log('\n=== Latency Statistics ===\n');
    console.log(`Min: ${stats.min.toFixed(2)}ms`);
    console.log(`Max: ${stats.max.toFixed(2)}ms`);
    console.log(`Avg: ${stats.avg.toFixed(2)}ms`);
    console.log(`Median: ${stats.median.toFixed(2)}ms`);
    console.log(`P90: ${stats.p90.toFixed(2)}ms`);
    console.log(`P95: ${stats.p95.toFixed(2)}ms`);
    console.log(`P99: ${stats.p99.toFixed(2)}ms`);
    console.log(`Std Dev: ${stats.stdDev.toFixed(2)}ms`);
  }

  if (failed.length > 0 && options.verbose) {
    console.log('\n=== Errors ===\n');
    const errorCounts = {};
    failed.forEach(r => {
      const key = r.error || `HTTP ${r.statusCode}`;
      errorCounts[key] = (errorCounts[key] || 0) + 1;
    });
    Object.entries(errorCounts).forEach(([error, count]) => {
      console.log(`${error}: ${count}`);
    });
  }

  // Histogram
  if (stats && options.verbose) {
    console.log('\n=== Latency Histogram ===\n');
    const buckets = [10, 25, 50, 100, 200, 500, 1000, 2000, 5000];
    let prev = 0;
    buckets.forEach(bucket => {
      const count = latencies.filter(l => l > prev && l <= bucket).length;
      const bar = '█'.repeat(Math.ceil(count / results.length * 50));
      console.log(`${prev.toString().padStart(5)}-${bucket.toString().padStart(5)}ms: ${bar} ${count}`);
      prev = bucket;
    });
    const remaining = latencies.filter(l => l > buckets[buckets.length - 1]).length;
    if (remaining > 0) {
      console.log(`${buckets[buckets.length - 1].toString().padStart(5)}+    ms: ${'█'.repeat(Math.ceil(remaining / results.length * 50))} ${remaining}`);
    }
  }
}

main().catch(console.error);

