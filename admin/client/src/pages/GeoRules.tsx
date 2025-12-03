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
  Globe, 
  Plus, 
  Search, 
  Edit, 
  Trash2, 
  Shield,
  Ban,
  CheckCircle,
  MapPin,
  AlertTriangle
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

// Country list (simplified)
const countries = [
  { code: "US", name: "United States" },
  { code: "GB", name: "United Kingdom" },
  { code: "DE", name: "Germany" },
  { code: "FR", name: "France" },
  { code: "SA", name: "Saudi Arabia" },
  { code: "AE", name: "United Arab Emirates" },
  { code: "KW", name: "Kuwait" },
  { code: "QA", name: "Qatar" },
  { code: "EG", name: "Egypt" },
  { code: "IN", name: "India" },
  { code: "CN", name: "China" },
  { code: "JP", name: "Japan" },
  { code: "BR", name: "Brazil" },
  { code: "RU", name: "Russia" },
];

// Mock data
const mockGeoRules = [
  {
    id: "1",
    name: "Block High-Risk Countries",
    scopeType: "global",
    scopeId: null,
    mode: "block",
    countries: ["CN", "RU", "IN"],
    priority: 100,
    status: "active",
    blockedClicks: 1250,
    createdAt: "2024-01-15",
  },
  {
    id: "2",
    name: "GCC Only - Premium Offer",
    scopeType: "offer",
    scopeId: "offer-123",
    mode: "allow",
    countries: ["SA", "AE", "KW", "QA"],
    priority: 50,
    status: "active",
    blockedClicks: 3420,
    createdAt: "2024-02-20",
  },
  {
    id: "3",
    name: "EU Compliance",
    scopeType: "advertiser",
    scopeId: "adv-456",
    mode: "allow",
    countries: ["DE", "FR", "GB"],
    priority: 75,
    status: "inactive",
    blockedClicks: 0,
    createdAt: "2024-03-10",
  },
];

export default function GeoRules() {
  const [searchQuery, setSearchQuery] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newRule, setNewRule] = useState({
    name: "",
    scopeType: "global",
    mode: "block",
    countries: [] as string[],
    priority: 50
  });

  const filteredRules = mockGeoRules.filter(rule => 
    rule.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreateRule = () => {
    if (!newRule.name || newRule.countries.length === 0) {
      toast.error("Please fill in all required fields");
      return;
    }
    toast.success(`Geo rule "${newRule.name}" created successfully`);
    setIsCreateOpen(false);
    setNewRule({ name: "", scopeType: "global", mode: "block", countries: [], priority: 50 });
  };

  const getModeBadge = (mode: string) => {
    if (mode === "allow") {
      return <Badge className="bg-green-500/10 text-green-500 border-green-500/20">Allow</Badge>;
    }
    return <Badge className="bg-red-500/10 text-red-500 border-red-500/20">Block</Badge>;
  };

  const getScopeBadge = (scopeType: string) => {
    switch (scopeType) {
      case "global":
        return <Badge variant="outline"><Globe className="h-3 w-3 mr-1" /> Global</Badge>;
      case "offer":
        return <Badge variant="secondary">Offer</Badge>;
      case "advertiser":
        return <Badge variant="secondary">Advertiser</Badge>;
      default:
        return <Badge variant="outline">{scopeType}</Badge>;
    }
  };

  const stats = [
    { title: "Total Rules", value: mockGeoRules.length, icon: Shield, color: "text-blue-500" },
    { title: "Active Rules", value: mockGeoRules.filter(r => r.status === 'active').length, icon: CheckCircle, color: "text-green-500" },
    { title: "Blocked Clicks", value: mockGeoRules.reduce((acc, r) => acc + r.blockedClicks, 0).toLocaleString(), icon: Ban, color: "text-red-500" },
    { title: "Countries Covered", value: new Set(mockGeoRules.flatMap(r => r.countries)).size, icon: MapPin, color: "text-purple-500" },
  ];

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Geo Rules</h1>
            <p className="text-muted-foreground mt-1">
              Manage country-based traffic rules
            </p>
          </div>
          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Create Rule
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[500px]">
              <DialogHeader>
                <DialogTitle>Create Geo Rule</DialogTitle>
                <DialogDescription>
                  Define country-based traffic rules
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="name">Rule Name *</Label>
                  <Input
                    id="name"
                    value={newRule.name}
                    onChange={(e) => setNewRule({ ...newRule, name: e.target.value })}
                    placeholder="e.g., Block High-Risk Countries"
                  />
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label>Scope</Label>
                    <Select 
                      value={newRule.scopeType} 
                      onValueChange={(v) => setNewRule({ ...newRule, scopeType: v })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="global">Global</SelectItem>
                        <SelectItem value="offer">Specific Offer</SelectItem>
                        <SelectItem value="advertiser">Advertiser</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  
                  <div className="grid gap-2">
                    <Label>Mode</Label>
                    <Select 
                      value={newRule.mode} 
                      onValueChange={(v) => setNewRule({ ...newRule, mode: v })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="block">Block Countries</SelectItem>
                        <SelectItem value="allow">Allow Only</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="grid gap-2">
                  <Label>Countries *</Label>
                  <div className="flex flex-wrap gap-2 p-3 border rounded-md min-h-[80px]">
                    {newRule.countries.map(code => {
                      const country = countries.find(c => c.code === code);
                      return (
                        <Badge 
                          key={code} 
                          variant="secondary"
                          className="cursor-pointer"
                          onClick={() => setNewRule({
                            ...newRule,
                            countries: newRule.countries.filter(c => c !== code)
                          })}
                        >
                          {country?.name || code} ×
                        </Badge>
                      );
                    })}
                  </div>
                  <Select 
                    onValueChange={(v) => {
                      if (!newRule.countries.includes(v)) {
                        setNewRule({ ...newRule, countries: [...newRule.countries, v] });
                      }
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Add country..." />
                    </SelectTrigger>
                    <SelectContent>
                      {countries.filter(c => !newRule.countries.includes(c.code)).map(country => (
                        <SelectItem key={country.code} value={country.code}>
                          {country.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="grid gap-2">
                  <Label>Priority (1-100)</Label>
                  <Input
                    type="number"
                    min="1"
                    max="100"
                    value={newRule.priority}
                    onChange={(e) => setNewRule({ ...newRule, priority: parseInt(e.target.value) || 50 })}
                  />
                  <p className="text-xs text-muted-foreground">Higher priority rules are evaluated first</p>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                <Button onClick={handleCreateRule}>Create Rule</Button>
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

        {/* Rules Table */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>All Geo Rules</CardTitle>
                <CardDescription>Country-based traffic filtering rules</CardDescription>
              </div>
              <div className="relative w-64">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search rules..."
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
                  <TableHead>Rule Name</TableHead>
                  <TableHead>Scope</TableHead>
                  <TableHead>Mode</TableHead>
                  <TableHead>Countries</TableHead>
                  <TableHead>Priority</TableHead>
                  <TableHead>Blocked</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-[100px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredRules.map((rule) => (
                  <TableRow key={rule.id}>
                    <TableCell className="font-medium">{rule.name}</TableCell>
                    <TableCell>{getScopeBadge(rule.scopeType)}</TableCell>
                    <TableCell>{getModeBadge(rule.mode)}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {rule.countries.slice(0, 3).map(code => (
                          <Badge key={code} variant="outline" className="text-xs">
                            {code}
                          </Badge>
                        ))}
                        {rule.countries.length > 3 && (
                          <Badge variant="outline" className="text-xs">
                            +{rule.countries.length - 3}
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>{rule.priority}</TableCell>
                    <TableCell className="text-red-500">{rule.blockedClicks.toLocaleString()}</TableCell>
                    <TableCell>
                      <Badge variant={rule.status === 'active' ? 'default' : 'secondary'}>
                        {rule.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button variant="ghost" size="icon">
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" className="text-destructive">
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
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

