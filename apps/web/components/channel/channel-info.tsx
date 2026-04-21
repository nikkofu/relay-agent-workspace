"use client"

import { useState } from "react"
import { 
  Bell, Hash, Lock, Plus, X, Trash2, Pin,
  Sparkles, Loader2, RefreshCw, FileText, Download, File
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { 
  Sheet, 
  SheetContent, 
  SheetHeader, 
  SheetTitle, 
  SheetTrigger 
} from "@/components/ui/sheet"
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from "@/components/ui/select"
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogFooter,
} from "@/components/ui/dialog"
import { Switch } from "@/components/ui/switch"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { useChannelStore } from "@/stores/channel-store"
import { useUserStore } from "@/stores/user-store"
import { useMessageStore } from "@/stores/message-store"
import { useFileStore } from "@/stores/file-store"
import { UserAvatar } from "@/components/common/user-avatar"
import { formatDistanceToNow } from "date-fns"

export function ChannelInfo({ trigger }: { trigger: React.ReactNode }) {
  const { 
    currentChannel, members, updateChannel, addMember, removeMember,
    currentChannelSummary, isSummaryLoading, generateChannelSummary,
    fetchChannelPreferences, updateChannelPreferences, leaveChannel
  } = useChannelStore()
  const { users, currentUser } = useUserStore()
  const { pinnedMessages, fetchPins, pinMessage } = useMessageStore()
  const { files, fetchFiles } = useFileStore()
  const [isEditing, setIsEditing] = useState(false)
  const [editForm, setEditForm] = useState({
    topic: currentChannel?.topic || "",
    purpose: currentChannel?.purpose || ""
  })
  const [showAddMember, setShowAddMember] = useState(false)
  
  const [prefs, setPrefs] = useState<any>(null)
  const [isConfirmingLeave, setIsConfirmingLeave] = useState(false)

  if (!currentChannel) return null

  const handleFetchPrefs = async () => {
    const data = await fetchChannelPreferences(currentChannel.id)
    setPrefs(data)
  }

  const handleUpdate = () => {
    updateChannel(currentChannel.id, editForm)
    setIsEditing(false)
  }

  const handleAddMember = (userId: string) => {
    addMember(currentChannel.id, userId)
    setShowAddMember(false)
  }

  return (
    <Sheet onOpenChange={(open) => {
      if (open && currentChannel) {
        fetchPins(currentChannel.id)
        fetchFiles({ channelId: currentChannel.id })
        handleFetchPrefs()
      }
    }}>
      <SheetTrigger asChild>
        {trigger}
      </SheetTrigger>
      <SheetContent className="w-[400px] sm:w-[540px] p-0 flex flex-col gap-0 border-l bg-white dark:bg-[#1a1d21]">
        <SheetHeader className="p-4 border-b shrink-0">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {currentChannel.type === "private" ? <Lock className="w-5 h-5" /> : <Hash className="w-5 h-5" />}
              <SheetTitle className="text-xl font-bold">{currentChannel.name}</SheetTitle>
            </div>
          </div>
        </SheetHeader>

        <Tabs defaultValue="about" className="flex-1 flex flex-col min-h-0">
          <div className="px-4 border-b bg-muted/30">
            <TabsList className="bg-transparent h-12 p-0 gap-6">
              <TabsTrigger value="about" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">About</TabsTrigger>
              <TabsTrigger value="members" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">Members <span className="ml-1 text-xs opacity-50">{members.length}</span></TabsTrigger>
              <TabsTrigger value="pins" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">Pins <span className="ml-1 text-xs opacity-50">{pinnedMessages.length}</span></TabsTrigger>
              <TabsTrigger value="files" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">Files <span className="ml-1 text-xs opacity-50">{files.length}</span></TabsTrigger>
              <TabsTrigger value="settings" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">Settings</TabsTrigger>
            </TabsList>
          </div>

          <ScrollArea className="flex-1">
            <TabsContent value="about" className="p-4 m-0 flex flex-col gap-6">
              {/* Topic Section */}
              <div className="space-y-2 group">
                <div className="flex items-center justify-between">
                  <Label className="text-sm font-bold">Topic</Label>
                  {!isEditing && (
                    <Button variant="ghost" size="sm" className="h-7 text-[#1164a3] hover:text-[#1164a3] hover:bg-[#1164a3]/10" onClick={() => setIsEditing(true)}>
                      Edit
                    </Button>
                  )}
                </div>
                {isEditing ? (
                  <div className="space-y-2">
                    <Input 
                      value={editForm.topic} 
                      onChange={(e) => setEditForm(s => ({ ...s, topic: e.target.value }))}
                      placeholder="Add a topic"
                      className="bg-white dark:bg-black"
                    />
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground italic">
                    {currentChannel.topic || "Add a topic to help people know what this channel is for."}
                  </p>
                )}
              </div>

              {/* AI Summary Section */}
              <div className="space-y-3 bg-purple-50/50 dark:bg-purple-900/10 p-4 rounded-xl border border-purple-100 dark:border-purple-800/30">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-purple-700 dark:text-purple-400">
                    <Sparkles className="w-4 h-4" />
                    <span className="text-sm font-bold">AI Channel Summary</span>
                  </div>
                  {currentChannelSummary && !isSummaryLoading && (
                    <Button variant="ghost" size="icon" className="h-6 w-6 text-purple-400 hover:text-purple-600 hover:bg-transparent" onClick={() => generateChannelSummary(currentChannel.id)}>
                      <RefreshCw className="w-3.5 h-3.5" />
                    </Button>
                  )}
                </div>
                
                {isSummaryLoading ? (
                  <div className="flex items-center gap-2 text-xs text-muted-foreground animate-pulse py-2">
                    <Loader2 className="w-3 h-3 animate-spin" />
                    Analyzing channel history...
                  </div>
                ) : currentChannelSummary ? (
                  <p className="text-sm leading-relaxed text-foreground/80 font-medium">
                    {currentChannelSummary}
                  </p>
                ) : (
                  <div className="space-y-3 py-1">
                    <p className="text-xs text-muted-foreground">
                      Get a high-level overview of recent activities, decisions, and discussions in this channel.
                    </p>
                    <Button 
                      size="sm" 
                      className="w-full bg-purple-600 hover:bg-purple-700 text-white font-bold text-xs h-8"
                      onClick={() => generateChannelSummary(currentChannel.id)}
                    >
                      Generate Summary
                    </Button>
                  </div>
                )}
              </div>

              {/* Purpose Section */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label className="text-sm font-bold">Purpose</Label>
                </div>
                {isEditing ? (
                  <div className="space-y-2">
                    <Textarea 
                      value={editForm.purpose} 
                      onChange={(e) => setEditForm(s => ({ ...s, purpose: e.target.value }))}
                      placeholder="Add a purpose"
                      className="bg-white dark:bg-black min-h-[100px]"
                    />
                    <div className="flex items-center gap-2 justify-end mt-2">
                      <Button variant="outline" size="sm" onClick={() => setIsEditing(false)}>Cancel</Button>
                      <Button size="sm" className="bg-[#007a5a] hover:bg-[#007a5a]/90" onClick={handleUpdate}>Save</Button>
                    </div>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">
                    {currentChannel.purpose || "Add a purpose to tell people what this channel is used for."}
                  </p>
                )}
              </div>

              <div className="pt-4 border-t space-y-4">
                <div className="space-y-3">
                  <div className="flex items-center justify-between py-1">
                    <div className="flex items-center gap-3">
                      <div className="p-2 rounded bg-muted">
                        <Bell className="w-4 h-4" />
                      </div>
                      <div>
                        <p className="text-sm font-bold">Notifications</p>
                        <p className="text-[10px] text-muted-foreground uppercase font-black tracking-tighter">
                          {prefs?.notification_level || 'all'} {prefs?.is_muted && '• Muted'}
                        </p>
                      </div>
                    </div>
                  </div>
                  
                  {prefs && (
                    <div className="space-y-4 p-4 bg-muted/20 rounded-xl border border-border/50">
                      <div className="space-y-2">
                        <Label className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">Notification Level</Label>
                        <Select 
                          value={prefs.notification_level} 
                          onValueChange={(val) => updateChannelPreferences(currentChannel.id, { notification_level: val }).then(handleFetchPrefs)}
                        >
                          <SelectTrigger className="h-8 text-xs bg-white dark:bg-black">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="all">All new messages</SelectItem>
                            <SelectItem value="mentions">Just mentions</SelectItem>
                            <SelectItem value="none">Nothing</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                      <div className="flex items-center justify-between">
                        <Label className="text-xs font-medium">Mute channel</Label>
                        <Switch 
                          checked={prefs.is_muted} 
                          onCheckedChange={(val) => updateChannelPreferences(currentChannel.id, { is_muted: val }).then(handleFetchPrefs)}
                        />
                      </div>
                    </div>
                  )}
                </div>
                
                <div className="pt-2">
                  <Button 
                    variant="ghost" 
                    className="w-full justify-start gap-3 p-0 h-auto hover:bg-transparent text-red-500 hover:text-red-600 group"
                    onClick={() => setIsConfirmingLeave(true)}
                  >
                    <div className="p-2 rounded bg-red-500/10 group-hover:bg-red-500/20 transition-colors">
                      <Trash2 className="w-4 h-4" />
                    </div>
                    <p className="text-sm font-bold">Leave channel</p>
                  </Button>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="members" className="p-0 m-0 flex flex-col">
              <div className="p-4 border-b">
                <Button 
                  variant="outline" 
                  className="w-full justify-start gap-2 h-10 border-dashed"
                  onClick={() => setShowAddMember(true)}
                >
                  <Plus className="w-4 h-4" />
                  Add People
                </Button>
              </div>

              {showAddMember && (
                <div className="p-4 bg-muted/30 border-b space-y-3">
                  <Label className="text-xs font-bold uppercase tracking-wider">Select a member to add</Label>
                  <div className="flex flex-col gap-1 max-h-[200px] overflow-y-auto pr-2">
                    {users
                      .filter(u => !members.find(m => m.user.id === u.id))
                      .map(user => (
                        <div 
                          key={user.id}
                          className="flex items-center justify-between p-2 hover:bg-white dark:hover:bg-black rounded cursor-pointer group"
                          onClick={() => handleAddMember(user.id)}
                        >
                          <div className="flex items-center gap-3">
                            <UserAvatar src={user.avatar} name={user.name} status={user.status} className="h-8 w-8" />
                            <span className="text-sm font-medium">{user.name}</span>
                          </div>
                          <Plus className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100" />
                        </div>
                      ))}
                  </div>
                  <Button variant="ghost" size="sm" className="w-full h-8" onClick={() => setShowAddMember(false)}>Cancel</Button>
                </div>
              )}

              <div className="flex flex-col">
                {members.map(({ user, role }) => (
                  <div key={user.id} className="flex items-center justify-between p-4 hover:bg-muted/30 group border-b last:border-0 transition-colors">
                    <div className="flex items-center gap-3">
                      <UserAvatar src={user.avatar} name={user.name} status={user.status} />
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-bold">{user.name}</span>
                          {user.id === currentUser?.id && <span className="text-[10px] bg-muted px-1.5 py-0.5 rounded uppercase font-bold text-muted-foreground">You</span>}
                          {role === 'admin' && <span className="text-[10px] bg-blue-500/10 text-blue-500 px-1.5 py-0.5 rounded uppercase font-bold">Admin</span>}
                        </div>
                        <p className="text-xs text-muted-foreground truncate max-w-[200px]">
                          {user.status === 'offline' && user.lastSeen 
                            ? `Last seen ${formatDistanceToNow(new Date(user.lastSeen), { addSuffix: true })}`
                            : (user.statusText || (user.status === 'online' ? "Active" : "Away"))}
                        </p>
                      </div>
                    </div>
                    {currentUser?.id !== user.id && (
                      <Button 
                        variant="ghost" 
                        size="icon" 
                        className="h-8 w-8 text-muted-foreground hover:text-red-500 hover:bg-red-500/10 opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={() => removeMember(currentChannel.id, user.id)}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            </TabsContent>

            <TabsContent value="pins" className="p-0 m-0 flex flex-col">
              {pinnedMessages.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-20 px-10 text-center text-muted-foreground">
                  <div className="w-12 h-12 rounded-full bg-muted flex items-center justify-center mb-4">
                    <Pin className="w-6 h-6 rotate-45" />
                  </div>
                  <p className="text-sm font-bold">No pinned messages</p>
                  <p className="text-xs mt-1 italic">Pin important messages, files, or common questions so anyone can find them.</p>
                </div>
              ) : (
                <div className="flex flex-col">
                  {pinnedMessages.map(({ message, user }) => (
                    <div key={message.id} className="p-4 hover:bg-muted/30 group border-b last:border-0 transition-colors">
                      <div className="flex items-start gap-3">
                        <UserAvatar src={user?.avatar} name={user?.name || "Unknown"} className="h-8 w-8" />
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-sm font-bold">{user?.name || "Someone"}</span>
                            <span className="text-[10px] text-muted-foreground uppercase font-medium ml-2">
                              {formatDistanceToNow(new Date(message.createdAt), { addSuffix: true })}
                            </span>
                          </div>
                          <div className="text-sm mt-1 break-words line-clamp-3 text-foreground" dangerouslySetInnerHTML={{ __html: message.content }} />
                        </div>
                        <Button 
                          variant="ghost" 
                          size="icon" 
                          className="h-8 w-8 text-muted-foreground hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
                          onClick={() => pinMessage(message.id)}
                        >
                          <X className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </TabsContent>

            <TabsContent value="files" className="p-0 m-0 flex flex-col">
              {files.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-20 px-10 text-center text-muted-foreground">
                  <div className="w-12 h-12 rounded-full bg-muted flex items-center justify-center mb-4">
                    <File className="w-6 h-6" />
                  </div>
                  <p className="text-sm font-bold">No files yet</p>
                  <p className="text-xs mt-1 italic">Files shared in this channel will appear here for everyone to find.</p>
                </div>
              ) : (
                <div className="flex flex-col">
                  {files.map((file) => (
                    <div key={file.id} className="p-4 hover:bg-muted/30 group border-b last:border-0 transition-colors">
                      <div className="flex items-center justify-between gap-3">
                        <div className="flex items-center gap-3 min-w-0">
                          <div className="w-8 h-8 rounded bg-blue-500/10 flex items-center justify-center shrink-0 text-blue-600">
                            <FileText className="w-4 h-4" />
                          </div>
                          <div className="min-w-0">
                            <p className="text-sm font-bold truncate">{file.name}</p>
                            <p className="text-[10px] text-muted-foreground uppercase font-bold tracking-wider">
                              {(file.size / 1024).toFixed(1)} KB • {formatDistanceToNow(new Date(file.createdAt), { addSuffix: true })}
                            </p>
                          </div>
                        </div>
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-foreground shrink-0" asChild>
                          <a href={file.url} target="_blank" rel="noopener noreferrer" download={file.name}>
                            <Download className="w-4 h-4" />
                          </a>
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </TabsContent>

            <TabsContent value="settings" className="p-4 m-0 space-y-6">
              <div className="space-y-4">
                <div className="space-y-1">
                  <p className="text-sm font-bold">Archive Channel</p>
                  <p className="text-xs text-muted-foreground">When you archive a channel, it will be closed to new messages. People can still browse the message history.</p>
                </div>
                <Button variant="outline" className="text-red-500 border-red-500 hover:bg-red-500/10 w-full justify-start">
                  Archive this channel
                </Button>
              </div>

              <div className="pt-4 border-t space-y-4">
                <div className="space-y-1">
                  <p className="text-sm font-bold">Channel ID</p>
                  <p className="text-xs text-muted-foreground bg-muted p-2 rounded font-mono">{currentChannel.id}</p>
                </div>
              </div>
            </TabsContent>
          </ScrollArea>
        </Tabs>
      </SheetContent>

      <Dialog open={isConfirmingLeave} onOpenChange={setIsConfirmingLeave}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle className="text-xl font-black">Leave #{currentChannel.name}?</DialogTitle>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-muted-foreground leading-relaxed">
              Are you sure you want to leave this channel? You can always join back later if the channel is public.
            </p>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setIsConfirmingLeave(false)}>Cancel</Button>
            <Button variant="destructive" onClick={() => leaveChannel(currentChannel.id)}>Leave Channel</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Sheet>
  )
}
