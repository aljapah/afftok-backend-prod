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
import { Plus, Globe, Languages } from "lucide-react";
import { toast } from "sonner";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";


export function CreateOfferDialog() {
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    titleAr: "",
    descriptionAr: "",
    termsAr: "",
    imageUrl: "",
    destinationUrl: "",
    category: "",
    payout: 0,
    commission: 0,
  });

  const utils = trpc.useUtils();
  const createMutation = trpc.offers.create.useMutation({
    onSuccess: () => {
      toast.success("Offer created successfully");
      utils.offers.list.invalidate();
      setOpen(false);
      setFormData({
        title: "",
        description: "",
        titleAr: "",
        descriptionAr: "",
        termsAr: "",
        imageUrl: "",
        destinationUrl: "",
        category: "",
        payout: 0,
        commission: 0,
      });
    },
    onError: (error) => {
      toast.error(`Failed to create offer: ${error.message}`);
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
          Create Offer
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[650px] max-h-[90vh] overflow-y-auto">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create New Offer</DialogTitle>
            <DialogDescription>
              Add a new offer to the platform. Fill in all required fields.
            </DialogDescription>
          </DialogHeader>
          
          <Tabs defaultValue="english" className="mt-4">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="english" className="flex items-center gap-2">
                <Globe className="h-4 w-4" />
                English
              </TabsTrigger>
              <TabsTrigger value="arabic" className="flex items-center gap-2">
                <Languages className="h-4 w-4" />
                العربية
              </TabsTrigger>
            </TabsList>
            
            {/* English Tab */}
            <TabsContent value="english" className="space-y-4 mt-4">
              <div className="grid gap-2">
                <Label htmlFor="title">Title *</Label>
                <Input
                  id="title"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  placeholder="Enter offer title in English"
                  required
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Enter offer description in English"
                  rows={3}
                />
              </div>
            </TabsContent>
            
            {/* Arabic Tab */}
            <TabsContent value="arabic" className="space-y-4 mt-4" dir="rtl">
              <div className="grid gap-2">
                <Label htmlFor="titleAr" className="text-right">العنوان بالعربي</Label>
                <Input
                  id="titleAr"
                  value={formData.titleAr}
                  onChange={(e) => setFormData({ ...formData, titleAr: e.target.value })}
                  placeholder="أدخل عنوان العرض بالعربي"
                  className="text-right"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="descriptionAr" className="text-right">الوصف بالعربي</Label>
                <Textarea
                  id="descriptionAr"
                  value={formData.descriptionAr}
                  onChange={(e) => setFormData({ ...formData, descriptionAr: e.target.value })}
                  placeholder="أدخل وصف العرض بالعربي"
                  className="text-right"
                  rows={3}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="termsAr" className="text-right">الشروط والأحكام</Label>
                <Textarea
                  id="termsAr"
                  value={formData.termsAr}
                  onChange={(e) => setFormData({ ...formData, termsAr: e.target.value })}
                  placeholder="أدخل شروط وأحكام العرض بالعربي"
                  className="text-right"
                  rows={3}
                />
              </div>
            </TabsContent>
          </Tabs>
          
          {/* Common Fields */}
          <div className="grid gap-4 py-4 border-t mt-4 pt-4">
            <h4 className="text-sm font-medium text-muted-foreground">Common Settings</h4>
            
            <div className="grid gap-2">
              <Label htmlFor="imageUrl">Image URL</Label>
              <Input
                id="imageUrl"
                type="url"
                value={formData.imageUrl}
                onChange={(e) => setFormData({ ...formData, imageUrl: e.target.value })}
                placeholder="https://example.com/image.png"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="destinationUrl">Destination URL *</Label>
              <Input
                id="destinationUrl"
                type="url"
                value={formData.destinationUrl}
                onChange={(e) => setFormData({ ...formData, destinationUrl: e.target.value })}
                placeholder="https://example.com/offer"
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="category">Category *</Label>
              <Select value={formData.category} onValueChange={(value) => setFormData({ ...formData, category: value })} required>
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
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="payout">Payout (cents) *</Label>
                <Input
                  id="payout"
                  type="number"
                  min="0"
                  value={formData.payout}
                  onChange={(e) => setFormData({ ...formData, payout: parseInt(e.target.value) || 0 })}
                  required
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="commission">Commission (cents) *</Label>
                <Input
                  id="commission"
                  type="number"
                  min="0"
                  value={formData.commission}
                  onChange={(e) => setFormData({ ...formData, commission: parseInt(e.target.value) || 0 })}
                  required
                />
              </div>
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