import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../models/offer.dart';
import '../models/user_offer.dart';
import '../utils/app_localizations.dart';

class DownloadPromptScreen extends StatelessWidget {
  final Offer offer;
  final UserOffer userOffer;
  final String promoterUsername;

  const DownloadPromptScreen({
    Key? key,
    required this.offer,
    required this.userOffer,
    required this.promoterUsername,
  }) : super(key: key);

  Future<void> _continueAsGuest() async {
    final Uri uri = Uri.parse(userOffer.userReferralLink);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  void _downloadApp() {
    // TODO: Navigate to App Store / Play Store
    // For now, just show a message
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);

    return Scaffold(
      backgroundColor: Colors.black,
      body: SafeArea(
        child: SingleChildScrollView(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Close button
                Align(
                  alignment: Alignment.topRight,
                  child: IconButton(
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(Icons.close, color: Colors.white),
                  ),
                ),

                const SizedBox(height: 20),

                // App logo/icon
                Container(
                  width: 120,
                  height: 120,
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFF8E2DE2), Color(0xFF4A00E0)],
                    ),
                    borderRadius: BorderRadius.circular(30),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF8E2DE2).withOpacity(0.5),
                        blurRadius: 30,
                        spreadRadius: 5,
                      ),
                    ],
                  ),
                  child: Center(
                    child: Image.asset(
                      'assets/logo.png',
                      width: 80,
                      height: 80,
                      errorBuilder: (context, error, stackTrace) {
                        return const Icon(
                          Icons.card_giftcard,
                          size: 60,
                          color: Colors.white,
                        );
                      },
                    ),
                  ),
                ),

                const SizedBox(height: 32),

                // Main heading
                Text(
                  lang.getTheBestExperience,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 28,
                    fontWeight: FontWeight.bold,
                    height: 1.2,
                  ),
                  textAlign: TextAlign.center,
                ),

                const SizedBox(height: 16),

                // Subheading
                Text(
                  lang.downloadAppDescription,
                  style: TextStyle(
                    color: Colors.grey[400],
                    fontSize: 16,
                    height: 1.5,
                  ),
                  textAlign: TextAlign.center,
                ),

                const SizedBox(height: 40),

                // Benefits
                _buildBenefit(
                  icon: Icons.flash_on,
                  title: lang.instantAccess,
                  description: lang.instantAccessDescription,
                ),
                const SizedBox(height: 20),
                _buildBenefit(
                  icon: Icons.track_changes,
                  title: lang.trackEarnings,
                  description: lang.trackEarningsDescription,
                ),
                const SizedBox(height: 20),
                _buildBenefit(
                  icon: Icons.notifications_active,
                  title: lang.getNotified,
                  description: lang.getNotifiedDescription,
                ),

                const SizedBox(height: 40),

                // Offer preview
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.grey[900],
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: Colors.grey[800]!),
                  ),
                  child: Column(
                    children: [
                      Text(
                        lang.youreAboutToAccess,
                        style: TextStyle(
                          color: Colors.grey[400],
                          fontSize: 14,
                        ),
                      ),
                      const SizedBox(height: 12),
                      Row(
                        children: [
                          Container(
                            width: 50,
                            height: 50,
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(10),
                            ),
                            padding: const EdgeInsets.all(8),
                            child: Image.network(
                              offer.logoUrl,
                              errorBuilder: (context, error, stackTrace) {
                                return Center(
                                  child: Text(
                                    offer.companyName[0],
                                    style: const TextStyle(
                                      fontSize: 20,
                                      fontWeight: FontWeight.bold,
                                      color: Colors.black87,
                                    ),
                                  ),
                                );
                              },
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  offer.companyName,
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  offer.reward,
                                  style: const TextStyle(
                                    color: Color(0xFF8E2DE2),
                                    fontSize: 14,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 40),

                // Download button
                ElevatedButton(
                  onPressed: _downloadApp,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF8E2DE2),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 18),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    elevation: 5,
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.download, size: 24),
                      const SizedBox(width: 12),
                      Text(
                        lang.downloadAffTok,
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 16),

                // Continue as guest
                TextButton(
                  onPressed: _continueAsGuest,
                  child: Text(
                    lang.continueAsGuest,
                    style: TextStyle(
                      color: Colors.grey[400],
                      fontSize: 16,
                      decoration: TextDecoration.underline,
                    ),
                  ),
                ),

                const SizedBox(height: 24),

                // Referral info
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.blue[900]?.withOpacity(0.2),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: Colors.blue[700]!.withOpacity(0.3),
                    ),
                  ),
                  child: Row(
                    children: [
                      Icon(
                        Icons.info_outline,
                        color: Colors.blue[300],
                        size: 20,
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          '${lang.recommendedBy} @$promoterUsername',
                          style: TextStyle(
                            color: Colors.blue[100],
                            fontSize: 13,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildBenefit({
    required IconData icon,
    required String title,
    required String description,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: [Color(0xFF8E2DE2), Color(0xFF4A00E0)],
            ),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(icon, color: Colors.white, size: 24),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                description,
                style: TextStyle(
                  color: Colors.grey[400],
                  fontSize: 14,
                  height: 1.4,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

