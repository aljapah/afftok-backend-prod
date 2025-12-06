import 'package:flutter/material.dart';

class AppLocalizations {
  final Locale locale;

  AppLocalizations(this.locale);

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  static final Map<String, Map<String, String>> _localizedValues = {
    'en': {
      // General
      'appName': 'AffTok',
      'tagline': 'Earn. Share. Repeat.',
      'skip': 'Skip',
      
      // Bottom Navigation
      'home': 'Home',
      'discover': 'Discover',
      'saved': 'Saved',
      'profile': 'Profile',
      
      // Home Screen
      'getReferralLink': 'Get Referral Link',
      'share': 'Share',
      'info': 'Info',
      'checkOut': 'Check out',
      'offerDetails': 'Offer Details',
      'reward': 'Reward',
      
      // Discover Screen
      'discoverTitle': 'Discover Offers',
      'all': 'All',
      'cashback': 'Cashback',
      'discounts': 'Discounts',
      'crypto': 'Crypto',
      'shopping': 'Shopping',
      'services': 'Services',
      'food': 'Food & Delivery',
      'travel': 'Travel',
      'finance': 'Finance',
      'entertainment': 'Entertainment',
      'education': 'Education',
      'noOffersFound': 'No offers found',
      'ecommerce': 'E-commerce',
      'technology': 'Technology',
      'utilities': 'Utilities',
      'foodRestaurants': 'Food & Restaurants',
      
      // Saved Screen
      'savedTitle': 'Saved Offers',
      'noSavedOffers': 'No saved offers yet',
      'startSaving': 'Start saving offers you like!',
      
      // Profile Screen
      'guestUser': 'Guest User',
      'earnings': 'Performance',
      'shared': 'Shared',
      'signIn': 'Sign In',
      'createAccount': 'Create Account',
      'referralCode': 'My Referral Code',
      'achievements': 'Achievements',
      'viewAll': 'View All',
      
      // Settings Screen
      'settings': 'Settings',
      'account': 'ACCOUNT',
      'editProfile': 'Edit Profile',
      'changeNamePhoto': 'Change your name and photo',
      'privacy': 'Privacy',
      'managePrivacy': 'Manage your privacy settings',
      'security': 'Security',
      'passwordAuth': 'Password and authentication',
      'preferences': 'PREFERENCES',
      'notifications': 'Notifications',
      'notificationsDesc': 'Get instant alerts when new referral offers are added',
      'notificationsSubtitle': 'Get instant alerts when new referral offers are added',
      'exclusiveOffers': 'Exclusive Offers Only',
      'exclusiveOffersDesc': 'Get notified about premium offers only',
      'exclusiveOffersSubtitle': 'Get notified about premium offers only',
      'darkMode': 'Dark Mode',
      'darkModeDesc': 'Always on (coming soon)',
      'language': 'Language',
      'selectPreferredCategories': 'Select your preferred categories',
      'customizeFeedSubtitle': 'Choose your favorite offer categories',
      'learnMoreAboutApp': 'Learn more about our app',
      'support': 'SUPPORT',
      'helpCenter': 'Help Center',
      'helpCenterDesc': 'Get help and support',
      'helpCenterSubtitle': 'Get help and support',

      'aboutReftok': 'About Reftok',
      'version': 'Version 1.0.0',
      'termsPrivacy': 'Terms & Privacy',
      'readPolicies': 'Read our policies',
      'logout': 'Logout',
      'signOutAccount': 'Sign Out',
      'loggedOut': 'You have been logged out',
      'save': 'Save',
      'comingSoon': 'Coming in Reftok v2 🔒',
      
      // Edit Profile
      'editProfileTitle': 'Edit Profile',
      'fullName': 'Full Name',
      'saveChanges': 'Save Changes',
      
      // Customize Feed
      'customizeFeedTitle': 'Customize Your Feed',
      'selectCategories': 'Select categories you\'re interested in',
      'applyPreferences': 'Apply Preferences',
      
      // Language Dialog
      'selectLanguage': 'Select Language',
      'english': 'English',
      'arabic': 'Arabic',
      
      // About Dialog
      'aboutTitle': 'About Reftok',
      'aboutDesc': 'Discover the best referral and cashback offers from top companies worldwide.',
      'aboutFullText': 'Welcome to Reftok - Where Sharing Pays! 🎉\n\nReftok is your ultimate platform for discovering and sharing the best referral offers and cashback deals from top brands worldwide. We believe that recommendations from friends are powerful, and they should be rewarded!\n\nHow It Works:\nSimply browse through hundreds of exclusive offers from leading companies, share your unique referral links with friends and family, and watch your successful referrals increase. Every successful referral makes you eligible for a reward directly from the partner company!\n\nWhy Reftok?\n🌟 Access Great Rewards: Get rewards from partner companies for every friend who signs up through your link\n🌟 Exclusive Offers: Access deals you won\'t find anywhere else\n🚀 Easy to Use: TikTok-style interface makes discovering offers fun and addictive\n🌍 Global Brands: Partner with trusted companies like Amazon, Uber, Airbnb, and more\n📱 Instant Tracking: Monitor your referrals and performance in real-time\n🎁 Daily Updates: New offers added every day\n\nOur Mission:\nTo create a community where everyone benefits - you benefit by sharing, your friends save money, and brands reach new customers. It\'s a win-win-win!\n\nJoin thousands of users who are already benefiting from AffTok. Download now and start turning your social network into a powerful promotion tool!\n\nReftok - Perform. Share. Repeat. ✨',
      'close': 'Close',
      
      // Logout Dialog
      'logoutTitle': 'Logout',
      'logoutConfirm': 'Are you sure you want to logout?',
      'cancel': 'Cancel',
      
      // Sign In Screen
      'signInTitle': 'Welcome Back',
      'welcomeBack': 'Welcome Back',
      'signInToContinue': 'Sign in to continue tracking',
      'signInSubtitle': 'Sign in to continue tracking',
      'email': 'Email',
      'password': 'Password',
      'forgotPassword': 'Forgot Password?',
      'signInButton': 'Sign In',
      'orContinueWith': 'Or continue with',
      'google': 'Google',
      'apple': 'Apple',
      'dontHaveAccount': 'Don\'t have an account?',
      'signUpLink': 'Sign Up',
      
      // Sign Up Screen
      'signUpTitle': 'Create Account',
      'signUpSubtitle': 'Start sharing and tracking today',
      'confirmPassword': 'Confirm Password',
      'agreeTerms': 'I agree to the Terms & Privacy Policy',
      'agreeToTerms': 'I agree to the',
      'termsOfService': 'Terms of Service',
      'and': 'and',
      'privacyPolicy': 'Privacy Policy',
      'signUpButton': 'Create Account',
      'signUp': 'Sign Up',
      'signUpToStart': 'Sign up to start sharing and tracking',
      'alreadyHaveAccount': 'Already have an account?',
      'signInLink': 'Sign In',
      'confirmPassword': 'Confirm Password',
      
      // Offer Categories
      'categoryAll': 'All',
      'categoryCrypto': 'Crypto',
      'categoryShopping': 'Shopping',
      'categoryServices': 'Services',
      'categoryFood': 'Food',
      'categoryTravel': 'Travel',
      'categoryFinance': 'Finance',
      
      // Teams Feature
      'teams': 'Teams',
      'myTeam': 'My Team',
      'leaderboard': 'Leaderboard',
      'challenges': 'Challenges',
      'teamPerformance': 'Team Performance',
      'referrals': 'Referrals',
      'clicks': 'Clicks',
      'earned': 'Achieved',
      'thisMonth': 'This Month',
      'teamGoal': 'Team Goal',
      'completed': 'Completed',
      'teamMembers': 'Team Members',
      'you': 'You',
      'inviteMembers': 'Invite Members',
      'topTeamsThisMonth': 'Top Teams This Month',
      'rank': 'Rank',
      'yourTeam': 'Your Team',
      'activeChallenges': 'Active Challenges',
      'progress': 'Progress',
      'timeLeft': 'Time Left',
      'currentRank': 'Current Rank',
      'createTeam': 'Create Team',
      'joinTeam': 'Join Team',
      'noTeamYet': 'You are not in a team yet',
      'joinCreateTeam': 'Join or create a team to compete and perform together!',
      'joinOrCreateTeam': 'Join an existing team or create your own to collaborate with others!',
      'browseTeams': 'Browse Teams',
      'teamName': 'Team Name',
      'teamStats': 'Team Statistics',
      'members': 'Members',
      'points': 'Points',
      'noMembersYet': 'No members yet',
      'leaveTeam': 'Leave Team',
      'confirmLeaveTeam': 'Are you sure you want to leave this team?',
      'confirmJoinTeam': 'Do you want to join team',
      'teamCreatedSuccessfully': 'Team created successfully!',
      'failedToCreateTeam': 'Failed to create team',
      'joinedTeamSuccessfully': 'Joined team successfully!',
      'failedToJoinTeam': 'Failed to join team',
      'leftTeamSuccessfully': 'Left team successfully!',
      'failedToLeaveTeam': 'Failed to leave team',
      'noTeamsYet': 'No teams available yet',
      'createFirstTeam': 'Be the first to create a team!',
      'full': 'Full',
      'join': 'Join',
      'leave': 'Leave',
      'create': 'Create',
      'cancel': 'Cancel',
      
      // Enhanced Profile
      'personalLink': 'Personal Link',
      'yourPersonalLink': 'Your Personal Referral Page',
      'copyLink': 'Copy Link',
      'shareLink': 'Share Link',
      'viewPage': 'View Page',
      'linkCopied': 'Link copied to clipboard!',
      'stats': 'Stats',
      'totalClicks': 'Total Clicks',
      'totalReferrals': 'Total Referrals',
      'totalEarnings': 'Total Results',
      'level': 'Level',
      'badges': 'Badges',
      'noBadgesYet': 'No badges yet',
      'earnBadges': 'Complete challenges to earn badges!',
      'myOffers': 'My Offers',
      'activeOffers': 'Active Offers',
      'copy': 'Copy',
      'yourPerformance': 'Your Performance',
      'bestOffer': 'Best Offer',
      'members': 'Members',
      'myActiveOffers': 'My Active Offers',
      'addMoreOffers': 'Add More Offers',
      'viewAllOffers': 'View All Offers',
      'myPublicPage': 'My Page',
      
      // Missing translations - Added
      'conversions': 'Conversions',
      'addOffer': 'Add Offer',
      'analytics': 'Analytics',
      'offerAddedSuccessfully': 'Offer added successfully!',
      'failedToAddOffer': 'Failed to add offer. Please try again.',
      'addOfferToMyList': 'Add to My Offers',
      'offerAdded': 'Offer Added ✓',
      'offerAlreadyAdded': 'Offer already added to your list',
      'conversionRate': 'Conversion Rate',
      'totalConversions': 'Total Conversions',
      'youWillEarn': 'Partner Reward',
      'aboutThisOffer': 'About This Offer',
      'usersEarned': 'Users Achieved',
      'rating': 'Rating',
      'addOfferInstructions': 'Add Offer Instructions',
      'registerAtCompany': 'Register at Company',
      'overview': 'Overview',
      'performanceChart': 'Performance Chart',
      'chartPlaceholder': 'Chart Placeholder',
      'topPerformingOffers': 'Top Performing Offers',
      'additionalStats': 'Additional Stats',
      'globalRank': 'Global Rank',
      'monthlyEarnings': 'Monthly Performance',
      'growthRate': 'Growth Rate',
      'allTime': 'All Time',
      'thisWeek': 'This Week',
      'description': 'Description',
      'performance': 'Performance',
      'yourReferralLink': 'Your Referral Link',
      'shareOffer': 'Share Offer',
      'openLink': 'Open Link',
      'myLink': 'My Link',
      'registeredOffers': 'Registered Offers',
      
      // Additional missing keys
      'addedDate': 'Added Date',
      'addedToMyOffers': 'Added to My Offers',
      'done': 'Done',
      'error': 'Error',
      'getNotified': 'Get Notified',
      'helped': 'Helped',
      'iRegistered': 'I Registered',
      'invalidLink': 'Invalid Link',
      'linkStats': 'Link Stats',
      'offerAdded': 'Offer Added',
      'offers': 'Offers',
      'ok': 'OK',
      'redirectNotice': 'Redirect Notice',
      'registerAndAddOffer': 'Register and Add Offer',
      'saveLink': 'Save Link',
      'scanQRCode': 'Scan QR Code',
      'shares': 'Shares',
      'startNow': 'Start Now',
      'viewDetails': 'View Details',
      'visitOffer': 'Visit Offer',
      'visits': 'Visits',
      'addedToFavorites': 'Added to Favorites',
      'removedFromFavorites': 'Removed from Favorites',
      
      // Additional Missing Keys
      'continueAsGuest': 'Continue as Guest',
      'recommendedBy': 'Recommended by',
      'myPersonalLink': 'My Personal Link',
      'qrCodeDescription': 'Scan this QR code to share your personal link',
      'personalLinkDescription': 'Share this link to promote all your offers',
      'noOffersAdded': 'No offers added yet',
      'startAddingOffers': 'Start adding offers to track your performance',
      'exploreOffers': 'Explore Offers',
      'registerInProgram': 'Register in Program',
      'registeredInProgram': 'Registered in Program',
      'performanceOverview': 'Performance Overview',
      'viewDetailedAnalytics': 'View Detailed Analytics',
      'recommendedOffers': 'Recommended Offers',
      'poweredByAffTok': 'Powered by AffTok',
      'startEarningToo': 'Start Performing Too',
      'registrationOpened': 'Registration Opened',
      'registrationInstructions': 'Follow these steps to complete registration',
      'registrationTip': 'Make sure to use your referral link when signing up',
      'enterReferralLink': 'Enter Referral Link',
      'enterReferralLinkDescription': 'Paste your unique referral link from the company',
      'referralLinkWarning': 'Please make sure you entered the correct referral link',
      'offerAddedSuccessfully': 'Offer Added Successfully',
      'offerAddedDescription': 'You can now start sharing and performing',
      'couldNotOpenWebsite': 'Could not open website',
      'openingWebsite': 'Opening website...',
      
      // Download Prompt Screen
      'getTheBestExperience': 'Get the Best Experience',
      'downloadAppDescription': 'Download the AffTok app for the full experience',
      'instantAccess': 'Instant Access',
      'instantAccessDescription': 'Quick access to all offers',
      'trackEarnings': 'Track Performance',
      'trackEarningsDescription': 'Monitor your performance in real-time',
      'getNotifiedDescription': 'Get notified about new offers',
      'youreAboutToAccess': 'You\'re about to access',
      'downloadAffTok': 'Download AffTok',
      
      // Settings Screen Additional Getters
      'privacyPolicyTitle': 'Privacy Policy',
      'privacyPolicySubtitle': 'Read our data and privacy rules.',
      'termsOfServiceTitle': 'Terms of Service',
      'termsOfServiceSubtitle': 'Review our terms and user agreement.',
      'aboutAffTokTitle': 'About AffTok',
      'aboutAffTokSubtitle': 'Learn more about our mission and story.',
      
      // About Screen
      'aboutScreenTitle': 'About App',
      'welcomeToAffTok': 'Welcome to AffTok',
      'whereSharingPays': 'Where Sharing Pays! 🎉',
      'ourMission': 'Our Mission',
      'missionText': 'To democratize affiliate marketing and make it accessible to everyone. We believe that your recommendations have value, and you should be rewarded for sharing great products and services with your network.',
      'whatMakesAffTokSpecial': 'What Makes AffTok Special',
      'tikTokStyleDiscovery': 'TikTok-Style Discovery',
      'tikTokStyleDiscoveryDesc': 'Swipe through hundreds of exclusive offers in an addictive, visually stunning feed.',
      'realEarnings': 'Performance Tracking',
      'realEarningsDesc': 'Monitor your successful referrals and the rewards you are eligible for from partners.',
      'teamPower': 'Team Power',
      'teamPowerDesc': 'Join or create teams to compete, collaborate, and perform together.',
      'transparentTracking': 'Transparent Tracking',
      'transparentTrackingDesc': 'Monitor your clicks, referrals, and performance in real-time.',
      'globalOpportunities': 'Global Opportunities',
      'globalOpportunitiesDesc': 'Access offers from leading companies worldwide including Amazon, Uber, Airbnb, Binance.',
      'gamifiedExperience': 'Gamified Experience',
      'gamifiedExperienceDesc': 'Level up, earn badges, and compete in challenges.',
      'howItWorks': 'How It Works',
      'discoverStep': 'Discover',
      'discoverStepDesc': 'Browse through our curated feed of affiliate offers.',
      'shareStep': 'Share',
      'shareStepDesc': 'Get your unique referral link with one tap.',
      'earnStep': 'Get Results',
      'earnStepDesc': 'When someone signs up through your link, you become eligible for a reward from the partner company.',
      'growStep': 'Grow',
      'growStepDesc': 'Join teams, participate in challenges, and level up.',
      'theNumbers': 'The Numbers',
      'activeUsers': '500,000+ Active Users Worldwide',
      'paidOut': '\$50M+ Partner Rewards Facilitated',
      'partnerCompanies': '1,000+ Partner Companies',
      'countriesSupported': '150+ Countries Supported',
      'averageUserRating': '4.8/5 Average User Rating',
      // Stats Labels
      'activeUsersLabel': 'Active Users Worldwide',
      'paidOutLabel': 'Partner Rewards Facilitated',
      'partnerCompaniesLabel': 'Partner Companies',
      'countriesSupportedLabel': 'Countries Supported',
      'averageUserRatingLabel': 'Average User Rating',
      'contactUs': 'Contact Us',
      'emailHello': 'Email: hello@afftok.com',
      'emailSupport': 'Support: support@afftok.com',
      'website': 'Website: www.afftok.com',
      'instagram': 'Instagram: @afftok',
      'twitter': 'Twitter: @afftok',
      'versionText': 'Version 1.0.0',
      'slogan': 'AffTok - Perform. Share. Repeat.',
      
      // Terms Screen
      'termsScreenTitle': 'Terms of Service',
      'effectiveDate': 'Effective Date: October 16, 2025',
      'acceptanceOfTerms': 'Acceptance of Terms',
      'acceptanceOfTermsDesc': 'By accessing or using AffTok, you agree to be bound by these Terms of Service. If you do not agree, please discontinue use immediately.',
      'descriptionOfService': 'Description of Service',
      'descriptionOfServiceDesc': 'AffTok is a mobile platform that connects users with affiliate marketing opportunities and cashback offers from partner companies.',
      'userAccounts': 'User Accounts',
      'accountRequirements': 'Account Requirements:',
      'ageRequirement': 'Must be at least 18 years old',
      'validEmail': 'Provide valid email address',
      'complyLaws': 'Comply with all applicable laws',
      'legitimateUse': 'Use the service for legitimate purposes only',
      'referralProgramRules': 'Referral Program Rules',
      'permittedActivities': 'Permitted Activities',
      'shareLinks': 'Share your unique referral links through legitimate channels',
      'promoteHonestly': 'Promote offers honestly and accurately',
      'complyPartnerTerms': 'Comply with partner company terms and conditions',
      'prohibitedActivities': 'Prohibited Activities',
      'noFakeAccounts': 'Creating fake accounts or referrals',
      'noBots': 'Using automated bots or scripts',
      'noSpamming': 'Spamming or unsolicited marketing',
      'noMisrepresentation': 'Misrepresenting offers or rewards',
      'noManipulation': 'Manipulating clicks or conversions',
      'noFraud': 'Engaging in fraudulent behavior',
      'earningsPayments': 'Rewards and Eligibility',
      'earningsDesc': 'Rewards are tracked and eligibility is confirmed when referrals meet partner company requirements. Rewards are paid directly by the partner company. Processing times vary by partner and may take 30-90 days.',
      'taxResponsibility': 'You are responsible for reporting and paying taxes on compensation received from partner companies in accordance with local laws.',
      'intellectualProperty': 'Intellectual Property',
      'intellectualPropertyDesc': 'All content, trademarks, and intellectual property on AffTok are owned by us or our licensors. You may not copy, modify, or distribute our content without permission.',
      'thirdPartyServices': 'Third-Party Services',
      'thirdPartyServicesDesc': 'AffTok integrates with partner companies and third-party services. We are not responsible for their actions, products, or services.',
      'disclaimers': 'Disclaimers',
      'disclaimersDesc': 'We provide the service "as is" without warranties. Performance results vary and are not guaranteed. AffTok does not provide financial, investment, or legal advice.',
      'limitationOfLiability': 'Limitation of Liability',
      'limitationOfLiabilityDesc': 'To the maximum extent permitted by law, AffTok shall not be liable for indirect, incidental, or consequential damages.',
      'changesToTerms': 'Changes to Terms',
      'changesToTermsDesc': 'We may modify these terms at any time. Continued use after changes constitutes acceptance.',
      'contactInformation': 'Contact Information',
      'contactInformationDesc': 'For questions about these Terms of Service:',
      'emailLegal': 'Email: legal@afftok.com',
      'websiteTerms': 'Website: www.afftok.com/terms',
      
      // Privacy Policy Screen
      'privacyPolicyScreenTitle': 'Privacy Policy',
      'introduction': 'Introduction',
      'introductionDesc': 'Welcome to AffTok. We respect your privacy and are committed to protecting your personal data. This privacy policy explains how we collect, use, and safeguard your information when you use our mobile application.',
      'informationWeCollect': 'Information We Collect',
      'personalInformation': 'Personal Information',
      'accountInformation': 'Account Information: Name, email address, and profile picture',
      'referralData': 'Referral Data: Your referral links, clicks, and successful referrals',
      'performanceMetrics': 'Performance Metrics: Results tracking, team statistics, and challenge progress',
      'automaticallyCollectedInfo': 'Automatically Collected Information',
      'deviceInformation': 'Device Information: Device type, operating system, and unique device identifiers',
      'usageData': 'Usage Data: App features used, time spent, and interaction patterns',
      'locationData': 'Location Data: General location (country/city level) for offer personalization',
      'howWeUseInfo': 'How We Use Your Information',
      'useInfoDesc': 'We use your information to:',
      'provideService': 'Provide and maintain the AffTok service',
      'trackReferrals': 'Track your referrals and performance results',
      'personalizeOffers': 'Personalize offers based on your interests',
      'communicateUpdates': 'Communicate important updates and notifications',
      'improveApp': 'Improve app performance and user experience',
      'preventFraud': 'Prevent fraud and ensure platform security',
      'informationSharing': 'Information Sharing',
      'informationSharingDesc': 'We DO NOT sell your personal information. We may share data with:',
      'partnerCompaniesShare': 'Partner Companies: To track referrals and confirm eligibility for rewards',
      'serviceProviders': 'Service Providers: For analytics, hosting, and data processing',
      'legalRequirements': 'Legal Requirements: When required by law or to protect our rights',
      'dataSecurity': 'Data Security',
      'dataSecurityDesc': 'We implement industry-standard security measures including:',
      'encryptedData': 'Encrypted data transmission (SSL/TLS)',
      'secureServers': 'Secure server infrastructure',
      'securityAudits': 'Regular security audits',
      'accessControls': 'Access controls and authentication',
      'yourRights': 'Your Rights',
      'yourRightsDesc': 'You have the right to:',
      'accessData': 'Access and receive a copy of your personal data',
      'correctData': 'Correct or update inaccurate data',
      'deleteData': 'Request deletion of your personal data',
      'optOutMarketing': 'Opt-out of marketing communications',
      'dataRetention': 'Data Retention',
      'dataRetentionDesc': 'We retain your personal data only for as long as necessary to provide the services and fulfill the purposes outlined in this policy.',
      'childrensPrivacy': 'Children\'s Privacy',
      'childrensPrivacyDesc': 'AffTok is not intended for users under the age of 18. We do not knowingly collect personal data from children.',
      'contactPrivacy': 'Contact for Privacy Questions',
      'contactPrivacyDesc': 'For questions about this Privacy Policy or your data rights:',
      'emailPrivacy': 'Email: privacy@afftok.com',
      'websitePrivacy': 'Website: www.afftok.com/privacy',
      'requestDeletion': 'Request Data Deletion',
      'exportData': 'Export My Data',
      'contactUsPrivacy': 'Contact Us',
      'contactUsPrivacyDesc': 'For privacy questions, contact us at privacy@afftok.com',
    },
    'ar': {
      // General
      'appName': 'أف توك',
      'tagline': 'شارك. حقق. كرر.',
      'skip': 'تخطي',
      
      // Bottom Navigation
      'home': 'الرئيسية',
      'discover': 'اكتشف',
      'saved': 'المحفوظة',
      'profile': 'الملف الشخصي',
      
      // Home Screen
      'getReferralLink': 'احصل على رابط الإحالة',
      'share': 'مشاركة',
      'info': 'معلومات',
      'checkOut': 'شاهد العرض',
      'offerDetails': 'تفاصيل العرض',
      'reward': 'النتيجة',
      
      // Discover Screen
      'discoverTitle': 'اكتشف العروض',
      'all': 'الكل',
      'cashback': 'استرداد نقدي',
      'discounts': 'خصومات',
      'crypto': 'عملات رقمية',
      'shopping': 'تسوق',
      'services': 'خدمات',
      'food': 'طعام وتوصيل',
      'travel': 'سفر',
      'finance': 'مالية',
      'entertainment': 'ترفيه',
      'education': 'تعليم',
      'noOffersFound': 'لا توجد عروض',
      'ecommerce': 'تجارة إلكترونية',
      'technology': 'تكنولوجيا',
      'utilities': 'خدمات عامة',
      'foodRestaurants': 'مطاعم',

      // Saved Screen
      'savedTitle': 'العروض المحفوظة',
      'noSavedOffers': 'لم يتم حفظ أي عروض بعد',
      'startSaving': 'ابدأ بحفظ العروض التي تعجبك!',
      
      // Profile Screen
      'guestUser': 'مستخدم زائر',
      'earnings': 'الأداء',
      'shared': 'تمت المشاركة',
      'signIn': 'تسجيل الدخول',
      'createAccount': 'إنشاء حساب',
      'referralCode': 'رمز الإحالة الخاص بي',
      'achievements': 'الإنجازات',
      'viewAll': 'عرض الكل',
      
      // Settings Screen
      'settings': 'الإعدادات',
      'account': 'الحساب',
      'editProfile': 'تعديل الملف الشخصي',
      'changeNamePhoto': 'تغيير الاسم والصورة',
      'privacy': 'الخصوصية',
      'managePrivacy': 'إدارة إعدادات الخصوصية',
      'security': 'الأمان',
      'passwordAuth': 'كلمة المرور والمصادقة',
      'preferences': 'التفضيلات',
      'notifications': 'الإشعارات',
      'notificationsDesc': 'تلقي تنبيهات فورية عند إضافة عروض إحالة جديدة',
      'notificationsSubtitle': 'تلقي تنبيهات فورية عند إضافة عروض إحالة جديدة',
      'exclusiveOffers': 'عروض حصرية فقط',
      'exclusiveOffersDesc': 'تلقي إشعارات حول العروض المميزة فقط',
      'exclusiveOffersSubtitle': 'تلقي إشعارات حول العروض المميزة فقط',
      'darkMode': 'الوضع الداكن',
      'darkModeDesc': 'دائمًا قيد التشغيل (قريبًا)',
      'language': 'اللغة',
      'selectPreferredCategories': 'اختر فئاتك المفضلة',
      'customizeFeedSubtitle': 'اختر فئات العروض المفضلة لديك',
      'learnMoreAboutApp': 'تعرف على المزيد عن تطبيقنا',
      'support': 'الدعم',
      'helpCenter': 'مركز المساعدة',
      'helpCenterDesc': 'الحصول على المساعدة والدعم',
      'helpCenterSubtitle': 'الحصول على المساعدة والدعم',

      'About App': 'عن التطبيق',
      'version': 'الإصدار 1.0.0',
      'termsPrivacy': 'الشروط والخصوصية',
      'readPolicies': 'قراءة سياساتنا',
      'logout': 'تسجيل الخروج',
      'signOutAccount': 'تسجيل الخروج من الحساب',
      'loggedOut': 'تم تسجيل خروجك',
      'save': 'حفظ',
      'comingSoon': 'قادم في Afftok الإصدار الثاني 🔒',
      
      // Edit Profile
      'editProfileTitle': 'تعديل الملف الشخصي',
      'fullName': 'الاسم الكامل',
      'saveChanges': 'حفظ التغييرات',
      
      // Customize Feed
      'customizeFeedTitle': 'تخصيص خلاصتك',
      'selectCategories': 'اختر الفئات التي تهتم بها',
      'applyPreferences': 'تطبيق التفضيلات',
      
      // Language Dialog
      'selectLanguage': 'اختر اللغة',
      'english': 'الإنجليزية',
      'arabic': 'العربية',
      
      // About Dialog
      'aboutTitle': 'عن التطبيق',
      'aboutDesc': 'اكتشف أفضل عروض الإحالة واسترداد النقود من كبرى الشركات العالمية.',
      'aboutFullText': 'مرحبًا بك في ريف توك - حيث المشاركة تجلب لك النتائج! 🎉\n\  منصتك المثالية لاكتشاف ومشاركة أفضل عروض الإحالة وصفقات استرداد النقود من كبرى العلامات التجارية حول العالم. نحن نؤمن بأن توصيات الأصدقاء قوية، ويجب مكافأتها!\n\nكيف يعمل:\nما عليك سوى تصفح المئات من العروض الحصرية من الشركات الرائدة، ومشاركة روابط الإحالة الفريدة الخاصة بك مع الأصدقاء والعائلة، وشاهد إحالاتك الناجحة تزداد. كل إحالة ناجحة تجعلك مؤهلاً للحصول على مكافأة مباشرة من الشركة الشريكة!\n\nلماذا افتوك؟\n🌟 الوصول إلى مكافآت رائعة: احصل على مكافآت من الشركات الشريكة مقابل كل صديق يسجل من خلال رابطك\n🌟 عروض حصرية: الوصول إلى صفقات لن تجدها في أي مكان آخر\n🚀 سهل الاستخدام: واجهة على نمط تيك توك تجعل اكتشاف العروض ممتعًا وإدمانيًا\n🌍 علامات تجارية عالمية: شراكة مع شركات موثوقة مثل أمازون، أوبر، إير بي إن بي، والمزيد\n📱 تتبع فوري: راقب إحالاتك وأداءك في الوقت الفعلي\n🎁 تحديثات يومية: عروض جديدة تضاف كل يوم\n\nمهمتنا:\nإنشاء مجتمع يستفيد منه الجميع - تستفيد أنت من خلال المشاركة، ويوفر أصدقاؤك المال، وتصل العلامات التجارية إلى عملاء جدد. إنه فوز للجميع!\n\nانضم إلى آلاف المستخدمين الذين يستفيدون بالفعل من أف توك. حمل الآن وابدأ بتحويل شبكتك الاجتماعية إلى أداة ترويج قوية!\n\nريف توك - شارك. حقق. كرر. ✨',
      'close': 'إغلاق',
      
      // Logout Dialog
      'logoutTitle': 'تسجيل الخروج',
      'logoutConfirm': 'هل أنت متأكد من أنك تريد تسجيل الخروج؟',
      'cancel': 'إلغاء',
      
      // Sign In Screen
      'signInTitle': 'مرحبًا بعودتك',
      'welcomeBack': 'مرحبًا بعودتك',
      'signInToContinue': 'سجل الدخول للمتابعة والتتبع',
      'signInSubtitle': 'سجل الدخول للمتابعة والتتبع',
      'email': 'البريد الإلكتروني',
      'password': 'كلمة المرور',
      'forgotPassword': 'نسيت كلمة المرور؟',
      'signInButton': 'تسجيل الدخول',
      'orContinueWith': 'أو تابع بواسطة',
      'google': 'جوجل',
      'apple': 'آبل',
      'dontHaveAccount': 'ليس لديك حساب؟',
      'signUpLink': 'إنشاء حساب',
      
      // Sign Up Screen
      'signUpTitle': 'إنشاء حساب',
      'signUpSubtitle': 'ابدأ بالمشاركة والتتبع اليوم',
      'confirmPassword': 'تأكيد كلمة المرور',
      'agreeTerms': 'أوافق على الشروط وسياسة الخصوصية',
      'agreeToTerms': 'أوافق على',
      'termsOfService': 'شروط الخدمة',
      'and': 'و',
      'privacyPolicy': 'سياسة الخصوصية',
      'signUpButton': 'إنشاء حساب',
      'signUp': 'إنشاء حساب',
      'signUpToStart': 'سجل لتبدأ بالمشاركة والتتبع',
      'alreadyHaveAccount': 'هل لديك حساب بالفعل؟',
      'signInLink': 'تسجيل الدخول',
      'confirmPassword': 'تأكيد كلمة المرور',
      
      // Offer Categories
      'categoryAll': 'الكل',
      'categoryCrypto': 'عملات رقمية',
      'categoryShopping': 'تسوق',
      'categoryServices': 'خدمات',
      'categoryFood': 'طعام',
      'categoryTravel': 'سفر',
      'categoryFinance': 'مالية',
      
      // Teams Feature
      'teams': 'الفرق',
      'myTeam': 'فريقك',
      'leaderboard': 'لوحة المتصدرين',
      'challenges': 'التحديات',
      'teamPerformance': 'أداء الفريق',
      'referrals': 'الإحالات',
      'clicks': 'النقرات',
      'earned': 'المكتسب',
      'thisMonth': 'هذا الشهر',
      'teamGoal': 'هدف الفريق',
      'completed': 'اكتمل',
      'teamMembers': 'أعضاء الفريق',
      'you': 'أنت',
      'inviteMembers': 'دعوة أعضاء',
      'topTeamsThisMonth': 'أفضل الفرق هذا الشهر',
      'rank': 'الترتيب',
      'yourTeam': 'فريقك',
      'activeChallenges': 'التحديات النشطة',
      'progress': 'التقدم',
      'timeLeft': 'الوقت المتبقي',
      'currentRank': 'الترتيب الحالي',
      'createTeam': 'إنشاء فريق',
      'joinTeam': 'الانضمام لفريق',
      'noTeamYet': 'أنت لست في أي فريق بعد',
      'joinCreateTeam': 'انضم أو أنشئ فريقًا للمنافسة والأداء معًا!',
      'joinOrCreateTeam': 'انضم لفريق موجود أو أنشئ فريقك الخاص للتعاون مع الآخرين!',
      'browseTeams': 'تصفح الفرق',
      'teamName': 'اسم الفريق',
      'teamStats': 'إحصائيات الفريق',
      'members': 'الأعضاء',
      'points': 'النقاط',
      'noMembersYet': 'لا يوجد أعضاء بعد',
      'leaveTeam': 'مغادرة الفريق',
      'confirmLeaveTeam': 'هل أنت متأكد أنك تريد مغادرة هذا الفريق؟',
      'confirmJoinTeam': 'هل تريد الانضمام للفريق',
      'teamCreatedSuccessfully': 'تم إنشاء الفريق بنجاح!',
      'failedToCreateTeam': 'فشل في إنشاء الفريق',
      'joinedTeamSuccessfully': 'تم الانضمام للفريق بنجاح!',
      'failedToJoinTeam': 'فشل في الانضمام للفريق',
      'leftTeamSuccessfully': 'تمت مغادرة الفريق بنجاح!',
      'failedToLeaveTeam': 'فشل في مغادرة الفريق',
      'noTeamsYet': 'لا توجد فرق متاحة حالياً',
      'createFirstTeam': 'كن أول من ينشئ فريقاً!',
      'full': 'مكتمل',
      'join': 'انضمام',
      'leave': 'مغادرة',
      'create': 'إنشاء',
      'cancel': 'إلغاء',
      
      // Enhanced Profile
      'personalLink': 'الرابط الشخصي',
      'yourPersonalLink': 'صفحة الإحالة الشخصية الخاصة بك',
      'copyLink': 'نسخ الرابط',
      'shareLink': 'مشاركة الرابط',
      'viewPage': 'عرض الصفحة',
      'linkCopied': 'تم نسخ الرابط إلى الحافظة!',
      'stats': 'الإحصائيات',
      'totalClicks': 'إجمالي النقرات',
      'totalReferrals': 'إجمالي الإحالات',
      'totalEarnings': 'إجمالي النتائج',
      'level': 'المستوى',
      'badges': 'الشارات',
      'noBadgesYet': 'لا توجد شارات بعد',
      'earnBadges': 'أكمل التحديات لكسب الشارات!',
      'myOffers': 'عروضي',
      'activeOffers': 'العروض النشطة',
      'copy': 'نسخ',
      'yourPerformance': 'أدائك',
      'bestOffer': 'أفضل عرض',
      'members': 'الأعضاء',
      'myActiveOffers': 'عروضي النشطة',
      'addMoreOffers': 'أضف المزيد من العروض',
      'viewAllOffers': 'عرض جميع العروض',
      'myPublicPage': 'صفحتي',
      
      // Missing translations - Added
      'conversions': 'تحويلات',
      'addOffer': 'إضافة عرض',
      'analytics': 'التحليلات',
      'conversionRate': 'معدل التحويل',
      'totalConversions': 'إجمالي التحويلات',
      'youWillEarn': 'مكافأة الشريك',
      'offerAddedSuccessfully': 'تمت إضافة العرض بنجاح!',
      'failedToAddOffer': 'فشل في إضافة العرض. حاول مرة أخرى.',
      'addOfferToMyList': 'إضافة لعروضي',
      'offerAdded': 'تمت الإضافة ✓',
      'offerAlreadyAdded': 'العرض مضاف مسبقاً لقائمتك',
      'aboutThisOffer': 'حول هذا العرض',
      'usersEarned': 'المستخدمون الذين حققوا',
      'rating': 'التقييم',
      'addOfferInstructions': 'تعليمات إضافة العرض',
      'registerAtCompany': 'التسجيل في الشركة',
      'overview': 'نظرة عامة',
      'performanceChart': 'مخطط الأداء',
      'chartPlaceholder': 'عنصر نائب للمخطط',
      'topPerformingOffers': 'أفضل العروض أداءً',
      'additionalStats': 'إحصائيات إضافية',
      'globalRank': 'الترتيب العالمي',
      'monthlyEarnings': 'الأداء الشهري',
      'growthRate': 'معدل النمو',
      'allTime': 'جميع الأوقات',
      'thisWeek': 'هذا الأسبوع',
      'description': 'الوصف',
      'performance': 'الأداء',
      'yourReferralLink': 'رابط الإحالة الخاص بك',
      'shareOffer': 'مشاركة العرض',
      'openLink': 'فتح الرابط',
      'myLink': 'رابطي',
      'registeredOffers': 'العروض المسجلة',
      
      // Additional missing keys
      'addedDate': 'تاريخ الإضافة',
      'addedToMyOffers': 'تمت الإضافة إلى عروضي',
      'done': 'تم',
      'error': 'خطأ',
      'getNotified': 'احصل على إشعارات',
      'helped': 'تمت المساعدة',
      'iRegistered': 'أنا مسجل',
      'invalidLink': 'رابط غير صالح',
      'linkStats': 'إحصائيات الرابط',
      'offerAdded': 'تمت إضافة العرض',
      'offers': 'العروض',
      'ok': 'موافق',
      'redirectNotice': 'إشعار إعادة التوجيه',
      'registerAndAddOffer': 'سجل وأضف العرض',
      'saveLink': 'حفظ الرابط',
      'scanQRCode': 'مسح رمز الاستجابة السريعة',
      'shares': 'المشاركات',
      'startNow': 'ابدأ الآن',
      'viewDetails': 'عرض التفاصيل',
      'visitOffer': 'زيارة العرض',
      'visits': 'الزيارات',
      'addedToFavorites': 'تمت الإضافة إلى المفضلة',
      'removedFromFavorites': 'تمت الإزالة من المفضلة',
      
      // Additional Missing Keys
      'continueAsGuest': 'المتابعة كزائر',
      'recommendedBy': 'موصى به من قبل',
      'myPersonalLink': 'رابطي الشخصي',
      'qrCodeDescription': 'امسح رمز الاستجابة السريعة هذا لمشاركة رابطك الشخصي',
      'personalLinkDescription': 'شارك هذا الرابط للترويج لجميع عروضك',
      'noOffersAdded': 'لم تتم إضافة عروض بعد',
      'startAddingOffers': 'ابدأ بإضافة العروض لتتبع أدائك',
      'exploreOffers': 'استكشاف العروض',
      'registerInProgram': 'التسجيل في البرنامج',
      'registeredInProgram': 'مسجل في البرنامج',
      'performanceOverview': 'نظرة عامة على الأداء',
      'viewDetailedAnalytics': 'عرض التحليلات التفصيلية',
      'recommendedOffers': 'العروض الموصى بها',
      'poweredByAffTok': 'مدعوم بواسطة أف توك',
      'startEarningToo': 'ابدأ بتحقيق النتائج أيضًا',
      'registrationOpened': 'فتح التسجيل',
      'registrationInstructions': 'اتبع هذه الخطوات لإكمال التسجيل',
      'registrationTip': 'تأكد من استخدام رابط الإحالة الخاص بك عند التسجيل',
      'enterReferralLink': 'أدخل رابط الإحالة',
      'enterReferralLinkDescription': 'الصق رابط الإحالة الفريد الخاص بك من الشركة',
      'referralLinkWarning': 'يرجى التأكد من إدخال رابط الإحالة الصحيح',
      'offerAddedSuccessfully': 'تمت إضافة العرض بنجاح',
      'offerAddedDescription': 'يمكنك الآن البدء بالمشاركة وتحقيق النتائج',
      'couldNotOpenWebsite': 'تعذر فتح الموقع',
      'openingWebsite': 'جارٍ فتح الموقع...',
      
      // Download Prompt Screen
      'getTheBestExperience': 'احصل على أفضل تجربة',
      'downloadAppDescription': 'قم بتحميل تطبيق أف توك للحصول على التجربة الكاملة',
      'instantAccess': 'وصول فوري',
      'instantAccessDescription': 'وصول سريع لجميع العروض',
      'trackEarnings': 'تتبع الأداء',
      'trackEarningsDescription': 'راقب أدائك في الوقت الفعلي',
      'getNotifiedDescription': 'تلقي إشعارات حول العروض الجديدة',
      'youreAboutToAccess': 'أنت على وشك الوصول إلى',
      'downloadAffTok': 'تحميل أف توك',
      
      // Settings Screen Additional Getters
      'privacyPolicyTitle': 'سياسة الخصوصية',
      'privacyPolicySubtitle': 'قراءة قواعد البيانات والخصوصية لدينا.',
      'termsOfServiceTitle': 'شروط الخدمة',
      'termsOfServiceSubtitle': 'مراجعة الشروط واتفاقية المستخدم.',
      'aboutAffTokTitle': 'عن التطبيق',
      'aboutAffTokSubtitle': 'تعرف على المزيد عن مهمتنا وقصتنا.',
      
      // About Screen
      'aboutScreenTitle': 'عن التطبيق',
      'welcomeToAffTok': 'مرحبًا بك في أف توك',
      'whereSharingPays': 'حيث المشاركة تجلب لك النتائج! 🎉',
      'ourMission': 'مهمتنا',
      'missionText': 'مهمتنا هي إضفاء الطابع الديمقراطي على التسويق بالعمولة وجعله متاحًا للجميع. نحن نؤمن بأن لتوصياتك قيمة، ويجب مكافأتك على مشاركة المنتجات والخدمات الرائعة مع شبكتك.',
      'whatMakesAffTokSpecial': 'ما الذي يجعل أف توك مميزًا',
      'Engaging Visual Content': 'محتوى مرئي جذاب',
      'Engaging Visual Content': 'تصفح المئات من العروض الحصرية في خلاصة جذابة ومذهلة بصريًا.',
      'realEarnings': 'تتبع الأداء',
      'realEarningsDesc': 'راقب إحالاتك الناجحة والمكافآت التي أنت مؤهل للحصول عليها من الشركاء.',
      'teamPower': 'قوة الفريق',
      'teamPowerDesc': 'انضم أو أنشئ فرقًا للمنافسة والتعاون وتحقيق النتائج معًا.',
      'transparentTracking': 'تتبع شفاف',
      'transparentTrackingDesc': 'راقب نقراتك وإحالاتك وأدائك في الوقت الفعلي.',
      'globalOpportunities': 'فرص عالمية',
      'globalOpportunitiesDesc': 'الوصول إلى عروض من الشركات الرائدة عالميًا بما في ذلك أمازون، أوبر، إير بي إن بي، باينانس.',
      'gamifiedExperience': 'تجربة محفزة باللعب',
      'gamifiedExperienceDesc': 'ارتقِ بالمستوى، واكسب الشارات، وتنافس في التحديات.',
      'howItWorks': 'كيف يعمل',
      'discoverStep': 'اكتشف',
      'discoverStepDesc': 'تصفح خلاصتنا المنسقة من عروض العمولة.',
      'shareStep': 'شارك',
      'shareStepDesc': 'احصل على رابط الإحالة الفريد الخاص بك بنقرة واحدة.',
      'earnStep': 'احصل على النتائج',
      'earnStepDesc': 'عندما يسجل شخص ما من خلال رابطك، تصبح مؤهلاً للحصول على مكافأة من الشركة الشريكة.',
      'growStep': 'تنمية',
      'growStepDesc': 'انضم إلى الفرق، وشارك في التحديات، وارتقِ بالمستوى.',
      'theNumbers': 'الأرقام',
      'activeUsers': '500,000+ مستخدم نشط حول العالم',
      'paidOut': '+50 مليون دولار مكافآت شريك تم تسهيلها',
      'partnerCompanies': '1,000+ شركة شريكة',
      'countriesSupported': '150+ دولة مدعومة',
      'averageUserRating': '4.8/5 متوسط تقييم المستخدمين',
      // Stats Labels
      'activeUsersLabel': 'مستخدم نشط حول العالم',
      'paidOutLabel': 'مكافآت شريك تم تسهيلها',
      'partnerCompaniesLabel': 'شركة شريكة',
      'countriesSupportedLabel': 'دولة مدعومة',
      'averageUserRatingLabel': 'متوسط تقييم المستخدمين',
      'contactUs': 'اتصل بنا',
      'emailHello': 'البريد الإلكتروني: hello@afftok.com',
      'emailSupport': 'الدعم: support@afftok.com',
      'website': 'الموقع الإلكتروني: www.afftok.com',
      'instagram': 'إنستغرام: @afftok',
      'twitter': 'تويتر: @afftok',
      'versionText': 'الإصدار 1.0.0',
      'slogan': 'أف توك - شارك. حقق. كرر.',
      
      // Terms Screen
      'termsScreenTitle': 'شروط الخدمة',
      'effectiveDate': 'تاريخ السريان: 16 أكتوبر 2025',
      'acceptanceOfTerms': 'قبول الشروط',
      'acceptanceOfTermsDesc': 'بالوصول إلى أف توك أو استخدامه، فإنك توافق على الالتزام بشروط الخدمة هذه. إذا كنت لا توافق، يرجى التوقف عن الاستخدام فورًا.',
      'descriptionOfService': 'وصف الخدمة',
      'descriptionOfServiceDesc': 'أف توك هي منصة هاتف محمول تربط المستخدمين بفرص التسويق بالعمولة وعروض استرداد النقود من الشركات الشريكة.',
      'userAccounts': 'حسابات المستخدمين',
      'accountRequirements': 'متطلبات الحساب:',
      'ageRequirement': 'يجب أن لا يقل العمر عن 18 عامًا',
      'validEmail': 'توفير عنوان بريد إلكتروني صالح',
      'complyLaws': 'الامتثال لجميع القوانين المعمول بها',
      'legitimateUse': 'استخدام الخدمة لأغراض مشروعة فقط',
      'referralProgramRules': 'قواعد برنامج الإحالة',
      'permittedActivities': 'الأنشطة المسموح بها',
      'shareLinks': 'مشاركة روابط الإحالة الفريدة الخاصة بك من خلال قنوات مشروعة',
      'promoteHonestly': 'الترويج للعروض بصدق ودقة',
      'complyPartnerTerms': 'الامتثال لشروط وأحكام الشركة الشريكة',
      'prohibitedActivities': 'الأنشطة المحظورة',
      'noFakeAccounts': 'إنشاء حسابات أو إحالات وهمية',
      'noBots': 'استخدام روبوتات أو برامج نصية آلية',
      'noSpamming': 'إرسال رسائل غير مرغوب فيها أو تسويق غير مرغوب فيه',
      'noMisrepresentation': 'تحريف العروض أو المكافآت',
      'noManipulation': 'التلاعب بالنقرات أو التحويلات',
      'noFraud': 'الانخراط في سلوك احتيالي',
      'earningsPayments': 'المكافآت والأهلية',
      'earningsDesc': 'يتم تتبع المكافآت وتأكيد الأهلية عندما تفي الإحالات بمتطلبات الشركة الشريكة. يتم دفع المكافآت مباشرة من قبل الشركة الشريكة. تختلف أوقات المعالجة حسب الشريك وقد تستغرق 30-90 يومًا.',
      'taxResponsibility': 'أنت مسؤول عن الإبلاغ عن الضرائب ودفعها على التعويضات المستلمة من الشركات الشريكة وفقًا للقوانين المحلية.',
      'intellectualProperty': 'الملكية الفكرية',
      'intellectualPropertyDesc': 'جميع المحتويات والعلامات التجارية والملكية الفكرية على أف توك مملوكة لنا أو لمرخصينا. لا يجوز لك نسخ أو تعديل أو توزيع المحتوى الخاص بنا دون إذن.',
      'thirdPartyServices': 'خدمات الطرف الثالث',
      'thirdPartyServicesDesc': 'يتكامل أف توك مع الشركات الشريكة وخدمات الطرف الثالث. نحن لسنا مسؤولين عن أفعالهم أو منتجاتهم أو خدماتهم.',
      'disclaimers': 'إخلاء المسؤولية',
      'disclaimersDesc': 'نحن نقدم الخدمة "كما هي" دون ضمانات. تختلف نتائج الأداء وليست مضمونة. لا يقدم أف توك نصائح مالية أو استثمارية أو قانونية.',
      'limitationOfLiability': 'تحديد المسؤولية',
      'limitationOfLiabilityDesc': 'إلى الحد الأقصى الذي يسمح به القانون، لن يكون أف توك مسؤولًا عن الأضرار غير المباشرة أو العرضية أو التبعية.',
      'changesToTerms': 'التغييرات على الشروط',
      'changesToTermsDesc': 'قد نقوم بتعديل هذه الشروط في أي وقت. يشكل الاستخدام المستمر بعد التغييرات قبولًا.',
      'contactInformation': 'معلومات الاتصال',
      'contactInformationDesc': 'للاستفسارات حول شروط الخدمة هذه:',
      'emailLegal': 'البريد الإلكتروني: legal@afftok.com',
      'websiteTerms': 'الموقع الإلكتروني: www.afftok.com/terms',
      
      // Privacy Policy Screen
      'privacyPolicyScreenTitle': 'سياسة الخصوصية',
      'lastUpdated': 'آخر تحديث: 16 أكتوبر 2025',
      'introduction': 'مقدمة',
      'introductionDesc': 'مرحبًا بك في أف توك. نحن نحترم خصوصيتك وملتزمون بحماية بياناتك الشخصية. توضح سياسة الخصوصية هذه كيفية جمعنا لمعلوماتك واستخدامها وحمايتها عند استخدامك لتطبيق الهاتف المحمول الخاص بنا.',
      'informationWeCollect': 'المعلومات التي نجمعها',
      'personalInformation': 'المعلومات الشخصية',
      'accountInformation': 'معلومات الحساب: الاسم، عنوان البريد الإلكتروني، وصورة الملف الشخصي',
      'referralData': 'بيانات الإحالة: روابط الإحالة الخاصة بك، النقرات، والإحالات الناجحة',
      'performanceMetrics': 'مقاييس الأداء: تتبع النتائج، إحصائيات الفريق، وتقدم التحدي',
      'automaticallyCollectedInfo': 'المعلومات التي يتم جمعها تلقائيًا',
      'deviceInformation': 'معلومات الجهاز: نوع الجهاز، نظام التشغيل، ومعرفات الجهاز الفريدة',
      'usageData': 'بيانات الاستخدام: ميزات التطبيق المستخدمة، والوقت المستغرق، وأنماط التفاعل',
      'locationData': 'بيانات الموقع: الموقع العام (على مستوى البلد/المدينة) لتخصيص العروض',
      'howWeUseInfo': 'كيف نستخدم معلوماتك',
      'useInfoDesc': 'نستخدم معلوماتك من أجل:',
      'provideService': 'توفير وصيانة خدمة أف توك',
      'trackReferrals': 'تتبع إحالاتك ونتائج أدائك',
      'personalizeOffers': 'تخصيص العروض بناءً على اهتماماتك',
      'communicateUpdates': 'توصيل التحديثات والإشعارات الهامة',
      'improveApp': 'تحسين أداء التطبيق وتجربة المستخدم',
      'preventFraud': 'منع الاحتيال وضمان أمان المنصة',
      'informationSharing': 'مشاركة المعلومات',
      'informationSharingDesc': 'نحن لا نبيع معلوماتك الشخصية. قد نشارك البيانات مع:',
      'partnerCompaniesShare': 'الشركات الشريكة: لتتبع الإحالات وتأكيد الأهلية للمكافآت',
      'serviceProviders': 'مقدمو الخدمات: للتحليلات والاستضافة ومعالجة البيانات',
      'legalRequirements': 'المتطلبات القانونية: عندما يقتضي القانون أو لحماية حقوقنا',
      'dataSecurity': 'أمان البيانات',
      'dataSecurityDesc': 'نحن نطبق إجراءات أمنية قياسية في الصناعة بما في ذلك:',
      'encryptedData': 'نقل البيانات المشفر (SSL/TLS)',
      'secureServers': 'بنية تحتية خادم آمنة',
      'securityAudits': 'تدقيقات أمنية منتظمة',
      'accessControls': 'ضوابط الوصول والمصادقة',
      'yourRights': 'حقوقك',
      'yourRightsDesc': 'لديك الحق في:',
      'accessData': 'الوصول إلى بياناتك الشخصية وتلقي نسخة منها',
      'correctData': 'تصحيح أو تحديث البيانات غير الدقيقة',
      'deleteData': 'طلب حذف بياناتك الشخصية',
      'optOutMarketing': 'إلغاء الاشتراك في الاتصالات التسويقية',
      'dataRetention': 'الاحتفاظ بالبيانات',
      'dataRetentionDesc': 'نحتفظ ببياناتك الشخصية فقط للمدة الضرورية لتقديم الخدمات وتحقيق الأغراض الموضحة في هذه السياسة.',
      'childrensPrivacy': 'خصوصية الأطفال',
      'childrensPrivacyDesc': 'أف توك ليس مخصصًا للمستخدمين الذين تقل أعمارهم عن 18 عامًا. نحن لا نجمع بيانات شخصية من الأطفال عن علم.',
      'contactPrivacy': 'جهة الاتصال لاستفسارات الخصوصية',
      'contactPrivacyDesc': 'للاستفسارات حول سياسة الخصوصية هذه أو حقوق بياناتك:',
      'emailPrivacy': 'البريد الإلكتروني: privacy@afftok.com',
      'websitePrivacy': 'الموقع الإلكتروني: www.afftok.com/privacy',
      'requestDeletion': 'طلب حذف البيانات',
      'exportData': 'تصدير بياناتي',
      'contactUsPrivacy': 'اتصل بنا',
      'contactUsPrivacyDesc': 'لأسئلة الخصوصية، تواصل معنا على privacy@afftok.com',
    },
  };

  String translate(String key) {
    return _localizedValues[locale.languageCode]?[key] ?? 'MISSING_TRANSLATION:$key';
  }

  // General
  String get appName => translate('appName');
  String get tagline => translate('tagline');
  String get skip => translate('skip');
  
  // Bottom Navigation
  String get home => translate('home');
  String get discover => translate('discover');
  String get saved => translate('saved');
  String get profile => translate('profile');
  
  // Home Screen
  String get getReferralLink => translate('getReferralLink');
  String get share => translate('share');
  String get info => translate('info');
  String get checkOut => translate('checkOut');
  String get offerDetails => translate('offerDetails');
  String get reward => translate('reward');
  
  // Discover Screen
  String get discoverTitle => translate('discoverTitle');
  String get all => translate('all');
  String get cashback => translate('cashback');
  String get discounts => translate('discounts');
  String get crypto => translate('crypto');
  String get shopping => translate('shopping');
  String get services => translate('services');
  String get food => translate('food');
  String get travel => translate('travel');
  String get finance => translate('finance');
  String get entertainment => translate('entertainment');
  String get education => translate('education');
  String get noOffersFound => translate('noOffersFound');
  
  // Getters for newly added Discover Screen keys
  String get ecommerce => translate('ecommerce');
  String get technology => translate('technology');
  String get utilities => translate('utilities');
  String get foodRestaurants => translate('foodRestaurants');
      
  // Saved Screen
  String get savedTitle => translate('savedTitle');
  String get noSavedOffers => translate('noSavedOffers');
  String get startSaving => translate('startSaving');
  
  // Profile Screen
  String get guestUser => translate('guestUser');
  String get earnings => translate('earnings');
  String get shared => translate('shared');
  String get signIn => translate('signIn');
  String get createAccount => translate('createAccount');
  String get referralCode => translate('referralCode');
  String get achievements => translate('achievements');
  String get viewAll => translate('viewAll');
  
  // Settings Screen
  String get settings => translate('settings');
  String get account => translate('account');
  String get editProfile => translate('editProfile');
  String get changeNamePhoto => translate('changeNamePhoto');
  String get privacy => translate('privacy');
  String get managePrivacy => translate('managePrivacy');
  String get security => translate('security');
  String get passwordAuth => translate('passwordAuth');
  String get preferences => translate('preferences');
  String get notifications => translate('notifications');
  String get notificationsDesc => translate('notificationsDesc');
  String get notificationsSubtitle => translate('notificationsSubtitle');
  String get exclusiveOffers => translate('exclusiveOffers');
  String get exclusiveOffersDesc => translate('exclusiveOffersDesc');
  String get exclusiveOffersSubtitle => translate('exclusiveOffersSubtitle');
  String get darkMode => translate('darkMode');
  String get darkModeDesc => translate('darkModeDesc');
  String get language => translate('language');
  String get customizeFeed => translate('customizeFeed');
  String get customizeFeedDesc => translate('customizeFeedDesc');
  String get customizeFeedSubtitle => translate('customizeFeedSubtitle');
  String get support => translate('support');
  String get helpCenter => translate('helpCenter');
  String get helpCenterDesc => translate('helpCenterDesc');
  String get helpCenterSubtitle => translate('helpCenterSubtitle');
  String get aboutReftok => translate('aboutReftok');
  String get version => translate('version');
  String get termsPrivacy => translate('termsPrivacy');
  String get readPolicies => translate('readPolicies');
  String get logout => translate('logout');
  String get signOutAccount => translate('signOutAccount');
  String get loggedOut => translate('loggedOut');
  String get save => translate('save');
  String get comingSoon => translate('comingSoon');
  
  // Edit Profile
  String get editProfileTitle => translate('editProfileTitle');
  String get fullName => translate('fullName');
  String get saveChanges => translate('saveChanges');
  
  // Customize Feed
  String get customizeFeedTitle => translate('customizeFeedTitle');
  String get selectCategories => translate('selectCategories');
  String get applyPreferences => translate('applyPreferences');
  
  // Language Dialog
  String get selectLanguage => translate('selectLanguage');
  String get english => translate('english');
  String get arabic => translate('arabic');
  
  // About Dialog
  String get aboutTitle => translate('aboutTitle');
  String get aboutDesc => translate('aboutDesc');
  String get aboutFullText => translate('aboutFullText');
  String get close => translate('close');
  
  // Logout Dialog
  String get logoutTitle => translate('logoutTitle');
  String get logoutConfirm => translate('logoutConfirm');
  String get cancel => translate('cancel');
  
  // Sign In Screen
  String get signInTitle => translate('signInTitle');
  String get welcomeBack => translate('welcomeBack');
  String get signInToContinue => translate('signInToContinue');
  String get signInSubtitle => translate('signInSubtitle');
  String get email => translate('email');
  String get password => translate('password');
  String get forgotPassword => translate('forgotPassword');
  String get signInButton => translate('signInButton');
  String get orContinueWith => translate('orContinueWith');
  String get google => translate('google');
  String get apple => translate('apple');
  String get dontHaveAccount => translate('dontHaveAccount');
  String get signUpLink => translate('signUpLink');
  
  // Sign Up Screen
  String get signUpTitle => translate('signUpTitle');
  String get signUpSubtitle => translate('signUpSubtitle');
  String get confirmPassword => translate('confirmPassword');
  String get agreeTerms => translate('agreeTerms');
  String get agreeToTerms => translate('agreeToTerms');
  String get termsOfService => translate('termsOfService');
  String get and => translate('and');
  String get privacyPolicy => translate('privacyPolicy');
  String get signUpButton => translate('signUpButton');
  String get signUp => translate('signUp');
  String get signUpToStart => translate('signUpToStart');
  String get alreadyHaveAccount => translate('alreadyHaveAccount');
  String get signInLink => translate('signInLink');
  
  // Offer Categories
  String get categoryAll => translate('categoryAll');
  String get categoryCrypto => translate('categoryCrypto');
  String get categoryShopping => translate('categoryShopping');
  String get categoryServices => translate('categoryServices');
  String get categoryFood => translate('categoryFood');
  String get categoryTravel => translate('categoryTravel');
  String get categoryFinance => translate('categoryFinance');
  
  // Teams Feature
  String get teams => translate('teams');
  String get myTeam => translate('myTeam');
  String get leaderboard => translate('leaderboard');
  String get challenges => translate('challenges');
  String get teamPerformance => translate('teamPerformance');
  String get referrals => translate('referrals');
  String get clicks => translate('clicks');
  String get earned => translate('earned');
  String get thisMonth => translate('thisMonth');
  String get teamGoal => translate('teamGoal');
  String get completed => translate('completed');
  String get teamMembers => translate('teamMembers');
  String get you => translate('you');
  String get inviteMembers => translate('inviteMembers');
  String get topTeamsThisMonth => translate('topTeamsThisMonth');
  String get rank => translate('rank');
  String get yourTeam => translate('yourTeam');
  String get activeChallenges => translate('activeChallenges');
  String get progress => translate('progress');
  String get timeLeft => translate('timeLeft');
  String get currentRank => translate('currentRank');
  String get createTeam => translate('createTeam');
  String get joinTeam => translate('joinTeam');
  String get noTeamYet => translate('noTeamYet');
  String get joinCreateTeam => translate('joinCreateTeam');
  String get joinOrCreateTeam => translate('joinOrCreateTeam');
  String get browseTeams => translate('browseTeams');
  String get teamName => translate('teamName');
  String get teamStats => translate('teamStats');
  String get points => translate('points');
  String get noMembersYet => translate('noMembersYet');
  String get leaveTeam => translate('leaveTeam');
  String get confirmLeaveTeam => translate('confirmLeaveTeam');
  String get confirmJoinTeam => translate('confirmJoinTeam');
  String get teamCreatedSuccessfully => translate('teamCreatedSuccessfully');
  String get failedToCreateTeam => translate('failedToCreateTeam');
  String get joinedTeamSuccessfully => translate('joinedTeamSuccessfully');
  String get failedToJoinTeam => translate('failedToJoinTeam');
  String get leftTeamSuccessfully => translate('leftTeamSuccessfully');
  String get failedToLeaveTeam => translate('failedToLeaveTeam');
  String get noTeamsYet => translate('noTeamsYet');
  String get createFirstTeam => translate('createFirstTeam');
  String get full => translate('full');
  String get join => translate('join');
  String get leave => translate('leave');
  String get create => translate('create');
  // cancel already defined above
  
  // Enhanced Profile
  String get personalLink => translate('personalLink');
  String get yourPersonalLink => translate('yourPersonalLink');
  String get copyLink => translate('copyLink');
  String get shareLink => translate('shareLink');
  String get viewPage => translate('viewPage');
  String get linkCopied => translate('linkCopied');
  String get stats => translate('stats');
  String get totalClicks => translate('totalClicks');
  String get totalReferrals => translate('totalReferrals');
  String get totalEarnings => translate('totalEarnings');
  String get level => translate('level');
  String get badges => translate('badges');
  String get noBadgesYet => translate('noBadgesYet');
  String get earnBadges => translate('earnBadges');
  String get myOffers => translate('myOffers');
  String get activeOffers => translate('activeOffers');
  String get copy => translate('copy');
  String get yourPerformance => translate('yourPerformance');
  String get bestOffer => translate('bestOffer');
  String get members => translate('members');
  String get myActiveOffers => translate('myActiveOffers');
  String get addMoreOffers => translate('addMoreOffers');
  String get viewAllOffers => translate('viewAllOffers');
  String get myPublicPage => translate('myPublicPage');
  
  // Missing translations - Added
  String get conversions => translate('conversions');
  String get addOffer => translate('addOffer');
  String get analytics => translate('analytics');
  String get conversionRate => translate('conversionRate');
  String get totalConversions => translate('totalConversions');
  String get youWillEarn => translate('youWillEarn');
  String get failedToAddOffer => translate('failedToAddOffer');
  String get addOfferToMyList => translate('addOfferToMyList');
  String get aboutThisOffer => translate('aboutThisOffer');
  String get usersEarned => translate('usersEarned');
  String get rating => translate('rating');
  String get addOfferInstructions => translate('addOfferInstructions');
  String get registerAtCompany => translate('registerAtCompany');
  String get overview => translate('overview');
  String get performanceChart => translate('performanceChart');
  String get chartPlaceholder => translate('chartPlaceholder');
  String get topPerformingOffers => translate('topPerformingOffers');
  String get additionalStats => translate('additionalStats');
  String get globalRank => translate('globalRank');
  String get monthlyEarnings => translate('monthlyEarnings');
  String get growthRate => translate('growthRate');
  String get allTime => translate('allTime');
  String get thisWeek => translate('thisWeek');
  String get description => translate('description');
  String get performance => translate('performance');
  String get yourReferralLink => translate('yourReferralLink');
  String get shareOffer => translate('shareOffer');
  String get openLink => translate('openLink');
  String get myLink => translate('myLink');
  String get registeredOffers => translate('registeredOffers');
  
  // Additional missing keys
  String get addedDate => translate('addedDate');
  String get addedToMyOffers => translate('addedToMyOffers');
  String get done => translate('done');
  String get error => translate('error');
  String get getNotified => translate('getNotified');
  String get helped => translate('helped');
  String get iRegistered => translate('iRegistered');
  String get invalidLink => translate('invalidLink');
  String get linkStats => translate('linkStats');
  String get offerAdded => translate('offerAdded');
  String get offerAlreadyAdded => translate('offerAlreadyAdded');
  String get offers => translate('offers');
  String get ok => translate('ok');
  String get redirectNotice => translate('redirectNotice');
  String get registerAndAddOffer => translate('registerAndAddOffer');
  String get saveLink => translate('saveLink');
  String get scanQRCode => translate('scanQRCode');
  String get shares => translate('shares');
  String get startNow => translate('startNow');
  String get viewDetails => translate('viewDetails');
  String get visitOffer => translate('visitOffer');
  String get visits => translate('visits');
  String get addedToFavorites => translate('addedToFavorites');
  String get removedFromFavorites => translate('removedFromFavorites');
  
  // Additional Missing Keys
  String get continueAsGuest => translate('continueAsGuest');
  String get recommendedBy => translate('recommendedBy');
  String get myPersonalLink => translate('myPersonalLink');
  String get qrCodeDescription => translate('qrCodeDescription');
  String get personalLinkDescription => translate('personalLinkDescription');
  String get noOffersAdded => translate('noOffersAdded');
  String get startAddingOffers => translate('startAddingOffers');
  String get exploreOffers => translate('exploreOffers');
  String get registerInProgram => translate('registerInProgram');
  String get registeredInProgram => translate('registeredInProgram');
  String get performanceOverview => translate('performanceOverview');
  String get viewDetailedAnalytics => translate('viewDetailedAnalytics');
  String get recommendedOffers => translate('recommendedOffers');
  String get poweredByAffTok => translate('poweredByAffTok');
  String get startEarningToo => translate('startEarningToo');
  String get registrationOpened => translate('registrationOpened');
  String get registrationInstructions => translate('registrationInstructions');
  String get registrationTip => translate('registrationTip');
  String get enterReferralLink => translate('enterReferralLink');
  String get enterReferralLinkDescription => translate('enterReferralLinkDescription');
  String get referralLinkWarning => translate('referralLinkWarning');
  String get offerAddedSuccessfully => translate('offerAddedSuccessfully');
  String get offerAddedDescription => translate('offerAddedDescription');
  String get couldNotOpenWebsite => translate('couldNotOpenWebsite');
  String get openingWebsite => translate('openingWebsite');
  
  // Download Prompt Screen
  String get getTheBestExperience => translate('getTheBestExperience');
  String get downloadAppDescription => translate('downloadAppDescription');
  String get instantAccess => translate('instantAccess');
  String get instantAccessDescription => translate('instantAccessDescription');
  String get trackEarnings => translate('trackEarnings');
  String get trackEarningsDescription => translate('trackEarningsDescription');
  String get getNotifiedDescription => translate('getNotifiedDescription');
  String get youreAboutToAccess => translate('youreAboutToAccess');
  String get downloadAffTok => translate('downloadAffTok');
  
  // Settings Screen Additional Getters
  String get privacyPolicyTitle => translate('privacyPolicyTitle');
  String get privacyPolicySubtitle => translate('privacyPolicySubtitle');
  String get termsOfServiceTitle => translate('termsOfServiceTitle');
  String get termsOfServiceSubtitle => translate('termsOfServiceSubtitle');
  String get aboutAffTokTitle => translate('aboutAffTokTitle');
  String get aboutAffTokSubtitle => translate('aboutAffTokSubtitle');
  
  // About Screen
  String get aboutScreenTitle => translate('aboutScreenTitle');
  String get welcomeToAffTok => translate('welcomeToAffTok');
  String get whereSharingPays => translate('whereSharingPays');
  String get ourMission => translate('ourMission');
  String get missionText => translate('missionText');
  String get whatMakesAffTokSpecial => translate('whatMakesAffTokSpecial');
  String get tikTokStyleDiscovery => translate('tikTokStyleDiscovery');
  String get tikTokStyleDiscoveryDesc => translate('tikTokStyleDiscoveryDesc');
  String get realEarnings => translate('realEarnings');
  String get realEarningsDesc => translate('realEarningsDesc');
  String get teamPower => translate('teamPower');
  String get teamPowerDesc => translate('teamPowerDesc');
  String get transparentTracking => translate('transparentTracking');
  String get transparentTrackingDesc => translate('transparentTrackingDesc');
  String get globalOpportunities => translate('globalOpportunities');
  String get globalOpportunitiesDesc => translate('globalOpportunitiesDesc');
  String get gamifiedExperience => translate('gamifiedExperience');
  String get gamifiedExperienceDesc => translate('gamifiedExperienceDesc');
  String get howItWorks => translate('howItWorks');
  String get discoverStep => translate('discoverStep');
  String get discoverStepDesc => translate('discoverStepDesc');
  String get shareStep => translate('shareStep');
  String get shareStepDesc => translate('shareStepDesc');
  String get earnStep => translate('earnStep');
  String get earnStepDesc => translate('earnStepDesc');
  String get growStep => translate('growStep');
  String get growStepDesc => translate('growStepDesc');
  String get theNumbers => translate('theNumbers');
  String get activeUsers => translate('activeUsers');
  String get paidOut => translate('paidOut');
  String get partnerCompanies => translate('partnerCompanies');
  String get countriesSupported => translate('countriesSupported');
  String get averageUserRating => translate('averageUserRating');
  // Stats Labels
  String get activeUsersLabel => translate('activeUsersLabel');
  String get paidOutLabel => translate('paidOutLabel');
  String get partnerCompaniesLabel => translate('partnerCompaniesLabel');
  String get countriesSupportedLabel => translate('countriesSupportedLabel');
  String get averageUserRatingLabel => translate('averageUserRatingLabel');
  String get contactUs => translate('contactUs');
  String get emailHello => translate('emailHello');
  String get emailSupport => translate('emailSupport');
  String get website => translate('website');
  String get instagram => translate('instagram');
  String get twitter => translate('twitter');
  String get versionText => translate('versionText');
  String get slogan => translate('slogan');
  
  // Terms Screen
  String get termsScreenTitle => translate('termsScreenTitle');
  String get effectiveDate => translate('effectiveDate');
  String get acceptanceOfTerms => translate('acceptanceOfTerms');
  String get acceptanceOfTermsDesc => translate('acceptanceOfTermsDesc');
  String get descriptionOfService => translate('descriptionOfService');
  String get descriptionOfServiceDesc => translate('descriptionOfServiceDesc');
  String get userAccounts => translate('userAccounts');
  String get accountRequirements => translate('accountRequirements');
  String get ageRequirement => translate('ageRequirement');
  String get validEmail => translate('validEmail');
  String get complyLaws => translate('complyLaws');
  String get legitimateUse => translate('legitimateUse');
  String get referralProgramRules => translate('referralProgramRules');
  String get permittedActivities => translate('permittedActivities');
  String get shareLinks => translate('shareLinks');
  String get promoteHonestly => translate('promoteHonestly');
  String get complyPartnerTerms => translate('complyPartnerTerms');
  String get prohibitedActivities => translate('prohibitedActivities');
  String get noFakeAccounts => translate('noFakeAccounts');
  String get noBots => translate('noBots');
  String get noSpamming => translate('noSpamming');
  String get noMisrepresentation => translate('noMisrepresentation');
  String get noManipulation => translate('noManipulation');
  String get noFraud => translate('noFraud');
  String get earningsPayments => translate('earningsPayments');
  String get earningsDesc => translate('earningsDesc');
  String get taxResponsibility => translate('taxResponsibility');
  String get intellectualProperty => translate('intellectualProperty');
  String get intellectualPropertyDesc => translate('intellectualPropertyDesc');
  String get thirdPartyServices => translate('thirdPartyServices');
  String get thirdPartyServicesDesc => translate('thirdPartyServicesDesc');
  String get disclaimers => translate('disclaimers');
  String get disclaimersDesc => translate('disclaimersDesc');
  String get limitationOfLiability => translate('limitationOfLiability');
  String get limitationOfLiabilityDesc => translate('limitationOfLiabilityDesc');
  String get changesToTerms => translate('changesToTerms');
  String get changesToTermsDesc => translate('changesToTermsDesc');
  String get contactInformation => translate('contactInformation');
  String get contactInformationDesc => translate('contactInformationDesc');
  String get emailLegal => translate('emailLegal');
  String get websiteTerms => translate('websiteTerms');
  
  // Privacy Policy Screen
  String get privacyPolicyScreenTitle => translate('privacyPolicyScreenTitle');
  String get lastUpdated => translate('lastUpdated');
  String get introduction => translate('introduction');
  String get introductionDesc => translate('introductionDesc');
  String get informationWeCollect => translate('informationWeCollect');
  String get personalInformation => translate('personalInformation');
  String get accountInformation => translate('accountInformation');
  String get referralData => translate('referralData');
  String get performanceMetrics => translate('performanceMetrics');
  String get automaticallyCollectedInfo => translate('automaticallyCollectedInfo');
  String get deviceInformation => translate('deviceInformation');
  String get usageData => translate('usageData');
  String get locationData => translate('locationData');
  String get howWeUseInfo => translate('howWeUseInfo');
  String get useInfoDesc => translate('useInfoDesc');
  String get provideService => translate('provideService');
  String get trackReferrals => translate('trackReferrals');
  String get personalizeOffers => translate('personalizeOffers');
  String get communicateUpdates => translate('communicateUpdates');
  String get improveApp => translate('improveApp');
  String get preventFraud => translate('preventFraud');
  String get informationSharing => translate('informationSharing');
  String get informationSharingDesc => translate('informationSharingDesc');
  String get partnerCompaniesShare => translate('partnerCompaniesShare');
  String get serviceProviders => translate('serviceProviders');
  String get legalRequirements => translate('legalRequirements');
  String get dataSecurity => translate('dataSecurity');
  String get dataSecurityDesc => translate('dataSecurityDesc');
  String get encryptedData => translate('encryptedData');
  String get secureServers => translate('secureServers');
  String get securityAudits => translate('securityAudits');
  String get accessControls => translate('accessControls');
  String get yourRights => translate('yourRights');
  String get yourRightsDesc => translate('yourRightsDesc');
  String get accessData => translate('accessData');
  String get correctData => translate('correctData');
  String get deleteData => translate('deleteData');
  String get optOutMarketing => translate('optOutMarketing');
  String get dataRetention => translate('dataRetention');
  String get dataRetentionDesc => translate('dataRetentionDesc');
  String get childrensPrivacy => translate('childrensPrivacy');
  String get childrensPrivacyDesc => translate('childrensPrivacyDesc');
  String get contactPrivacy => translate('contactPrivacy');
  String get contactPrivacyDesc => translate('contactPrivacyDesc');
  String get emailPrivacy => translate('emailPrivacy');
  String get websitePrivacy => translate('websitePrivacy');
  String get selectPreferredCategories => translate('selectPreferredCategories');
  String get learnMoreAboutApp => translate('learnMoreAboutApp');
  
  // Privacy Policy Screen - Missing getters
  String get requestDeletion => translate('requestDeletion');
  String get exportData => translate('exportData');
  String get contactUsPrivacy => translate('contactUsPrivacy');
  String get contactUsPrivacyDesc => translate('contactUsPrivacyDesc');
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  bool isSupported(Locale locale) {
    return ['en', 'ar'].contains(locale.languageCode);
  }

  @override
  Future<AppLocalizations> load(Locale locale) async {
    return AppLocalizations(locale);
  }

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}
