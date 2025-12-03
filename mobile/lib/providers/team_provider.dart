import 'package:flutter/foundation.dart';
import '../models/team.dart';
import '../services/team_service.dart';

class TeamProvider with ChangeNotifier {
  final TeamService _teamService = TeamService();

  List<Team> _teams = [];
  Team? _currentTeam;
  Team? _userTeam;
  bool _isLoading = false;
  String? _error;
  bool _isOwner = false;

  List<Team> get teams => _teams;
  Team? get currentTeam => _currentTeam;
  Team? get userTeam => _userTeam;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isInTeam => _userTeam != null;
  bool get isOwner => _isOwner;

  // Load all teams AND user's team
  Future<void> loadTeams() async {
    _setLoading(true);
    _error = null;

    try {
      // First, load user's team
      await _loadMyTeam();
      
      // Then load all teams
      final result = await _teamService.getAllTeams();

      if (result['success'] == true) {
        _teams = result['teams'] as List<Team>;
        print('[TeamProvider] Loaded ${_teams.length} teams');
      } else {
        _error = result['error'] as String?;
        print('[TeamProvider] Error loading teams: $_error');
      }
    } catch (e) {
      _error = e.toString();
      print('[TeamProvider] Exception loading teams: $e');
    }

    _setLoading(false);
  }

  // Load user's current team
  Future<void> _loadMyTeam() async {
    try {
      final result = await _teamService.getMyTeam();
      
      if (result['success'] == true) {
        _userTeam = result['team'] as Team?;
        _isOwner = result['is_owner'] as bool? ?? false;
        print('[TeamProvider] Loaded my team: ${_userTeam?.name}, isOwner: $_isOwner');
      } else {
        _userTeam = null;
        _isOwner = false;
        print('[TeamProvider] No team found for user');
      }
    } catch (e) {
      print('[TeamProvider] Exception loading my team: $e');
      _userTeam = null;
      _isOwner = false;
    }
  }

  // Refresh user's team only
  Future<void> refreshMyTeam() async {
    await _loadMyTeam();
    notifyListeners();
  }

  // Load specific team
  Future<void> loadTeam(String teamId) async {
    _setLoading(true);
    _error = null;

    try {
      final result = await _teamService.getTeam(teamId);

      if (result['success'] == true) {
        _currentTeam = result['team'] as Team;
        print('[TeamProvider] Loaded team: ${_currentTeam?.name}');
      } else {
        _error = result['error'] as String?;
      }
    } catch (e) {
      _error = e.toString();
    }

    _setLoading(false);
  }

  // Create a new team
  Future<bool> createTeam({
    required String name,
    String? description,
    String? logoUrl,
    int maxMembers = 10,
  }) async {
    _setLoading(true);
    _error = null;

    try {
      final result = await _teamService.createTeam(
        name: name,
        description: description,
        logoUrl: logoUrl,
        maxMembers: maxMembers,
      );

      if (result['success'] == true) {
        _userTeam = result['team'] as Team;
        _isOwner = true;
        print('[TeamProvider] Created team: ${_userTeam?.name}');
        await loadTeams(); // Refresh teams list
        _setLoading(false);
        return true;
      } else {
        _error = result['error'] as String?;
        print('[TeamProvider] Failed to create team: $_error');
      }
    } catch (e) {
      _error = e.toString();
      print('[TeamProvider] Exception creating team: $e');
    }

    _setLoading(false);
    return false;
  }

  // Join a team
  Future<bool> joinTeam(String teamId) async {
    // Check if already in a team
    if (_userTeam != null) {
      _error = 'أنت بالفعل في فريق. غادر فريقك الحالي أولاً';
      notifyListeners();
      return false;
    }
    
    _setLoading(true);
    _error = null;

    try {
      final result = await _teamService.joinTeam(teamId);

      if (result['success'] == true) {
        print('[TeamProvider] Joined team successfully');
        await loadTeams(); // This will also load myTeam
        _setLoading(false);
        return true;
      } else {
        _error = result['error'] as String?;
        print('[TeamProvider] Failed to join team: $_error');
      }
    } catch (e) {
      _error = e.toString();
      print('[TeamProvider] Exception joining team: $e');
    }

    _setLoading(false);
    return false;
  }

  // Leave current team
  Future<bool> leaveTeam() async {
    if (_userTeam == null) return false;

    _setLoading(true);
    _error = null;

    try {
      final result = await _teamService.leaveTeam(_userTeam!.id);

      if (result['success'] == true) {
        print('[TeamProvider] Left team successfully');
        _userTeam = null;
        _currentTeam = null;
        _isOwner = false;
        await loadTeams(); // Refresh teams list
        _setLoading(false);
        return true;
      } else {
        _error = result['error'] as String?;
        print('[TeamProvider] Failed to leave team: $_error');
      }
    } catch (e) {
      _error = e.toString();
      print('[TeamProvider] Exception leaving team: $e');
    }

    _setLoading(false);
    return false;
  }

  // Set user's current team (from user data)
  void setUserTeam(Team? team) {
    _userTeam = team;
    notifyListeners();
  }

  // Clear error
  void clearError() {
    _error = null;
    notifyListeners();
  }

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }
}
