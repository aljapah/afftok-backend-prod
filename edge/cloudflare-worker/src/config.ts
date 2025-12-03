/**
 * Edge Configuration
 */

import type { Env } from './index';

export class EdgeConfig {
  private env: Env;

  constructor(env: Env) {
    this.env = env;
  }

  get backendUrl(): string {
    return this.env.BACKEND_URL || 'https://api.afftok.com';
  }

  get signingSecret(): string {
    return this.env.SIGNING_SECRET || '';
  }

  get linkTTLSeconds(): number {
    return parseInt(this.env.LINK_TTL_SECONDS || '86400', 10);
  }

  get allowLegacyCodes(): boolean {
    return this.env.ALLOW_LEGACY_CODES === 'true';
  }

  get botDetectionEnabled(): boolean {
    return this.env.BOT_DETECTION_ENABLED !== 'false';
  }

  get geoRulesEnabled(): boolean {
    return this.env.GEO_RULES_ENABLED !== 'false';
  }

  get deviceRulesEnabled(): boolean {
    return this.env.DEVICE_RULES_ENABLED !== 'false';
  }

  get failoverEnabled(): boolean {
    return this.env.FAILOVER_ENABLED !== 'false';
  }

  get batchSize(): number {
    return parseInt(this.env.BATCH_SIZE || '50', 10);
  }

  get batchIntervalMs(): number {
    return parseInt(this.env.BATCH_INTERVAL_MS || '5000', 10);
  }

  get environment(): string {
    return this.env.ENVIRONMENT || 'development';
  }

  get isProduction(): boolean {
    return this.environment === 'production';
  }

  // Default fallback URLs
  get defaultFallbackUrl(): string {
    return 'https://afftok.com';
  }

  get suspendedTenantUrl(): string {
    return 'https://afftok.com/suspended';
  }

  get maintenanceUrl(): string {
    return 'https://afftok.com/maintenance';
  }

  // Timeouts
  get backendTimeoutMs(): number {
    return 5000;
  }

  get cacheRefreshIntervalMs(): number {
    return 30000; // 30 seconds
  }

  // Rate limits
  get maxClicksPerIpPerMinute(): number {
    return 100;
  }

  get maxClicksPerOfferPerMinute(): number {
    return 1000;
  }
}

