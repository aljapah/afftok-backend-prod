import Foundation
import CommonCrypto

/// Click tracking implementation for AffTok SDK
public class ClickTracker {
    
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
    
    /// Track a click event
    public func trackClick(params: ClickParams) async -> AfftokResponse {
        let payload = buildClickPayload(params: params)
        
        do {
            let response = try await sendRequest(endpoint: AfftokConfig.clickEndpoint, payload: payload)
            if response.success {
                log("Click tracked successfully: \(params.offerId)")
            } else {
                queue.enqueue(type: "click", payload: payload)
                log("Click queued for retry: \(params.offerId)")
            }
            return response
        } catch {
            queue.enqueue(type: "click", payload: payload)
            log("Click queued (offline): \(params.offerId), error: \(error.localizedDescription)")
            return AfftokResponse(
                success: false,
                message: "Click queued for offline retry",
                error: error.localizedDescription
            )
        }
    }
    
    /// Track click with signed link
    public func trackSignedClick(signedLink: String, params: ClickParams) async -> AfftokResponse {
        var payload = buildClickPayload(params: params)
        payload["signed_link"] = signedLink
        payload["link_validated"] = "true"
        
        do {
            let response = try await sendRequest(endpoint: AfftokConfig.clickEndpoint, payload: payload)
            if !response.success {
                queue.enqueue(type: "click", payload: payload)
            }
            return response
        } catch {
            queue.enqueue(type: "click", payload: payload)
            return AfftokResponse(
                success: false,
                message: "Signed click queued for retry",
                error: error.localizedDescription
            )
        }
    }
    
    /// Build click payload
    private func buildClickPayload(params: ClickParams) -> [String: Any] {
        let timestamp = Int64(Date().timeIntervalSince1970 * 1000)
        let nonce = generateNonce()
        let deviceInfo = DeviceInfo.getDeviceInfo()
        
        var payload: [String: Any] = [
            "api_key": options.apiKey,
            "advertiser_id": options.advertiserId,
            "offer_id": params.offerId,
            "timestamp": timestamp,
            "nonce": nonce,
            "device_info": deviceInfo
        ]
        
        if let userId = options.userId {
            payload["user_id"] = userId
        }
        if let trackingCode = params.trackingCode {
            payload["tracking_code"] = trackingCode
        }
        if let subId1 = params.subId1 {
            payload["sub_id_1"] = subId1
        }
        if let subId2 = params.subId2 {
            payload["sub_id_2"] = subId2
        }
        if let subId3 = params.subId3 {
            payload["sub_id_3"] = subId3
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
            if endpoint == AfftokConfig.clickEndpoint {
                return try await sendFallbackRequest(endpoint: AfftokConfig.fallbackClickEndpoint, payload: payload)
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
            return AfftokResponse(success: true, message: "Click tracked via fallback")
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
            print("[AffTok Click] \(message)")
        }
    }
}

