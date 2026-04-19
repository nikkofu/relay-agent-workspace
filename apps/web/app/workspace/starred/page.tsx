"use client"

import { useChannelStore } from "@/stores/channel-store"
import { useMessageStore } from "@/stores/message-store"
import { Star, Hash, Lock, Pin, ChevronRight } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useEffect } from "react"
import { formatDistanceToNow } from "date-fns"
import { useRouter } from "next/navigation"
import { UserAvatar } from "@/components/common/user-avatar"

export default function StarredPage() {
  const { channels, fetchStarredChannels, setCurrentChannelById } = useChannelStore()
  const { pinnedMessages, fetchPins } = useMessageStore()
  const router = useRouter()

  useEffect(() => {
    fetchStarredChannels()
    fetchPins() // Fetch global pins
  }, [fetchStarredChannels, fetchPins])

  const starredChannels = channels.filter(c => c.isStarred)

  const handleChannelClick = (channelId: string) => {
    setCurrentChannelById(channelId)
    router.push(`/workspace?c=${channelId}`)
  }

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-white dark:bg-[#1a1d21] z-10">
        <div className="flex items-center gap-2">
          <Star className="w-5 h-5 text-yellow-500 fill-yellow-500" />
          <h2 className="font-bold text-lg text-foreground">Starred & Pinned</h2>
        </div>
      </header>

      <ScrollArea className="flex-1">
        <div className="max-w-4xl mx-auto p-6 space-y-10 pb-20">
          {/* Starred Channels Section */}
          <section className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-border/50">
              <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground">Starred Channels</h3>
              <span className="text-xs bg-muted px-2 py-0.5 rounded-full">{starredChannels.length}</span>
            </div>
            
            {starredChannels.length === 0 ? (
              <p className="text-sm text-muted-foreground italic py-4 text-center">No starred channels yet.</p>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {starredChannels.map((channel) => (
                  <button
                    key={channel.id}
                    onClick={() => handleChannelClick(channel.id)}
                    className="flex items-center justify-between p-4 rounded-xl border border-border/50 hover:border-yellow-500/50 hover:bg-yellow-500/5 transition-all text-left group"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-lg bg-muted flex items-center justify-center">
                        {channel.type === 'private' ? <Lock className="w-4 h-4" /> : <Hash className="w-4 h-4" />}
                      </div>
                      <div>
                        <p className="text-sm font-bold text-foreground group-hover:text-yellow-600 dark:group-hover:text-yellow-400">
                          {channel.name}
                        </p>
                        <p className="text-xs text-muted-foreground truncate max-w-[180px]">
                          {channel.description || "No description"}
                        </p>
                      </div>
                    </div>
                    <ChevronRight className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                  </button>
                ))}
              </div>
            )}
          </section>

          {/* Pinned Messages Section */}
          <section className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-border/50">
              <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground">Global Pinned Messages</h3>
              <span className="text-xs bg-muted px-2 py-0.5 rounded-full">{pinnedMessages.length}</span>
            </div>

            {pinnedMessages.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <Pin className="w-8 h-8 mb-3 opacity-20 rotate-45" />
                <p className="text-sm italic">No pinned messages found.</p>
              </div>
            ) : (
              <div className="flex flex-col gap-4">
                {pinnedMessages.map(({ message, user, channel }) => (
                  <div 
                    key={message.id} 
                    className="p-4 rounded-xl border border-border/50 hover:bg-muted/30 transition-all group"
                  >
                    <div className="flex items-start gap-4">
                      <UserAvatar src={user?.avatar} name={user?.name || "Unknown"} className="h-9 w-9" />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between mb-1">
                          <div className="flex items-center gap-2">
                            <span className="text-sm font-bold">{user?.name || "Someone"}</span>
                            <span className="text-[10px] text-muted-foreground uppercase font-medium">
                              {formatDistanceToNow(new Date(message.createdAt), { addSuffix: true })}
                            </span>
                          </div>
                          <button 
                            onClick={() => handleChannelClick(channel?.id)}
                            className="text-[10px] font-bold text-purple-600 dark:text-purple-400 hover:underline flex items-center gap-1"
                          >
                            <Hash className="w-2.5 h-2.5" />
                            {channel?.name || "Unknown"}
                          </button>
                        </div>
                        <div 
                          className="text-sm text-foreground/80 break-words leading-relaxed" 
                          dangerouslySetInnerHTML={{ __html: message.content }} 
                        />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>
      </ScrollArea>
    </div>
  )
}
