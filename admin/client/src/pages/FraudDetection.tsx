import DashboardLayout from "@/components/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { 
  Shield, 
  AlertTriangle, 
  Ban, 
  Bot,
  Globe,
  Clock,
  TrendingUp,
  RefreshCw,
  Search,
  Eye,
  Lock,
  Unlock
} from "lucide-react";
import { useState, useEffect } from "react";
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell 
} from 'recharts';

// Mock data
const mockFraudStats = {
  totalBlocked: 4520,
  botsBlocked: 2150,
  geoBlocked: 1890,
  rateLimited: 480,
  riskScore: 15,
};

const mockRiskyIPs = [
  { ip: "192.168.1.100", attempts: 1250, lastSeen: "2 min ago", country: "CN", riskScore: 95, blocked: true },
  { ip: "10.0.0.55", attempts: 890, lastSeen: "5 min ago", country: "RU", riskScore: 88, blocked: true },
  { ip: "172.16.0.22", attempts: 450, lastSeen: "12 min ago", country: "IN", riskScore: 72, blocked: false },
  { ip: "203.45.67.89", attempts: 320, lastSeen: "25 min ago", country: "BR", riskScore: 65, blocked: false },
  { ip: "45.123.45.67", attempts: 180, lastSeen: "1 hour ago", country: "US", riskScore: 45, blocked: false },
];

const mockFraudEvents = [
  { id: "1", type: "bot_detected", ip: "192.168.1.100", country: "CN", timestamp: "2024-12-02 14:32:15", details: "Headless browser detected" },
  { id: "2", type: "rate_limit", ip: "10.0.0.55", country: "RU", timestamp: "2024-12-02 14:30:22", details: "100+ requests/min" },
  { id: "3", type: "geo_block", ip: "172.16.0.22", country: "IN", timestamp: "2024-12-02 14:28:45", details: "Blocked country" },
  { id: "4", type: "replay_attempt", ip: "203.45.67.89", country: "BR", timestamp: "2024-12-02 14:25:10", details: "Duplicate nonce detected" },
  { id: "5", type: "invalid_signature", ip: "45.123.45.67", country: "US", timestamp: "2024-12-02 14:20:33", details: "HMAC validation failed" },
];

const hourlyData = [
  { hour: "00:00", blocked: 120 },
  { hour: "04:00", blocked: 85 },
  { hour: "08:00", blocked: 210 },
  { hour: "12:00", blocked: 380 },
  { hour: "16:00", blocked: 450 },
  { hour: "20:00", blocked: 320 },
];

const fraudTypeData = [
  { name: "Bot Traffic", value: 45, color: "#ef4444" },
  { name: "Geo Blocked", value: 30, color: "#f59e0b" },
  { name: "Rate Limited", value: 15, color: "#3b82f6" },
  { name: "Invalid Links", value: 10, color: "#8b5cf6" },
];

export default function FraudDetection() {
  const [searchQuery, setSearchQuery] = useState("");
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState(new Date());

  const handleRefresh = () => {
    setIsRefreshing(true);
    setTimeout(() => {
      setLastUpdated(new Date());
      setIsRefreshing(false);
    }, 1000);
  };

  const getEventTypeBadge = (type: string) => {
    switch (type) {
      case "bot_detected":
        return <Badge className="bg-red-500/10 text-red-500"><Bot className="h-3 w-3 mr-1" /> Bot</Badge>;
      case "rate_limit":
        return <Badge className="bg-yellow-500/10 text-yellow-500"><Clock className="h-3 w-3 mr-1" /> Rate Limit</Badge>;
      case "geo_block":
        return <Badge className="bg-blue-500/10 text-blue-500"><Globe className="h-3 w-3 mr-1" /> Geo Block</Badge>;
      case "replay_attempt":
        return <Badge className="bg-purple-500/10 text-purple-500"><RefreshCw className="h-3 w-3 mr-1" /> Replay</Badge>;
      case "invalid_signature":
        return <Badge className="bg-orange-500/10 text-orange-500"><AlertTriangle className="h-3 w-3 mr-1" /> Invalid Sig</Badge>;
      default:
        return <Badge variant="outline">{type}</Badge>;
    }
  };

  const getRiskBadge = (score: number) => {
    if (score >= 80) return <Badge variant="destructive">{score}%</Badge>;
    if (score >= 50) return <Badge className="bg-yellow-500/10 text-yellow-500">{score}%</Badge>;
    return <Badge variant="outline">{score}%</Badge>;
  };

  const stats = [
    { title: "Total Blocked", value: mockFraudStats.totalBlocked.toLocaleString(), icon: Ban, color: "text-red-500", change: "+12%" },
    { title: "Bots Blocked", value: mockFraudStats.botsBlocked.toLocaleString(), icon: Bot, color: "text-orange-500", change: "+8%" },
    { title: "Geo Blocked", value: mockFraudStats.geoBlocked.toLocaleString(), icon: Globe, color: "text-blue-500", change: "+15%" },
    { title: "Risk Score", value: `${mockFraudStats.riskScore}%`, icon: Shield, color: "text-green-500", change: "-3%" },
  ];

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Fraud Detection</h1>
            <p className="text-muted-foreground mt-1">
              Monitor and block suspicious traffic
            </p>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-muted-foreground">
              Last updated: {lastUpdated.toLocaleTimeString()}
            </span>
            <Button onClick={handleRefresh} disabled={isRefreshing} variant="outline">
              <RefreshCw className={`h-4 w-4 mr-2 ${isRefreshing ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </div>

        {/* Stats */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {stats.map((stat) => {
            const Icon = stat.icon;
            return (
              <Card key={stat.title}>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
                  <Icon className={`h-4 w-4 ${stat.color}`} />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{stat.value}</div>
                  <p className={`text-xs ${stat.change.startsWith('+') ? 'text-red-500' : 'text-green-500'}`}>
                    {stat.change} from last hour
                  </p>
                </CardContent>
              </Card>
            );
          })}
        </div>

        {/* Charts */}
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Blocked Traffic (24h)</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={250}>
                <BarChart data={hourlyData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                  <XAxis dataKey="hour" stroke="#888" fontSize={10} />
                  <YAxis stroke="#888" fontSize={10} />
                  <Tooltip contentStyle={{ backgroundColor: '#1a1a1a', border: '1px solid #333' }} />
                  <Bar dataKey="blocked" fill="#ef4444" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Fraud by Type</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie
                    data={fraudTypeData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={5}
                    dataKey="value"
                    label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  >
                    {fraudTypeData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip contentStyle={{ backgroundColor: '#1a1a1a', border: '1px solid #333' }} />
                </PieChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </div>

        {/* Risky IPs */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Risky IP Addresses</CardTitle>
                <CardDescription>Top suspicious IPs by attempt count</CardDescription>
              </div>
              <div className="relative w-64">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search IPs..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>IP Address</TableHead>
                  <TableHead>Country</TableHead>
                  <TableHead>Attempts</TableHead>
                  <TableHead>Last Seen</TableHead>
                  <TableHead>Risk Score</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-[100px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {mockRiskyIPs.map((ip) => (
                  <TableRow key={ip.ip}>
                    <TableCell className="font-mono">{ip.ip}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{ip.country}</Badge>
                    </TableCell>
                    <TableCell className="text-red-500 font-medium">{ip.attempts.toLocaleString()}</TableCell>
                    <TableCell className="text-muted-foreground">{ip.lastSeen}</TableCell>
                    <TableCell>{getRiskBadge(ip.riskScore)}</TableCell>
                    <TableCell>
                      {ip.blocked ? (
                        <Badge variant="destructive"><Ban className="h-3 w-3 mr-1" /> Blocked</Badge>
                      ) : (
                        <Badge variant="outline">Active</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button variant="ghost" size="icon">
                          <Eye className="h-4 w-4" />
                        </Button>
                        {ip.blocked ? (
                          <Button variant="ghost" size="icon" className="text-green-500">
                            <Unlock className="h-4 w-4" />
                          </Button>
                        ) : (
                          <Button variant="ghost" size="icon" className="text-red-500">
                            <Lock className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Recent Events */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Fraud Events</CardTitle>
            <CardDescription>Latest detected suspicious activities</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>IP Address</TableHead>
                  <TableHead>Country</TableHead>
                  <TableHead>Timestamp</TableHead>
                  <TableHead>Details</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {mockFraudEvents.map((event) => (
                  <TableRow key={event.id}>
                    <TableCell>{getEventTypeBadge(event.type)}</TableCell>
                    <TableCell className="font-mono">{event.ip}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{event.country}</Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{event.timestamp}</TableCell>
                    <TableCell className="text-sm">{event.details}</TableCell>
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

