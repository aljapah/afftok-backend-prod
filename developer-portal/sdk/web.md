# Web SDK

Complete guide for integrating AffTok tracking into web applications.

---

## Installation

### CDN (Recommended)

```html
<script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>
```

### NPM

```bash
npm install afftok-js
# or
yarn add afftok-js
```

```javascript
import Afftok from 'afftok-js';
```

### ES Module

```html
<script type="module">
  import Afftok from 'https://cdn.afftok.com/sdk/afftok.esm.js';
  
  Afftok.init({ apiKey: 'afftok_live_sk_xxxxx' });
</script>
```

---

## Initialization

```javascript
Afftok.init({
  apiKey: 'afftok_live_sk_xxxxx',
  debug: true,
  enableOfflineQueue: true,
  maxQueueSize: 1000,
  flushIntervalMs: 30000,
  timeoutMs: 30000,
  retryAttempts: 3,
});
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
  cookieDomain?: string;
  cookieExpireDays?: number;
}
```

---

## Click Tracking

### Basic Click Tracking

```javascript
Afftok.trackClick('tc_abc123def456');
```

### Click with Metadata

```javascript
Afftok.trackClick('tc_abc123def456', {
  campaign: 'summer_sale',
  source: 'email',
  sub_id: 'user_123',
});
```

### Click with Callback

```javascript
Afftok.trackClick('tc_abc123def456', { campaign: 'summer_sale' })
  .then((result) => {
    console.log('Click tracked:', result.clickId);
  })
  .catch((error) => {
    console.error('Error:', error.message);
  });
```

### Async/Await

```javascript
async function trackUserClick() {
  try {
    const result = await Afftok.trackClick('tc_abc123def456', {
      campaign: 'summer_sale',
    });
    console.log('Click tracked:', result.clickId);
  } catch (error) {
    console.error('Error:', error.message);
  }
}
```

---

## Conversion Tracking

### Basic Conversion

```javascript
Afftok.trackConversion({
  offerId: 'off_xyz789',
  transactionId: `txn_${Date.now()}`,
});
```

### Conversion with Amount

```javascript
Afftok.trackConversion({
  offerId: 'off_xyz789',
  transactionId: 'order_12345',
  amount: 49.99,
  currency: 'USD',
});
```

### Full Conversion Tracking

```javascript
Afftok.trackConversion({
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
})
  .then((result) => {
    console.log('Conversion tracked:', result.conversionId);
  })
  .catch((error) => {
    console.error('Error:', error.message);
  });
```

### E-commerce Integration

```javascript
// Track purchase on checkout completion
function onCheckoutComplete(order) {
  Afftok.trackConversion({
    offerId: getOfferIdFromCookie(),
    transactionId: order.id,
    amount: order.total,
    currency: order.currency,
    metadata: {
      items: JSON.stringify(order.items),
      coupon: order.couponCode,
      shipping: order.shippingMethod,
    },
  });
}
```

---

## Device Fingerprinting

The SDK automatically generates a unique device fingerprint:

```javascript
// Get current device fingerprint
const fingerprint = Afftok.getDeviceFingerprint();
console.log('Device Fingerprint:', fingerprint);
```

### Custom Fingerprint Provider

```javascript
Afftok.setFingerprintProvider(() => {
  // Custom fingerprint logic
  return `custom_${navigator.userAgent}_${screen.width}x${screen.height}`;
});
```

---

## Offline Queue

Events are automatically queued when offline:

```javascript
// Check queue status
const queueSize = Afftok.getQueueSize();
console.log('Queued events:', queueSize);

// Force flush queue
Afftok.flushQueue()
  .then((result) => {
    console.log('Queue flushed:', result.flushedCount, 'events');
  })
  .catch((error) => {
    console.error('Flush failed:', error.message);
  });

// Clear queue
Afftok.clearQueue();
```

---

## URL Parameter Tracking

Automatically capture tracking codes from URL parameters:

```javascript
// Auto-track from URL on page load
Afftok.init({
  apiKey: 'afftok_live_sk_xxxxx',
  autoTrackUrlParams: true,
  urlParamName: 'tc', // Default: 'tc'
});

// Or manually
const urlParams = new URLSearchParams(window.location.search);
const trackingCode = urlParams.get('tc');
if (trackingCode) {
  Afftok.trackClick(trackingCode);
}
```

---

## Cookie Management

The SDK stores tracking information in cookies:

```javascript
// Configure cookie settings
Afftok.init({
  apiKey: 'afftok_live_sk_xxxxx',
  cookieDomain: '.example.com', // Cross-subdomain tracking
  cookieExpireDays: 30,
  cookieSecure: true,
  cookieSameSite: 'Lax',
});

// Get stored click ID
const clickId = Afftok.getStoredClickId();

// Get stored tracking code
const trackingCode = Afftok.getStoredTrackingCode();

// Clear stored data
Afftok.clearStoredData();
```

---

## User Identification

Associate events with a user:

```javascript
// Set user ID
Afftok.setUserId('user_12345');

// Set user properties
Afftok.setUserProperties({
  email: 'user@example.com',
  plan: 'premium',
  signup_date: '2024-01-15',
});

// Clear user (on logout)
Afftok.clearUser();
```

---

## Event Tracking

Track custom events:

```javascript
// Track page view
Afftok.trackEvent('page_view', {
  page: '/products',
  title: 'Products Page',
});

// Track button click
Afftok.trackEvent('button_click', {
  button: 'add_to_cart',
  product_id: 'prod_123',
});

// Track form submission
Afftok.trackEvent('form_submit', {
  form: 'signup',
  success: true,
});
```

---

## Error Handling

```javascript
// Error types
const ErrorCodes = {
  NETWORK_ERROR: 'NETWORK_ERROR',
  AUTH_ERROR: 'AUTH_ERROR',
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  RATE_LIMIT_ERROR: 'RATE_LIMIT_ERROR',
  SERVER_ERROR: 'SERVER_ERROR',
};

// Handle errors
Afftok.trackClick('tc_abc123')
  .then((result) => {
    console.log('Success:', result);
  })
  .catch((error) => {
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
  });
```

---

## Framework Integrations

### React

```jsx
import { useEffect } from 'react';
import Afftok from 'afftok-js';

// Initialize once
Afftok.init({ apiKey: 'afftok_live_sk_xxxxx' });

function useAfftok() {
  const trackClick = async (trackingCode, metadata) => {
    try {
      return await Afftok.trackClick(trackingCode, metadata);
    } catch (error) {
      console.error('Tracking error:', error);
      throw error;
    }
  };

  const trackConversion = async (options) => {
    try {
      return await Afftok.trackConversion(options);
    } catch (error) {
      console.error('Tracking error:', error);
      throw error;
    }
  };

  return { trackClick, trackConversion };
}

// Usage
function CheckoutButton({ orderId, amount }) {
  const { trackConversion } = useAfftok();

  const handleCheckout = async () => {
    await trackConversion({
      offerId: 'off_xyz789',
      transactionId: orderId,
      amount: amount,
      currency: 'USD',
    });
  };

  return <button onClick={handleCheckout}>Complete Purchase</button>;
}
```

### Vue.js

```javascript
// plugins/afftok.js
import Afftok from 'afftok-js';

export default {
  install(app) {
    Afftok.init({ apiKey: 'afftok_live_sk_xxxxx' });
    app.config.globalProperties.$afftok = Afftok;
  },
};

// main.js
import { createApp } from 'vue';
import App from './App.vue';
import afftok from './plugins/afftok';

createApp(App).use(afftok).mount('#app');

// Component usage
export default {
  methods: {
    async trackPurchase() {
      await this.$afftok.trackConversion({
        offerId: 'off_xyz789',
        transactionId: this.orderId,
        amount: this.total,
      });
    },
  },
};
```

### Next.js

```jsx
// lib/afftok.js
import Afftok from 'afftok-js';

let initialized = false;

export function initAfftok() {
  if (typeof window !== 'undefined' && !initialized) {
    Afftok.init({ apiKey: process.env.NEXT_PUBLIC_AFFTOK_API_KEY });
    initialized = true;
  }
}

export { Afftok };

// pages/_app.js
import { useEffect } from 'react';
import { initAfftok } from '../lib/afftok';

function MyApp({ Component, pageProps }) {
  useEffect(() => {
    initAfftok();
  }, []);

  return <Component {...pageProps} />;
}

export default MyApp;
```

### Angular

```typescript
// afftok.service.ts
import { Injectable } from '@angular/core';
import Afftok from 'afftok-js';

@Injectable({
  providedIn: 'root',
})
export class AfftokService {
  constructor() {
    Afftok.init({ apiKey: 'afftok_live_sk_xxxxx' });
  }

  async trackClick(trackingCode: string, metadata?: Record<string, string>) {
    return Afftok.trackClick(trackingCode, metadata);
  }

  async trackConversion(options: any) {
    return Afftok.trackConversion(options);
  }
}

// component.ts
@Component({
  selector: 'app-checkout',
  template: `<button (click)="onPurchase()">Buy Now</button>`,
})
export class CheckoutComponent {
  constructor(private afftok: AfftokService) {}

  async onPurchase() {
    await this.afftok.trackConversion({
      offerId: 'off_xyz789',
      transactionId: 'order_123',
      amount: 49.99,
    });
  }
}
```

---

## Landing Page Integration

```html
<!DOCTYPE html>
<html>
<head>
  <title>Special Offer</title>
  <script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>
</head>
<body>
  <h1>Special Offer!</h1>
  <button id="cta-button">Get Started</button>

  <script>
    // Initialize
    Afftok.init({
      apiKey: 'afftok_live_sk_xxxxx',
      autoTrackUrlParams: true,
    });

    // Track CTA click
    document.getElementById('cta-button').addEventListener('click', function() {
      Afftok.trackEvent('cta_click', {
        button: 'get_started',
        page: 'landing',
      });
      
      // Redirect to signup
      window.location.href = '/signup';
    });
  </script>
</body>
</html>
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
- Cookie operations
- Error details

---

## Complete Example

```html
<!DOCTYPE html>
<html>
<head>
  <title>E-commerce Store</title>
  <script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>
</head>
<body>
  <div id="app">
    <h1>Premium Subscription</h1>
    <p>$9.99/month</p>
    <button id="subscribe-btn">Subscribe Now</button>
    <div id="status"></div>
  </div>

  <script>
    // Initialize SDK
    Afftok.init({
      apiKey: 'afftok_live_sk_xxxxx',
      debug: true,
      autoTrackUrlParams: true,
    });

    // Track page view
    Afftok.trackEvent('page_view', {
      page: 'subscription',
      title: 'Premium Subscription',
    });

    // Handle subscription
    document.getElementById('subscribe-btn').addEventListener('click', async function() {
      const statusEl = document.getElementById('status');
      const button = this;
      
      button.disabled = true;
      statusEl.textContent = 'Processing...';

      try {
        // Your payment logic here...
        
        // Track conversion
        const result = await Afftok.trackConversion({
          offerId: 'off_premium_subscription',
          transactionId: 'txn_' + Date.now(),
          amount: 9.99,
          currency: 'USD',
          metadata: {
            product: 'premium_monthly',
            user_id: getUserId(),
          },
        });

        statusEl.textContent = 'Purchase successful! ID: ' + result.conversionId;
        console.log('Conversion tracked:', result);
      } catch (error) {
        statusEl.textContent = 'Error: ' + error.message;
        console.error('Tracking failed:', error);
      } finally {
        button.disabled = false;
      }
    });

    function getUserId() {
      // Return current user ID
      return 'user_12345';
    }
  </script>
</body>
</html>
```

---

## Browser Support

| Browser | Version |
|---------|---------|
| Chrome | 60+ |
| Firefox | 55+ |
| Safari | 11+ |
| Edge | 79+ |
| IE | Not supported |

---

## Next Steps

- [Android SDK](android.md) - Android integration guide
- [iOS SDK](ios.md) - iOS integration guide
- [API Reference](../api-reference/clicks.md) - Full API documentation

