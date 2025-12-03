import '../models/stats.dart';
import 'api_service.dart';

/// Service for fetching analytics and stats from the backend
/// Endpoints: /api/stats/me, /api/stats/daily, /api/clicks/my, /api/clicks/by-offer
class StatsService {
  final ApiService _apiService;

  StatsService([ApiService? apiService]) : _apiService = apiService ?? ApiService();

  /// Get comprehensive user analytics
  /// Endpoint: GET /api/stats/me
  Future<Map<String, dynamic>> getUserAnalytics() async {
    try {
      print('[StatsService] Fetching user analytics...');
      final response = await _apiService.get('/api/stats/me');
      
      if (response['success'] == true) {
        final data = response['data'] as Map<String, dynamic>?;
        final statsJson = data?['stats'] as Map<String, dynamic>? ?? {};
        
        return {
          'success': true,
          'stats': UserAnalytics.fromJson(statsJson),
        };
      } else {
        print('[StatsService] Failed to fetch analytics: ${response['error']}');
        return {
          'success': false,
          'error': response['error'] ?? 'Failed to fetch analytics',
          'stats': UserAnalytics.empty(),
        };
      }
    } catch (e) {
      print('[StatsService] Exception fetching analytics: $e');
      return {
        'success': false,
        'error': e.toString(),
        'stats': UserAnalytics.empty(),
      };
    }
  }

  /// Get daily stats for charts
  /// Endpoint: GET /api/stats/daily?days=7
  Future<Map<String, dynamic>> getDailyStats({int days = 7}) async {
    try {
      print('[StatsService] Fetching daily stats for $days days...');
      final response = await _apiService.get('/api/stats/daily?days=$days');
      
      if (response['success'] == true) {
        final data = response['data'] as Map<String, dynamic>?;
        final statsJson = data?['data'] as List? ?? [];
        
        final stats = statsJson
            .map((json) => DailyStats.fromJson(json as Map<String, dynamic>))
            .toList();
        
        return {
          'success': true,
          'data': stats,
          'days': data?['days'] ?? days,
        };
      } else {
        print('[StatsService] Failed to fetch daily stats: ${response['error']}');
        return {
          'success': false,
          'error': response['error'] ?? 'Failed to fetch daily stats',
          'data': <DailyStats>[],
        };
      }
    } catch (e) {
      print('[StatsService] Exception fetching daily stats: $e');
      return {
        'success': false,
        'error': e.toString(),
        'data': <DailyStats>[],
      };
    }
  }

  /// Get click stats grouped by offer
  /// Endpoint: GET /api/clicks/by-offer
  Future<Map<String, dynamic>> getClicksByOffer() async {
    try {
      print('[StatsService] Fetching clicks by offer...');
      final response = await _apiService.get('/api/clicks/by-offer');
      
      if (response['success'] == true) {
        final data = response['data'] as Map<String, dynamic>?;
        final statsJson = data?['stats'] as List? ?? [];
        
        final stats = statsJson
            .map((json) => OfferClickStats.fromJson(json as Map<String, dynamic>))
            .toList();
        
        return {
          'success': true,
          'stats': stats,
        };
      } else {
        print('[StatsService] Failed to fetch clicks by offer: ${response['error']}');
        return {
          'success': false,
          'error': response['error'] ?? 'Failed to fetch clicks by offer',
          'stats': <OfferClickStats>[],
        };
      }
    } catch (e) {
      print('[StatsService] Exception fetching clicks by offer: $e');
      return {
        'success': false,
        'error': e.toString(),
        'stats': <OfferClickStats>[],
      };
    }
  }

  /// Get recent click events
  /// Endpoint: GET /api/clicks/my
  Future<Map<String, dynamic>> getMyClicks() async {
    try {
      print('[StatsService] Fetching my clicks...');
      final response = await _apiService.get('/api/clicks/my');
      
      if (response['success'] == true) {
        final data = response['data'] as Map<String, dynamic>?;
        final clicksJson = data?['clicks'] as List? ?? [];
        final total = data?['total'] ?? 0;
        
        final clicks = clicksJson
            .map((json) => ClickEvent.fromJson(json as Map<String, dynamic>))
            .toList();
        
        return {
          'success': true,
          'clicks': clicks,
          'total': total,
        };
      } else {
        print('[StatsService] Failed to fetch my clicks: ${response['error']}');
        return {
          'success': false,
          'error': response['error'] ?? 'Failed to fetch clicks',
          'clicks': <ClickEvent>[],
          'total': 0,
        };
      }
    } catch (e) {
      print('[StatsService] Exception fetching my clicks: $e');
      return {
        'success': false,
        'error': e.toString(),
        'clicks': <ClickEvent>[],
        'total': 0,
      };
    }
  }

  /// Get click stats for a specific user offer
  /// Endpoint: GET /api/clicks/:id/stats
  Future<Map<String, dynamic>> getClickStatsForOffer(String userOfferId) async {
    try {
      print('[StatsService] Fetching click stats for offer $userOfferId...');
      final response = await _apiService.get('/api/clicks/$userOfferId/stats');
      
      if (response['success'] == true) {
        final data = response['data'] as Map<String, dynamic>? ?? {};
        return {
          'success': true,
          'stats': data,
        };
      } else {
        print('[StatsService] Failed to fetch click stats: ${response['error']}');
        return {
          'success': false,
          'error': response['error'] ?? 'Failed to fetch click stats',
        };
      }
    } catch (e) {
      print('[StatsService] Exception fetching click stats: $e');
      return {
        'success': false,
        'error': e.toString(),
      };
    }
  }
}

