"use client"

import type { ReactNode } from "react"
import { useRouter } from "next/navigation"
import {
  AlertTriangle,
  ArrowRight,
  AtSign,
  Bot,
  Briefcase,
  CheckSquare,
  Hash,
  Inbox,
  Layers3,
  Sparkles,
} from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { Badge } from "@/components/ui/badge"
import { HomeAppsToolsSection } from "@/components/layout/home-apps-tools-section"
import type {
  HomeActivityItem,
  HomeAISuggestionItem,
  HomeData,
  HomeWorkbenchListTaskItem,
  HomeMyWorkItem,
  HomeRecentActivityItem,
  HomeTodayItem,
  WorkspaceView,
} from "@/types"

interface HomeWorkbenchProps {
  homeData: HomeData | null
  workspaceViews: WorkspaceView[]
  isLoadingWorkspaceViews: boolean
  workspaceViewsError: string | null
}

function relative(value?: string | null) {
  if (!value) return null
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return null
  return formatDistanceToNow(date, { addSuffix: true })
}

function stripHtml(html: string): string {
  if (!html) return ""
  return html
    .replace(/<[^>]*>/g, "")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#039;/g, "'")
    .replace(/&nbsp;/g, " ")
    .trim()
}

function getTodayItemTone(item: HomeTodayItem) {
  if (item.type === "mention") return "violet"
  if (item.type === "ai_failure") return "rose"
  if (item.type === "list_item_due") return "amber"
  return "slate"
}

function toneClasses(tone: "violet" | "rose" | "amber" | "slate") {
  if (tone === "violet") return "bg-violet-500/10 text-violet-700 dark:text-violet-300"
  if (tone === "rose") return "bg-rose-500/10 text-rose-700 dark:text-rose-300"
  if (tone === "amber") return "bg-amber-500/10 text-amber-700 dark:text-amber-300"
  return "bg-muted text-muted-foreground"
}

function isListTaskItem(item: HomeTodayItem | HomeMyWorkItem): item is HomeWorkbenchListTaskItem {
  return (item.type === "list_item_due" || item.type === "list_item_assigned") && "item" in item
}

function SectionFrame({
  title,
  eyebrow,
  icon,
  children,
}: {
  title: string
  eyebrow: string
  icon: ReactNode
  children: ReactNode
}) {
  return (
    <div className="rounded-3xl border bg-white dark:bg-[#222529] shadow-sm">
      <div className="px-5 py-4 border-b flex items-center justify-between gap-3">
        <div>
          <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">{eyebrow}</p>
          <h3 className="mt-1 text-lg font-black tracking-tight flex items-center gap-2">
            {icon}
            {title}
          </h3>
        </div>
      </div>
      <div className="p-5">{children}</div>
    </div>
  )
}

function PrimaryWorkbenchSection({
  title,
  eyebrow,
  icon,
  items,
  emptyTitle,
  emptyBody,
  onOpen,
}: {
  title: string
  eyebrow: string
  icon: ReactNode
  items: Array<HomeTodayItem | HomeMyWorkItem>
  emptyTitle: string
  emptyBody: string
  onOpen: (item: HomeTodayItem | HomeMyWorkItem) => void
}) {
  return (
    <SectionFrame title={title} eyebrow={eyebrow} icon={icon}>
      <div className="space-y-3">
        {items.map((item) => {
          const tone = getTodayItemTone(item as HomeTodayItem)
          const when = relative((item as { occurred_at?: string | null }).occurred_at)
          const listItem = isListTaskItem(item) ? item.item : null
          return (
            <button
              key={item.id}
              type="button"
              onClick={() => onOpen(item)}
              className="w-full rounded-2xl border px-4 py-4 text-left transition-colors hover:border-violet-300 hover:bg-violet-500/5"
            >
              <div className="flex items-start gap-3">
                <div className={`mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${toneClasses(tone)}`}>
                  {item.type === "mention" && <AtSign className="w-4 h-4" />}
                  {item.type === "ai_failure" && <AlertTriangle className="w-4 h-4" />}
                  {(item.type === "list_item_due" || item.type === "list_item_assigned") && <CheckSquare className="w-4 h-4" />}
                  {item.type !== "mention" && item.type !== "ai_failure" && item.type !== "list_item_due" && item.type !== "list_item_assigned" && <Layers3 className="w-4 h-4" />}
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="text-sm font-bold leading-5">{item.summary || "Untitled work item"}</p>
                      {listItem?.content && listItem.content !== item.summary && (
                        <p className="text-sm text-muted-foreground mt-1 line-clamp-2">{listItem.content}</p>
                      )}
                    </div>
                    <ArrowRight className="w-4 h-4 text-muted-foreground shrink-0" />
                  </div>
                  <div className="mt-3 flex flex-wrap items-center gap-2">
                    <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
                      {item.type.replace(/_/g, " ")}
                    </Badge>
                    {when && (
                      <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-bold">
                        {when}
                      </span>
                    )}
                    {listItem?.due_at && (
                      <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-bold">
                        due {relative(listItem.due_at) || "soon"}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </button>
          )
        })}

        {items.length === 0 && (
          <div className="rounded-2xl border border-dashed px-5 py-8 text-center">
            <p className="text-sm font-bold">{emptyTitle}</p>
            <p className="text-xs text-muted-foreground mt-2">{emptyBody}</p>
          </div>
        )}
      </div>
    </SectionFrame>
  )
}

function CompactListSection({
  title,
  eyebrow,
  icon,
  items,
  emptyBody,
  onOpen,
}: {
  title: string
  eyebrow: string
  icon: ReactNode
  items: HomeRecentActivityItem[]
  emptyBody: string
  onOpen: (item: HomeRecentActivityItem) => void
}) {
  return (
    <SectionFrame title={title} eyebrow={eyebrow} icon={icon}>
      <div className="space-y-2">
        {items.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => onOpen(item)}
            className="w-full rounded-2xl border px-4 py-3 text-left transition-colors hover:border-violet-300 hover:bg-violet-500/5"
          >
            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <p className="text-sm font-bold truncate">#{item.channel_name || "channel"}</p>
                <p className="text-xs text-muted-foreground mt-1 line-clamp-1">{stripHtml(item.last_message || "") || "Recent channel activity"}</p>
              </div>
              <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-bold shrink-0">
                {relative(item.occurred_at) || "recent"}
              </span>
            </div>
          </button>
        ))}

        {items.length === 0 && (
          <div className="rounded-2xl border border-dashed px-4 py-6 text-xs text-muted-foreground">
            {emptyBody}
          </div>
        )}
      </div>
    </SectionFrame>
  )
}

function AISuggestionsSection({ items }: { items: HomeAISuggestionItem[] }) {
  const router = useRouter()

  return (
    <SectionFrame title="AI Suggestions" eyebrow="Grounded prompts" icon={<Bot className="w-5 h-5 text-fuchsia-600" />}>
      <div className="space-y-2">
        {items.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => router.push("/workspace/activity")}
            className="w-full rounded-2xl border px-4 py-3 text-left transition-colors hover:border-fuchsia-300 hover:bg-fuchsia-500/5"
          >
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="text-sm font-bold">{item.title || item.summary}</p>
                {item.title && <p className="text-xs text-muted-foreground mt-1 line-clamp-2">{item.summary}</p>}
              </div>
              <ArrowRight className="w-4 h-4 text-muted-foreground shrink-0" />
            </div>
            {item.source?.type && (
              <div className="mt-2 flex items-center gap-2">
                <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
                  {item.source.type.replace(/_/g, " ")}
                </Badge>
                {item.source.id && (
                  <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-bold truncate">
                    {item.source.id}
                  </span>
                )}
              </div>
            )}
          </button>
        ))}

        {items.length === 0 && (
          <div className="rounded-2xl border border-dashed px-4 py-6 text-xs text-muted-foreground">
            No grounded AI suggestions right now.
          </div>
        )}
      </div>
    </SectionFrame>
  )
}

function ActivitySection({ items }: { items: HomeActivityItem[] }) {
  const router = useRouter()

  const openItem = (item: HomeActivityItem) => {
    const channelId = item.channel?.id
    const messageId = item.message?.id
    const dmId = item.message?.dm_conversation_id
    if (channelId && messageId) {
      router.push(`/workspace?c=${channelId}&m=${messageId}`)
      return
    }
    if (channelId) {
      router.push(`/workspace?c=${channelId}`)
      return
    }
    if (dmId) {
      router.push(`/workspace/dms/${dmId}`)
      return
    }
    router.push("/workspace/activity")
  }

  return (
    <SectionFrame title="Activity" eyebrow="Recent summary" icon={<Inbox className="w-5 h-5 text-sky-600" />}>
      <div className="space-y-2">
        {items.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => openItem(item)}
            className="w-full rounded-2xl border px-4 py-3 text-left transition-colors hover:border-sky-300 hover:bg-sky-500/5"
          >
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="text-sm font-bold line-clamp-2">{item.summary}</p>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[10px] text-muted-foreground uppercase tracking-widest font-bold">
                  <span>{item.type.replace(/_/g, " ")}</span>
                  {item.target && <span>{item.target}</span>}
                </div>
              </div>
              <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-bold shrink-0">
                {relative(item.occurred_at) || "recent"}
              </span>
            </div>
          </button>
        ))}

        {items.length === 0 && (
          <div className="rounded-2xl border border-dashed px-4 py-6 text-xs text-muted-foreground">
            No recent activity yet.
          </div>
        )}
      </div>
    </SectionFrame>
  )
}

export function HomeWorkbench({
  homeData,
  workspaceViews,
  isLoadingWorkspaceViews,
  workspaceViewsError,
}: HomeWorkbenchProps) {
  const router = useRouter()
  const today = homeData?.today?.items || []
  const myWork = homeData?.my_work?.items || []
  const recentChannels = homeData?.recent_channels?.items || []
  const aiSuggestions = (homeData?.ai_suggestions?.items || []).filter(item => item.source?.type || item.source?.id)
  const activityItems = homeData?.activity?.items || []
  const appsTools = homeData?.apps_tools?.items || []

  const openPrimaryItem = (item: HomeTodayItem | HomeMyWorkItem) => {
    if (item.type === "mention" && "channel_id" in item && item.channel_id && "message_id" in item && item.message_id) {
      router.push(`/workspace?c=${item.channel_id}&m=${item.message_id}`)
      return
    }
    if (isListTaskItem(item) && item.item.source_channel_id) {
      router.push(`/workspace?c=${item.item.source_channel_id}`)
      return
    }
    if (item.type === "ai_failure") {
      router.push("/workspace/activity")
      return
    }
    router.push("/workspace/activity")
  }

  return (
    <div className="grid grid-cols-1 xl:grid-cols-12 gap-8 pt-4">
      <div className="xl:col-span-7 space-y-8">
        <PrimaryWorkbenchSection
          title="Today"
          eyebrow="Primary workbench"
          icon={<Sparkles className="w-5 h-5 text-violet-600" />}
          items={today}
          emptyTitle="You are clear for today"
          emptyBody="Urgent mentions, failed AI runs, and due-today work will land here first."
          onOpen={openPrimaryItem}
        />
        <PrimaryWorkbenchSection
          title="My Work"
          eyebrow="Owned by you"
          icon={<Briefcase className="w-5 h-5 text-emerald-600" />}
          items={myWork}
          emptyTitle="No personal backlog right now"
          emptyBody="Assigned work that is not already urgent will show up here."
          onOpen={openPrimaryItem}
        />
      </div>

      <div className="xl:col-span-5 space-y-6">
        <CompactListSection
          title="Recent Channels"
          eyebrow="Channel-scoped"
          icon={<Hash className="w-5 h-5 text-violet-600" />}
          items={recentChannels}
          emptyBody="No recent channels to surface right now."
          onOpen={(item) => router.push(`/workspace?c=${item.channel_id}`)}
        />
        <AISuggestionsSection items={aiSuggestions} />
        <HomeAppsToolsSection
          items={appsTools}
          workspaceViews={workspaceViews}
          isLoadingWorkspaceViews={isLoadingWorkspaceViews}
          workspaceViewsError={workspaceViewsError}
        />
        <ActivitySection items={activityItems} />
      </div>
    </div>
  )
}
