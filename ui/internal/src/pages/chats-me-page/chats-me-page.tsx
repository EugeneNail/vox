import { MouseEvent, useEffect, useLayoutEffect, useRef, useState } from "react";
import { NavLink, useParams } from "react-router-dom";
import { getAuthenticatedUserUuid } from "../../auth/auth-tokens";
import MessageComposer from "../../components/message-composer/message-composer";
import { AddMessageEvent, UpdateMessageEvent, useMessageWebSocket } from "../../contexts/message-web-socket-context/message-web-socket-context";
import { useApiClient } from "../../hooks/use-api-client";
import "./chats-me-page.sass";

type Chat = {
    uuid: string;
    name: string | null;
    avatar: string | null;
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
    createdAt: string;
    updatedAt: string;
    isPending?: boolean;
    pendingServerUuid?: string;
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
    const { addMessageListener, removeMessageListener, subscribeChat, unsubscribeChat, updateMessageListener } = useMessageWebSocket();
    const { chatUuid } = useParams();
    const messagesEndRef = useRef<HTMLDivElement | null>(null);
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
    const authenticatedUserUuid = getAuthenticatedUserUuid();
    const selectedChatUuid = chatUuid ?? null;

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

                setMessages([...data].reverse());
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

    useEffect(() => {
        if (!selectedChatUuid) {
            return;
        }

        subscribeChat(selectedChatUuid);

        return () => {
            unsubscribeChat(selectedChatUuid);
        };
    }, [selectedChatUuid, subscribeChat, unsubscribeChat]);

    useEffect(() => (
        addMessageListener((event) => {
            if (event.chatUuid !== selectedChatUuid) {
                return;
            }

            setMessages((currentMessages) => {
                if (currentMessages.some((message) => message.uuid === event.messageUuid && !message.isPending)) {
                    return currentMessages;
                }

                const confirmedMessage = addMessageEventToChatMessage(event);
                const pendingMessageIndex = currentMessages.findIndex((message) => (
                    message.isPending &&
                    (
                        message.pendingServerUuid === event.messageUuid ||
                        (
                            message.chatUuid === event.chatUuid &&
                            message.userUuid === event.userUuid &&
                            message.text === event.text
                        )
                    )
                ));

                if (pendingMessageIndex !== -1) {
                    return currentMessages.map((message, index) => (
                        index === pendingMessageIndex
                            ? confirmedMessage
                            : message
                    ));
                }

                // TODO: Sort confirmed messages by server createdAt while keeping pending messages stable.
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

    async function submitMessage(text: string) {
        if (editingMessage) {
            await editMessage(editingMessage, text);
            return;
        }

        await sendMessage(text);
    }

    async function sendMessage(text: string) {
        if (!selectedChat || !authenticatedUserUuid) {
            return;
        }

        const pendingMessageUuid = `pending-${crypto.randomUUID()}`;
        const createdAt = new Date().toISOString();

        setMessages((currentMessages) => [
            ...currentMessages,
            {
                uuid: pendingMessageUuid,
                chatUuid: selectedChat.uuid,
                userUuid: authenticatedUserUuid,
                text,
                createdAt,
                updatedAt: createdAt,
                isPending: true,
            },
        ]);

        try {
            const { data: messageUuid } = await apiClient.post<string>(`/api/v1/message/chats/${selectedChat.uuid}/messages`, {
                text,
            });

            setMessages((currentMessages) => {
                if (currentMessages.some((message) => message.uuid === messageUuid && !message.isPending)) {
                    return currentMessages.filter((message) => message.uuid !== pendingMessageUuid);
                }

                return currentMessages.map((message) => (
                    message.uuid === pendingMessageUuid
                        ? { ...message, uuid: messageUuid, pendingServerUuid: messageUuid }
                        : message
                ));
            });
        } catch (error) {
            setMessages((currentMessages) => currentMessages.filter((message) => message.uuid !== pendingMessageUuid));
            throw error;
        }
    }

    async function editMessage(message: ChatMessage, text: string) {
        await apiClient.put<string>(`/api/v1/message/messages/${message.uuid}`, {
            text,
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

    return (
        <section className="chats-me-page">
            <aside className="chats-me-page__sidebar" aria-label="Chats">
                <div className="chats-me-page__sidebar-header">
                    <p className="chats-me-page__eyebrow">Messages</p>
                    <h1 className="chats-me-page__title">Chats</h1>
                </div>

                <div className="chats-me-page__list">
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
                            <span className="chats-me-page__avatar" aria-hidden="true" />
                            <span className="chats-me-page__chat-name">{getChatTitle(chat, authenticatedUserUuid)}</span>
                        </NavLink>
                    ))}
                </div>
            </aside>

            <section className="chats-me-page__chat" aria-label="Chat">
                {selectedChat ? (
                    <div className="chats-me-page__chat-shell">
                        <header className="chats-me-page__chat-header">
                            <p className="chats-me-page__eyebrow">Selected chat</p>
                            <h2 className="chats-me-page__chat-title">{getChatTitle(selectedChat, authenticatedUserUuid)}</h2>
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
                                            {isThreadStart && <span className="chats-me-page__message-avatar" aria-hidden="true" />}
                                        </div>
                                        <div className="chats-me-page__message-body">
                                            {isThreadStart && (
                                                <div className="chats-me-page__message-header">
                                                    <p className="chats-me-page__message-author">{message.userUuid}</p>
                                                </div>
                                            )}
                                            <p className="chats-me-page__message-text">
                                                {renderMessageText(message.text)}
                                                {isMessageEdited(message) && (
                                                    <span className="chats-me-page__message-edited"> (edited)</span>
                                                )}
                                            </p>
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
                                <span className="chats-me-page__message-avatar" aria-hidden="true" />
                            </div>
                            <div className="chats-me-page__message-body">
                                <div className="chats-me-page__message-header">
                                    <p className="chats-me-page__message-author">{messagePendingDeletion.userUuid}</p>
                                </div>
                                <p className="chats-me-page__message-text">
                                    {renderMessageText(messagePendingDeletion.text)}
                                    {isMessageEdited(messagePendingDeletion) && (
                                        <span className="chats-me-page__message-edited"> (edited)</span>
                                    )}
                                </p>
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

function getChatTitle(chat: Chat, authenticatedUserUuid: string | null) {
    if (chat.name) {
        return chat.name;
    }

    const companionUuid = chat.memberUuids.find((memberUuid) => memberUuid !== authenticatedUserUuid);
    return companionUuid ?? chat.uuid;
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
        createdAt: event.createdAt,
        updatedAt: event.updatedAt,
    };
}

function updateMessageEventToChatMessage(event: UpdateMessageEvent): ChatMessage {
    return {
        uuid: event.messageUuid,
        chatUuid: event.chatUuid,
        userUuid: event.userUuid,
        text: event.text,
        createdAt: event.createdAt,
        updatedAt: event.updatedAt,
    };
}
