package com.afftok

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONObject
import java.util.concurrent.TimeUnit
import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec

/**
 * Conversion tracking implementation for AffTok SDK
 */
class ConversionTracker(
    private val options: AfftokOptions,
    private val fingerprintProvider: FingerprintProvider,
    private val queue: OfflineQueue
) {
    private val client: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(Config.CONNECTION_TIMEOUT_MS, TimeUnit.MILLISECONDS)
        .readTimeout(Config.READ_TIMEOUT_MS, TimeUnit.MILLISECONDS)
        .writeTimeout(Config.WRITE_TIMEOUT_MS, TimeUnit.MILLISECONDS)
        .build()
    
    private val jsonMediaType = "application/json; charset=utf-8".toMediaType()
    
    /**
     * Track a conversion event
     */
    suspend fun trackConversion(params: ConversionParams): AfftokResponse {
        val payload = buildConversionPayload(params)
        
        return withContext(Dispatchers.IO) {
            try {
                val response = sendRequest(Config.CONVERSION_ENDPOINT, payload)
                if (response.success) {
                    log("Conversion tracked successfully: ${params.transactionId}")
                } else {
                    // Queue for retry
                    queue.enqueue("conversion", payload)
                    log("Conversion queued for retry: ${params.transactionId}")
                }
                response
            } catch (e: Exception) {
                // Queue for offline retry
                queue.enqueue("conversion", payload)
                log("Conversion queued (offline): ${params.transactionId}, error: ${e.message}")
                AfftokResponse(
                    success = false,
                    error = e.message,
                    message = "Conversion queued for offline retry"
                )
            }
        }
    }
    
    /**
     * Track conversion with additional metadata
     */
    suspend fun trackConversionWithMeta(
        params: ConversionParams,
        metadata: Map<String, Any>
    ): AfftokResponse {
        val payload = buildConversionPayload(params).toMutableMap()
        payload["metadata"] = metadata
        
        return withContext(Dispatchers.IO) {
            try {
                val response = sendRequest(Config.CONVERSION_ENDPOINT, payload)
                if (!response.success) {
                    queue.enqueue("conversion", payload)
                }
                response
            } catch (e: Exception) {
                queue.enqueue("conversion", payload)
                AfftokResponse(
                    success = false,
                    error = e.message,
                    message = "Conversion with metadata queued for retry"
                )
            }
        }
    }
    
    /**
     * Build conversion payload
     */
    private fun buildConversionPayload(params: ConversionParams): Map<String, Any> {
        val timestamp = System.currentTimeMillis()
        val nonce = generateNonce()
        val deviceInfo = fingerprintProvider.getDeviceInfo()
        
        val payload = mutableMapOf<String, Any>(
            "api_key" to options.apiKey,
            "advertiser_id" to options.advertiserId,
            "offer_id" to params.offerId,
            "transaction_id" to params.transactionId,
            "status" to params.status,
            "currency" to params.currency,
            "timestamp" to timestamp,
            "nonce" to nonce,
            "device_info" to deviceInfo
        )
        
        options.userId?.let { payload["user_id"] = it }
        params.clickId?.let { payload["click_id"] = it }
        params.amount?.let { payload["amount"] = it }
        params.customParams?.let { payload["custom_params"] = it }
        
        // Add signature
        payload["signature"] = generateSignature(payload, timestamp, nonce)
        
        return payload
    }
    
    /**
     * Send HTTP request
     */
    private fun sendRequest(endpoint: String, payload: Map<String, Any>): AfftokResponse {
        val url = "${options.baseUrl}$endpoint"
        val json = JSONObject(payload).toString()
        val body = json.toRequestBody(jsonMediaType)
        
        val request = Request.Builder()
            .url(url)
            .post(body)
            .addHeader("Content-Type", "application/json")
            .addHeader("X-API-Key", options.apiKey)
            .addHeader("X-SDK-Version", Config.SDK_VERSION)
            .addHeader("X-SDK-Platform", Config.SDK_PLATFORM)
            .build()
        
        val response = client.newCall(request).execute()
        val responseBody = response.body?.string()
        
        return if (response.isSuccessful) {
            val jsonResponse = responseBody?.let { JSONObject(it) }
            AfftokResponse(
                success = true,
                message = jsonResponse?.optString("message"),
                data = jsonResponse?.let { parseJsonToMap(it) }
            )
        } else {
            // Try fallback endpoint
            if (endpoint == Config.CONVERSION_ENDPOINT) {
                return sendFallbackRequest(Config.FALLBACK_CONVERSION_ENDPOINT, payload)
            }
            AfftokResponse(
                success = false,
                error = "HTTP ${response.code}: $responseBody"
            )
        }
    }
    
    /**
     * Send request to fallback endpoint
     */
    private fun sendFallbackRequest(endpoint: String, payload: Map<String, Any>): AfftokResponse {
        return try {
            val url = "${options.baseUrl}$endpoint"
            val json = JSONObject(payload).toString()
            val body = json.toRequestBody(jsonMediaType)
            
            val request = Request.Builder()
                .url(url)
                .post(body)
                .addHeader("Content-Type", "application/json")
                .addHeader("X-API-Key", options.apiKey)
                .build()
            
            val response = client.newCall(request).execute()
            
            if (response.isSuccessful) {
                AfftokResponse(success = true, message = "Conversion tracked via fallback")
            } else {
                AfftokResponse(success = false, error = "Fallback failed: ${response.code}")
            }
        } catch (e: Exception) {
            AfftokResponse(success = false, error = "Fallback error: ${e.message}")
        }
    }
    
    /**
     * Generate HMAC-SHA256 signature
     */
    private fun generateSignature(payload: Map<String, Any>, timestamp: Long, nonce: String): String {
        val dataToSign = "${options.apiKey}|${options.advertiserId}|$timestamp|$nonce"
        
        return try {
            val mac = Mac.getInstance("HmacSHA256")
            val secretKey = SecretKeySpec(options.apiKey.toByteArray(), "HmacSHA256")
            mac.init(secretKey)
            val hmacBytes = mac.doFinal(dataToSign.toByteArray())
            hmacBytes.joinToString("") { "%02x".format(it) }
        } catch (e: Exception) {
            log("Signature generation failed: ${e.message}")
            ""
        }
    }
    
    /**
     * Generate random nonce
     */
    private fun generateNonce(): String {
        val chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
        return (1..32).map { chars.random() }.joinToString("")
    }
    
    /**
     * Parse JSON to Map
     */
    private fun parseJsonToMap(json: JSONObject): Map<String, Any> {
        val map = mutableMapOf<String, Any>()
        json.keys().forEach { key ->
            map[key] = json.get(key)
        }
        return map
    }
    
    private fun log(message: String) {
        if (options.debug) {
            println("[AffTok Conversion] $message")
        }
    }
}

