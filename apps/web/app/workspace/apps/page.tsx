"use client"

import Link from "next/link"
import { useEffect } from "react"
import { AppWindow, ArrowRight, Loader2, Sparkles } from "lucide-react"
import { useShallow } from "zustand/react/shallow"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { getBusinessAppModeLabel, getBusinessAppRoute } from "@/lib/business-apps"
import { useBusinessAppStore } from "@/stores/business-app-store"

export default function AppsHubPage() {
  const { apps, isLoadingApps, appsError, fetchApps } = useBusinessAppStore(useShallow((state) => ({
    apps: state.apps,
    isLoadingApps: state.isLoadingApps,
    appsError: state.appsError,
    fetchApps: state.fetchApps,
  })))

  useEffect(() => {
    fetchApps().catch(() => { /* handled in store */ })
  }, [fetchApps])

  return (
    <div className="flex flex-1 flex-col overflow-auto bg-white dark:bg-[#1a1d21]">
      <div className="border-b bg-gradient-to-br from-violet-500/8 via-white to-white px-6 py-5 dark:from-violet-500/10 dark:via-[#1a1d21] dark:to-[#1a1d21]">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-violet-500/10 text-violet-600 dark:text-violet-300">
                <AppWindow className="h-5 w-5" />
              </div>
              <div>
                <h1 className="text-xl font-black tracking-tight">App Hub</h1>
                <p className="mt-1 text-sm text-muted-foreground">Reusable business surfaces for Relay&apos;s AI-native workspace. Phase 73 ships the first instance: Sales App.</p>
              </div>
            </div>
          </div>
          <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">Business Apps</Badge>
        </div>
      </div>

      <div className="p-6">
        {isLoadingApps && apps.length === 0 ? (
          <div className="flex min-h-[260px] items-center justify-center gap-2 rounded-3xl border border-dashed text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading apps...
          </div>
        ) : appsError ? (
          <div className="rounded-2xl border border-rose-500/20 bg-rose-500/5 px-4 py-5 text-sm text-rose-700 dark:text-rose-300">{appsError}</div>
        ) : apps.length === 0 ? (
          <div className="flex min-h-[260px] flex-col items-center justify-center gap-3 rounded-3xl border border-dashed bg-muted/10 text-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
              <Sparkles className="h-6 w-6" />
            </div>
            <div className="space-y-1">
              <p className="text-sm font-black tracking-tight">No published apps yet</p>
              <p className="max-w-md text-sm text-muted-foreground">Once Codex and Gemini freeze more business app contracts, new app surfaces will appear here.</p>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
            {apps.map((app) => (
              <div key={app.key} className="rounded-3xl border bg-background/70 p-5 shadow-sm">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h2 className="text-lg font-black tracking-tight">{app.title}</h2>
                    <p className="mt-2 text-sm text-muted-foreground">{app.description}</p>
                  </div>
                  <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">
                    {app.primary_entity ?? "sales_order"}
                  </Badge>
                </div>
                <div className="mt-4 flex flex-wrap gap-2">
                  {app.modes.map((mode) => (
                    <Badge key={`${app.key}-${mode}`} variant="outline" className="text-[10px] font-black uppercase tracking-widest">
                      {getBusinessAppModeLabel(mode)}
                    </Badge>
                  ))}
                </div>
                <div className="mt-5 flex items-center justify-between gap-2 rounded-2xl bg-muted/30 px-4 py-3 text-sm text-muted-foreground">
                  <span>AI-native actions remain metadata-only until Codex freezes a write contract.</span>
                  <Button asChild size="sm" className="gap-1.5">
                    <Link href={getBusinessAppRoute(app.key)}>
                      Open
                      <ArrowRight className="h-3.5 w-3.5" />
                    </Link>
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
