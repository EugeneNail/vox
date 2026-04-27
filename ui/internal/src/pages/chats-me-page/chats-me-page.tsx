import { MouseEvent, useEffect, useLayoutEffect, useRef, useState, type CSSProperties } from "react";
import { NavLink, useNavigate, useParams } from "react-router-dom";
import { getAuthenticatedUserUuid } from "../../auth/auth-tokens";
import MessageComposer from "../../components/message-composer/message-composer";
import { AddMessageEvent, UpdateMessageEvent, useMessageWebSocket } from "../../contexts/message-web-socket-context/message-web-socket-context";
import { useApiClient } from "../../hooks/use-api-client";
import { buildAttachmentUrl, isImageAttachmentName, MessageAttachment } from "../../messages/message-attachments";
import { getCachedProfilesByUserUuids, getMissingOrStaleUserUuids, getStaleCachedUserUuids, PublicProfile, upsertProfiles } from "../../profiles/profile-cache";
import "./chats-me-page.sass";

type Chat = {
    uuid: string;
    name: string | null;
    avatar: string | null;
    isPrivate: boolean;
    createdByUserUuid: string;
    memberUuids: string[];
    createdAt: string;
    updatedAt: string;
};

type ChatMessage = {
    uuid: string;
    chatUuid: string;
    userUuid: string;
    text: string;
    attachments: MessageAttachment[];
    createdAt: string;
    updatedAt: string;
    isPending?: boolean;
};

type MessageContextMenu = {
    message: ChatMessage;
    x: number;
    y: number;
};

const linkPattern = /(https?:\/\/[^\s]+)/g;
const fullLinkPattern = /^https?:\/\/[^\s]+$/;
const messageThreadGapMs = 10 * 60 * 1000;

export default function ChatsMePage() {
    const apiClient = useApiClient();
    const navigate = useNavigate();
    const { addMessageListener, isConnected, removeMessageListener, updateMessageListener } = useMessageWebSocket();
    const { chatUuid } = useParams();
    const messagesEndRef = useRef<HTMLDivElement | null>(null);
    const messageReceivedAudioRef = useRef<HTMLAudioElement | null>(null);
    const searchInputRef = useRef<HTMLInputElement | null>(null);
    const [chats, setChats] = useState<Chat[]>([]);
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isMessagesLoading, setIsMessagesLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [messagesError, setMessagesError] = useState<string | null>(null);
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
    const [profilesByUserUuid, setProfilesByUserUuid] = useState<Record<string, PublicProfile>>({});
    const authenticatedUserUuid = getAuthenticatedUserUuid();
    const selectedChatUuid = chatUuid ?? null;
    const isSearchActive = isSearchFocused || searchQuery.length > 0;

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
        const staleCachedUserUuids = getStaleCachedUserUuids();
        if (staleCachedUserUuids.length === 0) {
            return;
        }

        let isMounted = true;

        async function refreshStaleProfiles() {
            try {
                const { data } = await apiClient.post<PublicProfile[]>("/api/v1/profile/profiles/batch", {
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
        let isMounted = true;

        async function loadChats() {
            try {
                const { data } = await apiClient.get<Chat[]>("/api/v1/message/chats");
                if (!isMounted) {
                    return;
                }

                setChats(data);
            } catch {
                if (!isMounted) {
                    return;
                }

                setError("Could not load chats.");
            } finally {
                if (isMounted) {
                    setIsLoading(false);
                }
            }
        }

        loadChats();

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
                const { data } = await apiClient.post<PublicProfile[]>("/api/v1/profile/profiles/batch", {
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
                const { data } = await apiClient.get<PublicProfile[]>(`/api/v1/profile/search?query=${encodeURIComponent(trimmedQuery)}&limit=10`);
                if (!isMounted) {
                    return;
                }

                const nextProfiles = data.filter((profile) => profile.userUuid !== authenticatedUserUuid);
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
        if (!selectedChatUuid || !isConnected) {
            return;
        }

        void apiClient.post(`/api/v1/message/chats/${selectedChatUuid}/open`).catch((error) => {
            console.error("opening chat view:", error);
        });
    }, [apiClient, isConnected, selectedChatUuid]);

    useEffect(() => {
        if (!selectedChatUuid) {
            setMessages([]);
            return;
        }

        let isMounted = true;

        async function loadMessages() {
            setIsMessagesLoading(true);
            setMessagesError(null);

            try {
                const { data } = await apiClient.get<ChatMessage[]>(`/api/v1/message/chats/${selectedChatUuid}/messages?length=250`);
                if (!isMounted) {
                    return;
                }

                setMessages([...data].reverse().map((message) => ({
                    ...message,
                    attachments: message.attachments ?? [],
                })));
            } catch {
                if (!isMounted) {
                    return;
                }

                setMessages([]);
                setMessagesError("Could not load messages.");
            } finally {
                if (isMounted) {
                    setIsMessagesLoading(false);
                }
            }
        }

        loadMessages();

        return () => {
            isMounted = false;
        };
    }, [apiClient, selectedChatUuid]);

    useLayoutEffect(() => {
        if (!selectedChatUuid || isMessagesLoading || messages.length === 0) {
            return;
        }

        messagesEndRef.current?.scrollIntoView({ block: "end" });
    }, [isMessagesLoading, messages.length, selectedChatUuid]);

    useEffect(() => {
        setEditingMessage(null);
        setMessageContextMenu(null);
        setMessagePendingDeletion(null);
    }, [selectedChatUuid]);

    useEffect(() => (
        addMessageListener((event) => {
            if (event.chatUuid !== selectedChatUuid) {
                return;
            }

            if (event.userUuid !== authenticatedUserUuid) {
                void messageReceivedAudioRef.current?.play();
            }

            setMessages((currentMessages) => {
                if (currentMessages.some((message) => message.uuid === event.messageUuid && !message.isPending)) {
                    return currentMessages;
                }

                const confirmedMessage = addMessageEventToChatMessage(event);
                return [
                    ...currentMessages,
                    confirmedMessage,
                ];
            });
        })
    ), [addMessageListener, selectedChatUuid]);

    useEffect(() => (
        updateMessageListener((event) => {
            if (event.chatUuid !== selectedChatUuid) {
                return;
            }

            setMessages((currentMessages) => currentMessages.map((message) => (
                message.uuid === event.messageUuid
                    ? updateMessageEventToChatMessage(event)
                    : message
            )));
        })
    ), [selectedChatUuid, updateMessageListener]);

    useEffect(() => (
        removeMessageListener((event) => {
            if (event.chatUuid !== selectedChatUuid) {
                return;
            }

            setMessages((currentMessages) => currentMessages.filter((message) => message.uuid !== event.messageUuid));
        })
    ), [removeMessageListener, selectedChatUuid]);

    useEffect(() => {
        function handleWindowClick() {
            setMessageContextMenu(null);
        }

        function handleWindowKeyDown(event: KeyboardEvent) {
            if (event.key === "Escape") {
                setMessageContextMenu(null);
                setMessagePendingDeletion(null);
            }
        }

        window.addEventListener("click", handleWindowClick);
        window.addEventListener("keydown", handleWindowKeyDown);

        return () => {
            window.removeEventListener("click", handleWindowClick);
            window.removeEventListener("keydown", handleWindowKeyDown);
        };
    }, []);

    const selectedChat = chats.find((chat) => chat.uuid === selectedChatUuid);

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
                text,
                attachments: pendingAttachments,
                createdAt,
                updatedAt: createdAt,
                isPending: true,
            },
        ]);

        try {
            const { data: messageUuid } = await apiClient.post<string>(`/api/v1/message/chats/${selectedChat.uuid}/messages`, {
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
        await apiClient.put<string>(`/api/v1/message/messages/${message.uuid}`, {
            text,
            attachments,
        });

        setEditingMessage(null);
    }

    async function deleteMessage(message: ChatMessage) {
        setIsDeletingMessage(true);
        try {
            await apiClient.delete<string>(`/api/v1/message/messages/${message.uuid}`);
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

            const { data: createdChatUuid } = await apiClient.post<string>("/api/v1/message/chats", {
                memberUuids: [profile.userUuid],
                name: null,
                avatar: null,
                isPrivate: true,
            });

            const { data: nextChats } = await apiClient.get<Chat[]>("/api/v1/message/chats");
            setChats(nextChats);
            clearSearch();
            navigate(`/chats/@me/${createdChatUuid}`);
        } finally {
            setIsCreatingChat(false);
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

    return (
        <section className="chats-me-page">
            <aside className="chats-me-page__sidebar" aria-label="Chats">
                <div className="chats-me-page__sidebar-header">
                    <p className="chats-me-page__eyebrow">Messages</p>
                    <h1 className="chats-me-page__title">Chats</h1>
                    <label className="chats-me-page__search">
                        <span className="material-symbols-rounded chats-me-page__search-icon" aria-hidden="true">search</span>
                        <input
                            className="chats-me-page__search-input"
                            type="text"
                            name="search"
                            placeholder="Search users"
                            ref={searchInputRef}
                            value={searchQuery}
                            onChange={(event) => setSearchQuery(event.target.value)}
                            onFocus={() => setIsSearchFocused(true)}
                            onBlur={() => setIsSearchFocused(false)}
                        />
                        {isSearchActive && (
                            <button
                                className="chats-me-page__search-clear"
                                type="button"
                                onMouseDown={(event) => event.preventDefault()}
                                onClick={(event) => {
                                    event.preventDefault();
                                    event.stopPropagation();
                                    clearSearch();
                                }}
                            >
                                <span className="material-symbols-rounded" aria-hidden="true">close</span>
                            </button>
                        )}
                    </label>
                </div>

                <div className="chats-me-page__list">
                    {isSearchActive ? (
                        <>
                            {searchQuery.trim().length === 0 && (
                                <p className="chats-me-page__state">Type a name.</p>
                            )}
                            {isSearchingProfiles && <p className="chats-me-page__state">Searching users...</p>}
                            {searchProfilesError && <p className="chats-me-page__state chats-me-page__state--error">{searchProfilesError}</p>}
                            {!isSearchingProfiles && !searchProfilesError && searchQuery.trim().length > 0 && searchProfiles.length === 0 && (
                                <p className="chats-me-page__state">No users found.</p>
                            )}
                            {searchProfiles.map((profile) => (
                                <button
                                    className="chats-me-page__chat-button chats-me-page__chat-button--search-result"
                                    disabled={isCreatingChat}
                                    key={profile.userUuid}
                                    type="button"
                                    onClick={() => void createPrivateChat(profile)}
                                >
                                    <UserAvatar
                                        className="chats-me-page__avatar"
                                        src={getProfileAvatarUrl(profile)}
                                        label={profile.name}
                                    />
                                    <span className="chats-me-page__chat-preview">
                                        <span className="chats-me-page__chat-name">{profile.name}</span>
                                    </span>
                                </button>
                            ))}
                        </>
                    ) : (
                        <>
                            {isLoading && <p className="chats-me-page__state">Loading chats...</p>}
                            {error && <p className="chats-me-page__state chats-me-page__state--error">{error}</p>}
                            {!isLoading && !error && chats.length === 0 && (
                                <p className="chats-me-page__state">No chats yet.</p>
                            )}
                            {chats.map((chat) => (
                                <NavLink
                                    className={
                                        ({ isActive }) => (
                                            isActive
                                                ? "chats-me-page__chat-button chats-me-page__chat-button--active"
                                                : "chats-me-page__chat-button"
                                        )
                                    }
                                    key={chat.uuid}
                                    to={`/chats/@me/${chat.uuid}`}
                                >
                                    <UserAvatar
                                        className="chats-me-page__avatar"
                                        src={getChatAvatarUrl(chat, authenticatedUserUuid, profilesByUserUuid)}
                                        label={getChatAvatarLabel(chat, authenticatedUserUuid, profilesByUserUuid)}
                                    />
                                    <span className="chats-me-page__chat-name">{getChatTitle(chat, authenticatedUserUuid, profilesByUserUuid)}</span>
                                </NavLink>
                            ))}
                        </>
                    )}
                </div>
            </aside>

            <section className="chats-me-page__chat" aria-label="Chat">
                {selectedChat ? (
                    <div className="chats-me-page__chat-shell">
                        <header className="chats-me-page__chat-header">
                            <p className="chats-me-page__eyebrow">Selected chat</p>
                            <h2 className="chats-me-page__chat-title">{getChatTitle(selectedChat, authenticatedUserUuid, profilesByUserUuid)}</h2>
                        </header>

                        <div className="chats-me-page__messages" aria-live="polite" onContextMenu={(event) => event.preventDefault()}>
                            {isMessagesLoading && <p className="chats-me-page__state">Loading messages...</p>}
                            {messagesError && <p className="chats-me-page__state chats-me-page__state--error">{messagesError}</p>}
                            {!isMessagesLoading && !messagesError && messages.length === 0 && (
                                <p className="chats-me-page__state">No messages yet.</p>
                            )}
                            {messages.map((message, index) => {
                                const isThreadStart = isMessageThreadStart(message, messages[index - 1]);

                                return (
                                    <article
                                        className={
                                            [
                                                "chats-me-page__message",
                                                isThreadStart ? "chats-me-page__message--thread-start" : "",
                                                message.isPending ? "chats-me-page__message--pending" : "",
                                                editingMessage?.uuid === message.uuid ? "chats-me-page__message--editing" : "",
                                            ].filter(Boolean).join(" ")
                                        }
                                        key={message.uuid}
                                        onContextMenu={(event) => handleMessageContextMenu(event, message)}
                                    >
                                        <div className="chats-me-page__message-avatar-cell">
                                            {isThreadStart && (
                                                <UserAvatar
                                                    className="chats-me-page__message-avatar"
                                                    src={getMessageAuthorAvatarUrl(message, profilesByUserUuid)}
                                                    label={getMessageAuthorAvatarLabel(message, profilesByUserUuid)}
                                                />
                                            )}
                                        </div>
                                        <div className="chats-me-page__message-body">
                                            {isThreadStart && (
                                                <div className="chats-me-page__message-header">
                                                    <p className="chats-me-page__message-author">{getMessageAuthorName(message, profilesByUserUuid)}</p>
                                                </div>
                                            )}
                                            {hasRenderableMessageText(message.text) && (
                                                <p className="chats-me-page__message-text">
                                                    {renderMessageText(message.text)}
                                                    {isMessageEdited(message) && (
                                                        <span className="chats-me-page__message-edited"> (edited)</span>
                                                    )}
                                                </p>
                                            )}
                                            {message.attachments.length > 0 && (
                                                <div className="chats-me-page__message-attachments" aria-label="Message attachments">
                                                    {message.attachments.map((attachment) => (
                                                        <MessageAttachmentView attachment={attachment} key={attachment.uuid} />
                                                    ))}
                                                </div>
                                            )}
                                        </div>
                                        <div className="chats-me-page__message-status">
                                            {isOwnConfirmedMessage(message, authenticatedUserUuid) && (
                                                <span className="material-symbols-rounded chats-me-page__message-delivery-icon" aria-label="Delivered">done</span>
                                            )}
                                            <time
                                                className="chats-me-page__message-time"
                                                dateTime={message.createdAt}
                                            >
                                                {formatMessageTime(message.createdAt)}
                                            </time>
                                        </div>
                                    </article>
                                );
                            })}
                            <div ref={messagesEndRef} className="chats-me-page__messages-end" aria-hidden="true" />
                        </div>

                        {messageContextMenu && (
                            <div
                                className="chats-me-page__context-menu"
                                style={{
                                    left: messageContextMenu.x,
                                    top: messageContextMenu.y,
                                }}
                                onClick={(event) => event.stopPropagation()}
                            >
                                {messageContextMenu.message.userUuid === authenticatedUserUuid && (
                                    <>
                                        <button
                                            className="chats-me-page__context-menu-button"
                                            type="button"
                                            onClick={() => {
                                                setEditingMessage(messageContextMenu.message);
                                                setMessageContextMenu(null);
                                            }}
                                        >
                                            <span className="material-symbols-rounded" aria-hidden="true">edit</span>
                                            Edit Message
                                        </button>
                                    </>
                                )}
                                <button
                                    className="chats-me-page__context-menu-button"
                                    type="button"
                                    onClick={() => void copyMessageText(messageContextMenu.message)}
                                >
                                    <span className="material-symbols-rounded" aria-hidden="true">content_copy</span>
                                    Copy text
                                </button>
                                <button
                                    className="chats-me-page__context-menu-button"
                                    type="button"
                                    onClick={() => void copyMessageLink(messageContextMenu.message)}
                                >
                                    <span className="material-symbols-rounded" aria-hidden="true">link</span>
                                    Copy Message Link
                                </button>
                                {messageContextMenu.message.userUuid === authenticatedUserUuid && (
                                    <div className="chats-me-page__context-menu-danger">
                                        <button
                                            className="chats-me-page__context-menu-button chats-me-page__context-menu-button--danger"
                                            type="button"
                                            onClick={() => {
                                                setMessagePendingDeletion(messageContextMenu.message);
                                                setMessageContextMenu(null);
                                            }}
                                        >
                                            <span className="material-symbols-rounded" aria-hidden="true">delete</span>
                                            Delete Message
                                        </button>
                                    </div>
                                )}
                            </div>
                        )}

                        <MessageComposer
                            disabled={!authenticatedUserUuid}
                            editingText={editingMessage?.text ?? null}
                            editingAttachments={editingMessage?.attachments ?? []}
                            onCancelEdit={() => setEditingMessage(null)}
                            onSubmit={submitMessage}
                        />
                    </div>
                ) : (
                    <div className="chats-me-page__chat-placeholder">
                        <p className="chats-me-page__eyebrow">No chat selected</p>
                        <h2 className="chats-me-page__chat-title">Choose a chat</h2>
                    </div>
                )}
            </section>

            <aside className="chats-me-page__details" aria-label="Chat details" />

            {messagePendingDeletion && (
                <div className="chats-me-page__modal-backdrop" role="presentation">
                    <section className="chats-me-page__delete-dialog" role="dialog" aria-modal="true" aria-labelledby="delete-message-title">
                        <div className="chats-me-page__delete-dialog-header">
                            <h2 className="chats-me-page__delete-dialog-title" id="delete-message-title">Delete Message</h2>
                            <p className="chats-me-page__delete-dialog-text">Are you sure you want to delete this message?</p>
                        </div>

                        <article className="chats-me-page__delete-message-preview">
                            <div className="chats-me-page__message-avatar-cell">
                                <UserAvatar
                                    className="chats-me-page__message-avatar"
                                    src={getMessageAuthorAvatarUrl(messagePendingDeletion, profilesByUserUuid)}
                                    label={getMessageAuthorAvatarLabel(messagePendingDeletion, profilesByUserUuid)}
                                />
                            </div>
                            <div className="chats-me-page__message-body">
                                <div className="chats-me-page__message-header">
                                    <p className="chats-me-page__message-author">{getMessageAuthorName(messagePendingDeletion, profilesByUserUuid)}</p>
                                </div>
                                {hasRenderableMessageText(messagePendingDeletion.text) && (
                                    <p className="chats-me-page__message-text">
                                        {renderMessageText(messagePendingDeletion.text)}
                                        {isMessageEdited(messagePendingDeletion) && (
                                            <span className="chats-me-page__message-edited"> (edited)</span>
                                        )}
                                    </p>
                                )}
                                {messagePendingDeletion.attachments.length > 0 && (
                                    <div className="chats-me-page__message-attachments" aria-label="Message attachments">
                                        {messagePendingDeletion.attachments.map((attachment) => (
                                            <MessageAttachmentView attachment={attachment} key={attachment.uuid} />
                                        ))}
                                    </div>
                                )}
                            </div>
                            <div className="chats-me-page__message-status">
                                <time
                                    className="chats-me-page__message-time"
                                    dateTime={messagePendingDeletion.createdAt}
                                >
                                    {formatMessageTime(messagePendingDeletion.createdAt)}
                                </time>
                            </div>
                        </article>

                        <div className="chats-me-page__delete-dialog-actions">
                            <button className="chats-me-page__delete-dialog-button" type="button" disabled={isDeletingMessage} onClick={() => setMessagePendingDeletion(null)}>
                                Cancel
                            </button>
                            <button className="chats-me-page__delete-dialog-button chats-me-page__delete-dialog-button--danger" type="button" disabled={isDeletingMessage} onClick={() => void deleteMessage(messagePendingDeletion)}>
                                Delete
                            </button>
                        </div>
                    </section>
                </div>
            )}
        </section>
    );
}

function getChatTitle(chat: Chat, authenticatedUserUuid: string | null, profilesByUserUuid: Record<string, PublicProfile>) {
    if (chat.name) {
        return chat.name;
    }

    if (!chat.isPrivate) {
        return chat.uuid;
    }

    const companionUuid = chat.memberUuids.find((memberUuid) => memberUuid !== authenticatedUserUuid);
    if (!companionUuid) {
        return chat.uuid;
    }

    return profilesByUserUuid[companionUuid]?.name ?? companionUuid;
}

function getChatAvatarUrl(chat: Chat, authenticatedUserUuid: string | null, profilesByUserUuid: Record<string, PublicProfile>) {
    if (chat.isPrivate) {
        const companionUuid = chat.memberUuids.find((memberUuid) => memberUuid !== authenticatedUserUuid);
        if (!companionUuid) {
            return null;
        }

        return getProfileAvatarUrl(profilesByUserUuid[companionUuid] ?? null);
    }

    if (!chat.avatar) {
        return null;
    }

    return buildAttachmentUrl(chat.avatar);
}

function getChatAvatarLabel(chat: Chat, authenticatedUserUuid: string | null, profilesByUserUuid: Record<string, PublicProfile>) {
    if (chat.isPrivate) {
        const companionUuid = chat.memberUuids.find((memberUuid) => memberUuid !== authenticatedUserUuid);
        if (!companionUuid) {
            return chat.uuid;
        }

        return profilesByUserUuid[companionUuid]?.name ?? companionUuid;
    }

    return getChatTitle(chat, authenticatedUserUuid, profilesByUserUuid);
}

function formatMessageTime(date: string) {
    return new Intl.DateTimeFormat(undefined, {
        hour: "2-digit",
        hour12: false,
        hourCycle: "h23",
        minute: "2-digit",
    }).format(new Date(date));
}

function isMessageThreadStart(message: ChatMessage, previousMessage?: ChatMessage) {
    if (!previousMessage) {
        return true;
    }

    if (previousMessage.userUuid !== message.userUuid) {
        return true;
    }

    return new Date(message.createdAt).getTime() - new Date(previousMessage.createdAt).getTime() >= messageThreadGapMs;
}

function isMessageEdited(message: ChatMessage) {
    return new Date(message.createdAt).getTime() !== new Date(message.updatedAt).getTime();
}

function isOwnConfirmedMessage(message: ChatMessage, authenticatedUserUuid: string | null) {
    return !message.isPending && message.userUuid === authenticatedUserUuid;
}

function getMessageAuthorName(message: ChatMessage, profilesByUserUuid: Record<string, PublicProfile>) {
    return profilesByUserUuid[message.userUuid]?.name ?? message.userUuid;
}

function getMessageAuthorAvatarUrl(message: ChatMessage, profilesByUserUuid: Record<string, PublicProfile>) {
    return getProfileAvatarUrl(profilesByUserUuid[message.userUuid] ?? null);
}

function getMessageAuthorAvatarLabel(message: ChatMessage, profilesByUserUuid: Record<string, PublicProfile>) {
    return getMessageAuthorName(message, profilesByUserUuid);
}

function renderMessageText(text: string) {
    return text.split(linkPattern).map((part, index) => {
        if (!fullLinkPattern.test(part)) {
            return part;
        }

        return (
            <a className="chats-me-page__message-link" href={part} key={`${part}-${index}`} target="_blank" rel="noreferrer">
                {part}
            </a>
        );
    });
}

function hasRenderableMessageText(text: string) {
    return text.trim().length > 0;
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

function addMessageEventToChatMessage(event: AddMessageEvent): ChatMessage {
    return {
        uuid: event.messageUuid,
        chatUuid: event.chatUuid,
        userUuid: event.userUuid,
        text: event.text,
        attachments: event.attachments ?? [],
        createdAt: event.createdAt,
        updatedAt: event.updatedAt,
    };
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

function updateMessageEventToChatMessage(event: UpdateMessageEvent): ChatMessage {
    return {
        uuid: event.messageUuid,
        chatUuid: event.chatUuid,
        userUuid: event.userUuid,
        text: event.text,
        attachments: event.attachments ?? [],
        createdAt: event.createdAt,
        updatedAt: event.updatedAt,
    };
}

function MessageAttachmentView({ attachment }: { attachment: MessageAttachment; }) {
    if (isImageAttachmentName(attachment.name)) {
        const url = buildAttachmentUrl(attachment.name);
        return (
            <a className="chats-me-page__message-attachment chats-me-page__message-attachment--image" href={url} target="_blank" rel="noreferrer">
                <img className="chats-me-page__message-attachment-image" src={url} alt={attachment.name} />
            </a>
        );
    }

    return (
        <a className="chats-me-page__message-attachment chats-me-page__message-attachment--document" href={buildAttachmentUrl(attachment.name)} target="_blank" rel="noreferrer">
            <span className="material-symbols-rounded chats-me-page__message-attachment-icon" aria-hidden="true">description</span>
            <span className="chats-me-page__message-attachment-name">{attachment.name}</span>
        </a>
    );
}

function getProfileAvatarUrl(profile: PublicProfile | null | undefined) {
    if (!profile?.avatar) {
        return null;
    }

    return buildAttachmentUrl(profile.avatar);
}

function UserAvatar({ className, src, label }: { className: string; src: string | null; label: string; }) {
    if (src) {
        return (
            <img
                className={`${className} ${className}--image`}
                src={src}
                alt=""
                aria-hidden="true"
                loading="lazy"
                decoding="async"
            />
        );
    }

    const placeholder = getAvatarPlaceholder(label);

    return (
        <span
            className={`${className} ${className}--placeholder`}
            style={placeholder.style}
            aria-hidden="true"
        >
            {placeholder.initials}
        </span>
    );
}

function getAvatarPlaceholder(label: string) {
    const initials = getAvatarInitials(label);
    const backgroundColor = hashToAvatarColor(label);

    return {
        initials,
        style: {
            backgroundColor,
            borderColor: "rgba(255, 255, 255, 0.16)",
            color: "#f8fafc",
        } satisfies CSSProperties,
    };
}

function getAvatarInitials(label: string) {
    const normalizedLabel = label.trim().replace(/\s+/g, " ");
    if (normalizedLabel.length === 0) {
        return "?";
    }

    const words = normalizedLabel.split(" ").filter(Boolean);
    if (words.length >= 2) {
        return `${firstAlphabeticCharacter(words[0])}${firstAlphabeticCharacter(words[1])}`.toUpperCase();
    }

    return normalizedLabel.slice(0, 2).toUpperCase();
}

function firstAlphabeticCharacter(value: string) {
    const match = value.match(/[a-zа-я0-9]/i);
    return match?.[0] ?? value[0] ?? "?";
}

function hashToAvatarColor(value: string) {
    const normalizedValue = value.trim().toLowerCase();
    let hash = 0;

    for (let index = 0; index < normalizedValue.length; index += 1) {
        hash = ((hash << 5) - hash + normalizedValue.charCodeAt(index)) | 0;
    }

    const hue = Math.abs(hash) % 360;
    return hslToHex(hue, 0.56, 0.46);
}

function hslToHex(hue: number, saturation: number, lightness: number) {
    const chroma = (1 - Math.abs(2 * lightness - 1)) * saturation;
    const hueSection = hue / 60;
    const secondary = chroma * (1 - Math.abs((hueSection % 2) - 1));
    let red = 0;
    let green = 0;
    let blue = 0;

    if (hueSection >= 0 && hueSection < 1) {
        red = chroma;
        green = secondary;
    } else if (hueSection < 2) {
        red = secondary;
        green = chroma;
    } else if (hueSection < 3) {
        green = chroma;
        blue = secondary;
    } else if (hueSection < 4) {
        green = secondary;
        blue = chroma;
    } else if (hueSection < 5) {
        red = secondary;
        blue = chroma;
    } else {
        red = chroma;
        blue = secondary;
    }

    const matchValue = lightness - chroma / 2;
    const toHexByte = (channel: number) => Math.round((channel + matchValue) * 255).toString(16).padStart(2, "0");

    return `#${toHexByte(red)}${toHexByte(green)}${toHexByte(blue)}`;
}

function profilesByUserUuidFromProfiles(profiles: PublicProfile[]) {
    return profiles.reduce<Record<string, PublicProfile>>((accumulator, profile) => {
        accumulator[profile.userUuid] = profile;
        return accumulator;
    }, {});
}

function collectReferencedUserUuids(chats: Chat[], messages: ChatMessage[], searchProfiles: PublicProfile[]) {
    return Array.from(new Set([
        ...chats.flatMap((chat) => chat.memberUuids),
        ...messages.map((message) => message.userUuid),
        ...searchProfiles.map((profile) => profile.userUuid),
    ]));
}

function generatePendingMessageUuid() {
    return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}
