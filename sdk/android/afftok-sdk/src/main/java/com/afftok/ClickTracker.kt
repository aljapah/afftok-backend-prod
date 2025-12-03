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
 * Click tracking implementation for AffTok SDK
 */
class ClickTracker(
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
     * Track a click event
     */
    suspend fun trackClick(params: ClickParams): AfftokResponse {
        val payload = buildClickPayload(params)
        
        return withContext(Dispatchers.IO) {
            try {
                val response = sendRequest(Config.CLICK_ENDPOINT, payload)
                if (response.success) {
                    log("Click tracked successfully: ${params.offerId}")
                } else {
                    // Queue for retry
                    queue.enqueue("click", payload)
                    log("Click queued for retry: ${params.offerId}")
                }
                response
            } catch (e: Exception) {
                // Queue for offline retry
                queue.enqueue("click", payload)
                log("Click queued (offline): ${params.offerId}, error: ${e.message}")
                AfftokResponse(
                    success = false,
                    error = e.message,
                    message = "Click queued for offline retry"
                )
            }
        }
    }
    
    /**
     * Track click with signed link
     */
    suspend fun trackSignedClick(signedLink: String, params: ClickParams): AfftokResponse {
        val payload = buildClickPayload(params).toMutableMap()
        payload["signed_link"] = signedLink
        payload["link_validated"] = true
        
        return withContext(Dispatchers.IO) {
            try {
                val response = sendRequest(Config.CLICK_ENDPOINT, payload)
                if (!response.success) {
                    queue.enqueue("click", payload)
                }
                response
            } catch (e: Exception) {
                queue.enqueue("click", payload)
                AfftokResponse(
                    success = false,
                    error = e.message,
                    message = "Signed click queued for retry"
                )
            }
        }
    }
    
    /**
     * Build click payload
     */
    private fun buildClickPayload(params: ClickParams): Map<String, Any> {
        val timestamp = System.currentTimeMillis()
        val nonce = generateNonce()
        val deviceInfo = fingerprintProvider.getDeviceInfo()
        
        val payload = mutableMapOf<String, Any>(
            "api_key" to options.apiKey,
            "advertiser_id" to options.advertiserId,
            "offer_id" to params.offerId,
            "timestamp" to timestamp,
            "nonce" to nonce,
            "device_info" to deviceInfo
        )
        
        options.userId?.let { payload["user_id"] = it }
        params.trackingCode?.let { payload["tracking_code"] = it }
        params.subId1?.let { payload["sub_id_1"] = it }
        params.subId2?.let { payload["sub_id_2"] = it }
        params.subId3?.let { payload["sub_id_3"] = it }
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
            if (endpoint == Config.CLICK_ENDPOINT) {
                return sendFallbackRequest(Config.FALLBACK_CLICK_ENDPOINT, payload)
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
                AfftokResponse(success = true, message = "Click tracked via fallback")
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
            println("[AffTok Click] $message")
        }
    }
}

