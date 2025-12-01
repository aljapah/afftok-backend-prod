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

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] ?? '',
      username: json['username'] ?? '',
      displayName: json['display_name'] ?? json['displayName'] ?? '',
      email: json['email'] ?? '',
      phone: json['phone'],
      avatarUrl: json['avatar_url'] ?? json['avatarUrl'],
      bio: json['bio'],
      level: UserLevel.values.firstWhere(
        (e) => e.name == (json['level'] ?? 'rookie'),
        orElse: () => UserLevel.rookie,
      ),
      stats: UserStats.fromJson(json['stats'] ?? {}),
      teamId: json['team_id'] ?? json['teamId'],
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
    );
  }

  String get personalLink => 'afftok.com/u/$username';
  
  bool get isInTeam => teamId != null;
  
  double get conversionRate => stats.conversionRate;
  
  int get totalClicks => stats.totalClicks;
  
  int get totalConversions => stats.totalConversions;
  
  double get totalEarnings => 0.0; // Will be calculated from conversions
  
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
  final int totalRegisteredOffers;
  final int monthlyClicks;
  final int monthlyConversions;
  final int globalRank;
  final Map<String, OfferStats> offerStats;

  UserStats({
    required this.totalClicks,
    required this.totalConversions,
    required this.totalRegisteredOffers,
    required this.monthlyClicks,
    required this.monthlyConversions,
    required this.globalRank,
    required this.offerStats,
  });

  factory UserStats.fromJson(Map<String, dynamic> json) {
    return UserStats(
      totalClicks: json['total_clicks'] ?? json['totalClicks'] ?? 0,
      totalConversions: json['total_conversions'] ?? json['totalConversions'] ?? 0,
      totalRegisteredOffers: json['total_registered_offers'] ?? json['totalRegisteredOffers'] ?? 0,
      monthlyClicks: json['monthly_clicks'] ?? json['monthlyClicks'] ?? 0,
      monthlyConversions: json['monthly_conversions'] ?? json['monthlyConversions'] ?? 0,
      globalRank: json['global_rank'] ?? json['globalRank'] ?? 0,
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



