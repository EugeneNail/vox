import { createContext, ReactNode, useEffect, useState } from "react";

import { authTokensChangedEventName, getLoginToken } from "../../auth/auth-tokens";

type MessageWebSocketContextValue = {
    isConnected: boolean;
};

type MessageWebSocketProviderProps = {
    children: ReactNode;
};

const MessageWebSocketContext = createContext<MessageWebSocketContextValue>({
    isConnected: false,
});

let messageWebSocket: WebSocket | null = null;
let messageWebSocketToken: string | null = null;

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
        }

        function handleClose() {
            setIsConnected(false);
        }

        function handleMessage(event: MessageEvent<string>) {
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
        <MessageWebSocketContext.Provider value={{ isConnected }}>
            {children}
        </MessageWebSocketContext.Provider>
    );
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
