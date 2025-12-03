# Smart Routing

AffTok's Smart Routing system intelligently distributes traffic based on multiple factors to maximize conversion rates and revenue.

## Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SMART ROUTING ENGINE                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  INCOMING CLICK                                                              │
│       │                                                                      │
│       ▼                                                                      │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                    │
│  │    GEO      │────>│   DEVICE    │────>│    ISP      │                    │
│  │   FILTER    │     │   FILTER    │     │   FILTER    │                    │
│  └─────────────┘     └─────────────┘     └─────────────┘                    │
│                             │                                                │
│                             ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                     ROUTING STRATEGY                                 │    │
│  │                                                                      │    │
│  │   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │    │
│  │   │ Round Robin │  │  Weighted   │  │  Smart CTR  │                 │    │
│  │   │             │  │             │  │             │                 │    │
│  │   │  A → B → C  │  │ A:50% B:30% │  │ Best perf.  │                 │    │
│  │   │  → A → B →  │  │    C:20%    │  │   first     │                 │    │
│  │   └─────────────┘  └─────────────┘  └─────────────┘                 │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                             │                                                │
│                             ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                      OFFER SELECTION                                 │    │
│  │                                                                      │    │
│  │   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │    │
│  │   │  Cap Check  │  │   A/B Test  │  │  Failover   │                 │    │
│  │   │             │  │             │  │             │                 │    │
│  │   │ Daily/Total │  │ Experiment  │  │  Fallback   │                 │    │
│  │   │   limits    │  │   groups    │  │   offers    │                 │    │
│  │   └─────────────┘  └─────────────┘  └─────────────┘                 │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                             │                                                │
│                             ▼                                                │
│                    FINAL DESTINATION URL                                     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Routing Strategies

### 1. Round Robin

Distributes traffic evenly across all eligible offers in sequence.

```json
{
  "strategy": "round_robin",
  "offers": ["offer_a", "offer_b", "offer_c"]
}
```

**Distribution**: A → B → C → A → B → C → ...

### 2. Weighted Distribution

Distributes traffic according to specified weights.

```json
{
  "strategy": "weighted",
  "offers": [
    { "offer_id": "offer_a", "weight": 50 },
    { "offer_id": "offer_b", "weight": 30 },
    { "offer_id": "offer_c", "weight": 20 }
  ]
}
```

**Distribution**: 50% → A, 30% → B, 20% → C

### 3. Smart CTR (Performance-Based)

Automatically routes more traffic to better-performing offers based on conversion rate.

```json
{
  "strategy": "smart_ctr",
  "offers": ["offer_a", "offer_b", "offer_c"],
  "learning_period": "24h",
  "exploration_rate": 0.1
}
```

**Algorithm**:
1. Initial exploration phase (equal distribution)
2. Calculate CTR for each offer
3. Allocate traffic proportional to performance
4. Maintain 10% exploration for new data

## Filtering Rules

### Geo Filtering

```json
{
  "geo_rules": {
    "mode": "allow",
    "countries": ["US", "CA", "GB", "AU"],
    "fallback_offer": "offer_global"
  }
}
```

### Device Filtering

```json
{
  "device_rules": {
    "mobile": "offer_mobile",
    "desktop": "offer_desktop",
    "tablet": "offer_mobile"
  }
}
```

### ISP/Carrier Filtering

```json
{
  "isp_rules": {
    "wifi": "offer_standard",
    "cellular": "offer_mobile_optimized",
    "blocked_isps": ["datacenter_provider_1"]
  }
}
```

## Offer Caps

### Daily Caps

```json
{
  "caps": {
    "daily_clicks": 10000,
    "daily_conversions": 500,
    "action_on_cap": "redirect_to_fallback"
  }
}
```

### Total Caps

```json
{
  "caps": {
    "total_clicks": 100000,
    "total_conversions": 5000,
    "total_revenue": 50000.00
  }
}
```

### Cap Behavior

| Action | Description |
|--------|-------------|
| `redirect_to_fallback` | Send to fallback offer |
| `pause_offer` | Stop accepting traffic |
| `notify_only` | Continue but send alert |

## A/B Testing

### Creating an A/B Test

```json
{
  "ab_test": {
    "name": "Landing Page Test",
    "variants": [
      { "id": "control", "offer_id": "offer_a", "weight": 50 },
      { "id": "variant_b", "offer_id": "offer_b", "weight": 50 }
    ],
    "metrics": ["ctr", "conversion_rate", "revenue"],
    "duration": "7d",
    "min_sample_size": 1000
  }
}
```

### Test Results

```json
{
  "ab_test_results": {
    "test_id": "test_123",
    "status": "completed",
    "winner": "variant_b",
    "confidence": 0.95,
    "results": {
      "control": {
        "clicks": 5000,
        "conversions": 150,
        "ctr": 0.03,
        "revenue": 1500.00
      },
      "variant_b": {
        "clicks": 5000,
        "conversions": 200,
        "ctr": 0.04,
        "revenue": 2000.00
      }
    }
  }
}
```

## Failover System

### Multi-Level Failover

```
Primary Offer (offer_a)
    │
    │ [Cap reached / Error]
    ▼
Secondary Offer (offer_b)
    │
    │ [Cap reached / Error]
    ▼
Tertiary Offer (offer_c)
    │
    │ [Cap reached / Error]
    ▼
Global Fallback URL
```

### Configuration

```json
{
  "failover": {
    "primary": "offer_a",
    "secondary": ["offer_b", "offer_c"],
    "fallback_url": "https://fallback.example.com",
    "triggers": ["cap_reached", "offer_paused", "geo_blocked"]
  }
}
```

## Smart Decision Cache

Routing decisions are cached at the edge for performance:

```
Cache Key: {tenant_id}:{offer_id}:{country}:{device}
Cache TTL: 5 minutes
```

### Cache Invalidation

Caches are invalidated when:
- Offer configuration changes
- Cap thresholds are reached
- A/B test status changes
- Manual cache clear

## API Endpoints

### Get Routing Config

```bash
GET /api/admin/edge/router
Authorization: Bearer <admin_token>
```

### Update Routing Rules

```bash
PUT /api/admin/offers/{offer_id}/routing
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "strategy": "weighted",
  "geo_rules": {...},
  "caps": {...}
}
```

## Metrics & Monitoring

### Routing Metrics

| Metric | Description |
|--------|-------------|
| `router_decisions_total` | Total routing decisions |
| `router_latency_ms` | Decision latency |
| `router_cache_hits` | Cache hit rate |
| `router_failovers` | Failover events |
| `offer_cap_reached` | Cap events |

### Dashboard

Access routing analytics at:
```
https://dashboard.afftok.com/routing
```

---

Next: [Quick Start Guide](../quickstart/README.md)

