import 'package:flutter/material.dart';
import '../models/offer.dart';
import '../services/offer_service.dart';
import '../widgets/offer_card.dart';
import '../widgets/side_action_bar.dart';
import '../utils/app_localizations.dart';
import 'discover_screen.dart';
import 'saved_screen.dart';
import 'profile_screen_enhanced.dart';
import 'teams_screen.dart';

class HomeFeedScreen extends StatefulWidget {
  const HomeFeedScreen({Key? key}) : super(key: key);

  @override
  State<HomeFeedScreen> createState() => _HomeFeedScreenState();
}

class _HomeFeedScreenState extends State<HomeFeedScreen> {
  final PageController _pageController = PageController();
  List<Offer> _offers = [];
  int _currentIndex = 0;
  int _selectedNavIndex = 0;
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadOffers();
  }

  Future<void> _loadOffers() async {
    setState(() => _isLoading = true);
    
    try {
      final offerService = OfferService();
      final result = await offerService.getAllOffers();
      
      setState(() {
        if (result['success'] == true) {
          final offers = result['offers'];
          
          // فحص إذا كانت القائمة null أو ليست List
          if (offers != null && offers is List<Offer>) {
            _offers = offers;
          } else if (offers is List) {
            // محاولة تحويل القائمة إذا كانت من نوع مختلف
            try {
              _offers = List<Offer>.from(offers);
            } catch (e) {
              print('Error converting offers: $e');
              _offers = [];
              _errorMessage = 'Error loading offers';
            }
          } else {
            _offers = [];
            _errorMessage = 'Invalid offers data';
          }
        } else {
          _offers = [];
          _errorMessage = result['error'] ?? 'Failed to load offers';
        }
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _offers = [];
        _errorMessage = 'Error: ${e.toString()}';
        _isLoading = false;
      });
      print('Error loading offers: $e');
    }
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    
    if (_isLoading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }
    
    // إذا لم تكن هناك عروض
    if (_offers.isEmpty) {
      return Scaffold(
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          actions: [
            IconButton(
              icon: const Icon(Icons.search, color: Colors.white),
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (context) => const DiscoverScreen()),
                );
              },
            ),
          ],
        ),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.inbox, size: 64, color: Colors.grey),
              const SizedBox(height: 16),
              Text(
                _errorMessage ?? 'No offers available',
                textAlign: TextAlign.center,
                style: const TextStyle(color: Colors.grey, fontSize: 16),
              ),
              const SizedBox(height: 24),
              ElevatedButton(
                onPressed: _loadOffers,
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        bottomNavigationBar: BottomNavigationBar(
          backgroundColor: Colors.black,
          selectedItemColor: Colors.white,
          unselectedItemColor: Colors.white54,
          type: BottomNavigationBarType.fixed,
          currentIndex: _selectedNavIndex,
          onTap: (index) {
            setState(() {
              _selectedNavIndex = index;
            });
            
            if (index == 1) {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => const DiscoverScreen()),
              );
            } else if (index == 2) {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => const TeamsScreen()),
              );
            } else if (index == 3) {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => const SavedScreen()),
              );
            } else if (index == 4) {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => const ProfileScreenEnhanced()),
              );
            }
          },
          items: [
            BottomNavigationBarItem(
              icon: const Icon(Icons.home),
              label: lang.home,
            ),
            BottomNavigationBarItem(
              icon: const Icon(Icons.explore),
              label: lang.discover,
            ),
            BottomNavigationBarItem(
              icon: const Icon(Icons.groups),
              label: lang.teams,
            ),
            BottomNavigationBarItem(
              icon: const Icon(Icons.favorite_border),
              label: lang.saved,
            ),
            BottomNavigationBarItem(
              icon: const Icon(Icons.person_outline),
              label: lang.profile,
            ),
          ],
        ),
      );
    }
    
    return Scaffold(
      extendBodyBehindAppBar: true,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.search, color: Colors.white),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => const DiscoverScreen()),
              );
            },
          ),
        ],
      ),
      body: Stack(
        children: [
          PageView.builder(
            controller: _pageController,
            scrollDirection: Axis.vertical,
            itemCount: _offers.length,
            onPageChanged: (index) {
              setState(() {
                _currentIndex = index;
              });
            },
            itemBuilder: (context, index) {
              if (index < _offers.length) {
                return OfferCard(offer: _offers[index]);
              }
              return const SizedBox.shrink();
            },
          ),
          
          if (_offers.isNotEmpty)
            Builder(
              builder: (context) {
                final textDirection = Directionality.of(context);
                return Positioned(
                  right: textDirection == TextDirection.rtl ? 8 : 8,
                  left: textDirection == TextDirection.rtl ? null : null,
                  bottom: 280,
                  child: SideActionBar(
                    offer: _offers[_currentIndex],
                  ),
                );
              },
            ),
        ],
      ),
      bottomNavigationBar: BottomNavigationBar(
        backgroundColor: Colors.black,
        selectedItemColor: Colors.white,
        unselectedItemColor: Colors.white54,
        type: BottomNavigationBarType.fixed,
        currentIndex: _selectedNavIndex,
        onTap: (index) {
          setState(() {
            _selectedNavIndex = index;
          });
          
          if (index == 1) {
            Navigator.push(
              context,
              MaterialPageRoute(builder: (context) => const DiscoverScreen()),
            );
          } else if (index == 2) {
            Navigator.push(
              context,
              MaterialPageRoute(builder: (context) => const TeamsScreen()),
            );
          } else if (index == 3) {
            Navigator.push(
              context,
              MaterialPageRoute(builder: (context) => const SavedScreen()),
            );
          } else if (index == 4) {
            Navigator.push(
              context,
              MaterialPageRoute(builder: (context) => const ProfileScreenEnhanced()),
            );
          }
        },
        items: [
          BottomNavigationBarItem(
            icon: const Icon(Icons.home),
            label: lang.home,
          ),
          BottomNavigationBarItem(
            icon: const Icon(Icons.explore),
            label: lang.discover,
          ),
          BottomNavigationBarItem(
            icon: const Icon(Icons.groups),
            label: lang.teams,
          ),
          BottomNavigationBarItem(
            icon: const Icon(Icons.favorite_border),
            label: lang.saved,
          ),
          BottomNavigationBarItem(
            icon: const Icon(Icons.person_outline),
            label: lang.profile,
          ),
        ],
      ),
    );
  }
}
