import '../models/team.dart';
import 'api_service.dart';

class TeamService {
  final ApiService _apiService = ApiService();

  // Get all teams
  Future<Map<String, dynamic>> getAllTeams() async {
    try {
      print('[TeamService] Fetching all teams...');
      final result = await _apiService.get('/api/teams');

      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>?;
        final teamsJson = data?['teams'] as List? ?? [];
        print('[TeamService] Found ${teamsJson.length} teams');

        final teams = teamsJson.map((json) => Team.fromJson(json)).toList();

        return {
          'success': true,
          'teams': teams,
        };
      } else {
        return {
          'success': false,
          'error': result['error'] ?? 'Failed to load teams',
        };
      }
    } catch (e) {
      print('[TeamService] Error fetching teams: $e');
      return {
        'success': false,
        'error': 'Error fetching teams: $e',
      };
    }
  }

  // Get single team by ID
  Future<Map<String, dynamic>> getTeam(String teamId) async {
    try {
      print('[TeamService] Fetching team $teamId...');
      final result = await _apiService.get('/api/teams/$teamId');

      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>?;
        final teamJson = data?['team'] as Map<String, dynamic>?;
        
        if (teamJson != null) {
          final team = Team.fromJson(teamJson);
          return {
            'success': true,
            'team': team,
          };
        }
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Team not found',
      };
    } catch (e) {
      print('[TeamService] Error fetching team: $e');
      return {
        'success': false,
        'error': 'Error fetching team: $e',
      };
    }
  }

  // Create a new team
  Future<Map<String, dynamic>> createTeam({
    required String name,
    String? description,
    String? logoUrl,
    int maxMembers = 10,
  }) async {
    try {
      print('[TeamService] Creating team: $name');
      final result = await _apiService.post('/api/teams', {
        'name': name,
        if (description != null) 'description': description,
        if (logoUrl != null) 'logo_url': logoUrl,
        'max_members': maxMembers,
      });

      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>?;
        final teamJson = data?['team'] as Map<String, dynamic>?;
        
        if (teamJson != null) {
          final team = Team.fromJson(teamJson);
          return {
            'success': true,
            'team': team,
            'message': data?['message'] ?? 'Team created successfully',
          };
        }
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Failed to create team',
      };
    } catch (e) {
      print('[TeamService] Error creating team: $e');
      return {
        'success': false,
        'error': 'Error creating team: $e',
      };
    }
  }

  // Join a team
  Future<Map<String, dynamic>> joinTeam(String teamId) async {
    try {
      print('[TeamService] Joining team $teamId...');
      final result = await _apiService.post('/api/teams/$teamId/join', {});

      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>?;
        return {
          'success': true,
          'message': data?['message'] ?? 'Joined team successfully',
        };
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Failed to join team',
      };
    } catch (e) {
      print('[TeamService] Error joining team: $e');
      return {
        'success': false,
        'error': 'Error joining team: $e',
      };
    }
  }

  // Leave a team
  Future<Map<String, dynamic>> leaveTeam(String teamId) async {
    try {
      print('[TeamService] Leaving team $teamId...');
      final result = await _apiService.post('/api/teams/$teamId/leave', {});

      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>?;
        return {
          'success': true,
          'message': data?['message'] ?? 'Left team successfully',
        };
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Failed to leave team',
      };
    } catch (e) {
      print('[TeamService] Error leaving team: $e');
      return {
        'success': false,
        'error': 'Error leaving team: $e',
      };
    }
  }
}

