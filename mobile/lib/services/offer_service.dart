import 'dart:convert';
import 'package:http/http.dart' as http;
import '../config/api_config.dart';
import '../models/offer.dart';
import '../models/user_offer.dart';

class OfferService {
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
          .replace(queryParameters: queryParams);

      final response = await http.get(
        uri,
        headers: {'Content-Type': 'application/json'},
      ).timeout(ApiConfig.connectTimeout);

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        
        // فحص إذا كانت البيانات موجودة وليست فارغة
        if (data['offers'] == null || (data['offers'] is! List)) {
          return {
            'success': true,
            'offers': [],
          };
        }
        
        final offersJson = data['offers'] as List;
        
        // فحص إذا كانت القائمة فارغة
        if (offersJson.isEmpty) {
          return {
            'success': true,
            'offers': [],
          };
        }
        
        final offers = offersJson.map((json) => Offer.fromJson(json)).toList();

        return {
          'success': true,
          'offers': offers,
        };
      } else {
        return {
          'success': false,
          'error': 'Failed to load offers: ${response.statusCode}',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Error fetching offers: $e',
      };
    }
  }

  Future<Map<String, dynamic>> getMyOffers() async {
    try {
      final uri = Uri.parse('${ApiConfig.baseUrl}/api/offers/my');

      final response = await http.get(
        uri,
        headers: {
          'Content-Type': 'application/json',
        },
      ).timeout(ApiConfig.connectTimeout);

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        
        // فحص إذا كانت البيانات موجودة وليست فارغة
        if (data['offers'] == null || (data['offers'] is! List)) {
          return {
            'success': true,
            'offers': [],
          };
        }
        
        final offersJson = data['offers'] as List;
        
        // فحص إذا كانت القائمة فارغة
        if (offersJson.isEmpty) {
          return {
            'success': true,
            'offers': [],
          };
        }
        
        final offers = offersJson.map((json) => UserOffer.fromJson(json)).toList();

        return {
          'success': true,
          'offers': offers,
        };
      } else {
        return {
          'success': false,
          'error': 'Failed to load my offers: ${response.statusCode}',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Error fetching my offers: $e',
      };
    }
  }

  Future<Map<String, dynamic>> joinOffer(String offerId) async {
    try {
      final uri = Uri.parse('${ApiConfig.baseUrl}/api/offers/$offerId/join');

      final response = await http.post(
        uri,
        headers: {
          'Content-Type': 'application/json',
        },
      ).timeout(ApiConfig.connectTimeout);

      if (response.statusCode == 200 || response.statusCode == 201) {
        final data = json.decode(response.body);

        return {
          'success': true,
          'message': data['message'] ?? 'Successfully joined offer',
        };
      } else {
        final data = json.decode(response.body);
        return {
          'success': false,
          'error': data['error'] ?? 'Failed to join offer',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Error joining offer: $e',
      };
    }
  }
}
