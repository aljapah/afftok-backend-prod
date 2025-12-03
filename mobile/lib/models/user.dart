class User {
  final String id;
  final String username;
  final String displayName;
  final String email;
  final String? phone;
  final String? avatarUrl;
  final String? bio;
  final UserLevel level;
  final UserStats stats;
  final String? teamId;
  final DateTime createdAt;

  User({
    required this.id,
    required this.username,
    required this.displayName,
    required this.email,
    this.phone,
    this.avatarUrl,
    this.bio,
    required this.level,
    required this.stats,
    this.teamId,
    required this.createdAt,
  });

  /// Factory that handles both nested stats and separate stats object
  /// Backend returns: { user: {...}, stats: {...} }
  /// This factory handles: user object directly OR combined user+stats
  factory User.fromJson(Map<String, dynamic> json, {Map<String, dynamic>? externalStats}) {
    // Determine user level from total_conversions or level field
    final totalConversions = externalStats?['total_conversions'] ?? 
                             json['stats']?['total_conversions'] ?? 
                             json['total_conversions'] ?? 0;
    
    UserLevel userLevel;
    final levelStr = json['level'];
    if (levelStr is String) {
      userLevel = UserLevel.values.firstWhere(
        (e) => e.name == levelStr,
        orElse: () => _getLevelFromConversions(totalConversions),
      );
    } else {
      userLevel = _getLevelFromConversions(totalConversions);
    }
    
    // Build stats from external stats object or nested stats
    final statsData = externalStats ?? json['stats'] ?? {};
    
    return User(
      id: json['id']?.toString() ?? '',
      username: json['username'] ?? '',
      displayName: json['full_name'] ?? json['display_name'] ?? json['displayName'] ?? json['username'] ?? '',
      email: json['email'] ?? '',
      phone: json['phone'],
      avatarUrl: json['avatar_url'] ?? json['avatarUrl'],
      bio: json['bio'],
      level: userLevel,
      stats: UserStats.fromJson(statsData),
      teamId: json['team_id'] ?? json['teamId'],
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'].toString()) ?? DateTime.now()
          : DateTime.now(),
    );
  }
  
  static UserLevel _getLevelFromConversions(int conversions) {
    if (conversions >= 500) return UserLevel.legend;
    if (conversions >= 201) return UserLevel.master;
    if (conversions >= 51) return UserLevel.expert;
    if (conversions >= 11) return UserLevel.pro;
    return UserLevel.rookie;
  }

  String get personalLink => 'afftok.com/u/$username';
  
  bool get isInTeam => teamId != null;
  
  double get conversionRate => stats.conversionRate;
  
  int get totalClicks => stats.totalClicks;
  
  int get totalConversions => stats.totalConversions;
  
  double get totalEarnings => stats.totalEarnings.toDouble();
  
  String get fullName => displayName;
  
  String get userLevelEmoji => level.emoji;
  
  String get userLevel => level.displayName;
  
  int get points => totalConversions * 10; // 10 points per conversion
}

enum UserLevel {
  rookie,    // 0-10 referrals
  pro,       // 11-50 referrals
  expert,    // 51-200 referrals
  master,    // 201-500 referrals
  legend,    // 500+ referrals
}

extension UserLevelExtension on UserLevel {
  String get displayName {
    switch (this) {
      case UserLevel.rookie:
        return 'Rookie';
      case UserLevel.pro:
        return 'Pro';
      case UserLevel.expert:
        return 'Expert';
      case UserLevel.master:
        return 'Master';
      case UserLevel.legend:
        return 'Legend';
    }
  }

  String get emoji {
    switch (this) {
      case UserLevel.rookie:
        return '🌱';
      case UserLevel.pro:
        return '⭐';
      case UserLevel.expert:
        return '💎';
      case UserLevel.master:
        return '👑';
      case UserLevel.legend:
        return '🏆';
    }
  }

  int get minReferrals {
    switch (this) {
      case UserLevel.rookie:
        return 0;
      case UserLevel.pro:
        return 11;
      case UserLevel.expert:
        return 51;
      case UserLevel.master:
        return 201;
      case UserLevel.legend:
        return 500;
    }
  }
}

class UserStats {
  final int totalClicks;
  final int totalConversions;
  final int totalEarnings;
  final int totalRegisteredOffers;
  final int monthlyClicks;
  final int monthlyConversions;
  final int globalRank;
  final Map<String, OfferStats> offerStats;

  UserStats({
    required this.totalClicks,
    required this.totalConversions,
    required this.totalEarnings,
    required this.totalRegisteredOffers,
    required this.monthlyClicks,
    required this.monthlyConversions,
    required this.globalRank,
    required this.offerStats,
  });

  factory UserStats.fromJson(Map<String, dynamic> json) {
    // Parse numeric values safely (backend may send int or double)
    int parseIntSafe(dynamic value) {
      if (value == null) return 0;
      if (value is int) return value;
      if (value is double) return value.toInt();
      if (value is String) return int.tryParse(value) ?? 0;
      return 0;
    }
    
    return UserStats(
      totalClicks: parseIntSafe(json['total_clicks'] ?? json['totalClicks']),
      totalConversions: parseIntSafe(json['total_conversions'] ?? json['totalConversions']),
      totalEarnings: parseIntSafe(json['total_earnings'] ?? json['totalEarnings']),
      totalRegisteredOffers: parseIntSafe(json['total_registered_offers'] ?? json['totalRegisteredOffers']),
      monthlyClicks: parseIntSafe(json['monthly_clicks'] ?? json['monthlyClicks']),
      monthlyConversions: parseIntSafe(json['monthly_conversions'] ?? json['monthlyConversions']),
      globalRank: parseIntSafe(json['global_rank'] ?? json['globalRank']),
      offerStats: {},
    );
  }

  double get conversionRate {
    if (totalClicks == 0) return 0;
    return ((totalConversions / totalClicks) * 100);
  }

  String? get bestOffer {
    if (offerStats.isEmpty) return null;
    var sorted = offerStats.entries.toList()
      ..sort((a, b) => b.value.conversions.compareTo(a.value.conversions));
    return sorted.first.key;
  }
  
  int get totalReferrals => totalConversions;
  
  double get monthlyGrowth {
    if (totalConversions == 0) return 0;
    return ((monthlyConversions / totalConversions) * 100);
  }
}

class OfferStats {
  final String offerId;
  final int clicks;
  final int conversions;

  OfferStats({
    required this.offerId,
    required this.clicks,
    required this.conversions,
  });
}



