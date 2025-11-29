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

interface EditBadgeDialogProps {
  badge: any;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditBadgeDialog({ badge, open, onOpenChange }: EditBadgeDialogProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [iconUrl, setIconUrl] = useState("");
  const [pointsReward, setPointsReward] = useState("");

  const utils = trpc.useUtils();
  const updateBadge = trpc.badges.update.useMutation({
    onSuccess: () => {
      toast.success("Badge updated successfully");
      utils.badges.list.invalidate();
      onOpenChange(false);
    },
    onError: (error) => {
      toast.error(error.message || "Failed to update badge");
    },
  });

  useEffect(() => {
    if (badge) {
      setName(badge.name || "");
      setDescription(badge.description || "");
      setIconUrl(badge.iconUrl || "");
      setPointsReward(badge.pointsReward?.toString() || "");
    }
  }, [badge]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!name || !description || !pointsReward) {
      toast.error("Please fill in all required fields");
      return;
    }

    const points = parseInt(pointsReward);
    if (isNaN(points) || points < 0 || points > 100000) {
      toast.error("Points reward must be between 0 and 100,000");
      return;
    }

    updateBadge.mutate({
      id: badge.id,
      name,
      description,
      iconUrl: iconUrl || undefined,
      pointsReward: points,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Edit Badge</DialogTitle>
          <DialogDescription>
            Update badge details. Click save when you're done.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Badge Name *</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Top Promoter"
                required
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="description">Description *</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Describe how to earn this badge..."
                rows={3}
                required
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="iconUrl">Icon URL</Label>
              <Input
                id="iconUrl"
                type="url"
                value={iconUrl}
                onChange={(e) => setIconUrl(e.target.value)}
                placeholder="https://example.com/badge-icon.png"
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="pointsReward">Points Reward *</Label>
              <Input
                id="pointsReward"
                type="number"
                value={pointsReward}
                onChange={(e) => setPointsReward(e.target.value)}
                placeholder="e.g., 1000"
                min="0"
                max="100000"
                required
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
            <Button type="submit" disabled={updateBadge.isPending}>
              {updateBadge.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
