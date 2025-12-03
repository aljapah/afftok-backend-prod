package com.afftok

import android.content.Context
import android.content.SharedPreferences
import kotlinx.coroutines.*
import org.json.JSONArray
import org.json.JSONObject
import java.util.UUID
import java.util.concurrent.ConcurrentLinkedQueue

/**
 * Offline queue for storing events when device is offline
 * Supports persistence and automatic retry with exponential backoff
 */
class OfflineQueue(
    private val context: Context,
    private val debug: Boolean = false
) {
    private val queue = ConcurrentLinkedQueue<QueueItem>()
    private val prefs: SharedPreferences = context.getSharedPreferences(Config.PREFS_NAME, Context.MODE_PRIVATE)
    private var flushJob: Job? = null
    private var isProcessing = false
    
    init {
        loadFromStorage()
    }
    
    /**
     * Add item to queue
     */
    fun enqueue(type: String, payload: Map<String, Any>): String {
        val id = UUID.randomUUID().toString()
        val item = QueueItem(
            id = id,
            type = type,
            payload = payload,
            timestamp = System.currentTimeMillis(),
            retryCount = 0,
            nextRetryTime = 0
        )
        
        if (queue.size >= Config.MAX_QUEUE_SIZE) {
            // Remove oldest item if queue is full
            queue.poll()
            log("Queue full, removed oldest item")
        }
        
        queue.add(item)
        saveToStorage()
        log("Enqueued $type event: $id")
        
        return id
    }
    
    /**
     * Get all pending items ready for processing
     */
    fun getPendingItems(): List<QueueItem> {
        val now = System.currentTimeMillis()
        return queue.filter { it.nextRetryTime <= now }
    }
    
    /**
     * Mark item as completed and remove from queue
     */
    fun markCompleted(id: String) {
        queue.removeIf { it.id == id }
        saveToStorage()
        log("Completed: $id")
    }
    
    /**
     * Mark item for retry with exponential backoff
     */
    fun markForRetry(id: String) {
        val item = queue.find { it.id == id } ?: return
        
        if (item.retryCount >= Config.MAX_RETRY_ATTEMPTS) {
            // Max retries reached, remove item
            queue.removeIf { it.id == id }
            log("Max retries reached, removing: $id")
            saveToStorage()
            return
        }
        
        // Calculate next retry time with exponential backoff + jitter
        val delay = minOf(
            Config.INITIAL_RETRY_DELAY_MS * (1 shl item.retryCount),
            Config.MAX_RETRY_DELAY_MS
        )
        val jitter = (Math.random() * delay * 0.1).toLong()
        
        val updatedItem = item.copy(
            retryCount = item.retryCount + 1,
            nextRetryTime = System.currentTimeMillis() + delay + jitter
        )
        
        queue.removeIf { it.id == id }
        queue.add(updatedItem)
        saveToStorage()
        
        log("Marked for retry (${ updatedItem.retryCount }/${Config.MAX_RETRY_ATTEMPTS}): $id")
    }
    
    /**
     * Get queue size
     */
    fun size(): Int = queue.size
    
    /**
     * Check if queue is empty
     */
    fun isEmpty(): Boolean = queue.isEmpty()
    
    /**
     * Clear all items
     */
    fun clear() {
        queue.clear()
        saveToStorage()
        log("Queue cleared")
    }
    
    /**
     * Start automatic flush job
     */
    fun startAutoFlush(interval: Long, processor: suspend (QueueItem) -> Boolean) {
        stopAutoFlush()
        
        flushJob = CoroutineScope(Dispatchers.IO).launch {
            while (isActive) {
                delay(interval)
                flush(processor)
            }
        }
        log("Auto-flush started with interval: ${interval}ms")
    }
    
    /**
     * Stop automatic flush job
     */
    fun stopAutoFlush() {
        flushJob?.cancel()
        flushJob = null
        log("Auto-flush stopped")
    }
    
    /**
     * Manually flush queue
     */
    suspend fun flush(processor: suspend (QueueItem) -> Boolean) {
        if (isProcessing) {
            log("Flush already in progress, skipping")
            return
        }
        
        isProcessing = true
        log("Starting flush, ${queue.size} items in queue")
        
        try {
            val pendingItems = getPendingItems()
            
            for (item in pendingItems) {
                try {
                    val success = processor(item)
                    if (success) {
                        markCompleted(item.id)
                    } else {
                        markForRetry(item.id)
                    }
                } catch (e: Exception) {
                    log("Error processing item ${item.id}: ${e.message}")
                    markForRetry(item.id)
                }
            }
        } finally {
            isProcessing = false
            log("Flush completed, ${queue.size} items remaining")
        }
    }
    
    /**
     * Save queue to SharedPreferences
     */
    private fun saveToStorage() {
        try {
            val jsonArray = JSONArray()
            queue.forEach { item ->
                val json = JSONObject().apply {
                    put("id", item.id)
                    put("type", item.type)
                    put("payload", JSONObject(item.payload))
                    put("timestamp", item.timestamp)
                    put("retryCount", item.retryCount)
                    put("nextRetryTime", item.nextRetryTime)
                }
                jsonArray.put(json)
            }
            prefs.edit().putString(Config.KEY_QUEUE, jsonArray.toString()).apply()
        } catch (e: Exception) {
            log("Error saving queue: ${e.message}")
        }
    }
    
    /**
     * Load queue from SharedPreferences
     */
    private fun loadFromStorage() {
        try {
            val jsonString = prefs.getString(Config.KEY_QUEUE, null) ?: return
            val jsonArray = JSONArray(jsonString)
            
            for (i in 0 until jsonArray.length()) {
                val json = jsonArray.getJSONObject(i)
                val payloadJson = json.getJSONObject("payload")
                val payload = mutableMapOf<String, Any>()
                payloadJson.keys().forEach { key ->
                    payload[key] = payloadJson.get(key)
                }
                
                val item = QueueItem(
                    id = json.getString("id"),
                    type = json.getString("type"),
                    payload = payload,
                    timestamp = json.getLong("timestamp"),
                    retryCount = json.optInt("retryCount", 0),
                    nextRetryTime = json.optLong("nextRetryTime", 0)
                )
                queue.add(item)
            }
            log("Loaded ${queue.size} items from storage")
        } catch (e: Exception) {
            log("Error loading queue: ${e.message}")
        }
    }
    
    private fun log(message: String) {
        if (debug) {
            println("[AffTok Queue] $message")
        }
    }
}

