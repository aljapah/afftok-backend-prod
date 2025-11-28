import DashboardLayout from "@/components/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { trpc } from "@/lib/trpc";
import { useRoute } from "wouter";
import { ArrowLeft, Mail, Phone, Calendar, Award, Activity } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useLocation } from "wouter";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export default function UserDetails() {
  const [, params] = useRoute("/users/:id");
  const [, navigate] = useLocation();
  const userId = params?.id;

  // Mock data - replace with real API later
  const user = {
    id: userId,
    name: "Ahmed Hassan",
    email: "ahmed@example.com",
    phone: "+965 9999 9999",
    joinedAt: "2024-01-15",
    totalClicks: 1250,
    totalConversions: 85,
    totalEarnings: 4250,
    level: 5,
    points: 3420,
  };

  const activityHistory = [
    { id: 1, type: "Click", offer: "Amazon Prime", date: "2024-03-15 14:30", status: "success" },
    { id: 2, type: "Conversion", offer: "Netflix", date: "2024-03-14 10:20", status: "success" },
    { id: 3, type: "Click", offer: "Spotify", date: "2024-03-13 16:45", status: "success" },
    { id: 4, type: "Click", offer: "Disney+", date: "2024-03-12 09:15", status: "failed" },
  ];

  const earnedBadges = [
    { id: 1, name: "First Click", icon: "🎯", earnedAt: "2024-01-15" },
    { id: 2, name: "10 Conversions", icon: "⭐", earnedAt: "2024-02-01" },
    { id: 3, name: "Level 5", icon: "🏆", earnedAt: "2024-03-01" },
  ];

  const userOffers = [
    { id: 1, title: "Amazon Prime", clicks: 450, conversions: 28, earnings: 1400 },
    { id: 2, title: "Netflix", clicks: 380, conversions: 25, earnings: 1250 },
    { id: 3, title: "Spotify", clicks: 420, conversions: 32, earnings: 1600 },
  ];

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header with Back Button */}
        <div className="flex items-center gap-4">
          <Button 
            variant="outline" 
            size="icon"
            onClick={() => navigate("/users")}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h1 className="text-3xl font-bold">User Details</h1>
        </div>

        {/* User Info Card */}
        <Card>
          <CardHeader>
            <CardTitle>User Information</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-2">
              <div className="space-y-4">
                <div className="flex items-center gap-3">
                  <div className="w-16 h-16 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white text-2xl font-bold">
                    {user.name.charAt(0)}
                  </div>
                  <div>
                    <h3 className="text-xl font-bold">{user.name}</h3>
                    <p className="text-sm text-muted-foreground">ID: {user.id}</p>
                  </div>
                </div>
                
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm">
                    <Mail className="h-4 w-4 text-muted-foreground" />
                    <span>{user.email}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <Phone className="h-4 w-4 text-muted-foreground" />
                    <span>{user.phone}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    <span>Joined: {new Date(user.joinedAt).toLocaleDateString()}</span>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 rounded-lg bg-purple-500/10 border border-purple-500/20">
                  <p className="text-sm text-muted-foreground">Total Clicks</p>
                  <p className="text-2xl font-bold text-purple-500">{user.totalClicks ?? 0}</p>
                </div>
                <div className="p-4 rounded-lg bg-pink-500/10 border border-pink-500/20">
                  <p className="text-sm text-muted-foreground">Conversions</p>
                  <p className="text-2xl font-bold text-pink-500">{user.totalConversions ?? 0}</p>
                </div>
                <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
                  <p className="text-sm text-muted-foreground">Total Earnings</p>
                  <p className="text-2xl font-bold text-green-500">${user.totalEarnings ?? 0}</p>
                </div>
                <div className="p-4 rounded-lg bg-blue-500/10 border border-blue-500/20">
                  <p className="text-sm text-muted-foreground">Level / Points</p>
                  <p className="text-2xl font-bold text-blue-500">{user.level ?? 1} / {user.points ?? 0}</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Earned Badges */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Award className="h-5 w-5" />
              Earned Badges
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-3">
              {earnedBadges.map((badge) => (
                <div 
                  key={badge.id}
                  className="p-4 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <div className="text-4xl">{badge.icon}</div>
                    <div>
                      <p className="font-semibold">{badge.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {new Date(badge.earnedAt).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* User Offers Performance */}
        <Card>
          <CardHeader>
            <CardTitle>Offers Performance</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Offer</TableHead>
                  <TableHead>Clicks</TableHead>
                  <TableHead>Conversions</TableHead>
                  <TableHead>Earnings</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {userOffers.map((offer) => (
                  <TableRow key={offer.id}>
                    <TableCell className="font-medium">{offer.title}</TableCell>
                    <TableCell>{offer.clicks ?? 0}</TableCell>
                    <TableCell>{offer.conversions ?? 0}</TableCell>
                    <TableCell className="text-green-500 font-semibold">
                      ${offer.earnings ?? 0}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Activity History */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Activity History
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Offer</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {activityHistory.map((activity) => (
                  <TableRow key={activity.id}>
                    <TableCell>
                      <Badge variant={activity.type === "Conversion" ? "default" : "outline"}>
                        {activity.type}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-medium">{activity.offer}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {activity.date}
                    </TableCell>
                    <TableCell>
                      <Badge 
                        variant={activity.status === "success" ? "default" : "destructive"}
                      >
                        {activity.status}
                      </Badge>
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
