import 'package:flutter/material.dart';
import 'package:afftok/utils/app_localizations.dart';

class PrivacyPolicyScreen extends StatelessWidget {
  const PrivacyPolicyScreen({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final lang = AppLocalizations.of(context);
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: Text(lang.privacyPolicyScreenTitle),
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
            _buildSection(lang.introduction),
            _buildText(lang.introductionDesc),
            
            _buildSection(lang.informationWeCollect),
            _buildSubSection(lang.personalInformation),
            _buildBullet(lang.accountInformation),
            _buildBullet(lang.referralData),
            _buildBullet(lang.performanceMetrics),
            
            _buildSubSection(lang.automaticallyCollectedInfo),
            _buildBullet(lang.deviceInformation),
            _buildBullet(lang.usageData),
            _buildBullet(lang.locationData),
            
            _buildSection(lang.howWeUseInfo),
            _buildText(lang.useInfoDesc),
            _buildBullet(lang.provideService),
            _buildBullet(lang.trackReferrals),
            _buildBullet(lang.personalizeOffers),
            _buildBullet(lang.communicateUpdates),
            _buildBullet(lang.improveApp),
            _buildBullet(lang.preventFraud),
            
            _buildSection(lang.informationSharing),
            _buildText(lang.informationSharingDesc),
            _buildBullet(lang.partnerCompaniesShare),
            _buildBullet(lang.serviceProviders),
            _buildBullet(lang.legalRequirements),
            
            _buildSection(lang.dataSecurity),
            _buildText(lang.dataSecurityDesc),
            _buildBullet(lang.encryptedData),
            _buildBullet(lang.secureServers),
            _buildBullet(lang.securityAudits),
            _buildBullet(lang.accessControls),
            
            _buildSection(lang.yourRights),
            _buildText(lang.yourRightsDesc),
            _buildBullet(lang.accessData),
            _buildBullet(lang.correctData),
            _buildBullet(lang.requestDeletion),
            _buildBullet(lang.optOutMarketing),
            _buildBullet(lang.exportData),
            
            _buildSection(lang.contactUsPrivacy),
            _buildText(lang.contactUsPrivacyDesc),
            _buildBullet(lang.emailPrivacy),
            
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
          fontSize: isSubtitle ? 14 : 20,
          fontWeight: FontWeight.bold,
          color: isSubtitle ? Colors.white70 : Colors.white,
        ),
      ),
    );
  }

  Widget _buildSubSection(String title) {
    return Padding(
      padding: const EdgeInsets.only(top: 15, bottom: 8),
      child: Text(
        title,
        style: const TextStyle(
          fontSize: 16,
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
