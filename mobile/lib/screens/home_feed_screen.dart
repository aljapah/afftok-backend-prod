import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/offer.dart';
import '../services/offer_service.dart';
import '../widgets/offer_card.dart';
import '../widgets/side_action_bar.dart';
import '../utils/app_localizations.dart';
import '../providers/auth_provider.dart';
import 'profile_screen_enhanced.dart';
import 'teams_screen.dart';
import 'ai_assistant_screen.dart';
import 'promoter_public_page.dart';
import 'leaderboard_screen.dart';

class HomeFeedScreen extends StatefulWidget {
  const HomeFeedScreen({Key? key}) : super(key: key);

  @override
  State<HomeFeedScreen> createState() => _HomeFeedScreenState();
}

class _HomeFeedScreenState extends State<HomeFeedScreen> {
  final PageController _pageController = PageController();
  List<Offer> _allOffers = [];
  List<Offer> _filteredOffers = [];
  int _currentIndex = 0;
  int _selectedNavIndex = 0; // 0 = Teams, 1 = Leaderboard, 2 = AI, 3 = My Page, 4 = Profile
  bool _isLoading = true;
  String? _errorMessage;
  
  // Filter categories
  Set<String> _selectedCategories = {};
  List<String> _availableCategories = [];
  bool _showAllCategories = true;

  @override
  void initState() {
    super.initState();
    _checkAuth();
    _loadOffers();
  }
  
  void _checkAuth() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final authProvider = Provider.of<AuthProvider>(context, listen: false);
      if (!authProvider.isLoggedIn || authProvider.currentUser == null) {
        // Session expired or not logged in - go back to role selection
        Navigator.pushNamedAndRemoveUntil(context, '/', (route) => false);
      }
    });
  }

  Future<void> _loadOffers() async {
    setState(() => _isLoading = true);
    
    try {
      final offerService = OfferService();
      final result = await offerService.getAllOffers();
      
      setState(() {
        if (result['success'] == true) {
          final offers = result['offers'];
          
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
          
          // Extract unique categories
          _availableCategories = _allOffers
              .map((o) => o.category)
              .where((c) => c.isNotEmpty)
              .toSet()
              .toList();
          
          // Apply filter
          _applyFilter();
        } else {
          _allOffers = [];
          _filteredOffers = [];
          _errorMessage = result['error'] ?? 'Failed to load offers';
        }
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _allOffers = [];
        _filteredOffers = [];
        _errorMessage = 'Error: ${e.toString()}';
        _isLoading = false;
      });
      print('Error loading offers: $e');
    }
  }

  void _applyFilter() {
    setState(() {
      if (_showAllCategories || _selectedCategories.isEmpty) {
        _filteredOffers = List.from(_allOffers);
      } else {
        _filteredOffers = _allOffers
            .where((o) => _selectedCategories.contains(o.category))
            .toList();
      }
      _currentIndex = 0;
      if (_pageController.hasClients) {
        _pageController.jumpToPage(0);
      }
    });
  }

  void _showCategoryFilter() {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => StatefulBuilder(
        builder: (context, setModalState) => Container(
          height: MediaQuery.of(context).size.height * 0.6,
          decoration: const BoxDecoration(
            color: Color(0xFF1A1A1A),
            borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
          ),
          child: Column(
            children: [
              // Handle
              Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.only(top: 12, bottom: 16),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.3),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              
              // Title
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                        ),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Icon(Icons.filter_list, color: Colors.white, size: 24),
                    ),
                    const SizedBox(width: 12),
                    Text(
                      isArabic ? 'فلتر التصنيفات' : 'Filter Categories',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
              
              const SizedBox(height: 20),
              
              // All categories option
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: GestureDetector(
                  onTap: () {
                    setModalState(() {
                      _showAllCategories = true;
                      _selectedCategories.clear();
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                    decoration: BoxDecoration(
                      gradient: _showAllCategories
                          ? const LinearGradient(
                              colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                            )
                          : null,
                      color: _showAllCategories ? null : Colors.white.withOpacity(0.05),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: _showAllCategories 
                            ? Colors.transparent 
                            : Colors.white.withOpacity(0.1),
                      ),
                    ),
                    child: Row(
                      children: [
                        Icon(
                          _showAllCategories 
                              ? Icons.check_circle 
                              : Icons.circle_outlined,
                          color: Colors.white,
                          size: 24,
                        ),
                        const SizedBox(width: 12),
                        Text(
                          isArabic ? '🌟 عرض الجميع' : '🌟 Show All',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        const Spacer(),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                          decoration: BoxDecoration(
                            color: Colors.white.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Text(
                            '${_allOffers.length}',
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              
              const SizedBox(height: 12),
              
              // Divider
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Row(
                  children: [
                    Expanded(child: Divider(color: Colors.white.withOpacity(0.1))),
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 12),
                      child: Text(
                        isArabic ? 'أو اختر تصنيفات محددة' : 'Or select specific',
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.5),
                          fontSize: 12,
                        ),
                      ),
                    ),
                    Expanded(child: Divider(color: Colors.white.withOpacity(0.1))),
                  ],
                ),
              ),
              
              const SizedBox(height: 12),
              
              // Categories list
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  itemCount: _getMergedCategories().length,
                  itemBuilder: (context, index) {
                    final category = _getMergedCategories()[index];
                    final isSelected = _selectedCategories.contains(category);
                    final count = _allOffers.where((o) => o.category.toLowerCase() == category.toLowerCase()).length;
                    final icon = _getCategoryIcon(category);
                    final displayName = isArabic ? _getCategoryNameAr(category) : category;
                    
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: GestureDetector(
                        onTap: () {
                          setModalState(() {
                            _showAllCategories = false;
                            if (isSelected) {
                              _selectedCategories.remove(category);
                              if (_selectedCategories.isEmpty) {
                                _showAllCategories = true;
                              }
                            } else {
                              _selectedCategories.add(category);
                            }
                          });
                        },
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                          decoration: BoxDecoration(
                            color: isSelected 
                                ? const Color(0xFFFF006E).withOpacity(0.2)
                                : Colors.white.withOpacity(0.05),
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: isSelected 
                                  ? const Color(0xFFFF006E).withOpacity(0.5)
                                  : Colors.white.withOpacity(0.1),
                            ),
                          ),
                          child: Row(
                            children: [
                              Icon(
                                isSelected 
                                    ? Icons.check_box 
                                    : Icons.check_box_outline_blank,
                                color: isSelected 
                                    ? const Color(0xFFFF006E) 
                                    : Colors.white.withOpacity(0.5),
                                size: 24,
                              ),
                              const SizedBox(width: 12),
                              Text(icon, style: const TextStyle(fontSize: 20)),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  displayName,
                                  style: TextStyle(
                                    color: isSelected ? Colors.white : Colors.white.withOpacity(0.8),
                                    fontSize: 16,
                                    fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                              const SizedBox(width: 8),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                decoration: BoxDecoration(
                                  color: isSelected 
                                      ? const Color(0xFFFF006E).withOpacity(0.3)
                                      : Colors.white.withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: Text(
                                  '$count',
                                  style: TextStyle(
                                    color: isSelected ? Colors.white : Colors.white.withOpacity(0.6),
                                    fontSize: 14,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ),
              
              // Apply button
              Padding(
                padding: const EdgeInsets.all(20),
                child: Container(
                  width: double.infinity,
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                    ),
                    borderRadius: BorderRadius.circular(16),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFFFF006E).withOpacity(0.3),
                        blurRadius: 12,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: ElevatedButton(
                    onPressed: () {
                      Navigator.pop(context);
                      _applyFilter();
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.transparent,
                      shadowColor: Colors.transparent,
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.check, size: 20),
                        const SizedBox(width: 8),
                        Text(
                          isArabic ? 'تطبيق الفلتر' : 'Apply Filter',
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  // التصنيفات الـ 8 فقط
  static const List<Map<String, String>> _allCategories = [
    {'name': 'E-commerce', 'icon': '🛒', 'nameAr': 'تسوق'},
    {'name': 'Finance', 'icon': '💰', 'nameAr': 'مالية'},
    {'name': 'Travel', 'icon': '✈️', 'nameAr': 'سفر'},
    {'name': 'Food', 'icon': '🍔', 'nameAr': 'طعام وتوصيل'},
    {'name': 'Education', 'icon': '📚', 'nameAr': 'تعليم'},
    {'name': 'Entertainment', 'icon': '🎬', 'nameAr': 'ترفيه'},
    {'name': 'Crypto', 'icon': '₿', 'nameAr': 'عملات رقمية'},
    {'name': 'Services', 'icon': '🔧', 'nameAr': 'خدمات'},
  ];

  String _getCategoryIcon(String category) {
    final cat = _allCategories.firstWhere(
      (c) => c['name']!.toLowerCase() == category.toLowerCase(),
      orElse: () => {'icon': '📦'},
    );
    return cat['icon'] ?? '📦';
  }
  
  String _getCategoryNameAr(String category) {
    final cat = _allCategories.firstWhere(
      (c) => c['name']!.toLowerCase() == category.toLowerCase(),
      orElse: () => {'nameAr': category},
    );
    return cat['nameAr'] ?? category;
  }

  List<String> _getMergedCategories() {
    // Merge actual categories with all predefined categories
    final Set<String> merged = {};
    
    // Add predefined categories
    for (var cat in _allCategories) {
      merged.add(cat['name']!);
    }
    
    // Add actual categories from offers (in case there are new ones)
    merged.addAll(_availableCategories);
    
    return merged.toList()..sort();
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    
    if (_isLoading) {
      return Scaffold(
        backgroundColor: Colors.black,
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const CircularProgressIndicator(color: Color(0xFFFF006E)),
              const SizedBox(height: 16),
              Text(
                isArabic ? 'جاري تحميل العروض...' : 'Loading offers...',
                style: const TextStyle(color: Colors.white70),
              ),
            ],
          ),
        ),
      );
    }
    
    // إذا لم تكن هناك عروض
    if (_filteredOffers.isEmpty) {
      return Scaffold(
        backgroundColor: Colors.black,
        body: Stack(
          children: [
            // Background gradient
            Container(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [
                    const Color(0xFF1A1A1A),
                    Colors.black,
                    const Color(0xFF0D0D0D),
                  ],
                ),
              ),
            ),
            
            // Subtle pattern overlay
            Positioned.fill(
              child: CustomPaint(
                painter: _EmptyStatePainter(),
              ),
            ),
            
            // Content
            Center(
              child: Padding(
                padding: const EdgeInsets.all(32),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    // Animated empty icon
                    Container(
                      width: 120,
                      height: 120,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        gradient: RadialGradient(
                          colors: [
                            const Color(0xFFFF006E).withOpacity(0.2),
                            const Color(0xFFFF006E).withOpacity(0.05),
                            Colors.transparent,
                          ],
                        ),
                      ),
                      child: Center(
                        child: Container(
                          width: 80,
                          height: 80,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            color: Colors.white.withOpacity(0.05),
                            border: Border.all(
                              color: const Color(0xFFFF006E).withOpacity(0.3),
                              width: 2,
                            ),
                          ),
                          child: const Icon(
                            Icons.search_off_rounded,
                            size: 40,
                            color: Color(0xFFFF006E),
                          ),
                        ),
                      ),
                    ),
                    
                    const SizedBox(height: 32),
                    
                    // Title
                    Text(
                      isArabic ? 'لا توجد إعلانات' : 'No Offers Found',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 24,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    
                    const SizedBox(height: 12),
                    
                    // Subtitle
                    Text(
                      _selectedCategories.isNotEmpty
                          ? (isArabic 
                              ? 'لا توجد إعلانات في التصنيفات المحددة\nجرب اختيار تصنيفات أخرى'
                              : 'No offers in selected categories\nTry selecting different categories')
                          : (_errorMessage ?? (isArabic 
                              ? 'لا توجد إعلانات متاحة حالياً'
                              : 'No offers available at the moment')),
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        color: Colors.white.withOpacity(0.6),
                        fontSize: 16,
                        height: 1.5,
                      ),
                    ),
                    
                    const SizedBox(height: 32),
                    
                    // Selected categories chips
                    if (_selectedCategories.isNotEmpty) ...[
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        alignment: WrapAlignment.center,
                        children: _selectedCategories.map((cat) => Container(
                          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                          decoration: BoxDecoration(
                            color: const Color(0xFFFF006E).withOpacity(0.2),
                            borderRadius: BorderRadius.circular(20),
                            border: Border.all(
                              color: const Color(0xFFFF006E).withOpacity(0.5),
                            ),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Text(_getCategoryIcon(cat), style: const TextStyle(fontSize: 14)),
                              const SizedBox(width: 6),
                              Text(
                                isArabic ? _getCategoryNameAr(cat) : cat,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 12,
                                ),
                              ),
                            ],
                          ),
                        )).toList(),
                      ),
                      
                      const SizedBox(height: 24),
                      
                      // Clear filter button
                      Container(
                        decoration: BoxDecoration(
                          gradient: const LinearGradient(
                            colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                          ),
                          borderRadius: BorderRadius.circular(16),
                          boxShadow: [
                            BoxShadow(
                              color: const Color(0xFFFF006E).withOpacity(0.3),
                              blurRadius: 12,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            setState(() {
                              _showAllCategories = true;
                              _selectedCategories.clear();
                            });
                            _applyFilter();
                          },
                          icon: const Icon(Icons.clear_all, size: 20),
                          label: Text(
                            isArabic ? 'عرض جميع الإعلانات' : 'Show All Offers',
                            style: const TextStyle(fontWeight: FontWeight.bold),
                          ),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.transparent,
                            shadowColor: Colors.transparent,
                            foregroundColor: Colors.white,
                            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16),
                            ),
                          ),
                        ),
                      ),
                    ] else ...[
                      // Retry button
                      OutlinedButton.icon(
                        onPressed: _loadOffers,
                        icon: const Icon(Icons.refresh, size: 20),
                        label: Text(isArabic ? 'إعادة المحاولة' : 'Retry'),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.white,
                          side: BorderSide(color: Colors.white.withOpacity(0.3)),
                          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16),
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ],
        ),
        bottomNavigationBar: _buildBottomNav(lang, isArabic),
      );
    }
    
    return Scaffold(
      extendBodyBehindAppBar: true,
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          PageView.builder(
            controller: _pageController,
            scrollDirection: Axis.vertical,
            itemCount: _filteredOffers.length,
            onPageChanged: (index) {
              setState(() {
                _currentIndex = index;
              });
            },
            itemBuilder: (context, index) {
              if (index < _filteredOffers.length) {
                return OfferCard(offer: _filteredOffers[index]);
              }
              return const SizedBox.shrink();
            },
          ),
          
          // Filter indicator
          if (!_showAllCategories && _selectedCategories.isNotEmpty)
            Positioned(
              top: MediaQuery.of(context).padding.top + 10,
              left: 16,
              right: 16,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: Colors.black.withOpacity(0.7),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: const Color(0xFFFF006E).withOpacity(0.5),
                  ),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.filter_alt, color: Color(0xFFFF006E), size: 18),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        _selectedCategories.join(', '),
                        style: const TextStyle(color: Colors.white, fontSize: 12),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    GestureDetector(
                      onTap: () {
                        setState(() {
                          _showAllCategories = true;
                          _selectedCategories.clear();
                        });
                        _applyFilter();
                      },
                      child: const Icon(Icons.close, color: Colors.white54, size: 18),
                    ),
                  ],
                ),
              ),
            ),
          
          if (_filteredOffers.isNotEmpty)
            Builder(
              builder: (context) {
                final textDirection = Directionality.of(context);
                return Positioned(
                  right: textDirection == TextDirection.rtl ? 8 : 8,
                  left: textDirection == TextDirection.rtl ? null : null,
                  bottom: 150,
                  child: SideActionBar(
                    offer: _filteredOffers[_currentIndex],
                    onFilterTap: _showCategoryFilter,
                    selectedCategories: _selectedCategories,
                  ),
                );
              },
            ),
        ],
      ),
      bottomNavigationBar: _buildBottomNav(lang, isArabic),
    );
  }

  Widget _buildBottomNav(AppLocalizations lang, bool isArabic) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.black,
        boxShadow: [
          BoxShadow(
            color: Colors.white.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 8),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              // 1. Teams (الفرق) - Left
              _buildNavItem(
                index: 0,
                icon: Icon(
                  Icons.groups,
                  color: _selectedNavIndex == 0 ? const Color(0xFFFF006E) : Colors.white54,
                  size: 26,
                ),
                label: lang.teams,
                onTap: () {
                  setState(() => _selectedNavIndex = 0);
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const TeamsScreen()),
                  );
                },
              ),
              
              // 2. Leaderboard (المتصدرين)
              _buildNavItem(
                index: 1,
                icon: Icon(
                  Icons.leaderboard,
                  color: _selectedNavIndex == 1 ? const Color(0xFFFF006E) : Colors.white54,
                  size: 26,
                ),
                label: isArabic ? 'المتصدرين' : 'Top',
                onTap: () {
                  setState(() => _selectedNavIndex = 1);
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const LeaderboardScreen()),
                  );
                },
              ),
              
              // 3. AI Assistant - Glowing Space Orb (Center)
              _buildNavItem(
                index: 2,
                icon: _buildAIOrb(),
                label: isArabic ? 'المساعد' : 'AI',
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const AIAssistantScreen()),
                  );
                },
              ),
              
              // 4. My Page (صفحتي)
              _buildNavItem(
                index: 3,
                icon: Icon(
                  Icons.web,
                  color: _selectedNavIndex == 3 ? const Color(0xFFFF006E) : Colors.white54,
                  size: 26,
                ),
                label: isArabic ? 'صفحتي' : 'My Page',
                onTap: () {
                  setState(() => _selectedNavIndex = 3);
                  final authProvider = Provider.of<AuthProvider>(context, listen: false);
                  final username = authProvider.currentUser?.username ?? '';
                  if (username.isNotEmpty) {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => PromoterPublicPage(username: username),
                      ),
                    );
                  } else {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text(isArabic ? 'يرجى تسجيل الدخول أولاً' : 'Please login first')),
                    );
                  }
                },
              ),
              
              // 5. Profile (الملف الشخصي) - Right
              _buildNavItem(
                index: 4,
                icon: Icon(
                  Icons.person,
                  color: _selectedNavIndex == 4 ? const Color(0xFFFF006E) : Colors.white54,
                  size: 26,
                ),
                label: lang.profile,
                onTap: () {
                  setState(() => _selectedNavIndex = 4);
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const ProfileScreenEnhanced()),
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildNavItem({
    required int index,
    required Widget icon,
    required String label,
    required VoidCallback onTap,
    bool isSelected = false,
  }) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            icon,
            const SizedBox(height: 2),
            Text(
              label,
              style: TextStyle(
                color: isSelected ? const Color(0xFFFF006E) : Colors.white54,
                fontSize: 10,
                fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAIOrb() {
    return Container(
      width: 32,
      height: 32,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        gradient: const RadialGradient(
          colors: [
            Color(0xFFFF006E),
            Color(0xFFFF4D00),
            Color(0xFF8B0000),
          ],
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFFFF006E).withOpacity(0.6),
            blurRadius: 12,
            spreadRadius: 2,
          ),
          BoxShadow(
            color: const Color(0xFFFF4D00).withOpacity(0.4),
            blurRadius: 20,
            spreadRadius: 4,
          ),
        ],
      ),
      child: const Center(
        child: Icon(
          Icons.auto_awesome,
          color: Colors.white,
          size: 18,
        ),
      ),
    );
  }
}

// Custom painter for empty state background pattern
class _EmptyStatePainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = const Color(0xFFFF006E).withOpacity(0.03)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;
    
    // Draw subtle circles
    for (int i = 0; i < 5; i++) {
      final radius = 100.0 + (i * 80);
      canvas.drawCircle(
        Offset(size.width / 2, size.height / 2),
        radius,
        paint,
      );
    }
    
    // Draw diagonal lines
    final linePaint = Paint()
      ..color = Colors.white.withOpacity(0.02)
      ..strokeWidth = 1;
    
    for (double i = -size.height; i < size.width + size.height; i += 50) {
      canvas.drawLine(
        Offset(i, 0),
        Offset(i + size.height, size.height),
        linePaint,
      );
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
