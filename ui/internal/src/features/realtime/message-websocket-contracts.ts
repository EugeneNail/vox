import { MessageAttachment } from "../../messages/message-attachments";

export type MessageWebSocketEvent = {
    type?: string;
    data?: unknown;
};

export type MessageCreatedEvent = {
    messageUuid: string;
    chatUuid: string;
    userUuid: string;
    recipientUuids: string[];
    revision: number;
    text: string;
    attachments: MessageAttachment[];
    createdAt: string;
    updatedAt: string;
};

export type LastSeenRevisionUpdatedEvent = {
    chatUuid: string;
    userUuid: string;
    lastSeenRevision: number;
};

export type MessageEditedEvent = {
    messageUuid: string;
    chatUuid: string;
    userUuid: string;
    text: string;
    attachments: MessageAttachment[];
    createdAt: string;
    updatedAt: string;
};

export type MessageDeletedEvent = {
    messageUuid: string;
    chatUuid: string;
    userUuid: string;
};

export function parseMessageWebSocketEvent(raw: string): MessageWebSocketEvent {
    return JSON.parse(raw) as MessageWebSocketEvent;
}

export function buildOpenChatCommand(chatUuid: string) {
    return JSON.stringify({
        type: "OpenChat",
        data: {
            chatUuid,
        },
    });
}

export function buildCloseChatCommand() {
    return JSON.stringify({
        type: "UnsubscribeChat",
    });
}
