package com.afftok

import android.content.Context
import android.os.Build
import android.provider.Settings
import java.security.MessageDigest
import java.util.UUID

/**
 * Device fingerprint provider for AffTok SDK
 */
class FingerprintProvider(private val context: Context) {
    
    private var cachedFingerprint: String? = null
    private var cachedDeviceId: String? = null
    
    /**
     * Get or generate device ID
     */
    fun getDeviceId(): String {
        if (cachedDeviceId != null) return cachedDeviceId!!
        
        val prefs = context.getSharedPreferences(Config.PREFS_NAME, Context.MODE_PRIVATE)
        var deviceId = prefs.getString(Config.KEY_DEVICE_ID, null)
        
        if (deviceId == null) {
            deviceId = generateDeviceId()
            prefs.edit().putString(Config.KEY_DEVICE_ID, deviceId).apply()
        }
        
        cachedDeviceId = deviceId
        return deviceId
    }
    
    /**
     * Generate unique device fingerprint
     */
    fun getFingerprint(): String {
        if (cachedFingerprint != null) return cachedFingerprint!!
        
        val components = listOf(
            getDeviceId(),
            Build.MANUFACTURER,
            Build.MODEL,
            Build.BRAND,
            Build.DEVICE,
            Build.PRODUCT,
            Build.BOARD,
            Build.HARDWARE,
            getAndroidId(),
            getScreenResolution(),
            getLanguage(),
            getTimezone()
        )
        
        val combined = components.joinToString("|")
        cachedFingerprint = sha256(combined)
        return cachedFingerprint!!
    }
    
    /**
     * Get device info map
     */
    fun getDeviceInfo(): Map<String, String> {
        return mapOf(
            "device_id" to getDeviceId(),
            "fingerprint" to getFingerprint(),
            "platform" to Config.SDK_PLATFORM,
            "sdk_version" to Config.SDK_VERSION,
            "os_version" to Build.VERSION.RELEASE,
            "api_level" to Build.VERSION.SDK_INT.toString(),
            "manufacturer" to Build.MANUFACTURER,
            "model" to Build.MODEL,
            "brand" to Build.BRAND,
            "screen" to getScreenResolution(),
            "language" to getLanguage(),
            "timezone" to getTimezone()
        )
    }
    
    private fun generateDeviceId(): String {
        val androidId = getAndroidId()
        return if (androidId.isNotEmpty() && androidId != "9774d56d682e549c") {
            sha256(androidId)
        } else {
            UUID.randomUUID().toString()
        }
    }
    
    private fun getAndroidId(): String {
        return try {
            Settings.Secure.getString(context.contentResolver, Settings.Secure.ANDROID_ID) ?: ""
        } catch (e: Exception) {
            ""
        }
    }
    
    private fun getScreenResolution(): String {
        return try {
            val metrics = context.resources.displayMetrics
            "${metrics.widthPixels}x${metrics.heightPixels}"
        } catch (e: Exception) {
            "unknown"
        }
    }
    
    private fun getLanguage(): String {
        return try {
            java.util.Locale.getDefault().language
        } catch (e: Exception) {
            "en"
        }
    }
    
    private fun getTimezone(): String {
        return try {
            java.util.TimeZone.getDefault().id
        } catch (e: Exception) {
            "UTC"
        }
    }
    
    private fun sha256(input: String): String {
        return try {
            val bytes = MessageDigest.getInstance("SHA-256").digest(input.toByteArray())
            bytes.joinToString("") { "%02x".format(it) }
        } catch (e: Exception) {
            input.hashCode().toString()
        }
    }
}

