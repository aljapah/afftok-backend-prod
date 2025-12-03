import DashboardLayout from "@/components/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
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
import { CreateOfferDialog } from "@/components/CreateOfferDialog";
import { EditOfferDialog } from "@/components/EditOfferDialog";
import { Trash2, Search, Pencil, Download, Check, X, Clock, AlertCircle } from "lucide-react";
import { toast } from "sonner";
import TableSkeleton from "@/components/TableSkeleton";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useState } from "react";
import { exportToCSV } from "@/lib/export";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Textarea } from "@/components/ui/textarea";

export default function Offers() {
  const { data: offers, isLoading } = trpc.offers.list.useQuery();
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;
  const [editingOffer, setEditingOffer] = useState<any>(null);
  const [rejectDialogOpen, setRejectDialogOpen] = useState(false);
  const [rejectingOffer, setRejectingOffer] = useState<any>(null);
  const [rejectReason, setRejectReason] = useState("");
  
  // Get pending offers count
  const pendingOffers = offers?.filter(o => o.status === "pending") || [];
  
  const filteredOffers = offers?.filter((offer) => {
    const matchesSearch = offer.title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      offer.description?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = statusFilter === "all" || offer.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const totalPages = Math.ceil((filteredOffers?.length || 0) / itemsPerPage);
  const paginatedOffers = filteredOffers?.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  );
  const utils = trpc.useUtils();
  
  const deleteMutation = trpc.offers.delete.useMutation({
    onSuccess: () => {
      toast.success("Offer deleted successfully");
      utils.offers.list.invalidate();
    },
    onError: (error) => {
      toast.error(`Failed to delete offer: ${error.message}`);
    },
  });

  const approveMutation = trpc.offers.approve.useMutation({
    onSuccess: () => {
      toast.success("Offer approved successfully! It's now visible to promoters.");
      utils.offers.list.invalidate();
    },
    onError: (error) => {
      toast.error(`Failed to approve offer: ${error.message}`);
    },
  });

  const rejectMutation = trpc.offers.reject.useMutation({
    onSuccess: () => {
      toast.success("Offer rejected.");
      utils.offers.list.invalidate();
      setRejectDialogOpen(false);
      setRejectingOffer(null);
      setRejectReason("");
    },
    onError: (error) => {
      toast.error(`Failed to reject offer: ${error.message}`);
    },
  });

  const handleDelete = (id: string, title: string) => {
    if (confirm(`Are you sure you want to delete "${title}"?`)) {
      deleteMutation.mutate({ id });
    }
  };

  const handleApprove = (id: string) => {
    approveMutation.mutate({ id });
  };

  const handleReject = () => {
    if (rejectingOffer) {
      rejectMutation.mutate({ id: rejectingOffer.id, reason: rejectReason });
    }
  };

  const openRejectDialog = (offer: any) => {
    setRejectingOffer(offer);
    setRejectReason("");
    setRejectDialogOpen(true);
  };

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Pending Offers Alert */}
        {pendingOffers.length > 0 && (
          <Card className="border-yellow-500/50 bg-yellow-500/10">
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <AlertCircle className="h-5 w-5 text-yellow-500" />
                <CardTitle className="text-yellow-500">
                  {pendingOffers.length} Pending Offer{pendingOffers.length > 1 ? 's' : ''} Awaiting Review
                </CardTitle>
              </div>
              <CardDescription>
                These offers were submitted by advertisers and need your approval before being visible to promoters.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {pendingOffers.slice(0, 3).map((offer) => (
                  <div key={offer.id} className="flex items-center justify-between p-3 bg-background/50 rounded-lg">
                    <div>
                      <p className="font-medium">{offer.title}</p>
                      <p className="text-sm text-muted-foreground">{offer.category} • ${(offer.payout / 100).toFixed(2)} payout</p>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        className="text-green-500 border-green-500 hover:bg-green-500/10"
                        onClick={() => handleApprove(offer.id)}
                        disabled={approveMutation.isPending}
                      >
                        <Check className="h-4 w-4 mr-1" />
                        Approve
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        className="text-red-500 border-red-500 hover:bg-red-500/10"
                        onClick={() => openRejectDialog(offer)}
                        disabled={rejectMutation.isPending}
                      >
                        <X className="h-4 w-4 mr-1" />
                        Reject
                      </Button>
                    </div>
                  </div>
                ))}
                {pendingOffers.length > 3 && (
                  <Button variant="link" className="text-yellow-500" onClick={() => setStatusFilter("pending")}>
                    View all {pendingOffers.length} pending offers →
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        )}

        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold">Offers</h1>
          <div className="flex items-center gap-4">
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="inactive">Inactive</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                if (offers && offers.length > 0) {
                  exportToCSV(offers, `afftok-offers-${new Date().toISOString().split('T')[0]}`);
                  toast.success('Offers exported successfully');
                } else {
                  toast.error('No offers to export');
                }
              }}
            >
              <Download className="h-4 w-4 mr-2" />
              Export CSV
            </Button>
            <div className="relative w-64">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search offers..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
            <CreateOfferDialog />
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>All Offers</CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead>Network</TableHead>
                    <TableHead>Category</TableHead>
                    <TableHead>Payout</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Clicks</TableHead>
                    <TableHead>Conversions</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableSkeleton rows={5} columns={8} />
                </TableBody>
              </Table>
            ) : filteredOffers && filteredOffers.length > 0 ? (
              <>
                <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead>Category</TableHead>
                    <TableHead>Payout</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Clicks</TableHead>
                    <TableHead>Conversions</TableHead>
                    <TableHead className="w-[100px]">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {paginatedOffers?.map((offer) => (
                    <TableRow key={offer.id}>
                      <TableCell className="font-medium">{offer.title}</TableCell>
                      <TableCell>{offer.category || "-"}</TableCell>
                      <TableCell>${(offer.payout / 100).toFixed(2)}</TableCell>
                      <TableCell>
                        <Badge variant={offer.status === "active" ? "default" : "secondary"}>
                          {offer.status}
                        </Badge>
                      </TableCell>
                      <TableCell>{offer.totalClicks ?? 0}</TableCell>
                      <TableCell>{offer.totalConversions ?? 0}</TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          {offer.status === "pending" && (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-green-500 hover:text-green-600 hover:bg-green-500/10"
                                onClick={() => handleApprove(offer.id)}
                                disabled={approveMutation.isPending}
                                title="Approve"
                              >
                                <Check className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-red-500 hover:text-red-600 hover:bg-red-500/10"
                                onClick={() => openRejectDialog(offer)}
                                disabled={rejectMutation.isPending}
                                title="Reject"
                              >
                                <X className="h-4 w-4" />
                              </Button>
                            </>
                          )}
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setEditingOffer(offer)}
                            title="Edit"
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDelete(offer.id, offer.title)}
                            disabled={deleteMutation.isPending}
                            title="Delete"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
                </Table>
                {totalPages > 1 && (
                  <div className="flex items-center justify-between mt-4">
                    <p className="text-sm text-muted-foreground">
                      Showing {((currentPage - 1) * itemsPerPage) + 1} to {Math.min(currentPage * itemsPerPage, filteredOffers.length)} of {filteredOffers.length} offers
                    </p>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                        disabled={currentPage === 1}
                      >
                        Previous
                      </Button>
                      <span className="flex items-center px-3 text-sm">
                        Page {currentPage} of {totalPages}
                      </span>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                        disabled={currentPage === totalPages}
                      >
                        Next
                      </Button>
                    </div>
                  </div>
                )}
              </>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                No offers found
              </div>
            )}
          </CardContent>
        </Card>
      </div>
      
      {editingOffer && (
        <EditOfferDialog
          offer={editingOffer}
          open={!!editingOffer}
          onOpenChange={(open) => !open && setEditingOffer(null)}
        />
      )}

      {/* Reject Dialog */}
      <AlertDialog open={rejectDialogOpen} onOpenChange={setRejectDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Reject Offer</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to reject "{rejectingOffer?.title}"? The advertiser will be notified.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <label className="text-sm font-medium">Reason (optional)</label>
            <Textarea
              placeholder="Enter rejection reason..."
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
              className="mt-2"
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => {
              setRejectDialogOpen(false);
              setRejectingOffer(null);
              setRejectReason("");
            }}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleReject}
              className="bg-red-500 hover:bg-red-600"
              disabled={rejectMutation.isPending}
            >
              {rejectMutation.isPending ? "Rejecting..." : "Reject Offer"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </DashboardLayout>
  );
}
