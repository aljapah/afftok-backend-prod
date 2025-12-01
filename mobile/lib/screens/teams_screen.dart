import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:provider/provider.dart';
import '../utils/app_localizations.dart';
import '../models/team.dart';
import '../models/user.dart';
import '../providers/auth_provider.dart';

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
    _tabController = TabController(length: 3, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        final user = authProvider.currentUser;
        
        if (user == null) {
          return Scaffold(
            backgroundColor: Colors.black,
            appBar: AppBar(
              backgroundColor: Colors.black,
              title: Text(lang.teams),
            ),
            body: const Center(
              child: CircularProgressIndicator(),
            ),
          );
        }
    
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
          if (user.isInTeam)
            IconButton(
              icon: const Icon(Icons.settings, color: Colors.white),
              onPressed: () {},
            ),
        ],
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: const Color(0xFFFF006E),
          labelColor: Colors.white,
          unselectedLabelColor: Colors.white.withValues(alpha: 0.5),
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
          _buildMyTeamTab(context, lang),
          _buildLeaderboardTab(context, lang),
          _buildChallengesTab(context, lang),
        ],
      ),
    );
      }
    );
  }

  Widget _buildMyTeamTab(BuildContext context, AppLocalizations lang) {
    final team = sampleTeam;
    
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildTeamHeader(context, team, lang),
          
          const SizedBox(height: 24),
          
          _buildTeamStats(team, lang),
          
          const SizedBox(height: 24),
          
          _buildTeamMembers(team, lang),
          
          const SizedBox(height: 24),
          
          _buildInviteButton(context, lang),
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
            Color(team.rank.color).withValues(alpha: 0.3),
            Color(team.rank.color).withValues(alpha: 0.1),
          ],
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: Color(team.rank.color).withValues(alpha: 0.5),
        ),
      ),
      child: Column(
        children: [
          Container(
            width: 80,
            height: 80,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              gradient: LinearGradient(
                colors: [
                  Color(team.rank.color),
                  Color(team.rank.color).withValues(alpha: 0.6),
                ],
              ),
              boxShadow: [
                BoxShadow(
                  color: Color(team.rank.color).withValues(alpha: 0.5),
                  blurRadius: 20,
                  spreadRadius: 2,
                ),
              ],
            ),
            child: Center(
              child: Text(
                team.rank.emoji,
                style: const TextStyle(fontSize: 40),
              ),
            ),
          ),
          
          const SizedBox(height: 16),
          
          Text(
            team.name,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 24,
              fontWeight: FontWeight.bold,
            ),
          ),
          
          const SizedBox(height: 8),
          
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            decoration: BoxDecoration(
              color: Color(team.rank.color).withValues(alpha: 0.2),
              borderRadius: BorderRadius.circular(20),
              border: Border.all(
                color: Color(team.rank.color).withValues(alpha: 0.5),
              ),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  '${team.rank.displayName} Team',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const SizedBox(width: 8),
                const Text('•', style: TextStyle(color: Colors.white)),
                const SizedBox(width: 8),
                Text(
                  '${lang.rank} #${team.stats.globalRank}',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
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
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          lang.teamPerformance,
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
              child: _StatCard(
                icon: Icons.people,
                label: lang.referrals,
                value: team.stats.totalReferrals.toString(),
                color: const Color(0xFFFF006E),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _StatCard(
                icon: Icons.touch_app,
                label: lang.clicks,
                value: team.stats.totalClicks.toString(),
                color: const Color(0xFF00D9FF),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _StatCard(
                icon: Icons.check_circle,
                label: lang.conversions,
                value: '${team.stats.totalConversions}',
                color: const Color(0xFF00FF88),
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.05),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: Colors.white.withValues(alpha: 0.1),
            ),
          ),
          child: Column(
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      const Text('📈', style: TextStyle(fontSize: 16)),
                      const SizedBox(width: 8),
                      Text(
                        lang.thisMonth,
                        style: TextStyle(
                          color: Colors.white.withValues(alpha: 0.7),
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                  Text(
                    '${team.stats.monthlyConversions} ${lang.conversions} (+${team.stats.conversionRate.toStringAsFixed(1)}%)',
                    style: const TextStyle(
                      color: Color(0xFF00FF88),
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          const Text('🎯', style: TextStyle(fontSize: 16)),
                          const SizedBox(width: 8),
                          Text(
                            lang.teamGoal,
                            style: TextStyle(
                              color: Colors.white.withValues(alpha: 0.7),
                              fontSize: 14,
                            ),
                          ),
                        ],
                      ),
                      Text(
                        '${team.stats.goalProgress}/${team.stats.goalTarget} ${lang.conversions}',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  ClipRRect(
                    borderRadius: BorderRadius.circular(10),
                    child: LinearProgressIndicator(
                      value: team.stats.goalPercentage / 100,
                      backgroundColor: Colors.white.withValues(alpha: 0.1),
                      valueColor: const AlwaysStoppedAnimation<Color>(
                        Color(0xFF00FF88),
                      ),
                      minHeight: 8,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '${team.stats.goalPercentage.toInt()}% ${lang.completed}',
                    style: TextStyle(
                      color: Colors.white.withValues(alpha: 0.6),
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildTeamMembers(Team team, AppLocalizations lang) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              lang.teamMembers,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(
              '${team.memberCount}/${team.maxMembers}',
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.6),
                fontSize: 14,
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        ListView.separated(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: team.members.length,
          separatorBuilder: (context, index) => const SizedBox(height: 12),
          itemBuilder: (context, index) {
            final member = team.members[index];
            final isCurrentUser = false; // TODO: pass from parent
            return _buildMemberCard(member, isCurrentUser, lang);
          },
        ),
      ],
    );
  }

  Widget _buildMemberCard(TeamMember member, bool isCurrentUser, AppLocalizations lang) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isCurrentUser
            ? const Color(0xFFFF006E).withValues(alpha: 0.1)
            : Colors.white.withValues(alpha: 0.05),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: isCurrentUser
              ? const Color(0xFFFF006E).withValues(alpha: 0.3)
              : Colors.white.withValues(alpha: 0.1),
        ),
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              gradient: LinearGradient(
                colors: member.teamRank <= 3
                    ? [const Color(0xFFFF006E), const Color(0xFFFF4D00)]
                    : [Colors.grey[700]!, Colors.grey[800]!],
              ),
            ),
            child: Center(
              child: Text(
                '#${member.teamRank}',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
          
          const SizedBox(width: 12),
          
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
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    if (member.isLeader) ...[
                      const SizedBox(width: 6),
                      const Text('👑', style: TextStyle(fontSize: 14)),
                    ],
                    if (isCurrentUser) ...[
                      const SizedBox(width: 6),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                        decoration: BoxDecoration(
                          color: const Color(0xFFFF006E),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          lang.you,
                          style: const TextStyle(
                            color: Colors.white,
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
                  '${member.referrals} ${lang.referrals} • ${member.conversions} ${lang.conversions}',
                  style: TextStyle(
                    color: Colors.white.withValues(alpha: 0.6),
                    fontSize: 12,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInviteButton(BuildContext context, AppLocalizations lang) {
    return SizedBox(
      width: double.infinity,
      height: 50,
      child: ElevatedButton.icon(
        onPressed: () {
          _showInviteDialog(context, lang);
        },
        icon: const Icon(Icons.person_add, color: Colors.white),
        label: Text(
          lang.inviteMembers,
          style: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
        style: ElevatedButton.styleFrom(
          backgroundColor: Colors.transparent,
          shadowColor: Colors.transparent,
          padding: EdgeInsets.zero,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(25),
          ),
        ),
      ).decorated(
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
          ),
          borderRadius: BorderRadius.circular(25),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFFFF006E).withValues(alpha: 0.3),
              blurRadius: 12,
              offset: const Offset(0, 4),
            ),
          ],
        ),
      ),
    );
  }

  void _showInviteDialog(BuildContext context, AppLocalizations lang) {
    final team = sampleTeam;
    final inviteLink = 'https://offtok.com/team/${team.name.toLowerCase().replaceAll(' ', '')}';
    
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (BuildContext context) {
        return Container(
          decoration: const BoxDecoration(
            color: Colors.black,
            borderRadius: BorderRadius.only(
              topLeft: Radius.circular(30),
              topRight: Radius.circular(30),
            ),
          ),
          padding: const EdgeInsets.all(30),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              
              Text(
                lang.inviteMembers,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 18,
                  fontWeight: FontWeight.bold
                ),
              ),
              const SizedBox(height: 30),
              
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(15),
                ),
                child: QrImageView(
                  data: inviteLink,
                  version: QrVersions.auto,
                  size: 200.0,
                  backgroundColor: Colors.white,
                ),
              ),
              const SizedBox(height: 20),
              
              Text(
                'رابط دعوة الفريق',
                style: TextStyle(
                  color: Colors.white.withValues(alpha: 0.7),
                  fontSize: 14
                ),
              ),
              const SizedBox(height: 10),
              
              
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 12),
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(color: Colors.white.withValues(alpha: 0.2)),
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        inviteLink,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    const SizedBox(width: 10),
                    InkWell(
                      onTap: () {
                        Clipboard.setData(ClipboardData(text: inviteLink));
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(
                            content: Text(lang.linkCopied),
                            backgroundColor: const Color(0xFFFF006E),
                            duration: const Duration(seconds: 2),
                          ),
                        );
                      },
                      child: const Icon(Icons.copy, color: Color(0xFFFF006E), size: 20),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              
              
              SizedBox(
                width: double.infinity,
                height: 50,
                child: ElevatedButton.icon(
                  onPressed: () {
                    final shareUrl = inviteLink;
                    Share.share(
                      shareUrl,
                      subject: lang.inviteMembers,
                    );
                  },
                  icon: const Icon(Icons.share, color: Colors.white),
                  label: Text(
                    lang.shareLink,
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.transparent,
                    shadowColor: Colors.transparent,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(25),
                    ),
                  ),
                ).decorated(
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                    ),
                    borderRadius: BorderRadius.circular(25),
                  ),
                ),
              ),
              const SizedBox(height: 10),
            ],
          ),
        );
      },
    );
  }

  Widget _buildLeaderboardTab(BuildContext context, AppLocalizations lang) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            lang.topTeamsThisMonth,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 20),
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: topTeams.length,
            separatorBuilder: (context, index) => const SizedBox(height: 12),
            itemBuilder: (context, index) {
              final team = topTeams[index];
              final isMyTeam = team.id == sampleTeam.id;
              return _buildLeaderboardCard(team, index + 1, isMyTeam, lang);
            },
          ),
        ],
      ),
    );
  }

  Widget _buildLeaderboardCard(Team team, int rank, bool isMyTeam, AppLocalizations lang) {
    String rankEmoji = '';
    if (rank == 1) rankEmoji = '🥇';
    else if (rank == 2) rankEmoji = '🥈';
    else if (rank == 3) rankEmoji = '🥉';
    
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: isMyTeam
            ? LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [
                  const Color(0xFFFF006E).withValues(alpha: 0.2),
                  const Color(0xFFFF4D00).withValues(alpha: 0.1),
                ],
              )
            : null,
        color: isMyTeam ? null : Colors.white.withValues(alpha: 0.05),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: isMyTeam
              ? const Color(0xFFFF006E).withValues(alpha: 0.3)
              : Colors.white.withValues(alpha: 0.1),
        ),
      ),
      child: Row(
        children: [
          Container(
            width: 50,
            height: 50,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              gradient: rank <= 3
                  ? LinearGradient(
                      colors: [
                        Color(team.rank.color),
                        Color(team.rank.color).withValues(alpha: 0.6),
                      ],
                    )
                  : null,
              color: rank > 3 ? Colors.grey[800] : null,
            ),
            child: Center(
              child: Text(
                rankEmoji.isNotEmpty ? rankEmoji : '#$rank',
                style: TextStyle(
                  color: Colors.white,
                  fontSize: rankEmoji.isNotEmpty ? 24 : 16,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
          
          const SizedBox(width: 16),
          
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
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    if (isMyTeam) ...[
                      const SizedBox(width: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: const Color(0xFFFF006E),
                          borderRadius: BorderRadius.circular(6),
                        ),
                        child: Text(
                          lang.yourTeam,
                          style: const TextStyle(
                            color: Colors.white,
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
                  '${team.rank.emoji} ${team.rank.displayName}',
                  style: TextStyle(
                    color: Colors.white.withValues(alpha: 0.6),
                    fontSize: 12,
                  ),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Text(
                      '${team.stats.totalReferrals} ${lang.referrals}',
                      style: TextStyle(
                        color: Colors.white.withValues(alpha: 0.8),
                        fontSize: 13,
                      ),
                    ),
                    const SizedBox(width: 12),
                    const Text('•', style: TextStyle(color: Colors.white)),
                    const SizedBox(width: 12),
                    Text(
                      '${team.stats.totalConversions} ${lang.conversions}',
                      style: const TextStyle(
                        color: Color(0xFF00FF88),
                        fontSize: 13,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChallengesTab(BuildContext context, AppLocalizations lang) {
    final challenge = activeChallenge;
    
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            lang.activeChallenges,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 20),
          _buildChallengeCard(challenge, lang),
        ],
      ),
    );
  }

  Widget _buildChallengeCard(TeamChallenge challenge, AppLocalizations lang) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            const Color(0xFFFF006E).withValues(alpha: 0.3),
            const Color(0xFFFF4D00).withValues(alpha: 0.1),
          ],
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: const Color(0xFFFF006E).withValues(alpha: 0.5),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 50,
                height: 50,
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Center(
                  child: Text('💎', style: TextStyle(fontSize: 24)),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      challenge.companyName,
                      style: TextStyle(
                        color: Colors.white.withValues(alpha: 0.7),
                        fontSize: 12,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      challenge.title,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          
          const SizedBox(height: 16),
          
          Text(
            challenge.description,
            style: TextStyle(
              color: Colors.white.withValues(alpha: 0.8),
              fontSize: 14,
              height: 1.5,
            ),
          ),
          
          const SizedBox(height: 20),
          
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    lang.progress,
                    style: TextStyle(
                      color: Colors.white.withValues(alpha: 0.7),
                      fontSize: 14,
                    ),
                  ),
                  Text(
                    '${challenge.currentProgress}/${challenge.targetReferrals}',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              ClipRRect(
                borderRadius: BorderRadius.circular(10),
                child: LinearProgressIndicator(
                  value: challenge.progressPercentage / 100,
                  backgroundColor: Colors.white.withValues(alpha: 0.2),
                  valueColor: const AlwaysStoppedAnimation<Color>(
                    Color(0xFF00FF88),
                  ),
                  minHeight: 10,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                '${challenge.progressPercentage.toInt()}% ${lang.completed}',
                style: TextStyle(
                  color: Colors.white.withValues(alpha: 0.6),
                  fontSize: 12,
                ),
              ),
            ],
          ),
          
          const SizedBox(height: 20),
          
          Row(
            children: [
              Expanded(
                child: _buildChallengeStat(
                  '⏰',
                  challenge.timeLeftText,
                  lang.timeLeft,
                ),
              ),
              Expanded(
                child: _buildChallengeStat(
                  '🏆',
                  '#${challenge.currentRank}',
                  lang.currentRank,
                ),
              ),
              Expanded(
                child: _buildChallengeStat(
                  '💰',
                  '\$${challenge.reward.toInt()}',
                  lang.reward,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildChallengeStat(String emoji, String value, String label) {
    return Column(
      children: [
        Text(emoji, style: const TextStyle(fontSize: 20)),
        const SizedBox(height: 8),
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 16,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: TextStyle(
            color: Colors.white.withValues(alpha: 0.6),
            fontSize: 11,
          ),
          textAlign: TextAlign.center,
        ),
      ],
    );
  }
}

class _StatCard extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  final Color color;

  const _StatCard({
    required this.icon,
    required this.label,
    required this.value,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: color.withValues(alpha: 0.3),
        ),
      ),
      child: Column(
        children: [
          Icon(icon, color: color, size: 24),
          const SizedBox(height: 8),
          Text(
            value,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 16,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            label,
            style: TextStyle(
              color: Colors.white.withValues(alpha: 0.6),
              fontSize: 11,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}

extension WidgetExtension on Widget {
  Widget decorated({required BoxDecoration decoration}) {
    return DecoratedBox(
      decoration: decoration,
      child: this,
    );
  }
}
