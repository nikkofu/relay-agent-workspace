"use client"

import { Hash, Lock } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { useChannelStore } from "@/stores/channel-store"
import { useState, useEffect } from "react"
import { toast } from "sonner"

export function CreateChannelDialog({ 
  open, 
  onOpenChange 
}: { 
  open: boolean, 
  onOpenChange: (open: boolean) => void 
}) {
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [isPrivate, setIsPrivate] = useState(false)
  const [loading, setLoading] = useState(false)
  const [mounted, setMounted] = useState(false)
  const { createChannel } = useChannelStore()

  useEffect(() => {
    setMounted(true)
  }, [])


  const handleCreate = async () => {
    if (!name) return
    setLoading(true)
    try {
      await createChannel(name, description, isPrivate)
      toast.success(`Channel #${name} created`)
      onOpenChange(false)
      setName("")
      setDescription("")
      setIsPrivate(false)
    } catch {
      toast.error("Failed to create channel")
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px] bg-white dark:bg-[#1a1d21]">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold">
            {mounted ? "Create a channel" : "Create"}
          </DialogTitle>
          <DialogDescription>
            Channels are where your team communicates. They’re best when organized around a topic — #marketing, for example.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="name" className="font-bold">Name</Label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                {isPrivate ? <Lock className="w-3.5 h-3.5" /> : <Hash className="w-3.5 h-3.5" />}
              </span>
              <Input
                id="name"
                placeholder="e.g. plan-budget"
                className="pl-9"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="description" className="font-bold">Description <span className="text-xs font-normal text-muted-foreground">(optional)</span></Label>
            <Textarea
              id="description"
              placeholder="What’s this channel about?"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="flex items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <Label className="font-bold">Make private</Label>
              <p className="text-xs text-muted-foreground">
                When a channel is set to private, it can only be viewed or joined by invitation.
              </p>
            </div>
            <Switch
              checked={isPrivate}
              onCheckedChange={setIsPrivate}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button 
            className="bg-[#007a5a] hover:bg-[#007a5a]/90 text-white" 
            onClick={handleCreate}
            disabled={!name || loading}
          >
            {loading ? "Creating..." : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
