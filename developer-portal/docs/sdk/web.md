# Web SDK

Official AffTok SDK for web applications.

## Requirements

- Modern browsers (ES6+ support)
- Chrome 60+, Firefox 55+, Safari 12+, Edge 79+

## Installation

### CDN (Recommended)

```html
<script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
```

### NPM

```bash
npm install @afftok/web-sdk
```

```javascript
import { Afftok } from '@afftok/web-sdk';
```

### Self-Hosted

Download `afftok.min.js` and host on your server:

```html
<script src="/js/afftok.min.js"></script>
```

## Initialization

Initialize the SDK as early as possible:

```html
<script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
<script>
  Afftok.init({
    apiKey: 'afftok_live_sk_your_api_key_here',
    advertiserId: 'adv_123456',
    userId: 'user_789',  // Optional
    debugMode: false,
    offlineQueueEnabled: true,
    maxOfflineEvents: 1000,
    flushIntervalSeconds: 30,
  });
</script>
```

### ES Modules

```javascript
import { Afftok } from '@afftok/web-sdk';

Afftok.init({
  apiKey: 'afftok_live_sk_your_api_key_here',
  advertiserId: 'adv_123456',
  debugMode: process.env.NODE_ENV === 'development',
});
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | string | Required | Your API key |
| `advertiserId` | string | Required | Your advertiser ID |
| `userId` | string | null | Your internal user ID |
| `debugMode` | boolean | false | Enable console logging |
| `offlineQueueEnabled` | boolean | true | Enable offline queue |
| `maxOfflineEvents` | number | 1000 | Max queued events |
| `flushIntervalSeconds` | number | 30 | Queue flush interval |
| `baseUrl` | string | null | Custom API endpoint |

## Track Click

Track when a user clicks on an offer:

```javascript
// Basic click tracking
Afftok.trackClick({
  offerId: 'off_123456',
  trackingCode: 'abc123.1701432000.xyz.sig',
});

// With metadata and callback
Afftok.trackClick({
  offerId: 'off_123456',
  trackingCode: 'abc123.1701432000.xyz.sig',
  metadata: {
    source: 'landing_page',
    campaign: 'summer_sale',
    placement: 'hero_banner',
  },
})
  .then((result) => {
    console.log('Click tracked:', result.clickId);
  })
  .catch((error) => {
    console.error('Error:', error.code, error.message);
  });
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
// Basic conversion
Afftok.trackConversion({
  offerId: 'off_123456',
  amount: 49.99,
  currency: 'USD',
});

// Full conversion with all options
Afftok.trackConversion({
  offerId: 'off_123456',
  amount: 49.99,
  currency: 'USD',
  orderId: 'order_12345',
  productId: 'prod_premium',
  metadata: {
    plan: 'annual',
    coupon: 'SAVE20',
  },
})
  .then((result) => {
    console.log('Conversion tracked:', result.conversionId);
  })
  .catch((error) => {
    console.error('Error:', error.code, error.message);
  });
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

## Async/Await

All SDK methods return Promises:

```javascript
async function trackPurchase() {
  try {
    const result = await Afftok.trackConversion({
      offerId: 'off_123456',
      amount: 49.99,
      currency: 'USD',
      orderId: 'order_12345',
    });
    console.log('Conversion ID:', result.conversionId);
  } catch (error) {
    console.error('Error:', error.code, error.message);
  }
}
```

## URL Parameter Tracking

Automatically track clicks from URL parameters:

```javascript
// On page load
document.addEventListener('DOMContentLoaded', () => {
  const params = new URLSearchParams(window.location.search);
  const trackingCode = params.get('ref');
  const offerId = params.get('offer');

  if (trackingCode && offerId) {
    Afftok.trackClick({
      offerId,
      trackingCode,
      metadata: {
        source: 'landing_page',
        url: window.location.href,
      },
    });
  }
});
```

### Auto-Track Helper

```javascript
// Enable automatic URL tracking
Afftok.init({
  apiKey: 'afftok_live_sk_...',
  advertiserId: 'adv_123456',
  autoTrackUrlParams: true,  // Automatically tracks ?ref= and ?offer=
  urlParamNames: {
    trackingCode: 'ref',     // Default
    offerId: 'offer',        // Default
  },
});
```

## E-commerce Integration

### Checkout Tracking

```javascript
// Track when user starts checkout
function onCheckoutStart(cart) {
  Afftok.trackClick({
    offerId: cart.offerId,
    metadata: {
      event: 'checkout_start',
      cart_value: cart.total,
      items: cart.items.length,
    },
  });
}

// Track successful purchase
function onPurchaseComplete(order) {
  Afftok.trackConversion({
    offerId: order.offerId,
    amount: order.total,
    currency: order.currency,
    orderId: order.id,
    metadata: {
      items: order.items.map((i) => i.sku).join(','),
      payment_method: order.paymentMethod,
    },
  });
}
```

### Thank You Page

```html
<!-- thank-you.html -->
<script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
<script>
  Afftok.init({
    apiKey: 'afftok_live_sk_...',
    advertiserId: 'adv_123456',
  });

  // Get order data from server-rendered variables
  const orderData = {
    offerId: '{{ order.offer_id }}',
    amount: {{ order.total }},
    currency: '{{ order.currency }}',
    orderId: '{{ order.id }}',
  };

  Afftok.trackConversion(orderData);
</script>
```

## Form Submission Tracking

```javascript
document.getElementById('signup-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  
  const formData = new FormData(e.target);
  
  try {
    // Submit form to your server
    const response = await fetch('/api/signup', {
      method: 'POST',
      body: formData,
    });
    
    const result = await response.json();
    
    // Track conversion
    await Afftok.trackConversion({
      offerId: 'off_signup',
      amount: 0,  // Lead conversion
      orderId: result.userId,
      metadata: {
        type: 'lead',
        source: 'signup_form',
      },
    });
    
    // Redirect to success page
    window.location.href = '/welcome';
  } catch (error) {
    console.error('Signup error:', error);
  }
});
```

## Button Click Tracking

```html
<button 
  class="cta-button"
  data-offer-id="off_123456"
  data-tracking-code="abc123.1701432000.xyz.sig"
  onclick="trackOfferClick(this)"
>
  Get Started
</button>

<script>
function trackOfferClick(button) {
  Afftok.trackClick({
    offerId: button.dataset.offerId,
    trackingCode: button.dataset.trackingCode,
    metadata: {
      button_text: button.textContent,
      page: window.location.pathname,
    },
  });
}
</script>
```

## Offline Queue

The SDK uses localStorage for offline queuing:

```javascript
// Check queue status
const queueSize = Afftok.getQueueSize();
console.log('Pending events:', queueSize);

// Force flush queue
Afftok.flushQueue()
  .then(({ sent, failed }) => {
    console.log(`Flushed: ${sent} sent, ${failed} failed`);
  });

// Clear queue
Afftok.clearQueue();
```

### Online/Offline Detection

```javascript
window.addEventListener('online', () => {
  console.log('Back online, flushing queue...');
  Afftok.flushQueue();
});

window.addEventListener('offline', () => {
  console.log('Offline, events will be queued');
});
```

## Device Fingerprinting

The SDK generates a browser fingerprint:

```javascript
// Get current fingerprint
const fingerprint = Afftok.getDeviceFingerprint();
console.log('Fingerprint:', fingerprint);
```

Fingerprint components:
- Canvas fingerprint
- WebGL renderer
- Screen resolution
- Timezone
- Language
- User agent
- Installed plugins

## User Identification

```javascript
// Set user ID
Afftok.setUserId('user_new_id');

// Get current user ID
const userId = Afftok.getUserId();

// Clear user ID (on logout)
Afftok.clearUserId();
```

## Error Handling

```javascript
Afftok.trackClick({ offerId: 'off_123456' })
  .then((result) => {
    console.log('Success:', result);
  })
  .catch((error) => {
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
  });
```

## React Integration

```jsx
import { useEffect, useState } from 'react';
import { Afftok } from '@afftok/web-sdk';

// Initialize once
Afftok.init({
  apiKey: process.env.REACT_APP_AFFTOK_API_KEY,
  advertiserId: process.env.REACT_APP_AFFTOK_ADVERTISER_ID,
});

function OfferPage({ offerId, trackingCode }) {
  const [isTracking, setIsTracking] = useState(false);

  useEffect(() => {
    if (trackingCode) {
      Afftok.trackClick({
        offerId,
        trackingCode,
        metadata: { source: 'offer_page' },
      });
    }
  }, [offerId, trackingCode]);

  const handlePurchase = async () => {
    setIsTracking(true);
    try {
      await Afftok.trackConversion({
        offerId,
        amount: 49.99,
        currency: 'USD',
        orderId: `order_${Date.now()}`,
      });
      // Handle success
    } catch (error) {
      console.error('Tracking error:', error);
    } finally {
      setIsTracking(false);
    }
  };

  return (
    <div>
      <h1>Special Offer</h1>
      <button onClick={handlePurchase} disabled={isTracking}>
        {isTracking ? 'Processing...' : 'Purchase Now'}
      </button>
    </div>
  );
}
```

## Vue Integration

```vue
<template>
  <div>
    <h1>Special Offer</h1>
    <button @click="handlePurchase" :disabled="isTracking">
      {{ isTracking ? 'Processing...' : 'Purchase Now' }}
    </button>
  </div>
</template>

<script>
import { Afftok } from '@afftok/web-sdk';

export default {
  props: ['offerId', 'trackingCode'],
  data() {
    return {
      isTracking: false,
    };
  },
  mounted() {
    if (this.trackingCode) {
      Afftok.trackClick({
        offerId: this.offerId,
        trackingCode: this.trackingCode,
      });
    }
  },
  methods: {
    async handlePurchase() {
      this.isTracking = true;
      try {
        await Afftok.trackConversion({
          offerId: this.offerId,
          amount: 49.99,
          currency: 'USD',
          orderId: `order_${Date.now()}`,
        });
      } catch (error) {
        console.error('Tracking error:', error);
      } finally {
        this.isTracking = false;
      }
    },
  },
};
</script>
```

## Testing

### Test Mode

```javascript
Afftok.init({
  apiKey: 'afftok_test_sk_...',  // Test key
  advertiserId: 'adv_test_123',
  debugMode: true,  // Enable console logging
});
```

### Mock Events

```javascript
if (process.env.NODE_ENV === 'development') {
  Afftok.generateTestClick();
  Afftok.generateTestConversion({ amount: 99.99 });
}
```

## Content Security Policy

If using CSP, add:

```
script-src 'self' https://cdn.afftok.com;
connect-src 'self' https://api.afftok.com;
```

## Google Tag Manager

```html
<!-- GTM Custom HTML Tag -->
<script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
<script>
  Afftok.init({
    apiKey: '{{Afftok API Key}}',
    advertiserId: '{{Afftok Advertiser ID}}',
  });
</script>
```

Create triggers for:
- Page View (for click tracking from URL params)
- Form Submit (for lead conversions)
- Custom Event (for purchase conversions)

## Complete Example

```html
<!DOCTYPE html>
<html>
<head>
  <title>Special Offer</title>
</head>
<body>
  <div id="offer">
    <h1>Premium Plan</h1>
    <p class="price">$49.99</p>
    <button id="purchase-btn">Purchase Now</button>
  </div>

  <script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
  <script>
    // Initialize SDK
    Afftok.init({
      apiKey: 'afftok_live_sk_your_key',
      advertiserId: 'adv_123456',
      debugMode: false,
    });

    // Track click from URL params
    document.addEventListener('DOMContentLoaded', () => {
      const params = new URLSearchParams(window.location.search);
      const trackingCode = params.get('ref');
      const offerId = params.get('offer') || 'off_default';

      if (trackingCode) {
        Afftok.trackClick({
          offerId,
          trackingCode,
          metadata: { page: 'offer_landing' },
        });
      }

      // Store for conversion
      window.currentOffer = { offerId, trackingCode };
    });

    // Track purchase
    document.getElementById('purchase-btn').addEventListener('click', async () => {
      const btn = document.getElementById('purchase-btn');
      btn.disabled = true;
      btn.textContent = 'Processing...';

      try {
        // Your purchase logic here...
        const orderId = 'order_' + Date.now();

        await Afftok.trackConversion({
          offerId: window.currentOffer.offerId,
          amount: 49.99,
          currency: 'USD',
          orderId,
        });

        alert('Purchase complete!');
        window.location.href = '/thank-you?order=' + orderId;
      } catch (error) {
        console.error('Error:', error);
        btn.disabled = false;
        btn.textContent = 'Purchase Now';
      }
    });
  </script>
</body>
</html>
```

## Changelog

### 1.0.0
- Initial release
- Click and conversion tracking
- Offline queue with localStorage
- Browser fingerprinting
- HMAC request signing
- Promise-based API

---

Next: [Webhook Documentation](../webhooks/README.md)

