import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../models/user.dart';
import '../models/user_offer.dart';
import '../services/api_service.dart';
import '../services/offer_service.dart';

class AuthProvider with ChangeNotifier {
  final ApiService _apiService = ApiService();
  final OfferService _offerService = OfferService();
  
  User? _currentUser;
  List<UserOffer> _userOffers = [];
  bool _isLoading = false;
  String? _error;
  DateTime? _lastRefresh;

  User? get currentUser => _currentUser;
  User? get user => _currentUser;
  List<UserOffer> get userOffers => _userOffers;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isLoggedIn => _currentUser != null;
  
  int get totalOffersAdded => _userOffers.length;
  int get totalClicks => _userOffers.fold(0, (sum, offer) => sum + offer.stats.clicks);
  int get totalConversions => _userOffers.fold(0, (sum, offer) => sum + offer.stats.conversions);
  double get overallConversionRate => totalClicks == 0 ? 0 : (totalConversions / totalClicks) * 100;
  
  // Check if data needs refresh (older than 30 seconds)
  bool get needsRefresh {
    if (_lastRefresh == null) return true;
    return DateTime.now().difference(_lastRefresh!).inSeconds > 30;
  }

  // Initialize
  Future<void> init() async {
    print('[AuthProvider] Initializing...');
    await _apiService.init();
    print('[AuthProvider] ApiService initialized');
    
    // Wait for token to be loaded
    int attempts = 0;
    while (_apiService.accessToken == null && attempts < 50) {
      await Future.delayed(const Duration(milliseconds: 100));
      attempts++;
    }
    
    print('[AuthProvider] isLoggedIn: ${_apiService.isLoggedIn}');
    if (_apiService.isLoggedIn) {
      print('[AuthProvider] Calling loadCurrentUser...');
      await loadCurrentUser();
    } else {
      print('[AuthProvider] User not logged in');
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

  /// Load current user and offers from API
  /// [forceRefresh] - if true, always fetch from API even if data exists
  Future<void> loadCurrentUser({bool forceRefresh = false}) async {
    // Skip if already loading
    if (_isLoading) {
      print('[AuthProvider] Already loading, skipping...');
      return;
    }
    
    _setLoading(true);
    _error = null;
    
    try {
      print('[AuthProvider] Loading current user (forceRefresh: $forceRefresh)...');
      final result = await _apiService.getMe();
      print('[AuthProvider] API Response success: ${result['success']}');
      
      if (result['success'] == true && result['user'] != null) {
        _currentUser = result['user'] as User;
        _error = null;
        _lastRefresh = DateTime.now();
        print('[AuthProvider] User loaded: ${_currentUser?.username}');
        print('[AuthProvider] Stats: clicks=${_currentUser?.stats.totalClicks}, conversions=${_currentUser?.stats.totalConversions}');
        await _loadUserOffers();
      } else {
        _currentUser = null;
        _error = result['error'] as String? ?? 'Failed to load user data';
        print('[AuthProvider] Failed: $_error');
      }
    } catch (e) {
      _currentUser = null;
      _error = 'Network error: ${e.toString()}';
      print('[AuthProvider] Exception: $e');
    }
    
    _setLoading(false);
    notifyListeners();
  }
  
  /// Force refresh all profile data
  Future<void> refreshProfile() async {
    print('[AuthProvider] Force refreshing profile...');
    await loadCurrentUser(forceRefresh: true);
  }
  
  Future<void> _loadUserOffers() async {
    try {
      print('[AuthProvider] Loading user offers...');
      final result = await _offerService.getMyOffers();
      print('[AuthProvider] User offers result: ${result['success']}, count: ${(result['offers'] as List?)?.length ?? 0}');
      
      if (result['success'] == true) {
        final offers = result['offers'];
        if (offers != null && offers is List<UserOffer>) {
          _userOffers = offers;
        } else if (offers is List) {
          try {
            _userOffers = List<UserOffer>.from(offers);
          } catch (e) {
            print('[AuthProvider] Error casting offers: $e');
            _userOffers = [];
          }
        } else {
          _userOffers = [];
        }
        print('[AuthProvider] Loaded ${_userOffers.length} user offers');
      } else {
        _userOffers = [];
        print('[AuthProvider] Failed to load offers: ${result['error']}');
      }
      notifyListeners();
    } catch (e) {
      _userOffers = [];
      print('[AuthProvider] Exception loading user offers: $e');
    }
  }
  
  /// Reload only user offers (useful after adding a new offer)
  Future<void> reloadUserOffers() async {
    print('[AuthProvider] Reloading user offers...');
    await _loadUserOffers();
  }
  
  Future<bool> addOfferToProfile(String offerId) async {
    // If user is null, try to load first
    if (_currentUser == null) {
      print('[AuthProvider] User is null, trying to load first...');
      await loadCurrentUser(forceRefresh: true);
      if (_currentUser == null) {
        print('[AuthProvider] Still no user after loading, cannot add offer');
        return false;
      }
    }
    
    try {
      print('[AuthProvider] Adding offer $offerId to profile for user ${_currentUser!.id}...');
      final result = await _offerService.joinOffer(offerId);
      print('[AuthProvider] Join offer result: $result');
      
      if (result['success'] == true) {
        print('[AuthProvider] Offer added successfully, refreshing data...');
        // Refresh both user stats and offers
        await loadCurrentUser(forceRefresh: true);
        return true;
      }
      print('[AuthProvider] Failed to add offer: ${result['error']}');
      return false;
    } catch (e) {
      print('[AuthProvider] Error adding offer: $e');
      return false;
    }
  }
  
  bool hasAddedOffer(String offerId) {
    return _userOffers.any((offer) => offer.offerId == offerId);
  }
  
  UserOffer? getUserOfferById(String offerId) {
    try {
      return _userOffers.firstWhere((offer) => offer.offerId == offerId);
    } catch (e) {
      return null;
    }
  }

  // Upload Profile Image to Imgur and update profile
  Future<bool> uploadProfileImage(String imagePath) async {
    if (_currentUser == null) {
      print('[AuthProvider] Cannot upload image: user is null');
      return false;
    }
    
    _setLoading(true);
    _error = null;

    try {
      print('[AuthProvider] Uploading profile image to Imgur...');
      
      // Read image file and convert to base64
      final File imageFile = File(imagePath);
      final bytes = await imageFile.readAsBytes();
      final base64Image = base64Encode(bytes);
      
      // Upload to Imgur (using anonymous upload)
      final imgurResponse = await http.post(
        Uri.parse('https://api.imgur.com/3/image'),
        headers: {
          'Authorization': 'Client-ID 546c25a59c58ad7', // Free Imgur Client ID
          'Content-Type': 'application/json',
        },
        body: jsonEncode({
          'image': base64Image,
          'type': 'base64',
        }),
      );
      
      print('[AuthProvider] Imgur response status: ${imgurResponse.statusCode}');
      
      if (imgurResponse.statusCode == 200) {
        final imgurData = jsonDecode(imgurResponse.body);
        final imageUrl = imgurData['data']['link'] as String;
        print('[AuthProvider] Image uploaded to Imgur: $imageUrl');
        
        // Now update profile with the new avatar URL
        final response = await _apiService.put('/api/profile', {
          'avatar_url': imageUrl,
        });
        print('[AuthProvider] Profile update response: $response');

        _setLoading(false);

        if (response['success'] == true) {
          print('[AuthProvider] Avatar URL updated successfully, refreshing user data...');
          await loadCurrentUser(forceRefresh: true);
          return true;
        } else {
          _error = response['error'] as String?;
          print('[AuthProvider] Profile update failed: $_error');
          notifyListeners();
          return false;
        }
      } else {
        print('[AuthProvider] Imgur upload failed: ${imgurResponse.body}');
        _setLoading(false);
        _error = 'Failed to upload image';
        notifyListeners();
        return false;
      }
    } catch (e) {
      print('[AuthProvider] Upload exception: $e');
      _setLoading(false);
      _error = e.toString();
      notifyListeners();
      return false;
    }
  }

  // Update Profile
  Future<bool> updateProfile({
    String? fullName,
    String? bio,
  }) async {
    if (_currentUser == null) {
      print('[AuthProvider] Cannot update profile: user is null');
      return false;
    }
    
    _setLoading(true);
    _error = null;

    try {
      print('[AuthProvider] Updating profile...');
      // Use the correct endpoint: PUT /api/profile
      final response = await _apiService.put('/api/profile', {
        if (fullName != null && fullName.isNotEmpty) 'full_name': fullName,
        if (bio != null) 'bio': bio,
      });
      print('[AuthProvider] Update response: $response');

      _setLoading(false);

      if (response['success'] == true) {
        print('[AuthProvider] Profile updated successfully, refreshing user data...');
        await loadCurrentUser(forceRefresh: true);
        return true;
      } else {
        _error = response['error'] as String?;
        print('[AuthProvider] Update failed: $_error');
        notifyListeners();
        return false;
      }
    } catch (e) {
      print('[AuthProvider] Update exception: $e');
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
    _error = null;
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
