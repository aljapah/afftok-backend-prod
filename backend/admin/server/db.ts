import { eq, sql } from "drizzle-orm";
import { drizzle } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import { users, networks, offers, badges, afftokUsers, teams, clicks, conversions } from "../drizzle/schema";
import type { InsertUser } from "../drizzle/schema";
import { ENV } from './_core/env';

let _db: ReturnType<typeof drizzle> | null = null;
let _client: ReturnType<typeof postgres> | null = null;

export async function getDb() {
  if (!_db && process.env.DATABASE_URL) {
    try {
      _client = postgres(process.env.DATABASE_URL);
      _db = drizzle(_client);
    } catch (error) {
      console.warn("[Database] Failed to connect:", error);
      _db = null;
    }
  }
  return _db;
}

export async function upsertUser(user: InsertUser): Promise<void> {
  if (!user.openId) {
    throw new Error("User openId is required for upsert");
  }

  const db = await getDb();
  if (!db) {
    console.warn("[Database] Cannot upsert user: database not available");
    return;
  }

  try {
    const values: InsertUser = {
      openId: user.openId,
    };
    const updateSet: Record<string, unknown> = {};

    const textFields = ["name", "email", "loginMethod"] as const;
    type TextField = (typeof textFields)[number];

    const assignNullable = (field: TextField) => {
      if (user[field] !== undefined) {
        values[field] = user[field];
        updateSet[field] = user[field];
      }
    };

    assignNullable("name");
    assignNullable("email");
    assignNullable("loginMethod");

    await db.insert(users).values(values).onConflictDoUpdate({
      target: users.openId,
      set: {
        ...updateSet,
        updatedAt: new Date(),
        lastSignedIn: new Date(),
      },
    });
  } catch (error) {
    console.error("[Database] Error upserting user:", error);
  }
}

export async function getAllAfftokUsers() {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  return db.select().from(afftokUsers);
}

export async function createAfftokUser(data: any) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const newUser = {
    username: data.username,
    email: data.email,
    passwordHash: data.password,
    fullName: data.fullName || null,
    role: data.role || 'user',
  };

  const result = await db.insert(afftokUsers).values(newUser).returning();
  return result[0];
}

export async function updateAfftokUser(data: { id: string; [key: string]: any }) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const { id, ...updates } = data;
  await db.update(afftokUsers).set(updates).where(eq(afftokUsers.id, id));
  return { success: true };
}

export async function deleteAfftokUser(id: string) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");
  
  await db.delete(afftokUsers).where(eq(afftokUsers.id, id));
  return { success: true };
}

export async function getClicksAnalytics(days: number = 30) {
  const db = await getDb();
  if (!db) return [];
  const { sql } = await import("drizzle-orm");
  
  const startDate = new Date();
  startDate.setDate(startDate.getDate() - days);
  
  const result = await db.execute(sql`
    SELECT 
      DATE(clicked_at) as date,
      COUNT(*) as count
    FROM clicks
    WHERE clicked_at >= ${startDate}
    GROUP BY DATE(clicked_at)
    ORDER BY date ASC
  `);
  
  return (result.rows as any[]).map((row: any) => ({
    date: row.date,
    clicks: parseInt(row.count),
  }));
}

export async function getConversionsAnalytics(days: number = 30) {
  const db = await getDb();
  if (!db) return [];
  const { sql } = await import("drizzle-orm");
  
  const startDate = new Date();
  startDate.setDate(startDate.getDate() - days);
  
  const result = await db.execute(sql`
    SELECT 
      DATE(converted_at) as date,
      COUNT(*) as count
    FROM conversions
    WHERE converted_at >= ${startDate}
    GROUP BY DATE(converted_at)
    ORDER BY date ASC
  `);
  
  return (result.rows as any[]).map((row: any) => ({
    date: row.date,
    conversions: parseInt(row.count),
  }));
}

export async function getAllNetworks() {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  return db.select().from(networks);
}

export async function createNetwork(data: any) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const newNetwork = {
    name: data.name,
    description: data.description || null,
    logoUrl: data.logoUrl || null,
    apiUrl: data.apiUrl || null,
    apiKey: data.apiKey || null,
    postbackUrl: data.postbackUrl || null,
    hmacSecret: data.hmacSecret || null,
  };

  const result = await db.insert(networks).values(newNetwork).returning();
  return result[0];
}

export async function updateNetwork(data: { id: string; [key: string]: any }) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const { id, ...updates } = data;
  await db.update(networks).set(updates).where(eq(networks.id, id));
  return { success: true };
}

export async function deleteNetwork(id: string) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");
  
  await db.delete(networks).where(eq(networks.id, id));
  return { success: true };
}

export async function getAllOffers() {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  return db.select().from(offers);
}

export async function createOffer(data: any) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const newOffer = {
    title: data.title,
    description: data.description || null,
    imageUrl: data.imageUrl || null,
    destinationUrl: data.destinationUrl,
    category: data.category || null,
    payout: data.payout,
    commission: data.commission,
    networkId: data.networkId || null,
  };

  const result = await db.insert(offers).values(newOffer).returning();
  return result[0];
}

export async function updateOffer(data: { id: string; [key: string]: any }) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const { id, ...updates } = data;
  await db.update(offers).set(updates).where(eq(offers.id, id));
  return { success: true };
}

export async function deleteOffer(id: string) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");
  
  await db.delete(offers).where(eq(offers.id, id));
  return { success: true };
}

export async function getDashboardStats() {
  const db = await getDb();
  if (!db) throw new Error("Database not available");
  const { sql } = await import("drizzle-orm");

  try {
    // Query each count separately to avoid SQL execution issues
    const usersCount = await db.execute(sql`SELECT COUNT(*) as count FROM afftok_users`);
    const offersCount = await db.execute(sql`SELECT COUNT(*) as count FROM offers`);
    const networksCount = await db.execute(sql`SELECT COUNT(*) as count FROM networks`);
    const teamsCount = await db.execute(sql`SELECT COUNT(*) as count FROM teams`);
    const clicksCount = await db.execute(sql`SELECT COUNT(*) as count FROM clicks`);
    const conversionsCount = await db.execute(sql`SELECT COUNT(*) as count FROM conversions`);
    const earningsSum = await db.execute(sql`SELECT COALESCE(SUM(amount), 0) as sum FROM conversions WHERE status = 'approved'`);

    const result = {
      totalUsers: parseInt((usersCount as any)?.[0]?.count || '0') || 0,
      totalOffers: parseInt((offersCount as any)?.[0]?.count || '0') || 0,
      totalNetworks: parseInt((networksCount as any)?.[0]?.count || '0') || 0,
      totalTeams: parseInt((teamsCount as any)?.[0]?.count || '0') || 0,
      totalClicks: parseInt((clicksCount as any)?.[0]?.count || '0') || 0,
      totalConversions: parseInt((conversionsCount as any)?.[0]?.count || '0') || 0,
      totalEarnings: parseFloat((earningsSum as any)?.[0]?.sum || '0') || 0,
    };
    return result;
  } catch (error) {
    console.error('Error fetching dashboard stats:', error);
    return {
      totalUsers: 0,
      totalOffers: 0,
      totalNetworks: 0,
      totalTeams: 0,
      totalClicks: 0,
      totalConversions: 0,
      totalEarnings: 0,
    };
  }
}

export async function getAllTeams() {
  const db = await getDb();
  if (!db) throw new Error("Database not available");
  
  return db.select().from(teams);
}

export async function createTeam(data: { name: string; description?: string | null; logoUrl?: string | null; ownerId: string; maxMembers?: number }) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const newTeam = {
    name: data.name,
    description: data.description || null,
    logoUrl: data.logoUrl || null,
    ownerId: data.ownerId,
    maxMembers: data.maxMembers || 10,
    memberCount: 0,
    totalPoints: 0,
  };

  const result = await db.insert(teams).values(newTeam).returning();
  return result[0];
}

export async function updateTeam(data: { id: string; [key: string]: any }) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const { id, ...updates } = data;
  await db.update(teams).set(updates).where(eq(teams.id, id));
  return { success: true };
}

export async function deleteTeam(id: string) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");
  
  await db.delete(teams).where(eq(teams.id, id));
  return { success: true };
}

export async function getAllBadges() {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  return db.select().from(badges);
}

export async function createBadge(data: any) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const newBadge = {
    name: data.name,
    description: data.description || null,
    iconUrl: data.iconUrl || null,
    points: data.pointsReward,
  };

  const result = await db.insert(badges).values(newBadge).returning();
  return result[0];
}

export async function updateBadge(data: { id: string; [key: string]: any }) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");

  const { id, ...updates } = data;
  
  if (updates.pointsReward !== undefined) {
    updates.points = updates.pointsReward;
    delete updates.pointsReward;
  }
  
  await db.update(badges).set(updates).where(eq(badges.id, id));
  return { success: true };
}

export async function deleteBadge(id: string) {
  const db = await getDb();
  if (!db) throw new Error("Database not available");
  
  await db.delete(badges).where(eq(badges.id, id));
  return { success: true };
}
