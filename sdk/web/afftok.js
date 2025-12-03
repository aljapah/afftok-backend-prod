/**
 * AffTok Web SDK
 * 
 * Official JavaScript SDK for AffTok affiliate tracking platform.
 * Supports click and conversion tracking with offline queue and automatic retry.
 * 
 * @version 1.0.0
 * @license MIT
 */

(function(window) {
  'use strict';

  // Configuration
  const Config = {
    DEFAULT_BASE_URL: 'https://api.afftok.com',
    CLICK_ENDPOINT: '/api/sdk/click',
    CONVERSION_ENDPOINT: '/api/sdk/conversion',
    FALLBACK_CLICK_ENDPOINT: '/api/c',
    FALLBACK_CONVERSION_ENDPOINT: '/api/convert',
    
    MAX_QUEUE_SIZE: 1000,
    MAX_RETRY_ATTEMPTS: 5,
    INITIAL_RETRY_DELAY_MS: 1000,
    MAX_RETRY_DELAY_MS: 300000, // 5 minutes
    FLUSH_INTERVAL_MS: 30000, // 30 seconds
    
    MAX_REQUESTS_PER_MINUTE: 60,
    CONNECTION_TIMEOUT_MS: 10000,
    
    QUEUE_KEY: 'afftok_offline_queue',
    DEVICE_ID_KEY: 'afftok_device_id',
    
    SDK_VERSION: '1.0.0',
    SDK_PLATFORM: 'web',
  };

  /**
   * Generate UUID v4
   */
  function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  }

  /**
   * SHA-256 hash function
   */
  async function sha256(message) {
    const msgBuffer = new TextEncoder().encode(message);
    const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
  }

  /**
   * HMAC-SHA256 signature
   */
  async function hmacSha256(key, message) {
    const encoder = new TextEncoder();
    const keyData = encoder.encode(key);
    const msgData = encoder.encode(message);
    
    const cryptoKey = await crypto.subtle.importKey(
      'raw',
      keyData,
      { name: 'HMAC', hash: 'SHA-256' },
      false,
      ['sign']
    );
    
    const signature = await crypto.subtle.sign('HMAC', cryptoKey, msgData);
    const hashArray = Array.from(new Uint8Array(signature));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
  }

  /**
   * AffTok SDK Class
   */
  class AfftokSDK {
    constructor() {
      this.isInitialized = false;
      this.options = null;
      this.queue = [];
      this.flushInterval = null;
      this.isProcessing = false;
      this.deviceId = null;
      this.fingerprint = null;
    }

    /**
     * Initialize the SDK
     * @param {Object} options - SDK options
     */
    async init(options) {
      if (this.isInitialized) {
        this._log('SDK already initialized');
        return;
      }

      this.options = {
        baseUrl: Config.DEFAULT_BASE_URL,
        debug: false,
        autoFlush: true,
        autoTrack: false,
        flushInterval: Config.FLUSH_INTERVAL_MS,
        ...options,
      };

      this._loadQueue();
      await this._initDeviceInfo();

      if (this.options.autoFlush) {
        this._startAutoFlush();
      }

      if (this.options.autoTrack) {
        this._setupAutoTrack();
      }

      // Listen for online/offline events
      window.addEventListener('online', () => this.flush());
      
      // Flush before page unload
      window.addEventListener('beforeunload', () => this._saveQueue());

      this.isInitialized = true;
      this._log('SDK initialized successfully');
      this._log(`Device ID: ${this.deviceId}`);
      this._log(`Pending queue items: ${this.queue.length}`);
    }

    /**
     * Track a click event
     * @param {Object} params - Click parameters
     * @returns {Promise<Object>} Response
     */
    async trackClick(params) {
      this._ensureInitialized();
      
      const payload = await this._buildClickPayload(params);
      
      try {
        const response = await this._sendRequest(Config.CLICK_ENDPOINT, payload);
        if (response.success) {
          this._log(`Click tracked successfully: ${params.offerId}`);
        } else {
          this._enqueue('click', payload);
          this._log(`Click queued for retry: ${params.offerId}`);
        }
        return response;
      } catch (error) {
        this._enqueue('click', payload);
        this._log(`Click queued (offline): ${params.offerId}, error: ${error.message}`);
        return {
          success: false,
          message: 'Click queued for offline retry',
          error: error.message,
        };
      }
    }

    /**
     * Track a click with signed link
     * @param {string} signedLink - Signed tracking link
     * @param {Object} params - Click parameters
     * @returns {Promise<Object>} Response
     */
    async trackSignedClick(signedLink, params) {
      this._ensureInitialized();
      
      const payload = await this._buildClickPayload(params);
      payload.signed_link = signedLink;
      payload.link_validated = true;
      
      try {
        const response = await this._sendRequest(Config.CLICK_ENDPOINT, payload);
        if (!response.success) {
          this._enqueue('click', payload);
        }
        return response;
      } catch (error) {
        this._enqueue('click', payload);
        return {
          success: false,
          message: 'Signed click queued for retry',
          error: error.message,
        };
      }
    }

    /**
     * Track a conversion event
     * @param {Object} params - Conversion parameters
     * @returns {Promise<Object>} Response
     */
    async trackConversion(params) {
      this._ensureInitialized();
      
      const payload = await this._buildConversionPayload(params);
      
      try {
        const response = await this._sendRequest(Config.CONVERSION_ENDPOINT, payload);
        if (response.success) {
          this._log(`Conversion tracked successfully: ${params.transactionId}`);
        } else {
          this._enqueue('conversion', payload);
          this._log(`Conversion queued for retry: ${params.transactionId}`);
        }
        return response;
      } catch (error) {
        this._enqueue('conversion', payload);
        this._log(`Conversion queued (offline): ${params.transactionId}, error: ${error.message}`);
        return {
          success: false,
          message: 'Conversion queued for offline retry',
          error: error.message,
        };
      }
    }

    /**
     * Track conversion with metadata
     * @param {Object} params - Conversion parameters
     * @param {Object} metadata - Additional metadata
     * @returns {Promise<Object>} Response
     */
    async trackConversionWithMeta(params, metadata) {
      this._ensureInitialized();
      
      const payload = await this._buildConversionPayload(params);
      payload.metadata = metadata;
      
      try {
        const response = await this._sendRequest(Config.CONVERSION_ENDPOINT, payload);
        if (!response.success) {
          this._enqueue('conversion', payload);
        }
        return response;
      } catch (error) {
        this._enqueue('conversion', payload);
        return {
          success: false,
          message: 'Conversion with metadata queued for retry',
          error: error.message,
        };
      }
    }

    /**
     * Enable automatic click tracking on links
     */
    autotrack() {
      this._ensureInitialized();
      this._setupAutoTrack();
    }

    /**
     * Manually enqueue an event
     * @param {string} type - Event type
     * @param {Object} payload - Event payload
     * @returns {string} Queue item ID
     */
    enqueue(type, payload) {
      this._ensureInitialized();
      return this._enqueue(type, payload);
    }

    /**
     * Manually flush the queue
     * @returns {Promise<void>}
     */
    async flush() {
      this._ensureInitialized();
      await this._flush();
    }

    /**
     * Get device fingerprint
     * @returns {string} Fingerprint
     */
    getFingerprint() {
      this._ensureInitialized();
      return this.fingerprint || '';
    }

    /**
     * Get device ID
     * @returns {string} Device ID
     */
    getDeviceId() {
      this._ensureInitialized();
      return this.deviceId || '';
    }

    /**
     * Get device info
     * @returns {Object} Device info
     */
    getDeviceInfo() {
      this._ensureInitialized();
      return this._getDeviceInfo();
    }

    /**
     * Get pending queue count
     * @returns {number} Count
     */
    getPendingCount() {
      return this.queue.length;
    }

    /**
     * Check if SDK is ready
     * @returns {boolean}
     */
    isReady() {
      return this.isInitialized;
    }

    /**
     * Get SDK version
     * @returns {string} Version
     */
    getVersion() {
      return Config.SDK_VERSION;
    }

    /**
     * Clear the queue
     */
    clearQueue() {
      this.queue = [];
      this._saveQueue();
    }

    /**
     * Shutdown the SDK
     */
    shutdown() {
      if (this.flushInterval) {
        clearInterval(this.flushInterval);
        this.flushInterval = null;
      }
      this._saveQueue();
      this.isInitialized = false;
      this._log('SDK shutdown');
    }

    // Private methods

    _ensureInitialized() {
      if (!this.isInitialized) {
        throw new Error('AffTok SDK not initialized. Call Afftok.init() first.');
      }
    }

    async _initDeviceInfo() {
      // Get or generate device ID
      this.deviceId = localStorage.getItem(Config.DEVICE_ID_KEY);
      
      if (!this.deviceId) {
        this.deviceId = generateUUID();
        localStorage.setItem(Config.DEVICE_ID_KEY, this.deviceId);
      }

      // Generate fingerprint
      this.fingerprint = await this._generateFingerprint();
    }

    async _generateFingerprint() {
      const components = [
        this.deviceId,
        navigator.userAgent,
        navigator.language,
        screen.width + 'x' + screen.height,
        screen.colorDepth,
        new Date().getTimezoneOffset(),
        navigator.hardwareConcurrency || 'unknown',
        navigator.platform || 'unknown',
      ];
      
      const data = components.join('|');
      return await sha256(data);
    }

    _getDeviceInfo() {
      return {
        device_id: this.deviceId,
        fingerprint: this.fingerprint,
        platform: Config.SDK_PLATFORM,
        sdk_version: Config.SDK_VERSION,
        user_agent: navigator.userAgent,
        language: navigator.language,
        screen: `${screen.width}x${screen.height}`,
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        referrer: document.referrer,
        url: window.location.href,
      };
    }

    async _buildClickPayload(params) {
      const timestamp = Date.now();
      const nonce = this._generateNonce();
      
      const payload = {
        api_key: this.options.apiKey,
        advertiser_id: this.options.advertiserId,
        offer_id: params.offerId,
        timestamp,
        nonce,
        device_info: this._getDeviceInfo(),
      };
      
      if (this.options.userId) payload.user_id = this.options.userId;
      if (params.trackingCode) payload.tracking_code = params.trackingCode;
      if (params.subId1) payload.sub_id_1 = params.subId1;
      if (params.subId2) payload.sub_id_2 = params.subId2;
      if (params.subId3) payload.sub_id_3 = params.subId3;
      if (params.customParams) payload.custom_params = params.customParams;
      
      payload.signature = await this._generateSignature(timestamp, nonce);
      
      return payload;
    }

    async _buildConversionPayload(params) {
      const timestamp = Date.now();
      const nonce = this._generateNonce();
      
      const payload = {
        api_key: this.options.apiKey,
        advertiser_id: this.options.advertiserId,
        offer_id: params.offerId,
        transaction_id: params.transactionId,
        status: params.status || 'pending',
        currency: params.currency || 'USD',
        timestamp,
        nonce,
        device_info: this._getDeviceInfo(),
      };
      
      if (this.options.userId) payload.user_id = this.options.userId;
      if (params.clickId) payload.click_id = params.clickId;
      if (params.amount !== undefined) payload.amount = params.amount;
      if (params.customParams) payload.custom_params = params.customParams;
      
      payload.signature = await this._generateSignature(timestamp, nonce);
      
      return payload;
    }

    async _sendRequest(endpoint, payload) {
      const url = `${this.options.baseUrl}${endpoint}`;
      
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), Config.CONNECTION_TIMEOUT_MS);
      
      try {
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-API-Key': this.options.apiKey,
            'X-SDK-Version': Config.SDK_VERSION,
            'X-SDK-Platform': Config.SDK_PLATFORM,
          },
          body: JSON.stringify(payload),
          signal: controller.signal,
        });
        
        clearTimeout(timeoutId);
        
        if (response.ok) {
          const data = await response.json();
          return { success: true, ...data };
        } else {
          // Try fallback
          if (endpoint === Config.CLICK_ENDPOINT) {
            return this._sendFallbackRequest(Config.FALLBACK_CLICK_ENDPOINT, payload);
          } else if (endpoint === Config.CONVERSION_ENDPOINT) {
            return this._sendFallbackRequest(Config.FALLBACK_CONVERSION_ENDPOINT, payload);
          }
          return { success: false, error: `HTTP ${response.status}` };
        }
      } catch (error) {
        clearTimeout(timeoutId);
        throw error;
      }
    }

    async _sendFallbackRequest(endpoint, payload) {
      try {
        const url = `${this.options.baseUrl}${endpoint}`;
        
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-API-Key': this.options.apiKey,
          },
          body: JSON.stringify(payload),
        });
        
        if (response.ok) {
          return { success: true, message: 'Tracked via fallback' };
        }
        return { success: false, error: `Fallback failed: ${response.status}` };
      } catch (error) {
        return { success: false, error: `Fallback error: ${error.message}` };
      }
    }

    async _generateSignature(timestamp, nonce) {
      const dataToSign = `${this.options.apiKey}|${this.options.advertiserId}|${timestamp}|${nonce}`;
      return await hmacSha256(this.options.apiKey, dataToSign);
    }

    _generateNonce() {
      const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
      let result = '';
      for (let i = 0; i < 32; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
      }
      return result;
    }

    _enqueue(type, payload) {
      const id = generateUUID();
      const item = {
        id,
        type,
        payload,
        timestamp: Date.now(),
        retryCount: 0,
        nextRetryTime: 0,
      };
      
      if (this.queue.length >= Config.MAX_QUEUE_SIZE) {
        this.queue.shift();
        this._log('Queue full, removed oldest item');
      }
      
      this.queue.push(item);
      this._saveQueue();
      this._log(`Enqueued ${type} event: ${id}`);
      
      return id;
    }

    _startAutoFlush() {
      if (this.flushInterval) {
        clearInterval(this.flushInterval);
      }
      this.flushInterval = setInterval(() => this._flush(), this.options.flushInterval);
      this._log(`Auto-flush started with interval: ${this.options.flushInterval}ms`);
    }

    async _flush() {
      if (this.isProcessing) {
        this._log('Flush already in progress, skipping');
        return;
      }
      
      if (!navigator.onLine) {
        this._log('Offline, skipping flush');
        return;
      }
      
      this.isProcessing = true;
      this._log(`Starting flush, ${this.queue.length} items in queue`);
      
      const now = Date.now();
      const pendingItems = this.queue.filter(item => item.nextRetryTime <= now);
      
      for (const item of pendingItems) {
        try {
          const success = await this._processQueueItem(item);
          if (success) {
            this.queue = this.queue.filter(i => i.id !== item.id);
            this._log(`Completed: ${item.id}`);
          } else {
            this._markForRetry(item);
          }
        } catch (error) {
          this._log(`Error processing item ${item.id}: ${error.message}`);
          this._markForRetry(item);
        }
      }
      
      this._saveQueue();
      this.isProcessing = false;
      this._log(`Flush completed, ${this.queue.length} items remaining`);
    }

    async _processQueueItem(item) {
      try {
        let response;
        
        if (item.type === 'click') {
          response = await this._sendRequest(Config.CLICK_ENDPOINT, item.payload);
        } else if (item.type === 'conversion') {
          response = await this._sendRequest(Config.CONVERSION_ENDPOINT, item.payload);
        } else {
          return false;
        }
        
        return response.success;
      } catch (error) {
        return false;
      }
    }

    _markForRetry(item) {
      if (item.retryCount >= Config.MAX_RETRY_ATTEMPTS) {
        this.queue = this.queue.filter(i => i.id !== item.id);
        this._log(`Max retries reached, removing: ${item.id}`);
        return;
      }
      
      const delay = Math.min(
        Config.INITIAL_RETRY_DELAY_MS * Math.pow(2, item.retryCount),
        Config.MAX_RETRY_DELAY_MS
      );
      const jitter = Math.random() * delay * 0.1;
      
      item.retryCount++;
      item.nextRetryTime = Date.now() + delay + jitter;
      
      this._log(`Marked for retry (${item.retryCount}/${Config.MAX_RETRY_ATTEMPTS}): ${item.id}`);
    }

    _loadQueue() {
      try {
        const jsonString = localStorage.getItem(Config.QUEUE_KEY);
        if (jsonString) {
          this.queue = JSON.parse(jsonString);
          this._log(`Loaded ${this.queue.length} items from storage`);
        }
      } catch (error) {
        this._log(`Error loading queue: ${error.message}`);
      }
    }

    _saveQueue() {
      try {
        localStorage.setItem(Config.QUEUE_KEY, JSON.stringify(this.queue));
      } catch (error) {
        this._log(`Error saving queue: ${error.message}`);
      }
    }

    _setupAutoTrack() {
      document.addEventListener('click', (event) => {
        const link = event.target.closest('a[data-afftok-offer]');
        if (link) {
          const offerId = link.getAttribute('data-afftok-offer');
          const trackingCode = link.getAttribute('data-afftok-code');
          const subId1 = link.getAttribute('data-afftok-sub1');
          const subId2 = link.getAttribute('data-afftok-sub2');
          const subId3 = link.getAttribute('data-afftok-sub3');
          
          this.trackClick({
            offerId,
            trackingCode,
            subId1,
            subId2,
            subId3,
          });
        }
      });
      
      this._log('Auto-track enabled for [data-afftok-offer] links');
    }

    _log(message) {
      if (this.options?.debug) {
        console.log(`[AffTok SDK] ${message}`);
      }
    }
  }

  // Create and expose singleton
  const Afftok = new AfftokSDK();
  
  // Expose to window
  window.Afftok = Afftok;
  
  // AMD support
  if (typeof define === 'function' && define.amd) {
    define([], function() { return Afftok; });
  }
  
  // CommonJS support
  if (typeof module === 'object' && module.exports) {
    module.exports = Afftok;
  }

})(typeof window !== 'undefined' ? window : this);

