import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import '../models/offer.dart';
import '../utils/app_localizations.dart';
import '../providers/auth_provider.dart';

class OfferCard extends StatefulWidget {
  final Offer offer;
  final VoidCallback? onFavoriteChanged;

  const OfferCard({Key? key, required this.offer, this.onFavoriteChanged}) : super(key: key);

  @override
  State<OfferCard> createState() => _OfferCardState();
}

class _OfferCardState extends State<OfferCard> with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _scaleAnimation;
  late Animation<double> _fadeAnimation;
  late Animation<Offset> _slideAnimation;

  bool isAdded = false;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 1200),
      vsync: this,
    );

    _scaleAnimation = Tween<double>(begin: 1.0, end: 1.05).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );

    _fadeAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(
        parent: _controller,
        curve: const Interval(0.0, 0.5, curve: Curves.easeIn),
      ),
    );

    _slideAnimation = Tween<Offset>(
      begin: const Offset(0, -0.3),
      end: Offset.zero,
    ).animate(
      CurvedAnimation(
        parent: _controller,
        curve: const Interval(0.2, 0.7, curve: Curves.easeOut),
      ),
    );

    _controller.forward();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _openRegistration(String url) async {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final userID = authProvider.currentUser?.id;
    
    // Check if already added
    if (authProvider.hasAddedOffer(widget.offer.id)) {
      setState(() => isAdded = true);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(AppLocalizations.of(context)!.offerAlreadyAdded),
          backgroundColor: const Color(0xFF8E2DE2), // App theme color
        ),
      );
      return;
    }
    
    String finalUrl = url;
    if (userID != null && userID.isNotEmpty) {
      final separator = url.contains('?') ? '&' : '?';
      finalUrl = '$url${separator}promoter=$userID';
    }
    
    final Uri uri = Uri.parse(finalUrl);

    if (await canLaunchUrl(uri)) {
      // First, add the offer to user's profile via API
      print('[OfferCard] Adding offer ${widget.offer.id} to profile...');
      final success = await authProvider.addOfferToProfile(widget.offer.id);
      print('[OfferCard] Add offer result: $success');
      
      // Always mark as added and open URL (even if API says already joined)
      setState(() => isAdded = true);
      
      // Open the registration URL
      await launchUrl(uri, mode: LaunchMode.externalApplication);

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            '${widget.offer.companyName} ${AppLocalizations.of(context)!.offerAdded}',
          ),
          backgroundColor: Colors.green,
        ),
      );
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Could not launch registration link'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    final gradients = [
      [const Color(0xFF8E2DE2), const Color(0xFF4A00E0)],
      [const Color(0xFFFF006E), const Color(0xFFFF4D00)],
      [const Color(0xFF00D9FF), const Color(0xFF0099FF)],
      [const Color(0xFF00FF88), const Color(0xFF00CC66)],
    ];

    final gradientIndex = widget.offer.companyName.hashCode % gradients.length;
    final gradient = gradients[gradientIndex];

    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return Transform.scale(
          scale: _scaleAnimation.value,
          child: Stack(
            fit: StackFit.expand,
            children: [
              Image.network(
                widget.offer.imageUrl,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) {
                  return Container(
                    color: Colors.grey[900],
                    child: const Center(
                      child: Icon(Icons.image_not_supported, size: 50, color: Colors.white30),
                    ),
                  );
                },
              ),

              Container(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [
                      Colors.transparent,
                      Colors.black.withOpacity(0.3),
                      Colors.black.withOpacity(0.8),
                    ],
                    stops: const [0.0, 0.5, 1.0],
                  ),
                ),
              ),

              Positioned(
                left: 16,
                right: 80,
                bottom: 100,
                child: FadeTransition(
                  opacity: _fadeAnimation,
                  child: SlideTransition(
                    position: _slideAnimation,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          width: 60,
                          height: 60,
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(12),
                            boxShadow: [
                              BoxShadow(
                                color: Colors.black.withOpacity(0.3),
                                blurRadius: 10,
                                offset: const Offset(0, 4),
                              ),
                            ],
                          ),
                          padding: const EdgeInsets.all(8),
                          child: Image.network(
                            widget.offer.logoUrl,
                            errorBuilder: (context, error, stackTrace) {
                              return Center(
                                child: Text(
                                  widget.offer.companyName[0],
                                  style: const TextStyle(
                                    fontSize: 24,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.black87,
                                  ),
                                ),
                              );
                            },
                          ),
                        ),

                        const SizedBox(height: 16),

                        Row(
                          mainAxisAlignment: MainAxisAlignment.start,
                          children: [
                            Expanded(
                              child: Text(
                                widget.offer.companyName,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 24,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                            const SizedBox(width: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                              decoration: BoxDecoration(
                                color: Colors.amber.withOpacity(0.2),
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(
                                  color: Colors.amber.withOpacity(0.5),
                                  width: 1,
                                ),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  const Icon(Icons.star, color: Colors.amber, size: 16),
                                  const SizedBox(width: 4),
                                  Text(
                                    widget.offer.rating.toString(),
                                    style: const TextStyle(
                                      color: Colors.amber,
                                      fontSize: 14,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),

                        const SizedBox(height: 8),

                        TweenAnimationBuilder(
                          duration: const Duration(seconds: 2),
                          tween: Tween<double>(begin: 0.95, end: 1.05),
                          curve: Curves.easeInOut,
                          builder: (context, double scale, child) {
                            return Transform.scale(
                              scale: scale,
                              child: Container(
                                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                                decoration: BoxDecoration(
                                  gradient: LinearGradient(colors: gradient),
                                  borderRadius: BorderRadius.circular(20),
                                  boxShadow: [
                                    BoxShadow(
                                      color: gradient[0].withOpacity(0.5),
                                      blurRadius: 12,
                                      spreadRadius: 2,
                                    ),
                                  ],
                                ),
                                child: Text(
                                  widget.offer.reward,
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            );
                          },
                          onEnd: () {
                            if (mounted) setState(() {});
                          },
                        ),

                        const SizedBox(height: 12),

                        Text(
                          widget.offer.description,
                          style: const TextStyle(color: Colors.white, fontSize: 15),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),

                        const SizedBox(height: 12),

                        Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(Icons.people, color: Colors.white70, size: 16),
                            const SizedBox(width: 6),
                            Text(
                              '${widget.offer.usersCount} users earned',
                              style: TextStyle(
                                color: Colors.white.withOpacity(0.7),
                                fontSize: 13,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ],
                        ),

                        const SizedBox(height: 12),

                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                          decoration: BoxDecoration(
                            color: Colors.white.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: Colors.white.withOpacity(0.3)),
                          ),
                          child: Text(
                            widget.offer.category,
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 12,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ),

                        const SizedBox(height: 20),

                        SizedBox(
                          width: double.infinity,
                          height: 50,
                          child: ElevatedButton(
                            onPressed: isAdded
                                ? null
                                : () => _openRegistration(widget.offer.offerUrl),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: isAdded ? Colors.grey : Colors.white,
                              foregroundColor: Colors.black,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(25),
                              ),
                              elevation: 8,
                              shadowColor: Colors.white.withOpacity(0.5),
                            ),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(isAdded ? Icons.check : Icons.person_add_alt_1, size: 20),
                                const SizedBox(width: 8),
                                Text(
                                  isAdded
                                      ? lang.addedToMyOffers
                                      : lang.registerAndAddOffer,
                                  style: const TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),

                        const SizedBox(height: 8),

                        Center(
                          child: Text(
                            lang.redirectNotice,
                            style: const TextStyle(
                              color: Colors.white70,
                              fontSize: 12,
                            ),
                          ),
                        ),

                        const SizedBox(height: 8),

                        SizedBox(
                          width: double.infinity,
                          height: 50,
                          child: IconButton(
                            icon: Icon(
                              widget.offer.isRegistered ? Icons.favorite : Icons.favorite_border,
                              color: widget.offer.isRegistered ? Colors.red : Colors.white,
                            ),
                            onPressed: widget.onFavoriteChanged,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
