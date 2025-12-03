# SDK Overview

AffTok provides official SDKs for all major platforms to simplify integration and ensure reliable tracking.

---

## Available SDKs

| Platform | Language | Package |
|----------|----------|---------|
| Android | Kotlin | `com.afftok:sdk:1.0.0` |
| iOS | Swift | `Afftok` (CocoaPods/SPM) |
| Flutter | Dart | `afftok_sdk` |
| React Native | JavaScript | `@afftok/react-native` |
| Web | JavaScript | `afftok.js` |

---

## Core Features

All SDKs provide:

- **Click Tracking** - Track user clicks with full attribution
- **Conversion Tracking** - Report conversions with custom data
- **Signed Links** - Cryptographically secure tracking URLs
- **HMAC-SHA256 Signing** - Request authentication
- **Offline Queue** - Store events when offline
- **Retry Logic** - Automatic retry with exponential backoff
- **Device Fingerprinting** - Unique device identification
- **Rate Limiting** - Respect API limits

---

## Quick Comparison

| Feature | Android | iOS | Flutter | React Native | Web |
|---------|---------|-----|---------|--------------|-----|
| Click Tracking | ✅ | ✅ | ✅ | ✅ | ✅ |
| Conversion Tracking | ✅ | ✅ | ✅ | ✅ | ✅ |
| Offline Queue | ✅ | ✅ | ✅ | ✅ | ✅ |
| Device Fingerprint | ✅ | ✅ | ✅ | ✅ | ✅ |
| Deep Links | ✅ | ✅ | ✅ | ✅ | ❌ |
| Push Attribution | ✅ | ✅ | ✅ | ✅ | ❌ |
| Install Referrer | ✅ | ❌ | ✅ | ✅ | ❌ |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Your App                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   AffTok SDK                         │    │
│  │  ┌─────────┐  ┌──────────┐  ┌──────────────────┐    │    │
│  │  │  Click  │  │Conversion│  │ Device Fingerprint│    │    │
│  │  │ Tracker │  │ Tracker  │  │     Provider      │    │    │
│  │  └────┬────┘  └────┬─────┘  └────────┬─────────┘    │    │
│  │       │            │                  │              │    │
│  │  ┌────▼────────────▼──────────────────▼─────────┐   │    │
│  │  │              Event Queue                      │   │    │
│  │  │  (Offline Storage + Retry Logic)              │   │    │
│  │  └────────────────────┬──────────────────────────┘   │    │
│  │                       │                              │    │
│  │  ┌────────────────────▼──────────────────────────┐   │    │
│  │  │              Request Signer                    │   │    │
│  │  │  (HMAC-SHA256 + Timestamp + Nonce)             │   │    │
│  │  └────────────────────┬──────────────────────────┘   │    │
│  └───────────────────────┼──────────────────────────────┘    │
└──────────────────────────┼──────────────────────────────────┘
                           │
                           ▼
              ┌─────────────────────────┐
              │     AffTok API          │
              │  api.afftok.com         │
              └─────────────────────────┘
```

---

## Installation

### Android (Gradle)

```groovy
dependencies {
    implementation 'com.afftok:sdk:1.0.0'
}
```

### iOS (CocoaPods)

```ruby
pod 'Afftok', '~> 1.0'
```

### iOS (Swift Package Manager)

```swift
.package(url: "https://github.com/afftok/afftok-ios-sdk.git", from: "1.0.0")
```

### Flutter

```yaml
dependencies:
  afftok_sdk: ^1.0.0
```

### React Native

```bash
npm install @afftok/react-native
# or
yarn add @afftok/react-native
```

### Web

```html
<script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>
```

Or via npm:

```bash
npm install afftok-js
```

---

## Basic Usage

### Initialize SDK

**Android (Kotlin):**

```kotlin
Afftok.init(
    context = applicationContext,
    apiKey = "afftok_live_sk_xxxxx",
    config = AfftokConfig(
        debug = BuildConfig.DEBUG,
        enableOfflineQueue = true
    )
)
```

**iOS (Swift):**

```swift
Afftok.shared.initialize(
    apiKey: "afftok_live_sk_xxxxx",
    config: AfftokConfig(
        debug: true,
        enableOfflineQueue: true
    )
)
```

**Flutter:**

```dart
await Afftok.init(
  apiKey: 'afftok_live_sk_xxxxx',
  config: AfftokConfig(
    debug: true,
    enableOfflineQueue: true,
  ),
);
```

**React Native:**

```javascript
import Afftok from '@afftok/react-native';

Afftok.init({
  apiKey: 'afftok_live_sk_xxxxx',
  debug: __DEV__,
  enableOfflineQueue: true,
});
```

**Web:**

```javascript
Afftok.init({
  apiKey: 'afftok_live_sk_xxxxx',
  debug: true,
});
```

---

### Track Click

**Android:**

```kotlin
Afftok.trackClick(
    trackingCode = "tc_abc123",
    metadata = mapOf(
        "campaign" to "summer_sale",
        "source" to "email"
    )
)
```

**iOS:**

```swift
Afftok.shared.trackClick(
    trackingCode: "tc_abc123",
    metadata: [
        "campaign": "summer_sale",
        "source": "email"
    ]
)
```

**Flutter:**

```dart
await Afftok.trackClick(
  trackingCode: 'tc_abc123',
  metadata: {
    'campaign': 'summer_sale',
    'source': 'email',
  },
);
```

**React Native:**

```javascript
await Afftok.trackClick('tc_abc123', {
  campaign: 'summer_sale',
  source: 'email',
});
```

**Web:**

```javascript
Afftok.trackClick('tc_abc123', {
  campaign: 'summer_sale',
  source: 'email',
});
```

---

### Track Conversion

**Android:**

```kotlin
Afftok.trackConversion(
    offerId = "off_xyz789",
    transactionId = "txn_${System.currentTimeMillis()}",
    amount = 49.99,
    currency = "USD",
    metadata = mapOf(
        "product_id" to "prod_123",
        "category" to "electronics"
    )
)
```

**iOS:**

```swift
Afftok.shared.trackConversion(
    offerId: "off_xyz789",
    transactionId: "txn_\(Date().timeIntervalSince1970)",
    amount: 49.99,
    currency: "USD",
    metadata: [
        "product_id": "prod_123",
        "category": "electronics"
    ]
)
```

**Flutter:**

```dart
await Afftok.trackConversion(
  offerId: 'off_xyz789',
  transactionId: 'txn_${DateTime.now().millisecondsSinceEpoch}',
  amount: 49.99,
  currency: 'USD',
  metadata: {
    'product_id': 'prod_123',
    'category': 'electronics',
  },
);
```

**React Native:**

```javascript
await Afftok.trackConversion({
  offerId: 'off_xyz789',
  transactionId: `txn_${Date.now()}`,
  amount: 49.99,
  currency: 'USD',
  metadata: {
    product_id: 'prod_123',
    category: 'electronics',
  },
});
```

**Web:**

```javascript
Afftok.trackConversion({
  offerId: 'off_xyz789',
  transactionId: `txn_${Date.now()}`,
  amount: 49.99,
  currency: 'USD',
  metadata: {
    product_id: 'prod_123',
    category: 'electronics',
  },
});
```

---

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | string | required | Your API key |
| `debug` | boolean | false | Enable debug logging |
| `enableOfflineQueue` | boolean | true | Queue events when offline |
| `maxQueueSize` | integer | 1000 | Maximum offline queue size |
| `flushInterval` | integer | 30000 | Queue flush interval (ms) |
| `timeout` | integer | 30000 | Request timeout (ms) |
| `retryAttempts` | integer | 3 | Max retry attempts |
| `baseUrl` | string | api.afftok.com | API base URL |

---

## Error Handling

All SDKs provide consistent error handling:

```kotlin
// Android
Afftok.trackClick("tc_abc123") { result ->
    when (result) {
        is Result.Success -> println("Click tracked: ${result.data}")
        is Result.Error -> println("Error: ${result.error.message}")
    }
}
```

```swift
// iOS
Afftok.shared.trackClick(trackingCode: "tc_abc123") { result in
    switch result {
    case .success(let data):
        print("Click tracked: \(data)")
    case .failure(let error):
        print("Error: \(error.localizedDescription)")
    }
}
```

```dart
// Flutter
try {
  await Afftok.trackClick(trackingCode: 'tc_abc123');
  print('Click tracked');
} on AfftokException catch (e) {
  print('Error: ${e.message}');
}
```

---

## Next Steps

- [Android SDK](android.md) - Full Android documentation
- [iOS SDK](ios.md) - Full iOS documentation
- [Flutter SDK](flutter.md) - Full Flutter documentation
- [React Native SDK](react-native.md) - Full React Native documentation
- [Web SDK](web.md) - Full Web SDK documentation

