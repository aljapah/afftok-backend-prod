import 'package:flutter/foundation.dart';
import '../models/offer.dart';
import '../models/user_offer.dart';
import '../services/offer_service.dart';

class OfferProvider with ChangeNotifier {
  final OfferService _offerService = OfferService();
  
  List<Offer> _offers = [];
  List<UserOffer> _myOffers = [];
  bool _isLoading = false;
  String? _error;

  List<Offer> get offers => _offers;
  List<UserOffer> get myOffers => _myOffers;
  bool get isLoading => _isLoading;
  String? get error => _error;

  // Load all offers
  Future<void> loadOffers({
    String? status,
    String? category,
    String? sort,
    String? order,
  }) async {
    _setLoading(true);
    _error = null;

    final result = await _offerService.getAllOffers(
      status: status,
      category: category,
      sort: sort,
      order: order,
    );

    _setLoading(false);

    if (result['success'] == true) {
      _offers = result['offers'] as List<Offer>;
      notifyListeners();
    } else {
      _error = result['error'] as String;
      notifyListeners();
    }
  }

  // Load my offers
  Future<void> loadMyOffers() async {
    final result = await _offerService.getMyOffers();

    if (result['success'] == true) {
      _myOffers = result['offers'] as List<UserOffer>;
      notifyListeners();
    }
  }

  // Join offer
  Future<Map<String, dynamic>> joinOffer(String offerId) async {
    final result = await _offerService.joinOffer(offerId);
    
    if (result['success'] == true) {
      // Reload my offers
      await loadMyOffers();
    }
    
    return result;
  }

  // Check if user has joined offer
  bool hasJoinedOffer(String offerId) {
    return _myOffers.any((uo) => uo.offerId == offerId);
  }

  // Get user offer by offer ID
  UserOffer? getUserOffer(String offerId) {
    try {
      return _myOffers.firstWhere((uo) => uo.offerId == offerId);
    } catch (e) {
      return null;
    }
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
