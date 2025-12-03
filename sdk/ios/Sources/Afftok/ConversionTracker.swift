import Foundation
import CommonCrypto

/// Conversion tracking implementation for AffTok SDK
public class ConversionTracker {
    
    private let options: AfftokOptions
    private let queue: OfflineQueue
    private let session: URLSession
    
    public init(options: AfftokOptions, queue: OfflineQueue) {
        self.options = options
        self.queue = queue
        
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = AfftokConfig.connectionTimeout
        config.timeoutIntervalForResource = AfftokConfig.readTimeout
        self.session = URLSession(configuration: config)
    }
    
    /// Track a conversion event
    public func trackConversion(params: ConversionParams) async -> AfftokResponse {
        let payload = buildConversionPayload(params: params)
        
        do {
            let response = try await sendRequest(endpoint: AfftokConfig.conversionEndpoint, payload: payload)
            if response.success {
                log("Conversion tracked successfully: \(params.transactionId)")
            } else {
                queue.enqueue(type: "conversion", payload: payload)
                log("Conversion queued for retry: \(params.transactionId)")
            }
            return response
        } catch {
            queue.enqueue(type: "conversion", payload: payload)
            log("Conversion queued (offline): \(params.transactionId), error: \(error.localizedDescription)")
            return AfftokResponse(
                success: false,
                message: "Conversion queued for offline retry",
                error: error.localizedDescription
            )
        }
    }
    
    /// Track conversion with additional metadata
    public func trackConversionWithMeta(params: ConversionParams, metadata: [String: Any]) async -> AfftokResponse {
        var payload = buildConversionPayload(params: params)
        payload["metadata"] = metadata
        
        do {
            let response = try await sendRequest(endpoint: AfftokConfig.conversionEndpoint, payload: payload)
            if !response.success {
                queue.enqueue(type: "conversion", payload: payload)
            }
            return response
        } catch {
            queue.enqueue(type: "conversion", payload: payload)
            return AfftokResponse(
                success: false,
                message: "Conversion with metadata queued for retry",
                error: error.localizedDescription
            )
        }
    }
    
    /// Build conversion payload
    private func buildConversionPayload(params: ConversionParams) -> [String: Any] {
        let timestamp = Int64(Date().timeIntervalSince1970 * 1000)
        let nonce = generateNonce()
        let deviceInfo = DeviceInfo.getDeviceInfo()
        
        var payload: [String: Any] = [
            "api_key": options.apiKey,
            "advertiser_id": options.advertiserId,
            "offer_id": params.offerId,
            "transaction_id": params.transactionId,
            "status": params.status,
            "currency": params.currency,
            "timestamp": timestamp,
            "nonce": nonce,
            "device_info": deviceInfo
        ]
        
        if let userId = options.userId {
            payload["user_id"] = userId
        }
        if let clickId = params.clickId {
            payload["click_id"] = clickId
        }
        if let amount = params.amount {
            payload["amount"] = amount
        }
        if let customParams = params.customParams {
            payload["custom_params"] = customParams
        }
        
        // Add signature
        payload["signature"] = generateSignature(timestamp: timestamp, nonce: nonce)
        
        return payload
    }
    
    /// Send HTTP request
    private func sendRequest(endpoint: String, payload: [String: Any]) async throws -> AfftokResponse {
        guard let url = URL(string: "\(options.baseUrl)\(endpoint)") else {
            throw URLError(.badURL)
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue(options.apiKey, forHTTPHeaderField: "X-API-Key")
        request.setValue(AfftokConfig.sdkVersion, forHTTPHeaderField: "X-SDK-Version")
        request.setValue(AfftokConfig.sdkPlatform, forHTTPHeaderField: "X-SDK-Platform")
        request.httpBody = try JSONSerialization.data(withJSONObject: payload)
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw URLError(.badServerResponse)
        }
        
        if httpResponse.statusCode >= 200 && httpResponse.statusCode < 300 {
            let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any]
            return AfftokResponse(
                success: true,
                message: json?["message"] as? String,
                data: json
            )
        } else {
            // Try fallback endpoint
            if endpoint == AfftokConfig.conversionEndpoint {
                return try await sendFallbackRequest(endpoint: AfftokConfig.fallbackConversionEndpoint, payload: payload)
            }
            return AfftokResponse(
                success: false,
                error: "HTTP \(httpResponse.statusCode)"
            )
        }
    }
    
    /// Send request to fallback endpoint
    private func sendFallbackRequest(endpoint: String, payload: [String: Any]) async throws -> AfftokResponse {
        guard let url = URL(string: "\(options.baseUrl)\(endpoint)") else {
            throw URLError(.badURL)
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue(options.apiKey, forHTTPHeaderField: "X-API-Key")
        request.httpBody = try JSONSerialization.data(withJSONObject: payload)
        
        let (_, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw URLError(.badServerResponse)
        }
        
        if httpResponse.statusCode >= 200 && httpResponse.statusCode < 300 {
            return AfftokResponse(success: true, message: "Conversion tracked via fallback")
        } else {
            return AfftokResponse(success: false, error: "Fallback failed: \(httpResponse.statusCode)")
        }
    }
    
    /// Generate HMAC-SHA256 signature
    private func generateSignature(timestamp: Int64, nonce: String) -> String {
        let dataToSign = "\(options.apiKey)|\(options.advertiserId)|\(timestamp)|\(nonce)"
        
        guard let keyData = options.apiKey.data(using: .utf8),
              let messageData = dataToSign.data(using: .utf8) else {
            return ""
        }
        
        var hmac = [UInt8](repeating: 0, count: Int(CC_SHA256_DIGEST_LENGTH))
        
        keyData.withUnsafeBytes { keyBytes in
            messageData.withUnsafeBytes { messageBytes in
                CCHmac(CCHmacAlgorithm(kCCHmacAlgSHA256),
                       keyBytes.baseAddress, keyData.count,
                       messageBytes.baseAddress, messageData.count,
                       &hmac)
            }
        }
        
        return hmac.map { String(format: "%02x", $0) }.joined()
    }
    
    /// Generate random nonce
    private func generateNonce() -> String {
        let chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
        return String((0..<32).map { _ in chars.randomElement()! })
    }
    
    private func log(_ message: String) {
        if options.debug {
            print("[AffTok Conversion] \(message)")
        }
    }
}

