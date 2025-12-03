import { pgTable, uuid, varchar, text, integer, decimal, timestamp, pgEnum } from "drizzle-orm/pg-core";

export const users = pgTable("users", {
  id: uuid("id").primaryKey().defaultRandom(),
  openId: varchar("open_id", { length: 64 }).notNull().unique(),
  name: text("name"),
  email: varchar("email", { length: 320 }),
  loginMethod: varchar("login_method", { length: 64 }),
  role: varchar("role", { length: 20 }).default("user").notNull(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
  updatedAt: timestamp("updated_at").defaultNow().notNull(),
  lastSignedIn: timestamp("last_signed_in").defaultNow().notNull(),
});

export type InsertUser = typeof users.$inferInsert;
export type SelectUser = typeof users.$inferSelect;

export const userRoleEnum = pgEnum("user_role", ["user", "admin"]);
export const statusEnum = pgEnum("status", ["active", "inactive", "suspended", "pending"]);
export const payoutTypeEnum = pgEnum("payout_type", ["cpa", "cps", "cpl"]);

export const adminUsers = pgTable("admin_users", {
  id: uuid("id").primaryKey().defaultRandom(),
  username: varchar("username", { length: 50 }).notNull().unique(),
  email: varchar("email", { length: 255 }).notNull().unique(),
  passwordHash: varchar("password_hash", { length: 255 }).notNull(),
  fullName: varchar("full_name", { length: 100 }),
  role: varchar("role", { length: 20 }).default("admin").notNull(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
  updatedAt: timestamp("updated_at").defaultNow().notNull(),
});

export const afftokUsers = pgTable("afftok_users", {
  id: uuid("id").primaryKey().defaultRandom(),
  username: varchar("username", { length: 50 }).notNull().unique(),
  email: varchar("email", { length: 255 }).notNull().unique(),
  passwordHash: varchar("password_hash", { length: 255 }).notNull(),
  fullName: varchar("full_name", { length: 100 }),
  avatarUrl: text("avatar_url"),
  role: varchar("role", { length: 20 }).default("user").notNull(),
  status: varchar("status", { length: 20 }).default("active").notNull(),
  points: integer("points").default(0).notNull(),
  level: integer("level").default(1).notNull(),
  totalClicks: integer("total_clicks").default(0).notNull(),
  totalConversions: integer("total_conversions").default(0).notNull(),
  totalEarnings: integer("total_earnings").default(0).notNull(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
  updatedAt: timestamp("updated_at").defaultNow().notNull(),
});

export const networks = pgTable("networks", {
  id: uuid("id").primaryKey().defaultRandom(),
  name: varchar("name", { length: 100 }).notNull(),
  description: text("description"),
  logoUrl: text("logo_url"),
  apiUrl: text("api_url"),
  apiKey: text("api_key"),
  postbackUrl: text("postback_url"),
  hmacSecret: text("hmac_secret"),
  status: varchar("status", { length: 20 }).default("active").notNull(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
  updatedAt: timestamp("updated_at").defaultNow().notNull(),
});

export const offers = pgTable("offers", {
  id: uuid("id").primaryKey().defaultRandom(),
  networkId: uuid("network_id"),
  externalOfferId: varchar("external_offer_id", { length: 100 }),
  // English Fields
  title: varchar("title", { length: 255 }).notNull(),
  description: text("description"),
  // Arabic Fields (للمستخدم العربي)
  titleAr: varchar("title_ar", { length: 255 }),
  descriptionAr: text("description_ar"),
  termsAr: text("terms_ar"),
  // Common Fields
  imageUrl: text("image_url"),
  logoUrl: text("logo_url"),
  destinationUrl: text("destination_url").notNull(),
  category: varchar("category", { length: 50 }),
  payout: integer("payout").default(0).notNull(),
  commission: integer("commission").default(0).notNull(),
  payoutType: varchar("payout_type", { length: 20 }).default("cpa").notNull(),
  rating: decimal("rating", { precision: 3, scale: 2 }).default("0.0").notNull(),
  usersCount: integer("users_count").default(0).notNull(),
  status: varchar("status", { length: 20 }).default("pending").notNull(),
  totalClicks: integer("total_clicks").default(0).notNull(),
  totalConversions: integer("total_conversions").default(0).notNull(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
  updatedAt: timestamp("updated_at").defaultNow().notNull(),
});

export const userOffers = pgTable("user_offers", {
  id: uuid("id").primaryKey().defaultRandom(),
  userId: uuid("user_id").notNull(),
  offerId: uuid("offer_id").notNull(),
  affiliateLink: text("affiliate_link").notNull(),
  shortLink: text("short_link"),
  status: varchar("status", { length: 20 }).default("active").notNull(),
  earnings: integer("earnings").default(0).notNull(),
  joinedAt: timestamp("joined_at").defaultNow().notNull(),
});

export const clicks = pgTable("clicks", {
  id: uuid("id").primaryKey().defaultRandom(),
  userOfferId: uuid("user_offer_id").notNull(),
  ipAddress: varchar("ip_address", { length: 45 }),
  userAgent: text("user_agent"),
  device: varchar("device", { length: 50 }),
  browser: varchar("browser", { length: 50 }),
  os: varchar("os", { length: 50 }),
  country: varchar("country", { length: 2 }),
  city: varchar("city", { length: 100 }),
  clickedAt: timestamp("clicked_at").defaultNow().notNull(),
});

export const conversions = pgTable("conversions", {
  id: uuid("id").primaryKey().defaultRandom(),
  userOfferId: uuid("user_offer_id").notNull(),
  externalConversionId: varchar("external_conversion_id", { length: 100 }),
  amount: integer("amount").default(0).notNull(),
  status: varchar("status", { length: 20 }).default("pending").notNull(),
  ipAddress: varchar("ip_address", { length: 45 }),
  convertedAt: timestamp("converted_at").defaultNow().notNull(),
  approvedAt: timestamp("approved_at"),
});

export const teams = pgTable("teams", {
  id: uuid("id").primaryKey().defaultRandom(),
  name: varchar("name", { length: 100 }).notNull(),
  description: text("description"),
  logoUrl: text("logo_url"),
  ownerId: uuid("owner_id").notNull(),
  maxMembers: integer("max_members").default(10).notNull(),
  memberCount: integer("member_count").default(0).notNull(),
  totalPoints: integer("total_points").default(0).notNull(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
  updatedAt: timestamp("updated_at").defaultNow().notNull(),
});

export const teamMembers = pgTable("team_members", {
  id: uuid("id").primaryKey().defaultRandom(),
  teamId: uuid("team_id").notNull(),
  userId: uuid("user_id").notNull(),
  role: varchar("role", { length: 20 }).default("member").notNull(),
  points: integer("points").default(0).notNull(),
  joinedAt: timestamp("joined_at").defaultNow().notNull(),
});

export const badges = pgTable("badges", {
  id: uuid("id").primaryKey().defaultRandom(),
  name: varchar("name", { length: 100 }).notNull(),
  description: text("description"),
  iconUrl: text("icon_url"),
  requirement: text("requirement"),
  points: integer("points").default(0).notNull(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
});

export const userBadges = pgTable("user_badges", {
  id: uuid("id").primaryKey().defaultRandom(),
  userId: uuid("user_id").notNull(),
  badgeId: uuid("badge_id").notNull(),
  earnedAt: timestamp("earned_at").defaultNow().notNull(),
});
