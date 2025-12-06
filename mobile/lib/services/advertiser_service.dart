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
      ).timeout(const Duration(seconds: 30));

      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else if (response.statusCode == 401) {
        throw Exception('Session expired. Please login again.');
      } else if (response.statusCode == 403) {
        throw Exception('Access denied. You must be an advertiser.');
      } else {
        final errorBody = response.body.isNotEmpty ? json.decode(response.body) : {};
        throw Exception(errorBody['error'] ?? 'Failed to load dashboard: ${response.statusCode}');
      }
    } on http.ClientException {
      throw Exception('Network error. Please check your connection.');
    } catch (e) {
      if (e.toString().contains('TimeoutException')) {
        throw Exception('Request timed out. Please try again.');
      }
      rethrow;
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
      ).timeout(const Duration(seconds: 30));

      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else if (response.statusCode == 401) {
        throw Exception('Session expired. Please login again.');
      } else if (response.statusCode == 403) {
        throw Exception('Access denied. You must be an advertiser.');
      } else {
        final errorBody = response.body.isNotEmpty ? json.decode(response.body) : {};
        throw Exception(errorBody['error'] ?? 'Failed to load offers: ${response.statusCode}');
      }
    } on http.ClientException {
      throw Exception('Network error. Please check your connection.');
    } catch (e) {
      if (e.toString().contains('TimeoutException')) {
        throw Exception('Request timed out. Please try again.');
      }
      rethrow;
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

  /// Get all promoters who joined advertiser's offers with their stats
  Future<Map<String, dynamic>> getPromoters(String token) async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/advertiser/promoters'),
        headers: {
          'Authorization': 'Bearer $token',
          'Content-Type': 'application/json',
        },
      ).timeout(const Duration(seconds: 30));

      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else if (response.statusCode == 401) {
        throw Exception('Session expired. Please login again.');
      } else if (response.statusCode == 403) {
        throw Exception('Access denied. You must be an advertiser.');
      } else {
        final errorBody = response.body.isNotEmpty ? json.decode(response.body) : {};
        throw Exception(errorBody['error'] ?? 'Failed to load promoters: ${response.statusCode}');
      }
    } on http.ClientException {
      throw Exception('Network error. Please check your connection.');
    } catch (e) {
      if (e.toString().contains('TimeoutException')) {
        throw Exception('Request timed out. Please try again.');
      }
      rethrow;
    }
  }
}
