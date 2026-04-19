"use client"

import { useState, useEffect } from "react"
import { Mail, UserPlus } from "lucide-react"
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
import { useWorkspaceStore } from "@/stores/workspace-store"
import { toast } from "sonner"

export function InviteMemberDialog({ 
  open, 
  onOpenChange 
}: { 
  open: boolean, 
  onOpenChange: (open: boolean) => void 
}) {
  const [email, setEmail] = useState("")
  const [loading, setLoading] = useState(false)
  const [mounted, setMounted] = useState(false)
  const { inviteMember, currentWorkspace } = useWorkspaceStore()

  useEffect(() => {
    setMounted(true)
  }, [])

  const handleInvite = async () => {
    if (!email) return
    setLoading(true)
    try {
      await inviteMember(email)
      toast.success(`Invitation sent to ${email}`)
      onOpenChange(false)
      setEmail("")
    } catch {
      toast.error("Failed to send invitation")
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px] bg-white dark:bg-[#1a1d21]">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold flex items-center gap-2">
            <UserPlus className="w-5 h-5 text-purple-600" />
            Invite people to {mounted ? (currentWorkspace?.name || "Workspace") : "Workspace"}
          </DialogTitle>
          <DialogDescription>
            They will receive an email to join this workspace and start collaborating.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="email" className="font-bold">Email address</Label>
            <div className="relative">
              <Mail className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground w-4 h-4" />
              <Input
                id="email"
                type="email"
                placeholder="name@example.com"
                className="pl-9"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button 
            className="bg-[#3f0e40] hover:bg-[#3f0e40]/90 text-white" 
            onClick={handleInvite}
            disabled={!email || loading}
          >
            {loading ? "Sending..." : "Send invitation"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
