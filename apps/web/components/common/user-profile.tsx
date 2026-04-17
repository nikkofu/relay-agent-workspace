"use client"

import { User } from "@/types"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Sparkles, MessageCircle, Mail, Clock, Calendar } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { useRouter } from "next/navigation"

interface UserProfileProps {
  user: User
}

export function UserProfile({ user }: UserProfileProps) {
  const router = useRouter()
  
  const statusColors: Record<string, string> = {
    online: "bg-green-500",
    away: "bg-yellow-500",
    busy: "bg-red-500",
    offline: "bg-slate-500"
  }

  const handleMessageClick = () => {
    router.push(`/workspace/dms?u=${user.id}`)
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
        <div>
          <div className="flex items-center gap-2">
            <h3 className="text-xl font-black tracking-tight">{user.name}</h3>
            <div className={`h-2.5 w-2.5 rounded-full ${statusColors[user.status] || statusColors.offline}`} />
          </div>
          <p className="text-sm text-muted-foreground font-medium">{user.statusText || (user.status === 'online' ? "Active" : "Away")}</p>
        </div>

        {/* AI Insight Section */}
        <div className="bg-purple-50 dark:bg-purple-900/20 p-3 rounded-lg border border-purple-100 dark:border-purple-800/30">
          <div className="flex items-center gap-2 mb-1.5 text-purple-600 dark:text-purple-400">
            <Sparkles className="w-3.5 h-3.5" />
            <span className="text-[10px] font-bold uppercase tracking-wider">AI Collaboration Insight</span>
          </div>
          <p className="text-[11px] leading-relaxed text-foreground/80 italic font-medium">
            &quot;{user.aiInsight || `${user.name} is present in the workspace. Insight data is still loading.`}&quot;
          </p>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-2">
          <Button 
            className="flex-1 h-9 font-bold bg-[#007a5a] hover:bg-[#007a5a]/90 text-white"
            onClick={handleMessageClick}
          >
            <MessageCircle className="w-4 h-4 mr-2" /> Message
          </Button>
          <Button variant="outline" size="icon" className="h-9 w-9">
            <Mail className="w-4 h-4" />
          </Button>
        </div>

        <Separator className="opacity-50" />

        {/* Details Grid */}
        <div className="grid grid-cols-1 gap-3">
          <DetailItem icon={Clock} label="Local time" value="10:45 AM" />
          <DetailItem icon={Calendar} label="Working days" value="Mon - Fri" />
        </div>
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
