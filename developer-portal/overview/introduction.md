# Introduction to AffTok

AffTok is an enterprise-grade affiliate tracking platform designed for high-performance, real-time tracking of clicks, conversions, and affiliate attribution.

## What is AffTok?

AffTok provides a complete tracking infrastructure for:

- **Advertisers** - Track campaign performance, manage offers, and process conversions
- **Publishers/Affiliates** - Generate tracking links, monitor clicks, and earn commissions
- **Networks** - Manage multiple advertisers and publishers with multi-tenant isolation

## Key Features

### 🚀 High-Performance Tracking

- Process millions of clicks per day
- Sub-50ms redirect latency
- Global CDN edge routing
- Zero-drop tracking guarantee

### 🔐 Enterprise Security

- HMAC-SHA256 signed tracking links
- API key authentication with rotation
- Rate limiting and fraud detection
- Geo-blocking and IP allowlisting
- Multi-tenant data isolation

### 📊 Real-Time Analytics

- Live click and conversion stats
- Daily/weekly/monthly breakdowns
- Offer-level and user-level metrics
- Fraud detection insights

### 🔄 Flexible Integrations

- Native SDKs (Android, iOS, Flutter, React Native, Web)
- Server-to-server postback support
- Webhook delivery with retry logic
- RESTful API for all operations

## Platform Components

```
┌─────────────────────────────────────────────────────────────────┐
│                        AffTok Platform                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Edge CDN   │  │   Backend    │  │   Admin      │          │
│  │   Routing    │──│   API        │──│   Panel      │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                 │                 │                   │
│         ▼                 ▼                 ▼                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Click      │  │   Redis      │  │   PostgreSQL │          │
│  │   Processor  │──│   Cache      │──│   Database   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                 │                 │                   │
│         ▼                 ▼                 ▼                   │
│  ┌──────────────────────────────────────────────────┐          │
│  │              Observability Layer                 │          │
│  │     (Logging, Metrics, Fraud Detection)          │          │
│  └──────────────────────────────────────────────────┘          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Use Cases

### Affiliate Marketing

Track affiliate referrals, attribute conversions, and calculate commissions automatically.

### Mobile App Install Tracking

Use native SDKs to track app installs and in-app events with device fingerprinting.

### E-commerce Conversion Tracking

Integrate with your checkout flow to track purchases and revenue.

### Lead Generation

Track form submissions, sign-ups, and qualified leads across campaigns.

## Getting Started

1. **Create an Account** - Sign up at [dashboard.afftok.com](https://dashboard.afftok.com)
2. **Generate API Key** - Navigate to Settings → API Keys
3. **Create an Offer** - Set up your first tracking offer
4. **Integrate** - Use our SDKs or API to start tracking

## API Base URLs

| Environment | Base URL |
|-------------|----------|
| Production | `https://api.afftok.com` |
| Sandbox | `https://sandbox.api.afftok.com` |

## SDK Support

| Platform | Package |
|----------|---------|
| Android | `com.afftok:afftok-sdk:1.0.0` |
| iOS | `AfftokSDK` (Swift Package) |
| Flutter | `afftok: ^1.0.0` |
| React Native | `@afftok/react-native-sdk` |
| Web | `https://cdn.afftok.com/sdk/afftok.min.js` |

## Next Steps

- [Architecture Overview](architecture.md) - Understand how AffTok works
- [Quick Start Guide](../quick-start/getting-started.md) - Get up and running in 5 minutes
- [API Reference](../api-reference/authentication.md) - Explore the full API

## Support

Need help? Contact us:

- **Email**: support@afftok.com
- **Documentation**: https://docs.afftok.com
- **Status Page**: https://status.afftok.com

