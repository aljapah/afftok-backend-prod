/**
 * Smart Click Router - Advanced routing logic at edge
 * GEO, Device, ISP, Offer Caps, Rotation, A/B Testing
 */

import type { Env } from './index';
import type { GeoInfo } from './geo-validator';
import type { DeviceInfo } from './device-detector';

export interface RoutingRule {
  id: string;
  name: string;
  priority: number;
  conditions: RoutingCondition[];
  action: RoutingAction;
  status: 'active' | 'inactive';
}

export interface RoutingCondition {
  type: 'geo' | 'device' | 'isp' | 'connection' | 'time' | 'cap' | 'custom';
  operator: 'eq' | 'neq' | 'in' | 'not_in' | 'gt' | 'lt' | 'regex';
  field: string;
  value: any;
}

export interface RoutingAction {
  type: 'redirect' | 'rotate' | 'ab_test' | 'fallback' | 'block';
  destination?: string;
  destinations?: WeightedDestination[];
  ab_variants?: ABVariant[];
  fallback_chain?: string[];
}

export interface WeightedDestination {
  url: string;
  weight: number;
  offer_id?: string;
}

export interface ABVariant {
  id: string;
  url: string;
  percentage: number;
}

export interface OfferConfig {
  id: string;
  tenant_id: string;
  advertiser_id: string;
  landing_url: string;
  fallback_url?: string;
  daily_cap?: number;
  total_cap?: number;
  current_clicks?: number;
  current_daily_clicks?: number;
  status: 'active' | 'paused' | 'capped';
  routing_rules?: RoutingRule[];
  ab_test?: ABTestConfig;
  rotation?: RotationConfig;
}

export interface ABTestConfig {
  enabled: boolean;
  variants: ABVariant[];
}

export interface RotationConfig {
  mode: 'round_robin' | 'weighted' | 'smart_ctr';
  destinations: WeightedDestination[];
}

export interface RoutingDecision {
  final_destination: string;
  rule_applied: string;
  router_source: 'edge';
  latency_ms: number;
  offer_id?: string;
  variant_id?: string;
  rotation_index?: number;
  capped?: boolean;
  blocked?: boolean;
  block_reason?: string;
}

export class SmartRouter {
  private env: Env;

  constructor(env: Env) {
    this.env = env;
  }

  /**
   * Route click based on all rules
   */
  async route(
    offerConfig: OfferConfig,
    geoInfo: GeoInfo,
    deviceInfo: DeviceInfo,
    request: Request,
    startTime: number
  ): Promise<RoutingDecision> {
    const decision: RoutingDecision = {
      final_destination: offerConfig.landing_url,
      rule_applied: 'default',
      router_source: 'edge',
      latency_ms: 0,
      offer_id: offerConfig.id
    };

    try {
      // 1. Check offer caps
      const capResult = await this.checkCaps(offerConfig);
      if (capResult.capped) {
        decision.capped = true;
        decision.rule_applied = 'cap_exceeded';
        decision.final_destination = offerConfig.fallback_url || this.getDefaultFallback();
        decision.latency_ms = Date.now() - startTime;
        return decision;
      }

      // 2. Apply routing rules
      if (offerConfig.routing_rules && offerConfig.routing_rules.length > 0) {
        const ruleResult = await this.applyRules(
          offerConfig.routing_rules,
          geoInfo,
          deviceInfo,
          request
        );

        if (ruleResult) {
          if (ruleResult.blocked) {
            decision.blocked = true;
            decision.block_reason = ruleResult.block_reason;
            decision.final_destination = offerConfig.fallback_url || this.getDefaultFallback();
          } else if (ruleResult.destination) {
            decision.final_destination = ruleResult.destination;
            decision.rule_applied = ruleResult.rule_name;
          }
        }
      }

      // 3. A/B Testing
      if (!decision.blocked && offerConfig.ab_test?.enabled) {
        const variant = this.selectABVariant(offerConfig.ab_test.variants);
        if (variant) {
          decision.final_destination = variant.url;
          decision.variant_id = variant.id;
          decision.rule_applied = `ab_test:${variant.id}`;
        }
      }

      // 4. Rotation
      if (!decision.blocked && !decision.variant_id && offerConfig.rotation) {
        const rotated = await this.applyRotation(offerConfig);
        if (rotated) {
          decision.final_destination = rotated.url;
          decision.rotation_index = rotated.index;
          decision.rule_applied = `rotation:${offerConfig.rotation.mode}`;
        }
      }

    } catch (error) {
      console.error('Routing error:', error);
      // Fall back to default
    }

    decision.latency_ms = Date.now() - startTime;
    return decision;
  }

  /**
   * Check offer caps
   */
  private async checkCaps(offerConfig: OfferConfig): Promise<{ capped: boolean; reason?: string }> {
    // Check total cap
    if (offerConfig.total_cap && offerConfig.current_clicks) {
      if (offerConfig.current_clicks >= offerConfig.total_cap) {
        return { capped: true, reason: 'total_cap_exceeded' };
      }
    }

    // Check daily cap
    if (offerConfig.daily_cap && offerConfig.current_daily_clicks) {
      if (offerConfig.current_daily_clicks >= offerConfig.daily_cap) {
        return { capped: true, reason: 'daily_cap_exceeded' };
      }
    }

    // Check from KV cache
    try {
      const capKey = `cap:${offerConfig.id}:${this.getTodayKey()}`;
      const cached = await this.env.ROUTING_CACHE.get(capKey);
      if (cached) {
        const dailyClicks = parseInt(cached, 10);
        if (offerConfig.daily_cap && dailyClicks >= offerConfig.daily_cap) {
          return { capped: true, reason: 'daily_cap_exceeded' };
        }
      }
    } catch {
      // Ignore cache errors
    }

    return { capped: false };
  }

  /**
   * Apply routing rules
   */
  private async applyRules(
    rules: RoutingRule[],
    geoInfo: GeoInfo,
    deviceInfo: DeviceInfo,
    request: Request
  ): Promise<{ destination?: string; rule_name: string; blocked?: boolean; block_reason?: string } | null> {
    // Sort by priority
    const sortedRules = [...rules].sort((a, b) => b.priority - a.priority);

    for (const rule of sortedRules) {
      if (rule.status !== 'active') continue;

      const matches = this.evaluateConditions(rule.conditions, geoInfo, deviceInfo, request);
      
      if (matches) {
        switch (rule.action.type) {
          case 'redirect':
            return { destination: rule.action.destination, rule_name: rule.name };

          case 'rotate':
            if (rule.action.destinations) {
              const selected = this.selectWeighted(rule.action.destinations);
              return { destination: selected.url, rule_name: rule.name };
            }
            break;

          case 'ab_test':
            if (rule.action.ab_variants) {
              const variant = this.selectABVariant(rule.action.ab_variants);
              if (variant) {
                return { destination: variant.url, rule_name: `${rule.name}:${variant.id}` };
              }
            }
            break;

          case 'fallback':
            if (rule.action.fallback_chain && rule.action.fallback_chain.length > 0) {
              return { destination: rule.action.fallback_chain[0], rule_name: rule.name };
            }
            break;

          case 'block':
            return { blocked: true, block_reason: rule.name, rule_name: rule.name };
        }
      }
    }

    return null;
  }

  /**
   * Evaluate all conditions
   */
  private evaluateConditions(
    conditions: RoutingCondition[],
    geoInfo: GeoInfo,
    deviceInfo: DeviceInfo,
    request: Request
  ): boolean {
    for (const condition of conditions) {
      if (!this.evaluateCondition(condition, geoInfo, deviceInfo, request)) {
        return false;
      }
    }
    return true;
  }

  /**
   * Evaluate single condition
   */
  private evaluateCondition(
    condition: RoutingCondition,
    geoInfo: GeoInfo,
    deviceInfo: DeviceInfo,
    request: Request
  ): boolean {
    let fieldValue: any;

    switch (condition.type) {
      case 'geo':
        fieldValue = this.getGeoField(condition.field, geoInfo);
        break;
      case 'device':
        fieldValue = this.getDeviceField(condition.field, deviceInfo);
        break;
      case 'isp':
        fieldValue = geoInfo.asOrganization;
        break;
      case 'connection':
        fieldValue = (request.cf as any)?.[condition.field];
        break;
      case 'time':
        fieldValue = this.getTimeField(condition.field);
        break;
      default:
        return true;
    }

    return this.compareValues(fieldValue, condition.operator, condition.value);
  }

  /**
   * Get geo field value
   */
  private getGeoField(field: string, geoInfo: GeoInfo): any {
    switch (field) {
      case 'country': return geoInfo.country;
      case 'region': return geoInfo.region;
      case 'city': return geoInfo.city;
      case 'continent': return geoInfo.continent;
      case 'isEU': return geoInfo.isEU;
      case 'asn': return geoInfo.asn;
      default: return null;
    }
  }

  /**
   * Get device field value
   */
  private getDeviceField(field: string, deviceInfo: DeviceInfo): any {
    switch (field) {
      case 'type': return deviceInfo.type;
      case 'browser': return deviceInfo.browser;
      case 'os': return deviceInfo.os;
      case 'isMobile': return deviceInfo.isMobile;
      case 'isTablet': return deviceInfo.isTablet;
      case 'isDesktop': return deviceInfo.isDesktop;
      default: return null;
    }
  }

  /**
   * Get time field value
   */
  private getTimeField(field: string): any {
    const now = new Date();
    switch (field) {
      case 'hour': return now.getUTCHours();
      case 'day': return now.getUTCDay();
      case 'date': return now.getUTCDate();
      case 'month': return now.getUTCMonth() + 1;
      default: return null;
    }
  }

  /**
   * Compare values with operator
   */
  private compareValues(fieldValue: any, operator: string, conditionValue: any): boolean {
    switch (operator) {
      case 'eq':
        return fieldValue === conditionValue;
      case 'neq':
        return fieldValue !== conditionValue;
      case 'in':
        return Array.isArray(conditionValue) && conditionValue.includes(fieldValue);
      case 'not_in':
        return Array.isArray(conditionValue) && !conditionValue.includes(fieldValue);
      case 'gt':
        return fieldValue > conditionValue;
      case 'lt':
        return fieldValue < conditionValue;
      case 'regex':
        return new RegExp(conditionValue).test(String(fieldValue));
      default:
        return false;
    }
  }

  /**
   * Select A/B variant based on percentage
   */
  private selectABVariant(variants: ABVariant[]): ABVariant | null {
    const random = Math.random() * 100;
    let cumulative = 0;

    for (const variant of variants) {
      cumulative += variant.percentage;
      if (random <= cumulative) {
        return variant;
      }
    }

    return variants[0] || null;
  }

  /**
   * Select weighted destination
   */
  private selectWeighted(destinations: WeightedDestination[]): WeightedDestination {
    const totalWeight = destinations.reduce((sum, d) => sum + d.weight, 0);
    const random = Math.random() * totalWeight;
    let cumulative = 0;

    for (const dest of destinations) {
      cumulative += dest.weight;
      if (random <= cumulative) {
        return dest;
      }
    }

    return destinations[0];
  }

  /**
   * Apply rotation
   */
  private async applyRotation(offerConfig: OfferConfig): Promise<{ url: string; index: number } | null> {
    if (!offerConfig.rotation || offerConfig.rotation.destinations.length === 0) {
      return null;
    }

    const { mode, destinations } = offerConfig.rotation;

    switch (mode) {
      case 'round_robin':
        return this.roundRobinRotation(offerConfig.id, destinations);

      case 'weighted':
        const selected = this.selectWeighted(destinations);
        return { url: selected.url, index: destinations.indexOf(selected) };

      case 'smart_ctr':
        // Would need CTR data from backend
        return this.roundRobinRotation(offerConfig.id, destinations);

      default:
        return null;
    }
  }

  /**
   * Round robin rotation with KV state
   */
  private async roundRobinRotation(
    offerId: string,
    destinations: WeightedDestination[]
  ): Promise<{ url: string; index: number }> {
    try {
      const key = `rotation:${offerId}`;
      const cached = await this.env.ROUTING_CACHE.get(key);
      let index = cached ? parseInt(cached, 10) : 0;

      // Next index
      const nextIndex = (index + 1) % destinations.length;
      await this.env.ROUTING_CACHE.put(key, String(nextIndex), {
        expirationTtl: 86400 // 24 hours
      });

      return { url: destinations[index].url, index };
    } catch {
      return { url: destinations[0].url, index: 0 };
    }
  }

  /**
   * Get today's key for daily caps
   */
  private getTodayKey(): string {
    return new Date().toISOString().split('T')[0];
  }

  /**
   * Get default fallback URL
   */
  private getDefaultFallback(): string {
    return 'https://afftok.com';
  }

  /**
   * Increment cap counter
   */
  async incrementCap(offerId: string): Promise<void> {
    try {
      const capKey = `cap:${offerId}:${this.getTodayKey()}`;
      const current = await this.env.ROUTING_CACHE.get(capKey);
      const newValue = (current ? parseInt(current, 10) : 0) + 1;
      await this.env.ROUTING_CACHE.put(capKey, String(newValue), {
        expirationTtl: 86400 // 24 hours
      });
    } catch {
      // Ignore errors
    }
  }
}

