import { useEffect, useRef } from 'react'
import { API_BASE_URL } from '@/lib/constants'
import { useMessageStore } from '@/stores/message-store'

export function useWebsocket() {
  const socketRef = useRef<WebSocket | null>(null)
  const { addMessage } = useMessageStore()

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
          // Handle incoming message from WS
          // Map backend payload to frontend Message type
          addMessage({
            id: data.payload.id,
            content: data.payload.content,
            senderId: data.payload.user_id,
            channelId: data.payload.channel_id,
            createdAt: data.payload.created_at,
            reactions: [],
            attachments: []
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
