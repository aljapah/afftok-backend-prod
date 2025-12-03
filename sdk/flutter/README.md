# AffTok Flutter SDK

Official Flutter SDK for AffTok affiliate tracking platform.

## Features

- 🚀 Click & Conversion Tracking
- 📴 Offline Queue with Automatic Retry
- 🔐 HMAC-SHA256 Signature Verification
- 🔄 Exponential Backoff Retry
- 📱 Device Fingerprinting
- 🎯 Zero-Drop Tracking Compatible
- 🌐 Cross-Platform (iOS & Android)

## Requirements

- Flutter 3.0+
- Dart 3.0+

## Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  afftok: ^1.0.0
```

Then run:

```bash
flutter pub get
```

## Quick Start

### 1. Initialize the SDK

Initialize in your main.dart:

```dart
import 'package:afftok/afftok.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  await Afftok.instance.initialize(AfftokOptions(
    apiKey: 'your_api_key',
    advertiserId: 'your_advertiser_id',
    userId: 'optional_user_id',  // Optional
    debug: true,                  // Enable logging
  ));
  
  runApp(MyApp());
}
```

### 2. Track Clicks

```dart
import 'package:afftok/afftok.dart';

// Track a click
final response = await Afftok.instance.trackClick(ClickParams(
  offerId: 'offer_123',
  trackingCode: 'abc123',    // Optional
  subId1: 'campaign_1',      // Optional
  subId2: 'creative_2',      // Optional
  customParams: {            // Optional
    'source': 'facebook',
    'campaign': 'summer_sale',
  },
));

if (response.success) {
  print('Click tracked!');
}
```

### 3. Track Conversions

```dart
import 'package:afftok/afftok.dart';

final response = await Afftok.instance.trackConversion(ConversionParams(
  offerId: 'offer_123',
  transactionId: 'txn_abc123',
  amount: 29.99,
  currency: 'USD',
  status: 'approved',    // 'pending', 'approved', 'rejected'
  clickId: 'click_xyz',  // Optional, for attribution
));

if (response.success) {
  print('Conversion tracked!');
}
```

### 4. Track with Signed Links

```dart
// When using pre-signed tracking links from AffTok
final response = await Afftok.instance.trackSignedClick(
  'https://track.afftok.com/c/abc123.1234567890.xyz.signature',
  ClickParams(offerId: 'offer_123'),
);
```

## Offline Queue

The SDK automatically queues events when offline and retries with exponential backoff.

```dart
// Check pending items
final pendingCount = Afftok.instance.getPendingCount();

// Manually flush queue
await Afftok.instance.flush();

// Clear queue
Afftok.instance.clearQueue();
```

## Device Information

```dart
// Get device fingerprint
final fingerprint = Afftok.instance.getFingerprint();

// Get device ID
final deviceId = Afftok.instance.getDeviceId();

// Get full device info
final deviceInfo = Afftok.instance.getDeviceInfo();
```

## Configuration Options

```dart
AfftokOptions(
  apiKey: 'your_api_key',              // Required
  advertiserId: 'your_advertiser',      // Required
  userId: null,                         // Optional user identifier
  baseUrl: 'https://api.afftok.com',   // Custom API URL
  debug: false,                         // Enable debug logging
  autoFlush: true,                      // Auto-flush queue
  flushInterval: Duration(seconds: 30), // Flush interval
)
```

## Error Handling

```dart
final response = await Afftok.instance.trackClick(params);

if (response.success) {
  // Event tracked successfully
  final data = response.data;
  print('Data: $data');
} else if (response.error != null) {
  // Error occurred
  print('Error: ${response.error}');
} else if (response.message != null) {
  // Event queued for retry
  print('Queued: ${response.message}');
}
```

## Widget Integration

```dart
class TrackingButton extends StatefulWidget {
  @override
  _TrackingButtonState createState() => _TrackingButtonState();
}

class _TrackingButtonState extends State<TrackingButton> {
  bool _isTracking = false;

  Future<void> _trackClick() async {
    setState(() => _isTracking = true);
    
    final response = await Afftok.instance.trackClick(
      ClickParams(offerId: 'offer_123'),
    );
    
    setState(() => _isTracking = false);
    
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(response.success ? 'Tracked!' : 'Error')),
    );
  }

  @override
  Widget build(BuildContext context) {
    return ElevatedButton(
      onPressed: _isTracking ? null : _trackClick,
      child: _isTracking 
        ? CircularProgressIndicator() 
        : Text('Track Click'),
    );
  }
}
```

## Provider Integration

```dart
import 'package:provider/provider.dart';

class TrackingProvider extends ChangeNotifier {
  int _pendingCount = 0;
  int get pendingCount => _pendingCount;

  Future<void> trackClick(String offerId) async {
    final response = await Afftok.instance.trackClick(
      ClickParams(offerId: offerId),
    );
    _pendingCount = Afftok.instance.getPendingCount();
    notifyListeners();
  }
}
```

## Best Practices

1. **Initialize Early**: Initialize in `main()` before `runApp()`
2. **Use Debug Mode**: Enable during development
3. **Handle Responses**: Always check response status
4. **Unique Transaction IDs**: Use unique IDs for conversions
5. **Shutdown Properly**: Call `Afftok.instance.shutdown()` on app exit

## Platform-Specific Setup

### Android

Add to `AndroidManifest.xml`:

```xml
<uses-permission android:name="android.permission.INTERNET" />
```

### iOS

No additional setup required.

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- Issues: https://github.com/afftok/flutter-sdk/issues

## License

MIT License - see LICENSE file for details.

