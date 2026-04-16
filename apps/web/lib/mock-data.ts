import { User, Workspace, Channel, Message } from "@/types"

export const USERS: User[] = [
  {
    id: "user-1",
    name: "Nikko Fu",
    email: "nikko@example.com",
    avatar: "https://github.com/nikkofu.png",
    status: "online",
    statusText: "Building Relay 🚀",
  },
  {
    id: "user-2",
    name: "AI Assistant",
    email: "ai@acim.ai",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=ai",
    status: "online",
    statusText: "Ready to help",
  },
  {
    id: "user-3",
    name: "John Doe",
    email: "john@example.com",
    avatar: "https://i.pravatar.cc/150?u=john",
    status: "away",
    lastSeen: "10m ago",
  },
  {
    id: "user-4",
    name: "Jane Smith",
    email: "jane@example.com",
    avatar: "https://i.pravatar.cc/150?u=jane",
    status: "offline",
    lastSeen: "2h ago",
  },
]

export const WORKSPACES: Workspace[] = [
  {
    id: "ws-1",
    name: "Acme Corp",
    slug: "acme-corp",
    logo: "AC",
  },
  {
    id: "ws-2",
    name: "Side Project",
    slug: "side-project",
    logo: "SP",
  },
]

export const CHANNELS: Channel[] = [
  {
    id: "ch-1",
    name: "general",
    description: "Company-wide announcements and work-based matters",
    type: "public",
    workspaceId: "ws-1",
    memberCount: 15,
    isStarred: true,
  },
  {
    id: "ch-2",
    name: "random",
    description: "Non-work-related banter and water cooler talk",
    type: "public",
    workspaceId: "ws-1",
    memberCount: 12,
  },
  {
    id: "ch-3",
    name: "engineering",
    description: "Technical discussions and code reviews",
    type: "public",
    workspaceId: "ws-1",
    memberCount: 8,
    unreadCount: 3,
  },
  {
    id: "ch-4",
    name: "design",
    description: "UI/UX design critiques and inspiration",
    type: "public",
    workspaceId: "ws-1",
    memberCount: 5,
  },
  {
    id: "ch-5",
    name: "ai-lab",
    description: "Exploring the future of AI-native applications",
    type: "public",
    workspaceId: "ws-1",
    memberCount: 10,
    isStarred: true,
    unreadCount: 12,
  },
  {
    id: "ch-6",
    name: "private-deal",
    type: "private",
    workspaceId: "ws-1",
    memberCount: 3,
  },
]

export const MESSAGES: Message[] = [
  {
    id: "msg-1",
    content: "Welcome to the team everyone! Glad to have you here.",
    senderId: "user-1",
    channelId: "ch-1",
    createdAt: "2026-04-14T10:00:00.000Z",
    reactions: [{ emoji: "👋", count: 3, userIds: ["user-2", "user-3", "user-4"] }],
  },
  {
    id: "msg-2",
    content: "Thanks @Nikko Fu! Excited to get started on the AI integration.",
    senderId: "user-3",
    channelId: "ch-1",
    createdAt: "2026-04-14T11:30:00.000Z",
    replyCount: 2,
    lastReplyAt: "2026-04-14T12:00:00.000Z",
  },
  {
    id: "msg-3",
    content: "I've drafted the implementation plan for Relay Agent Workspace. Take a look: https://github.com/nikkofu/relay-agent-workspace",
    senderId: "user-1",
    channelId: "ch-5",
    createdAt: "2026-04-15T08:00:00.000Z",
    attachments: [
      {
        id: "att-1",
        type: "link",
        url: "https://github.com/nikkofu/relay-agent-workspace",
        name: "Relay Agent Workspace Repository",
      },
    ],
  },
  {
    id: "msg-4",
    content: "The AI components look promising. Can we add support for streaming responses?",
    senderId: "user-4",
    channelId: "ch-5",
    createdAt: "2026-04-15T08:30:00.000Z",
  },
  {
    id: "msg-5",
    content: "Absolutely. Vercel AI SDK provides great primitives for that. I'll include it in Phase 3.",
    senderId: "user-1",
    channelId: "ch-5",
    createdAt: "2026-04-15T09:00:00.000Z",
    reactions: [{ emoji: "✅", count: 2, userIds: ["user-4", "user-2"] }],
  },
  {
    id: "msg-6",
    content: "Here is a quick summary of today's progress:\n- Setup project with Next.js 15\n- Initialized shadcn/ui components\n- Defined core types and mock data",
    senderId: "user-2",
    channelId: "ch-5",
    createdAt: "2026-04-15T09:30:00.000Z",
  },
]
