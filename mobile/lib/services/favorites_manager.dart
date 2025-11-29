import 'package:shared_preferences/shared_preferences.dart';
import 'dart:convert';
import '../models/offer.dart';

class FavoritesManager {
  static const String _key = 'saved_offers';
  static FavoritesManager? _instance;
  static SharedPreferences? _prefs;

  FavoritesManager._();

  static Future<FavoritesManager> getInstance() async {
    if (_instance == null) {
      _instance = FavoritesManager._();
      _prefs = await SharedPreferences.getInstance();
    }
    return _instance!;
  }

  // حفظ عرض في المفضلة
  Future<bool> saveOffer(Offer offer) async {
    try {
      if (_prefs == null) {
        _prefs = await SharedPreferences.getInstance();
      }
      
      final savedOffers = await getSavedOffers();
      
      // تحقق إذا كان العرض موجود مسبقاً
      if (savedOffers.any((o) => o.id == offer.id)) {
        return false; // العرض موجود مسبقاً
      }
      
      savedOffers.add(offer);
      
      final List<String> offerJsonList = savedOffers
          .map((o) => json.encode(o.toJson()))
          .toList();
      
      return await _prefs!.setStringList(_key, offerJsonList);
    } catch (e) {
      print('Error saving offer: $e');
      return false;
    }
  }

  // حذف عرض من المفضلة
  Future<bool> removeOffer(String offerId) async {
    try {
      if (_prefs == null) {
        _prefs = await SharedPreferences.getInstance();
      }
      
      final savedOffers = await getSavedOffers();
      
      if (savedOffers.isEmpty) {
        return false;
      }
      
      savedOffers.removeWhere((o) => o.id == offerId);
      
      final List<String> offerJsonList = savedOffers
          .map((o) => json.encode(o.toJson()))
          .toList();
      
      return await _prefs!.setStringList(_key, offerJsonList);
    } catch (e) {
      print('Error removing offer: $e');
      return false;
    }
  }

  // جلب جميع العروض المحفوظة
  Future<List<Offer>> getSavedOffers() async {
    try {
      if (_prefs == null) {
        _prefs = await SharedPreferences.getInstance();
      }
      
      final List<String>? offerJsonList = _prefs!.getStringList(_key);
      
      if (offerJsonList == null || offerJsonList.isEmpty) {
        return [];
      }
      
      final List<Offer> offers = [];
      for (var jsonStr in offerJsonList) {
        try {
          final offer = Offer.fromJson(json.decode(jsonStr));
          offers.add(offer);
        } catch (e) {
          print('Error parsing offer: $e');
          continue;
        }
      }
      
      return offers;
    } catch (e) {
      print('Error getting saved offers: $e');
      return [];
    }
  }

  // التحقق إذا كان العرض محفوظ
  Future<bool> isOfferSaved(String offerId) async {
    try {
      final savedOffers = await getSavedOffers();
      return savedOffers.any((o) => o.id == offerId);
    } catch (e) {
      print('Error checking if offer is saved: $e');
      return false;
    }
  }

  // مسح جميع المفضلة
  Future<bool> clearAll() async {
    try {
      if (_prefs == null) {
        _prefs = await SharedPreferences.getInstance();
      }
      
      return await _prefs!.remove(_key);
    } catch (e) {
      print('Error clearing favorites: $e');
      return false;
    }
  }

  // عدد العروض المحفوظة
  Future<int> getSavedCount() async {
    final savedOffers = await getSavedOffers();
    return savedOffers.length;
  }
}
