/**
 * Device Detector - Device/Browser/OS detection at edge
 */

export interface DeviceInfo {
  type: 'mobile' | 'tablet' | 'desktop' | 'bot' | 'unknown';
  browser: string;
  browserVersion: string;
  os: string;
  osVersion: string;
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isBot: boolean;
  brand?: string;
  model?: string;
}

export interface DeviceRule {
  id: string;
  scope_type: 'offer' | 'advertiser' | 'tenant' | 'global';
  scope_id: string;
  mode: 'allow' | 'block';
  device_types: string[];
  browsers?: string[];
  os_types?: string[];
  priority: number;
  status: 'active' | 'inactive';
}

export interface DeviceValidationResult {
  allowed: boolean;
  reason?: string;
  rule?: DeviceRule;
  deviceInfo: DeviceInfo;
}

// Mobile device patterns
const MOBILE_PATTERNS = [
  /android.*mobile/i,
  /iphone/i,
  /ipod/i,
  /blackberry/i,
  /windows phone/i,
  /opera mini/i,
  /opera mobi/i,
  /mobile.*firefox/i,
  /mobile.*safari/i,
  /webos/i,
  /palm/i,
  /symbian/i,
  /j2me/i,
  /midp/i,
  /cldc/i,
  /netfront/i,
  /up\.browser/i,
  /up\.link/i,
  /fennec/i,
  /maemo/i,
  /meego/i,
  /mobile/i,
];

// Tablet patterns
const TABLET_PATTERNS = [
  /ipad/i,
  /android(?!.*mobile)/i,
  /tablet/i,
  /kindle/i,
  /playbook/i,
  /silk/i,
  /surface/i,
  /xoom/i,
  /nexus\s*[0-9]+/i,
];

// Browser patterns
const BROWSER_PATTERNS: Array<{ name: string; pattern: RegExp; versionPattern?: RegExp }> = [
  { name: 'Chrome', pattern: /chrome/i, versionPattern: /chrome\/(\d+)/i },
  { name: 'Firefox', pattern: /firefox/i, versionPattern: /firefox\/(\d+)/i },
  { name: 'Safari', pattern: /safari/i, versionPattern: /version\/(\d+)/i },
  { name: 'Edge', pattern: /edg/i, versionPattern: /edg\/(\d+)/i },
  { name: 'Opera', pattern: /opera|opr/i, versionPattern: /(?:opera|opr)\/(\d+)/i },
  { name: 'IE', pattern: /msie|trident/i, versionPattern: /(?:msie\s|rv:)(\d+)/i },
  { name: 'Samsung', pattern: /samsungbrowser/i, versionPattern: /samsungbrowser\/(\d+)/i },
  { name: 'UC Browser', pattern: /ucbrowser/i, versionPattern: /ucbrowser\/(\d+)/i },
];

// OS patterns
const OS_PATTERNS: Array<{ name: string; pattern: RegExp; versionPattern?: RegExp }> = [
  { name: 'iOS', pattern: /iphone|ipad|ipod/i, versionPattern: /os\s(\d+)/i },
  { name: 'Android', pattern: /android/i, versionPattern: /android\s(\d+)/i },
  { name: 'Windows', pattern: /windows/i, versionPattern: /windows\snt\s(\d+\.\d+)/i },
  { name: 'macOS', pattern: /mac\sos|macintosh/i, versionPattern: /mac\sos\sx\s(\d+)/i },
  { name: 'Linux', pattern: /linux/i },
  { name: 'Chrome OS', pattern: /cros/i },
];

export class DeviceDetector {
  /**
   * Detect device from user-agent
   */
  detect(userAgent: string): DeviceInfo {
    const ua = userAgent.toLowerCase();

    // Detect device type
    let type: DeviceInfo['type'] = 'unknown';
    let isMobile = false;
    let isTablet = false;
    let isDesktop = false;
    let isBot = false;

    // Check for bot first
    if (this.isBot(ua)) {
      type = 'bot';
      isBot = true;
    } else if (TABLET_PATTERNS.some(p => p.test(userAgent))) {
      type = 'tablet';
      isTablet = true;
    } else if (MOBILE_PATTERNS.some(p => p.test(userAgent))) {
      type = 'mobile';
      isMobile = true;
    } else {
      type = 'desktop';
      isDesktop = true;
    }

    // Detect browser
    let browser = 'Unknown';
    let browserVersion = '';
    for (const b of BROWSER_PATTERNS) {
      if (b.pattern.test(userAgent)) {
        browser = b.name;
        if (b.versionPattern) {
          const match = userAgent.match(b.versionPattern);
          if (match) {
            browserVersion = match[1];
          }
        }
        break;
      }
    }

    // Detect OS
    let os = 'Unknown';
    let osVersion = '';
    for (const o of OS_PATTERNS) {
      if (o.pattern.test(userAgent)) {
        os = o.name;
        if (o.versionPattern) {
          const match = userAgent.match(o.versionPattern);
          if (match) {
            osVersion = match[1];
          }
        }
        break;
      }
    }

    // Detect brand/model for mobile
    const { brand, model } = this.detectBrandModel(userAgent);

    return {
      type,
      browser,
      browserVersion,
      os,
      osVersion,
      isMobile,
      isTablet,
      isDesktop,
      isBot,
      brand,
      model
    };
  }

  /**
   * Check if user-agent is a bot
   */
  private isBot(ua: string): boolean {
    const botPatterns = [
      'bot', 'crawl', 'spider', 'slurp', 'mediapartners',
      'wget', 'curl', 'python', 'java/', 'libwww', 'httpie'
    ];
    return botPatterns.some(p => ua.includes(p));
  }

  /**
   * Detect brand and model
   */
  private detectBrandModel(userAgent: string): { brand?: string; model?: string } {
    // iPhone
    if (/iphone/i.test(userAgent)) {
      return { brand: 'Apple', model: 'iPhone' };
    }

    // iPad
    if (/ipad/i.test(userAgent)) {
      return { brand: 'Apple', model: 'iPad' };
    }

    // Samsung
    const samsungMatch = userAgent.match(/samsung[- ]?([\w-]+)/i);
    if (samsungMatch) {
      return { brand: 'Samsung', model: samsungMatch[1] };
    }

    // Huawei
    const huaweiMatch = userAgent.match(/huawei[- ]?([\w-]+)/i);
    if (huaweiMatch) {
      return { brand: 'Huawei', model: huaweiMatch[1] };
    }

    // Xiaomi
    const xiaomiMatch = userAgent.match(/xiaomi[- ]?([\w-]+)/i);
    if (xiaomiMatch) {
      return { brand: 'Xiaomi', model: xiaomiMatch[1] };
    }

    // Google Pixel
    const pixelMatch = userAgent.match(/pixel[- ]?([\w-]+)?/i);
    if (pixelMatch) {
      return { brand: 'Google', model: `Pixel ${pixelMatch[1] || ''}`.trim() };
    }

    return {};
  }

  /**
   * Validate device against rules
   */
  validate(
    deviceInfo: DeviceInfo,
    rule?: DeviceRule
  ): DeviceValidationResult {
    if (!rule) {
      return {
        allowed: true,
        deviceInfo
      };
    }

    let matches = false;

    // Check device type
    if (rule.device_types.length > 0) {
      matches = rule.device_types.includes(deviceInfo.type) ||
                rule.device_types.includes('*');
    }

    // Check browser
    if (rule.browsers && rule.browsers.length > 0) {
      const browserMatch = rule.browsers.includes(deviceInfo.browser.toLowerCase()) ||
                           rule.browsers.includes('*');
      matches = matches && browserMatch;
    }

    // Check OS
    if (rule.os_types && rule.os_types.length > 0) {
      const osMatch = rule.os_types.includes(deviceInfo.os.toLowerCase()) ||
                      rule.os_types.includes('*');
      matches = matches && osMatch;
    }

    // Apply mode
    let allowed: boolean;
    if (rule.mode === 'allow') {
      allowed = matches;
    } else {
      allowed = !matches;
    }

    return {
      allowed,
      reason: allowed ? undefined : `device_${rule.mode}_${deviceInfo.type}`,
      rule,
      deviceInfo
    };
  }

  /**
   * Get device type string
   */
  getDeviceTypeString(deviceInfo: DeviceInfo): string {
    if (deviceInfo.isMobile) return 'mobile';
    if (deviceInfo.isTablet) return 'tablet';
    if (deviceInfo.isDesktop) return 'desktop';
    if (deviceInfo.isBot) return 'bot';
    return 'unknown';
  }

  /**
   * Get connection type from Cloudflare
   */
  getConnectionType(request: Request): string {
    const cf = request.cf as any;
    
    // Check for HTTP/2 or HTTP/3
    if (cf?.httpProtocol) {
      return cf.httpProtocol;
    }

    return 'unknown';
  }

  /**
   * Estimate connection speed from Cloudflare hints
   */
  estimateConnectionSpeed(request: Request): 'slow' | 'medium' | 'fast' | 'unknown' {
    const cf = request.cf as any;

    // Use RTT if available
    if (cf?.clientTcpRtt) {
      const rtt = cf.clientTcpRtt;
      if (rtt > 500) return 'slow';
      if (rtt > 100) return 'medium';
      return 'fast';
    }

    return 'unknown';
  }
}

