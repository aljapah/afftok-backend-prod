/**
 * Failover Manager - Handle backend unavailability
 */

import type { Env } from './index';
import type { OfferConfig } from './smart-router';

export interface FailoverConfig {
  enabled: boolean;
  cache_ttl_seconds: number;
  fallback_url: string;
  suspended_url: string;
  maintenance_url: string;
}

export interface BackendHealthStatus {
  healthy: boolean;
  last_check: string;
  last_success: string;
  consecutive_failures: number;
  latency_ms?: number;
}

export interface TenantStatus {
  id: string;
  status: 'active' | 'suspended' | 'deleted';
  fallback_url?: string;
}

export class FailoverManager {
  private env: Env;
  private healthCache: BackendHealthStatus | null = null;

  constructor(env: Env) {
    this.env = env;
  }

  /**
   * Check if backend is healthy
   */
  async isBackendHealthy(): Promise<boolean> {
    // Check cached status first
    if (this.healthCache) {
      const lastCheck = new Date(this.healthCache.last_check).getTime();
      if (Date.now() - lastCheck < 10000) { // 10 seconds cache
        return this.healthCache.healthy;
      }
    }

    // Check KV cache
    try {
      const cached = await this.env.ROUTING_CACHE.get('backend_health');
      if (cached) {
        const status = JSON.parse(cached) as BackendHealthStatus;
        const lastCheck = new Date(status.last_check).getTime();
        if (Date.now() - lastCheck < 30000) { // 30 seconds
          this.healthCache = status;
          return status.healthy;
        }
      }
    } catch {
      // Ignore cache errors
    }

    // Perform health check
    return this.performHealthCheck();
  }

  /**
   * Perform backend health check
   */
  private async performHealthCheck(): Promise<boolean> {
    const backendUrl = this.env.BACKEND_URL || 'https://api.afftok.com';
    const healthEndpoint = `${backendUrl}/health`;

    const startTime = Date.now();
    let healthy = false;
    let latency = 0;

    try {
      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), 5000);

      const response = await fetch(healthEndpoint, {
        method: 'GET',
        signal: controller.signal
      });

      clearTimeout(timeout);
      latency = Date.now() - startTime;
      healthy = response.ok;

    } catch {
      healthy = false;
      latency = Date.now() - startTime;
    }

    // Update status
    const previousStatus = this.healthCache;
    const status: BackendHealthStatus = {
      healthy,
      last_check: new Date().toISOString(),
      last_success: healthy ? new Date().toISOString() : (previousStatus?.last_success || ''),
      consecutive_failures: healthy ? 0 : (previousStatus?.consecutive_failures || 0) + 1,
      latency_ms: latency
    };

    this.healthCache = status;

    // Cache in KV
    try {
      await this.env.ROUTING_CACHE.put('backend_health', JSON.stringify(status), {
        expirationTtl: 60
      });
    } catch {
      // Ignore cache errors
    }

    return healthy;
  }

  /**
   * Get cached offer config
   */
  async getCachedOfferConfig(trackingCode: string): Promise<OfferConfig | null> {
    try {
      const key = `offer_cache:${trackingCode}`;
      const cached = await this.env.OFFER_CACHE.get(key);
      if (cached) {
        return JSON.parse(cached);
      }
    } catch {
      // Ignore errors
    }
    return null;
  }

  /**
   * Cache offer config
   */
  async cacheOfferConfig(trackingCode: string, config: OfferConfig): Promise<void> {
    try {
      const key = `offer_cache:${trackingCode}`;
      await this.env.OFFER_CACHE.put(key, JSON.stringify(config), {
        expirationTtl: 1800 // 30 minutes
      });
    } catch {
      // Ignore errors
    }
  }

  /**
   * Get tenant status
   */
  async getTenantStatus(tenantId: string): Promise<TenantStatus | null> {
    try {
      const key = `tenant_status:${tenantId}`;
      const cached = await this.env.ROUTING_CACHE.get(key);
      if (cached) {
        return JSON.parse(cached);
      }
    } catch {
      // Ignore errors
    }
    return null;
  }

  /**
   * Cache tenant status
   */
  async cacheTenantStatus(status: TenantStatus): Promise<void> {
    try {
      const key = `tenant_status:${status.id}`;
      await this.env.ROUTING_CACHE.put(key, JSON.stringify(status), {
        expirationTtl: 300 // 5 minutes
      });
    } catch {
      // Ignore errors
    }
  }

  /**
   * Get failover redirect URL
   */
  getFailoverUrl(reason: 'backend_down' | 'tenant_suspended' | 'maintenance' | 'error'): string {
    switch (reason) {
      case 'backend_down':
        return 'https://afftok.com/offline';
      case 'tenant_suspended':
        return 'https://afftok.com/suspended';
      case 'maintenance':
        return 'https://afftok.com/maintenance';
      case 'error':
      default:
        return 'https://afftok.com';
    }
  }

  /**
   * Handle failover redirect
   */
  async handleFailover(
    trackingCode: string,
    reason: 'backend_down' | 'tenant_suspended' | 'maintenance' | 'error'
  ): Promise<Response> {
    // Try to get cached offer config
    const cachedConfig = await this.getCachedOfferConfig(trackingCode);
    
    let redirectUrl: string;
    
    if (cachedConfig && cachedConfig.fallback_url) {
      redirectUrl = cachedConfig.fallback_url;
    } else if (cachedConfig && cachedConfig.landing_url) {
      redirectUrl = cachedConfig.landing_url;
    } else {
      redirectUrl = this.getFailoverUrl(reason);
    }

    return Response.redirect(redirectUrl, 302);
  }

  /**
   * Store click locally during failover
   */
  async storeLocalClick(event: any): Promise<void> {
    try {
      const key = `local_click:${Date.now()}:${Math.random().toString(36).substring(7)}`;
      await this.env.CLICK_QUEUE.put(key, JSON.stringify(event), {
        expirationTtl: 86400 // 24 hours
      });
    } catch {
      // Ignore errors
    }
  }

  /**
   * Get local clicks count
   */
  async getLocalClicksCount(): Promise<number> {
    try {
      const list = await this.env.CLICK_QUEUE.list({ prefix: 'local_click:', limit: 1000 });
      return list.keys.length;
    } catch {
      return 0;
    }
  }

  /**
   * Flush local clicks when backend recovers
   */
  async flushLocalClicks(): Promise<{ flushed: number; failed: number }> {
    let flushed = 0;
    let failed = 0;

    try {
      const list = await this.env.CLICK_QUEUE.list({ prefix: 'local_click:', limit: 100 });

      for (const key of list.keys) {
        try {
          const data = await this.env.CLICK_QUEUE.get(key.name);
          if (data) {
            // Send to backend
            const response = await fetch(`${this.env.BACKEND_URL}/api/internal/edge-click`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: data
            });

            if (response.ok) {
              await this.env.CLICK_QUEUE.delete(key.name);
              flushed++;
            } else {
              failed++;
            }
          }
        } catch {
          failed++;
        }
      }
    } catch {
      // Ignore errors
    }

    return { flushed, failed };
  }

  /**
   * Get failover status
   */
  async getFailoverStatus(): Promise<{
    backend_healthy: boolean;
    last_health_check: string;
    consecutive_failures: number;
    local_clicks_pending: number;
    cached_offers: number;
  }> {
    const healthy = await this.isBackendHealthy();
    const localClicks = await this.getLocalClicksCount();

    let cachedOffers = 0;
    try {
      const list = await this.env.OFFER_CACHE.list({ prefix: 'offer_cache:', limit: 1000 });
      cachedOffers = list.keys.length;
    } catch {
      // Ignore
    }

    return {
      backend_healthy: healthy,
      last_health_check: this.healthCache?.last_check || '',
      consecutive_failures: this.healthCache?.consecutive_failures || 0,
      local_clicks_pending: localClicks,
      cached_offers: cachedOffers
    };
  }
}

