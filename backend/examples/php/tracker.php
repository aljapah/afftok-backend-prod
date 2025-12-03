<?php
/**
 * AffTok Server-to-Server Integration - PHP Example
 * 
 * This example shows how to send postbacks/conversions from your server
 * to AffTok using the Server-to-Server API.
 */

class AfftokTracker {
    private $apiKey;
    private $advertiserId;
    private $baseUrl;
    private $timeout;

    public function __construct($apiKey, $advertiserId, $baseUrl = 'https://api.afftok.com') {
        $this->apiKey = $apiKey;
        $this->advertiserId = $advertiserId;
        $this->baseUrl = $baseUrl;
        $this->timeout = 30;
    }

    /**
     * Generate HMAC-SHA256 signature
     */
    private function generateSignature($timestamp, $nonce) {
        $dataToSign = "{$this->apiKey}|{$this->advertiserId}|{$timestamp}|{$nonce}";
        return hash_hmac('sha256', $dataToSign, $this->apiKey);
    }

    /**
     * Generate random nonce
     */
    private function generateNonce($length = 32) {
        $chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        $nonce = '';
        for ($i = 0; $i < $length; $i++) {
            $nonce .= $chars[random_int(0, strlen($chars) - 1)];
        }
        return $nonce;
    }

    /**
     * Send HTTP POST request
     */
    private function sendRequest($endpoint, $payload) {
        $url = $this->baseUrl . $endpoint;
        
        $ch = curl_init($url);
        curl_setopt_array($ch, [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_POST => true,
            CURLOPT_POSTFIELDS => json_encode($payload),
            CURLOPT_HTTPHEADER => [
                'Content-Type: application/json',
                'X-API-Key: ' . $this->apiKey,
            ],
            CURLOPT_TIMEOUT => $this->timeout,
            CURLOPT_SSL_VERIFYPEER => true,
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        curl_close($ch);

        if ($error) {
            return ['success' => false, 'error' => $error];
        }

        $data = json_decode($response, true);
        
        if ($httpCode >= 200 && $httpCode < 300) {
            return ['success' => true, 'data' => $data];
        }

        return ['success' => false, 'error' => $data ?? "HTTP $httpCode"];
    }

    /**
     * Send a postback/conversion to AffTok
     * 
     * @param array $params Conversion parameters
     * @return array Result
     */
    public function sendPostback($params) {
        $timestamp = round(microtime(true) * 1000);
        $nonce = $this->generateNonce();
        $signature = $this->generateSignature($timestamp, $nonce);

        $payload = [
            'api_key' => $this->apiKey,
            'advertiser_id' => $this->advertiserId,
            'offer_id' => $params['offer_id'],
            'transaction_id' => $params['transaction_id'],
            'status' => $params['status'] ?? 'approved',
            'currency' => $params['currency'] ?? 'USD',
            'timestamp' => $timestamp,
            'nonce' => $nonce,
            'signature' => $signature,
        ];

        // Add optional fields
        if (isset($params['click_id'])) {
            $payload['click_id'] = $params['click_id'];
        }
        if (isset($params['amount'])) {
            $payload['amount'] = $params['amount'];
        }
        if (isset($params['custom_params'])) {
            $payload['custom_params'] = $params['custom_params'];
        }

        return $this->sendRequest('/api/postback', $payload);
    }

    /**
     * Track a click event (server-side)
     * 
     * @param array $params Click parameters
     * @return array Result
     */
    public function trackClick($params) {
        $timestamp = round(microtime(true) * 1000);
        $nonce = $this->generateNonce();
        $signature = $this->generateSignature($timestamp, $nonce);

        $payload = [
            'api_key' => $this->apiKey,
            'advertiser_id' => $this->advertiserId,
            'offer_id' => $params['offer_id'],
            'timestamp' => $timestamp,
            'nonce' => $nonce,
            'signature' => $signature,
        ];

        // Add optional fields
        $optionalFields = ['tracking_code', 'sub_id_1', 'sub_id_2', 'sub_id_3', 'ip', 'user_agent', 'custom_params'];
        foreach ($optionalFields as $field) {
            if (isset($params[$field])) {
                $payload[$field] = $params[$field];
            }
        }

        return $this->sendRequest('/api/sdk/click', $payload);
    }

    /**
     * Batch send multiple conversions
     * 
     * @param array $conversions Array of conversion parameters
     * @return array Results
     */
    public function sendBatchPostbacks($conversions) {
        $results = [];
        
        foreach ($conversions as $conversion) {
            $result = $this->sendPostback($conversion);
            $results[] = [
                'transaction_id' => $conversion['transaction_id'],
                'success' => $result['success'],
                'error' => $result['error'] ?? null,
            ];
            
            // Small delay to avoid rate limiting
            usleep(100000); // 100ms
        }
        
        return $results;
    }
}

// Example usage
if (php_sapi_name() === 'cli' || !isset($_SERVER['REQUEST_URI'])) {
    // Configuration
    $apiKey = getenv('AFFTOK_API_KEY') ?: 'your_api_key';
    $advertiserId = getenv('AFFTOK_ADVERTISER_ID') ?: 'your_advertiser_id';
    
    $tracker = new AfftokTracker($apiKey, $advertiserId);
    
    echo "AffTok Server-to-Server Integration Example (PHP)\n\n";

    // Example 1: Send a simple conversion
    echo "1. Sending a simple conversion...\n";
    $result = $tracker->sendPostback([
        'offer_id' => 'offer_123',
        'transaction_id' => 'txn_' . time(),
        'amount' => 29.99,
        'status' => 'approved',
    ]);
    print_r($result);

    // Example 2: Send a conversion with click attribution
    echo "\n2. Sending a conversion with click attribution...\n";
    $result = $tracker->sendPostback([
        'offer_id' => 'offer_123',
        'transaction_id' => 'txn_' . time() . '_2',
        'click_id' => 'click_abc123',
        'amount' => 49.99,
        'currency' => 'EUR',
        'status' => 'approved',
        'custom_params' => [
            'product_id' => 'prod_456',
            'category' => 'electronics',
        ],
    ]);
    print_r($result);

    // Example 3: Track a server-side click
    echo "\n3. Tracking a server-side click...\n";
    $result = $tracker->trackClick([
        'offer_id' => 'offer_123',
        'tracking_code' => 'campaign_summer_2024',
        'sub_id_1' => 'source_google',
        'ip' => '192.168.1.1',
        'user_agent' => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    ]);
    print_r($result);

    // Example 4: Batch send conversions
    echo "\n4. Batch sending conversions...\n";
    $batchResults = $tracker->sendBatchPostbacks([
        ['offer_id' => 'offer_123', 'transaction_id' => 'batch_1_' . time(), 'amount' => 10.00],
        ['offer_id' => 'offer_123', 'transaction_id' => 'batch_2_' . time(), 'amount' => 20.00],
        ['offer_id' => 'offer_123', 'transaction_id' => 'batch_3_' . time(), 'amount' => 30.00],
    ]);
    print_r($batchResults);
}

