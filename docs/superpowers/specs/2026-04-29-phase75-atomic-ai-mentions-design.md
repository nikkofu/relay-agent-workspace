# Phase 75 Atomic Mentions And Channel AI Replies Design

## Product Definition

Phase 75 turns channel mentions into structured workspace objects instead of plain text. A user mention such as `@AI Assistant` must behave as one indivisible editor chip, persist as deterministic mention metadata, and trigger an AI reply when the mentioned user is classified as an AI or bot user.

This phase extends the Slack-like channel composer and AI-native messaging loop. It does not replace DMs, rebuild the whole editor, or add autonomous execution without an explicit user message.

## Goals

- Make user mentions atomic in the channel/thread composer.
- Delete the whole mention chip when Backspace is pressed at the chip boundary.
- Preserve render-ready mention metadata when messages are sent, reloaded, or received through websocket events.
- Add an explicit user classification contract for `human`, `bot`, and `ai` users while preserving existing AI-user heuristics as fallback.
- Trigger channel AI auto-replies when a posted channel message mentions an AI/bot user.
- Reuse the existing AI DM side-channel experience: thinking, reasoning, tool calls, execution trace where available, and token usage.
- Keep `/ask` compatible, but make AI mention replies the richer default channel-native path.

## Non-Goals

- Do not build a new rich-text editor from scratch.
- Do not auto-run AI when a human user is mentioned.
- Do not trigger AI on drafts; only posted messages can trigger replies.
- Do not let AI replies recursively trigger more AI replies.
- Do not introduce a separate AI chat product inside channels.
- Do not block Sales App Phase 74 follow-up work.

## Mention Chip Contract

Windsurf owns the Web implementation.

Composer mention chips should be inline atom nodes with these attributes:

```json
{
  "kind": "user",
  "user_id": "user-2",
  "name": "AI Assistant",
  "user_type": "ai",
  "mention_text": "@AI Assistant"
}
```

The serialized HTML must include deterministic data attributes so the backend can parse mentions without relying only on fuzzy text matching:

```html
<span
  data-mention-kind="user"
  data-mention-user-id="user-2"
  data-mention-name="AI Assistant"
  data-mention-user-type="ai"
  contenteditable="false"
>@AI Assistant</span>
```

The chip must be:

- inline
- atomic
- selectable
- copy/paste safe enough to keep readable `@Name` text
- visually distinct from normal text
- compatible with existing `@entity:` autocomplete behavior

Backspace behavior:

- If the caret is immediately after a mention chip, Backspace removes the whole chip.
- If the mention chip is selected, Backspace/Delete removes the whole chip.
- Backspace must not remove only the final character of the displayed name.

## Mention Picker Contract

`MentionPopover` should return structured objects instead of a plain name string.

Recommended shape:

```ts
type ComposerMentionTarget =
  | {
      kind: "user"
      user_id: string
      name: string
      user_type?: "human" | "bot" | "ai"
      avatar?: string
    }
  | {
      kind: "group"
      group_id: string
      handle: string
      name: string
    }
```

User groups can remain non-AI mentions in this phase. They should not trigger AI replies.

## Backend Mention Parsing Contract

Gemini owns the backend implementation.

`CreateMessage` should accept the existing `content` field and continue to support text fallback parsing. It should additionally parse structured mention spans from sanitized HTML:

- `data-mention-kind="user"`
- `data-mention-user-id`
- `data-mention-name`
- `data-mention-user-type`

If both structured spans and text fallback identify the same user in the same message, persist only one `MessageMention` row because the table already enforces message/user/kind uniqueness.

`message.metadata.user_mentions[]` remains the render-ready source for message bubbles, activity rows, and notification surfaces. The realtime `message.created` payload must include metadata so newly received messages render the same as reloaded messages.

## AI User Classification

Add a durable user type field exposed to Web:

- `human`
- `bot`
- `ai`

Recommended backend field:

```go
UserType string `json:"user_type"`
```

Seed `AI Assistant` (`user-2`) as `ai`.

Fallback remains required for old rows and local data:

- name contains `ai assistant`
- name contains `assistant`
- email starts with `ai@`
- email contains `ai-assistant`

The fallback should be used only when `user_type` is empty.

## Channel AI Reply Contract

When `CreateMessage` saves a channel message:

1. Persist user mentions.
2. Refresh message metadata.
3. Broadcast the user message.
4. Detect AI/bot users among persisted mentions.
5. Start a background AI channel reply for the first eligible AI mention.

Eligibility:

- channel message only
- mentioned user type is `ai` or `bot`
- sender is not the mentioned AI user
- message is not already an AI-generated reply
- no recursive trigger from AI replies

The AI reply should be saved as a normal channel `Message` from the mentioned AI user. Its metadata should include canonical `metadata.ai_sidecar` so `MessageItem` can reuse the existing reasoning/tool/usage blocks.

Recommended AI reply metadata:

```json
{
  "ai_sidecar": {
    "reasoning": {},
    "tool_calls": [],
    "usage": {}
  },
  "ai_mention_reply": {
    "trigger_message_id": "msg-123",
    "mentioned_user_id": "user-2",
    "scope_type": "channel",
    "channel_id": "ch-1"
  }
}
```

## Streaming Contract

Reuse the existing unified AI stream semantics. The Web can either consume a new channel-scoped websocket event or a generalized stream event, but it must show progress in channels similar to AI DMs.

Recommended websocket event:

- `channel.ai.stream.chunk`

Recommended payload:

```json
{
  "temp_id": "channel-ai-stream-1",
  "channel_id": "ch-1",
  "trigger_message_id": "msg-123",
  "ai_user_id": "user-2",
  "kind": "answer",
  "chunk": "Working through the request...",
  "is_final": false
}
```

Supported stream kinds should align with the existing side-channel contract:

- `answer`
- `reasoning`
- `tool_call`
- `usage`
- `error`

On final completion, the backend persists the final message and broadcasts `message.created`. The Web should remove the temporary streaming row when the persisted message arrives.

## Existing Gaps To Close

- `MessageComposer` currently inserts mentions as plain text through `editor.commands.insertContent(name + " ")`.
- `MentionPopover` currently passes only a name string and identifies AI Assistant with `user.id === "user-2"`.
- `CreateMessage` currently calls `persistMessageMentions` with text content only.
- `triggerAIDMReply` exists for DMs, but channel mention replies do not reuse it yet.
- `/ask` creates a channel AI placeholder with reserved user id `ai-assistant`; Phase 75 should prefer the real mentioned AI user.
- `useWebsocket` currently maps `message.created` into a minimal message and drops `metadata`; this blocks live AI sidecar rendering until refresh.

## Ownership

- Gemini owns backend/API/tests:
  - user type field and seed update
  - structured mention parsing
  - channel AI mention reply trigger
  - stream/realtime contract
  - backend tests
- Windsurf owns Web/UI:
  - Tiptap mention atom
  - structured mention picker result
  - channel AI streaming UI
  - metadata-preserving websocket mapping
  - Web checks
- Codex owns contract, sequencing, collaboration docs, release coordination, and integration audit.

## Acceptance Criteria

- Selecting `@AI Assistant` inserts one atomic chip, not plain text.
- Backspace after the chip deletes the whole mention.
- Posted messages persist `message.metadata.user_mentions[]`.
- Reloaded messages and realtime `message.created` messages render the same mention/AI metadata.
- Mentioning a human user does not trigger AI.
- Mentioning `AI Assistant` in a channel triggers one AI reply from that AI user.
- AI replies show answer content plus reasoning/tool/usage metadata in the channel message list.
- AI replies do not recursively mention-trigger themselves.
- Existing `/ask`, DM AI replies, entity mentions, file attachments, and Sales App routes remain working.
