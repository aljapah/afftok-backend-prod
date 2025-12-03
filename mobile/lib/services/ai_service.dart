import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

enum AIProvider {
  groq,
  openai,
  // claude, // يمكن إضافته لاحقاً
}

class AIService {
  static const String _apiKeyStorageKey = 'user_ai_api_key';
  static const String _providerStorageKey = 'user_ai_provider';
  
  static final AIService _instance = AIService._internal();
  factory AIService() => _instance;
  AIService._internal();
  
  String? _apiKey;
  AIProvider _provider = AIProvider.groq;
  
  bool get hasApiKey => _apiKey != null && _apiKey!.isNotEmpty;
  AIProvider get provider => _provider;
  String get providerName {
    switch (_provider) {
      case AIProvider.groq:
        return 'Groq';
      case AIProvider.openai:
        return 'OpenAI';
    }
  }
  
  // ============ INITIALIZATION ============
  
  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _apiKey = prefs.getString(_apiKeyStorageKey);
    final providerIndex = prefs.getInt(_providerStorageKey) ?? 0;
    _provider = AIProvider.values[providerIndex];
  }
  
  // ============ API KEY MANAGEMENT ============
  
  Future<bool> setApiKey(String apiKey, AIProvider provider) async {
    try {
      // Validate the key first
      final isValid = await validateApiKey(apiKey, provider);
      if (!isValid) return false;
      
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_apiKeyStorageKey, apiKey);
      await prefs.setInt(_providerStorageKey, provider.index);
      
      _apiKey = apiKey;
      _provider = provider;
      
      return true;
    } catch (e) {
      print('[AIService] Error setting API key: $e');
      return false;
    }
  }
  
  Future<void> removeApiKey() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_apiKeyStorageKey);
    await prefs.remove(_providerStorageKey);
    _apiKey = null;
  }
  
  Future<bool> validateApiKey(String apiKey, AIProvider provider) async {
    try {
      // Simple validation request
      final response = await chat(
        'Say "OK" only.',
        apiKey: apiKey,
        providerOverride: provider,
      );
      return response != null && response.isNotEmpty;
    } catch (e) {
      print('[AIService] API key validation failed: $e');
      return false;
    }
  }
  
  // ============ CHAT FUNCTION ============
  
  Future<String?> chat(
    String message, {
    String? systemPrompt,
    String? apiKey,
    AIProvider? providerOverride,
    List<Map<String, String>>? conversationHistory,
  }) async {
    final key = apiKey ?? _apiKey;
    final currentProvider = providerOverride ?? _provider;
    
    if (key == null || key.isEmpty) {
      return null;
    }
    
    try {
      switch (currentProvider) {
        case AIProvider.groq:
          return await _chatWithGroq(key, message, systemPrompt, conversationHistory);
        case AIProvider.openai:
          return await _chatWithOpenAI(key, message, systemPrompt, conversationHistory);
      }
    } catch (e) {
      print('[AIService] Chat error: $e');
      return null;
    }
  }
  
  // ============ GROQ API ============
  
  Future<String?> _chatWithGroq(
    String apiKey,
    String message,
    String? systemPrompt,
    List<Map<String, String>>? history,
  ) async {
    const endpoint = 'https://api.groq.com/openai/v1/chat/completions';
    
    final messages = <Map<String, String>>[];
    
    // System prompt
    if (systemPrompt != null) {
      messages.add({'role': 'system', 'content': systemPrompt});
    } else {
      messages.add({
        'role': 'system',
        'content': _getDefaultSystemPrompt(),
      });
    }
    
    // Conversation history
    if (history != null) {
      messages.addAll(history);
    }
    
    // User message
    messages.add({'role': 'user', 'content': message});
    
    final response = await http.post(
      Uri.parse(endpoint),
      headers: {
        'Authorization': 'Bearer $apiKey',
        'Content-Type': 'application/json',
      },
      body: jsonEncode({
        'model': 'llama-3.1-70b-versatile',
        'messages': messages,
        'temperature': 0.7,
        'max_tokens': 1024,
      }),
    );
    
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      return data['choices'][0]['message']['content'];
    } else {
      print('[AIService] Groq error: ${response.statusCode} - ${response.body}');
      return null;
    }
  }
  
  // ============ OPENAI API ============
  
  Future<String?> _chatWithOpenAI(
    String apiKey,
    String message,
    String? systemPrompt,
    List<Map<String, String>>? history,
  ) async {
    const endpoint = 'https://api.openai.com/v1/chat/completions';
    
    final messages = <Map<String, String>>[];
    
    // System prompt
    if (systemPrompt != null) {
      messages.add({'role': 'system', 'content': systemPrompt});
    } else {
      messages.add({
        'role': 'system',
        'content': _getDefaultSystemPrompt(),
      });
    }
    
    // Conversation history
    if (history != null) {
      messages.addAll(history);
    }
    
    // User message
    messages.add({'role': 'user', 'content': message});
    
    final response = await http.post(
      Uri.parse(endpoint),
      headers: {
        'Authorization': 'Bearer $apiKey',
        'Content-Type': 'application/json',
      },
      body: jsonEncode({
        'model': 'gpt-4o-mini',
        'messages': messages,
        'temperature': 0.7,
        'max_tokens': 1024,
      }),
    );
    
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      return data['choices'][0]['message']['content'];
    } else {
      print('[AIService] OpenAI error: ${response.statusCode} - ${response.body}');
      return null;
    }
  }
  
  // ============ SYSTEM PROMPTS ============
  
  String _getDefaultSystemPrompt() {
    return '''أنت مساعد ذكي متخصص في التسويق بالعمولة (Affiliate Marketing) لتطبيق AffTok.

مهمتك:
- مساعدة المروّجين على زيادة نقراتهم وتحويلاتهم
- تقديم نصائح عملية ومفيدة
- الإجابة على أسئلتهم بشكل مختصر ومفيد
- اقتراح استراتيجيات للنجاح

ملاحظات مهمة:
- لا تتحدث عن الأرباح المالية أو المبالغ - التطبيق للتتبع فقط
- ركّز على النقرات والتحويلات والأداء
- كن إيجابياً ومشجعاً
- أجب بالعربية إذا كان السؤال بالعربية، وبالإنجليزية إذا كان بالإنجليزية
- اجعل إجاباتك مختصرة ومفيدة (لا تزيد عن 200 كلمة)''';
  }
  
  String getAnalysisPrompt({
    required int totalClicks,
    required int totalConversions,
    required int offersCount,
    required List<Map<String, dynamic>> offersData,
  }) {
    final offersInfo = offersData.map((o) => 
      '- ${o['title']}: ${o['clicks']} نقرة, ${o['conversions']} تحويل, فئة: ${o['category']}'
    ).join('\n');
    
    return '''حلل أداء هذا المروّج وقدم نصائح مخصصة:

الإحصائيات العامة:
- إجمالي النقرات: $totalClicks
- إجمالي التحويلات: $totalConversions
- عدد العروض: $offersCount

تفاصيل العروض:
$offersInfo

قدم:
1. تحليل سريع للأداء (3 أسطر)
2. نقاط القوة (2 نقاط)
3. نقاط التحسين (2 نقاط)
4. خطة عمل مقترحة (3 خطوات)''';
  }
  
  String getContentSuggestionPrompt({
    required String offerTitle,
    required String offerCategory,
    required String platform,
  }) {
    return '''اقترح محتوى ترويجي لهذا العرض:

العرض: $offerTitle
الفئة: $offerCategory
المنصة: $platform

اكتب:
1. عنوان جذاب (سطر واحد)
2. نص المنشور (3-4 أسطر)
3. هاشتاقات مناسبة (5 هاشتاقات)
4. أفضل وقت للنشر''';
  }
}

