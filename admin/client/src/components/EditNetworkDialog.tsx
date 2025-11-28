import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { trpc } from "@/lib/trpc";
import { toast } from "sonner";

interface EditNetworkDialogProps {
  network: any;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditNetworkDialog({ network, open, onOpenChange }: EditNetworkDialogProps) {
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    logoUrl: "",
  });

  const utils = trpc.useUtils();
  const updateNetwork = trpc.networks.update.useMutation({
    onSuccess: () => {
      toast.success("Network updated successfully");
      utils.networks.list.invalidate();
      onOpenChange(false);
    },
    onError: (error) => {
      toast.error(error.message || "Failed to update network");
    },
  });

  useEffect(() => {
    if (network) {
      setFormData({
        name: network.name || "",
        description: network.description || "",
        logoUrl: network.logoUrl || "",
      });
    }
  }, [network]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.name) {
      toast.error("Network name is required");
      return;
    }

    updateNetwork.mutate({
      id: network.id,
      ...formData,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>Edit Network</DialogTitle>
          <DialogDescription>
            Update network information. Click save when you're done.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Network Name *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g., AffTok Finance Network"
                required
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Brief description of the network"
                rows={3}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="logoUrl">Logo URL</Label>
              <Input
                id="logoUrl"
                type="url"
                value={formData.logoUrl}
                onChange={(e) => setFormData({ ...formData, logoUrl: e.target.value })}
                placeholder="https://example.com/logo.png"
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={updateNetwork.isPending}>
              {updateNetwork.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}