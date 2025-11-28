import 'package:flutter/material.dart';
import '../utils/app_localizations.dart';

class SecurityScreen extends StatelessWidget {
  const SecurityScreen({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    final isArabic = lang.locale.languageCode == 'ar';

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        title: Text(
          isArabic ? 'إعدادات الأمان' : 'Security Settings',
          style: const TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.bold,
            fontSize: 22,
          ),
        ),
        iconTheme: const IconThemeData(color: Colors.white),
        centerTitle: true,
      ),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: SingleChildScrollView(
          child: Text(
            isArabic
                ? '''
نحن نأخذ أمان حسابك بجدية تامة.

يتم تشفير جميع البيانات أثناء الإرسال والتخزين لضمان الحماية القصوى.

قريبًا سنضيف ميزة التحقق بخطوتين (2FA) لتوفير تسجيل دخول آمن، مع إمكانية تغيير كلمة المرور وإدارة الجلسات النشطة.

نوصي دائمًا باستخدام كلمة مرور قوية وفريدة، وتجنب مشاركتها مع أي طرف ثالث.

كما ننفذ بشكل مستمر تدابير أمان محدثة للحفاظ على خصوصيتك وأمان بياناتك في جميع الأوقات.
'''
                : '''
We take your account security very seriously.

All data transmissions and storage are encrypted to ensure maximum protection.

Soon, we will introduce Two-Factor Authentication (2FA) for enhanced login security, along with options to change your password and manage active sessions.

We always recommend using a strong and unique password and never sharing it with anyone.

We also continuously implement advanced security measures to maintain your privacy and data protection at all times.
''',
            style: const TextStyle(
              color: Colors.white70,
              height: 1.7,
              fontSize: 16,
              fontWeight: FontWeight.w400,
            ),
            textAlign: isArabic ? TextAlign.right : TextAlign.left,
          ),
        ),
      ),
    );
  }
}
