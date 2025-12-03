import DashboardLayout from "@/components/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { 
  Activity, 
  Server, 
  Database, 
  HardDrive, 
  Cpu, 
  MemoryStick, 
  Wifi, 
  RefreshCw,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Clock,
  Zap,
  TrendingUp,
  Globe
} from "lucide-react";
import { useState, useEffect } from "react";
import { 
  LineChart, 
  Line, 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer 
} from 'recharts';

// Mock data - replace with actual API calls
const generateMockData = () => ({
  system: {
    status: "healthy",
    uptime: "15d 7h 32m",
    version: "1.0.0",
    environment: "production"
  },
  services: [
    { name: "API Server", status: "healthy", latency: 12, uptime: 99.99 },
    { name: "PostgreSQL", status: "healthy", latency: 5, uptime: 99.95 },
    { name: "Redis", status: "healthy", latency: 2, uptime: 99.99 },
    { name: "Edge Workers", status: "healthy", latency: 8, uptime: 99.97 },
    { name: "WAL Service", status: "healthy", latency: 3, uptime: 100 },
    { name: "Stream Consumers", status: "warning", latency: 45, uptime: 98.5 },
  ],
  metrics: {
    cpu: 45,
    memory: 62,
    disk: 38,
    network: 25,
    goroutines: 1250,
    connections: 85
  },
  performance: {
    clicksPerSecond: 125,
    conversionsPerMinute: 42,
    avgLatencyMs: 18,
    p99LatencyMs: 85,
    errorRate: 0.02
  }
});

// Generate time series data
const generateTimeSeriesData = () => {
  const data = [];
  const now = new Date();
  for (let i = 30; i >= 0; i--) {
    data.push({
      time: new Date(now.getTime() - i * 60000).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
      clicks: Math.floor(Math.random() * 150) + 50,
      latency: Math.floor(Math.random() * 30) + 10,
      errors: Math.floor(Math.random() * 5),
    });
  }
  return data;
};

export default function Monitoring() {
  const [data, setData] = useState(generateMockData());
  const [timeSeriesData, setTimeSeriesData] = useState(generateTimeSeriesData());
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState(new Date());

  // Auto-refresh every 30 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      setData(generateMockData());
      setTimeSeriesData(generateTimeSeriesData());
      setLastUpdated(new Date());
    }, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleRefresh = () => {
    setIsRefreshing(true);
    setTimeout(() => {
      setData(generateMockData());
      setTimeSeriesData(generateTimeSeriesData());
      setLastUpdated(new Date());
      setIsRefreshing(false);
    }, 1000);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "healthy": return "text-green-500";
      case "warning": return "text-yellow-500";
      case "critical": return "text-red-500";
      default: return "text-gray-500";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "healthy": return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      case "warning": return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case "critical": return <XCircle className="h-4 w-4 text-red-500" />;
      default: return <Clock className="h-4 w-4 text-gray-500" />;
    }
  };

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">System Monitoring</h1>
            <p className="text-muted-foreground mt-1">
              Real-time system health and performance metrics
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

        {/* System Status Banner */}
        <Card className={`border-l-4 ${data.system.status === 'healthy' ? 'border-l-green-500' : 'border-l-yellow-500'}`}>
          <CardContent className="py-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                {getStatusIcon(data.system.status)}
                <div>
                  <h3 className="font-semibold">System Status: {data.system.status.toUpperCase()}</h3>
                  <p className="text-sm text-muted-foreground">
                    Uptime: {data.system.uptime} | Version: {data.system.version} | Environment: {data.system.environment}
                  </p>
                </div>
              </div>
              <Badge variant={data.system.status === 'healthy' ? 'default' : 'destructive'}>
                {data.system.status === 'healthy' ? 'All Systems Operational' : 'Issues Detected'}
              </Badge>
            </div>
          </CardContent>
        </Card>

        {/* Performance Metrics */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Clicks/sec</CardTitle>
              <Zap className="h-4 w-4 text-purple-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.performance.clicksPerSecond}</div>
              <p className="text-xs text-muted-foreground">+12% from avg</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Conversions/min</CardTitle>
              <TrendingUp className="h-4 w-4 text-green-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.performance.conversionsPerMinute}</div>
              <p className="text-xs text-muted-foreground">+5% from avg</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
              <Activity className="h-4 w-4 text-blue-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.performance.avgLatencyMs}ms</div>
              <p className="text-xs text-muted-foreground">p99: {data.performance.p99LatencyMs}ms</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
              <AlertTriangle className="h-4 w-4 text-yellow-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.performance.errorRate}%</div>
              <p className="text-xs text-green-500">Below threshold</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Goroutines</CardTitle>
              <Server className="h-4 w-4 text-cyan-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.metrics.goroutines}</div>
              <p className="text-xs text-muted-foreground">Active workers</p>
            </CardContent>
          </Card>
        </div>

        {/* Charts */}
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Click Traffic (Last 30 min)</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={250}>
                <AreaChart data={timeSeriesData}>
                  <defs>
                    <linearGradient id="colorClicks" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                  <XAxis dataKey="time" stroke="#888" fontSize={10} />
                  <YAxis stroke="#888" fontSize={10} />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#1a1a1a', border: '1px solid #333' }}
                  />
                  <Area type="monotone" dataKey="clicks" stroke="#8b5cf6" fillOpacity={1} fill="url(#colorClicks)" />
                </AreaChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Response Latency (Last 30 min)</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={250}>
                <LineChart data={timeSeriesData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                  <XAxis dataKey="time" stroke="#888" fontSize={10} />
                  <YAxis stroke="#888" fontSize={10} />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#1a1a1a', border: '1px solid #333' }}
                  />
                  <Line type="monotone" dataKey="latency" stroke="#10b981" strokeWidth={2} dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </div>

        {/* Services Status */}
        <Card>
          <CardHeader>
            <CardTitle>Services Health</CardTitle>
            <CardDescription>Status of all system components</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {data.services.map((service) => (
                <div 
                  key={service.name}
                  className="flex items-center justify-between p-4 rounded-lg border bg-card"
                >
                  <div className="flex items-center gap-3">
                    {getStatusIcon(service.status)}
                    <div>
                      <p className="font-medium">{service.name}</p>
                      <p className="text-xs text-muted-foreground">
                        Latency: {service.latency}ms | Uptime: {service.uptime}%
                      </p>
                    </div>
                  </div>
                  <Badge variant={service.status === 'healthy' ? 'outline' : 'destructive'}>
                    {service.status}
                  </Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Resource Usage */}
        <Card>
          <CardHeader>
            <CardTitle>Resource Usage</CardTitle>
            <CardDescription>Server resource utilization</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Cpu className="h-4 w-4 text-blue-500" />
                    <span className="text-sm font-medium">CPU</span>
                  </div>
                  <span className="text-sm font-bold">{data.metrics.cpu}%</span>
                </div>
                <Progress value={data.metrics.cpu} className="h-2" />
              </div>
              
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <MemoryStick className="h-4 w-4 text-purple-500" />
                    <span className="text-sm font-medium">Memory</span>
                  </div>
                  <span className="text-sm font-bold">{data.metrics.memory}%</span>
                </div>
                <Progress value={data.metrics.memory} className="h-2" />
              </div>
              
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <HardDrive className="h-4 w-4 text-green-500" />
                    <span className="text-sm font-medium">Disk</span>
                  </div>
                  <span className="text-sm font-bold">{data.metrics.disk}%</span>
                </div>
                <Progress value={data.metrics.disk} className="h-2" />
              </div>
              
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Wifi className="h-4 w-4 text-cyan-500" />
                    <span className="text-sm font-medium">Network</span>
                  </div>
                  <span className="text-sm font-bold">{data.metrics.network}%</span>
                </div>
                <Progress value={data.metrics.network} className="h-2" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
}

