import Foundation

/// AffTok SDK - Main Entry Point
///
/// The AffTok SDK provides click and conversion tracking for iOS applications.
/// It supports offline queuing, automatic retry with exponential backoff,
/// HMAC-SHA256 signature verification, and zero-drop tracking.
///
/// Usage:
/// ```swift
/// // Initialize
/// Afftok.shared.initialize(options: AfftokOptions(
///     apiKey: "your_api_key",
///     advertiserId: "your_advertiser_id",
///     debug: true
/// ))
///
/// // Track click
/// await Afftok.shared.trackClick(params: ClickParams(offerId: "offer123"))
///
/// // Track conversion
/// await Afftok.shared.trackConversion(params: ConversionParams(
///     offerId: "offer123",
///     transactionId: "txn_abc123",
///     amount: 29.99
/// ))
/// ```
public class Afftok {
    
    /// Shared singleton instance
    public static let shared = Afftok()
    
    private var isInitialized = false
    private var options: AfftokOptions?
    private var queue: OfflineQueue?
    private var clickTracker: ClickTracker?
    private var conversionTracker: ConversionTracker?
    
    private init() {}
    
    /// Initialize the AffTok SDK
    ///
    /// - Parameter options: SDK configuration options
    public func initialize(options: AfftokOptions) {
        guard !isInitialized else {
            log("SDK already initialized")
            return
        }
        
        self.options = options
        
        // Initialize components
        queue = OfflineQueue(debug: options.debug)
        clickTracker = ClickTracker(options: options, queue: queue!)
        conversionTracker = ConversionTracker(options: options, queue: queue!)
        
        // Start auto-flush if enabled
        if options.autoFlush {
            startAutoFlush()
        }
        
        isInitialized = true
        log("SDK initialized successfully")
        log("Device ID: \(DeviceInfo.getDeviceId())")
        log("Pending queue items: \(queue?.size() ?? 0)")
    }
    
    /// Track a click event
    ///
    /// - Parameter params: Click parameters
    /// - Returns: AfftokResponse with tracking result
    public func trackClick(params: ClickParams) async -> AfftokResponse {
        guard let tracker = clickTracker else {
            return AfftokResponse(success: false, error: "SDK not initialized")
        }
        return await tracker.trackClick(params: params)
    }
    
    /// Track a click event with completion handler
    public func trackClick(params: ClickParams, completion: @escaping (AfftokResponse) -> Void) {
        Task {
            let response = await trackClick(params: params)
            DispatchQueue.main.async {
                completion(response)
            }
        }
    }
    
    /// Track a click with a signed tracking link
    ///
    /// - Parameters:
    ///   - signedLink: The signed tracking link from AffTok
    ///   - params: Additional click parameters
    /// - Returns: AfftokResponse with tracking result
    public func trackSignedClick(signedLink: String, params: ClickParams) async -> AfftokResponse {
        guard let tracker = clickTracker else {
            return AfftokResponse(success: false, error: "SDK not initialized")
        }
        return await tracker.trackSignedClick(signedLink: signedLink, params: params)
    }
    
    /// Track a conversion event
    ///
    /// - Parameter params: Conversion parameters
    /// - Returns: AfftokResponse with tracking result
    public func trackConversion(params: ConversionParams) async -> AfftokResponse {
        guard let tracker = conversionTracker else {
            return AfftokResponse(success: false, error: "SDK not initialized")
        }
        return await tracker.trackConversion(params: params)
    }
    
    /// Track a conversion event with completion handler
    public func trackConversion(params: ConversionParams, completion: @escaping (AfftokResponse) -> Void) {
        Task {
            let response = await trackConversion(params: params)
            DispatchQueue.main.async {
                completion(response)
            }
        }
    }
    
    /// Track conversion with additional metadata
    public func trackConversionWithMeta(params: ConversionParams, metadata: [String: Any]) async -> AfftokResponse {
        guard let tracker = conversionTracker else {
            return AfftokResponse(success: false, error: "SDK not initialized")
        }
        return await tracker.trackConversionWithMeta(params: params, metadata: metadata)
    }
    
    /// Manually enqueue an event for later processing
    ///
    /// - Parameters:
    ///   - type: Event type ("click" or "conversion")
    ///   - payload: Event payload
    /// - Returns: Queue item ID
    @discardableResult
    public func enqueue(type: String, payload: [String: Any]) -> String {
        guard let q = queue else { return "" }
        return q.enqueue(type: type, payload: payload)
    }
    
    /// Manually flush the offline queue
    public func flush() async {
        guard let q = queue else { return }
        await q.flush { [weak self] item in
            await self?.processQueueItem(item: item) ?? false
        }
    }
    
    /// Flush queue with completion handler
    public func flush(completion: @escaping () -> Void) {
        Task {
            await flush()
            DispatchQueue.main.async {
                completion()
            }
        }
    }
    
    /// Get the device fingerprint
    public func getFingerprint() -> String {
        return DeviceInfo.getFingerprint()
    }
    
    /// Get the device ID
    public func getDeviceId() -> String {
        return DeviceInfo.getDeviceId()
    }
    
    /// Get device info
    public func getDeviceInfo() -> [String: String] {
        return DeviceInfo.getDeviceInfo()
    }
    
    /// Get number of pending queue items
    public func getPendingCount() -> Int {
        return queue?.size() ?? 0
    }
    
    /// Check if SDK is initialized
    public func isReady() -> Bool {
        return isInitialized
    }
    
    /// Get SDK version
    public func getVersion() -> String {
        return AfftokConfig.sdkVersion
    }
    
    /// Clear all pending queue items
    public func clearQueue() {
        queue?.clear()
    }
    
    /// Shutdown the SDK
    public func shutdown() {
        queue?.stopAutoFlush()
        isInitialized = false
        log("SDK shutdown")
    }
    
    // MARK: - Private Methods
    
    private func startAutoFlush() {
        guard let q = queue, let opts = options else { return }
        q.startAutoFlush(interval: opts.flushInterval) { [weak self] item in
            await self?.processQueueItem(item: item) ?? false
        }
    }
    
    private func processQueueItem(item: QueueItem) async -> Bool {
        switch item.type {
        case "click":
            let params = ClickParams(
                offerId: item.payload["offer_id"] ?? "",
                trackingCode: item.payload["tracking_code"],
                subId1: item.payload["sub_id_1"],
                subId2: item.payload["sub_id_2"],
                subId3: item.payload["sub_id_3"]
            )
            let response = await trackClick(params: params)
            return response.success
            
        case "conversion":
            let params = ConversionParams(
                offerId: item.payload["offer_id"] ?? "",
                transactionId: item.payload["transaction_id"] ?? "",
                clickId: item.payload["click_id"],
                amount: Double(item.payload["amount"] ?? ""),
                currency: item.payload["currency"] ?? "USD",
                status: item.payload["status"] ?? "pending"
            )
            let response = await trackConversion(params: params)
            return response.success
            
        default:
            log("Unknown event type: \(item.type)")
            return false
        }
    }
    
    private func log(_ message: String) {
        if options?.debug == true {
            print("[AffTok SDK] \(message)")
        }
    }
}

