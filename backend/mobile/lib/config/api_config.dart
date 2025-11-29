class ApiConfig {
  static const String baseUrl = 'https://afftok-backend-prod-production.up.railway.app';
  
  static const String apiPrefix = '/api';
  
  static const String register = '$apiPrefix/auth/register';
  static const String login = '$apiPrefix/auth/login';
  static const String logout = '$apiPrefix/auth/logout';
  static const String refreshToken = '$apiPrefix/auth/refresh';
  static const String getMe = '$apiPrefix/auth/me';
  
  static const String users = '$apiPrefix/users';
  static const String profile = '$apiPrefix/profile';
  
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);
}
