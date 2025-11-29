import DashboardLayout from "@/components/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { trpc } from "@/lib/trpc";
import { CreateBadgeDialog } from "@/components/CreateBadgeDialog";
import { EditBadgeDialog } from "@/components/EditBadgeDialog";
import { Trash2, Award, Pencil } from "lucide-react";
import { toast } from "sonner";
import { useState } from "react";

export default function Badges() {
  const { data: badges, isLoading } = trpc.badges.list.useQuery();
  const [editingBadge, setEditingBadge] = useState<any>(null);
  
  const utils = trpc.useUtils();
  const deleteMutation = trpc.badges.delete.useMutation({
    onSuccess: () => {
      toast.success("Badge deleted successfully");
      utils.badges.list.invalidate();
    },
    onError: (error) => {
      toast.error(`Failed to delete badge: ${error.message}`);
    },
  });

  const handleDelete = (id: string, name: string) => {
    if (confirm(`Are you sure you want to delete "${name}"?`)) {
      deleteMutation.mutate({ id });
    }
  };

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold">Badges</h1>
          <CreateBadgeDialog />
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Achievement Badges</CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="text-center py-8">Loading...</div>
            ) : badges && badges.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Badge</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Points Reward</TableHead>
                    <TableHead>Earned By</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {badges.map((badge) => (
                    <TableRow key={badge.id}>
                      <TableCell>
                        <div className="flex items-center gap-3">
                          {badge.iconUrl ? (
                            <img src={badge.iconUrl} alt={badge.name} className="w-8 h-8" />
                          ) : (
                            <Award className="w-8 h-8 text-primary" />
                          )}
                          <span className="font-medium">{badge.name}</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground max-w-md">
                        {badge.description}
                      </TableCell>
                      <TableCell className="font-semibold">{badge.pointsReward ?? 0}</TableCell>
                      <TableCell>0 users</TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setEditingBadge(badge)}
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDelete(badge.id, badge.name)}
                            disabled={deleteMutation.isPending}
                          >
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
                No badges found
              </div>
            )}
          </CardContent>
        </Card>
      </div>
      
      {editingBadge && (
        <EditBadgeDialog
          badge={editingBadge}
          open={!!editingBadge}
          onOpenChange={(open) => !open && setEditingBadge(null)}
        />
      )}
    </DashboardLayout>
  );
}
