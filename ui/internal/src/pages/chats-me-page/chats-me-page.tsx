import { useEffect, useState } from "react";
import { getAuthenticatedUserUuid } from "../../auth/auth-tokens";
import MessageComposer from "../../components/message-composer/message-composer";
import { useApiClient } from "../../hooks/use-api-client";
import "./chats-me-page.sass";

type DirectChat = {
    uuid: string;
    firstMemberUuid: string;
    secondMemberUuid: string;
};

export default function ChatsMePage() {
    const apiClient = useApiClient();
    const [directChats, setDirectChats] = useState<DirectChat[]>([]);
    const [selectedChatUuid, setSelectedChatUuid] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const authenticatedUserUuid = getAuthenticatedUserUuid();

    useEffect(() => {
        let isMounted = true;

        async function loadDirectChats() {
            try {
                const { data } = await apiClient.get<DirectChat[]>("/api/v1/message/direct-chats");
                if (!isMounted) {
                    return;
                }

                setDirectChats(data);
                setSelectedChatUuid(data[0]?.uuid ?? null);
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

    const selectedChat = directChats.find((chat) => chat.uuid === selectedChatUuid);

    async function sendMessage(text: string) {
        if (!selectedChat) {
            return;
        }

        await apiClient.post(`/api/v1/message/chats/${selectedChat.uuid}/messages`, {
            text,
        });
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
                        <button
                            className={
                                chat.uuid === selectedChatUuid
                                    ? "chats-me-page__chat-button chats-me-page__chat-button--active"
                                    : "chats-me-page__chat-button"
                            }
                            key={chat.uuid}
                            type="button"
                            onClick={() => setSelectedChatUuid(chat.uuid)}
                        >
                            <span className="chats-me-page__avatar" aria-hidden="true" />
                            <span className="chats-me-page__chat-name">{getDirectChatTitle(chat, authenticatedUserUuid)}</span>
                        </button>
                    ))}
                </div>
            </aside>

            <section className="chats-me-page__chat" aria-label="Chat">
                {selectedChat ? (
                    <div className="chats-me-page__chat-shell">
                        <div className="chats-me-page__chat-placeholder">
                            <p className="chats-me-page__eyebrow">Selected chat</p>
                            <h2 className="chats-me-page__chat-title">{getDirectChatTitle(selectedChat, authenticatedUserUuid)}</h2>
                            <p className="chats-me-page__chat-text">
                                Message history will appear here.
                            </p>
                        </div>

                        <MessageComposer onSubmit={sendMessage} />
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
