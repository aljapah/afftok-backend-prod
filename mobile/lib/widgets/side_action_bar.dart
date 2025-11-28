import 'package:flutter/material.dart';
import 'package:share_plus/share_plus.dart';
import '../models/offer.dart';
import '../utils/app_localizations.dart';
import '../services/favorites_manager.dart';

class SideActionBar extends StatefulWidget {
  final Offer offer;
  final VoidCallback? onFavoriteChanged;

  const SideActionBar({
    Key? key,
    required this.offer,
    this.onFavoriteChanged,
  }) : super(key: key);

  @override
  State<SideActionBar> createState() => _SideActionBarState();
}

class _SideActionBarState extends State<SideActionBar> with SingleTickerProviderStateMixin {
  bool _isSaved = false;
  late AnimationController _animController;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      duration: const Duration(milliseconds: 300),
      vsync: this,
    );
    _checkIfSaved();
  }

  Future<void> _checkIfSaved() async {
    try {
      final favManager = await FavoritesManager.getInstance();
      final isSaved = await favManager.isOfferSaved(widget.offer.id);
      if (mounted) {
        setState(() {
          _isSaved = isSaved;
        });
      }
    } catch (e) {
      print('Error checking if saved: $e');
    }
  }

  Future<void> _toggleFavorite() async {
    if (_isLoading) return;

    setState(() {
      _isLoading = true;
    });

    try {
      final favManager = await FavoritesManager.getInstance();
      
      if (_isSaved) {
        // حذف من المفضلة
        final success = await favManager.removeOffer(widget.offer.id);
        if (success && mounted) {
          setState(() {
            _isSaved = false;
            _isLoading = false;
          });
          
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                AppLocalizations.of(context)!.removedFromFavorites ?? 'تم الحذف من المفضلة',
                style: const TextStyle(color: Colors.white),
              ),
              backgroundColor: Colors.black87,
              duration: const Duration(seconds: 1),
              behavior: SnackBarBehavior.floating,
            ),
          );
          
          // استدعاء callback
          if (widget.onFavoriteChanged != null) {
            widget.onFavoriteChanged!();
          }
        }
      } else {
        // إضافة للمفضلة
        final success = await favManager.saveOffer(widget.offer);
        if (success && mounted) {
          setState(() {
            _isSaved = true;
            _isLoading = false;
          });
          
          _animController.forward().then((_) => _animController.reverse());
          
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                AppLocalizations.of(context)!.addedToFavorites ?? 'تم الحفظ في المفضلة ❤️',
                style: const TextStyle(color: Colors.white),
              ),
              backgroundColor: Colors.black87,
              duration: const Duration(seconds: 1),
              behavior: SnackBarBehavior.floating,
            ),
          );
        }
      }
    } catch (e) {
      print('Error toggling favorite: $e');
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        // Save button
        _ActionButton(
          icon: _isSaved ? Icons.favorite : Icons.favorite_border,
          label: _isSaved ? lang.saved : lang.save ?? 'حفظ',
          onTap: _toggleFavorite,
          color: _isSaved ? Colors.red : Colors.white,
          isLoading: _isLoading,
        ),

        const SizedBox(height: 24),

        // Share button
        _ActionButton(
          icon: Icons.share,
          label: lang.share,
          onTap: () {
            Share.share(
              '${lang.checkOut} ${widget.offer.companyName}! ${widget.offer.reward}\n${widget.offer.offerUrl}',
              subject: widget.offer.companyName,
            );
          },
        ),

        const SizedBox(height: 24),

        // Info button
        _ActionButton(
          icon: Icons.info_outline,
          label: lang.info,
          onTap: () {
            _showInfoBottomSheet(context);
          },
        ),
      ],
    );
  }

  void _showInfoBottomSheet(BuildContext context) {
    final lang = AppLocalizations.of(context)!;
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF1a1a1a),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 50,
                  height: 50,
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Center(
                    child: Text(
                      widget.offer.companyName[0],
                      style: const TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: Colors.black,
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
                        widget.offer.companyName,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Text(
                        widget.offer.category,
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.6),
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 24),
            Text(
              lang.offerDetails,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 12),
            Text(
              widget.offer.description,
              style: TextStyle(
                color: Colors.white.withOpacity(0.8),
                fontSize: 14,
                height: 1.5,
              ),
            ),
            const SizedBox(height: 24),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFFF006E).withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: const Color(0xFFFF006E).withOpacity(0.3),
                ),
              ),
              child: Row(
                children: [
                  const Icon(
                    Icons.card_giftcard,
                    color: Color(0xFFFF006E),
                    size: 24,
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          lang.reward,
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.6),
                            fontSize: 12,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          widget.offer.reward,
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
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
}

class _ActionButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onTap;
  final Color color;
  final bool isLoading;

  const _ActionButton({
    Key? key,
    required this.icon,
    required this.label,
    required this.onTap,
    this.color = Colors.white,
    this.isLoading = false,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: isLoading ? null : onTap,
      child: Column(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: Colors.black.withOpacity(0.3),
              shape: BoxShape.circle,
              border: Border.all(
                color: Colors.white.withOpacity(0.2),
                width: 1,
              ),
            ),
            child: isLoading
                ? const Padding(
                    padding: EdgeInsets.all(12),
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                    ),
                  )
                : Icon(
                    icon,
                    color: color,
                    size: 28,
                  ),
          ),
          const SizedBox(height: 4),
          Text(
            label,
            style: TextStyle(
              color: color,
              fontSize: 12,
              fontWeight: FontWeight.w500,
              shadows: [
                Shadow(
                  color: Colors.black.withOpacity(0.5),
                  blurRadius: 4,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

