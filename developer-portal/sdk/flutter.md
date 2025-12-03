# Flutter SDK

Complete guide for integrating AffTok tracking into Flutter applications.

---

## Requirements

- Flutter 3.0+
- Dart 2.17+
- iOS 12.0+ / Android SDK 21+

---

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

---

## Initialization

Initialize the SDK in your `main.dart`:

```dart
import 'package:afftok_sdk/afftok_sdk.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  await Afftok.init(
    apiKey: 'afftok_live_sk_xxxxx',
    config: AfftokConfig(
      debug: kDebugMode,
      enableOfflineQueue: true,
      maxQueueSize: 1000,
      flushIntervalMs: 30000,
      timeoutMs: 30000,
      retryAttempts: 3,
    ),
  );
  
  runApp(MyApp());
}
```

### Configuration Options

```dart
class AfftokConfig {
  final bool debug;
  final bool enableOfflineQueue;
  final int maxQueueSize;
  final int flushIntervalMs;
  final int timeoutMs;
  final int retryAttempts;
  final String? baseUrl;
  final bool enableDeviceFingerprint;
  
  const AfftokConfig({
    this.debug = false,
    this.enableOfflineQueue = true,
    this.maxQueueSize = 1000,
    this.flushIntervalMs = 30000,
    this.timeoutMs = 30000,
    this.retryAttempts = 3,
    this.baseUrl,
    this.enableDeviceFingerprint = true,
  });
}
```

---

## Click Tracking

### Basic Click Tracking

```dart
await Afftok.trackClick(trackingCode: 'tc_abc123def456');
```

### Click with Metadata

```dart
await Afftok.trackClick(
  trackingCode: 'tc_abc123def456',
  metadata: {
    'campaign': 'summer_sale',
    'source': 'push_notification',
    'sub_id': 'user_123',
  },
);
```

### Click with Result Handling

```dart
try {
  final result = await Afftok.trackClick(
    trackingCode: 'tc_abc123def456',
    metadata: {'campaign': 'summer_sale'},
  );
  print('Click tracked: ${result.clickId}');
} on AfftokException catch (e) {
  print('Error: ${e.message}');
}
```

---

## Conversion Tracking

### Basic Conversion

```dart
await Afftok.trackConversion(
  offerId: 'off_xyz789',
  transactionId: 'txn_${DateTime.now().millisecondsSinceEpoch}',
);
```

### Conversion with Amount

```dart
await Afftok.trackConversion(
  offerId: 'off_xyz789',
  transactionId: 'order_12345',
  amount: 49.99,
  currency: 'USD',
);
```

### Full Conversion Tracking

```dart
try {
  final result = await Afftok.trackConversion(
    offerId: 'off_xyz789',
    transactionId: 'order_12345',
    clickId: 'clk_abc123',  // Optional: link to specific click
    amount: 49.99,
    currency: 'USD',
    status: ConversionStatus.approved,
    metadata: {
      'product_id': 'prod_456',
      'category': 'electronics',
      'user_type': 'new',
    },
  );
  print('Conversion tracked: ${result.conversionId}');
} on AfftokException catch (e) {
  print('Error: ${e.message}');
}
```

### Conversion Status

```dart
enum ConversionStatus {
  pending,
  approved,
  rejected,
}
```

---

## Device Fingerprinting

The SDK automatically generates a unique device fingerprint:

```dart
// Get current device fingerprint
final fingerprint = await Afftok.getDeviceFingerprint();
print('Device Fingerprint: $fingerprint');
```

### Custom Fingerprint Provider

```dart
Afftok.setFingerprintProvider(() async {
  // Custom fingerprint logic
  final deviceInfo = await DeviceInfoPlugin().deviceInfo;
  return 'custom_${deviceInfo.data['id']}';
});
```

---

## Offline Queue

Events are automatically queued when offline:

```dart
// Check queue status
final queueSize = await Afftok.getQueueSize();
print('Queued events: $queueSize');

// Force flush queue
try {
  final result = await Afftok.flushQueue();
  print('Queue flushed: ${result.flushedCount} events');
} on AfftokException catch (e) {
  print('Flush failed: ${e.message}');
}

// Clear queue
await Afftok.clearQueue();
```

---

## Deep Link Handling

### Using uni_links

```dart
import 'package:uni_links/uni_links.dart';

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
        await Afftok.handleDeepLink(Uri.parse(initialLink));
      }
    } catch (e) {
      print('Error getting initial link: $e');
    }
    
    // Handle incoming links
    _linkSubscription = linkStream.listen((String? link) async {
      if (link != null) {
        try {
          final result = await Afftok.handleDeepLink(Uri.parse(link));
          print('Deep link tracked: $result');
        } on AfftokException catch (e) {
          print('Error: ${e.message}');
        }
      }
    });
  }
  
  @override
  void dispose() {
    _linkSubscription?.cancel();
    super.dispose();
  }
}
```

### Using go_router

```dart
import 'package:go_router/go_router.dart';

final router = GoRouter(
  routes: [
    GoRoute(
      path: '/c/:trackingCode',
      redirect: (context, state) async {
        final trackingCode = state.pathParameters['trackingCode'];
        if (trackingCode != null) {
          await Afftok.trackClick(trackingCode: trackingCode);
        }
        return '/home';
      },
    ),
  ],
);
```

### Generate Deep Link

```dart
final deepLink = Afftok.generateDeepLink(
  trackingCode: 'tc_abc123',
  scheme: 'myapp',
  path: '/offer/details',
);
// Result: myapp://offer/details?tc=tc_abc123
```

---

## Push Notification Attribution

### Firebase Cloud Messaging

```dart
import 'package:firebase_messaging/firebase_messaging.dart';

class PushNotificationService {
  final FirebaseMessaging _fcm = FirebaseMessaging.instance;
  
  Future<void> initialize() async {
    // Handle foreground messages
    FirebaseMessaging.onMessage.listen(_handleMessage);
    
    // Handle background/terminated messages
    FirebaseMessaging.onMessageOpenedApp.listen(_handleMessage);
    
    // Check for initial message
    final initialMessage = await _fcm.getInitialMessage();
    if (initialMessage != null) {
      _handleMessage(initialMessage);
    }
  }
  
  void _handleMessage(RemoteMessage message) {
    final trackingCode = message.data['afftok_tc'];
    if (trackingCode != null) {
      Afftok.trackClick(
        trackingCode: trackingCode,
        metadata: {
          'source': 'push_notification',
          'campaign': message.data['campaign'] ?? '',
        },
      );
    }
  }
}
```

---

## User Identification

Associate events with a user:

```dart
// Set user ID
await Afftok.setUserId('user_12345');

// Set user properties
await Afftok.setUserProperties({
  'email': 'user@example.com',
  'plan': 'premium',
  'signup_date': '2024-01-15',
});

// Clear user (on logout)
await Afftok.clearUser();
```

---

## Error Handling

```dart
abstract class AfftokException implements Exception {
  String get message;
}

class NetworkException extends AfftokException {
  @override
  final String message;
  NetworkException(this.message);
}

class AuthException extends AfftokException {
  @override
  final String message;
  AuthException(this.message);
}

class ValidationException extends AfftokException {
  @override
  final String message;
  ValidationException(this.message);
}

class RateLimitException extends AfftokException {
  @override
  final String message;
  final Duration retryAfter;
  RateLimitException(this.message, this.retryAfter);
}

class ServerException extends AfftokException {
  @override
  final String message;
  final int statusCode;
  ServerException(this.message, this.statusCode);
}

// Handle errors
try {
  await Afftok.trackClick(trackingCode: 'tc_abc123');
} on NetworkException catch (e) {
  // Event queued for retry
  print('Network error (event queued): ${e.message}');
} on RateLimitException catch (e) {
  print('Rate limited, retry after ${e.retryAfter.inSeconds}s');
} on AuthException catch (e) {
  print('Auth error: ${e.message}');
} on AfftokException catch (e) {
  print('Error: ${e.message}');
}
```

---

## State Management Integration

### Provider

```dart
import 'package:provider/provider.dart';

class TrackingProvider extends ChangeNotifier {
  bool _isTracking = false;
  String? _lastClickId;
  
  bool get isTracking => _isTracking;
  String? get lastClickId => _lastClickId;
  
  Future<void> trackClick(String trackingCode) async {
    _isTracking = true;
    notifyListeners();
    
    try {
      final result = await Afftok.trackClick(trackingCode: trackingCode);
      _lastClickId = result.clickId;
    } finally {
      _isTracking = false;
      notifyListeners();
    }
  }
}

// Usage
Consumer<TrackingProvider>(
  builder: (context, provider, child) {
    return ElevatedButton(
      onPressed: provider.isTracking
          ? null
          : () => provider.trackClick('tc_abc123'),
      child: provider.isTracking
          ? CircularProgressIndicator()
          : Text('Track Click'),
    );
  },
)
```

### Riverpod

```dart
import 'package:flutter_riverpod/flutter_riverpod.dart';

final trackingProvider = FutureProvider.family<ClickResult, String>(
  (ref, trackingCode) async {
    return await Afftok.trackClick(trackingCode: trackingCode);
  },
);

// Usage
Consumer(
  builder: (context, ref, child) {
    final tracking = ref.watch(trackingProvider('tc_abc123'));
    
    return tracking.when(
      data: (result) => Text('Tracked: ${result.clickId}'),
      loading: () => CircularProgressIndicator(),
      error: (error, stack) => Text('Error: $error'),
    );
  },
)
```

### BLoC

```dart
import 'package:flutter_bloc/flutter_bloc.dart';

// Events
abstract class TrackingEvent {}

class TrackClickEvent extends TrackingEvent {
  final String trackingCode;
  TrackClickEvent(this.trackingCode);
}

// States
abstract class TrackingState {}

class TrackingInitial extends TrackingState {}
class TrackingLoading extends TrackingState {}
class TrackingSuccess extends TrackingState {
  final String clickId;
  TrackingSuccess(this.clickId);
}
class TrackingError extends TrackingState {
  final String message;
  TrackingError(this.message);
}

// BLoC
class TrackingBloc extends Bloc<TrackingEvent, TrackingState> {
  TrackingBloc() : super(TrackingInitial()) {
    on<TrackClickEvent>(_onTrackClick);
  }
  
  Future<void> _onTrackClick(
    TrackClickEvent event,
    Emitter<TrackingState> emit,
  ) async {
    emit(TrackingLoading());
    try {
      final result = await Afftok.trackClick(trackingCode: event.trackingCode);
      emit(TrackingSuccess(result.clickId));
    } on AfftokException catch (e) {
      emit(TrackingError(e.message));
    }
  }
}
```

---

## Debug Mode

Enable detailed logging:

```dart
AfftokConfig(debug: true)

// Or at runtime
Afftok.setDebugMode(true);
```

Debug logs include:
- All API requests/responses
- Queue operations
- Fingerprint generation
- Error details

---

## Complete Example

```dart
import 'package:flutter/material.dart';
import 'package:afftok_sdk/afftok_sdk.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  await Afftok.init(
    apiKey: 'afftok_live_sk_xxxxx',
    config: AfftokConfig(debug: true),
  );
  
  runApp(MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: SubscriptionScreen(),
    );
  }
}

class SubscriptionScreen extends StatefulWidget {
  @override
  _SubscriptionScreenState createState() => _SubscriptionScreenState();
}

class _SubscriptionScreenState extends State<SubscriptionScreen> {
  bool _isPurchasing = false;
  
  @override
  void initState() {
    super.initState();
    // Track screen view
    Afftok.trackEvent('screen_view', metadata: {'screen': 'subscription'});
  }
  
  Future<void> _completePurchase() async {
    setState(() => _isPurchasing = true);
    
    try {
      // Your purchase logic here...
      
      // Track conversion
      final result = await Afftok.trackConversion(
        offerId: 'off_premium_subscription',
        transactionId: 'txn_${DateTime.now().millisecondsSinceEpoch}',
        amount: 9.99,
        currency: 'USD',
        metadata: {
          'product': 'premium_monthly',
          'user_id': await _getCurrentUserId(),
        },
      );
      
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Purchase tracked: ${result.conversionId}')),
      );
    } on AfftokException catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Tracking failed: ${e.message}')),
      );
    } finally {
      setState(() => _isPurchasing = false);
    }
  }
  
  Future<String> _getCurrentUserId() async {
    // Return current user ID
    return 'user_12345';
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Premium Subscription')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              '\$9.99/month',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
            SizedBox(height: 20),
            ElevatedButton(
              onPressed: _isPurchasing ? null : _completePurchase,
              child: _isPurchasing
                  ? SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text('Subscribe Now'),
            ),
          ],
        ),
      ),
    );
  }
}
```

---

## Platform-Specific Setup

### Android

Add to `android/app/src/main/AndroidManifest.xml`:

```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

### iOS

No additional setup required.

---

## Next Steps

- [React Native SDK](react-native.md) - React Native integration guide
- [Web SDK](web.md) - Web integration guide
- [API Reference](../api-reference/clicks.md) - Full API documentation

