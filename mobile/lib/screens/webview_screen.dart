import 'package:flutter/material.dart';
import 'package:webview_flutter/webview_flutter.dart';
import 'package:provider/provider.dart';
import '../utils/app_localizations.dart';
import '../providers/auth_provider.dart';

class WebViewScreen extends StatefulWidget {
  final String url;
  final String title;
  final String offerId;

  const WebViewScreen({
    Key? key,
    required this.url,
    required this.title,
    required this.offerId,
  }) : super(key: key);

  @override
  State<WebViewScreen> createState() => _WebViewScreenState();
}

class _WebViewScreenState extends State<WebViewScreen> {
  late WebViewController _webViewController;
  bool _isLoading = true;
  bool _isJoining = false;
  bool _hasJoined = false;

  @override
  void initState() {
    super.initState();
    _initializeWebView();
    _checkIfAlreadyJoined();
  }

  void _checkIfAlreadyJoined() {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    setState(() {
      _hasJoined = authProvider.hasAddedOffer(widget.offerId);
    });
  }

  void _initializeWebView() {
    _webViewController = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..setNavigationDelegate(
        NavigationDelegate(
          onPageStarted: (String url) {
            setState(() {
              _isLoading = true;
            });
          },
          onPageFinished: (String url) {
            setState(() {
              _isLoading = false;
            });
          },
          onWebResourceError: (WebResourceError error) {
            setState(() {
              _isLoading = false;
            });
          },
        ),
      )
      ..loadRequest(Uri.parse(widget.url));
  }

  Future<void> _joinOffer() async {
    print('[WebViewScreen] _joinOffer called, offerId: ${widget.offerId}');
    print('[WebViewScreen] _isJoining: $_isJoining, _hasJoined: $_hasJoined');
    
    if (_isJoining || _hasJoined) {
      print('[WebViewScreen] Already joining or joined, returning');
      return;
    }

    setState(() {
      _isJoining = true;
    });

    print('[WebViewScreen] Calling addOfferToProfile...');
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final success = await authProvider.addOfferToProfile(widget.offerId);
    print('[WebViewScreen] addOfferToProfile result: $success');

    setState(() {
      _isJoining = false;
      _hasJoined = success;
    });

    if (success) {
      print('[WebViewScreen] Offer added successfully!');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(AppLocalizations.of(context).offerAddedSuccessfully),
          backgroundColor: Colors.green,
          duration: const Duration(seconds: 2),
        ),
      );
      // Go back to previous screen
      Navigator.pop(context, true);
    } else {
      print('[WebViewScreen] Failed to add offer');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(AppLocalizations.of(context).failedToAddOffer),
          backgroundColor: Colors.red,
          duration: const Duration(seconds: 2),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: Text(widget.title, style: const TextStyle(color: Colors.white)),
        leading: IconButton(
          icon: const Icon(Icons.close, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: Column(
        children: [
          Expanded(
            child: Stack(
              children: [
                WebViewWidget(controller: _webViewController),
                if (_isLoading)
                  Container(
                    color: Colors.black.withOpacity(0.3),
                    child: const Center(
                      child: CircularProgressIndicator(
                        color: Color(0xFF8E2DE2),
                      ),
                    ),
                  ),
              ],
            ),
          ),
          // Bottom bar with join button
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.grey[900],
              border: Border(
                top: BorderSide(color: Colors.grey[800]!),
              ),
            ),
            child: SafeArea(
              child: SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _hasJoined ? null : (_isJoining ? null : _joinOffer),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: _hasJoined ? Colors.green : const Color(0xFF8E2DE2),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: _isJoining
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                            color: Colors.white,
                            strokeWidth: 2,
                          ),
                        )
                      : Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              _hasJoined ? Icons.check_circle : Icons.add_circle_outline,
                              size: 24,
                            ),
                            const SizedBox(width: 8),
                            Text(
                              _hasJoined ? lang.offerAdded : lang.addOfferToMyList,
                              style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
