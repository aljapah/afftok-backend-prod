import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../../providers/auth_provider.dart';
import 'advertiser_dashboard_screen.dart';

class AdvertiserRegisterScreen extends StatefulWidget {
  const AdvertiserRegisterScreen({super.key});

  @override
  State<AdvertiserRegisterScreen> createState() => _AdvertiserRegisterScreenState();
}

class _AdvertiserRegisterScreenState extends State<AdvertiserRegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _usernameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  final _fullNameController = TextEditingController();
  final _companyNameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _websiteController = TextEditingController();
  final _countryController = TextEditingController();
  
  bool _isLoading = false;
  bool _obscurePassword = true;
  bool _obscureConfirmPassword = true;
  bool _isLogin = true; // Start with login screen first
  bool _rememberMe = false;

  @override
  void initState() {
    super.initState();
    _loadSavedCredentials();
  }

  Future<void> _loadSavedCredentials() async {
    final prefs = await SharedPreferences.getInstance();
    final savedUsername = prefs.getString('advertiser_saved_username');
    final savedPassword = prefs.getString('advertiser_saved_password');
    final rememberMe = prefs.getBool('advertiser_remember_me') ?? false;

    if (rememberMe && savedUsername != null && savedPassword != null) {
      setState(() {
        _usernameController.text = savedUsername;
        _passwordController.text = savedPassword;
        _rememberMe = true;
      });
    }
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    _fullNameController.dispose();
    _companyNameController.dispose();
    _phoneController.dispose();
    _websiteController.dispose();
    _countryController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    final authProvider = Provider.of<AuthProvider>(context, listen: false);

    bool success;
    if (_isLogin) {
      success = await authProvider.login(
        username: _usernameController.text.trim(),
        password: _passwordController.text,
      );
    } else {
      success = await authProvider.registerAdvertiser(
        username: _usernameController.text.trim(),
        email: _emailController.text.trim(),
        password: _passwordController.text,
        fullName: _fullNameController.text.trim(),
        companyName: _companyNameController.text.trim(),
        phone: _phoneController.text.trim().isEmpty ? null : _phoneController.text.trim(),
        website: _websiteController.text.trim().isEmpty ? null : _websiteController.text.trim(),
        country: _countryController.text.trim().isEmpty ? null : _countryController.text.trim(),
      );
    }

    setState(() => _isLoading = false);

    if (success && mounted) {
      // Save credentials if remember me is checked
      final prefs = await SharedPreferences.getInstance();
      if (_rememberMe && _isLogin) {
        await prefs.setString('advertiser_saved_username', _usernameController.text.trim());
        await prefs.setString('advertiser_saved_password', _passwordController.text);
        await prefs.setBool('advertiser_remember_me', true);
      } else {
        await prefs.remove('advertiser_saved_username');
        await prefs.remove('advertiser_saved_password');
        await prefs.setBool('advertiser_remember_me', false);
      }
      
      Navigator.pushAndRemoveUntil(
        context,
        MaterialPageRoute(builder: (context) => const AdvertiserDashboardScreen()),
        (route) => false,
      );
    } else if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(authProvider.error ?? 'حدث خطأ'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';

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
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24.0),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Back Button
                    IconButton(
                      onPressed: () => Navigator.pop(context),
                      icon: const Icon(Icons.arrow_back_ios, color: Colors.white),
                    ),
                    
                    const SizedBox(height: 20),
                    
                    // Header
                    Center(
                      child: Column(
                        children: [
                          Container(
                            width: 80,
                            height: 80,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              gradient: const LinearGradient(
                                colors: [Color(0xFF6C63FF), Color(0xFF9D4EDD)],
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: const Color(0xFF6C63FF).withOpacity(0.4),
                                  blurRadius: 20,
                                  spreadRadius: 3,
                                ),
                              ],
                            ),
                            child: const Icon(
                              Icons.business_center_rounded,
                              color: Colors.white,
                              size: 40,
                            ),
                          ),
                          
                          const SizedBox(height: 16),
                          
                          Text(
                            _isLogin
                              ? (isArabic ? 'تسجيل دخول المعلن' : 'Advertiser Login')
                              : (isArabic ? 'تسجيل معلن جديد' : 'Advertiser Registration'),
                            style: const TextStyle(
                              fontSize: 24,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                            ),
                          ),
                          
                          const SizedBox(height: 8),
                          
                          Text(
                            isArabic 
                              ? 'أضف عروضك واحصل على مروجين مجاناً'
                              : 'Add your offers and get promoters for free',
                            style: TextStyle(
                              fontSize: 14,
                              color: Colors.white.withOpacity(0.7),
                            ),
                          ),
                        ],
                      ),
                    ),
                  
                  const SizedBox(height: 32),
                  
                  // Form Fields
                  _buildTextField(
                    controller: _usernameController,
                    label: isArabic ? 'اسم المستخدم' : 'Username',
                    icon: Icons.person_outline,
                    validator: (v) => v!.isEmpty ? (isArabic ? 'مطلوب' : 'Required') : null,
                  ),
                  
                  if (!_isLogin) ...[
                    const SizedBox(height: 16),
                    _buildTextField(
                      controller: _emailController,
                      label: isArabic ? 'البريد الإلكتروني' : 'Email',
                      icon: Icons.email_outlined,
                      keyboardType: TextInputType.emailAddress,
                      validator: (v) {
                        if (v!.isEmpty) return isArabic ? 'مطلوب' : 'Required';
                        if (!v.contains('@')) return isArabic ? 'بريد غير صالح' : 'Invalid email';
                        return null;
                      },
                    ),
                  ],
                  
                  const SizedBox(height: 16),
                  _buildTextField(
                    controller: _passwordController,
                    label: isArabic ? 'كلمة المرور' : 'Password',
                    icon: Icons.lock_outline,
                    obscureText: _obscurePassword,
                    suffixIcon: IconButton(
                      icon: Icon(
                        _obscurePassword ? Icons.visibility_off : Icons.visibility,
                        color: Colors.grey,
                      ),
                      onPressed: () => setState(() => _obscurePassword = !_obscurePassword),
                    ),
                    validator: (v) {
                      if (v!.isEmpty) return isArabic ? 'مطلوب' : 'Required';
                      if (v.length < 6) return isArabic ? '6 أحرف على الأقل' : 'Min 6 characters';
                      return null;
                    },
                  ),
                  
                  if (!_isLogin) ...[
                    const SizedBox(height: 16),
                    _buildTextField(
                      controller: _confirmPasswordController,
                      label: isArabic ? 'تأكيد كلمة المرور' : 'Confirm Password',
                      icon: Icons.lock_outline,
                      obscureText: _obscureConfirmPassword,
                      suffixIcon: IconButton(
                        icon: Icon(
                          _obscureConfirmPassword ? Icons.visibility_off : Icons.visibility,
                          color: Colors.grey,
                        ),
                        onPressed: () => setState(() => _obscureConfirmPassword = !_obscureConfirmPassword),
                      ),
                      validator: (v) {
                        if (v!.isEmpty) return isArabic ? 'مطلوب' : 'Required';
                        if (v != _passwordController.text) {
                          return isArabic ? 'كلمات المرور غير متطابقة' : 'Passwords do not match';
                        }
                        return null;
                      },
                    ),
                    
                    const SizedBox(height: 24),
                    
                    // Section: Personal Info
                    _buildSectionTitle(isArabic ? 'المعلومات الشخصية' : 'Personal Information'),
                    
                    const SizedBox(height: 16),
                    _buildTextField(
                      controller: _fullNameController,
                      label: isArabic ? 'الاسم الكامل' : 'Full Name',
                      icon: Icons.badge_outlined,
                    ),
                    
                    const SizedBox(height: 16),
                    _buildTextField(
                      controller: _phoneController,
                      label: isArabic ? 'رقم الهاتف' : 'Phone Number',
                      icon: Icons.phone_outlined,
                      keyboardType: TextInputType.phone,
                    ),
                    
                    const SizedBox(height: 24),
                    
                    // Section: Company Info
                    _buildSectionTitle(isArabic ? 'معلومات الشركة' : 'Company Information'),
                    
                    const SizedBox(height: 16),
                    _buildTextField(
                      controller: _companyNameController,
                      label: isArabic ? 'اسم الشركة *' : 'Company Name *',
                      icon: Icons.business_outlined,
                      validator: (v) => v!.isEmpty ? (isArabic ? 'مطلوب' : 'Required') : null,
                    ),
                    
                    const SizedBox(height: 16),
                    _buildTextField(
                      controller: _websiteController,
                      label: isArabic ? 'الموقع الإلكتروني' : 'Website',
                      icon: Icons.language_outlined,
                      keyboardType: TextInputType.url,
                    ),
                    
                    const SizedBox(height: 16),
                    _buildTextField(
                      controller: _countryController,
                      label: isArabic ? 'الدولة' : 'Country',
                      icon: Icons.location_on_outlined,
                    ),
                  ],
                  
                  // Remember Me (only for login)
                  if (_isLogin) ...[
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Checkbox(
                          value: _rememberMe,
                          onChanged: (value) {
                            setState(() => _rememberMe = value ?? false);
                          },
                          activeColor: const Color(0xFF6C63FF),
                          checkColor: Colors.white,
                          side: BorderSide(color: Colors.white.withOpacity(0.5)),
                        ),
                        Text(
                          isArabic ? 'تذكرني' : 'Remember me',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.8),
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ],
                  
                  const SizedBox(height: 24),
                  
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
                            _isLogin
                              ? (isArabic ? 'تسجيل الدخول' : 'Login')
                              : (isArabic ? 'إنشاء الحساب' : 'Create Account'),
                            style: const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                            ),
                          ),
                    ),
                  ),
                  
                  const SizedBox(height: 20),
                  
                  // Toggle Login/Register
                  Center(
                    child: TextButton(
                      onPressed: () => setState(() => _isLogin = !_isLogin),
                      child: Text(
                        _isLogin
                          ? (isArabic ? 'ليس لديك حساب؟ سجل الآن' : "Don't have an account? Register")
                          : (isArabic ? 'لديك حساب بالفعل؟ سجل دخول' : 'Already have an account? Login'),
                        style: const TextStyle(
                          color: Color(0xFF6C63FF),
                          fontSize: 14,
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
    );
  }

  Widget _buildSectionTitle(String title) {
    return Row(
      children: [
        Container(
          width: 4,
          height: 20,
          decoration: BoxDecoration(
            color: const Color(0xFF6C63FF),
            borderRadius: BorderRadius.circular(2),
          ),
        ),
        const SizedBox(width: 8),
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
    bool obscureText = false,
    Widget? suffixIcon,
    TextInputType? keyboardType,
    String? Function(String?)? validator,
  }) {
    return TextFormField(
      controller: controller,
      obscureText: obscureText,
      keyboardType: keyboardType,
      validator: validator,
      style: const TextStyle(color: Colors.white),
      decoration: InputDecoration(
        labelText: label,
        labelStyle: TextStyle(color: Colors.white.withOpacity(0.7)),
        prefixIcon: Icon(icon, color: const Color(0xFF6C63FF)),
        suffixIcon: suffixIcon,
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
}
