import 'package:flutter/foundation.dart';
import '../models/stats.dart';
import '../services/stats_service.dart';

/// Provider for managing user analytics and stats state
class StatsProvider with ChangeNotifier {
  final StatsService _statsService = StatsService();

  // State
  UserAnalytics? _analytics;
  List<DailyStats> _dailyStats = [];
  List<OfferClickStats> _offerStats = [];
  List<ClickEvent> _recentClicks = [];
  
  bool _isLoading = false;
  String? _error;
  DateTime? _lastFetch;

  // Getters
  UserAnalytics? get analytics => _analytics;
  List<DailyStats> get dailyStats => _dailyStats;
  List<OfferClickStats> get offerStats => _offerStats;
  List<ClickEvent> get recentClicks => _recentClicks;
  bool get isLoading => _isLoading;
  String? get error => _error;
  
  // Computed stats from analytics
  int get totalClicks => _analytics?.totalClicks ?? 0;
  int get totalConversions => _analytics?.totalConversions ?? 0;
  int get totalEarnings => _analytics?.totalEarnings ?? 0;
  double get conversionRate => _analytics?.conversionRate ?? 0.0;
  int get activeOffers => _analytics?.activeOffers ?? 0;
  
  // Time-based stats
  int get clicksToday => _analytics?.clicksToday ?? 0;
  int get clicksThisWeek => _analytics?.clicksThisWeek ?? 0;
  int get clicksThisMonth => _analytics?.clicksThisMonth ?? 0;
  
  int get conversionsToday => _analytics?.conversionsToday ?? 0;
  int get conversionsThisWeek => _analytics?.conversionsThisWeek ?? 0;
  int get conversionsThisMonth => _analytics?.conversionsThisMonth ?? 0;

  // Check if data needs refresh (older than 1 minute)
  bool get needsRefresh {
    if (_lastFetch == null) return true;
    return DateTime.now().difference(_lastFetch!).inMinutes >= 1;
  }

  /// Load all analytics data
  Future<void> loadAllStats({bool forceRefresh = false}) async {
    if (_isLoading) return;
    if (!forceRefresh && !needsRefresh && _analytics != null) {
      print('[StatsProvider] Using cached stats');
      return;
    }

    _setLoading(true);
    _error = null;

    try {
      print('[StatsProvider] Loading all stats...');
      
      // Load analytics and daily stats in parallel
      final results = await Future.wait([
        _statsService.getUserAnalytics(),
        _statsService.getDailyStats(days: 7),
        _statsService.getClicksByOffer(),
      ]);

      // Process analytics
      final analyticsResult = results[0];
      if (analyticsResult['success'] == true) {
        _analytics = analyticsResult['stats'] as UserAnalytics?;
        print('[StatsProvider] Analytics loaded: clicks=${_analytics?.totalClicks}, conversions=${_analytics?.totalConversions}');
      } else {
        _error = analyticsResult['error'] as String?;
      }

      // Process daily stats
      final dailyResult = results[1];
      if (dailyResult['success'] == true) {
        _dailyStats = dailyResult['data'] as List<DailyStats>? ?? [];
        print('[StatsProvider] Daily stats loaded: ${_dailyStats.length} days');
      }

      // Process offer stats
      final offerResult = results[2];
      if (offerResult['success'] == true) {
        _offerStats = offerResult['stats'] as List<OfferClickStats>? ?? [];
        print('[StatsProvider] Offer stats loaded: ${_offerStats.length} offers');
      }

      _lastFetch = DateTime.now();
    } catch (e) {
      print('[StatsProvider] Error loading stats: $e');
      _error = e.toString();
    }

    _setLoading(false);
  }

  /// Load only user analytics
  Future<void> loadAnalytics({bool forceRefresh = false}) async {
    if (_isLoading) return;
    if (!forceRefresh && !needsRefresh && _analytics != null) return;

    _setLoading(true);
    _error = null;

    try {
      final result = await _statsService.getUserAnalytics();
      
      if (result['success'] == true) {
        _analytics = result['stats'] as UserAnalytics?;
        _lastFetch = DateTime.now();
      } else {
        _error = result['error'] as String?;
      }
    } catch (e) {
      _error = e.toString();
    }

    _setLoading(false);
  }

  /// Load daily stats for charts
  Future<void> loadDailyStats({int days = 7, bool forceRefresh = false}) async {
    if (_isLoading) return;
    if (!forceRefresh && _dailyStats.isNotEmpty) return;

    _setLoading(true);
    _error = null;

    try {
      final result = await _statsService.getDailyStats(days: days);
      
      if (result['success'] == true) {
        _dailyStats = result['data'] as List<DailyStats>? ?? [];
      } else {
        _error = result['error'] as String?;
      }
    } catch (e) {
      _error = e.toString();
    }

    _setLoading(false);
  }

  /// Load recent clicks
  Future<void> loadRecentClicks({bool forceRefresh = false}) async {
    if (_isLoading) return;
    if (!forceRefresh && _recentClicks.isNotEmpty) return;

    _setLoading(true);
    _error = null;

    try {
      final result = await _statsService.getMyClicks();
      
      if (result['success'] == true) {
        _recentClicks = result['clicks'] as List<ClickEvent>? ?? [];
      } else {
        _error = result['error'] as String?;
      }
    } catch (e) {
      _error = e.toString();
    }

    _setLoading(false);
  }

  /// Load offer stats
  Future<void> loadOfferStats({bool forceRefresh = false}) async {
    if (_isLoading) return;
    if (!forceRefresh && _offerStats.isNotEmpty) return;

    _setLoading(true);
    _error = null;

    try {
      final result = await _statsService.getClicksByOffer();
      
      if (result['success'] == true) {
        _offerStats = result['stats'] as List<OfferClickStats>? ?? [];
      } else {
        _error = result['error'] as String?;
      }
    } catch (e) {
      _error = e.toString();
    }

    _setLoading(false);
  }

  /// Force refresh all data
  Future<void> refresh() async {
    await loadAllStats(forceRefresh: true);
  }

  /// Clear all stats (on logout)
  void clearStats() {
    _analytics = null;
    _dailyStats = [];
    _offerStats = [];
    _recentClicks = [];
    _error = null;
    _lastFetch = null;
    notifyListeners();
  }

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }
  
  void clearError() {
    _error = null;
    notifyListeners();
  }
}

