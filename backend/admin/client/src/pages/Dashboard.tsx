import DashboardLayout from "@/components/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { trpc } from "@/lib/trpc";
import { Users, Tag, MousePointerClick, TrendingUp, BarChart3, Network, UsersRound } from "lucide-react";
import EmptyState from "@/components/EmptyState";
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { useMemo } from "react";

export default function Dashboard() {
  const { data: stats, isLoading } = trpc.dashboard.stats.useQuery();
  const { data: clicksRaw } = trpc.dashboard.clicksAnalytics.useQuery({ days: 7 });
  const { data: conversionsRaw } = trpc.dashboard.conversionsAnalytics.useQuery({ days: 7 });
  
  // Format data for charts
  const clicksData = useMemo(() => {
    if (!clicksRaw || clicksRaw.length === 0) return [];
    const formatted = clicksRaw.map((item: any) => ({
      name: new Date(item.date).toLocaleDateString('en-US', { weekday: 'short' }),
      clicks: item.clicks || 0,
    }));
    // Return empty if all clicks are 0
    const hasData = formatted.some((item: any) => item.clicks > 0);
    return hasData ? formatted : [];
  }, [clicksRaw]);
  
  const conversionsData = useMemo(() => {
    if (!conversionsRaw || conversionsRaw.length === 0) return [];
    const formatted = conversionsRaw.map((item: any) => ({
      name: new Date(item.date).toLocaleDateString('en-US', { weekday: 'short' }),
      conversions: item.conversions || 0,
    }));
    // Return empty if all conversions are 0
    const hasData = formatted.some((item: any) => item.conversions > 0);
    return hasData ? formatted : [];
  }, [conversionsRaw]);

  const statsCards = [
    {
      title: "Total Users",
      value: stats?.totalUsers ?? 0,
      icon: Users,
      color: "text-blue-500",
    },
    {
      title: "Total Offers",
      value: stats?.totalOffers ?? 0,
      icon: Tag,
      color: "text-green-500",
    },
    {
      title: "Total Networks",
      value: stats?.totalNetworks ?? 0,
      icon: Network,
      color: "text-orange-500",
    },
    {
      title: "Total Teams",
      value: stats?.totalTeams ?? 0,
      icon: UsersRound,
      color: "text-cyan-500",
    },
    {
      title: "Total Clicks",
      value: stats?.totalClicks ?? 0,
      icon: MousePointerClick,
      color: "text-purple-500",
    },
    {
      title: "Total Conversions",
      value: stats?.totalConversions ?? 0,
      icon: TrendingUp,
      color: "text-pink-500",
    },
  ];

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <h1 className="text-3xl font-bold">Dashboard</h1>

        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {statsCards.map((stat) => {
            const Icon = stat.icon;
            return (
              <Card key={stat.title}>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
                  <Icon className={`h-4 w-4 ${stat.color}`} />
                </CardHeader>
                <CardContent>
                  {isLoading ? (
                    <div className="text-2xl font-bold animate-pulse">...</div>
                  ) : (
                    <div className="text-2xl font-bold">{(stat.value ?? 0).toLocaleString()}</div>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>

        {/* Charts */}
        <div className="grid gap-4 md:grid-cols-2">
          {/* Clicks Chart */}
          <Card>
            <CardHeader>
              <CardTitle>Clicks Overview (Last 7 Days)</CardTitle>
            </CardHeader>
            <CardContent>
              {clicksData.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={clicksData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                    <XAxis dataKey="name" stroke="#888" />
                    <YAxis stroke="#888" />
                    <Tooltip 
                      contentStyle={{ backgroundColor: '#1a1a1a', border: '1px solid #333' }}
                      labelStyle={{ color: '#fff' }}
                    />
                    <Legend />
                    <Line 
                      type="monotone" 
                      dataKey="clicks" 
                      stroke="#8b5cf6" 
                      strokeWidth={2}
                      dot={{ fill: '#8b5cf6' }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <EmptyState
                  icon={BarChart3}
                  title="No Click Data Yet"
                  description="Start promoting offers to see click analytics here. Run 'pnpm seed' to add sample data."
                />
              )}
            </CardContent>
          </Card>

          {/* Conversions Chart */}
          <Card>
            <CardHeader>
              <CardTitle>Conversions Overview (Last 7 Days)</CardTitle>
            </CardHeader>
            <CardContent>
              {conversionsData.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={conversionsData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                    <XAxis dataKey="name" stroke="#888" />
                    <YAxis stroke="#888" />
                    <Tooltip 
                      contentStyle={{ backgroundColor: '#1a1a1a', border: '1px solid #333' }}
                      labelStyle={{ color: '#fff' }}
                    />
                    <Legend />
                    <Bar dataKey="conversions" fill="#ec4899" radius={[8, 8, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <EmptyState
                  icon={TrendingUp}
                  title="No Conversion Data Yet"
                  description="Conversions will appear here once users complete offers. Run 'pnpm seed' to add sample data."
                />
              )}
            </CardContent>
          </Card>
        </div>

        {/* Welcome Card */}
        <Card>
          <CardHeader>
            <CardTitle>Welcome to AffTok Admin Panel</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">
              Manage your affiliate marketing platform from this dashboard.
            </p>
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
}
