import 'package:flutter/foundation.dart';
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

  // Initialize
  Future<void> init() async {
    print('[AuthProvider] Initializing...');
    await _apiService.init();
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

  Future<void> loadCurrentUser() async {
    _setLoading(true);
    _error = null;
    
    try {
      print('[AuthProvider] Loading current user...');
      final result = await _apiService.getMe();
      print('[AuthProvider] API Response: $result');
      
      if (result['success'] == true && result['user'] != null) {
        _currentUser = result['user'] as User;
        _error = null;
        print('[AuthProvider] User loaded: ${_currentUser?.username}');
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
  
  Future<void> _loadUserOffers() async {
    try {
      final result = await _offerService.getMyOffers();
      if (result['success'] == true) {
        final offers = result['offers'];
        if (offers != null && offers is List<UserOffer>) {
          _userOffers = offers;
        } else if (offers is List) {
          try {
            _userOffers = List<UserOffer>.from(offers);
          } catch (e) {
            _userOffers = [];
          }
        } else {
          _userOffers = [];
        }
      } else {
        _userOffers = [];
      }
      notifyListeners();
    } catch (e) {
      _userOffers = [];
      print('Error loading user offers: $e');
    }
  }
  
  Future<bool> addOfferToProfile(String offerId) async {
    if (_currentUser == null) return false;
    
    try {
      final result = await _offerService.joinOffer(offerId);
      if (result['success'] == true) {
        await _loadUserOffers();
        return true;
      }
      return false;
    } catch (e) {
      print('Error adding offer: $e');
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

  // Upload Profile Image
  Future<bool> uploadProfileImage(String imagePath) async {
    if (_currentUser == null) return false;
    
    _setLoading(true);
    _error = null;

    try {
      final response = await _apiService.uploadFile(
        '/users/${_currentUser!.id}/avatar',
        imagePath,
      );

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
