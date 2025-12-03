# Integration Guides

Step-by-step guides for integrating AffTok with your systems.

## Available Guides

| Guide | Description |
|-------|-------------|
| [Server-to-Server](./server-to-server.md) | Backend integration examples |
| [Landing Page](./landing-page.md) | Tracking on landing pages |
| [Mobile App](./mobile-app.md) | Mobile app integration |
| [E-commerce](./ecommerce.md) | Shop and checkout tracking |
| [WordPress](./wordpress.md) | WordPress plugin integration |

---

## Server-to-Server Integration

### Node.js Example

```javascript
const crypto = require('crypto');
const axios = require('axios');

class AfftokClient {
  constructor(apiKey, advertiserId) {
    this.apiKey = apiKey;
    this.advertiserId = advertiserId;
    this.baseUrl = 'https://api.afftok.com';
  }

  generateSignature(payload, timestamp, nonce) {
    const message = `${JSON.stringify(payload)}${timestamp}${nonce}`;
    return crypto
      .createHmac('sha256', this.apiKey)
      .update(message)
      .digest('hex');
  }

  async sendPostback(data) {
    const timestamp = Math.floor(Date.now() / 1000).toString();
    const nonce = crypto.randomBytes(16).toString('hex');
    const signature = this.generateSignature(data, timestamp, nonce);

    const response = await axios.post(
      `${this.baseUrl}/api/postback`,
      data,
      {
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
          'X-Afftok-Timestamp': timestamp,
          'X-Afftok-Nonce': nonce,
          'X-Afftok-Signature': `sha256=${signature}`,
        },
      }
    );

    return response.data;
  }

  async trackConversion(conversionData) {
    return this.sendPostback({
      click_id: conversionData.clickId,
      external_id: conversionData.orderId,
      amount: conversionData.amount,
      currency: conversionData.currency || 'USD',
      status: 'approved',
      advertiser_id: this.advertiserId,
    });
  }
}

// Usage
const client = new AfftokClient(
  process.env.AFFTOK_API_KEY,
  process.env.AFFTOK_ADVERTISER_ID
);

// After successful purchase
await client.trackConversion({
  clickId: 'click_123',
  orderId: 'order_456',
  amount: 49.99,
  currency: 'USD',
});
```

### Python Example

```python
import os
import time
import hmac
import hashlib
import secrets
import json
import httpx

class AfftokClient:
    def __init__(self, api_key: str, advertiser_id: str):
        self.api_key = api_key
        self.advertiser_id = advertiser_id
        self.base_url = 'https://api.afftok.com'

    def _generate_signature(self, payload: dict, timestamp: str, nonce: str) -> str:
        message = f"{json.dumps(payload)}{timestamp}{nonce}"
        return hmac.new(
            self.api_key.encode(),
            message.encode(),
            hashlib.sha256
        ).hexdigest()

    async def send_postback(self, data: dict) -> dict:
        timestamp = str(int(time.time()))
        nonce = secrets.token_hex(16)
        signature = self._generate_signature(data, timestamp, nonce)

        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{self.base_url}/api/postback",
                json=data,
                headers={
                    'Content-Type': 'application/json',
                    'X-API-Key': self.api_key,
                    'X-Afftok-Timestamp': timestamp,
                    'X-Afftok-Nonce': nonce,
                    'X-Afftok-Signature': f'sha256={signature}',
                }
            )
            return response.json()

    async def track_conversion(
        self,
        click_id: str,
        order_id: str,
        amount: float,
        currency: str = 'USD'
    ) -> dict:
        return await self.send_postback({
            'click_id': click_id,
            'external_id': order_id,
            'amount': amount,
            'currency': currency,
            'status': 'approved',
            'advertiser_id': self.advertiser_id,
        })

# Usage
import asyncio

async def main():
    client = AfftokClient(
        os.environ['AFFTOK_API_KEY'],
        os.environ['AFFTOK_ADVERTISER_ID']
    )

    result = await client.track_conversion(
        click_id='click_123',
        order_id='order_456',
        amount=49.99,
        currency='USD'
    )
    print(result)

asyncio.run(main())
```

### PHP Example

```php
<?php

class AfftokClient {
    private string $apiKey;
    private string $advertiserId;
    private string $baseUrl = 'https://api.afftok.com';

    public function __construct(string $apiKey, string $advertiserId) {
        $this->apiKey = $apiKey;
        $this->advertiserId = $advertiserId;
    }

    private function generateSignature(array $payload, string $timestamp, string $nonce): string {
        $message = json_encode($payload) . $timestamp . $nonce;
        return hash_hmac('sha256', $message, $this->apiKey);
    }

    public function sendPostback(array $data): array {
        $timestamp = (string) time();
        $nonce = bin2hex(random_bytes(16));
        $signature = $this->generateSignature($data, $timestamp, $nonce);

        $ch = curl_init();
        curl_setopt_array($ch, [
            CURLOPT_URL => $this->baseUrl . '/api/postback',
            CURLOPT_POST => true,
            CURLOPT_POSTFIELDS => json_encode($data),
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HTTPHEADER => [
                'Content-Type: application/json',
                'X-API-Key: ' . $this->apiKey,
                'X-Afftok-Timestamp: ' . $timestamp,
                'X-Afftok-Nonce: ' . $nonce,
                'X-Afftok-Signature: sha256=' . $signature,
            ],
        ]);

        $response = curl_exec($ch);
        curl_close($ch);

        return json_decode($response, true);
    }

    public function trackConversion(
        string $clickId,
        string $orderId,
        float $amount,
        string $currency = 'USD'
    ): array {
        return $this->sendPostback([
            'click_id' => $clickId,
            'external_id' => $orderId,
            'amount' => $amount,
            'currency' => $currency,
            'status' => 'approved',
            'advertiser_id' => $this->advertiserId,
        ]);
    }
}

// Usage
$client = new AfftokClient(
    getenv('AFFTOK_API_KEY'),
    getenv('AFFTOK_ADVERTISER_ID')
);

$result = $client->trackConversion(
    'click_123',
    'order_456',
    49.99,
    'USD'
);

print_r($result);
```

### Go Example

```go
package main

import (
    "bytes"
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "time"
)

type AfftokClient struct {
    APIKey       string
    AdvertiserID string
    BaseURL      string
}

func NewAfftokClient(apiKey, advertiserId string) *AfftokClient {
    return &AfftokClient{
        APIKey:       apiKey,
        AdvertiserID: advertiserId,
        BaseURL:      "https://api.afftok.com",
    }
}

func (c *AfftokClient) generateSignature(payload []byte, timestamp, nonce string) string {
    message := string(payload) + timestamp + nonce
    mac := hmac.New(sha256.New, []byte(c.APIKey))
    mac.Write([]byte(message))
    return hex.EncodeToString(mac.Sum(nil))
}

func (c *AfftokClient) generateNonce() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}

func (c *AfftokClient) SendPostback(data map[string]interface{}) (map[string]interface{}, error) {
    payload, _ := json.Marshal(data)
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    nonce := c.generateNonce()
    signature := c.generateSignature(payload, timestamp, nonce)

    req, _ := http.NewRequest("POST", c.BaseURL+"/api/postback", bytes.NewBuffer(payload))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", c.APIKey)
    req.Header.Set("X-Afftok-Timestamp", timestamp)
    req.Header.Set("X-Afftok-Nonce", nonce)
    req.Header.Set("X-Afftok-Signature", "sha256="+signature)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}

func (c *AfftokClient) TrackConversion(clickID, orderID string, amount float64, currency string) (map[string]interface{}, error) {
    return c.SendPostback(map[string]interface{}{
        "click_id":      clickID,
        "external_id":   orderID,
        "amount":        amount,
        "currency":      currency,
        "status":        "approved",
        "advertiser_id": c.AdvertiserID,
    })
}

func main() {
    client := NewAfftokClient(
        os.Getenv("AFFTOK_API_KEY"),
        os.Getenv("AFFTOK_ADVERTISER_ID"),
    )

    result, err := client.TrackConversion("click_123", "order_456", 49.99, "USD")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Result:", result)
}
```

---

## Landing Page Integration

### Basic HTML

```html
<!DOCTYPE html>
<html>
<head>
    <title>Special Offer</title>
</head>
<body>
    <h1>Premium Plan - $49.99</h1>
    <button id="cta-button">Get Started</button>

    <script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
    <script>
        // Initialize
        Afftok.init({
            apiKey: 'afftok_live_sk_...',
            advertiserId: 'adv_123456',
        });

        // Track click from URL params
        const params = new URLSearchParams(window.location.search);
        const trackingCode = params.get('ref');
        const offerId = params.get('offer') || 'off_default';

        if (trackingCode) {
            Afftok.trackClick({
                offerId,
                trackingCode,
                metadata: { page: 'landing' }
            });
        }

        // Store for later conversion
        sessionStorage.setItem('afftok_offer', offerId);
        sessionStorage.setItem('afftok_ref', trackingCode);

        // CTA button
        document.getElementById('cta-button').addEventListener('click', () => {
            window.location.href = '/checkout';
        });
    </script>
</body>
</html>
```

### Checkout Page

```html
<!DOCTYPE html>
<html>
<head>
    <title>Checkout</title>
</head>
<body>
    <form id="checkout-form">
        <input type="text" name="email" placeholder="Email" required>
        <input type="text" name="card" placeholder="Card Number" required>
        <button type="submit">Complete Purchase</button>
    </form>

    <script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
    <script>
        Afftok.init({
            apiKey: 'afftok_live_sk_...',
            advertiserId: 'adv_123456',
        });

        document.getElementById('checkout-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            // Process payment...
            const orderId = 'order_' + Date.now();
            
            // Track conversion
            const offerId = sessionStorage.getItem('afftok_offer');
            
            await Afftok.trackConversion({
                offerId,
                amount: 49.99,
                currency: 'USD',
                orderId,
            });

            // Redirect to thank you page
            window.location.href = '/thank-you?order=' + orderId;
        });
    </script>
</body>
</html>
```

---

## Mobile App Integration

### React Native

```javascript
import { useEffect } from 'react';
import { Linking } from 'react-native';
import { Afftok } from '@afftok/react-native';

// Initialize in App.js
Afftok.init({
    apiKey: 'afftok_live_sk_...',
    advertiserId: 'adv_123456',
});

// Handle deep links
function App() {
    useEffect(() => {
        const handleUrl = ({ url }) => {
            const params = new URL(url).searchParams;
            const ref = params.get('ref');
            const offer = params.get('offer');

            if (ref && offer) {
                Afftok.trackClick({
                    offerId: offer,
                    trackingCode: ref,
                    metadata: { source: 'deep_link' }
                });
            }
        };

        Linking.getInitialURL().then(url => {
            if (url) handleUrl({ url });
        });

        const subscription = Linking.addEventListener('url', handleUrl);
        return () => subscription.remove();
    }, []);

    return <MainNavigator />;
}

// In purchase screen
async function handlePurchase(offerId, amount) {
    // Process payment...
    
    await Afftok.trackConversion({
        offerId,
        amount,
        currency: 'USD',
        orderId: `order_${Date.now()}`,
    });
}
```

### Flutter

```dart
import 'package:afftok_sdk/afftok_sdk.dart';
import 'package:uni_links/uni_links.dart';

void main() async {
    WidgetsFlutterBinding.ensureInitialized();
    
    await Afftok.init(
        apiKey: 'afftok_live_sk_...',
        advertiserId: 'adv_123456',
    );
    
    runApp(MyApp());
}

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
        try {
            final initialLink = await getInitialLink();
            if (initialLink != null) _handleDeepLink(initialLink);
        } catch (e) {}

        _linkSubscription = linkStream.listen((link) {
            if (link != null) _handleDeepLink(link);
        });
    }

    void _handleDeepLink(String link) {
        final uri = Uri.parse(link);
        final ref = uri.queryParameters['ref'];
        final offer = uri.queryParameters['offer'];

        if (ref != null && offer != null) {
            Afftok.trackClick(
                offerId: offer,
                trackingCode: ref,
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

// Purchase handling
Future<void> handlePurchase(String offerId, double amount) async {
    // Process payment...
    
    await Afftok.trackConversion(
        offerId: offerId,
        amount: amount,
        currency: 'USD',
        orderId: 'order_${DateTime.now().millisecondsSinceEpoch}',
    );
}
```

---

## E-commerce Integration

### Shopify

Add to theme.liquid before `</head>`:

```liquid
<script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
<script>
    Afftok.init({
        apiKey: '{{ settings.afftok_api_key }}',
        advertiserId: '{{ settings.afftok_advertiser_id }}',
    });

    // Track clicks from URL
    const params = new URLSearchParams(window.location.search);
    const ref = params.get('ref');
    if (ref) {
        sessionStorage.setItem('afftok_ref', ref);
        Afftok.trackClick({
            offerId: 'shopify_store',
            trackingCode: ref,
        });
    }
</script>
```

Add to checkout thank you page (Additional Scripts):

```liquid
<script src="https://cdn.afftok.com/sdk/v1/afftok.min.js"></script>
<script>
    Afftok.init({
        apiKey: '{{ settings.afftok_api_key }}',
        advertiserId: '{{ settings.afftok_advertiser_id }}',
    });

    Afftok.trackConversion({
        offerId: 'shopify_store',
        amount: {{ checkout.total_price | money_without_currency }},
        currency: '{{ checkout.currency }}',
        orderId: '{{ checkout.order_number }}',
    });
</script>
```

### WooCommerce

Add to functions.php:

```php
// Track click from URL
add_action('wp_head', function() {
    if (isset($_GET['ref'])) {
        WC()->session->set('afftok_ref', sanitize_text_field($_GET['ref']));
    }
});

// Track conversion after order
add_action('woocommerce_thankyou', function($order_id) {
    $order = wc_get_order($order_id);
    $ref = WC()->session->get('afftok_ref');
    
    if (!$ref) return;
    
    $api_key = get_option('afftok_api_key');
    $advertiser_id = get_option('afftok_advertiser_id');
    
    $data = [
        'click_id' => $ref,
        'external_id' => $order_id,
        'amount' => $order->get_total(),
        'currency' => $order->get_currency(),
        'status' => 'approved',
        'advertiser_id' => $advertiser_id,
    ];
    
    $timestamp = time();
    $nonce = bin2hex(random_bytes(16));
    $signature = hash_hmac('sha256', json_encode($data) . $timestamp . $nonce, $api_key);
    
    wp_remote_post('https://api.afftok.com/api/postback', [
        'headers' => [
            'Content-Type' => 'application/json',
            'X-API-Key' => $api_key,
            'X-Afftok-Timestamp' => $timestamp,
            'X-Afftok-Nonce' => $nonce,
            'X-Afftok-Signature' => 'sha256=' . $signature,
        ],
        'body' => json_encode($data),
    ]);
    
    WC()->session->set('afftok_ref', null);
});
```

---

## Testing Your Integration

### 1. Use Test Keys

```javascript
Afftok.init({
    apiKey: 'afftok_test_sk_...',  // Test key
    advertiserId: 'adv_test_123',
    debugMode: true,
});
```

### 2. Generate Test Events

```javascript
// Test click
await Afftok.generateTestClick();

// Test conversion
await Afftok.generateTestConversion({ amount: 99.99 });
```

### 3. Verify in Admin Panel

Check the admin dashboard for:
- Click recorded
- Conversion recorded
- Correct amounts
- Proper attribution

---

Next: [Testing & QA](../testing/README.md)

