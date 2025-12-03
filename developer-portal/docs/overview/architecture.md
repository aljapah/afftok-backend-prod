# System Architecture

AffTok is built on a modern, scalable architecture designed for high-throughput affiliate tracking.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AFFTOK ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                   │
│  │   Mobile     │    │     Web      │    │   Server     │                   │
│  │    SDKs      │    │     SDK      │    │   to Server  │                   │
│  │ (iOS/Android │    │ (JavaScript) │    │   (API)      │                   │
│  │  Flutter/RN) │    │              │    │              │                   │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘                   │
│         │                   │                   │                            │
│         └───────────────────┼───────────────────┘                            │
│                             │                                                │
│                             ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                     CLOUDFLARE EDGE LAYER                            │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  │    │
│  │  │   Link      │  │    Bot      │  │    Geo      │  │   Smart    │  │    │
│  │  │ Validation  │  │  Detection  │  │   Rules     │  │   Router   │  │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                             │                                                │
│                             ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                       BACKEND API LAYER                              │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  │    │
│  │  │   Click     │  │  Postback   │  │   Stats     │  │   Admin    │  │    │
│  │  │  Handler    │  │  Handler    │  │   Handler   │  │   APIs     │  │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                             │                                                │
│         ┌───────────────────┼───────────────────┐                            │
│         ▼                   ▼                   ▼                            │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                    │
│  │   Redis     │     │ PostgreSQL  │     │   Redis     │                    │
│  │   Cache     │     │  Database   │     │  Streams    │                    │
│  │  (L1 + L2)  │     │ (Primary +  │     │  (Events)   │                    │
│  │             │     │  Replica)   │     │             │                    │
│  └─────────────┘     └─────────────┘     └─────────────┘                    │
│                             │                                                │
│                             ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                      ASYNC WORKERS                                   │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  │    │
│  │  │   Click     │  │  Postback   │  │  Analytics  │  │  Webhook   │  │    │
│  │  │  Worker     │  │  Worker     │  │   Worker    │  │  Worker    │  │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Edge Layer (Cloudflare Workers)

The Edge Layer handles all incoming traffic at the network edge, providing:

- **Link Validation**: HMAC signature verification, TTL checks, replay protection
- **Bot Detection**: User-agent analysis, behavioral patterns, datacenter IP detection
- **Geo Rules**: Country-level allow/block lists
- **Smart Router**: A/B testing, offer rotation, traffic distribution

### 2. Backend API Layer

The Go-based backend provides:

- **Click Handler**: Records clicks, generates fingerprints, deduplication
- **Postback Handler**: Processes conversions from advertisers
- **Stats Handler**: Aggregated analytics and reporting
- **Admin APIs**: System management and configuration

### 3. Data Layer

#### PostgreSQL (Primary Database)
- Clicks and conversions storage
- User and offer management
- Partitioned tables for high-volume data
- Read replicas for scalability

#### Redis (Caching & Streams)
- L1/L2 caching for hot data
- Click deduplication
- Rate limiting
- Event streaming for async processing

### 4. Async Workers

Background workers handle:

- Click batch processing
- Postback retries
- Analytics aggregation
- Webhook delivery with retry logic

## Data Flow

### Click Tracking Flow

```
User Click → Edge Layer → Validate Link → Check Geo Rules → 
Bot Detection → Record Click → Redirect to Offer
```

### Conversion Flow

```
Advertiser Postback → Validate API Key → Match Click → 
Create Conversion → Update Stats → Trigger Webhooks
```

## Scalability Features

| Feature | Implementation |
|---------|----------------|
| Horizontal Scaling | Kubernetes HPA/VPA |
| Database Scaling | Read replicas, partitioning |
| Cache Scaling | Redis cluster mode |
| Edge Scaling | Cloudflare global network |
| Async Processing | Worker pools with queues |

## High Availability

- **Multi-region deployment** via Cloudflare
- **Database failover** with automatic promotion
- **Zero-drop tracking** with WAL and local queues
- **Circuit breakers** for external dependencies

## Security Layers

1. **Network**: Cloudflare WAF, DDoS protection
2. **Transport**: TLS 1.3, certificate pinning
3. **Application**: API key auth, HMAC signatures
4. **Data**: Encryption at rest, tenant isolation

---

Next: [Tracking Flow](./tracking-flow.md)

