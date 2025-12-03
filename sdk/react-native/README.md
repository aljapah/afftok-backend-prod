# AffTok React Native SDK

Official React Native SDK for AffTok affiliate tracking platform.

## Features

- 🚀 Click & Conversion Tracking
- 📴 Offline Queue with Automatic Retry
- 🔐 HMAC-SHA256 Signature Verification
- 🔄 Exponential Backoff Retry
- 📱 Device Fingerprinting
- 🎯 Zero-Drop Tracking Compatible
- 🌐 Cross-Platform (iOS & Android)

## Requirements

- React Native 0.60+
- React 16.8+

## Installation

```bash
npm install @afftok/react-native-sdk
# or
yarn add @afftok/react-native-sdk
```

Install peer dependencies:

```bash
npm install @react-native-async-storage/async-storage
# or
yarn add @react-native-async-storage/async-storage
```

### iOS Setup

```bash
cd ios && pod install
```

## Quick Start

### 1. Initialize the SDK

```javascript
import Afftok from '@afftok/react-native-sdk';

// In your App.js or index.js
useEffect(() => {
  Afftok.initialize({
    apiKey: 'your_api_key',
    advertiserId: 'your_advertiser_id',
    userId: 'optional_user_id',  // Optional
    debug: __DEV__,              // Enable logging in dev
  });
}, []);
```

### 2. Track Clicks

```javascript
import Afftok from '@afftok/react-native-sdk';

const handleClick = async () => {
  const response = await Afftok.trackClick({
    offerId: 'offer_123',
    trackingCode: 'abc123',    // Optional
    subId1: 'campaign_1',      // Optional
    subId2: 'creative_2',      // Optional
    customParams: {            // Optional
      source: 'facebook',
      campaign: 'summer_sale',
    },
  });

  if (response.success) {
    console.log('Click tracked!');
  }
};
```

### 3. Track Conversions

```javascript
import Afftok from '@afftok/react-native-sdk';

const handlePurchase = async () => {
  const response = await Afftok.trackConversion({
    offerId: 'offer_123',
    transactionId: 'txn_abc123',
    amount: 29.99,
    currency: 'USD',
    status: 'approved',    // 'pending', 'approved', 'rejected'
    clickId: 'click_xyz',  // Optional, for attribution
  });

  if (response.success) {
    console.log('Conversion tracked!');
  }
};
```

### 4. Track with Signed Links

```javascript
// When using pre-signed tracking links from AffTok
const response = await Afftok.trackSignedClick(
  'https://track.afftok.com/c/abc123.1234567890.xyz.signature',
  { offerId: 'offer_123' }
);
```

## Offline Queue

The SDK automatically queues events when offline and retries with exponential backoff.

```javascript
// Check pending items
const pendingCount = Afftok.getPendingCount();

// Manually flush queue
await Afftok.flush();

// Clear queue
Afftok.clearQueue();
```

## Device Information

```javascript
// Get device fingerprint
const fingerprint = Afftok.getFingerprint();

// Get device ID
const deviceId = Afftok.getDeviceId();

// Get full device info
const deviceInfo = Afftok.getDeviceInfo();
```

## Configuration Options

```javascript
Afftok.initialize({
  apiKey: 'your_api_key',              // Required
  advertiserId: 'your_advertiser',      // Required
  userId: null,                         // Optional user identifier
  baseUrl: 'https://api.afftok.com',   // Custom API URL
  debug: false,                         // Enable debug logging
  autoFlush: true,                      // Auto-flush queue
  flushInterval: 30000,                 // Flush interval in ms
});
```

## Error Handling

```javascript
const response = await Afftok.trackClick(params);

if (response.success) {
  // Event tracked successfully
  console.log('Data:', response.data);
} else if (response.error) {
  // Error occurred
  console.error('Error:', response.error);
} else if (response.message) {
  // Event queued for retry
  console.warn('Queued:', response.message);
}
```

## Hooks Example

```javascript
import React, { useState, useCallback } from 'react';
import { Button, Text } from 'react-native';
import Afftok from '@afftok/react-native-sdk';

const TrackingButton = ({ offerId }) => {
  const [isTracking, setIsTracking] = useState(false);
  const [result, setResult] = useState(null);

  const handleTrack = useCallback(async () => {
    setIsTracking(true);
    
    const response = await Afftok.trackClick({ offerId });
    
    setIsTracking(false);
    setResult(response.success ? 'Tracked!' : 'Error');
  }, [offerId]);

  return (
    <>
      <Button
        title={isTracking ? 'Tracking...' : 'Track Click'}
        onPress={handleTrack}
        disabled={isTracking}
      />
      {result && <Text>{result}</Text>}
    </>
  );
};
```

## TypeScript Support

TypeScript definitions are included:

```typescript
import Afftok, { AfftokConfig } from '@afftok/react-native-sdk';

interface TrackingParams {
  offerId: string;
  trackingCode?: string;
  subId1?: string;
  subId2?: string;
  subId3?: string;
  customParams?: Record<string, string>;
}

const params: TrackingParams = {
  offerId: 'offer_123',
};

const response = await Afftok.trackClick(params);
```

## Best Practices

1. **Initialize Early**: Initialize in your App component's `useEffect`
2. **Use Debug Mode**: Enable during development
3. **Handle Responses**: Always check response status
4. **Unique Transaction IDs**: Use unique IDs for conversions
5. **Shutdown Properly**: Call `Afftok.shutdown()` on app exit

## Troubleshooting

### AsyncStorage Error

If you see AsyncStorage errors, make sure you've installed the peer dependency:

```bash
npm install @react-native-async-storage/async-storage
```

### iOS Build Errors

Run pod install:

```bash
cd ios && pod install
```

### Network Issues

The SDK automatically queues failed requests. Check:

```javascript
console.log('Pending:', Afftok.getPendingCount());
```

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- Issues: https://github.com/afftok/react-native-sdk/issues

## License

MIT License - see LICENSE file for details.

