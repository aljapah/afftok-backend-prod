# iOS SDK

Official AffTok SDK for iOS applications.

## Requirements

- iOS 13.0+
- Xcode 14.0+
- Swift 5.7+

## Installation

### Swift Package Manager (Recommended)

Add to your `Package.swift`:

```swift
dependencies: [
    .package(url: "https://github.com/afftok/ios-sdk.git", from: "1.0.0")
]
```

Or in Xcode:
1. File → Add Packages...
2. Enter: `https://github.com/afftok/ios-sdk.git`
3. Select version: `1.0.0`

### CocoaPods

```ruby
# Podfile
pod 'AfftokSDK', '~> 1.0.0'
```

Then run:
```bash
pod install
```

### Carthage

```
github "afftok/ios-sdk" ~> 1.0.0
```

## Initialization

Initialize the SDK in your `AppDelegate`:

```swift
import AfftokSDK

@main
class AppDelegate: UIResponder, UIApplicationDelegate {
    
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
    ) -> Bool {
        
        let config = AfftokConfig(
            apiKey: "afftok_live_sk_your_api_key_here",
            advertiserId: "adv_123456",
            userId: "user_789",  // Optional
            debugMode: false,
            offlineQueueEnabled: true,
            maxOfflineEvents: 1000,
            flushIntervalSeconds: 30
        )
        
        Afftok.initialize(config: config)
        
        return true
    }
}
```

### SwiftUI App

```swift
import SwiftUI
import AfftokSDK

@main
struct MyApp: App {
    init() {
        let config = AfftokConfig(
            apiKey: "afftok_live_sk_your_api_key_here",
            advertiserId: "adv_123456"
        )
        Afftok.initialize(config: config)
    }
    
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | String | Required | Your API key |
| `advertiserId` | String | Required | Your advertiser ID |
| `userId` | String? | nil | Your internal user ID |
| `debugMode` | Bool | false | Enable verbose logging |
| `offlineQueueEnabled` | Bool | true | Enable offline queue |
| `maxOfflineEvents` | Int | 1000 | Max queued events |
| `flushIntervalSeconds` | Int | 30 | Queue flush interval |
| `baseUrl` | String? | nil | Custom API endpoint |

## Track Click

Track when a user clicks on an offer:

```swift
import AfftokSDK

// Basic click tracking
Afftok.trackClick(
    offerId: "off_123456",
    trackingCode: "abc123.1701432000.xyz.sig"
)

// With metadata and completion handler
Afftok.trackClick(
    offerId: "off_123456",
    trackingCode: "abc123.1701432000.xyz.sig",
    metadata: [
        "source": "email_campaign",
        "creative": "banner_a"
    ]
) { result in
    switch result {
    case .success(let response):
        print("Click tracked: \(response.clickId)")
    case .failure(let error):
        print("Error: \(error.code) - \(error.message)")
    }
}
```

### Click Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `offerId` | String | Yes | Offer identifier |
| `trackingCode` | String? | No | Signed tracking code |
| `subId` | String? | No | Sub-affiliate ID |
| `metadata` | [String: Any]? | No | Custom key-value pairs |
| `completion` | Callback? | No | Completion handler |

## Track Conversion

Track when a user completes a conversion:

```swift
import AfftokSDK

// Basic conversion
Afftok.trackConversion(
    offerId: "off_123456",
    amount: 49.99,
    currency: "USD"
)

// Full conversion with all options
Afftok.trackConversion(
    offerId: "off_123456",
    amount: 49.99,
    currency: "USD",
    orderId: "order_12345",
    productId: "prod_premium",
    metadata: [
        "plan": "annual",
        "coupon": "SAVE20"
    ]
) { result in
    switch result {
    case .success(let response):
        print("Conversion tracked: \(response.conversionId)")
    case .failure(let error):
        print("Error: \(error.code) - \(error.message)")
    }
}
```

### Conversion Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `offerId` | String | Yes | Offer identifier |
| `amount` | Double | Yes | Conversion value |
| `currency` | String | No | ISO currency code (default: USD) |
| `orderId` | String? | No | Your order/transaction ID |
| `productId` | String? | No | Product identifier |
| `metadata` | [String: Any]? | No | Custom key-value pairs |
| `completion` | Callback? | No | Completion handler |

## Async/Await Support

The SDK provides async versions of all methods:

```swift
import AfftokSDK

// Track click with async/await
Task {
    do {
        let result = try await Afftok.trackClick(
            offerId: "off_123456",
            trackingCode: "abc123.1701432000.xyz.sig"
        )
        print("Click ID: \(result.clickId)")
    } catch let error as AfftokError {
        print("Error: \(error.code) - \(error.message)")
    }
}

// Track conversion with async/await
Task {
    do {
        let result = try await Afftok.trackConversion(
            offerId: "off_123456",
            amount: 49.99,
            currency: "USD",
            orderId: "order_12345"
        )
        print("Conversion ID: \(result.conversionId)")
    } catch let error as AfftokError {
        print("Error: \(error.code)")
    }
}
```

## Deep Link Handling

### UIKit

```swift
class AppDelegate: UIResponder, UIApplicationDelegate {
    
    func application(
        _ app: UIApplication,
        open url: URL,
        options: [UIApplication.OpenURLOptionsKey: Any] = [:]
    ) -> Bool {
        handleDeepLink(url)
        return true
    }
    
    private func handleDeepLink(_ url: URL) {
        guard let components = URLComponents(url: url, resolvingAgainstBaseURL: true),
              let queryItems = components.queryItems else {
            return
        }
        
        let trackingCode = queryItems.first(where: { $0.name == "ref" })?.value
        let offerId = queryItems.first(where: { $0.name == "offer" })?.value
        
        if let trackingCode = trackingCode, let offerId = offerId {
            Afftok.trackClick(
                offerId: offerId,
                trackingCode: trackingCode,
                metadata: ["source": "deep_link"]
            )
        }
    }
}
```

### SwiftUI

```swift
struct ContentView: View {
    var body: some View {
        Text("Hello")
            .onOpenURL { url in
                handleDeepLink(url)
            }
    }
    
    private func handleDeepLink(_ url: URL) {
        guard let components = URLComponents(url: url, resolvingAgainstBaseURL: true),
              let queryItems = components.queryItems else {
            return
        }
        
        if let trackingCode = queryItems.first(where: { $0.name == "ref" })?.value,
           let offerId = queryItems.first(where: { $0.name == "offer" })?.value {
            Afftok.trackClick(
                offerId: offerId,
                trackingCode: trackingCode,
                metadata: ["source": "deep_link"]
            )
        }
    }
}
```

## Universal Links

Configure in `Associated Domains`:

```
applinks:track.yourdomain.com
```

Handle in AppDelegate:

```swift
func application(
    _ application: UIApplication,
    continue userActivity: NSUserActivity,
    restorationHandler: @escaping ([UIUserActivityRestoring]?) -> Void
) -> Bool {
    guard userActivity.activityType == NSUserActivityTypeBrowsingWeb,
          let url = userActivity.webpageURL else {
        return false
    }
    
    handleDeepLink(url)
    return true
}
```

## Offline Queue

The SDK automatically queues events when offline:

```swift
// Check queue status
let queueSize = Afftok.getQueueSize()
print("Pending events: \(queueSize)")

// Force flush queue
Afftok.flushQueue { sent, failed in
    print("Flushed: \(sent) sent, \(failed) failed")
}

// Clear queue
Afftok.clearQueue()
```

### Background Sync

Enable background fetch in Xcode:
1. Signing & Capabilities → Background Modes
2. Enable "Background fetch"

```swift
func application(
    _ application: UIApplication,
    performFetchWithCompletionHandler completionHandler: @escaping (UIBackgroundFetchResult) -> Void
) {
    Afftok.flushQueue { sent, failed in
        if sent > 0 {
            completionHandler(.newData)
        } else if failed > 0 {
            completionHandler(.failed)
        } else {
            completionHandler(.noData)
        }
    }
}
```

## Device Fingerprinting

The SDK automatically generates a device fingerprint:

```swift
// Get current fingerprint
let fingerprint = Afftok.getDeviceFingerprint()
print("Fingerprint: \(fingerprint)")
```

Fingerprint components:
- Vendor ID (identifierForVendor)
- Device model
- Screen resolution
- Timezone
- Language
- App version

## User Identification

```swift
// Set user ID
Afftok.setUserId("user_new_id")

// Get current user ID
let userId = Afftok.getUserId()

// Clear user ID (on logout)
Afftok.clearUserId()
```

## Error Handling

```swift
Afftok.trackClick(offerId: "off_123456") { result in
    switch result {
    case .success(let response):
        print("Success: \(response.clickId)")
        
    case .failure(let error):
        switch error.code {
        case .networkError:
            // Event queued for retry
            print("Offline, event queued")
            
        case .invalidApiKey:
            // Check API key configuration
            print("Invalid API key")
            
        case .rateLimited:
            // Too many requests
            print("Rate limited, retry later")
            
        case .validationError:
            // Invalid parameters
            print("Validation: \(error.message)")
            
        default:
            print("Error: \(error.message)")
        }
    }
}
```

### Error Types

```swift
enum AfftokErrorCode {
    case networkError
    case invalidApiKey
    case rateLimited
    case validationError
    case serverError
    case unknown
}

struct AfftokError: Error {
    let code: AfftokErrorCode
    let message: String
    let details: [String: Any]?
}
```

## Combine Support

```swift
import Combine
import AfftokSDK

class ViewModel: ObservableObject {
    private var cancellables = Set<AnyCancellable>()
    
    func trackClick() {
        Afftok.trackClickPublisher(
            offerId: "off_123456",
            trackingCode: "abc123"
        )
        .sink(
            receiveCompletion: { completion in
                if case .failure(let error) = completion {
                    print("Error: \(error.message)")
                }
            },
            receiveValue: { response in
                print("Click ID: \(response.clickId)")
            }
        )
        .store(in: &cancellables)
    }
}
```

## Testing

### Test Mode

```swift
let config = AfftokConfig(
    apiKey: "afftok_test_sk_...",  // Test key
    advertiserId: "adv_test_123",
    debugMode: true
)
Afftok.initialize(config: config)
```

### Mock Events

```swift
#if DEBUG
// Generate test events
Afftok.generateTestClick()
Afftok.generateTestConversion(amount: 99.99)
#endif
```

## Complete Example

```swift
import SwiftUI
import AfftokSDK

@main
struct MyApp: App {
    init() {
        Afftok.initialize(config: AfftokConfig(
            apiKey: "afftok_live_sk_your_key",
            advertiserId: "adv_123456",
            debugMode: false
        ))
    }
    
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}

struct OfferView: View {
    let offerId: String
    let trackingCode: String?
    
    @State private var isPurchasing = false
    @State private var showSuccess = false
    
    var body: some View {
        VStack {
            Text("Special Offer")
                .font(.title)
            
            Button("Purchase Now - $49.99") {
                purchase()
            }
            .disabled(isPurchasing)
        }
        .onAppear {
            trackOfferView()
        }
        .alert("Purchase Complete!", isPresented: $showSuccess) {
            Button("OK", role: .cancel) {}
        }
    }
    
    private func trackOfferView() {
        if let trackingCode = trackingCode {
            Afftok.trackClick(
                offerId: offerId,
                trackingCode: trackingCode,
                metadata: ["screen": "offer_detail"]
            )
        }
    }
    
    private func purchase() {
        isPurchasing = true
        let orderId = "order_\(Int(Date().timeIntervalSince1970))"
        
        // Your purchase logic here...
        
        Afftok.trackConversion(
            offerId: offerId,
            amount: 49.99,
            currency: "USD",
            orderId: orderId
        ) { result in
            isPurchasing = false
            if case .success = result {
                showSuccess = true
            }
        }
    }
}
```

## Privacy

The SDK collects:
- Device identifier (vendor ID)
- Device model
- OS version
- App version
- Timezone
- Language

Add to your Privacy Policy that you use analytics for affiliate tracking.

## Changelog

### 1.0.0
- Initial release
- Click and conversion tracking
- Offline queue support
- Device fingerprinting
- HMAC request signing
- Async/await support
- Combine publishers

---

Next: [Flutter SDK](./flutter.md)

