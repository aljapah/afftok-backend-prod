/// Stats models matching the updated backend analytics service
/// Backend endpoints: /api/stats/me, /api/stats/daily

// User stats response from /api/stats/me
class UserAnalytics {
  final int totalClicks;
  final int totalConversions;
  final int totalEarnings;
  final double conversionRate;
  final int activeOffers;
  
  // Time-based stats
  final int clicksToday;
  final int clicksThisWeek;
  final int clicksThisMonth;
  
  final int conversionsToday;
  final int conversionsThisWeek;
  final int conversionsThisMonth;

  UserAnalytics({
    required this.totalClicks,
    required this.totalConversions,
    required this.totalEarnings,
    required this.conversionRate,
    required this.activeOffers,
    required this.clicksToday,
    required this.clicksThisWeek,
    required this.clicksThisMonth,
    required this.conversionsToday,
    required this.conversionsThisWeek,
    required this.conversionsThisMonth,
  });

  factory UserAnalytics.fromJson(Map<String, dynamic> json) {
    return UserAnalytics(
      totalClicks: _parseInt(json['total_clicks']),
      totalConversions: _parseInt(json['total_conversions']),
      totalEarnings: _parseInt(json['total_earnings']),
      conversionRate: _parseDouble(json['conversion_rate']),
      activeOffers: _parseInt(json['active_offers']),
      clicksToday: _parseInt(json['clicks_today']),
      clicksThisWeek: _parseInt(json['clicks_this_week']),
      clicksThisMonth: _parseInt(json['clicks_this_month']),
      conversionsToday: _parseInt(json['conversions_today']),
      conversionsThisWeek: _parseInt(json['conversions_this_week']),
      conversionsThisMonth: _parseInt(json['conversions_this_month']),
    );
  }

  // Empty/default stats
  factory UserAnalytics.empty() {
    return UserAnalytics(
      totalClicks: 0,
      totalConversions: 0,
      totalEarnings: 0,
      conversionRate: 0.0,
      activeOffers: 0,
      clicksToday: 0,
      clicksThisWeek: 0,
      clicksThisMonth: 0,
      conversionsToday: 0,
      conversionsThisWeek: 0,
      conversionsThisMonth: 0,
    );
  }

  static int _parseInt(dynamic value) {
    if (value == null) return 0;
    if (value is int) return value;
    if (value is double) return value.toInt();
    if (value is String) return int.tryParse(value) ?? 0;
    return 0;
  }

  static double _parseDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is double) return value;
    if (value is int) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }
}

// Daily stats for charts from /api/stats/daily
class DailyStats {
  final String date;
  final int clicks;
  final int conversions;

  DailyStats({
    required this.date,
    required this.clicks,
    required this.conversions,
  });

  factory DailyStats.fromJson(Map<String, dynamic> json) {
    return DailyStats(
      date: json['date']?.toString() ?? '',
      clicks: UserAnalytics._parseInt(json['clicks']),
      conversions: UserAnalytics._parseInt(json['conversions']),
    );
  }

  double get conversionRate {
    if (clicks == 0) return 0.0;
    return (conversions / clicks) * 100;
  }
}

// Click stats by offer from /api/clicks/by-offer
class OfferClickStats {
  final String offerId;
  final String offerTitle;
  final int totalClicks;
  final int uniqueClicks;
  final int conversions;

  OfferClickStats({
    required this.offerId,
    required this.offerTitle,
    required this.totalClicks,
    required this.uniqueClicks,
    required this.conversions,
  });

  factory OfferClickStats.fromJson(Map<String, dynamic> json) {
    return OfferClickStats(
      offerId: json['offer_id']?.toString() ?? '',
      offerTitle: json['offer_title']?.toString() ?? '',
      totalClicks: UserAnalytics._parseInt(json['total_clicks']),
      uniqueClicks: UserAnalytics._parseInt(json['unique_clicks']),
      conversions: UserAnalytics._parseInt(json['conversions']),
    );
  }

  double get conversionRate {
    if (totalClicks == 0) return 0.0;
    return (conversions / totalClicks) * 100;
  }
}

// Click event from /api/clicks/my
class ClickEvent {
  final String id;
  final String userOfferId;
  final String ipAddress;
  final String device;
  final String browser;
  final String os;
  final String? country;
  final String? city;
  final String? referrer;
  final DateTime clickedAt;
  final bool isUnique;

  ClickEvent({
    required this.id,
    required this.userOfferId,
    required this.ipAddress,
    required this.device,
    required this.browser,
    required this.os,
    this.country,
    this.city,
    this.referrer,
    required this.clickedAt,
    this.isUnique = true,
  });

  factory ClickEvent.fromJson(Map<String, dynamic> json) {
    return ClickEvent(
      id: json['id']?.toString() ?? '',
      userOfferId: json['user_offer_id']?.toString() ?? '',
      ipAddress: json['ip_address']?.toString() ?? '',
      device: json['device']?.toString() ?? 'unknown',
      browser: json['browser']?.toString() ?? 'unknown',
      os: json['os']?.toString() ?? 'unknown',
      country: json['country']?.toString(),
      city: json['city']?.toString(),
      referrer: json['referrer']?.toString(),
      clickedAt: DateTime.tryParse(json['clicked_at']?.toString() ?? '') ?? DateTime.now(),
      isUnique: json['is_unique'] ?? true,
    );
  }
}

// Conversion event
class ConversionEvent {
  final String id;
  final String userOfferId;
  final String? clickId;
  final String? externalConversionId;
  final int amount;
  final int commission;
  final String currency;
  final String status;
  final DateTime convertedAt;
  final DateTime? approvedAt;

  ConversionEvent({
    required this.id,
    required this.userOfferId,
    this.clickId,
    this.externalConversionId,
    required this.amount,
    required this.commission,
    required this.currency,
    required this.status,
    required this.convertedAt,
    this.approvedAt,
  });

  factory ConversionEvent.fromJson(Map<String, dynamic> json) {
    return ConversionEvent(
      id: json['id']?.toString() ?? '',
      userOfferId: json['user_offer_id']?.toString() ?? '',
      clickId: json['click_id']?.toString(),
      externalConversionId: json['external_conversion_id']?.toString(),
      amount: UserAnalytics._parseInt(json['amount']),
      commission: UserAnalytics._parseInt(json['commission']),
      currency: json['currency']?.toString() ?? 'USD',
      status: json['status']?.toString() ?? 'pending',
      convertedAt: DateTime.tryParse(json['converted_at']?.toString() ?? '') ?? DateTime.now(),
      approvedAt: json['approved_at'] != null 
          ? DateTime.tryParse(json['approved_at'].toString()) 
          : null,
    );
  }

  bool get isPending => status == 'pending';
  bool get isApproved => status == 'approved';
  bool get isRejected => status == 'rejected';
}

// JoinOffer response with new tracking link
class JoinOfferResponse {
  final bool success;
  final String? error;
  final String? userOfferId;
  final String? affiliateLink;
  final String? shortLink;
  final String? trackingUrl;

  JoinOfferResponse({
    required this.success,
    this.error,
    this.userOfferId,
    this.affiliateLink,
    this.shortLink,
    this.trackingUrl,
  });

  factory JoinOfferResponse.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>? ?? json;
    final userOffer = data['user_offer'] as Map<String, dynamic>?;
    
    return JoinOfferResponse(
      success: json['success'] == true || data['success'] == true,
      error: json['error']?.toString() ?? data['error']?.toString(),
      userOfferId: userOffer?['id']?.toString(),
      affiliateLink: data['affiliate_link']?.toString() ?? userOffer?['affiliate_link']?.toString(),
      shortLink: data['short_link']?.toString() ?? userOffer?['short_link']?.toString(),
      trackingUrl: data['tracking_url']?.toString(),
    );
  }
}

