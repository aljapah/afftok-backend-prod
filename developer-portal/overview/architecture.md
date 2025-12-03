# System Architecture

AffTok is built on a modern, scalable architecture designed for high-throughput tracking with enterprise-grade reliability.

## Architecture Diagram

```
                                    ┌─────────────────────────────────────┐
                                    │           Client Layer              │
                                    ├─────────────────────────────────────┤
                                    │  Mobile Apps │ Web │ Landing Pages  │
                                    │  (SDK)       │(SDK)│ (JS/Pixel)     │
                                    └───────────────┬─────────────────────┘
                                                    │
                                                    ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                              Edge Layer (CDN)                                 │
├───────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │  Cloudflare │  │   Fastly    │  │   Vercel    │  │   Custom    │          │
│  │   Worker    │  │   Compute   │  │   Edge      │  │   Edge      │          │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘          │
│         │                │                │                │                  │
│         └────────────────┴────────────────┴────────────────┘                  │
│                                   │                                           │
│  Features:                        │                                           │
│  • Link Validation               │                                           │
│  • Bot Detection                 │                                           │
│  • Geo Validation                │                                           │
│  • Smart Routing                 │                                           │
│  • Instant Redirect              │                                           │
└───────────────────────────────────┬───────────────────────────────────────────┘
                                    │
                                    ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                            API Gateway Layer                                  │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐               │
│  │  Rate Limiter   │  │  Auth Middleware│  │  Tenant Resolver│               │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘               │
│           │                    │                    │                         │
│           └────────────────────┴────────────────────┘                         │
│                                │                                              │
└────────────────────────────────┬──────────────────────────────────────────────┘
                                 │
                                 ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                           Application Layer                                   │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   Click     │  │  Conversion │  │   Stats     │  │   Admin     │          │
│  │   Handler   │  │   Handler   │  │   Handler   │  │   Handler   │          │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘          │
│         │                │                │                │                  │
│  ┌──────┴──────┐  ┌──────┴──────┐  ┌──────┴──────┐  ┌──────┴──────┐          │
│  │   Click     │  │  Postback   │  │  Analytics  │  │   Webhook   │          │
│  │   Service   │  │   Service   │  │   Service   │  │   Service   │          │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘          │
│         │                │                │                │                  │
│  ┌──────┴──────┐  ┌──────┴──────┐  ┌──────┴──────┐  ┌──────┴──────┐          │
│  │   Link      │  │   Fraud     │  │   Geo Rule  │  │   API Key   │          │
│  │   Signing   │  │   Detection │  │   Service   │  │   Service   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘          │
│                                                                               │
└───────────────────────────────────┬───────────────────────────────────────────┘
                                    │
                                    ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                             Data Layer                                        │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────────────────────┐  ┌─────────────────────────────┐            │
│  │         PostgreSQL          │  │           Redis             │            │
│  ├─────────────────────────────┤  ├─────────────────────────────┤            │
│  │ • Users                     │  │ • Session Cache             │            │
│  │ • Offers                    │  │ • Click Deduplication       │            │
│  │ • Clicks (Partitioned)      │  │ • Rate Limit Counters       │            │
│  │ • Conversions               │  │ • Stats Cache               │            │
│  │ • API Keys                  │  │ • Geo Rules Cache           │            │
│  │ • Geo Rules                 │  │ • Replay Protection         │            │
│  │ • Webhooks                  │  │ • Stream Queues             │            │
│  │ • Tenants                   │  │                             │            │
│  └─────────────────────────────┘  └─────────────────────────────┘            │
│                                                                               │
│  ┌─────────────────────────────┐  ┌─────────────────────────────┐            │
│  │      Write-Ahead Log        │  │      Redis Streams          │            │
│  ├─────────────────────────────┤  ├─────────────────────────────┤            │
│  │ • Crash Recovery            │  │ • Click Events              │            │
│  │ • Zero-Drop Guarantee       │  │ • Conversion Events         │            │
│  │ • Event Replay              │  │ • Postback Events           │            │
│  └─────────────────────────────┘  └─────────────────────────────┘            │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

## Component Overview

### Edge Layer

The edge layer handles initial request processing at CDN locations worldwide:

| Component | Purpose |
|-----------|---------|
| **Link Validator** | Validates HMAC signatures, TTL, and nonces |
| **Bot Detector** | Filters bot traffic using UA analysis and heuristics |
| **Geo Validator** | Enforces geographic restrictions |
| **Smart Router** | Routes traffic based on rules (A/B testing, caps, etc.) |
| **Click Queue** | Buffers clicks for async backend delivery |

### API Gateway

The gateway layer provides security and routing:

| Component | Purpose |
|-----------|---------|
| **Rate Limiter** | Enforces per-IP and per-key rate limits |
| **Auth Middleware** | Validates JWT tokens and API keys |
| **Tenant Resolver** | Identifies tenant from headers, domain, or API key |

### Application Layer

Core business logic services:

| Service | Purpose |
|---------|---------|
| **Click Service** | Processes and stores click events |
| **Postback Service** | Handles conversion postbacks |
| **Analytics Service** | Aggregates statistics |
| **Webhook Service** | Manages webhook delivery |
| **Link Signing Service** | Generates and validates signed links |
| **Fraud Detection** | Identifies suspicious activity |
| **Geo Rule Service** | Evaluates geographic restrictions |

### Data Layer

Persistent storage and caching:

| Component | Purpose |
|-----------|---------|
| **PostgreSQL** | Primary data store (users, offers, clicks, conversions) |
| **Redis** | Caching, rate limiting, deduplication, queues |
| **WAL** | Write-ahead log for zero-drop guarantee |
| **Redis Streams** | Event streaming for async processing |

## Data Flow

### Click Tracking Flow

```
User Click → Edge CDN → Validate Link → Check Geo Rules → Check Bot
     │                                                        │
     │                                                        ▼
     │                                              Record Click (Async)
     │                                                        │
     └──────────────────── Redirect ──────────────────────────┘
                              │
                              ▼
                       Destination URL
```

### Conversion Flow

```
Conversion Event → API Gateway → Auth Check → Validate Signature
                                                      │
                                                      ▼
                                            Find Associated Click
                                                      │
                                                      ▼
                                            Create Conversion Record
                                                      │
                                                      ▼
                                            Update Stats (Atomic)
                                                      │
                                                      ▼
                                            Trigger Webhooks
```

## Technology Stack

| Layer | Technology |
|-------|------------|
| Edge | Cloudflare Workers, TypeScript |
| Backend | Go 1.21+, Gin Framework |
| Database | PostgreSQL 15+ |
| Cache | Redis 7+ |
| Message Queue | Redis Streams |
| Container | Docker, Kubernetes |
| CDN | Cloudflare, Fastly |

## Scalability Features

### Horizontal Scaling

- Stateless API servers behind load balancer
- Redis cluster for distributed caching
- PostgreSQL read replicas for read scaling
- Kubernetes HPA for auto-scaling

### Performance Optimizations

- L1 (in-memory) + L2 (Redis) caching
- Click table partitioning by month
- Async processing with worker pools
- Connection pooling for DB and Redis

### High Availability

- Multi-region deployment
- Automatic failover
- Zero-drop tracking with WAL
- Circuit breakers for external services

## Security Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Security Layers                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Layer 1: Edge Security                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ • WAF Rules           • DDoS Protection             │   │
│  │ • Bot Detection       • Rate Limiting               │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Layer 2: Transport Security                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ • TLS 1.3             • Certificate Pinning         │   │
│  │ • HSTS                • Secure Headers              │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Layer 3: Application Security                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ • HMAC Signatures     • JWT Authentication          │   │
│  │ • API Key Hashing     • Input Validation            │   │
│  │ • Replay Protection   • Tenant Isolation            │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Layer 4: Data Security                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ • Encryption at Rest  • Encryption in Transit       │   │
│  │ • Key Rotation        • Audit Logging               │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Next Steps

- [Tracking Flow](tracking-flow.md) - Detailed click-to-conversion flow
- [Smart Routing](smart-routing.md) - Traffic routing and optimization
- [Quick Start](../quick-start/getting-started.md) - Get started with integration

