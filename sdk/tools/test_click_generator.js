#!/usr/bin/env node

/**
 * AffTok Click Generator Test Tool
 * 
 * Generates test clicks for validating AffTok integration.
 * 
 * Usage:
 *   node test_click_generator.js --count 10 --offer offer_123
 *   node test_click_generator.js --help
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
    delay: 100,
    verbose: false,
    help: false,
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

// Send HTTP request
function sendRequest(payload) {
  return new Promise((resolve, reject) => {
    const url = new URL(`${config.baseUrl}/api/sdk/click`);
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

// Generate a single click
async function generateClick(offerId, index) {
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
    tracking_code: `test_campaign_${index}`,
    sub_id_1: `source_${index % 5}`,
    sub_id_2: `medium_${index % 3}`,
    sub_id_3: `campaign_${index % 2}`,
    device_info: {
      device_id: `test_device_${index}`,
      fingerprint: crypto.randomBytes(16).toString('hex'),
      platform: 'test-tool',
      sdk_version: '1.0.0',
    },
  };

  const start = Date.now();
  const response = await sendRequest(payload);
  const latency = Date.now() - start;

  return {
    index,
    success: response.statusCode >= 200 && response.statusCode < 300,
    statusCode: response.statusCode,
    latency,
    clickId: response.data?.data?.click_id,
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
AffTok Click Generator Test Tool

Usage:
  node test_click_generator.js [options]

Options:
  --count, -c      Number of clicks to generate (default: 1)
  --offer, -o      Offer ID (default: test_offer)
  --delay, -d      Delay between clicks in ms (default: 100)
  --verbose, -v    Enable verbose output
  --api-key        API key (or set AFFTOK_API_KEY env var)
  --advertiser-id  Advertiser ID (or set AFFTOK_ADVERTISER_ID env var)
  --base-url       Base URL (or set AFFTOK_BASE_URL env var)
  --help, -h       Show this help

Examples:
  node test_click_generator.js --count 10 --offer offer_123
  node test_click_generator.js -c 100 -o test_offer -d 50 -v

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

  console.log('=== AffTok Click Generator ===\n');
  console.log(`Base URL: ${config.baseUrl}`);
  console.log(`Offer ID: ${options.offer}`);
  console.log(`Count: ${options.count}`);
  console.log(`Delay: ${options.delay}ms\n`);

  const results = [];
  const startTime = Date.now();

  for (let i = 0; i < options.count; i++) {
    try {
      const result = await generateClick(options.offer, i);
      results.push(result);

      if (options.verbose) {
        const status = result.success ? '✓' : '✗';
        console.log(`${status} Click ${i + 1}: ${result.statusCode} (${result.latency}ms) ${result.clickId || result.error || ''}`);
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
  console.log(`Total Time: ${totalTime}ms`);
  console.log(`Avg Latency: ${avgLatency}ms`);
  console.log(`Min Latency: ${minLatency}ms`);
  console.log(`Max Latency: ${maxLatency}ms`);
  console.log(`Throughput: ${((results.length / totalTime) * 1000).toFixed(2)} req/s`);

  if (failed > 0) {
    console.log('\n=== Errors ===\n');
    results.filter(r => !r.success).forEach(r => {
      console.log(`Click ${r.index}: ${r.error || `HTTP ${r.statusCode}`}`);
    });
  }
}

main().catch(console.error);

