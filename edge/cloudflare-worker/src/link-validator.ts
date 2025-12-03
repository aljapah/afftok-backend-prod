/**
 * Link Validator - HMAC Signature + TTL + Replay Protection
 */

import type { Env } from './index';

export interface SignedLinkComponents {
  trackingCode: string;
  timestamp: number;
  nonce: string;
  signature: string;
}

export interface LinkValidationResult {
  valid: boolean;
  reason?: string;
  components?: SignedLinkComponents;
  isLegacy?: boolean;
  trackingCode?: string;
}

export class LinkValidator {
  private env: Env;
  private encoder: TextEncoder;

  constructor(env: Env) {
    this.env = env;
    this.encoder = new TextEncoder();
  }

  /**
   * Parse and validate a signed tracking link
   * Format: trackingCode.timestamp.nonce.signature
   */
  async validateSignedLink(signedCode: string): Promise<LinkValidationResult> {
    // Check for legacy format (no dots)
    if (!signedCode.includes('.')) {
      if (this.env.ALLOW_LEGACY_CODES === 'true') {
        return {
          valid: true,
          isLegacy: true,
          trackingCode: signedCode,
          reason: 'legacy_code_allowed'
        };
      }
      return {
        valid: false,
        reason: 'legacy_codes_not_allowed'
      };
    }

    // Parse components
    const parts = signedCode.split('.');
    if (parts.length !== 4) {
      return {
        valid: false,
        reason: 'invalid_format'
      };
    }

    const [trackingCode, timestampStr, nonce, signature] = parts;
    const timestamp = parseInt(timestampStr, 10);

    if (isNaN(timestamp)) {
      return {
        valid: false,
        reason: 'invalid_timestamp'
      };
    }

    const components: SignedLinkComponents = {
      trackingCode,
      timestamp,
      nonce,
      signature
    };

    // 1. Validate TTL
    const ttlSeconds = parseInt(this.env.LINK_TTL_SECONDS || '86400', 10);
    const now = Math.floor(Date.now() / 1000);
    const age = now - timestamp;

    if (age > ttlSeconds) {
      return {
        valid: false,
        reason: 'link_expired',
        components
      };
    }

    // Allow 5 minutes clock skew
    if (timestamp > now + 300) {
      return {
        valid: false,
        reason: 'timestamp_in_future',
        components
      };
    }

    // 2. Validate signature
    const expectedSignature = await this.createSignature(trackingCode, timestamp, nonce);
    if (!this.secureCompare(signature, expectedSignature)) {
      return {
        valid: false,
        reason: 'invalid_signature',
        components
      };
    }

    // 3. Check replay (nonce)
    const isReplay = await this.checkReplay(nonce);
    if (isReplay) {
      return {
        valid: false,
        reason: 'replay_attempt',
        components
      };
    }

    // Store nonce to prevent replay
    await this.storeNonce(nonce, ttlSeconds);

    return {
      valid: true,
      components,
      trackingCode
    };
  }

  /**
   * Create HMAC-SHA256 signature
   */
  private async createSignature(trackingCode: string, timestamp: number, nonce: string): Promise<string> {
    const secret = this.env.SIGNING_SECRET;
    const message = `${trackingCode}.${timestamp}.${nonce}`;

    const key = await crypto.subtle.importKey(
      'raw',
      this.encoder.encode(secret),
      { name: 'HMAC', hash: 'SHA-256' },
      false,
      ['sign']
    );

    const signatureBuffer = await crypto.subtle.sign(
      'HMAC',
      key,
      this.encoder.encode(message)
    );

    // Convert to hex and take first 12 chars
    const signatureArray = Array.from(new Uint8Array(signatureBuffer));
    const fullSignature = signatureArray.map(b => b.toString(16).padStart(2, '0')).join('');
    return fullSignature.substring(0, 12);
  }

  /**
   * Constant-time string comparison
   */
  private secureCompare(a: string, b: string): boolean {
    if (a.length !== b.length) {
      return false;
    }

    let result = 0;
    for (let i = 0; i < a.length; i++) {
      result |= a.charCodeAt(i) ^ b.charCodeAt(i);
    }
    return result === 0;
  }

  /**
   * Check if nonce was already used (replay attack)
   */
  private async checkReplay(nonce: string): Promise<boolean> {
    try {
      const key = `nonce:${nonce}`;
      const existing = await this.env.NONCE_CACHE.get(key);
      return existing !== null;
    } catch {
      // If KV fails, allow the request (fail open for availability)
      return false;
    }
  }

  /**
   * Store nonce to prevent replay
   */
  private async storeNonce(nonce: string, ttlSeconds: number): Promise<void> {
    try {
      const key = `nonce:${nonce}`;
      await this.env.NONCE_CACHE.put(key, '1', {
        expirationTtl: ttlSeconds
      });
    } catch {
      // Log but don't fail
      console.error('Failed to store nonce:', nonce);
    }
  }

  /**
   * Generate a new signed link (for testing)
   */
  async generateSignedLink(trackingCode: string): Promise<string> {
    const timestamp = Math.floor(Date.now() / 1000);
    const nonce = this.generateNonce();
    const signature = await this.createSignature(trackingCode, timestamp, nonce);
    return `${trackingCode}.${timestamp}.${nonce}.${signature}`;
  }

  /**
   * Generate random nonce
   */
  private generateNonce(): string {
    const array = new Uint8Array(8);
    crypto.getRandomValues(array);
    return Array.from(array).map(b => b.toString(16).padStart(2, '0')).join('');
  }
}

