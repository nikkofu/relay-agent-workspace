import { useEffect, useRef } from 'react'
import { toast } from 'sonner'
import { API_BASE_URL } from '@/lib/constants'
import { useMessageStore } from '@/stores/message-store'
import { useCollabStore } from '@/stores/collab-store'
import { usePresenceStore } from '@/stores/presence-store'
import { useArtifactStore } from '@/stores/artifact-store'
import { useActivityStore } from '@/stores/activity-store'
import { useDirectoryStore } from '@/stores/directory-store'
import { useFileStore } from '@/stores/file-store'
import { useKnowledgeStore } from '@/stores/knowledge-store'
import { useChannelStore } from '@/stores/channel-store'
import { useWorkspaceStore } from '@/stores/workspace-store'
import { useUserStore } from '@/stores/user-store'

export function useWebsocket() {
  const socketRef = useRef<WebSocket | null>(null)
  const { addMessage } = useMessageStore()
  const { setCollabData } = useCollabStore()
  const { updatePresence, setTyping } = usePresenceStore()
  const { updateArtifactLocally } = useArtifactStore()
  const { markAsReadLocally } = useActivityStore()
  const { fetchWorkflowRuns } = useDirectoryStore()

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
        } else if (data.type === 'artifact.updated') {
          updateArtifactLocally(data.payload)
        } else if (data.type === 'notifications.read') {
          markAsReadLocally(data.payload.item_ids || [])
        } else if (data.type === 'workflow.run.updated') {
          fetchWorkflowRuns()
        } else if (data.type === 'file.extraction.updated') {
          const { file_id, status, is_searchable, is_citable, content_summary, last_indexed_at, needs_ocr } = data.payload || {}
          if (file_id) {
            useFileStore.getState().updateFileLocally(file_id, {
              extraction_status: status,
              is_searchable,
              is_citable,
              content_summary,
              last_indexed_at,
              needs_ocr,
            })
          }
        } else if (data.type === 'knowledge.entity.created') {
          const entity = data.payload?.entity || data.payload
          if (entity?.id) useKnowledgeStore.getState().handleEntityCreated(entity)
        } else if (data.type === 'knowledge.entity.updated') {
          const entity = data.payload?.entity || data.payload
          if (entity?.id) useKnowledgeStore.getState().handleEntityUpdated(entity)
        } else if (data.type === 'knowledge.entity.ref.created') {
          const ref = data.payload?.ref || data.payload
          if (ref?.entity_id) {
            useKnowledgeStore.getState().pushLiveUpdate({
              type: 'ref.created',
              entityId: ref.entity_id,
              payload: ref,
              ts: Date.now(),
            })
          }
          const activeChannelId = useChannelStore.getState().currentChannel?.id
          if (activeChannelId) {
            useKnowledgeStore.getState().fetchChannelKnowledge(activeChannelId)
            useKnowledgeStore.getState().fetchChannelKnowledgeSummary(activeChannelId)
            const entityTitle = ref?.entity_title
            if (entityTitle) {
              toast(
                `📋 ${entityTitle} auto-linked`,
                {
                  description: ref?.source_snippet ? `"${(ref.source_snippet as string).slice(0, 80)}…"` : 'A message or file mentioned this entity.',
                  duration: 4000,
                  action: { label: 'View', onClick: () => window.location.href = `/workspace/knowledge/${ref.entity_id}` },
                }
              )
            }
          }
        } else if (data.type === 'knowledge.event.created') {
          const event = data.payload?.event || data.payload
          if (event?.entity_id) {
            useKnowledgeStore.getState().pushLiveUpdate({
              type: 'event.created',
              entityId: event.entity_id,
              payload: event,
              ts: Date.now(),
            })
          }
        } else if (data.type === 'knowledge.link.created') {
          const link = data.payload?.link || data.payload
          if (link?.from_entity_id) {
            useKnowledgeStore.getState().pushLiveUpdate({
              type: 'link.created',
              entityId: link.from_entity_id,
              payload: link,
              ts: Date.now(),
            })
          }
        } else if (data.type === 'knowledge.entity.activity.spiked') {
          const payload = data.payload || {}
          const currentUserId = useUserStore.getState().currentUser?.id
          const userIds: string[] = payload.user_ids || []
          if (currentUserId && userIds.includes(currentUserId)) {
            const entity = payload.entity
            const entityId: string = entity?.id || payload.entity_id
            const entityTitle: string = entity?.title || payload.entity_title || 'An entity'
            const delta: number = payload.delta ?? 0
            if (entityId) useKnowledgeStore.getState().markEntitySpiking(entityId)
            toast(
              `⚡ ${entityTitle} is spiking`,
              {
                description: delta > 0 ? `+${delta} new references this period` : 'Activity spike detected',
                duration: 8000,
                action: entityId ? {
                  label: 'View',
                  onClick: () => window.location.href = `/workspace/knowledge/${entityId}`,
                } : undefined,
              }
            )
          }
        } else if (data.type === 'knowledge.digest.published') {
          const payload = data.payload || {}
          const channelId = payload.channel_id || payload.channel?.id
          // Update the live-update bus so the inbox page reacts
          useKnowledgeStore.getState().applyDigestPublished({
            channel_id: channelId,
            message: payload.message,
            digest: payload.digest,
          })
          // Refresh the cross-channel inbox
          const scope = useKnowledgeStore.getState().knowledgeInboxScope
          useKnowledgeStore.getState().fetchKnowledgeInbox(scope, 50).catch(() => {})
          // If current channel matches, re-fetch its digest & summary
          const activeChannelId = useChannelStore.getState().currentChannel?.id
          if (activeChannelId && activeChannelId === channelId) {
            useKnowledgeStore.getState().fetchChannelKnowledgeSummary(activeChannelId).catch(() => {})
          }
          // Refresh home digest aggregation so Home stays in sync
          useWorkspaceStore.getState().fetchHome?.().catch?.(() => {})
          // Subtle toast with jump
          const channelName = payload?.channel?.name
          const messageId = payload?.message?.id
          toast(
            `📰 New ${payload?.digest?.window || 'weekly'} digest${channelName ? ` in #${channelName}` : ''}`,
            {
              description: payload?.digest?.headline || 'Auto-published knowledge digest',
              duration: 5000,
              action: messageId && channelId ? {
                label: 'View',
                onClick: () => window.location.href = `/workspace?c=${channelId}&m=${messageId}`,
              } : undefined,
            }
          )
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
