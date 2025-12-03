/**
 * AffTok Server-to-Server Integration - Node.js Example
 * 
 * This example shows how to send postbacks/conversions from your server
 * to AffTok using the Server-to-Server API.
 */

const axios = require('axios');
const crypto = require('crypto');

// Configuration
const config = {
  apiKey: process.env.AFFTOK_API_KEY || 'your_api_key',
  advertiserId: process.env.AFFTOK_ADVERTISER_ID || 'your_advertiser_id',
  baseUrl: process.env.AFFTOK_BASE_URL || 'https://api.afftok.com',
};

/**
 * Generate HMAC-SHA256 signature
 */
function generateSignature(apiKey, advertiserId, timestamp, nonce) {
  const dataToSign = `${apiKey}|${advertiserId}|${timestamp}|${nonce}`;
  return crypto.createHmac('sha256', apiKey).update(dataToSign).digest('hex');
}

/**
 * Generate random nonce
 */
function generateNonce(length = 32) {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

/**
 * Send a postback/conversion to AffTok
 * 
 * @param {Object} params - Conversion parameters
 * @param {string} params.offerId - Offer ID
 * @param {string} params.transactionId - Unique transaction ID
 * @param {string} [params.clickId] - Click ID for attribution
 * @param {number} [params.amount] - Conversion amount
 * @param {string} [params.currency] - Currency code (default: USD)
 * @param {string} [params.status] - Status: pending, approved, rejected
 * @param {Object} [params.customParams] - Additional custom parameters
 */
async function sendPostback(params) {
  const timestamp = Date.now();
  const nonce = generateNonce();
  const signature = generateSignature(config.apiKey, config.advertiserId, timestamp, nonce);

  const payload = {
    api_key: config.apiKey,
    advertiser_id: config.advertiserId,
    offer_id: params.offerId,
    transaction_id: params.transactionId,
    status: params.status || 'approved',
    currency: params.currency || 'USD',
    timestamp,
    nonce,
    signature,
  };

  // Add optional fields
  if (params.clickId) payload.click_id = params.clickId;
  if (params.amount !== undefined) payload.amount = params.amount;
  if (params.customParams) payload.custom_params = params.customParams;

  try {
    const response = await axios.post(`${config.baseUrl}/api/postback`, payload, {
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': config.apiKey,
      },
      timeout: 30000,
    });

    console.log('Postback sent successfully:', response.data);
    return { success: true, data: response.data };
  } catch (error) {
    console.error('Postback failed:', error.response?.data || error.message);
    return { success: false, error: error.response?.data || error.message };
  }
}

/**
 * Track a click event (server-side)
 */
async function trackClick(params) {
  const timestamp = Date.now();
  const nonce = generateNonce();
  const signature = generateSignature(config.apiKey, config.advertiserId, timestamp, nonce);

  const payload = {
    api_key: config.apiKey,
    advertiser_id: config.advertiserId,
    offer_id: params.offerId,
    timestamp,
    nonce,
    signature,
  };

  // Add optional fields
  if (params.trackingCode) payload.tracking_code = params.trackingCode;
  if (params.subId1) payload.sub_id_1 = params.subId1;
  if (params.subId2) payload.sub_id_2 = params.subId2;
  if (params.subId3) payload.sub_id_3 = params.subId3;
  if (params.ip) payload.ip = params.ip;
  if (params.userAgent) payload.user_agent = params.userAgent;
  if (params.customParams) payload.custom_params = params.customParams;

  try {
    const response = await axios.post(`${config.baseUrl}/api/sdk/click`, payload, {
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': config.apiKey,
      },
      timeout: 30000,
    });

    console.log('Click tracked successfully:', response.data);
    return { success: true, data: response.data };
  } catch (error) {
    console.error('Click tracking failed:', error.response?.data || error.message);
    return { success: false, error: error.response?.data || error.message };
  }
}

/**
 * Batch send multiple conversions
 */
async function sendBatchPostbacks(conversions) {
  const results = [];
  
  for (const conversion of conversions) {
    const result = await sendPostback(conversion);
    results.push({
      transactionId: conversion.transactionId,
      ...result,
    });
    
    // Small delay to avoid rate limiting
    await new Promise(resolve => setTimeout(resolve, 100));
  }
  
  return results;
}

// Example usage
async function main() {
  console.log('AffTok Server-to-Server Integration Example\n');

  // Example 1: Send a simple conversion
  console.log('1. Sending a simple conversion...');
  await sendPostback({
    offerId: 'offer_123',
    transactionId: `txn_${Date.now()}`,
    amount: 29.99,
    status: 'approved',
  });

  // Example 2: Send a conversion with click attribution
  console.log('\n2. Sending a conversion with click attribution...');
  await sendPostback({
    offerId: 'offer_123',
    transactionId: `txn_${Date.now()}`,
    clickId: 'click_abc123',
    amount: 49.99,
    currency: 'EUR',
    status: 'approved',
    customParams: {
      product_id: 'prod_456',
      category: 'electronics',
    },
  });

  // Example 3: Track a server-side click
  console.log('\n3. Tracking a server-side click...');
  await trackClick({
    offerId: 'offer_123',
    trackingCode: 'campaign_summer_2024',
    subId1: 'source_google',
    ip: '192.168.1.1',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
  });

  // Example 4: Batch send conversions
  console.log('\n4. Batch sending conversions...');
  const batchResults = await sendBatchPostbacks([
    { offerId: 'offer_123', transactionId: `batch_1_${Date.now()}`, amount: 10.00 },
    { offerId: 'offer_123', transactionId: `batch_2_${Date.now()}`, amount: 20.00 },
    { offerId: 'offer_123', transactionId: `batch_3_${Date.now()}`, amount: 30.00 },
  ]);
  console.log('Batch results:', batchResults);
}

// Run if executed directly
if (require.main === module) {
  main().catch(console.error);
}

// Export for use as module
module.exports = {
  sendPostback,
  trackClick,
  sendBatchPostbacks,
  generateSignature,
  config,
};

