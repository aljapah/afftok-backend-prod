import DashboardLayout from "@/components/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { trpc } from "@/lib/trpc";
import { TrendingUp, DollarSign, MousePointerClick, Target, Download } from "lucide-react";
import { 
  LineChart, 
  Line, 
  BarChart, 
  Bar, 
  PieChart,
  Pie,
  Cell,
  AreaChart,
  Area,
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer, 
  Legend 
} from 'recharts';
import { useMemo } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { exportToCSV, exportToExcel } from "@/lib/export";
import { toast } from "sonner";

const COLORS = ['#8b5cf6', '#ec4899', '#10b981', '#f59e0b', '#3b82f6'];

export default function Analytics() {
  const { data: stats } = trpc.dashboard.stats.useQuery();
  const { data: clicksRaw } = trpc.analytics.getClicks.useQuery({ days: 30 });
  const { data: conversionsRaw } = trpc.analytics.getConversions.useQuery({ days: 30 });
  
  // Format data for charts
  const clicksData = useMemo(() => {
    if (!clicksRaw) return [];
    return clicksRaw.map((item: any) => ({
      date: new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      clicks: item.clicks,
    }));
  }, [clicksRaw]);
  
  const conversionsData = useMemo(() => {
    if (!conversionsRaw) return [];
    return conversionsRaw.map((item: any) => ({
      date: new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      conversions: item.conversions,
    }));
  }, [conversionsRaw]);

  // Combined data for area chart
  const combinedData = useMemo(() => {
    if (!clicksRaw || !conversionsRaw) return [];
    const dataMap = new Map();
    
    clicksRaw.forEach((item: any) => {
      const date = new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
      dataMap.set(date, { date, clicks: item.clicks, conversions: 0 });
    });
    
    conversionsRaw.forEach((item: any) => {
      const date = new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
      const existing = dataMap.get(date) || { date, clicks: 0, conversions: 0 };
      existing.conversions = item.conversions;
      dataMap.set(date, existing);
    });
    
    return Array.from(dataMap.values());
  }, [clicksRaw, conversionsRaw]);

  // Mock data for pie chart (replace with real data later)
  const offerDistribution = [
    { name: 'E-commerce', value: 35 },
    { name: 'Subscription', value: 25 },
    { name: 'Digital Products', value: 20 },
    { name: 'Services', value: 15 },
    { name: 'Other', value: 5 },
  ];

  // Mock top performers
  const topPerformers = [
    { id: 1, name: 'Ahmed Hassan', clicks: 1250, conversions: 85, revenue: 4250, conversionRate: 6.8 },
    { id: 2, name: 'Sara Ali', clicks: 980, conversions: 72, revenue: 3600, conversionRate: 7.3 },
    { id: 3, name: 'Mohammed Khalid', clicks: 850, conversions: 58, revenue: 2900, conversionRate: 6.8 },
    { id: 4, name: 'Fatima Omar', clicks: 720, conversions: 51, revenue: 2550, conversionRate: 7.1 },
    { id: 5, name: 'Youssef Ibrahim', clicks: 650, conversions: 42, revenue: 2100, conversionRate: 6.5 },
  ];

  const statsCards = [
    {
      title: "Total Revenue",
      value: "$" + (((stats?.totalConversions ?? 0) as number) * 50).toLocaleString(),
      icon: DollarSign,
      color: "text-green-500",
      change: "+12.5%",
    },
    {
      title: "Avg. Conversion Rate",
      value: (() => {
        const clicks = (stats?.totalClicks ?? 0) as number;
        const conversions = (stats?.totalConversions ?? 0) as number;
        return clicks > 0 ? ((conversions / clicks) * 100).toFixed(2) + "%" : "0%";
      })(),
      icon: Target,
      color: "text-blue-500",
      change: "+2.3%",
    },
    {
      title: "Total Clicks",
      value: (stats?.totalClicks ?? 0).toLocaleString(),
      icon: MousePointerClick,
      color: "text-purple-500",
      change: "+18.2%",
    },
    {
      title: "Growth Rate",
      value: "+24.5%",
      icon: TrendingUp,
      color: "text-pink-500",
      change: "+5.1%",
    },
  ];

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold">Analytics</h1>
          <div className="flex items-center gap-4">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                if (combinedData && combinedData.length > 0) {
                  exportToExcel(combinedData, `afftok-analytics-${new Date().toISOString().split('T')[0]}`);
                  toast.success('Analytics exported successfully');
                } else {
                  toast.error('No analytics data to export');
                }
              }}
            >
              <Download className="h-4 w-4 mr-2" />
              Export Excel
            </Button>
            <div className="text-sm text-muted-foreground">Last 30 Days</div>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {statsCards.map((stat) => {
            const Icon = stat.icon;
            return (
              <Card key={stat.title}>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
                  <Icon className={`h-4 w-4 ${stat.color}`} />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{stat.value}</div>
                  <p className="text-xs text-green-500 mt-1">{stat.change} from last month</p>
                </CardContent>
              </Card>
            );
          })}
        </div>

        {/* Combined Area Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Performance Overview (Last 30 Days)</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={350}>
              <AreaChart data={combinedData.length > 0 ? combinedData : [{ date: 'No Data', clicks: 0, conversions: 0 }]}>
                <defs>
                  <linearGradient id="colorClicks" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="colorConversions" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#ec4899" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#ec4899" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                <XAxis dataKey="date" stroke="#888" />
                <YAxis stroke="#888" />
                <Tooltip 
                  contentStyle={{ backgroundColor: '#1a1a1a', border: '1px solid #333' }}
                  labelStyle={{ color: '#fff' }}
                />
                <Legend />
                <Area 
                  type="monotone" 
                  dataKey="clicks" 
                  stroke="#8b5cf6" 
                  fillOpacity={1}
                  fill="url(#colorClicks)"
                />
                <Area 
                  type="monotone" 
                  dataKey="conversions" 
                  stroke="#ec4899" 
                  fillOpacity={1}
                  fill="url(#colorConversions)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* Charts Row */}
        <div className="grid gap-4 md:grid-cols-2">
          {/* Pie Chart - Offer Distribution */}
          <Card>
            <CardHeader>
              <CardTitle>Offers by Category</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={offerDistribution}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="value"
                  >
                    {offerDistribution.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#1a1a1a', border: '1px solid #333' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* Bar Chart - Monthly Comparison */}
          <Card>
            <CardHeader>
              <CardTitle>Clicks vs Conversions</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={combinedData.slice(-7).length > 0 ? combinedData.slice(-7) : [{ date: 'No Data', clicks: 0, conversions: 0 }]}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                  <XAxis dataKey="date" stroke="#888" />
                  <YAxis stroke="#888" />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#1a1a1a', border: '1px solid #333' }}
                    labelStyle={{ color: '#fff' }}
                  />
                  <Legend />
                  <Bar dataKey="clicks" fill="#8b5cf6" radius={[8, 8, 0, 0]} />
                  <Bar dataKey="conversions" fill="#ec4899" radius={[8, 8, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </div>

        {/* Top Performers Table */}
        <Card>
          <CardHeader>
            <CardTitle>Top Performers</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Rank</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Clicks</TableHead>
                  <TableHead>Conversions</TableHead>
                  <TableHead>Revenue</TableHead>
                  <TableHead>Conv. Rate</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {topPerformers.map((performer, index) => (
                  <TableRow key={performer.id}>
                    <TableCell>
                      <Badge variant={index === 0 ? "default" : "outline"}>
                        #{index + 1}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-medium">{performer.name}</TableCell>
                    <TableCell>{(performer.clicks ?? 0).toLocaleString()}</TableCell>
                    <TableCell>{performer.conversions ?? 0}</TableCell>
                    <TableCell className="text-green-500 font-semibold">
                      ${(performer.revenue ?? 0).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{performer.conversionRate}%</Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
}
