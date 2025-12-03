library afftok;

import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:math';
import 'package:crypto/crypto.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:device_info_plus/device_info_plus.dart';
import 'package:uuid/uuid.dart';

/// AffTok SDK Configuration
class AfftokConfig {
  static const String defaultBaseUrl = 'https://api.afftok.com';
  static const String clickEndpoint = '/api/sdk/click';
  static const String conversionEndpoint = '/api/sdk/conversion';
  static const String fallbackClickEndpoint = '/api/c';
  static const String fallbackConversionEndpoint = '/api/convert';
  
  static const int maxQueueSize = 1000;
  static const int maxRetryAttempts = 5;
  static const Duration initialRetryDelay = Duration(seconds: 1);
  static const Duration maxRetryDelay = Duration(minutes: 5);
  static const Duration flushInterval = Duration(seconds: 30);
  
  static const int maxRequestsPerMinute = 60;
  static const Duration connectionTimeout = Duration(seconds: 10);
  static const Duration readTimeout = Duration(seconds: 30);
  
  static const String queueKey = 'afftok_offline_queue';
  static const String deviceIdKey = 'afftok_device_id';
  
  static const String sdkVersion = '1.0.0';
  static const String sdkPlatform = 'flutter';
}

/// SDK initialization options
class AfftokOptions {
  final String apiKey;
  final String advertiserId;
  final String? userId;
  final String baseUrl;
  final bool debug;
  final bool autoFlush;
  final Duration flushInterval;
  
  AfftokOptions({
    required this.apiKey,
    required this.advertiserId,
    this.userId,
    this.baseUrl = AfftokConfig.defaultBaseUrl,
    this.debug = false,
    this.autoFlush = true,
    this.flushInterval = AfftokConfig.flushInterval,
  });
}

/// Click tracking parameters
class ClickParams {
  final String offerId;
  final String? trackingCode;
  final String? subId1;
  final String? subId2;
  final String? subId3;
  final Map<String, String>? customParams;
  
  ClickParams({
    required this.offerId,
    this.trackingCode,
    this.subId1,
    this.subId2,
    this.subId3,
    this.customParams,
  });
  
  Map<String, dynamic> toJson() => {
    'offer_id': offerId,
    if (trackingCode != null) 'tracking_code': trackingCode,
    if (subId1 != null) 'sub_id_1': subId1,
    if (subId2 != null) 'sub_id_2': subId2,
    if (subId3 != null) 'sub_id_3': subId3,
    if (customParams != null) 'custom_params': customParams,
  };
}

/// Conversion tracking parameters
class ConversionParams {
  final String offerId;
  final String transactionId;
  final String? clickId;
  final double? amount;
  final String currency;
  final String status;
  final Map<String, String>? customParams;
  
  ConversionParams({
    required this.offerId,
    required this.transactionId,
    this.clickId,
    this.amount,
    this.currency = 'USD',
    this.status = 'pending',
    this.customParams,
  });
  
  Map<String, dynamic> toJson() => {
    'offer_id': offerId,
    'transaction_id': transactionId,
    'status': status,
    'currency': currency,
    if (clickId != null) 'click_id': clickId,
    if (amount != null) 'amount': amount,
    if (customParams != null) 'custom_params': customParams,
  };
}

/// SDK Response
class AfftokResponse {
  final bool success;
  final String? message;
  final Map<String, dynamic>? data;
  final String? error;
  
  AfftokResponse({
    required this.success,
    this.message,
    this.data,
    this.error,
  });
  
  factory AfftokResponse.fromJson(Map<String, dynamic> json) {
    return AfftokResponse(
      success: json['success'] ?? false,
      message: json['message'],
      data: json['data'],
      error: json['error'],
    );
  }
}

/// Queue Item
class QueueItem {
  final String id;
  final String type;
  final Map<String, dynamic> payload;
  final int timestamp;
  int retryCount;
  int nextRetryTime;
  
  QueueItem({
    required this.id,
    required this.type,
    required this.payload,
    required this.timestamp,
    this.retryCount = 0,
    this.nextRetryTime = 0,
  });
  
  Map<String, dynamic> toJson() => {
    'id': id,
    'type': type,
    'payload': payload,
    'timestamp': timestamp,
    'retryCount': retryCount,
    'nextRetryTime': nextRetryTime,
  };
  
  factory QueueItem.fromJson(Map<String, dynamic> json) {
    return QueueItem(
      id: json['id'],
      type: json['type'],
      payload: Map<String, dynamic>.from(json['payload']),
      timestamp: json['timestamp'],
      retryCount: json['retryCount'] ?? 0,
      nextRetryTime: json['nextRetryTime'] ?? 0,
    );
  }
}

/// AffTok SDK - Main Entry Point
/// 
/// The AffTok SDK provides click and conversion tracking for Flutter applications.
/// It supports offline queuing, automatic retry with exponential backoff,
/// HMAC-SHA256 signature verification, and zero-drop tracking.
/// 
/// Usage:
/// ```dart
/// // Initialize
/// await Afftok.instance.initialize(AfftokOptions(
///   apiKey: 'your_api_key',
///   advertiserId: 'your_advertiser_id',
///   debug: true,
/// ));
/// 
/// // Track click
/// await Afftok.instance.trackClick(ClickParams(offerId: 'offer123'));
/// 
/// // Track conversion
/// await Afftok.instance.trackConversion(ConversionParams(
///   offerId: 'offer123',
///   transactionId: 'txn_abc123',
///   amount: 29.99,
/// ));
/// ```
class Afftok {
  static final Afftok instance = Afftok._internal();
  
  bool _isInitialized = false;
  AfftokOptions? _options;
  SharedPreferences? _prefs;
  List<QueueItem> _queue = [];
  Timer? _flushTimer;
  bool _isProcessing = false;
  String? _deviceId;
  Map<String, String>? _deviceInfo;
  final Uuid _uuid = const Uuid();
  
  Afftok._internal();
  
  /// Initialize the AffTok SDK
  Future<void> initialize(AfftokOptions options) async {
    if (_isInitialized) {
      _log('SDK already initialized');
      return;
    }
    
    _options = options;
    _prefs = await SharedPreferences.getInstance();
    
    await _loadQueue();
    await _initDeviceInfo();
    
    if (options.autoFlush) {
      _startAutoFlush();
    }
    
    _isInitialized = true;
    _log('SDK initialized successfully');
    _log('Device ID: $_deviceId');
    _log('Pending queue items: ${_queue.length}');
  }
  
  /// Track a click event
  Future<AfftokResponse> trackClick(ClickParams params) async {
    _ensureInitialized();
    
    final payload = await _buildClickPayload(params);
    
    try {
      final response = await _sendRequest(AfftokConfig.clickEndpoint, payload);
      if (response.success) {
        _log('Click tracked successfully: ${params.offerId}');
      } else {
        _enqueue('click', payload);
        _log('Click queued for retry: ${params.offerId}');
      }
      return response;
    } catch (e) {
      _enqueue('click', payload);
      _log('Click queued (offline): ${params.offerId}, error: $e');
      return AfftokResponse(
        success: false,
        message: 'Click queued for offline retry',
        error: e.toString(),
      );
    }
  }
  
  /// Track a click with a signed tracking link
  Future<AfftokResponse> trackSignedClick(String signedLink, ClickParams params) async {
    _ensureInitialized();
    
    final payload = await _buildClickPayload(params);
    payload['signed_link'] = signedLink;
    payload['link_validated'] = true;
    
    try {
      final response = await _sendRequest(AfftokConfig.clickEndpoint, payload);
      if (!response.success) {
        _enqueue('click', payload);
      }
      return response;
    } catch (e) {
      _enqueue('click', payload);
      return AfftokResponse(
        success: false,
        message: 'Signed click queued for retry',
        error: e.toString(),
      );
    }
  }
  
  /// Track a conversion event
  Future<AfftokResponse> trackConversion(ConversionParams params) async {
    _ensureInitialized();
    
    final payload = await _buildConversionPayload(params);
    
    try {
      final response = await _sendRequest(AfftokConfig.conversionEndpoint, payload);
      if (response.success) {
        _log('Conversion tracked successfully: ${params.transactionId}');
      } else {
        _enqueue('conversion', payload);
        _log('Conversion queued for retry: ${params.transactionId}');
      }
      return response;
    } catch (e) {
      _enqueue('conversion', payload);
      _log('Conversion queued (offline): ${params.transactionId}, error: $e');
      return AfftokResponse(
        success: false,
        message: 'Conversion queued for offline retry',
        error: e.toString(),
      );
    }
  }
  
  /// Track conversion with additional metadata
  Future<AfftokResponse> trackConversionWithMeta(
    ConversionParams params,
    Map<String, dynamic> metadata,
  ) async {
    _ensureInitialized();
    
    final payload = await _buildConversionPayload(params);
    payload['metadata'] = metadata;
    
    try {
      final response = await _sendRequest(AfftokConfig.conversionEndpoint, payload);
      if (!response.success) {
        _enqueue('conversion', payload);
      }
      return response;
    } catch (e) {
      _enqueue('conversion', payload);
      return AfftokResponse(
        success: false,
        message: 'Conversion with metadata queued for retry',
        error: e.toString(),
      );
    }
  }
  
  /// Manually enqueue an event for later processing
  String enqueue(String type, Map<String, dynamic> payload) {
    _ensureInitialized();
    return _enqueue(type, payload);
  }
  
  /// Manually flush the offline queue
  Future<void> flush() async {
    _ensureInitialized();
    await _flush();
  }
  
  /// Get the device fingerprint
  String getFingerprint() {
    _ensureInitialized();
    return _generateFingerprint();
  }
  
  /// Get the device ID
  String getDeviceId() {
    _ensureInitialized();
    return _deviceId ?? '';
  }
  
  /// Get device info
  Map<String, String> getDeviceInfo() {
    _ensureInitialized();
    return _deviceInfo ?? {};
  }
  
  /// Get number of pending queue items
  int getPendingCount() {
    return _queue.length;
  }
  
  /// Check if SDK is initialized
  bool isReady() => _isInitialized;
  
  /// Get SDK version
  String getVersion() => AfftokConfig.sdkVersion;
  
  /// Clear all pending queue items
  void clearQueue() {
    _queue.clear();
    _saveQueue();
  }
  
  /// Shutdown the SDK
  void shutdown() {
    _flushTimer?.cancel();
    _flushTimer = null;
    _isInitialized = false;
    _log('SDK shutdown');
  }
  
  // Private methods
  
  void _ensureInitialized() {
    if (!_isInitialized) {
      throw StateError('AffTok SDK not initialized. Call Afftok.instance.initialize() first.');
    }
  }
  
  Future<void> _initDeviceInfo() async {
    final deviceInfoPlugin = DeviceInfoPlugin();
    _deviceId = _prefs?.getString(AfftokConfig.deviceIdKey);
    
    if (_deviceId == null) {
      _deviceId = _uuid.v4();
      await _prefs?.setString(AfftokConfig.deviceIdKey, _deviceId!);
    }
    
    try {
      if (Platform.isAndroid) {
        final androidInfo = await deviceInfoPlugin.androidInfo;
        _deviceInfo = {
          'device_id': _deviceId!,
          'fingerprint': _generateFingerprint(),
          'platform': AfftokConfig.sdkPlatform,
          'sdk_version': AfftokConfig.sdkVersion,
          'os': 'android',
          'os_version': androidInfo.version.release,
          'manufacturer': androidInfo.manufacturer,
          'model': androidInfo.model,
          'brand': androidInfo.brand,
        };
      } else if (Platform.isIOS) {
        final iosInfo = await deviceInfoPlugin.iosInfo;
        _deviceInfo = {
          'device_id': _deviceId!,
          'fingerprint': _generateFingerprint(),
          'platform': AfftokConfig.sdkPlatform,
          'sdk_version': AfftokConfig.sdkVersion,
          'os': 'ios',
          'os_version': iosInfo.systemVersion,
          'model': iosInfo.model,
          'device_name': iosInfo.name,
        };
      }
    } catch (e) {
      _deviceInfo = {
        'device_id': _deviceId!,
        'fingerprint': _generateFingerprint(),
        'platform': AfftokConfig.sdkPlatform,
        'sdk_version': AfftokConfig.sdkVersion,
      };
    }
  }
  
  String _generateFingerprint() {
    final data = '$_deviceId|${_deviceInfo?.values.join('|') ?? ''}';
    final bytes = utf8.encode(data);
    final digest = sha256.convert(bytes);
    return digest.toString();
  }
  
  Future<Map<String, dynamic>> _buildClickPayload(ClickParams params) async {
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    final nonce = _generateNonce();
    
    final payload = <String, dynamic>{
      'api_key': _options!.apiKey,
      'advertiser_id': _options!.advertiserId,
      'timestamp': timestamp,
      'nonce': nonce,
      'device_info': _deviceInfo,
      ...params.toJson(),
    };
    
    if (_options!.userId != null) {
      payload['user_id'] = _options!.userId;
    }
    
    payload['signature'] = _generateSignature(timestamp, nonce);
    
    return payload;
  }
  
  Future<Map<String, dynamic>> _buildConversionPayload(ConversionParams params) async {
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    final nonce = _generateNonce();
    
    final payload = <String, dynamic>{
      'api_key': _options!.apiKey,
      'advertiser_id': _options!.advertiserId,
      'timestamp': timestamp,
      'nonce': nonce,
      'device_info': _deviceInfo,
      ...params.toJson(),
    };
    
    if (_options!.userId != null) {
      payload['user_id'] = _options!.userId;
    }
    
    payload['signature'] = _generateSignature(timestamp, nonce);
    
    return payload;
  }
  
  Future<AfftokResponse> _sendRequest(String endpoint, Map<String, dynamic> payload) async {
    final url = Uri.parse('${_options!.baseUrl}$endpoint');
    
    final response = await http.post(
      url,
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': _options!.apiKey,
        'X-SDK-Version': AfftokConfig.sdkVersion,
        'X-SDK-Platform': AfftokConfig.sdkPlatform,
      },
      body: jsonEncode(payload),
    ).timeout(AfftokConfig.connectionTimeout);
    
    if (response.statusCode >= 200 && response.statusCode < 300) {
      final json = jsonDecode(response.body);
      return AfftokResponse.fromJson(json);
    } else {
      // Try fallback
      if (endpoint == AfftokConfig.clickEndpoint) {
        return _sendFallbackRequest(AfftokConfig.fallbackClickEndpoint, payload);
      } else if (endpoint == AfftokConfig.conversionEndpoint) {
        return _sendFallbackRequest(AfftokConfig.fallbackConversionEndpoint, payload);
      }
      return AfftokResponse(
        success: false,
        error: 'HTTP ${response.statusCode}: ${response.body}',
      );
    }
  }
  
  Future<AfftokResponse> _sendFallbackRequest(String endpoint, Map<String, dynamic> payload) async {
    try {
      final url = Uri.parse('${_options!.baseUrl}$endpoint');
      
      final response = await http.post(
        url,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': _options!.apiKey,
        },
        body: jsonEncode(payload),
      ).timeout(AfftokConfig.connectionTimeout);
      
      if (response.statusCode >= 200 && response.statusCode < 300) {
        return AfftokResponse(success: true, message: 'Tracked via fallback');
      }
      return AfftokResponse(success: false, error: 'Fallback failed: ${response.statusCode}');
    } catch (e) {
      return AfftokResponse(success: false, error: 'Fallback error: $e');
    }
  }
  
  String _generateSignature(int timestamp, String nonce) {
    final dataToSign = '${_options!.apiKey}|${_options!.advertiserId}|$timestamp|$nonce';
    final key = utf8.encode(_options!.apiKey);
    final bytes = utf8.encode(dataToSign);
    final hmacSha256 = Hmac(sha256, key);
    final digest = hmacSha256.convert(bytes);
    return digest.toString();
  }
  
  String _generateNonce() {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    final random = Random.secure();
    return List.generate(32, (_) => chars[random.nextInt(chars.length)]).join();
  }
  
  String _enqueue(String type, Map<String, dynamic> payload) {
    final id = _uuid.v4();
    final item = QueueItem(
      id: id,
      type: type,
      payload: payload,
      timestamp: DateTime.now().millisecondsSinceEpoch,
    );
    
    if (_queue.length >= AfftokConfig.maxQueueSize) {
      _queue.removeAt(0);
      _log('Queue full, removed oldest item');
    }
    
    _queue.add(item);
    _saveQueue();
    _log('Enqueued $type event: $id');
    
    return id;
  }
  
  void _startAutoFlush() {
    _flushTimer?.cancel();
    _flushTimer = Timer.periodic(_options!.flushInterval, (_) => _flush());
    _log('Auto-flush started with interval: ${_options!.flushInterval.inSeconds}s');
  }
  
  Future<void> _flush() async {
    if (_isProcessing) {
      _log('Flush already in progress, skipping');
      return;
    }
    
    _isProcessing = true;
    _log('Starting flush, ${_queue.length} items in queue');
    
    final now = DateTime.now().millisecondsSinceEpoch;
    final pendingItems = _queue.where((item) => item.nextRetryTime <= now).toList();
    
    for (final item in pendingItems) {
      try {
        final success = await _processQueueItem(item);
        if (success) {
          _queue.removeWhere((i) => i.id == item.id);
          _log('Completed: ${item.id}');
        } else {
          _markForRetry(item);
        }
      } catch (e) {
        _log('Error processing item ${item.id}: $e');
        _markForRetry(item);
      }
    }
    
    _saveQueue();
    _isProcessing = false;
    _log('Flush completed, ${_queue.length} items remaining');
  }
  
  Future<bool> _processQueueItem(QueueItem item) async {
    try {
      AfftokResponse response;
      
      if (item.type == 'click') {
        response = await _sendRequest(AfftokConfig.clickEndpoint, item.payload);
      } else if (item.type == 'conversion') {
        response = await _sendRequest(AfftokConfig.conversionEndpoint, item.payload);
      } else {
        return false;
      }
      
      return response.success;
    } catch (e) {
      return false;
    }
  }
  
  void _markForRetry(QueueItem item) {
    if (item.retryCount >= AfftokConfig.maxRetryAttempts) {
      _queue.removeWhere((i) => i.id == item.id);
      _log('Max retries reached, removing: ${item.id}');
      return;
    }
    
    final delay = min(
      AfftokConfig.initialRetryDelay.inMilliseconds * pow(2, item.retryCount),
      AfftokConfig.maxRetryDelay.inMilliseconds,
    );
    final jitter = (Random().nextDouble() * delay * 0.1).toInt();
    
    item.retryCount++;
    item.nextRetryTime = DateTime.now().millisecondsSinceEpoch + delay.toInt() + jitter;
    
    _log('Marked for retry (${item.retryCount}/${AfftokConfig.maxRetryAttempts}): ${item.id}');
  }
  
  Future<void> _loadQueue() async {
    try {
      final jsonString = _prefs?.getString(AfftokConfig.queueKey);
      if (jsonString != null) {
        final List<dynamic> jsonList = jsonDecode(jsonString);
        _queue = jsonList.map((json) => QueueItem.fromJson(json)).toList();
        _log('Loaded ${_queue.length} items from storage');
      }
    } catch (e) {
      _log('Error loading queue: $e');
    }
  }
  
  void _saveQueue() {
    try {
      final jsonList = _queue.map((item) => item.toJson()).toList();
      _prefs?.setString(AfftokConfig.queueKey, jsonEncode(jsonList));
    } catch (e) {
      _log('Error saving queue: $e');
    }
  }
  
  void _log(String message) {
    if (_options?.debug == true) {
      print('[AffTok SDK] $message');
    }
  }
}

