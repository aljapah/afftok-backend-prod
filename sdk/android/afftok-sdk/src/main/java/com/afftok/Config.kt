package com.afftok

/**
 * AffTok SDK Configuration
 */
object Config {
    // API Endpoints
    const val DEFAULT_BASE_URL = "https://api.afftok.com"
    const val CLICK_ENDPOINT = "/api/sdk/click"
    const val CONVERSION_ENDPOINT = "/api/sdk/conversion"
    
    // Fallback Endpoints
    const val FALLBACK_CLICK_ENDPOINT = "/api/c"
    const val FALLBACK_CONVERSION_ENDPOINT = "/api/convert"
    
    // Queue Settings
    const val MAX_QUEUE_SIZE = 1000
    const val MAX_RETRY_ATTEMPTS = 5
    const val INITIAL_RETRY_DELAY_MS = 1000L
    const val MAX_RETRY_DELAY_MS = 300000L // 5 minutes
    const val FLUSH_INTERVAL_MS = 30000L // 30 seconds
    
    // Rate Limiting
    const val MAX_REQUESTS_PER_MINUTE = 60
    const val RATE_LIMIT_WINDOW_MS = 60000L
    
    // Timeouts
    const val CONNECTION_TIMEOUT_MS = 10000L
    const val READ_TIMEOUT_MS = 30000L
    const val WRITE_TIMEOUT_MS = 30000L
    
    // Storage Keys
    const val PREFS_NAME = "afftok_sdk_prefs"
    const val KEY_QUEUE = "offline_queue"
    const val KEY_DEVICE_ID = "device_id"
    const val KEY_LAST_FLUSH = "last_flush"
    
    // SDK Version
    const val SDK_VERSION = "1.0.0"
    const val SDK_PLATFORM = "android"
}

/**
 * SDK initialization options
 */
data class AfftokOptions(
    val apiKey: String,
    val advertiserId: String,
    val userId: String? = null,
    val baseUrl: String = Config.DEFAULT_BASE_URL,
    val debug: Boolean = false,
    val autoFlush: Boolean = true,
    val flushInterval: Long = Config.FLUSH_INTERVAL_MS
)

/**
 * Click tracking parameters
 */
data class ClickParams(
    val offerId: String,
    val trackingCode: String? = null,
    val subId1: String? = null,
    val subId2: String? = null,
    val subId3: String? = null,
    val customParams: Map<String, String>? = null
)

/**
 * Conversion tracking parameters
 */
data class ConversionParams(
    val offerId: String,
    val clickId: String? = null,
    val transactionId: String,
    val amount: Double? = null,
    val currency: String = "USD",
    val status: String = "pending",
    val customParams: Map<String, String>? = null
)

/**
 * SDK Response
 */
data class AfftokResponse(
    val success: Boolean,
    val message: String? = null,
    val data: Map<String, Any>? = null,
    val error: String? = null
)

/**
 * Queue Item
 */
data class QueueItem(
    val id: String,
    val type: String, // "click" or "conversion"
    val payload: Map<String, Any>,
    val timestamp: Long,
    val retryCount: Int = 0,
    val nextRetryTime: Long = 0
)

