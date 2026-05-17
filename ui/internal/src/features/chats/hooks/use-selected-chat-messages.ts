import { useEffect, useState } from "react";
import { useApiClient } from "../../../hooks/use-api-client";
import { Chat, ChatMessage, listChatMessages } from "../chat-api";

type UseSelectedChatMessagesParams = {
    selectedChatUuid: string | null;
    selectedChat: Chat | undefined;
    localLastSeenRevisionByChatUuid: Record<string, number>;
};

export function useSelectedChatMessages({
    selectedChatUuid,
    selectedChat,
    localLastSeenRevisionByChatUuid,
}: UseSelectedChatMessagesParams) {
    const apiClient = useApiClient();
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [isMessagesLoading, setIsMessagesLoading] = useState(false);
    const [messagesError, setMessagesError] = useState<string | null>(null);
    const [targetVisibleRevision, setTargetVisibleRevision] = useState<number | null>(null);

    useEffect(() => {
        if (!selectedChatUuid) {
            setMessages([]);
            setTargetVisibleRevision(null);
            return;
        }

        let isMounted = true;
        const currentSelectedChatUuid = selectedChatUuid;
        const selectedChatLastSeenRevision = Math.max(
            localLastSeenRevisionByChatUuid[currentSelectedChatUuid] ?? 0,
            selectedChat?.currentUserLastSeenRevision ?? 0,
        );
        const requestRevision = Math.max(0, selectedChatLastSeenRevision - 100);

        async function loadMessages() {
            setIsMessagesLoading(true);
            setMessagesError(null);

            try {
                const data = await listChatMessages(apiClient, currentSelectedChatUuid, requestRevision);
                if (!isMounted) {
                    return;
                }

                setMessages(data.map((message) => ({
                    ...message,
                    attachments: message.attachments ?? [],
                })));
                setTargetVisibleRevision(selectedChatLastSeenRevision);
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

        void loadMessages();

        return () => {
            isMounted = false;
        };
    }, [apiClient, localLastSeenRevisionByChatUuid, selectedChat?.currentUserLastSeenRevision, selectedChatUuid]);

    return {
        messages,
        setMessages,
        isMessagesLoading,
        messagesError,
        targetVisibleRevision,
        setTargetVisibleRevision,
    };
}
