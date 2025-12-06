import 'dart:convert';
import 'package:http/http.dart' as http;
import '../config/api_config.dart';
import '../models/offer.dart';
import '../models/user_offer.dart';
import '../services/api_service.dart';

class OfferService {
  final ApiService _apiService = ApiService();

  Future<Map<String, dynamic>> getAllOffers({
    String? status,
    String? category,
    String? sort,
    String? order,
  }) async {
    try {
      final queryParams = <String, String>{};
      if (status != null) queryParams['status'] = status;
      if (category != null) queryParams['category'] = category;
      if (sort != null) queryParams['sort'] = sort;
      if (order != null) queryParams['order'] = order;
      
      final uri = Uri.parse('${ApiConfig.baseUrl}/api/offers')
          .replace(queryParameters: queryParams.isNotEmpty ? queryParams : null);
      
      print('[OfferService] Fetching all offers from: $uri');
      
      final response = await http.get(
        uri,
        headers: {'Content-Type': 'application/json'},
      ).timeout(ApiConfig.connectTimeout);
      
      print('[OfferService] Response status: ${response.statusCode}');
      print('[OfferService] Response body: ${response.body}');
      
      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        
        if (data['offers'] == null || (data['offers'] is! List)) {
          print('[OfferService] No offers array in response');
          return {
            'success': true,
            'offers': <Offer>[],
          };
        }
        
        final offersJson = data['offers'] as List;
        print('[OfferService] Found ${offersJson.length} offers in response');
        
        if (offersJson.isEmpty) {
          return {
            'success': true,
            'offers': <Offer>[],
          };
        }
        
        final offers = <Offer>[];
        for (final json in offersJson) {
          try {
            final offer = Offer.fromJson(json as Map<String, dynamic>);
            offers.add(offer);
          } catch (e) {
            print('[OfferService] Error parsing offer: $e');
          }
        }
        
        print('[OfferService] Successfully parsed ${offers.length} offers');
        return {
          'success': true,
          'offers': offers,
        };
      } else {
        print('[OfferService] Failed with status: ${response.statusCode}');
        return {
          'success': false,
          'error': 'Failed to load offers: ${response.statusCode}',
        };
      }
    } catch (e) {
      print('[OfferService] Exception: $e');
      return {
        'success': false,
        'error': 'Error fetching offers: $e',
      };
    }
  }

  Future<Map<String, dynamic>> getMyOffers() async {
    try {
      print('[OfferService] Fetching my offers...');
      final result = await _apiService.get('/api/offers/my');
      print('[OfferService] Result: ${result['success']}');
      
      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>?;
        print('[OfferService] Data: $data');
        
        if (data == null || data['offers'] == null || (data['offers'] is! List)) {
          print('[OfferService] No offers found in response');
          return {
            'success': true,
            'offers': <UserOffer>[],
          };
        }
        
        final offersJson = data['offers'] as List;
        print('[OfferService] Found ${offersJson.length} offers');
        
        if (offersJson.isEmpty) {
          return {
            'success': true,
            'offers': <UserOffer>[],
          };
        }
        
        final offers = <UserOffer>[];
        for (final json in offersJson) {
          try {
            final userOffer = UserOffer.fromJson(json as Map<String, dynamic>);
            offers.add(userOffer);
            print('[OfferService] Parsed offer: ${userOffer.id}, title: ${userOffer.offerTitle}');
          } catch (e) {
            print('[OfferService] Error parsing offer: $e, json: $json');
          }
        }
        
        return {
          'success': true,
          'offers': offers,
        };
      } else {
        print('[OfferService] Failed: ${result['error']}');
        return {
          'success': false,
          'error': result['error'] ?? 'Failed to load offers',
        };
      }
    } catch (e) {
      print('[OfferService] Exception: $e');
      return {
        'success': false,
        'error': 'Error fetching offers: $e',
      };
    }
  }

  /// Join an offer and get unique tracking link
  /// Backend returns: { success, message, user_offer, affiliate_link, tracking_url, short_link }
  Future<Map<String, dynamic>> joinOffer(String offerId) async {
    try {
      print('[OfferService] Joining offer $offerId...');
      print('[OfferService] API endpoint: /api/offers/$offerId/join');
      
      final result = await _apiService.post('/api/offers/$offerId/join', {});
      print('[OfferService] Join API response: $result');
      print('[OfferService] Join result success: ${result['success']}');
      
      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>? ?? {};
        final userOfferData = data['user_offer'] as Map<String, dynamic>?;
        
        print('[OfferService] Offer joined successfully!');
        print('[OfferService] Affiliate link: ${data['affiliate_link']}');
        print('[OfferService] Short link: ${data['short_link']}');
        print('[OfferService] Tracking URL: ${data['tracking_url']}');
        
        return {
          'success': true,
          'message': data['message'] ?? 'Offer joined successfully',
          'userOffer': userOfferData != null ? UserOffer.fromJson(userOfferData) : null,
          'userOfferRaw': userOfferData,
          'affiliateLink': data['affiliate_link'],
          'shortLink': data['short_link'],
          'trackingUrl': data['tracking_url'],
        };
      } else {
        // Check if it's "already joined" error - treat as success
        final error = result['error']?.toString() ?? '';
        if (error.toLowerCase().contains('already joined') || 
            error.toLowerCase().contains('already added')) {
          print('[OfferService] Offer already joined - treating as success');
          return {
            'success': true,
            'alreadyJoined': true,
            'message': error,
          };
        }
        
        print('[OfferService] Join failed with error: ${result['error']}');
        return {
          'success': false,
          'error': result['error'] ?? 'Failed to join offer',
        };
      }
    } catch (e) {
      print('[OfferService] Exception joining offer: $e');
      return {
        'success': false,
        'error': 'Error joining offer: $e',
      };
    }
  }

  /// Get a single offer by ID
  Future<Map<String, dynamic>> getOfferById(String offerId) async {
    try {
      print('[OfferService] Fetching offer by ID: $offerId');
      
      final response = await http.get(
        Uri.parse('${ApiConfig.baseUrl}/api/offers/$offerId'),
        headers: {'Content-Type': 'application/json'},
      ).timeout(ApiConfig.connectTimeout);
      
      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final offerJson = data['offer'] as Map<String, dynamic>?;
        
        if (offerJson != null) {
          return {
            'success': true,
            'offer': Offer.fromJson(offerJson),
          };
        } else {
          return {
            'success': false,
            'error': 'Offer not found',
          };
        }
      } else {
        return {
          'success': false,
          'error': 'Failed to load offer: ${response.statusCode}',
        };
      }
    } catch (e) {
      print('[OfferService] Exception: $e');
      return {
        'success': false,
        'error': 'Error fetching offer: $e',
      };
    }
  }

  Future<Map<String, dynamic>> getOfferStats(String offerId) async {
    try {
      final result = await _apiService.get('/api/offers/$offerId/stats');
      
      if (result['success'] == true) {
        return {
          'success': true,
          'stats': result['data'],
        };
      } else {
        return {
          'success': false,
          'error': result['error'] ?? 'Failed to load stats',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Error fetching stats: $e',
      };
    }
  }
}
