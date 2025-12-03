import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/auth_provider.dart';
import '../../services/advertiser_service.dart';

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
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              widget.existingOffer != null
                ? (isArabic ? 'تم تحديث العرض وإرساله للمراجعة' : 'Offer updated and sent for review')
                : (isArabic ? 'تم إنشاء العرض وإرساله للمراجعة' : 'Offer created and sent for review'),
            ),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, true);
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
                        _buildTextField(
                          controller: _imageUrlController,
                          label: isArabic ? 'رابط صورة العرض' : 'Offer Image URL',
                          icon: Icons.image,
                          keyboardType: TextInputType.url,
                        ),
                        
                        const SizedBox(height: 16),
                        _buildTextField(
                          controller: _logoUrlController,
                          label: isArabic ? 'رابط الشعار' : 'Logo URL',
                          icon: Icons.branding_watermark,
                          keyboardType: TextInputType.url,
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
                        
                        const SizedBox(height: 32),
                        
                        // Submit Button
                        SizedBox(
                          width: double.infinity,
                          height: 56,
                          child: ElevatedButton(
                            onPressed: _isLoading ? null : _submit,
                            style: ElevatedButton.styleFrom(
                              backgroundColor: const Color(0xFF6C63FF),
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
}
