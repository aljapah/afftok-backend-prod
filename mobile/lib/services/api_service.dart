import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../config/api_config.dart';
import '../models/user.dart';

class ApiService {
  static final ApiService _instance = ApiService._internal();
  factory ApiService() => _instance;
  ApiService._internal();

  String? _accessToken;
  String? _refreshToken;

  // Initialize tokens from storage
  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _accessToken = prefs.getString('access_token');
    _refreshToken = prefs.getString('refresh_token');
    print('[ApiService] init() - accessToken: ${_accessToken?.substring(0, 20)}...');
    print('[ApiService] init() - refreshToken: ${_refreshToken?.substring(0, 20)}...');
  }

  // Save tokens to storage
  Future<void> _saveTokens(String accessToken, String refreshToken) async {
    _accessToken = accessToken;
    _refreshToken = refreshToken;
    
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('access_token', accessToken);
    await prefs.setString('refresh_token', refreshToken);
    print('[ApiService] Tokens saved successfully');
  }

  // Clear tokens
  Future<void> clearTokens() async {
    _accessToken = null;
    _refreshToken = null;
    
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('access_token');
    await prefs.remove('refresh_token');
  }

  // Get headers
  Map<String, String> _getHeaders({bool includeAuth = false}) {
    final headers = {
      'Content-Type': 'application/json',
    };
    
    if (includeAuth && _accessToken != null) {
      headers['Authorization'] = 'Bearer $_accessToken';
    }
    
    return headers;
  }

  // Register
  Future<Map<String, dynamic>> register({
    required String username,
    required String email,
    required String password,
    String? fullName,
  }) async {
    try {
      final response = await http.post(
        Uri.parse('${ApiConfig.baseUrl}${ApiConfig.register}'),
        headers: _getHeaders(),
        body: jsonEncode({
          'username': username,
          'email': email,
          'password': password,
          'full_name': fullName,
        }),
      ).timeout(const Duration(seconds: 30));

      if (response.body.isEmpty) {
        return {
          'success': false,
          'error': 'Empty response from server',
        };
      }

      final data = jsonDecode(response.body);

      if (response.statusCode == 201) {
        // Save tokens
        if (data['access_token'] != null && data['refresh_token'] != null) {
          await _saveTokens(
            data['access_token'] as String,
            data['refresh_token'] as String,
          );
        }
        
        return {
          'success': true,
          'user': data['user'] != null ? User.fromJson(data['user']) : null,
          'message': data['message'],
        };
      } else {
        return {
          'success': false,
          'error': data['error'] ?? 'Registration failed',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Network error: ${e.toString()}',
      };
    }
  }

  // Login
  Future<Map<String, dynamic>> login({
    required String username,
    required String password,
  }) async {
    try {
      final response = await http.post(
        Uri.parse('${ApiConfig.baseUrl}${ApiConfig.login}'),
        headers: _getHeaders(),
        body: jsonEncode({
          'username': username,
          'password': password,
        }),
      ).timeout(const Duration(seconds: 30));

      if (response.body.isEmpty) {
        return {
          'success': false,
          'error': 'Empty response from server',
        };
      }

      final data = jsonDecode(response.body);

      if (response.statusCode == 200) {
        // Save tokens
        if (data['access_token'] != null && data['refresh_token'] != null) {
          await _saveTokens(
            data['access_token'] as String,
            data['refresh_token'] as String,
          );
        }
        
        return {
          'success': true,
          'user': data['user'] != null ? User.fromJson(data['user']) : null,
          'message': data['message'],
        };
      } else {
        return {
          'success': false,
          'error': data['error'] ?? 'Login failed',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Network error: ${e.toString()}',
      };
    }
  }

  // Get current user
  Future<Map<String, dynamic>> getMe() async {
    try {
      final response = await http.get(
        Uri.parse('${ApiConfig.baseUrl}${ApiConfig.getMe}'),
        headers: _getHeaders(includeAuth: true),
      ).timeout(const Duration(seconds: 30));

      if (response.body.isEmpty) {
        return {
          'success': false,
          'error': 'Empty response from server',
        };
      }

      final data = jsonDecode(response.body);

      if (response.statusCode == 200) {
        return {
          'success': true,
          'user': data['user'] != null ? User.fromJson(data['user']) : null,
        };
      } else {
        return {
          'success': false,
          'error': data['error'] ?? 'Failed to get user info',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Network error: ${e.toString()}',
      };
    }
  }

  // Logout
  Future<void> logout() async {
    try {
      await http.post(
        Uri.parse('${ApiConfig.baseUrl}${ApiConfig.logout}'),
        headers: _getHeaders(includeAuth: true),
      ).timeout(const Duration(seconds: 30));
    } catch (e) {
      // Ignore errors on logout
    } finally {
      await clearTokens();
    }
  }

  // Generic GET request
  Future<Map<String, dynamic>> get(String endpoint) async {
    try {
      final response = await http.get(
        Uri.parse('${ApiConfig.baseUrl}$endpoint'),
        headers: _getHeaders(includeAuth: true),
      ).timeout(const Duration(seconds: 30));

      if (response.body.isEmpty) {
        return {
          'success': false,
          'error': 'Empty response from server',
        };
      }

      final data = jsonDecode(response.body);

      if (response.statusCode == 200) {
        return {
          'success': true,
          'data': data,
        };
      } else {
        return {
          'success': false,
          'error': data['error'] ?? 'Request failed',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Network error: ${e.toString()}',
      };
    }
  }

  // Generic PUT request
  Future<Map<String, dynamic>> put(String endpoint, Map<String, dynamic> body) async {
    try {
      final response = await http.put(
        Uri.parse('${ApiConfig.baseUrl}$endpoint'),
        headers: _getHeaders(includeAuth: true),
        body: jsonEncode(body),
      ).timeout(const Duration(seconds: 30));

      if (response.body.isEmpty) {
        return {
          'success': false,
          'error': 'Empty response from server',
        };
      }

      final data = jsonDecode(response.body);

      if (response.statusCode == 200) {
        return {
          'success': true,
          'data': data,
        };
      } else {
        return {
          'success': false,
          'error': data['error'] ?? 'Update failed',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Network error: ${e.toString()}',
      };
    }
  }

  // Generic DELETE request
  Future<Map<String, dynamic>> delete(String endpoint) async {
    try {
      final response = await http.delete(
        Uri.parse('${ApiConfig.baseUrl}$endpoint'),
        headers: _getHeaders(includeAuth: true),
      ).timeout(const Duration(seconds: 30));

      if (response.body.isEmpty) {
        return {
          'success': false,
          'error': 'Empty response from server',
        };
      }

      final data = jsonDecode(response.body);

      if (response.statusCode == 200) {
        return {
          'success': true,
          'data': data,
        };
      } else {
        return {
          'success': false,
          'error': data['error'] ?? 'Delete failed',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Network error: ${e.toString()}',
      };
    }
  }

  // Upload file
  Future<Map<String, dynamic>> uploadFile(String endpoint, String filePath) async {
    try {
      final file = File(filePath);
      final request = http.MultipartRequest(
        'POST',
        Uri.parse('${ApiConfig.baseUrl}$endpoint'),
      );

      request.headers.addAll(_getHeaders(includeAuth: true));
      request.files.add(
        await http.MultipartFile.fromPath('avatar', file.path),
      );

      final response = await request.send().timeout(const Duration(seconds: 30));
      final responseBody = await response.stream.bytesToString();

      if (responseBody.isEmpty) {
        return {
          'success': false,
          'error': 'Empty response from server',
        };
      }

      final data = jsonDecode(responseBody);

      if (response.statusCode == 200) {
        return {
          'success': true,
          'data': data,
        };
      } else {
        return {
          'success': false,
          'error': data['error'] ?? 'Upload failed',
        };
      }
    } catch (e) {
      return {
        'success': false,
        'error': 'Network error: ${e.toString()}',
      };
    }
  }

  // Check if user is logged in
  bool get isLoggedIn => _accessToken != null;
  
  // Get access token
  String? get accessToken => _accessToken;
}
