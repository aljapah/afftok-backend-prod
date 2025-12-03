# Operational Guidelines

Best practices for operating and maintaining your AffTok integration.

## Company Integration

### Onboarding Checklist

1. **Account Setup**
   - Create advertiser account
   - Generate API keys
   - Configure webhook endpoints
   - Set up geo rules

2. **Technical Integration**
   - Implement SDK/API integration
   - Configure postback URL
   - Test in sandbox environment
   - Verify signature handling

3. **Go-Live**
   - Switch to production keys
   - Enable monitoring
   - Configure alerts
   - Document integration

---

## Required Postback Data

### Minimum Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `click_id` | string | Original click ID from tracking |
| `external_id` | string | Your unique conversion ID |
| `amount` | number | Conversion value |
| `status` | string | approved/pending/rejected |

### Recommended Fields

| Field | Type | Description |
|-------|------|-------------|
| `currency` | string | ISO currency code |
| `product_id` | string | Product identifier |
| `order_id` | string | Order reference |
| `customer_id` | string | Customer identifier |
| `metadata` | object | Additional data |

### Example Postback

```json
{
  "click_id": "click_abc123",
  "external_id": "order_456789",
  "amount": 49.99,
  "currency": "USD",
  "status": "approved",
  "product_id": "prod_premium",
  "order_id": "ORD-2024-001234",
  "metadata": {
    "plan": "annual",
    "source": "web",
    "coupon": "SAVE20"
  }
}
```

---

## API Key Management

### Key Rotation Schedule

| Key Type | Rotation Frequency |
|----------|-------------------|
| Production | Every 90 days |
| Test | Every 180 days |
| Admin | Every 30 days |

### Rotation Process

1. **Generate new key**

   ```bash
   curl -X POST https://api.afftok.com/api/admin/api-keys/key_123/rotate \
     -H "Authorization: Bearer ADMIN_TOKEN"
   ```

2. **Update integrations**
   - Update environment variables
   - Deploy changes
   - Verify functionality

3. **Monitor old key usage**
   - Check for any remaining traffic
   - Investigate if old key still in use

4. **Revoke old key**

   ```bash
   curl -X POST https://api.afftok.com/api/admin/api-keys/old_key_123/revoke \
     -H "Authorization: Bearer ADMIN_TOKEN"
   ```

### Emergency Key Revocation

If a key is compromised:

1. **Immediately revoke**

   ```bash
   curl -X POST https://api.afftok.com/api/admin/api-keys/key_123/revoke \
     -H "Authorization: Bearer ADMIN_TOKEN" \
     -d '{"reason": "Security incident"}'
   ```

2. **Generate replacement**

3. **Review audit logs**

4. **Notify affected parties**

---

## Monitoring Best Practices

### Key Metrics to Monitor

| Metric | Alert Threshold |
|--------|-----------------|
| Click latency | > 100ms avg |
| Error rate | > 1% |
| Conversion rate | < 50% of baseline |
| Webhook failures | > 5% |
| Queue depth | > 1000 |

### Dashboard Setup

Track these in your monitoring system:

```
# Prometheus metrics example
afftok_clicks_total
afftok_conversions_total
afftok_click_latency_seconds
afftok_webhook_failures_total
afftok_queue_depth
afftok_error_rate
```

### Alert Configuration

```yaml
# Example alert rules
alerts:
  - name: high_error_rate
    condition: error_rate > 0.01
    duration: 5m
    severity: warning
    
  - name: click_latency_high
    condition: avg_click_latency > 100ms
    duration: 10m
    severity: warning
    
  - name: webhook_failures
    condition: webhook_failure_rate > 0.05
    duration: 5m
    severity: critical
    
  - name: queue_backlog
    condition: queue_depth > 1000
    duration: 15m
    severity: warning
```

---

## Incident Response

### Severity Levels

| Level | Description | Response Time |
|-------|-------------|---------------|
| P1 | Complete outage | 15 minutes |
| P2 | Major degradation | 1 hour |
| P3 | Minor issues | 4 hours |
| P4 | Low impact | 24 hours |

### Response Procedures

#### P1: Complete Outage

1. **Acknowledge** - Confirm incident within 15 minutes
2. **Assess** - Determine scope and impact
3. **Communicate** - Notify stakeholders
4. **Mitigate** - Implement temporary fix
5. **Resolve** - Deploy permanent solution
6. **Post-mortem** - Document and learn

#### P2: Major Degradation

1. **Acknowledge** - Confirm within 1 hour
2. **Investigate** - Identify root cause
3. **Communicate** - Update status page
4. **Fix** - Deploy solution
5. **Verify** - Confirm resolution
6. **Document** - Update runbooks

### Runbook: High Error Rate

```markdown
## High Error Rate Runbook

### Symptoms
- Error rate > 1%
- Increased 4xx/5xx responses

### Investigation
1. Check API logs for error patterns
2. Review recent deployments
3. Check database connectivity
4. Verify Redis health
5. Review rate limit status

### Resolution
1. If deployment issue: rollback
2. If database issue: failover to replica
3. If Redis issue: restart/failover
4. If rate limit: adjust thresholds

### Escalation
- If unresolved after 30 minutes: escalate to on-call
```

---

## Capacity Planning

### Traffic Estimation

| Metric | Calculation |
|--------|-------------|
| Daily clicks | Offers × Avg clicks/offer |
| Daily conversions | Clicks × Conversion rate |
| API calls | Clicks + Conversions + Stats queries |
| Storage | Clicks × 1KB + Conversions × 2KB |

### Scaling Guidelines

| Traffic Level | Recommended Setup |
|---------------|-------------------|
| < 10K clicks/day | Single instance |
| 10K-100K clicks/day | 2-3 instances + Redis |
| 100K-1M clicks/day | 5+ instances + Redis cluster |
| > 1M clicks/day | Contact for enterprise setup |

### Resource Requirements

```yaml
# Per instance (100K clicks/day)
cpu: 2 cores
memory: 4GB
storage: 50GB SSD
network: 100Mbps

# Redis
memory: 2GB per 1M cached items
```

---

## Backup & Recovery

### Backup Schedule

| Data Type | Frequency | Retention |
|-----------|-----------|-----------|
| Database | Daily | 30 days |
| Redis | Hourly | 7 days |
| Logs | Daily | 90 days |
| Config | On change | Unlimited |

### Recovery Procedures

#### Database Recovery

```bash
# Point-in-time recovery
pg_restore -d afftok_db backup_2024_12_01.dump

# Verify data integrity
SELECT COUNT(*) FROM clicks WHERE created_at > '2024-12-01';
```

#### Redis Recovery

```bash
# Restore from RDB
redis-cli CONFIG SET dir /backup
redis-cli CONFIG SET dbfilename dump.rdb
redis-cli DEBUG RELOAD
```

### Disaster Recovery Plan

1. **RTO (Recovery Time Objective)**: 4 hours
2. **RPO (Recovery Point Objective)**: 1 hour

| Scenario | Recovery Steps |
|----------|---------------|
| Single instance failure | Auto-failover to healthy instance |
| Database failure | Promote replica, restore from backup |
| Redis failure | Failover to replica, rebuild cache |
| Complete region failure | Failover to DR region |

---

## Maintenance Windows

### Scheduled Maintenance

| Type | Frequency | Duration | Impact |
|------|-----------|----------|--------|
| Security patches | Weekly | 15 min | None (rolling) |
| Minor updates | Monthly | 30 min | Minimal |
| Major updates | Quarterly | 2 hours | Planned downtime |
| Database maintenance | Weekly | 30 min | None (replica) |

### Maintenance Notification

Notify users at least:
- 24 hours for minor maintenance
- 7 days for major maintenance
- 30 days for breaking changes

### Zero-Downtime Deployment

```bash
# Rolling deployment
kubectl rollout restart deployment/afftok-api

# Canary deployment
kubectl set image deployment/afftok-api \
  api=afftok/api:v2.0.0 \
  --record

# Monitor
kubectl rollout status deployment/afftok-api
```

---

## Compliance

### Data Handling

| Requirement | Implementation |
|-------------|---------------|
| GDPR | Data export, deletion APIs |
| CCPA | Opt-out mechanisms |
| PCI DSS | No card data stored |
| SOC 2 | Audit logging, access controls |

### Data Retention

| Data Type | Default Retention | Configurable |
|-----------|-------------------|--------------|
| Clicks | 90 days | Yes |
| Conversions | 2 years | Yes |
| Logs | 30 days | Yes |
| Audit trails | 1 year | No |

### Data Export

```bash
# Request data export
curl -X POST https://api.afftok.com/api/admin/data-export \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{"user_id": "user_123", "format": "json"}'
```

### Data Deletion

```bash
# Request data deletion
curl -X POST https://api.afftok.com/api/admin/data-delete \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{"user_id": "user_123", "confirm": true}'
```

---

## Support Escalation

### Support Tiers

| Tier | Response Time | Scope |
|------|---------------|-------|
| L1 | 4 hours | Basic issues, documentation |
| L2 | 2 hours | Technical issues, integration |
| L3 | 1 hour | Critical issues, escalations |
| Engineering | 30 min | P1 incidents only |

### Contact Information

| Channel | Use Case |
|---------|----------|
| support@afftok.com | General support |
| security@afftok.com | Security issues |
| urgent@afftok.com | P1 incidents |
| Slack | Enterprise customers |

### Escalation Path

```
L1 Support → L2 Support → L3 Support → Engineering → On-Call
```

### Information to Provide

When contacting support:

1. **Account ID** - Your tenant/advertiser ID
2. **Correlation ID** - From error responses
3. **Timestamp** - When issue occurred
4. **Steps to reproduce** - Detailed steps
5. **Expected vs actual** - What should happen
6. **Logs/screenshots** - Supporting evidence

---

## Documentation

### Keep Updated

- Integration documentation
- API key inventory
- Webhook endpoints
- Geo rule configurations
- Contact information

### Change Log

Document all changes:

```markdown
## 2024-12-01
- Rotated production API key
- Updated webhook endpoint to v2
- Added geo rule for EU countries

## 2024-11-15
- Initial integration completed
- Configured monitoring alerts
```

---

## Summary

### Daily Operations

- [ ] Monitor error rates
- [ ] Check queue depths
- [ ] Review alerts
- [ ] Verify webhook delivery

### Weekly Operations

- [ ] Review performance metrics
- [ ] Check security logs
- [ ] Verify backups
- [ ] Update documentation

### Monthly Operations

- [ ] Rotate API keys (if due)
- [ ] Review access permissions
- [ ] Capacity planning review
- [ ] Compliance audit

### Quarterly Operations

- [ ] Major version updates
- [ ] Disaster recovery test
- [ ] Security assessment
- [ ] Performance optimization

