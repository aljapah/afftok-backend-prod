import Foundation

/// AffTok SDK Configuration
public struct AfftokConfig {
    // API Endpoints
    public static let defaultBaseURL = "https://api.afftok.com"
    public static let clickEndpoint = "/api/sdk/click"
    public static let conversionEndpoint = "/api/sdk/conversion"
    
    // Fallback Endpoints
    public static let fallbackClickEndpoint = "/api/c"
    public static let fallbackConversionEndpoint = "/api/convert"
    
    // Queue Settings
    public static let maxQueueSize = 1000
    public static let maxRetryAttempts = 5
    public static let initialRetryDelayMs: TimeInterval = 1.0
    public static let maxRetryDelayMs: TimeInterval = 300.0 // 5 minutes
    public static let flushInterval: TimeInterval = 30.0 // 30 seconds
    
    // Rate Limiting
    public static let maxRequestsPerMinute = 60
    public static let rateLimitWindow: TimeInterval = 60.0
    
    // Timeouts
    public static let connectionTimeout: TimeInterval = 10.0
    public static let readTimeout: TimeInterval = 30.0
    
    // Storage Keys
    public static let queueKey = "afftok_offline_queue"
    public static let deviceIdKey = "afftok_device_id"
    public static let lastFlushKey = "afftok_last_flush"
    
    // SDK Version
    public static let sdkVersion = "1.0.0"
    public static let sdkPlatform = "ios"
}

/// SDK initialization options
public struct AfftokOptions {
    public let apiKey: String
    public let advertiserId: String
    public var userId: String?
    public var baseUrl: String
    public var debug: Bool
    public var autoFlush: Bool
    public var flushInterval: TimeInterval
    
    public init(
        apiKey: String,
        advertiserId: String,
        userId: String? = nil,
        baseUrl: String = AfftokConfig.defaultBaseURL,
        debug: Bool = false,
        autoFlush: Bool = true,
        flushInterval: TimeInterval = AfftokConfig.flushInterval
    ) {
        self.apiKey = apiKey
        self.advertiserId = advertiserId
        self.userId = userId
        self.baseUrl = baseUrl
        self.debug = debug
        self.autoFlush = autoFlush
        self.flushInterval = flushInterval
    }
}

/// Click tracking parameters
public struct ClickParams {
    public let offerId: String
    public var trackingCode: String?
    public var subId1: String?
    public var subId2: String?
    public var subId3: String?
    public var customParams: [String: String]?
    
    public init(
        offerId: String,
        trackingCode: String? = nil,
        subId1: String? = nil,
        subId2: String? = nil,
        subId3: String? = nil,
        customParams: [String: String]? = nil
    ) {
        self.offerId = offerId
        self.trackingCode = trackingCode
        self.subId1 = subId1
        self.subId2 = subId2
        self.subId3 = subId3
        self.customParams = customParams
    }
}

/// Conversion tracking parameters
public struct ConversionParams {
    public let offerId: String
    public let transactionId: String
    public var clickId: String?
    public var amount: Double?
    public var currency: String
    public var status: String
    public var customParams: [String: String]?
    
    public init(
        offerId: String,
        transactionId: String,
        clickId: String? = nil,
        amount: Double? = nil,
        currency: String = "USD",
        status: String = "pending",
        customParams: [String: String]? = nil
    ) {
        self.offerId = offerId
        self.transactionId = transactionId
        self.clickId = clickId
        self.amount = amount
        self.currency = currency
        self.status = status
        self.customParams = customParams
    }
}

/// SDK Response
public struct AfftokResponse {
    public let success: Bool
    public var message: String?
    public var data: [String: Any]?
    public var error: String?
    
    public init(success: Bool, message: String? = nil, data: [String: Any]? = nil, error: String? = nil) {
        self.success = success
        self.message = message
        self.data = data
        self.error = error
    }
}

/// Queue Item
public struct QueueItem: Codable {
    public let id: String
    public let type: String
    public let payload: [String: String]
    public let timestamp: TimeInterval
    public var retryCount: Int
    public var nextRetryTime: TimeInterval
    
    public init(id: String, type: String, payload: [String: String], timestamp: TimeInterval, retryCount: Int = 0, nextRetryTime: TimeInterval = 0) {
        self.id = id
        self.type = type
        self.payload = payload
        self.timestamp = timestamp
        self.retryCount = retryCount
        self.nextRetryTime = nextRetryTime
    }
}

