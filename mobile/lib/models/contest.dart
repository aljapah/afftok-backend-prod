class Contest {
  final String id;
  final String title;
  final String? titleAr;
  final String? description;
  final String? descriptionAr;
  final String? imageUrl;
  
  // Prize Info
  final String? prizeTitle;
  final String? prizeTitleAr;
  final String? prizeDescription;
  final String? prizeDescriptionAr;
  final double prizeAmount;
  final String prizeCurrency;
  
  // Contest Type & Target
  final String contestType; // team, individual
  final String targetType; // clicks, conversions, referrals, points
  final int targetValue;
  
  // Conditions
  final int minClicks;
  final int minConversions;
  final int minMembers;
  final int maxParticipants;
  
  // Dates
  final DateTime startDate;
  final DateTime endDate;
  
  // Status
  final String status; // draft, active, ended, cancelled
  final int participantsCount;
  
  final DateTime createdAt;
  
  Contest({
    required this.id,
    required this.title,
    this.titleAr,
    this.description,
    this.descriptionAr,
    this.imageUrl,
    this.prizeTitle,
    this.prizeTitleAr,
    this.prizeDescription,
    this.prizeDescriptionAr,
    this.prizeAmount = 0.0,
    this.prizeCurrency = 'USD',
    this.contestType = 'team',
    this.targetType = 'clicks',
    this.targetValue = 100,
    this.minClicks = 0,
    this.minConversions = 0,
    this.minMembers = 1,
    this.maxParticipants = 0,
    required this.startDate,
    required this.endDate,
    this.status = 'draft',
    this.participantsCount = 0,
    required this.createdAt,
  });
  
  bool get isActive {
    final now = DateTime.now();
    return status == 'active' && now.isAfter(startDate) && now.isBefore(endDate);
  }
  
  bool get isEnded => DateTime.now().isAfter(endDate) || status == 'ended';
  
  Duration get timeLeft => endDate.difference(DateTime.now());
  
  String get timeLeftFormatted {
    if (isEnded) return 'انتهت';
    
    final duration = timeLeft;
    if (duration.inDays > 0) {
      return '${duration.inDays} يوم';
    } else if (duration.inHours > 0) {
      return '${duration.inHours} ساعة';
    } else if (duration.inMinutes > 0) {
      return '${duration.inMinutes} دقيقة';
    } else {
      return 'أقل من دقيقة';
    }
  }
  
  String getLocalizedTitle(String languageCode) {
    if (languageCode == 'ar' && titleAr != null && titleAr!.isNotEmpty) {
      return titleAr!;
    }
    return title;
  }
  
  String? getLocalizedDescription(String languageCode) {
    if (languageCode == 'ar' && descriptionAr != null && descriptionAr!.isNotEmpty) {
      return descriptionAr;
    }
    return description;
  }
  
  String? getLocalizedPrizeTitle(String languageCode) {
    if (languageCode == 'ar' && prizeTitleAr != null && prizeTitleAr!.isNotEmpty) {
      return prizeTitleAr;
    }
    return prizeTitle;
  }
  
  factory Contest.fromJson(Map<String, dynamic> json) {
    return Contest(
      id: json['id']?.toString() ?? '',
      title: json['title'] ?? '',
      titleAr: json['title_ar'],
      description: json['description'],
      descriptionAr: json['description_ar'],
      imageUrl: json['image_url'],
      prizeTitle: json['prize_title'],
      prizeTitleAr: json['prize_title_ar'],
      prizeDescription: json['prize_description'],
      prizeDescriptionAr: json['prize_description_ar'],
      prizeAmount: (json['prize_amount'] as num?)?.toDouble() ?? 0.0,
      prizeCurrency: json['prize_currency'] ?? 'USD',
      contestType: json['contest_type'] ?? 'team',
      targetType: json['target_type'] ?? 'clicks',
      targetValue: json['target_value'] ?? 100,
      minClicks: json['min_clicks'] ?? 0,
      minConversions: json['min_conversions'] ?? 0,
      minMembers: json['min_members'] ?? 1,
      maxParticipants: json['max_participants'] ?? 0,
      startDate: json['start_date'] != null 
          ? DateTime.parse(json['start_date']) 
          : DateTime.now(),
      endDate: json['end_date'] != null 
          ? DateTime.parse(json['end_date']) 
          : DateTime.now().add(const Duration(days: 30)),
      status: json['status'] ?? 'draft',
      participantsCount: json['participants_count'] ?? 0,
      createdAt: json['created_at'] != null 
          ? DateTime.parse(json['created_at']) 
          : DateTime.now(),
    );
  }
  
  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'title_ar': titleAr,
      'description': description,
      'description_ar': descriptionAr,
      'image_url': imageUrl,
      'prize_title': prizeTitle,
      'prize_title_ar': prizeTitleAr,
      'prize_description': prizeDescription,
      'prize_description_ar': prizeDescriptionAr,
      'prize_amount': prizeAmount,
      'prize_currency': prizeCurrency,
      'contest_type': contestType,
      'target_type': targetType,
      'target_value': targetValue,
      'min_clicks': minClicks,
      'min_conversions': minConversions,
      'min_members': minMembers,
      'max_participants': maxParticipants,
      'start_date': startDate.toIso8601String(),
      'end_date': endDate.toIso8601String(),
      'status': status,
      'participants_count': participantsCount,
      'created_at': createdAt.toIso8601String(),
    };
  }
}

class ContestParticipant {
  final String id;
  final String contestId;
  final String? teamId;
  final String? userId;
  final int currentClicks;
  final int currentConversions;
  final int currentPoints;
  final int progress;
  final int rank;
  final String status; // active, winner, completed, disqualified
  final DateTime joinedAt;
  
  // Optional nested objects
  final String? teamName;
  final String? userName;
  final String? avatarUrl;
  
  ContestParticipant({
    required this.id,
    required this.contestId,
    this.teamId,
    this.userId,
    this.currentClicks = 0,
    this.currentConversions = 0,
    this.currentPoints = 0,
    this.progress = 0,
    this.rank = 0,
    this.status = 'active',
    required this.joinedAt,
    this.teamName,
    this.userName,
    this.avatarUrl,
  });
  
  bool get isWinner => status == 'winner';
  
  factory ContestParticipant.fromJson(Map<String, dynamic> json) {
    String? teamName;
    String? userName;
    String? avatarUrl;
    
    if (json['team'] != null) {
      teamName = json['team']['name'];
      avatarUrl = json['team']['logo_url'];
    }
    if (json['user'] != null) {
      userName = json['user']['full_name'] ?? json['user']['username'];
      avatarUrl ??= json['user']['avatar_url'];
    }
    
    return ContestParticipant(
      id: json['id']?.toString() ?? '',
      contestId: json['contest_id']?.toString() ?? '',
      teamId: json['team_id']?.toString(),
      userId: json['user_id']?.toString(),
      currentClicks: json['current_clicks'] ?? 0,
      currentConversions: json['current_conversions'] ?? 0,
      currentPoints: json['current_points'] ?? 0,
      progress: json['progress'] ?? 0,
      rank: json['rank'] ?? 0,
      status: json['status'] ?? 'active',
      joinedAt: json['joined_at'] != null 
          ? DateTime.parse(json['joined_at']) 
          : DateTime.now(),
      teamName: teamName,
      userName: userName,
      avatarUrl: avatarUrl,
    );
  }
}

