import { MouseEvent, type CSSProperties, useEffect, useLayoutEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { getAuthenticatedUserUuid, getLoginToken } from "../../auth/auth-tokens";
import {
    MessageCreatedEvent,
    MessageEditedEvent,
    useMessageWebSocket,
} from "../../contexts/message-web-socket-context/message-web-socket-context";
import {
    addChatMembers,
    changeChatAvatar,
    createDirectChat,
    createGroupChat as createGroupChatRequest,
    createMessage,
    deleteMessage as deleteMessageRequest,
    editMessage as editMessageRequest,
    kickChatMember,
    listChats,
    renameChat,
    setLastSeenRevision,
} from "../../features/chats/chat-api";
import { useChatsList } from "../../features/chats/hooks/use-chats-list";
import { useSelectedChatMessages } from "../../features/chats/hooks/use-selected-chat-messages";
import { ChatDetailsPanel } from "../../features/chats/components/chat-details-panel";
import { ChatMessagesPanel, type ChatMessageViewMode, MessageContextMenu } from "../../features/chats/components/chat-messages-panel";
import { ChatsSidebar } from "../../features/chats/components/chats-sidebar";
import { DeleteMessageDialog } from "../../features/chats/components/delete-message-dialog";
import { GroupMembersDialog } from "../../features/chats/components/group-members-dialog";
import {
    Chat,
    ChatMessage,
    collectReferencedUserUuids,
    getTargetLastSeenRevision,
    profilesByUserUuidFromProfiles,
} from "../../features/chats/components/chats-ui";
import { loadProfilesBatch, searchProfiles } from "../../features/profile/profile-api";
import { useApiClient } from "../../hooks/use-api-client";
import { getCachedProfilesByUserUuids, getMissingOrStaleUserUuids, getStaleCachedUserUuids, PublicProfile, upsertProfiles } from "../../profiles/profile-cache";
import "./chats-me-page.sass";

enum ChatType {
    Direct = 0,
    Group = 1,
}

enum ChatMemberRole {
    Owner = "owner",
    Admin = "admin",
    Member = "member",
}

const lastSeenRevisionSyncIntervalMs = 30_000;
const chatBottomThresholdPx = 24;
const chatsSidebarWidthStorageKey = "vox.chats.sidebarWidth";
const chatMessageViewModeStorageKey = "vox.chats.messageViewMode";
const chatsSidebarMinWidthPx = 230;
const chatsSidebarMaxWidthPx = 520;
const chatsSidebarDefaultWidthPx = 350;
type MobilePanel = "list" | "chat" | "details";

async function searchProfilesByQuery(apiClient: ReturnType<typeof useApiClient>, query: string, authenticatedUserUuid: string | null) {
    const data = await searchProfiles(apiClient, query, 10);
    return data.filter((profile) => profile.userUuid !== authenticatedUserUuid);
}

export default function ChatsMePage() {
    const apiClient = useApiClient();
    const navigate = useNavigate();
    const {
        isConnected,
        messageCreatedListener,
        messageDeletedListener,
        messageEditedListener,
        lastSeenRevisionUpdatedListener,
        subscribeChat,
        unsubscribeChat,
    } = useMessageWebSocket();
    const { chatUuid } = useParams();
    const messagesContainerRef = useRef<HTMLDivElement | null>(null);
    const messagesEndRef = useRef<HTMLDivElement | null>(null);
    const messageElementByRevisionRef = useRef<Record<number, HTMLElement | null>>({});
    const messageReceivedAudioRef = useRef<HTMLAudioElement | null>(null);
    const searchInputRef = useRef<HTMLInputElement | null>(null);
    const chatsRef = useRef<Chat[]>([]);
    const localLastSeenRevisionByChatUuidRef = useRef<Record<string, number>>({});
    const selectedChatUuidRef = useRef<string | null>(null);
    const isSelectedChatScrolledToBottomRef = useRef(false);
    const shouldAutoScrollSelectedChatOnNextMessageRef = useRef(false);
    const shouldSkipNextBottomScrollSyncRef = useRef(false);
    const pendingLastSeenRevisionSyncByChatUuidRef = useRef<Record<string, number>>({});
    const sidebarResizeStateRef = useRef<{ startX: number; startWidth: number; } | null>(null);
    const {
        chats,
        setChats,
        isLoading,
        error,
    } = useChatsList();
    const [editingMessage, setEditingMessage] = useState<ChatMessage | null>(null);
    const [messageContextMenu, setMessageContextMenu] = useState<MessageContextMenu | null>(null);
    const [messagePendingDeletion, setMessagePendingDeletion] = useState<ChatMessage | null>(null);
    const [isDeletingMessage, setIsDeletingMessage] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [searchProfiles, setSearchProfiles] = useState<PublicProfile[]>([]);
    const [isSearchFocused, setIsSearchFocused] = useState(false);
    const [isSearchingProfiles, setIsSearchingProfiles] = useState(false);
    const [searchProfilesError, setSearchProfilesError] = useState<string | null>(null);
    const [isCreatingChat, setIsCreatingChat] = useState(false);
    const [isCreateGroupChatModalOpen, setIsCreateGroupChatModalOpen] = useState(false);
    const [groupChatSearchQuery, setGroupChatSearchQuery] = useState("");
    const [groupChatSearchProfiles, setGroupChatSearchProfiles] = useState<PublicProfile[]>([]);
    const [isGroupChatSearchFocused, setIsGroupChatSearchFocused] = useState(false);
    const [isSearchingGroupChatProfiles, setIsSearchingGroupChatProfiles] = useState(false);
    const [groupChatSearchProfilesError, setGroupChatSearchProfilesError] = useState<string | null>(null);
    const [groupChatInvitees, setGroupChatInvitees] = useState<PublicProfile[]>([]);
    const [isCreatingGroupChat, setIsCreatingGroupChat] = useState(false);
    const [groupChatName, setGroupChatName] = useState("");
    const [isAddMembersModalOpen, setIsAddMembersModalOpen] = useState(false);
    const [addMembersSearchQuery, setAddMembersSearchQuery] = useState("");
    const [addMembersSearchProfiles, setAddMembersSearchProfiles] = useState<PublicProfile[]>([]);
    const [isAddMembersSearchFocused, setIsAddMembersSearchFocused] = useState(false);
    const [isSearchingAddMembersProfiles, setIsSearchingAddMembersProfiles] = useState(false);
    const [addMembersSearchProfilesError, setAddMembersSearchProfilesError] = useState<string | null>(null);
    const [addMembersInvitees, setAddMembersInvitees] = useState<PublicProfile[]>([]);
    const [isAddingMembers, setIsAddingMembers] = useState(false);
    const [isKickingMember, setIsKickingMember] = useState(false);
    const [localLastSeenRevisionByChatUuid, setLocalLastSeenRevisionByChatUuid] = useState<Record<string, number>>({});
    const [profilesByUserUuid, setProfilesByUserUuid] = useState<Record<string, PublicProfile>>({});
    const [sidebarWidth, setSidebarWidth] = useState(() => loadChatsSidebarWidth());
    const [isSidebarResizing, setIsSidebarResizing] = useState(false);
    const [messageViewMode, setMessageViewMode] = useState<ChatMessageViewMode>(() => loadChatMessageViewMode());
    const [mobilePanel, setMobilePanel] = useState<MobilePanel>(() => (chatUuid ? "chat" : "list"));
    const authenticatedUserUuid = getAuthenticatedUserUuid();
    const selectedChatUuid = chatUuid ?? null;
    const selectedChat = chats.find((chat) => chat.uuid === selectedChatUuid);
    const isMobileLayout = useIsPortraitLayout();
    const {
        messages,
        setMessages,
        isMessagesLoading,
        messagesError,
        targetVisibleRevision,
        setTargetVisibleRevision,
    } = useSelectedChatMessages({
        selectedChatUuid,
        selectedChat,
        localLastSeenRevisionByChatUuid,
    });
    const isSearchActive = isSearchFocused || searchQuery.length > 0;
    const groupChatSearchInputRef = useRef<HTMLInputElement | null>(null);
    const addMembersSearchInputRef = useRef<HTMLInputElement | null>(null);
    const groupChatSearchQueryTrimmed = groupChatSearchQuery.trim();
    const groupChatInviteeUserUuids = new Set(groupChatInvitees.map((profile) => profile.userUuid));
    const groupChatVisibleSearchProfiles = [
        ...groupChatInvitees,
        ...groupChatSearchProfiles.filter((profile) => !groupChatInviteeUserUuids.has(profile.userUuid)),
    ];

    useEffect(() => {
        chatsRef.current = chats;

        const nextLocalLastSeenRevisionByChatUuid: Record<string, number> = {};
        chats.forEach((chat) => {
            const currentLocalRevision = localLastSeenRevisionByChatUuidRef.current[chat.uuid] ?? 0;
            nextLocalLastSeenRevisionByChatUuid[chat.uuid] = Math.max(currentLocalRevision, chat.currentUserLastSeenRevision);
        });

        localLastSeenRevisionByChatUuidRef.current = nextLocalLastSeenRevisionByChatUuid;
        setLocalLastSeenRevisionByChatUuid((currentLocalLastSeenRevisionByChatUuid) => (
            areLastSeenRevisionMapsEqual(currentLocalLastSeenRevisionByChatUuid, nextLocalLastSeenRevisionByChatUuid)
                ? currentLocalLastSeenRevisionByChatUuid
                : nextLocalLastSeenRevisionByChatUuid
        ));
    }, [chats]);

    useEffect(() => {
        selectedChatUuidRef.current = selectedChatUuid;
    }, [selectedChatUuid]);

    useEffect(() => {
        window.localStorage.setItem(chatsSidebarWidthStorageKey, String(sidebarWidth));
    }, [sidebarWidth]);

    useEffect(() => {
        window.localStorage.setItem(chatMessageViewModeStorageKey, messageViewMode);
    }, [messageViewMode]);

    useEffect(() => {
        if (!isMobileLayout) {
            return;
        }

        setMobilePanel(selectedChatUuid ? "chat" : "list");
    }, [isMobileLayout, selectedChatUuid]);

    useEffect(() => {
        const audio = new Audio("/message-received.mp3");
        audio.preload = "auto";
        messageReceivedAudioRef.current = audio;

        return () => {
            audio.pause();
            audio.src = "";
            messageReceivedAudioRef.current = null;
        };
    }, []);

    useEffect(() => {
        if (!isSidebarResizing) {
            return;
        }

        function handleMouseMove(event: globalThis.MouseEvent) {
            const resizeState = sidebarResizeStateRef.current;
            if (!resizeState) {
                return;
            }

            const widthDelta = event.clientX - resizeState.startX;
            setSidebarWidth(clampChatsSidebarWidth(resizeState.startWidth + widthDelta));
        }

        function handleMouseUp() {
            sidebarResizeStateRef.current = null;
            setIsSidebarResizing(false);
        }

        window.addEventListener("mousemove", handleMouseMove);
        window.addEventListener("mouseup", handleMouseUp);

        return () => {
            window.removeEventListener("mousemove", handleMouseMove);
            window.removeEventListener("mouseup", handleMouseUp);
        };
    }, [isSidebarResizing]);

    useEffect(() => {
        const staleCachedUserUuids = getStaleCachedUserUuids();
        if (staleCachedUserUuids.length === 0) {
            return;
        }

        let isMounted = true;

        async function refreshStaleProfiles() {
            try {
                const data = await loadProfilesBatch(apiClient, {
                    userUuids: staleCachedUserUuids,
                });
                if (!isMounted || data.length === 0) {
                    return;
                }

                upsertProfiles(data);
                setProfilesByUserUuid((currentProfilesByUserUuid) => ({
                    ...currentProfilesByUserUuid,
                    ...profilesByUserUuidFromProfiles(data),
                }));
            } catch {
                // Keep stale cache entries until a later refresh succeeds.
            }
        }

        void refreshStaleProfiles();

        return () => {
            isMounted = false;
        };
    }, [apiClient]);

    useEffect(() => {
        const referencedUserUuids = collectReferencedUserUuids(chats, messages, searchProfiles);
        if (referencedUserUuids.length === 0) {
            return;
        }

        const cachedProfiles = getCachedProfilesByUserUuids(referencedUserUuids);
        if (Object.keys(cachedProfiles).length > 0) {
            setProfilesByUserUuid((currentProfilesByUserUuid) => ({
                ...currentProfilesByUserUuid,
                ...cachedProfiles,
            }));
        }

        const missingOrStaleUserUuids = getMissingOrStaleUserUuids(referencedUserUuids);
        if (missingOrStaleUserUuids.length === 0) {
            return;
        }

        let isMounted = true;

        async function loadProfiles() {
            try {
                const data = await loadProfilesBatch(apiClient, {
                    userUuids: missingOrStaleUserUuids,
                });
                if (!isMounted || data.length === 0) {
                    return;
                }

                upsertProfiles(data);
                setProfilesByUserUuid((currentProfilesByUserUuid) => ({
                    ...currentProfilesByUserUuid,
                    ...profilesByUserUuidFromProfiles(data),
                }));
            } catch {
                // Keep rendering cached or fallback values if profile loading fails.
            }
        }

        void loadProfiles();

        return () => {
            isMounted = false;
        };
    }, [apiClient, chats, messages, searchProfiles]);

    useEffect(() => {
        if (!isSearchActive) {
            setSearchProfiles([]);
            setSearchProfilesError(null);
            setIsSearchingProfiles(false);
            return;
        }

        const trimmedQuery = searchQuery.trim();
        if (trimmedQuery.length === 0) {
            setSearchProfiles([]);
            setSearchProfilesError(null);
            setIsSearchingProfiles(false);
            return;
        }

        let isMounted = true;
        const timeoutId = window.setTimeout(async () => {
            setIsSearchingProfiles(true);
            setSearchProfilesError(null);

            try {
                const nextProfiles = await searchProfilesByQuery(apiClient, trimmedQuery, authenticatedUserUuid);
                if (!isMounted) {
                    return;
                }

                upsertProfiles(nextProfiles);
                setProfilesByUserUuid((currentProfilesByUserUuid) => ({
                    ...currentProfilesByUserUuid,
                    ...profilesByUserUuidFromProfiles(nextProfiles),
                }));
                setSearchProfiles(nextProfiles);
            } catch {
                if (!isMounted) {
                    return;
                }

                setSearchProfiles([]);
                setSearchProfilesError("Could not search users.");
            } finally {
                if (isMounted) {
                    setIsSearchingProfiles(false);
                }
            }
        }, 300);

        return () => {
            isMounted = false;
            window.clearTimeout(timeoutId);
        };
    }, [apiClient, authenticatedUserUuid, isSearchActive, searchQuery]);

    useEffect(() => {
        if (!isCreateGroupChatModalOpen) {
            setGroupChatName("");
            setGroupChatSearchQuery("");
            setGroupChatSearchProfiles([]);
            setGroupChatSearchProfilesError(null);
            setIsSearchingGroupChatProfiles(false);
            setIsGroupChatSearchFocused(false);
            setGroupChatInvitees([]);
            return;
        }

        const timeoutId = window.setTimeout(() => {
            groupChatSearchInputRef.current?.focus();
        }, 0);

        return () => {
            window.clearTimeout(timeoutId);
        };
    }, [isCreateGroupChatModalOpen]);

    useEffect(() => {
        if (!isCreateGroupChatModalOpen) {
            return;
        }

        const trimmedQuery = groupChatSearchQuery.trim();
        if (trimmedQuery.length === 0) {
            setGroupChatSearchProfiles([]);
            setGroupChatSearchProfilesError(null);
            setIsSearchingGroupChatProfiles(false);
            return;
        }

        let isMounted = true;
        const timeoutId = window.setTimeout(async () => {
            setIsSearchingGroupChatProfiles(true);
            setGroupChatSearchProfilesError(null);

            try {
                const nextProfiles = await searchProfilesByQuery(apiClient, trimmedQuery, authenticatedUserUuid);
                if (!isMounted) {
                    return;
                }

                upsertProfiles(nextProfiles);
                setProfilesByUserUuid((currentProfilesByUserUuid) => ({
                    ...currentProfilesByUserUuid,
                    ...profilesByUserUuidFromProfiles(nextProfiles),
                }));
                setGroupChatSearchProfiles(nextProfiles);
            } catch {
                if (!isMounted) {
                    return;
                }

                setGroupChatSearchProfiles([]);
                setGroupChatSearchProfilesError("Could not search users.");
            } finally {
                if (isMounted) {
                    setIsSearchingGroupChatProfiles(false);
                }
            }
        }, 300);

        return () => {
            isMounted = false;
            window.clearTimeout(timeoutId);
        };
    }, [apiClient, authenticatedUserUuid, groupChatSearchQuery, isCreateGroupChatModalOpen]);

    useEffect(() => {
        if (!isAddMembersModalOpen) {
            setAddMembersSearchQuery("");
            setAddMembersSearchProfiles([]);
            setAddMembersSearchProfilesError(null);
            setIsSearchingAddMembersProfiles(false);
            setIsAddMembersSearchFocused(false);
            setAddMembersInvitees([]);
            return;
        }

        const timeoutId = window.setTimeout(() => {
            addMembersSearchInputRef.current?.focus();
        }, 0);

        return () => {
            window.clearTimeout(timeoutId);
        };
    }, [isAddMembersModalOpen]);

    useEffect(() => {
        if (!isAddMembersModalOpen) {
            return;
        }

        const trimmedQuery = addMembersSearchQuery.trim();
        if (trimmedQuery.length === 0) {
            setAddMembersSearchProfiles([]);
            setAddMembersSearchProfilesError(null);
            setIsSearchingAddMembersProfiles(false);
            return;
        }

        let isMounted = true;
        const timeoutId = window.setTimeout(async () => {
            setIsSearchingAddMembersProfiles(true);
            setAddMembersSearchProfilesError(null);

            try {
                const nextProfiles = await searchProfilesByQuery(apiClient, trimmedQuery, authenticatedUserUuid);
                if (!isMounted) {
                    return;
                }

                upsertProfiles(nextProfiles);
                setProfilesByUserUuid((currentProfilesByUserUuid) => ({
                    ...currentProfilesByUserUuid,
                    ...profilesByUserUuidFromProfiles(nextProfiles),
                }));
                setAddMembersSearchProfiles(nextProfiles);
            } catch {
                if (!isMounted) {
                    return;
                }

                setAddMembersSearchProfiles([]);
                setAddMembersSearchProfilesError("Could not search users.");
            } finally {
                if (isMounted) {
                    setIsSearchingAddMembersProfiles(false);
                }
            }
        }, 300);

        return () => {
            isMounted = false;
            window.clearTimeout(timeoutId);
        };
    }, [apiClient, authenticatedUserUuid, addMembersSearchQuery, isAddMembersModalOpen]);

    useEffect(() => {
        if (!isConnected) {
            return;
        }

        if (!selectedChatUuid) {
            unsubscribeChat("");
            return;
        }

        subscribeChat(selectedChatUuid);

        return () => {
            unsubscribeChat(selectedChatUuid);
        };
    }, [isConnected, selectedChatUuid, subscribeChat, unsubscribeChat]);

    useEffect(() => {
        if (!selectedChatUuid) {
            isSelectedChatScrolledToBottomRef.current = false;
            messageElementByRevisionRef.current = {};
            return;
        }
    }, [selectedChatUuid, setMessages]);

    useLayoutEffect(() => {
        if (!selectedChatUuid || isMessagesLoading || messages.length === 0) {
            return;
        }

        if (targetVisibleRevision !== null) {
            scrollMessagesContainerToRevision(
                messagesContainerRef.current,
                messageElementByRevisionRef.current,
                findFirstVisibleRevision(messages, targetVisibleRevision),
            );
            setTargetVisibleRevision(null);
            isSelectedChatScrolledToBottomRef.current = isMessagesContainerScrolledToBottom(messagesContainerRef.current);
            return;
        }

        if (shouldAutoScrollSelectedChatOnNextMessageRef.current) {
            shouldSkipNextBottomScrollSyncRef.current = true;
            scrollMessagesContainerToBottom(messagesContainerRef.current, messagesEndRef.current);
            shouldAutoScrollSelectedChatOnNextMessageRef.current = false;
            isSelectedChatScrolledToBottomRef.current = true;
            return;
        }

        isSelectedChatScrolledToBottomRef.current = isMessagesContainerScrolledToBottom(messagesContainerRef.current);
    }, [isMessagesLoading, messages, selectedChatUuid, setTargetVisibleRevision, targetVisibleRevision]);

    useEffect(() => {
        setEditingMessage(null);
        setMessageContextMenu(null);
        setMessagePendingDeletion(null);
    }, [selectedChatUuid]);

    useEffect(() => (
        messageCreatedListener((event) => {
            setChats((currentChats) => currentChats.map((chat) => {
                if (chat.uuid !== event.chatUuid) {
                    return chat;
                }

                return {
                    ...chat,
                    revision: Math.max(chat.revision, event.revision),
                    lastMessage: messageCreatedEventToChatMessage(event),
                };
            }));

            if (event.chatUuid !== selectedChatUuidRef.current) {
                return;
            }

            const wasSelectedChatScrolledToBottom = isMessagesContainerScrolledToBottom(messagesContainerRef.current);
            if (wasSelectedChatScrolledToBottom) {
                shouldAutoScrollSelectedChatOnNextMessageRef.current = true;
                updateLocalLastSeenRevision(event.chatUuid, event.revision);
            }

            if (event.userUuid !== authenticatedUserUuid) {
                void messageReceivedAudioRef.current?.play();
            }

            setMessages((currentMessages) => {
                if (currentMessages.some((message) => message.uuid === event.messageUuid && !message.isPending)) {
                    return currentMessages;
                }

                const confirmedMessage = messageCreatedEventToChatMessage(event);
                return [
                    ...currentMessages,
                    confirmedMessage,
                ];
            });

        })
    ), [authenticatedUserUuid, messageCreatedListener]);

    useEffect(() => (
        messageEditedListener((event) => {
            if (event.chatUuid !== selectedChatUuidRef.current) {
                return;
            }

            setMessages((currentMessages) => currentMessages.map((message) => (
                message.uuid === event.messageUuid
                    ? messageEditedEventToChatMessage(event, message.revision)
                    : message
            )));
        })
    ), [messageEditedListener]);

    useEffect(() => (
        messageDeletedListener((event) => {
            if (event.chatUuid !== selectedChatUuidRef.current) {
                return;
            }

            setMessages((currentMessages) => currentMessages.filter((message) => message.uuid !== event.messageUuid));
        })
    ), [messageDeletedListener]);

    useEffect(() => (
        lastSeenRevisionUpdatedListener((event) => {
            if (event.userUuid !== authenticatedUserUuid) {
                return;
            }

            updateLocalLastSeenRevision(event.chatUuid, event.lastSeenRevision);

            setChats((currentChats) => currentChats.map((chat) => (
                chat.uuid === event.chatUuid
                    ? {
                        ...chat,
                        currentUserLastSeenRevision: event.lastSeenRevision,
                    }
                    : chat
            )));

            if (pendingLastSeenRevisionSyncByChatUuidRef.current[event.chatUuid] === event.lastSeenRevision) {
                delete pendingLastSeenRevisionSyncByChatUuidRef.current[event.chatUuid];
            }
        })
    ), [authenticatedUserUuid, lastSeenRevisionUpdatedListener]);

    useEffect(() => {
        const intervalId = window.setInterval(() => {
            const dirtyChats = getDirtyChats(chatsRef.current);
            dirtyChats.forEach((chat) => {
                void syncChatLastSeenRevisionIfNeeded(chat.uuid);
            });
        }, lastSeenRevisionSyncIntervalMs);

        return () => {
            window.clearInterval(intervalId);
        };
    }, []);

    useEffect(() => {
        function handlePageHide() {
            syncDirtyChatsWithKeepalive();
        }

        function handleBeforeUnload() {
            syncDirtyChatsWithKeepalive();
        }

        window.addEventListener("pagehide", handlePageHide);
        window.addEventListener("beforeunload", handleBeforeUnload);

        return () => {
            window.removeEventListener("pagehide", handlePageHide);
            window.removeEventListener("beforeunload", handleBeforeUnload);
        };
    }, []);

    useEffect(() => {
        return () => {
            if (selectedChatUuid) {
                void syncChatLastSeenRevisionIfNeeded(selectedChatUuid);
            }
        };
    }, [selectedChatUuid]);

    useEffect(() => {
        function handleWindowClick() {
            setMessageContextMenu(null);
        }

        function handleWindowKeyDown(event: KeyboardEvent) {
            if (event.key === "Escape") {
                setMessageContextMenu(null);
                setMessagePendingDeletion(null);
                closeCreateGroupChatModal();
                closeAddMembersModal();
            }
        }

        window.addEventListener("click", handleWindowClick);
        window.addEventListener("keydown", handleWindowKeyDown);

        return () => {
            window.removeEventListener("click", handleWindowClick);
            window.removeEventListener("keydown", handleWindowKeyDown);
        };
    }, []);

    const isSelectedChatGroup = selectedChat?.chatType === ChatType.Group;
    const isSelectedChatOwnedByAuthenticatedUser = selectedChat?.createdByUserUuid === authenticatedUserUuid;
    const canDeleteAnyMessage = isSelectedChatGroup && (
        selectedChat?.currentUserRole === ChatMemberRole.Owner
            || selectedChat?.currentUserRole === ChatMemberRole.Admin
    );
    const selectedChatMemberUserUuids = new Set(selectedChat?.memberUuids ?? []);
    const addMembersInviteeUserUuids = new Set(addMembersInvitees.map((profile) => profile.userUuid));
    const addMembersVisibleSearchProfiles = [
        ...addMembersInvitees,
        ...addMembersSearchProfiles.filter((profile) => (
            !selectedChatMemberUserUuids.has(profile.userUuid)
                && !addMembersInviteeUserUuids.has(profile.userUuid)
        )),
    ];

    async function submitMessage(text: string, attachments: string[]) {
        if (editingMessage) {
            await editMessage(editingMessage, text, attachments);
            return;
        }

        await sendMessage(text, attachments);
    }

    async function sendMessage(text: string, attachments: string[]) {
        if (!selectedChat || !authenticatedUserUuid) {
            return;
        }

        shouldAutoScrollSelectedChatOnNextMessageRef.current = true;

        const pendingMessageUuid = `pending-${generatePendingMessageUuid()}`;
        const createdAt = new Date().toISOString();
        const pendingAttachments = attachments.map((name) => ({
            uuid: name,
            name,
        }));

        setMessages((currentMessages) => [
            ...currentMessages,
            {
                uuid: pendingMessageUuid,
                chatUuid: selectedChat.uuid,
                userUuid: authenticatedUserUuid,
                revision: selectedChat.revision + 1,
                text,
                attachments: pendingAttachments,
                createdAt,
                updatedAt: createdAt,
                isPending: true,
            },
        ]);

        try {
            const messageUuid = await createMessage(apiClient, selectedChat.uuid, {
                text,
                attachments,
            });

            setMessages((currentMessages) => confirmPendingMessage(currentMessages, pendingMessageUuid, messageUuid));
        } catch (error) {
            setMessages((currentMessages) => currentMessages.filter((message) => message.uuid !== pendingMessageUuid));
            throw error;
        }
    }

    async function editMessage(message: ChatMessage, text: string, attachments: string[]) {
        await editMessageRequest(apiClient, message.uuid, {
            text,
            attachments,
        });

        setEditingMessage(null);
    }

    async function deleteMessage(message: ChatMessage) {
        setIsDeletingMessage(true);
        try {
            await deleteMessageRequest(apiClient, message.uuid);
            setMessagePendingDeletion(null);
            if (editingMessage?.uuid === message.uuid) {
                setEditingMessage(null);
            }
        } finally {
            setIsDeletingMessage(false);
        }
    }

    function handleMessageContextMenu(event: MouseEvent<HTMLElement>, message: ChatMessage) {
        event.preventDefault();

        if (message.isPending) {
            return;
        }

        setMessageContextMenu({
            message,
            x: event.clientX,
            y: event.clientY,
        });
    }

    async function copyMessageText(message: ChatMessage) {
        try {
            await navigator.clipboard.writeText(message.text);
        } catch {
            copyTextWithFallback(message.text);
        } finally {
            setMessageContextMenu(null);
        }
    }

    async function copyMessageLink(message: ChatMessage) {
        const messageLink = `${window.location.origin}/chats/@me/${message.chatUuid}/${message.uuid}`;

        try {
            await navigator.clipboard.writeText(messageLink);
        } catch {
            copyTextWithFallback(messageLink);
        } finally {
            setMessageContextMenu(null);
        }
    }

    async function createPrivateChat(profile: PublicProfile) {
        setIsCreatingChat(true);
        try {
            upsertProfiles([profile]);
            setProfilesByUserUuid((currentProfilesByUserUuid) => ({
                ...currentProfilesByUserUuid,
                [profile.userUuid]: profile,
            }));

            const createdChatUuid = await createDirectChat(apiClient, {
                interlocutorUuid: profile.userUuid,
            });

            const nextChats = await listChats(apiClient);
            setChats(nextChats);
            clearSearch();
            navigate(`/chats/@me/${createdChatUuid}`);
        } finally {
            setIsCreatingChat(false);
        }
    }

    function openCreateGroupChatModal() {
        setIsCreateGroupChatModalOpen(true);
    }

    function closeCreateGroupChatModal() {
        setIsCreateGroupChatModalOpen(false);
        setGroupChatName("");
        setGroupChatSearchQuery("");
        setGroupChatSearchProfiles([]);
        setGroupChatSearchProfilesError(null);
        setIsSearchingGroupChatProfiles(false);
        setIsGroupChatSearchFocused(false);
        setGroupChatInvitees([]);
    }

    function toggleGroupChatInvitee(profile: PublicProfile) {
        setGroupChatInvitees((currentInvitees) => {
            if (currentInvitees.some((invitee) => invitee.userUuid === profile.userUuid)) {
                return currentInvitees.filter((invitee) => invitee.userUuid !== profile.userUuid);
            }

            return [...currentInvitees, profile];
        });
    }

    async function createGroupChat() {
        setIsCreatingGroupChat(true);
        try {
            upsertProfiles(groupChatInvitees);
            setProfilesByUserUuid((currentProfilesByUserUuid) => ({
                ...currentProfilesByUserUuid,
                ...profilesByUserUuidFromProfiles(groupChatInvitees),
            }));

            const createdChatUuid = await createGroupChatRequest(apiClient, {
                memberUuids: groupChatInvitees.map((profile) => profile.userUuid),
                name: groupChatName.trim().length > 0 ? groupChatName.trim() : null,
            });

            const nextChats = await listChats(apiClient);
            setChats(nextChats);
            closeCreateGroupChatModal();
            navigate(`/chats/@me/${createdChatUuid}`);
        } finally {
            setIsCreatingGroupChat(false);
        }
    }

    function openAddMembersModal() {
        setIsAddMembersModalOpen(true);
    }

    function closeAddMembersModal() {
        setIsAddMembersModalOpen(false);
        setAddMembersSearchQuery("");
        setAddMembersSearchProfiles([]);
        setAddMembersSearchProfilesError(null);
        setIsSearchingAddMembersProfiles(false);
        setIsAddMembersSearchFocused(false);
        setAddMembersInvitees([]);
    }

    function toggleAddMembersInvitee(profile: PublicProfile) {
        setAddMembersInvitees((currentInvitees) => {
            if (currentInvitees.some((invitee) => invitee.userUuid === profile.userUuid)) {
                return currentInvitees.filter((invitee) => invitee.userUuid !== profile.userUuid);
            }

            return [...currentInvitees, profile];
        });
    }

    async function addMembersToChat() {
        if (!selectedChat) {
            return;
        }

        setIsAddingMembers(true);
        try {
            upsertProfiles(addMembersInvitees);
            setProfilesByUserUuid((currentProfilesByUserUuid) => ({
                ...currentProfilesByUserUuid,
                ...profilesByUserUuidFromProfiles(addMembersInvitees),
            }));

            await addChatMembers(apiClient, selectedChat.uuid, {
                memberUuids: addMembersInvitees.map((profile) => profile.userUuid),
            });

            const nextChats = await listChats(apiClient);
            setChats(nextChats);
            closeAddMembersModal();
        } finally {
            setIsAddingMembers(false);
        }
    }

    async function saveChatDetails(payload: { name?: string; avatar?: string; }) {
        if (!selectedChat) {
            return;
        }

        if (payload.name) {
            await renameChat(apiClient, selectedChat.uuid, {
                name: payload.name,
            });
        }

        if (payload.avatar) {
            await changeChatAvatar(apiClient, selectedChat.uuid, {
                avatar: payload.avatar,
            });
        }

        const nextChats = await listChats(apiClient);
        setChats(nextChats);
    }

    async function kickMemberFromChat(memberUuid: string) {
        if (!selectedChat) {
            return;
        }

        setIsKickingMember(true);
        try {
            await kickChatMember(apiClient, selectedChat.uuid, memberUuid);
            const nextChats = await listChats(apiClient);
            setChats(nextChats);
        } finally {
            setIsKickingMember(false);
        }
    }

    function clearSearch() {
        searchInputRef.current?.blur();
        setSearchQuery("");
        setSearchProfiles([]);
        setSearchProfilesError(null);
        setIsSearchFocused(false);
        setIsSearchingProfiles(false);
    }

    function updateLocalLastSeenRevision(chatUuidToUpdate: string, revision: number) {
        const nextRevision = Math.max(localLastSeenRevisionByChatUuidRef.current[chatUuidToUpdate] ?? 0, revision);
        if (localLastSeenRevisionByChatUuidRef.current[chatUuidToUpdate] === nextRevision) {
            return;
        }

        localLastSeenRevisionByChatUuidRef.current = {
            ...localLastSeenRevisionByChatUuidRef.current,
            [chatUuidToUpdate]: nextRevision,
        };
        setLocalLastSeenRevisionByChatUuid((currentLocalLastSeenRevisionByChatUuid) => ({
            ...currentLocalLastSeenRevisionByChatUuid,
            [chatUuidToUpdate]: nextRevision,
        }));
    }

    function handleMessagesScroll() {
        const isScrolledToBottom = isMessagesContainerScrolledToBottom(messagesContainerRef.current);
        isSelectedChatScrolledToBottomRef.current = isScrolledToBottom;

        if (!isScrolledToBottom || !selectedChatUuidRef.current) {
            return;
        }

        if (shouldSkipNextBottomScrollSyncRef.current) {
            shouldSkipNextBottomScrollSyncRef.current = false;
            return;
        }

        void syncChatLastSeenRevisionIfNeeded(selectedChatUuidRef.current);
    }

    function syncDirtyChatsWithKeepalive() {
        const dirtyChats = getDirtyChats(chatsRef.current);
        dirtyChats.forEach((chat) => {
            void syncChatLastSeenRevisionIfNeeded(chat.uuid, { keepalive: true });
        });
    }

    async function syncChatLastSeenRevisionIfNeeded(chatUuidToSync: string, options?: { keepalive?: boolean; }) {
        const chat = chatsRef.current.find((item) => item.uuid === chatUuidToSync);
        if (!chat) {
            return;
        }

        const targetLastSeenRevision = getTargetLastSeenRevision(chat, chatUuidToSync, selectedChatUuidRef.current, isSelectedChatScrolledToBottomRef.current, localLastSeenRevisionByChatUuidRef.current);
        if (targetLastSeenRevision <= chat.currentUserLastSeenRevision) {
            return;
        }

        if (pendingLastSeenRevisionSyncByChatUuidRef.current[chatUuidToSync] === targetLastSeenRevision) {
            return;
        }

        pendingLastSeenRevisionSyncByChatUuidRef.current[chatUuidToSync] = targetLastSeenRevision;

        try {
            await postLastSeenRevision(chatUuidToSync, targetLastSeenRevision, options);
        } catch {
            if (pendingLastSeenRevisionSyncByChatUuidRef.current[chatUuidToSync] === targetLastSeenRevision) {
                delete pendingLastSeenRevisionSyncByChatUuidRef.current[chatUuidToSync];
            }
        }
    }

    async function postLastSeenRevision(chatUuidToSync: string, revision: number, options?: { keepalive?: boolean; }) {
        if (options?.keepalive) {
            const loginToken = getLoginToken();
            if (!loginToken) {
                return;
            }

            await fetch(`/api/v1/message/chats/${chatUuidToSync}/last-seen-revision`, {
                method: "POST",
                keepalive: true,
                headers: {
                    Authorization: `Bearer ${loginToken}`,
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ revision }),
            });
            return;
        }

        await setLastSeenRevision(apiClient, chatUuidToSync, {
            revision,
        });
    }

    function startSidebarResize(event: MouseEvent<HTMLButtonElement>) {
        event.preventDefault();

        sidebarResizeStateRef.current = {
            startX: event.clientX,
            startWidth: sidebarWidth,
        };
        setIsSidebarResizing(true);
    }

    function openMobileChatDetails() {
        if (!isMobileLayout || !selectedChat) {
            return;
        }

        setMobilePanel("details");
    }

    function returnToMobileChat() {
        if (!isMobileLayout || !selectedChat) {
            return;
        }

        setMobilePanel("chat");
    }

    function returnToMobileChatsList() {
        if (!isMobileLayout) {
            return;
        }

        setMobilePanel("list");
        navigate("/chats/@me");
    }

    const mobileLayoutClassName = isMobileLayout ? `chats-me-page--mobile chats-me-page--mobile-${mobilePanel}` : "";
    const rootClassName = [
        "chats-me-page",
        isSidebarResizing ? "chats-me-page--resizing" : "",
        mobileLayoutClassName,
    ].filter(Boolean).join(" ");

    return (
        <section
            className={rootClassName}
            style={{ "--chats-sidebar-width": `${sidebarWidth}px` } as CSSProperties}
        >
            <ChatsSidebar
                authenticatedUserUuid={authenticatedUserUuid}
                chats={chats}
                clearSearch={clearSearch}
                createPrivateChat={createPrivateChat}
                error={error}
                isCreatingChat={isCreatingChat}
                isLoading={isLoading}
                isSearchActive={isSearchActive}
                isSearchFocused={isSearchFocused}
                isSearchingProfiles={isSearchingProfiles}
                localLastSeenRevisionByChatUuid={localLastSeenRevisionByChatUuid}
                openCreateGroupChatModal={openCreateGroupChatModal}
                profilesByUserUuid={profilesByUserUuid}
                searchInputRef={searchInputRef}
                searchProfiles={searchProfiles}
                searchProfilesError={searchProfilesError}
                searchQuery={searchQuery}
                startResize={startSidebarResize}
                setIsSearchFocused={setIsSearchFocused}
                setSearchQuery={setSearchQuery}
            />

            <ChatMessagesPanel
                authenticatedUserUuid={authenticatedUserUuid}
                canDeleteAnyMessage={canDeleteAnyMessage}
                editingMessage={editingMessage}
                handleMessageContextMenu={handleMessageContextMenu}
                isKickingMember={isKickingMember}
                isMessagesLoading={isMessagesLoading}
                isSelectedChatGroup={isSelectedChatGroup}
                isSelectedChatOwnedByAuthenticatedUser={isSelectedChatOwnedByAuthenticatedUser}
                messageContextMenu={messageContextMenu}
                messageViewMode={messageViewMode}
                messageElementByRevisionRef={messageElementByRevisionRef}
                messages={messages}
                messagesContainerRef={messagesContainerRef}
                messagesEndRef={messagesEndRef}
                messagesError={messagesError}
                onCancelEdit={() => setEditingMessage(null)}
                onCloseMessageContextMenu={() => setMessageContextMenu(null)}
                onCopyMessageLink={copyMessageLink}
                onCopyMessageText={copyMessageText}
                onDeleteMessageRequest={(message) => setMessagePendingDeletion(message)}
                onEditMessageRequest={setEditingMessage}
                onKickMemberRequest={(memberUuid) => void kickMemberFromChat(memberUuid)}
                onScroll={handleMessagesScroll}
                onSubmit={submitMessage}
                isMobileLayout={isMobileLayout}
                onOpenChatDetails={openMobileChatDetails}
                onBackToChatsList={returnToMobileChatsList}
                profilesByUserUuid={profilesByUserUuid}
                selectedChat={selectedChat ?? null}
                setMessageViewMode={setMessageViewMode}
            />

            <ChatDetailsPanel
                authenticatedUserUuid={authenticatedUserUuid}
                isKickingMember={isKickingMember}
                isSelectedChatGroup={isSelectedChatGroup}
                isSelectedChatOwnedByAuthenticatedUser={isSelectedChatOwnedByAuthenticatedUser}
                onKickMember={(memberUuid) => void kickMemberFromChat(memberUuid)}
                onOpenAddMembersModal={openAddMembersModal}
                onCloseDetails={returnToMobileChat}
                isMobileLayout={isMobileLayout}
                onSaveChat={saveChatDetails}
                profilesByUserUuid={profilesByUserUuid}
                selectedChat={selectedChat ?? null}
            />

            <DeleteMessageDialog
                isDeletingMessage={isDeletingMessage}
                message={messagePendingDeletion}
                onCancel={() => setMessagePendingDeletion(null)}
                onConfirm={deleteMessage}
                profilesByUserUuid={profilesByUserUuid}
            />

            <GroupMembersDialog
                actionText="Create group"
                busyText="Creating..."
                emptyStateText="No users found."
                error={groupChatSearchProfilesError}
                inputName="group-search"
                inputPlaceholder="Search users"
                invitees={groupChatInvitees}
                isBusy={isCreatingGroupChat}
                isFocused={isGroupChatSearchFocused}
                isOpen={isCreateGroupChatModalOpen}
                isSearching={isSearchingGroupChatProfiles}
                name={groupChatName}
                onClose={closeCreateGroupChatModal}
                onConfirm={createGroupChat}
                onFocusChange={setIsGroupChatSearchFocused}
                onNameChange={setGroupChatName}
                onQueryChange={setGroupChatSearchQuery}
                onToggleInvitee={toggleGroupChatInvitee}
                query={groupChatSearchQuery}
                queryInputRef={groupChatSearchInputRef}
                queryTrimmed={groupChatSearchQueryTrimmed}
                secondaryText="Add people now or create the group first and invite later."
                selectedUserUuids={groupChatInviteeUserUuids}
                title="Create a group chat"
                titleEyebrow="New group"
                visibleProfiles={groupChatVisibleSearchProfiles}
            />

            <GroupMembersDialog
                actionText="Add members"
                busyText="Adding..."
                emptyStateText="No users found."
                error={addMembersSearchProfilesError}
                inputName="members-search"
                inputPlaceholder="Search users"
                invitees={addMembersInvitees}
                isBusy={isAddingMembers}
                isFocused={isAddMembersSearchFocused}
                isOpen={isAddMembersModalOpen && Boolean(selectedChat)}
                isSearching={isSearchingAddMembersProfiles}
                onClose={closeAddMembersModal}
                onConfirm={addMembersToChat}
                onFocusChange={setIsAddMembersSearchFocused}
                onQueryChange={setAddMembersSearchQuery}
                onToggleInvitee={toggleAddMembersInvitee}
                query={addMembersSearchQuery}
                queryInputRef={addMembersSearchInputRef}
                queryTrimmed={addMembersSearchQuery.trim()}
                secondaryText="Pick users to add to this chat. Already selected users stay visible until you uncheck them."
                selectedUserUuids={addMembersInviteeUserUuids}
                title="Add members"
                titleEyebrow="Group members"
                visibleProfiles={addMembersVisibleSearchProfiles}
            />
        </section>
    );
}

function useIsPortraitLayout() {
    const [isPortraitLayout, setIsPortraitLayout] = useState(() => (
        typeof window !== "undefined" && window.innerWidth <= window.innerHeight
    ));

    useEffect(() => {
        const mediaQuery = window.matchMedia("(orientation: portrait)");
        const updateLayout = () => {
            setIsPortraitLayout(mediaQuery.matches);
        };

        updateLayout();

        if (typeof mediaQuery.addEventListener === "function") {
            mediaQuery.addEventListener("change", updateLayout);

            return () => {
                mediaQuery.removeEventListener("change", updateLayout);
            };
        }

        mediaQuery.addListener(updateLayout);

        return () => {
            mediaQuery.removeListener(updateLayout);
        };
    }, []);

    return isPortraitLayout;
}

function copyTextWithFallback(text: string) {
    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    textarea.style.pointerEvents = "none";

    document.body.append(textarea);
    textarea.select();
    document.execCommand("copy");
    textarea.remove();
}

function messageCreatedEventToChatMessage(event: MessageCreatedEvent): ChatMessage {
    return {
        uuid: event.messageUuid,
        chatUuid: event.chatUuid,
        userUuid: event.userUuid,
        revision: event.revision,
        text: event.text,
        attachments: event.attachments ?? [],
        createdAt: event.createdAt,
        updatedAt: event.updatedAt,
    };
}

function getDirtyChats(chats: Chat[]) {
    return chats.filter((chat) => chat.currentUserLastSeenRevision < chat.revision);
}

function isMessagesContainerScrolledToBottom(container: HTMLDivElement | null) {
    if (!container) {
        return false;
    }

    return container.scrollHeight - container.scrollTop - container.clientHeight <= chatBottomThresholdPx;
}

function scrollMessagesContainerToBottom(container: HTMLDivElement | null, messagesEnd: HTMLDivElement | null) {
    if (container) {
        container.scrollTop = container.scrollHeight;
    }

    messagesEnd?.scrollIntoView({ block: "end" });
}

function findFirstVisibleRevision(messages: ChatMessage[], targetRevision: number) {
    const matchingMessage = messages.find((message) => message.revision >= targetRevision);
    return matchingMessage?.revision ?? null;
}

function scrollMessagesContainerToRevision(
    container: HTMLDivElement | null,
    messageElementByRevision: Record<number, HTMLElement | null>,
    revision: number | null,
) {
    if (!container) {
        return;
    }

    if (revision === null) {
        container.scrollTop = container.scrollHeight;
        return;
    }

    const targetMessage = messageElementByRevision[revision];
    if (!targetMessage) {
        container.scrollTop = container.scrollHeight;
        return;
    }

    container.scrollTop = Math.max(0, targetMessage.offsetTop - container.clientHeight / 3);
}

function areLastSeenRevisionMapsEqual(left: Record<string, number>, right: Record<string, number>) {
    const leftKeys = Object.keys(left);
    const rightKeys = Object.keys(right);
    if (leftKeys.length !== rightKeys.length) {
        return false;
    }

    return leftKeys.every((key) => left[key] === right[key]);
}

function confirmPendingMessage(currentMessages: ChatMessage[], pendingMessageUuid: string, confirmedMessageUuid: string) {
    const confirmedMessageExists = currentMessages.some((message) => message.uuid === confirmedMessageUuid && !message.isPending);
    if (confirmedMessageExists) {
        return currentMessages.filter((message) => message.uuid !== pendingMessageUuid);
    }

    let didReplacePendingMessage = false;
    const confirmedMessages = currentMessages.map((message) => {
        if (message.uuid !== pendingMessageUuid) {
            return message;
        }

        didReplacePendingMessage = true;
        return {
            ...message,
            uuid: confirmedMessageUuid,
            isPending: false,
        };
    });

    return didReplacePendingMessage
        ? confirmedMessages
        : currentMessages;
}

function messageEditedEventToChatMessage(event: MessageEditedEvent, revision: number): ChatMessage {
    return {
        uuid: event.messageUuid,
        chatUuid: event.chatUuid,
        userUuid: event.userUuid,
        revision,
        text: event.text,
        attachments: event.attachments ?? [],
        createdAt: event.createdAt,
        updatedAt: event.updatedAt,
    };
}

function generatePendingMessageUuid() {
    return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function loadChatsSidebarWidth() {
    if (typeof window === "undefined") {
        return chatsSidebarDefaultWidthPx;
    }

    const rawValue = window.localStorage.getItem(chatsSidebarWidthStorageKey);
    if (!rawValue) {
        return chatsSidebarDefaultWidthPx;
    }

    const parsedValue = Number(rawValue);
    if (!Number.isFinite(parsedValue)) {
        return chatsSidebarDefaultWidthPx;
    }

    return clampChatsSidebarWidth(parsedValue);
}

function clampChatsSidebarWidth(width: number) {
    return Math.min(chatsSidebarMaxWidthPx, Math.max(chatsSidebarMinWidthPx, Math.round(width)));
}

function loadChatMessageViewMode(): ChatMessageViewMode {
    if (typeof window === "undefined") {
        return "bubble";
    }

    const rawValue = window.localStorage.getItem(chatMessageViewModeStorageKey);
    return rawValue === "plain" ? "plain" : "bubble";
}
