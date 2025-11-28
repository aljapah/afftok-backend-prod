import 'package:flutter/material.dart';
import '../utils/app_localizations.dart';
import 'signin_screen.dart';

class SignUpScreen extends StatefulWidget {
  const SignUpScreen({Key? key}) : super(key: key);

  @override
  State<SignUpScreen> createState() => _SignUpScreenState();
}

class _SignUpScreenState extends State<SignUpScreen> {
  bool _obscurePassword = true;
  bool _obscureConfirmPassword = true;
  bool _agreedToTerms = false;
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();

  @override
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 20),
            
            // Title
            Text(
              lang.createAccount,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 32,
                fontWeight: FontWeight.bold,
              ),
            ),
            
            const SizedBox(height: 8),
            
            Text(
              lang.signUpToStart,
              style: TextStyle(
                color: Colors.white.withOpacity(0.6),
                fontSize: 16,
              ),
            ),
            
            const SizedBox(height: 40),
            
            // Full Name field
            TextField(
              controller: _nameController,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                labelText: lang.fullName,
                labelStyle: const TextStyle(color: Colors.white60),
                prefixIcon: const Icon(Icons.person_outline, color: Colors.white60),
                enabledBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Colors.white30),
                  borderRadius: BorderRadius.circular(12),
                ),
                focusedBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Color(0xFFFF006E), width: 2),
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
            
            const SizedBox(height: 20),
            
            // Email field
            TextField(
              controller: _emailController,
              style: const TextStyle(color: Colors.white),
              keyboardType: TextInputType.emailAddress,
              decoration: InputDecoration(
                labelText: lang.email,
                labelStyle: const TextStyle(color: Colors.white60),
                prefixIcon: const Icon(Icons.email_outlined, color: Colors.white60),
                enabledBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Colors.white30),
                  borderRadius: BorderRadius.circular(12),
                ),
                focusedBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Color(0xFFFF006E), width: 2),
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
            
            const SizedBox(height: 20),
            
            // Password field
            TextField(
              controller: _passwordController,
              style: const TextStyle(color: Colors.white),
              obscureText: _obscurePassword,
              decoration: InputDecoration(
                labelText: lang.password,
                labelStyle: const TextStyle(color: Colors.white60),
                prefixIcon: const Icon(Icons.lock_outline, color: Colors.white60),
                suffixIcon: IconButton(
                  icon: Icon(
                    _obscurePassword ? Icons.visibility_off : Icons.visibility,
                    color: Colors.white60,
                  ),
                  onPressed: () {
                    setState(() {
                      _obscurePassword = !_obscurePassword;
                    });
                  },
                ),
                enabledBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Colors.white30),
                  borderRadius: BorderRadius.circular(12),
                ),
                focusedBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Color(0xFFFF006E), width: 2),
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
            
            const SizedBox(height: 20),
            
            // Confirm Password field
            TextField(
              controller: _confirmPasswordController,
              style: const TextStyle(color: Colors.white),
              obscureText: _obscureConfirmPassword,
              decoration: InputDecoration(
                labelText: lang.confirmPassword,
                labelStyle: const TextStyle(color: Colors.white60),
                prefixIcon: const Icon(Icons.lock_outline, color: Colors.white60),
                suffixIcon: IconButton(
                  icon: Icon(
                    _obscureConfirmPassword ? Icons.visibility_off : Icons.visibility,
                    color: Colors.white60,
                  ),
                  onPressed: () {
                    setState(() {
                      _obscureConfirmPassword = !_obscureConfirmPassword;
                    });
                  },
                ),
                enabledBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Colors.white30),
                  borderRadius: BorderRadius.circular(12),
                ),
                focusedBorder: OutlineInputBorder(
                  borderSide: const BorderSide(color: Color(0xFFFF006E), width: 2),
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
            
            const SizedBox(height: 20),
            
            // Terms checkbox
            Row(
              children: [
                Checkbox(
                  value: _agreedToTerms,
                  onChanged: (value) {
                    setState(() {
                      _agreedToTerms = value ?? false;
                    });
                  },
                  activeColor: const Color(0xFFFF006E),
                  side: const BorderSide(color: Colors.white30),
                ),
                Expanded(
                  child: Text.rich(
                    TextSpan(
                      text: lang.agreeToTerms,
                      style: TextStyle(
                        color: Colors.white.withOpacity(0.6),
                        fontSize: 14,
                      ),
                      children: [
                        TextSpan(
                          text: ' ${lang.termsOfService}',
                          style: const TextStyle(
                            color: Color(0xFFFF006E),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        TextSpan(
                          text: ' ${lang.and} ',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.6),
                          ),
                        ),
                        TextSpan(
                          text: lang.privacyPolicy,
                          style: const TextStyle(
                            color: Color(0xFFFF006E),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
            
            const SizedBox(height: 24),
            
            // Sign Up Button
            SizedBox(
              width: double.infinity,
              height: 56,
              child: ElevatedButton(
                onPressed: _agreedToTerms ? () {} : null,
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.transparent,
                  shadowColor: Colors.transparent,
                  disabledBackgroundColor: Colors.grey[800],
                  padding: EdgeInsets.zero,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(28),
                  ),
                ),
                child: _agreedToTerms
                    ? Ink(
                        decoration: BoxDecoration(
                          gradient: const LinearGradient(
                            colors: [Color(0xFFFF006E), Color(0xFFFF4D00)],
                          ),
                          borderRadius: BorderRadius.circular(28),
                          boxShadow: [
                            BoxShadow(
                              color: const Color(0xFFFF006E).withOpacity(0.4),
                              blurRadius: 16,
                              offset: const Offset(0, 8),
                            ),
                          ],
                        ),
                        child: Container(
                          alignment: Alignment.center,
                          child: Text(
                            lang.createAccount,
                            style: const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                            ),
                          ),
                        ),
                      )
                    : Container(
                        alignment: Alignment.center,
                        child: Text(
                          lang.createAccount,
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                            color: Colors.white.withOpacity(0.5),
                          ),
                        ),
                      ),
              ),
            ),
            
            const SizedBox(height: 32),
            
            // Divider
            Row(
              children: [
                Expanded(child: Divider(color: Colors.white.withOpacity(0.2))),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Text(
                    lang.orContinueWith,
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.6),
                      fontSize: 14,
                    ),
                  ),
                ),
                Expanded(child: Divider(color: Colors.white.withOpacity(0.2))),
              ],
            ),
            
            const SizedBox(height: 32),
            
            // Social signup buttons
            Row(
              children: [
                Expanded(
                  child: _SocialButton(
                    icon: Icons.g_mobiledata,
                    label: 'Google',
                    onPressed: () {},
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: _SocialButton(
                    icon: Icons.apple,
                    label: 'Apple',
                    onPressed: () {},
                  ),
                ),
              ],
            ),
            
            const SizedBox(height: 32),
            
            // Sign in link
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  lang.alreadyHaveAccount,
                  style: TextStyle(
                    color: Colors.white.withOpacity(0.6),
                    fontSize: 14,
                  ),
                ),
                TextButton(
                  onPressed: () {
                    Navigator.pushReplacement(
                      context,
                      MaterialPageRoute(builder: (context) => const SignInScreen()),
                    );
                  },
                  child: Text(
                    lang.signIn,
                    style: const TextStyle(
                      color: Color(0xFFFF006E),
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _SocialButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onPressed;

  const _SocialButton({
    Key? key,
    required this.icon,
    required this.label,
    required this.onPressed,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return OutlinedButton(
      onPressed: onPressed,
      style: OutlinedButton.styleFrom(
        side: const BorderSide(color: Colors.white30),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
        ),
        padding: const EdgeInsets.symmetric(vertical: 16),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, color: Colors.white, size: 24),
          const SizedBox(width: 8),
          Text(
            label,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 16,
            ),
          ),
        ],
      ),
    );
  }
}

