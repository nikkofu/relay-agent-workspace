"use client"

import { useEffect } from "react"
import { useNotificationStore } from "@/stores/notification-store"
import { Bell, Shield, Settings } from "lucide-react"
import { Switch } from "@/components/ui/switch"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"

export default function SettingsPage() {
  const { preferences, isLoading, fetchPreferences, updatePreferences } = useNotificationStore()

  useEffect(() => {
    fetchPreferences()
  }, [fetchPreferences])

  if (isLoading || !preferences) {
    return <div className="flex-1 flex items-center justify-center">Loading preferences...</div>
  }

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      <header className="h-14 px-6 flex items-center border-b shrink-0">
        <Settings className="w-5 h-5 mr-2 text-slate-600" />
        <h1 className="text-lg font-black tracking-tight uppercase">Workspace Settings</h1>
      </header>

      <ScrollArea className="flex-1">
        <div className="p-10 max-w-3xl mx-auto space-y-12">
          {/* Notifications Section */}
          <section className="space-y-6">
            <div className="flex items-center gap-2">
              <Bell className="w-5 h-5 text-purple-600" />
              <h2 className="text-xl font-black tracking-tight">Notifications</h2>
            </div>
            
            <Card className="shadow-none border-border/50">
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-bold">Preferences</CardTitle>
                <CardDescription className="text-xs">Control how and when you receive notifications.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <p className="text-sm font-bold">Inbox Notifications</p>
                    <p className="text-[11px] text-muted-foreground">Receive alerts for new messages in your inbox.</p>
                  </div>
                  <Switch 
                    checked={preferences.inboxEnabled}
                    onCheckedChange={(val) => updatePreferences({ inboxEnabled: val })}
                  />
                </div>
                
                <Separator />

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <p className="text-sm font-bold">Mentions & Replies</p>
                    <p className="text-[11px] text-muted-foreground">Get notified when someone @mentions you or replies to your threads.</p>
                  </div>
                  <Switch 
                    checked={preferences.mentionsEnabled}
                    onCheckedChange={(val) => updatePreferences({ mentionsEnabled: val })}
                  />
                </div>

                <Separator />

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <p className="text-sm font-bold">Direct Messages</p>
                    <p className="text-[11px] text-muted-foreground">Always receive notifications for private direct messages.</p>
                  </div>
                  <Switch 
                    checked={preferences.dmEnabled}
                    onCheckedChange={(val) => updatePreferences({ dmEnabled: val })}
                  />
                </div>

                <Separator />

                <div className="flex items-center justify-between pt-2">
                  <div className="space-y-0.5">
                    <p className="text-sm font-black text-red-600">Mute All Notifications</p>
                    <p className="text-[11px] text-muted-foreground italic">Temporarily disable all notification alerts.</p>
                  </div>
                  <Switch 
                    checked={preferences.muteAll}
                    onCheckedChange={(val) => updatePreferences({ muteAll: val })}
                    className="data-[state=checked]:bg-red-500"
                  />
                </div>
              </CardContent>
            </Card>
          </section>

          {/* Privacy & Safety */}
          <section className="space-y-6">
            <div className="flex items-center gap-2">
              <Shield className="w-5 h-5 text-blue-600" />
              <h2 className="text-xl font-black tracking-tight">Privacy & Safety</h2>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Button variant="outline" className="h-auto p-4 justify-start text-left flex-col gap-1 items-start border-border/50">
                <span className="text-sm font-bold">Blocking</span>
                <span className="text-[10px] text-muted-foreground font-normal">Manage the people you&apos;ve blocked.</span>
              </Button>
              <Button variant="outline" className="h-auto p-4 justify-start text-left flex-col gap-1 items-start border-border/50">
                <span className="text-sm font-bold">App Permissions</span>
                <span className="text-[10px] text-muted-foreground font-normal">Control which apps can access your data.</span>
              </Button>
            </div>
          </section>
        </div>
      </ScrollArea>
    </div>
  )
}
