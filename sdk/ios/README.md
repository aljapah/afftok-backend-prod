# AffTok iOS SDK

Official iOS SDK for AffTok affiliate tracking platform.

## Features

- 🚀 Click & Conversion Tracking
- 📴 Offline Queue with Automatic Retry
- 🔐 HMAC-SHA256 Signature Verification
- 🔄 Exponential Backoff Retry
- 📱 Device Fingerprinting
- 🎯 Zero-Drop Tracking Compatible

## Requirements

- iOS 13.0+
- Swift 5.5+
- Xcode 13.0+

## Installation

### Swift Package Manager

Add to your `Package.swift`:

```swift
dependencies: [
    .package(url: "https://github.com/afftok/ios-sdk.git", from: "1.0.0")
]
```

Or in Xcode:
1. File → Add Packages
2. Enter: `https://github.com/afftok/ios-sdk.git`
3. Select version: 1.0.0

### CocoaPods

Add to your `Podfile`:

```ruby
pod 'AfftokSDK', '~> 1.0.0'
```

### Carthage

Add to your `Cartfile`:

```
github "afftok/ios-sdk" ~> 1.0.0
```

## Quick Start

### 1. Initialize the SDK

Initialize in your AppDelegate or SceneDelegate:

```swift
import Afftok

@main
class AppDelegate: UIResponder, UIApplicationDelegate {
    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        
        Afftok.shared.initialize(options: AfftokOptions(
            apiKey: "your_api_key",
            advertiserId: "your_advertiser_id",
            userId: "optional_user_id",  // Optional
            debug: true                   // Enable logging
        ))
        
        return true
    }
}
```

### 2. Track Clicks

```swift
import Afftok

// Using async/await
Task {
    let response = await Afftok.shared.trackClick(params: ClickParams(
        offerId: "offer_123",
        trackingCode: "abc123",  // Optional
        subId1: "campaign_1",    // Optional
        subId2: "creative_2",    // Optional
        customParams: [          // Optional
            "source": "facebook",
            "campaign": "summer_sale"
        ]
    ))
    
    if response.success {
        print("Click tracked!")
    }
}

// Using completion handler
Afftok.shared.trackClick(params: ClickParams(offerId: "offer_123")) { response in
    if response.success {
        print("Click tracked!")
    }
}
```

### 3. Track Conversions

```swift
import Afftok

Task {
    let response = await Afftok.shared.trackConversion(params: ConversionParams(
        offerId: "offer_123",
        transactionId: "txn_abc123",
        amount: 29.99,
        currency: "USD",
        status: "approved",   // "pending", "approved", "rejected"
        clickId: "click_xyz"  // Optional, for attribution
    ))
    
    if response.success {
        print("Conversion tracked!")
    }
}
```

### 4. Track with Signed Links

```swift
// When using pre-signed tracking links from AffTok
Task {
    let response = await Afftok.shared.trackSignedClick(
        signedLink: "https://track.afftok.com/c/abc123.1234567890.xyz.signature",
        params: ClickParams(offerId: "offer_123")
    )
}
```

## Offline Queue

The SDK automatically queues events when offline and retries with exponential backoff.

```swift
// Check pending items
let pendingCount = Afftok.shared.getPendingCount()

// Manually flush queue
Afftok.shared.flush {
    print("Queue flushed!")
}

// Clear queue
Afftok.shared.clearQueue()
```

## Device Information

```swift
// Get device fingerprint
let fingerprint = Afftok.shared.getFingerprint()

// Get device ID
let deviceId = Afftok.shared.getDeviceId()

// Get full device info
let deviceInfo = Afftok.shared.getDeviceInfo()
```

## Configuration Options

```swift
AfftokOptions(
    apiKey: "your_api_key",              // Required
    advertiserId: "your_advertiser",      // Required
    userId: nil,                          // Optional user identifier
    baseUrl: "https://api.afftok.com",   // Custom API URL
    debug: false,                         // Enable debug logging
    autoFlush: true,                      // Auto-flush queue
    flushInterval: 30.0                   // Flush interval in seconds
)
```

## Error Handling

```swift
let response = await Afftok.shared.trackClick(params: params)

if response.success {
    // Event tracked successfully
    if let data = response.data {
        print("Response data: \(data)")
    }
} else if let error = response.error {
    // Error occurred
    print("Error: \(error)")
} else if let message = response.message {
    // Event queued for retry
    print("Queued: \(message)")
}
```

## SwiftUI Integration

```swift
import SwiftUI
import Afftok

struct ContentView: View {
    @State private var isTracking = false
    
    var body: some View {
        Button("Track Click") {
            isTracking = true
            Task {
                let response = await Afftok.shared.trackClick(
                    params: ClickParams(offerId: "offer_123")
                )
                isTracking = false
                print("Tracked: \(response.success)")
            }
        }
        .disabled(isTracking)
    }
}
```

## Background Tasks

For iOS 13+, register background tasks for queue flushing:

```swift
// In AppDelegate
func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
    
    BGTaskScheduler.shared.register(forTaskWithIdentifier: "com.afftok.flush", using: nil) { task in
        Task {
            await Afftok.shared.flush()
            task.setTaskCompleted(success: true)
        }
    }
    
    return true
}
```

## Best Practices

1. **Initialize Early**: Initialize the SDK in `didFinishLaunchingWithOptions`
2. **Use Debug Mode**: Enable debug mode during development
3. **Handle Responses**: Always check response status
4. **Unique Transaction IDs**: Use unique IDs for conversions
5. **Shutdown Properly**: Call `Afftok.shared.shutdown()` when app terminates

## Privacy

Add to your `Info.plist` if required:

```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <false/>
</dict>
```

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- Issues: https://github.com/afftok/ios-sdk/issues

## License

MIT License - see LICENSE file for details.

