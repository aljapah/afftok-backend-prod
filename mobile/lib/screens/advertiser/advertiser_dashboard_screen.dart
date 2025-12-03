import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/auth_provider.dart';
import '../../services/advertiser_service.dart';
import '../role_selection_screen.dart';
import 'create_offer_screen.dart';

class AdvertiserDashboardScreen extends StatefulWidget {
  const AdvertiserDashboardScreen({super.key});

  @override
  State<AdvertiserDashboardScreen> createState() => _AdvertiserDashboardScreenState();
}

class _AdvertiserDashboardScreenState extends State<AdvertiserDashboardScreen> {
  final AdvertiserService _advertiserService = AdvertiserService();
  Map<String, dynamic>? _dashboard;
  List<dynamic> _offers = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadDashboard();
  }

  Future<void> _loadDashboard() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final authProvider = Provider.of<AuthProvider>(context, listen: false);
      final dashboard = await _advertiserService.getDashboard(authProvider.token!);
      final offersResponse = await _advertiserService.getMyOffers(authProvider.token!);
      
      setState(() {
        _dashboard = dashboard;
        _offers = offersResponse['offers'] ?? [];
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    final authProvider = Provider.of<AuthProvider>(context);
    final user = authProvider.user;

    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          // Background "AffTok" shadow text
          Positioned.fill(
            child: Center(
              child: Transform.rotate(
                angle: -0.2,
                child: Text(
                  'AffTok',
                  style: TextStyle(
                    fontSize: 120,
                    fontWeight: FontWeight.w900,
                    color: Colors.white.withOpacity(0.03),
                    letterSpacing: 8,
                  ),
                ),
              ),
            ),
          ),
          Positioned(
            top: MediaQuery.of(context).size.height * 0.15,
            left: -50,
            child: Text(
              'AffTok',
              style: TextStyle(
                fontSize: 80,
                fontWeight: FontWeight.w900,
                color: Colors.white.withOpacity(0.02),
                letterSpacing: 4,
              ),
            ),
          ),
          Positioned(
            bottom: MediaQuery.of(context).size.height * 0.1,
            right: -30,
            child: Text(
              'AffTok',
              style: TextStyle(
                fontSize: 60,
                fontWeight: FontWeight.w900,
                color: Colors.white.withOpacity(0.02),
                letterSpacing: 4,
              ),
            ),
          ),
          // Main Content
          SafeArea(
            child: RefreshIndicator(
            onRefresh: _loadDashboard,
            child: CustomScrollView(
              slivers: [
                // App Bar
                SliverAppBar(
                  backgroundColor: Colors.transparent,
                  floating: true,
                  title: Text(
                    isArabic ? 'لوحة تحكم المعلن' : 'Advertiser Dashboard',
                    style: const TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  actions: [
                    IconButton(
                      icon: const Icon(Icons.logout, color: Colors.white),
                      onPressed: () async {
                        await authProvider.logout();
                        if (mounted) {
                          Navigator.pushAndRemoveUntil(
                            context,
                            MaterialPageRoute(builder: (context) => const RoleSelectionScreen()),
                            (route) => false,
                          );
                        }
                      },
                    ),
                  ],
                ),
                
                SliverPadding(
                  padding: const EdgeInsets.all(16),
                  sliver: SliverList(
                    delegate: SliverChildListDelegate([
                      // Welcome Card
                      _buildWelcomeCard(user?.companyName ?? user?.fullName ?? 'Advertiser', isArabic),
                      
                      const SizedBox(height: 20),
                      
                      // Stats Cards
                      if (_isLoading)
                        const Center(child: CircularProgressIndicator())
                      else if (_error != null)
                        _buildErrorWidget(isArabic)
                      else ...[
                        _buildStatsGrid(isArabic),
                        
                        const SizedBox(height: 24),
                        
                        // Quick Actions
                        _buildQuickActions(isArabic),
                        
                        const SizedBox(height: 24),
                        
                        // My Offers Section
                        _buildOffersSection(isArabic),
                      ],
                    ]),
                  ),
                ),
              ],
            ),
          ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () async {
          final result = await Navigator.push(
            context,
            MaterialPageRoute(builder: (context) => const CreateOfferScreen()),
          );
          if (result == true) {
            _loadDashboard();
          }
        },
        backgroundColor: const Color(0xFF6C63FF),
        icon: const Icon(Icons.add),
        label: Text(isArabic ? 'إضافة عرض' : 'Add Offer'),
      ),
    );
  }

  Widget _buildWelcomeCard(String name, bool isArabic) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF6C63FF), Color(0xFF9D4EDD)],
        ),
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF6C63FF).withOpacity(0.3),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 60,
            height: 60,
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.2),
              shape: BoxShape.circle,
            ),
            child: const Icon(Icons.business, color: Colors.white, size: 30),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  isArabic ? 'مرحباً،' : 'Welcome,',
                  style: TextStyle(
                    color: Colors.white.withOpacity(0.8),
                    fontSize: 14,
                  ),
                ),
                Text(
                  name,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 22,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatsGrid(bool isArabic) {
    final offers = _dashboard?['offers'] ?? {};
    final stats = _dashboard?['stats'] ?? {};

    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 12,
      crossAxisSpacing: 12,
      childAspectRatio: 1.5,
      children: [
        _buildStatCard(
          isArabic ? 'العروض النشطة' : 'Active Offers',
          '${offers['active'] ?? 0}',
          Icons.check_circle_outline,
          const Color(0xFF4CAF50),
        ),
        _buildStatCard(
          isArabic ? 'قيد المراجعة' : 'Pending',
          '${offers['pending'] ?? 0}',
          Icons.hourglass_empty,
          const Color(0xFFFF9800),
        ),
        _buildStatCard(
          isArabic ? 'إجمالي النقرات' : 'Total Clicks',
          '${stats['total_clicks'] ?? 0}',
          Icons.touch_app,
          const Color(0xFF2196F3),
        ),
        _buildStatCard(
          isArabic ? 'إجمالي التحويلات' : 'Conversions',
          '${stats['total_conversions'] ?? 0}',
          Icons.trending_up,
          const Color(0xFFE91E63),
        ),
      ],
    );
  }

  Widget _buildStatCard(String title, String value, IconData icon, Color color) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Icon(icon, color: color, size: 24),
          const SizedBox(height: 8),
          Text(
            value,
            style: TextStyle(
              color: color,
              fontSize: 24,
              fontWeight: FontWeight.bold,
            ),
          ),
          Text(
            title,
            style: TextStyle(
              color: Colors.white.withOpacity(0.7),
              fontSize: 12,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildQuickActions(bool isArabic) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          isArabic ? 'إجراءات سريعة' : 'Quick Actions',
          style: const TextStyle(
            color: Colors.white,
            fontSize: 18,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: _buildActionButton(
                isArabic ? 'إضافة عرض جديد' : 'Add New Offer',
                Icons.add_circle_outline,
                const Color(0xFF6C63FF),
                () async {
                  final result = await Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const CreateOfferScreen()),
                  );
                  if (result == true) {
                    _loadDashboard();
                  }
                },
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _buildActionButton(
                isArabic ? 'تحديث البيانات' : 'Refresh Data',
                Icons.refresh,
                const Color(0xFF4CAF50),
                _loadDashboard,
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildActionButton(String title, IconData icon, Color color, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: color.withOpacity(0.1),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: color.withOpacity(0.3)),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, color: color, size: 20),
            const SizedBox(width: 8),
            Flexible(
              child: Text(
                title,
                style: TextStyle(
                  color: color,
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                ),
                textAlign: TextAlign.center,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildOffersSection(bool isArabic) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              isArabic ? 'عروضي' : 'My Offers',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(
              '${_offers.length} ${isArabic ? 'عرض' : 'offers'}',
              style: TextStyle(
                color: Colors.white.withOpacity(0.5),
                fontSize: 14,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        
        if (_offers.isEmpty)
          _buildEmptyOffersWidget(isArabic)
        else
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _offers.length,
            separatorBuilder: (_, __) => const SizedBox(height: 12),
            itemBuilder: (context, index) {
              final offer = _offers[index];
              return _buildOfferCard(offer, isArabic);
            },
          ),
      ],
    );
  }

  Widget _buildOfferCard(Map<String, dynamic> offer, bool isArabic) {
    final status = offer['status'] ?? 'pending';
    final statusColor = _getStatusColor(status);
    final statusText = _getStatusText(status, isArabic);

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white.withOpacity(0.1)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              // Offer Image/Logo
              Container(
                width: 50,
                height: 50,
                decoration: BoxDecoration(
                  color: const Color(0xFF6C63FF).withOpacity(0.2),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: offer['logo_url'] != null && offer['logo_url'].toString().isNotEmpty
                  ? ClipRRect(
                      borderRadius: BorderRadius.circular(12),
                      child: Image.network(
                        offer['logo_url'],
                        fit: BoxFit.cover,
                        errorBuilder: (_, __, ___) => const Icon(
                          Icons.local_offer,
                          color: Color(0xFF6C63FF),
                        ),
                      ),
                    )
                  : const Icon(Icons.local_offer, color: Color(0xFF6C63FF)),
              ),
              
              const SizedBox(width: 12),
              
              // Offer Info
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      offer['title'] ?? 'Untitled',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      offer['category'] ?? 'General',
                      style: TextStyle(
                        color: Colors.white.withOpacity(0.5),
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
              
              // Status Badge
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: statusColor.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  statusText,
                  style: TextStyle(
                    color: statusColor,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ],
          ),
          
          // Rejection Reason (if rejected)
          if (status == 'rejected' && offer['rejection_reason'] != null) ...[
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: Colors.red.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                children: [
                  const Icon(Icons.info_outline, color: Colors.red, size: 16),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '${isArabic ? 'سبب الرفض: ' : 'Reason: '}${offer['rejection_reason']}',
                      style: const TextStyle(
                        color: Colors.red,
                        fontSize: 12,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
          
          const SizedBox(height: 12),
          
          // Stats Row
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _buildOfferStat(
                Icons.people_outline,
                '${offer['promoters_count'] ?? 0}',
                isArabic ? 'مروج' : 'Promoters',
              ),
              _buildOfferStat(
                Icons.touch_app,
                '${offer['total_clicks'] ?? 0}',
                isArabic ? 'نقرة' : 'Clicks',
              ),
              _buildOfferStat(
                Icons.trending_up,
                '${offer['total_conversions'] ?? 0}',
                isArabic ? 'تحويل' : 'Conv.',
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildOfferStat(IconData icon, String value, String label) {
    return Column(
      children: [
        Row(
          children: [
            Icon(icon, color: Colors.white54, size: 16),
            const SizedBox(width: 4),
            Text(
              value,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
                fontWeight: FontWeight.bold,
              ),
            ),
          ],
        ),
        Text(
          label,
          style: TextStyle(
            color: Colors.white.withOpacity(0.5),
            fontSize: 10,
          ),
        ),
      ],
    );
  }

  Widget _buildEmptyOffersWidget(bool isArabic) {
    return Container(
      padding: const EdgeInsets.all(40),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        children: [
          Icon(
            Icons.inbox_outlined,
            color: Colors.white.withOpacity(0.3),
            size: 60,
          ),
          const SizedBox(height: 16),
          Text(
            isArabic ? 'لا توجد عروض بعد' : 'No offers yet',
            style: TextStyle(
              color: Colors.white.withOpacity(0.5),
              fontSize: 16,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            isArabic 
              ? 'أضف عرضك الأول للبدء'
              : 'Add your first offer to get started',
            style: TextStyle(
              color: Colors.white.withOpacity(0.3),
              fontSize: 14,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorWidget(bool isArabic) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.red.withOpacity(0.1),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        children: [
          const Icon(Icons.error_outline, color: Colors.red, size: 40),
          const SizedBox(height: 12),
          Text(
            isArabic ? 'حدث خطأ في تحميل البيانات' : 'Error loading data',
            style: const TextStyle(color: Colors.red, fontSize: 16),
          ),
          const SizedBox(height: 8),
          ElevatedButton(
            onPressed: _loadDashboard,
            child: Text(isArabic ? 'إعادة المحاولة' : 'Retry'),
          ),
        ],
      ),
    );
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'active':
        return const Color(0xFF4CAF50);
      case 'pending':
        return const Color(0xFFFF9800);
      case 'rejected':
        return Colors.red;
      case 'paused':
        return Colors.grey;
      default:
        return Colors.grey;
    }
  }

  String _getStatusText(String status, bool isArabic) {
    switch (status) {
      case 'active':
        return isArabic ? 'نشط' : 'Active';
      case 'pending':
        return isArabic ? 'قيد المراجعة' : 'Pending';
      case 'rejected':
        return isArabic ? 'مرفوض' : 'Rejected';
      case 'paused':
        return isArabic ? 'متوقف' : 'Paused';
      default:
        return status;
    }
  }
}
