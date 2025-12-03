import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';
import 'package:provider/provider.dart';
import '../utils/app_localizations.dart';
import '../models/team.dart';
import '../providers/auth_provider.dart';
import '../providers/team_provider.dart';

class TeamsScreen extends StatefulWidget {
  const TeamsScreen({Key? key}) : super(key: key);

  @override
  State<TeamsScreen> createState() => _TeamsScreenState();
}

class _TeamsScreenState extends State<TeamsScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    
    // Load teams when screen opens
    WidgetsBinding.instance.addPostFrameCallback((_) {
      Provider.of<TeamProvider>(context, listen: false).loadTeams();
    });
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
              ],
            ),
          ),
          body: TabBarView(
            controller: _tabController,
            children: [
              _buildMyTeamTab(context, lang, teamProvider),
              _buildLeaderboardTab(context, lang, teamProvider),
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
    return Column(
      children: [
        // Invite Button
        Container(
          width: double.infinity,
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
            ),
            borderRadius: BorderRadius.circular(12),
          ),
          child: ElevatedButton.icon(
            onPressed: () => _shareTeamInvite(team),
            icon: const Icon(Icons.share, color: Colors.white),
            label: Text(lang.inviteMembers),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.transparent,
              shadowColor: Colors.transparent,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(vertical: 16),
            ),
          ),
        ),
        
        const SizedBox(height: 12),
        
        // Leave Team Button
        TextButton(
          onPressed: () => _confirmLeaveTeam(context, lang, teamProvider),
          child: Text(
            lang.leaveTeam,
            style: const TextStyle(color: Colors.red),
          ),
        ),
      ],
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
    return Container(
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
          
          // Join Button
          if (teamProvider.userTeam == null)
            ElevatedButton(
              onPressed: team.isFull
                  ? null
                  : () => _confirmJoinTeam(context, lang, teamProvider, team),
              style: ElevatedButton.styleFrom(
                backgroundColor: team.isFull ? Colors.grey : const Color(0xFFFF006E),
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              ),
              child: Text(team.isFull ? lang.full : lang.join),
            ),
        ],
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
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
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
            onPressed: () => Navigator.pop(context),
            child: Text(lang.cancel, style: const TextStyle(color: Colors.white54)),
          ),
          ElevatedButton(
            onPressed: () async {
              Navigator.pop(context);
              
              final success = await teamProvider.joinTeam(team.id);
              
              if (success && mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(lang.joinedTeamSuccessfully),
                    backgroundColor: Colors.green,
                  ),
                );
                _tabController.animateTo(0);
              } else if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
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
    final inviteLink = 'https://afftok.com/team/${team.id}';
    Share.share(
      'Join my team "${team.name}" on AffTok!\n\n$inviteLink',
    );
  }
}
