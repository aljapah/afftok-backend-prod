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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Label } from "@/components/ui/label";
import { 
  Building2, 
  Plus, 
  Search, 
  MoreVertical, 
  Edit, 
  Trash2, 
  Ban, 
  CheckCircle,
  Users,
  MousePointerClick,
  DollarSign,
  Globe,
  Settings
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

// Mock data - replace with actual API calls
const mockTenants = [
  {
    id: "1",
    name: "AffTok Main",
    slug: "afftok-main",
    status: "active",
    plan: "enterprise",
    adminEmail: "admin@afftok.com",
    usersCount: 1250,
    offersCount: 45,
    clicksToday: 15420,
    revenue: 45000,
    createdAt: "2024-01-15",
  },
  {
    id: "2",
    name: "Partner Corp",
    slug: "partner-corp",
    status: "active",
    plan: "pro",
    adminEmail: "admin@partner.com",
    usersCount: 340,
    offersCount: 12,
    clicksToday: 3200,
    revenue: 12000,
    createdAt: "2024-03-20",
  },
  {
    id: "3",
    name: "Test Tenant",
    slug: "test-tenant",
    status: "suspended",
    plan: "free",
    adminEmail: "test@test.com",
    usersCount: 5,
    offersCount: 2,
    clicksToday: 0,
    revenue: 0,
    createdAt: "2024-06-01",
  },
];

export default function Tenants() {
  const [searchQuery, setSearchQuery] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newTenant, setNewTenant] = useState({
    name: "",
    slug: "",
    adminEmail: "",
    plan: "free"
  });

  const filteredTenants = mockTenants.filter(tenant => 
    tenant.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    tenant.slug.toLowerCase().includes(searchQuery.toLowerCase()) ||
    tenant.adminEmail.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreateTenant = () => {
    if (!newTenant.name || !newTenant.slug || !newTenant.adminEmail) {
      toast.error("Please fill in all required fields");
      return;
    }
    toast.success(`Tenant "${newTenant.name}" created successfully`);
    setIsCreateOpen(false);
    setNewTenant({ name: "", slug: "", adminEmail: "", plan: "free" });
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return <Badge className="bg-green-500/10 text-green-500 border-green-500/20">Active</Badge>;
      case "suspended":
        return <Badge variant="destructive">Suspended</Badge>;
      case "pending":
        return <Badge variant="secondary">Pending</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getPlanBadge = (plan: string) => {
    switch (plan) {
      case "enterprise":
        return <Badge className="bg-purple-500/10 text-purple-500 border-purple-500/20">Enterprise</Badge>;
      case "pro":
        return <Badge className="bg-blue-500/10 text-blue-500 border-blue-500/20">Pro</Badge>;
      case "free":
        return <Badge variant="outline">Free</Badge>;
      default:
        return <Badge variant="outline">{plan}</Badge>;
    }
  };

  const stats = [
    { title: "Total Tenants", value: mockTenants.length, icon: Building2, color: "text-blue-500" },
    { title: "Active Tenants", value: mockTenants.filter(t => t.status === 'active').length, icon: CheckCircle, color: "text-green-500" },
    { title: "Total Users", value: mockTenants.reduce((acc, t) => acc + t.usersCount, 0), icon: Users, color: "text-purple-500" },
    { title: "Total Revenue", value: `$${(mockTenants.reduce((acc, t) => acc + t.revenue, 0) / 100).toLocaleString()}`, icon: DollarSign, color: "text-green-500" },
  ];

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Tenants</h1>
            <p className="text-muted-foreground mt-1">
              Manage multi-tenant organizations
            </p>
          </div>
          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Create Tenant
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create New Tenant</DialogTitle>
                <DialogDescription>
                  Add a new organization to the platform
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="name">Organization Name *</Label>
                  <Input
                    id="name"
                    value={newTenant.name}
                    onChange={(e) => setNewTenant({ ...newTenant, name: e.target.value })}
                    placeholder="e.g., Acme Corporation"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="slug">Slug *</Label>
                  <Input
                    id="slug"
                    value={newTenant.slug}
                    onChange={(e) => setNewTenant({ ...newTenant, slug: e.target.value.toLowerCase().replace(/\s+/g, '-') })}
                    placeholder="e.g., acme-corp"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="adminEmail">Admin Email *</Label>
                  <Input
                    id="adminEmail"
                    type="email"
                    value={newTenant.adminEmail}
                    onChange={(e) => setNewTenant({ ...newTenant, adminEmail: e.target.value })}
                    placeholder="admin@example.com"
                  />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                <Button onClick={handleCreateTenant}>Create Tenant</Button>
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

        {/* Tenants Table */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>All Tenants</CardTitle>
                <CardDescription>Manage tenant organizations</CardDescription>
              </div>
              <div className="relative w-64">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search tenants..."
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
                  <TableHead>Tenant</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Plan</TableHead>
                  <TableHead>Users</TableHead>
                  <TableHead>Offers</TableHead>
                  <TableHead>Clicks Today</TableHead>
                  <TableHead>Revenue</TableHead>
                  <TableHead className="w-[50px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredTenants.map((tenant) => (
                  <TableRow key={tenant.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
                          <Building2 className="h-5 w-5 text-primary" />
                        </div>
                        <div>
                          <p className="font-medium">{tenant.name}</p>
                          <p className="text-xs text-muted-foreground">{tenant.slug}</p>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{getStatusBadge(tenant.status)}</TableCell>
                    <TableCell>{getPlanBadge(tenant.plan)}</TableCell>
                    <TableCell>{tenant.usersCount.toLocaleString()}</TableCell>
                    <TableCell>{tenant.offersCount}</TableCell>
                    <TableCell>{tenant.clicksToday.toLocaleString()}</TableCell>
                    <TableCell className="font-medium text-green-500">
                      ${(tenant.revenue / 100).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem>
                            <Edit className="h-4 w-4 mr-2" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Settings className="h-4 w-4 mr-2" />
                            Settings
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Globe className="h-4 w-4 mr-2" />
                            Domains
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          {tenant.status === 'active' ? (
                            <DropdownMenuItem className="text-yellow-500">
                              <Ban className="h-4 w-4 mr-2" />
                              Suspend
                            </DropdownMenuItem>
                          ) : (
                            <DropdownMenuItem className="text-green-500">
                              <CheckCircle className="h-4 w-4 mr-2" />
                              Activate
                            </DropdownMenuItem>
                          )}
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
      </div>
    </DashboardLayout>
  );
}

