# AffTok Mobile SDK Guide

Complete guide for integrating AffTok tracking into mobile applications.

## Overview

The AffTok Mobile SDK supports:
- **Android** (Kotlin/Java)
- **iOS** (Swift/Objective-C)
- **Flutter** (Dart)
- **React Native** (JavaScript)

All SDKs share the same core features:
- Click & Conversion Tracking
- Offline Queue with Automatic Retry
- HMAC-SHA256 Signature Verification
- Device Fingerprinting
- Zero-Drop Tracking

---

## Initialization

### Android (Kotlin)

```kotlin
import com.afftok.Afftok
import com.afftok.AfftokOptions

class MyApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        
        Afftok.init(this, AfftokOptions(
            apiKey = "your_api_key",
            advertiserId = "your_advertiser_id",
            userId = "optional_user_id",
            debug = BuildConfig.DEBUG
        ))
    }
}
```

### iOS (Swift)

```swift
import Afftok

@main
class AppDelegate: UIResponder, UIApplicationDelegate {
    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        
        Afftok.shared.initialize(options: AfftokOptions(
            apiKey: "your_api_key",
            advertiserId: "your_advertiser_id",
            userId: "optional_user_id",
            debug: true
        ))
        
        return true
    }
}
```

### Flutter

```dart
import 'package:afftok/afftok.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  await Afftok.instance.initialize(AfftokOptions(
    apiKey: 'your_api_key',
    advertiserId: 'your_advertiser_id',
    userId: 'optional_user_id',
    debug: true,
  ));
  
  runApp(MyApp());
}
```

### React Native

```javascript
import Afftok from '@afftok/react-native-sdk';

useEffect(() => {
  Afftok.initialize({
    apiKey: 'your_api_key',
    advertiserId: 'your_advertiser_id',
    userId: 'optional_user_id',
    debug: __DEV__,
  });
}, []);
```

---

## Click Tracking

### Basic Click

```kotlin
// Android
Afftok.trackClick(ClickParams(offerId = "offer_123"))

// iOS
await Afftok.shared.trackClick(params: ClickParams(offerId: "offer_123"))

// Flutter
await Afftok.instance.trackClick(ClickParams(offerId: 'offer_123'));

// React Native
await Afftok.trackClick({ offerId: 'offer_123' });
```

### Click with Parameters

```kotlin
// Android
Afftok.trackClick(ClickParams(
    offerId = "offer_123",
    trackingCode = "campaign_summer",
    subId1 = "facebook",
    subId2 = "banner_1",
    subId3 = "mobile",
    customParams = mapOf(
        "source" to "app",
        "version" to "2.0"
    )
))
```

### Signed Click

```kotlin
// Using pre-signed tracking links
Afftok.trackSignedClick(
    signedLink = "https://track.afftok.com/c/abc123.1234567890.xyz.signature",
    params = ClickParams(offerId = "offer_123")
)
```

---

## Conversion Tracking

### Basic Conversion

```kotlin
// Android
Afftok.trackConversion(ConversionParams(
    offerId = "offer_123",
    transactionId = "txn_abc123"
))
```

### Conversion with Amount

```kotlin
// Android
Afftok.trackConversion(ConversionParams(
    offerId = "offer_123",
    transactionId = "txn_abc123",
    amount = 29.99,
    currency = "USD",
    status = "approved"
))
```

### Conversion with Click Attribution

```kotlin
// Android
Afftok.trackConversion(ConversionParams(
    offerId = "offer_123",
    transactionId = "txn_abc123",
    clickId = "click_xyz",  // Link to original click
    amount = 49.99,
    status = "approved"
))
```

### Conversion with Metadata

```dart
// Flutter
await Afftok.instance.trackConversionWithMeta(
  ConversionParams(
    offerId: 'offer_123',
    transactionId: 'txn_abc123',
    amount: 29.99,
  ),
  {
    'product_id': 'prod_456',
    'category': 'electronics',
    'user_type': 'premium',
  },
);
```

---

## Offline Queue

The SDK automatically queues events when offline and retries with exponential backoff.

### Check Queue Status

```kotlin
// Android
val pendingCount = Afftok.getPendingCount()
val isReady = Afftok.isReady()
```

### Manual Flush

```kotlin
// Android
Afftok.flush {
    Log.d("AffTok", "Queue flushed!")
}

// iOS
await Afftok.shared.flush()

// Flutter
await Afftok.instance.flush();
```

### Clear Queue

```kotlin
Afftok.clearQueue()
```

---

## Device Information

### Get Device ID

```kotlin
val deviceId = Afftok.getDeviceId()
// Returns: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

### Get Fingerprint

```kotlin
val fingerprint = Afftok.getFingerprint()
// Returns: "abc123def456..."
```

### Get Full Device Info

```kotlin
val deviceInfo = Afftok.getDeviceInfo()
// Returns: Map with device_id, fingerprint, platform, os_version, model, etc.
```

---

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | String | Required | Your AffTok API key |
| `advertiserId` | String | Required | Your advertiser ID |
| `userId` | String? | null | Optional user identifier |
| `baseUrl` | String | `https://api.afftok.com` | API base URL |
| `debug` | Boolean | false | Enable debug logging |
| `autoFlush` | Boolean | true | Auto-flush queue |
| `flushInterval` | Long/Duration | 30s | Queue flush interval |

---

## Error Handling

### Response Structure

```kotlin
data class AfftokResponse(
    val success: Boolean,
    val message: String?,
    val data: Map<String, Any>?,
    val error: String?
)
```

### Handling Responses

```kotlin
val response = Afftok.trackClick(params)

when {
    response.success -> {
        // Event tracked successfully
        Log.d("AffTok", "Success: ${response.data}")
    }
    response.error != null -> {
        // Error occurred
        Log.e("AffTok", "Error: ${response.error}")
    }
    else -> {
        // Event queued for retry
        Log.w("AffTok", "Queued: ${response.message}")
    }
}
```

---

## Best Practices

### 1. Initialize Early

Initialize the SDK as early as possible in your app lifecycle.

### 2. Use Debug Mode

Enable debug mode during development to see detailed logs.

### 3. Handle Responses

Always check the response status and handle errors appropriately.

### 4. Unique Transaction IDs

Use unique transaction IDs for conversions to prevent duplicates.

### 5. Proper Shutdown

Call `shutdown()` when your app terminates:

```kotlin
override fun onDestroy() {
    Afftok.shutdown()
    super.onDestroy()
}
```

### 6. Background Tracking

For iOS, register background tasks for queue flushing:

```swift
BGTaskScheduler.shared.register(forTaskWithIdentifier: "com.afftok.flush", using: nil) { task in
    Task {
        await Afftok.shared.flush()
        task.setTaskCompleted(success: true)
    }
}
```

---

## Troubleshooting

### SDK Not Initializing

1. Check API key and advertiser ID
2. Verify network connectivity
3. Enable debug mode for logs

### Events Not Tracking

1. Check `isReady()` returns true
2. Verify offer ID is valid
3. Check queue with `getPendingCount()`

### High Queue Size

1. Check network connectivity
2. Verify API endpoint is reachable
3. Call `flush()` manually

---

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- GitHub: https://github.com/afftok

