import 'dart:convert';
import 'package:http/http.dart' as http;
import 'api_service.dart';

class AdvertiserService {
  final String _baseUrl = ApiService.baseUrl;

  /// Get advertiser dashboard summary
  Future<Map<String, dynamic>> getDashboard(String token) async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/advertiser/dashboard'),
        headers: {
          'Authorization': 'Bearer $token',
          'Content-Type': 'application/json',
        },
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('Failed to load dashboard: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error loading dashboard: $e');
    }
  }

  /// Get all offers created by the advertiser
  Future<Map<String, dynamic>> getMyOffers(String token) async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/advertiser/offers'),
        headers: {
          'Authorization': 'Bearer $token',
          'Content-Type': 'application/json',
        },
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('Failed to load offers: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error loading offers: $e');
    }
  }

  /// Get stats for a specific offer
  Future<Map<String, dynamic>> getOfferStats(String token, String offerId) async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/advertiser/offers/$offerId/stats'),
        headers: {
          'Authorization': 'Bearer $token',
          'Content-Type': 'application/json',
        },
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('Failed to load offer stats: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error loading offer stats: $e');
    }
  }

  /// Create a new offer (will be pending approval)
  Future<bool> createOffer(String token, Map<String, dynamic> offerData) async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/advertiser/offers'),
        headers: {
          'Authorization': 'Bearer $token',
          'Content-Type': 'application/json',
        },
        body: json.encode(offerData),
      );

      if (response.statusCode == 201) {
        return true;
      } else {
        final data = json.decode(response.body);
        throw Exception(data['error'] ?? 'Failed to create offer');
      }
    } catch (e) {
      throw Exception('Error creating offer: $e');
    }
  }

  /// Update an existing offer (only pending/rejected offers can be updated)
  Future<bool> updateOffer(String token, String offerId, Map<String, dynamic> offerData) async {
    try {
      final response = await http.put(
        Uri.parse('$_baseUrl/api/advertiser/offers/$offerId'),
        headers: {
          'Authorization': 'Bearer $token',
          'Content-Type': 'application/json',
        },
        body: json.encode(offerData),
      );

      if (response.statusCode == 200) {
        return true;
      } else {
        final data = json.decode(response.body);
        throw Exception(data['error'] ?? 'Failed to update offer');
      }
    } catch (e) {
      throw Exception('Error updating offer: $e');
    }
  }

  /// Delete an offer (only pending offers can be deleted)
  Future<bool> deleteOffer(String token, String offerId) async {
    try {
      final response = await http.delete(
        Uri.parse('$_baseUrl/api/advertiser/offers/$offerId'),
        headers: {
          'Authorization': 'Bearer $token',
          'Content-Type': 'application/json',
        },
      );

      if (response.statusCode == 200) {
        return true;
      } else {
        final data = json.decode(response.body);
        throw Exception(data['error'] ?? 'Failed to delete offer');
      }
    } catch (e) {
      throw Exception('Error deleting offer: $e');
    }
  }
}
