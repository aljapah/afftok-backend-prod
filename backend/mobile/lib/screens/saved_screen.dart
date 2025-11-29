import 'package:flutter/material.dart';
import '../utils/app_localizations.dart';
import '../models/offer.dart';
import '../services/favorites_manager.dart';
import '../widgets/offer_card.dart';

class SavedScreen extends StatefulWidget {
  const SavedScreen({Key? key}) : super(key: key);

  @override
  State<SavedScreen> createState() => _SavedScreenState();
}

class _SavedScreenState extends State<SavedScreen> {
  List<Offer> _savedOffers = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadSavedOffers();
  }

  Future<void> _loadSavedOffers() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final favManager = await FavoritesManager.getInstance();
      final offers = await favManager.getSavedOffers();
      
      if (mounted) {
        setState(() {
          _savedOffers = offers;
          _isLoading = false;
        });
      }
    } catch (e) {
      print('Error loading saved offers: $e');
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _toggleFavorite(Offer offer) async {
    try {
      final favManager = await FavoritesManager.getInstance();
      if (await favManager.isOfferSaved(offer.id)) {
        await favManager.removeOffer(offer.id);
      } else {
        await favManager.saveOffer(offer);
      }
      await _loadSavedOffers();
    } catch (e) {
      print('Error toggling favorite: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: Text(
          lang.savedTitle,
          style: const TextStyle(color: Colors.white),
        ),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
        actions: [
          if (_savedOffers.isNotEmpty)
            IconButton(
              icon: const Icon(Icons.refresh, color: Colors.white),
              onPressed: _loadSavedOffers,
            ),
        ],
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(
                color: Colors.white,
              ),
            )
          : _savedOffers.isEmpty
              ? _buildEmptyState(lang)
              : _buildOffersList(),
    );
  }

  Widget _buildEmptyState(AppLocalizations lang) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.favorite_border,
            size: 100,
            color: Colors.white.withOpacity(0.3),
          ),
          const SizedBox(height: 24),
          Text(
            lang.noSavedOffers,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            lang.startSaving,
            style: TextStyle(
              color: Colors.white.withOpacity(0.6),
              fontSize: 16,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildOffersList() {
    return RefreshIndicator(
      onRefresh: _loadSavedOffers,
      color: const Color(0xFF8B5CF6),
      backgroundColor: Colors.black,
      child: PageView.builder(
        scrollDirection: Axis.vertical,
        itemCount: _savedOffers.length,
        itemBuilder: (context, index) {
          return OfferCard(
            offer: _savedOffers[index],
            onFavoriteChanged: () => _toggleFavorite(_savedOffers[index]),
          );
        },
      ),
    );
  }
}
