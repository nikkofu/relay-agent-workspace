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
import { useListStore, mapListItem } from '@/stores/list-store'
import { useToolStore, mapToolRun } from '@/stores/tool-store'

export function useWebsocket() {
  const socketRef = useRef<WebSocket | null>(null)
  const { addMessage, addStreamingChunk } = useMessageStore()
  const { setCollabData } = useCollabStore()
  const { updatePresence, setTyping } = usePresenceStore()
  const { updateArtifactLocally } = useArtifactStore()
  const { markAsReadLocally, appendUnifiedFeedItem, appendMentionItem } = useActivityStore()
  const { fetchWorkflowRuns } = useDirectoryStore()
  const { addItemLocally, updateItemLocally, removeItemLocally } = useListStore()
  const { addRunLocally, updateRunLocally } = useToolStore()

  useEffect(() => {
    // Convert http://... to ws://...
    const wsUrl = API_BASE_URL.replace(/^http/, 'ws') + '/realtime'
    
    console.log("Connecting to WebSocket:", wsUrl)
    const socket = new WebSocket(wsUrl)
    socketRef.current = socket

    socket.onopen = () => {
      console.log("WebSocket connected ✅")
      // Phase 61: bulk-hydrate presence on connect/reconnect
      const channelId = useChannelStore.getState().currentChannel?.id
      usePresenceStore.getState().bulkHydratePresence(channelId || undefined).catch(() => {})
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
          // Phase 64C: append to unified feed for live updates
          if (data.payload.channel_id && !data.payload.thread_id) {
            appendUnifiedFeedItem({
              id: `message:${data.payload.id}`,
              event_type: 'message',
              workspace_id: data.payload.workspace_id,
              actor_id: data.payload.user_id,
              channel_id: data.payload.channel_id,
              title: `Message in channel`,
              body: data.payload.content,
              link: data.payload.channel_id ? `/workspace?c=${data.payload.channel_id}` : undefined,
              occurred_at: data.payload.created_at ?? new Date().toISOString(),
            })
          } else if (data.payload.dm_id && !data.payload.thread_id) {
            appendUnifiedFeedItem({
              id: `dm_message:${data.payload.id}`,
              event_type: 'dm_message',
              actor_id: data.payload.user_id,
              dm_id: data.payload.dm_id,
              title: 'DM message',
              body: data.payload.content,
              occurred_at: data.payload.created_at ?? new Date().toISOString(),
            })
          }
        } else if (data.type === 'message.deleted') {
          const { message_id } = data.payload
          useMessageStore.getState().deleteMessageLocally(message_id)
        } else if (data.type === 'message.updated') {
          const updatedMsg = data.payload.message || data.payload
          useMessageStore.getState().updateMessageLocally(updatedMsg)
        } else if (data.type === 'reaction.updated') {
          const updatedMsg = data.payload.message || data.payload
          useMessageStore.getState().updateMessageLocally(updatedMsg)
          // Phase 64C: append reaction to unified feed
          const reaction = data.payload.reaction
          if (reaction) {
            appendUnifiedFeedItem({
              id: `reaction:${reaction.id ?? reaction.message_id + ':' + reaction.emoji + ':' + reaction.user_id}`,
              event_type: 'reaction',
              actor_id: reaction.user_id,
              channel_id: data.payload.channel_id,
              title: `Reacted ${reaction.emoji}`,
              body: data.payload.message?.content,
              occurred_at: reaction.created_at ?? new Date().toISOString(),
              meta: { message_id: reaction.message_id, emoji: reaction.emoji },
            })
          }
        } else if (data.type === 'mention.created') {
          // Phase 65A: live user mention — wire to mentions store + unified feed
          const p = data.payload || {}
          const currentUserId = useUserStore.getState().currentUser?.id
          const mentionId = `mention:${p.message_id}:${p.mentioned_user_id}`
          const channelName = p.channel_id ? `#${p.channel_id}` : 'DM'
          const actorName = p.mentioned_by_user_id
            ? (useUserStore.getState().users.find((u: any) => u.id === p.mentioned_by_user_id)?.name ?? 'Someone')
            : 'Someone'
          // Append to unified feed rail
          appendUnifiedFeedItem({
            id: mentionId,
            event_type: 'mention',
            workspace_id: p.workspace_id,
            actor_id: p.mentioned_by_user_id,
            actor_name: actorName,
            channel_id: p.channel_id || undefined,
            dm_id: p.dm_id || undefined,
            link: p.channel_id
              ? `/workspace?c=${p.channel_id}&m=${p.message_id}`
              : p.dm_id
              ? `/workspace/dms/${p.dm_id}`
              : undefined,
            title: `${actorName} mentioned you${p.channel_id ? ` in ${channelName}` : p.dm_id ? ' in a DM' : ''}`,
            body: p.mention_text,
            occurred_at: new Date().toISOString(),
            meta: {
              mention_kind: p.mention_kind ?? 'user',
              message_id: p.message_id,
              mentioned_user_id: p.mentioned_user_id,
              mentioned_by_user_id: p.mentioned_by_user_id,
              mention_text: p.mention_text,
            },
          })
          // Prepend to Mentions tab if this mention is for the current user
          if (p.mentioned_user_id && p.mentioned_user_id === currentUserId) {
            appendMentionItem({
              id: mentionId,
              type: 'mention',
              user: { id: p.mentioned_by_user_id, name: actorName },
              channel: p.channel_id ? { id: p.channel_id } : undefined,
              message: p.message_id ? { id: p.message_id, dm_id: p.dm_id } : undefined,
              target: p.channel_id ? channelName : 'DM',
              summary: `mentioned you${p.channel_id ? ` in ${channelName}` : ' in a DM'}`,
              occurredAt: new Date().toISOString(),
              isRead: false,
            })
            toast(`${actorName} mentioned you`, {
              description: p.mention_text ? `"${String(p.mention_text).slice(0, 80)}"` : undefined,
              duration: 5000,
              action: p.channel_id || p.dm_id ? {
                label: 'View',
                onClick: () => {
                  if (p.channel_id) window.location.href = `/workspace?c=${p.channel_id}`
                  else if (p.dm_id) window.location.href = `/workspace/dms/${p.dm_id}`
                }
              } : undefined,
            })
          }
        } else if (data.type === 'list.item.created') {
          // Phase 67B: handle live list item creation. The WS payload
          // arrives in raw snake_case from Go; pipe through `mapListItem`
          // so the store receives the camelCase shape it stores everywhere
          // else (`listId`, `isCompleted`, `createdAt`, …).
          const item = data.payload.item
          if (item) {
            addItemLocally(mapListItem(item))
            useWorkspaceStore.getState().markHomeExecutionStale?.()
          }
        } else if (data.type === 'list.item.updated') {
          // Phase 67B: handle live list item updates (snake_case → camelCase).
          const item = data.payload.item
          const listId = data.payload.list_id
          if (item && listId) {
            updateItemLocally(listId, mapListItem(item))
            useWorkspaceStore.getState().markHomeExecutionStale?.()
          }
        } else if (data.type === 'list.item.deleted') {
          // Phase 67B: handle live list item deletion
          const { list_id, item_id } = data.payload
          if (list_id && item_id) {
            removeItemLocally(list_id, item_id)
            useWorkspaceStore.getState().markHomeExecutionStale?.()
          }
        } else if (data.type === 'tool.run.started') {
          // Phase 67B: handle live tool run starts (snake_case → camelCase).
          const run = data.payload.run
          if (run) {
            addRunLocally(mapToolRun(run))
            useWorkspaceStore.getState().markHomeExecutionStale?.()
          }
        } else if (data.type === 'tool.run.completed' || data.type === 'tool.run.updated') {
          // Phase 67B: handle live tool run completion or progress (snake_case → camelCase).
          const run = data.payload.run
          if (run) {
            updateRunLocally(mapToolRun(run))
            useWorkspaceStore.getState().markHomeExecutionStale?.()
            // Phase 70B: notify ToolRunDetailPanel so it can stop polling immediately.
            if (data.type === 'tool.run.completed') {
              window.dispatchEvent(new CustomEvent('tool-run-completed', { detail: { runId: run.id } }))
            }
          }
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
          // Phase 64C: append artifact updates to unified feed
          const artifact = data.payload
          if (artifact?.id) {
            appendUnifiedFeedItem({
              id: `artifact_updated:${artifact.id}`,
              event_type: 'artifact_updated',
              workspace_id: artifact.workspace_id,
              actor_id: artifact.updated_by,
              channel_id: artifact.channel_id,
              title: `Artifact updated · ${artifact.title ?? 'Untitled'}`,
              body: [artifact.type, artifact.status].filter(Boolean).join(' · '),
              link: artifact.channel_id ? `/workspace?c=${artifact.channel_id}` : undefined,
              occurred_at: artifact.updated_at ?? new Date().toISOString(),
              meta: { artifact_id: artifact.id, artifact_type: artifact.type },
            })
          }
        } else if (data.type === 'dm.stream.chunk') {
          // Bug 1 fix: live AI DM response streaming
          const { temp_id, dm_id, chunk, is_final } = data.payload || {}
          if (temp_id && dm_id !== undefined) {
            addStreamingChunk(temp_id, dm_id, chunk ?? '', !!is_final)
          }
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
        } else if (data.type === 'knowledge.trending.changed') {
          // Phase 60: live trending rerank
          const payload = data.payload || {}
          if (payload.workspace_id && Array.isArray(payload.items)) {
            useKnowledgeStore.getState().applyTrendingChanged({
              workspace_id: payload.workspace_id,
              days: payload.days ?? 7,
              items: payload.items,
            })
          }
        } else if (data.type === 'knowledge.followed.stats.changed') {
          // Phase 61: push follow-stat deltas so the stats strip updates without polling
          const payload = data.payload || {}
          if (payload.stats) {
            useKnowledgeStore.getState().applyFollowedStatsChanged(payload.stats)
          }
        } else if (data.type === 'knowledge.entity.brief.generated') {
          // Phase 62: multi-tab sync of generated entity brief
          const payload = data.payload || {}
          if (payload.brief) {
            useKnowledgeStore.getState().applyEntityBriefGenerated(payload.brief)
          }
        } else if (data.type === 'knowledge.entity.brief.changed') {
          // Phase 63A: cached entity brief invalidation (show stale pulse)
          const payload = data.payload || {}
          const b = payload.brief
          if (b && b.entity_id) {
            useKnowledgeStore.getState().applyEntityBriefChanged({
              entity_id: b.entity_id,
              workspace_id: b.workspace_id,
              title: b.title,
              reason: b.reason,
              changed_at: b.changed_at || new Date().toISOString(),
              stale: true,
            })
          }
        } else if (data.type === 'notifications.bulk_read') {
          // Phase 62: multi-tab inbox read-state sync
          const payload = data.payload || {}
          if (Array.isArray(payload.item_ids)) {
            useKnowledgeStore.getState().applyNotificationsBulkRead(payload.item_ids)
          }
        } else if (data.type === 'channel.summary.updated') {
          // Phase 63F: live channel auto-summarize refresh (always-on rolling summary).
          // Backend payload: { channel_id, workspace_id, reason, summary, setting }.
          const payload = data.payload || {}
          if (payload.channel_id) {
            useKnowledgeStore.getState().applyChannelSummaryUpdated({
              channel_id: payload.channel_id,
              workspace_id: payload.workspace_id,
              reason: payload.reason,
              summary: payload.summary,
              setting: payload.setting,
            })
          }
        } else if (data.type === 'knowledge.compose.suggestion.generated') {
          // Phase 63F + 63G: co-drafting observer signal. The store prefers the
          // persisted `activity` row (63G) and falls back to synthesizing from
          // `compose` (63F parity). Payload: { compose?, activity? }.
          const payload = data.payload || {}
          if (payload.compose || payload.activity) {
            useKnowledgeStore.getState().applyComposeSuggestionGenerated({
              compose: payload.compose,
              activity: payload.activity,
            })
            // Phase 64C: append compose_activity events to unified feed
            if (payload.activity) {
              appendUnifiedFeedItem({
                id: `compose_activity:${payload.activity.id}`,
                event_type: 'compose_activity',
                workspace_id: payload.activity.workspace_id,
                actor_id: payload.activity.user_id,
                channel_id: payload.activity.channel_id,
                title: `Compose activity · ${payload.activity.entity_title}`,
                body: payload.activity.content,
                link: payload.activity.channel_id ? `/workspace?c=${payload.activity.channel_id}` : undefined,
                occurred_at: payload.activity.created_at ?? new Date().toISOString(),
              })
            }
          }
        } else if (
          data.type === 'knowledge.entity.brief.regen.queued' ||
          data.type === 'knowledge.entity.brief.regen.started' ||
          data.type === 'knowledge.entity.brief.regen.failed'
        ) {
          // Phase 63H: entity brief automation job lifecycle.
          // Payload: { job: AIAutomationJob, entity: KnowledgeEntity, reason? }
          const payload = data.payload || {}
          useKnowledgeStore.getState().applyEntityBriefAutomationEvent(data.type, payload)
          // Phase 64C: live-append succeeded automation jobs to unified feed
          const job = payload.job
          const entity = payload.entity
          if (job) {
            appendUnifiedFeedItem({
              id: `automation_job:${job.id}`,
              event_type: 'automation_job',
              workspace_id: job.workspace_id,
              entity_id: entity?.id ?? job.scope_id,
              entity_title: entity?.title,
              entity_kind: entity?.kind,
              title: `Automation · ${(job.job_type ?? 'job').replace(/_/g, ' ')} · ${job.status}`,
              body: job.last_error ?? job.trigger_reason,
              link: entity?.id ? `/workspace/knowledge/${entity.id}` : undefined,
              occurred_at: job.finished_at ?? job.started_at ?? job.created_at ?? new Date().toISOString(),
              meta: { status: job.status, attempt_count: job.attempt_count, job_type: job.job_type },
            })
          }
        } else if (
          data.type === 'schedule.event.booked' ||
          data.type === 'schedule.event.cancelled'
        ) {
          // Phase 63H: AI schedule booking lifecycle.
          // Payload: { booking: AIScheduleBooking }
          const payload = data.payload || {}
          useKnowledgeStore.getState().applyScheduleBookingEvent(data.type, payload)
          // Phase 64C: live-append to unified feed
          const booking = payload.booking
          if (booking && data.type === 'schedule.event.booked') {
            appendUnifiedFeedItem({
              id: `schedule_booking:${booking.id}`,
              event_type: 'schedule_booking',
              workspace_id: booking.workspace_id,
              actor_id: booking.requested_by,
              channel_id: booking.channel_id,
              dm_id: booking.dm_conversation_id,
              title: booking.title ?? 'Schedule booking',
              body: booking.description,
              link: booking.channel_id ? `/workspace?c=${booking.channel_id}` : '/workspace/dms',
              occurred_at: booking.created_at ?? new Date().toISOString(),
            })
          }
        } else if (data.type === 'knowledge.entity.ask.answered') {
          // Phase 63I: cross-entity ask answered — prepend to shared ask feed.
          // Payload: { item: KnowledgeAskRecentItem }
          const payload = data.payload || {}
          useKnowledgeStore.getState().applyEntityAskAnswered({ item: payload.item })
          // Phase 64C: live-append to unified feed
          const item = payload.item
          if (item) {
            appendUnifiedFeedItem({
              id: `knowledge_ask:${item.id}`,
              event_type: 'knowledge_ask',
              workspace_id: item.workspace_id,
              actor_id: item.user_id,
              entity_id: item.entity_id,
              entity_title: item.entity_title,
              entity_kind: item.entity_kind,
              title: `Ask AI · ${item.entity_title ?? 'entity'}`,
              body: item.question,
              link: item.entity_id ? `/workspace/knowledge/${item.entity_id}` : undefined,
              occurred_at: item.answered_at ?? item.created_at ?? new Date().toISOString(),
              meta: { citation_count: item.citation_count, provider: item.provider, model: item.model },
            })
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
