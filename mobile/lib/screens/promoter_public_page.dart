import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:webview_flutter/webview_flutter.dart';
import 'package:flutter/services.dart' show rootBundle;
import 'package:url_launcher/url_launcher.dart';
import '../utils/app_localizations.dart';
import '../providers/auth_provider.dart';

class PromoterPublicPage extends StatefulWidget {
  final String username;

  const PromoterPublicPage({
    Key? key,
    required this.username,
  }) : super(key: key);

  @override
  State<PromoterPublicPage> createState() => _PromoterPublicPageState();
}

class _PromoterPublicPageState extends State<PromoterPublicPage> {
  late final WebViewController _webViewController;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _initializeWebView();
    // Refresh user offers when page opens
    WidgetsBinding.instance.addPostFrameCallback((_) {
      Provider.of<AuthProvider>(context, listen: false).reloadUserOffers();
    });
  }

  void _initializeWebView() async {
    _webViewController = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..setNavigationDelegate(
        NavigationDelegate(
          onNavigationRequest: (NavigationRequest request) {
            print('[PromoterPublicPage] Navigation request: ${request.url}');
            
            // Allow initial page load and base URL
            if (request.url == 'https://afftok-backend-prod-production.up.railway.app/' ||
                request.url == 'about:blank') {
              return NavigationDecision.navigate;
            }
            
            // Handle mailto links
            if (request.url.startsWith('mailto:')) {
              print('[PromoterPublicPage] Opening email: ${request.url}');
              _openEmailApp();
              return NavigationDecision.prevent;
            }
            
            // Handle external social media links
            if (request.url.startsWith('http') &&
                (request.url.contains('instagram.com') ||
                 request.url.contains('tiktok.com') ||
                 request.url.contains('twitter.com') ||
                 request.url.contains('youtube.com'))) {
              launchUrl(Uri.parse(request.url), mode: LaunchMode.externalApplication);
              return NavigationDecision.prevent;
            }
            
            // Handle affiliate links - open in external browser
            if (request.url.contains('/api/c/')) {
              print('[PromoterPublicPage] Opening affiliate link externally: ${request.url}');
              launchUrl(Uri.parse(request.url), mode: LaunchMode.externalApplication);
              return NavigationDecision.prevent;
            }
            
            // Handle privacy and terms pages
            if (request.url.contains('privacy.html') || request.url.contains('terms.html')) {
              _showHtmlPage(request.url);
              return NavigationDecision.prevent;
            }
            
            return NavigationDecision.navigate;
          },
          onPageFinished: (_) {
            _injectUserData();
            setState(() {
              _isLoading = false;
            });
          },
        ),
      );

    // Always use the local HTML from assets folder
    print('[PromoterPublicPage] Loading landing page from assets');
    await _loadLocalHtml();
  }
  
  Future<void> _refreshPage() async {
    setState(() {
      _isLoading = true;
    });
    
    // Reload user offers from API
    await Provider.of<AuthProvider>(context, listen: false).reloadUserOffers();
    
    // Re-inject user data
    _injectUserData();
    
    setState(() {
      _isLoading = false;
    });
  }
  
  Future<void> _loadLocalHtml() async {
    try {
      final htmlContent = await rootBundle.loadString('assets/html/promoter_landing.html');
      print('[PromoterPublicPage] Loaded HTML from assets');
      _webViewController.loadHtmlString(
        htmlContent, 
        baseUrl: 'https://afftok-backend-prod-production.up.railway.app/'
      );
    } catch (e) {
      print('[PromoterPublicPage] Error loading HTML from assets: $e');
    }
  }


  void _showHtmlPage(String url) async {
    String title = 'سياسة الخصوصية';
    String filePath = 'assets/html/privacy.html';

    if (url.contains('terms')) {
      title = 'شروط الاستخدام';
      filePath = 'assets/html/terms.html';
    }

    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => _HtmlPageScreen(title: title, filePath: filePath),
      ),
    );
  }
  
  Future<void> _openEmailApp() async {
    // Use a simple mailto URL string
    final String emailUrl = 'mailto:support@afftokapp.com?subject=AffTok%20Support';
    
    print('[PromoterPublicPage] Attempting to launch email: $emailUrl');
    
    try {
      // Try to launch directly without checking canLaunch
      await launchUrl(
        Uri.parse(emailUrl),
        mode: LaunchMode.externalApplication,
      );
    } catch (e) {
      print('[PromoterPublicPage] Error launching email: $e');
      // If it fails, try with platform default
      try {
        await launchUrl(Uri.parse(emailUrl));
      } catch (e2) {
        print('[PromoterPublicPage] Second attempt failed: $e2');
      }
    }
  }

  void _injectUserData() {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final user = authProvider.currentUser;
    final userOffers = authProvider.userOffers;

    if (user != null) {
      print('[PromoterPublicPage] Injecting user data: ${user.username}');
      print('[PromoterPublicPage] User offers count: ${userOffers.length}');
      
      // Build offers JSON array
      final offersJson = userOffers.map((offer) {
        final title = offer.offerInfo?.title ?? offer.offerTitle;
        final description = offer.offerInfo?.description ?? '';
        final imageUrl = offer.offerInfo?.imageUrl ?? '';
        final category = offer.offerInfo?.category ?? '';
        final offerId = offer.offerId;
        
        return '''{
          "id": "$offerId",
          "title": "${_escapeJs(title)}",
          "description": "${_escapeJs(description)}",
          "image_url": "$imageUrl",
          "category": "$category",
          "payout": 0,
          "payout_type": "cpa"
        }''';
      }).join(',');
      
      final jsCode = '''
        console.log('Flutter: Injecting user data');
        
        // Set promoter ID for API calls
        window.promoterId = "${user.id}";
        
        // Update profile section
        document.getElementById('profile-name').textContent = "${_escapeJs(user.displayName)}";
        document.getElementById('profile-username').textContent = "@${user.username}";
        
        // Update bio if available
        const profileBio = document.getElementById('profile-bio');
        if (profileBio) {
          const bioText = "${_escapeJs(user.bio ?? '')}";
          if (bioText) {
            profileBio.textContent = bioText;
            profileBio.style.display = 'block';
          } else {
            profileBio.style.display = 'none';
          }
        }
        
        // Update profile image if available
        const profileImg = document.getElementById('profile-image');
        if (profileImg && "${user.avatarUrl ?? ''}") {
          profileImg.src = "${user.avatarUrl ?? ''}";
        }
        
        // Store user data globally
        window.promoterData = {
          name: "${_escapeJs(user.displayName)}",
          username: "${user.username}",
          bio: "${_escapeJs(user.bio ?? '')}",
          rating: 4.5,
          offersCount: ${userOffers.length},
          clicksCount: ${user.stats.totalClicks}
        };
        
        // Inject offers data
        window.offersData = [$offersJson];
        
        // Update offers grid
        const API_BASE = 'https://afftok-backend-prod-production.up.railway.app';
        const offersGrid = document.getElementById('offers-grid');
        if (offersGrid && window.offersData && window.offersData.length > 0) {
          offersGrid.innerHTML = '';
          window.offersData.forEach(offer => {
            const card = document.createElement('div');
            card.className = 'offer-card';
            
            const imageUrl = offer.image_url || 'https://via.placeholder.com/400x200?text=' + encodeURIComponent(offer.title);
            const affiliateLink = API_BASE + '/api/c/' + offer.id + '?promoter=' + window.promoterId;
            const currentLang = document.documentElement.lang || 'ar';
            const buttonText = currentLang === 'ar' ? 'احصل على الرابط' : 'Get Link';
            
            card.innerHTML = \`
              <img src="\${imageUrl}" alt="\${offer.title}" class="offer-image" onerror="this.src='https://via.placeholder.com/400x200?text=\${encodeURIComponent(offer.title)}'">
              <div class="offer-content">
                <h3 class="offer-name">\${offer.title}</h3>
                <p class="offer-description">\${offer.description || 'عرض حصري'}</p>
                <a href="\${affiliateLink}" target="_blank" class="offer-button">
                  <i class="fas fa-external-link-alt"></i>
                  \${buttonText}
                </a>
              </div>
            \`;
            offersGrid.appendChild(card);
          });
        } else if (offersGrid) {
          offersGrid.innerHTML = '<p style="grid-column: 1/-1; text-align: center; color: var(--text-secondary); padding: 40px;">لا توجد عروض حالياً</p>';
        }
        
        console.log('Flutter: User data injected successfully');
      ''';

      _webViewController.runJavaScript(jsCode);
    } else {
      print('[PromoterPublicPage] No user data available');
    }
  }
  
  String _escapeJs(String text) {
    return text
      .replaceAll('\\', '\\\\')
      .replaceAll('"', '\\"')
      .replaceAll("'", "\\'")
      .replaceAll('\n', '\\n')
      .replaceAll('\r', '\\r');
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);

    return PopScope(
      canPop: true,
      child: Scaffold(
        backgroundColor: Colors.black,
        appBar: AppBar(
          backgroundColor: Colors.black,
          title: Text(lang.myPublicPage ?? "صفحتي", style: const TextStyle(color: Colors.white)),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back, color: Colors.white),
            onPressed: () => Navigator.pop(context),
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.refresh, color: Colors.white),
              onPressed: _refreshPage,
            ),
          ],
        ),
        body: Stack(
          children: [
            WebViewWidget(controller: _webViewController),
            if (_isLoading)
              const Center(
                child: CircularProgressIndicator(color: Colors.white),
              ),
          ],
        ),
      ),
    );
  }
}

// Separate screen for Privacy and Terms pages with proper scrolling
class _HtmlPageScreen extends StatefulWidget {
  final String title;
  final String filePath;

  const _HtmlPageScreen({
    required this.title,
    required this.filePath,
  });

  @override
  State<_HtmlPageScreen> createState() => _HtmlPageScreenState();
}

class _HtmlPageScreenState extends State<_HtmlPageScreen> {
  late final WebViewController _controller;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _initController();
  }

  void _initController() async {
    _controller = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..setNavigationDelegate(
        NavigationDelegate(
          onPageFinished: (_) {
            // Inject scroll fix
            _controller.runJavaScript('''
              document.documentElement.style.cssText = 'overflow: auto !important; height: auto !important; -webkit-overflow-scrolling: touch !important;';
              document.body.style.cssText = 'overflow: auto !important; height: auto !important; -webkit-overflow-scrolling: touch !important; min-height: 100vh;';
              
              var content = document.querySelector('.content');
              if (content) {
                content.style.cssText = 'max-height: none !important; overflow: visible !important; height: auto !important;';
              }
              
              var sections = document.querySelectorAll('.section-content');
              sections.forEach(function(s) {
                s.style.cssText = 'max-height: none !important; overflow: visible !important; height: auto !important;';
              });
            ''');
            setState(() {
              _isLoading = false;
            });
          },
        ),
      );

    try {
      final htmlContent = await rootBundle.loadString(widget.filePath);
      _controller.loadHtmlString(htmlContent);
    } catch (e) {
      print('Error loading HTML: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: Text(widget.title, style: const TextStyle(color: Colors.white)),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: Stack(
        children: [
          WebViewWidget(controller: _controller),
          if (_isLoading)
            const Center(
              child: CircularProgressIndicator(color: Colors.white),
            ),
        ],
      ),
    );
  }
}
