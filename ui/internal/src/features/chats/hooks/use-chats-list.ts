import { useEffect, useState } from "react";
import { listChats, Chat } from "../chat-api";
import { useApiClient } from "../../../hooks/use-api-client";

export function useChatsList() {
    const apiClient = useApiClient();
    const [chats, setChats] = useState<Chat[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        let isMounted = true;

        async function loadChats() {
            try {
                const data = await listChats(apiClient);
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

        void loadChats();

        return () => {
            isMounted = false;
        };
    }, [apiClient]);

    return {
        chats,
        setChats,
        isLoading,
        setIsLoading,
        error,
        setError,
    };
}
