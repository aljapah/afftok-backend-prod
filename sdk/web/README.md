# AffTok Web SDK

Official JavaScript SDK for AffTok affiliate tracking platform.

## Features

- 🚀 Click & Conversion Tracking
- 📴 Offline Queue with Automatic Retry
- 🔐 HMAC-SHA256 Signature Verification
- 🔄 Exponential Backoff Retry
- 🌐 Browser Fingerprinting
- 🎯 Zero-Drop Tracking Compatible
- 📦 No Dependencies (Vanilla JS)
- ⚡ Auto-tracking Support

## Installation

### CDN

```html
<!-- Production (minified) -->
<script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>

<!-- Development -->
<script src="https://cdn.afftok.com/sdk/afftok.js"></script>
```

### NPM

```bash
npm install @afftok/web-sdk
```

```javascript
import Afftok from '@afftok/web-sdk';
```

### Download

Download `afftok.min.js` and include it in your project:

```html
<script src="/path/to/afftok.min.js"></script>
```

## Quick Start

### 1. Initialize the SDK

```html
<script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>
<script>
  Afftok.init({
    apiKey: 'your_api_key',
    advertiserId: 'your_advertiser_id',
    userId: 'optional_user_id',  // Optional
    debug: true,                  // Enable logging
  });
</script>
```

### 2. Track Clicks

```javascript
// Track a click
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
```

### 3. Track Conversions

```javascript
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
```

### 4. Auto-Tracking

Enable automatic click tracking on links with data attributes:

```html
<a href="https://example.com" 
   data-afftok-offer="offer_123"
   data-afftok-code="abc123"
   data-afftok-sub1="campaign_1">
  Click Me
</a>

<script>
  Afftok.init({
    apiKey: 'your_api_key',
    advertiserId: 'your_advertiser_id',
    autoTrack: true,  // Enable auto-tracking
  });
</script>
```

Or enable manually:

```javascript
Afftok.autotrack();
```

### 5. Track with Signed Links

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
Afftok.init({
  apiKey: 'your_api_key',              // Required
  advertiserId: 'your_advertiser',      // Required
  userId: null,                         // Optional user identifier
  baseUrl: 'https://api.afftok.com',   // Custom API URL
  debug: false,                         // Enable debug logging
  autoFlush: true,                      // Auto-flush queue
  autoTrack: false,                     // Auto-track clicks
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

## E-commerce Integration

### Shopify

```html
{% if first_time_accessed %}
<script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>
<script>
  Afftok.init({
    apiKey: 'your_api_key',
    advertiserId: 'your_advertiser_id',
  });
  
  Afftok.trackConversion({
    offerId: '{{ shop.permanent_domain }}',
    transactionId: '{{ order.order_number }}',
    amount: {{ order.total_price | money_without_currency }},
    currency: '{{ shop.currency }}',
    status: 'approved',
  });
</script>
{% endif %}
```

### WooCommerce

```php
add_action('woocommerce_thankyou', 'afftok_track_conversion');
function afftok_track_conversion($order_id) {
    $order = wc_get_order($order_id);
    ?>
    <script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>
    <script>
        Afftok.init({
            apiKey: 'your_api_key',
            advertiserId: 'your_advertiser_id',
        });
        
        Afftok.trackConversion({
            offerId: 'woocommerce',
            transactionId: '<?php echo $order_id; ?>',
            amount: <?php echo $order->get_total(); ?>,
            currency: '<?php echo $order->get_currency(); ?>',
            status: 'approved',
        });
    </script>
    <?php
}
```

## Google Tag Manager

Create a Custom HTML tag:

```html
<script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>
<script>
  Afftok.init({
    apiKey: '{{Constant - AffTok API Key}}',
    advertiserId: '{{Constant - AffTok Advertiser ID}}',
  });
  
  // Track page view as click
  Afftok.trackClick({
    offerId: '{{Page Path}}',
    customParams: {
      page_title: '{{Page Title}}',
      referrer: '{{Referrer}}',
    },
  });
</script>
```

## Best Practices

1. **Initialize Early**: Initialize as soon as possible
2. **Use Debug Mode**: Enable during development
3. **Handle Responses**: Always check response status
4. **Unique Transaction IDs**: Use unique IDs for conversions
5. **Test Offline**: Test offline queue functionality

## Browser Support

- Chrome 60+
- Firefox 55+
- Safari 11+
- Edge 79+

Note: Requires `crypto.subtle` for HMAC signatures. Falls back gracefully in older browsers.

## Support

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- Issues: https://github.com/afftok/web-sdk/issues

## License

MIT License - see LICENSE file for details.

