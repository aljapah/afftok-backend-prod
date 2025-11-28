class Team {
  final String id;
  final String name;
  final String? logoUrl;
  final String? description;
  final TeamRank rank;
  final List<TeamMember> members;
  final TeamStats stats;
  final int maxMembers;
  final DateTime createdAt;

  Team({
    required this.id,
    required this.name,
    this.logoUrl,
    this.description,
    required this.rank,
    required this.members,
    required this.stats,
    this.maxMembers = 10,
    required this.createdAt,
  });

  int get memberCount => members.length;
  bool get isFull => memberCount >= maxMembers;
  TeamMember? get leader => members.firstWhere(
        (m) => m.isLeader,
        orElse: () => members.first,
      );

  // Factory constructor to parse from API response
  factory Team.fromJson(Map<String, dynamic> json) {
    return Team(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      logoUrl: json['logo_url'],
      description: json['description'],
      rank: _parseRank(json['total_points'] ?? 0),
      members: (json['members'] as List<dynamic>?)
              ?.map((m) => TeamMember.fromJson(m))
              .toList() ??
          [],
      stats: TeamStats.fromJson(json['stats'] ?? {}),
      maxMembers: json['max_members'] ?? 10,
      createdAt: DateTime.parse(json['created_at'] ?? DateTime.now().toIso8601String()),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'logo_url': logoUrl,
      'description': description,
      'max_members': maxMembers,
      'created_at': createdAt.toIso8601String(),
    };
  }

  static TeamRank _parseRank(int points) {
    if (points >= 1000) return TeamRank.diamond;
    if (points >= 500) return TeamRank.platinum;
    if (points >= 250) return TeamRank.gold;
    if (points >= 100) return TeamRank.silver;
    return TeamRank.bronze;
  }
}

enum TeamRank {
  bronze,
  silver,
  gold,
  platinum,
  diamond,
}

extension TeamRankExtension on TeamRank {
  String get displayName {
    switch (this) {
      case TeamRank.bronze:
        return 'Bronze';
      case TeamRank.silver:
        return 'Silver';
      case TeamRank.gold:
        return 'Gold';
      case TeamRank.platinum:
        return 'Platinum';
      case TeamRank.diamond:
        return 'Diamond';
    }
  }

  String get emoji {
    switch (this) {
      case TeamRank.bronze:
        return '🥉';
      case TeamRank.silver:
        return '🥈';
      case TeamRank.gold:
        return '🥇';
      case TeamRank.platinum:
        return '💠';
      case TeamRank.diamond:
        return '💎';
    }
  }

  int get color {
    switch (this) {
      case TeamRank.bronze:
        return 0xFFCD7F32;
      case TeamRank.silver:
        return 0xFFC0C0C0;
      case TeamRank.gold:
        return 0xFFFFD700;
      case TeamRank.platinum:
        return 0xFFE5E4E2;
      case TeamRank.diamond:
        return 0xFFB9F2FF;
    }
  }
}

class TeamMember {
  final String userId;
  final String username;
  final String displayName;
  final String? avatarUrl;
  final bool isLeader;
  final int referrals;
  final int conversions;
  final int teamRank;
  final DateTime joinedAt;

  TeamMember({
    required this.userId,
    required this.username,
    required this.displayName,
    this.avatarUrl,
    this.isLeader = false,
    required this.referrals,
    required this.conversions,
    required this.teamRank,
    required this.joinedAt,
  });

  factory TeamMember.fromJson(Map<String, dynamic> json) {
    return TeamMember(
      userId: json['user_id'] ?? '',
      username: json['username'] ?? '',
      displayName: json['display_name'] ?? json['full_name'] ?? '',
      avatarUrl: json['avatar_url'],
      isLeader: json['role'] == 'owner' || json['role'] == 'leader',
      referrals: json['referrals'] ?? 0,
      conversions: json['conversions'] ?? 0,
      teamRank: json['team_rank'] ?? 0,
      joinedAt: DateTime.parse(json['joined_at'] ?? DateTime.now().toIso8601String()),
    );
  }
}

class TeamStats {
  final int totalReferrals;
  final int totalClicks;
  final int totalConversions;
  final int monthlyReferrals;
  final int monthlyConversions;
  final int globalRank;
  final int goalProgress;
  final int goalTarget;

  TeamStats({
    required this.totalReferrals,
    required this.totalClicks,
    required this.totalConversions,
    required this.monthlyReferrals,
    required this.monthlyConversions,
    required this.globalRank,
    required this.goalProgress,
    required this.goalTarget,
  });

  double get goalPercentage => goalTarget > 0 ? (goalProgress / goalTarget * 100).clamp(0, 100) : 0;
  
  double get conversionRate {
    if (totalClicks == 0) return 0;
    return ((totalConversions / totalClicks) * 100);
  }

  factory TeamStats.fromJson(Map<String, dynamic> json) {
    return TeamStats(
      totalReferrals: json['total_referrals'] ?? 0,
      totalClicks: json['total_clicks'] ?? 0,
      totalConversions: json['total_conversions'] ?? 0,
      monthlyReferrals: json['monthly_referrals'] ?? 0,
      monthlyConversions: json['monthly_conversions'] ?? 0,
      globalRank: json['global_rank'] ?? 0,
      goalProgress: json['goal_progress'] ?? 0,
      goalTarget: json['goal_target'] ?? 100,
    );
  }
}

class TeamChallenge {
  final String id;
  final String title;
  final String description;
  final String companyName;
  final String? companyLogo;
  final double reward;
  final int targetReferrals;
  final DateTime startDate;
  final DateTime endDate;
  final int currentProgress;
  final int currentRank;
  final int totalTeams;

  TeamChallenge({
    required this.id,
    required this.title,
    required this.description,
    required this.companyName,
    this.companyLogo,
    required this.reward,
    required this.targetReferrals,
    required this.startDate,
    required this.endDate,
    required this.currentProgress,
    required this.currentRank,
    required this.totalTeams,
  });

  double get progressPercentage =>
      (currentProgress / targetReferrals * 100).clamp(0, 100);

  Duration get timeLeft => endDate.difference(DateTime.now());

  bool get isActive => DateTime.now().isBefore(endDate);

  String get timeLeftText {
    if (!isActive) return 'Ended';
    final days = timeLeft.inDays;
    if (days > 0) return '$days days left';
    final hours = timeLeft.inHours;
    if (hours > 0) return '$hours hours left';
    return '${timeLeft.inMinutes} minutes left';
  }

  factory TeamChallenge.fromJson(Map<String, dynamic> json) {
    return TeamChallenge(
      id: json['id'] ?? '',
      title: json['title'] ?? '',
      description: json['description'] ?? '',
      companyName: json['company_name'] ?? '',
      companyLogo: json['company_logo'],
      reward: (json['reward'] ?? 0).toDouble(),
      targetReferrals: json['target_referrals'] ?? 0,
      startDate: DateTime.parse(json['start_date'] ?? DateTime.now().toIso8601String()),
      endDate: DateTime.parse(json['end_date'] ?? DateTime.now().toIso8601String()),
      currentProgress: json['current_progress'] ?? 0,
      currentRank: json['current_rank'] ?? 0,
      totalTeams: json['total_teams'] ?? 0,
    );
  }
}

// Sample team data
final sampleTeam = Team(
  id: 'team_cryptokings',
  name: 'CryptoKings',
  description: 'Top crypto affiliate team',
  rank: TeamRank.gold,
  members: [
    TeamMember(
      userId: 'user_001',
      username: 'abomohammed',
      displayName: 'Abo Mohammed',
      isLeader: true,
      referrals: 45,
      conversions: 35,
      teamRank: 1,
      joinedAt: DateTime(2024, 6, 15),
    ),
    TeamMember(
      userId: 'user_002',
      username: 'ahmedali',
      displayName: 'Ahmed Ali',
      referrals: 38,
      conversions: 30,
      teamRank: 2,
      joinedAt: DateTime(2024, 6, 20),
    ),
    TeamMember(
      userId: 'user_003',
      username: 'sarakhan',
      displayName: 'Sara Khan',
      referrals: 32,
      conversions: 28,
      teamRank: 3,
      joinedAt: DateTime(2024, 7, 1),
    ),
    TeamMember(
      userId: 'user_004',
      username: 'mohammed',
      displayName: 'Mohammed',
      referrals: 28,
      conversions: 22,
      teamRank: 4,
      joinedAt: DateTime(2024, 7, 5),
    ),
    TeamMember(
      userId: 'user_005',
      username: 'fatima',
      displayName: 'Fatima',
      referrals: 25,
      conversions: 20,
      teamRank: 5,
      joinedAt: DateTime(2024, 7, 10),
    ),
    TeamMember(
      userId: 'user_006',
      username: 'omar',
      displayName: 'Omar',
      referrals: 22,
      conversions: 18,
      teamRank: 6,
      joinedAt: DateTime(2024, 7, 15),
    ),
    TeamMember(
      userId: 'user_007',
      username: 'layla',
      displayName: 'Layla',
      referrals: 18,
      conversions: 15,
      teamRank: 7,
      joinedAt: DateTime(2024, 8, 1),
    ),
    TeamMember(
      userId: 'user_008',
      username: 'khalid',
      displayName: 'Khalid',
      referrals: 15,
      conversions: 12,
      teamRank: 8,
      joinedAt: DateTime(2024, 8, 10),
    ),
  ],
  stats: TeamStats(
    totalReferrals: 450,
    totalClicks: 5200,
    totalConversions: 180,
    monthlyReferrals: 180,
    monthlyConversions: 65,
    globalRank: 3,
    goalProgress: 180,
    goalTarget: 300,
  ),
  createdAt: DateTime(2024, 6, 15),
);

// Sample leaderboard data
final topTeams = [
  Team(
    id: 'team_001',
    name: 'AffiliateKings',
    rank: TeamRank.diamond,
    members: [],
    stats: TeamStats(
      totalReferrals: 850,
      totalClicks: 9500,
      totalConversions: 420,
      monthlyReferrals: 320,
      monthlyConversions: 150,
      globalRank: 1,
      goalProgress: 420,
      goalTarget: 500,
    ),
    createdAt: DateTime(2024, 5, 1),
  ),
  Team(
    id: 'team_002',
    name: 'MarketMasters',
    rank: TeamRank.platinum,
    members: [],
    stats: TeamStats(
      totalReferrals: 720,
      totalClicks: 8200,
      totalConversions: 310,
      monthlyReferrals: 280,
      monthlyConversions: 120,
      globalRank: 2,
      goalProgress: 310,
      goalTarget: 400,
    ),
    createdAt: DateTime(2024, 5, 10),
  ),
  sampleTeam,
];

// Sample active challenge
final activeChallenge = TeamChallenge(
  id: 'challenge_001',
  title: 'Binance Special Campaign',
  description: 'First 3 teams to reach 100 referrals win \$1,000 bonus each!',
  companyName: 'Binance',
  companyLogo: 'https://logo.clearbit.com/binance.com',
  reward: 1000.0,
  targetReferrals: 100,
  startDate: DateTime(2024, 10, 1 ),
  endDate: DateTime(2024, 10, 20),
  currentProgress: 67,
  currentRank: 2,
  totalTeams: 45,
);
