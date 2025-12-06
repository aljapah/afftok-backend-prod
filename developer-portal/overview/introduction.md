# Introduction to AffTok

AffTok is an enterprise-grade affiliate tracking platform designed for high-performance, real-time tracking of clicks, conversions, and affiliate attribution. It operates as a **Two-Sided Marketplace** connecting promoters with advertisers.

## What is AffTok?

AffTok provides a complete tracking infrastructure for:

- **Promoters (مروّجين)** - Market offers, track performance, earn points, compete in leaderboards
- **Advertisers (معلنين)** - Create offers, manage campaigns, track promoter performance
- **Admins** - Full platform management, approval workflows, analytics

## Business Model

AffTok is **100% free** for both promoters and advertisers:
- No subscription fees
- No transaction fees
- Platform focuses on **tracking only**
- Companies pay promoters directly
- AffTok earns commission from companies

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

### 🤖 AI Assistant (NEW)

- Personal statistics dashboard
- Smart offer suggestions
- Performance analysis with recommendations
- Success tips and strategies
- Points & levels gamification
- BYOK (Bring Your Own Key) for advanced chat

### 👥 Teams & Contests (NEW)

- Create and join teams
- Global leaderboard rankings
- Competitive contests with prizes
- Team-based challenges

### 🏢 Advertisers Portal (NEW)

- Advertiser registration and onboarding
- Offer creation with approval workflow
- Performance analytics
- Promoter management

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
│  │   Mobile     │  │   Backend    │  │   Admin      │          │
│  │   App        │──│   (Go)       │──│   Panel      │          │
│  │  (Flutter)   │  │              │  │   (React)    │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                 │                 │                   │
│         ▼                 ▼                 ▼                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   AI         │  │   Redis      │  │   PostgreSQL │          │
│  │   Assistant  │──│   Cache      │──│   Database   │          │
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

## User Roles

| Role | Description | Access |
|------|-------------|--------|
| Promoter | Markets offers using tracking links | Mobile App |
| Advertiser | Creates and manages offers | Mobile App |
| Admin | Platform administrator | Admin Panel |

## Mobile App Features

### For Promoters
- Browse and join offers
- Get unique tracking links
- View personal statistics
- AI Assistant for guidance
- Join teams and contests
- Track points and levels
- Global leaderboard

### For Advertisers
- Create company account
- Submit offers for approval
- Track offer performance
- View promoter statistics

## Use Cases

### Affiliate Marketing

Track affiliate referrals, attribute conversions, and monitor performance in real-time.

### Influencer Marketing

Provide influencers with unique tracking links to measure their campaign effectiveness.

### E-commerce Promotion

Partner with promoters to drive traffic and sales to your online store.

### Lead Generation

Track form submissions, sign-ups, and qualified leads across campaigns.

## Getting Started

### For Promoters
1. **Download the App** - Available on iOS and Android
2. **Create Account** - Sign up as a promoter
3. **Browse Offers** - Find offers that match your audience
4. **Share Links** - Get your unique tracking link
5. **Track Performance** - Monitor clicks and conversions

### For Advertisers
1. **Download the App** - Available on iOS and Android
2. **Register Company** - Provide company details
3. **Create Offers** - Submit offers for approval
4. **Track Results** - Monitor promoter performance

### For Developers
1. **Get API Key** - Generate from Admin Panel
2. **Read Documentation** - Explore API reference
3. **Integrate SDKs** - Use native SDKs for tracking
4. **Test & Deploy** - Use sandbox for testing

## API Base URLs

| Environment | Base URL |
|-------------|----------|
| Production | `https://afftok-backend-prod-production.up.railway.app/api` |
| Website | `https://afftokapp.com` |

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
- [Quick Start Guide](../quick-start/getting-started.md) - Get up and running
- [API Reference](../api-reference/authentication.md) - Explore the full API
- [AI Assistant](../docs/features/ai-assistant.md) - Learn about the AI features
- [Teams & Contests](../docs/features/teams-contests.md) - Explore social features

## Support

Need help? Contact us:

- **Email**: support@afftokapp.com
- **Website**: [afftokapp.com](https://afftokapp.com)
- **Twitter**: [@afftokapp](https://twitter.com/afftokapp)
- **Instagram**: [@afftokapp](https://instagram.com/afftokapp)

---

*Last updated: December 2024*
