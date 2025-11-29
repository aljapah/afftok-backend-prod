// Model for tracking offers added by users
class UserOffer {
  final String id;
  final String userId;
  final String offerId;
  final String userReferralLink; // The user's unique referral link for this offer
  final DateTime addedAt;
  final UserOfferStats stats;
  final bool isActive;

  UserOffer({
    required this.id,
    required this.userId,
    required this.offerId,
    required this.userReferralLink,
    required this.addedAt,
    required this.stats,
    this.isActive = true,
  });

  // Calculate conversion rate
  double get conversionRate {
    if (stats.clicks == 0) return 0;
    return (stats.conversions / stats.clicks) * 100;
  }

  // Getters for compatibility
  int get clicks => stats.clicks;
  int get conversions => stats.conversions;
  double get earnings => 0.0; // Earnings removed
  String get affiliateLink => userReferralLink;
  int get totalClicks => stats.clicks;
  int get totalConversions => stats.conversions;
  
  // fromJson factory
  factory UserOffer.fromJson(Map<String, dynamic> json) {
    return UserOffer(
      id: json['id'] ?? '',
      userId: json['user_id'] ?? json['userId'] ?? '',
      offerId: json['offer_id'] ?? json['offerId'] ?? '',
      userReferralLink: json['user_referral_link'] ?? json['userReferralLink'] ?? json['affiliate_link'] ?? '',
      addedAt: json['added_at'] != null ? DateTime.parse(json['added_at']) : DateTime.now(),
      stats: UserOfferStats.fromJson(json['stats'] ?? {}),
      isActive: json['is_active'] ?? json['isActive'] ?? true,
    );
  }
}

class UserOfferStats {
  final int clicks;
  final int conversions;
  final int shares;
  final DateTime lastActivity;

  UserOfferStats({
    required this.clicks,
    required this.conversions,
    required this.shares,
    required this.lastActivity,
  });
  
  double get earnings => conversions * 5.0; // $5 per conversion
  
  factory UserOfferStats.fromJson(Map<String, dynamic> json) {
    return UserOfferStats(
      clicks: json['clicks'] ?? 0,
      conversions: json['conversions'] ?? 0,
      shares: json['shares'] ?? 0,
      lastActivity: json['last_activity'] != null ? DateTime.parse(json['last_activity']) : DateTime.now(),
    );
  }

  // Copy with method for updating stats
  UserOfferStats copyWith({
    int? clicks,
    int? conversions,
    // double? earnings, // ❌ تم إزالته
    int? shares,
    DateTime? lastActivity,
  }) {
    return UserOfferStats(
      clicks: clicks ?? this.clicks,
      conversions: conversions ?? this.conversions,
      // earnings: earnings ?? this.earnings, // ❌ تم إزالته
      shares: shares ?? this.shares,
      lastActivity: lastActivity ?? this.lastActivity,
    );
  }
}

// Sample data for testing
final List<UserOffer> sampleUserOffers = [
  UserOffer(
    id: 'uo_001',
    userId: 'user_001',
    offerId: '1', // Binance
    userReferralLink: 'https://www.binance.com/en/activity/referral?ref=ABC123',
    addedAt: DateTime(2024, 9, 1),
    stats: UserOfferStats(
      clicks: 450,
      conversions: 25,
      // earnings removed
      shares: 120,
      lastActivity: DateTime.now().subtract(const Duration(hours: 2)),
    ),
  ),
  UserOffer(
    id: 'uo_002',
    userId: 'user_001',
    offerId: '5', // Amazon
    userReferralLink: 'https://www.amazon.com/?tag=abomohammed-20',
    addedAt: DateTime(2024, 9, 5),
    stats: UserOfferStats(
      clicks: 280,
      conversions: 12,
      // earnings removed
      shares: 85,
      lastActivity: DateTime.now().subtract(const Duration(hours: 5)),
    ),
  ),
  UserOffer(
    id: 'uo_003',
    userId: 'user_001',
    offerId: '10', // Uber
    userReferralLink: 'https://www.uber.com/invite/abomohammed',
    addedAt: DateTime(2024, 9, 10),
    stats: UserOfferStats(
      clicks: 180,
      conversions: 8,
      // earnings removed
      shares: 60,
      lastActivity: DateTime.now().subtract(const Duration(days: 1)),
    ),
  ),
  UserOffer(
    id: 'uo_004',
    userId: 'user_001',
    offerId: '14', // PayPal
    userReferralLink: 'https://www.paypal.com/invite/abomohammed',
    addedAt: DateTime(2024, 9, 15),
    stats: UserOfferStats(
      clicks: 95,
      conversions: 5,
      // earnings removed
      shares: 30,
      lastActivity: DateTime.now().subtract(const Duration(days: 2)),
    ),
  ),
  UserOffer(
    id: 'uo_005',
    userId: 'user_001',
    offerId: '18', // Airbnb
    userReferralLink: 'https://www.airbnb.com/c/abomohammed',
    addedAt: DateTime(2024, 9, 20),
    stats: UserOfferStats(
      clicks: 245,
      conversions: 10,
      // earnings removed
      shares: 95,
      lastActivity: DateTime.now().subtract(const Duration(hours: 12)),
    ),
  ),
];

// Helper function to check if user has added an offer
bool hasUserAddedOffer(String userId, String offerId) {
  return sampleUserOffers.any(
    (userOffer) => userOffer.userId == userId && userOffer.offerId == offerId,
  );
}

// Helper function to get user's referral link for an offer
String? getUserReferralLink(String userId, String offerId) {
  try {
    final userOffer = sampleUserOffers.firstWhere(
      (uo) => uo.userId == userId && uo.offerId == offerId,
    );
    return userOffer.userReferralLink;
  } catch (e) {
    return null;
  }
}

// Helper function to get all offers added by user
List<UserOffer> getUserOffers(String userId) {
  return sampleUserOffers.where((uo) => uo.userId == userId).toList()
    ..sort((a, b) => b.addedAt.compareTo(a.addedAt)); // Sort by most recent
}

// Helper function to get user offer by ID
UserOffer? getUserOfferById(String userOfferId) {
  try {
    return sampleUserOffers.firstWhere((uo) => uo.id == userOfferId);
  } catch (e) {
    return null;
  }
}

