# Android SDK

Official AffTok SDK for Android applications.

## Requirements

- Android API 21+ (Android 5.0 Lollipop)
- Kotlin 1.6+ or Java 8+

## Installation

### Gradle (Kotlin DSL)

```kotlin
// build.gradle.kts
dependencies {
    implementation("com.afftok:sdk:1.0.0")
}
```

### Gradle (Groovy)

```groovy
// build.gradle
dependencies {
    implementation 'com.afftok:sdk:1.0.0'
}
```

### Maven

```xml
<dependency>
    <groupId>com.afftok</groupId>
    <artifactId>sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

## Permissions

Add to `AndroidManifest.xml`:

```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

## Initialization

Initialize the SDK in your `Application` class:

```kotlin
import com.afftok.Afftok
import com.afftok.AfftokConfig

class MyApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        
        val config = AfftokConfig.Builder()
            .apiKey("afftok_live_sk_your_api_key_here")
            .advertiserId("adv_123456")
            .userId("user_789")  // Optional: your internal user ID
            .debugMode(BuildConfig.DEBUG)
            .offlineQueueEnabled(true)
            .maxOfflineEvents(1000)
            .flushIntervalSeconds(30)
            .build()
        
        Afftok.init(this, config)
    }
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | String | Required | Your API key |
| `advertiserId` | String | Required | Your advertiser ID |
| `userId` | String | null | Your internal user ID |
| `debugMode` | Boolean | false | Enable verbose logging |
| `offlineQueueEnabled` | Boolean | true | Enable offline queue |
| `maxOfflineEvents` | Int | 1000 | Max queued events |
| `flushIntervalSeconds` | Int | 30 | Queue flush interval |
| `retryIntervalSeconds` | Int | 5 | Initial retry delay |
| `baseUrl` | String | Production URL | Custom API endpoint |

## Track Click

Track when a user clicks on an offer:

```kotlin
import com.afftok.Afftok
import com.afftok.ClickEvent

// Basic click tracking
Afftok.trackClick(
    offerId = "off_123456",
    trackingCode = "abc123.1701432000.xyz.sig"
)

// With metadata
val metadata = mapOf(
    "source" to "email_campaign",
    "creative" to "banner_a",
    "placement" to "header"
)

Afftok.trackClick(
    offerId = "off_123456",
    trackingCode = "abc123.1701432000.xyz.sig",
    metadata = metadata,
    callback = object : Afftok.Callback {
        override fun onSuccess(result: Map<String, Any>) {
            Log.d("Afftok", "Click tracked: ${result["click_id"]}")
        }
        
        override fun onError(error: AfftokError) {
            Log.e("Afftok", "Error: ${error.code} - ${error.message}")
        }
    }
)
```

### Click Event Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `offerId` | String | Yes | Offer identifier |
| `trackingCode` | String | No | Signed tracking code |
| `subId` | String | No | Sub-affiliate ID |
| `metadata` | Map | No | Custom key-value pairs |
| `callback` | Callback | No | Success/error callback |

## Track Conversion

Track when a user completes a conversion:

```kotlin
import com.afftok.Afftok
import com.afftok.ConversionEvent

// Basic conversion
Afftok.trackConversion(
    offerId = "off_123456",
    amount = 49.99,
    currency = "USD"
)

// Full conversion with all options
Afftok.trackConversion(
    offerId = "off_123456",
    amount = 49.99,
    currency = "USD",
    orderId = "order_12345",
    productId = "prod_premium",
    metadata = mapOf(
        "plan" to "annual",
        "coupon" to "SAVE20"
    ),
    callback = object : Afftok.Callback {
        override fun onSuccess(result: Map<String, Any>) {
            Log.d("Afftok", "Conversion tracked: ${result["conversion_id"]}")
        }
        
        override fun onError(error: AfftokError) {
            Log.e("Afftok", "Error: ${error.code} - ${error.message}")
        }
    }
)
```

### Conversion Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `offerId` | String | Yes | Offer identifier |
| `amount` | Double | Yes | Conversion value |
| `currency` | String | No | ISO currency code (default: USD) |
| `orderId` | String | No | Your order/transaction ID |
| `productId` | String | No | Product identifier |
| `metadata` | Map | No | Custom key-value pairs |
| `callback` | Callback | No | Success/error callback |

## Deep Link Handling

Track clicks from deep links:

```kotlin
class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        handleDeepLink(intent)
    }
    
    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        handleDeepLink(intent)
    }
    
    private fun handleDeepLink(intent: Intent) {
        intent.data?.let { uri ->
            // Extract tracking code from deep link
            val trackingCode = uri.getQueryParameter("ref")
            val offerId = uri.getQueryParameter("offer")
            
            if (trackingCode != null && offerId != null) {
                Afftok.trackClick(
                    offerId = offerId,
                    trackingCode = trackingCode,
                    metadata = mapOf("source" to "deep_link")
                )
            }
        }
    }
}
```

## Coroutines Support

The SDK provides suspend functions for Kotlin coroutines:

```kotlin
import com.afftok.Afftok
import kotlinx.coroutines.launch

// In a CoroutineScope
lifecycleScope.launch {
    try {
        val result = Afftok.trackClickAsync(
            offerId = "off_123456",
            trackingCode = "abc123.1701432000.xyz.sig"
        )
        Log.d("Afftok", "Click tracked: ${result.clickId}")
    } catch (e: AfftokException) {
        Log.e("Afftok", "Error: ${e.code} - ${e.message}")
    }
}

// Track conversion with coroutines
lifecycleScope.launch {
    try {
        val result = Afftok.trackConversionAsync(
            offerId = "off_123456",
            amount = 49.99,
            currency = "USD",
            orderId = "order_12345"
        )
        Log.d("Afftok", "Conversion ID: ${result.conversionId}")
    } catch (e: AfftokException) {
        Log.e("Afftok", "Error: ${e.code}")
    }
}
```

## Offline Queue

The SDK automatically queues events when offline:

```kotlin
// Check queue status
val queueSize = Afftok.getQueueSize()
Log.d("Afftok", "Pending events: $queueSize")

// Force flush queue (when online)
Afftok.flushQueue(object : Afftok.FlushCallback {
    override fun onComplete(sent: Int, failed: Int) {
        Log.d("Afftok", "Flushed: $sent sent, $failed failed")
    }
})

// Clear queue (discard all pending events)
Afftok.clearQueue()
```

## Device Fingerprinting

The SDK automatically generates a device fingerprint:

```kotlin
// Get current fingerprint
val fingerprint = Afftok.getDeviceFingerprint()
Log.d("Afftok", "Fingerprint: $fingerprint")

// Fingerprint is automatically included in all events
```

Fingerprint components:
- Android ID
- Device model and manufacturer
- Screen resolution
- Timezone
- Language
- App version

## User Identification

Set or update user ID after initialization:

```kotlin
// Set user ID
Afftok.setUserId("user_new_id")

// Get current user ID
val userId = Afftok.getUserId()

// Clear user ID (on logout)
Afftok.clearUserId()
```

## Error Handling

```kotlin
Afftok.trackClick(
    offerId = "off_123456",
    callback = object : Afftok.Callback {
        override fun onSuccess(result: Map<String, Any>) {
            // Handle success
        }
        
        override fun onError(error: AfftokError) {
            when (error.code) {
                "NETWORK_ERROR" -> {
                    // Event queued for retry
                    Log.w("Afftok", "Offline, event queued")
                }
                "INVALID_API_KEY" -> {
                    // Check API key configuration
                    Log.e("Afftok", "Invalid API key")
                }
                "RATE_LIMITED" -> {
                    // Too many requests
                    Log.w("Afftok", "Rate limited, retry later")
                }
                "VALIDATION_ERROR" -> {
                    // Invalid parameters
                    Log.e("Afftok", "Validation: ${error.message}")
                }
                else -> {
                    Log.e("Afftok", "Error: ${error.message}")
                }
            }
        }
    }
)
```

## ProGuard Rules

If using ProGuard/R8, add these rules:

```proguard
# AffTok SDK
-keep class com.afftok.** { *; }
-keepclassmembers class com.afftok.** { *; }

# Keep callback interfaces
-keep interface com.afftok.Afftok$Callback { *; }
-keep interface com.afftok.Afftok$FlushCallback { *; }
```

## Testing

### Test Mode

Use test API keys during development:

```kotlin
val config = AfftokConfig.Builder()
    .apiKey("afftok_test_sk_...")  // Test key
    .advertiserId("adv_test_123")
    .debugMode(true)
    .build()

Afftok.init(this, config)
```

### Mock Events

Generate test events:

```kotlin
// Only available in debug mode
if (BuildConfig.DEBUG) {
    Afftok.generateTestClick()
    Afftok.generateTestConversion(amount = 99.99)
}
```

## Complete Example

```kotlin
class MyApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        
        Afftok.init(this, AfftokConfig.Builder()
            .apiKey("afftok_live_sk_your_key")
            .advertiserId("adv_123456")
            .debugMode(BuildConfig.DEBUG)
            .build()
        )
    }
}

class OfferActivity : AppCompatActivity() {
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_offer)
        
        // Track offer view click
        val offerId = intent.getStringExtra("offer_id") ?: return
        val trackingCode = intent.getStringExtra("tracking_code")
        
        if (trackingCode != null) {
            Afftok.trackClick(
                offerId = offerId,
                trackingCode = trackingCode,
                metadata = mapOf("screen" to "offer_detail")
            )
        }
        
        // Purchase button
        findViewById<Button>(R.id.btnPurchase).setOnClickListener {
            processPurchase(offerId)
        }
    }
    
    private fun processPurchase(offerId: String) {
        // Your purchase logic...
        val orderId = "order_${System.currentTimeMillis()}"
        val amount = 49.99
        
        // Track conversion
        Afftok.trackConversion(
            offerId = offerId,
            amount = amount,
            currency = "USD",
            orderId = orderId,
            callback = object : Afftok.Callback {
                override fun onSuccess(result: Map<String, Any>) {
                    Toast.makeText(
                        this@OfferActivity,
                        "Purchase complete!",
                        Toast.LENGTH_SHORT
                    ).show()
                }
                
                override fun onError(error: AfftokError) {
                    // Conversion will retry automatically
                    Log.w("Afftok", "Conversion tracking error: ${error.message}")
                }
            }
        )
    }
}
```

## Changelog

### 1.0.0
- Initial release
- Click and conversion tracking
- Offline queue support
- Device fingerprinting
- HMAC request signing

---

Next: [iOS SDK](./ios.md)

