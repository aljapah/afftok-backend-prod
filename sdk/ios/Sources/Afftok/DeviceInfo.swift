import Foundation
import UIKit
import CommonCrypto

/// Device information provider for AffTok SDK
public class DeviceInfo {
    
    private static var cachedFingerprint: String?
    private static var cachedDeviceId: String?
    
    /// Get or generate device ID
    public static func getDeviceId() -> String {
        if let cached = cachedDeviceId {
            return cached
        }
        
        if let storedId = UserDefaults.standard.string(forKey: AfftokConfig.deviceIdKey) {
            cachedDeviceId = storedId
            return storedId
        }
        
        let newId = generateDeviceId()
        UserDefaults.standard.set(newId, forKey: AfftokConfig.deviceIdKey)
        cachedDeviceId = newId
        return newId
    }
    
    /// Generate unique device fingerprint
    public static func getFingerprint() -> String {
        if let cached = cachedFingerprint {
            return cached
        }
        
        let components = [
            getDeviceId(),
            UIDevice.current.model,
            UIDevice.current.systemName,
            UIDevice.current.systemVersion,
            getScreenResolution(),
            getLanguage(),
            getTimezone(),
            getCarrier()
        ]
        
        let combined = components.joined(separator: "|")
        cachedFingerprint = sha256(combined)
        return cachedFingerprint!
    }
    
    /// Get device info dictionary
    public static func getDeviceInfo() -> [String: String] {
        return [
            "device_id": getDeviceId(),
            "fingerprint": getFingerprint(),
            "platform": AfftokConfig.sdkPlatform,
            "sdk_version": AfftokConfig.sdkVersion,
            "os_version": UIDevice.current.systemVersion,
            "model": UIDevice.current.model,
            "device_name": UIDevice.current.name,
            "screen": getScreenResolution(),
            "language": getLanguage(),
            "timezone": getTimezone(),
            "carrier": getCarrier()
        ]
    }
    
    private static func generateDeviceId() -> String {
        if let identifierForVendor = UIDevice.current.identifierForVendor {
            return sha256(identifierForVendor.uuidString)
        }
        return UUID().uuidString
    }
    
    private static func getScreenResolution() -> String {
        let screen = UIScreen.main
        let bounds = screen.bounds
        let scale = screen.scale
        return "\(Int(bounds.width * scale))x\(Int(bounds.height * scale))"
    }
    
    private static func getLanguage() -> String {
        return Locale.current.languageCode ?? "en"
    }
    
    private static func getTimezone() -> String {
        return TimeZone.current.identifier
    }
    
    private static func getCarrier() -> String {
        // Note: CTCarrier is deprecated in iOS 16+
        return "unknown"
    }
    
    private static func sha256(_ input: String) -> String {
        guard let data = input.data(using: .utf8) else {
            return String(input.hashValue)
        }
        
        var hash = [UInt8](repeating: 0, count: Int(CC_SHA256_DIGEST_LENGTH))
        data.withUnsafeBytes {
            _ = CC_SHA256($0.baseAddress, CC_LONG(data.count), &hash)
        }
        
        return hash.map { String(format: "%02x", $0) }.joined()
    }
}

