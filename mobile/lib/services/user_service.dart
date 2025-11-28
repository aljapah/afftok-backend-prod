import 'api_service.dart';
import '../models/user.dart';

class UserService {
  final ApiService _apiService = ApiService();
  
  // Get user by ID
  Future<User?> getUserById(String userId) async {
    try {
      final response = await _apiService.get('/users/$userId');
      if (response['success'] == true && response['data'] != null) {
        return User.fromJson(response['data']);
      }
      return null;
    } catch (e) {
      print('Error fetching user: $e');
      return null;
    }
  }
  
  // Update user
  Future<bool> updateUser(String userId, Map<String, dynamic> data) async {
    try {
      final response = await _apiService.put('/users/$userId', data);
      return response['success'] == true;
    } catch (e) {
      print('Error updating user: $e');
      return false;
    }
  }
  
  // Get user stats
  Future<UserStats?> getUserStats(String userId) async {
    try {
      final user = await getUserById(userId);
      return user?.stats;
    } catch (e) {
      print('Error fetching user stats: $e');
      return null;
    }
  }
}
