class Team {
  final String id;
  final String name;
  final String? logoUrl;
  final String? description;
  final String ownerId;
  final String? inviteCode;
  final String? inviteUrl;
  final TeamRank rank;
  final List<TeamMember> members;
  final List<TeamMember> pendingMembers;
  final TeamStats stats;
  final int maxMembers;
  final DateTime createdAt;

  Team({
    required this.id,
    required this.name,
    this.logoUrl,
    this.description,
    required this.ownerId,
    this.inviteCode,
    this.inviteUrl,
    required this.rank,
    required this.members,
    this.pendingMembers = const [],
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
    final totalPoints = json['total_points'] ?? 0;
    final totalClicks = json['total_clicks'] ?? 0;
    final totalConversions = json['total_conversions'] ?? 0;
    final memberCount = json['member_count'] ?? 0;
    
    // Parse members - only active members
    final membersList = (json['members'] as List<dynamic>?)
        ?.map((m) => TeamMember.fromJson(m))
        .where((m) => m.status == 'active')
        .toList() ?? [];
    
    // Parse pending members if available
    final pendingList = (json['pending_members'] as List<dynamic>?)
        ?.map((m) => TeamMember.fromJson(m))
        .toList() ?? [];
    
    return Team(
      id: json['id']?.toString() ?? '',
      name: json['name'] ?? '',
      logoUrl: json['logo_url'],
      description: json['description'],
      ownerId: json['owner_id']?.toString() ?? '',
      inviteCode: json['invite_code'],
      inviteUrl: json['invite_url'],
      rank: _parseRank(totalPoints),
      members: membersList,
      pendingMembers: pendingList,
      stats: TeamStats(
        totalPoints: totalPoints,
        totalClicks: totalClicks,
        totalConversions: totalConversions,
        memberCount: memberCount,
        totalReferrals: totalClicks,
      ),
      maxMembers: json['max_members'] ?? 10,
      createdAt: json['created_at'] != null 
          ? DateTime.parse(json['created_at']) 
          : DateTime.now(),
    );
  }
  
  bool isOwner(String userId) => ownerId == userId;

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
  final String id;
  final String teamId;
  final String userId;
  final String username;
  final String displayName;
  final String? avatarUrl;
  final String role;
  final String status; // pending, active, rejected
  final int points;
  final int referrals;
  final int conversions;
  final int teamRank;
  final DateTime joinedAt;
  final DateTime? approvedAt;

  TeamMember({
    required this.id,
    required this.teamId,
    required this.userId,
    required this.username,
    required this.displayName,
    this.avatarUrl,
    required this.role,
    this.status = 'active',
    this.points = 0,
    this.referrals = 0,
    this.conversions = 0,
    this.teamRank = 0,
    required this.joinedAt,
    this.approvedAt,
  });

  bool get isLeader => role == 'owner';
  bool get isAdmin => role == 'admin' || role == 'owner';
  bool get isPending => status == 'pending';
  bool get isActive => status == 'active';

  factory TeamMember.fromJson(Map<String, dynamic> json) {
    // Handle nested user object from backend
    final userJson = json['user'] as Map<String, dynamic>? ?? {};
    
    return TeamMember(
      id: json['id']?.toString() ?? '',
      teamId: json['team_id']?.toString() ?? '',
      userId: json['user_id']?.toString() ?? userJson['id']?.toString() ?? '',
      username: userJson['username'] ?? json['username'] ?? '',
      displayName: userJson['full_name'] ?? userJson['display_name'] ?? json['display_name'] ?? '',
      avatarUrl: userJson['avatar_url'] ?? json['avatar_url'],
      role: json['role'] ?? 'member',
      status: json['status'] ?? 'active',
      points: json['points'] ?? 0,
      referrals: userJson['total_clicks'] ?? json['referrals'] ?? 0,
      conversions: userJson['total_conversions'] ?? json['conversions'] ?? 0,
      teamRank: json['team_rank'] ?? 0,
      joinedAt: json['joined_at'] != null 
          ? DateTime.parse(json['joined_at']) 
          : DateTime.now(),
      approvedAt: json['approved_at'] != null 
          ? DateTime.parse(json['approved_at']) 
          : null,
    );
  }
}

class TeamStats {
  final int totalReferrals;
  final int totalClicks;
  final int totalConversions;
  final int totalPoints;
  final int memberCount;
  final int monthlyReferrals;
  final int monthlyConversions;
  final int globalRank;
  final int goalProgress;
  final int goalTarget;

  TeamStats({
    this.totalReferrals = 0,
    this.totalClicks = 0,
    this.totalConversions = 0,
    this.totalPoints = 0,
    this.memberCount = 0,
    this.monthlyReferrals = 0,
    this.monthlyConversions = 0,
    this.globalRank = 0,
    this.goalProgress = 0,
    this.goalTarget = 100,
  });

  double get goalPercentage => goalTarget > 0 ? (goalProgress / goalTarget * 100).clamp(0, 100) : 0;
  
  double get conversionRate {
    if (totalClicks == 0) return 0;
    return ((totalConversions / totalClicks) * 100);
  }

  factory TeamStats.fromJson(Map<String, dynamic> json) {
    return TeamStats(
      totalReferrals: json['total_referrals'] ?? json['total_clicks'] ?? 0,
      totalClicks: json['total_clicks'] ?? 0,
      totalConversions: json['total_conversions'] ?? 0,
      totalPoints: json['total_points'] ?? 0,
      memberCount: json['member_count'] ?? 0,
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

