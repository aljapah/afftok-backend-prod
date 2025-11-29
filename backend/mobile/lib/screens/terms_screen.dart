import 'package:flutter/material.dart';
import 'package:afftok/utils/app_localizations.dart';

class TermsScreen extends StatelessWidget {
  const TermsScreen({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: Text(lang.termsScreenTitle),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildSection(lang.effectiveDate, isSubtitle: true),
            const SizedBox(height: 20),
            
            _buildSection(lang.acceptanceOfTerms),
            _buildText(lang.acceptanceOfTermsDesc),
            
            _buildSection(lang.descriptionOfService),
            _buildText(lang.descriptionOfServiceDesc),
            
            _buildSection(lang.userAccounts),
            _buildText(lang.accountRequirements),
            _buildBullet(lang.ageRequirement),
            _buildBullet(lang.validEmail),
            _buildBullet(lang.complyLaws),
            _buildBullet(lang.legitimateUse),
            
            _buildSection(lang.referralProgramRules),
            _buildSubSection(lang.permittedActivities),
            _buildBullet(lang.shareLinks),
            _buildBullet(lang.promoteHonestly),
            _buildBullet(lang.complyPartnerTerms),
            
            _buildSubSection(lang.prohibitedActivities),
            _buildBullet(lang.noFakeAccounts),
            _buildBullet(lang.noBots),
            _buildBullet(lang.noSpamming),
            _buildBullet(lang.noMisrepresentation),
            _buildBullet(lang.noManipulation),
            _buildBullet(lang.noFraud),
            
            _buildSection(lang.earningsPayments),
            _buildText(lang.earningsDesc),
            _buildText(lang.taxResponsibility),
            
            _buildSection(lang.intellectualProperty),
            _buildText(lang.intellectualPropertyDesc),
            
            _buildSection(lang.thirdPartyServices),
            _buildText(lang.thirdPartyServicesDesc),
            
            _buildSection(lang.disclaimers),
            _buildText(lang.disclaimersDesc),
            
            _buildSection(lang.limitationOfLiability),
            _buildText(lang.limitationOfLiabilityDesc),
            
            _buildSection(lang.changesToTerms),
            _buildText(lang.changesToTermsDesc),
            
            _buildSection(lang.contactInformation),
            _buildText(lang.contactInformationDesc),
            _buildBullet(lang.emailLegal),
            _buildBullet(lang.websiteTerms),
            _buildBullet(lang.emailSupport),
            
            const SizedBox(height: 30),
            Center(
              child: Text(
                lang.slogan,
                style: const TextStyle(
                  color: Colors.white54,
                  fontSize: 12,
                  fontStyle: FontStyle.italic,
                ),
              ),
            ),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }

  Widget _buildSection(String title, {bool isSubtitle = false}) {
    return Padding(
      padding: const EdgeInsets.only(top: 20, bottom: 10),
      child: Text(
        title,
        style: TextStyle(
          fontSize: isSubtitle ? 14 : 18,
          fontWeight: FontWeight.bold,
          color: isSubtitle ? Colors.white70 : Colors.white,
        ),
      ),
    );
  }

  Widget _buildSubSection(String title) {
    return Padding(
      padding: const EdgeInsets.only(top: 12, bottom: 8),
      child: Text(
        title,
        style: const TextStyle(
          fontSize: 15,
          fontWeight: FontWeight.w600,
          color: Colors.white,
        ),
      ),
    );
  }

  Widget _buildText(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Text(
        text,
        style: const TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.bold,
          color: Colors.white70,
          height: 1.5,
        ),
      ),
    );
  }

  Widget _buildBullet(String text) {
    return Padding(
      padding: const EdgeInsets.only(left: 16, bottom: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '• ',
            style: TextStyle(
              fontSize: 14,
              color: Color(0xFFE91E63),
              fontWeight: FontWeight.bold,
            ),
          ),
          Expanded(
            child: Text(
              text,
              style: const TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.bold,
                color: Colors.white70,
                height: 1.5,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
