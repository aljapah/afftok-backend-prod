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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { trpc } from "@/lib/trpc";
import { toast } from "sonner";

interface EditOfferDialogProps {
  offer: any;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditOfferDialog({ offer, open, onOpenChange }: EditOfferDialogProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState("");
  const [payout, setPayout] = useState("");
  const [targetUrl, setTargetUrl] = useState("");
  const [imageUrl, setImageUrl] = useState("");
  const [requirements, setRequirements] = useState("");
  const [status, setStatus] = useState<"active" | "inactive" | "pending">("active");

  const utils = trpc.useUtils();
  const updateOffer = trpc.offers.update.useMutation({
    onSuccess: () => {
      toast.success("Offer updated successfully");
      utils.offers.list.invalidate();
      onOpenChange(false);
    },
    onError: (error) => {
      toast.error(error.message || "Failed to update offer");
    },
  });

  useEffect(() => {
    if (offer) {
      setTitle(offer.title || "");
      setDescription(offer.description || "");
      setCategory(offer.category || "");
      setPayout(offer.payout?.toString() || "");
      setTargetUrl(offer.targetUrl || "");
      setImageUrl(offer.imageUrl || "");
      setRequirements(offer.requirements || "");
      setStatus(offer.status || "active");
    }
  }, [offer]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!title || !description || !category || !payout || !targetUrl) {
      toast.error("Please fill in all required fields");
      return;
    }

    const payoutNum = parseInt(payout);
    if (isNaN(payoutNum) || payoutNum < 0 || payoutNum > 1000000) {
      toast.error("Payout must be between 0 and 1,000,000 cents");
      return;
    }

    updateOffer.mutate({
      id: offer.id,
      title,
      description,
      category,
      payout: payoutNum,
      destinationUrl: targetUrl,
      imageUrl: imageUrl || undefined,
      status: status,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Offer</DialogTitle>
          <DialogDescription>
            Update offer details. Click save when you're done.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="title">Title *</Label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g., Get 50% off on Premium Subscription"
                required
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="description">Description *</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Detailed description of the offer..."
                rows={4}
                required
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="category">Category *</Label>
                <Select value={category} onValueChange={setCategory} required>
                  <SelectTrigger>
                    <SelectValue placeholder="Select category" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="Cashback">Cashback</SelectItem>
                    <SelectItem value="E-commerce">E-commerce</SelectItem>
                    <SelectItem value="Finance">Finance</SelectItem>
                    <SelectItem value="Travel">Travel</SelectItem>
                    <SelectItem value="Education">Education</SelectItem>
                    <SelectItem value="Technology">Technology</SelectItem>
                    <SelectItem value="Utilities">Utilities</SelectItem>
                    <SelectItem value="Food & Restaurants">Food & Restaurants</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="payout">Payout (cents) *</Label>
                <Input
                  id="payout"
                  type="number"
                  value={payout}
                  onChange={(e) => setPayout(e.target.value)}
                  placeholder="e.g., 500 ($5.00)"
                  min="0"
                  max="1000000"
                  required
                />
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="targetUrl">Target URL *</Label>
              <Input
                id="targetUrl"
                type="url"
                value={targetUrl}
                onChange={(e) => setTargetUrl(e.target.value)}
                placeholder="https://example.com/offer"
                required
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="imageUrl">Image URL</Label>
              <Input
                id="imageUrl"
                type="url"
                value={imageUrl}
                onChange={(e) => setImageUrl(e.target.value)}
                placeholder="https://example.com/image.jpg"
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="requirements">Requirements</Label>
              <Textarea
                id="requirements"
                value={requirements}
                onChange={(e) => setRequirements(e.target.value)}
                placeholder="e.g., 18+, US only"
                rows={2}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="status">Status *</Label>
              <Select value={status} onValueChange={(value) => setStatus(value as "active" | "inactive" | "pending")} required>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="inactive">Inactive</SelectItem>
                </SelectContent>
              </Select>
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
            <Button type="submit" disabled={updateOffer.isPending}>
              {updateOffer.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}