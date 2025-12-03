# React Native SDK

Complete guide for integrating AffTok tracking into React Native applications.

---

## Requirements

- React Native 0.60+
- iOS 12.0+ / Android SDK 21+
- Node.js 14+

---

## Installation

```bash
npm install @afftok/react-native
# or
yarn add @afftok/react-native
```

### iOS Setup

```bash
cd ios && pod install
```

### Android Setup

No additional setup required.

---

## Initialization

Initialize the SDK in your app entry point:

```javascript
import Afftok from '@afftok/react-native';

// In App.js or index.js
Afftok.init({
  apiKey: 'afftok_live_sk_xxxxx',
  debug: __DEV__,
  enableOfflineQueue: true,
  maxQueueSize: 1000,
  flushIntervalMs: 30000,
  timeoutMs: 30000,
  retryAttempts: 3,
});
```

### TypeScript Support

```typescript
import Afftok, { AfftokConfig } from '@afftok/react-native';

const config: AfftokConfig = {
  apiKey: 'afftok_live_sk_xxxxx',
  debug: __DEV__,
  enableOfflineQueue: true,
};

Afftok.init(config);
```

### Configuration Options

```typescript
interface AfftokConfig {
  apiKey: string;
  debug?: boolean;
  enableOfflineQueue?: boolean;
  maxQueueSize?: number;
  flushIntervalMs?: number;
  timeoutMs?: number;
  retryAttempts?: number;
  baseUrl?: string;
  enableDeviceFingerprint?: boolean;
}
```

---

## Click Tracking

### Basic Click Tracking

```javascript
await Afftok.trackClick('tc_abc123def456');
```

### Click with Metadata

```javascript
await Afftok.trackClick('tc_abc123def456', {
  campaign: 'summer_sale',
  source: 'push_notification',
  sub_id: 'user_123',
});
```

### Click with Result Handling

```javascript
try {
  const result = await Afftok.trackClick('tc_abc123def456', {
    campaign: 'summer_sale',
  });
  console.log('Click tracked:', result.clickId);
} catch (error) {
  console.error('Error:', error.message);
}
```

---

## Conversion Tracking

### Basic Conversion

```javascript
await Afftok.trackConversion({
  offerId: 'off_xyz789',
  transactionId: `txn_${Date.now()}`,
});
```

### Conversion with Amount

```javascript
await Afftok.trackConversion({
  offerId: 'off_xyz789',
  transactionId: 'order_12345',
  amount: 49.99,
  currency: 'USD',
});
```

### Full Conversion Tracking

```javascript
try {
  const result = await Afftok.trackConversion({
    offerId: 'off_xyz789',
    transactionId: 'order_12345',
    clickId: 'clk_abc123', // Optional: link to specific click
    amount: 49.99,
    currency: 'USD',
    status: 'approved',
    metadata: {
      product_id: 'prod_456',
      category: 'electronics',
      user_type: 'new',
    },
  });
  console.log('Conversion tracked:', result.conversionId);
} catch (error) {
  console.error('Error:', error.message);
}
```

### Conversion Status

```typescript
type ConversionStatus = 'pending' | 'approved' | 'rejected';
```

---

## Device Fingerprinting

The SDK automatically generates a unique device fingerprint:

```javascript
// Get current device fingerprint
const fingerprint = await Afftok.getDeviceFingerprint();
console.log('Device Fingerprint:', fingerprint);
```

### Custom Fingerprint Provider

```javascript
Afftok.setFingerprintProvider(async () => {
  // Custom fingerprint logic
  const deviceId = await getDeviceId();
  return `custom_${deviceId}`;
});
```

---

## Offline Queue

Events are automatically queued when offline:

```javascript
// Check queue status
const queueSize = await Afftok.getQueueSize();
console.log('Queued events:', queueSize);

// Force flush queue
try {
  const result = await Afftok.flushQueue();
  console.log('Queue flushed:', result.flushedCount, 'events');
} catch (error) {
  console.error('Flush failed:', error.message);
}

// Clear queue
await Afftok.clearQueue();
```

---

## Deep Link Handling

### React Navigation

```javascript
import { Linking } from 'react-native';
import { NavigationContainer } from '@react-navigation/native';

function App() {
  const linking = {
    prefixes: ['myapp://', 'https://myapp.com'],
    config: {
      screens: {
        Home: 'home',
        Offer: 'c/:trackingCode',
      },
    },
  };

  return (
    <NavigationContainer
      linking={linking}
      onStateChange={(state) => {
        // Handle deep link tracking
        const route = state?.routes[state.index];
        if (route?.name === 'Offer' && route.params?.trackingCode) {
          Afftok.trackClick(route.params.trackingCode);
        }
      }}
    >
      {/* Your screens */}
    </NavigationContainer>
  );
}
```

### Manual Deep Link Handling

```javascript
import { Linking } from 'react-native';

useEffect(() => {
  // Handle initial URL
  Linking.getInitialURL().then((url) => {
    if (url) {
      handleDeepLink(url);
    }
  });

  // Handle incoming URLs
  const subscription = Linking.addEventListener('url', ({ url }) => {
    handleDeepLink(url);
  });

  return () => {
    subscription.remove();
  };
}, []);

async function handleDeepLink(url) {
  try {
    const result = await Afftok.handleDeepLink(url);
    console.log('Deep link tracked:', result);
  } catch (error) {
    console.error('Error:', error.message);
  }
}
```

### Generate Deep Link

```javascript
const deepLink = Afftok.generateDeepLink({
  trackingCode: 'tc_abc123',
  scheme: 'myapp',
  path: '/offer/details',
});
// Result: myapp://offer/details?tc=tc_abc123
```

---

## Push Notification Attribution

### Firebase Cloud Messaging

```javascript
import messaging from '@react-native-firebase/messaging';

useEffect(() => {
  // Handle foreground messages
  const unsubscribe = messaging().onMessage(async (remoteMessage) => {
    handlePushMessage(remoteMessage);
  });

  // Handle background/quit messages
  messaging().onNotificationOpenedApp((remoteMessage) => {
    handlePushMessage(remoteMessage);
  });

  // Check for initial notification
  messaging()
    .getInitialNotification()
    .then((remoteMessage) => {
      if (remoteMessage) {
        handlePushMessage(remoteMessage);
      }
    });

  return unsubscribe;
}, []);

function handlePushMessage(remoteMessage) {
  const trackingCode = remoteMessage.data?.afftok_tc;
  if (trackingCode) {
    Afftok.trackClick(trackingCode, {
      source: 'push_notification',
      campaign: remoteMessage.data?.campaign || '',
    });
  }
}
```

---

## User Identification

Associate events with a user:

```javascript
// Set user ID
await Afftok.setUserId('user_12345');

// Set user properties
await Afftok.setUserProperties({
  email: 'user@example.com',
  plan: 'premium',
  signup_date: '2024-01-15',
});

// Clear user (on logout)
await Afftok.clearUser();
```

---

## Error Handling

```typescript
interface AfftokError extends Error {
  code: string;
  message: string;
  retryAfter?: number;
}

// Error codes
const ErrorCodes = {
  NETWORK_ERROR: 'NETWORK_ERROR',
  AUTH_ERROR: 'AUTH_ERROR',
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  RATE_LIMIT_ERROR: 'RATE_LIMIT_ERROR',
  SERVER_ERROR: 'SERVER_ERROR',
};
```

```javascript
try {
  await Afftok.trackClick('tc_abc123');
} catch (error) {
  switch (error.code) {
    case 'NETWORK_ERROR':
      console.log('Network error (event queued):', error.message);
      break;
    case 'RATE_LIMIT_ERROR':
      console.log(`Rate limited, retry after ${error.retryAfter}ms`);
      break;
    case 'AUTH_ERROR':
      console.log('Auth error:', error.message);
      break;
    default:
      console.log('Error:', error.message);
  }
}
```

---

## React Hooks

### useAfftok Hook

```javascript
import { useState, useCallback } from 'react';
import Afftok from '@afftok/react-native';

function useAfftok() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  const trackClick = useCallback(async (trackingCode, metadata) => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await Afftok.trackClick(trackingCode, metadata);
      return result;
    } catch (e) {
      setError(e);
      throw e;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const trackConversion = useCallback(async (options) => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await Afftok.trackConversion(options);
      return result;
    } catch (e) {
      setError(e);
      throw e;
    } finally {
      setIsLoading(false);
    }
  }, []);

  return {
    isLoading,
    error,
    trackClick,
    trackConversion,
  };
}

// Usage
function MyComponent() {
  const { isLoading, error, trackClick } = useAfftok();

  const handlePress = async () => {
    try {
      await trackClick('tc_abc123', { campaign: 'summer_sale' });
    } catch (e) {
      console.error('Tracking failed:', e.message);
    }
  };

  return (
    <TouchableOpacity onPress={handlePress} disabled={isLoading}>
      <Text>{isLoading ? 'Tracking...' : 'Track Click'}</Text>
    </TouchableOpacity>
  );
}
```

### useDeepLink Hook

```javascript
import { useEffect } from 'react';
import { Linking } from 'react-native';
import Afftok from '@afftok/react-native';

function useDeepLink(onDeepLink) {
  useEffect(() => {
    // Handle initial URL
    Linking.getInitialURL().then((url) => {
      if (url) {
        handleUrl(url);
      }
    });

    // Handle incoming URLs
    const subscription = Linking.addEventListener('url', ({ url }) => {
      handleUrl(url);
    });

    return () => {
      subscription.remove();
    };
  }, []);

  async function handleUrl(url) {
    try {
      const result = await Afftok.handleDeepLink(url);
      onDeepLink?.(result);
    } catch (error) {
      console.error('Deep link error:', error.message);
    }
  }
}

// Usage
function App() {
  useDeepLink((result) => {
    console.log('Deep link handled:', result);
  });

  return <MainNavigator />;
}
```

---

## Debug Mode

Enable detailed logging:

```javascript
Afftok.init({
  apiKey: 'afftok_live_sk_xxxxx',
  debug: true,
});

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

```javascript
import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import Afftok from '@afftok/react-native';

// Initialize SDK
Afftok.init({
  apiKey: 'afftok_live_sk_xxxxx',
  debug: __DEV__,
  enableOfflineQueue: true,
});

function SubscriptionScreen() {
  const [isPurchasing, setIsPurchasing] = useState(false);

  useEffect(() => {
    // Track screen view
    Afftok.trackEvent('screen_view', { screen: 'subscription' });
  }, []);

  const completePurchase = async () => {
    setIsPurchasing(true);

    try {
      // Your purchase logic here...

      // Track conversion
      const result = await Afftok.trackConversion({
        offerId: 'off_premium_subscription',
        transactionId: `txn_${Date.now()}`,
        amount: 9.99,
        currency: 'USD',
        metadata: {
          product: 'premium_monthly',
          user_id: getCurrentUserId(),
        },
      });

      console.log('Conversion tracked:', result.conversionId);
      alert('Purchase successful!');
    } catch (error) {
      console.error('Tracking failed:', error.message);
      alert('Purchase failed. Please try again.');
    } finally {
      setIsPurchasing(false);
    }
  };

  const getCurrentUserId = () => {
    // Return current user ID
    return 'user_12345';
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Premium Subscription</Text>
      <Text style={styles.price}>$9.99/month</Text>

      <TouchableOpacity
        style={[styles.button, isPurchasing && styles.buttonDisabled]}
        onPress={completePurchase}
        disabled={isPurchasing}
      >
        {isPurchasing ? (
          <ActivityIndicator color="#fff" />
        ) : (
          <Text style={styles.buttonText}>Subscribe Now</Text>
        )}
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  price: {
    fontSize: 20,
    color: '#666',
    marginBottom: 30,
  },
  button: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 30,
    paddingVertical: 15,
    borderRadius: 8,
    minWidth: 200,
    alignItems: 'center',
  },
  buttonDisabled: {
    backgroundColor: '#999',
  },
  buttonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
  },
});

export default SubscriptionScreen;
```

---

## TypeScript Types

```typescript
// Types
interface ClickResult {
  clickId: string;
  trackingCode: string;
  timestamp: string;
}

interface ConversionResult {
  conversionId: string;
  offerId: string;
  transactionId: string;
  amount?: number;
  status: ConversionStatus;
  timestamp: string;
}

interface ConversionOptions {
  offerId: string;
  transactionId: string;
  clickId?: string;
  amount?: number;
  currency?: string;
  status?: ConversionStatus;
  metadata?: Record<string, string>;
}

type ConversionStatus = 'pending' | 'approved' | 'rejected';

interface DeepLinkResult {
  trackingCode: string;
  url: string;
  handled: boolean;
}

interface QueueFlushResult {
  flushedCount: number;
  failedCount: number;
}
```

---

## Next Steps

- [Web SDK](web.md) - Web integration guide
- [Flutter SDK](flutter.md) - Flutter integration guide
- [API Reference](../api-reference/clicks.md) - Full API documentation

