# Flutter SDK

Official AffTok SDK for Flutter applications.

## Requirements

- Flutter 2.0+
- Dart 2.12+
- iOS 13.0+ / Android API 21+

## Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  afftok_sdk: ^1.0.0
```

Then run:

```bash
flutter pub get
```

## Initialization

Initialize the SDK early in your app:

```dart
import 'package:afftok_sdk/afftok_sdk.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  await Afftok.init(
    apiKey: 'afftok_live_sk_your_api_key_here',
    advertiserId: 'adv_123456',
    userId: 'user_789',  // Optional
    debugMode: kDebugMode,
    offlineQueueEnabled: true,
    maxOfflineEvents: 1000,
    flushIntervalSeconds: 30,
  );
  
  runApp(MyApp());
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | String | Required | Your API key |
| `advertiserId` | String | Required | Your advertiser ID |
| `userId` | String? | null | Your internal user ID |
| `debugMode` | bool | false | Enable verbose logging |
| `offlineQueueEnabled` | bool | true | Enable offline queue |
| `maxOfflineEvents` | int | 1000 | Max queued events |
| `flushIntervalSeconds` | int | 30 | Queue flush interval |
| `baseUrl` | String? | null | Custom API endpoint |

## Track Click

Track when a user clicks on an offer:

```dart
import 'package:afftok_sdk/afftok_sdk.dart';

// Basic click tracking
await Afftok.trackClick(
  offerId: 'off_123456',
  trackingCode: 'abc123.1701432000.xyz.sig',
);

// With metadata
try {
  final result = await Afftok.trackClick(
    offerId: 'off_123456',
    trackingCode: 'abc123.1701432000.xyz.sig',
    metadata: {
      'source': 'email_campaign',
      'creative': 'banner_a',
    },
  );
  print('Click tracked: ${result.clickId}');
} on AfftokException catch (e) {
  print('Error: ${e.code} - ${e.message}');
}
```

### Click Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `offerId` | String | Yes | Offer identifier |
| `trackingCode` | String? | No | Signed tracking code |
| `subId` | String? | No | Sub-affiliate ID |
| `metadata` | Map<String, dynamic>? | No | Custom key-value pairs |

## Track Conversion

Track when a user completes a conversion:

```dart
import 'package:afftok_sdk/afftok_sdk.dart';

// Basic conversion
await Afftok.trackConversion(
  offerId: 'off_123456',
  amount: 49.99,
  currency: 'USD',
);

// Full conversion with all options
try {
  final result = await Afftok.trackConversion(
    offerId: 'off_123456',
    amount: 49.99,
    currency: 'USD',
    orderId: 'order_12345',
    productId: 'prod_premium',
    metadata: {
      'plan': 'annual',
      'coupon': 'SAVE20',
    },
  );
  print('Conversion tracked: ${result.conversionId}');
} on AfftokException catch (e) {
  print('Error: ${e.code} - ${e.message}');
}
```

### Conversion Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `offerId` | String | Yes | Offer identifier |
| `amount` | double | Yes | Conversion value |
| `currency` | String | No | ISO currency code (default: USD) |
| `orderId` | String? | No | Your order/transaction ID |
| `productId` | String? | No | Product identifier |
| `metadata` | Map<String, dynamic>? | No | Custom key-value pairs |

## Deep Link Handling

### Using uni_links

```yaml
dependencies:
  uni_links: ^0.5.1
```

```dart
import 'package:uni_links/uni_links.dart';
import 'package:afftok_sdk/afftok_sdk.dart';

class MyApp extends StatefulWidget {
  @override
  _MyAppState createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  StreamSubscription? _linkSubscription;
  
  @override
  void initState() {
    super.initState();
    _initDeepLinks();
  }
  
  Future<void> _initDeepLinks() async {
    // Handle initial link
    try {
      final initialLink = await getInitialLink();
      if (initialLink != null) {
        _handleDeepLink(initialLink);
      }
    } catch (e) {
      print('Error getting initial link: $e');
    }
    
    // Handle incoming links
    _linkSubscription = linkStream.listen((String? link) {
      if (link != null) {
        _handleDeepLink(link);
      }
    });
  }
  
  void _handleDeepLink(String link) {
    final uri = Uri.parse(link);
    final trackingCode = uri.queryParameters['ref'];
    final offerId = uri.queryParameters['offer'];
    
    if (trackingCode != null && offerId != null) {
      Afftok.trackClick(
        offerId: offerId,
        trackingCode: trackingCode,
        metadata: {'source': 'deep_link'},
      );
    }
  }
  
  @override
  void dispose() {
    _linkSubscription?.cancel();
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return MaterialApp(home: HomeScreen());
  }
}
```

### Using go_router

```dart
import 'package:go_router/go_router.dart';
import 'package:afftok_sdk/afftok_sdk.dart';

final router = GoRouter(
  routes: [
    GoRoute(
      path: '/offer/:id',
      builder: (context, state) {
        final offerId = state.pathParameters['id']!;
        final trackingCode = state.uri.queryParameters['ref'];
        
        // Track click from deep link
        if (trackingCode != null) {
          Afftok.trackClick(
            offerId: offerId,
            trackingCode: trackingCode,
          );
        }
        
        return OfferScreen(offerId: offerId);
      },
    ),
  ],
);
```

## Provider Integration

Use with Provider for state management:

```dart
import 'package:provider/provider.dart';
import 'package:afftok_sdk/afftok_sdk.dart';

class TrackingProvider extends ChangeNotifier {
  bool _isTracking = false;
  String? _lastClickId;
  String? _lastConversionId;
  
  bool get isTracking => _isTracking;
  String? get lastClickId => _lastClickId;
  String? get lastConversionId => _lastConversionId;
  
  Future<void> trackClick({
    required String offerId,
    String? trackingCode,
    Map<String, dynamic>? metadata,
  }) async {
    _isTracking = true;
    notifyListeners();
    
    try {
      final result = await Afftok.trackClick(
        offerId: offerId,
        trackingCode: trackingCode,
        metadata: metadata,
      );
      _lastClickId = result.clickId;
    } catch (e) {
      print('Click tracking error: $e');
    } finally {
      _isTracking = false;
      notifyListeners();
    }
  }
  
  Future<void> trackConversion({
    required String offerId,
    required double amount,
    String currency = 'USD',
    String? orderId,
  }) async {
    _isTracking = true;
    notifyListeners();
    
    try {
      final result = await Afftok.trackConversion(
        offerId: offerId,
        amount: amount,
        currency: currency,
        orderId: orderId,
      );
      _lastConversionId = result.conversionId;
    } catch (e) {
      print('Conversion tracking error: $e');
    } finally {
      _isTracking = false;
      notifyListeners();
    }
  }
}

// Usage
void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Afftok.init(apiKey: '...', advertiserId: '...');
  
  runApp(
    ChangeNotifierProvider(
      create: (_) => TrackingProvider(),
      child: MyApp(),
    ),
  );
}

class OfferScreen extends StatelessWidget {
  final String offerId;
  
  const OfferScreen({required this.offerId});
  
  @override
  Widget build(BuildContext context) {
    final tracking = context.watch<TrackingProvider>();
    
    return Scaffold(
      body: Column(
        children: [
          ElevatedButton(
            onPressed: tracking.isTracking
                ? null
                : () => tracking.trackConversion(
                    offerId: offerId,
                    amount: 49.99,
                  ),
            child: tracking.isTracking
                ? CircularProgressIndicator()
                : Text('Purchase'),
          ),
        ],
      ),
    );
  }
}
```

## Riverpod Integration

```dart
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:afftok_sdk/afftok_sdk.dart';

// Click tracking provider
final clickTrackingProvider = FutureProvider.family<ClickResult, ClickParams>(
  (ref, params) async {
    return await Afftok.trackClick(
      offerId: params.offerId,
      trackingCode: params.trackingCode,
      metadata: params.metadata,
    );
  },
);

class ClickParams {
  final String offerId;
  final String? trackingCode;
  final Map<String, dynamic>? metadata;
  
  ClickParams({
    required this.offerId,
    this.trackingCode,
    this.metadata,
  });
}

// Usage
class OfferScreen extends ConsumerWidget {
  final String offerId;
  final String? trackingCode;
  
  const OfferScreen({required this.offerId, this.trackingCode});
  
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    // Track click on first build
    useEffect(() {
      if (trackingCode != null) {
        ref.read(clickTrackingProvider(ClickParams(
          offerId: offerId,
          trackingCode: trackingCode,
        )));
      }
      return null;
    }, []);
    
    return Scaffold(
      body: Center(child: Text('Offer: $offerId')),
    );
  }
}
```

## Offline Queue

The SDK automatically queues events when offline:

```dart
// Check queue status
final queueSize = await Afftok.getQueueSize();
print('Pending events: $queueSize');

// Force flush queue
final result = await Afftok.flushQueue();
print('Sent: ${result.sent}, Failed: ${result.failed}');

// Clear queue
await Afftok.clearQueue();
```

## Device Fingerprinting

The SDK automatically generates a device fingerprint:

```dart
// Get current fingerprint
final fingerprint = await Afftok.getDeviceFingerprint();
print('Fingerprint: $fingerprint');
```

Fingerprint components:
- Platform-specific device ID
- Device model
- Screen resolution
- Timezone
- Language
- App version

## User Identification

```dart
// Set user ID
await Afftok.setUserId('user_new_id');

// Get current user ID
final userId = await Afftok.getUserId();

// Clear user ID (on logout)
await Afftok.clearUserId();
```

## Error Handling

```dart
try {
  await Afftok.trackClick(offerId: 'off_123456');
} on AfftokException catch (e) {
  switch (e.code) {
    case AfftokErrorCode.networkError:
      // Event queued for retry
      print('Offline, event queued');
      break;
    case AfftokErrorCode.invalidApiKey:
      // Check API key configuration
      print('Invalid API key');
      break;
    case AfftokErrorCode.rateLimited:
      // Too many requests
      print('Rate limited, retry later');
      break;
    case AfftokErrorCode.validationError:
      // Invalid parameters
      print('Validation: ${e.message}');
      break;
    default:
      print('Error: ${e.message}');
  }
} catch (e) {
  print('Unexpected error: $e');
}
```

### Error Types

```dart
enum AfftokErrorCode {
  networkError,
  invalidApiKey,
  rateLimited,
  validationError,
  serverError,
  unknown,
}

class AfftokException implements Exception {
  final AfftokErrorCode code;
  final String message;
  final Map<String, dynamic>? details;
  
  AfftokException({
    required this.code,
    required this.message,
    this.details,
  });
}
```

## Testing

### Test Mode

```dart
await Afftok.init(
  apiKey: 'afftok_test_sk_...',  // Test key
  advertiserId: 'adv_test_123',
  debugMode: true,
);
```

### Mock Events

```dart
// Only in debug mode
if (kDebugMode) {
  await Afftok.generateTestClick();
  await Afftok.generateTestConversion(amount: 99.99);
}
```

## Complete Example

```dart
import 'package:flutter/material.dart';
import 'package:afftok_sdk/afftok_sdk.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  await Afftok.init(
    apiKey: 'afftok_live_sk_your_key',
    advertiserId: 'adv_123456',
    debugMode: false,
  );
  
  runApp(MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: OfferScreen(
        offerId: 'off_123456',
        trackingCode: 'abc123.1701432000.xyz.sig',
      ),
    );
  }
}

class OfferScreen extends StatefulWidget {
  final String offerId;
  final String? trackingCode;
  
  const OfferScreen({
    required this.offerId,
    this.trackingCode,
  });
  
  @override
  _OfferScreenState createState() => _OfferScreenState();
}

class _OfferScreenState extends State<OfferScreen> {
  bool _isPurchasing = false;
  
  @override
  void initState() {
    super.initState();
    _trackOfferView();
  }
  
  Future<void> _trackOfferView() async {
    if (widget.trackingCode != null) {
      try {
        await Afftok.trackClick(
          offerId: widget.offerId,
          trackingCode: widget.trackingCode,
          metadata: {'screen': 'offer_detail'},
        );
      } catch (e) {
        print('Click tracking error: $e');
      }
    }
  }
  
  Future<void> _purchase() async {
    setState(() => _isPurchasing = true);
    
    final orderId = 'order_${DateTime.now().millisecondsSinceEpoch}';
    
    try {
      await Afftok.trackConversion(
        offerId: widget.offerId,
        amount: 49.99,
        currency: 'USD',
        orderId: orderId,
      );
      
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Purchase complete!')),
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e')),
      );
    } finally {
      setState(() => _isPurchasing = false);
    }
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Special Offer')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              'Premium Plan',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
            SizedBox(height: 16),
            Text('\$49.99', style: TextStyle(fontSize: 32)),
            SizedBox(height: 32),
            ElevatedButton(
              onPressed: _isPurchasing ? null : _purchase,
              child: _isPurchasing
                  ? SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text('Purchase Now'),
            ),
          ],
        ),
      ),
    );
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
- Provider/Riverpod compatible

---

Next: [React Native SDK](./react-native.md)

