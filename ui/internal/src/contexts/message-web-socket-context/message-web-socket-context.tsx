import { createContext, ReactNode, useContext, useEffect, useState } from "react";

import { authTokensChangedEventName, getLoginToken } from "../../auth/auth-tokens";

type MessageWebSocketContextValue = {
    isConnected: boolean;
    subscribeDirectChat: (directChatUuid: string) => void;
    unsubscribeDirectChat: (directChatUuid: string) => void;
    addMessageListener: (listener: AddMessageListener) => () => void;
};

type MessageWebSocketProviderProps = {
    children: ReactNode;
};

type MessageWebSocketEvent = {
    type?: string;
    data?: unknown;
};

export type AddMessageEvent = {
    messageUuid: string;
    chatUuid: string;
    userUuid: string;
    text: string;
    createdAt: string;
    updatedAt: string;
};

type AddMessageListener = (event: AddMessageEvent) => void;
type MessageWebSocketListeners = {
    handleOpen: () => void;
    handleClose: () => void;
    handleMessage: (event: MessageEvent<string>) => void;
};

const MessageWebSocketContext = createContext<MessageWebSocketContextValue>({
    isConnected: false,
    subscribeDirectChat,
    unsubscribeDirectChat,
    addMessageListener,
});

let messageWebSocket: WebSocket | null = null;
let messageWebSocketToken: string | null = null;
let messageWebSocketReconnectTimeoutId: number | null = null;
let messageWebSocketReconnectAttempt = 0;
const directChatSubscriptions = new Set<string>();
const addMessageListeners = new Set<AddMessageListener>();
const messageWebSocketReconnectBaseDelayMs = 500;
const messageWebSocketReconnectMaxDelayMs = 5000;

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

        const activeLoginToken = loginToken;
        let shouldReconnect = true;

        function handleOpen() {
            messageWebSocketReconnectAttempt = 0;
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
            if (shouldReconnect) {
                scheduleMessageWebSocketReconnect(activeLoginToken, {
                    handleOpen,
                    handleClose,
                    handleMessage,
                });
            }
        }

        function handleMessage(event: MessageEvent<string>) {
            try {
                const websocketEvent = JSON.parse(event.data) as MessageWebSocketEvent;
                if (websocketEvent.type === "chat.add_message") {
                    addMessageListeners.forEach((listener) => {
                        listener(websocketEvent.data as AddMessageEvent);
                    });
                    return;
                }
            } catch {
                // Keep logging raw messages for temporary connection probes such as "pong".
            }

            console.log("message websocket:", event.data);
        }

        const webSocket = openMessageWebSocket(activeLoginToken, {
            handleOpen,
            handleClose,
            handleMessage,
        });
        setIsConnected(webSocket.readyState === WebSocket.OPEN);

        return () => {
            shouldReconnect = false;
            closeMessageWebSocket();
        };
    }, [loginToken]);

    return (
        <MessageWebSocketContext.Provider value={{ isConnected, subscribeDirectChat, unsubscribeDirectChat, addMessageListener }}>
            {children}
        </MessageWebSocketContext.Provider>
    );
}

export function useMessageWebSocket() {
    return useContext(MessageWebSocketContext);
}

function openMessageWebSocket(loginToken: string, listeners: MessageWebSocketListeners) {
    if (
        messageWebSocket &&
        messageWebSocketToken === loginToken &&
        (messageWebSocket.readyState === WebSocket.CONNECTING || messageWebSocket.readyState === WebSocket.OPEN)
    ) {
        return messageWebSocket;
    }

    closeMessageWebSocket();
    clearMessageWebSocketReconnect();

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url = new URL("/api/v1/message/ws", `${protocol}//${window.location.host}`);
    url.searchParams.set("token", loginToken);

    messageWebSocket = new WebSocket(url);
    messageWebSocketToken = loginToken;
    messageWebSocket.addEventListener("open", listeners.handleOpen);
    messageWebSocket.addEventListener("close", listeners.handleClose);
    messageWebSocket.addEventListener("message", listeners.handleMessage);
    messageWebSocket.addEventListener("close", () => {
        messageWebSocket = null;
        messageWebSocketToken = null;
    });

    return messageWebSocket;
}

function closeMessageWebSocket() {
    clearMessageWebSocketReconnect();

    if (!messageWebSocket) {
        return;
    }

    messageWebSocket.close();
    messageWebSocket = null;
    messageWebSocketToken = null;
}

function scheduleMessageWebSocketReconnect(loginToken: string, listeners: MessageWebSocketListeners) {
    if (messageWebSocketReconnectTimeoutId !== null) {
        return;
    }

    const delay = Math.min(
        messageWebSocketReconnectBaseDelayMs * 2 ** messageWebSocketReconnectAttempt,
        messageWebSocketReconnectMaxDelayMs,
    );
    messageWebSocketReconnectAttempt += 1;

    messageWebSocketReconnectTimeoutId = window.setTimeout(() => {
        messageWebSocketReconnectTimeoutId = null;
        openMessageWebSocket(loginToken, listeners);
    }, delay);
}

function clearMessageWebSocketReconnect() {
    if (messageWebSocketReconnectTimeoutId === null) {
        return;
    }

    window.clearTimeout(messageWebSocketReconnectTimeoutId);
    messageWebSocketReconnectTimeoutId = null;
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

function addMessageListener(listener: AddMessageListener) {
    addMessageListeners.add(listener);

    return () => {
        addMessageListeners.delete(listener);
    };
}
