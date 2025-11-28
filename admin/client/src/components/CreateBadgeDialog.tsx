import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { trpc } from "@/lib/trpc";
import { Plus } from "lucide-react";
import { toast } from "sonner";

export function CreateBadgeDialog() {
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    iconUrl: "",
    pointsReward: 0,
  });

  const utils = trpc.useUtils();
  const createMutation = trpc.badges.create.useMutation({
    onSuccess: () => {
      toast.success("Badge created successfully");
      utils.badges.list.invalidate();
      setOpen(false);
      setFormData({ name: "", description: "", iconUrl: "", pointsReward: 0 });
    },
    onError: (error) => {
      toast.error(`Failed to create badge: ${error.message}`);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create Badge
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create New Badge</DialogTitle>
            <DialogDescription>
              Add a new badge to the gamification system.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Name *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="iconUrl">Icon URL</Label>
              <Input
                id="iconUrl"
                type="url"
                value={formData.iconUrl}
                onChange={(e) => setFormData({ ...formData, iconUrl: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="pointsReward">Points Reward *</Label>
              <Input
                id="pointsReward"
                type="number"
                min="0"
                value={formData.pointsReward}
                onChange={(e) => setFormData({ ...formData, pointsReward: parseInt(e.target.value) || 0 })}
                required
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
