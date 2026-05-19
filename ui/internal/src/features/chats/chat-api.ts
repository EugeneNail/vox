import { AxiosInstance } from "axios";
import { MessageAttachment } from "../../messages/message-attachments";

export type Chat = {
    uuid: string;
    name: string | null;
    avatar: string | null;
    chatType: number;
    revision: number;
    createdByUserUuid: string;
    memberUuids: string[];
    currentUserRole: string;
    currentUserLastSeenRevision: number;
    lastMessage: ChatMessage | null;
    createdAt: string;
    updatedAt: string;
};

export type ChatMessage = {
    uuid: string;
    chatUuid: string;
    userUuid: string;
    revision: number;
    text: string;
    attachments: MessageAttachment[];
    createdAt: string;
    updatedAt: string;
    isPending?: boolean;
};

export type CreateDirectChatRequest = {
    interlocutorUuid: string;
};

export type CreateGroupChatRequest = {
    memberUuids: string[];
    name: string | null;
};

export type AddChatMembersRequest = {
    memberUuids: string[];
};

export type RenameChatRequest = {
    name: string;
};

export type ChangeChatAvatarRequest = {
    avatar: string;
};

export type CreateMessageRequest = {
    text: string;
    attachments: string[];
};

export type EditMessageRequest = {
    text: string;
    attachments: string[];
};

export type SetLastSeenRevisionRequest = {
    revision: number;
};

export async function listChats(apiClient: AxiosInstance) {
    const { data } = await apiClient.get<Chat[]>("/api/v1/message/chats");
    return data;
}

export async function listChatMessages(apiClient: AxiosInstance, chatUuid: string, revision: number) {
    const { data } = await apiClient.get<ChatMessage[]>(`/api/v1/message/chats/${chatUuid}/messages?revision=${revision}`);
    return data;
}

export async function createDirectChat(apiClient: AxiosInstance, payload: CreateDirectChatRequest) {
    const { data } = await apiClient.post<string>("/api/v1/message/chats/direct", payload);
    return data;
}

export async function createGroupChat(apiClient: AxiosInstance, payload: CreateGroupChatRequest) {
    const { data } = await apiClient.post<string>("/api/v1/message/chats/group", payload);
    return data;
}

export async function addChatMembers(apiClient: AxiosInstance, chatUuid: string, payload: AddChatMembersRequest) {
    await apiClient.post(`/api/v1/message/chats/${chatUuid}/members`, payload);
}

export async function renameChat(apiClient: AxiosInstance, chatUuid: string, payload: RenameChatRequest) {
    await apiClient.patch(`/api/v1/message/chats/${chatUuid}/name`, payload);
}

export async function changeChatAvatar(apiClient: AxiosInstance, chatUuid: string, payload: ChangeChatAvatarRequest) {
    await apiClient.patch(`/api/v1/message/chats/${chatUuid}/avatar`, payload);
}

export async function kickChatMember(apiClient: AxiosInstance, chatUuid: string, memberUuid: string) {
    await apiClient.delete(`/api/v1/message/chats/${chatUuid}/members/${memberUuid}`);
}

export async function createMessage(apiClient: AxiosInstance, chatUuid: string, payload: CreateMessageRequest) {
    const { data } = await apiClient.post<string>(`/api/v1/message/chats/${chatUuid}/messages`, payload);
    return data;
}

export async function editMessage(apiClient: AxiosInstance, messageUuid: string, payload: EditMessageRequest) {
    await apiClient.put<string>(`/api/v1/message/messages/${messageUuid}`, payload);
}

export async function deleteMessage(apiClient: AxiosInstance, messageUuid: string) {
    await apiClient.delete<string>(`/api/v1/message/messages/${messageUuid}`);
}

export async function setLastSeenRevision(apiClient: AxiosInstance, chatUuid: string, payload: SetLastSeenRevisionRequest) {
    await apiClient.post(`/api/v1/message/chats/${chatUuid}/last-seen-revision`, payload);
}
