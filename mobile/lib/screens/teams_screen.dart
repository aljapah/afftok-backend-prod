import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';
import 'package:provider/provider.dart';
import 'package:qr_flutter/qr_flutter.dart';
import '../utils/app_localizations.dart';
import '../models/team.dart';
import '../models/contest.dart';
import '../providers/auth_provider.dart';
import '../providers/team_provider.dart';
import '../services/team_service.dart';
import '../services/contest_service.dart';
import '../services/api_service.dart';

class TeamsScreen extends StatefulWidget {
  const TeamsScreen({Key? key}) : super(key: key);

  @override
  State<TeamsScreen> createState() => _TeamsScreenState();
}

class _TeamsScreenState extends State<TeamsScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  late ContestService _contestService;
  List<Contest> _activeContests = [];
  bool _loadingContests = false;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    
    // Load teams when screen opens
    WidgetsBinding.instance.addPostFrameCallback((_) {
      Provider.of<TeamProvider>(context, listen: false).loadTeams();
      _loadContests();
    });
  }

  Future<void> _loadContests() async {
    final apiService = Provider.of<ApiService>(context, listen: false);
    _contestService = ContestService(apiService);
    
    setState(() => _loadingContests = true);
    
    try {
      final contests = await _contestService.getActiveContests();
      if (mounted) {
        setState(() {
          _activeContests = contests;
          _loadingContests = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loadingContests = false);
      }
    }
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    
    return Consumer2<AuthProvider, TeamProvider>(
      builder: (context, authProvider, teamProvider, child) {
        return Scaffold(
          backgroundColor: Colors.black,
          appBar: AppBar(
            backgroundColor: Colors.black,
            title: Text(
              lang.teams,
              style: const TextStyle(color: Colors.white),
            ),
            leading: IconButton(
              icon: const Icon(Icons.arrow_back, color: Colors.white),
              onPressed: () => Navigator.pop(context),
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.refresh, color: Colors.white),
                onPressed: () => teamProvider.loadTeams(),
              ),
            ],
            bottom: TabBar(
              controller: _tabController,
              indicatorColor: const Color(0xFFFF006E),
              labelColor: Colors.white,
              unselectedLabelColor: Colors.white.withOpacity(0.5),
              tabs: [
                Tab(text: lang.myTeam),
                Tab(text: lang.leaderboard),
                Tab(text: lang.challenges),
              ],
            ),
          ),
          body: TabBarView(
            controller: _tabController,
            children: [
              _buildMyTeamTab(context, lang, teamProvider),
              _buildLeaderboardTab(context, lang, teamProvider),
              _buildContestsTab(context, lang),
            ],
          ),
        );
      },
    );
  }

  Widget _buildMyTeamTab(BuildContext context, AppLocalizations lang, TeamProvider teamProvider) {
    final userTeam = teamProvider.userTeam;
    
    if (userTeam == null) {
      return _buildNoTeamView(context, lang, teamProvider);
    }
    
    return RefreshIndicator(
      onRefresh: () => teamProvider.loadTeams(),
      color: const Color(0xFFFF006E),
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildTeamHeader(context, userTeam, lang),
            const SizedBox(height: 24),
            _buildTeamStats(userTeam, lang),
            const SizedBox(height: 24),
            _buildTeamMembers(userTeam, lang),
            const SizedBox(height: 24),
            _buildTeamActions(context, lang, teamProvider, userTeam),
          ],
        ),
      ),
    );
  }

  Widget _buildNoTeamView(BuildContext context, AppLocalizations lang, TeamProvider teamProvider) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          const SizedBox(height: 60),
          Container(
            width: 120,
            height: 120,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              gradient: LinearGradient(
                colors: [
                  const Color(0xFFFF006E).withOpacity(0.2),
                  const Color(0xFFFF4D00).withOpacity(0.1),
                ],
              ),
            ),
            child: const Icon(
              Icons.group_outlined,
              size: 60,
              color: Color(0xFFFF006E),
            ),
          ),
          const SizedBox(height: 24),
          Text(
            lang.noTeamYet,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 24,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            lang.joinOrCreateTeam,
            textAlign: TextAlign.center,
            style: TextStyle(
              color: Colors.white.withOpacity(0.6),
              fontSize: 14,
            ),
          ),
          const SizedBox(height: 32),
          
          // Create Team Button
          Container(
            width: double.infinity,
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
              ),
              borderRadius: BorderRadius.circular(12),
            ),
            child: ElevatedButton.icon(
              onPressed: () => _showCreateTeamDialog(context, lang, teamProvider),
              icon: const Icon(Icons.add, color: Colors.white),
              label: Text(lang.createTeam),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.transparent,
                shadowColor: Colors.transparent,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
            ),
          ),
          
          const SizedBox(height: 16),
          
          // Browse Teams Button
          OutlinedButton.icon(
            onPressed: () {
              _tabController.animateTo(1);
            },
            icon: const Icon(Icons.search, color: Color(0xFFFF006E)),
            label: Text(
              lang.browseTeams,
              style: const TextStyle(color: Color(0xFFFF006E)),
            ),
            style: OutlinedButton.styleFrom(
              side: const BorderSide(color: Color(0xFFFF006E)),
              padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 24),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTeamHeader(BuildContext context, Team team, AppLocalizations lang) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color(team.rank.color).withOpacity(0.3),
            Color(team.rank.color).withOpacity(0.1),
          ],
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: Color(team.rank.color).withOpacity(0.5),
        ),
      ),
      child: Column(
        children: [
          // Team Logo
          Container(
            width: 80,
            height: 80,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              gradient: LinearGradient(
                colors: [
                  Color(team.rank.color),
                  Color(team.rank.color).withOpacity(0.7),
                ],
              ),
            ),
            child: team.logoUrl != null
                ? ClipOval(
                    child: Image.network(
                      team.logoUrl!,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Center(
                        child: Text(
                          team.name.substring(0, 1).toUpperCase(),
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 32,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ),
                  )
                : Center(
                    child: Text(
                      team.name.substring(0, 1).toUpperCase(),
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 32,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
          ),
          const SizedBox(height: 16),
          
          // Team Name
          Text(
            team.name,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 24,
              fontWeight: FontWeight.bold,
            ),
          ),
          
          // Description
          if (team.description != null && team.description!.isNotEmpty) ...[
            const SizedBox(height: 8),
            Text(
              team.description!,
              textAlign: TextAlign.center,
              style: TextStyle(
                color: Colors.white.withOpacity(0.7),
                fontSize: 14,
              ),
            ),
          ],
          
          const SizedBox(height: 12),
          
          // Rank Badge
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            decoration: BoxDecoration(
              color: Color(team.rank.color),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(team.rank.emoji, style: const TextStyle(fontSize: 16)),
                const SizedBox(width: 8),
                Text(
                  team.rank.displayName,
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTeamStats(Team team, AppLocalizations lang) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white.withOpacity(0.1)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            lang.teamStats,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 18,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: _buildStatItem(
                  icon: Icons.people,
                  label: lang.members,
                  value: '${team.memberCount}/${team.maxMembers}',
                ),
              ),
              Expanded(
                child: _buildStatItem(
                  icon: Icons.touch_app,
                  label: lang.clicks,
                  value: team.stats.totalClicks.toString(),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: _buildStatItem(
                  icon: Icons.trending_up,
                  label: lang.conversions,
                  value: team.stats.totalConversions.toString(),
                ),
              ),
              Expanded(
                child: _buildStatItem(
                  icon: Icons.stars,
                  label: lang.points,
                  value: team.stats.totalPoints.toString(),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildStatItem({
    required IconData icon,
    required String label,
    required String value,
  }) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: const Color(0xFFFF006E).withOpacity(0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: const Color(0xFFFF006E), size: 20),
        ),
        const SizedBox(width: 12),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              value,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(
              label,
              style: TextStyle(
                color: Colors.white.withOpacity(0.5),
                fontSize: 12,
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildTeamMembers(Team team, AppLocalizations lang) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white.withOpacity(0.1)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${lang.members} (${team.members.length})',
            style: const TextStyle(
              color: Colors.white,
              fontSize: 18,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 16),
          if (team.members.isEmpty)
            Center(
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: Text(
                  lang.noMembersYet,
                  style: TextStyle(color: Colors.white.withOpacity(0.5)),
                ),
              ),
            )
          else
            ...team.members.map((member) => _buildMemberItem(member)),
        ],
      ),
    );
  }

  Widget _buildMemberItem(TeamMember member) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          // Avatar
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              gradient: LinearGradient(
                colors: member.isLeader
                    ? [const Color(0xFFFFD700), const Color(0xFFFFA500)]
                    : [const Color(0xFFFF006E), const Color(0xFFFF4D00)],
              ),
            ),
            child: member.avatarUrl != null
                ? ClipOval(
                    child: Image.network(
                      member.avatarUrl!,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Center(
                        child: Text(
                          member.displayName.isNotEmpty
                              ? member.displayName[0].toUpperCase()
                              : '?',
                          style: const TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ),
                  )
                : Center(
                    child: Text(
                      member.displayName.isNotEmpty
                          ? member.displayName[0].toUpperCase()
                          : '?',
                      style: const TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
          ),
          const SizedBox(width: 12),
          
          // Info
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      member.displayName,
                      style: const TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    if (member.isLeader) ...[
                      const SizedBox(width: 6),
                      const Text('👑', style: TextStyle(fontSize: 12)),
                    ],
                  ],
                ),
                Text(
                  '@${member.username}',
                  style: TextStyle(
                    color: Colors.white.withOpacity(0.5),
                    fontSize: 12,
                  ),
                ),
              ],
            ),
          ),
          
          // Stats
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                '${member.conversions}',
                style: const TextStyle(
                  color: Color(0xFFFF006E),
                  fontWeight: FontWeight.bold,
                ),
              ),
              Text(
                'conversions',
                style: TextStyle(
                  color: Colors.white.withOpacity(0.5),
                  fontSize: 10,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildTeamActions(BuildContext context, AppLocalizations lang, TeamProvider teamProvider, Team team) {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final isOwner = authProvider.currentUser != null && team.isOwner(authProvider.currentUser!.id);
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Invite Section with QR Code
        _buildInviteSection(context, team, isArabic),
        
        const SizedBox(height: 24),
        
        // Pending Members (Owner Only)
        if (isOwner && team.pendingMembers.isNotEmpty) ...[
          _buildPendingMembersSection(context, team, teamProvider, isArabic),
          const SizedBox(height: 24),
        ],
        
        // Actions
        Row(
          children: [
            // Share Button
            Expanded(
              child: Container(
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                  ),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: ElevatedButton.icon(
                  onPressed: () => _shareTeamInvite(team),
                  icon: const Icon(Icons.share, color: Colors.white, size: 20),
                  label: Text(isArabic ? 'مشاركة' : 'Share'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.transparent,
                    shadowColor: Colors.transparent,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            // QR Button
            Container(
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.white.withOpacity(0.2)),
              ),
              child: IconButton(
                onPressed: () => _showQRCodeDialog(context, team, isArabic),
                icon: const Icon(Icons.qr_code, color: Colors.white),
                tooltip: isArabic ? 'رمز QR' : 'QR Code',
              ),
            ),
          ],
        ),
        
        const SizedBox(height: 16),
        
        // Leave Team Button (or Delete for owner)
        if (isOwner)
          TextButton.icon(
            onPressed: () => _confirmDeleteTeam(context, lang, teamProvider, team),
            icon: const Icon(Icons.delete_outline, color: Colors.red, size: 20),
            label: Text(
              isArabic ? 'حذف الفريق' : 'Delete Team',
              style: const TextStyle(color: Colors.red),
            ),
          )
        else
          TextButton.icon(
            onPressed: () => _confirmLeaveTeam(context, lang, teamProvider),
            icon: const Icon(Icons.exit_to_app, color: Colors.red, size: 20),
            label: Text(
              lang.leaveTeam,
              style: const TextStyle(color: Colors.red),
            ),
          ),
      ],
    );
  }
  
  Widget _buildInviteSection(BuildContext context, Team team, bool isArabic) {
    final inviteUrl = team.inviteUrl ?? 'https://afftok.com/join/${team.inviteCode ?? team.id}';
    
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white.withOpacity(0.1)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.link, color: Color(0xFFFF006E), size: 20),
              const SizedBox(width: 8),
              Text(
                isArabic ? 'رابط الدعوة' : 'Invite Link',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            decoration: BoxDecoration(
              color: Colors.black.withOpacity(0.3),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    inviteUrl,
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.7),
                      fontSize: 13,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                const SizedBox(width: 8),
                GestureDetector(
                  onTap: () {
                    Clipboard.setData(ClipboardData(text: inviteUrl));
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: Text(isArabic ? 'تم نسخ الرابط ✓' : 'Link copied ✓'),
                        backgroundColor: const Color(0xFFFF006E),
                        duration: const Duration(seconds: 2),
                      ),
                    );
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFF006E),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      isArabic ? 'نسخ' : 'Copy',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
          if (team.inviteCode != null) ...[
            const SizedBox(height: 8),
            Text(
              '${isArabic ? "كود الدعوة:" : "Invite Code:"} ${team.inviteCode}',
              style: TextStyle(
                color: Colors.white.withOpacity(0.5),
                fontSize: 12,
              ),
            ),
          ],
        ],
      ),
    );
  }
  
  Widget _buildPendingMembersSection(BuildContext context, Team team, TeamProvider teamProvider, bool isArabic) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.orange.withOpacity(0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.orange.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.pending_actions, color: Colors.orange, size: 20),
              const SizedBox(width: 8),
              Text(
                isArabic ? 'طلبات الانضمام (${team.pendingMembers.length})' : 'Join Requests (${team.pendingMembers.length})',
                style: const TextStyle(
                  color: Colors.orange,
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          ...team.pendingMembers.map((member) => _buildPendingMemberTile(context, team, member, teamProvider, isArabic)),
        ],
      ),
    );
  }
  
  Widget _buildPendingMemberTile(BuildContext context, Team team, TeamMember member, TeamProvider teamProvider, bool isArabic) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.black.withOpacity(0.2),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Row(
        children: [
          CircleAvatar(
            radius: 20,
            backgroundColor: Colors.white.withOpacity(0.1),
            backgroundImage: member.avatarUrl != null && member.avatarUrl!.isNotEmpty
                ? NetworkImage(member.avatarUrl!)
                : null,
            child: member.avatarUrl == null || member.avatarUrl!.isEmpty
                ? Text(
                    member.username.isNotEmpty ? member.username[0].toUpperCase() : '?',
                    style: const TextStyle(color: Colors.white),
                  )
                : null,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  member.displayName.isNotEmpty ? member.displayName : member.username,
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Text(
                  '@${member.username}',
                  style: TextStyle(
                    color: Colors.white.withOpacity(0.5),
                    fontSize: 12,
                  ),
                ),
              ],
            ),
          ),
          // Approve Button
          IconButton(
            onPressed: () => _approveMember(context, team, member, teamProvider, isArabic),
            icon: const Icon(Icons.check_circle, color: Colors.green),
            tooltip: isArabic ? 'قبول' : 'Approve',
          ),
          // Reject Button
          IconButton(
            onPressed: () => _rejectMember(context, team, member, teamProvider, isArabic),
            icon: const Icon(Icons.cancel, color: Colors.red),
            tooltip: isArabic ? 'رفض' : 'Reject',
          ),
        ],
      ),
    );
  }
  
  void _showQRCodeDialog(BuildContext context, Team team, bool isArabic) {
    final inviteUrl = team.inviteUrl ?? 'https://afftok.com/join/${team.inviteCode ?? team.id}';
    
    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: Colors.transparent,
        child: Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: const Color(0xFF1a1a1a),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                isArabic ? 'رمز QR للدعوة' : 'Invite QR Code',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                team.name,
                style: TextStyle(
                  color: Colors.white.withOpacity(0.7),
                  fontSize: 14,
                ),
              ),
              const SizedBox(height: 20),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: QrImageView(
                  data: inviteUrl,
                  version: QrVersions.auto,
                  size: 200,
                  backgroundColor: Colors.white,
                  foregroundColor: Colors.black,
                ),
              ),
              const SizedBox(height: 16),
              Text(
                isArabic ? 'امسح الرمز للانضمام للفريق' : 'Scan to join the team',
                style: TextStyle(
                  color: Colors.white.withOpacity(0.5),
                  fontSize: 12,
                ),
              ),
              const SizedBox(height: 20),
              TextButton(
                onPressed: () => Navigator.pop(context),
                child: Text(
                  isArabic ? 'إغلاق' : 'Close',
                  style: const TextStyle(color: Color(0xFFFF006E)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
  
  Future<void> _approveMember(BuildContext context, Team team, TeamMember member, TeamProvider teamProvider, bool isArabic) async {
    final teamService = TeamService();
    final result = await teamService.approveMember(team.id, member.id);
    
    if (result['success'] == true) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(isArabic ? 'تمت الموافقة على ${member.username}' : '${member.username} approved'),
          backgroundColor: Colors.green,
        ),
      );
      teamProvider.loadTeams();
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(result['error'] ?? 'Error'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }
  
  Future<void> _rejectMember(BuildContext context, Team team, TeamMember member, TeamProvider teamProvider, bool isArabic) async {
    final teamService = TeamService();
    final result = await teamService.rejectMember(team.id, member.id);
    
    if (result['success'] == true) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(isArabic ? 'تم رفض ${member.username}' : '${member.username} rejected'),
          backgroundColor: Colors.orange,
        ),
      );
      teamProvider.loadTeams();
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(result['error'] ?? 'Error'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }
  
  void _confirmDeleteTeam(BuildContext context, AppLocalizations lang, TeamProvider teamProvider, Team team) {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    final scaffoldMessenger = ScaffoldMessenger.of(context);
    
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        backgroundColor: const Color(0xFF1a1a1a),
        title: Text(
          isArabic ? 'حذف الفريق' : 'Delete Team',
          style: const TextStyle(color: Colors.white),
        ),
        content: Text(
          isArabic 
              ? 'هل أنت متأكد من حذف الفريق "${team.name}"؟ سيتم إزالة جميع الأعضاء.'
              : 'Are you sure you want to delete "${team.name}"? All members will be removed.',
          style: const TextStyle(color: Colors.white70),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: Text(lang.cancel, style: const TextStyle(color: Colors.white54)),
          ),
          ElevatedButton(
            onPressed: () async {
              Navigator.pop(dialogContext);
              
              final teamService = TeamService();
              final result = await teamService.deleteTeam(team.id);
              
              if (result['success'] == true) {
                scaffoldMessenger.showSnackBar(
                  SnackBar(
                    content: Text(isArabic ? 'تم حذف الفريق' : 'Team deleted'),
                    backgroundColor: Colors.green,
                  ),
                );
                teamProvider.loadTeams();
              } else {
                scaffoldMessenger.showSnackBar(
                  SnackBar(
                    content: Text(result['error'] ?? 'Error'),
                    backgroundColor: Colors.red,
                  ),
                );
              }
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: Text(isArabic ? 'حذف' : 'Delete'),
          ),
        ],
      ),
    );
  }

  Widget _buildLeaderboardTab(BuildContext context, AppLocalizations lang, TeamProvider teamProvider) {
    final teams = teamProvider.teams;
    
    if (teams.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.leaderboard_outlined,
              size: 64,
              color: Colors.white.withOpacity(0.3),
            ),
            const SizedBox(height: 16),
            Text(
              lang.noTeamsYet,
              style: TextStyle(
                color: Colors.white.withOpacity(0.5),
                fontSize: 16,
              ),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () => _showCreateTeamDialog(context, lang, teamProvider),
              icon: const Icon(Icons.add),
              label: Text(lang.createFirstTeam),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFFFF006E),
                foregroundColor: Colors.white,
              ),
            ),
          ],
        ),
      );
    }
    
    return RefreshIndicator(
      onRefresh: () => teamProvider.loadTeams(),
      color: const Color(0xFFFF006E),
      child: ListView.builder(
        padding: const EdgeInsets.all(20),
        itemCount: teams.length,
        itemBuilder: (context, index) {
          final team = teams[index];
          return _buildLeaderboardItem(context, lang, teamProvider, team, index + 1);
        },
      ),
    );
  }

  Widget _buildLeaderboardItem(BuildContext context, AppLocalizations lang, TeamProvider teamProvider, Team team, int rank) {
    return GestureDetector(
      onTap: () => _showTeamDetailsSheet(context, lang, teamProvider, team, rank),
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white.withOpacity(0.05),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: rank <= 3
                ? Color(team.rank.color).withOpacity(0.5)
                : Colors.white.withOpacity(0.1),
          ),
        ),
        child: Row(
          children: [
            // Rank
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: rank == 1
                    ? const Color(0xFFFFD700)
                    : rank == 2
                        ? const Color(0xFFC0C0C0)
                        : rank == 3
                            ? const Color(0xFFCD7F32)
                            : Colors.white.withOpacity(0.1),
              ),
              child: Center(
                child: Text(
                  '$rank',
                  style: TextStyle(
                    color: rank <= 3 ? Colors.black : Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            
            // Team Info
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Text(
                        team.name,
                        style: const TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(team.rank.emoji),
                    ],
                  ),
                  Text(
                    '${team.memberCount} ${lang.members} • ${team.stats.totalPoints} ${lang.points}',
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.5),
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
            
            // Arrow indicator (tap to see details)
            Icon(
              Icons.chevron_right,
              color: Colors.white.withOpacity(0.3),
            ),
          ],
        ),
      ),
    );
  }

  // Show team details with members
  void _showTeamDetailsSheet(BuildContext context, AppLocalizations lang, TeamProvider teamProvider, Team team, int rank) {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.75,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        builder: (context, scrollController) => Container(
          decoration: const BoxDecoration(
            color: Color(0xFF1A1A1A),
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: Column(
            children: [
              // Handle
              Container(
                margin: const EdgeInsets.symmetric(vertical: 12),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.3),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              
              // Team Header
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 24),
                child: Column(
                  children: [
                    // Rank Badge
                    Container(
                      width: 60,
                      height: 60,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        gradient: LinearGradient(
                          colors: rank == 1
                              ? [const Color(0xFFFFD700), const Color(0xFFFFB800)]
                              : rank == 2
                                  ? [const Color(0xFFC0C0C0), const Color(0xFF909090)]
                                  : rank == 3
                                      ? [const Color(0xFFCD7F32), const Color(0xFFA0522D)]
                                      : [const Color(0xFFFF006E), const Color(0xFFFF4D00)],
                        ),
                      ),
                      child: Center(
                        child: Text(
                          '#$rank',
                          style: TextStyle(
                            color: rank <= 3 ? Colors.black : Colors.white,
                            fontWeight: FontWeight.bold,
                            fontSize: 20,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),
                    
                    // Team Name
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          team.name,
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 24,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(team.rank.emoji, style: const TextStyle(fontSize: 24)),
                      ],
                    ),
                    if (team.description != null && team.description!.isNotEmpty) ...[
                      const SizedBox(height: 8),
                      Text(
                        team.description!,
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.6),
                          fontSize: 14,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ],
                    
                    const SizedBox(height: 16),
                    
                    // Stats Row
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                      children: [
                        _buildStatItem(
                          icon: Icons.people,
                          value: '${team.memberCount}',
                          label: lang.members,
                        ),
                        _buildStatItem(
                          icon: Icons.touch_app,
                          value: '${team.stats.totalClicks}',
                          label: isArabic ? 'نقرات' : 'Clicks',
                        ),
                        _buildStatItem(
                          icon: Icons.check_circle,
                          value: '${team.stats.totalConversions}',
                          label: isArabic ? 'تحويلات' : 'Conversions',
                        ),
                        _buildStatItem(
                          icon: Icons.star,
                          value: '${team.stats.totalPoints}',
                          label: lang.points,
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              
              const SizedBox(height: 16),
              
              // Divider
              Divider(color: Colors.white.withOpacity(0.1)),
              
              // Members Header
              Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    const Icon(Icons.group, color: Color(0xFFFF006E), size: 20),
                    const SizedBox(width: 8),
                    Text(
                      isArabic ? 'الأعضاء' : 'Team Members',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
              
              // Members List
              Expanded(
                child: ListView.builder(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  itemCount: team.members.length,
                  itemBuilder: (context, index) {
                    final member = team.members[index];
                    final isOwner = member.role == 'owner';
                    
                    return Container(
                      margin: const EdgeInsets.only(bottom: 8),
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.05),
                        borderRadius: BorderRadius.circular(12),
                        border: isOwner
                            ? Border.all(color: const Color(0xFFFFD700).withOpacity(0.5))
                            : null,
                      ),
                      child: Row(
                        children: [
                          // Member Avatar
                          Container(
                            width: 48,
                            height: 48,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              gradient: LinearGradient(
                                colors: isOwner
                                    ? [const Color(0xFFFFD700), const Color(0xFFFFB800)]
                                    : [const Color(0xFFFF006E), const Color(0xFFFF4D00)],
                              ),
                            ),
                            child: member.avatarUrl != null && member.avatarUrl!.isNotEmpty
                                ? ClipOval(
                                    child: Image.network(
                                      member.avatarUrl!,
                                      fit: BoxFit.cover,
                                      errorBuilder: (_, __, ___) => Center(
                                        child: Text(
                                          (member.displayName.isNotEmpty ? member.displayName : member.username)[0].toUpperCase(),
                                          style: const TextStyle(
                                            color: Colors.white,
                                            fontWeight: FontWeight.bold,
                                            fontSize: 18,
                                          ),
                                        ),
                                      ),
                                    ),
                                  )
                                : Center(
                                    child: Text(
                                      (member.displayName.isNotEmpty ? member.displayName : member.username)[0].toUpperCase(),
                                      style: const TextStyle(
                                        color: Colors.white,
                                        fontWeight: FontWeight.bold,
                                        fontSize: 18,
                                      ),
                                    ),
                                  ),
                          ),
                          const SizedBox(width: 12),
                          
                          // Member Info
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  children: [
                                    Flexible(
                                      child: Text(
                                        member.displayName.isNotEmpty ? member.displayName : member.username,
                                        style: const TextStyle(
                                          color: Colors.white,
                                          fontWeight: FontWeight.bold,
                                          fontSize: 15,
                                        ),
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                    ),
                                    if (isOwner) ...[
                                      const SizedBox(width: 8),
                                      Container(
                                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                                        decoration: BoxDecoration(
                                          color: const Color(0xFFFFD700),
                                          borderRadius: BorderRadius.circular(10),
                                        ),
                                        child: Text(
                                          isArabic ? 'القائد' : 'Leader',
                                          style: const TextStyle(
                                            color: Colors.black,
                                            fontSize: 10,
                                            fontWeight: FontWeight.bold,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  '@${member.username}',
                                  style: TextStyle(
                                    color: Colors.white.withOpacity(0.5),
                                    fontSize: 12,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          
                          // Member Stats
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Row(
                                children: [
                                  Icon(
                                    Icons.touch_app,
                                    size: 14,
                                    color: Colors.white.withOpacity(0.5),
                                  ),
                                  const SizedBox(width: 4),
                                  Text(
                                    '${member.referrals}',
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 4),
                              Row(
                                children: [
                                  Icon(
                                    Icons.star,
                                    size: 14,
                                    color: const Color(0xFFFFD700).withOpacity(0.7),
                                  ),
                                  const SizedBox(width: 4),
                                  Text(
                                    '${member.points}',
                                    style: TextStyle(
                                      color: const Color(0xFFFFD700).withOpacity(0.9),
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ],
                      ),
                    );
                  },
                ),
              ),
              
              // Join Button (if not in any team)
              if (teamProvider.userTeam == null)
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Container(
                    width: double.infinity,
                    decoration: BoxDecoration(
                      gradient: team.isFull
                          ? null
                          : const LinearGradient(
                              colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                            ),
                      color: team.isFull ? Colors.grey : null,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: ElevatedButton(
                      onPressed: team.isFull
                          ? null
                          : () {
                              Navigator.pop(context);
                              _confirmJoinTeam(context, lang, teamProvider, team);
                            },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.transparent,
                        shadowColor: Colors.transparent,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                      child: Text(
                        team.isFull
                            ? (isArabic ? 'الفريق ممتلئ' : 'Team is Full')
                            : (isArabic ? 'انضم للفريق' : 'Join This Team'),
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  void _showCreateTeamDialog(BuildContext context, AppLocalizations lang, TeamProvider teamProvider) {
    final nameController = TextEditingController();
    final descriptionController = TextEditingController();
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: Colors.grey[900],
        title: Text(
          lang.createTeam,
          style: const TextStyle(color: Colors.white),
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: nameController,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                labelText: lang.teamName,
                labelStyle: TextStyle(color: Colors.white.withOpacity(0.5)),
                enabledBorder: OutlineInputBorder(
                  borderSide: BorderSide(color: Colors.white.withOpacity(0.2)),
                  borderRadius: BorderRadius.circular(12),
                ),
                focusedBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Color(0xFFFF006E)),
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: descriptionController,
              style: const TextStyle(color: Colors.white),
              maxLines: 3,
              decoration: InputDecoration(
                labelText: lang.description,
                labelStyle: TextStyle(color: Colors.white.withOpacity(0.5)),
                enabledBorder: OutlineInputBorder(
                  borderSide: BorderSide(color: Colors.white.withOpacity(0.2)),
                  borderRadius: BorderRadius.circular(12),
                ),
                focusedBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Color(0xFFFF006E)),
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(lang.cancel, style: const TextStyle(color: Colors.white54)),
          ),
          ElevatedButton(
            onPressed: () async {
              if (nameController.text.trim().isEmpty) return;
              
              Navigator.pop(context);
              
              final success = await teamProvider.createTeam(
                name: nameController.text.trim(),
                description: descriptionController.text.trim(),
              );
              
              if (success && mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(lang.teamCreatedSuccessfully),
                    backgroundColor: Colors.green,
                  ),
                );
              } else if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(teamProvider.error ?? lang.failedToCreateTeam),
                    backgroundColor: Colors.red,
                  ),
                );
              }
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFFFF006E),
            ),
            child: Text(lang.create),
          ),
        ],
      ),
    );
  }

  void _confirmJoinTeam(BuildContext context, AppLocalizations lang, TeamProvider teamProvider, Team team) {
    final scaffoldMessenger = ScaffoldMessenger.of(context);
    
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        backgroundColor: Colors.grey[900],
        title: Text(
          lang.joinTeam,
          style: const TextStyle(color: Colors.white),
        ),
        content: Text(
          '${lang.confirmJoinTeam} "${team.name}"?',
          style: const TextStyle(color: Colors.white70),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: Text(lang.cancel, style: const TextStyle(color: Colors.white54)),
          ),
          ElevatedButton(
            onPressed: () async {
              Navigator.pop(dialogContext);
              
              final success = await teamProvider.joinTeam(team.id);
              
              if (success && mounted) {
                scaffoldMessenger.showSnackBar(
                  SnackBar(
                    content: Text(lang.joinedTeamSuccessfully),
                    backgroundColor: Colors.green,
                  ),
                );
                _tabController.animateTo(0);
              } else if (mounted) {
                scaffoldMessenger.showSnackBar(
                  SnackBar(
                    content: Text(teamProvider.error ?? lang.failedToJoinTeam),
                    backgroundColor: Colors.red,
                  ),
                );
              }
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFFFF006E),
            ),
            child: Text(lang.join),
          ),
        ],
      ),
    );
  }

  void _confirmLeaveTeam(BuildContext context, AppLocalizations lang, TeamProvider teamProvider) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: Colors.grey[900],
        title: Text(
          lang.leaveTeam,
          style: const TextStyle(color: Colors.white),
        ),
        content: Text(
          lang.confirmLeaveTeam,
          style: const TextStyle(color: Colors.white70),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(lang.cancel, style: const TextStyle(color: Colors.white54)),
          ),
          ElevatedButton(
            onPressed: () async {
              Navigator.pop(context);
              
              final success = await teamProvider.leaveTeam();
              
              if (success && mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(lang.leftTeamSuccessfully),
                    backgroundColor: Colors.green,
                  ),
                );
              } else if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(teamProvider.error ?? lang.failedToLeaveTeam),
                    backgroundColor: Colors.red,
                  ),
                );
              }
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red,
            ),
            child: Text(lang.leave),
          ),
        ],
      ),
    );
  }

  void _shareTeamInvite(Team team) {
    final inviteLink = team.inviteUrl ?? 'https://afftok.com/join/${team.inviteCode ?? team.id}';
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    
    final message = isArabic
        ? 'انضم لفريقي "${team.name}" على AffTok!\n\n$inviteLink'
        : 'Join my team "${team.name}" on AffTok!\n\n$inviteLink';
    
    Share.share(message);
  }

  // ============================================
  // CONTESTS TAB
  // ============================================

  Widget _buildContestsTab(BuildContext context, AppLocalizations lang) {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    
    if (_loadingContests) {
      return const Center(
        child: CircularProgressIndicator(color: Color(0xFFFF006E)),
      );
    }
    
    if (_activeContests.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.emoji_events_outlined,
              size: 80,
              color: Colors.white.withOpacity(0.3),
            ),
            const SizedBox(height: 16),
            Text(
              isArabic ? 'لا توجد مسابقات حالياً' : 'No active contests',
              style: TextStyle(
                color: Colors.white.withOpacity(0.5),
                fontSize: 18,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              isArabic ? 'ترقب المسابقات القادمة!' : 'Stay tuned for upcoming contests!',
              style: TextStyle(
                color: Colors.white.withOpacity(0.3),
                fontSize: 14,
              ),
            ),
          ],
        ),
      );
    }
    
    return RefreshIndicator(
      onRefresh: _loadContests,
      color: const Color(0xFFFF006E),
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _activeContests.length,
        itemBuilder: (context, index) {
          final contest = _activeContests[index];
          return _buildContestCard(context, contest, isArabic);
        },
      ),
    );
  }

  Widget _buildContestCard(BuildContext context, Contest contest, bool isArabic) {
    final title = contest.getLocalizedTitle(isArabic ? 'ar' : 'en');
    final description = contest.getLocalizedDescription(isArabic ? 'ar' : 'en');
    final prizeTitle = contest.getLocalizedPrizeTitle(isArabic ? 'ar' : 'en');
    
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            const Color(0xFFFF006E).withOpacity(0.2),
            const Color(0xFFFF6B35).withOpacity(0.1),
          ],
        ),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFFFF006E).withOpacity(0.3),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header with image
          if (contest.imageUrl != null && contest.imageUrl!.isNotEmpty)
            ClipRRect(
              borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
              child: Image.network(
                contest.imageUrl!,
                height: 150,
                width: double.infinity,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) {
                  return Container(
                    height: 150,
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          const Color(0xFFFF006E),
                          const Color(0xFFFF6B35),
                        ],
                      ),
                    ),
                    child: const Center(
                      child: Icon(Icons.emoji_events, color: Colors.white, size: 60),
                    ),
                  );
                },
              ),
            )
          else
            Container(
              height: 120,
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [
                    const Color(0xFFFF006E),
                    const Color(0xFFFF6B35),
                  ],
                ),
                borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
              ),
              child: const Center(
                child: Icon(Icons.emoji_events, color: Colors.white, size: 60),
              ),
            ),
          
          // Content
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Title and Type Badge
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        title,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: contest.contestType == 'team' 
                            ? Colors.blue.withOpacity(0.3)
                            : Colors.green.withOpacity(0.3),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        contest.contestType == 'team'
                            ? (isArabic ? 'فرق' : 'Teams')
                            : (isArabic ? 'أفراد' : 'Individual'),
                        style: TextStyle(
                          color: contest.contestType == 'team' 
                              ? Colors.blue[300]
                              : Colors.green[300],
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ],
                ),
                
                if (description != null && description.isNotEmpty) ...[
                  const SizedBox(height: 8),
                  Text(
                    description,
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.7),
                      fontSize: 14,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
                
                const SizedBox(height: 16),
                
                // Prize Section
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.amber.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: Colors.amber.withOpacity(0.3)),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.military_tech, color: Colors.amber, size: 28),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              isArabic ? 'الجائزة' : 'Prize',
                              style: TextStyle(
                                color: Colors.amber.withOpacity(0.7),
                                fontSize: 12,
                              ),
                            ),
                            Text(
                              prizeTitle ?? (contest.prizeAmount > 0 
                                  ? '${contest.prizeAmount} ${contest.prizeCurrency}'
                                  : (isArabic ? 'جائزة خاصة' : 'Special Prize')),
                              style: const TextStyle(
                                color: Colors.amber,
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                
                const SizedBox(height: 12),
                
                // Target & Stats Row
                Row(
                  children: [
                    _buildContestStatItem(
                      icon: _getTargetIcon(contest.targetType),
                      label: isArabic ? 'الهدف' : 'Target',
                      value: '${contest.targetValue} ${_getTargetLabel(contest.targetType, isArabic)}',
                    ),
                    const SizedBox(width: 16),
                    _buildContestStatItem(
                      icon: Icons.groups,
                      label: isArabic ? 'المشاركون' : 'Participants',
                      value: '${contest.participantsCount}',
                    ),
                  ],
                ),
                
                const SizedBox(height: 12),
                
                // Time Left
                Row(
                  children: [
                    Icon(
                      Icons.timer_outlined,
                      color: contest.isEnded ? Colors.red : Colors.orange,
                      size: 18,
                    ),
                    const SizedBox(width: 6),
                    Text(
                      contest.isEnded 
                          ? (isArabic ? 'انتهت المسابقة' : 'Contest ended')
                          : '${isArabic ? 'المتبقي:' : 'Time left:'} ${contest.timeLeftFormatted}',
                      style: TextStyle(
                        color: contest.isEnded ? Colors.red : Colors.orange,
                        fontSize: 14,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                
                const SizedBox(height: 16),
                
                // Action Button
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton.icon(
                    onPressed: contest.isActive ? () => _joinContest(contest) : null,
                    icon: Icon(
                      contest.isActive ? Icons.add : Icons.lock_outline,
                      size: 18,
                    ),
                    label: Text(
                      contest.isActive 
                          ? (isArabic ? 'انضم للمسابقة' : 'Join Contest')
                          : (isArabic ? 'غير متاحة' : 'Not Available'),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: contest.isActive 
                          ? const Color(0xFFFF006E)
                          : Colors.grey[700],
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 12),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildContestStatItem({
    required IconData icon,
    required String label,
    required String value,
  }) {
    return Expanded(
      child: Row(
        children: [
          Icon(icon, color: Colors.white54, size: 18),
          const SizedBox(width: 6),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: TextStyle(
                  color: Colors.white.withOpacity(0.5),
                  fontSize: 11,
                ),
              ),
              Text(
                value,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  IconData _getTargetIcon(String targetType) {
    switch (targetType) {
      case 'clicks':
        return Icons.touch_app;
      case 'conversions':
        return Icons.shopping_cart;
      case 'referrals':
        return Icons.people;
      case 'points':
        return Icons.stars;
      default:
        return Icons.flag;
    }
  }

  String _getTargetLabel(String targetType, bool isArabic) {
    switch (targetType) {
      case 'clicks':
        return isArabic ? 'نقرة' : 'clicks';
      case 'conversions':
        return isArabic ? 'تحويل' : 'conversions';
      case 'referrals':
        return isArabic ? 'إحالة' : 'referrals';
      case 'points':
        return isArabic ? 'نقطة' : 'points';
      default:
        return '';
    }
  }

  Future<void> _joinContest(Contest contest) async {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    final scaffoldMessenger = ScaffoldMessenger.of(context);
    
    final result = await _contestService.joinContest(contest.id);
    final success = result['success'] as bool;
    final message = result['message'] as String;
    
    if (success) {
      scaffoldMessenger.showSnackBar(
        SnackBar(
          content: Text(isArabic ? 'تم الانضمام للمسابقة بنجاح!' : message),
          backgroundColor: Colors.green,
        ),
      );
      _loadContests();
    } else {
      scaffoldMessenger.showSnackBar(
        SnackBar(
          content: Text(message),
          backgroundColor: Colors.red,
        ),
      );
    }
  }
}
