import { createContext, ReactNode, useContext, useEffect, useState } from "react";

import { authTokensChangedEventName, getLoginToken } from "../../auth/auth-tokens";
import { getValidLoginToken } from "../../auth/refresh-login-token";
import { MessageAttachment } from "../../messages/message-attachments";

type MessageWebSocketContextValue = {
    isConnected: boolean;
    subscribeChat: (chatUuid: string) => void;
    unsubscribeChat: (chatUuid: string) => void;
    addMessageListener: (listener: AddMessageListener) => () => void;
    updateMessageListener: (listener: UpdateMessageListener) => () => void;
    removeMessageListener: (listener: RemoveMessageListener) => () => void;
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
    attachments: MessageAttachment[];
    createdAt: string;
    updatedAt: string;
};

export type UpdateMessageEvent = {
    messageUuid: string;
    chatUuid: string;
    userUuid: string;
    text: string;
    attachments: MessageAttachment[];
    createdAt: string;
    updatedAt: string;
};

export type RemoveMessageEvent = {
    messageUuid: string;
    chatUuid: string;
    userUuid: string;
};

type AddMessageListener = (event: AddMessageEvent) => void;
type UpdateMessageListener = (event: UpdateMessageEvent) => void;
type RemoveMessageListener = (event: RemoveMessageEvent) => void;
type MessageWebSocketListeners = {
    handleOpen: () => void;
    handleClose: () => void;
    handleMessage: (event: MessageEvent<string>) => void;
};

const MessageWebSocketContext = createContext<MessageWebSocketContextValue>({
    isConnected: false,
    subscribeChat,
    unsubscribeChat,
    addMessageListener,
    updateMessageListener,
    removeMessageListener,
});

let messageWebSocket: WebSocket | null = null;
let messageWebSocketToken: string | null = null;
let messageWebSocketReconnectTimeoutId: number | null = null;
let messageWebSocketReconnectAttempt = 0;
const addMessageListeners = new Set<AddMessageListener>();
const updateMessageListeners = new Set<UpdateMessageListener>();
const removeMessageListeners = new Set<RemoveMessageListener>();
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

        let shouldReconnect = true;
        // Defer the side effect so React StrictMode can cancel the probe mount
        // before it creates and immediately closes a CONNECTING WebSocket.
        let openTimeoutId: number | null = window.setTimeout(() => {
            if (!shouldReconnect) {
                return;
            }

            void openMessageWebSocketWithFreshToken(() => shouldReconnect, {
                handleOpen,
                handleClose,
                handleMessage,
            }, (webSocket) => {
                setIsConnected(webSocket.readyState === WebSocket.OPEN);
            });
            openTimeoutId = null;
        }, 0);

        function handleOpen() {
            messageWebSocketReconnectAttempt = 0;
            setIsConnected(true);
        }

        function handleClose() {
            setIsConnected(false);
            if (shouldReconnect) {
                scheduleMessageWebSocketReconnect(() => shouldReconnect, {
                    handleOpen,
                    handleClose,
                    handleMessage,
                });
            }
        }

        function handleMessage(event: MessageEvent<string>) {
            try {
                const websocketEvent = JSON.parse(event.data) as MessageWebSocketEvent;
                if (websocketEvent.type === "MessageCreated") {
                    addMessageListeners.forEach((listener) => {
                        listener(websocketEvent.data as AddMessageEvent);
                    });
                    return;
                }

                if (websocketEvent.type === "chat.update_message") {
                    updateMessageListeners.forEach((listener) => {
                        listener(websocketEvent.data as UpdateMessageEvent);
                    });
                    return;
                }

                if (websocketEvent.type === "chat.remove_message") {
                    removeMessageListeners.forEach((listener) => {
                        listener(websocketEvent.data as RemoveMessageEvent);
                    });
                    return;
                }
            } catch {
                // Keep logging raw messages for temporary connection probes such as "pong".
            }

            console.log("message websocket:", event.data);
        }

        return () => {
            shouldReconnect = false;
            if (openTimeoutId !== null) {
                window.clearTimeout(openTimeoutId);
            }
            closeMessageWebSocket();
        };
    }, [loginToken]);

    return (
        <MessageWebSocketContext.Provider value={{ isConnected, subscribeChat, unsubscribeChat, addMessageListener, updateMessageListener, removeMessageListener }}>
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
    const url = new URL("/api/v1/realtime/ws", `${protocol}//${window.location.host}`);
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

async function openMessageWebSocketWithFreshToken(shouldOpen: () => boolean, listeners: MessageWebSocketListeners, handleOpened?: (webSocket: WebSocket) => void) {
    const loginToken = await getValidLoginToken();
    if (!shouldOpen()) {
        return;
    }

    const webSocket = openMessageWebSocket(loginToken, listeners);
    handleOpened?.(webSocket);
}

function scheduleMessageWebSocketReconnect(shouldOpen: () => boolean, listeners: MessageWebSocketListeners) {
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
        void openMessageWebSocketWithFreshToken(shouldOpen, listeners);
    }, delay);
}

function clearMessageWebSocketReconnect() {
    if (messageWebSocketReconnectTimeoutId === null) {
        return;
    }

    window.clearTimeout(messageWebSocketReconnectTimeoutId);
    messageWebSocketReconnectTimeoutId = null;
}

function subscribeChat(chatUuid: string) {
    void chatUuid;
}

function unsubscribeChat(chatUuid: string) {
    void chatUuid;
}

function addMessageListener(listener: AddMessageListener) {
    addMessageListeners.add(listener);

    return () => {
        addMessageListeners.delete(listener);
    };
}

function updateMessageListener(listener: UpdateMessageListener) {
    updateMessageListeners.add(listener);

    return () => {
        updateMessageListeners.delete(listener);
    };
}

function removeMessageListener(listener: RemoveMessageListener) {
    removeMessageListeners.add(listener);

    return () => {
        removeMessageListeners.delete(listener);
    };
}
