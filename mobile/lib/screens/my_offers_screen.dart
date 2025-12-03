import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';
import 'package:provider/provider.dart';
import '../models/user_offer.dart';
import '../utils/app_localizations.dart';
import '../providers/auth_provider.dart';

class MyOffersScreen extends StatefulWidget {
  const MyOffersScreen({Key? key}) : super(key: key);

  @override
  State<MyOffersScreen> createState() => _MyOffersScreenState();
}

class _MyOffersScreenState extends State<MyOffersScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _refreshOffers();
    });
  }

  Future<void> _refreshOffers() async {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    await authProvider.reloadUserOffers();
  }

  void _copyLink(BuildContext context, String link) {
    Clipboard.setData(ClipboardData(text: link));
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(AppLocalizations.of(context).linkCopied),
        backgroundColor: Colors.green,
        duration: const Duration(seconds: 2),
      ),
    );
  }

  void _shareOffer(String companyName, String link) {
    Share.share(
      'Check out this amazing offer from $companyName! 🎁\n$link',
      subject: 'Amazing offer on AffTok',
    );
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        final user = authProvider.currentUser;
        
        if (user == null) {
          return Scaffold(
            backgroundColor: Colors.black,
            appBar: AppBar(
              backgroundColor: Colors.transparent,
              title: Text(lang.myOffers),
            ),
            body: const Center(
              child: CircularProgressIndicator(),
            ),
          );
        }
        
        // Use offers from AuthProvider instead of sample data
        final userOffers = authProvider.userOffers;

        return Scaffold(
          backgroundColor: Colors.black,
          appBar: AppBar(
            backgroundColor: Colors.transparent,
            elevation: 0,
            title: Text(
              lang.myOffers,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
            leading: IconButton(
              icon: const Icon(Icons.arrow_back, color: Colors.white),
              onPressed: () => Navigator.pop(context),
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.refresh, color: Colors.white),
                onPressed: _refreshOffers,
              ),
            ],
          ),
          body: RefreshIndicator(
            onRefresh: _refreshOffers,
            color: const Color(0xFF8E2DE2),
            backgroundColor: Colors.grey[900],
            child: userOffers.isEmpty
                ? _buildEmptyState(context, lang)
                : ListView.builder(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.all(16),
                    itemCount: userOffers.length,
                    itemBuilder: (context, index) {
                      final userOffer = userOffers[index];
                      return _buildUserOfferCard(context, userOffer, lang);
                    },
                  ),
          ),
        );
      },
    );
  }

  Widget _buildEmptyState(BuildContext context, AppLocalizations lang) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.inventory_2_outlined,
              size: 100,
              color: Colors.grey[700],
            ),
            const SizedBox(height: 24),
            Text(
              lang.noOffersAdded,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 22,
                fontWeight: FontWeight.bold,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            Text(
              lang.startAddingOffers,
              style: TextStyle(
                color: Colors.grey[400],
                fontSize: 16,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            ElevatedButton.icon(
              onPressed: () => Navigator.pop(context),
              icon: const Icon(Icons.explore),
              label: Text(lang.exploreOffers),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF8E2DE2),
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(
                  horizontal: 32,
                  vertical: 16,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildUserOfferCard(
    BuildContext context,
    UserOffer userOffer,
    AppLocalizations lang,
  ) {
    // Get offer info from the userOffer (populated by backend)
    final offerTitle = userOffer.offerInfo?.title ?? userOffer.offerTitle;
    final offerLogo = userOffer.offerInfo?.logoUrl ?? '';
    final offerCategory = userOffer.offerInfo?.category ?? '';
    
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.grey[900],
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey[800]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header with company info
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                // Company logo
                Container(
                  width: 50,
                  height: 50,
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  padding: const EdgeInsets.all(8),
                  child: offerLogo.isNotEmpty
                      ? Image.network(
                          offerLogo,
                          errorBuilder: (context, error, stackTrace) {
                            return Center(
                              child: Text(
                                offerTitle.isNotEmpty ? offerTitle[0] : '?',
                                style: const TextStyle(
                                  fontSize: 20,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.black87,
                                ),
                              ),
                            );
                          },
                        )
                      : Center(
                          child: Text(
                            offerTitle.isNotEmpty ? offerTitle[0] : '?',
                            style: const TextStyle(
                              fontSize: 20,
                              fontWeight: FontWeight.bold,
                              color: Colors.black87,
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
                        offerTitle,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      if (offerCategory.isNotEmpty) ...[
                        const SizedBox(height: 4),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            gradient: const LinearGradient(
                              colors: [Color(0xFF8E2DE2), Color(0xFF4A00E0)],
                            ),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            offerCategory,
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: userOffer.isActive
                        ? Colors.green[900]?.withOpacity(0.3)
                        : Colors.red[900]?.withOpacity(0.3),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(
                    userOffer.isActive ? Icons.check_circle : Icons.pause_circle,
                    color: userOffer.isActive ? Colors.green[400] : Colors.red[400],
                    size: 24,
                  ),
                ),
              ],
            ),
          ),

          const Divider(color: Colors.grey, height: 1),

          // Stats
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildStat(
                  icon: Icons.touch_app,
                  label: lang.clicks,
                  value: '${userOffer.stats.clicks}',
                  color: Colors.blue,
                ),
                _buildStat(
                  icon: Icons.people,
                  label: lang.conversions,
                  value: '${userOffer.stats.conversions}',
                  color: Colors.green,
                ),
                _buildStat(
                  icon: Icons.attach_money,
                  label: lang.earnings,
                  value: '\$${userOffer.stats.earnings.toStringAsFixed(0)}',
                  color: Colors.amber,
                ),
              ],
            ),
          ),

          const Divider(color: Colors.grey, height: 1),

          // Action buttons
          Padding(
            padding: const EdgeInsets.all(12),
            child: Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () => _copyLink(context, userOffer.userReferralLink),
                    icon: const Icon(Icons.copy, size: 18),
                    label: Text(lang.copyLink),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: Colors.white,
                      side: const BorderSide(color: Colors.grey),
                      padding: const EdgeInsets.symmetric(vertical: 12),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () => _shareOffer(
                      offerTitle,
                      userOffer.userReferralLink,
                    ),
                    icon: const Icon(Icons.share, size: 18),
                    label: Text(lang.share),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: Colors.white,
                      side: const BorderSide(color: Colors.grey),
                      padding: const EdgeInsets.symmetric(vertical: 12),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
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

  Widget _buildStat({
    required IconData icon,
    required String label,
    required String value,
    required Color color,
  }) {
    return Column(
      children: [
        Icon(icon, color: color, size: 24),
        const SizedBox(height: 6),
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 18,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 2),
        Text(
          label,
          style: TextStyle(
            color: Colors.grey[400],
            fontSize: 11,
          ),
        ),
      ],
    );
  }
}

