"use client"

import { useState } from "react"
import { 
  Users, Info, Bell, Star, Hash, Lock, 
  ChevronRight, Plus, X, Pencil, Trash2, Pin 
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
Sheet,
SheetContent,
SheetDescription,
SheetHeader,
SheetTitle,
SheetTrigger,
} from "@/components/ui/sheet"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { useChannelStore } from "@/stores/channel-store"
import { useUserStore } from "@/stores/user-store"
import { useMessageStore } from "@/stores/message-store"
import { UserAvatar } from "@/components/common/user-avatar"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"

export function ChannelInfo({ trigger }: { trigger: React.ReactNode }) {
  const { currentChannel, members, updateChannel, addMember, removeMember } = useChannelStore()
  const { users, currentUser } = useUserStore()
  const { pinnedMessages, fetchPins, pinMessage } = useMessageStore()
  const [isEditing, setIsEditing] = useState(false)
  const [editForm, setEditForm] = useState({
    topic: currentChannel?.topic || "",
    purpose: currentChannel?.purpose || ""
  })
  const [showAddMember, setShowAddMember] = useState(false)

  if (!currentChannel) return null

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
                <div className="flex items-center justify-between py-1">
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded bg-muted">
                      <Bell className="w-4 h-4" />
                    </div>
                    <div>
                      <p className="text-sm font-bold">Notifications</p>
                      <p className="text-xs text-muted-foreground">All new messages</p>
                    </div>
                  </div>
                  <ChevronRight className="w-4 h-4 text-muted-foreground" />
                </div>
                
                <div className="flex items-center justify-between py-1">
                  <div className="flex items-center gap-3 text-red-500">
                    <div className="p-2 rounded bg-red-500/10">
                      <Trash2 className="w-4 h-4" />
                    </div>
                    <p className="text-sm font-bold">Leave channel</p>
                  </div>
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
                        <p className="text-xs text-muted-foreground truncate max-w-[200px]">{user.statusText || "Online"}</p>
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
    </Sheet>
  )
}
