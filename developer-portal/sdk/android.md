# Android SDK

Complete guide for integrating AffTok tracking into Android applications.

---

## Requirements

- Android SDK 21+ (Android 5.0 Lollipop)
- Kotlin 1.6+ or Java 8+
- Gradle 7.0+

---

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

---

## Initialization

Initialize the SDK in your `Application` class:

```kotlin
import com.afftok.Afftok
import com.afftok.AfftokConfig

class MyApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        
        Afftok.init(
            context = this,
            apiKey = "afftok_live_sk_xxxxx",
            config = AfftokConfig(
                debug = BuildConfig.DEBUG,
                enableOfflineQueue = true,
                maxQueueSize = 1000,
                flushIntervalMs = 30000,
                timeoutMs = 30000,
                retryAttempts = 3
            )
        )
    }
}
```

### Configuration Options

```kotlin
data class AfftokConfig(
    val debug: Boolean = false,
    val enableOfflineQueue: Boolean = true,
    val maxQueueSize: Int = 1000,
    val flushIntervalMs: Long = 30000,
    val timeoutMs: Long = 30000,
    val retryAttempts: Int = 3,
    val baseUrl: String = "https://api.afftok.com",
    val enableDeviceFingerprint: Boolean = true,
    val enableInstallReferrer: Boolean = true
)
```

---

## Click Tracking

### Basic Click Tracking

```kotlin
Afftok.trackClick(
    trackingCode = "tc_abc123def456"
)
```

### Click with Metadata

```kotlin
Afftok.trackClick(
    trackingCode = "tc_abc123def456",
    metadata = mapOf(
        "campaign" to "summer_sale",
        "source" to "push_notification",
        "sub_id" to "user_123"
    )
)
```

### Click with Callback

```kotlin
Afftok.trackClick(
    trackingCode = "tc_abc123def456",
    metadata = mapOf("campaign" to "summer_sale")
) { result ->
    when (result) {
        is AfftokResult.Success -> {
            Log.d("Afftok", "Click tracked: ${result.data.clickId}")
        }
        is AfftokResult.Error -> {
            Log.e("Afftok", "Error: ${result.error.message}")
        }
    }
}
```

### Coroutine Support

```kotlin
import kotlinx.coroutines.launch

lifecycleScope.launch {
    try {
        val result = Afftok.trackClickAsync(
            trackingCode = "tc_abc123def456",
            metadata = mapOf("campaign" to "summer_sale")
        )
        Log.d("Afftok", "Click tracked: ${result.clickId}")
    } catch (e: AfftokException) {
        Log.e("Afftok", "Error: ${e.message}")
    }
}
```

---

## Conversion Tracking

### Basic Conversion

```kotlin
Afftok.trackConversion(
    offerId = "off_xyz789",
    transactionId = "txn_${System.currentTimeMillis()}"
)
```

### Conversion with Amount

```kotlin
Afftok.trackConversion(
    offerId = "off_xyz789",
    transactionId = "order_12345",
    amount = 49.99,
    currency = "USD"
)
```

### Full Conversion Tracking

```kotlin
Afftok.trackConversion(
    offerId = "off_xyz789",
    transactionId = "order_12345",
    clickId = "clk_abc123",  // Optional: link to specific click
    amount = 49.99,
    currency = "USD",
    status = ConversionStatus.APPROVED,
    metadata = mapOf(
        "product_id" to "prod_456",
        "category" to "electronics",
        "user_type" to "new"
    )
) { result ->
    when (result) {
        is AfftokResult.Success -> {
            Log.d("Afftok", "Conversion tracked: ${result.data.conversionId}")
        }
        is AfftokResult.Error -> {
            Log.e("Afftok", "Error: ${result.error.message}")
        }
    }
}
```

### Conversion Status

```kotlin
enum class ConversionStatus {
    PENDING,
    APPROVED,
    REJECTED
}
```

---

## Device Fingerprinting

The SDK automatically generates a unique device fingerprint:

```kotlin
// Get current device fingerprint
val fingerprint = Afftok.getDeviceFingerprint()
Log.d("Afftok", "Device Fingerprint: $fingerprint")
```

### Custom Fingerprint Provider

```kotlin
Afftok.setFingerprintProvider(object : FingerprintProvider {
    override fun getFingerprint(): String {
        // Custom fingerprint logic
        return "custom_fingerprint_${Build.FINGERPRINT}"
    }
})
```

---

## Offline Queue

Events are automatically queued when offline:

```kotlin
// Check queue status
val queueSize = Afftok.getQueueSize()
Log.d("Afftok", "Queued events: $queueSize")

// Force flush queue
Afftok.flushQueue { result ->
    when (result) {
        is AfftokResult.Success -> {
            Log.d("Afftok", "Queue flushed: ${result.data.flushedCount} events")
        }
        is AfftokResult.Error -> {
            Log.e("Afftok", "Flush failed: ${result.error.message}")
        }
    }
}

// Clear queue
Afftok.clearQueue()
```

---

## Deep Link Handling

### Handle Deep Links

```kotlin
class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Handle deep link
        intent?.data?.let { uri ->
            Afftok.handleDeepLink(uri) { result ->
                when (result) {
                    is AfftokResult.Success -> {
                        Log.d("Afftok", "Deep link tracked: ${result.data}")
                    }
                    is AfftokResult.Error -> {
                        Log.e("Afftok", "Error: ${result.error.message}")
                    }
                }
            }
        }
    }
    
    override fun onNewIntent(intent: Intent?) {
        super.onNewIntent(intent)
        intent?.data?.let { uri ->
            Afftok.handleDeepLink(uri)
        }
    }
}
```

### Generate Deep Link

```kotlin
val deepLink = Afftok.generateDeepLink(
    trackingCode = "tc_abc123",
    scheme = "myapp",
    path = "/offer/details"
)
// Result: myapp://offer/details?tc=tc_abc123
```

---

## Install Referrer

Automatically capture Google Play Install Referrer:

```kotlin
// Enable in config
AfftokConfig(
    enableInstallReferrer = true
)

// Or manually capture
Afftok.captureInstallReferrer { result ->
    when (result) {
        is AfftokResult.Success -> {
            Log.d("Afftok", "Install referrer: ${result.data.referrer}")
        }
        is AfftokResult.Error -> {
            Log.e("Afftok", "Error: ${result.error.message}")
        }
    }
}
```

---

## Push Notification Attribution

Track conversions from push notifications:

```kotlin
class MyFirebaseMessagingService : FirebaseMessagingService() {
    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        // Extract tracking code from push data
        val trackingCode = remoteMessage.data["afftok_tc"]
        
        if (trackingCode != null) {
            Afftok.trackClick(
                trackingCode = trackingCode,
                metadata = mapOf(
                    "source" to "push_notification",
                    "campaign" to remoteMessage.data["campaign"] ?: ""
                )
            )
        }
    }
}
```

---

## User Identification

Associate events with a user:

```kotlin
// Set user ID
Afftok.setUserId("user_12345")

// Set user properties
Afftok.setUserProperties(mapOf(
    "email" to "user@example.com",
    "plan" to "premium",
    "signup_date" to "2024-01-15"
))

// Clear user (on logout)
Afftok.clearUser()
```

---

## Error Handling

```kotlin
sealed class AfftokException : Exception() {
    data class NetworkError(override val message: String) : AfftokException()
    data class AuthError(override val message: String) : AfftokException()
    data class ValidationError(override val message: String) : AfftokException()
    data class RateLimitError(override val message: String, val retryAfter: Long) : AfftokException()
    data class ServerError(override val message: String, val code: Int) : AfftokException()
}

// Handle errors
Afftok.trackClick("tc_abc123") { result ->
    when (result) {
        is AfftokResult.Error -> {
            when (val error = result.error) {
                is AfftokException.NetworkError -> {
                    // Event queued for retry
                    Log.w("Afftok", "Network error, event queued")
                }
                is AfftokException.RateLimitError -> {
                    Log.w("Afftok", "Rate limited, retry after ${error.retryAfter}ms")
                }
                is AfftokException.AuthError -> {
                    Log.e("Afftok", "Invalid API key")
                }
                else -> {
                    Log.e("Afftok", "Error: ${error.message}")
                }
            }
        }
        else -> {}
    }
}
```

---

## ProGuard Rules

Add to your `proguard-rules.pro`:

```proguard
-keep class com.afftok.** { *; }
-keepattributes Signature
-keepattributes *Annotation*
```

---

## Permissions

Required permissions (automatically added by SDK):

```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

Optional permissions:

```xml
<!-- For install referrer -->
<uses-permission android:name="com.google.android.finsky.permission.BIND_GET_INSTALL_REFERRER_SERVICE" />
```

---

## Debug Mode

Enable detailed logging:

```kotlin
AfftokConfig(
    debug = true
)

// Or at runtime
Afftok.setDebugMode(true)
```

Debug logs include:
- All API requests/responses
- Queue operations
- Fingerprint generation
- Error details

---

## Complete Example

```kotlin
class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        
        // Handle deep link
        intent?.data?.let { uri ->
            Afftok.handleDeepLink(uri)
        }
        
        // Track screen view
        Afftok.trackEvent("screen_view", mapOf("screen" to "main"))
        
        // Setup purchase button
        findViewById<Button>(R.id.purchaseButton).setOnClickListener {
            completePurchase()
        }
    }
    
    private fun completePurchase() {
        // Your purchase logic...
        
        // Track conversion
        Afftok.trackConversion(
            offerId = "off_premium_subscription",
            transactionId = "txn_${System.currentTimeMillis()}",
            amount = 9.99,
            currency = "USD",
            metadata = mapOf(
                "product" to "premium_monthly",
                "user_id" to getCurrentUserId()
            )
        ) { result ->
            when (result) {
                is AfftokResult.Success -> {
                    Toast.makeText(this, "Purchase tracked!", Toast.LENGTH_SHORT).show()
                }
                is AfftokResult.Error -> {
                    Log.e("Afftok", "Tracking failed: ${result.error.message}")
                }
            }
        }
    }
}
```

---

## Next Steps

- [iOS SDK](ios.md) - iOS integration guide
- [API Reference](../api-reference/clicks.md) - Full API documentation
- [Testing Tools](../testing/tools.md) - Test your integration

