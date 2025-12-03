import 'package:flutter/material.dart';
import 'package:afftok/utils/app_localizations.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:url_launcher/url_launcher.dart';

class AboutScreen extends StatelessWidget {
  const AboutScreen({Key? key}) : super(key: key);

  Future<void> _launchURL(String url) async {
    try {
      final Uri uri = Uri.parse(url);
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } catch (e) {
      print('Error launching URL: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    
    final contactTitle = isArabic ? 'للتواصل' : 'Contact Us';

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: Text(lang.aboutScreenTitle),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(
              child: Image.asset(
                'assets/logo.png',
                width: 100,
                height: 100,
              ),
            ),
            const SizedBox(height: 20),
            Center(
              child: Text(
                lang.welcomeToAffTok,
                style: const TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
            ),
            const SizedBox(height: 10),
            Center(
              child: Text(
                lang.whereSharingPays,
                style: const TextStyle(
                  fontSize: 18,
                  color: Color(0xFFE91E63),
                ),
              ),
            ),
            const SizedBox(height: 30),

            _buildSection(lang.ourMission),
            _buildText(lang.missionText),

            _buildSection(lang.whatMakesAffTokSpecial),
            _buildFeature(Icons.attach_money, lang.realEarnings, lang.realEarningsDesc),
            _buildFeature(Icons.group, lang.teamPower, lang.teamPowerDesc),
            _buildFeature(Icons.trending_up, lang.transparentTracking, lang.transparentTrackingDesc),
            _buildFeature(Icons.public, lang.globalOpportunities, lang.globalOpportunitiesDesc),
            _buildFeature(Icons.emoji_events, lang.gamifiedExperience, lang.gamifiedExperienceDesc),

            _buildSection(lang.howItWorks),
            _buildStep('1', lang.discoverStep, lang.discoverStepDesc),
            _buildStep('2', lang.shareStep, lang.shareStepDesc),
            _buildStep('3', lang.earnStep, lang.earnStepDesc),
            _buildStep('4', lang.growStep, lang.growStepDesc),

            _buildSection(contactTitle),
            
            _buildContactLink(
              icon: Icons.email_outlined,
              title: 'support@afftokapp.com',
              url: 'mailto:support@afftokapp.com',
              onTap: () => _launchURL('mailto:support@afftokapp.com'),
            ),
            
            _buildContactLink(
              icon: FontAwesomeIcons.instagram,
              title: 'Afftok_app',
              url: 'https://instagram.com/Afftok_app',
              onTap: () => _launchURL('https://instagram.com/Afftok_app'),
            ),

            _buildContactLink(
              icon: FontAwesomeIcons.xTwitter,
              title: 'afftokapp',
              url: 'https://twitter.com/afftokapp',
              onTap: () => _launchURL('https://twitter.com/afftokapp'),
            ),

            _buildContactLink(
              icon: FontAwesomeIcons.tiktok,
              title: 'afftok_app',
              url: 'https://www.tiktok.com/@afftok_app',
              onTap: () => _launchURL('https://www.tiktok.com/@afftok_app'),
            ),
            
            _buildContactLink(
              icon: FontAwesomeIcons.youtube,
              title: 'afftok',
              url: 'https://youtube.com/@afftok',
              onTap: () => _launchURL('https://youtube.com/@afftok'),
            ),

            const SizedBox(height: 30),
            Center(
              child: Text(
                lang.versionText,
                style: const TextStyle(
                  color: Colors.white54,
                  fontSize: 12,
                ),
              ),
            ),
            const SizedBox(height: 10),
            Center(
              child: Text(
                lang.slogan,
                style: const TextStyle(
                  color: Colors.white54,
                  fontSize: 12,
                  fontStyle: FontStyle.italic,
                ),
              ),
            ),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }

  Widget _buildContactLink({
    required IconData icon,
    required String title,
    required String url,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: Row(
          children: [
            Icon(
              icon,
              color: const Color(0xFFE91E63),
              size: 20,
            ),
            const SizedBox(width: 12),
            Text(
              title,
              style: const TextStyle(
                fontSize: 16,
                color: Colors.white,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSection(String title) {
    return Padding(
      padding: const EdgeInsets.only(top: 25, bottom: 15),
      child: Text(
        title,
        style: const TextStyle(
          fontSize: 20,
          fontWeight: FontWeight.bold,
          color: Colors.white,
        ),
      ),
    );
  }

  Widget _buildText(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Text(
        text,
        style: const TextStyle(
          fontSize: 14,
          color: Colors.white70,
          height: 1.5,
        ),
      ),
    );
  }

  Widget _buildFeature(IconData icon, String title, String description) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: const Color(0xFFE91E63).withOpacity(0.2),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(icon, color: const Color(0xFFE91E63), size: 24),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  description,
                  style: const TextStyle(
                    fontSize: 14,
                    color: Colors.white70,
                    height: 1.4,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStep(String number, String title, String description) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [Color(0xFF9C27B0), Color(0xFFE91E63)],
              ),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Center(
              child: Text(
                number,
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  description,
                  style: const TextStyle(
                    fontSize: 14,
                    color: Colors.white70,
                    height: 1.4,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
