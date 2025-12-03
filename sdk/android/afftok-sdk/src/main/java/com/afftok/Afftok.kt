package com.afftok

import android.content.Context
import kotlinx.coroutines.*

/**
 * AffTok SDK - Main Entry Point
 * 
 * The AffTok SDK provides click and conversion tracking for Android applications.
 * It supports offline queuing, automatic retry with exponential backoff,
 * HMAC-SHA256 signature verification, and zero-drop tracking.
 * 
 * Usage:
 * ```kotlin
 * // Initialize
 * Afftok.init(context, AfftokOptions(
 *     apiKey = "your_api_key",
 *     advertiserId = "your_advertiser_id",
 *     debug = true
 * ))
 * 
 * // Track click
 * Afftok.trackClick(ClickParams(offerId = "offer123"))
 * 
 * // Track conversion
 * Afftok.trackConversion(ConversionParams(
 *     offerId = "offer123",
 *     transactionId = "txn_abc123",
 *     amount = 29.99
 * ))
 * ```
 */
object Afftok {
    private var isInitialized = false
    private lateinit var options: AfftokOptions
    private lateinit var context: Context
    private lateinit var fingerprintProvider: FingerprintProvider
    private lateinit var queue: OfflineQueue
    private lateinit var clickTracker: ClickTracker
    private lateinit var conversionTracker: ConversionTracker
    
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    
    /**
     * Initialize the AffTok SDK
     * 
     * @param context Android application context
     * @param options SDK configuration options
     */
    fun init(context: Context, options: AfftokOptions) {
        if (isInitialized) {
            log("SDK already initialized")
            return
        }
        
        this.context = context.applicationContext
        this.options = options
        
        // Initialize components
        fingerprintProvider = FingerprintProvider(this.context)
        queue = OfflineQueue(this.context, options.debug)
        clickTracker = ClickTracker(options, fingerprintProvider, queue)
        conversionTracker = ConversionTracker(options, fingerprintProvider, queue)
        
        // Start auto-flush if enabled
        if (options.autoFlush) {
            startAutoFlush()
        }
        
        isInitialized = true
        log("SDK initialized successfully")
        log("Device ID: ${fingerprintProvider.getDeviceId()}")
        log("Pending queue items: ${queue.size()}")
    }
    
    /**
     * Track a click event
     * 
     * @param params Click parameters
     * @return AfftokResponse with tracking result
     */
    suspend fun trackClick(params: ClickParams): AfftokResponse {
        ensureInitialized()
        return clickTracker.trackClick(params)
    }
    
    /**
     * Track a click event (non-suspend version using callback)
     */
    fun trackClick(params: ClickParams, callback: (AfftokResponse) -> Unit) {
        ensureInitialized()
        scope.launch {
            val response = clickTracker.trackClick(params)
            withContext(Dispatchers.Main) {
                callback(response)
            }
        }
    }
    
    /**
     * Track a click with a signed tracking link
     * 
     * @param signedLink The signed tracking link from AffTok
     * @param params Additional click parameters
     * @return AfftokResponse with tracking result
     */
    suspend fun trackSignedClick(signedLink: String, params: ClickParams): AfftokResponse {
        ensureInitialized()
        return clickTracker.trackSignedClick(signedLink, params)
    }
    
    /**
     * Track a conversion event
     * 
     * @param params Conversion parameters
     * @return AfftokResponse with tracking result
     */
    suspend fun trackConversion(params: ConversionParams): AfftokResponse {
        ensureInitialized()
        return conversionTracker.trackConversion(params)
    }
    
    /**
     * Track a conversion event (non-suspend version using callback)
     */
    fun trackConversion(params: ConversionParams, callback: (AfftokResponse) -> Unit) {
        ensureInitialized()
        scope.launch {
            val response = conversionTracker.trackConversion(params)
            withContext(Dispatchers.Main) {
                callback(response)
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
        ensureInitialized()
        return conversionTracker.trackConversionWithMeta(params, metadata)
    }
    
    /**
     * Manually enqueue an event for later processing
     * 
     * @param type Event type ("click" or "conversion")
     * @param payload Event payload
     * @return Queue item ID
     */
    fun enqueue(type: String, payload: Map<String, Any>): String {
        ensureInitialized()
        return queue.enqueue(type, payload)
    }
    
    /**
     * Manually flush the offline queue
     */
    suspend fun flush() {
        ensureInitialized()
        queue.flush { item ->
            processQueueItem(item)
        }
    }
    
    /**
     * Flush queue with callback
     */
    fun flush(callback: () -> Unit) {
        ensureInitialized()
        scope.launch {
            flush()
            withContext(Dispatchers.Main) {
                callback()
            }
        }
    }
    
    /**
     * Get the device fingerprint
     */
    fun getFingerprint(): String {
        ensureInitialized()
        return fingerprintProvider.getFingerprint()
    }
    
    /**
     * Get the device ID
     */
    fun getDeviceId(): String {
        ensureInitialized()
        return fingerprintProvider.getDeviceId()
    }
    
    /**
     * Get device info
     */
    fun getDeviceInfo(): Map<String, String> {
        ensureInitialized()
        return fingerprintProvider.getDeviceInfo()
    }
    
    /**
     * Get number of pending queue items
     */
    fun getPendingCount(): Int {
        ensureInitialized()
        return queue.size()
    }
    
    /**
     * Check if SDK is initialized
     */
    fun isReady(): Boolean = isInitialized
    
    /**
     * Get SDK version
     */
    fun getVersion(): String = Config.SDK_VERSION
    
    /**
     * Clear all pending queue items
     */
    fun clearQueue() {
        ensureInitialized()
        queue.clear()
    }
    
    /**
     * Shutdown the SDK
     */
    fun shutdown() {
        if (!isInitialized) return
        
        queue.stopAutoFlush()
        scope.cancel()
        isInitialized = false
        log("SDK shutdown")
    }
    
    // Private methods
    
    private fun startAutoFlush() {
        queue.startAutoFlush(options.flushInterval) { item ->
            processQueueItem(item)
        }
    }
    
    private suspend fun processQueueItem(item: QueueItem): Boolean {
        return try {
            val response = when (item.type) {
                "click" -> {
                    val params = ClickParams(
                        offerId = item.payload["offer_id"] as? String ?: "",
                        trackingCode = item.payload["tracking_code"] as? String,
                        subId1 = item.payload["sub_id_1"] as? String,
                        subId2 = item.payload["sub_id_2"] as? String,
                        subId3 = item.payload["sub_id_3"] as? String
                    )
                    clickTracker.trackClick(params)
                }
                "conversion" -> {
                    val params = ConversionParams(
                        offerId = item.payload["offer_id"] as? String ?: "",
                        clickId = item.payload["click_id"] as? String,
                        transactionId = item.payload["transaction_id"] as? String ?: "",
                        amount = item.payload["amount"] as? Double,
                        currency = item.payload["currency"] as? String ?: "USD",
                        status = item.payload["status"] as? String ?: "pending"
                    )
                    conversionTracker.trackConversion(params)
                }
                else -> AfftokResponse(success = false, error = "Unknown event type")
            }
            response.success
        } catch (e: Exception) {
            log("Error processing queue item: ${e.message}")
            false
        }
    }
    
    private fun ensureInitialized() {
        if (!isInitialized) {
            throw IllegalStateException("AffTok SDK not initialized. Call Afftok.init() first.")
        }
    }
    
    private fun log(message: String) {
        if (::options.isInitialized && options.debug) {
            println("[AffTok SDK] $message")
        }
    }
}

