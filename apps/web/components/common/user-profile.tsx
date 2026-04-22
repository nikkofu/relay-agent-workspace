"use client"

import { User } from "@/types"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Sparkles, MessageCircle, Mail, Clock, Calendar, Users, MapPin, Phone } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { useUIStore } from "@/stores/ui-store"
import { format, formatDistanceToNow } from "date-fns"
import { useUserStore } from "@/stores/user-store"
import { useState } from "react"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogDescription,
  DialogTrigger,
  DialogFooter
} from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from "@/components/ui/select"

interface UserProfileProps {
  user: User
}

export function UserProfile({ user }: UserProfileProps) {
  const { currentUser, updateStatus, updateProfile } = useUserStore()
  const { openDockedChat } = useUIStore()
  const [isEditingStatus, setIsEditingStatus] = useState(false)
  const [isEditingProfile, setIsEditingProfile] = useState(false)
  const [newStatusText, setNewStatusText] = useState(user.statusText || "")
  const [newStatusEmoji, setNewStatusEmoji] = useState(user.statusEmoji || "💬")
  const [expiryDuration, setExpiryDuration] = useState("0")
  
  const [editForm, setEditForm] = useState({
    title: user.title || "",
    department: user.department || "",
    timezone: user.timezone || "UTC+8",
    pronouns: user.pronouns || "",
    location: user.location || "",
    phone: user.phone || "",
    bio: user.bio || ""
  })
  
  const statusColors: Record<string, string> = {
    online: "bg-green-500",
    away: "bg-amber-500",
    busy: "bg-red-500",
    offline: "bg-slate-500"
  }

  const handleMessageClick = () => {
    openDockedChat(user.id)
  }

  const handleStatusUpdate = async () => {
    if (currentUser) {
      await updateStatus(currentUser.id, {
        status: currentUser.status,
        statusText: newStatusText,
        statusEmoji: newStatusEmoji,
        expiresInMinutes: parseInt(expiryDuration)
      })
      setIsEditingStatus(false)
    }
  }

  const handleProfileUpdate = async () => {
    if (currentUser) {
      await updateProfile(currentUser.id, editForm)
      setIsEditingProfile(false)
    }
  }

  return (
    <div className="w-[300px] flex flex-col bg-white dark:bg-[#1a1d21] overflow-hidden">
      {/* Header with Background and Large Avatar */}
      <div className="h-24 bg-gradient-to-br from-purple-600 to-blue-600 relative">
        <div className="absolute -bottom-10 left-4 p-1 rounded-xl bg-white dark:bg-[#1a1d21] shadow-xl">
          <Avatar className="h-20 w-20 rounded-lg">
            <AvatarImage src={user.avatar} />
            <AvatarFallback className="rounded-lg text-2xl font-bold">{user.name.substring(0, 2).toUpperCase()}</AvatarFallback>
          </Avatar>
        </div>
      </div>

      <div className="px-4 pt-12 pb-4 flex flex-col gap-4">
        {/* Name and Status */}
        <div className="flex flex-col gap-0.5">
          <div className="flex items-center gap-2">
            <h3 className="text-xl font-black tracking-tight text-foreground">{user.name}</h3>
            {user.pronouns && <span className="text-xs text-muted-foreground font-normal">({user.pronouns})</span>}
            {user.statusEmoji && <span className="text-lg">{user.statusEmoji}</span>}
            <div className={`h-2.5 w-2.5 rounded-full ${statusColors[user.status] || statusColors.offline}`} />
          </div>
          <div className="flex flex-col">
            {currentUser?.id === user.id ? (
              <Popover open={isEditingStatus} onOpenChange={setIsEditingStatus}>
                <PopoverTrigger asChild>
                  <button className="text-sm text-muted-foreground font-medium hover:text-purple-600 transition-colors text-left truncate">
                    {user.statusText || "Set a status"}
                  </button>
                </PopoverTrigger>
                <PopoverContent className="w-[280px] p-3" align="start">
                  <div className="space-y-4">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">Update your status</h4>
                    <div className="flex gap-2">
                      <div className="w-10">
                        <Input 
                          placeholder="Icon" 
                          value={newStatusEmoji}
                          onChange={(e) => setNewStatusEmoji(e.target.value)}
                          className="text-center px-0 bg-white dark:bg-black"
                        />
                      </div>
                      <Input 
                        placeholder="What's happening?" 
                        value={newStatusText}
                        onChange={(e) => setNewStatusText(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleStatusUpdate()}
                        className="flex-1 bg-white dark:bg-black"
                      />
                    </div>
                    
                    <div className="space-y-1.5">
                      <Label className="text-[10px] uppercase font-bold text-muted-foreground">Clear after</Label>
                      <Select value={expiryDuration} onValueChange={setExpiryDuration}>
                        <SelectTrigger className="h-8 text-xs bg-white dark:bg-black">
                          <SelectValue placeholder="Don't clear" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="0">Don't clear</SelectItem>
                          <SelectItem value="30">30 minutes</SelectItem>
                          <SelectItem value="60">1 hour</SelectItem>
                          <SelectItem value="240">4 hours</SelectItem>
                          <SelectItem value="1440">Today</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="flex justify-end gap-2 pt-2">
                      <Button variant="ghost" size="sm" className="h-8 text-xs" onClick={() => setIsEditingStatus(false)}>Cancel</Button>
                      <Button size="sm" className="h-8 text-xs bg-[#3f0e40] text-white hover:bg-[#3f0e40]/90" onClick={handleStatusUpdate}>Save</Button>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
            ) : (
              <p className="text-sm text-muted-foreground font-medium">{user.statusText || (user.status === 'online' ? "Active" : "Away")}</p>
            )}
            
            {user.statusExpiresAt && (
              <p className="text-[9px] text-purple-600 font-bold uppercase tracking-tight mt-1 flex items-center gap-1">
                <Clock className="w-2.5 h-2.5" />
                Until {format(new Date(user.statusExpiresAt), 'p')}
              </p>
            )}
            
            {user.status === 'offline' && user.lastSeen && (
              <p className="text-[10px] text-muted-foreground italic mt-0.5">
                Last seen {formatDistanceToNow(new Date(user.lastSeen), { addSuffix: true })}
              </p>
            )}
          </div>
        </div>

        {/* AI Insight Section */}
        <div className="bg-purple-50 dark:bg-purple-900/20 p-3 rounded-lg border border-purple-100 dark:border-purple-800/30">
          <div className="flex items-center gap-2 mb-1.5 text-purple-600 dark:text-purple-400">
            <Sparkles className="w-3.5 h-3.5" />
            <span className="text-[10px] font-bold uppercase tracking-wider">AI Collaboration Insight</span>
          </div>
          <p className="text-[11px] leading-relaxed text-foreground/80 italic font-medium">
            &quot;{user.aiInsight || `${user.name} is an active collaborator. Insight data will refine over time.`}&quot;
          </p>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-2">
          {currentUser?.id === user.id ? (
            <Dialog open={isEditingProfile} onOpenChange={setIsEditingProfile}>
              <DialogTrigger asChild>
                <Button variant="outline" className="flex-1 h-9 font-bold">Edit Profile</Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                  <DialogTitle className="text-xl font-black">Edit your profile</DialogTitle>
                  <DialogDescription className="text-xs">
                    Make changes to your public profile here. Click save when you're done.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4 max-h-[60vh] overflow-y-auto px-1 custom-scrollbar">
                  <div className="grid gap-2">
                    <Label htmlFor="title">Job Title</Label>
                    <Input 
                      id="title" 
                      value={editForm.title} 
                      onChange={(e) => setEditForm(prev => ({ ...prev, title: e.target.value }))}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="dept">Department</Label>
                    <Input 
                      id="dept" 
                      value={editForm.department}
                      onChange={(e) => setEditForm(prev => ({ ...prev, department: e.target.value }))}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="pronouns">Pronouns</Label>
                    <Input 
                      id="pronouns" 
                      placeholder="e.g. they/them"
                      value={editForm.pronouns}
                      onChange={(e) => setEditForm(prev => ({ ...prev, pronouns: e.target.value }))}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="bio">Bio</Label>
                    <Input 
                      id="bio" 
                      placeholder="Tell us about yourself"
                      value={editForm.bio}
                      onChange={(e) => setEditForm(prev => ({ ...prev, bio: e.target.value }))}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="loc">Location</Label>
                    <Input 
                      id="loc" 
                      value={editForm.location}
                      onChange={(e) => setEditForm(prev => ({ ...prev, location: e.target.value }))}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="phone">Phone</Label>
                    <Input 
                      id="phone" 
                      value={editForm.phone}
                      onChange={(e) => setEditForm(prev => ({ ...prev, phone: e.target.value }))}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="tz">Timezone</Label>
                    <Input 
                      id="tz" 
                      value={editForm.timezone}
                      onChange={(e) => setEditForm(prev => ({ ...prev, timezone: e.target.value }))}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="ghost" onClick={() => setIsEditingProfile(false)}>Cancel</Button>
                  <Button className="bg-[#3f0e40] text-white" onClick={handleProfileUpdate}>Save Changes</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          ) : (
            <Button 
              className="flex-1 h-9 font-bold bg-[#007a5a] hover:bg-[#007a5a]/90 text-white"
              onClick={handleMessageClick}
            >
              <MessageCircle className="w-4 h-4 mr-2" /> Message
            </Button>
          )}
          <Button variant="outline" size="icon" className="h-9 w-9">
            <Mail className="w-4 h-4 text-foreground" />
          </Button>
        </div>

        <Separator className="opacity-50" />

        <div className="grid grid-cols-1 gap-3">
          <DetailItem icon={Clock} label="Local time" value={user.profile?.localTime || "10:45 AM"} />
          <DetailItem icon={Calendar} label="Role / Title" value={user.title || "Software Engineer"} />
          <DetailItem icon={Users} label="Department" value={user.department || "Engineering"} />
          {user.location && <DetailItem icon={MapPin} label="Location" value={user.location} />}
          {user.phone && <DetailItem icon={Phone} label="Phone" value={user.phone} />}
        </div>
        {user.bio && (
          <div className="mt-2 p-3 bg-muted/30 rounded-lg border border-dashed">
            <p className="text-[10px] uppercase font-bold text-muted-foreground mb-1 tracking-widest">About</p>
            <p className="text-xs text-foreground/80 leading-relaxed italic">{user.bio}</p>
          </div>
        )}
      </div>
    </div>
  )
}

function DetailItem({ icon: Icon, label, value }: { icon: React.ElementType, label: string, value: string }) {
  return (
    <div className="flex items-start gap-3 group">
      <div className="w-8 h-8 rounded bg-muted flex items-center justify-center shrink-0">
        <Icon className="w-4 h-4 text-muted-foreground" />
      </div>
      <div className="flex flex-col">
        <span className="text-[10px] text-muted-foreground font-bold uppercase tracking-tighter">{label}</span>
        <span className="text-xs font-bold text-foreground/90">{value}</span>
      </div>
    </div>
  )
}
