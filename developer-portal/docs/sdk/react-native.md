# React Native SDK

Official AffTok SDK for React Native applications.

## Requirements

- React Native 0.63+
- iOS 13.0+ / Android API 21+

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

## Initialization

Initialize the SDK in your app entry point:

```javascript
import { Afftok } from '@afftok/react-native';

// In App.js or index.js
Afftok.init({
  apiKey: 'afftok_live_sk_your_api_key_here',
  advertiserId: 'adv_123456',
  userId: 'user_789',  // Optional
  debugMode: __DEV__,
  offlineQueueEnabled: true,
  maxOfflineEvents: 1000,
  flushIntervalSeconds: 30,
});

export default function App() {
  return <MainNavigator />;
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | string | Required | Your API key |
| `advertiserId` | string | Required | Your advertiser ID |
| `userId` | string | null | Your internal user ID |
| `debugMode` | boolean | false | Enable verbose logging |
| `offlineQueueEnabled` | boolean | true | Enable offline queue |
| `maxOfflineEvents` | number | 1000 | Max queued events |
| `flushIntervalSeconds` | number | 30 | Queue flush interval |
| `baseUrl` | string | null | Custom API endpoint |

## Track Click

Track when a user clicks on an offer:

```javascript
import { Afftok } from '@afftok/react-native';

// Basic click tracking
await Afftok.trackClick({
  offerId: 'off_123456',
  trackingCode: 'abc123.1701432000.xyz.sig',
});

// With metadata
try {
  const result = await Afftok.trackClick({
    offerId: 'off_123456',
    trackingCode: 'abc123.1701432000.xyz.sig',
    metadata: {
      source: 'email_campaign',
      creative: 'banner_a',
    },
  });
  console.log('Click tracked:', result.clickId);
} catch (error) {
  console.error('Error:', error.code, error.message);
}
```

### Click Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `offerId` | string | Yes | Offer identifier |
| `trackingCode` | string | No | Signed tracking code |
| `subId` | string | No | Sub-affiliate ID |
| `metadata` | object | No | Custom key-value pairs |

## Track Conversion

Track when a user completes a conversion:

```javascript
import { Afftok } from '@afftok/react-native';

// Basic conversion
await Afftok.trackConversion({
  offerId: 'off_123456',
  amount: 49.99,
  currency: 'USD',
});

// Full conversion with all options
try {
  const result = await Afftok.trackConversion({
    offerId: 'off_123456',
    amount: 49.99,
    currency: 'USD',
    orderId: 'order_12345',
    productId: 'prod_premium',
    metadata: {
      plan: 'annual',
      coupon: 'SAVE20',
    },
  });
  console.log('Conversion tracked:', result.conversionId);
} catch (error) {
  console.error('Error:', error.code, error.message);
}
```

### Conversion Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `offerId` | string | Yes | Offer identifier |
| `amount` | number | Yes | Conversion value |
| `currency` | string | No | ISO currency code (default: USD) |
| `orderId` | string | No | Your order/transaction ID |
| `productId` | string | No | Product identifier |
| `metadata` | object | No | Custom key-value pairs |

## Deep Link Handling

### Using React Navigation

```javascript
import { Linking } from 'react-native';
import { NavigationContainer } from '@react-navigation/native';
import { Afftok } from '@afftok/react-native';

const linking = {
  prefixes: ['myapp://', 'https://myapp.com'],
  config: {
    screens: {
      Offer: 'offer/:id',
    },
  },
};

function App() {
  useEffect(() => {
    // Handle initial URL
    Linking.getInitialURL().then((url) => {
      if (url) handleDeepLink(url);
    });

    // Handle incoming URLs
    const subscription = Linking.addEventListener('url', ({ url }) => {
      handleDeepLink(url);
    });

    return () => subscription.remove();
  }, []);

  const handleDeepLink = (url) => {
    const parsed = new URL(url);
    const trackingCode = parsed.searchParams.get('ref');
    const offerId = parsed.searchParams.get('offer');

    if (trackingCode && offerId) {
      Afftok.trackClick({
        offerId,
        trackingCode,
        metadata: { source: 'deep_link' },
      });
    }
  };

  return (
    <NavigationContainer linking={linking}>
      <MainNavigator />
    </NavigationContainer>
  );
}
```

### Using Expo

```javascript
import * as Linking from 'expo-linking';
import { Afftok } from '@afftok/react-native';

function App() {
  useEffect(() => {
    const handleUrl = ({ url }) => {
      const { queryParams } = Linking.parse(url);
      
      if (queryParams.ref && queryParams.offer) {
        Afftok.trackClick({
          offerId: queryParams.offer,
          trackingCode: queryParams.ref,
          metadata: { source: 'deep_link' },
        });
      }
    };

    Linking.getInitialURL().then((url) => {
      if (url) handleUrl({ url });
    });

    const subscription = Linking.addEventListener('url', handleUrl);
    return () => subscription.remove();
  }, []);

  return <MainNavigator />;
}
```

## Hooks

The SDK provides React hooks for easier integration:

```javascript
import { useAfftok, useTrackClick, useTrackConversion } from '@afftok/react-native';

function OfferScreen({ offerId, trackingCode }) {
  const { isInitialized, queueSize } = useAfftok();
  
  const {
    trackClick,
    isLoading: isClickLoading,
    error: clickError,
    result: clickResult,
  } = useTrackClick();
  
  const {
    trackConversion,
    isLoading: isConversionLoading,
    error: conversionError,
    result: conversionResult,
  } = useTrackConversion();

  useEffect(() => {
    if (trackingCode) {
      trackClick({
        offerId,
        trackingCode,
        metadata: { screen: 'offer_detail' },
      });
    }
  }, [offerId, trackingCode]);

  const handlePurchase = async () => {
    await trackConversion({
      offerId,
      amount: 49.99,
      currency: 'USD',
      orderId: `order_${Date.now()}`,
    });
  };

  return (
    <View>
      <Text>Offer: {offerId}</Text>
      <Button
        title="Purchase"
        onPress={handlePurchase}
        disabled={isConversionLoading}
      />
      {conversionResult && <Text>Success!</Text>}
      {conversionError && <Text>Error: {conversionError.message}</Text>}
    </View>
  );
}
```

## Context Provider

Wrap your app with the Afftok provider for global access:

```javascript
import { AfftokProvider, useAfftokContext } from '@afftok/react-native';

function App() {
  return (
    <AfftokProvider
      config={{
        apiKey: 'afftok_live_sk_your_key',
        advertiserId: 'adv_123456',
        debugMode: __DEV__,
      }}
    >
      <MainNavigator />
    </AfftokProvider>
  );
}

function OfferScreen() {
  const { trackClick, trackConversion, isReady } = useAfftokContext();
  
  if (!isReady) {
    return <ActivityIndicator />;
  }
  
  return (
    <View>
      <Button
        title="Track Click"
        onPress={() => trackClick({ offerId: 'off_123' })}
      />
    </View>
  );
}
```

## Offline Queue

The SDK automatically queues events when offline:

```javascript
import { Afftok } from '@afftok/react-native';

// Check queue status
const queueSize = await Afftok.getQueueSize();
console.log('Pending events:', queueSize);

// Force flush queue
const { sent, failed } = await Afftok.flushQueue();
console.log(`Flushed: ${sent} sent, ${failed} failed`);

// Clear queue
await Afftok.clearQueue();
```

### Network Status Handling

```javascript
import NetInfo from '@react-native-community/netinfo';
import { Afftok } from '@afftok/react-native';

function useNetworkAwareTracking() {
  const [isConnected, setIsConnected] = useState(true);

  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener((state) => {
      setIsConnected(state.isConnected);
      
      // Flush queue when coming back online
      if (state.isConnected) {
        Afftok.flushQueue();
      }
    });

    return () => unsubscribe();
  }, []);

  return { isConnected };
}
```

## Device Fingerprinting

The SDK automatically generates a device fingerprint:

```javascript
import { Afftok } from '@afftok/react-native';

// Get current fingerprint
const fingerprint = await Afftok.getDeviceFingerprint();
console.log('Fingerprint:', fingerprint);
```

## User Identification

```javascript
import { Afftok } from '@afftok/react-native';

// Set user ID
await Afftok.setUserId('user_new_id');

// Get current user ID
const userId = await Afftok.getUserId();

// Clear user ID (on logout)
await Afftok.clearUserId();
```

## Error Handling

```javascript
import { Afftok, AfftokError } from '@afftok/react-native';

try {
  await Afftok.trackClick({ offerId: 'off_123456' });
} catch (error) {
  if (error instanceof AfftokError) {
    switch (error.code) {
      case 'NETWORK_ERROR':
        // Event queued for retry
        console.log('Offline, event queued');
        break;
      case 'INVALID_API_KEY':
        // Check API key configuration
        console.error('Invalid API key');
        break;
      case 'RATE_LIMITED':
        // Too many requests
        console.warn('Rate limited, retry later');
        break;
      case 'VALIDATION_ERROR':
        // Invalid parameters
        console.error('Validation:', error.message);
        break;
      default:
        console.error('Error:', error.message);
    }
  }
}
```

## TypeScript Support

The SDK includes TypeScript definitions:

```typescript
import {
  Afftok,
  AfftokConfig,
  ClickParams,
  ConversionParams,
  ClickResult,
  ConversionResult,
  AfftokError,
} from '@afftok/react-native';

const config: AfftokConfig = {
  apiKey: 'afftok_live_sk_your_key',
  advertiserId: 'adv_123456',
  debugMode: true,
};

Afftok.init(config);

const trackClick = async (params: ClickParams): Promise<ClickResult> => {
  return await Afftok.trackClick(params);
};

const trackConversion = async (params: ConversionParams): Promise<ConversionResult> => {
  return await Afftok.trackConversion(params);
};
```

## Testing

### Test Mode

```javascript
Afftok.init({
  apiKey: 'afftok_test_sk_...',  // Test key
  advertiserId: 'adv_test_123',
  debugMode: true,
});
```

### Mock Events

```javascript
if (__DEV__) {
  await Afftok.generateTestClick();
  await Afftok.generateTestConversion({ amount: 99.99 });
}
```

### Jest Mocking

```javascript
// __mocks__/@afftok/react-native.js
export const Afftok = {
  init: jest.fn(),
  trackClick: jest.fn().mockResolvedValue({ clickId: 'test_click' }),
  trackConversion: jest.fn().mockResolvedValue({ conversionId: 'test_conv' }),
  getQueueSize: jest.fn().mockResolvedValue(0),
  flushQueue: jest.fn().mockResolvedValue({ sent: 0, failed: 0 }),
  clearQueue: jest.fn(),
  setUserId: jest.fn(),
  getUserId: jest.fn().mockResolvedValue(null),
  clearUserId: jest.fn(),
  getDeviceFingerprint: jest.fn().mockResolvedValue('test_fingerprint'),
};
```

## Complete Example

```javascript
import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  Button,
  StyleSheet,
  Alert,
  ActivityIndicator,
} from 'react-native';
import { Afftok } from '@afftok/react-native';

// Initialize in app entry
Afftok.init({
  apiKey: 'afftok_live_sk_your_key',
  advertiserId: 'adv_123456',
  debugMode: __DEV__,
});

function OfferScreen({ route }) {
  const { offerId, trackingCode } = route.params;
  const [isPurchasing, setIsPurchasing] = useState(false);

  useEffect(() => {
    // Track offer view
    if (trackingCode) {
      Afftok.trackClick({
        offerId,
        trackingCode,
        metadata: { screen: 'offer_detail' },
      }).catch((error) => {
        console.warn('Click tracking error:', error);
      });
    }
  }, [offerId, trackingCode]);

  const handlePurchase = async () => {
    setIsPurchasing(true);
    const orderId = `order_${Date.now()}`;

    try {
      // Your purchase logic here...

      await Afftok.trackConversion({
        offerId,
        amount: 49.99,
        currency: 'USD',
        orderId,
      });

      Alert.alert('Success', 'Purchase complete!');
    } catch (error) {
      console.error('Conversion error:', error);
      // Conversion will retry automatically if offline
    } finally {
      setIsPurchasing(false);
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Premium Plan</Text>
      <Text style={styles.price}>$49.99</Text>
      
      {isPurchasing ? (
        <ActivityIndicator size="large" />
      ) : (
        <Button title="Purchase Now" onPress={handlePurchase} />
      )}
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
    fontSize: 32,
    marginBottom: 30,
  },
});

export default OfferScreen;
```

## Changelog

### 1.0.0
- Initial release
- Click and conversion tracking
- Offline queue support
- Device fingerprinting
- HMAC request signing
- React hooks
- TypeScript support

---

Next: [Web SDK](./web.md)

