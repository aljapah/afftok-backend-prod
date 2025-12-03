class Offer {
  final String id;
  final String companyName;
  final String logoUrl;
  final String imageUrl;
  final String reward;
  final String description;
  final String category;
  final String offerUrl;
  final String registrationUrl;
  final bool isRegistered;

  final String? networkId;
  final String? networkName;
  final String? trackingParameter;
  final String? trackingLinkTemplate;

  final double rating;
  final String usersCount;

  Offer({
    required this.id,
    required this.companyName,
    required this.logoUrl,
    required this.imageUrl,
    required this.reward,
    required this.description,
    required this.category,
    required this.offerUrl,
    required this.registrationUrl,
    required this.rating,
    required this.usersCount,
    this.isRegistered = false,
    this.networkId,
    this.networkName,
    this.trackingParameter,
    this.trackingLinkTemplate,
  });

  factory Offer.fromJson(Map<String, dynamic> json) {
    final payout = json['payout'] ?? json['commission'] ?? 0;
    final payoutType = json['payout_type'] ?? 'cpa';

    String rewardText = '\$$payout per $payoutType';
    
    // Parse rating safely
    double rating = 4.5;
    if (json['rating'] != null) {
      if (json['rating'] is double) {
        rating = json['rating'];
      } else if (json['rating'] is int) {
        rating = (json['rating'] as int).toDouble();
      } else if (json['rating'] is String) {
        rating = double.tryParse(json['rating']) ?? 4.5;
      }
    }
    
    print('[Offer.fromJson] Parsing offer: ${json['title']}, id: ${json['id']}');

    return Offer(
      id: json['id']?.toString() ?? '',
      companyName: json['title'] ?? json['name'] ?? 'Offer',
      logoUrl: json['logo_url'] ?? '',
      imageUrl: json['image_url'] ?? '',
      reward: rewardText,
      description: json['description'] ?? '',
      category: json['category'] ?? 'General',
      offerUrl: json['destination_url'] ?? '',
      registrationUrl: json['destination_url'] ?? '',
      rating: rating,
      usersCount: json['users_count']?.toString() ?? '0',
      isRegistered: json['is_registered'] ?? false,
      networkId: json['network_id']?.toString(),
      networkName: json['network_name'],
      trackingParameter: json['tracking_parameter'],
      trackingLinkTemplate: json['tracking_link_template'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': companyName,
      'logo_url': logoUrl,
      'image_url': imageUrl,
      'description': description,
      'category': category,
      'destination_url': offerUrl,
      'payout': reward,
      'rating': rating,
      'users_count': usersCount,
      'is_registered': isRegistered,
      'network_id': networkId,
      'network_name': networkName,
      'tracking_parameter': trackingParameter,
      'tracking_link_template': trackingLinkTemplate,
    };
  }

  static List<Offer> getSampleOffers() {
    return [
      Offer(
        id: '1',
        companyName: 'Binance',
        logoUrl: 'https://logo.clearbit.com/binance.com',
        imageUrl: 'https://images.unsplash.com/photo-1621761191319-c6fb62004040?w=800',
        reward: 'Earn \$10 per Signup',
        description: 'Get \$10 bonus for every verified referral',
        category: 'Crypto',
        offerUrl: 'https://www.binance.com/en/activity/referral',
        registrationUrl: 'https://www.binance.com/en/activity/referral/affiliate/register',
        rating: 4.7,
        usersCount: '500K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '2',
        companyName: 'Coinbase',
        logoUrl: 'https://logo.clearbit.com/coinbase.com',
        imageUrl: 'https://images.unsplash.com/photo-1622630998477-20aa696ecb05?w=800',
        reward: '\$10 Bitcoin Bonus',
        description: 'Earn \$10 in Bitcoin for each referral',
        category: 'Crypto',
        offerUrl: 'https://www.coinbase.com/join',
        registrationUrl: 'https://www.coinbase.com/join/affiliate/register',
        rating: 4.6,
        usersCount: '300K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '3',
        companyName: 'Crypto.com',
        logoUrl: 'https://logo.clearbit.com/crypto.com',
        imageUrl: 'https://images.unsplash.com/photo-1639762681485-074b7f938ba0?w=800',
        reward: '\$25 Sign Up Bonus',
        description: 'Get \$25 when you stake CRO',
        category: 'Crypto',
        offerUrl: 'https://crypto.com/app',
        registrationUrl: 'https://crypto.com/app/affiliate/register',
        rating: 4.5,
        usersCount: '200K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '4',
        companyName: 'Kraken',
        logoUrl: 'https://logo.clearbit.com/kraken.com',
        imageUrl: 'https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?w=800',
        reward: '\$10 Trading Credit',
        description: 'Earn \$10 for each verified referral',
        category: 'Crypto',
        offerUrl: 'https://www.kraken.com/sign-up',
        registrationUrl: 'https://www.kraken.com/sign-up/affiliate/register',
        rating: 4.4,
        usersCount: '150K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '5',
        companyName: 'Amazon',
        logoUrl: 'https://logo.clearbit.com/amazon.com',
        imageUrl: 'https://images.unsplash.com/photo-1523474253046-8cd2748b5fd2?w=800',
        reward: '8% Cashback',
        description: 'Earn 8% cashback on all purchases',
        category: 'Shopping',
        offerUrl: 'https://www.amazon.com',
        registrationUrl: 'https://www.amazon.com/affiliate/register',
        rating: 4.9,
        usersCount: '1M+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '6',
        companyName: 'Nike',
        logoUrl: 'https://logo.clearbit.com/nike.com',
        imageUrl: 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800',
        reward: '15% Off First Order',
        description: 'Get 15% discount on your first purchase',
        category: 'Shopping',
        offerUrl: 'https://www.nike.com',
        registrationUrl: 'https://www.nike.com/affiliate/register',
        rating: 4.8,
        usersCount: '400K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '7',
        companyName: 'SHEIN',
        logoUrl: 'https://logo.clearbit.com/shein.com',
        imageUrl: 'https://images.unsplash.com/photo-1445205170230-053b83016050?w=800',
        reward: '20% Off + Free Shipping',
        description: 'Get 20% off and free shipping on first order',
        category: 'Shopping',
        offerUrl: 'https://www.shein.com',
        registrationUrl: 'https://www.shein.com/affiliate/register',
        rating: 4.6,
        usersCount: '350K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '8',
        companyName: 'AliExpress',
        logoUrl: 'https://logo.clearbit.com/aliexpress.com',
        imageUrl: 'https://images.unsplash.com/photo-1472851294608-062f824d29cc?w=800',
        reward: '\$5 Coupon',
        description: 'Get \$5 off on orders over \$50',
        category: 'Shopping',
        offerUrl: 'https://www.aliexpress.com',
        registrationUrl: 'https://www.aliexpress.com/affiliate/register',
        rating: 4.5,
        usersCount: '600K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '9',
        companyName: 'Zara',
        logoUrl: 'https://logo.clearbit.com/zara.com',
        imageUrl: 'https://images.unsplash.com/photo-1490481651871-ab68de25d43d?w=800',
        reward: '10% Student Discount',
        description: 'Students get 10% off all items',
        category: 'Shopping',
        offerUrl: 'https://www.zara.com',
        registrationUrl: 'https://www.zara.com/affiliate/register',
        rating: 4.7,
        usersCount: '250K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '10',
        companyName: 'Uber',
        logoUrl: 'https://logo.clearbit.com/uber.com',
        imageUrl: 'https://images.unsplash.com/photo-1449965408869-eaa3f722e40d?w=800',
        reward: '\$15 Off First Ride',
        description: 'Get \$15 off your first Uber ride',
        category: 'Services',
        offerUrl: 'https://www.uber.com',
        registrationUrl: 'https://www.uber.com/affiliate/register',
        rating: 4.6,
        usersCount: '800K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '11',
        companyName: 'Shopify',
        logoUrl: 'https://logo.clearbit.com/shopify.com',
        imageUrl: 'https://images.unsplash.com/photo-1556742049-0cfed4f6a45d?w=800',
        reward: '14-Day Free Trial',
        description: 'Start your online store with free trial',
        category: 'Services',
        offerUrl: 'https://www.shopify.com',
        registrationUrl: 'https://www.shopify.com/affiliate/register',
        rating: 4.7,
        usersCount: '200K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '12',
        companyName: 'Fiverr',
        logoUrl: 'https://logo.clearbit.com/fiverr.com',
        imageUrl: 'https://images.unsplash.com/photo-1522202176988-66273c2fd55f?w=800',
        reward: '20% Off First Order',
        description: 'Get 20% discount on your first service',
        category: 'Services',
        offerUrl: 'https://www.fiverr.com',
        registrationUrl: 'https://www.fiverr.com/affiliate/register',
        rating: 4.6,
        usersCount: '180K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '13',
        companyName: 'Upwork',
        logoUrl: 'https://logo.clearbit.com/upwork.com',
        imageUrl: 'https://images.unsplash.com/photo-1553877522-43269d4ea984?w=800',
        reward: '\$100 Credit',
        description: 'Get \$100 credit for your first project',
        category: 'Services',
        offerUrl: 'https://www.upwork.com',
        registrationUrl: 'https://www.upwork.com/affiliate/register',
        rating: 4.5,
        usersCount: '220K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '14',
        companyName: 'PayPal',
        logoUrl: 'https://logo.clearbit.com/paypal.com',
        imageUrl: 'https://images.unsplash.com/photo-1563013544-824ae1b704d3?w=800',
        reward: '\$10 Sign Up Bonus',
        description: 'Get \$10 when you sign up and verify',
        category: 'Finance',
        offerUrl: 'https://www.paypal.com',
        registrationUrl: 'https://www.paypal.com/affiliate/register',
        rating: 4.8,
        usersCount: '900K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '15',
        companyName: 'Revolut',
        logoUrl: 'https://logo.clearbit.com/revolut.com',
        imageUrl: 'https://images.unsplash.com/photo-1556742111-a301076d9d18?w=800',
        reward: 'Free Premium Trial',
        description: '3 months free Premium subscription',
        category: 'Finance',
        offerUrl: 'https://www.revolut.com',
        registrationUrl: 'https://www.revolut.com/affiliate/register',
        rating: 4.7,
        usersCount: '300K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '16',
        companyName: 'Wise',
        logoUrl: 'https://logo.clearbit.com/wise.com',
        imageUrl: 'https://images.unsplash.com/photo-1579621970563-ebec7560ff3e?w=800',
        reward: 'Fee-Free Transfer',
        description: 'First transfer up to \$500 fee-free',
        category: 'Finance',
        offerUrl: 'https://wise.com',
        registrationUrl: 'https://wise.com/affiliate/register',
        rating: 4.6,
        usersCount: '250K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
      Offer(
        id: '17',
        companyName: 'Robinhood',
        logoUrl: 'https://logo.clearbit.com/robinhood.com',
        imageUrl: 'https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?w=800',
        reward: 'Free Stock',
        description: 'Get a free stock worth up to \$200',
        category: 'Finance',
        offerUrl: 'https://robinhood.com',
        registrationUrl: 'https://robinhood.com/affiliate/register',
        rating: 4.5,
        usersCount: '400K+',
        networkName: 'ArabClicks',
        trackingParameter: 'subid',
      ),
    ];
  }
}
