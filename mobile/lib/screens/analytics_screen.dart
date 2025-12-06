import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../utils/app_localizations.dart';
import '../providers/auth_provider.dart';

class AnalyticsScreen extends StatefulWidget {
  const AnalyticsScreen({Key? key}) : super(key: key);

  @override
  State<AnalyticsScreen> createState() => _AnalyticsScreenState();
}

class _AnalyticsScreenState extends State<AnalyticsScreen> {
  String _selectedPeriod = 'all';

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
              title: Text(lang.analytics),
            ),
            body: const Center(
              child: CircularProgressIndicator(),
            ),
          );
        }

        final totalClicks = authProvider.totalClicks;
        final totalConversions = authProvider.totalConversions;
        final conversionRate = authProvider.overallConversionRate;
        final offersCount = authProvider.totalOffersAdded;

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
                  _buildPeriodSelector(lang),
                  const SizedBox(height: 24),
                  Text(
                    'Overview',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _buildOverviewCard(
                    icon: Icons.visibility,
                    label: lang.visits,
                    value: '$totalClicks',
                    color: Colors.blue,
                  ),
                  const SizedBox(height: 12),
                  _buildOverviewCard(
                    icon: Icons.people,
                    label: lang.conversions,
                    value: '$totalConversions',
                    color: Colors.green,
                  ),
                  const SizedBox(height: 12),
                  _buildOverviewCard(
                    icon: Icons.trending_up,
                    label: lang.conversionRate,
                    value: '${conversionRate.toStringAsFixed(1)}%',
                    color: Colors.orange,
                  ),
                  const SizedBox(height: 12),
                  _buildOverviewCard(
                    icon: Icons.inventory_2,
                    label: 'Active Offers',
                    value: '$offersCount',
                    color: Colors.purple,
                  ),
                  const SizedBox(height: 32),
                  Text(
                    'Offers Performance',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _buildOffersPerformance(authProvider, lang),
                  const SizedBox(height: 32),
                  Text(
                    'Monthly Performance',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _buildStatRow(
                    icon: Icons.visibility,
                    label: lang.visits,
                    value: '${user.stats.monthlyClicks}',
                    color: Colors.blue,
                  ),
                  const SizedBox(height: 8),
                  _buildStatRow(
                    icon: Icons.people,
                    label: lang.conversions,
                    value: '${user.stats.monthlyConversions}',
                    color: Colors.green,
                  ),
                  const SizedBox(height: 8),
                  _buildStatRow(
                    icon: Icons.trending_up,
                    label: 'Monthly Growth',
                    value: '${user.stats.monthlyGrowth.toStringAsFixed(1)}%',
                    color: Colors.orange,
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

  Widget _buildOffersPerformance(AuthProvider authProvider, AppLocalizations lang) {
    final userOffers = authProvider.userOffers;
    
    if (userOffers.isEmpty) {
      return Container(
        padding: const EdgeInsets.symmetric(vertical: 24),
        child: Center(
          child: Text(
            lang.noOffersAdded,
            style: TextStyle(
              color: Colors.white.withOpacity(0.5),
              fontSize: 14,
            ),
          ),
        ),
      );
    }
    
    return Column(
      children: userOffers.map((offer) {
        return Column(
          children: [
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.05),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.white.withOpacity(0.1)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Offer ${offer.offerId}',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Clicks: ${offer.stats.clicks}',
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.7),
                          fontSize: 12,
                        ),
                      ),
                      Text(
                        'Conversions: ${offer.stats.conversions}',
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.7),
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Text(
                    'Rate: ${offer.conversionRate.toStringAsFixed(1)}%',
                    style: const TextStyle(
                      color: Color(0xFF00FF88),
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 8),
          ],
        );
      }).toList(),
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
        padding: const EdgeInsets.symmetric(vertical: 10),
        decoration: BoxDecoration(
          color: isSelected ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(10),
        ),
        child: Text(
          label,
          textAlign: TextAlign.center,
          style: TextStyle(
            color: isSelected ? Colors.black : Colors.white,
            fontWeight: FontWeight.bold,
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
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        children: [
          Icon(icon, color: color, size: 32),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: TextStyle(
                    color: Colors.grey[400],
                    fontSize: 12,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  value,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 24,
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

  Widget _buildStatRow({
    required IconData icon,
    required String label,
    required String value,
    required Color color,
  }) {
    return Row(
      children: [
        Icon(icon, color: color, size: 16),
        const SizedBox(width: 8),
        Text(
          label,
          style: TextStyle(
            color: Colors.grey[400],
            fontSize: 12,
          ),
        ),
        const Spacer(),
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 14,
            fontWeight: FontWeight.bold,
          ),
        ),
      ],
    );
  }
}
