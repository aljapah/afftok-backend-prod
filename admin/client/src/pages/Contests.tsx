import DashboardLayout from "@/components/DashboardLayout";
import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
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
import { Badge } from "@/components/ui/badge";
import { Trophy, Target, Users, Calendar, DollarSign, Clock, Plus, Edit, Trash2, Play, Square } from "lucide-react";
import { format } from "date-fns";

export default function Contests() {
  const utils = trpc.useUtils();

  const { data: contests, isLoading } = trpc.contests.list.useQuery();

  const createContest = trpc.contests.create.useMutation({
    onSuccess: () => utils.contests.list.invalidate(),
  });

  const updateContest = trpc.contests.update.useMutation({
    onSuccess: () => utils.contests.list.invalidate(),
  });

  const deleteContest = trpc.contests.delete.useMutation({
    onSuccess: () => utils.contests.list.invalidate(),
  });

  const activateContest = trpc.contests.activate.useMutation({
    onSuccess: () => utils.contests.list.invalidate(),
  });

  const endContest = trpc.contests.end.useMutation({
    onSuccess: () => utils.contests.list.invalidate(),
  });

  const [open, setOpen] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [current, setCurrent] = useState<any>(null);

  // Form fields
  const [title, setTitle] = useState("");
  const [titleAr, setTitleAr] = useState("");
  const [description, setDescription] = useState("");
  const [descriptionAr, setDescriptionAr] = useState("");
  const [imageUrl, setImageUrl] = useState("");
  const [prizeTitle, setPrizeTitle] = useState("");
  const [prizeTitleAr, setPrizeTitleAr] = useState("");
  const [prizeDescription, setPrizeDescription] = useState("");
  const [prizeAmount, setPrizeAmount] = useState("0");
  const [prizeCurrency, setPrizeCurrency] = useState("USD");
  const [contestType, setContestType] = useState("individual");
  const [targetType, setTargetType] = useState("clicks");
  const [targetValue, setTargetValue] = useState("100");
  const [minClicks, setMinClicks] = useState("0");
  const [minConversions, setMinConversions] = useState("0");
  const [minMembers, setMinMembers] = useState("1");
  const [maxParticipants, setMaxParticipants] = useState("0");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [status, setStatus] = useState("draft");

  const resetForm = () => {
    setTitle("");
    setTitleAr("");
    setDescription("");
    setDescriptionAr("");
    setImageUrl("");
    setPrizeTitle("");
    setPrizeTitleAr("");
    setPrizeDescription("");
    setPrizeAmount("0");
    setPrizeCurrency("USD");
    setContestType("individual");
    setTargetType("clicks");
    setTargetValue("100");
    setMinClicks("0");
    setMinConversions("0");
    setMinMembers("1");
    setMaxParticipants("0");
    setStartDate("");
    setEndDate("");
    setStatus("draft");
    setCurrent(null);
    setEditMode(false);
  };

  const openCreateModal = () => {
    resetForm();
    setOpen(true);
  };

  const openEditModal = (contest: any) => {
    setCurrent(contest);
    setTitle(contest.title);
    setTitleAr(contest.titleAr ?? "");
    setDescription(contest.description ?? "");
    setDescriptionAr(contest.descriptionAr ?? "");
    setImageUrl(contest.imageUrl ?? "");
    setPrizeTitle(contest.prizeTitle ?? "");
    setPrizeTitleAr(contest.prizeTitleAr ?? "");
    setPrizeDescription(contest.prizeDescription ?? "");
    setPrizeAmount(contest.prizeAmount?.toString() || "0");
    setPrizeCurrency(contest.prizeCurrency || "USD");
    setContestType(contest.contestType || "individual");
    setTargetType(contest.targetType || "clicks");
    setTargetValue(contest.targetValue?.toString() || "100");
    setMinClicks(contest.minClicks?.toString() || "0");
    setMinConversions(contest.minConversions?.toString() || "0");
    setMinMembers(contest.minMembers?.toString() || "1");
    setMaxParticipants(contest.maxParticipants?.toString() || "0");
    setStartDate(contest.startDate ? format(new Date(contest.startDate), "yyyy-MM-dd'T'HH:mm") : "");
    setEndDate(contest.endDate ? format(new Date(contest.endDate), "yyyy-MM-dd'T'HH:mm") : "");
    setStatus(contest.status || "draft");
    setEditMode(true);
    setOpen(true);
  };

  const handleSave = () => {
    const data = {
      title,
      titleAr: titleAr || null,
      description: description || null,
      descriptionAr: descriptionAr || null,
      imageUrl: imageUrl || null,
      prizeTitle: prizeTitle || null,
      prizeTitleAr: prizeTitleAr || null,
      prizeDescription: prizeDescription || null,
      prizeAmount: parseFloat(prizeAmount) || 0,
      prizeCurrency,
      contestType,
      targetType,
      targetValue: parseInt(targetValue) || 100,
      minClicks: parseInt(minClicks) || 0,
      minConversions: parseInt(minConversions) || 0,
      minMembers: parseInt(minMembers) || 1,
      maxParticipants: parseInt(maxParticipants) || 0,
      startDate: startDate ? new Date(startDate).toISOString() : new Date().toISOString(),
      endDate: endDate ? new Date(endDate).toISOString() : new Date().toISOString(),
      status,
    };

    if (editMode) {
      updateContest.mutate({ id: current.id, ...data });
    } else {
      createContest.mutate(data);
    }
    setOpen(false);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return <Badge className="bg-green-500/20 text-green-400 border-green-500/30">Active</Badge>;
      case "draft":
        return <Badge className="bg-gray-500/20 text-gray-400 border-gray-500/30">Draft</Badge>;
      case "ended":
        return <Badge className="bg-blue-500/20 text-blue-400 border-blue-500/30">Ended</Badge>;
      case "cancelled":
        return <Badge className="bg-red-500/20 text-red-400 border-red-500/30">Cancelled</Badge>;
      default:
        return <Badge>{status}</Badge>;
    }
  };

  const getTargetIcon = (type: string) => {
    switch (type) {
      case "clicks": return "🖱️";
      case "conversions": return "🛒";
      case "referrals": return "👥";
      case "points": return "⭐";
      default: return "🎯";
    }
  };

  if (isLoading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-pink-500"></div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Trophy className="h-8 w-8 text-yellow-500" />
            المسابقات / Contests
          </h1>
          <p className="text-muted-foreground mt-1">إدارة المسابقات والتحديات للمروجين والفرق</p>
        </div>
        <Button onClick={openCreateModal} className="gap-2">
          <Plus className="h-4 w-4" />
          إنشاء مسابقة جديدة
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4 mb-6">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-green-500/20 rounded-lg">
                <Play className="h-5 w-5 text-green-400" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">مسابقات نشطة</p>
                <p className="text-2xl font-bold">{contests?.filter(c => c.status === 'active').length || 0}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-gray-500/20 rounded-lg">
                <Clock className="h-5 w-5 text-gray-400" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">مسودات</p>
                <p className="text-2xl font-bold">{contests?.filter(c => c.status === 'draft').length || 0}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-500/20 rounded-lg">
                <Trophy className="h-5 w-5 text-blue-400" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">منتهية</p>
                <p className="text-2xl font-bold">{contests?.filter(c => c.status === 'ended').length || 0}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-yellow-500/20 rounded-lg">
                <Users className="h-5 w-5 text-yellow-400" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">إجمالي المشاركين</p>
                <p className="text-2xl font-bold">{contests?.reduce((a, c) => a + (c.participantsCount || 0), 0) || 0}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Contests Table */}
      <Card>
        <CardHeader>
          <CardTitle>قائمة المسابقات</CardTitle>
          <CardDescription>جميع المسابقات والتحديات في النظام</CardDescription>
        </CardHeader>
        <CardContent>
          {contests && contests.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>العنوان</TableHead>
                  <TableHead>النوع</TableHead>
                  <TableHead>الهدف</TableHead>
                  <TableHead>الجائزة</TableHead>
                  <TableHead>المشاركون</TableHead>
                  <TableHead>التاريخ</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead>الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {contests.map((contest) => (
                  <TableRow key={contest.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        {contest.imageUrl ? (
                          <img src={contest.imageUrl} alt="" className="w-10 h-10 rounded object-cover" />
                        ) : (
                          <div className="w-10 h-10 rounded bg-gradient-to-br from-pink-500 to-orange-500 flex items-center justify-center">
                            <Trophy className="h-5 w-5 text-white" />
                          </div>
                        )}
                        <div>
                          <p className="font-medium">{contest.title}</p>
                          {contest.titleAr && <p className="text-xs text-muted-foreground">{contest.titleAr}</p>}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {contest.contestType === 'team' ? '👥 فرق' : '👤 أفراد'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <span>{getTargetIcon(contest.targetType)}</span>
                        <span>{contest.targetValue}</span>
                        <span className="text-xs text-muted-foreground">
                          {contest.targetType === 'clicks' ? 'نقرة' : 
                           contest.targetType === 'conversions' ? 'تحويل' : 
                           contest.targetType === 'referrals' ? 'إحالة' : 'نقطة'}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-amber-400 font-medium">
                        {contest.prizeTitle || (contest.prizeAmount > 0 ? `${contest.prizeAmount} ${contest.prizeCurrency}` : 'جائزة خاصة')}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Users className="h-4 w-4 text-muted-foreground" />
                        <span>{contest.participantsCount || 0}</span>
                        {contest.maxParticipants > 0 && (
                          <span className="text-muted-foreground">/ {contest.maxParticipants}</span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-xs">
                        <div className="flex items-center gap-1">
                          <Calendar className="h-3 w-3" />
                          {format(new Date(contest.startDate), 'dd/MM/yyyy')}
                        </div>
                        <div className="text-muted-foreground">
                          → {format(new Date(contest.endDate), 'dd/MM/yyyy')}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{getStatusBadge(contest.status)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        {contest.status === 'draft' && (
                          <Button
                            size="sm"
                            variant="ghost"
                            className="h-8 w-8 p-0 text-green-400 hover:text-green-300"
                            onClick={() => activateContest.mutate({ id: contest.id })}
                            title="تفعيل"
                          >
                            <Play className="h-4 w-4" />
                          </Button>
                        )}
                        {contest.status === 'active' && (
                          <Button
                            size="sm"
                            variant="ghost"
                            className="h-8 w-8 p-0 text-red-400 hover:text-red-300"
                            onClick={() => endContest.mutate({ id: contest.id })}
                            title="إنهاء"
                          >
                            <Square className="h-4 w-4" />
                          </Button>
                        )}
                        <Button
                          size="sm"
                          variant="ghost"
                          className="h-8 w-8 p-0"
                          onClick={() => openEditModal(contest)}
                          title="تعديل"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          className="h-8 w-8 p-0 text-red-400 hover:text-red-300"
                          onClick={() => deleteContest.mutate({ id: contest.id })}
                          title="حذف"
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
            <div className="text-center py-12">
              <Trophy className="h-16 w-16 mx-auto text-muted-foreground/30 mb-4" />
              <p className="text-lg text-muted-foreground">لا توجد مسابقات</p>
              <p className="text-sm text-muted-foreground/70 mb-4">ابدأ بإنشاء أول مسابقة</p>
              <Button onClick={openCreateModal} className="gap-2">
                <Plus className="h-4 w-4" />
                إنشاء مسابقة
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create/Edit Modal */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Trophy className="h-5 w-5 text-yellow-500" />
              {editMode ? "تعديل المسابقة" : "إنشاء مسابقة جديدة"}
            </DialogTitle>
          </DialogHeader>

          <div className="space-y-6">
            {/* Basic Info */}
            <div className="space-y-4">
              <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wide">معلومات أساسية</h3>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium mb-2 block">العنوان (English) *</label>
                  <Input
                    placeholder="Contest Title"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                  />
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">العنوان (عربي)</label>
                  <Input
                    placeholder="عنوان المسابقة"
                    value={titleAr}
                    onChange={(e) => setTitleAr(e.target.value)}
                    dir="rtl"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium mb-2 block">الوصف (English)</label>
                  <Textarea
                    placeholder="Contest description..."
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    rows={2}
                  />
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">الوصف (عربي)</label>
                  <Textarea
                    placeholder="وصف المسابقة..."
                    value={descriptionAr}
                    onChange={(e) => setDescriptionAr(e.target.value)}
                    rows={2}
                    dir="rtl"
                  />
                </div>
              </div>

              <div>
                <label className="text-sm font-medium mb-2 block">صورة المسابقة (URL)</label>
                <Input
                  placeholder="https://example.com/image.png"
                  value={imageUrl}
                  onChange={(e) => setImageUrl(e.target.value)}
                />
              </div>
            </div>

            {/* Prize Info */}
            <div className="space-y-4">
              <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                <DollarSign className="h-4 w-4" />
                الجائزة
              </h3>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium mb-2 block">عنوان الجائزة (English)</label>
                  <Input
                    placeholder="iPhone 15 Pro Max"
                    value={prizeTitle}
                    onChange={(e) => setPrizeTitle(e.target.value)}
                  />
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">عنوان الجائزة (عربي)</label>
                  <Input
                    placeholder="آيفون 15 برو ماكس"
                    value={prizeTitleAr}
                    onChange={(e) => setPrizeTitleAr(e.target.value)}
                    dir="rtl"
                  />
                </div>
              </div>

              <div>
                <label className="text-sm font-medium mb-2 block">وصف الجائزة</label>
                <Textarea
                  placeholder="تفاصيل إضافية عن الجائزة..."
                  value={prizeDescription}
                  onChange={(e) => setPrizeDescription(e.target.value)}
                  rows={2}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium mb-2 block">قيمة الجائزة (رقمياً)</label>
                  <Input
                    type="number"
                    placeholder="1000"
                    value={prizeAmount}
                    onChange={(e) => setPrizeAmount(e.target.value)}
                    min="0"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">العملة</label>
                  <Select value={prizeCurrency} onValueChange={setPrizeCurrency}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="USD">USD - دولار</SelectItem>
                      <SelectItem value="SAR">SAR - ريال سعودي</SelectItem>
                      <SelectItem value="AED">AED - درهم إماراتي</SelectItem>
                      <SelectItem value="EGP">EGP - جنيه مصري</SelectItem>
                      <SelectItem value="EUR">EUR - يورو</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>

            {/* Contest Rules */}
            <div className="space-y-4">
              <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                <Target className="h-4 w-4" />
                شروط المسابقة
              </h3>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium mb-2 block">نوع المسابقة</label>
                  <Select value={contestType} onValueChange={setContestType}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="individual">👤 أفراد (Individual)</SelectItem>
                      <SelectItem value="team">👥 فرق (Teams)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">نوع الهدف</label>
                  <Select value={targetType} onValueChange={setTargetType}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="clicks">🖱️ نقرات (Clicks)</SelectItem>
                      <SelectItem value="conversions">🛒 تحويلات (Conversions)</SelectItem>
                      <SelectItem value="referrals">👥 إحالات (Referrals)</SelectItem>
                      <SelectItem value="points">⭐ نقاط (Points)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="text-sm font-medium mb-2 block">قيمة الهدف</label>
                  <Input
                    type="number"
                    placeholder="100"
                    value={targetValue}
                    onChange={(e) => setTargetValue(e.target.value)}
                    min="1"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">الحد الأقصى للمشاركين</label>
                  <Input
                    type="number"
                    placeholder="0 = غير محدود"
                    value={maxParticipants}
                    onChange={(e) => setMaxParticipants(e.target.value)}
                    min="0"
                  />
                </div>
                {contestType === 'team' && (
                  <div>
                    <label className="text-sm font-medium mb-2 block">الحد الأدنى لأعضاء الفريق</label>
                    <Input
                      type="number"
                      placeholder="1"
                      value={minMembers}
                      onChange={(e) => setMinMembers(e.target.value)}
                      min="1"
                    />
                  </div>
                )}
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium mb-2 block">الحد الأدنى للنقرات (للمشاركة)</label>
                  <Input
                    type="number"
                    placeholder="0"
                    value={minClicks}
                    onChange={(e) => setMinClicks(e.target.value)}
                    min="0"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">الحد الأدنى للتحويلات (للمشاركة)</label>
                  <Input
                    type="number"
                    placeholder="0"
                    value={minConversions}
                    onChange={(e) => setMinConversions(e.target.value)}
                    min="0"
                  />
                </div>
              </div>
            </div>

            {/* Timing */}
            <div className="space-y-4">
              <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                التوقيت
              </h3>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium mb-2 block">تاريخ البدء *</label>
                  <Input
                    type="datetime-local"
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                  />
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">تاريخ الانتهاء *</label>
                  <Input
                    type="datetime-local"
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                  />
                </div>
              </div>
            </div>

            {/* Status */}
            <div>
              <label className="text-sm font-medium mb-2 block">الحالة</label>
              <Select value={status} onValueChange={setStatus}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="draft">📝 مسودة (Draft)</SelectItem>
                  <SelectItem value="active">✅ نشطة (Active)</SelectItem>
                  <SelectItem value="ended">🏁 منتهية (Ended)</SelectItem>
                  <SelectItem value="cancelled">❌ ملغاة (Cancelled)</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              إلغاء
            </Button>
            <Button 
              onClick={handleSave} 
              disabled={!title || !startDate || !endDate}
              className="gap-2"
            >
              {editMode ? "حفظ التغييرات" : "إنشاء المسابقة"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </DashboardLayout>
  );
}

