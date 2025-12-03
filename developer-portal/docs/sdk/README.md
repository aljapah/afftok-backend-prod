# SDK Documentation

AffTok provides official SDKs for all major platforms to simplify tracking integration.

## Available SDKs

| Platform | Language | Package |
|----------|----------|---------|
| [Android](./android.md) | Kotlin/Java | `com.afftok:sdk` |
| [iOS](./ios.md) | Swift | `AfftokSDK` (SPM/CocoaPods) |
| [Flutter](./flutter.md) | Dart | `afftok_sdk` |
| [React Native](./react-native.md) | JavaScript | `@afftok/react-native` |
| [Web](./web.md) | JavaScript | `afftok.js` |

## Common Features

All SDKs provide:

- **Automatic Device Fingerprinting** - Unique device identification
- **Offline Queue** - Events stored when offline, synced when online
- **Retry Logic** - Exponential backoff for failed requests
- **Signed Requests** - HMAC-SHA256 request signing
- **Rate Limit Handling** - Automatic backoff on 429 responses
- **Debug Mode** - Verbose logging for development

## Quick Comparison

| Feature | Android | iOS | Flutter | React Native | Web |
|---------|---------|-----|---------|--------------|-----|
| Min Version | API 21 | iOS 13 | 2.0 | 0.63 | ES6 |
| Size | ~50KB | ~45KB | ~40KB | ~35KB | ~15KB |
| Offline Queue | ✅ | ✅ | ✅ | ✅ | ✅ |
| Background Sync | ✅ | ✅ | ✅ | ✅ | ❌ |
| Fingerprinting | ✅ | ✅ | ✅ | ✅ | ✅ |

## Installation Overview

### Android (Gradle)

```groovy
dependencies {
    implementation 'com.afftok:sdk:1.0.0'
}
```

### iOS (Swift Package Manager)

```swift
dependencies: [
    .package(url: "https://github.com/afftok/ios-sdk.git", from: "1.0.0")
]
```

### Flutter (pubspec.yaml)

```yaml
dependencies:
  afftok_sdk: ^1.0.0
```

### React Native (npm)

```bash
npm install @afftok/react-native
```

### Web (CDN)

```html
<script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
```

## Basic Usage Pattern

All SDKs follow the same pattern:

```
1. Initialize with API Key and Advertiser ID
2. Track clicks with offer/tracking data
3. Track conversions with amount and metadata
```

### Initialization (Pseudocode)

```
Afftok.init({
  apiKey: "afftok_live_sk_...",
  advertiserId: "adv_123456",
  userId: "user_789",
  debugMode: false
})
```

### Track Click

```
Afftok.trackClick({
  offerId: "off_123456",
  trackingCode: "abc123.1701432000.xyz.sig",
  metadata: {
    source: "email_campaign",
    creative: "banner_a"
  }
})
```

### Track Conversion

```
Afftok.trackConversion({
  offerId: "off_123456",
  amount: 49.99,
  currency: "USD",
  orderId: "order_12345",
  metadata: {
    product: "premium_plan"
  }
})
```

## Authentication

SDKs authenticate using API Keys:

```
API Key Format: afftok_live_sk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

- **Live Keys** (`afftok_live_sk_...`) - Production environment
- **Test Keys** (`afftok_test_sk_...`) - Sandbox environment

## Request Signing

All SDK requests are signed with HMAC-SHA256:

```
Signature = HMAC-SHA256(payload + timestamp + nonce, api_key)
```

Headers sent:
- `X-Afftok-Timestamp` - Unix timestamp
- `X-Afftok-Nonce` - Random string
- `X-Afftok-Signature` - HMAC signature

## Offline Queue

When network is unavailable:

1. Events are stored locally (SharedPreferences/UserDefaults/localStorage)
2. Queue is checked periodically (default: 30 seconds)
3. Events are sent with exponential backoff on failure
4. Maximum queue size: 1000 events (configurable)
5. Events older than 7 days are discarded

## Device Fingerprinting

The SDK generates a unique device fingerprint using:

| Android | iOS | Web |
|---------|-----|-----|
| Android ID | Vendor ID | Canvas fingerprint |
| Build info | Device model | WebGL renderer |
| Screen size | Screen size | Screen size |
| Timezone | Timezone | Timezone |
| Language | Language | Language |

## Error Handling

All SDKs provide error callbacks:

```javascript
Afftok.trackClick({...})
  .then(result => console.log("Success:", result))
  .catch(error => console.error("Error:", error.code, error.message))
```

Common error codes:
- `NETWORK_ERROR` - Network unavailable
- `INVALID_API_KEY` - API key validation failed
- `RATE_LIMITED` - Too many requests
- `VALIDATION_ERROR` - Invalid parameters

## Debug Mode

Enable verbose logging:

```javascript
Afftok.init({
  apiKey: "...",
  debugMode: true  // Enables console logging
})
```

Debug mode logs:
- All API requests/responses
- Queue operations
- Fingerprint generation
- Retry attempts

## Best Practices

### 1. Initialize Early

Initialize the SDK as early as possible in your app lifecycle:

```kotlin
// Android - in Application.onCreate()
class MyApp : Application() {
    override fun onCreate() {
        super.onCreate()
        Afftok.init(this, config)
    }
}
```

### 2. Handle Deep Links

Track clicks from deep links:

```swift
// iOS
func application(_ app: UIApplication, open url: URL) -> Bool {
    if let trackingCode = url.queryParameters["ref"] {
        Afftok.trackClick(trackingCode: trackingCode)
    }
    return true
}
```

### 3. Track Conversions Server-Side

For high-value conversions, also send server-side:

```javascript
// Client
Afftok.trackConversion({ amount: 99.99, orderId: "123" })

// Server (backup)
fetch('/api/postback', {
  method: 'POST',
  headers: { 'X-API-Key': 'afftok_live_sk_...' },
  body: JSON.stringify({ amount: 99.99, order_id: "123" })
})
```

### 4. Test in Sandbox

Use test API keys during development:

```javascript
Afftok.init({
  apiKey: "afftok_test_sk_...",  // Test key
  debugMode: true
})
```

## Migration Guide

### From v0.x to v1.0

```javascript
// Old (v0.x)
Afftok.track('click', { offer: '123' })

// New (v1.0)
Afftok.trackClick({ offerId: '123' })
```

Key changes:
- Separate `trackClick` and `trackConversion` methods
- Required `offerId` parameter
- New `metadata` object for custom data

---

Select your platform for detailed documentation:

- [Android SDK](./android.md)
- [iOS SDK](./ios.md)
- [Flutter SDK](./flutter.md)
- [React Native SDK](./react-native.md)
- [Web SDK](./web.md)

