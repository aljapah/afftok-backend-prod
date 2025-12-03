import 'package:flutter/material.dart';
import 'package:flutter/services.dart' show rootBundle;
import 'package:webview_flutter/webview_flutter.dart';
import 'package:url_launcher/url_launcher.dart';
import '../utils/app_localizations.dart';

class TermsScreen extends StatefulWidget {
  const TermsScreen({Key? key}) : super(key: key);

  @override
  State<TermsScreen> createState() => _TermsScreenState();
}

class _TermsScreenState extends State<TermsScreen> {
  WebViewController? _webViewController;
  bool _isLoading = true;
  bool _isInitialized = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_isInitialized) {
      _isInitialized = true;
      _initializeWebView();
    }
  }

  void _initializeWebView() async {
    final lang = AppLocalizations.of(context);
    final isArabic = lang.locale.languageCode == 'ar';
    
    _webViewController = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..setNavigationDelegate(
        NavigationDelegate(
          onNavigationRequest: (NavigationRequest request) {
            // Handle mailto links
            if (request.url.startsWith('mailto:')) {
              _launchEmail(request.url);
              return NavigationDecision.prevent;
            }
            return NavigationDecision.navigate;
          },
          onPageFinished: (_) {
            // Hide elements and set language
            _hideAppElements();
            _setLanguage(isArabic);
            if (mounted) {
              setState(() {
                _isLoading = false;
              });
            }
          },
        ),
      );

    try {
      final htmlContent = await rootBundle.loadString('assets/html/terms.html');
      _webViewController?.loadHtmlString(
        htmlContent,
        baseUrl: 'https://afftok.com/',
      );
    } catch (e) {
      print('[TermsScreen] Error loading HTML: $e');
    }
  }

  void _hideAppElements() {
    _webViewController?.runJavaScript("""
      var langBtn = document.querySelector('.language-toggle');
      if (langBtn) langBtn.style.display = 'none';
      var backBtn = document.querySelector('.back-btn');
      if (backBtn) backBtn.style.display = 'none';
      var footerSection = document.querySelector('.footer-section');
      if (footerSection) footerSection.style.display = 'none';
      var footer = document.querySelector('footer');
      if (footer) footer.style.display = 'none';
    """);
  }

  void _setLanguage(bool isArabic) {
    final targetLang = isArabic ? 'ar' : 'en';
    _webViewController?.runJavaScript("""
      var html = document.documentElement;
      var body = document.body;
      var targetLang = '$targetLang';
      
      if (targetLang === 'en') {
        html.lang = 'en';
        html.dir = 'ltr';
        body.classList.add('en');
        body.classList.remove('ar');
      } else {
        html.lang = 'ar';
        html.dir = 'rtl';
        body.classList.remove('en');
        body.classList.add('ar');
      }
      
      // Use the page's built-in translations
      if (typeof translations !== 'undefined' && translations[targetLang]) {
        currentLang = targetLang;
        Object.keys(translations[targetLang]).forEach(function(key) {
          var element = document.getElementById(key);
          if (element) {
            element.innerHTML = translations[targetLang][key];
          }
        });
      }
    """);
  }

  Future<void> _launchEmail(String emailUrl) async {
    try {
      final Uri uri = Uri.parse(emailUrl);
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } catch (e) {
      print('[TermsScreen] Error launching email: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    final isArabic = lang.locale.languageCode == 'ar';

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        elevation: 0,
        title: Text(
          isArabic ? 'شروط الاستخدام' : 'Terms of Use',
          style: const TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.bold,
            fontSize: 20,
          ),
        ),
        iconTheme: const IconThemeData(color: Colors.white),
        centerTitle: true,
      ),
      body: Stack(
        children: [
          if (_webViewController != null)
            WebViewWidget(controller: _webViewController!),
          if (_isLoading)
            const Center(
              child: CircularProgressIndicator(color: Color(0xFFFF006E)),
            ),
        ],
      ),
    );
  }
}
