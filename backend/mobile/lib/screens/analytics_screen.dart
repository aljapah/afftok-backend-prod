import 'package:flutter/material.dart';
import '../models/user.dart';
import '../models/user_offer.dart';
import '../models/offer.dart';
import '../utils/app_localizations.dart';

class AnalyticsScreen extends StatefulWidget {
  const AnalyticsScreen({Key? key}) : super(key: key);

  @override
  State<AnalyticsScreen> createState() => _AnalyticsScreenState();
}

class _AnalyticsScreenState extends State<AnalyticsScreen> {
  String _selectedPeriod = 'all'; // all, month, week

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    final userOffers = getUserOffers(currentUser.id);
    final allOffers = Offer.getSampleOffers();

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        title: Text(
          lang.analytics,
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
      ),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(20.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Period selector
              _buildPeriodSelector(lang),

              const SizedBox(height: 24),

              // Overview cards
              Text(
                lang.overview,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 16),

              Row(
                children: [
                  Expanded(
                    child: _buildOverviewCard(
                      icon: Icons.touch_app,
                      label: lang.totalClicks,
                      value: '${currentUser.stats.totalClicks}',
                      color: Colors.blue,
                      trend: '+12%',
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _buildOverviewCard(
                      icon: Icons.people,
                      label: lang.totalConversions,
                      value: '${currentUser.stats.totalReferrals}',
                      color: Colors.green,
                      trend: '+8%',
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 12),

              Row(
                children: [
                  Expanded(
                    child: _buildOverviewCard(
                      icon: Icons.attach_money,
                      label: lang.totalEarnings,
                      value: '\$${currentUser.stats.totalRegisteredOffers.toStringAsFixed(0)}',
                      color: Colors.amber,
                      trend: '+15%',
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _buildOverviewCard(
                      icon: Icons.trending_up,
                      label: lang.conversionRate,
                      value: '${(currentUser.stats.totalConversions/ currentUser.stats.totalClicks * 100).toStringAsFixed(1)}%',
                      color: Colors.purple,
                      trend: '+3%',
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 32),

              // Performance chart placeholder
              Text(
                lang.performanceChart,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 16),

              Container(
                height: 200,
                decoration: BoxDecoration(
                  color: Colors.grey[900],
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: Colors.grey[800]!),
                ),
                child: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.show_chart,
                        size: 60,
                        color: Colors.grey[700],
                      ),
                      const SizedBox(height: 12),
                      Text(
                        lang.chartPlaceholder,
                        style: TextStyle(
                          color: Colors.grey[600],
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 32),

              // Top performing offers
              Text(
                lang.topPerformingOffers,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 16),

              ...userOffers.take(5).map((userOffer) {
                final offer = allOffers.firstWhere(
                  (o) => o.id == userOffer.offerId,
                  orElse: () => allOffers.first,
                );
                return _buildOfferPerformanceCard(offer, userOffer, lang);
              }).toList(),

              const SizedBox(height: 32),

              // Additional stats
              Text(
                lang.additionalStats,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 16),

              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.grey[900],
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: Colors.grey[800]!),
                ),
                child: Column(
                  children: [
                    _buildStatRow(
                      icon: Icons.emoji_events,
                      label: lang.globalRank,
                      value: '#${currentUser.stats.globalRank}',
                      color: Colors.amber,
                    ),
                    const Divider(color: Colors.grey, height: 32),
                    _buildStatRow(
                      icon: Icons.inventory_2,
                      label: lang.activeOffers,
                      value: '${userOffers.length}',
                      color: Colors.blue,
                    ),
                    const Divider(color: Colors.grey, height: 32),
                    _buildStatRow(
                      icon: Icons.calendar_month,
                      label: lang.monthlyEarnings,
                      value: '\$${currentUser.stats.monthlyConversions.toStringAsFixed(0)}',
                      color: Colors.green,
                    ),
                    const Divider(color: Colors.grey, height: 32),
                    _buildStatRow(
                      icon: Icons.trending_up,
                      label: lang.growthRate,
                      value: '${currentUser.stats.conversionRate.toStringAsFixed(1)}%',
                      color: Colors.purple,
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildPeriodSelector(AppLocalizations lang) {
    return Container(
      padding: const EdgeInsets.all(4),
      decoration: BoxDecoration(
        color: Colors.grey[900],
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          Expanded(
            child: _buildPeriodButton(lang.allTime, 'all'),
          ),
          Expanded(
            child: _buildPeriodButton(lang.thisMonth, 'month'),
          ),
          Expanded(
            child: _buildPeriodButton(lang.thisWeek, 'week'),
          ),
        ],
      ),
    );
  }

  Widget _buildPeriodButton(String label, String value) {
    final isSelected = _selectedPeriod == value;
    return GestureDetector(
      onTap: () => setState(() => _selectedPeriod = value),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xFF8E2DE2) : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Text(
          label,
          textAlign: TextAlign.center,
          style: TextStyle(
            color: isSelected ? Colors.white : Colors.grey[400],
            fontSize: 14,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
          ),
        ),
      ),
    );
  }

  Widget _buildOverviewCard({
    required IconData icon,
    required String label,
    required String value,
    required Color color,
    required String trend,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey[900],
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey[800]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Icon(icon, color: color, size: 24),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.green[900]?.withOpacity(0.3),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  trend,
                  style: TextStyle(
                    color: Colors.green[400],
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            value,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 24,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            label,
            style: TextStyle(
              color: Colors.grey[400],
              fontSize: 12,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildOfferPerformanceCard(
    Offer offer,
    UserOffer userOffer,
    AppLocalizations lang,
  ) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey[900],
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey[800]!),
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(8),
            ),
            padding: const EdgeInsets.all(6),
            child: Image.network(
              offer.logoUrl,
              errorBuilder: (context, error, stackTrace) {
                return Center(
                  child: Text(
                    offer.companyName[0],
                    style: const TextStyle(
                      fontSize: 16,
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
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  '${userOffer.stats.clicks} ${lang.clicks} • ${userOffer.stats.conversions} ${lang.conversions}',
                  style: TextStyle(
                    color: Colors.grey[400],
                    fontSize: 12,
                  ),
                ),
              ],
            ),
          ),
          Text(
            '\$${userOffer.stats.earnings.toStringAsFixed(0)}',
            style: const TextStyle(
              color: Colors.amber,
              fontSize: 16,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatRow({
    required IconData icon,
    required String label,
    required String value,
    required Color color,
  }) {
    return Row(
      children: [
        Icon(icon, color: color, size: 24),
        const SizedBox(width: 12),
        Expanded(
          child: Text(
            label,
            style: TextStyle(
              color: Colors.grey[300],
              fontSize: 15,
            ),
          ),
        ),
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 16,
            fontWeight: FontWeight.bold,
          ),
        ),
      ],
    );
  }
}

