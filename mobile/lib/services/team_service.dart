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

  // Get my team (with pending members if owner)
  Future<Map<String, dynamic>> getMyTeam() async {
    try {
      print('[TeamService] Fetching my team...');
      final result = await _apiService.get('/api/teams/my');

      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>?;
        final teamJson = data?['team'] as Map<String, dynamic>?;
        
        if (teamJson != null) {
          // Add pending_members to team json if available
          if (data?['pending_members'] != null) {
            teamJson['pending_members'] = data!['pending_members'];
          }
          if (data?['invite_code'] != null) {
            teamJson['invite_code'] = data!['invite_code'];
          }
          if (data?['invite_url'] != null) {
            teamJson['invite_url'] = data!['invite_url'];
          }
          
          final team = Team.fromJson(teamJson);
          return {
            'success': true,
            'team': team,
            'is_owner': data?['is_owner'] ?? false,
          };
        }
      }
      return {
        'success': false,
        'error': result['error'] ?? 'You are not in any team',
      };
    } catch (e) {
      print('[TeamService] Error fetching my team: $e');
      return {
        'success': false,
        'error': 'Error fetching my team: $e',
      };
    }
  }

  // Join team by invite code
  Future<Map<String, dynamic>> joinTeamByInviteCode(String inviteCode) async {
    try {
      print('[TeamService] Joining team by invite code: $inviteCode');
      final result = await _apiService.post('/api/teams/join/$inviteCode', {});

      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>?;
        return {
          'success': true,
          'message': data?['message'] ?? 'Join request sent successfully',
          'status': data?['status'] ?? 'pending',
          'team_name': data?['team_name'],
        };
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Failed to join team',
      };
    } catch (e) {
      print('[TeamService] Error joining team by code: $e');
      return {
        'success': false,
        'error': 'Error joining team: $e',
      };
    }
  }

  // Approve pending member (owner only)
  Future<Map<String, dynamic>> approveMember(String teamId, String memberId) async {
    try {
      print('[TeamService] Approving member $memberId in team $teamId');
      final result = await _apiService.post('/api/teams/$teamId/approve/$memberId', {});

      if (result['success'] == true) {
        return {
          'success': true,
          'message': 'Member approved successfully',
        };
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Failed to approve member',
      };
    } catch (e) {
      print('[TeamService] Error approving member: $e');
      return {
        'success': false,
        'error': 'Error approving member: $e',
      };
    }
  }

  // Reject pending member (owner only)
  Future<Map<String, dynamic>> rejectMember(String teamId, String memberId) async {
    try {
      print('[TeamService] Rejecting member $memberId in team $teamId');
      final result = await _apiService.post('/api/teams/$teamId/reject/$memberId', {});

      if (result['success'] == true) {
        return {
          'success': true,
          'message': 'Member rejected',
        };
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Failed to reject member',
      };
    } catch (e) {
      print('[TeamService] Error rejecting member: $e');
      return {
        'success': false,
        'error': 'Error rejecting member: $e',
      };
    }
  }

  // Remove active member (owner only)
  Future<Map<String, dynamic>> removeMember(String teamId, String memberId) async {
    try {
      print('[TeamService] Removing member $memberId from team $teamId');
      final result = await _apiService.delete('/api/teams/$teamId/members/$memberId');

      if (result['success'] == true) {
        return {
          'success': true,
          'message': 'Member removed successfully',
        };
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Failed to remove member',
      };
    } catch (e) {
      print('[TeamService] Error removing member: $e');
      return {
        'success': false,
        'error': 'Error removing member: $e',
      };
    }
  }

  // Regenerate invite code (owner only)
  Future<Map<String, dynamic>> regenerateInviteCode(String teamId) async {
    try {
      print('[TeamService] Regenerating invite code for team $teamId');
      final result = await _apiService.post('/api/teams/$teamId/regenerate-invite', {});

      if (result['success'] == true) {
        final data = result['data'] as Map<String, dynamic>?;
        return {
          'success': true,
          'invite_code': data?['invite_code'],
          'invite_url': data?['invite_url'],
        };
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Failed to regenerate invite code',
      };
    } catch (e) {
      print('[TeamService] Error regenerating invite code: $e');
      return {
        'success': false,
        'error': 'Error regenerating invite code: $e',
      };
    }
  }

  // Delete team (owner only)
  Future<Map<String, dynamic>> deleteTeam(String teamId) async {
    try {
      print('[TeamService] Deleting team $teamId');
      final result = await _apiService.delete('/api/teams/$teamId');

      if (result['success'] == true) {
        return {
          'success': true,
          'message': 'Team deleted successfully',
        };
      }
      return {
        'success': false,
        'error': result['error'] ?? 'Failed to delete team',
      };
    } catch (e) {
      print('[TeamService] Error deleting team: $e');
      return {
        'success': false,
        'error': 'Error deleting team: $e',
      };
    }
  }
}

