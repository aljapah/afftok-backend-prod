/**
 * AffTok Conversion Tracking Pixel
 * Version: 1.0.0
 * 
 * Usage:
 * 1. Add this script to your thank you/confirmation page:
 *    <script src="https://go.afftokapp.com/pixel.js"></script>
 * 
 * 2. Track conversions:
 *    AffTok.track('purchase', { value: 100, currency: 'USD', order_id: 'ORD-123' });
 * 
 * Or use simple image pixel:
 *    <img src="https://go.afftokapp.com/api/pixel/convert?img=1" width="1" height="1" />
 */

(function(window, document) {
    'use strict';

    var AffTok = window.AffTok || {};
    var API_BASE = 'https://go.afftokapp.com';
    var COOKIE_NAME = 'afftok_click_id';
    var COOKIE_DAYS = 30;

    /**
     * Get cookie value by name
     */
    function getCookie(name) {
        var nameEQ = name + '=';
        var cookies = document.cookie.split(';');
        for (var i = 0; i < cookies.length; i++) {
            var cookie = cookies[i].trim();
            if (cookie.indexOf(nameEQ) === 0) {
                return decodeURIComponent(cookie.substring(nameEQ.length));
            }
        }
        return null;
    }

    /**
     * Set cookie
     */
    function setCookie(name, value, days) {
        var expires = '';
        if (days) {
            var date = new Date();
            date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
            expires = '; expires=' + date.toUTCString();
        }
        document.cookie = name + '=' + encodeURIComponent(value) + expires + '; path=/; SameSite=Lax';
    }

    /**
     * Get click_id from URL parameters
     */
    function getClickIDFromURL() {
        var params = new URLSearchParams(window.location.search);
        return params.get('click_id') || params.get('aff_id') || null;
    }

    /**
     * Get stored click_id (from cookie or URL)
     */
    function getClickID() {
        // First check URL (fresh click)
        var urlClickID = getClickIDFromURL();
        if (urlClickID) {
            // Store in cookie for future conversions
            setCookie(COOKIE_NAME, urlClickID, COOKIE_DAYS);
            return urlClickID;
        }
        
        // Fall back to cookie (returning visitor)
        return getCookie(COOKIE_NAME);
    }

    /**
     * Send conversion to AffTok
     */
    function sendConversion(eventType, data, callback) {
        var clickID = getClickID();
        
        if (!clickID) {
            console.log('[AffTok] No click_id found - visitor did not come from affiliate link');
            if (callback) callback(false, 'No click_id');
            return;
        }

        var payload = {
            click_id: clickID,
            event: eventType,
            amount: data.value || data.amount || 0,
            currency: data.currency || 'USD',
            order_id: data.order_id || data.orderId || '',
            timestamp: new Date().toISOString(),
            page_url: window.location.href,
            referrer: document.referrer
        };

        // Send via fetch (modern browsers)
        if (window.fetch) {
            fetch(API_BASE + '/api/pixel/convert', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload),
                credentials: 'include'
            })
            .then(function(response) {
                return response.json();
            })
            .then(function(result) {
                console.log('[AffTok] Conversion tracked:', result);
                if (callback) callback(true, result);
            })
            .catch(function(error) {
                console.error('[AffTok] Error tracking conversion:', error);
                // Fallback to image pixel
                sendImagePixel(payload);
                if (callback) callback(false, error);
            });
        } else {
            // Fallback for older browsers
            sendImagePixel(payload);
            if (callback) callback(true, null);
        }
    }

    /**
     * Send conversion via image pixel (fallback)
     */
    function sendImagePixel(payload) {
        var img = new Image(1, 1);
        var params = new URLSearchParams();
        params.append('click_id', payload.click_id);
        params.append('amount', payload.amount);
        params.append('currency', payload.currency);
        params.append('order_id', payload.order_id);
        params.append('img', '1');
        
        img.src = API_BASE + '/api/pixel/convert?' + params.toString();
    }

    /**
     * Track event
     */
    AffTok.track = function(eventType, data, callback) {
        data = data || {};
        
        switch (eventType) {
            case 'purchase':
            case 'conversion':
            case 'sale':
                sendConversion('purchase', data, callback);
                break;
            case 'lead':
            case 'signup':
            case 'register':
                sendConversion('lead', data, callback);
                break;
            case 'install':
                sendConversion('install', data, callback);
                break;
            default:
                sendConversion(eventType, data, callback);
        }
    };

    /**
     * Simple conversion tracking (alias)
     */
    AffTok.conversion = function(data, callback) {
        AffTok.track('purchase', data, callback);
    };

    /**
     * Get current click_id
     */
    AffTok.getClickID = function() {
        return getClickID();
    };

    /**
     * Check if visitor came from affiliate
     */
    AffTok.hasAffiliate = function() {
        return !!getClickID();
    };

    /**
     * Initialize - store click_id from URL if present
     */
    AffTok.init = function() {
        var clickID = getClickIDFromURL();
        if (clickID) {
            setCookie(COOKIE_NAME, clickID, COOKIE_DAYS);
            console.log('[AffTok] Click ID stored:', clickID);
        }
    };

    // Auto-initialize on page load
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', AffTok.init);
    } else {
        AffTok.init();
    }

    // Expose to global scope
    window.AffTok = AffTok;

    console.log('[AffTok] Pixel loaded. Version 1.0.0');

})(window, document);

