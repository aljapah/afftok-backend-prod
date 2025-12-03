/**
 * Geo Validator - Country/Region validation at edge
 */

import type { Env } from './index';

export interface GeoInfo {
  country: string;
  region: string;
  city: string;
  continent: string;
  timezone: string;
  isEU: boolean;
  latitude: number;
  longitude: number;
  asn: number;
  asOrganization: string;
}

export interface GeoRule {
  id: string;
  scope_type: 'offer' | 'advertiser' | 'tenant' | 'global';
  scope_id: string;
  mode: 'allow' | 'block';
  countries: string[];
  regions?: string[];
  priority: number;
  status: 'active' | 'inactive';
}

export interface GeoValidationResult {
  allowed: boolean;
  reason?: string;
  rule?: GeoRule;
  geoInfo: GeoInfo;
}

export class GeoValidator {
  private env: Env;

  constructor(env: Env) {
    this.env = env;
  }

  /**
   * Extract geo info from Cloudflare request
   */
  extractGeoInfo(request: Request): GeoInfo {
    const cf = request.cf as any;

    return {
      country: cf?.country || 'XX',
      region: cf?.region || '',
      city: cf?.city || '',
      continent: cf?.continent || '',
      timezone: cf?.timezone || 'UTC',
      isEU: cf?.isEUCountry || false,
      latitude: cf?.latitude || 0,
      longitude: cf?.longitude || 0,
      asn: cf?.asn || 0,
      asOrganization: cf?.asOrganization || ''
    };
  }

  /**
   * Validate geo against rules
   */
  async validate(
    geoInfo: GeoInfo,
    offerId: string,
    advertiserId?: string,
    tenantId?: string
  ): Promise<GeoValidationResult> {
    // Get effective rule
    const rule = await this.getEffectiveRule(offerId, advertiserId, tenantId);

    if (!rule) {
      // No rule = allow all
      return {
        allowed: true,
        geoInfo
      };
    }

    // Check country
    const countryMatch = rule.countries.includes(geoInfo.country) ||
                         rule.countries.includes('*');

    // Check region if specified
    let regionMatch = true;
    if (rule.regions && rule.regions.length > 0) {
      regionMatch = rule.regions.includes(geoInfo.region) ||
                    rule.regions.includes('*');
    }

    const matches = countryMatch && regionMatch;

    // Apply mode
    let allowed: boolean;
    if (rule.mode === 'allow') {
      allowed = matches;
    } else {
      allowed = !matches;
    }

    return {
      allowed,
      reason: allowed ? undefined : `geo_${rule.mode}_${geoInfo.country}`,
      rule,
      geoInfo
    };
  }

  /**
   * Get effective geo rule (offer > advertiser > tenant > global)
   */
  private async getEffectiveRule(
    offerId: string,
    advertiserId?: string,
    tenantId?: string
  ): Promise<GeoRule | null> {
    // Try offer-level rule
    let rule = await this.getRuleFromCache('offer', offerId);
    if (rule) return rule;

    // Try advertiser-level rule
    if (advertiserId) {
      rule = await this.getRuleFromCache('advertiser', advertiserId);
      if (rule) return rule;
    }

    // Try tenant-level rule
    if (tenantId) {
      rule = await this.getRuleFromCache('tenant', tenantId);
      if (rule) return rule;
    }

    // Try global rule
    rule = await this.getRuleFromCache('global', 'global');
    return rule;
  }

  /**
   * Get rule from KV cache
   */
  private async getRuleFromCache(
    scopeType: string,
    scopeId: string
  ): Promise<GeoRule | null> {
    try {
      const key = `geo_rule:${scopeType}:${scopeId}`;
      const cached = await this.env.ROUTING_CACHE.get(key);
      
      if (cached) {
        return JSON.parse(cached);
      }
    } catch {
      // Ignore cache errors
    }
    return null;
  }

  /**
   * Cache geo rule
   */
  async cacheRule(rule: GeoRule): Promise<void> {
    const key = `geo_rule:${rule.scope_type}:${rule.scope_id}`;
    await this.env.ROUTING_CACHE.put(key, JSON.stringify(rule), {
      expirationTtl: 600 // 10 minutes
    });
  }

  /**
   * Invalidate geo rule cache
   */
  async invalidateRule(scopeType: string, scopeId: string): Promise<void> {
    const key = `geo_rule:${scopeType}:${scopeId}`;
    await this.env.ROUTING_CACHE.delete(key);
  }

  /**
   * Check if country is in EU
   */
  isEUCountry(country: string): boolean {
    const euCountries = new Set([
      'AT', 'BE', 'BG', 'HR', 'CY', 'CZ', 'DK', 'EE', 'FI', 'FR',
      'DE', 'GR', 'HU', 'IE', 'IT', 'LV', 'LT', 'LU', 'MT', 'NL',
      'PL', 'PT', 'RO', 'SK', 'SI', 'ES', 'SE'
    ]);
    return euCountries.has(country);
  }

  /**
   * Check if country is in GDPR scope
   */
  isGDPRCountry(country: string): boolean {
    // EU + EEA + UK
    const gdprCountries = new Set([
      'AT', 'BE', 'BG', 'HR', 'CY', 'CZ', 'DK', 'EE', 'FI', 'FR',
      'DE', 'GR', 'HU', 'IE', 'IT', 'LV', 'LT', 'LU', 'MT', 'NL',
      'PL', 'PT', 'RO', 'SK', 'SI', 'ES', 'SE',
      'IS', 'LI', 'NO', // EEA
      'GB' // UK
    ]);
    return gdprCountries.has(country);
  }

  /**
   * Get country name
   */
  getCountryName(code: string): string {
    const countries: Record<string, string> = {
      'US': 'United States',
      'GB': 'United Kingdom',
      'DE': 'Germany',
      'FR': 'France',
      'CA': 'Canada',
      'AU': 'Australia',
      'JP': 'Japan',
      'KW': 'Kuwait',
      'SA': 'Saudi Arabia',
      'AE': 'United Arab Emirates',
      'EG': 'Egypt',
      'JO': 'Jordan',
      'LB': 'Lebanon',
      'QA': 'Qatar',
      'BH': 'Bahrain',
      'OM': 'Oman',
      // Add more as needed
    };
    return countries[code] || code;
  }
}

