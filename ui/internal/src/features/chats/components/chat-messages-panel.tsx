import { type RefObject } from "react";
import MessageComposer from "../../../components/message-composer/message-composer";
import { type PublicProfile } from "../../../profiles/profile-cache";
import {
    Chat,
    ChatMessage,
    MessageAttachments,
    MessageText,
    UserAvatar,
    formatMessageTime,
    getChatTitle,
    getMessageAuthorAvatarLabel,
    getMessageAuthorAvatarUrl,
    getMessageAuthorName,
    isMessageThreadStart,
    isOwnConfirmedMessage,
} from "./chats-ui";

export type MessageContextMenu = {
    message: ChatMessage;
    x: number;
    y: number;
};

type ChatMessagesPanelProps = {
    authenticatedUserUuid: string | null;
    canDeleteAnyMessage: boolean;
    editingMessage: ChatMessage | null;
    handleMessageContextMenu: (event: React.MouseEvent<HTMLElement>, message: ChatMessage) => void;
    isKickingMember: boolean;
    isMessagesLoading: boolean;
    isSelectedChatGroup: boolean;
    isSelectedChatOwnedByAuthenticatedUser: boolean;
    messageContextMenu: MessageContextMenu | null;
    messageElementByRevisionRef: RefObject<Record<number, HTMLElement | null>>;
    messages: ChatMessage[];
    messagesContainerRef: RefObject<HTMLDivElement | null>;
    messagesEndRef: RefObject<HTMLDivElement | null>;
    messagesError: string | null;
    onCancelEdit: () => void;
    onCloseMessageContextMenu: () => void;
    onCopyMessageLink: (message: ChatMessage) => Promise<void>;
    onCopyMessageText: (message: ChatMessage) => Promise<void>;
    onDeleteMessageRequest: (message: ChatMessage) => void;
    onEditMessageRequest: (message: ChatMessage) => void;
    onKickMemberRequest: (memberUuid: string) => void;
    onScroll: () => void;
    onSubmit: (text: string, attachments: string[]) => Promise<void>;
    profilesByUserUuid: Record<string, PublicProfile>;
    selectedChat: Chat | null;
};

export function ChatMessagesPanel(props: ChatMessagesPanelProps) {
    const {
        authenticatedUserUuid,
        canDeleteAnyMessage,
        editingMessage,
        handleMessageContextMenu,
        isKickingMember,
        isMessagesLoading,
        isSelectedChatGroup,
        isSelectedChatOwnedByAuthenticatedUser,
        messageContextMenu,
        messageElementByRevisionRef,
        messages,
        messagesContainerRef,
        messagesEndRef,
        messagesError,
        onCancelEdit,
        onCloseMessageContextMenu,
        onCopyMessageLink,
        onCopyMessageText,
        onDeleteMessageRequest,
        onEditMessageRequest,
        onKickMemberRequest,
        onScroll,
        onSubmit,
        profilesByUserUuid,
        selectedChat,
    } = props;

    if (!selectedChat) {
        return (
            <section className="chats-me-page__chat" aria-label="Chat">
                <div className="chats-me-page__chat-placeholder">
                    <p className="chats-me-page__eyebrow">No chat selected</p>
                    <h2 className="chats-me-page__chat-title">Choose a chat</h2>
                </div>
            </section>
        );
    }

    return (
        <section className="chats-me-page__chat" aria-label="Chat">
            <div className="chats-me-page__chat-shell">
                <header className="chats-me-page__chat-header">
                    <p className="chats-me-page__eyebrow">Selected chat</p>
                    <h2 className="chats-me-page__chat-title">{getChatTitle(selectedChat, authenticatedUserUuid, profilesByUserUuid)}</h2>
                </header>

                <div
                    ref={messagesContainerRef}
                    className="chats-me-page__messages"
                    aria-live="polite"
                    onContextMenu={(event) => event.preventDefault()}
                    onScroll={onScroll}
                >
                    {isMessagesLoading && <p className="chats-me-page__state">Loading messages...</p>}
                    {messagesError && <p className="chats-me-page__state chats-me-page__state--error">{messagesError}</p>}
                    {!isMessagesLoading && !messagesError && messages.length === 0 && (
                        <p className="chats-me-page__state">No messages yet.</p>
                    )}
                    {messages.map((message, index) => (
                        <MessageRow
                            authenticatedUserUuid={authenticatedUserUuid}
                            editingMessageUuid={editingMessage?.uuid ?? null}
                            key={message.uuid}
                            message={message}
                            messageElementByRevisionRef={messageElementByRevisionRef}
                            previousMessage={messages[index - 1]}
                            profilesByUserUuid={profilesByUserUuid}
                            onContextMenu={handleMessageContextMenu}
                        />
                    ))}
                    <div ref={messagesEndRef} className="chats-me-page__messages-end" aria-hidden="true" />
                </div>

                {messageContextMenu && (
                    <div
                        className="chats-me-page__context-menu"
                        style={{
                            left: messageContextMenu.x,
                            top: messageContextMenu.y,
                        }}
                        onClick={(event) => event.stopPropagation()}
                    >
                        {messageContextMenu.message.userUuid === authenticatedUserUuid && (
                            <button
                                className="chats-me-page__context-menu-button"
                                type="button"
                                onClick={() => {
                                    onEditMessageRequest(messageContextMenu.message);
                                    onCloseMessageContextMenu();
                                }}
                            >
                                <span className="material-symbols-rounded" aria-hidden="true">edit</span>
                                Edit Message
                            </button>
                        )}
                        <button
                            className="chats-me-page__context-menu-button"
                            type="button"
                            onClick={() => void onCopyMessageText(messageContextMenu.message)}
                        >
                            <span className="material-symbols-rounded" aria-hidden="true">content_copy</span>
                            Copy text
                        </button>
                        <button
                            className="chats-me-page__context-menu-button"
                            type="button"
                            onClick={() => void onCopyMessageLink(messageContextMenu.message)}
                        >
                            <span className="material-symbols-rounded" aria-hidden="true">link</span>
                            Copy Message Link
                        </button>
                        {(messageContextMenu.message.userUuid === authenticatedUserUuid || canDeleteAnyMessage || (isSelectedChatGroup && isSelectedChatOwnedByAuthenticatedUser && messageContextMenu.message.userUuid !== authenticatedUserUuid)) && (
                            <div className="chats-me-page__context-menu-danger">
                                {(messageContextMenu.message.userUuid === authenticatedUserUuid || canDeleteAnyMessage) && (
                                    <button
                                        className="chats-me-page__context-menu-button chats-me-page__context-menu-button--danger"
                                        type="button"
                                        onClick={() => {
                                            onDeleteMessageRequest(messageContextMenu.message);
                                            onCloseMessageContextMenu();
                                        }}
                                    >
                                        <span className="material-symbols-rounded" aria-hidden="true">delete</span>
                                        Delete Message
                                    </button>
                                )}
                                {isSelectedChatGroup && isSelectedChatOwnedByAuthenticatedUser && messageContextMenu.message.userUuid !== authenticatedUserUuid && (
                                    <button
                                        className="chats-me-page__context-menu-button chats-me-page__context-menu-button--danger"
                                        type="button"
                                        disabled={isKickingMember}
                                        onClick={() => {
                                            onCloseMessageContextMenu();
                                            onKickMemberRequest(messageContextMenu.message.userUuid);
                                        }}
                                    >
                                        <span className="material-symbols-rounded" aria-hidden="true">person_remove</span>
                                        Kick User
                                    </button>
                                )}
                            </div>
                        )}
                    </div>
                )}

                <MessageComposer
                    disabled={!authenticatedUserUuid}
                    editingText={editingMessage?.text ?? null}
                    editingAttachments={editingMessage?.attachments ?? []}
                    onCancelEdit={onCancelEdit}
                    onSubmit={onSubmit}
                />
            </div>
        </section>
    );
}

function MessageRow({
    authenticatedUserUuid,
    editingMessageUuid,
    message,
    messageElementByRevisionRef,
    previousMessage,
    profilesByUserUuid,
    onContextMenu,
}: {
    authenticatedUserUuid: string | null;
    editingMessageUuid: string | null;
    message: ChatMessage;
    messageElementByRevisionRef: RefObject<Record<number, HTMLElement | null>>;
    previousMessage?: ChatMessage;
    profilesByUserUuid: Record<string, PublicProfile>;
    onContextMenu: (event: React.MouseEvent<HTMLElement>, message: ChatMessage) => void;
}) {
    const isThreadStart = isMessageThreadStart(message, previousMessage);

    return (
        <article
            className={
                [
                    "chats-me-page__message",
                    isThreadStart ? "chats-me-page__message--thread-start" : "",
                    message.isPending ? "chats-me-page__message--pending" : "",
                    editingMessageUuid === message.uuid ? "chats-me-page__message--editing" : "",
                ].filter(Boolean).join(" ")
            }
            ref={(element) => {
                messageElementByRevisionRef.current[message.revision] = element;
            }}
            onContextMenu={(event) => onContextMenu(event, message)}
        >
            <div className="chats-me-page__message-avatar-cell">
                {isThreadStart && (
                    <UserAvatar
                        className="chats-me-page__message-avatar"
                        src={getMessageAuthorAvatarUrl(message, profilesByUserUuid)}
                        label={getMessageAuthorAvatarLabel(message, profilesByUserUuid)}
                    />
                )}
            </div>
            <div className="chats-me-page__message-body">
                {isThreadStart && (
                    <div className="chats-me-page__message-header">
                        <p className="chats-me-page__message-author">{getMessageAuthorName(message, profilesByUserUuid)}</p>
                    </div>
                )}
                <MessageText message={message} />
                <MessageAttachments attachments={message.attachments} />
            </div>
            <div className="chats-me-page__message-status">
                {isOwnConfirmedMessage(message, authenticatedUserUuid) && (
                    <span className="material-symbols-rounded chats-me-page__message-delivery-icon" aria-label="Delivered">done</span>
                )}
                <time
                    className="chats-me-page__message-time"
                    dateTime={message.createdAt}
                >
                    {formatMessageTime(message.createdAt)}
                </time>
            </div>
        </article>
    );
}
