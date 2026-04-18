import { useEffect, useRef } from 'react'
import { API_BASE_URL } from '@/lib/constants'
import { useMessageStore } from '@/stores/message-store'
import { useCollabStore } from '@/stores/collab-store'
import { usePresenceStore } from '@/stores/presence-store'

export function useWebsocket() {
  const socketRef = useRef<WebSocket | null>(null)
  const { addMessage } = useMessageStore()
  const { setCollabData } = useCollabStore()
  const { updatePresence, setTyping } = usePresenceStore()

  useEffect(() => {
    // Convert http://... to ws://...
    const wsUrl = API_BASE_URL.replace(/^http/, 'ws') + '/realtime'
    
    console.log("Connecting to WebSocket:", wsUrl)
    const socket = new WebSocket(wsUrl)
    socketRef.current = socket

    socket.onopen = () => {
      console.log("WebSocket connected ✅")
    }

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        console.log("WS Event received:", data.type)

        if (data.type === 'message.created') {
          addMessage({
            id: data.payload.id,
            content: data.payload.content,
            senderId: data.payload.user_id,
            channelId: data.payload.channel_id,
            dmId: data.payload.dm_id,
            createdAt: data.payload.created_at,
            reactions: [],
            attachments: [],
            replyCount: data.payload.reply_count,
            lastReplyAt: data.payload.last_reply_at
          })
        } else if (data.type === 'message.deleted') {
          const { message_id } = data.payload
          useMessageStore.getState().deleteMessageLocally(message_id)
        } else if (data.type === 'message.updated' || data.type === 'reaction.updated') {
          const updatedMsg = data.payload.message || data.payload
          useMessageStore.getState().updateMessageLocally(updatedMsg)
        } else if (data.type === 'agent_collab.sync') {
          console.log("Syncing Agent Collab data from backend...")
          setCollabData(data.payload)
        } else if (data.type === 'presence.updated') {
          console.log("Presence updated for user:", data.payload.user?.id)
          updatePresence(data.payload.user)
        } else if (data.type === 'typing.updated') {
          setTyping({
            userId: data.payload.user_id,
            isTyping: data.payload.is_typing,
            channelId: data.payload.channel_id,
            dmId: data.payload.dm_id,
            threadId: data.payload.thread_id
          })
        }
      } catch (err) {
        console.error("Failed to parse WS message:", err)
      }
    }

    socket.onerror = (error) => {
      console.error("WebSocket error ❌:", error)
    }

    socket.onclose = () => {
      console.log("WebSocket disconnected 🔌")
    }

    return () => {
      socket.close()
    }
  }, [addMessage])
}
