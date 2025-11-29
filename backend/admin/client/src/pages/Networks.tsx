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
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { trpc } from "@/lib/trpc";
import { CreateNetworkDialog } from "@/components/CreateNetworkDialog";
import { EditNetworkDialog } from "@/components/EditNetworkDialog";
import { Trash2, Pencil } from "lucide-react";
import { toast } from "sonner";
import { useState } from "react";

export default function Networks() {
  const { data: networks, isLoading } = trpc.networks.list.useQuery();
  const [editingNetwork, setEditingNetwork] = useState<any>(null);
  
  const utils = trpc.useUtils();
  const deleteMutation = trpc.networks.delete.useMutation({
    onSuccess: () => {
      toast.success("Network deleted successfully");
      utils.networks.list.invalidate();
    },
    onError: (error) => {
      toast.error(`Failed to delete network: ${error.message}`);
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
          <h1 className="text-3xl font-bold">Networks</h1>
          <CreateNetworkDialog />
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Affiliate Networks</CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="text-center py-8">Loading...</div>
            ) : networks && networks.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Total Offers</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {networks.map((network) => (
                    <TableRow key={network.id}>
                      <TableCell className="font-medium">{network.name}</TableCell>
                      <TableCell>
                        <Badge variant={network.status === "active" ? "default" : "secondary"}>
                          {network.status}
                        </Badge>
                      </TableCell>
                      <TableCell>0</TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setEditingNetwork(network)}
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDelete(network.id, network.name)}
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
                No networks found
              </div>
            )}
          </CardContent>
        </Card>
      </div>
      
      {editingNetwork && (
        <EditNetworkDialog
          network={editingNetwork}
          open={!!editingNetwork}
          onOpenChange={(open) => !open && setEditingNetwork(null)}
        />
      )}
    </DashboardLayout>
  );
}