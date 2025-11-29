import 'api_service.dart';
import '../models/user.dart';

class UserService {
  final ApiService _apiService = ApiService();
  
  // Get user by ID
  Future<User?> getUserById(String userId) async {
    try {
      if (userId.isEmpty) {
        print('Error: userId is empty');
        return null;
      }
      
      final response = await _apiService.get('/users/$userId');
      
      if (response['success'] == true && response['data'] != null) {
        try {
          return User.fromJson(response['data']);
        } catch (e) {
          print('Error parsing user data: $e');
          return null;
        }
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
      if (userId.isEmpty) {
        print('Error: userId is empty');
        return false;
      }
      
      if (data.isEmpty) {
        print('Error: data is empty');
        return false;
      }
      
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
      if (userId.isEmpty) {
        print('Error: userId is empty');
        return null;
      }
      
      final user = await getUserById(userId);
      return user?.stats;
    } catch (e) {
      print('Error fetching user stats: $e');
      return null;
    }
  }
}
