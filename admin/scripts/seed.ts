import { drizzle } from "drizzle-orm/mysql2";
import mysql from "mysql2/promise";
import * as schema from "../drizzle/schema";

const DATABASE_URL = process.env.DATABASE_URL || "mysql://root:password@localhost:3306/afftok";

async function seed() {
  console.log("🌱 Starting database seed...");

  const connection = await mysql.createConnection(DATABASE_URL);
  const db = drizzle(connection, { schema, mode: "default" });

  try {
    // Clear existing data (optional - comment out if you want to keep existing data)
    console.log("🗑️  Clearing existing data...");
    await db.delete(schema.conversions);
    await db.delete(schema.clicks);
    await db.delete(schema.userBadges);
    await db.delete(schema.badges);
    await db.delete(schema.teamMembers);
    await db.delete(schema.teams);
    await db.delete(schema.offers);
    await db.delete(schema.networks);
    await db.delete(schema.users);

    // Seed Users
    console.log("👥 Seeding users...");
    const usernames = [
      "ahmed_ali", "sara_mohamed", "omar_hassan", "fatima_ahmed", "khalid_ibrahim",
      "layla_yousef", "mohammed_salem", "aisha_khalid", "abdullah_omar", "noor_ahmed",
      "hassan_ali", "mariam_hassan", "youssef_mohammed", "huda_salem", "ali_abdullah",
      "zainab_omar", "ibrahim_khalid", "salma_yousef", "faisal_ahmed", "reem_hassan",
      "tariq_ali", "noura_mohammed", "saeed_ibrahim", "lina_salem", "waleed_omar",
      "dina_khalid", "majid_yousef", "hana_ahmed", "rashid_hassan", "rana_ali",
      "adel_mohammed", "mona_ibrahim", "karim_salem", "leila_omar", "nasser_khalid",
      "amira_yousef", "hamza_ahmed", "yasmin_hassan", "bilal_ali", "samira_mohammed",
      "fahad_ibrahim", "nada_salem", "mansour_omar", "rania_khalid", "talal_yousef",
      "hiba_ahmed", "sami_hassan", "dalia_ali", "jamal_mohammed", "maya_ibrahim"
    ];

    const userIds: number[] = [];
    for (let i = 0; i < usernames.length; i++) {
      const result = await db.insert(schema.users).values({
        username: usernames[i],
        email: `${usernames[i]}@example.com`,
        passwordHash: "$2a$10$dummyhashforseeding1234567890",
        fullName: usernames[i].replace("_", " ").split(" ").map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(" "),
        phoneNumber: `+96650${Math.floor(Math.random() * 10000000).toString().padStart(7, "0")}`,
        status: Math.random() > 0.1 ? "active" : "inactive",
        points: Math.floor(Math.random() * 5000),
        level: Math.floor(Math.random() * 10) + 1,
        totalEarnings: Math.floor(Math.random() * 10000),
        totalClicks: Math.floor(Math.random() * 1000),
        totalConversions: Math.floor(Math.random() * 100),
      });
      userIds.push(Number(result.insertId));
    }
    console.log(`✅ Created ${userIds.length} users`);

    // Seed Networks
    console.log("🌐 Seeding networks...");
    const networkNames = [
      "Amazon Associates", "ClickBank", "ShareASale", "CJ Affiliate", "Rakuten Marketing",
      "Impact", "Awin", "FlexOffers", "Pepperjam", "Refersion"
    ];
    const networkIds: number[] = [];
    for (const name of networkNames) {
      const result = await db.insert(schema.networks).values({
        name,
        description: `${name} - Leading affiliate marketing network`,
        websiteUrl: `https://${name.toLowerCase().replace(/\s+/g, "")}.com`,
        status: Math.random() > 0.2 ? "active" : "inactive",
      });
      networkIds.push(Number(result.insertId));
    }
    console.log(`✅ Created ${networkIds.length} networks`);

    // Seed Offers
    console.log("🎁 Seeding offers...");
    const offerTitles = [
      "Amazon Prime 30-Day Free Trial", "Shopify 14-Day Free Trial", "Netflix Premium Subscription",
      "Udemy Course Bundle", "NordVPN 2-Year Plan", "Adobe Creative Cloud", "Grammarly Premium",
      "Canva Pro Subscription", "Spotify Premium Family", "YouTube Premium",
      "Microsoft 365 Personal", "Dropbox Plus", "Zoom Pro Plan", "Slack Business+",
      "Mailchimp Standard", "HubSpot CRM", "Salesforce Essentials", "QuickBooks Online",
      "FreshBooks Accounting", "Xero Accounting", "Wix Premium", "Squarespace Business",
      "WordPress Premium", "Bluehost Hosting", "HostGator Cloud", "GoDaddy Pro",
      "Namecheap Premium DNS", "Cloudflare Pro", "AWS Free Tier", "Google Cloud Credits"
    ];

    const categories = ["Software", "E-commerce", "Entertainment", "Education", "Business Tools", "Web Services"];
    const offerIds: number[] = [];
    
    for (let i = 0; i < offerTitles.length; i++) {
      const result = await db.insert(schema.offers).values({
        title: offerTitles[i],
        description: `Get exclusive access to ${offerTitles[i]} with special affiliate pricing`,
        category: categories[Math.floor(Math.random() * categories.length)],
        payout: Math.floor(Math.random() * 10000) + 1000, // 10-110 SAR in halalas
        destinationUrl: `https://example.com/offer/${i + 1}`,
        imageUrl: `https://picsum.photos/seed/${i}/400/300`,
        status: Math.random() > 0.2 ? "active" : Math.random() > 0.5 ? "inactive" : "pending",
        networkId: networkIds[Math.floor(Math.random() * networkIds.length)],
        totalClicks: 0,
        totalConversions: 0,
      });
      offerIds.push(Number(result.insertId));
    }
    console.log(`✅ Created ${offerIds.length} offers`);

    // Seed Teams
    console.log("👨‍👩‍👧‍👦 Seeding teams...");
    const teamNames = [
      "Alpha Team", "Beta Squad", "Gamma Force", "Delta Warriors", "Epsilon Elite",
      "Zeta Champions", "Eta Masters", "Theta Legends", "Iota Heroes", "Kappa Kings"
    ];
    const teamIds: number[] = [];
    for (const name of teamNames) {
      const result = await db.insert(schema.teams).values({
        name,
        description: `${name} - Top performing affiliate marketing team`,
        memberCount: Math.floor(Math.random() * 10) + 2,
        totalPoints: Math.floor(Math.random() * 50000),
      });
      teamIds.push(Number(result.insertId));
    }
    console.log(`✅ Created ${teamIds.length} teams`);

    // Assign users to teams
    console.log("🔗 Assigning users to teams...");
    let teamMemberCount = 0;
    for (const userId of userIds) {
      if (Math.random() > 0.3) { // 70% of users are in teams
        await db.insert(schema.teamMembers).values({
          teamId: teamIds[Math.floor(Math.random() * teamIds.length)],
          userId,
          role: Math.random() > 0.8 ? "leader" : "member",
        });
        teamMemberCount++;
      }
    }
    console.log(`✅ Created ${teamMemberCount} team memberships`);

    // Seed Badges
    console.log("🏆 Seeding badges...");
    const badgeData = [
      { name: "First Click", description: "Achieved your first click", icon: "🎯", pointsReward: 10 },
      { name: "10 Clicks", description: "Reached 10 clicks", icon: "🔟", pointsReward: 50 },
      { name: "First Conversion", description: "Your first successful conversion", icon: "💰", pointsReward: 100 },
      { name: "10 Conversions", description: "Reached 10 conversions", icon: "💎", pointsReward: 500 },
      { name: "Top Performer", description: "Ranked in top 10 this month", icon: "⭐", pointsReward: 1000 },
      { name: "Team Player", description: "Joined a team", icon: "👥", pointsReward: 50 },
      { name: "Level 5", description: "Reached level 5", icon: "🚀", pointsReward: 200 },
      { name: "Level 10", description: "Reached level 10", icon: "🏅", pointsReward: 500 },
      { name: "Early Bird", description: "Active in first week", icon: "🌅", pointsReward: 25 },
      { name: "Consistent", description: "Active for 30 days straight", icon: "📅", pointsReward: 300 },
      { name: "High Roller", description: "Earned 5000+ SAR", icon: "💵", pointsReward: 1000 },
      { name: "Social Butterfly", description: "Referred 5 friends", icon: "🦋", pointsReward: 250 },
      { name: "Master Marketer", description: "100+ conversions", icon: "🎓", pointsReward: 2000 },
      { name: "Speed Demon", description: "10 conversions in 24 hours", icon: "⚡", pointsReward: 500 },
      { name: "Loyalty King", description: "Active for 6 months", icon: "👑", pointsReward: 1500 },
    ];

    const badgeIds: number[] = [];
    for (const badge of badgeData) {
      const result = await db.insert(schema.badges).values(badge);
      badgeIds.push(Number(result.insertId));
    }
    console.log(`✅ Created ${badgeIds.length} badges`);

    // Award badges to users
    console.log("🎖️ Awarding badges to users...");
    let badgeAwardCount = 0;
    for (const userId of userIds) {
      const numBadges = Math.floor(Math.random() * 5); // 0-4 badges per user
      const awardedBadges = new Set<number>();
      for (let i = 0; i < numBadges; i++) {
        const badgeId = badgeIds[Math.floor(Math.random() * badgeIds.length)];
        if (!awardedBadges.has(badgeId)) {
          await db.insert(schema.userBadges).values({
            userId,
            badgeId,
          });
          awardedBadges.add(badgeId);
          badgeAwardCount++;
        }
      }
    }
    console.log(`✅ Awarded ${badgeAwardCount} badges`);

    // Seed Clicks (distributed over last 30 days)
    console.log("🖱️ Seeding clicks...");
    const now = new Date();
    let clickCount = 0;
    for (let day = 0; day < 30; day++) {
      const clicksThisDay = Math.floor(Math.random() * 30) + 10; // 10-40 clicks per day
      for (let i = 0; i < clicksThisDay; i++) {
        const clickDate = new Date(now);
        clickDate.setDate(clickDate.getDate() - day);
        clickDate.setHours(Math.floor(Math.random() * 24));
        clickDate.setMinutes(Math.floor(Math.random() * 60));

        await db.insert(schema.clicks).values({
          userId: userIds[Math.floor(Math.random() * userIds.length)],
          offerId: offerIds[Math.floor(Math.random() * offerIds.length)],
          clickedAt: clickDate,
          ipAddress: `192.168.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`,
          userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        });
        clickCount++;
      }
    }
    console.log(`✅ Created ${clickCount} clicks`);

    // Seed Conversions (30% of clicks convert)
    console.log("💵 Seeding conversions...");
    const conversionRate = 0.3;
    const expectedConversions = Math.floor(clickCount * conversionRate);
    let conversionCount = 0;

    for (let day = 0; day < 30; day++) {
      const conversionsThisDay = Math.floor((Math.random() * 30 + 10) * conversionRate);
      for (let i = 0; i < conversionsThisDay; i++) {
        const conversionDate = new Date(now);
        conversionDate.setDate(conversionDate.getDate() - day);
        conversionDate.setHours(Math.floor(Math.random() * 24));
        conversionDate.setMinutes(Math.floor(Math.random() * 60));

        const userId = userIds[Math.floor(Math.random() * userIds.length)];
        const offerId = offerIds[Math.floor(Math.random() * offerIds.length)];

        await db.insert(schema.conversions).values({
          userId,
          offerId,
          convertedAt: conversionDate,
          commission: Math.floor(Math.random() * 5000) + 500, // 5-55 SAR in halalas
          status: Math.random() > 0.1 ? "approved" : "pending",
        });
        conversionCount++;
      }
    }
    console.log(`✅ Created ${conversionCount} conversions`);

    console.log("\n🎉 Database seeding completed successfully!");
    console.log(`📊 Summary:`);
    console.log(`   - Users: ${userIds.length}`);
    console.log(`   - Networks: ${networkIds.length}`);
    console.log(`   - Offers: ${offerIds.length}`);
    console.log(`   - Teams: ${teamIds.length}`);
    console.log(`   - Team Members: ${teamMemberCount}`);
    console.log(`   - Badges: ${badgeIds.length}`);
    console.log(`   - Badge Awards: ${badgeAwardCount}`);
    console.log(`   - Clicks: ${clickCount}`);
    console.log(`   - Conversions: ${conversionCount}`);

  } catch (error) {
    console.error("❌ Error seeding database:", error);
    throw error;
  } finally {
    await connection.end();
  }
}

seed()
  .then(() => {
    console.log("✅ Seed script finished");
    process.exit(0);
  })
  .catch((error) => {
    console.error("❌ Seed script failed:", error);
    process.exit(1);
  });
