// Model for tracking offers added by users
// Updated to support new secure tracking links from backend
class UserOffer {
  final String id;
  final String userId;
  final String offerId;
  final String userReferralLink; // The user's unique referral link for this offer
  final String? shortLink; // Short tracking code for sharing
  final String? trackingUrl; // Full tracking URL for clicks
  final DateTime addedAt;
  final UserOfferStats stats;
  final bool isActive;
  final OfferInfo? offerInfo; // Nested offer details from backend

  UserOffer({
    required this.id,
    required this.userId,
    required this.offerId,
    required this.userReferralLink,
    this.shortLink,
    this.trackingUrl,
    required this.addedAt,
    required this.stats,
    this.isActive = true,
    this.offerInfo,
  });

  // Calculate conversion rate
  double get conversionRate {
    if (stats.clicks == 0) return 0;
    return (stats.conversions / stats.clicks) * 100;
  }

  // Getters for compatibility
  int get clicks => stats.clicks;
  int get conversions => stats.conversions;
  double get earnings => stats.earnings;
  String get affiliateLink => userReferralLink;
  int get totalClicks => stats.clicks;
  int get totalConversions => stats.conversions;
  
  // Get the best link to share (prefer short link if available)
  String get shareableLink => shortLink ?? userReferralLink;
  
  // Get full tracking URL for clicks
  String get fullTrackingUrl => trackingUrl ?? '/api/c/$offerId?promoter=$userId';
  
  // Get offer title from nested offer or fallback
  String get offerTitle => offerInfo?.title ?? 'Offer';
  String get offerLogoUrl => offerInfo?.logoUrl ?? '';
  String get offerCategory => offerInfo?.category ?? '';
  String get offerImageUrl => offerInfo?.imageUrl ?? '';
  
  // fromJson factory - handles backend response structure
  // Backend returns: { id, user_id, offer_id, affiliate_link, short_link, status, joined_at, offer: {...}, stats: {...} }
  factory UserOffer.fromJson(Map<String, dynamic> json) {
    final offerData = json['offer'] as Map<String, dynamic>?;
    
    // Parse status to boolean
    final status = json['status']?.toString().toLowerCase();
    final isActive = status == 'active' || status == null;
    
    // Build stats from response - check both nested stats and direct fields
    final statsData = json['stats'] as Map<String, dynamic>? ?? {};
    
    // Also check for direct click/conversion counts (backend sends both)
    if (statsData.isEmpty) {
      // Try to get from direct fields
      final directClicks = json['total_clicks'] ?? 0;
      final directConversions = json['total_conversions'] ?? 0;
      if (directClicks > 0 || directConversions > 0) {
        statsData['clicks'] = directClicks;
        statsData['conversions'] = directConversions;
      }
    }
    
    return UserOffer(
      id: json['id']?.toString() ?? '',
      userId: json['user_id']?.toString() ?? json['userId']?.toString() ?? '',
      offerId: json['offer_id']?.toString() ?? json['offerId']?.toString() ?? offerData?['id']?.toString() ?? '',
      userReferralLink: json['affiliate_link'] ?? json['user_referral_link'] ?? json['userReferralLink'] ?? '',
      shortLink: json['short_link']?.toString(),
      trackingUrl: json['tracking_url']?.toString(),
      addedAt: _parseDateTime(json['joined_at'] ?? json['added_at']),
      stats: UserOfferStats.fromJson(statsData),
      isActive: isActive,
      offerInfo: offerData != null ? OfferInfo.fromJson(offerData) : null,
    );
  }
  
  static DateTime _parseDateTime(dynamic value) {
    if (value == null) return DateTime.now();
    if (value is DateTime) return value;
    if (value is String) {
      return DateTime.tryParse(value) ?? DateTime.now();
    }
    return DateTime.now();
  }
  
  // Copy with method for state updates
  UserOffer copyWith({
    String? id,
    String? userId,
    String? offerId,
    String? userReferralLink,
    String? shortLink,
    String? trackingUrl,
    DateTime? addedAt,
    UserOfferStats? stats,
    bool? isActive,
    OfferInfo? offerInfo,
  }) {
    return UserOffer(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      offerId: offerId ?? this.offerId,
      userReferralLink: userReferralLink ?? this.userReferralLink,
      shortLink: shortLink ?? this.shortLink,
      trackingUrl: trackingUrl ?? this.trackingUrl,
      addedAt: addedAt ?? this.addedAt,
      stats: stats ?? this.stats,
      isActive: isActive ?? this.isActive,
      offerInfo: offerInfo ?? this.offerInfo,
    );
  }
}

// Lightweight offer info from nested offer in UserOffer response
class OfferInfo {
  final String id;
  final String title;
  final String? description;
  final String? logoUrl;
  final String? imageUrl;
  final String? category;
  final int totalClicks;
  final int totalConversions;

  OfferInfo({
    required this.id,
    required this.title,
    this.description,
    this.logoUrl,
    this.imageUrl,
    this.category,
    this.totalClicks = 0,
    this.totalConversions = 0,
  });

  factory OfferInfo.fromJson(Map<String, dynamic> json) {
    return OfferInfo(
      id: json['id']?.toString() ?? '',
      title: json['title'] ?? 'Offer',
      description: json['description'],
      logoUrl: json['logo_url'],
      imageUrl: json['image_url'],
      category: json['category'],
      totalClicks: _parseIntSafe(json['total_clicks']),
      totalConversions: _parseIntSafe(json['total_conversions']),
    );
  }
  
  static int _parseIntSafe(dynamic value) {
    if (value == null) return 0;
    if (value is int) return value;
    if (value is double) return value.toInt();
    if (value is String) return int.tryParse(value) ?? 0;
    return 0;
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
      clicks: _parseIntSafe(json['clicks']),
      conversions: _parseIntSafe(json['conversions']),
      shares: _parseIntSafe(json['shares']),
      lastActivity: _parseDateTime(json['last_activity']),
    );
  }
  
  static int _parseIntSafe(dynamic value) {
    if (value == null) return 0;
    if (value is int) return value;
    if (value is double) return value.toInt();
    if (value is String) return int.tryParse(value) ?? 0;
    return 0;
  }
  
  static DateTime _parseDateTime(dynamic value) {
    if (value == null) return DateTime.now();
    if (value is DateTime) return value;
    if (value is String) {
      return DateTime.tryParse(value) ?? DateTime.now();
    }
    return DateTime.now();
  }

  // Copy with method for updating stats
  UserOfferStats copyWith({
    int? clicks,
    int? conversions,
    int? shares,
    DateTime? lastActivity,
  }) {
    return UserOfferStats(
      clicks: clicks ?? this.clicks,
      conversions: conversions ?? this.conversions,
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

