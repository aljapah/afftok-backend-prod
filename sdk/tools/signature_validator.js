#!/usr/bin/env node

/**
 * AffTok Signature Validator Tool
 * 
 * Validates and generates HMAC-SHA256 signatures for AffTok API requests.
 * 
 * Usage:
 *   node signature_validator.js --generate --api-key xxx --advertiser-id yyy
 *   node signature_validator.js --validate --signature abc123 --api-key xxx --advertiser-id yyy --timestamp 123 --nonce xyz
 *   node signature_validator.js --help
 */

const crypto = require('crypto');

// Parse command line arguments
function parseArgs() {
  const args = process.argv.slice(2);
  const options = {
    mode: null, // 'generate' or 'validate'
    apiKey: process.env.AFFTOK_API_KEY || '',
    advertiserId: process.env.AFFTOK_ADVERTISER_ID || '',
    timestamp: null,
    nonce: null,
    signature: null,
    help: false,
  };

  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '--generate':
      case '-g':
        options.mode = 'generate';
        break;
      case '--validate':
      case '-v':
        options.mode = 'validate';
        break;
      case '--api-key':
      case '-k':
        options.apiKey = args[++i];
        break;
      case '--advertiser-id':
      case '-a':
        options.advertiserId = args[++i];
        break;
      case '--timestamp':
      case '-t':
        options.timestamp = parseInt(args[++i], 10);
        break;
      case '--nonce':
      case '-n':
        options.nonce = args[++i];
        break;
      case '--signature':
      case '-s':
        options.signature = args[++i];
        break;
      case '--help':
      case '-h':
        options.help = true;
        break;
    }
  }

  return options;
}

// Generate HMAC-SHA256 signature
function generateSignature(apiKey, advertiserId, timestamp, nonce) {
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  return crypto.createHmac('sha256', apiKey).update(dataToSign).digest('hex');
}

// Generate random nonce
function generateNonce(length = 32) {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  const randomBytes = crypto.randomBytes(length);
  for (let i = 0; i < length; i++) {
    result += chars[randomBytes[i] % chars.length];
  }
  return result;
}

// Validate signature
function validateSignature(apiKey, advertiserId, timestamp, nonce, providedSignature) {
  const expectedSignature = generateSignature(apiKey, advertiserId, timestamp, nonce);
  const isValid = crypto.timingSafeEqual(
    Buffer.from(expectedSignature),
    Buffer.from(providedSignature)
  );
  return {
    isValid,
    expectedSignature,
    providedSignature,
  };
}

// Print help
function printHelp() {
  console.log(`
AffTok Signature Validator Tool

Usage:
  node signature_validator.js [mode] [options]

Modes:
  --generate, -g   Generate a new signature
  --validate, -v   Validate an existing signature

Options:
  --api-key, -k        API key (or set AFFTOK_API_KEY env var)
  --advertiser-id, -a  Advertiser ID (or set AFFTOK_ADVERTISER_ID env var)
  --timestamp, -t      Unix timestamp in milliseconds
  --nonce, -n          32-character nonce string
  --signature, -s      Signature to validate (for validate mode)
  --help, -h           Show this help

Examples:
  # Generate new signature
  node signature_validator.js --generate --api-key your_key --advertiser-id your_id

  # Validate existing signature
  node signature_validator.js --validate \\
    --api-key your_key \\
    --advertiser-id your_id \\
    --timestamp 1234567890000 \\
    --nonce abc123... \\
    --signature def456...

Environment Variables:
  AFFTOK_API_KEY        Your API key
  AFFTOK_ADVERTISER_ID  Your advertiser ID
`);
}

// Generate mode
function runGenerate(options) {
  if (!options.apiKey || !options.advertiserId) {
    console.error('Error: API key and advertiser ID are required');
    console.error('Use --api-key and --advertiser-id or set environment variables');
    process.exit(1);
  }

  const timestamp = options.timestamp || Date.now();
  const nonce = options.nonce || generateNonce();
  const signature = generateSignature(options.apiKey, options.advertiserId, timestamp, nonce);
  const dataToSign = `${options.apiKey}|${options.advertiserId}|${timestamp}|${nonce}`;

  console.log('=== Generated Signature ===\n');
  console.log(`API Key: ${options.apiKey.substring(0, 10)}...${options.apiKey.substring(options.apiKey.length - 4)}`);
  console.log(`Advertiser ID: ${options.advertiserId}`);
  console.log(`Timestamp: ${timestamp}`);
  console.log(`Nonce: ${nonce}`);
  console.log(`\nData to Sign: ${dataToSign}`);
  console.log(`\nSignature: ${signature}`);
  
  console.log('\n=== JSON Payload ===\n');
  console.log(JSON.stringify({
    api_key: options.apiKey,
    advertiser_id: options.advertiserId,
    timestamp,
    nonce,
    signature,
  }, null, 2));

  console.log('\n=== cURL Example ===\n');
  console.log(`curl -X POST https://api.afftok.com/api/sdk/click \\
  -H "Content-Type: application/json" \\
  -H "X-API-Key: ${options.apiKey}" \\
  -d '{
    "api_key": "${options.apiKey}",
    "advertiser_id": "${options.advertiserId}",
    "offer_id": "your_offer_id",
    "timestamp": ${timestamp},
    "nonce": "${nonce}",
    "signature": "${signature}"
  }'`);
}

// Validate mode
function runValidate(options) {
  if (!options.apiKey || !options.advertiserId) {
    console.error('Error: API key and advertiser ID are required');
    process.exit(1);
  }

  if (!options.timestamp || !options.nonce || !options.signature) {
    console.error('Error: Timestamp, nonce, and signature are required for validation');
    process.exit(1);
  }

  const result = validateSignature(
    options.apiKey,
    options.advertiserId,
    options.timestamp,
    options.nonce,
    options.signature
  );

  console.log('=== Signature Validation ===\n');
  console.log(`API Key: ${options.apiKey.substring(0, 10)}...`);
  console.log(`Advertiser ID: ${options.advertiserId}`);
  console.log(`Timestamp: ${options.timestamp}`);
  console.log(`Nonce: ${options.nonce}`);
  console.log(`\nProvided Signature: ${options.signature}`);
  console.log(`Expected Signature: ${result.expectedSignature}`);
  console.log(`\nResult: ${result.isValid ? '✓ VALID' : '✗ INVALID'}`);

  if (!result.isValid) {
    console.log('\n=== Debugging ===\n');
    console.log('Common issues:');
    console.log('1. Incorrect API key used for signing');
    console.log('2. Wrong order of parameters in data_to_sign');
    console.log('3. Timestamp mismatch (should be in milliseconds)');
    console.log('4. Nonce contains invalid characters');
    console.log('\nExpected format: {api_key}|{advertiser_id}|{timestamp}|{nonce}');
    
    const dataToSign = `${options.apiKey}|${options.advertiserId}|${options.timestamp}|${options.nonce}`;
    console.log(`\nData to sign: ${dataToSign}`);
  }

  process.exit(result.isValid ? 0 : 1);
}

// Interactive mode
function runInteractive() {
  const readline = require('readline');
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  console.log('=== AffTok Signature Tool (Interactive) ===\n');

  const askQuestion = (question) => {
    return new Promise(resolve => rl.question(question, resolve));
  };

  (async () => {
    const mode = await askQuestion('Mode (generate/validate): ');
    const apiKey = await askQuestion('API Key: ');
    const advertiserId = await askQuestion('Advertiser ID: ');

    if (mode === 'generate') {
      const useCurrentTime = await askQuestion('Use current timestamp? (y/n): ');
      const timestamp = useCurrentTime.toLowerCase() === 'y' 
        ? Date.now() 
        : parseInt(await askQuestion('Timestamp: '), 10);
      
      const generateNewNonce = await askQuestion('Generate new nonce? (y/n): ');
      const nonce = generateNewNonce.toLowerCase() === 'y'
        ? generateNonce()
        : await askQuestion('Nonce: ');

      runGenerate({ apiKey, advertiserId, timestamp, nonce });
    } else if (mode === 'validate') {
      const timestamp = parseInt(await askQuestion('Timestamp: '), 10);
      const nonce = await askQuestion('Nonce: ');
      const signature = await askQuestion('Signature: ');

      runValidate({ apiKey, advertiserId, timestamp, nonce, signature });
    } else {
      console.log('Invalid mode. Use "generate" or "validate".');
    }

    rl.close();
  })();
}

// Main function
function main() {
  const options = parseArgs();

  if (options.help) {
    printHelp();
    return;
  }

  if (!options.mode) {
    // Interactive mode if no mode specified
    if (process.stdin.isTTY) {
      runInteractive();
    } else {
      printHelp();
    }
    return;
  }

  if (options.mode === 'generate') {
    runGenerate(options);
  } else if (options.mode === 'validate') {
    runValidate(options);
  } else {
    console.error('Invalid mode. Use --generate or --validate');
    process.exit(1);
  }
}

main();

