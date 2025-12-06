import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:provider/provider.dart';
import '../utils/app_localizations.dart';
import '../models/user.dart';
import '../models/user_offer.dart';
import '../models/team.dart';
import '../providers/auth_provider.dart';
import 'settings_screen.dart';
import 'teams_screen.dart';
import 'home_feed_screen.dart';
import 'webview_screen.dart';
import 'promoter_public_page.dart';
import 'advertiser/advertiser_dashboard_screen.dart';
import 'role_selection_screen.dart';

class WebViewPage extends StatelessWidget {
  final String url;
  final String title;

  const WebViewPage({Key? key, required this.url, required this.title})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(title),
        backgroundColor: Colors.black,
      ),
      body: WebViewScreen(url: url, title: title, offerId: ''),
    );
  }
}

class ProfileScreenEnhanced extends StatefulWidget {
  const ProfileScreenEnhanced({Key? key}) : super(key: key);

  @override
  State<ProfileScreenEnhanced> createState() => _ProfileScreenEnhancedState();
}

class _ProfileScreenEnhancedState extends State<ProfileScreenEnhanced> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _refreshData();
    });
  }
  
  Future<void> _refreshData() async {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    // Only refresh if logged in
    if (authProvider.isLoggedIn) {
      await authProvider.loadCurrentUser(forceRefresh: true);
    }
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        // If not logged in, redirect to role selection
        if (!authProvider.isLoggedIn) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            Navigator.pushAndRemoveUntil(
              context,
              MaterialPageRoute(builder: (context) => const RoleSelectionScreen()),
              (route) => false,
            );
          });
          return Scaffold(
            backgroundColor: Colors.black,
            body: const Center(child: CircularProgressIndicator()),
          );
        }
        
        final user = authProvider.currentUser;
        
        // If user is null and still loading, show loading indicator
        if (user == null && authProvider.isLoading) {
          return Scaffold(
            backgroundColor: Colors.black,
            appBar: AppBar(
              backgroundColor: Colors.black,
              title: Text(lang.profile),
            ),
            body: const Center(
              child: CircularProgressIndicator(),
            ),
          );
        }
        
        // If user is null and not loading, go back
        if (user == null) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (Navigator.canPop(context)) {
              Navigator.pop(context);
            }
          });
          return Scaffold(
            backgroundColor: Colors.black,
            body: const Center(child: CircularProgressIndicator()),
          );
        }
        
        // If user is advertiser, redirect to advertiser dashboard
        if (user.isAdvertiser) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            Navigator.pushReplacement(
              context,
              MaterialPageRoute(builder: (context) => const AdvertiserDashboardScreen()),
            );
          });
          return Scaffold(
            backgroundColor: Colors.black,
            body: const Center(child: CircularProgressIndicator()),
          );
        }
        
        return Scaffold(
          backgroundColor: Colors.black,
          appBar: AppBar(
            backgroundColor: Colors.black,
            title: Text(
              lang.profile,
              style: const TextStyle(color: Colors.white),
            ),
            leading: IconButton(
              icon: const Icon(Icons.arrow_back, color: Colors.white),
              onPressed: () => Navigator.pop(context),
            ),
            actions: [
              // Refresh button
              IconButton(
                icon: const Icon(Icons.refresh, color: Colors.white),
                onPressed: () => _refreshData(),
              ),
              IconButton(
                icon: const Icon(Icons.settings, color: Colors.white),
                onPressed: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const SettingsScreen()),
                  );
                },
              ),
            ],
          ),
          body: RefreshIndicator(
            onRefresh: _refreshData,
            color: const Color(0xFFFF006E),
            backgroundColor: Colors.grey[900],
            child: SingleChildScrollView(
              physics: const AlwaysScrollableScrollPhysics(),
              child: Column(
                children: [
                  const SizedBox(height: 24),
                  
                  // Profile Header
                  _buildProfileHeader(context, user, lang),
                  
                  const SizedBox(height: 24),
                  
                  // Personal Link Card
                  _buildPersonalLinkCard(context, user, lang),
                  
                  const SizedBox(height: 24),
                  
                  // Payment Method Card
                  _buildPaymentMethodCard(context, user, lang),
                  
                  const SizedBox(height: 24),
                  
                  // Stats Cards
                  _buildStatsCards(context, user, lang),
                  
                  const SizedBox(height: 24),
                  
                  // Team Section (if in team)
                  if (user.isInTeam) ...[
                    _buildTeamSection(context, lang),
                    const SizedBox(height: 24),
                  ],
                  
                  // Add New Offers Button
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 20),
                    child: SizedBox(
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        onPressed: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(builder: (context) => const HomeFeedScreen()),
                          );
                        },
                        icon: const Icon(Icons.add, color: Colors.white),
                        label: Text(
                          lang.addMoreOffers,
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFFFF006E),
                          padding: const EdgeInsets.symmetric(vertical: 16),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                      ),
                    ),
                  ),
                  
                  const SizedBox(height: 32),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildProfileHeader(BuildContext context, User user, AppLocalizations lang) {
    return Column(
      children: [
        // Avatar with level badge
        Stack(
          clipBehavior: Clip.none,
          children: [
            TweenAnimationBuilder(
              duration: const Duration(milliseconds: 800),
              tween: Tween<double>(begin: 0, end: 1),
              builder: (context, double value, child) {
                return Transform.scale(
                  scale: value,
                  child: Container(
                    width: 100,
                    height: 100,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: const LinearGradient(
                        colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFFFF006E).withOpacity(0.5),
                          blurRadius: 20,
                          spreadRadius: 2,
                        ),
                      ],
                    ),
                    padding: const EdgeInsets.all(3),
                    child: Container(
                      decoration: const BoxDecoration(
                        shape: BoxShape.circle,
                        color: Colors.black,
                      ),
                        child: Container(
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: Colors.grey[800],
                        ),
                        child: user.avatarUrl != null && user.avatarUrl!.isNotEmpty
                            ? ClipOval(
                                child: Image.network(
                                  user.avatarUrl!,
                                  fit: BoxFit.cover,
                                  errorBuilder: (context, error, stackTrace) {
                                    return const Icon(
                                      Icons.person,
                                      size: 40,
                                      color: Colors.white,
                                    );
                                  },
                                ),
                              )
                            : const Icon(
                                Icons.person,
                                size: 40,
                                color: Colors.white,
                              ),
                      ),
                    ),
                  ),
                );
              },
            ),
            // Level Badge
            Positioned(
              bottom: -5,
              right: -5,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                  ),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.black, width: 2),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      user.level.emoji,
                      style: const TextStyle(fontSize: 12),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      user.level.displayName,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
        
        const SizedBox(height: 16),
        
        // Display Name
        Text(
          user.displayName,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 24,
            fontWeight: FontWeight.bold,
          ),
        ),
        
        const SizedBox(height: 4),
        
        // Username
        Text(
          '@${user.username}',
          style: TextStyle(
            color: Colors.white.withOpacity(0.6),
            fontSize: 14,
          ),
        ),
        
        // Bio (if exists)
        if (user.bio != null && user.bio!.isNotEmpty) ...[
          const SizedBox(height: 8),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Text(
              user.bio!,
              textAlign: TextAlign.center,
              style: TextStyle(
                color: Colors.white.withOpacity(0.7),
                fontSize: 13,
              ),
              maxLines: 3,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
        
        const SizedBox(height: 8),
        
        // Rank Badge
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(0.1),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(
              color: Colors.white.withOpacity(0.2),
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Text('🏆', style: TextStyle(fontSize: 14)),
              const SizedBox(width: 6),
              Text(
                '${lang.rank} #${user.stats.globalRank}',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildPersonalLinkCard(BuildContext context, User user, AppLocalizations lang) {
    // Unique link with unique code (falls back to username if no code)
    final uniqueCode = user.uniqueCode;
    final uniqueLink = (uniqueCode != null && uniqueCode.isNotEmpty)
        ? 'https://afftok-backend-prod-production.up.railway.app/api/r/$uniqueCode'
        : 'https://afftok-backend-prod-production.up.railway.app/api/promoter/user/${user.username}';
    
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              const Color(0xFFFF006E).withOpacity(0.25),
              const Color(0xFFFF4D00).withOpacity(0.15),
            ],
          ),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: const Color(0xFFFF006E).withOpacity(0.4),
            width: 1.5,
          ),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFFFF006E).withOpacity(0.2),
              blurRadius: 15,
              spreadRadius: 0,
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: const Color(0xFFFF006E).withOpacity(0.2),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Icon(
                    Icons.link,
                    color: Color(0xFFFF006E),
                    size: 20,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        lang.locale.languageCode == 'ar' 
                            ? '🔗 رابطك الفريد' 
                            : '🔗 Your Unique Link',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        lang.locale.languageCode == 'ar'
                            ? 'شاركه لكسب العمولات'
                            : 'Share it to earn commissions',
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.6),
                          fontSize: 11,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            // Clickable unique link
            GestureDetector(
              onTap: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (context) => PromoterPublicPage(username: user.username),
                  ),
                );
              },
              child: Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: Colors.black.withOpacity(0.4),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: const Color(0xFFFF006E).withOpacity(0.3),
                  ),
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.open_in_new,
                      color: Color(0xFFFF006E),
                      size: 18,
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        uniqueLink,
                        style: const TextStyle(
                          color: Color(0xFFFF006E),
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          decoration: TextDecoration.underline,
                          decorationColor: Color(0xFFFF006E),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 14),
            Row(
              children: [
                Expanded(
                  child: _ActionButton(
                    icon: Icons.copy,
                    label: lang.copy,
                    onTap: () {
                      Clipboard.setData(ClipboardData(text: uniqueLink));
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(lang.linkCopied),
                          backgroundColor: const Color(0xFFFF006E),
                          duration: const Duration(seconds: 2),
                        ),
                      );
                    },
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _ActionButton(
                    icon: Icons.share,
                    label: lang.share,
                    onTap: () {
                      Share.share(
                        uniqueLink,
                        subject: 'Check out my AffTok profile!',
                      );
                    },
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _ActionButton(
                    icon: Icons.qr_code,
                    label: 'QR',
                    onTap: () {
                      _showQRCodeDialog(context, user, lang);
                    },
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPaymentMethodCard(BuildContext context, User user, AppLocalizations lang) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              const Color(0xFF00FF88).withOpacity(0.15),
              const Color(0xFF00D9FF).withOpacity(0.08),
            ],
          ),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: const Color(0xFF00FF88).withOpacity(0.3),
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: const Color(0xFF00FF88).withOpacity(0.2),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Icon(
                    Icons.account_balance_wallet,
                    color: Color(0xFF00FF88),
                    size: 20,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        lang.locale.languageCode == 'ar' 
                            ? 'طريقة استلام الأرباح' 
                            : 'Payment Method',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        lang.locale.languageCode == 'ar'
                            ? 'يراها المعلن للتواصل معك'
                            : 'Visible to advertisers',
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.5),
                          fontSize: 11,
                        ),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  onPressed: () => _showEditPaymentMethodDialog(context, user, lang),
                  icon: Icon(
                    user.paymentMethod != null && user.paymentMethod!.isNotEmpty
                        ? Icons.edit
                        : Icons.add_circle_outline,
                    color: const Color(0xFF00FF88),
                    size: 22,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: Colors.black.withOpacity(0.3),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: Colors.white.withOpacity(0.1),
                ),
              ),
              child: user.paymentMethod != null && user.paymentMethod!.isNotEmpty
                  ? Row(
                      children: [
                        const Icon(
                          Icons.check_circle,
                          color: Color(0xFF00FF88),
                          size: 18,
                        ),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Text(
                            user.paymentMethod!,
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 14,
                            ),
                          ),
                        ),
                      ],
                    )
                  : Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.info_outline,
                          color: Colors.white.withOpacity(0.4),
                          size: 18,
                        ),
                        const SizedBox(width: 8),
                        Text(
                          lang.locale.languageCode == 'ar'
                              ? 'لم يتم تحديد طريقة الدفع بعد'
                              : 'No payment method set',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.5),
                            fontSize: 13,
                          ),
                        ),
                      ],
                    ),
            ),
            if (user.paymentMethod == null || user.paymentMethod!.isEmpty) ...[
              const SizedBox(height: 12),
              Text(
                lang.locale.languageCode == 'ar'
                    ? 'أمثلة: PayPal: email@example.com أو STC Pay: 0551234567'
                    : 'Examples: PayPal: email@example.com or Bank: IBAN...',
                style: TextStyle(
                  color: Colors.white.withOpacity(0.4),
                  fontSize: 11,
                  fontStyle: FontStyle.italic,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  void _showEditPaymentMethodDialog(BuildContext context, User user, AppLocalizations lang) {
    final controller = TextEditingController(text: user.paymentMethod ?? '');
    final isArabic = lang.locale.languageCode == 'ar';
    
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Padding(
        padding: EdgeInsets.only(
          bottom: MediaQuery.of(context).viewInsets.bottom,
        ),
        child: Container(
          padding: const EdgeInsets.all(24),
          decoration: const BoxDecoration(
            color: Color(0xFF1A1A1A),
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: const Color(0xFF00FF88).withOpacity(0.2),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(
                      Icons.account_balance_wallet,
                      color: Color(0xFF00FF88),
                      size: 24,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          isArabic ? 'طريقة استلام الأرباح' : 'Payment Method',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 20,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          isArabic
                              ? 'أدخل طريقة الدفع المفضلة لديك'
                              : 'Enter your preferred payment method',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.6),
                            fontSize: 13,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              
              // Text Field
              TextField(
                controller: controller,
                maxLines: 3,
                style: const TextStyle(color: Colors.white),
                decoration: InputDecoration(
                  hintText: isArabic
                      ? 'أدخل طريقة الدفع المفضلة لديك...'
                      : 'Enter your preferred payment method...',
                  hintStyle: TextStyle(color: Colors.white.withOpacity(0.3)),
                  filled: true,
                  fillColor: Colors.white.withOpacity(0.05),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: Colors.white.withOpacity(0.2)),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: Colors.white.withOpacity(0.2)),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(color: Color(0xFF00FF88)),
                  ),
                ),
              ),
              const SizedBox(height: 24),
              
              // Save Button
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () async {
                    Navigator.pop(context);
                    await _updatePaymentMethod(context, controller.text);
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF00FF88),
                    foregroundColor: Colors.black,
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: Text(
                    isArabic ? 'حفظ' : 'Save',
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 8),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _updatePaymentMethod(BuildContext context, String paymentMethod) async {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    
    try {
      // Call API to update payment method
      await authProvider.updateProfile(paymentMethod: paymentMethod);
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              AppLocalizations.of(context).locale.languageCode == 'ar'
                  ? 'تم حفظ طريقة الدفع بنجاح'
                  : 'Payment method saved successfully',
            ),
            backgroundColor: const Color(0xFF00FF88),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Widget _buildStatsCards(BuildContext context, User user, AppLocalizations lang) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            lang.yourPerformance,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 18,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: _StatCard(
                  icon: Icons.touch_app,
                  label: lang.clicks,
                  value: user.stats.totalClicks.toString(),
                  color: const Color(0xFF00D9FF),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _StatCard(
                  icon: Icons.people,
                  label: lang.referrals,
                  value: user.stats.totalConversions.toString(),
                  color: const Color(0xFFFF006E),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _StatCard(
                  icon: Icons.attach_money,
                  label: lang.registeredOffers,
                  value: '${user.stats.totalRegisteredOffers}',
                  color: const Color(0xFF00FF88),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.05),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: Colors.white.withOpacity(0.1),
              ),
            ),
            child: Column(
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        const Text('📈', style: TextStyle(fontSize: 16)),
                        const SizedBox(width: 8),
                        Text(
                          lang.thisMonth,
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.7),
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                    Text(
                      '${user.stats.monthlyConversions} ${lang.conversions} (+${user.stats.conversionRate.toInt()}%)',
                      style: const TextStyle(
                        color: Color(0xFF00FF88),
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        const Text('🔥', style: TextStyle(fontSize: 16)),
                        const SizedBox(width: 8),
                        Text(
                          lang.bestOffer,
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.7),
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                    Text(
                      '${user.stats.bestOffer} (${user.stats.offerStats[user.stats.bestOffer]?.conversions ?? 0} conversions)',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTeamSection(BuildContext context, AppLocalizations lang) {
    
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: GestureDetector(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (context) => const TeamsScreen()),
          );
        },
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [
                const Color(0xFFFF006E).withOpacity(0.2),
                const Color(0xFFFF006E).withOpacity(0.05),
              ],
            ),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(
              color: const Color(0xFFFF006E).withOpacity(0.3),
            ),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Icon(
                    Icons.groups,
                    color: Colors.white,
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    lang.myTeam,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const Spacer(),
                  const Icon(
                    Icons.arrow_forward_ios,
                    color: Colors.white,
                    size: 16,
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  const Text(
                    '👥',
                    style: TextStyle(fontSize: 24),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          lang.myTeam,
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'عرض الفرق',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.7),
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    child: _buildTeamStat(
                      '0',
                      lang.members,
                    ),
                  ),
                  Expanded(
                    child: _buildTeamStat(
                      '0',
                      lang.referrals,
                    ),
                  ),
                  Expanded(
                    child: _buildTeamStat(
                      '0',
                      lang.conversions,
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildTeamStat(String value, String label) {
    return Column(
      children: [
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 16,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: TextStyle(
            color: Colors.white.withOpacity(0.6),
            fontSize: 11,
          ),
        ),
      ],
    );
  }

  static void _showQRCodeDialog(BuildContext context, User user, AppLocalizations lang) {
    final uniqueCode = user.uniqueCode;
    final uniqueLink = (uniqueCode != null && uniqueCode.isNotEmpty)
        ? 'https://afftok-backend-prod-production.up.railway.app/api/r/$uniqueCode'
        : 'https://afftok-backend-prod-production.up.railway.app/api/promoter/user/${user.username}';
    
    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: Colors.transparent,
        child: Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
            ),
            borderRadius: BorderRadius.circular(20),
            boxShadow: [
              BoxShadow(
                color: const Color(0xFFFF006E).withOpacity(0.6),
                blurRadius: 30,
                spreadRadius: 5,
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                lang.locale.languageCode == 'ar' ? '🔗 رابطك الفريد' : '🔗 Your Unique Link',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 20),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                ),
                child: QrImageView(
                  data: uniqueLink,
                  version: QrVersions.auto,
                  size: 200.0,
                  backgroundColor: Colors.white,
                ),
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  uniqueLink,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                  ),
                  textAlign: TextAlign.center,
                ),
              ),
              const SizedBox(height: 20),
              ElevatedButton(
                onPressed: () => Navigator.pop(context),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.white,
                  foregroundColor: const Color(0xFFFF006E),
                  padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                child: Text(
                  lang.locale.languageCode == 'ar' ? 'إغلاق' : 'Close',
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatCard extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  final Color color;

  const _StatCard({
    required this.icon,
    required this.label,
    required this.value,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: color.withOpacity(0.3),
        ),
      ),
      child: Column(
        children: [
          Icon(icon, color: color, size: 24),
          const SizedBox(height: 8),
          Text(
            value,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 18,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            label,
            style: TextStyle(
              color: Colors.white.withOpacity(0.6),
              fontSize: 11,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}

class _ActionButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  const _ActionButton({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 10),
        decoration: BoxDecoration(
          color: Colors.white.withOpacity(0.1),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, color: Colors.white, size: 16),
            const SizedBox(width: 6),
            Text(
              label,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 12,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
