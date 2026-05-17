# UI Rewrite Plan

## Goal

Rewrite the entire UI from scratch to:

- replace the current visual language with a Telegram-inspired one;
- remove accumulated frontend technical debt;
- preserve the currently fixed backend contract and business behavior;
- keep the rewrite incremental enough to detect regressions early.

The backend is out of scope. Only `ui` changes are allowed.

## Hard Constraints

- Do not change HTTP endpoints, websocket event types, auth flow, or payload shapes unless the current UI already violates them and the fix is strictly local to the UI.
- Treat current business behavior as the source of truth until an explicit product decision says otherwise.
- Keep BEM naming in CSS.
- Keep 4-space indentation in `.tsx` files.
- Do not optimize for speed of rewriting at the cost of hidden behavior regressions.

## Business Logic That Must Stay Intact

### Authentication

- login flow;
- signup flow;
- login token storage and refresh;
- guest-only pages vs authenticated pages.

### Chats

- list chats for the authenticated user;
- open a chat and subscribe/open it over websocket;
- load messages from `revision`;
- request message history using a revision derived from last seen revision;
- scroll to the message at the current last seen revision after loading;
- optimistic message sending;
- message edit flow;
- message delete flow;
- chat preview updates from websocket;
- unread badge behavior based on revision math;
- add members to a group chat;
- kick members from a group chat;
- direct chat creation;
- group chat creation.

### Read State

- maintain local last seen revision state;
- sync last seen revision when the selected chat reaches bottom;
- keepalive sync on page leave / unload;
- react to `LastSeenRevisionUpdated` websocket events.

### Profile

- load profile data;
- update textual profile fields;
- avatar upload / paste / drag-and-drop behavior;
- crop flow and submission.

## Current Risk Areas

The rewrite must specifically target the largest debt concentrations:

- `ui/internal/src/pages/chats-me-page/chats-me-page.tsx`
- `ui/internal/src/pages/profile-page/profile-page.tsx`
- `ui/internal/src/contexts/message-web-socket-context/message-web-socket-context.tsx`
- `ui/internal/src/components/message-composer/message-composer.tsx`

These files currently mix transport, state orchestration, derived state, and presentation too aggressively.

## Rewrite Strategy

Do not replace the app in one pass.

Rewrite in this order:

1. Stabilize integration contracts.
2. Extract stateful application hooks from monolithic pages.
3. Introduce a new design system and global style tokens.
4. Rebuild layouts and simple pages.
5. Rebuild complex pages on top of extracted hooks.
6. Remove old monoliths only after behavior parity is reached.

## Target Architecture

### 1. Integration Layer

Create explicit UI-side modules for:

- HTTP request/response contracts;
- websocket event parsing and dispatch;
- auth token lifecycle;
- API violation mapping;
- message/chat/profile data mappers.

Goal:

- pages should not directly know transport-level wire details unless unavoidable.

### 2. Application Layer

Move orchestration logic from giant pages into dedicated hooks.

Target hooks for chats:

- `useChatsList`
- `useSelectedChatMessages`
- `useChatRealtime`
- `useLastSeenRevisionSync`
- `useChatActions`
- `useChatMembers`

Target hooks for profile:

- `useProfileForm`
- `useAvatarCrop`
- `useProfileSubmission`

Goal:

- side effects live in hooks;
- page components become composition shells;
- selectors and derived state become explicit utilities instead of ad-hoc `useEffect` chains.

### 3. Presentation Layer

Introduce reusable UI primitives before rewriting pages:

- button;
- text input;
- textarea;
- modal;
- card / panel;
- avatar;
- badge;
- dropdown / context menu;
- empty state;
- error state;
- loading state;
- form helper / validation message.

Goal:

- pages are assembled from stable components;
- visual changes become centralized instead of page-local.

## Delivery Phases

### Phase 0. Baseline Audit

Produce a behavior checklist for each page and main user scenario.

Artifacts:

- page inventory;
- interaction inventory;
- websocket event inventory;
- manual regression checklist.

### Phase 1. UI Contract Isolation

Refactor without visual redesign yet:

- extract typed API modules;
- extract websocket command/event helpers;
- isolate auth/session logic;
- isolate chat/profile mappers.

Acceptance:

- no user-facing behavior change;
- app still builds and runs with the same flows.

### Phase 2. State Extraction

Split `ChatsMePage` and `ProfilePage` into hooks plus dumb components.

Acceptance:

- large page files shrink materially;
- side effects are concentrated in hooks;
- presentation components become mostly prop-driven.

### Phase 3. Design System

Introduce a Telegram-inspired style foundation:

- typography scale;
- color tokens;
- spacing tokens;
- border radii;
- shadow system;
- motion rules;
- responsive breakpoints.

Acceptance:

- simple pages can switch to the new design language using shared primitives only.

### Phase 4. Layout Rewrite

Rebuild:

- `GuestLayout`;
- `DesktopLayout`;
- top-level app spacing and responsive behavior.

Acceptance:

- navigation shells reflect the new system;
- page-level rewrites stop duplicating structural CSS.

### Phase 5. Simple Page Rewrite

Rebuild first:

- home;
- login;
- signup;
- not-found.

Acceptance:

- validates design system before touching the chat page.

### Phase 6. Profile Rewrite

Rebuild the profile page with the new primitives and extracted hooks.

Acceptance:

- profile data updates still work;
- avatar crop/upload behavior remains intact.

### Phase 7. Chats Rewrite

Rebuild the chat experience last.

Break it into subareas:

- sidebar / chat list;
- selected chat header;
- message list;
- message composer;
- details panel;
- modals and context menu.

Acceptance:

- all chat business flows still work;
- websocket-driven updates remain behaviorally equivalent;
- revision-based loading and last seen sync remain intact.

### Phase 8. Cleanup

- remove legacy components and dead helpers;
- collapse duplicate SASS;
- remove transitional wrappers;
- document the final frontend structure.

## Regression Control

Each phase must be validated against the following checklist:

- login works;
- signup works;
- logout/guest gating works;
- chats list loads;
- selected chat loads messages;
- messages load from revision correctly;
- scroll lands near the current last seen revision;
- optimistic message send still resolves correctly;
- incoming websocket messages update the selected chat and list previews;
- unread counts remain correct;
- edit/delete message still works;
- group create/add member/kick member still works;
- profile update still works;
- avatar upload/crop still works.

## Suggested Commit Groups

Keep commits split by semantic area:

1. `chore (ui): document rewrite plan`
2. `refactor (ui): isolate api and websocket contracts`
3. `refactor (ui): extract chat page state hooks`
4. `refactor (ui): extract profile page state hooks`
5. `feat (ui): introduce telegram-inspired design system`
6. `feat (ui): rewrite guest and desktop layouts`
7. `feat (ui): rewrite auth and static pages`
8. `feat (ui): rewrite profile page`
9. `feat (ui): rewrite chats experience`
10. `chore (ui): remove legacy components and styles`

## Non-Goals For The First Rewrite

- backend API redesign;
- websocket protocol redesign;
- product behavior changes disguised as UI cleanup;
- speculative general-purpose state libraries unless the current code proves they are needed.
