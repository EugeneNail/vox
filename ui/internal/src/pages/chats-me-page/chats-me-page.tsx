import { useEffect, useState } from "react";
import { NavLink, useParams } from "react-router-dom";
import { getAuthenticatedUserUuid } from "../../auth/auth-tokens";
import MessageComposer from "../../components/message-composer/message-composer";
import { useApiClient } from "../../hooks/use-api-client";
import "./chats-me-page.sass";

type DirectChat = {
    uuid: string;
    firstMemberUuid: string;
    secondMemberUuid: string;
};

type ChatMessage = {
    uuid: string;
    chatUuid: string;
    userUuid: string;
    text: string;
    createdAt: string;
    updatedAt: string;
};

const linkPattern = /(https?:\/\/[^\s]+)/g;
const fullLinkPattern = /^https?:\/\/[^\s]+$/;
const messageThreadGapMs = 10 * 60 * 1000;

export default function ChatsMePage() {
    const apiClient = useApiClient();
    const { directChatUuid } = useParams();
    const [directChats, setDirectChats] = useState<DirectChat[]>([]);
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isMessagesLoading, setIsMessagesLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [messagesError, setMessagesError] = useState<string | null>(null);
    const authenticatedUserUuid = getAuthenticatedUserUuid();
    const selectedChatUuid = directChatUuid ?? null;

    useEffect(() => {
        let isMounted = true;

        async function loadDirectChats() {
            try {
                const { data } = await apiClient.get<DirectChat[]>("/api/v1/message/direct-chats");
                if (!isMounted) {
                    return;
                }

                setDirectChats(data);
            } catch {
                if (!isMounted) {
                    return;
                }

                setError("Could not load direct chats.");
            } finally {
                if (isMounted) {
                    setIsLoading(false);
                }
            }
        }

        loadDirectChats();

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
                const { data } = await apiClient.get<ChatMessage[]>(`/api/v1/message/direct-chats/${selectedChatUuid}/messages?length=250`);
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

    const selectedChat = directChats.find((chat) => chat.uuid === selectedChatUuid);

    async function sendMessage(text: string) {
        if (!selectedChat || !authenticatedUserUuid) {
            return;
        }

        const { data: messageUuid } = await apiClient.post<string>(`/api/v1/message/chats/${selectedChat.uuid}/messages`, {
            text,
        });

        const createdAt = new Date().toISOString();
        setMessages((currentMessages) => [
            ...currentMessages,
            {
                uuid: messageUuid,
                chatUuid: selectedChat.uuid,
                userUuid: authenticatedUserUuid,
                text,
                createdAt,
                updatedAt: createdAt,
            },
        ]);
    }

    return (
        <section className="chats-me-page">
            <aside className="chats-me-page__sidebar" aria-label="Direct chats">
                <div className="chats-me-page__sidebar-header">
                    <p className="chats-me-page__eyebrow">Direct messages</p>
                    <h1 className="chats-me-page__title">Chats</h1>
                </div>

                <div className="chats-me-page__list">
                    {isLoading && <p className="chats-me-page__state">Loading chats...</p>}
                    {error && <p className="chats-me-page__state chats-me-page__state--error">{error}</p>}
                    {!isLoading && !error && directChats.length === 0 && (
                        <p className="chats-me-page__state">No direct chats yet.</p>
                    )}
                    {directChats.map((chat) => (
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
                            <span className="chats-me-page__chat-name">{getDirectChatTitle(chat, authenticatedUserUuid)}</span>
                        </NavLink>
                    ))}
                </div>
            </aside>

            <section className="chats-me-page__chat" aria-label="Chat">
                {selectedChat ? (
                    <div className="chats-me-page__chat-shell">
                        <div className="chats-me-page__messages-panel">
                            <p className="chats-me-page__eyebrow">Selected chat</p>
                            <h2 className="chats-me-page__chat-title">{getDirectChatTitle(selectedChat, authenticatedUserUuid)}</h2>

                            <div className="chats-me-page__messages" aria-live="polite">
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
                                                ].filter(Boolean).join(" ")
                                            }
                                            key={message.uuid}
                                        >
                                            <div className="chats-me-page__message-avatar-cell">
                                                {isThreadStart && <span className="chats-me-page__message-avatar" aria-hidden="true" />}
                                            </div>
                                            <div className="chats-me-page__message-body">
                                                {isThreadStart && (
                                                    <div className="chats-me-page__message-header">
                                                        <p className="chats-me-page__message-author">{message.userUuid}</p>
                                                        <time className="chats-me-page__message-time" dateTime={message.createdAt}>
                                                            {formatMessageTime(message.createdAt)}
                                                        </time>
                                                    </div>
                                                )}
                                                <p className="chats-me-page__message-text">{renderMessageText(message.text)}</p>
                                            </div>
                                        </article>
                                    );
                                })}
                            </div>
                        </div>

                        <MessageComposer disabled={!authenticatedUserUuid} onSubmit={sendMessage} />
                    </div>
                ) : (
                    <div className="chats-me-page__chat-placeholder">
                        <p className="chats-me-page__eyebrow">No chat selected</p>
                        <h2 className="chats-me-page__chat-title">Choose a direct chat</h2>
                    </div>
                )}
            </section>

            <aside className="chats-me-page__details" aria-label="Chat details" />
        </section>
    );
}

function getDirectChatTitle(chat: DirectChat, authenticatedUserUuid: string | null) {
    if (chat.firstMemberUuid === authenticatedUserUuid) {
        return chat.secondMemberUuid;
    }

    return chat.firstMemberUuid;
}

function formatMessageTime(date: string) {
    return new Intl.DateTimeFormat(undefined, {
        hour: "2-digit",
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
