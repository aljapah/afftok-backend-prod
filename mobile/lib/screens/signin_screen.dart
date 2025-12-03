import 'package:flutter/material.dart';
import '../utils/app_localizations.dart';
import 'signup_screen.dart';
import 'home_feed_screen.dart';
import '../services/auth_service.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

class SignInScreen extends StatefulWidget {
  final bool isNewUser;
  
  const SignInScreen({Key? key, this.isNewUser = false}) : super(key: key);

  @override
  State<SignInScreen> createState() => _SignInScreenState();
}

class _SignInScreenState extends State<SignInScreen> {
  final _formKey = GlobalKey<FormState>();
  bool _obscurePassword = true;
  bool _rememberMe = false;
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _authService = AuthService();
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadSavedCredentials();
    _checkAutoLogin();
  }

  void _loadSavedCredentials() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final savedEmail = prefs.getString('saved_email');
      final savedPassword = prefs.getString('saved_password');
      final rememberMe = prefs.getBool('remember_me') ?? false;

      if (savedEmail != null && savedPassword != null && mounted) {
        setState(() {
          _emailController.text = savedEmail;
          _passwordController.text = savedPassword;
          _rememberMe = rememberMe;
        });
      }
    } catch (e) {
      print('Error loading saved credentials: $e');
    }
  }

  void _checkAutoLogin() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final rememberMe = prefs.getBool('remember_me') ?? false;
      
      if (rememberMe && mounted) {
        final savedEmail = prefs.getString('saved_email');
        final savedPassword = prefs.getString('saved_password');
        
        if (savedEmail != null && savedPassword != null) {
          setState(() {
            _isLoading = true;
          });
          
          try {
            await _authService.login(savedEmail, savedPassword);
            
            if (mounted) {
              await Provider.of<AuthProvider>(context, listen: false).loadCurrentUser();
              Navigator.pushAndRemoveUntil(
                context,
                MaterialPageRoute(builder: (context) => const HomeFeedScreen()),
                (route) => false, // Remove all previous routes
              );
            }
          } catch (e) {
            if (mounted) {
              setState(() {
                _isLoading = false;
              });
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text('Auto-login failed: ${e.toString().replaceFirst("Exception: ", "")}'),
                  backgroundColor: Colors.redAccent,
                ),
              );
            }
          }
        }
      }
    } catch (e) {
      print('Error checking auto-login: $e');
    }
  }

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _signIn() async {
    if (_formKey.currentState!.validate()) {
      setState(() {
        _isLoading = true;
      });
      try {
        await _authService.login(
          _emailController.text.trim(),
          _passwordController.text.trim(),
        );
        
        final prefs = await SharedPreferences.getInstance();
        if (_rememberMe) {
          await prefs.setString('saved_email', _emailController.text.trim());
          await prefs.setString('saved_password', _passwordController.text.trim());
          await prefs.setBool('remember_me', true);
        } else {
          await prefs.remove('saved_email');
          await prefs.remove('saved_password');
          await prefs.setBool('remember_me', false);
        }
        
        if (mounted) {
          await Provider.of<AuthProvider>(context, listen: false).loadCurrentUser();
          Navigator.pushAndRemoveUntil(
            context,
            MaterialPageRoute(builder: (context) => const HomeFeedScreen()),
            (route) => false, // Remove all previous routes
          );
        }
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Login Failed: ${e.toString().replaceFirst("Exception: ", "")}'),
              backgroundColor: Colors.redAccent,
            ),
          );
        }
      } finally {
        if (mounted) {
          setState(() {
            _isLoading = false;
          });
        }
      }
    }
  }

  void _signInWithGoogle() async {
    setState(() {
      _isLoading = true;
    });
    try {
      await _authService.googleSignIn();
      if (mounted) {
        await Provider.of<AuthProvider>(context, listen: false).loadCurrentUser();
        Navigator.pushAndRemoveUntil(
          context,
          MaterialPageRoute(builder: (context) => const HomeFeedScreen()),
          (route) => false, // Remove all previous routes
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Google Sign-In Failed: ${e.toString().replaceFirst("Exception: ", "")}'),
            backgroundColor: Colors.redAccent,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 20),
              
              Text(
                lang.welcomeBack,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 32,
                  fontWeight: FontWeight.bold,
                ),
              ),
              
              const SizedBox(height: 8),
              
              Text(
                lang.signInToContinue,
                style: TextStyle(
                  color: Colors.white.withOpacity(0.6),
                  fontSize: 16,
                ),
              ),
              
              const SizedBox(height: 40),
              
              TextFormField(
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
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Please enter your email';
                  }
                  if (!RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$').hasMatch(value)) {
                    return 'Please enter a valid email';
                  }
                  return null;
                },
              ),
              
              const SizedBox(height: 20),
              
              TextFormField(
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
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Please enter your password';
                  }
                  return null;
                },
              ),
              
              const SizedBox(height: 12),
              
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Checkbox(
                        value: _rememberMe,
                        onChanged: (value) {
                          setState(() {
                            _rememberMe = value ?? false;
                          });
                        },
                        activeColor: const Color(0xFFFF006E),
                        checkColor: Colors.white,
                      ),
                      Text(
                        'Remember me',
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.8),
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                  TextButton(
                    onPressed: () {},
                    child: Text(
                      lang.forgotPassword,
                      style: const TextStyle(
                        color: Color(0xFFFF006E),
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                ],
              ),
              
              const SizedBox(height: 24),
              
              SizedBox(
                width: double.infinity,
                height: 56,
                child: ElevatedButton(
                  onPressed: _isLoading ? null : _signIn,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.transparent,
                    shadowColor: Colors.transparent,
                    padding: EdgeInsets.zero,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(28),
                    ),
                  ),
                  child: Ink(
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
                      child: _isLoading
                          ? const CircularProgressIndicator(color: Colors.white)
                          : Text(
                              lang.signIn,
                              style: const TextStyle(
                                fontSize: 18,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                    ),
                  ),
                ),
              ),
              
              const SizedBox(height: 32),
              
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
              
              Center(
                child: _SocialButton(
                  icon: Icons.g_mobiledata,
                  label: 'Google',
                  onPressed: _isLoading ? null : _signInWithGoogle,
                ),
              ),
              
              const SizedBox(height: 32),
              
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    lang.dontHaveAccount,
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.6),
                      fontSize: 14,
                    ),
                  ),
                  TextButton(
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(builder: (context) => const SignUpScreen()),
                      );
                    },
                    child: Text(
                      lang.signUp,
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
      ),
    );
  }
}

class _SocialButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback? onPressed;

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
        padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 32),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: Colors.white, size: 24),
          const SizedBox(width: 12),
          Text(
            label,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 16,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }
}
