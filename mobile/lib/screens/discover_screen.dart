import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../utils/app_localizations.dart';
import '../models/offer.dart';
import '../services/offer_service.dart';

class DiscoverScreen extends StatefulWidget {
  const DiscoverScreen({Key? key}) : super(key: key);

  @override
  State<DiscoverScreen> createState() => _DiscoverScreenState();
}

class _DiscoverScreenState extends State<DiscoverScreen> {
  String _selectedCategory = 'All';
  List<Offer> _allOffers = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadOffers();
  }

  Future<void> _loadOffers() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    
    try {
      final offerService = OfferService();
      final category = _selectedCategory == 'All' ? null : _selectedCategory;
      final result = await offerService.getAllOffers(category: category);
      
      if (mounted) {
        setState(() {
          if (result['success'] == true) {
            final offers = result['offers'];
            
            // فحص إذا كانت القائمة null أو ليست List
            if (offers != null && offers is List<Offer>) {
              _allOffers = offers;
            } else if (offers is List) {
              try {
                _allOffers = List<Offer>.from(offers);
              } catch (e) {
                print('Error converting offers: $e');
                _allOffers = [];
                _errorMessage = 'Error loading offers';
              }
            } else {
              _allOffers = [];
              _errorMessage = 'Invalid offers data';
            }
          } else {
            _allOffers = [];
            _errorMessage = result['error'] as String? ?? 'Failed to load offers';
          }
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _allOffers = [];
          _errorMessage = 'Error: ${e.toString()}';
          _isLoading = false;
        });
      }
      print('Error loading offers: $e');
    }
  }

  List<Offer> get _filteredOffers {
    return _allOffers;
  }

  Future<void> _launchUrl(String url) async {
    try {
      final Uri uri = Uri.parse(url);
      if (!await launchUrl(uri, mode: LaunchMode.externalApplication)) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Could not launch $url')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    
    final categories = [
      {'key': 'All', 'label': lang.all, 'icon': Icons.apps},
      {'key': 'Cashback', 'label': lang.cashback, 'icon': Icons.monetization_on},
      {'key': 'E-commerce', 'label': lang.ecommerce, 'icon': Icons.shopping_bag},
      {'key': 'Finance', 'label': lang.finance, 'icon': Icons.account_balance},
      {'key': 'Travel', 'label': lang.travel, 'icon': Icons.flight},
      {'key': 'Education', 'label': lang.education, 'icon': Icons.school},
      {'key': 'Technology', 'label': lang.technology, 'icon': Icons.devices},
      {'key': 'Utilities', 'label': lang.utilities, 'icon': Icons.lightbulb_outline},
      {'key': 'Food & Restaurants', 'label': lang.foodRestaurants, 'icon': Icons.restaurant},
    ];

    if (_isLoading) {
      return Scaffold(
        backgroundColor: Colors.black,
        appBar: AppBar(
          backgroundColor: Colors.black,
          title: Text(
            lang.discoverTitle,
            style: const TextStyle(color: Colors.white),
          ),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back, color: Colors.white),
            onPressed: () => Navigator.pop(context),
          ),
        ),
        body: const Center(
          child: CircularProgressIndicator(color: Color(0xFFFF006E)),
        ),
      );
    }

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: Text(
          lang.discoverTitle,
          style: const TextStyle(color: Colors.white),
        ),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: Column(
        children: [
          SizedBox(
            height: 50,
            child: ListView.builder(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 16),
              itemCount: categories.length,
              itemBuilder: (context, index) {
                final category = categories[index];
                final isSelected = _selectedCategory == category['key'];
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: FilterChip(
                    label: Text(category['label'] as String),
                    selected: isSelected,
                    onSelected: (selected) {
                      if (selected) {
                        setState(() {
                          _selectedCategory = category['key'] as String;
                        });
                        _loadOffers();
                      }
                    },
                    backgroundColor: Colors.grey[900],
                    selectedColor: const Color(0xFFFF006E).withOpacity(0.3),
                    checkmarkColor: Colors.white,
                    labelStyle: TextStyle(
                      color: isSelected ? Colors.white : Colors.white70,
                    ),
                    side: BorderSide(
                      color: isSelected ? const Color(0xFFFF006E) : Colors.grey[700]!,
                    ),
                  ),
                );
              },
            ),
          ),
          
          const SizedBox(height: 8),
          
          if (_errorMessage != null)
            Padding(
              padding: const EdgeInsets.all(16),
              child: Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.red.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.red),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.error_outline, color: Colors.red),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        _errorMessage!,
                        style: const TextStyle(color: Colors.white),
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.red),
                      onPressed: () {
                        setState(() => _errorMessage = null);
                      },
                    ),
                  ],
                ),
              ),
            ),
          
          Expanded(
            child: _filteredOffers.isEmpty
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.search_off,
                          size: 64,
                          color: Colors.grey[700],
                        ),
                        const SizedBox(height: 16),
                        Text(
                          'لا توجد عروض في هذه الفئة',
                          style: TextStyle(
                            color: Colors.grey[600],
                            fontSize: 16,
                          ),
                        ),
                        const SizedBox(height: 24),
                        ElevatedButton(
                          onPressed: _loadOffers,
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFFFF006E),
                          ),
                          child: const Text('إعادة محاولة'),
                        ),
                      ],
                    ),
                  )
                : GridView.builder(
                    padding: const EdgeInsets.all(16),
                    gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 2,
                      childAspectRatio: 0.75,
                      crossAxisSpacing: 12,
                      mainAxisSpacing: 12,
                    ),
                    itemCount: _filteredOffers.length,
                    itemBuilder: (context, index) {
                      if (index >= _filteredOffers.length) {
                        return const SizedBox.shrink();
                      }
                      
                      final offer = _filteredOffers[index];
                      return GestureDetector(
                        onTap: () => _launchUrl(offer.offerUrl),
                        child: Container(
                          decoration: BoxDecoration(
                            color: Colors.grey[900],
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Expanded(
                                child: Stack(
                                  children: [
                                    Container(
                                      decoration: BoxDecoration(
                                        borderRadius: const BorderRadius.vertical(
                                          top: Radius.circular(12),
                                        ),
                                        color: Colors.grey[800],
                                        image: offer.imageUrl.isNotEmpty
                                            ? DecorationImage(
                                                image: NetworkImage(offer.imageUrl),
                                                fit: BoxFit.cover,
                                              )
                                            : null,
                                      ),
                                      child: offer.imageUrl.isEmpty
                                          ? const Center(
                                              child: Icon(Icons.image, color: Colors.grey, size: 50),
                                            )
                                          : null,
                                    ),
                                    Container(
                                      decoration: BoxDecoration(
                                        borderRadius: const BorderRadius.vertical(
                                          top: Radius.circular(12),
                                        ),
                                        gradient: LinearGradient(
                                          begin: Alignment.topCenter,
                                          end: Alignment.bottomCenter,
                                          colors: [
                                            Colors.transparent,
                                            Colors.black.withOpacity(0.7),
                                          ],
                                        ),
                                      ),
                                    ),
                                    Positioned(
                                      top: 8,
                                      left: 8,
                                      child: Container(
                                        width: 40,
                                        height: 40,
                                        decoration: BoxDecoration(
                                          color: Colors.white,
                                          borderRadius: BorderRadius.circular(8),
                                          boxShadow: [
                                            BoxShadow(
                                              color: Colors.black.withOpacity(0.3),
                                              blurRadius: 8,
                                            ),
                                          ],
                                        ),
                                        padding: const EdgeInsets.all(6),
                                        child: offer.logoUrl.isNotEmpty
                                            ? Image.network(
                                                offer.logoUrl,
                                                errorBuilder: (context, error, stackTrace) {
                                                  return Center(
                                                    child: Text(
                                                      offer.companyName.isNotEmpty
                                                          ? offer.companyName[0]
                                                          : '?',
                                                      style: const TextStyle(
                                                        fontSize: 16,
                                                        fontWeight: FontWeight.bold,
                                                        color: Colors.black87,
                                                      ),
                                                    ),
                                                  );
                                                },
                                              )
                                            : Center(
                                                child: Text(
                                                  offer.companyName.isNotEmpty
                                                      ? offer.companyName[0]
                                                      : '?',
                                                  style: const TextStyle(
                                                    fontSize: 16,
                                                    fontWeight: FontWeight.bold,
                                                    color: Colors.black87,
                                                  ),
                                                ),
                                              ),
                                      ),
                                    ),
                                    Positioned(
                                      bottom: 8,
                                      left: 8,
                                      child: Container(
                                        padding: const EdgeInsets.symmetric(
                                          horizontal: 8,
                                          vertical: 4,
                                        ),
                                        decoration: BoxDecoration(
                                          color: Colors.black.withOpacity(0.6),
                                          borderRadius: BorderRadius.circular(8),
                                          border: Border.all(
                                            color: Colors.white.withOpacity(0.3),
                                          ),
                                        ),
                                        child: Text(
                                          offer.category,
                                          style: const TextStyle(
                                            color: Colors.white,
                                            fontSize: 10,
                                            fontWeight: FontWeight.w500,
                                          ),
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              Padding(
                                padding: const EdgeInsets.all(12),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      offer.companyName,
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                      style: const TextStyle(
                                        color: Colors.white,
                                        fontWeight: FontWeight.bold,
                                        fontSize: 12,
                                      ),
                                    ),
                                    const SizedBox(height: 4),
                                    Text(
                                      offer.reward,
                                      style: const TextStyle(
                                        color: Color(0xFFFF006E),
                                        fontWeight: FontWeight.bold,
                                        fontSize: 11,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}
