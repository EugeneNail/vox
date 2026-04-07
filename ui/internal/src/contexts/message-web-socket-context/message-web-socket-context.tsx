import { createContext, ReactNode, useContext, useEffect, useState } from "react";

import { authTokensChangedEventName, getLoginToken } from "../../auth/auth-tokens";

type MessageWebSocketContextValue = {
    isConnected: boolean;
    subscribeDirectChat: (directChatUuid: string) => void;
    unsubscribeDirectChat: (directChatUuid: string) => void;
    addMessageCreatedListener: (listener: MessageCreatedListener) => () => void;
};

type MessageWebSocketProviderProps = {
    children: ReactNode;
};

type MessageWebSocketEvent = {
    type?: string;
    data?: unknown;
};

export type MessageCreatedEvent = {
    messageUuid: string;
    chatUuid: string;
    userUuid: string;
    text: string;
    createdAt: string;
    updatedAt: string;
};

type MessageCreatedListener = (event: MessageCreatedEvent) => void;

const MessageWebSocketContext = createContext<MessageWebSocketContextValue>({
    isConnected: false,
    subscribeDirectChat,
    unsubscribeDirectChat,
    addMessageCreatedListener,
});

let messageWebSocket: WebSocket | null = null;
let messageWebSocketToken: string | null = null;
const directChatSubscriptions = new Set<string>();
const messageCreatedListeners = new Set<MessageCreatedListener>();

export function MessageWebSocketProvider({ children }: MessageWebSocketProviderProps) {
    const [loginToken, setLoginToken] = useState(() => getLoginToken());
    const [isConnected, setIsConnected] = useState(false);

    useEffect(() => {
        function handleAuthTokensChanged() {
            setLoginToken(getLoginToken());
        }

        window.addEventListener(authTokensChangedEventName, handleAuthTokensChanged);
        window.addEventListener("storage", handleAuthTokensChanged);

        return () => {
            window.removeEventListener(authTokensChangedEventName, handleAuthTokensChanged);
            window.removeEventListener("storage", handleAuthTokensChanged);
        };
    }, []);

    useEffect(() => {
        if (!loginToken) {
            closeMessageWebSocket();
            setIsConnected(false);
            return;
        }

        const webSocket = openMessageWebSocket(loginToken);
        setIsConnected(webSocket.readyState === WebSocket.OPEN);

        function handleOpen() {
            setIsConnected(true);
            directChatSubscriptions.forEach((directChatUuid) => {
                sendMessageWebSocketCommand({
                    type: "chat.subscribe",
                    chatUuid: directChatUuid,
                });
            });
        }

        function handleClose() {
            setIsConnected(false);
        }

        function handleMessage(event: MessageEvent<string>) {
            try {
                const websocketEvent = JSON.parse(event.data) as MessageWebSocketEvent;
                if (websocketEvent.type === "message.created") {
                    messageCreatedListeners.forEach((listener) => {
                        listener(websocketEvent.data as MessageCreatedEvent);
                    });
                    return;
                }
            } catch {
                // Keep logging raw messages for temporary connection probes such as "pong".
            }

            console.log("message websocket:", event.data);
        }

        webSocket.addEventListener("open", handleOpen);
        webSocket.addEventListener("close", handleClose);
        webSocket.addEventListener("message", handleMessage);

        return () => {
            webSocket.removeEventListener("open", handleOpen);
            webSocket.removeEventListener("close", handleClose);
            webSocket.removeEventListener("message", handleMessage);
        };
    }, [loginToken]);

    return (
        <MessageWebSocketContext.Provider value={{ isConnected, subscribeDirectChat, unsubscribeDirectChat, addMessageCreatedListener }}>
            {children}
        </MessageWebSocketContext.Provider>
    );
}

export function useMessageWebSocket() {
    return useContext(MessageWebSocketContext);
}

function openMessageWebSocket(loginToken: string) {
    if (
        messageWebSocket &&
        messageWebSocketToken === loginToken &&
        (messageWebSocket.readyState === WebSocket.CONNECTING || messageWebSocket.readyState === WebSocket.OPEN)
    ) {
        return messageWebSocket;
    }

    closeMessageWebSocket();

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url = new URL("/api/v1/message/ws", `${protocol}//${window.location.host}`);
    url.searchParams.set("token", loginToken);

    messageWebSocket = new WebSocket(url);
    messageWebSocketToken = loginToken;
    messageWebSocket.addEventListener("close", () => {
        messageWebSocket = null;
        messageWebSocketToken = null;
    });

    return messageWebSocket;
}

function closeMessageWebSocket() {
    if (!messageWebSocket) {
        return;
    }

    messageWebSocket.close();
    messageWebSocket = null;
    messageWebSocketToken = null;
}

function subscribeDirectChat(directChatUuid: string) {
    directChatSubscriptions.add(directChatUuid);
    sendMessageWebSocketCommand({
        type: "chat.subscribe",
        chatUuid: directChatUuid,
    });
}

function unsubscribeDirectChat(directChatUuid: string) {
    directChatSubscriptions.delete(directChatUuid);
    sendMessageWebSocketCommand({
        type: "chat.unsubscribe",
        chatUuid: directChatUuid,
    });
}

function sendMessageWebSocketCommand(command: { type: string; chatUuid: string }) {
    if (!messageWebSocket || messageWebSocket.readyState !== WebSocket.OPEN) {
        return;
    }

    messageWebSocket.send(JSON.stringify(command));
}

function addMessageCreatedListener(listener: MessageCreatedListener) {
    messageCreatedListeners.add(listener);

    return () => {
        messageCreatedListeners.delete(listener);
    };
}
