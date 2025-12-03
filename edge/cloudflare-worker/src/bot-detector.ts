/**
 * Bot Detector - Edge-side bot filtering
 * 100+ bot signatures, headless browsers, Tor, DC ASNs
 */

import type { Env } from './index';

export interface BotDetectionResult {
  isBot: boolean;
  confidence: number;
  riskScore: number;
  indicators: string[];
  category?: string;
}

// Known bot user-agent patterns
const BOT_PATTERNS: RegExp[] = [
  // Search engines
  /googlebot/i,
  /bingbot/i,
  /yandexbot/i,
  /baiduspider/i,
  /duckduckbot/i,
  /slurp/i,
  /ia_archiver/i,
  /sogou/i,
  /exabot/i,
  /facebot/i,
  /facebookexternalhit/i,
  /twitterbot/i,
  /linkedinbot/i,
  /pinterest/i,
  /whatsapp/i,
  /telegrambot/i,
  /slackbot/i,
  /discordbot/i,
  /applebot/i,
  
  // Crawlers
  /crawl/i,
  /spider/i,
  /bot\b/i,
  /scraper/i,
  /wget/i,
  /curl/i,
  /httpie/i,
  /python-requests/i,
  /python-urllib/i,
  /go-http-client/i,
  /java\//i,
  /libwww/i,
  /lwp-/i,
  /apache-httpclient/i,
  /okhttp/i,
  /axios/i,
  /node-fetch/i,
  /got\//i,
  /superagent/i,
  /request\//i,
  /undici/i,
  
  // Headless browsers
  /headless/i,
  /phantomjs/i,
  /slimerjs/i,
  /puppeteer/i,
  /playwright/i,
  /selenium/i,
  /webdriver/i,
  /chromedriver/i,
  /geckodriver/i,
  /cypress/i,
  /nightmare/i,
  /zombie/i,
  /casperjs/i,
  
  // SEO tools
  /semrush/i,
  /ahrefs/i,
  /moz\./i,
  /majestic/i,
  /screaming frog/i,
  /seokicks/i,
  /sistrix/i,
  /dotbot/i,
  /rogerbot/i,
  
  // Monitoring
  /uptimerobot/i,
  /pingdom/i,
  /statuscake/i,
  /newrelic/i,
  /datadog/i,
  /site24x7/i,
  /gtmetrix/i,
  /pagespeed/i,
  /lighthouse/i,
  
  // Security scanners
  /nikto/i,
  /nessus/i,
  /acunetix/i,
  /burp/i,
  /zap/i,
  /nmap/i,
  /masscan/i,
  /sqlmap/i,
  /wpscan/i,
  
  // Misc bots
  /feedfetcher/i,
  /mediapartners/i,
  /adsbot/i,
  /apis-google/i,
  /bytespider/i,
  /petalbot/i,
  /mj12bot/i,
  /ahrefsbot/i,
  /blexbot/i,
  /dataforseo/i,
  /ccbot/i,
  /gptbot/i,
  /claudebot/i,
  /anthropic/i,
  /chatgpt/i,
  /openai/i,
  /cohere/i,
  /perplexity/i,
  /diffbot/i,
  /commoncrawl/i,
];

// Known data center ASNs
const DC_ASNS: Set<number> = new Set([
  // AWS
  16509, 14618,
  // Google Cloud
  15169, 396982,
  // Microsoft Azure
  8075, 8068, 8069,
  // DigitalOcean
  14061,
  // Linode
  63949,
  // Vultr
  20473,
  // OVH
  16276,
  // Hetzner
  24940,
  // Cloudflare
  13335,
  // Fastly
  54113,
  // Akamai
  20940, 16625,
  // Oracle Cloud
  31898,
  // IBM Cloud
  36351,
  // Alibaba Cloud
  45102,
  // Tencent Cloud
  132203,
  // Scaleway
  12876,
  // Contabo
  51167,
  // HostGator
  46606,
  // GoDaddy
  26496,
  // Rackspace
  33070,
  // Leaseweb
  60781,
]);

// Tor exit node indicators
const TOR_INDICATORS = [
  'tor',
  '.onion',
  'exit',
];

// Suspicious header patterns
const SUSPICIOUS_HEADERS: string[] = [
  'x-forwarded-for',
  'via',
  'x-real-ip',
];

export class BotDetector {
  private env: Env;

  constructor(env: Env) {
    this.env = env;
  }

  /**
   * Detect if request is from a bot
   */
  detect(request: Request): BotDetectionResult {
    const indicators: string[] = [];
    let riskScore = 0;
    let category: string | undefined;

    const userAgent = request.headers.get('user-agent') || '';
    const cf = request.cf as any;

    // 1. Check user-agent patterns
    for (const pattern of BOT_PATTERNS) {
      if (pattern.test(userAgent)) {
        indicators.push(`bot_pattern:${pattern.source.substring(0, 20)}`);
        riskScore += 50;
        category = 'known_bot';
        break;
      }
    }

    // 2. Empty or missing user-agent
    if (!userAgent || userAgent.length < 10) {
      indicators.push('empty_or_short_ua');
      riskScore += 40;
    }

    // 3. Check for headless browser indicators
    if (this.isHeadlessBrowser(userAgent)) {
      indicators.push('headless_browser');
      riskScore += 60;
      category = 'headless';
    }

    // 4. Check ASN (data center)
    if (cf?.asn && DC_ASNS.has(cf.asn)) {
      indicators.push(`dc_asn:${cf.asn}`);
      riskScore += 30;
      category = category || 'datacenter';
    }

    // 5. Check for Tor
    if (cf?.isEUCountry === false && this.isTorExit(userAgent, request)) {
      indicators.push('tor_exit');
      riskScore += 70;
      category = 'tor';
    }

    // 6. Missing standard headers
    const missingHeaders = this.checkMissingHeaders(request);
    if (missingHeaders.length > 0) {
      indicators.push(...missingHeaders.map(h => `missing:${h}`));
      riskScore += missingHeaders.length * 10;
    }

    // 7. Impossible browser combinations
    if (this.hasImpossibleBrowser(userAgent)) {
      indicators.push('impossible_browser');
      riskScore += 50;
    }

    // 8. Check Accept-Language
    const acceptLang = request.headers.get('accept-language');
    if (!acceptLang || acceptLang === '*') {
      indicators.push('no_accept_language');
      riskScore += 15;
    }

    // 9. Check Accept-Encoding
    const acceptEnc = request.headers.get('accept-encoding');
    if (!acceptEnc) {
      indicators.push('no_accept_encoding');
      riskScore += 15;
    }

    // 10. Check Connection header
    const connection = request.headers.get('connection');
    if (!connection) {
      indicators.push('no_connection_header');
      riskScore += 10;
    }

    // 11. Suspicious refresh patterns (would need state)
    // This is handled at backend level

    // Calculate confidence
    const confidence = Math.min(riskScore / 100, 1);
    const isBot = riskScore >= 50;

    return {
      isBot,
      confidence,
      riskScore: Math.min(riskScore, 100),
      indicators,
      category
    };
  }

  /**
   * Check for headless browser signatures
   */
  private isHeadlessBrowser(userAgent: string): boolean {
    const headlessIndicators = [
      'headlesschrome',
      'headless',
      'phantomjs',
      'slimerjs',
      'puppeteer',
      'playwright',
      'selenium',
      'webdriver',
      'chromedriver',
      'geckodriver',
    ];

    const ua = userAgent.toLowerCase();
    return headlessIndicators.some(indicator => ua.includes(indicator));
  }

  /**
   * Check for Tor exit indicators
   */
  private isTorExit(userAgent: string, request: Request): boolean {
    const ua = userAgent.toLowerCase();
    return TOR_INDICATORS.some(indicator => ua.includes(indicator));
  }

  /**
   * Check for missing standard browser headers
   */
  private checkMissingHeaders(request: Request): string[] {
    const missing: string[] = [];
    const standardHeaders = [
      'accept',
      'accept-language',
      'accept-encoding',
    ];

    for (const header of standardHeaders) {
      if (!request.headers.get(header)) {
        missing.push(header);
      }
    }

    return missing;
  }

  /**
   * Check for impossible browser combinations
   */
  private hasImpossibleBrowser(userAgent: string): boolean {
    const ua = userAgent.toLowerCase();

    // Chrome on Windows claiming to be Safari
    if (ua.includes('chrome') && ua.includes('safari') && !ua.includes('mobile')) {
      // This is actually normal for Chrome
      return false;
    }

    // IE on Mac
    if (ua.includes('msie') && ua.includes('mac')) {
      return true;
    }

    // Very old browser versions with modern features
    if (ua.includes('chrome/1.') || ua.includes('firefox/1.')) {
      return true;
    }

    // Android browser on iOS
    if (ua.includes('android') && ua.includes('iphone')) {
      return true;
    }

    return false;
  }

  /**
   * Get bot category description
   */
  getCategoryDescription(category?: string): string {
    switch (category) {
      case 'known_bot': return 'Known bot or crawler';
      case 'headless': return 'Headless browser automation';
      case 'datacenter': return 'Data center IP';
      case 'tor': return 'Tor exit node';
      default: return 'Unknown';
    }
  }
}

