import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../../providers/auth_provider.dart';
import '../../services/advertiser_service.dart';

class PromotersScreen extends StatefulWidget {
  const PromotersScreen({super.key});

  @override
  State<PromotersScreen> createState() => _PromotersScreenState();
}

class _PromotersScreenState extends State<PromotersScreen> {
  final AdvertiserService _advertiserService = AdvertiserService();
  List<dynamic> _promoters = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadPromoters();
  }

  Future<void> _loadPromoters() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final authProvider = Provider.of<AuthProvider>(context, listen: false);
      final token = authProvider.token;
      
      // Check if token is null or empty
      if (token == null || token.isEmpty) {
        setState(() {
          _error = 'Not authenticated. Please login again.';
          _isLoading = false;
        });
        return;
      }
      
      final response = await _advertiserService.getPromoters(token);
      
      if (mounted) {
        setState(() {
          _promoters = response['promoters'] ?? [];
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString().replaceAll('Exception: ', '');
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: Text(
          isArabic ? 'سجل المروجين' : 'Promoters Log',
          style: const TextStyle(color: Colors.white),
        ),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh, color: Colors.white),
            onPressed: _loadPromoters,
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFF6C63FF)))
          : _error != null
              ? _buildErrorState(isArabic)
              : _promoters.isEmpty
                  ? _buildEmptyState(isArabic)
                  : _buildPromotersList(isArabic),
    );
  }

  Widget _buildErrorState(bool isArabic) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.error_outline,
            size: 64,
            color: Colors.red.withOpacity(0.5),
          ),
          const SizedBox(height: 16),
          Text(
            isArabic ? 'حدث خطأ' : 'An error occurred',
            style: TextStyle(
              color: Colors.white.withOpacity(0.7),
              fontSize: 18,
            ),
          ),
          const SizedBox(height: 8),
          TextButton(
            onPressed: _loadPromoters,
            child: Text(
              isArabic ? 'إعادة المحاولة' : 'Retry',
              style: const TextStyle(color: Color(0xFF6C63FF)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState(bool isArabic) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.people_outline,
            size: 80,
            color: Colors.white.withOpacity(0.3),
          ),
          const SizedBox(height: 16),
          Text(
            isArabic ? 'لا يوجد مروجين بعد' : 'No promoters yet',
            style: TextStyle(
              color: Colors.white.withOpacity(0.7),
              fontSize: 18,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            isArabic 
                ? 'عندما يسجل المروجين في عروضك، ستراهم هنا'
                : 'When promoters join your offers, they will appear here',
            style: TextStyle(
              color: Colors.white.withOpacity(0.5),
              fontSize: 14,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  Widget _buildPromotersList(bool isArabic) {
    return RefreshIndicator(
      onRefresh: _loadPromoters,
      color: const Color(0xFF6C63FF),
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _promoters.length,
        itemBuilder: (context, index) {
          final promoter = _promoters[index];
          return _buildPromoterCard(promoter, isArabic);
        },
      ),
    );
  }

  Widget _buildPromoterCard(Map<String, dynamic> promoter, bool isArabic) {
    final username = promoter['username'] ?? 'Unknown';
    final fullName = promoter['full_name'] ?? username;
    final email = promoter['email'] ?? '';
    final clicks = promoter['clicks'] ?? 0;
    final conversions = promoter['conversions'] ?? 0;
    final paymentMethod = promoter['payment_method'] ?? '';
    final offerTitle = promoter['offer_title'] ?? '';
    final joinedAt = promoter['joined_at'] ?? '';

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            const Color(0xFF1A1A2E),
            const Color(0xFF16213E).withOpacity(0.8),
          ],
        ),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: Colors.white.withOpacity(0.1),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header with avatar and name
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                // Avatar
                Container(
                  width: 50,
                  height: 50,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    gradient: const LinearGradient(
                      colors: [Color(0xFF6C63FF), Color(0xFF9D4EDD)],
                    ),
                  ),
                  child: Center(
                    child: Text(
                      fullName.isNotEmpty ? fullName[0].toUpperCase() : 'U',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                // Name and username
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        fullName,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Text(
                        '@$username',
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.6),
                          fontSize: 13,
                        ),
                      ),
                    ],
                  ),
                ),
                // Offer badge
                if (offerTitle.isNotEmpty)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: const Color(0xFF6C63FF).withOpacity(0.2),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text(
                      offerTitle.length > 15 ? '${offerTitle.substring(0, 15)}...' : offerTitle,
                      style: const TextStyle(
                        color: Color(0xFF6C63FF),
                        fontSize: 11,
                      ),
                    ),
                  ),
              ],
            ),
          ),

          // Stats Row
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: Colors.black.withOpacity(0.3),
            ),
            child: Row(
              children: [
                Expanded(
                  child: _buildStatItem(
                    Icons.touch_app,
                    isArabic ? 'نقرات' : 'Clicks',
                    clicks.toString(),
                    const Color(0xFF00D9FF),
                  ),
                ),
                Container(
                  width: 1,
                  height: 40,
                  color: Colors.white.withOpacity(0.1),
                ),
                Expanded(
                  child: _buildStatItem(
                    Icons.check_circle,
                    isArabic ? 'تحويلات' : 'Conversions',
                    conversions.toString(),
                    const Color(0xFF00FF88),
                  ),
                ),
                Container(
                  width: 1,
                  height: 40,
                  color: Colors.white.withOpacity(0.1),
                ),
                Expanded(
                  child: _buildStatItem(
                    Icons.percent,
                    isArabic ? 'معدل' : 'Rate',
                    clicks > 0 
                        ? '${((conversions / clicks) * 100).toStringAsFixed(1)}%'
                        : '0%',
                    const Color(0xFFFF6B6B),
                  ),
                ),
              ],
            ),
          ),

          // Payment Method Section
          if (paymentMethod.isNotEmpty)
            Container(
              margin: const EdgeInsets.all(16),
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFF00FF88).withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: const Color(0xFF00FF88).withOpacity(0.3),
                ),
              ),
              child: Row(
                children: [
                  const Icon(
                    Icons.account_balance_wallet,
                    color: Color(0xFF00FF88),
                    size: 20,
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          isArabic ? 'طريقة الدفع' : 'Payment Method',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.6),
                            fontSize: 11,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          paymentMethod,
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.copy, color: Color(0xFF00FF88), size: 18),
                    onPressed: () {
                      Clipboard.setData(ClipboardData(text: paymentMethod));
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(isArabic ? 'تم النسخ' : 'Copied'),
                          backgroundColor: const Color(0xFF00FF88),
                          duration: const Duration(seconds: 1),
                        ),
                      );
                    },
                  ),
                ],
              ),
            )
          else
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  Icon(
                    Icons.info_outline,
                    color: Colors.orange.withOpacity(0.7),
                    size: 18,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    isArabic 
                        ? 'لم يحدد المروج طريقة الدفع بعد'
                        : 'Promoter hasn\'t set payment method',
                    style: TextStyle(
                      color: Colors.orange.withOpacity(0.7),
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),

          // Contact Button
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
            child: Row(
              children: [
                if (email.isNotEmpty)
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: () {
                        Clipboard.setData(ClipboardData(text: email));
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(
                            content: Text(isArabic ? 'تم نسخ البريد' : 'Email copied'),
                            backgroundColor: const Color(0xFF6C63FF),
                            duration: const Duration(seconds: 1),
                          ),
                        );
                      },
                      icon: const Icon(Icons.email, size: 16),
                      label: Text(
                        isArabic ? 'نسخ البريد' : 'Copy Email',
                        style: const TextStyle(fontSize: 12),
                      ),
                      style: OutlinedButton.styleFrom(
                        foregroundColor: const Color(0xFF6C63FF),
                        side: const BorderSide(color: Color(0xFF6C63FF)),
                        padding: const EdgeInsets.symmetric(vertical: 8),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatItem(IconData icon, String label, String value, Color color) {
    return Column(
      children: [
        Icon(icon, color: color, size: 20),
        const SizedBox(height: 4),
        Text(
          value,
          style: TextStyle(
            color: color,
            fontSize: 18,
            fontWeight: FontWeight.bold,
          ),
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
}

