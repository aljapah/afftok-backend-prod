import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../utils/app_localizations.dart';

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
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _openInBrowser();
  }

  Future<void> _openInBrowser() async {
    try {
      final Uri uri = Uri.parse(widget.url);
      if (await canLaunchUrl(uri)) {
        await launchUrl(
          uri,
          mode: LaunchMode.externalApplication,
        );
        // After opening, show success dialog
        if (mounted) {
          _showRegistrationDialog();
        }
      } else {
        if (mounted) {
          _showErrorDialog();
        }
      }
    } catch (e) {
      if (mounted) {
        _showErrorDialog();
      }
    }
  }

  void _showRegistrationDialog() {
    final lang = AppLocalizations.of(context);
    
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        backgroundColor: Colors.grey[900],
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
        ),
        title: Row(
          children: [
            const Icon(Icons.info_outline, color: Colors.blue),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                lang.registrationOpened,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 18,
                ),
              ),
            ),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              lang.registrationInstructions,
              style: TextStyle(
                color: Colors.grey[300],
                fontSize: 14,
                height: 1.5,
              ),
            ),
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.blue[900]?.withOpacity(0.3),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.blue[700]!.withOpacity(0.5)),
              ),
              child: Row(
                children: [
                  Icon(
                    Icons.lightbulb_outline,
                    color: Colors.blue[300],
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      lang.registrationTip,
                      style: TextStyle(
                        color: Colors.blue[100],
                        fontSize: 12,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context); // Close dialog
              Navigator.pop(context); // Go back to previous screen
            },
            child: Text(
              lang.cancel,
              style: TextStyle(color: Colors.grey[400]),
            ),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context); // Close dialog
              _showReferralLinkDialog();
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF8E2DE2),
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: Text(lang.iRegistered),
          ),
        ],
      ),
    );
  }

  void _showReferralLinkDialog() {
    final lang = AppLocalizations.of(context);
    final TextEditingController linkController = TextEditingController();

    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        backgroundColor: Colors.grey[900],
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
        ),
        title: Row(
          children: [
            const Icon(Icons.link, color: Colors.green),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                lang.enterReferralLink,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 18,
                ),
              ),
            ),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              lang.enterReferralLinkDescription,
              style: TextStyle(
                color: Colors.grey[300],
                fontSize: 14,
                height: 1.5,
              ),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: linkController,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                hintText: 'https://example.com/ref/yourcode',
                hintStyle: TextStyle(color: Colors.grey[600]),
                filled: true,
                fillColor: Colors.grey[850],
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide.none,
                ),
                prefixIcon: const Icon(Icons.link, color: Colors.grey),
              ),
              keyboardType: TextInputType.url,
              maxLines: 2,
            ),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.orange[900]?.withOpacity(0.3),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.orange[700]!.withOpacity(0.5)),
              ),
              child: Row(
                children: [
                  Icon(
                    Icons.warning_amber_outlined,
                    color: Colors.orange[300],
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      lang.referralLinkWarning,
                      style: TextStyle(
                        color: Colors.orange[100],
                        fontSize: 11,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context); // Close dialog
              Navigator.pop(context); // Go back to previous screen
            },
            child: Text(
              lang.cancel,
              style: TextStyle(color: Colors.grey[400]),
            ),
          ),
          ElevatedButton(
            onPressed: () {
              final link = linkController.text.trim();
              if (link.isNotEmpty && Uri.tryParse(link) != null) {
                // TODO: Save the referral link to backend
                Navigator.pop(context); // Close dialog
                Navigator.pop(context); // Go back to previous screen
                _showSuccessDialog();
              } else {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(lang.invalidLink),
                    backgroundColor: Colors.red,
                  ),
                );
              }
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.green,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: Text(lang.saveLink),
          ),
        ],
      ),
    );
  }

  void _showSuccessDialog() {
    final lang = AppLocalizations.of(context);
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: Colors.grey[900],
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.green[900]?.withOpacity(0.3),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.check_circle,
                color: Colors.green,
                size: 60,
              ),
            ),
            const SizedBox(height: 24),
            Text(
              lang.offerAddedSuccessfully,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            Text(
              lang.offerAddedDescription,
              style: TextStyle(
                color: Colors.grey[400],
                fontSize: 14,
              ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
        actions: [
          Center(
            child: ElevatedButton(
              onPressed: () {
                Navigator.pop(context);
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF8E2DE2),
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(
                  horizontal: 32,
                  vertical: 12,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
              child: Text(lang.done),
            ),
          ),
        ],
      ),
    );
  }

  void _showErrorDialog() {
    final lang = AppLocalizations.of(context);
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: Colors.grey[900],
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
        ),
        title: Row(
          children: [
            const Icon(Icons.error_outline, color: Colors.red),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                lang.error,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 18,
                ),
              ),
            ),
          ],
        ),
        content: Text(
          lang.couldNotOpenWebsite,
          style: TextStyle(
            color: Colors.grey[300],
            fontSize: 14,
          ),
        ),
        actions: [
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context); // Close dialog
              Navigator.pop(context); // Go back
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF8E2DE2),
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: Text(lang.ok),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const CircularProgressIndicator(
              color: Color(0xFF8E2DE2),
            ),
            const SizedBox(height: 24),
            Text(
              AppLocalizations.of(context).openingWebsite,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

