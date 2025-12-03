import DashboardLayout from "@/components/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Textarea } from "@/components/ui/textarea";
import { 
  Webhook, 
  Plus, 
  Search, 
  MoreVertical,
  Edit, 
  Trash2, 
  Play,
  Pause,
  RefreshCw,
  CheckCircle,
  XCircle,
  Clock,
  Send,
  AlertTriangle
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

// Mock data
const mockWebhooks = [
  {
    id: "1",
    name: "Conversion Postback",
    url: "https://partner.com/postback",
    triggerType: "conversion",
    status: "active",
    signatureMode: "hmac",
    lastTriggered: "2 min ago",
    successRate: 98.5,
    totalDeliveries: 1250,
    failedDeliveries: 19,
  },
  {
    id: "2",
    name: "Click Notification",
    url: "https://analytics.example.com/clicks",
    triggerType: "click",
    status: "active",
    signatureMode: "jwt",
    lastTriggered: "30 sec ago",
    successRate: 99.2,
    totalDeliveries: 15420,
    failedDeliveries: 124,
  },
  {
    id: "3",
    name: "Fraud Alert",
    url: "https://security.internal/alerts",
    triggerType: "fraud",
    status: "paused",
    signatureMode: "none",
    lastTriggered: "1 hour ago",
    successRate: 95.0,
    totalDeliveries: 340,
    failedDeliveries: 17,
  },
];

const mockDLQ = [
  { id: "dlq_1", webhookName: "Conversion Postback", error: "Connection timeout", attempts: 5, lastAttempt: "5 min ago" },
  { id: "dlq_2", webhookName: "Click Notification", error: "HTTP 500", attempts: 3, lastAttempt: "15 min ago" },
];

const triggerTypes = [
  { value: "click", label: "Click Event" },
  { value: "conversion", label: "Conversion Event" },
  { value: "postback", label: "Postback Received" },
  { value: "fraud", label: "Fraud Detected" },
  { value: "user_signup", label: "User Signup" },
];

const signatureModes = [
  { value: "none", label: "None" },
  { value: "hmac", label: "HMAC-SHA256" },
  { value: "jwt", label: "JWT" },
];

export default function Webhooks() {
  const [searchQuery, setSearchQuery] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newWebhook, setNewWebhook] = useState({
    name: "",
    url: "",
    triggerType: "conversion",
    signatureMode: "hmac",
    secret: "",
  });

  const filteredWebhooks = mockWebhooks.filter(webhook => 
    webhook.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    webhook.url.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreateWebhook = () => {
    if (!newWebhook.name || !newWebhook.url) {
      toast.error("Please fill in all required fields");
      return;
    }
    toast.success(`Webhook "${newWebhook.name}" created successfully`);
    setIsCreateOpen(false);
    setNewWebhook({ name: "", url: "", triggerType: "conversion", signatureMode: "hmac", secret: "" });
  };

  const handleTestWebhook = (name: string) => {
    toast.info(`Testing webhook "${name}"...`);
    setTimeout(() => toast.success(`Webhook "${name}" test successful!`), 1500);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return <Badge className="bg-green-500/10 text-green-500"><CheckCircle className="h-3 w-3 mr-1" /> Active</Badge>;
      case "paused":
        return <Badge variant="secondary"><Pause className="h-3 w-3 mr-1" /> Paused</Badge>;
      case "error":
        return <Badge variant="destructive"><XCircle className="h-3 w-3 mr-1" /> Error</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getSignatureBadge = (mode: string) => {
    switch (mode) {
      case "hmac":
        return <Badge variant="outline" className="text-blue-500">HMAC</Badge>;
      case "jwt":
        return <Badge variant="outline" className="text-purple-500">JWT</Badge>;
      default:
        return <Badge variant="outline">None</Badge>;
    }
  };

  const stats = [
    { title: "Total Webhooks", value: mockWebhooks.length, icon: Webhook, color: "text-blue-500" },
    { title: "Active", value: mockWebhooks.filter(w => w.status === 'active').length, icon: CheckCircle, color: "text-green-500" },
    { title: "Total Deliveries", value: mockWebhooks.reduce((acc, w) => acc + w.totalDeliveries, 0).toLocaleString(), icon: Send, color: "text-purple-500" },
    { title: "DLQ Items", value: mockDLQ.length, icon: AlertTriangle, color: "text-yellow-500" },
  ];

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Webhooks</h1>
            <p className="text-muted-foreground mt-1">
              Manage webhook pipelines and deliveries
            </p>
          </div>
          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Create Webhook
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[500px]">
              <DialogHeader>
                <DialogTitle>Create Webhook</DialogTitle>
                <DialogDescription>
                  Configure a new webhook endpoint
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="name">Name *</Label>
                  <Input
                    id="name"
                    value={newWebhook.name}
                    onChange={(e) => setNewWebhook({ ...newWebhook, name: e.target.value })}
                    placeholder="e.g., Conversion Postback"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="url">URL *</Label>
                  <Input
                    id="url"
                    type="url"
                    value={newWebhook.url}
                    onChange={(e) => setNewWebhook({ ...newWebhook, url: e.target.value })}
                    placeholder="https://example.com/webhook"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label>Trigger</Label>
                    <Select 
                      value={newWebhook.triggerType} 
                      onValueChange={(v) => setNewWebhook({ ...newWebhook, triggerType: v })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {triggerTypes.map(t => (
                          <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-2">
                    <Label>Signature</Label>
                    <Select 
                      value={newWebhook.signatureMode} 
                      onValueChange={(v) => setNewWebhook({ ...newWebhook, signatureMode: v })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {signatureModes.map(s => (
                          <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                {newWebhook.signatureMode !== 'none' && (
                  <div className="grid gap-2">
                    <Label htmlFor="secret">Secret Key</Label>
                    <Input
                      id="secret"
                      type="password"
                      value={newWebhook.secret}
                      onChange={(e) => setNewWebhook({ ...newWebhook, secret: e.target.value })}
                      placeholder="Enter signing secret"
                    />
                  </div>
                )}
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                <Button onClick={handleCreateWebhook}>Create Webhook</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
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
                </CardContent>
              </Card>
            );
          })}
        </div>

        {/* Webhooks Table */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>All Webhooks</CardTitle>
                <CardDescription>Manage webhook pipelines</CardDescription>
              </div>
              <div className="relative w-64">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search webhooks..."
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
                  <TableHead>Name</TableHead>
                  <TableHead>URL</TableHead>
                  <TableHead>Trigger</TableHead>
                  <TableHead>Signature</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Success Rate</TableHead>
                  <TableHead>Last Triggered</TableHead>
                  <TableHead className="w-[80px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredWebhooks.map((webhook) => (
                  <TableRow key={webhook.id}>
                    <TableCell className="font-medium">{webhook.name}</TableCell>
                    <TableCell className="font-mono text-xs max-w-[200px] truncate">{webhook.url}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{webhook.triggerType}</Badge>
                    </TableCell>
                    <TableCell>{getSignatureBadge(webhook.signatureMode)}</TableCell>
                    <TableCell>{getStatusBadge(webhook.status)}</TableCell>
                    <TableCell>
                      <span className={webhook.successRate >= 98 ? 'text-green-500' : webhook.successRate >= 95 ? 'text-yellow-500' : 'text-red-500'}>
                        {webhook.successRate}%
                      </span>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{webhook.lastTriggered}</TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleTestWebhook(webhook.name)}>
                            <Play className="h-4 w-4 mr-2" />
                            Test
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Edit className="h-4 w-4 mr-2" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            {webhook.status === 'active' ? (
                              <><Pause className="h-4 w-4 mr-2" /> Pause</>
                            ) : (
                              <><Play className="h-4 w-4 mr-2" /> Activate</>
                            )}
                          </DropdownMenuItem>
                          <DropdownMenuItem className="text-destructive">
                            <Trash2 className="h-4 w-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Dead Letter Queue */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-yellow-500" />
              Dead Letter Queue
            </CardTitle>
            <CardDescription>Failed webhook deliveries awaiting retry</CardDescription>
          </CardHeader>
          <CardContent>
            {mockDLQ.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Webhook</TableHead>
                    <TableHead>Error</TableHead>
                    <TableHead>Attempts</TableHead>
                    <TableHead>Last Attempt</TableHead>
                    <TableHead className="w-[100px]">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {mockDLQ.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className="font-medium">{item.webhookName}</TableCell>
                      <TableCell className="text-red-500">{item.error}</TableCell>
                      <TableCell>{item.attempts}</TableCell>
                      <TableCell className="text-muted-foreground">{item.lastAttempt}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Button variant="ghost" size="sm">
                            <RefreshCw className="h-4 w-4 mr-1" />
                            Retry
                          </Button>
                          <Button variant="ghost" size="sm" className="text-destructive">
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <CheckCircle className="h-12 w-12 mx-auto mb-4 text-green-500" />
                <p>No failed deliveries</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
}

