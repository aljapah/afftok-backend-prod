/**
 * Edge Router - Main routing logic
 */

import type { Env, EdgeClickEvent } from './index';
import type { LinkValidator, LinkValidationResult } from './link-validator';
import type { BotDetector, BotDetectionResult } from './bot-detector';
import type { GeoValidator, GeoInfo, GeoValidationResult } from './geo-validator';
import type { DeviceDetector, DeviceInfo, DeviceValidationResult } from './device-detector';
import type { ClickQueue } from './click-queue';
import type { SmartRouter, OfferConfig, RoutingDecision } from './smart-router';
import type { FailoverManager } from './failover';
import type { EdgeMetrics } from './metrics';
import type { EdgeConfig } from './config';

export interface RouterDependencies {
  linkValidator: LinkValidator;
  botDetector: BotDetector;
  geoValidator: GeoValidator;
  deviceDetector: DeviceDetector;
  clickQueue: ClickQueue;
  smartRouter: SmartRouter;
  failover: FailoverManager;
  metrics: EdgeMetrics;
  config: EdgeConfig;
}

export class EdgeRouter {
  private env: Env;
  private deps: RouterDependencies;

  constructor(env: Env, deps: RouterDependencies) {
    this.env = env;
    this.deps = deps;
  }

  /**
   * Handle click tracking request
   */
  async handleClick(request: Request, ctx: ExecutionContext, startTime: number): Promise<Response> {
    const url = new URL(request.url);
    const signedCode = url.pathname.replace('/c/', '');
    const edgeLocation = (request.cf as any)?.colo || 'unknown';

    // Extract request info
    const ip = request.headers.get('cf-connecting-ip') || 
               request.headers.get('x-forwarded-for')?.split(',')[0] || 
               'unknown';
    const userAgent = request.headers.get('user-agent') || '';
    const referer = request.headers.get('referer') || '';
    const acceptLanguage = request.headers.get('accept-language') || '';

    // 1. Bot Detection
    if (this.deps.config.botDetectionEnabled) {
      const botResult = this.deps.botDetector.detect(request);
      if (botResult.isBot && botResult.riskScore >= 70) {
        ctx.waitUntil(this.deps.metrics.trackBotDetection('bot'));
        ctx.waitUntil(this.deps.metrics.trackClick(edgeLocation, Date.now() - startTime, true, 'bot'));
        
        // Still redirect but don't track
        return this.redirectWithoutTracking(signedCode, 'bot_blocked');
      }
    }

    // 2. Link Validation
    const linkResult = await this.deps.linkValidator.validateSignedLink(signedCode);
    if (!linkResult.valid) {
      ctx.waitUntil(this.deps.metrics.trackLinkValidation(
        linkResult.reason as any || 'invalid_signature'
      ));
      ctx.waitUntil(this.deps.metrics.trackClick(edgeLocation, Date.now() - startTime, true, linkResult.reason));
      
      return this.redirectWithoutTracking(signedCode, linkResult.reason || 'invalid_link');
    }

    const trackingCode = linkResult.trackingCode || linkResult.components?.trackingCode || signedCode;

    // 3. Get offer config (from cache or backend)
    let offerConfig = await this.deps.failover.getCachedOfferConfig(trackingCode);
    
    if (!offerConfig) {
      // Check if backend is healthy
      const backendHealthy = await this.deps.failover.isBackendHealthy();
      
      if (!backendHealthy) {
        ctx.waitUntil(this.deps.metrics.trackFailover('backend_down'));
        return this.deps.failover.handleFailover(trackingCode, 'backend_down');
      }

      // Fetch from backend
      offerConfig = await this.fetchOfferConfig(trackingCode);
      
      if (offerConfig) {
        ctx.waitUntil(this.deps.failover.cacheOfferConfig(trackingCode, offerConfig));
      }
    }

    if (!offerConfig) {
      ctx.waitUntil(this.deps.metrics.trackError('offer_not_found'));
      return this.redirectWithoutTracking(signedCode, 'offer_not_found');
    }

    // 4. Geo Validation
    const geoInfo = this.deps.geoValidator.extractGeoInfo(request);
    
    if (this.deps.config.geoRulesEnabled) {
      const geoResult = await this.deps.geoValidator.validate(
        geoInfo,
        offerConfig.id,
        offerConfig.advertiser_id,
        offerConfig.tenant_id
      );

      if (!geoResult.allowed) {
        ctx.waitUntil(this.deps.metrics.trackGeoBlock(geoInfo.country));
        ctx.waitUntil(this.deps.metrics.trackClick(edgeLocation, Date.now() - startTime, true, 'geo_blocked'));
        
        // Redirect to fallback
        const fallbackUrl = offerConfig.fallback_url || this.deps.config.defaultFallbackUrl;
        return Response.redirect(fallbackUrl, 302);
      }
    }

    // 5. Device Detection & Validation
    const deviceInfo = this.deps.deviceDetector.detect(userAgent);
    
    if (this.deps.config.deviceRulesEnabled && deviceInfo.isBot) {
      ctx.waitUntil(this.deps.metrics.trackClick(edgeLocation, Date.now() - startTime, true, 'device_bot'));
      return this.redirectWithoutTracking(signedCode, 'device_bot');
    }

    // 6. Smart Routing
    const routingDecision = await this.deps.smartRouter.route(
      offerConfig,
      geoInfo,
      deviceInfo,
      request,
      startTime
    );

    if (routingDecision.blocked) {
      ctx.waitUntil(this.deps.metrics.trackClick(edgeLocation, Date.now() - startTime, true, routingDecision.block_reason));
      const fallbackUrl = offerConfig.fallback_url || this.deps.config.defaultFallbackUrl;
      return Response.redirect(fallbackUrl, 302);
    }

    // 7. Create click event
    const clickEvent: EdgeClickEvent = {
      tracking_code: trackingCode,
      tenant_id: offerConfig.tenant_id,
      offer_id: offerConfig.id,
      user_offer_id: trackingCode, // Will be resolved by backend
      country: geoInfo.country,
      region: geoInfo.region,
      city: geoInfo.city,
      device: deviceInfo.type,
      browser: deviceInfo.browser,
      os: deviceInfo.os,
      ip: ip,
      timestamp: new Date().toISOString(),
      user_agent: userAgent,
      referer: referer,
      accept_language: acceptLanguage,
      edge_location: edgeLocation,
      latency_ms: Date.now() - startTime,
      router_decision: routingDecision.rule_applied,
      final_destination: routingDecision.final_destination,
      meta: {
        variant_id: routingDecision.variant_id,
        rotation_index: routingDecision.rotation_index,
        connection_type: (request.cf as any)?.httpProtocol,
        asn: geoInfo.asn,
        is_legacy: linkResult.isLegacy
      }
    };

    // 8. Queue click for async processing (non-blocking)
    ctx.waitUntil(this.deps.clickQueue.queueClick(clickEvent));

    // 9. Update cap counter
    ctx.waitUntil(this.deps.smartRouter.incrementCap(offerConfig.id));

    // 10. Track metrics
    ctx.waitUntil(this.deps.metrics.trackClick(edgeLocation, Date.now() - startTime, false));
    if (routingDecision.variant_id) {
      ctx.waitUntil(this.deps.metrics.trackRouting('ab_test'));
    }
    if (routingDecision.rotation_index !== undefined) {
      ctx.waitUntil(this.deps.metrics.trackRouting('rotation'));
    }

    // 11. Redirect immediately (< 20ms target)
    return Response.redirect(routingDecision.final_destination, 302);
  }

  /**
   * Redirect without tracking (for blocked/invalid clicks)
   */
  private async redirectWithoutTracking(signedCode: string, reason: string): Promise<Response> {
    // Try to get fallback URL from cache
    const offerConfig = await this.deps.failover.getCachedOfferConfig(signedCode);
    const fallbackUrl = offerConfig?.fallback_url || this.deps.config.defaultFallbackUrl;
    
    return Response.redirect(fallbackUrl, 302);
  }

  /**
   * Fetch offer config from backend
   */
  private async fetchOfferConfig(trackingCode: string): Promise<OfferConfig | null> {
    try {
      const backendUrl = this.deps.config.backendUrl;
      const response = await fetch(`${backendUrl}/api/internal/edge/offer/${trackingCode}`, {
        method: 'GET',
        headers: {
          'X-Edge-Request': 'true',
          'X-Edge-Location': 'cloudflare'
        }
      });

      if (response.ok) {
        const data = await response.json() as { data: OfferConfig };
        return data.data;
      }
    } catch {
      // Ignore errors
    }
    return null;
  }

  /**
   * Handle /edge/verify endpoint
   */
  async handleVerify(request: Request): Promise<Response> {
    const url = new URL(request.url);
    const link = url.searchParams.get('link') || '';

    if (!link) {
      return new Response(JSON.stringify({
        error: 'Missing link parameter'
      }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' }
      });
    }

    const result = await this.deps.linkValidator.validateSignedLink(link);

    return new Response(JSON.stringify({
      valid: result.valid,
      reason: result.reason,
      is_legacy: result.isLegacy,
      tracking_code: result.trackingCode,
      components: result.components ? {
        tracking_code: result.components.trackingCode,
        timestamp: result.components.timestamp,
        nonce: result.components.nonce
      } : undefined,
      timestamp: new Date().toISOString()
    }), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  /**
   * Handle /edge/diagnostics endpoint
   */
  async handleDiagnostics(request: Request): Promise<Response> {
    const cf = request.cf as any;
    const metrics = await this.deps.metrics.getMetrics();
    const queueStatus = await this.deps.clickQueue.getQueueStatus();
    const failoverStatus = await this.deps.failover.getFailoverStatus();

    return new Response(JSON.stringify({
      edge_location: cf?.colo || 'unknown',
      environment: this.deps.config.environment,
      config: {
        bot_detection: this.deps.config.botDetectionEnabled,
        geo_rules: this.deps.config.geoRulesEnabled,
        device_rules: this.deps.config.deviceRulesEnabled,
        failover: this.deps.config.failoverEnabled,
        link_ttl_seconds: this.deps.config.linkTTLSeconds,
        allow_legacy: this.deps.config.allowLegacyCodes,
        batch_size: this.deps.config.batchSize
      },
      metrics: {
        clicks_processed: metrics.clicks_processed,
        clicks_blocked: metrics.clicks_blocked,
        bots_blocked: metrics.bots_blocked,
        geo_blocks: metrics.geo_blocks,
        avg_latency_ms: metrics.avg_latency_ms,
        p99_latency_ms: metrics.p99_latency_ms,
        errors: metrics.errors
      },
      queue: queueStatus,
      failover: failoverStatus,
      request_info: {
        country: cf?.country,
        region: cf?.region,
        city: cf?.city,
        asn: cf?.asn,
        colo: cf?.colo
      },
      timestamp: new Date().toISOString()
    }), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  /**
   * Handle /edge/emit endpoint (log batching)
   */
  async handleEmit(request: Request, ctx: ExecutionContext): Promise<Response> {
    try {
      const body = await request.json() as { events: EdgeClickEvent[] };
      
      if (!body.events || !Array.isArray(body.events)) {
        return new Response(JSON.stringify({
          error: 'Invalid request body'
        }), {
          status: 400,
          headers: { 'Content-Type': 'application/json' }
        });
      }

      // Queue all events
      for (const event of body.events) {
        ctx.waitUntil(this.deps.clickQueue.queueClick(event));
      }

      return new Response(JSON.stringify({
        success: true,
        queued: body.events.length,
        timestamp: new Date().toISOString()
      }), {
        headers: { 'Content-Type': 'application/json' }
      });

    } catch (error) {
      return new Response(JSON.stringify({
        error: 'Failed to process events'
      }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  /**
   * Handle /edge/status endpoint
   */
  async handleStatus(request: Request): Promise<Response> {
    const metrics = await this.deps.metrics.getMetrics();
    const failoverStatus = await this.deps.failover.getFailoverStatus();

    return new Response(JSON.stringify({
      status: failoverStatus.backend_healthy ? 'healthy' : 'degraded',
      edge_location: (request.cf as any)?.colo || 'unknown',
      backend_healthy: failoverStatus.backend_healthy,
      metrics: {
        rps: Math.round(metrics.clicks_processed / 3600), // Approx RPS
        blocked_rate: metrics.clicks_processed > 0 
          ? (metrics.clicks_blocked / metrics.clicks_processed * 100).toFixed(2) + '%'
          : '0%',
        avg_latency_ms: metrics.avg_latency_ms,
        error_rate: metrics.total_requests > 0
          ? (metrics.errors / metrics.total_requests * 100).toFixed(2) + '%'
          : '0%'
      },
      timestamp: new Date().toISOString()
    }), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  /**
   * Handle /edge/cache/refresh endpoint
   */
  async handleCacheRefresh(request: Request, ctx: ExecutionContext): Promise<Response> {
    try {
      const body = await request.json() as { 
        type: 'offer' | 'geo_rule' | 'tenant' | 'all';
        id?: string;
      };

      let refreshed = 0;

      switch (body.type) {
        case 'offer':
          if (body.id) {
            await this.env.OFFER_CACHE.delete(`offer_cache:${body.id}`);
            refreshed = 1;
          }
          break;

        case 'geo_rule':
          if (body.id) {
            await this.env.ROUTING_CACHE.delete(`geo_rule:offer:${body.id}`);
            await this.env.ROUTING_CACHE.delete(`geo_rule:advertiser:${body.id}`);
            await this.env.ROUTING_CACHE.delete(`geo_rule:tenant:${body.id}`);
            refreshed = 3;
          }
          break;

        case 'tenant':
          if (body.id) {
            await this.env.ROUTING_CACHE.delete(`tenant_status:${body.id}`);
            refreshed = 1;
          }
          break;

        case 'all':
          // Clear all caches (expensive operation)
          const lists = await Promise.all([
            this.env.OFFER_CACHE.list({ limit: 1000 }),
            this.env.ROUTING_CACHE.list({ limit: 1000 })
          ]);

          for (const list of lists) {
            for (const key of list.keys) {
              ctx.waitUntil(this.env.OFFER_CACHE.delete(key.name));
              ctx.waitUntil(this.env.ROUTING_CACHE.delete(key.name));
              refreshed++;
            }
          }
          break;
      }

      return new Response(JSON.stringify({
        success: true,
        refreshed,
        timestamp: new Date().toISOString()
      }), {
        headers: { 'Content-Type': 'application/json' }
      });

    } catch (error) {
      return new Response(JSON.stringify({
        error: 'Failed to refresh cache'
      }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }
}

