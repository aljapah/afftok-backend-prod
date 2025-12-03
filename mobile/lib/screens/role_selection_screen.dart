import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../utils/app_localizations.dart';
import '../providers/language_provider.dart';
import 'signin_screen.dart';
import 'advertiser/advertiser_register_screen.dart';

class RoleSelectionScreen extends StatelessWidget {
  const RoleSelectionScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';

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
          // Second layer for depth
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
            child: Padding(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                children: [
                  // Language Toggle Button
                  Align(
                    alignment: Alignment.topLeft,
                    child: Consumer<LanguageProvider>(
                      builder: (context, langProvider, child) {
                        final isCurrentlyArabic = langProvider.locale.languageCode == 'ar';
                        return GestureDetector(
                          onTap: () {
                            langProvider.setLocale(
                              isCurrentlyArabic 
                                ? const Locale('en', '') 
                                : const Locale('ar', ''),
                            );
                          },
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                            decoration: BoxDecoration(
                              color: Colors.white.withOpacity(0.1),
                              borderRadius: BorderRadius.circular(20),
                              border: Border.all(color: Colors.white.withOpacity(0.2)),
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                const Icon(Icons.language, color: Colors.white, size: 18),
                                const SizedBox(width: 6),
                                Text(
                                  isCurrentlyArabic ? 'English' : 'العربية',
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 14,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                  
                  const SizedBox(height: 20),
                  
                  // App Logo - Bigger
                  ClipRRect(
                    borderRadius: BorderRadius.circular(24),
                    child: Image.asset(
                      'assets/logo.png',
                      width: 130,
                      height: 130,
                      errorBuilder: (context, error, stackTrace) {
                        // Fallback if logo not found
                        return Container(
                          width: 130,
                          height: 130,
                          decoration: const BoxDecoration(
                            shape: BoxShape.circle,
                            gradient: LinearGradient(
                              colors: [Color(0xFFE53935), Color(0xFFFF6B6B)],
                            ),
                          ),
                          child: const Center(
                            child: Text(
                              'A',
                              style: TextStyle(
                                fontSize: 48,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                  
                  const SizedBox(height: 24),
                  
                  // Title
                  Text(
                    isArabic ? 'مرحباً بك في AffTok' : 'Welcome to AffTok',
                    style: const TextStyle(
                      fontSize: 28,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  
                  const SizedBox(height: 12),
                  
                  Text(
                    isArabic 
                      ? 'اختر نوع حسابك للمتابعة'
                      : 'Choose your account type to continue',
                    style: TextStyle(
                      fontSize: 16,
                      color: Colors.white.withOpacity(0.7),
                    ),
                    textAlign: TextAlign.center,
                  ),
                  
                  const SizedBox(height: 60),
                  
                // Promoter Card
                _buildRoleCard(
                  context,
                  icon: Icons.campaign_rounded,
                  title: isArabic ? 'مسوّق' : 'Promoter',
                  subtitle: isArabic 
                    ? 'انشر العروض واكسب من كل تحويل'
                    : 'Share offers and earn from conversions',
                  gradientColors: const [Color(0xFFE53935), Color(0xFFFF6B6B)],
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const SignInScreen(),
                      ),
                    );
                  },
                ),
                
                const SizedBox(height: 20),
                
                // Advertiser Card
                _buildRoleCard(
                  context,
                  icon: Icons.business_center_rounded,
                  title: isArabic ? 'معلن' : 'Advertiser',
                  subtitle: isArabic 
                    ? 'أضف عروضك واحصل على مروجين'
                    : 'Add your offers and get promoters',
                  gradientColors: const [Color(0xFF6C63FF), Color(0xFF9D4EDD)],
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const AdvertiserRegisterScreen(),
                      ),
                    );
                  },
                ),
                  
                  const Spacer(),
                  
                  // Footer
                  Text(
                    isArabic 
                      ? 'منصة مجانية 100% للمروجين والمعلنين'
                      : '100% free platform for promoters and advertisers',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.white.withOpacity(0.5),
                    ),
                    textAlign: TextAlign.center,
                  ),
                  
                  const SizedBox(height: 20),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRoleCard(
    BuildContext context, {
    required IconData icon,
    required String title,
    required String subtitle,
    required List<Color> gradientColors,
    required VoidCallback onTap,
  }) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 24, horizontal: 20),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(20),
          color: Colors.white.withOpacity(0.05),
          border: Border.all(
            color: gradientColors[0].withOpacity(0.3),
            width: 1,
          ),
        ),
        child: Row(
          children: [
            // Arrow (RTL support)
            Icon(
              Icons.arrow_back_ios,
              color: gradientColors[0].withOpacity(0.5),
              size: 18,
            ),
            
            const SizedBox(width: 12),
            
            // Content - Centered
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  // Title (مسوّق / معلن) - Big and centered
                  Text(
                    title,
                    style: TextStyle(
                      fontSize: 26,
                      fontWeight: FontWeight.bold,
                      color: gradientColors[0],
                    ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 8),
                  // Subtitle description
                  Text(
                    subtitle,
                    style: TextStyle(
                      fontSize: 13,
                      color: Colors.white.withOpacity(0.6),
                    ),
                    textAlign: TextAlign.center,
                  ),
                ],
              ),
            ),
            
            const SizedBox(width: 12),
            
            // Icon
            Container(
              width: 55,
              height: 55,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: LinearGradient(colors: gradientColors),
                boxShadow: [
                  BoxShadow(
                    color: gradientColors[0].withOpacity(0.4),
                    blurRadius: 12,
                    spreadRadius: 2,
                  ),
                ],
              ),
              child: Icon(icon, color: Colors.white, size: 26),
            ),
          ],
        ),
      ),
    );
  }
}
