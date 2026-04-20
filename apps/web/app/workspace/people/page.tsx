"use client"

import { useEffect, useState } from "react"
import { useUserStore } from "@/stores/user-store"
import { useDirectoryStore } from "@/stores/directory-store"
import { Search, Filter, Users, Building, Clock, ChevronRight, Plus, Trash2, Shield, MoreVertical } from "lucide-react"
import { 
  Tabs, 
  TabsContent, 
  TabsList, 
  TabsTrigger 
} from "@/components/ui/tabs"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from "@/components/ui/select"
import { UserAvatar } from "@/components/common/user-avatar"

export default function PeopleDirectoryPage() {
  const { users, fetchUsers } = useUserStore()
  const { userGroups, fetchUserGroups, createGroup, deleteGroup } = useDirectoryStore()
  
  const [activeTab, setActiveTab] = useState("people")
  const [q, setQ] = useState("")
  const [department, setDepartment] = useState("all")
  const [status, setStatus] = useState("all")
  const [userGroupId, setUserGroupId] = useState("all")

  // Group Create State
  const [isCreatingGroup, setIsCreatingGroup] = useState(false)
  const [newGroup, setNewGroup] = useState({ name: "", handle: "", description: "" })

  useEffect(() => {
    fetchUserGroups()
  }, [fetchUserGroups])

  useEffect(() => {
    const params: any = {}
    if (q) params.q = q
    if (department !== "all") params.department = department
    if (status !== "all") params.status = status
    if (userGroupId !== "all") params.userGroupId = userGroupId
    
    fetchUsers(params)
  }, [q, department, status, userGroupId, fetchUsers])

  const handleCreateGroup = async () => {
    await createGroup(newGroup.name, newGroup.handle, newGroup.description)
    setIsCreatingGroup(false)
    setNewGroup({ name: "", handle: "", description: "" })
  }

  const departments = Array.from(new Set(users.map(u => u.department).filter(Boolean)))

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      <header className="h-14 px-6 flex items-center border-b shrink-0 justify-between">
        <div className="flex items-center gap-2">
          <Users className="w-5 h-5 text-purple-600" />
          <h1 className="text-lg font-black tracking-tight uppercase">People & Groups</h1>
        </div>
        {activeTab === "groups" && (
          <Dialog open={isCreatingGroup} onOpenChange={setIsCreatingGroup}>
            <DialogTrigger asChild>
              <Button size="sm" className="h-8 font-bold bg-[#3f0e40] text-white">
                <Plus className="w-3.5 h-3.5 mr-1.5" /> Create Group
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle className="text-xl font-black">Create User Group</DialogTitle>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="name">Group Name</Label>
                  <Input 
                    id="name" 
                    placeholder="e.g. Engineering Team" 
                    value={newGroup.name}
                    onChange={e => setNewGroup(p => ({ ...p, name: e.target.value }))}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="handle">Handle</Label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground text-sm">@</span>
                    <Input 
                      id="handle" 
                      className="pl-7" 
                      placeholder="eng-team" 
                      value={newGroup.handle}
                      onChange={e => setNewGroup(p => ({ ...p, handle: e.target.value }))}
                    />
                  </div>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="desc">Description</Label>
                  <Input 
                    id="desc" 
                    placeholder="Short summary of this group" 
                    value={newGroup.description}
                    onChange={e => setNewGroup(p => ({ ...p, description: e.target.value }))}
                  />
                </div>
              </div>
              <DialogFooter>
                <Button variant="ghost" onClick={() => setIsCreatingGroup(false)}>Cancel</Button>
                <Button className="bg-[#3f0e40] text-white" onClick={handleCreateGroup}>Create Group</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        )}
      </header>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col min-h-0">
        <div className="px-6 border-b bg-muted/5">
          <TabsList className="h-12 bg-transparent gap-6">
            <TabsTrigger value="people" className="h-12 rounded-none border-b-2 border-transparent data-[state=active]:border-purple-600 data-[state=active]:bg-transparent px-0 text-xs font-bold uppercase tracking-widest">Directory</TabsTrigger>
            <TabsTrigger value="groups" className="h-12 rounded-none border-b-2 border-transparent data-[state=active]:border-purple-600 data-[state=active]:bg-transparent px-0 text-xs font-bold uppercase tracking-widest">User Groups</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="people" className="flex-1 flex flex-col min-h-0 m-0">
          {/* Filter Bar */}
          <div className="p-4 border-b bg-muted/10 flex flex-wrap gap-3">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input 
                placeholder="Search by name or email..." 
                className="pl-9 bg-white dark:bg-[#1a1d21]"
                value={q}
                onChange={(e) => setQ(e.target.value)}
              />
            </div>

            <Select value={department} onValueChange={setDepartment}>
              <SelectTrigger className="w-[160px] bg-white dark:bg-[#1a1d21]">
                <SelectValue placeholder="Department" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Departments</SelectItem>
                {departments.map(dept => (
                  <SelectItem key={dept} value={dept!}>{dept}</SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select value={status} onValueChange={setStatus}>
              <SelectTrigger className="w-[140px] bg-white dark:bg-[#1a1d21]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="online">Online</SelectItem>
                <SelectItem value="away">Away</SelectItem>
                <SelectItem value="busy">Busy</SelectItem>
                <SelectItem value="offline">Offline</SelectItem>
              </SelectContent>
            </Select>

            <Select value={userGroupId} onValueChange={setUserGroupId}>
              <SelectTrigger className="w-[160px] bg-white dark:bg-[#1a1d21]">
                <SelectValue placeholder="User Group" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Groups</SelectItem>
                {userGroups.map(group => (
                  <SelectItem key={group.id} value={group.id}>{group.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Button variant="ghost" size="icon" onClick={() => {
              setQ("")
              setDepartment("all")
              setStatus("all")
              setUserGroupId("all")
            }}>
              <Filter className="w-4 h-4" />
            </Button>
          </div>

          <ScrollArea className="flex-1">
            <div className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {users.map(user => (
                  <div 
                    key={user.id}
                    className="p-4 rounded-xl border bg-white dark:bg-card/50 hover:shadow-md transition-all group cursor-pointer"
                  >
                    <div className="flex items-start gap-4">
                      <UserAvatar src={user.avatar} name={user.name} status={user.status} className="w-12 h-12" />
                      <div className="flex-1 min-w-0">
                        <h3 className="font-bold text-sm truncate group-hover:text-purple-600 transition-colors">{user.name}</h3>
                        <p className="text-[11px] text-muted-foreground truncate">{user.title || 'No Title'}</p>
                        <div className="flex flex-col gap-1.5 mt-3">
                          <div className="flex items-center gap-2 text-[10px] text-muted-foreground font-medium">
                            <Building className="w-3 h-3" />
                            {user.department || 'General'}
                          </div>
                          <div className="flex items-center gap-2 text-[10px] text-muted-foreground font-medium">
                            <Clock className="w-3 h-3" />
                            {user.profile?.localTime || 'Local Time Unknown'}
                          </div>
                        </div>
                      </div>
                      <ChevronRight className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity self-center" />
                    </div>
                    {user.statusText && (
                      <div className="mt-3 pt-3 border-t border-dashed">
                        <p className="text-[10px] italic text-muted-foreground line-clamp-1">&ldquo;{user.statusText}&rdquo;</p>
                      </div>
                    )}
                  </div>
                ))}
              </div>
              
              {users.length === 0 && (
                <div className="py-20 text-center flex flex-col items-center gap-3">
                  <Users className="w-12 h-12 text-muted-foreground/20" />
                  <p className="text-muted-foreground font-medium">No results found matching your filters.</p>
                  <Button variant="link" onClick={() => {
                    setQ("")
                    setDepartment("all")
                    setStatus("all")
                    setUserGroupId("all")
                  }}>Clear all filters</Button>
                </div>
              )}
            </div>
          </ScrollArea>
        </TabsContent>

        <TabsContent value="groups" className="flex-1 flex flex-col min-h-0 m-0">
          <ScrollArea className="flex-1">
            <div className="p-8 max-w-4xl mx-auto space-y-4">
              {userGroups.map(group => (
                <div key={group.id} className="p-4 rounded-xl border bg-white dark:bg-card/50 flex items-center justify-between group/card">
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 rounded-xl bg-purple-500/10 flex items-center justify-center text-purple-600">
                      <Shield className="w-6 h-6" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="font-bold text-sm">{group.name}</h3>
                        <Badge variant="outline" className="text-[9px] font-black uppercase tracking-tighter h-4">@{group.handle}</Badge>
                      </div>
                      <p className="text-[11px] text-muted-foreground mt-0.5 line-clamp-1">{group.description || 'No description provided.'}</p>
                      <p className="text-[10px] font-bold text-purple-600/70 mt-1 uppercase tracking-widest">{group.memberCount} Members</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button variant="ghost" size="sm" className="h-8 text-xs font-bold">View Members</Button>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground opacity-0 group-hover/card:opacity-100 transition-opacity">
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem>Edit Group</DropdownMenuItem>
                        <DropdownMenuItem className="text-red-600" onClick={() => deleteGroup(group.id)}>
                          <Trash2 className="w-3.5 h-3.5 mr-2" />
                          Delete Group
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </div>
              ))}

              {userGroups.length === 0 && (
                <div className="py-20 text-center flex flex-col items-center gap-3">
                  <Shield className="w-12 h-12 text-muted-foreground/20" />
                  <p className="text-muted-foreground font-medium">No user groups created yet.</p>
                  <Button variant="link" onClick={() => setIsCreatingGroup(true)}>Create your first group</Button>
                </div>
              )}
            </div>
          </ScrollArea>
        </TabsContent>
      </Tabs>
    </div>
  )
}
