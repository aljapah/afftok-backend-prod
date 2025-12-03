import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../utils/app_localizations.dart';
import '../providers/language_provider.dart';
import '../providers/auth_provider.dart';
import 'edit_profile_screen.dart';
import 'privacy_screen.dart';
import 'security_screen.dart';
import 'terms_screen.dart';
import 'about_screen.dart';
import 'role_selection_screen.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({Key? key}) : super(key: key);

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> with TickerProviderStateMixin {
  bool _notificationsEnabled = true;
  bool _exclusiveOffersOnly = false;
  List<String> _selectedCategories = ['Shopping', 'Crypto'];

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    final languageProvider = Provider.of<LanguageProvider>(context);
    final isArabic = languageProvider.locale.languageCode == 'ar';
    
    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          Positioned.fill(
            child: Opacity(
              opacity: 0.05,
              child: Center(
                child: Transform.rotate(
                  angle: -0.3,
                  child: const Text(
                    'AFFTOK',
                    style: TextStyle(
                      fontSize: 120,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: 20,
                    ),
                  ),
                ),
              ),
            ),
          ),
          
          SafeArea(
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      IconButton(
                        icon: const Icon(Icons.arrow_back, color: Colors.white),
                        onPressed: () => Navigator.pop(context),
                      ),
                      Text(
                        lang.settings,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),
                
                Expanded(
                  child: ListView(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    children: [
                      const SizedBox(height: 10),
                      
                      _SectionHeader(title: lang.account),
                      const _GradientDivider(),
                      const SizedBox(height: 12),
                      
                      _AnimatedSettingsTile(
                        icon: Icons.person_outline,
                        title: lang.editProfile,
                        subtitle: lang.changeNamePhoto,
                        onTap: () async {
                          final result = await Navigator.push(
                            context,
                            MaterialPageRoute(builder: (context) => const EditProfileScreen()),
                          );
                          // Refresh profile data if changes were made
                          if (result == true && mounted) {
                            Provider.of<AuthProvider>(context, listen: false).loadCurrentUser(forceRefresh: true);
                          }
                        },
                      ),
                      
                      _AnimatedSettingsTile(
                        icon: Icons.security,
                        title: lang.security,
                        subtitle: lang.passwordAuth,
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(builder: (context) => const SecurityScreen()),
                          );
                        },
                      ),
                      
                      const SizedBox(height: 24),
                      
                      _SectionHeader(title: lang.preferences),
                      const _GradientDivider(),
                      const SizedBox(height: 12),
                      
                      _AnimatedSettingsTile(
                        icon: Icons.notifications_outlined,
                        title: lang.notifications,
                        subtitle: lang.notificationsSubtitle,
                        trailing: Switch(
                          value: _notificationsEnabled,
                          onChanged: (value) {
                            setState(() {
                              _notificationsEnabled = value;
                            });
                          },
                          activeColor: const Color(0xFFFF006E),
                        ),
                        onTap: () {
                          setState(() {
                            _notificationsEnabled = !_notificationsEnabled;
                          });
                        },
                      ),
                      
                      _AnimatedSettingsTile(
                        icon: Icons.language,
                        title: lang.language,
                        subtitle: isArabic ? 'العربية' : 'English',
                        onTap: () => _showLanguageDialog(context),
                      ),
                      
                      _AnimatedSettingsTile(
                        icon: Icons.tune,
                        title: lang.selectPreferredCategories,
                        subtitle: '',
                        onTap: () => _showCustomizeFeedBottomSheet(context),
                      ),
                      
                      const SizedBox(height: 24),
                      
                      _SectionHeader(title: lang.support),
                      const _GradientDivider(),
                      const SizedBox(height: 12),
                      
                      _AnimatedSettingsTile(
                        icon: Icons.privacy_tip_outlined,
                        title: lang.privacyPolicyTitle,
                        subtitle: lang.privacyPolicySubtitle,
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(builder: (context) => const PrivacyScreen()),
                          );
                        },
                      ),
                      _AnimatedSettingsTile(
                        icon: Icons.description_outlined,
                        title: lang.termsOfServiceTitle,
                        subtitle: lang.termsOfServiceSubtitle,
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(builder: (context) => const TermsScreen()),
                          );
                        },
                      ),
                      _AnimatedSettingsTile(
                        icon: Icons.info_outline,
                        title: lang.aboutScreenTitle,
                        subtitle: lang.learnMoreAboutApp,
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(builder: (context) => const AboutScreen()),
                          );
                        },
                      ),
                      _AnimatedSettingsTile(
                        icon: Icons.logout,
                        title: lang.logout,
                        subtitle: lang.signOutAccount,
                        onTap: () => _showLogoutDialog(context),
                        titleColor: Colors.red,
                      ),
                      
                      const SizedBox(height: 32),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  void _showLanguageDialog(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    final languageProvider = Provider.of<LanguageProvider>(context, listen: false);
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xFF1a1a1a),
        title: Text(
          lang.selectLanguage,
          style: const TextStyle(color: Colors.white),
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              title: const Text('English', style: TextStyle(color: Colors.white)),
              leading: Radio<String>(
                value: 'en',
                groupValue: languageProvider.locale.languageCode,
                onChanged: (value) {
                  languageProvider.setLocale(const Locale('en'));
                  Navigator.pop(context);
                },
                activeColor: const Color(0xFFFF006E),
              ),
            ),
            ListTile(
              title: const Text('العربية', style: TextStyle(color: Colors.white)),
              leading: Radio<String>(
                value: 'ar',
                groupValue: languageProvider.locale.languageCode,
                onChanged: (value) {
                  languageProvider.setLocale(const Locale('ar'));
                  Navigator.pop(context);
                },
                activeColor: const Color(0xFFFF006E),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showCustomizeFeedBottomSheet(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF1a1a1a),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => StatefulBuilder(
        builder: (context, setModalState) => Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                lang.selectCategories,
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 24),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  lang.shopping,
                  lang.crypto,
                  lang.services,
                  lang.food,
                  lang.travel,
                  lang.finance,
                  lang.entertainment,
                  lang.education,
                ].map((category) {
                  final isSelected = _selectedCategories.contains(category);
                  return FilterChip(
                    label: Text(category),
                    selected: isSelected,
                    onSelected: (selected) {
                      setModalState(() {
                        if (selected) {
                          _selectedCategories.add(category);
                        } else {
                          _selectedCategories.remove(category);
                        }
                      });
                      setState(() {});
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
                  );
                }).toList(),
              ),
              const SizedBox(height: 24),
            ],
          ),
        ),
      ),
    );
  }

  void _showComingSoonToast(BuildContext context, String feature) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('$feature - Coming in AffTok v2 🔒'),
        backgroundColor: const Color(0xFF1a1a1a),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  void _showAboutDialog(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: const Color(0xFF1a1a1a),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
        ),
        child: Container(
          constraints: const BoxConstraints(maxWidth: 500),
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 50,
                    height: 50,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: const LinearGradient(
                        colors: [Color(0xFF8E2DE2), Color(0xFFFF006E)],
                      ),
                    ),
                    child: const Center(
                      child: Text(
                        'A',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 28,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Text(
                      lang.aboutScreenTitle,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 24,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close, color: Colors.white70),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              
              Flexible(
                child: SingleChildScrollView(
                  child: Text(
                    lang.aboutFullText,
                    style: const TextStyle(
                      color: Colors.white70,
                      fontSize: 15,
                      height: 1.6,
                    ),
                  ),
                ),
              ),
              
              const SizedBox(height: 24),
              
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => Navigator.pop(context),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF8E2DE2),
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: Text(
                    lang.close,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
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

  void _showLogoutDialog(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    // Save references BEFORE dialog
    final navigator = Navigator.of(context);
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        backgroundColor: const Color(0xFF1a1a1a),
        title: Text(
          lang.logout,
          style: const TextStyle(color: Colors.white),
        ),
        content: Text(
          lang.logoutConfirm,
          style: const TextStyle(color: Colors.white70),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: Text(
              lang.cancel,
              style: const TextStyle(color: Colors.white70),
            ),
          ),
          TextButton(
            onPressed: () async {
              // Close dialog first
              Navigator.pop(dialogContext);
              
              // Perform actual logout
              await authProvider.logout();
              
              // Navigate to role selection screen and clear all previous routes
              navigator.pushAndRemoveUntil(
                MaterialPageRoute(builder: (_) => const RoleSelectionScreen()),
                (route) => false,
              );
            },
            child: Text(
              lang.logout,
              style: const TextStyle(color: Colors.red),
            ),
          ),
        ],
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;

  const _SectionHeader({Key? key, required this.title}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 4, bottom: 8),
      child: Text(
        title,
        style: TextStyle(
          color: Colors.white.withOpacity(0.5),
          fontSize: 12,
          fontWeight: FontWeight.bold,
          letterSpacing: 1.5,
        ),
      ),
    );
  }
}

class _GradientDivider extends StatelessWidget {
  const _GradientDivider({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 2,
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            const Color(0xFFFF006E).withOpacity(0.3),
            const Color(0xFFFF4D00).withOpacity(0.3),
            Colors.transparent,
          ],
        ),
        borderRadius: BorderRadius.circular(2),
      ),
    );
  }
}

class _AnimatedSettingsTile extends StatefulWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final Widget? trailing;
  final VoidCallback onTap;
  final Color? titleColor;

  const _AnimatedSettingsTile({
    Key? key,
    required this.icon,
    required this.title,
    required this.subtitle,
    this.trailing,
    required this.onTap,
    this.titleColor,
  }) : super(key: key);

  @override
  State<_AnimatedSettingsTile> createState() => _AnimatedSettingsTileState();
}

class _AnimatedSettingsTileState extends State<_AnimatedSettingsTile> {
  bool _isPressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTapDown: (_) => setState(() => _isPressed = true),
      onTapUp: (_) {
        setState(() => _isPressed = false);
        widget.onTap();
      },
      onTapCancel: () => setState(() => _isPressed = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        transform: Matrix4.translationValues(_isPressed ? 4 : 0, 0, 0),
        margin: const EdgeInsets.only(bottom: 8),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: const Color(0xFF1a1a1a),
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.2),
              blurRadius: 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Row(
          children: [
            Icon(widget.icon, color: Colors.white, size: 24),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    widget.title,
                    style: TextStyle(
                      color: widget.titleColor ?? Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    widget.subtitle,
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.6),
                      fontSize: 13,
                    ),
                  ),
                ],
              ),
            ),
            if (widget.trailing != null)
              widget.trailing!
            else
              const Icon(
                Icons.arrow_forward_ios,
                color: Colors.white30,
                size: 16,
              ),
          ],
        ),
      ),
    );
  }
}
