# Tracking Flow

This document explains the complete lifecycle of a tracking event in AffTok, from initial click to conversion attribution.

## Click → Postback → Conversion Lifecycle

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        COMPLETE TRACKING LIFECYCLE                          │
└─────────────────────────────────────────────────────────────────────────────┘

     USER                    AFFTOK                   ADVERTISER
       │                       │                          │
       │  1. Click Link        │                          │
       │──────────────────────>│                          │
       │                       │                          │
       │                       │ 2. Validate & Record     │
       │                       │    Click                 │
       │                       │                          │
       │  3. Redirect          │                          │
       │<──────────────────────│                          │
       │                       │                          │
       │  4. Land on Offer     │                          │
       │──────────────────────────────────────────────────>│
       │                       │                          │
       │  5. Complete Action   │                          │
       │  (Purchase/Signup)    │                          │
       │──────────────────────────────────────────────────>│
       │                       │                          │
       │                       │  6. Postback             │
       │                       │<─────────────────────────│
       │                       │                          │
       │                       │ 7. Match & Attribute     │
       │                       │    Conversion            │
       │                       │                          │
       │                       │ 8. Update Stats          │
       │                       │                          │
       │                       │ 9. Trigger Webhooks      │
       │                       │                          │
```

## Detailed Flow Breakdown

### Phase 1: Click Tracking

#### Step 1: User Clicks Tracking Link

The user clicks a tracking link in one of these formats:

```
# Signed Link (Recommended)
https://track.afftok.com/c/abc123.1701234567890.xyz789.a1b2c3d4

# Legacy Link
https://track.afftok.com/c/abc123
```

#### Step 2: Edge Validation

At the Cloudflare Edge, the following checks occur:

```
┌─────────────────────────────────────────────────────────────┐
│                    EDGE VALIDATION                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Parse Link Components                                   │
│     ├── tracking_code: abc123                               │
│     ├── timestamp: 1701234567890                            │
│     ├── nonce: xyz789                                       │
│     └── signature: a1b2c3d4                                 │
│                                                             │
│  2. Validate Signature (HMAC-SHA256)                        │
│     signature = HMAC(secret, "abc123.1701234567890.xyz789") │
│                                                             │
│  3. Check TTL (Default: 5 minutes)                          │
│     if (now - timestamp > TTL) → EXPIRED                    │
│                                                             │
│  4. Check Replay (Nonce in Redis)                           │
│     if (nonce exists) → REPLAY_ATTEMPT                      │
│                                                             │
│  5. Bot Detection                                           │
│     ├── User-Agent analysis                                 │
│     ├── IP reputation check                                 │
│     └── Behavioral patterns                                 │
│                                                             │
│  6. Geo Rules                                               │
│     ├── Check offer geo rules                               │
│     ├── Check advertiser geo rules                          │
│     └── Check global geo rules                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

#### Step 3: Click Recording

If all validations pass, the click is recorded:

```json
{
  "click_id": "clk_a1b2c3d4e5f6",
  "user_offer_id": "uo_xyz789",
  "offer_id": "off_123456",
  "promoter_id": "usr_abc123",
  "ip_address": "203.0.113.42",
  "user_agent": "Mozilla/5.0...",
  "device": "mobile",
  "browser": "Chrome",
  "os": "Android",
  "country": "US",
  "city": "New York",
  "fingerprint": "fp_hash_abc123",
  "clicked_at": "2024-12-01T12:00:00Z"
}
```

#### Step 4: Redirect

The user is immediately redirected to the offer landing page:

```
HTTP/1.1 302 Found
Location: https://advertiser.com/offer?click_id=clk_a1b2c3d4e5f6&sub1=campaign1
```

### Phase 2: Conversion Attribution

#### Step 5: User Completes Action

The user completes the desired action on the advertiser's site (purchase, signup, etc.)

#### Step 6: Advertiser Sends Postback

The advertiser sends a postback to AffTok:

```bash
POST https://api.afftok.com/api/postback
Content-Type: application/json
X-API-Key: afftok_live_sk_xxxxx

{
  "click_id": "clk_a1b2c3d4e5f6",
  "transaction_id": "txn_advertiser_12345",
  "amount": 49.99,
  "currency": "USD",
  "status": "approved"
}
```

#### Step 7: Click Matching & Attribution

AffTok matches the postback to the original click:

```
┌─────────────────────────────────────────────────────────────┐
│                   ATTRIBUTION ENGINE                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Find Click by click_id                                  │
│     SELECT * FROM clicks WHERE id = 'clk_a1b2c3d4e5f6'      │
│                                                             │
│  2. Validate Attribution Window                             │
│     if (now - click_time > 30 days) → EXPIRED               │
│                                                             │
│  3. Check Duplicate Conversion                              │
│     if (transaction_id exists) → DUPLICATE                  │
│                                                             │
│  4. Create Conversion Record                                │
│     INSERT INTO conversions (...)                           │
│                                                             │
│  5. Update Statistics                                       │
│     ├── user_offers.total_conversions++                     │
│     ├── user_offers.earnings += payout                      │
│     ├── offers.total_conversions++                          │
│     └── users.total_earnings += payout                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

#### Step 8: Conversion Record

```json
{
  "conversion_id": "conv_xyz789abc",
  "click_id": "clk_a1b2c3d4e5f6",
  "user_offer_id": "uo_xyz789",
  "offer_id": "off_123456",
  "promoter_id": "usr_abc123",
  "transaction_id": "txn_advertiser_12345",
  "amount": 49.99,
  "currency": "USD",
  "payout": 10.00,
  "status": "approved",
  "converted_at": "2024-12-01T12:30:00Z"
}
```

#### Step 9: Webhook Notifications

AffTok triggers webhooks to notify interested parties:

```json
POST https://your-server.com/webhook
Content-Type: application/json
X-AffTok-Signature: sha256=abc123...

{
  "event": "conversion.created",
  "timestamp": 1701345600000,
  "data": {
    "conversion_id": "conv_xyz789abc",
    "offer_id": "off_123456",
    "amount": 49.99,
    "payout": 10.00,
    "status": "approved"
  }
}
```

## Click Deduplication

AffTok prevents duplicate clicks using fingerprinting:

```
Fingerprint = SHA256(
  IP Address +
  User Agent +
  Offer ID +
  Timestamp (rounded to minute)
)
```

Duplicate clicks within 60 seconds are rejected.

## Attribution Windows

| Attribution Type | Default Window |
|------------------|----------------|
| Click-to-Conversion | 30 days |
| View-through (future) | 24 hours |
| Multi-touch (future) | 30 days |

## Error Handling

### Click Errors

| Error | Behavior |
|-------|----------|
| Invalid signature | Log fraud event, redirect anyway |
| Expired link | Log warning, redirect anyway |
| Geo blocked | Log event, redirect to fallback |
| Bot detected | Log fraud event, block or redirect |

### Postback Errors

| Error | Response |
|-------|----------|
| Invalid API key | 401 Unauthorized |
| Click not found | 404 Not Found |
| Duplicate conversion | 409 Conflict |
| Invalid parameters | 400 Bad Request |

---

Next: [Smart Routing](./smart-routing.md)

