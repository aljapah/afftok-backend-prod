/**
 * Edge Metrics - Performance and operational metrics
 */

import type { Env } from './index';

export interface EdgeMetricsData {
  // Request metrics
  total_requests: number;
  clicks_processed: number;
  clicks_blocked: number;
  
  // Bot metrics
  bots_blocked: number;
  dc_traffic_blocked: number;
  ua_anomalies: number;
  ip_anomalies: number;
  
  // Geo metrics
  geo_blocks: number;
  
  // Device metrics
  device_blocks: number;
  
  // Link validation
  signature_failures: number;
  expired_links: number;
  replay_attempts: number;
  legacy_links: number;
  
  // Routing metrics
  router_decisions: number;
  ab_tests: number;
  rotations: number;
  caps_exceeded: number;
  
  // Performance
  avg_latency_ms: number;
  p99_latency_ms: number;
  
  // Errors
  errors: number;
  
  // Queue metrics
  queue_size: number;
  queue_flushes: number;
  
  // Failover
  failover_redirects: number;
  backend_errors: number;
  
  // Per region
  by_region: Record<string, number>;
  
  // Timestamps
  period_start: string;
  period_end: string;
}

export class EdgeMetrics {
  private env: Env;
  private localCounters: Map<string, number> = new Map();
  private latencies: number[] = [];

  constructor(env: Env) {
    this.env = env;
  }

  /**
   * Increment a counter
   */
  async incrementCounter(name: string, value: number = 1): Promise<void> {
    // Local increment
    const current = this.localCounters.get(name) || 0;
    this.localCounters.set(name, current + value);

    // Persist to KV periodically
    try {
      const key = `metric:${name}:${this.getCurrentPeriod()}`;
      const cached = await this.env.ROUTING_CACHE.get(key);
      const newValue = (cached ? parseInt(cached, 10) : 0) + value;
      await this.env.ROUTING_CACHE.put(key, String(newValue), {
        expirationTtl: 3600 // 1 hour
      });
    } catch {
      // Ignore errors
    }
  }

  /**
   * Record latency
   */
  recordLatency(latencyMs: number): void {
    this.latencies.push(latencyMs);
    
    // Keep only last 1000 samples
    if (this.latencies.length > 1000) {
      this.latencies.shift();
    }
  }

  /**
   * Get average latency
   */
  getAverageLatency(): number {
    if (this.latencies.length === 0) return 0;
    const sum = this.latencies.reduce((a, b) => a + b, 0);
    return Math.round(sum / this.latencies.length);
  }

  /**
   * Get P99 latency
   */
  getP99Latency(): number {
    if (this.latencies.length === 0) return 0;
    const sorted = [...this.latencies].sort((a, b) => a - b);
    const index = Math.floor(sorted.length * 0.99);
    return sorted[index] || sorted[sorted.length - 1];
  }

  /**
   * Track click event
   */
  async trackClick(
    edgeLocation: string,
    latencyMs: number,
    blocked: boolean,
    blockReason?: string
  ): Promise<void> {
    this.recordLatency(latencyMs);

    await this.incrementCounter('clicks_processed');
    await this.incrementCounter(`region:${edgeLocation}`);

    if (blocked) {
      await this.incrementCounter('clicks_blocked');
      if (blockReason) {
        await this.incrementCounter(`block:${blockReason}`);
      }
    }
  }

  /**
   * Track bot detection
   */
  async trackBotDetection(
    category: 'bot' | 'dc' | 'ua_anomaly' | 'ip_anomaly'
  ): Promise<void> {
    switch (category) {
      case 'bot':
        await this.incrementCounter('bots_blocked');
        break;
      case 'dc':
        await this.incrementCounter('dc_traffic_blocked');
        break;
      case 'ua_anomaly':
        await this.incrementCounter('ua_anomalies');
        break;
      case 'ip_anomaly':
        await this.incrementCounter('ip_anomalies');
        break;
    }
  }

  /**
   * Track link validation
   */
  async trackLinkValidation(
    result: 'valid' | 'invalid_signature' | 'expired' | 'replay' | 'legacy'
  ): Promise<void> {
    switch (result) {
      case 'invalid_signature':
        await this.incrementCounter('signature_failures');
        break;
      case 'expired':
        await this.incrementCounter('expired_links');
        break;
      case 'replay':
        await this.incrementCounter('replay_attempts');
        break;
      case 'legacy':
        await this.incrementCounter('legacy_links');
        break;
    }
  }

  /**
   * Track routing decision
   */
  async trackRouting(
    decision: 'ab_test' | 'rotation' | 'cap_exceeded' | 'rule_applied'
  ): Promise<void> {
    await this.incrementCounter('router_decisions');
    
    switch (decision) {
      case 'ab_test':
        await this.incrementCounter('ab_tests');
        break;
      case 'rotation':
        await this.incrementCounter('rotations');
        break;
      case 'cap_exceeded':
        await this.incrementCounter('caps_exceeded');
        break;
    }
  }

  /**
   * Track geo block
   */
  async trackGeoBlock(country: string): Promise<void> {
    await this.incrementCounter('geo_blocks');
    await this.incrementCounter(`geo_block:${country}`);
  }

  /**
   * Track failover
   */
  async trackFailover(reason: string): Promise<void> {
    await this.incrementCounter('failover_redirects');
    await this.incrementCounter(`failover:${reason}`);
  }

  /**
   * Track error
   */
  async trackError(errorType: string): Promise<void> {
    await this.incrementCounter('errors');
    await this.incrementCounter(`error:${errorType}`);
  }

  /**
   * Get current metrics
   */
  async getMetrics(): Promise<EdgeMetricsData> {
    const period = this.getCurrentPeriod();
    const metrics: EdgeMetricsData = {
      total_requests: 0,
      clicks_processed: 0,
      clicks_blocked: 0,
      bots_blocked: 0,
      dc_traffic_blocked: 0,
      ua_anomalies: 0,
      ip_anomalies: 0,
      geo_blocks: 0,
      device_blocks: 0,
      signature_failures: 0,
      expired_links: 0,
      replay_attempts: 0,
      legacy_links: 0,
      router_decisions: 0,
      ab_tests: 0,
      rotations: 0,
      caps_exceeded: 0,
      avg_latency_ms: this.getAverageLatency(),
      p99_latency_ms: this.getP99Latency(),
      errors: 0,
      queue_size: 0,
      queue_flushes: 0,
      failover_redirects: 0,
      backend_errors: 0,
      by_region: {},
      period_start: period,
      period_end: new Date().toISOString()
    };

    try {
      // Fetch all metrics for current period
      const list = await this.env.ROUTING_CACHE.list({ prefix: `metric:`, limit: 100 });

      for (const key of list.keys) {
        const value = await this.env.ROUTING_CACHE.get(key.name);
        if (!value) continue;

        const count = parseInt(value, 10);
        const metricName = key.name.replace(`metric:`, '').replace(`:${period}`, '');

        // Map to metrics object
        if (metricName === 'clicks_processed') metrics.clicks_processed = count;
        else if (metricName === 'clicks_blocked') metrics.clicks_blocked = count;
        else if (metricName === 'bots_blocked') metrics.bots_blocked = count;
        else if (metricName === 'dc_traffic_blocked') metrics.dc_traffic_blocked = count;
        else if (metricName === 'ua_anomalies') metrics.ua_anomalies = count;
        else if (metricName === 'ip_anomalies') metrics.ip_anomalies = count;
        else if (metricName === 'geo_blocks') metrics.geo_blocks = count;
        else if (metricName === 'device_blocks') metrics.device_blocks = count;
        else if (metricName === 'signature_failures') metrics.signature_failures = count;
        else if (metricName === 'expired_links') metrics.expired_links = count;
        else if (metricName === 'replay_attempts') metrics.replay_attempts = count;
        else if (metricName === 'legacy_links') metrics.legacy_links = count;
        else if (metricName === 'router_decisions') metrics.router_decisions = count;
        else if (metricName === 'ab_tests') metrics.ab_tests = count;
        else if (metricName === 'rotations') metrics.rotations = count;
        else if (metricName === 'caps_exceeded') metrics.caps_exceeded = count;
        else if (metricName === 'errors') metrics.errors = count;
        else if (metricName === 'failover_redirects') metrics.failover_redirects = count;
        else if (metricName === 'backend_errors') metrics.backend_errors = count;
        else if (metricName.startsWith('region:')) {
          const region = metricName.replace('region:', '');
          metrics.by_region[region] = count;
        }
      }

      metrics.total_requests = metrics.clicks_processed + metrics.clicks_blocked;

    } catch {
      // Ignore errors
    }

    return metrics;
  }

  /**
   * Get current period key (hourly)
   */
  private getCurrentPeriod(): string {
    const now = new Date();
    return `${now.getUTCFullYear()}-${String(now.getUTCMonth() + 1).padStart(2, '0')}-${String(now.getUTCDate()).padStart(2, '0')}T${String(now.getUTCHours()).padStart(2, '0')}`;
  }

  /**
   * Reset metrics
   */
  async resetMetrics(): Promise<void> {
    this.localCounters.clear();
    this.latencies = [];
  }
}

