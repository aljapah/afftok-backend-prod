import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'dart:math' as math;
import '../providers/auth_provider.dart';
import '../providers/offer_provider.dart';
import '../services/ai_service.dart';
import 'ai_settings_screen.dart';

class AIAssistantScreen extends StatefulWidget {
  const AIAssistantScreen({super.key});

  @override
  State<AIAssistantScreen> createState() => _AIAssistantScreenState();
}

class _AIAssistantScreenState extends State<AIAssistantScreen>
    with TickerProviderStateMixin {
  late AnimationController _pulseController;
  late AnimationController _rotateController;
  late Animation<double> _pulseAnimation;
  
  final AIService _aiService = AIService();
  final TextEditingController _chatController = TextEditingController();
  final List<Map<String, String>> _chatHistory = [];
  bool _isChatLoading = false;
  
  // AffTok Colors
  static const Color primaryRed = Color(0xFFE53935);
  static const Color primaryPink = Color(0xFFFF006E);
  static const Color accentOrange = Color(0xFFFF7043);
  static const Color successGreen = Color(0xFF4CAF50);
  static const Color infoBlue = Color(0xFF2196F3);
  static const Color warningYellow = Color(0xFFFFB300);

  @override
  void initState() {
    super.initState();
    _initAI();
    
    _pulseController = AnimationController(
      duration: const Duration(seconds: 2),
      vsync: this,
    )..repeat(reverse: true);
    
    _pulseAnimation = Tween<double>(begin: 1.0, end: 1.1).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
    
    _rotateController = AnimationController(
      duration: const Duration(seconds: 10),
      vsync: this,
    )..repeat();
  }
  
  Future<void> _initAI() async {
    await _aiService.init();
    setState(() {});
  }

  @override
  void dispose() {
    _pulseController.dispose();
    _rotateController.dispose();
    _chatController.dispose();
    super.dispose();
  }

  bool get _isArabic => Localizations.localeOf(context).languageCode == 'ar';
  
  // ============ AI CHAT ============
  
  void _openChat() {
    if (!_aiService.hasApiKey) {
      _showNoApiKeyDialog();
      return;
    }
    _showChatSheet();
  }
  
  void _showNoApiKeyDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xFF1A1A2E),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Row(
          children: [
            const Icon(Icons.key, color: primaryRed),
            const SizedBox(width: 12),
            Text(
              _isArabic ? 'مفتاح AI مطلوب' : 'AI Key Required',
              style: const TextStyle(color: Colors.white),
            ),
          ],
        ),
        content: Text(
          _isArabic 
              ? 'لاستخدام المحادثة الذكية، تحتاج إضافة مفتاح API الخاص بك.\n\nهذه الميزة مجانية تقريباً مع Groq!'
              : 'To use smart chat, you need to add your own API key.\n\nThis feature is almost free with Groq!',
          style: TextStyle(color: Colors.grey[400]),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(_isArabic ? 'لاحقاً' : 'Later', style: TextStyle(color: Colors.grey[500])),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              Navigator.push(context, MaterialPageRoute(builder: (_) => const AISettingsScreen()));
            },
            style: ElevatedButton.styleFrom(backgroundColor: primaryRed),
            child: Text(_isArabic ? 'إضافة مفتاح' : 'Add Key'),
          ),
        ],
      ),
    );
  }
  
  void _clearChat(StateSetter setModalState) {
    setModalState(() {
      _chatHistory.clear();
    });
    setState(() {});
  }
  
  void _showChatSheet() {
    final ScrollController chatScrollController = ScrollController();
    
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (modalContext) => StatefulBuilder(
        builder: (context, setModalState) => Padding(
          padding: EdgeInsets.only(
            bottom: MediaQuery.of(context).viewInsets.bottom,
          ),
          child: Container(
            height: MediaQuery.of(context).size.height * 0.9,
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [Color(0xFF1A1A2E), Color(0xFF0F0F1A)],
              ),
              borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
            ),
            child: Column(
              children: [
                // Handle & Header
                Container(
                  margin: const EdgeInsets.only(top: 12),
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey[600],
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          gradient: const LinearGradient(colors: [primaryRed, primaryPink]),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: const Icon(Icons.chat, color: Colors.white, size: 22),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _isArabic ? 'محادثة ذكية' : 'Smart Chat',
                              style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
                            ),
                            Text(
                              _aiService.providerName,
                              style: TextStyle(color: Colors.grey[500], fontSize: 11),
                            ),
                          ],
                        ),
                      ),
                      // New Chat Button
                      if (_chatHistory.isNotEmpty)
                        IconButton(
                          icon: const Icon(Icons.refresh, color: Colors.white70, size: 22),
                          onPressed: () {
                            showDialog(
                              context: context,
                              builder: (ctx) => AlertDialog(
                                backgroundColor: const Color(0xFF1A1A2E),
                                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                                title: Text(
                                  _isArabic ? 'محادثة جديدة؟' : 'New Chat?',
                                  style: const TextStyle(color: Colors.white),
                                ),
                                content: Text(
                                  _isArabic ? 'سيتم مسح المحادثة الحالية' : 'Current chat will be cleared',
                                  style: TextStyle(color: Colors.grey[400]),
                                ),
                                actions: [
                                  TextButton(
                                    onPressed: () => Navigator.pop(ctx),
                                    child: Text(_isArabic ? 'إلغاء' : 'Cancel', style: TextStyle(color: Colors.grey[500])),
                                  ),
                                  ElevatedButton(
                                    onPressed: () {
                                      Navigator.pop(ctx);
                                      _clearChat(setModalState);
                                    },
                                    style: ElevatedButton.styleFrom(backgroundColor: primaryRed),
                                    child: Text(_isArabic ? 'مسح' : 'Clear'),
                                  ),
                                ],
                              ),
                            );
                          },
                          tooltip: _isArabic ? 'محادثة جديدة' : 'New Chat',
                        ),
                      IconButton(
                        icon: const Icon(Icons.close, color: Colors.grey, size: 22),
                        onPressed: () => Navigator.pop(context),
                      ),
                    ],
                  ),
                ),
                Divider(color: Colors.grey[800], height: 1),
                
                // Chat Messages
                Expanded(
                  child: _chatHistory.isEmpty
                      ? Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.chat_bubble_outline, color: Colors.grey[700], size: 56),
                              const SizedBox(height: 16),
                              Text(
                                _isArabic ? 'اسألني أي شيء عن التسويق!' : 'Ask me anything about marketing!',
                                style: TextStyle(color: Colors.grey[600], fontSize: 15),
                              ),
                              const SizedBox(height: 8),
                              Text(
                                _isArabic ? 'مثال: كيف أزيد تحويلاتي؟' : 'Example: How to increase conversions?',
                                style: TextStyle(color: Colors.grey[700], fontSize: 13),
                              ),
                            ],
                          ),
                        )
                      : ListView.builder(
                          controller: chatScrollController,
                          padding: const EdgeInsets.all(16),
                          itemCount: _chatHistory.length,
                          itemBuilder: (context, index) {
                            final message = _chatHistory[index];
                            final isUser = message['role'] == 'user';
                            return _buildChatBubble(message['content']!, isUser);
                          },
                        ),
                ),
                
                // Loading Indicator
                if (_isChatLoading)
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                    child: Row(
                      children: [
                        const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(color: primaryRed, strokeWidth: 2),
                        ),
                        const SizedBox(width: 10),
                        Text(
                          _isArabic ? 'جاري التفكير...' : 'Thinking...',
                          style: TextStyle(color: Colors.grey[500], fontSize: 13),
                        ),
                      ],
                    ),
                  ),
                
                // Input Field - Safe Area
                SafeArea(
                  top: false,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    decoration: BoxDecoration(
                      color: const Color(0xFF0F0F1A),
                      border: Border(top: BorderSide(color: Colors.grey[850]!)),
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: Container(
                            decoration: BoxDecoration(
                              color: Colors.white.withOpacity(0.08),
                              borderRadius: BorderRadius.circular(24),
                            ),
                            child: TextField(
                              controller: _chatController,
                              style: const TextStyle(color: Colors.white, fontSize: 15),
                              maxLines: 4,
                              minLines: 1,
                              textInputAction: TextInputAction.send,
                              decoration: InputDecoration(
                                hintText: _isArabic ? 'اكتب سؤالك...' : 'Type your question...',
                                hintStyle: TextStyle(color: Colors.grey[600], fontSize: 14),
                                border: InputBorder.none,
                                contentPadding: const EdgeInsets.symmetric(horizontal: 18, vertical: 10),
                              ),
                              onSubmitted: (_) {
                                _sendMessage(setModalState);
                                // Scroll to bottom after sending
                                Future.delayed(const Duration(milliseconds: 100), () {
                                  if (chatScrollController.hasClients) {
                                    chatScrollController.animateTo(
                                      chatScrollController.position.maxScrollExtent,
                                      duration: const Duration(milliseconds: 300),
                                      curve: Curves.easeOut,
                                    );
                                  }
                                });
                              },
                            ),
                          ),
                        ),
                        const SizedBox(width: 10),
                        Container(
                          decoration: const BoxDecoration(
                            gradient: LinearGradient(colors: [primaryRed, primaryPink]),
                            shape: BoxShape.circle,
                          ),
                          child: IconButton(
                            icon: const Icon(Icons.send, color: Colors.white, size: 20),
                            onPressed: () {
                              _sendMessage(setModalState);
                              // Scroll to bottom after sending
                              Future.delayed(const Duration(milliseconds: 100), () {
                                if (chatScrollController.hasClients) {
                                  chatScrollController.animateTo(
                                    chatScrollController.position.maxScrollExtent,
                                    duration: const Duration(milliseconds: 300),
                                    curve: Curves.easeOut,
                                  );
                                }
                              });
                            },
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
  
  Widget _buildChatBubble(String message, bool isUser) {
    return Align(
      alignment: isUser ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(14),
        constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.8),
        decoration: BoxDecoration(
          gradient: isUser 
              ? const LinearGradient(colors: [primaryRed, primaryPink])
              : null,
          color: isUser ? null : Colors.white.withOpacity(0.1),
          borderRadius: BorderRadius.only(
            topLeft: const Radius.circular(16),
            topRight: const Radius.circular(16),
            bottomLeft: Radius.circular(isUser ? 16 : 4),
            bottomRight: Radius.circular(isUser ? 4 : 16),
          ),
        ),
        child: Text(
          message,
          style: TextStyle(
            color: isUser ? Colors.white : Colors.grey[300],
            fontSize: 15,
            height: 1.4,
          ),
        ),
      ),
    );
  }
  
  // Maximum messages to send to AI (to reduce cost)
  static const int _maxHistoryToSend = 10;
  
  Future<void> _sendMessage(StateSetter setModalState) async {
    final message = _chatController.text.trim();
    if (message.isEmpty) return;
    
    _chatController.clear();
    
    setModalState(() {
      _chatHistory.add({'role': 'user', 'content': message});
      _isChatLoading = true;
    });
    
    // Also update parent state
    setState(() {});
    
    try {
      // Only send last N messages to reduce cost
      final historyToSend = _chatHistory.length > _maxHistoryToSend
          ? _chatHistory.sublist(_chatHistory.length - _maxHistoryToSend)
          : _chatHistory;
      
      final response = await _aiService.chat(message, conversationHistory: historyToSend);
      
      if (response != null) {
        setModalState(() {
          _chatHistory.add({'role': 'assistant', 'content': response});
        });
      } else {
        setModalState(() {
          _chatHistory.add({
            'role': 'assistant',
            'content': _isArabic ? 'عذراً، حدث خطأ. حاول مرة أخرى.' : 'Sorry, an error occurred. Please try again.',
          });
        });
      }
    } catch (e) {
      setModalState(() {
        _chatHistory.add({
          'role': 'assistant',
          'content': _isArabic ? 'عذراً، حدث خطأ. حاول مرة أخرى.' : 'Sorry, an error occurred. Please try again.',
        });
      });
    } finally {
      setModalState(() {
        _isChatLoading = false;
      });
      setState(() {});
    }
  }

  // ============ ENHANCED STATS DISPLAY ============
  
  void _showMyStats() {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final user = authProvider.user;
    
    if (user == null) return;
    
    final stats = user.stats;
    final totalClicks = stats.totalClicks;
    final totalConversions = stats.totalConversions;
    final conversionRate = totalClicks > 0 
        ? (totalConversions / totalClicks) * 100
        : 0.0;
    final globalRank = stats.globalRank;
    final offersCount = authProvider.userOffers.length;
    
    _showEnhancedStatsSheet(
      totalClicks: totalClicks,
      totalConversions: totalConversions,
      conversionRate: conversionRate,
      globalRank: globalRank,
      offersCount: offersCount,
    );
  }
  
  void _showEnhancedStatsSheet({
    required int totalClicks,
    required int totalConversions,
    required double conversionRate,
    required int globalRank,
    required int offersCount,
  }) {
    // Calculate max for progress bars
    final maxClicks = math.max(totalClicks, 100);
    final maxConversions = math.max(totalConversions, 10);
    
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
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [Color(0xFF1A1A2E), Color(0xFF0F0F1A)],
            ),
            borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
          ),
          child: Column(
            children: [
              // Handle
              Container(
                margin: const EdgeInsets.only(top: 12),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey[600],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              
              // Header
              Padding(
                padding: const EdgeInsets.all(24),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: [primaryRed, primaryPink],
                        ),
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: const Icon(Icons.analytics, color: Colors.white, size: 28),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            _isArabic ? 'إحصائياتك' : 'Your Stats',
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 24,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          Text(
                            _isArabic ? 'نظرة شاملة على أدائك' : 'Overview of your performance',
                            style: TextStyle(
                              color: Colors.grey[400],
                              fontSize: 14,
                            ),
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.grey),
                      onPressed: () => Navigator.pop(context),
                    ),
                  ],
                ),
              ),
              
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  child: Column(
                    children: [
                      // Stats Cards Grid
                      Row(
                        children: [
                          Expanded(
                            child: _buildStatCard(
                              icon: Icons.touch_app,
                              iconColor: accentOrange,
                              label: _isArabic ? 'النقرات' : 'Clicks',
                              value: totalClicks.toString(),
                              progress: totalClicks / maxClicks,
                              progressColor: accentOrange,
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: _buildStatCard(
                              icon: Icons.check_circle,
                              iconColor: successGreen,
                              label: _isArabic ? 'التحويلات' : 'Conversions',
                              value: totalConversions.toString(),
                              progress: totalConversions / maxConversions,
                              progressColor: successGreen,
                            ),
                          ),
                        ],
                      ),
                      
                      const SizedBox(height: 12),
                      
                      Row(
                        children: [
                          Expanded(
                            child: _buildStatCard(
                              icon: Icons.trending_up,
                              iconColor: infoBlue,
                              label: _isArabic ? 'معدل التحويل' : 'Conv. Rate',
                              value: '${conversionRate.toStringAsFixed(1)}%',
                              progress: conversionRate / 100,
                              progressColor: infoBlue,
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: _buildStatCard(
                              icon: Icons.emoji_events,
                              iconColor: warningYellow,
                              label: _isArabic ? 'ترتيبك' : 'Your Rank',
                              value: '#$globalRank',
                              isRank: true,
                              rankEmoji: globalRank == 1 ? '🥇' : globalRank <= 3 ? '🥈' : '🏆',
                            ),
                          ),
                        ],
                      ),
                      
                      const SizedBox(height: 12),
                      
                      // Offers Count
                      _buildWideStatCard(
                        icon: Icons.local_offer,
                        iconColor: primaryPink,
                        label: _isArabic ? 'العروض المُفعّلة' : 'Active Offers',
                        value: offersCount.toString(),
                        subtitle: _isArabic 
                            ? 'عرض${offersCount > 1 ? "اً" : ""} نشط${offersCount > 1 ? "اً" : ""}'
                            : '$offersCount active offer${offersCount > 1 ? "s" : ""}',
                      ),
                      
                      const SizedBox(height: 24),
                      
                      // Smart Advice Section
                      _buildAdviceSection(totalClicks, totalConversions, conversionRate),
                      
                      const SizedBox(height: 24),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
  
  Widget _buildStatCard({
    required IconData icon,
    required Color iconColor,
    required String label,
    required String value,
    double? progress,
    Color? progressColor,
    bool isRank = false,
    String? rankEmoji,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: iconColor.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: iconColor.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(icon, color: iconColor, size: 20),
              ),
              const Spacer(),
              if (isRank && rankEmoji != null)
                Text(rankEmoji, style: const TextStyle(fontSize: 24)),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            value,
            style: TextStyle(
              color: Colors.white,
              fontSize: 28,
              fontWeight: FontWeight.bold,
              shadows: [
                Shadow(
                  color: iconColor.withOpacity(0.5),
                  blurRadius: 10,
                ),
              ],
            ),
          ),
          const SizedBox(height: 4),
          Text(
            label,
            style: TextStyle(
              color: Colors.grey[400],
              fontSize: 13,
            ),
          ),
          if (progress != null) ...[
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: progress.clamp(0.0, 1.0),
                backgroundColor: Colors.grey[800],
                valueColor: AlwaysStoppedAnimation<Color>(progressColor ?? iconColor),
                minHeight: 6,
              ),
            ),
          ],
        ],
      ),
    );
  }
  
  Widget _buildWideStatCard({
    required IconData icon,
    required Color iconColor,
    required String label,
    required String value,
    required String subtitle,
  }) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            iconColor.withOpacity(0.15),
            iconColor.withOpacity(0.05),
          ],
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: iconColor.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: iconColor.withOpacity(0.2),
              borderRadius: BorderRadius.circular(14),
            ),
            child: Icon(icon, color: iconColor, size: 28),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: TextStyle(
                    color: Colors.grey[400],
                    fontSize: 14,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  subtitle,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 16,
                  ),
                ),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            decoration: BoxDecoration(
              color: iconColor,
              borderRadius: BorderRadius.circular(20),
            ),
            child: Text(
              value,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
    );
  }
  
  Widget _buildAdviceSection(int clicks, int conversions, double rate) {
    String title;
    String advice;
    IconData icon;
    Color color;
    
    if (clicks == 0) {
      title = _isArabic ? '🚀 ابدأ الآن!' : '🚀 Start Now!';
      advice = _isArabic 
          ? 'شارك روابطك على منصات التواصل الاجتماعي للحصول على أول نقراتك. كلما زادت المشاركة، زادت فرص النجاح!'
          : 'Share your links on social media to get your first clicks. The more you share, the higher your chances of success!';
      icon = Icons.rocket_launch;
      color = accentOrange;
    } else if (rate < 1) {
      title = _isArabic ? '💡 نصيحة للتحسين' : '💡 Improvement Tip';
      advice = _isArabic
          ? 'جرّب استهداف جمهور أكثر اهتماماً بالعروض. اختر المنصات التي يتواجد فيها جمهورك المستهدف.'
          : 'Try targeting an audience more interested in offers. Choose platforms where your target audience is present.';
      icon = Icons.lightbulb;
      color = warningYellow;
    } else if (rate < 5) {
      title = _isArabic ? '👍 أداء جيد!' : '👍 Good Performance!';
      advice = _isArabic
          ? 'أنت على الطريق الصحيح! استمر في زيادة النقرات وجرّب أنواعاً مختلفة من المحتوى.'
          : 'You\'re on the right track! Keep increasing clicks and try different types of content.';
      icon = Icons.thumb_up;
      color = infoBlue;
    } else {
      title = _isArabic ? '🌟 أداء ممتاز!' : '🌟 Excellent!';
      advice = _isArabic
          ? 'تهانينا! أنت من أفضل المروّجين. استمر بنفس الاستراتيجية وفكّر في زيادة حجم النشر.'
          : 'Congratulations! You\'re among the top promoters. Keep the same strategy and consider increasing your posting volume.';
      icon = Icons.star;
      color = successGreen;
    }
    
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            color.withOpacity(0.2),
            color.withOpacity(0.05),
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: color.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(icon, color: color, size: 24),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  title,
                  style: TextStyle(
                    color: color,
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Text(
            advice,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 15,
              height: 1.5,
            ),
          ),
        ],
      ),
    );
  }

  // ============ SUGGEST OFFERS ============
  
  void _suggestOffers() {
    final offerProvider = Provider.of<OfferProvider>(context, listen: false);
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    
    final allOffers = offerProvider.offers;
    final userOfferIds = authProvider.userOffers.map((uo) => uo.offerId).toSet();
    
    final availableOffers = allOffers.where((o) => !userOfferIds.contains(o.id)).toList();
    final topOffers = availableOffers.take(3).toList();
    
    _showSuggestionsSheet(topOffers, availableOffers.isEmpty);
  }
  
  void _showSuggestionsSheet(List offers, bool allJoined) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.7,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        builder: (context, scrollController) => Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [Color(0xFF1A1A2E), Color(0xFF0F0F1A)],
            ),
            borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
          ),
          child: Column(
            children: [
              Container(
                margin: const EdgeInsets.only(top: 12),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey[600],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              
              Padding(
                padding: const EdgeInsets.all(24),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: [accentOrange, warningYellow],
                        ),
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: const Icon(Icons.lightbulb, color: Colors.white, size: 28),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Text(
                        _isArabic ? 'العروض المقترحة' : 'Suggested Offers',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.grey),
                      onPressed: () => Navigator.pop(context),
                    ),
                  ],
                ),
              ),
              
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  child: allJoined
                      ? _buildAllJoinedMessage()
                      : Column(
                          children: [
                            ...offers.asMap().entries.map((entry) {
                              final index = entry.key;
                              final offer = entry.value;
                              return _buildOfferSuggestionCard(offer, index + 1);
                            }),
                            const SizedBox(height: 24),
                          ],
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
  
  Widget _buildAllJoinedMessage() {
    return Container(
      padding: const EdgeInsets.all(32),
      child: Column(
        children: [
          const Text('🎉', style: TextStyle(fontSize: 64)),
          const SizedBox(height: 16),
          Text(
            _isArabic ? 'أنت مشترك في جميع العروض!' : 'You\'ve joined all offers!',
            style: const TextStyle(
              color: Colors.white,
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 12),
          Text(
            _isArabic 
                ? 'استمر في الترويج للعروض الحالية لزيادة نقراتك وتحويلاتك'
                : 'Keep promoting your current offers to increase clicks and conversions',
            style: TextStyle(
              color: Colors.grey[400],
              fontSize: 16,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
  
  Widget _buildOfferSuggestionCard(dynamic offer, int rank) {
    final colors = [primaryRed, accentOrange, primaryPink];
    final color = colors[(rank - 1) % colors.length];
    
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: color,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Center(
              child: Text(
                '#$rank',
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                ),
              ),
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  offer.companyName ?? 'Offer',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const SizedBox(height: 4),
                Row(
                  children: [
                    Icon(Icons.category, size: 14, color: Colors.grey[500]),
                    const SizedBox(width: 4),
                    Text(
                      offer.category ?? '',
                      style: TextStyle(
                        color: Colors.grey[500],
                        fontSize: 13,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            decoration: BoxDecoration(
              color: successGreen.withOpacity(0.2),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Text(
              offer.reward ?? '',
              style: const TextStyle(
                color: successGreen,
                fontWeight: FontWeight.bold,
                fontSize: 14,
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ============ ANALYZE PERFORMANCE ============
  
  void _analyzePerformance() {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final user = authProvider.user;
    if (user == null) return;
    
    final stats = user.stats;
    final userOffers = authProvider.userOffers;
    
    _showAnalysisSheet(stats, userOffers);
  }
  
  void _showAnalysisSheet(dynamic stats, List userOffers) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.8,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        builder: (context, scrollController) => Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [Color(0xFF1A1A2E), Color(0xFF0F0F1A)],
            ),
            borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
          ),
          child: Column(
            children: [
              Container(
                margin: const EdgeInsets.only(top: 12),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey[600],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              
              Padding(
                padding: const EdgeInsets.all(24),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: [primaryPink, primaryRed],
                        ),
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: const Icon(Icons.analytics, color: Colors.white, size: 28),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Text(
                        _isArabic ? 'تحليل الأداء' : 'Performance Analysis',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.grey),
                      onPressed: () => Navigator.pop(context),
                    ),
                  ],
                ),
              ),
              
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Offers Analysis
                      Text(
                        _isArabic ? '📦 تحليل العروض' : '📦 Offers Analysis',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 16),
                      
                      ...userOffers.map((uo) => _buildOfferAnalysisCard(uo)),
                      
                      const SizedBox(height: 24),
                      
                      // Action Plan
                      _buildActionPlan(stats.totalClicks, stats.totalConversions),
                      
                      const SizedBox(height: 24),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
  
  Widget _buildOfferAnalysisCard(dynamic uo) {
    final clicks = uo.totalClicks;
    final conversions = uo.totalConversions;
    
    Color statusColor;
    String statusText;
    IconData statusIcon;
    
    if (clicks == 0) {
      statusColor = Colors.grey;
      statusText = _isArabic ? 'لم يبدأ' : 'Not started';
      statusIcon = Icons.hourglass_empty;
    } else if (conversions > 0) {
      statusColor = successGreen;
      statusText = _isArabic ? 'ممتاز' : 'Excellent';
      statusIcon = Icons.check_circle;
    } else {
      statusColor = warningYellow;
      statusText = _isArabic ? 'يحتاج تحسين' : 'Needs work';
      statusIcon = Icons.warning_amber;
    }
    
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: statusColor.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  uo.offerTitle,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: statusColor.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(statusIcon, color: statusColor, size: 14),
                    const SizedBox(width: 4),
                    Text(
                      statusText,
                      style: TextStyle(
                        color: statusColor,
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              _buildMiniStat(Icons.touch_app, '$clicks', _isArabic ? 'نقرة' : 'clicks', accentOrange),
              const SizedBox(width: 24),
              _buildMiniStat(Icons.check_circle, '$conversions', _isArabic ? 'تحويل' : 'conv.', successGreen),
            ],
          ),
        ],
      ),
    );
  }
  
  Widget _buildMiniStat(IconData icon, String value, String label, Color color) {
    return Row(
      children: [
        Icon(icon, color: color, size: 16),
        const SizedBox(width: 6),
        Text(
          value,
          style: TextStyle(
            color: color,
            fontSize: 16,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(width: 4),
        Text(
          label,
          style: TextStyle(
            color: Colors.grey[500],
            fontSize: 13,
          ),
        ),
      ],
    );
  }
  
  Widget _buildActionPlan(int clicks, int conversions) {
    final rate = clicks > 0 ? (conversions / clicks) * 100 : 0.0;
    
    List<Map<String, dynamic>> actions;
    
    if (clicks == 0) {
      actions = [
        {'icon': Icons.share, 'text': _isArabic ? 'شارك روابطك على منصات التواصل' : 'Share links on social media', 'color': primaryRed},
        {'icon': Icons.people, 'text': _isArabic ? 'استهدف الجمهور المهتم بالعروض' : 'Target audience interested in offers', 'color': accentOrange},
        {'icon': Icons.edit, 'text': _isArabic ? 'استخدم محتوى جذاب مع الروابط' : 'Use engaging content with links', 'color': primaryPink},
      ];
    } else if (rate < 2) {
      actions = [
        {'icon': Icons.swap_horiz, 'text': _isArabic ? 'جرّب عروضاً مختلفة' : 'Try different offers', 'color': primaryRed},
        {'icon': Icons.gps_fixed, 'text': _isArabic ? 'استهدف جمهوراً أكثر تخصصاً' : 'Target more specific audience', 'color': accentOrange},
        {'icon': Icons.brush, 'text': _isArabic ? 'حسّن طريقة عرض الروابط' : 'Improve how you present links', 'color': primaryPink},
      ];
    } else {
      actions = [
        {'icon': Icons.repeat, 'text': _isArabic ? 'استمر بنفس الاستراتيجية' : 'Continue same strategy', 'color': successGreen},
        {'icon': Icons.trending_up, 'text': _isArabic ? 'زِد حجم النشر' : 'Increase posting volume', 'color': accentOrange},
        {'icon': Icons.new_releases, 'text': _isArabic ? 'جرّب عروضاً جديدة' : 'Try new offers', 'color': primaryPink},
      ];
    }
    
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: primaryRed.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.lightbulb, color: warningYellow, size: 24),
              const SizedBox(width: 8),
              Text(
                _isArabic ? 'خطة العمل المقترحة' : 'Suggested Action Plan',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          const SizedBox(height: 20),
          ...actions.asMap().entries.map((entry) {
            final index = entry.key + 1;
            final action = entry.value;
            return Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Row(
                children: [
                  Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      color: (action['color'] as Color).withOpacity(0.2),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Center(
                      child: Text(
                        '$index',
                        style: TextStyle(
                          color: action['color'] as Color,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Icon(action['icon'] as IconData, color: action['color'] as Color, size: 20),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      action['text'] as String,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 15,
                      ),
                    ),
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  // ============ TIPS ============
  
  void _showTips() {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.85,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        builder: (context, scrollController) => Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [Color(0xFF1A1A2E), Color(0xFF0F0F1A)],
            ),
            borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
          ),
          child: Column(
            children: [
              Container(
                margin: const EdgeInsets.only(top: 12),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey[600],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              
              Padding(
                padding: const EdgeInsets.all(24),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: [successGreen, infoBlue],
                        ),
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: const Icon(Icons.tips_and_updates, color: Colors.white, size: 28),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Text(
                        _isArabic ? 'نصائح للنجاح' : 'Tips for Success',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.grey),
                      onPressed: () => Navigator.pop(context),
                    ),
                  ],
                ),
              ),
              
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  child: Column(
                    children: [
                      _buildTipCard(
                        number: '1',
                        emoji: '⏰',
                        title: _isArabic ? 'اختر الوقت المناسب' : 'Choose the Right Time',
                        tips: _isArabic 
                            ? ['أفضل أوقات النشر: المساء والعطلات', 'تجنب النشر في ساعات العمل']
                            : ['Best times: Evenings & weekends', 'Avoid posting during work hours'],
                        color: primaryRed,
                      ),
                      _buildTipCard(
                        number: '2',
                        emoji: '🎯',
                        title: _isArabic ? 'استهدف الجمهور الصحيح' : 'Target the Right Audience',
                        tips: _isArabic
                            ? ['شارك عروض التمويل مع المهتمين', 'شارك عروض التسوق مع المتسوقين']
                            : ['Share finance offers with interested users', 'Share shopping offers with shoppers'],
                        color: accentOrange,
                      ),
                      _buildTipCard(
                        number: '3',
                        emoji: '✍️',
                        title: _isArabic ? 'اكتب محتوى جذاب' : 'Write Engaging Content',
                        tips: _isArabic
                            ? ['استخدم عبارات مثيرة للاهتمام', 'أضف تجربتك الشخصية']
                            : ['Use attention-grabbing phrases', 'Add your personal experience'],
                        color: primaryPink,
                      ),
                      _buildTipCard(
                        number: '4',
                        emoji: '📱',
                        title: _isArabic ? 'نوّع منصات النشر' : 'Diversify Platforms',
                        tips: _isArabic
                            ? ['تويتر، انستقرام، تيك توك', 'المجموعات والمنتديات المتخصصة']
                            : ['Twitter, Instagram, TikTok', 'Specialized groups & forums'],
                        color: infoBlue,
                      ),
                      _buildTipCard(
                        number: '5',
                        emoji: '📊',
                        title: _isArabic ? 'تابع وحلل النتائج' : 'Track & Analyze Results',
                        tips: _isArabic
                            ? ['راقب أي العروض تحقق نتائج', 'ركّز على ما ينجح']
                            : ['Monitor which offers perform best', 'Focus on what works'],
                        color: successGreen,
                      ),
                      
                      // Final Tip
                      Container(
                        margin: const EdgeInsets.only(bottom: 24, top: 8),
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          gradient: LinearGradient(
                            colors: [
                              warningYellow.withOpacity(0.2),
                              warningYellow.withOpacity(0.05),
                            ],
                          ),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(color: warningYellow.withOpacity(0.4)),
                        ),
                        child: Row(
                          children: [
                            const Text('💡', style: TextStyle(fontSize: 32)),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Text(
                                _isArabic 
                                    ? 'تذكر: الاستمرارية هي مفتاح النجاح!'
                                    : 'Remember: Consistency is key to success!',
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 16,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
  
  Widget _buildTipCard({
    required String number,
    required String emoji,
    required String title,
    required List<String> tips,
    required Color color,
  }) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 36,
                height: 36,
                decoration: BoxDecoration(
                  color: color,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Center(
                  child: Text(
                    number,
                    style: const TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                      fontSize: 18,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Text(emoji, style: const TextStyle(fontSize: 24)),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  title,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 17,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          ...tips.map((tip) => Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(Icons.check_circle, color: color, size: 18),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    tip,
                    style: TextStyle(
                      color: Colors.grey[300],
                      fontSize: 15,
                      height: 1.4,
                    ),
                  ),
                ),
              ],
            ),
          )),
        ],
      ),
    );
  }

  // ============ MAIN BUILD ============

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF0A0A0A),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        title: Text(
          _isArabic ? 'المساعد الذكي' : 'AI Assistant',
          style: const TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.bold,
          ),
        ),
        centerTitle: true,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings, color: Colors.white),
            onPressed: () async {
              await Navigator.push(
                context, 
                MaterialPageRoute(builder: (_) => const AISettingsScreen()),
              );
              _initAI(); // Refresh after settings
            },
            tooltip: _isArabic ? 'الإعدادات' : 'Settings',
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
        child: Column(
          children: [
            // Animated Space Orb
            SizedBox(
              height: 160,
              child: AnimatedBuilder(
                animation: Listenable.merge([_pulseController, _rotateController]),
                builder: (context, child) {
                  return Transform.scale(
                    scale: _pulseAnimation.value,
                    child: Stack(
                      alignment: Alignment.center,
                      children: [
                        ...List.generate(2, (index) {
                          return Transform.rotate(
                            angle: _rotateController.value * 2 * math.pi + (index * math.pi / 2),
                            child: Container(
                              width: 100 + (index * 20),
                              height: 100 + (index * 20),
                              decoration: BoxDecoration(
                                shape: BoxShape.circle,
                                border: Border.all(
                                  color: [primaryRed, primaryPink][index].withOpacity(0.3),
                                  width: 2,
                                ),
                              ),
                            ),
                          );
                        }),
                        Container(
                          width: 80,
                          height: 80,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            gradient: const RadialGradient(
                              colors: [
                                Color(0xFFFFFFFF),
                                Color(0xFFFF8A80),
                                Color(0xFFE53935),
                                Color(0xFFFF006E),
                                Color(0xFF1A1A2E),
                              ],
                              stops: [0.0, 0.2, 0.4, 0.7, 1.0],
                              center: Alignment(-0.3, -0.3),
                            ),
                            boxShadow: [
                              BoxShadow(
                                color: primaryRed.withOpacity(0.5),
                                blurRadius: 25,
                                spreadRadius: 5,
                              ),
                            ],
                          ),
                          child: const Center(
                            child: Icon(Icons.auto_awesome, size: 32, color: Colors.white),
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
            
            ShaderMask(
              shaderCallback: (bounds) => const LinearGradient(
                colors: [primaryRed, primaryPink],
              ).createShader(bounds),
              child: Text(
                _isArabic ? 'مساعدك الذكي' : 'Your AI Assistant',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
            
            const SizedBox(height: 6),
            
            Text(
              _isArabic ? 'كيف يمكنني مساعدتك؟' : 'How can I help you?',
              style: TextStyle(color: Colors.grey[400], fontSize: 15),
            ),
            
            const SizedBox(height: 28),
            
            // AI Chat Button (Premium Feature)
            _buildChatButton(),
            
            const SizedBox(height: 14),
            
            _buildActionButton(
              icon: Icons.analytics_outlined,
              title: _isArabic ? 'عرض نقراتي وتحويلاتي' : 'View My Clicks & Conversions',
              subtitle: _isArabic ? 'إحصائيات مفصلة عن أدائك' : 'Detailed stats about your performance',
              color: primaryRed,
              onTap: _showMyStats,
            ),
            
            const SizedBox(height: 14),
            
            _buildActionButton(
              icon: Icons.lightbulb_outline,
              title: _isArabic ? 'اقتراح أفضل العروض لي' : 'Suggest Best Offers for Me',
              subtitle: _isArabic ? 'عروض مناسبة لك بناءً على التحليل' : 'Offers suitable based on analysis',
              color: accentOrange,
              onTap: _suggestOffers,
            ),
            
            const SizedBox(height: 14),
            
            _buildActionButton(
              icon: Icons.trending_up,
              title: _isArabic ? 'تحليل وتحسين أدائي' : 'Analyze & Improve My Performance',
              subtitle: _isArabic ? 'تحليل شامل مع توصيات' : 'Comprehensive analysis with recommendations',
              color: primaryPink,
              onTap: _analyzePerformance,
            ),
            
            const SizedBox(height: 14),
            
            _buildActionButton(
              icon: Icons.tips_and_updates_outlined,
              title: _isArabic ? 'نصائح لزيادة تحويلاتي' : 'Tips to Increase My Conversions',
              subtitle: _isArabic ? 'استراتيجيات فعّالة للنجاح' : 'Effective strategies for success',
              color: successGreen,
              onTap: _showTips,
            ),
            
            const SizedBox(height: 28),
            
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.05),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.white.withOpacity(0.1)),
              ),
              child: Row(
                children: [
                  Icon(Icons.auto_awesome, color: primaryPink, size: 18),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      _isArabic 
                          ? 'المساعد يستخدم بياناتك الفعلية لتقديم نصائح مخصصة'
                          : 'Uses your real data for personalized insights',
                      style: TextStyle(color: Colors.grey[400], fontSize: 12),
                    ),
                  ),
                ],
              ),
            ),
            
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }
  
  Widget _buildActionButton({
    required IconData icon,
    required String title,
    required String subtitle,
    required Color color,
    required VoidCallback onTap,
  }) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(16),
        child: Container(
          padding: const EdgeInsets.all(18),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [color.withOpacity(0.15), color.withOpacity(0.05)],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: color.withOpacity(0.3)),
          ),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: color.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(icon, color: color, size: 26),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 15,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    const SizedBox(height: 3),
                    Text(
                      subtitle,
                      style: TextStyle(color: Colors.grey[500], fontSize: 12),
                    ),
                  ],
                ),
              ),
              Icon(Icons.arrow_forward_ios, color: color, size: 16),
            ],
          ),
        ),
      ),
    );
  }
  
  Widget _buildChatButton() {
    final hasKey = _aiService.hasApiKey;
    
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: _openChat,
        borderRadius: BorderRadius.circular(20),
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: hasKey 
                  ? [primaryRed, primaryPink]
                  : [Colors.grey.withOpacity(0.3), Colors.grey.withOpacity(0.1)],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(20),
            boxShadow: hasKey ? [
              BoxShadow(
                color: primaryRed.withOpacity(0.3),
                blurRadius: 15,
                offset: const Offset(0, 5),
              ),
            ] : null,
          ),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(14),
                ),
                child: const Icon(Icons.chat, color: Colors.white, size: 28),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Text(
                          _isArabic ? 'محادثة ذكية مع AI' : 'Smart Chat with AI',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 17,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        if (!hasKey) ...[
                          const SizedBox(width: 8),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                            decoration: BoxDecoration(
                              color: Colors.white.withOpacity(0.2),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Text(
                              _isArabic ? 'BYOK' : 'BYOK',
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
                      hasKey
                          ? (_isArabic ? 'اسألني أي سؤال عن التسويق!' : 'Ask me any marketing question!')
                          : (_isArabic ? 'أضف مفتاح API لتفعيل هذه الميزة' : 'Add API key to unlock this feature'),
                      style: TextStyle(
                        color: Colors.white.withOpacity(hasKey ? 0.9 : 0.6),
                        fontSize: 13,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.2),
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  hasKey ? Icons.arrow_forward : Icons.lock_outline,
                  color: Colors.white,
                  size: 18,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
