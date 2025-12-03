import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/contest.dart';
import '../config/api_config.dart';
import 'api_service.dart';

class ContestService {
  final ApiService _apiService;
  
  ContestService(this._apiService);
  
  String get _baseUrl => ApiConfig.baseUrl;
  Map<String, String> get _headers => {
    'Content-Type': 'application/json',
    if (_apiService.accessToken != null)
      'Authorization': 'Bearer ${_apiService.accessToken}',
  };
  
  /// Get all active contests
  Future<List<Contest>> getActiveContests() async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/contests'),
        headers: _headers,
      );
      
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final contests = (data['contests'] as List?)
            ?.map((c) => Contest.fromJson(c))
            .toList() ?? [];
        return contests;
      }
      return [];
    } catch (e) {
      print('Error fetching active contests: $e');
      return [];
    }
  }
  
  /// Get a single contest with details
  Future<Contest?> getContest(String contestId) async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/contests/$contestId'),
        headers: _headers,
      );
      
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return Contest.fromJson(data['contest']);
      }
      return null;
    } catch (e) {
      print('Error fetching contest: $e');
      return null;
    }
  }
  
  /// Get contest leaderboard
  Future<List<ContestParticipant>> getContestLeaderboard(String contestId) async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/contests/$contestId/leaderboard'),
        headers: _headers,
      );
      
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final leaderboard = (data['leaderboard'] as List?)
            ?.map((p) => ContestParticipant.fromJson(p))
            .toList() ?? [];
        return leaderboard;
      }
      return [];
    } catch (e) {
      print('Error fetching leaderboard: $e');
      return [];
    }
  }
  
  /// Join a contest - returns {success: bool, message: String}
  Future<Map<String, dynamic>> joinContest(String contestId) async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/contests/$contestId/join'),
        headers: _headers,
      );
      
      final data = jsonDecode(response.body);
      
      if (response.statusCode == 200) {
        return {'success': true, 'message': data['message'] ?? 'Joined successfully'};
      } else {
        return {'success': false, 'message': data['error'] ?? 'Failed to join'};
      }
    } catch (e) {
      print('Error joining contest: $e');
      return {'success': false, 'message': 'Connection error'};
    }
  }
  
  /// Get contests the user is participating in
  Future<List<ContestParticipant>> getMyContests() async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/contests/my'),
        headers: _headers,
      );
      
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final contests = (data['contests'] as List?)
            ?.map((c) => ContestParticipant.fromJson(c))
            .toList() ?? [];
        return contests;
      }
      return [];
    } catch (e) {
      print('Error fetching my contests: $e');
      return [];
    }
  }
}
