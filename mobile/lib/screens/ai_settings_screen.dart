import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../services/ai_service.dart';

class AISettingsScreen extends StatefulWidget {
  const AISettingsScreen({super.key});

  @override
  State<AISettingsScreen> createState() => _AISettingsScreenState();
}

class _AISettingsScreenState extends State<AISettingsScreen> {
  final _apiKeyController = TextEditingController();
  final _aiService = AIService();
  
  AIProvider _selectedProvider = AIProvider.groq;
  bool _isLoading = false;
  bool _obscureKey = true;
  String? _errorMessage;
  String? _successMessage;
  
  static const Color primaryRed = Color(0xFFE53935);
  static const Color primaryPink = Color(0xFFFF006E);
  static const Color successGreen = Color(0xFF4CAF50);

  bool get _isArabic => Localizations.localeOf(context).languageCode == 'ar';

  @override
  void initState() {
    super.initState();
    _loadCurrentSettings();
  }
  
  Future<void> _loadCurrentSettings() async {
    await _aiService.init();
    setState(() {
      _selectedProvider = _aiService.provider;
    });
  }

  @override
  void dispose() {
    _apiKeyController.dispose();
    super.dispose();
  }

  Future<void> _saveApiKey() async {
    final apiKey = _apiKeyController.text.trim();
    
    if (apiKey.isEmpty) {
      setState(() {
        _errorMessage = _isArabic ? 'الرجاء إدخال مفتاح API' : 'Please enter API key';
      });
      return;
    }
    
    setState(() {
      _isLoading = true;
      _errorMessage = null;
      _successMessage = null;
    });
    
    try {
      final success = await _aiService.setApiKey(apiKey, _selectedProvider);
      
      if (success) {
        setState(() {
          _successMessage = _isArabic 
              ? 'تم حفظ المفتاح بنجاح! ✅' 
              : 'API key saved successfully! ✅';
          _apiKeyController.clear();
        });
      } else {
        setState(() {
          _errorMessage = _isArabic 
              ? 'المفتاح غير صالح. تأكد من صحته وحاول مرة أخرى.' 
              : 'Invalid API key. Please check and try again.';
        });
      }
    } catch (e) {
      setState(() {
        _errorMessage = _isArabic ? 'حدث خطأ. حاول مرة أخرى.' : 'An error occurred. Try again.';
      });
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }
  
  Future<void> _removeApiKey() async {
    await _aiService.removeApiKey();
    setState(() {
      _successMessage = _isArabic ? 'تم حذف المفتاح' : 'API key removed';
    });
  }
  
  Future<void> _pasteFromClipboard() async {
    try {
      final clipboardData = await Clipboard.getData(Clipboard.kTextPlain);
      if (clipboardData != null && clipboardData.text != null && clipboardData.text!.isNotEmpty) {
        setState(() {
          _apiKeyController.text = clipboardData.text!;
          _successMessage = _isArabic ? 'تم لصق المفتاح ✓' : 'Key pasted ✓';
        });
      } else {
        setState(() {
          _errorMessage = _isArabic ? 'الحافظة فارغة' : 'Clipboard is empty';
        });
      }
    } catch (e) {
      setState(() {
        _errorMessage = _isArabic ? 'فشل اللصق. جرب النسخ واللصق يدوياً' : 'Paste failed. Try manual copy-paste';
      });
    }
  }
  
  void _goToAssistant() {
    // Go back to AI Assistant screen
    Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF0A0A0A),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        title: Text(
          _isArabic ? 'إعدادات الذكاء الاصطناعي' : 'AI Settings',
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
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header Card
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    primaryRed.withOpacity(0.2),
                    primaryPink.withOpacity(0.1),
                  ],
                ),
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: primaryRed.withOpacity(0.3)),
              ),
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: primaryRed.withOpacity(0.2),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(Icons.key, color: primaryRed, size: 28),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _isArabic ? 'BYOK - استخدم مفتاحك الخاص' : 'BYOK - Bring Your Own Key',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          _isArabic 
                              ? 'أضف مفتاح AI الخاص بك لتفعيل الميزات المتقدمة'
                              : 'Add your own AI key to unlock advanced features',
                          style: TextStyle(
                            color: Colors.grey[400],
                            fontSize: 13,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            
            const SizedBox(height: 32),
            
            // Current Status
            if (_aiService.hasApiKey) ...[
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: successGreen.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: successGreen.withOpacity(0.3)),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.check_circle, color: successGreen),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            _isArabic ? 'المفتاح مُفعّل' : 'API Key Active',
                            style: const TextStyle(
                              color: successGreen,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          Text(
                            '${_isArabic ? 'المزود:' : 'Provider:'} ${_aiService.providerName}',
                            style: TextStyle(
                              color: Colors.grey[400],
                              fontSize: 13,
                            ),
                          ),
                        ],
                      ),
                    ),
                    TextButton(
                      onPressed: _removeApiKey,
                      child: Text(
                        _isArabic ? 'حذف' : 'Remove',
                        style: const TextStyle(color: Colors.red),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
            ],
            
            // Provider Selection
            Text(
              _isArabic ? 'اختر المزود' : 'Select Provider',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 12),
            
            Row(
              children: [
                Expanded(
                  child: _buildProviderCard(
                    provider: AIProvider.openai,
                    name: 'OpenAI',
                    description: 'GPT-4o Mini',
                    icon: Icons.auto_awesome,
                    recommended: true,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: _buildProviderCard(
                    provider: AIProvider.groq,
                    name: 'Groq',
                    description: 'Llama 3.1 70B',
                    icon: Icons.bolt,
                  ),
                ),
              ],
            ),
            
            const SizedBox(height: 24),
            
            // API Key Input
            Text(
              _isArabic ? 'مفتاح API' : 'API Key',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 12),
            
            Container(
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.05),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.white.withOpacity(0.1)),
              ),
              child: TextField(
                controller: _apiKeyController,
                obscureText: _obscureKey,
                style: const TextStyle(color: Colors.white),
                enableInteractiveSelection: true,
                contextMenuBuilder: (context, editableTextState) {
                  return AdaptiveTextSelectionToolbar.editableText(
                    editableTextState: editableTextState,
                  );
                },
                decoration: InputDecoration(
                  hintText: _selectedProvider == AIProvider.groq 
                      ? 'gsk_...' 
                      : 'sk-...',
                  hintStyle: TextStyle(color: Colors.grey[600]),
                  border: InputBorder.none,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                  suffixIcon: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        icon: Icon(Icons.content_paste, color: Colors.grey[500], size: 20),
                        onPressed: _pasteFromClipboard,
                        tooltip: _isArabic ? 'لصق' : 'Paste',
                      ),
                      IconButton(
                        icon: Icon(
                          _obscureKey ? Icons.visibility : Icons.visibility_off,
                          color: Colors.grey[500],
                          size: 20,
                        ),
                        onPressed: () => setState(() => _obscureKey = !_obscureKey),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            
            const SizedBox(height: 8),
            
            // How to get key
            GestureDetector(
              onTap: () => _showHowToGetKey(),
              child: Row(
                children: [
                  Icon(Icons.help_outline, color: Colors.grey[500], size: 16),
                  const SizedBox(width: 6),
                  Text(
                    _isArabic ? 'كيف أحصل على مفتاح API؟' : 'How to get an API key?',
                    style: TextStyle(
                      color: Colors.grey[500],
                      fontSize: 13,
                      decoration: TextDecoration.underline,
                    ),
                  ),
                ],
              ),
            ),
            
            // Error/Success Messages
            if (_errorMessage != null) ...[
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.red.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.error, color: Colors.red, size: 20),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        _errorMessage!,
                        style: const TextStyle(color: Colors.red, fontSize: 13),
                      ),
                    ),
                  ],
                ),
              ),
            ],
            
            if (_successMessage != null) ...[
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: successGreen.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.check_circle, color: successGreen, size: 20),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        _successMessage!,
                        style: const TextStyle(color: successGreen, fontSize: 13),
                      ),
                    ),
                  ],
                ),
              ),
            ],
            
            const SizedBox(height: 24),
            
            // Save Button
            SizedBox(
              width: double.infinity,
              height: 52,
              child: ElevatedButton(
                onPressed: _isLoading ? null : _saveApiKey,
                style: ElevatedButton.styleFrom(
                  backgroundColor: primaryRed,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                child: _isLoading
                    ? const SizedBox(
                        width: 24,
                        height: 24,
                        child: CircularProgressIndicator(
                          color: Colors.white,
                          strokeWidth: 2,
                        ),
                      )
                    : Text(
                        _isArabic ? 'حفظ المفتاح' : 'Save API Key',
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
              ),
            ),
            
            const SizedBox(height: 32),
            
            // Features List
            Text(
              _isArabic ? 'الميزات المُفعّلة مع المفتاح' : 'Features Unlocked with Key',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 16),
            
            _buildFeatureItem(
              icon: Icons.chat,
              title: _isArabic ? 'محادثة حرة مع AI' : 'Free Chat with AI',
              description: _isArabic ? 'اسأل أي سؤال واحصل على إجابة ذكية' : 'Ask any question and get smart answers',
              onTap: _aiService.hasApiKey ? () => _goToAssistant() : null,
            ),
            _buildFeatureItem(
              icon: Icons.analytics,
              title: _isArabic ? 'تحليل عميق' : 'Deep Analysis',
              description: _isArabic ? 'تحليل شامل لأدائك مع توصيات مخصصة' : 'Comprehensive analysis with personalized recommendations',
              onTap: _aiService.hasApiKey ? () => _goToAssistant() : null,
            ),
            _buildFeatureItem(
              icon: Icons.edit_note,
              title: _isArabic ? 'اقتراح محتوى' : 'Content Suggestions',
              description: _isArabic ? 'اقتراحات لمحتوى ترويجي جذاب' : 'Suggestions for engaging promotional content',
              onTap: _aiService.hasApiKey ? () => _goToAssistant() : null,
            ),
            _buildFeatureItem(
              icon: Icons.lightbulb,
              title: _isArabic ? 'استراتيجيات مخصصة' : 'Custom Strategies',
              description: _isArabic ? 'خطط عمل مبنية على بياناتك' : 'Action plans based on your data',
              onTap: _aiService.hasApiKey ? () => _goToAssistant() : null,
            ),
            
            const SizedBox(height: 32),
            
            // Privacy Note
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.05),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(Icons.lock, color: Colors.grey[500], size: 20),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      _isArabic 
                          ? 'مفتاحك يُخزن بشكل آمن على جهازك فقط ولا يُرسل لخوادمنا. الاتصال مباشر بينك وبين مزود AI.'
                          : 'Your key is stored securely on your device only and is never sent to our servers. Connection is direct between you and the AI provider.',
                      style: TextStyle(
                        color: Colors.grey[500],
                        fontSize: 12,
                        height: 1.4,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
  
  Widget _buildProviderCard({
    required AIProvider provider,
    required String name,
    required String description,
    required IconData icon,
    bool recommended = false,
  }) {
    final isSelected = _selectedProvider == provider;
    
    return GestureDetector(
      onTap: () => setState(() => _selectedProvider = provider),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isSelected ? primaryRed.withOpacity(0.15) : Colors.white.withOpacity(0.05),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isSelected ? primaryRed : Colors.white.withOpacity(0.1),
            width: isSelected ? 2 : 1,
          ),
        ),
        child: Column(
          children: [
            if (recommended)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                margin: const EdgeInsets.only(bottom: 8),
                decoration: BoxDecoration(
                  color: successGreen,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  _isArabic ? 'موصى به' : 'Recommended',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            Icon(
              icon,
              color: isSelected ? primaryRed : Colors.grey[500],
              size: 28,
            ),
            const SizedBox(height: 8),
            Text(
              name,
              style: TextStyle(
                color: isSelected ? Colors.white : Colors.grey[400],
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              description,
              style: TextStyle(
                color: Colors.grey[600],
                fontSize: 11,
              ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
  
  Widget _buildFeatureItem({
    required IconData icon,
    required String title,
    required String description,
    VoidCallback? onTap,
  }) {
    final isEnabled = onTap != null;
    
    return GestureDetector(
      onTap: isEnabled ? onTap : () {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(_isArabic ? 'أضف مفتاح API أولاً لتفعيل هذه الميزة' : 'Add API key first to enable this feature'),
            backgroundColor: primaryRed,
            duration: const Duration(seconds: 2),
          ),
        );
      },
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isEnabled ? primaryPink.withOpacity(0.1) : Colors.white.withOpacity(0.05),
          borderRadius: BorderRadius.circular(12),
          border: isEnabled ? Border.all(color: primaryPink.withOpacity(0.3)) : null,
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: isEnabled ? primaryPink.withOpacity(0.2) : primaryPink.withOpacity(0.15),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(icon, color: isEnabled ? primaryPink : Colors.grey, size: 22),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      color: isEnabled ? Colors.white : Colors.grey[400],
                      fontWeight: FontWeight.w600,
                      fontSize: 14,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    description,
                    style: TextStyle(
                      color: Colors.grey[500],
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
            if (isEnabled)
              Icon(Icons.arrow_forward_ios, color: primaryPink, size: 16)
            else
              Icon(Icons.lock_outline, color: Colors.grey[600], size: 18),
          ],
        ),
      ),
    );
  }
  
  void _showHowToGetKey() {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF1A1A2E),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (context) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              _isArabic ? 'كيف تحصل على مفتاح API؟' : 'How to get an API key?',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 20),
            
            _buildStepItem(
              number: '1',
              title: 'Groq (${_isArabic ? 'موصى به' : 'Recommended'})',
              steps: _isArabic
                  ? ['اذهب إلى console.groq.com', 'سجّل حساب مجاني', 'انسخ مفتاح API']
                  : ['Go to console.groq.com', 'Create free account', 'Copy API key'],
            ),
            const SizedBox(height: 16),
            
            _buildStepItem(
              number: '2',
              title: 'OpenAI',
              steps: _isArabic
                  ? ['اذهب إلى platform.openai.com', 'سجّل حساب وأضف رصيد', 'أنشئ مفتاح API']
                  : ['Go to platform.openai.com', 'Create account & add credit', 'Create API key'],
            ),
            
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
  
  Widget _buildStepItem({
    required String number,
    required String title,
    required List<String> steps,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.05),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 28,
                height: 28,
                decoration: BoxDecoration(
                  color: primaryRed,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Center(
                  child: Text(
                    number,
                    style: const TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Text(
                title,
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          ...steps.asMap().entries.map((entry) => Padding(
            padding: const EdgeInsets.only(bottom: 6),
            child: Row(
              children: [
                Text(
                  '${entry.key + 1}. ',
                  style: TextStyle(color: Colors.grey[500]),
                ),
                Text(
                  entry.value,
                  style: TextStyle(color: Colors.grey[400]),
                ),
              ],
            ),
          )),
        ],
      ),
    );
  }
}

