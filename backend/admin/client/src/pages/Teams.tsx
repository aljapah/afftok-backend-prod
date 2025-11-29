import DashboardLayout from "@/components/DashboardLayout";
import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { trpc } from "@/lib/trpc";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export default function Teams() {
  const utils = trpc.useUtils();

  const { data: teams, isLoading } = trpc.teams.list.useQuery();
  const { data: users } = trpc.users.list.useQuery();

  const createTeam = trpc.teams.create.useMutation({
    onSuccess: () => utils.teams.list.invalidate(),
  });

  const updateTeam = trpc.teams.update.useMutation({
    onSuccess: () => utils.teams.list.invalidate(),
  });

  const deleteTeam = trpc.teams.delete.useMutation({
    onSuccess: () => utils.teams.list.invalidate(),
  });

  const [open, setOpen] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [current, setCurrent] = useState<any>(null);

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [logoUrl, setLogoUrl] = useState("");
  const [ownerId, setOwnerId] = useState("");
  const [maxMembers, setMaxMembers] = useState("10");

  const resetForm = () => {
    setName("");
    setDescription("");
    setLogoUrl("");
    setOwnerId("");
    setMaxMembers("10");
    setCurrent(null);
    setEditMode(false);
  };

  const openCreateModal = () => {
    resetForm();
    setOpen(true);
  };

  const openEditModal = (team: any) => {
    setCurrent(team);
    setName(team.name);
    setDescription(team.description ?? "");
    setLogoUrl(team.logoUrl ?? "");
    setOwnerId(team.ownerId);
    setMaxMembers(team.maxMembers?.toString() || "10");
    setEditMode(true);
    setOpen(true);
  };

  const handleSave = () => {
    if (editMode) {
      updateTeam.mutate({
        id: current.id,
        name,
        description: description || null,
        logoUrl: logoUrl || null,
        ownerId,
        maxMembers: parseInt(maxMembers),
      });
    } else {
      createTeam.mutate({
        name,
        description: description || null,
        logoUrl: logoUrl || null,
        ownerId,
        maxMembers: parseInt(maxMembers),
      });
    }
    setOpen(false);
  };

  if (isLoading) {
    return <DashboardLayout>Loading...</DashboardLayout>;
  }

  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Teams</h1>
        <Button onClick={openCreateModal}>Create New Team</Button>
      </div>

      <div>
        <Card>
          <CardHeader>
            <CardTitle>Teams List</CardTitle>
          </CardHeader>
          <CardContent>
            {teams && teams.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Logo</TableHead>
                    <TableHead>Members</TableHead>
                    <TableHead>Points</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {teams.map((team) => (
                    <TableRow key={team.id}>
                      <TableCell className="font-medium">{team.name}</TableCell>
                      <TableCell>{team.description}</TableCell>
                      <TableCell>
                        {team.logoUrl && (
                          <img src={team.logoUrl} alt={team.name} className="w-8 h-8 rounded" />
                        )}
                      </TableCell>
                      <TableCell>{team.memberCount || 0}</TableCell>
                      <TableCell>{team.totalPoints || 0}</TableCell>
                      <TableCell className="flex space-x-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => openEditModal(team)}
                        >
                          Edit
                        </Button>

                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => deleteTeam.mutate({ id: team.id })}
                        >
                          Delete
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                No teams found
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Modal */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editMode ? "Edit Team" : "Create Team"}</DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-2 block">Team Name *</label>
              <Input
                placeholder="Enter team name"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block">Description</label>
              <Textarea
                placeholder="Enter team description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
              />
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block">Logo URL</label>
              <Input
                placeholder="https://example.com/logo.png"
                value={logoUrl}
                onChange={(e) => setLogoUrl(e.target.value)}
              />
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block">Team Owner *</label>
              <Select value={ownerId} onValueChange={setOwnerId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select team owner" />
                </SelectTrigger>
                <SelectContent>
                  {users?.map((user) => (
                    <SelectItem key={user.id} value={user.id}>
                      {user.fullName || user.username} ({user.email})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block">Max Members</label>
              <Input
                type="number"
                placeholder="10"
                value={maxMembers}
                onChange={(e) => setMaxMembers(e.target.value)}
                min="1"
                max="100"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={!name || !ownerId}>
              {editMode ? "Save Changes" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </DashboardLayout>
  );
}
