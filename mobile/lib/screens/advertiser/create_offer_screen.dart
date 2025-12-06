import 'dart:io';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:image_picker/image_picker.dart';
import '../../providers/auth_provider.dart';
import '../../services/advertiser_service.dart';
import '../../services/imgbb_service.dart';

class CreateOfferScreen extends StatefulWidget {
  final Map<String, dynamic>? existingOffer; // For editing

  const CreateOfferScreen({super.key, this.existingOffer});

  @override
  State<CreateOfferScreen> createState() => _CreateOfferScreenState();
}

class _CreateOfferScreenState extends State<CreateOfferScreen> {
  final _formKey = GlobalKey<FormState>();
  final _advertiserService = AdvertiserService();
  
  // English fields
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  
  // Arabic fields
  final _titleArController = TextEditingController();
  final _descriptionArController = TextEditingController();
  final _termsArController = TextEditingController();
  
  // URLs
  final _imageUrlController = TextEditingController();
  final _logoUrlController = TextEditingController();
  final _destinationUrlController = TextEditingController();
  
  // Other fields
  final _payoutController = TextEditingController();
  final _commissionController = TextEditingController();
  
  String _selectedCategory = 'general';
  String _selectedPayoutType = 'cpa';
  bool _isLoading = false;
  bool _isUploadingImage = false;
  bool _isUploadingLogo = false;
  bool _agreedToTerms = false;
  File? _selectedImage;
  File? _selectedLogo;
  final ImagePicker _picker = ImagePicker();

  final List<Map<String, String>> _categories = [
    {'value': 'general', 'en': 'General', 'ar': 'عام'},
    {'value': 'finance', 'en': 'Finance', 'ar': 'مالية'},
    {'value': 'ecommerce', 'en': 'E-Commerce', 'ar': 'تجارة إلكترونية'},
    {'value': 'gaming', 'en': 'Gaming', 'ar': 'ألعاب'},
    {'value': 'health', 'en': 'Health', 'ar': 'صحة'},
    {'value': 'education', 'en': 'Education', 'ar': 'تعليم'},
    {'value': 'technology', 'en': 'Technology', 'ar': 'تقنية'},
    {'value': 'travel', 'en': 'Travel', 'ar': 'سفر'},
    {'value': 'food', 'en': 'Food & Delivery', 'ar': 'طعام وتوصيل'},
    {'value': 'entertainment', 'en': 'Entertainment', 'ar': 'ترفيه'},
  ];

  final List<Map<String, String>> _payoutTypes = [
    {'value': 'cpa', 'en': 'CPA (Cost Per Action)', 'ar': 'CPA (لكل إجراء)'},
    {'value': 'cpl', 'en': 'CPL (Cost Per Lead)', 'ar': 'CPL (لكل عميل محتمل)'},
    {'value': 'cps', 'en': 'CPS (Cost Per Sale)', 'ar': 'CPS (لكل بيع)'},
    {'value': 'cpi', 'en': 'CPI (Cost Per Install)', 'ar': 'CPI (لكل تثبيت)'},
  ];

  @override
  void initState() {
    super.initState();
    if (widget.existingOffer != null) {
      _populateForm(widget.existingOffer!);
    }
  }

  void _populateForm(Map<String, dynamic> offer) {
    _titleController.text = offer['title'] ?? '';
    _descriptionController.text = offer['description'] ?? '';
    _titleArController.text = offer['title_ar'] ?? '';
    _descriptionArController.text = offer['description_ar'] ?? '';
    _termsArController.text = offer['terms_ar'] ?? '';
    _imageUrlController.text = offer['image_url'] ?? '';
    _logoUrlController.text = offer['logo_url'] ?? '';
    _destinationUrlController.text = offer['destination_url'] ?? '';
    _payoutController.text = (offer['payout'] ?? 0).toString();
    _commissionController.text = (offer['commission'] ?? 0).toString();
    _selectedCategory = offer['category'] ?? 'general';
    _selectedPayoutType = offer['payout_type'] ?? 'cpa';
  }

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    _titleArController.dispose();
    _descriptionArController.dispose();
    _termsArController.dispose();
    _imageUrlController.dispose();
    _logoUrlController.dispose();
    _destinationUrlController.dispose();
    _payoutController.dispose();
    _commissionController.dispose();
    super.dispose();
  }

  Future<void> _pickAndUploadImage({required bool isLogo}) async {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    
    try {
      final XFile? pickedFile = await _picker.pickImage(
        source: ImageSource.gallery,
        maxWidth: 1200,
        maxHeight: 1200,
        imageQuality: 85,
      );

      if (pickedFile == null) return;

      setState(() {
        if (isLogo) {
          _isUploadingLogo = true;
          _selectedLogo = File(pickedFile.path);
        } else {
          _isUploadingImage = true;
          _selectedImage = File(pickedFile.path);
        }
      });

      // Upload to ImgBB
      final imageUrl = await ImgBBService.uploadImage(File(pickedFile.path));

      setState(() {
        if (isLogo) {
          _isUploadingLogo = false;
          if (imageUrl != null) {
            _logoUrlController.text = imageUrl;
          }
        } else {
          _isUploadingImage = false;
          if (imageUrl != null) {
            _imageUrlController.text = imageUrl;
          }
        }
      });

      if (imageUrl == null && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(isArabic ? 'فشل رفع الصورة' : 'Failed to upload image'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } catch (e) {
      setState(() {
        _isUploadingImage = false;
        _isUploadingLogo = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(e.toString()),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    
    final offerData = {
      'title': _titleController.text.trim(),
      'description': _descriptionController.text.trim(),
      'title_ar': _titleArController.text.trim(),
      'description_ar': _descriptionArController.text.trim(),
      'terms_ar': _termsArController.text.trim(),
      'image_url': _imageUrlController.text.trim(),
      'logo_url': _logoUrlController.text.trim(),
      'destination_url': _destinationUrlController.text.trim(),
      'category': _selectedCategory,
      'payout': int.tryParse(_payoutController.text) ?? 0,
      'commission': int.tryParse(_commissionController.text) ?? 0,
      'payout_type': _selectedPayoutType,
    };

    try {
      bool success;
      if (widget.existingOffer != null) {
        success = await _advertiserService.updateOffer(
          authProvider.token!,
          widget.existingOffer!['id'],
          offerData,
        );
      } else {
        success = await _advertiserService.createOffer(
          authProvider.token!,
          offerData,
        );
      }

      setState(() => _isLoading = false);

      if (success && mounted) {
        final isArabic = Localizations.localeOf(context).languageCode == 'ar';
        
        if (widget.existingOffer != null) {
          // Editing - just show snackbar
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(isArabic ? 'تم تحديث العرض وإرساله للمراجعة' : 'Offer updated and sent for review'),
              backgroundColor: Colors.green,
            ),
          );
          Navigator.pop(context, true);
        } else {
          // New offer - show payment details dialog
          await _showPaymentDetailsDialog(context, isArabic);
          if (mounted) {
            Navigator.pop(context, true);
          }
        }
      }
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(e.toString()),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _showPaymentDetailsDialog(BuildContext context, bool isArabic) async {
    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => Dialog(
        backgroundColor: const Color(0xFF1A1A2E),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        child: Container(
          constraints: const BoxConstraints(maxWidth: 450),
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Success Icon
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.green.withOpacity(0.2),
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.check_circle, color: Colors.green, size: 48),
              ),
              const SizedBox(height: 20),
              Text(
                isArabic ? 'تم إرسال العرض للمراجعة!' : 'Offer Submitted for Review!',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 12),
              Text(
                isArabic
                  ? 'سيتم مراجعة عرضك وتفعيله خلال 24-48 ساعة. فيما يلي بيانات الدفع الخاصة بنسبة المنصة:'
                  : 'Your offer will be reviewed and activated within 24-48 hours. Below are the platform commission payment details:',
                style: TextStyle(
                  color: Colors.white.withOpacity(0.7),
                  fontSize: 14,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 24),
              
              // Billing Info - No bank details
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.05),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFF6C63FF).withOpacity(0.3)),
                ),
                child: Column(
                  children: [
                    const Icon(Icons.check_circle, color: Colors.green, size: 40),
                    const SizedBox(height: 12),
                    Text(
                      isArabic
                        ? 'يتم سداد المستحقات عبر تحويل بنكي وفق التفاصيل المرفقة داخل الفاتورة الشهرية.'
                        : 'Payments are made via bank transfer according to the details in the monthly invoice.',
                      style: TextStyle(
                        color: Colors.white.withOpacity(0.9),
                        fontSize: 14,
                        height: 1.5,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 12),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.calendar_today, color: Color(0xFF6C63FF), size: 16),
                        const SizedBox(width: 6),
                        Text(
                          isArabic ? 'الفاتورة: يوم 1 من كل شهر' : 'Invoice: 1st of each month',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.7),
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => Navigator.pop(context),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF6C63FF),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: Text(
                    isArabic ? 'فهمت، متابعة' : 'Got it, Continue',
                    style: const TextStyle(fontSize: 16, color: Colors.white, fontWeight: FontWeight.bold),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showOfferPreview(BuildContext context, bool isArabic) {
    final title = isArabic && _titleArController.text.isNotEmpty 
        ? _titleArController.text 
        : _titleController.text;
    final description = isArabic && _descriptionArController.text.isNotEmpty 
        ? _descriptionArController.text 
        : _descriptionController.text;
    final imageUrl = _imageUrlController.text;
    final logoUrl = _logoUrlController.text;
    final payout = _payoutController.text;
    final commission = _commissionController.text;
    final category = _categories.firstWhere(
      (c) => c['value'] == _selectedCategory,
      orElse: () => {'ar': 'عام', 'en': 'General'},
    );

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        maxChildSize: 0.95,
        minChildSize: 0.5,
        builder: (context, scrollController) => Container(
          decoration: const BoxDecoration(
            color: Colors.black,
            borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
          ),
          child: Column(
            children: [
              // Handle
              Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.symmetric(vertical: 12),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.3),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              // Header
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      isArabic ? 'معاينة العرض' : 'Offer Preview',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.white),
                      onPressed: () => Navigator.pop(context),
                    ),
                  ],
                ),
              ),
              const Divider(color: Colors.white24),
              // Preview Content - Similar to Offer Card
              Expanded(
                child: SingleChildScrollView(
                  controller: scrollController,
                  child: Column(
                    children: [
                      // Offer Card Preview
                      Container(
                        margin: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: Colors.black,
                          borderRadius: BorderRadius.circular(24),
                          border: Border.all(color: Colors.white.withOpacity(0.1)),
                        ),
                        child: Stack(
                          children: [
                            // Background Image
                            ClipRRect(
                              borderRadius: BorderRadius.circular(24),
                              child: imageUrl.isNotEmpty
                                  ? Image.network(
                                      imageUrl,
                                      height: 500,
                                      width: double.infinity,
                                      fit: BoxFit.cover,
                                      errorBuilder: (_, __, ___) => Container(
                                        height: 500,
                                        color: Colors.grey.shade900,
                                        child: const Icon(Icons.image, color: Colors.white30, size: 80),
                                      ),
                                    )
                                  : Container(
                                      height: 500,
                                      color: Colors.grey.shade900,
                                      child: const Icon(Icons.image, color: Colors.white30, size: 80),
                                    ),
                            ),
                            // Gradient Overlay
                            Positioned.fill(
                              child: Container(
                                decoration: BoxDecoration(
                                  borderRadius: BorderRadius.circular(24),
                                  gradient: LinearGradient(
                                    begin: Alignment.topCenter,
                                    end: Alignment.bottomCenter,
                                    colors: [
                                      Colors.transparent,
                                      Colors.black.withOpacity(0.3),
                                      Colors.black.withOpacity(0.9),
                                    ],
                                    stops: const [0.3, 0.6, 1.0],
                                  ),
                                ),
                              ),
                            ),
                            // Logo
                            if (logoUrl.isNotEmpty)
                              Positioned(
                                top: 16,
                                left: 16,
                                child: Container(
                                  width: 50,
                                  height: 50,
                                  decoration: BoxDecoration(
                                    color: Colors.white,
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  child: ClipRRect(
                                    borderRadius: BorderRadius.circular(12),
                                    child: Image.network(
                                      logoUrl,
                                      fit: BoxFit.cover,
                                      errorBuilder: (_, __, ___) => const Icon(Icons.business),
                                    ),
                                  ),
                                ),
                              ),
                            // Category Badge
                            Positioned(
                              top: 16,
                              right: 16,
                              child: Container(
                                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                                decoration: BoxDecoration(
                                  color: const Color(0xFFFF006E),
                                  borderRadius: BorderRadius.circular(20),
                                ),
                                child: Text(
                                  isArabic ? category['ar']! : category['en']!,
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 12,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ),
                            // Bottom Content
                            Positioned(
                              bottom: 0,
                              left: 0,
                              right: 0,
                              child: Padding(
                                padding: const EdgeInsets.all(20),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    // Title
                                    Text(
                                      title.isEmpty ? (isArabic ? 'عنوان العرض' : 'Offer Title') : title,
                                      style: const TextStyle(
                                        color: Colors.white,
                                        fontSize: 22,
                                        fontWeight: FontWeight.bold,
                                        height: 1.3,
                                      ),
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                    const SizedBox(height: 8),
                                    // Description
                                    Text(
                                      description.isEmpty 
                                          ? (isArabic ? 'وصف العرض' : 'Offer description') 
                                          : description,
                                      style: TextStyle(
                                        color: Colors.white.withOpacity(0.8),
                                        fontSize: 14,
                                        height: 1.4,
                                      ),
                                      maxLines: 3,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                    const SizedBox(height: 16),
                                    // Commission Badge
                                    Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                                      decoration: BoxDecoration(
                                        gradient: const LinearGradient(
                                          colors: [Color(0xFFFF006E), Color(0xFFFF4D94)],
                                        ),
                                        borderRadius: BorderRadius.circular(20),
                                      ),
                                      child: Row(
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          const Icon(Icons.monetization_on, color: Colors.white, size: 18),
                                          const SizedBox(width: 6),
                                          Text(
                                            isArabic 
                                                ? 'العمولة: ${commission.isEmpty ? "0" : commission} نقطة'
                                                : 'Commission: ${commission.isEmpty ? "0" : commission} pts',
                                            style: const TextStyle(
                                              color: Colors.white,
                                              fontSize: 14,
                                              fontWeight: FontWeight.bold,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    const SizedBox(height: 16),
                                    // Action Button (Demo)
                                    Container(
                                      width: double.infinity,
                                      padding: const EdgeInsets.symmetric(vertical: 14),
                                      decoration: BoxDecoration(
                                        color: Colors.white,
                                        borderRadius: BorderRadius.circular(30),
                                      ),
                                      child: Row(
                                        mainAxisAlignment: MainAxisAlignment.center,
                                        children: [
                                          Icon(Icons.add_circle_outline, color: Colors.grey.shade800),
                                          const SizedBox(width: 8),
                                          Text(
                                            isArabic ? 'سجّل وأضف العرض' : 'Sign up & Add Offer',
                                            style: TextStyle(
                                              color: Colors.grey.shade800,
                                              fontWeight: FontWeight.bold,
                                              fontSize: 16,
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
                      // Info
                      Padding(
                        padding: const EdgeInsets.all(16),
                        child: Text(
                          isArabic 
                              ? '👆 هكذا سيظهر عرضك للمروجين'
                              : '👆 This is how your offer will appear to promoters',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.7),
                            fontSize: 14,
                          ),
                          textAlign: TextAlign.center,
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

  void _showFullAgreement(BuildContext context, bool isArabic) {
    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: const Color(0xFF1A1A2E),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        child: Container(
          constraints: const BoxConstraints(maxWidth: 500, maxHeight: 600),
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                children: [
                  const Icon(Icons.description, color: Color(0xFF6C63FF)),
                  const SizedBox(width: 12),
                  Text(
                    isArabic ? 'اتفاقية المعلنين - AffTok' : 'Advertiser Agreement - AffTok',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 20),
              Expanded(
                child: SingleChildScrollView(
                  child: Text(
                    isArabic ? '''
نسبة المنصّة
تحصل منصّة AffTok على نسبة قدرها 10% من قيمة العمولات التي يقوم المعلن بدفعها للمروجين، ويُعد هذا الاتفاق ماليًا داخليًا بين الطرفين وغير معلن للمستخدمين أو أطراف خارجية.

التزامات المعلن
• دفع عمولات المروجين مباشرة دون أي تدخل مالي أو تنفيذي من المنصّة.
• سداد نسبة المنصّة وفق الفواتير الشهرية الصادرة عنها.
• تزويد المنصّة بإثبات الدفع خلال المدة المحددة عند الطلب.

التزامات المنصّة
• توفير نظام تتبّع دقيق وآمن للعروض والارتباطات.
• إصدار فواتير شهرية واضحة بنسبة 10% المتفق عليها.
• عدم استلام أو توزيع أو تحويل أي مبالغ تخص المروجين، وتبقى العلاقة المالية الخاصة بالعمولات مباشرة بين المعلن والمروج.

آلية الدفع والمواعيد
• تُصدر الفاتورة الشهرية في اليوم الأول من كل شهر ميلادي.
• يلتزم المعلن بالسداد خلال 7 أيام عمل من تاريخ إصدار الفاتورة.
• في حال كان المبلغ المستحق أقل من 10 د.ك، يُرحَّل للشهر التالي.

وسيلة الدفع
تتم عملية السداد الخاصة بنسبة المنصّة عبر تحويل بنكي مباشر لحساب الشركة، وتُزوَّد بيانات الحساب للمعلن عند بدء تفعيل الاتفاق.

الإنهاء
• يحق لأي طرف إنهاء هذه الاتفاقية بإشعار خطي مسبق مدته 30 يومًا.
• تُسدَّد جميع المستحقات المالية قبل سريان الإنهاء.
''' : '''
Platform Commission
AffTok platform receives 10% of the commissions paid by the advertiser to promoters. This is an internal financial agreement between both parties and is not disclosed to users or external parties.

Advertiser Obligations
• Pay promoter commissions directly without any financial or operational intervention from the platform.
• Pay the platform's share according to the monthly invoices issued.
• Provide payment proof within the specified period upon request.

Platform Obligations
• Provide an accurate and secure tracking system for offers and links.
• Issue clear monthly invoices for the agreed 10%.
• Not receive, distribute, or transfer any amounts related to promoters; the financial relationship regarding commissions remains directly between the advertiser and the promoter.

Payment Schedule
• Monthly invoices are issued on the first day of each calendar month.
• The advertiser must pay within 7 business days from the invoice date.
• If the amount due is less than 10 KWD, it will be carried over to the next month.

Payment Method
Platform commission payments are made via direct bank transfer to the company account. Account details are provided to the advertiser upon agreement activation.

Termination
• Either party may terminate this agreement with 30 days written notice.
• All outstanding amounts must be paid before termination takes effect.
''',
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.9),
                      fontSize: 14,
                      height: 1.6,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => Navigator.pop(context),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF6C63FF),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: Text(
                    isArabic ? 'إغلاق' : 'Close',
                    style: const TextStyle(fontSize: 16, color: Colors.white),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';
    final isEditing = widget.existingOffer != null;

    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          // Background "AffTok" shadow text
          Positioned.fill(
            child: Center(
              child: Transform.rotate(
                angle: -0.2,
                child: Text(
                  'AffTok',
                  style: TextStyle(
                    fontSize: 120,
                    fontWeight: FontWeight.w900,
                    color: Colors.white.withOpacity(0.03),
                    letterSpacing: 8,
                  ),
                ),
              ),
            ),
          ),
          Positioned(
            top: MediaQuery.of(context).size.height * 0.1,
            left: -50,
            child: Text(
              'AffTok',
              style: TextStyle(
                fontSize: 80,
                fontWeight: FontWeight.w900,
                color: Colors.white.withOpacity(0.02),
                letterSpacing: 4,
              ),
            ),
          ),
          Positioned(
            bottom: MediaQuery.of(context).size.height * 0.05,
            right: -30,
            child: Text(
              'AffTok',
              style: TextStyle(
                fontSize: 60,
                fontWeight: FontWeight.w900,
                color: Colors.white.withOpacity(0.02),
                letterSpacing: 4,
              ),
            ),
          ),
          // Main Content
          SafeArea(
            child: Column(
            children: [
              // Header
              Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    IconButton(
                      onPressed: () => Navigator.pop(context),
                      icon: const Icon(Icons.arrow_back_ios, color: Colors.white),
                    ),
                    Expanded(
                      child: Text(
                        isEditing
                          ? (isArabic ? 'تعديل العرض' : 'Edit Offer')
                          : (isArabic ? 'إضافة عرض جديد' : 'Add New Offer'),
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ),
                    const SizedBox(width: 48), // Balance the back button
                  ],
                ),
              ),
              
              // Form
              Expanded(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: Form(
                    key: _formKey,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Info Banner
                        Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: const Color(0xFF6C63FF).withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: const Color(0xFF6C63FF).withOpacity(0.3),
                            ),
                          ),
                          child: Row(
                            children: [
                              const Icon(Icons.info_outline, color: Color(0xFF6C63FF), size: 20),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Text(
                                  isArabic
                                    ? 'سيتم مراجعة العرض قبل نشره للمروجين'
                                    : 'Your offer will be reviewed before being published',
                                  style: const TextStyle(
                                    color: Color(0xFF6C63FF),
                                    fontSize: 12,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                        
                        const SizedBox(height: 24),
                        
                        // English Section
                        _buildSectionTitle(isArabic ? 'المعلومات بالإنجليزية' : 'English Information', Icons.language),
                        
                        const SizedBox(height: 16),
                        _buildTextField(
                          controller: _titleController,
                          label: isArabic ? 'عنوان العرض (إنجليزي) *' : 'Offer Title (English) *',
                          icon: Icons.title,
                          validator: (v) => v!.isEmpty ? (isArabic ? 'مطلوب' : 'Required') : null,
                        ),
                        
                        const SizedBox(height: 16),
                        _buildTextField(
                          controller: _descriptionController,
                          label: isArabic ? 'الوصف (إنجليزي)' : 'Description (English)',
                          icon: Icons.description,
                          maxLines: 3,
                        ),
                        
                        const SizedBox(height: 24),
                        
                        // Arabic Section
                        _buildSectionTitle(isArabic ? 'المعلومات بالعربية' : 'Arabic Information', Icons.translate),
                        
                        const SizedBox(height: 16),
                        _buildTextField(
                          controller: _titleArController,
                          label: isArabic ? 'عنوان العرض (عربي)' : 'Offer Title (Arabic)',
                          icon: Icons.title,
                        ),
                        
                        const SizedBox(height: 16),
                        _buildTextField(
                          controller: _descriptionArController,
                          label: isArabic ? 'الوصف (عربي)' : 'Description (Arabic)',
                          icon: Icons.description,
                          maxLines: 3,
                        ),
                        
                        const SizedBox(height: 16),
                        _buildTextField(
                          controller: _termsArController,
                          label: isArabic ? 'الشروط والأحكام (عربي)' : 'Terms & Conditions (Arabic)',
                          icon: Icons.gavel,
                          maxLines: 3,
                        ),
                        
                        const SizedBox(height: 24),
                        
                        // URLs Section
                        _buildSectionTitle(isArabic ? 'الروابط والصور' : 'URLs & Images', Icons.link),
                        
                        const SizedBox(height: 16),
                        _buildTextField(
                          controller: _destinationUrlController,
                          label: isArabic ? 'رابط العرض *' : 'Destination URL *',
                          icon: Icons.open_in_new,
                          keyboardType: TextInputType.url,
                          validator: (v) {
                            if (v!.isEmpty) return isArabic ? 'مطلوب' : 'Required';
                            if (!v.startsWith('http')) return isArabic ? 'رابط غير صالح' : 'Invalid URL';
                            return null;
                          },
                        ),
                        
                        const SizedBox(height: 16),
                        _buildImagePicker(
                          label: isArabic ? 'صورة العرض' : 'Offer Image',
                          urlController: _imageUrlController,
                          selectedImage: _selectedImage,
                          isUploading: _isUploadingImage,
                          onPick: () => _pickAndUploadImage(isLogo: false),
                          isArabic: isArabic,
                        ),
                        
                        const SizedBox(height: 16),
                        _buildImagePicker(
                          label: isArabic ? 'الشعار' : 'Logo',
                          urlController: _logoUrlController,
                          selectedImage: _selectedLogo,
                          isUploading: _isUploadingLogo,
                          onPick: () => _pickAndUploadImage(isLogo: true),
                          isArabic: isArabic,
                        ),
                        
                        const SizedBox(height: 24),
                        
                        // Category & Payout Section
                        _buildSectionTitle(isArabic ? 'التصنيف والعمولة' : 'Category & Payout', Icons.category),
                        
                        const SizedBox(height: 16),
                        _buildDropdown(
                          label: isArabic ? 'التصنيف' : 'Category',
                          value: _selectedCategory,
                          items: _categories.map((c) => DropdownMenuItem(
                            value: c['value'],
                            child: Text(isArabic ? c['ar']! : c['en']!),
                          )).toList(),
                          onChanged: (v) => setState(() => _selectedCategory = v!),
                        ),
                        
                        const SizedBox(height: 16),
                        _buildDropdown(
                          label: isArabic ? 'نوع الدفع' : 'Payout Type',
                          value: _selectedPayoutType,
                          items: _payoutTypes.map((p) => DropdownMenuItem(
                            value: p['value'],
                            child: Text(isArabic ? p['ar']! : p['en']!),
                          )).toList(),
                          onChanged: (v) => setState(() => _selectedPayoutType = v!),
                        ),
                        
                        const SizedBox(height: 16),
                        Row(
                          children: [
                            Expanded(
                              child: _buildTextField(
                                controller: _payoutController,
                                label: isArabic ? 'المبلغ للمعلن' : 'Payout',
                                icon: Icons.attach_money,
                                keyboardType: TextInputType.number,
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: _buildTextField(
                                controller: _commissionController,
                                label: isArabic ? 'عمولة المروج' : 'Commission',
                                icon: Icons.percent,
                                keyboardType: TextInputType.number,
                              ),
                            ),
                          ],
                        ),
                        
                        const SizedBox(height: 24),
                        
                        // Platform Terms Agreement
                        Container(
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: Colors.white.withOpacity(0.05),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: Colors.white.withOpacity(0.1)),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Icon(Icons.handshake, color: const Color(0xFF6C63FF), size: 20),
                                  const SizedBox(width: 8),
                                  Text(
                                    isArabic ? 'اتفاقية المنصة' : 'Platform Agreement',
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 12),
                              Text(
                                isArabic
                                  ? 'بإرسال هذا العرض، أوافق على أن منصة AffTok تحصل على نسبة 10% من قيمة العمولات المدفوعة للمروجين. هذا الاتفاق مالي داخلي بين الطرفين.'
                                  : 'By submitting this offer, I agree that AffTok platform receives 10% of the commissions paid to promoters. This is an internal financial agreement between both parties.',
                                style: TextStyle(
                                  color: Colors.white.withOpacity(0.7),
                                  fontSize: 13,
                                  height: 1.5,
                                ),
                              ),
                              const SizedBox(height: 8),
                              GestureDetector(
                                onTap: () => _showFullAgreement(context, isArabic),
                                child: Text(
                                  isArabic ? 'عرض الاتفاقية الكاملة ←' : 'View full agreement →',
                                  style: const TextStyle(
                                    color: Color(0xFF6C63FF),
                                    fontSize: 13,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ),
                              const SizedBox(height: 12),
                              Row(
                                children: [
                                  SizedBox(
                                    width: 24,
                                    height: 24,
                                    child: Checkbox(
                                      value: _agreedToTerms,
                                      onChanged: (value) {
                                        setState(() => _agreedToTerms = value ?? false);
                                      },
                                      activeColor: const Color(0xFF6C63FF),
                                      shape: RoundedRectangleBorder(
                                        borderRadius: BorderRadius.circular(4),
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Expanded(
                                    child: Text(
                                      isArabic
                                        ? 'أوافق على شروط وأحكام المنصة'
                                        : 'I agree to the platform terms and conditions',
                                      style: TextStyle(
                                        color: Colors.white.withOpacity(0.9),
                                        fontSize: 13,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),
                        
                        const SizedBox(height: 24),
                        
                        // Preview Button
                        SizedBox(
                          width: double.infinity,
                          height: 50,
                          child: OutlinedButton.icon(
                            onPressed: () => _showOfferPreview(context, isArabic),
                            icon: const Icon(Icons.visibility, color: Color(0xFF6C63FF)),
                            label: Text(
                              isArabic ? 'معاينة العرض' : 'Preview Offer',
                              style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                                color: Color(0xFF6C63FF),
                              ),
                            ),
                            style: OutlinedButton.styleFrom(
                              side: const BorderSide(color: Color(0xFF6C63FF), width: 2),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(16),
                              ),
                            ),
                          ),
                        ),
                        
                        const SizedBox(height: 16),
                        
                        // Submit Button
                        SizedBox(
                          width: double.infinity,
                          height: 56,
                          child: ElevatedButton(
                            onPressed: (_isLoading || !_agreedToTerms) ? null : _submit,
                            style: ElevatedButton.styleFrom(
                              backgroundColor: _agreedToTerms ? const Color(0xFF6C63FF) : Colors.grey,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(16),
                              ),
                            ),
                            child: _isLoading
                              ? const CircularProgressIndicator(color: Colors.white)
                              : Text(
                                  isEditing
                                    ? (isArabic ? 'تحديث العرض' : 'Update Offer')
                                    : (isArabic ? 'إرسال للمراجعة' : 'Submit for Review'),
                                  style: const TextStyle(
                                    fontSize: 18,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.white,
                                  ),
                                ),
                          ),
                        ),
                        
                        const SizedBox(height: 20),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionTitle(String title, IconData icon) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: const Color(0xFF6C63FF).withOpacity(0.2),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: const Color(0xFF6C63FF), size: 18),
        ),
        const SizedBox(width: 12),
        Text(
          title,
          style: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
      ],
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required IconData icon,
    int maxLines = 1,
    TextInputType? keyboardType,
    String? Function(String?)? validator,
  }) {
    return TextFormField(
      controller: controller,
      maxLines: maxLines,
      keyboardType: keyboardType,
      validator: validator,
      style: const TextStyle(color: Colors.white),
      decoration: InputDecoration(
        labelText: label,
        labelStyle: TextStyle(color: Colors.white.withOpacity(0.7)),
        prefixIcon: Icon(icon, color: const Color(0xFF6C63FF)),
        filled: true,
        fillColor: Colors.white.withOpacity(0.05),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.white.withOpacity(0.1)),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.white.withOpacity(0.1)),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF6C63FF)),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Colors.red),
        ),
      ),
    );
  }

  Widget _buildDropdown({
    required String label,
    required String value,
    required List<DropdownMenuItem<String>> items,
    required void Function(String?) onChanged,
  }) {
    return DropdownButtonFormField<String>(
      value: value,
      items: items,
      onChanged: onChanged,
      dropdownColor: const Color(0xFF1A1A2E),
      style: const TextStyle(color: Colors.white),
      decoration: InputDecoration(
        labelText: label,
        labelStyle: TextStyle(color: Colors.white.withOpacity(0.7)),
        filled: true,
        fillColor: Colors.white.withOpacity(0.05),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.white.withOpacity(0.1)),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.white.withOpacity(0.1)),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF6C63FF)),
        ),
      ),
    );
  }

  Widget _buildImagePicker({
    required String label,
    required TextEditingController urlController,
    required File? selectedImage,
    required bool isUploading,
    required VoidCallback onPick,
    required bool isArabic,
  }) {
    final hasImage = urlController.text.isNotEmpty || selectedImage != null;
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: TextStyle(
            color: Colors.white.withOpacity(0.7),
            fontSize: 14,
          ),
        ),
        const SizedBox(height: 8),
        GestureDetector(
          onTap: isUploading ? null : onPick,
          child: Container(
            height: 150,
            width: double.infinity,
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.05),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: hasImage 
                  ? const Color(0xFF6C63FF) 
                  : Colors.white.withOpacity(0.1),
                width: hasImage ? 2 : 1,
              ),
            ),
            child: isUploading
              ? const Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      CircularProgressIndicator(color: Color(0xFF6C63FF)),
                      SizedBox(height: 12),
                      Text(
                        'جاري الرفع...',
                        style: TextStyle(color: Colors.white70),
                      ),
                    ],
                  ),
                )
              : hasImage
                ? Stack(
                    children: [
                      // Show image
                      ClipRRect(
                        borderRadius: BorderRadius.circular(11),
                        child: selectedImage != null
                          ? Image.file(
                              selectedImage,
                              width: double.infinity,
                              height: 150,
                              fit: BoxFit.cover,
                            )
                          : Image.network(
                              urlController.text,
                              width: double.infinity,
                              height: 150,
                              fit: BoxFit.cover,
                              errorBuilder: (_, __, ___) => const Center(
                                child: Icon(Icons.broken_image, color: Colors.white54, size: 40),
                              ),
                            ),
                      ),
                      // Change button
                      Positioned(
                        top: 8,
                        right: 8,
                        child: Container(
                          decoration: BoxDecoration(
                            color: Colors.black.withOpacity(0.6),
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: IconButton(
                            icon: const Icon(Icons.edit, color: Colors.white, size: 20),
                            onPressed: onPick,
                            constraints: const BoxConstraints(
                              minWidth: 36,
                              minHeight: 36,
                            ),
                            padding: EdgeInsets.zero,
                          ),
                        ),
                      ),
                      // Success indicator
                      Positioned(
                        bottom: 8,
                        left: 8,
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                          decoration: BoxDecoration(
                            color: Colors.green.withOpacity(0.9),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Icon(Icons.check, color: Colors.white, size: 14),
                              const SizedBox(width: 4),
                              Text(
                                isArabic ? 'تم الرفع' : 'Uploaded',
                                style: const TextStyle(color: Colors.white, fontSize: 11),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ],
                  )
                : Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(
                        Icons.add_photo_alternate_outlined,
                        color: Color(0xFF6C63FF),
                        size: 40,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        isArabic ? 'اضغط لاختيار صورة' : 'Tap to select image',
                        style: const TextStyle(
                          color: Color(0xFF6C63FF),
                          fontSize: 14,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        isArabic ? 'من المحفوظات' : 'from gallery',
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.5),
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
          ),
        ),
      ],
    );
  }
}
