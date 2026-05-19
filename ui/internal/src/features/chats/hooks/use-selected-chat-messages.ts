import { useEffect, useState } from "react";
import { useApiClient } from "../../../hooks/use-api-client";
import { ChatMessage, listChatMessages } from "../chat-api";

type UseSelectedChatMessagesParams = {
    selectedChatUuid: string | null;
    selectedChatLastSeenRevision: number;
};

export function useSelectedChatMessages({
    selectedChatUuid,
    selectedChatLastSeenRevision,
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
    }, [apiClient, selectedChatUuid]);

    useEffect(() => {
        if (!selectedChatUuid || messages.length === 0) {
            return;
        }

        if (targetVisibleRevision !== 0) {
            return;
        }

        if (selectedChatLastSeenRevision <= 0) {
            return;
        }

        setTargetVisibleRevision(selectedChatLastSeenRevision);
    }, [messages.length, selectedChatLastSeenRevision, selectedChatUuid, targetVisibleRevision]);

    return {
        messages,
        setMessages,
        isMessagesLoading,
        messagesError,
        targetVisibleRevision,
        setTargetVisibleRevision,
    };
}
