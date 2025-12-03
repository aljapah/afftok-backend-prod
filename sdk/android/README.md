# AffTok Android SDK

Official Android SDK for AffTok affiliate tracking platform.

## Features

- 🚀 Click & Conversion Tracking
- 📴 Offline Queue with Automatic Retry
- 🔐 HMAC-SHA256 Signature Verification
- 🔄 Exponential Backoff Retry
- 📱 Device Fingerprinting
- 🎯 Zero-Drop Tracking Compatible

## Requirements

- Android API 21+ (Lollipop)
- Kotlin 1.5+

## Installation

### Gradle

Add to your `build.gradle` (app level):

```gradle
dependencies {
    implementation 'com.afftok:afftok-sdk:1.0.0'
}
```

### Maven

```xml
<dependency>
    <groupId>com.afftok</groupId>
    <artifactId>afftok-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

## Quick Start

### 1. Initialize the SDK

Initialize in your Application class or main Activity:

```kotlin
import com.afftok.Afftok
import com.afftok.AfftokOptions

class MyApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        
        Afftok.init(this, AfftokOptions(
            apiKey = "your_api_key",
            advertiserId = "your_advertiser_id",
            userId = "optional_user_id",  // Optional
            debug = BuildConfig.DEBUG     // Enable logging in debug builds
        ))
    }
}
```

### 2. Track Clicks

```kotlin
import com.afftok.Afftok
import com.afftok.ClickParams
import kotlinx.coroutines.launch

// Using coroutines
lifecycleScope.launch {
    val response = Afftok.trackClick(ClickParams(
        offerId = "offer_123",
        trackingCode = "abc123",  // Optional
        subId1 = "campaign_1",   // Optional
        subId2 = "creative_2",   // Optional
        customParams = mapOf(    // Optional
            "source" to "facebook",
            "campaign" to "summer_sale"
        )
    ))
    
    if (response.success) {
        Log.d("AffTok", "Click tracked!")
    }
}

// Using callback
Afftok.trackClick(ClickParams(offerId = "offer_123")) { response ->
    if (response.success) {
        Log.d("AffTok", "Click tracked!")
    }
}
```

### 3. Track Conversions

```kotlin
import com.afftok.Afftok
import com.afftok.ConversionParams

lifecycleScope.launch {
    val response = Afftok.trackConversion(ConversionParams(
        offerId = "offer_123",
        transactionId = "txn_abc123",
        amount = 29.99,
        currency = "USD",
        status = "approved",  // "pending", "approved", "rejected"
        clickId = "click_xyz" // Optional, for attribution
    ))
    
    if (response.success) {
        Log.d("AffTok", "Conversion tracked!")
    }
}
```

### 4. Track with Signed Links

```kotlin
// When using pre-signed tracking links from AffTok
Afftok.trackSignedClick(
    signedLink = "https://track.afftok.com/c/abc123.1234567890.xyz.signature",
    params = ClickParams(offerId = "offer_123")
)
```

## Offline Queue

The SDK automatically queues events when offline and retries with exponential backoff.

```kotlin
// Check pending items
val pendingCount = Afftok.getPendingCount()

// Manually flush queue
Afftok.flush {
    Log.d("AffTok", "Queue flushed!")
}

// Clear queue
Afftok.clearQueue()
```

## Device Information

```kotlin
// Get device fingerprint
val fingerprint = Afftok.getFingerprint()

// Get device ID
val deviceId = Afftok.getDeviceId()

// Get full device info
val deviceInfo = Afftok.getDeviceInfo()
```

## Configuration Options

```kotlin
AfftokOptions(
    apiKey = "your_api_key",           // Required
    advertiserId = "your_advertiser",   // Required
    userId = null,                      // Optional user identifier
    baseUrl = "https://api.afftok.com", // Custom API URL
    debug = false,                      // Enable debug logging
    autoFlush = true,                   // Auto-flush queue
    flushInterval = 30000L              // Flush interval in ms
)
```

## Error Handling

```kotlin
val response = Afftok.trackClick(params)

when {
    response.success -> {
        // Event tracked successfully
        val data = response.data
    }
    response.error != null -> {
        // Error occurred
        Log.e("AffTok", "Error: ${response.error}")
    }
    else -> {
        // Event queued for retry
        Log.w("AffTok", "Event queued: ${response.message}")
    }
}
```

## ProGuard Rules

If using ProGuard, add these rules:

```proguard
-keep class com.afftok.** { *; }
-keepclassmembers class com.afftok.** { *; }
```

## Permissions

Add to your `AndroidManifest.xml`:

```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

## Best Practices

1. **Initialize Early**: Initialize the SDK in your Application class
2. **Use Debug Mode**: Enable debug mode during development
3. **Handle Responses**: Always check response status
4. **Unique Transaction IDs**: Use unique IDs for conversions
5. **Shutdown Properly**: Call `Afftok.shutdown()` when app terminates

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- Issues: https://github.com/afftok/android-sdk/issues

## License

MIT License - see LICENSE file for details.

