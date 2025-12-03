# iOS SDK

Complete guide for integrating AffTok tracking into iOS applications.

---

## Requirements

- iOS 12.0+
- Swift 5.5+
- Xcode 13.0+

---

## Installation

### CocoaPods

```ruby
# Podfile
pod 'Afftok', '~> 1.0'
```

```bash
pod install
```

### Swift Package Manager

In Xcode:
1. File → Add Packages...
2. Enter: `https://github.com/afftok/afftok-ios-sdk.git`
3. Select version: `1.0.0`

Or in `Package.swift`:

```swift
dependencies: [
    .package(url: "https://github.com/afftok/afftok-ios-sdk.git", from: "1.0.0")
]
```

### Carthage

```
github "afftok/afftok-ios-sdk" ~> 1.0
```

---

## Initialization

Initialize the SDK in your `AppDelegate`:

```swift
import Afftok

@main
class AppDelegate: UIResponder, UIApplicationDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
    ) -> Bool {
        
        Afftok.shared.initialize(
            apiKey: "afftok_live_sk_xxxxx",
            config: AfftokConfig(
                debug: true,
                enableOfflineQueue: true,
                maxQueueSize: 1000,
                flushInterval: 30,
                timeout: 30,
                retryAttempts: 3
            )
        )
        
        return true
    }
}
```

### SwiftUI Initialization

```swift
import SwiftUI
import Afftok

@main
struct MyApp: App {
    init() {
        Afftok.shared.initialize(
            apiKey: "afftok_live_sk_xxxxx",
            config: AfftokConfig(debug: true)
        )
    }
    
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
```

### Configuration Options

```swift
struct AfftokConfig {
    let debug: Bool
    let enableOfflineQueue: Bool
    let maxQueueSize: Int
    let flushInterval: TimeInterval
    let timeout: TimeInterval
    let retryAttempts: Int
    let baseUrl: String
    let enableDeviceFingerprint: Bool
    
    init(
        debug: Bool = false,
        enableOfflineQueue: Bool = true,
        maxQueueSize: Int = 1000,
        flushInterval: TimeInterval = 30,
        timeout: TimeInterval = 30,
        retryAttempts: Int = 3,
        baseUrl: String = "https://api.afftok.com",
        enableDeviceFingerprint: Bool = true
    )
}
```

---

## Click Tracking

### Basic Click Tracking

```swift
Afftok.shared.trackClick(trackingCode: "tc_abc123def456")
```

### Click with Metadata

```swift
Afftok.shared.trackClick(
    trackingCode: "tc_abc123def456",
    metadata: [
        "campaign": "summer_sale",
        "source": "push_notification",
        "sub_id": "user_123"
    ]
)
```

### Click with Completion Handler

```swift
Afftok.shared.trackClick(
    trackingCode: "tc_abc123def456",
    metadata: ["campaign": "summer_sale"]
) { result in
    switch result {
    case .success(let data):
        print("Click tracked: \(data.clickId)")
    case .failure(let error):
        print("Error: \(error.localizedDescription)")
    }
}
```

### Async/Await Support

```swift
Task {
    do {
        let result = try await Afftok.shared.trackClick(
            trackingCode: "tc_abc123def456",
            metadata: ["campaign": "summer_sale"]
        )
        print("Click tracked: \(result.clickId)")
    } catch {
        print("Error: \(error.localizedDescription)")
    }
}
```

---

## Conversion Tracking

### Basic Conversion

```swift
Afftok.shared.trackConversion(
    offerId: "off_xyz789",
    transactionId: "txn_\(Date().timeIntervalSince1970)"
)
```

### Conversion with Amount

```swift
Afftok.shared.trackConversion(
    offerId: "off_xyz789",
    transactionId: "order_12345",
    amount: 49.99,
    currency: "USD"
)
```

### Full Conversion Tracking

```swift
Afftok.shared.trackConversion(
    offerId: "off_xyz789",
    transactionId: "order_12345",
    clickId: "clk_abc123",  // Optional: link to specific click
    amount: 49.99,
    currency: "USD",
    status: .approved,
    metadata: [
        "product_id": "prod_456",
        "category": "electronics",
        "user_type": "new"
    ]
) { result in
    switch result {
    case .success(let data):
        print("Conversion tracked: \(data.conversionId)")
    case .failure(let error):
        print("Error: \(error.localizedDescription)")
    }
}
```

### Conversion Status

```swift
enum ConversionStatus: String {
    case pending
    case approved
    case rejected
}
```

---

## Device Fingerprinting

The SDK automatically generates a unique device fingerprint:

```swift
// Get current device fingerprint
let fingerprint = Afftok.shared.getDeviceFingerprint()
print("Device Fingerprint: \(fingerprint)")
```

### Custom Fingerprint Provider

```swift
Afftok.shared.setFingerprintProvider { () -> String in
    // Custom fingerprint logic
    let identifier = UIDevice.current.identifierForVendor?.uuidString ?? ""
    return "custom_\(identifier)"
}
```

---

## Offline Queue

Events are automatically queued when offline:

```swift
// Check queue status
let queueSize = Afftok.shared.getQueueSize()
print("Queued events: \(queueSize)")

// Force flush queue
Afftok.shared.flushQueue { result in
    switch result {
    case .success(let data):
        print("Queue flushed: \(data.flushedCount) events")
    case .failure(let error):
        print("Flush failed: \(error.localizedDescription)")
    }
}

// Clear queue
Afftok.shared.clearQueue()
```

---

## Deep Link Handling

### UIKit

```swift
class SceneDelegate: UIResponder, UIWindowSceneDelegate {
    func scene(
        _ scene: UIScene,
        openURLContexts URLContexts: Set<UIOpenURLContext>
    ) {
        guard let url = URLContexts.first?.url else { return }
        
        Afftok.shared.handleDeepLink(url) { result in
            switch result {
            case .success(let data):
                print("Deep link tracked: \(data)")
            case .failure(let error):
                print("Error: \(error.localizedDescription)")
            }
        }
    }
    
    func scene(
        _ scene: UIScene,
        continue userActivity: NSUserActivity
    ) {
        guard let url = userActivity.webpageURL else { return }
        Afftok.shared.handleDeepLink(url)
    }
}
```

### SwiftUI

```swift
struct ContentView: View {
    var body: some View {
        Text("Hello, World!")
            .onOpenURL { url in
                Afftok.shared.handleDeepLink(url)
            }
    }
}
```

### Generate Deep Link

```swift
let deepLink = Afftok.shared.generateDeepLink(
    trackingCode: "tc_abc123",
    scheme: "myapp",
    path: "/offer/details"
)
// Result: myapp://offer/details?tc=tc_abc123
```

---

## Universal Links

Configure `apple-app-site-association` on your domain:

```json
{
  "applinks": {
    "apps": [],
    "details": [
      {
        "appID": "TEAM_ID.com.yourcompany.app",
        "paths": ["/c/*"]
      }
    ]
  }
}
```

Handle Universal Links:

```swift
func application(
    _ application: UIApplication,
    continue userActivity: NSUserActivity,
    restorationHandler: @escaping ([UIUserActivityRestoring]?) -> Void
) -> Bool {
    guard let url = userActivity.webpageURL else { return false }
    
    return Afftok.shared.handleUniversalLink(url) { result in
        switch result {
        case .success(let data):
            print("Universal link tracked: \(data)")
        case .failure(let error):
            print("Error: \(error.localizedDescription)")
        }
    }
}
```

---

## Push Notification Attribution

### APNs

```swift
func application(
    _ application: UIApplication,
    didReceiveRemoteNotification userInfo: [AnyHashable: Any],
    fetchCompletionHandler completionHandler: @escaping (UIBackgroundFetchResult) -> Void
) {
    // Extract tracking code from push payload
    if let trackingCode = userInfo["afftok_tc"] as? String {
        Afftok.shared.trackClick(
            trackingCode: trackingCode,
            metadata: [
                "source": "push_notification",
                "campaign": userInfo["campaign"] as? String ?? ""
            ]
        )
    }
    
    completionHandler(.newData)
}
```

### User Notification Center

```swift
extension AppDelegate: UNUserNotificationCenterDelegate {
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let userInfo = response.notification.request.content.userInfo
        
        if let trackingCode = userInfo["afftok_tc"] as? String {
            Afftok.shared.trackClick(
                trackingCode: trackingCode,
                metadata: ["source": "push_notification"]
            )
        }
        
        completionHandler()
    }
}
```

---

## User Identification

Associate events with a user:

```swift
// Set user ID
Afftok.shared.setUserId("user_12345")

// Set user properties
Afftok.shared.setUserProperties([
    "email": "user@example.com",
    "plan": "premium",
    "signup_date": "2024-01-15"
])

// Clear user (on logout)
Afftok.shared.clearUser()
```

---

## Error Handling

```swift
enum AfftokError: Error {
    case networkError(String)
    case authError(String)
    case validationError(String)
    case rateLimitError(String, retryAfter: TimeInterval)
    case serverError(String, code: Int)
}

// Handle errors
Afftok.shared.trackClick(trackingCode: "tc_abc123") { result in
    switch result {
    case .success(let data):
        print("Click tracked: \(data.clickId)")
    case .failure(let error):
        switch error {
        case .networkError(let message):
            print("Network error (event queued): \(message)")
        case .rateLimitError(let message, let retryAfter):
            print("Rate limited, retry after \(retryAfter)s: \(message)")
        case .authError(let message):
            print("Auth error: \(message)")
        default:
            print("Error: \(error.localizedDescription)")
        }
    }
}
```

---

## App Tracking Transparency

Handle iOS 14+ ATT framework:

```swift
import AppTrackingTransparency

func requestTrackingPermission() {
    if #available(iOS 14, *) {
        ATTrackingManager.requestTrackingAuthorization { status in
            switch status {
            case .authorized:
                // Full tracking enabled
                Afftok.shared.setTrackingEnabled(true)
            case .denied, .restricted:
                // Limited tracking
                Afftok.shared.setTrackingEnabled(false)
            case .notDetermined:
                break
            @unknown default:
                break
            }
        }
    }
}
```

---

## Debug Mode

Enable detailed logging:

```swift
AfftokConfig(debug: true)

// Or at runtime
Afftok.shared.setDebugMode(true)
```

Debug logs include:
- All API requests/responses
- Queue operations
- Fingerprint generation
- Error details

---

## Complete Example

```swift
import SwiftUI
import Afftok

struct ContentView: View {
    @State private var isPurchasing = false
    
    var body: some View {
        VStack(spacing: 20) {
            Text("Premium Subscription")
                .font(.title)
            
            Text("$9.99/month")
                .font(.headline)
            
            Button(action: completePurchase) {
                if isPurchasing {
                    ProgressView()
                } else {
                    Text("Subscribe Now")
                }
            }
            .buttonStyle(.borderedProminent)
            .disabled(isPurchasing)
        }
        .padding()
        .onAppear {
            // Track screen view
            Afftok.shared.trackEvent("screen_view", metadata: ["screen": "subscription"])
        }
    }
    
    func completePurchase() {
        isPurchasing = true
        
        // Your purchase logic here...
        
        // Track conversion
        Task {
            do {
                let result = try await Afftok.shared.trackConversion(
                    offerId: "off_premium_subscription",
                    transactionId: "txn_\(Date().timeIntervalSince1970)",
                    amount: 9.99,
                    currency: "USD",
                    metadata: [
                        "product": "premium_monthly",
                        "user_id": getCurrentUserId()
                    ]
                )
                print("Conversion tracked: \(result.conversionId)")
            } catch {
                print("Tracking failed: \(error.localizedDescription)")
            }
            
            isPurchasing = false
        }
    }
    
    func getCurrentUserId() -> String {
        // Return current user ID
        return "user_12345"
    }
}
```

---

## Privacy Manifest

Add to your app's privacy manifest (`PrivacyInfo.xcprivacy`):

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>NSPrivacyTracking</key>
    <false/>
    <key>NSPrivacyTrackingDomains</key>
    <array>
        <string>api.afftok.com</string>
    </array>
    <key>NSPrivacyCollectedDataTypes</key>
    <array>
        <dict>
            <key>NSPrivacyCollectedDataType</key>
            <string>NSPrivacyCollectedDataTypeDeviceID</string>
            <key>NSPrivacyCollectedDataTypeLinked</key>
            <false/>
            <key>NSPrivacyCollectedDataTypeTracking</key>
            <false/>
            <key>NSPrivacyCollectedDataTypePurposes</key>
            <array>
                <string>NSPrivacyCollectedDataTypePurposeAnalytics</string>
            </array>
        </dict>
    </array>
</dict>
</plist>
```

---

## Next Steps

- [Android SDK](android.md) - Android integration guide
- [Flutter SDK](flutter.md) - Flutter integration guide
- [API Reference](../api-reference/clicks.md) - Full API documentation

