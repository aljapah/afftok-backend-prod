# Smart Routing

AffTok's Smart Routing engine enables intelligent traffic distribution based on multiple criteria including geography, device type, offer caps, A/B testing, and performance optimization.

## Overview

Smart Routing operates at the edge layer, making routing decisions in under 10ms before redirecting users to the optimal destination.

```
┌─────────────────────────────────────────────────────────────────┐
│                      Smart Routing Engine                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Input: Click Request                                           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  • User Agent      • IP Address       • Country          │   │
│  │  • Device Type     • Connection       • Time of Day      │   │
│  │  • Referrer        • Tracking Code    • Sub IDs          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                           │                                     │
│                           ▼                                     │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                   Routing Rules                          │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐    │   │
│  │  │   Geo   │  │ Device  │  │  Caps   │  │  A/B    │    │   │
│  │  │  Rules  │  │  Rules  │  │  Check  │  │  Test   │    │   │
│  │  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘    │   │
│  │       │            │            │            │          │   │
│  │       └────────────┴────────────┴────────────┘          │   │
│  │                           │                              │   │
│  │                           ▼                              │   │
│  │  ┌─────────────────────────────────────────────────┐    │   │
│  │  │              Offer Selection                     │    │   │
│  │  │  • Round Robin    • Weighted     • Smart CTR    │    │   │
│  │  └─────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────┘   │
│                           │                                     │
│                           ▼                                     │
│  Output: Final Destination URL                                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Routing Criteria

### Geographic Routing

Route traffic based on user location:

```json
{
  "rule_type": "geo",
  "conditions": {
    "countries": ["US", "CA", "GB"],
    "mode": "allow"
  },
  "action": {
    "type": "route",
    "destination": "offer_us_123"
  },
  "fallback": {
    "destination": "offer_global_456"
  }
}
```

**Supported Geo Criteria:**
- Country (ISO 3166-1 alpha-2)
- Region/State
- City
- ISP/ASN
- Connection type (mobile, broadband, etc.)

### Device Routing

Route based on device characteristics:

```json
{
  "rule_type": "device",
  "conditions": {
    "device_types": ["mobile", "tablet"],
    "os": ["iOS", "Android"],
    "min_os_version": "12.0"
  },
  "action": {
    "type": "route",
    "destination": "offer_mobile_789"
  }
}
```

**Supported Device Criteria:**
- Device type (mobile, tablet, desktop)
- Operating system
- OS version
- Browser
- Browser version
- Screen resolution

### Offer Caps

Limit traffic to offers based on caps:

```json
{
  "rule_type": "cap",
  "offer_id": "off_123",
  "caps": {
    "daily_clicks": 10000,
    "daily_conversions": 500,
    "monthly_budget": 50000.00
  },
  "on_cap_reached": {
    "action": "fallback",
    "destination": "offer_backup_456"
  }
}
```

**Cap Types:**
- Daily clicks
- Daily conversions
- Daily revenue
- Monthly budget
- Total conversions
- Per-affiliate caps

### A/B Testing

Split traffic between variations:

```json
{
  "rule_type": "ab_test",
  "test_id": "test_landing_page_v2",
  "variations": [
    {
      "id": "control",
      "weight": 50,
      "destination": "https://example.com/landing-v1"
    },
    {
      "id": "variant_a",
      "weight": 30,
      "destination": "https://example.com/landing-v2"
    },
    {
      "id": "variant_b",
      "weight": 20,
      "destination": "https://example.com/landing-v3"
    }
  ],
  "tracking": {
    "track_variation": true,
    "sub_id_field": "sub3"
  }
}
```

## Routing Modes

### Round Robin

Distribute traffic evenly across offers:

```
Offer A ─────┐
             │
Offer B ─────┼──── Round Robin ──── 33% each
             │
Offer C ─────┘
```

### Weighted Distribution

Distribute based on configured weights:

```
Offer A (weight: 50) ─────┐
                          │
Offer B (weight: 30) ─────┼──── Weighted ──── A: 50%, B: 30%, C: 20%
                          │
Offer C (weight: 20) ─────┘
```

### Smart CTR Optimization

Automatically optimize based on performance:

```
┌─────────────────────────────────────────────────────────────┐
│                   Smart CTR Algorithm                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Collect performance data (last 24 hours)                │
│     ├── Offer A: 1000 clicks, 50 conversions (5% CR)        │
│     ├── Offer B: 800 clicks, 60 conversions (7.5% CR)       │
│     └── Offer C: 600 clicks, 24 conversions (4% CR)         │
│                                                             │
│  2. Calculate weights based on CR                           │
│     ├── Offer A: 5 / 16.5 = 30.3%                           │
│     ├── Offer B: 7.5 / 16.5 = 45.5%                         │
│     └── Offer C: 4 / 16.5 = 24.2%                           │
│                                                             │
│  3. Apply exploration factor (10%)                          │
│     └── 10% traffic goes to random offer for exploration    │
│                                                             │
│  4. Route traffic accordingly                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Multi-Offer Fallback

Configure fallback chains for reliability:

```json
{
  "primary_offer": "off_123",
  "fallback_chain": [
    {
      "offer_id": "off_456",
      "trigger": "cap_reached"
    },
    {
      "offer_id": "off_789",
      "trigger": "geo_blocked"
    },
    {
      "url": "https://fallback.example.com",
      "trigger": "all_failed"
    }
  ]
}
```

**Fallback Triggers:**
- `cap_reached` - Offer hit daily/monthly cap
- `geo_blocked` - User's country not allowed
- `device_blocked` - User's device not supported
- `offer_paused` - Offer temporarily disabled
- `offer_expired` - Offer end date passed
- `all_failed` - No other offers available

## Network Failover

Handle network-level failures:

```
┌─────────────────────────────────────────────────────────────┐
│                   Network Failover                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Primary Network                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Network A (Priority 1)                              │   │
│  │  ├── Health check: every 30s                         │   │
│  │  ├── Timeout: 5s                                     │   │
│  │  └── Status: HEALTHY                                 │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                 │
│                           ▼ (if unhealthy)                  │
│  Failover Network                                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Network B (Priority 2)                              │   │
│  │  ├── Auto-activate on primary failure                │   │
│  │  └── Status: STANDBY                                 │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Configuration Examples

### Geo-Based Routing

Route US traffic to US offer, others to global:

```json
{
  "name": "US Traffic Router",
  "rules": [
    {
      "priority": 1,
      "conditions": {
        "geo": {
          "countries": ["US"],
          "mode": "allow"
        }
      },
      "destination": {
        "offer_id": "off_us_premium",
        "url": "https://us.example.com/offer"
      }
    },
    {
      "priority": 2,
      "conditions": {
        "default": true
      },
      "destination": {
        "offer_id": "off_global",
        "url": "https://global.example.com/offer"
      }
    }
  ]
}
```

### Device + Geo Routing

Mobile users in specific countries:

```json
{
  "name": "Mobile App Install Campaign",
  "rules": [
    {
      "priority": 1,
      "conditions": {
        "device": {
          "types": ["mobile"],
          "os": ["iOS"]
        },
        "geo": {
          "countries": ["US", "CA", "GB", "AU"]
        }
      },
      "destination": {
        "offer_id": "off_ios_tier1",
        "url": "https://apps.apple.com/app/id123456"
      }
    },
    {
      "priority": 2,
      "conditions": {
        "device": {
          "types": ["mobile"],
          "os": ["Android"]
        },
        "geo": {
          "countries": ["US", "CA", "GB", "AU"]
        }
      },
      "destination": {
        "offer_id": "off_android_tier1",
        "url": "https://play.google.com/store/apps/details?id=com.example"
      }
    },
    {
      "priority": 3,
      "conditions": {
        "default": true
      },
      "destination": {
        "url": "https://example.com/not-available"
      }
    }
  ]
}
```

### Time-Based Routing

Route based on time of day:

```json
{
  "name": "Time-Based Router",
  "rules": [
    {
      "priority": 1,
      "conditions": {
        "time": {
          "hours": [9, 10, 11, 12, 13, 14, 15, 16, 17],
          "timezone": "America/New_York"
        }
      },
      "destination": {
        "offer_id": "off_business_hours"
      }
    },
    {
      "priority": 2,
      "conditions": {
        "default": true
      },
      "destination": {
        "offer_id": "off_after_hours"
      }
    }
  ]
}
```

## Routing Decision Response

Each routing decision returns metadata:

```json
{
  "decision": {
    "final_destination": "https://example.com/offer?click_id=abc123",
    "offer_id": "off_123",
    "rule_applied": "geo_us_premium",
    "router_source": "edge",
    "latency_ms": 3,
    "cache_hit": true
  },
  "context": {
    "country": "US",
    "device": "mobile",
    "os": "iOS",
    "variation": "control"
  },
  "fallbacks_tried": []
}
```

## Performance

| Metric | Target | Typical |
|--------|--------|---------|
| Routing decision | < 10ms | 3-5ms |
| Cache hit rate | > 90% | 95% |
| Rule evaluation | < 5ms | 1-2ms |
| Fallback latency | < 20ms | 10ms |

## Best Practices

1. **Order rules by priority** - Most specific rules first
2. **Always have a default** - Catch-all rule prevents lost traffic
3. **Use caching** - Cache routing decisions for repeated visitors
4. **Monitor performance** - Track which rules are triggering
5. **Test thoroughly** - Use sandbox to validate routing logic
6. **Set reasonable caps** - Avoid over-delivery

## Next Steps

- [Geo Rules API](../api-reference/geo-rules.md) - Configure geographic restrictions
- [Offers API](../api-reference/offers.md) - Manage offer destinations
- [Analytics](../api-reference/stats.md) - Monitor routing performance

