import 'package:flutter/material.dart';
import '../utils/app_localizations.dart';

class PrivacyScreen extends StatelessWidget {
  const PrivacyScreen({Key? key}) : super(key: key);

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
          isArabic ? 'سياسة الخصوصية' : 'Privacy Policy',
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
نحن في تطبيق AffTok نحرص على حماية خصوصية المستخدمين وسرية بياناتهم.

يتم جمع المعلومات الأساسية مثل البريد الإلكتروني أو الاسم فقط لتحسين تجربة المستخدم وتقديم العروض المناسبة.

لا نقوم ببيع أو مشاركة أي بيانات مع أطراف خارجية دون إذن المستخدم، ونستخدم تقنيات تشفير متقدمة لحماية بياناتك.

يمكنك في أي وقت طلب حذف بياناتك أو تعطيل حسابك نهائيًا من خلال التواصل مع فريق الدعم.

نلتزم بجميع معايير الخصوصية المعتمدة لضمان أمان وسرية بياناتك بشكل كامل.
'''
                : '''
At AffTok, we are committed to protecting our users’ privacy and keeping their information secure.

We collect only essential data, such as name or email, to improve user experience and provide relevant offers.

We never sell or share personal information with third parties without consent, and we use advanced encryption technologies to protect your data.

You can request to delete your data or deactivate your account at any time by contacting our support team.

We strictly comply with international privacy standards to ensure the highest level of data protection.
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
