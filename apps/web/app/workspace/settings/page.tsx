"use client"

import { useEffect, useState } from "react"
import { useNotificationStore } from "@/stores/notification-store"
import { useUserStore } from "@/stores/user-store"
import { useTheme } from "next-themes"
import { Bell, Shield, Settings, User, Palette, Sun, Moon, Monitor, Check } from "lucide-react"
import { API_BASE_URL } from "@/lib/constants"
import { Switch } from "@/components/ui/switch"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { toast } from "sonner"

type Tab = "profile" | "appearance" | "notifications" | "privacy"

const TABS: { id: Tab; label: string; icon: React.ElementType }[] = [
  { id: "profile", label: "Profile", icon: User },
  { id: "appearance", label: "Appearance", icon: Palette },
  { id: "notifications", label: "Notifications", icon: Bell },
  { id: "privacy", label: "Privacy", icon: Shield },
]

const THEME_OPTIONS = [
  { value: "light", label: "Light", icon: Sun },
  { value: "dark", label: "Dark", icon: Moon },
  { value: "system", label: "System", icon: Monitor },
]

const DENSITY_OPTIONS = [
  { value: "comfortable", label: "Comfortable", description: "More space between messages" },
  { value: "compact", label: "Compact", description: "Tighter spacing, more content visible" },
]

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<Tab>("profile")
  const { preferences, isLoading: notifLoading, fetchPreferences, updatePreferences } = useNotificationStore()
  const { currentUser, updateProfile } = useUserStore()
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)
  const [density, setDensity] = useState("comfortable")
  const [isSaving, setIsSaving] = useState(false)

  const [profileForm, setProfileForm] = useState({
    name: "",
    title: "",
    department: "",
    timezone: "",
    pronouns: "",
    location: "",
    phone: "",
    bio: "",
  })

  useEffect(() => {
    setMounted(true)
    fetchPreferences()
    const loadSettings = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/me/settings`)
        if (res.ok) {
          const data = await res.json()
          const s = data.settings
          if (s.theme) setTheme(s.theme)
          if (s.message_density) {
            setDensity(s.message_density)
            localStorage.setItem("relay-density", s.message_density)
          }
        }
      } catch {
        const saved = localStorage.getItem("relay-density")
        if (saved) setDensity(saved)
      }
    }
    loadSettings()
  }, [fetchPreferences, setTheme])

  useEffect(() => {
    if (currentUser) {
      setProfileForm({
        name: currentUser.name || "",
        title: currentUser.title || "",
        department: currentUser.department || "",
        timezone: currentUser.timezone || "UTC+8",
        pronouns: currentUser.pronouns || "",
        location: currentUser.location || "",
        phone: currentUser.phone || "",
        bio: currentUser.bio || "",
      })
    }
  }, [currentUser])

  const handleProfileSave = async () => {
    if (!currentUser) return
    setIsSaving(true)
    await updateProfile(currentUser.id, {
      title: profileForm.title,
      department: profileForm.department,
      timezone: profileForm.timezone,
      pronouns: profileForm.pronouns,
      location: profileForm.location,
      phone: profileForm.phone,
      bio: profileForm.bio,
    })
    setIsSaving(false)
    toast.success("Profile updated")
  }

  const patchSettings = async (patch: Record<string, string>) => {
    try {
      await fetch(`${API_BASE_URL}/me/settings`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(patch),
      })
    } catch (e) {
      console.error("Failed to persist settings", e)
    }
  }

  const handleThemeChange = (val: string) => {
    setTheme(val)
    patchSettings({ theme: val })
  }

  const handleDensityChange = (val: string) => {
    setDensity(val)
    localStorage.setItem("relay-density", val)
    patchSettings({ message_density: val })
    toast.success(`Density set to ${val}`)
  }

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      <header className="h-14 px-6 flex items-center border-b shrink-0">
        <Settings className="w-5 h-5 mr-2 text-purple-600" />
        <h1 className="text-lg font-black tracking-tight">Settings</h1>
      </header>

      <div className="flex flex-1 min-h-0">
        {/* Sidebar tabs */}
        <nav className="w-52 border-r shrink-0 py-4 flex flex-col gap-0.5 px-3">
          {TABS.map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                "flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm font-semibold transition-all text-left",
                activeTab === tab.id
                  ? "bg-purple-500/10 text-purple-600 dark:text-purple-400"
                  : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
              )}
            >
              <tab.icon className="w-4 h-4 shrink-0" />
              {tab.label}
            </button>
          ))}
        </nav>

        <ScrollArea className="flex-1">
          <div className="p-8 max-w-2xl space-y-8">

            {/* ── Profile Tab ── */}
            {activeTab === "profile" && (
              <div className="space-y-6">
                <div>
                  <h2 className="text-lg font-black">Profile</h2>
                  <p className="text-sm text-muted-foreground mt-0.5">Your personal information visible to workspace members.</p>
                </div>

                {/* Avatar row */}
                <Card className="shadow-none border-border/50">
                  <CardContent className="pt-5 flex items-center gap-5">
                    <Avatar className="h-16 w-16 rounded-xl">
                      <AvatarImage src={currentUser?.avatar} />
                      <AvatarFallback className="text-lg font-black bg-purple-500/20 text-purple-600 rounded-xl">
                        {currentUser?.name?.substring(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-bold text-base">{currentUser?.name}</p>
                      <p className="text-xs text-muted-foreground">{currentUser?.email}</p>
                      <Badge variant="outline" className={cn(
                        "text-[9px] mt-1.5 font-black uppercase h-4 px-1.5",
                        currentUser?.status === "online" ? "border-green-500/30 text-green-600 bg-green-500/5" :
                        currentUser?.status === "away" ? "border-amber-500/30 text-amber-600 bg-amber-500/5" :
                        "border-slate-500/30 text-slate-500 bg-slate-500/5"
                      )}>{currentUser?.status}</Badge>
                    </div>
                  </CardContent>
                </Card>

                {/* Form fields */}
                <Card className="shadow-none border-border/50">
                  <CardHeader className="pb-4">
                    <CardTitle className="text-sm font-bold">Personal Info</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-1.5">
                        <Label className="text-xs font-bold">Display Name</Label>
                        <Input value={profileForm.name} disabled className="bg-muted/30 text-xs" />
                        <p className="text-[10px] text-muted-foreground">Contact an admin to change your name.</p>
                      </div>
                      <div className="space-y-1.5">
                        <Label className="text-xs font-bold">Title / Role</Label>
                        <Input
                          value={profileForm.title}
                          onChange={e => setProfileForm(p => ({ ...p, title: e.target.value }))}
                          placeholder="e.g. Software Engineer"
                          className="text-xs"
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-1.5">
                        <Label className="text-xs font-bold">Department</Label>
                        <Input
                          value={profileForm.department}
                          onChange={e => setProfileForm(p => ({ ...p, department: e.target.value }))}
                          placeholder="e.g. Engineering"
                          className="text-xs"
                        />
                      </div>
                      <div className="space-y-1.5">
                        <Label className="text-xs font-bold">Timezone</Label>
                        <Input
                          value={profileForm.timezone}
                          onChange={e => setProfileForm(p => ({ ...p, timezone: e.target.value }))}
                          placeholder="e.g. UTC+8"
                          className="text-xs"
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-1.5">
                        <Label className="text-xs font-bold">Pronouns</Label>
                        <Input
                          value={profileForm.pronouns}
                          onChange={e => setProfileForm(p => ({ ...p, pronouns: e.target.value }))}
                          placeholder="e.g. they/them"
                          className="text-xs"
                        />
                      </div>
                      <div className="space-y-1.5">
                        <Label className="text-xs font-bold">Location</Label>
                        <Input
                          value={profileForm.location}
                          onChange={e => setProfileForm(p => ({ ...p, location: e.target.value }))}
                          placeholder="e.g. San Francisco, CA"
                          className="text-xs"
                        />
                      </div>
                    </div>

                    <div className="space-y-1.5">
                      <Label className="text-xs font-bold">Phone</Label>
                      <Input
                        value={profileForm.phone}
                        onChange={e => setProfileForm(p => ({ ...p, phone: e.target.value }))}
                        placeholder="+1 (555) 000-0000"
                        className="text-xs"
                      />
                    </div>

                    <div className="space-y-1.5">
                      <Label className="text-xs font-bold">Bio</Label>
                      <Textarea
                        value={profileForm.bio}
                        onChange={e => setProfileForm(p => ({ ...p, bio: e.target.value }))}
                        placeholder="A short intro about yourself…"
                        className="text-xs resize-none"
                        rows={3}
                      />
                    </div>
                  </CardContent>
                </Card>

                <div className="flex justify-end">
                  <Button
                    onClick={handleProfileSave}
                    disabled={isSaving}
                    className="bg-purple-600 hover:bg-purple-700 text-white font-bold"
                  >
                    {isSaving ? "Saving…" : "Save Profile"}
                  </Button>
                </div>
              </div>
            )}

            {/* ── Appearance Tab ── */}
            {activeTab === "appearance" && mounted && (
              <div className="space-y-6">
                <div>
                  <h2 className="text-lg font-black">Appearance</h2>
                  <p className="text-sm text-muted-foreground mt-0.5">Customize how Relay looks and feels.</p>
                </div>

                <Card className="shadow-none border-border/50">
                  <CardHeader className="pb-3">
                    <CardTitle className="text-sm font-bold">Theme</CardTitle>
                    <CardDescription className="text-xs">Choose between light, dark, or follow your system setting.</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-3 gap-3">
                      {THEME_OPTIONS.map(opt => (
                        <button
                          key={opt.value}
                          onClick={() => handleThemeChange(opt.value)}
                          className={cn(
                            "relative flex flex-col items-center gap-2.5 p-4 rounded-xl border-2 transition-all",
                            theme === opt.value
                              ? "border-purple-500 bg-purple-500/5"
                              : "border-border/60 hover:border-border"
                          )}
                        >
                          {theme === opt.value && (
                            <span className="absolute top-2 right-2 w-4 h-4 rounded-full bg-purple-500 flex items-center justify-center">
                              <Check className="w-2.5 h-2.5 text-white" strokeWidth={3} />
                            </span>
                          )}
                          <div className={cn(
                            "w-10 h-10 rounded-lg flex items-center justify-center",
                            opt.value === "light" ? "bg-slate-100 text-slate-700" :
                            opt.value === "dark" ? "bg-slate-800 text-slate-200" :
                            "bg-gradient-to-br from-slate-100 to-slate-800 text-slate-500"
                          )}>
                            <opt.icon className="w-5 h-5" />
                          </div>
                          <span className="text-xs font-bold">{opt.label}</span>
                        </button>
                      ))}
                    </div>
                  </CardContent>
                </Card>

                <Card className="shadow-none border-border/50">
                  <CardHeader className="pb-3">
                    <CardTitle className="text-sm font-bold">Message Density</CardTitle>
                    <CardDescription className="text-xs">Control the spacing between messages in channels and DMs.</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-2">
                    {DENSITY_OPTIONS.map(opt => (
                      <button
                        key={opt.value}
                        onClick={() => handleDensityChange(opt.value)}
                        className={cn(
                          "w-full flex items-center justify-between p-3 rounded-lg border-2 transition-all text-left",
                          density === opt.value
                            ? "border-purple-500 bg-purple-500/5"
                            : "border-border/60 hover:border-border"
                        )}
                      >
                        <div>
                          <p className="text-sm font-bold">{opt.label}</p>
                          <p className="text-[11px] text-muted-foreground">{opt.description}</p>
                        </div>
                        {density === opt.value && (
                          <div className="w-5 h-5 rounded-full bg-purple-500 flex items-center justify-center shrink-0">
                            <Check className="w-3 h-3 text-white" strokeWidth={3} />
                          </div>
                        )}
                      </button>
                    ))}
                  </CardContent>
                </Card>
              </div>
            )}

            {/* ── Notifications Tab ── */}
            {activeTab === "notifications" && (
              <div className="space-y-6">
                <div>
                  <h2 className="text-lg font-black">Notifications</h2>
                  <p className="text-sm text-muted-foreground mt-0.5">Control how and when you receive alerts.</p>
                </div>

                {notifLoading || !preferences ? (
                  <p className="text-sm text-muted-foreground">Loading preferences…</p>
                ) : (
                  <Card className="shadow-none border-border/50">
                    <CardContent className="pt-5 space-y-5">
                      {[
                        { key: "inboxEnabled" as const, label: "Inbox Notifications", desc: "Receive alerts for new messages in your inbox." },
                        { key: "mentionsEnabled" as const, label: "Mentions & Replies", desc: "Get notified when someone @mentions you or replies to your threads." },
                        { key: "dmEnabled" as const, label: "Direct Messages", desc: "Always receive notifications for private direct messages." },
                      ].map(({ key, label, desc }, i, arr) => (
                        <div key={key}>
                          <div className="flex items-center justify-between">
                            <div>
                              <p className="text-sm font-bold">{label}</p>
                              <p className="text-[11px] text-muted-foreground">{desc}</p>
                            </div>
                            <Switch
                              checked={preferences[key]}
                              onCheckedChange={(val) => updatePreferences({ [key]: val })}
                            />
                          </div>
                          {i < arr.length - 1 && <Separator className="mt-5" />}
                        </div>
                      ))}
                      <Separator />
                      <div className="flex items-center justify-between pt-1">
                        <div>
                          <p className="text-sm font-black text-red-600">Mute All Notifications</p>
                          <p className="text-[11px] text-muted-foreground italic">Temporarily silence all alerts.</p>
                        </div>
                        <Switch
                          checked={preferences.muteAll}
                          onCheckedChange={(val) => updatePreferences({ muteAll: val })}
                          className="data-[state=checked]:bg-red-500"
                        />
                      </div>
                    </CardContent>
                  </Card>
                )}
              </div>
            )}

            {/* ── Privacy Tab ── */}
            {activeTab === "privacy" && (
              <div className="space-y-6">
                <div>
                  <h2 className="text-lg font-black">Privacy & Safety</h2>
                  <p className="text-sm text-muted-foreground mt-0.5">Manage your data and privacy preferences.</p>
                </div>
                <div className="grid grid-cols-1 gap-3">
                  {[
                    { label: "Blocking", desc: "Manage the people you've blocked." },
                    { label: "App Permissions", desc: "Control which apps can access your data." },
                    { label: "Data Export", desc: "Download a copy of your Relay data." },
                    { label: "Delete Account", desc: "Permanently remove your account and data.", danger: true },
                  ].map(item => (
                    <Button
                      key={item.label}
                      variant="outline"
                      className={cn(
                        "h-auto p-4 justify-start text-left flex-col gap-1 items-start border-border/50",
                        item.danger && "border-red-500/20 hover:border-red-500/40 hover:bg-red-500/5"
                      )}
                    >
                      <span className={cn("text-sm font-bold", item.danger && "text-red-600")}>{item.label}</span>
                      <span className="text-[10px] text-muted-foreground font-normal">{item.desc}</span>
                    </Button>
                  ))}
                </div>
              </div>
            )}

          </div>
        </ScrollArea>
      </div>
    </div>
  )
}
