import { COOKIE_NAME } from "@shared/const";
import { getSessionCookieOptions } from "./_core/cookies";
import { systemRouter } from "./_core/systemRouter";
import { publicProcedure, router } from "./_core/trpc";
import { z } from "zod";

export const appRouter = router({
  system: systemRouter,
  auth: router({
    me: publicProcedure.query(opts => opts.ctx.user),
    logout: publicProcedure.mutation(({ ctx }) => {
      const cookieOptions = getSessionCookieOptions(ctx.req);
      ctx.res.clearCookie(COOKIE_NAME, { ...cookieOptions, maxAge: -1 });
      return {
        success: true,
      } as const;
    }),
  }),

  dashboard: router({
    stats: publicProcedure.query(async () => {
      const { getDashboardStats } = await import("./db");
      return getDashboardStats();
    }),
    clicksAnalytics: publicProcedure
      .input((input: unknown) => {
        return z.object({ days: z.number().int().min(1).max(365).optional() }).parse(input);
      })
      .query(async ({ input }) => {
        const { getClicksAnalytics } = await import("./db");
        return getClicksAnalytics(input.days);
      }),
    conversionsAnalytics: publicProcedure
      .input((input: unknown) => {
        return z.object({ days: z.number().int().min(1).max(365).optional() }).parse(input);
      })
      .query(async ({ input }) => {
        const { getConversionsAnalytics } = await import("./db");
        return getConversionsAnalytics(input.days);
      }),
  }),
  
  users: router({
    list: publicProcedure.query(async () => {
      const { getAllAfftokUsers } = await import("./db");
      return getAllAfftokUsers();
    }),
    create: publicProcedure
      .input((input: unknown) => {
        return z.object({
          username: z.string().min(3),
          email: z.string().email(),
          password: z.string().min(6),
          fullName: z.string().optional(),
          role: z.enum(['user', 'admin']).default('user'),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { createAfftokUser } = await import("./db");
        return createAfftokUser(input);
      }),
    update: publicProcedure
      .input((input: unknown) => {
        return z.object({
          id: z.string(),
          username: z.string().min(3).optional(),
          email: z.string().email().optional(),
          fullName: z.string().optional(),
          role: z.enum(['user', 'admin']).default('user').optional(),
          status: z.enum(['active', 'inactive', 'suspended', 'pending']).optional(),
          points: z.number().int().min(0).optional(),
          level: z.number().int().min(1).optional(),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { updateAfftokUser } = await import("./db");
        return updateAfftokUser(input);
      }),
    delete: publicProcedure
      .input((input: unknown) => {
        return z.object({ id: z.string() }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { deleteAfftokUser } = await import("./db");
        return deleteAfftokUser(input.id);
      }),
  }),
  
  networks: router({
    list: publicProcedure.query(async () => {
      const { getAllNetworks } = await import("./db");
      return getAllNetworks();
    }),
    create: publicProcedure
      .input((input: unknown) => {
        return z.object({
          name: z.string().min(1),
          description: z.string().optional(),
          logoUrl: z.string().url().optional(),
          apiUrl: z.string().url().optional(),
          apiKey: z.string().optional(),
          postbackUrl: z.string().url().optional(),
          hmacSecret: z.string().optional(),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { createNetwork } = await import("./db");
        return createNetwork(input);
      }),
    update: publicProcedure
      .input((input: unknown) => {
        return z.object({
          id: z.string(),
          name: z.string().min(1).optional(),
          description: z.string().optional(),
          logoUrl: z.string().url().optional(),
          apiUrl: z.string().url().optional(),
          apiKey: z.string().optional(),
          postbackUrl: z.string().url().optional(),
          hmacSecret: z.string().optional(),
          status: z.enum(['active', 'inactive', 'suspended', 'pending']).optional(),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { updateNetwork } = await import("./db");
        return updateNetwork(input);
      }),
    delete: publicProcedure
      .input((input: unknown) => {
        return z.object({ id: z.string() }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { deleteNetwork } = await import("./db");
        return deleteNetwork(input.id);
      }),
  }),
  
  offers: router({
    list: publicProcedure.query(async () => {
      const { getAllOffers } = await import("./db");
      return getAllOffers();
    }),
    create: publicProcedure
      .input((input: unknown) => {
        return z.object({
          title: z.string().min(1),
          description: z.string().optional(),
          imageUrl: z.string().url().optional(),
          destinationUrl: z.string().url(),
          category: z.string().optional(),
          payout: z.number().int().min(0),
          commission: z.number().int().min(0),
          networkId: z.string().optional(),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { createOffer } = await import("./db");
        return createOffer(input);
      }),
    update: publicProcedure
      .input((input: unknown) => {
        return z.object({
          id: z.string(),
          title: z.string().min(1).optional(),
          description: z.string().optional(),
          imageUrl: z.string().url().optional(),
          destinationUrl: z.string().url().optional(),
          category: z.string().optional(),
          payout: z.number().int().min(0).optional(),
          commission: z.number().int().min(0).optional(),
          status: z.enum(['active', 'inactive', 'suspended', 'pending']).optional(),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { updateOffer } = await import("./db");
        return updateOffer(input);
      }),
    delete: publicProcedure
      .input((input: unknown) => {
        return z.object({ id: z.string() }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { deleteOffer } = await import("./db");
        return deleteOffer(input.id);
      }),
  }),
  
  teams: router({
    list: publicProcedure.query(async () => {
      const { getAllTeams } = await import("./db");
      return getAllTeams();
    }),
    create: publicProcedure
      .input((input: unknown) => {
        return z.object({
          name: z.string().min(1),
          description: z.string().optional(),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { createTeam, getAllAfftokUsers } = await import("./db");
        
        const users = await getAllAfftokUsers();
        if (!users || users.length === 0) {
          throw new Error("No users found. Please create a user first.");
        }
        
        const defaultOwnerId = users[0].id;
        
        return createTeam({
            ...input,
            ownerId: input.ownerId || defaultOwnerId,
        });
      }),
    update: publicProcedure
      .input((input: unknown) => {
        return z.object({
          id: z.string(),
          name: z.string().min(1).optional(),
          description: z.string().optional(),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { updateTeam } = await import("./db");
        return updateTeam(input);
      }),
    delete: publicProcedure
      .input((input: unknown) => {
        return z.object({ id: z.string() }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { deleteTeam } = await import("./db");
        return deleteTeam(input.id);
      }),
  }),
  
  badges: router({
    list: publicProcedure.query(async () => {
      const { getAllBadges } = await import("./db");
      return getAllBadges();
    }),
    create: publicProcedure
      .input((input: unknown) => {
        return z.object({
          name: z.string().min(1),
          description: z.string().optional(),
          iconUrl: z.string().optional(),
          pointsReward: z.number().int().min(0),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { createBadge } = await import("./db");
        return createBadge(input);
      }),
    update: publicProcedure
      .input((input: unknown) => {
        return z.object({
          id: z.string(),
          name: z.string().min(1).optional(),
          description: z.string().optional(),
          iconUrl: z.string().optional(),
          pointsReward: z.number().int().min(0).optional(),
        }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { updateBadge } = await import("./db");
        return updateBadge(input);
      }),
    delete: publicProcedure
      .input((input: unknown) => {
        return z.object({ id: z.string() }).parse(input);
      })
      .mutation(async ({ input }) => {
        const { deleteBadge } = await import("./db");
        return deleteBadge(input.id);
      }),
  }),
});
