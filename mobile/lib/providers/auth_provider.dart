import 'package:flutter/foundation.dart';
import '../models/user.dart';
import '../services/api_service.dart';

class AuthProvider with ChangeNotifier {
  final ApiService _apiService = ApiService();
  
  User? _currentUser;
  bool _isLoading = false;
  String? _error;

  User? get currentUser => _currentUser;
  User? get user => _currentUser; // Alias for currentUser
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isLoggedIn => _currentUser != null;

  // Initialize
  Future<void> init() async {
    await _apiService.init();
    if (_apiService.isLoggedIn) {
      await loadCurrentUser();
    }
  }

  // Register
  Future<bool> register({
    required String username,
    required String email,
    required String password,
    String? fullName,
  }) async {
    _setLoading(true);
    _error = null;

    final result = await _apiService.register(
      username: username,
      email: email,
      password: password,
      fullName: fullName,
    );

    _setLoading(false);

    if (result['success'] == true) {
      _currentUser = result['user'] as User;
      notifyListeners();
      return true;
    } else {
      _error = result['error'] as String;
      notifyListeners();
      return false;
    }
  }

  // Login
  Future<bool> login({
    required String username,
    required String password,
  }) async {
    _setLoading(true);
    _error = null;

    final result = await _apiService.login(
      username: username,
      password: password,
    );

    _setLoading(false);

    if (result['success'] == true) {
      _currentUser = result['user'] as User;
      notifyListeners();
      return true;
    } else {
      _error = result['error'] as String;
      notifyListeners();
      return false;
    }
  }

  // Load current user
  Future<void> loadCurrentUser() async {
    final result = await _apiService.getMe();
    
    if (result['success'] == true) {
      _currentUser = result['user'] as User;
      notifyListeners();
    }
  }

  // Update Profile
  Future<bool> updateProfile({
    String? fullName,
    String? email,
    String? phone,
  }) async {
    if (_currentUser == null) return false;
    
    _setLoading(true);
    _error = null;

    try {
      final response = await _apiService.put('/users/${_currentUser!.id}', {
        if (fullName != null) 'full_name': fullName,
        if (email != null) 'email': email,
        if (phone != null) 'phone': phone,
      });

      _setLoading(false);

      if (response['success'] == true) {
        await loadCurrentUser();
        return true;
      } else {
        _error = response['error'] as String?;
        notifyListeners();
        return false;
      }
    } catch (e) {
      _setLoading(false);
      _error = e.toString();
      notifyListeners();
      return false;
    }
  }

  // Logout
  Future<void> logout() async {
    await _apiService.logout();
    _currentUser = null;
    notifyListeners();
  }

  // Clear error
  void clearError() {
    _error = null;
    notifyListeners();
  }

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }
}
