import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'dart:math' as math;
import '../providers/auth_provider.dart';
import '../providers/offer_provider.dart';
import '../utils/app_localizations.dart';
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
  
  final _aiService = AIService();
  String? _currentResponse;
  String? _currentTitle;
  bool _isLoadingAI = false;
  
  // Chat
  final _chatController = TextEditingController();
  final _scrollController = ScrollController();
  final List<Map<String, String>> _chatHistory = [];
  bool _showChat = false;
  
  // AffTok Colors
  static const Color primaryRed = Color(0xFFE53935);
  static const Color primaryPink = Color(0xFFFF006E);
  static const Color accentOrange = Color(0xFFFF7043);

  @override
  void initState() {
    super.initState();
    _aiService.init();
    
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

  @override
  void dispose() {
    _pulseController.dispose();
    _rotateController.dispose();
    _chatController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  bool get _isArabic => Localizations.localeOf(context).languageCode == 'ar';

  // ============ GUIDED ASSISTANT HANDLERS ============
  
  void _showMyStats() {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final user = authProvider.user;
    
    if (user == null) return;
    
    final stats = user.stats;
    final totalClicks = stats.totalClicks;
    final totalConversions = stats.totalConversions;
    final conversionRate = totalClicks > 0 ? ((totalConversions / totalClicks) * 100) : 0.0;
    final globalRank = stats.globalRank;
    final offersCount = authProvider.userOffers.length;
    
    _showStatsSheet(totalClicks, totalConversions, conversionRate, globalRank, offersCount);
  }
  
  void _showStatsSheet(int clicks, int conversions, double rate, int rank, int offersCount) {
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
            color: Color(0xFF1A1A1A),
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: Column(
            children: [
              Container(margin: const EdgeInsets.only(top: 12), width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey[600], borderRadius: BorderRadius.circular(2))),
              Padding(
                padding: const EdgeInsets.all(20),
                child: Row(
                  children: [
                    const Text('📊', style: TextStyle(fontSize: 28)),
                    const SizedBox(width: 12),
                    Expanded(child: Text(_isArabic ? 'إحصائياتك' : 'Your Stats', style: const TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.bold))),
                    IconButton(icon: const Icon(Icons.close, color: Colors.grey), onPressed: () => Navigator.pop(context)),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Stats Grid
                      Row(
                        children: [
                          Expanded(child: _buildStatCard('🔥', _isArabic ? 'النقرات' : 'Clicks', '$clicks', primaryRed)),
                          const SizedBox(width: 12),
                          Expanded(child: _buildStatCard('✅', _isArabic ? 'التحويلات' : 'Conversions', '$conversions', Colors.green)),
                        ],
                      ),
                      const SizedBox(height: 12),
                      Row(
                        children: [
                          Expanded(child: _buildStatCard('📈', _isArabic ? 'المعدل' : 'Rate', '${rate.toStringAsFixed(1)}%', accentOrange)),
                          const SizedBox(width: 12),
                          Expanded(child: _buildStatCard('🏆', _isArabic ? 'الترتيب' : 'Rank', '#$rank', primaryPink)),
                        ],
                      ),
                      const SizedBox(height: 12),
                      _buildStatCard('📦', _isArabic ? 'العروض المُفعّلة' : 'Active Offers', '$offersCount', Colors.purple),
                      
                      const SizedBox(height: 24),
                      
                      // Advice
                      Text(_isArabic ? 'نصيحة' : 'Advice', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
                      const SizedBox(height: 12),
                      _buildAdviceCard(clicks, conversions),
                      
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
  
  Widget _buildAdviceCard(int clicks, int conversions) {
    String advice;
    IconData icon;
    Color color;
    
    if (clicks == 0) {
      advice = _isArabic ? 'ابدأ بمشاركة روابطك لتحصل على نقرات!' : 'Start sharing your links to get clicks!';
      icon = Icons.rocket_launch;
      color = primaryRed;
    } else if (conversions == 0) {
      advice = _isArabic ? 'جرّب عروضاً مختلفة لتحسين معدل التحويل' : 'Try different offers to improve conversion rate';
      icon = Icons.lightbulb;
      color = accentOrange;
    } else {
      advice = _isArabic ? 'أداء ممتاز! استمر على هذا المنوال' : 'Excellent! Keep up the great work';
      icon = Icons.star;
      color = Colors.green;
    }
    
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(color: color.withOpacity(0.2), borderRadius: BorderRadius.circular(10)),
            child: Icon(icon, color: color, size: 24),
          ),
          const SizedBox(width: 14),
          Expanded(child: Text(advice, style: const TextStyle(color: Colors.white, fontSize: 14))),
        ],
      ),
    );
  }
  
  void _suggestOffers() {
    final offerProvider = Provider.of<OfferProvider>(context, listen: false);
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    
    // Refresh offers first
    offerProvider.loadOffers();
    
    final allOffers = offerProvider.offers;
    final userOfferIds = authProvider.userOffers.map((uo) => uo.offerId).toSet();
    
    final availableOffers = allOffers.where((o) => !userOfferIds.contains(o.id)).toList();
    
    _showSuggestionsSheet(availableOffers);
  }
  
  void _showSuggestionsSheet(List availableOffers) {
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
            color: Color(0xFF1A1A1A),
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: Column(
            children: [
              Container(margin: const EdgeInsets.only(top: 12), width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey[600], borderRadius: BorderRadius.circular(2))),
              Padding(
                padding: const EdgeInsets.all(20),
                child: Row(
                  children: [
                    const Text('💡', style: TextStyle(fontSize: 28)),
                    const SizedBox(width: 12),
                    Expanded(child: Text(_isArabic ? 'العروض المقترحة' : 'Suggested Offers', style: const TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.bold))),
                    IconButton(icon: const Icon(Icons.close, color: Colors.grey), onPressed: () => Navigator.pop(context)),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (availableOffers.isEmpty) ...[
                        Container(
                          padding: const EdgeInsets.all(30),
                          decoration: BoxDecoration(
                            color: Colors.green.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(color: Colors.green.withOpacity(0.3)),
                          ),
                          child: Column(
                            children: [
                              const Text('🎉', style: TextStyle(fontSize: 50)),
                              const SizedBox(height: 16),
                              Text(
                                _isArabic ? 'أنت مشترك في جميع العروض!' : 'You\'ve joined all offers!',
                                style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
                                textAlign: TextAlign.center,
                              ),
                              const SizedBox(height: 8),
                              Text(
                                _isArabic ? 'استمر في الترويج للعروض الحالية' : 'Keep promoting your current offers',
                                style: TextStyle(color: Colors.grey[400], fontSize: 14),
                                textAlign: TextAlign.center,
                              ),
                            ],
                          ),
                        ),
                      ] else ...[
                        Text(
                          _isArabic ? 'عروض متاحة للانضمام (${availableOffers.length})' : 'Available to join (${availableOffers.length})',
                          style: TextStyle(color: Colors.grey[400], fontSize: 14),
                        ),
                        const SizedBox(height: 16),
                        ...availableOffers.take(5).map((offer) => _buildSuggestionCard(offer)),
                      ],
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
  
  Widget _buildSuggestionCard(dynamic offer) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: accentOrange.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          Container(
            width: 50,
            height: 50,
            decoration: BoxDecoration(
              color: accentOrange.withOpacity(0.2),
              borderRadius: BorderRadius.circular(12),
            ),
            child: const Center(child: Text('📦', style: TextStyle(fontSize: 24))),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  offer.companyName ?? 'Offer',
                  style: const TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.w600),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 4),
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                      decoration: BoxDecoration(color: primaryPink.withOpacity(0.2), borderRadius: BorderRadius.circular(6)),
                      child: Text(offer.category ?? '', style: const TextStyle(color: primaryPink, fontSize: 11)),
                    ),
                    const SizedBox(width: 8),
                    Text(offer.reward ?? '', style: TextStyle(color: Colors.grey[500], fontSize: 12)),
                  ],
                ),
              ],
            ),
          ),
          Icon(Icons.arrow_forward_ios, color: accentOrange, size: 16),
        ],
      ),
    );
  }
  
  void _analyzePerformance() {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    
    final user = authProvider.user;
    if (user == null) return;
    
    final stats = user.stats;
    final clicks = stats.totalClicks;
    final conversions = stats.totalConversions;
    final rate = clicks > 0 ? (conversions / clicks) * 100 : 0.0;
    final userOffers = authProvider.userOffers;
    
    _showAnalysisSheet(clicks, conversions, rate, userOffers);
  }
  
  void _showAnalysisSheet(int clicks, int conversions, double rate, List userOffers) {
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
            color: Color(0xFF1A1A1A),
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: Column(
            children: [
              Container(
                margin: const EdgeInsets.only(top: 12),
                width: 40,
                height: 4,
                decoration: BoxDecoration(color: Colors.grey[600], borderRadius: BorderRadius.circular(2)),
              ),
              Padding(
                padding: const EdgeInsets.all(20),
                child: Row(
                  children: [
                    const Text('📈', style: TextStyle(fontSize: 28)),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        _isArabic ? 'تحليل الأداء' : 'Performance Analysis',
                        style: const TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.bold),
                      ),
                    ),
                    IconButton(icon: const Icon(Icons.close, color: Colors.grey), onPressed: () => Navigator.pop(context)),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Summary stats
                      Text(_isArabic ? 'ملخص الأداء' : 'Performance Summary', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
                      const SizedBox(height: 16),
                      Row(
                        children: [
                          Expanded(child: _buildStatCard('📱', _isArabic ? 'النقرات' : 'Clicks', '$clicks', primaryRed)),
                          const SizedBox(width: 12),
                          Expanded(child: _buildStatCard('✅', _isArabic ? 'التحويلات' : 'Conversions', '$conversions', Colors.green)),
                          const SizedBox(width: 12),
                          Expanded(child: _buildStatCard('📈', _isArabic ? 'المعدل' : 'Rate', '${rate.toStringAsFixed(1)}%', accentOrange)),
                        ],
                      ),
                      
                      const SizedBox(height: 24),
                      
                      // Offers analysis
                      Text(_isArabic ? 'تحليل العروض' : 'Offers Analysis', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
                      const SizedBox(height: 16),
                      
                      if (userOffers.isEmpty)
                        Container(
                          padding: const EdgeInsets.all(20),
                          decoration: BoxDecoration(
                            color: Colors.white.withOpacity(0.05),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Center(
                            child: Text(
                              _isArabic ? 'لم تنضم لأي عروض بعد' : 'No offers joined yet',
                              style: TextStyle(color: Colors.grey[500]),
                            ),
                          ),
                        )
                      else
                        ...userOffers.map((uo) => _buildOfferAnalysisCard(uo)),
                      
                      const SizedBox(height: 24),
                      
                      // Recommendations
                      Text(_isArabic ? 'التوصيات' : 'Recommendations', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
                      const SizedBox(height: 16),
                      ..._getRecommendationWidgets(clicks, conversions, userOffers.length),
                      
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
  
  Widget _buildStatCard(String emoji, String title, String value, Color color) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Column(
        children: [
          Text(emoji, style: const TextStyle(fontSize: 22)),
          const SizedBox(height: 6),
          Text(value, style: TextStyle(color: color, fontSize: 18, fontWeight: FontWeight.bold)),
          Text(title, style: TextStyle(color: Colors.grey[500], fontSize: 11)),
        ],
      ),
    );
  }
  
  Widget _buildOfferAnalysisCard(dynamic uo) {
    final offerClicks = uo.totalClicks;
    final offerConversions = uo.totalConversions;
    final status = offerClicks == 0 ? 'not_started' : (offerConversions > 0 ? 'excellent' : 'needs_work');
    
    Color statusColor;
    String statusText;
    IconData statusIcon;
    
    switch (status) {
      case 'excellent':
        statusColor = Colors.green;
        statusText = _isArabic ? 'ممتاز' : 'Excellent';
        statusIcon = Icons.check_circle;
        break;
      case 'needs_work':
        statusColor = Colors.orange;
        statusText = _isArabic ? 'يحتاج تحسين' : 'Needs Work';
        statusIcon = Icons.warning;
        break;
      default:
        statusColor = Colors.grey;
        statusText = _isArabic ? 'لم يبدأ' : 'Not Started';
        statusIcon = Icons.circle_outlined;
    }
    
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(12),
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
                  style: const TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.w600),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(color: statusColor.withOpacity(0.2), borderRadius: BorderRadius.circular(8)),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(statusIcon, color: statusColor, size: 14),
                    const SizedBox(width: 4),
                    Text(statusText, style: TextStyle(color: statusColor, fontSize: 11, fontWeight: FontWeight.w600)),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Row(
            children: [
              _buildMiniStat(Icons.touch_app, '$offerClicks', _isArabic ? 'نقرة' : 'clicks', Colors.blue),
              const SizedBox(width: 16),
              _buildMiniStat(Icons.trending_up, '$offerConversions', _isArabic ? 'تحويل' : 'conv', Colors.green),
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
        Text(value, style: TextStyle(color: color, fontSize: 14, fontWeight: FontWeight.bold)),
        const SizedBox(width: 4),
        Text(label, style: TextStyle(color: Colors.grey[500], fontSize: 12)),
      ],
    );
  }
  
  List<Widget> _getRecommendationWidgets(int clicks, int conversions, int offersCount) {
    final List<Widget> widgets = [];
    
    if (clicks < 10) {
      widgets.add(_buildRecommendationCard(
        Icons.share,
        _isArabic ? 'زيادة النقرات' : 'Increase Clicks',
        _isArabic ? 'شارك روابطك أكثر على المنصات المختلفة' : 'Share your links more on different platforms',
        primaryRed,
      ));
    }
    if (conversions == 0 && clicks > 0) {
      widgets.add(_buildRecommendationCard(
        Icons.edit,
        _isArabic ? 'تحسين المحتوى' : 'Improve Content',
        _isArabic ? 'جرّب كتابة محتوى أكثر جاذبية حول الروابط' : 'Try writing more engaging content around your links',
        accentOrange,
      ));
    }
    if (offersCount < 3) {
      widgets.add(_buildRecommendationCard(
        Icons.add_box,
        _isArabic ? 'تنويع العروض' : 'Diversify Offers',
        _isArabic ? 'أضف المزيد من العروض لزيادة فرص التحويل' : 'Add more offers to increase conversion chances',
        Colors.purple,
      ));
    }
    if (widgets.isEmpty) {
      widgets.add(_buildRecommendationCard(
        Icons.star,
        _isArabic ? 'أداء ممتاز!' : 'Excellent Performance!',
        _isArabic ? 'استمر على هذا المنوال وحافظ على نشاطك' : 'Keep up the great work and stay active',
        Colors.green,
      ));
    }
    
    return widgets;
  }
  
  Widget _buildRecommendationCard(IconData icon, String title, String desc, Color color) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(color: color.withOpacity(0.2), borderRadius: BorderRadius.circular(10)),
            child: Icon(icon, color: color, size: 20),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: TextStyle(color: color, fontSize: 14, fontWeight: FontWeight.w600)),
                const SizedBox(height: 2),
                Text(desc, style: TextStyle(color: Colors.grey[400], fontSize: 12)),
              ],
            ),
          ),
        ],
      ),
    );
  }
  
  void _showTips() {
    _showTipsSheet();
  }
  
  void _showTipsSheet() {
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
            color: Color(0xFF1A1A1A),
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: Column(
            children: [
              Container(margin: const EdgeInsets.only(top: 12), width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey[600], borderRadius: BorderRadius.circular(2))),
              Padding(
                padding: const EdgeInsets.all(20),
                child: Row(
                  children: [
                    const Text('🚀', style: TextStyle(fontSize: 28)),
                    const SizedBox(width: 12),
                    Expanded(child: Text(_isArabic ? 'نصائح للنجاح' : 'Tips for Success', style: const TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.bold))),
                    IconButton(icon: const Icon(Icons.close, color: Colors.grey), onPressed: () => Navigator.pop(context)),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Column(
                    children: [
                      _buildTipCard(Icons.schedule, _isArabic ? 'اختر الوقت المناسب' : 'Choose the Right Time', _isArabic ? 'أفضل الأوقات: المساء والعطلات' : 'Best times: Evenings & weekends', primaryRed),
                      _buildTipCard(Icons.people, _isArabic ? 'استهدف الجمهور الصحيح' : 'Target Right Audience', _isArabic ? 'شارك العروض مع المهتمين بها' : 'Share offers with interested users', Colors.blue),
                      _buildTipCard(Icons.edit, _isArabic ? 'اكتب محتوى جذاب' : 'Write Engaging Content', _isArabic ? 'أضف تجربتك الشخصية' : 'Add your personal experience', Colors.purple),
                      _buildTipCard(Icons.share, _isArabic ? 'نوّع منصات النشر' : 'Diversify Platforms', _isArabic ? 'تويتر، انستقرام، تيك توك' : 'Twitter, Instagram, TikTok', Colors.green),
                      _buildTipCard(Icons.analytics, _isArabic ? 'تابع وحلل النتائج' : 'Track & Analyze', _isArabic ? 'ركّز على ما ينجح' : 'Focus on what works', accentOrange),
                      const SizedBox(height: 16),
                      Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          gradient: LinearGradient(colors: [primaryPink.withOpacity(0.15), primaryRed.withOpacity(0.1)]),
                          borderRadius: BorderRadius.circular(14),
                        ),
                        child: Row(
                          children: [
                            const Text('💡', style: TextStyle(fontSize: 24)),
                            const SizedBox(width: 12),
                            Expanded(child: Text(_isArabic ? 'الاستمرارية هي مفتاح النجاح!' : 'Consistency is key to success!', style: const TextStyle(color: Colors.white, fontSize: 13))),
                          ],
                        ),
                      ),
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
  
  Widget _buildTipCard(IconData icon, String title, String desc, Color color) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: color.withOpacity(0.2)),
      ),
      child: Row(
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(color: color.withOpacity(0.15), borderRadius: BorderRadius.circular(12)),
            child: Icon(icon, color: color, size: 22),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: const TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.w600)),
                const SizedBox(height: 4),
                Text(desc, style: TextStyle(color: Colors.grey[500], fontSize: 12)),
              ],
            ),
          ),
        ],
      ),
    );
  }
  
  void _showPointsSystem() {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final user = authProvider.user;
    final clicks = user?.stats.totalClicks ?? 0;
    final conversions = user?.stats.totalConversions ?? 0;
    final currentPoints = (clicks * 2) + (conversions * 20);
    
    setState(() {
      _currentTitle = _isArabic ? '🏆 نظام النقاط' : '🏆 Points System';
    });
    
    _showPointsSheet(currentPoints);
    
  }
  
  void _showPointsSheet(int currentPoints) {
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
            color: Color(0xFF1A1A1A),
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: Column(
            children: [
              Container(
                margin: const EdgeInsets.only(top: 12),
                width: 40,
                height: 4,
                decoration: BoxDecoration(color: Colors.grey[600], borderRadius: BorderRadius.circular(2)),
              ),
              Padding(
                padding: const EdgeInsets.all(20),
                child: Row(
                  children: [
                    const Text('🏆', style: TextStyle(fontSize: 28)),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        _isArabic ? 'نظام النقاط' : 'Points System',
                        style: const TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.bold),
                      ),
                    ),
                    IconButton(icon: const Icon(Icons.close, color: Colors.grey), onPressed: () => Navigator.pop(context)),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // How to earn points
                      Text(_isArabic ? 'كيف تكسب النقاط؟' : 'How to earn points?', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
                      const SizedBox(height: 16),
                      _buildPointRow('📱', _isArabic ? 'نقرة فريدة' : 'Unique Click', '+2'),
                      _buildPointRow('✅', _isArabic ? 'تحويل ناجح' : 'Successful Conversion', '+20'),
                      // TODO: Enable when implemented
                      // _buildPointRow('🎯', _isArabic ? 'أول تحويل في عرض' : 'First Conv. in Offer', '+50'),
                      // _buildPointRow('📅', _isArabic ? 'نشاط يومي' : 'Daily Activity', '+5'),
                      _buildPointRow('👥', _isArabic ? 'دعوة صديق' : 'Invite Friend', '+30'),
                      
                      const SizedBox(height: 24),
                      
                      // Levels
                      Text(_isArabic ? 'المستويات' : 'Levels', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
                      const SizedBox(height: 16),
                      _buildLevelRow('🌱', _isArabic ? 'مبتدئ' : 'Beginner', '0 - 100'),
                      _buildLevelRow('⭐', _isArabic ? 'نشط' : 'Active', '101 - 500'),
                      _buildLevelRow('💎', _isArabic ? 'محترف' : 'Pro', '501 - 2,000'),
                      _buildLevelRow('👑', _isArabic ? 'خبير' : 'Expert', '2,001 - 10,000'),
                      _buildLevelRow('🏆', _isArabic ? 'أسطوري' : 'Legendary', '10,000+'),
                      
                      const SizedBox(height: 24),
                      
                      // Your status
                      Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          gradient: LinearGradient(colors: [primaryRed.withOpacity(0.15), primaryPink.withOpacity(0.1)]),
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(color: primaryRed.withOpacity(0.3)),
                        ),
                        child: Column(
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(_isArabic ? 'نقاطك الحالية' : 'Your Points', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
                                Text('$currentPoints', style: const TextStyle(color: primaryPink, fontSize: 24, fontWeight: FontWeight.bold)),
                              ],
                            ),
                            const SizedBox(height: 12),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(_isArabic ? 'المستوى' : 'Level', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
                                Text(_isArabic ? _getLevelArabic(currentPoints) : _getLevelEnglish(currentPoints), style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.w600)),
                              ],
                            ),
                            const SizedBox(height: 12),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(_isArabic ? 'للصعود' : 'To Next Level', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
                                Text('${_getPointsToNextLevel(currentPoints)} ${_isArabic ? 'نقطة' : 'pts'}', style: TextStyle(color: accentOrange, fontSize: 16, fontWeight: FontWeight.w600)),
                              ],
                            ),
                          ],
                        ),
                      ),
                      
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
  
  Widget _buildPointRow(String emoji, String title, String points) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          Text(emoji, style: const TextStyle(fontSize: 20)),
          const SizedBox(width: 12),
          Expanded(child: Text(title, style: const TextStyle(color: Colors.white, fontSize: 15))),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(color: primaryPink.withOpacity(0.2), borderRadius: BorderRadius.circular(8)),
            child: Text(points, style: const TextStyle(color: primaryPink, fontSize: 14, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }
  
  Widget _buildLevelRow(String emoji, String title, String range) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.03),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Row(
        children: [
          Text(emoji, style: const TextStyle(fontSize: 18)),
          const SizedBox(width: 12),
          Expanded(child: Text(title, style: const TextStyle(color: Colors.white, fontSize: 14))),
          Text(range, style: TextStyle(color: Colors.grey[500], fontSize: 13)),
        ],
      ),
    );
  }
  
  String _getLevelArabic(int points) {
    if (points >= 10000) return '🏆 أسطوري';
    if (points >= 2000) return '👑 خبير';
    if (points >= 500) return '💎 محترف';
    if (points >= 100) return '⭐ نشط';
    return '🌱 مبتدئ';
  }
  
  String _getLevelEnglish(int points) {
    if (points >= 10000) return '🏆 Legendary';
    if (points >= 2000) return '👑 Expert';
    if (points >= 500) return '💎 Pro';
    if (points >= 100) return '⭐ Active';
    return '🌱 Beginner';
  }
  
  int _getPointsToNextLevel(int points) {
    if (points < 100) return 100 - points;
    if (points < 500) return 500 - points;
    if (points < 2000) return 2000 - points;
    if (points < 10000) return 10000 - points;
    return 0;
  }
  
  void _showResponseSheet() {
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
            color: Color(0xFF1A1A1A),
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
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
                padding: const EdgeInsets.all(20),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        _currentTitle ?? '',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 22,
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
              const Divider(color: Colors.grey, height: 1),
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  padding: const EdgeInsets.all(20),
                  child: Text(
                    _currentResponse ?? '',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 16,
                      height: 1.6,
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

  // ============ FREE CHAT WITH AI ============
  
  Future<void> _sendChatMessage() async {
    final message = _chatController.text.trim();
    if (message.isEmpty) return;
    
    if (!_aiService.hasApiKey) {
      Navigator.push(context, MaterialPageRoute(builder: (_) => const AISettingsScreen()));
      return;
    }
    
    setState(() {
      _chatHistory.add({'role': 'user', 'content': message});
      _chatController.clear();
      _isLoadingAI = true;
    });
    
    _scrollToBottom();
    
    final response = await _aiService.chat(
      message,
      conversationHistory: _chatHistory.length > 10 
          ? _chatHistory.sublist(_chatHistory.length - 10) 
          : _chatHistory,
    );
    
    setState(() {
      _isLoadingAI = false;
      if (response != null) {
        _chatHistory.add({'role': 'assistant', 'content': response});
      } else {
        _chatHistory.add({
          'role': 'assistant', 
          'content': _isArabic ? 'عذراً، حدث خطأ. حاول مرة أخرى.' : 'Sorry, an error occurred. Please try again.'
        });
      }
    });
    
    _scrollToBottom();
  }
  
  void _scrollToBottom() {
    Future.delayed(const Duration(milliseconds: 100), () {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    });
  }
  
  void _newChat() {
    setState(() {
      _chatHistory.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    if (_showChat) {
      return _buildChatScreen();
    }
    return _buildAssistantScreen();
  }

  Widget _buildAssistantScreen() {
    return Scaffold(
      backgroundColor: const Color(0xFF0A0A0A),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        title: Text(
          _isArabic ? 'المساعد الذكي' : 'AI Assistant',
          style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
        ),
        centerTitle: true,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
        // Settings button removed - AI settings accessible via chat button
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
                              colors: [Color(0xFFFFFFFF), Color(0xFFFF8A80), Color(0xFFE53935), Color(0xFFFF006E), Color(0xFF1A1A2E)],
                              stops: [0.0, 0.2, 0.4, 0.7, 1.0],
                              center: Alignment(-0.3, -0.3),
                            ),
                            boxShadow: [BoxShadow(color: primaryRed.withOpacity(0.5), blurRadius: 30, spreadRadius: 5)],
                          ),
                          child: const Center(child: Icon(Icons.auto_awesome, size: 32, color: Colors.white)),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
            
            ShaderMask(
              shaderCallback: (bounds) => const LinearGradient(colors: [primaryRed, primaryPink]).createShader(bounds),
              child: Text(
                _isArabic ? 'مساعدك الذكي' : 'Your AI Assistant',
                style: const TextStyle(color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold),
              ),
            ),
            
            const SizedBox(height: 4),
            Text(_isArabic ? 'كيف يمكنني مساعدتك؟' : 'How can I help you?', style: TextStyle(color: Colors.grey[400], fontSize: 14)),
            
            const SizedBox(height: 24),
            
            // Free Chat Button
            _buildChatButton(),
            
            const SizedBox(height: 20),
            
            // Divider
            Row(
              children: [
                Expanded(child: Divider(color: Colors.grey[800])),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Text(_isArabic ? 'أو اختر' : 'or choose', style: TextStyle(color: Colors.grey[600], fontSize: 12)),
                ),
                Expanded(child: Divider(color: Colors.grey[800])),
              ],
            ),
            
            const SizedBox(height: 20),
            
            // Guided Actions
            _buildActionButton(icon: Icons.analytics_outlined, title: _isArabic ? 'إحصائياتي' : 'My Stats', subtitle: _isArabic ? 'عرض نقراتي وتحويلاتي' : 'View clicks & conversions', color: primaryRed, onTap: _showMyStats),
            const SizedBox(height: 12),
            // TODO: Enable later - Suggestions button
            // _buildActionButton(icon: Icons.lightbulb_outline, title: _isArabic ? 'اقتراحات' : 'Suggestions', subtitle: _isArabic ? 'أفضل العروض لي' : 'Best offers for me', color: accentOrange, onTap: _suggestOffers),
            // const SizedBox(height: 12),
            // TODO: Enable later - Performance Analysis button  
            // _buildActionButton(icon: Icons.trending_up, title: _isArabic ? 'تحليل أدائي' : 'Analyze Performance', subtitle: _isArabic ? 'تحليل شامل مع توصيات' : 'Full analysis with tips', color: primaryPink, onTap: _analyzePerformance),
            // const SizedBox(height: 12),
            _buildActionButton(icon: Icons.tips_and_updates_outlined, title: _isArabic ? 'نصائح النجاح' : 'Success Tips', subtitle: _isArabic ? 'استراتيجيات فعّالة' : 'Effective strategies', color: primaryRed, onTap: _showTips),
            const SizedBox(height: 12),
            _buildActionButton(icon: Icons.emoji_events_outlined, title: _isArabic ? 'نظام النقاط' : 'Points System', subtitle: _isArabic ? 'كيف أحصل على نقاط؟' : 'How to earn points?', color: const Color(0xFFFFD700), onTap: _showPointsSystem),
            
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
  
  Widget _buildChatButton() {
    return GestureDetector(
      onTap: () {
        if (_aiService.hasApiKey) {
          setState(() => _showChat = true);
        } else {
          Navigator.push(context, MaterialPageRoute(builder: (_) => const AISettingsScreen()));
        }
      },
      child: Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [primaryPink.withOpacity(0.2), primaryRed.withOpacity(0.1)],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: primaryPink.withOpacity(0.4), width: 2),
          boxShadow: [BoxShadow(color: primaryPink.withOpacity(0.2), blurRadius: 15, offset: const Offset(0, 5))],
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(color: primaryPink.withOpacity(0.3), borderRadius: BorderRadius.circular(14)),
              child: const Icon(Icons.chat_bubble_outline, color: Colors.white, size: 28),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _isArabic ? 'محادثة حرة مع AI' : 'Free Chat with AI',
                    style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    _aiService.hasApiKey 
                        ? (_isArabic ? 'اسأل أي سؤال واحصل على إجابة ذكية' : 'Ask anything and get smart answers')
                        : (_isArabic ? 'أضف مفتاح API لتفعيل الدردشة' : 'Add API key to enable chat'),
                    style: TextStyle(color: Colors.grey[400], fontSize: 13),
                  ),
                ],
              ),
            ),
            Icon(
              _aiService.hasApiKey ? Icons.arrow_forward_ios : Icons.lock_outline,
              color: primaryPink,
              size: 20,
            ),
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
        borderRadius: BorderRadius.circular(14),
        child: Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            gradient: LinearGradient(colors: [color.withOpacity(0.12), color.withOpacity(0.04)], begin: Alignment.topLeft, end: Alignment.bottomRight),
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: color.withOpacity(0.25)),
          ),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(color: color.withOpacity(0.2), borderRadius: BorderRadius.circular(10)),
                child: Icon(icon, color: color, size: 24),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(title, style: const TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.w600)),
                    const SizedBox(height: 2),
                    Text(subtitle, style: TextStyle(color: Colors.grey[500], fontSize: 12)),
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

  Widget _buildChatScreen() {
    return Scaffold(
      backgroundColor: const Color(0xFF0A0A0A),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        title: Row(
          children: [
            Container(
              width: 36,
              height: 36,
              decoration: BoxDecoration(
                gradient: const LinearGradient(colors: [primaryRed, primaryPink]),
                borderRadius: BorderRadius.circular(10),
              ),
              child: const Icon(Icons.auto_awesome, color: Colors.white, size: 20),
            ),
            const SizedBox(width: 12),
            Text(_isArabic ? 'دردشة AI' : 'AI Chat', style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 18)),
          ],
        ),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios, color: Colors.white),
          onPressed: () => setState(() => _showChat = false),
        ),
        actions: [
          IconButton(icon: const Icon(Icons.refresh, color: Colors.white70), onPressed: _newChat, tooltip: _isArabic ? 'محادثة جديدة' : 'New chat'),
          IconButton(icon: const Icon(Icons.settings, color: Colors.white70), onPressed: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const AISettingsScreen()))),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: _chatHistory.isEmpty
                ? Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.chat_bubble_outline, size: 60, color: Colors.grey[700]),
                        const SizedBox(height: 16),
                        Text(_isArabic ? 'ابدأ المحادثة...' : 'Start chatting...', style: TextStyle(color: Colors.grey[600], fontSize: 16)),
                      ],
                    ),
                  )
                : ListView.builder(
                    controller: _scrollController,
                    padding: const EdgeInsets.all(16),
                    itemCount: _chatHistory.length + (_isLoadingAI ? 1 : 0),
                    itemBuilder: (context, index) {
                      if (_isLoadingAI && index == _chatHistory.length) {
                        return _buildTypingIndicator();
                      }
                      final msg = _chatHistory[index];
                      return _buildChatBubble(msg['content']!, msg['role'] == 'user');
                    },
                  ),
          ),
          _buildChatInput(),
        ],
      ),
    );
  }
  
  Widget _buildChatBubble(String message, bool isUser) {
    return Align(
      alignment: isUser ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.75),
        decoration: BoxDecoration(
          color: isUser ? primaryPink : Colors.white.withOpacity(0.1),
          borderRadius: BorderRadius.only(
            topLeft: const Radius.circular(16),
            topRight: const Radius.circular(16),
            bottomLeft: Radius.circular(isUser ? 16 : 4),
            bottomRight: Radius.circular(isUser ? 4 : 16),
          ),
        ),
        child: Text(message, style: TextStyle(color: isUser ? Colors.white : Colors.white.withOpacity(0.9), fontSize: 15, height: 1.4)),
      ),
    );
  }
  
  Widget _buildTypingIndicator() {
    return Align(
      alignment: Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(color: Colors.white.withOpacity(0.1), borderRadius: BorderRadius.circular(16)),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: List.generate(3, (i) => Container(
            margin: EdgeInsets.only(right: i < 2 ? 4 : 0),
            width: 8,
            height: 8,
            decoration: BoxDecoration(color: primaryPink.withOpacity(0.6), shape: BoxShape.circle),
          )),
        ),
      ),
    );
  }
  
  Widget _buildChatInput() {
    return Container(
      padding: EdgeInsets.only(left: 16, right: 8, top: 12, bottom: MediaQuery.of(context).padding.bottom + 12),
      decoration: BoxDecoration(
        color: const Color(0xFF1A1A1A),
        border: Border(top: BorderSide(color: Colors.grey[800]!)),
      ),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _chatController,
              style: const TextStyle(color: Colors.white),
              maxLines: 4,
              minLines: 1,
              textInputAction: TextInputAction.send,
              onSubmitted: (_) => _sendChatMessage(),
              decoration: InputDecoration(
                hintText: _isArabic ? 'اكتب رسالتك...' : 'Type a message...',
                hintStyle: TextStyle(color: Colors.grey[600]),
                border: InputBorder.none,
                contentPadding: const EdgeInsets.symmetric(horizontal: 4),
              ),
            ),
          ),
          const SizedBox(width: 8),
          Container(
            decoration: BoxDecoration(
              gradient: const LinearGradient(colors: [primaryRed, primaryPink]),
              borderRadius: BorderRadius.circular(12),
            ),
            child: IconButton(
              icon: _isLoadingAI 
                  ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
                  : const Icon(Icons.send, color: Colors.white),
              onPressed: _isLoadingAI ? null : _sendChatMessage,
            ),
          ),
        ],
      ),
    );
  }
}
