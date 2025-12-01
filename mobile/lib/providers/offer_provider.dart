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

    try {
      final result = await _offerService.getAllOffers(
        status: status,
        category: category,
        sort: sort,
        order: order,
      );

      _setLoading(false);

      if (result['success'] == true) {
        final offers = result['offers'];
        
        // فحص إذا كانت القائمة null أو ليست List
        if (offers != null && offers is List<Offer>) {
          _offers = offers;
        } else if (offers is List) {
          try {
            _offers = List<Offer>.from(offers);
          } catch (e) {
            print('Error converting offers: $e');
            _offers = [];
            _error = 'Error loading offers';
          }
        } else {
          _offers = [];
          _error = 'Invalid offers data';
        }
        notifyListeners();
      } else {
        _error = result['error'] as String? ?? 'Failed to load offers';
        _offers = [];
        notifyListeners();
      }
    } catch (e) {
      _setLoading(false);
      _error = 'Error: ${e.toString()}';
      _offers = [];
      notifyListeners();
      print('Error loading offers: $e');
    }
  }

  // Load my offers
  Future<void> loadMyOffers() async {
    try {
      final result = await _offerService.getMyOffers();

      if (result['success'] == true) {
        final offers = result['offers'];
        
        // فحص إذا كانت القائمة null أو ليست List
        if (offers != null && offers is List<UserOffer>) {
          _myOffers = offers;
        } else if (offers is List) {
          try {
            _myOffers = List<UserOffer>.from(offers);
          } catch (e) {
            print('Error converting my offers: $e');
            _myOffers = [];
          }
        } else {
          _myOffers = [];
        }
        notifyListeners();
      } else {
        _myOffers = [];
      }
    } catch (e) {
      _myOffers = [];
      print('Error loading my offers: $e');
    }
  }

  Future<Map<String, dynamic>> joinOffer(String offerId) async {
    try {
      final result = await _offerService.joinOffer(offerId);
      
      if (result['success'] == true) {
        await loadMyOffers();
      }
      
      return result;
    } catch (e) {
      return {
        'success': false,
        'error': 'Error joining offer: ${e.toString()}',
      };
    }
  }
  
  Future<void> refreshMyOffers() async {
    await loadMyOffers();
  }
  
  void notifyOffersChanged() {
    notifyListeners();
  }

  // Check if user has joined offer
  bool hasJoinedOffer(String offerId) {
    try {
      return _myOffers.any((uo) => uo.offerId == offerId);
    } catch (e) {
      print('Error checking if offer is joined: $e');
      return false;
    }
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
