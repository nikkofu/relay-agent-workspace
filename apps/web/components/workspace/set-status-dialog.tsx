"use client"

import { useState } from "react"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select"
import { Loader2, Smile } from "lucide-react"
import { useUserStore } from "@/stores/user-store"
import { toast } from "sonner"

const STATUS_OPTIONS = [
  { value: "online",  label: "Active",     dot: "bg-green-500" },
  { value: "away",    label: "Away",       dot: "bg-amber-500" },
  { value: "busy",    label: "Do not disturb", dot: "bg-red-500" },
  { value: "offline", label: "Invisible",  dot: "bg-slate-400" },
]

interface SetStatusDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function SetStatusDialog({ open, onOpenChange }: SetStatusDialogProps) {
  const { currentUser, updateStatus } = useUserStore()
  const [status, setStatus] = useState<string>(currentUser?.status || "online")
  const [statusText, setStatusText] = useState(currentUser?.statusText || "")
  const [isSaving, setIsSaving] = useState(false)

  const handleSave = async () => {
    if (!currentUser) return
    setIsSaving(true)
    await updateStatus(currentUser.id, { status, statusText })
    setIsSaving(false)
    toast.success("Status updated")
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Smile className="w-4 h-4 text-purple-600" />
            Set a Status
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="space-y-1.5">
            <Label className="text-xs font-bold">Availability</Label>
            <Select value={status} onValueChange={setStatus}>
              <SelectTrigger className="h-9">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {STATUS_OPTIONS.map(opt => (
                  <SelectItem key={opt.value} value={opt.value}>
                    <div className="flex items-center gap-2">
                      <span className={`w-2.5 h-2.5 rounded-full ${opt.dot}`} />
                      {opt.label}
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label className="text-xs font-bold">Status Message</Label>
            <Input
              placeholder="What are you up to?"
              value={statusText}
              onChange={e => setStatusText(e.target.value)}
              className="text-sm"
              maxLength={100}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" size="sm" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button size="sm" onClick={handleSave} disabled={isSaving} className="bg-purple-600 hover:bg-purple-700 text-white">
            {isSaving ? <Loader2 className="w-3.5 h-3.5 animate-spin mr-1" /> : null}
            Save Status
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
