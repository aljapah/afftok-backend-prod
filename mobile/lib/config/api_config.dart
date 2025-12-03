/// API configuration for AffTok mobile app
/// Updated to match the new backend tracking engine
class ApiConfig {
  static const String baseUrl = 'https://afftok-backend-prod-production.up.railway.app';
  
  static const String apiPrefix = '/api';
  
  // Auth endpoints
  static const String register = '$apiPrefix/auth/register';
  static const String login = '$apiPrefix/auth/login';
  static const String logout = '$apiPrefix/auth/logout';
  static const String refreshToken = '$apiPrefix/auth/refresh';
  static const String getMe = '$apiPrefix/auth/me';
  
  // User endpoints
  static const String users = '$apiPrefix/users';
  static const String profile = '$apiPrefix/profile';
  
  // Offers endpoints
  static const String offers = '$apiPrefix/offers';
  static const String myOffers = '$apiPrefix/offers/my';
  static String joinOffer(String offerId) => '$apiPrefix/offers/$offerId/join';
  
  // Stats/Analytics endpoints (new)
  static const String statsMe = '$apiPrefix/stats/me';
  static const String statsDaily = '$apiPrefix/stats/daily';
  
  // Clicks endpoints
  static const String myClicks = '$apiPrefix/clicks/my';
  static const String clicksByOffer = '$apiPrefix/clicks/by-offer';
  static String clickStats(String userOfferId) => '$apiPrefix/clicks/$userOfferId/stats';
  
  // Click tracking endpoint (for referral links)
  static String trackClick(String trackingCode) => '$apiPrefix/c/$trackingCode';
  
  // Teams endpoints
  static const String teams = '$apiPrefix/teams';
  static String joinTeam(String teamId) => '$apiPrefix/teams/$teamId/join';
  static String leaveTeam(String teamId) => '$apiPrefix/teams/$teamId/leave';
  
  // Badges endpoints
  static const String badges = '$apiPrefix/badges';
  static const String myBadges = '$apiPrefix/badges/my';
  
  // Networks endpoints
  static const String networks = '$apiPrefix/networks';
  
  // Promoter public page
  static String promoterPage(String promoterId) => '$apiPrefix/promoter/$promoterId';
  static String promoterPageByUsername(String username) => '$apiPrefix/promoter/user/$username';
  
  // Rating endpoint
  static const String ratePromoter = '$apiPrefix/rate-promoter';
  
  // Postback (for external integrations)
  static const String postback = '$apiPrefix/postback';
  
  // Timeouts
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);
  
  // Build full URL
  static String fullUrl(String endpoint) => '$baseUrl$endpoint';
}
