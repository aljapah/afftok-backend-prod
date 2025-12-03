/**
 * AffTok Edge Router - Cloudflare Worker
 * 
 * Global CDN Edge Layer for ultra-fast click tracking
 * Target: < 20ms redirect worldwide
 */

import { EdgeRouter } from './router';
import { LinkValidator } from './link-validator';
import { BotDetector } from './bot-detector';
import { GeoValidator } from './geo-validator';
import { DeviceDetector } from './device-detector';
import { ClickQueue } from './click-queue';
import { EdgeMetrics } from './metrics';
import { EdgeConfig } from './config';
import { SmartRouter } from './smart-router';
import { FailoverManager } from './failover';

export interface Env {
  // KV Namespaces
  ROUTING_CACHE: KVNamespace;
  OFFER_CACHE: KVNamespace;
  NONCE_CACHE: KVNamespace;
  CLICK_QUEUE: KVNamespace;
  
  // Secrets
  SIGNING_SECRET: string;
  
  // Configuration
  BACKEND_URL: string;
  LINK_TTL_SECONDS: string;
  ALLOW_LEGACY_CODES: string;
  BOT_DETECTION_ENABLED: string;
  GEO_RULES_ENABLED: string;
  DEVICE_RULES_ENABLED: string;
  FAILOVER_ENABLED: string;
  BATCH_SIZE: string;
  BATCH_INTERVAL_MS: string;
  ENVIRONMENT: string;
}

export interface EdgeClickEvent {
  tracking_code: string;
  tenant_id: string;
  offer_id: string;
  user_offer_id: string;
  country: string;
  region: string;
  city: string;
  device: string;
  browser: string;
  os: string;
  ip: string;
  timestamp: string;
  user_agent: string;
  referer: string;
  accept_language: string;
  edge_location: string;
  latency_ms: number;
  router_decision: string;
  final_destination: string;
  meta: Record<string, any>;
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const startTime = Date.now();
    const url = new URL(request.url);
    const path = url.pathname;

    // Initialize services
    const config = new EdgeConfig(env);
    const metrics = new EdgeMetrics(env);
    const linkValidator = new LinkValidator(env);
    const botDetector = new BotDetector(env);
    const geoValidator = new GeoValidator(env);
    const deviceDetector = new DeviceDetector();
    const clickQueue = new ClickQueue(env);
    const smartRouter = new SmartRouter(env);
    const failover = new FailoverManager(env);
    const router = new EdgeRouter(env, {
      linkValidator,
      botDetector,
      geoValidator,
      deviceDetector,
      clickQueue,
      smartRouter,
      failover,
      metrics,
      config
    });

    try {
      // ============================================
      // CLICK TRACKING: /c/*
      // ============================================
      if (path.startsWith('/c/')) {
        return await router.handleClick(request, ctx, startTime);
      }

      // ============================================
      // EDGE VERIFICATION: /edge/verify
      // ============================================
      if (path === '/edge/verify') {
        return await router.handleVerify(request);
      }

      // ============================================
      // EDGE DIAGNOSTICS: /edge/diagnostics
      // ============================================
      if (path === '/edge/diagnostics') {
        return await router.handleDiagnostics(request);
      }

      // ============================================
      // EDGE EMIT (Log Batching): /edge/emit
      // ============================================
      if (path === '/edge/emit' && request.method === 'POST') {
        return await router.handleEmit(request, ctx);
      }

      // ============================================
      // EDGE STATUS: /edge/status
      // ============================================
      if (path === '/edge/status') {
        return await router.handleStatus(request);
      }

      // ============================================
      // EDGE HEALTH: /edge/health
      // ============================================
      if (path === '/edge/health') {
        return new Response(JSON.stringify({
          status: 'healthy',
          edge_location: request.cf?.colo || 'unknown',
          timestamp: new Date().toISOString(),
          latency_ms: Date.now() - startTime
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      }

      // ============================================
      // EDGE CACHE REFRESH: /edge/cache/refresh
      // ============================================
      if (path === '/edge/cache/refresh' && request.method === 'POST') {
        return await router.handleCacheRefresh(request, ctx);
      }

      // ============================================
      // EDGE QUEUE FLUSH: /edge/queue/flush
      // ============================================
      if (path === '/edge/queue/flush' && request.method === 'POST') {
        return await clickQueue.flushToBackend(ctx);
      }

      // 404 for unknown paths
      return new Response('Not Found', { status: 404 });

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      
      // Log error
      console.error('Edge Router Error:', {
        path,
        error: errorMessage,
        timestamp: new Date().toISOString()
      });

      // Increment error counter
      ctx.waitUntil(metrics.incrementCounter('edge_errors'));

      return new Response(JSON.stringify({
        error: 'Internal Edge Error',
        code: 'EDGE_ERROR',
        timestamp: new Date().toISOString()
      }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  },

  // Scheduled handler for batch processing
  async scheduled(event: ScheduledEvent, env: Env, ctx: ExecutionContext): Promise<void> {
    const clickQueue = new ClickQueue(env);
    ctx.waitUntil(clickQueue.flushToBackend(ctx));
  }
};

