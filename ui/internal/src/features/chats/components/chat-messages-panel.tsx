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
    getChatAvatarLabel,
    getChatAvatarUrl,
    getMessageAuthorFirstName,
    hasRenderableMessageText,
    getMessageAuthorAvatarLabel,
    getMessageAuthorAvatarUrl,
    getMessageAuthorName,
    isMessageEdited,
    isMessageThreadStart,
    isOwnConfirmedMessage,
} from "./chats-ui";

export type MessageContextMenu = {
    message: ChatMessage;
    x: number;
    y: number;
};

export type ChatMessageViewMode = "bubble" | "plain";

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
    messageViewMode: ChatMessageViewMode;
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
    setMessageViewMode: (mode: ChatMessageViewMode) => void;
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
        messageViewMode,
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
        setMessageViewMode,
    } = props;

    if (!selectedChat) {
        return (
            <section className="chats-me-page__chat" aria-label="Chat">
                <div className="chats-me-page__chat-placeholder">
                    <p className="chats-me-page__state">Choose a chat</p>
                </div>
            </section>
        );
    }

    return (
        <section className="chats-me-page__chat" aria-label="Chat">
            <div className={`chats-me-page__chat-shell chats-me-page__chat-shell--${messageViewMode}`}>
                <header className="chats-me-page__chat-topbar">
                    <UserAvatar
                        className="chats-me-page__chat-topbar-avatar"
                        src={getChatAvatarUrl(selectedChat, authenticatedUserUuid, profilesByUserUuid)}
                        label={getChatAvatarLabel(selectedChat, authenticatedUserUuid, profilesByUserUuid)}
                    />
                    <div className="chats-me-page__chat-topbar-copy">
                        <h2 className="chats-me-page__chat-topbar-title">{selectedChat.uuid}</h2>
                        <p className="chats-me-page__chat-topbar-meta">{selectedChat.memberUuids.length} members</p>
                    </div>
                </header>
                <div
                    ref={messagesContainerRef}
                    className={`chats-me-page__messages chats-me-page__messages--${messageViewMode}`}
                    aria-live="polite"
                    onContextMenu={(event) => event.preventDefault()}
                    onScroll={onScroll}
                >
                    <div className="chats-me-page__messages-toolbar">
                        <div className="chats-me-page__message-view-switcher" aria-label="Message view mode">
                            <button
                                className={messageViewMode === "bubble" ? "chats-me-page__message-view-button chats-me-page__message-view-button--active" : "chats-me-page__message-view-button"}
                                type="button"
                                onClick={() => setMessageViewMode("bubble")}
                            >
                                Bubbles
                            </button>
                            <button
                                className={messageViewMode === "plain" ? "chats-me-page__message-view-button chats-me-page__message-view-button--active" : "chats-me-page__message-view-button"}
                                type="button"
                                onClick={() => setMessageViewMode("plain")}
                            >
                                Plain
                            </button>
                        </div>
                    </div>
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
                            messageViewMode={messageViewMode}
                            messageElementByRevisionRef={messageElementByRevisionRef}
                            nextMessage={messages[index + 1]}
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
    messageViewMode,
    messageElementByRevisionRef,
    nextMessage,
    previousMessage,
    profilesByUserUuid,
    onContextMenu,
}: {
    authenticatedUserUuid: string | null;
    editingMessageUuid: string | null;
    message: ChatMessage;
    messageViewMode: ChatMessageViewMode;
    messageElementByRevisionRef: RefObject<Record<number, HTMLElement | null>>;
    nextMessage?: ChatMessage;
    previousMessage?: ChatMessage;
    profilesByUserUuid: Record<string, PublicProfile>;
    onContextMenu: (event: React.MouseEvent<HTMLElement>, message: ChatMessage) => void;
}) {
    const isOwnMessage = message.userUuid === authenticatedUserUuid;
    const isThreadStart = isMessageThreadStart(message, previousMessage);
    const isThreadEnd = !nextMessage || isMessageThreadStart(nextMessage, message);
    const shouldShowBubbleHeader = messageViewMode === "bubble" && isThreadStart;
    const shouldShowBubbleAvatar = messageViewMode === "bubble" && isThreadEnd;
    const shouldShowPlainAvatar = messageViewMode === "plain" && isThreadStart;
    const hasMessageText = hasRenderableMessageText(message.text);
    const isEditedMessage = isMessageEdited(message);
    const hasDeliveryMark = isOwnConfirmedMessage(message, authenticatedUserUuid);
    const bubbleInlineMeta = (
        <span className="chats-me-page__message-inline-meta">
            {isEditedMessage && (
                <span className="chats-me-page__message-edited-meta">edited</span>
            )}
            <time
                className="chats-me-page__message-time"
                dateTime={message.createdAt}
            >
                {formatMessageTime(message.createdAt)}
            </time>
            {hasDeliveryMark && (
                <span className="material-symbols-rounded chats-me-page__message-delivery-icon" aria-label="Delivered">done</span>
            )}
        </span>
    );

    return (
        <article
            className={
                [
                    "chats-me-page__message",
                    `chats-me-page__message--${messageViewMode}`,
                    isOwnMessage ? "chats-me-page__message--own" : "chats-me-page__message--foreign",
                    isThreadStart ? "chats-me-page__message--thread-start" : "",
                    isThreadEnd ? "chats-me-page__message--thread-end" : "",
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
                {(shouldShowBubbleAvatar || shouldShowPlainAvatar) && (
                    <UserAvatar
                        className="chats-me-page__message-avatar"
                        src={getMessageAuthorAvatarUrl(message, profilesByUserUuid)}
                        label={getMessageAuthorAvatarLabel(message, profilesByUserUuid)}
                    />
                )}
            </div>
            <div className="chats-me-page__message-content">
                {messageViewMode === "bubble" ? (
                    <>
                        <div className="chats-me-page__message-bubble">
                            {shouldShowBubbleHeader && !isOwnMessage && (
                                <div className="chats-me-page__message-header chats-me-page__message-header--bubble">
                                    <p className="chats-me-page__message-author">{getMessageAuthorFirstName(message, profilesByUserUuid)}</p>
                                </div>
                            )}
                            <div className="chats-me-page__message-body">
                                <MessageText
                                    message={message}
                                    showEditedInline={false}
                                    trailingMeta={hasMessageText ? bubbleInlineMeta : null}
                                />
                                <MessageAttachments attachments={message.attachments} />
                            </div>
                            {!hasMessageText && (
                                <div className="chats-me-page__message-status">
                                    {bubbleInlineMeta}
                                </div>
                            )}
                        </div>
                    </>
                ) : (
                    <>
                        {isThreadStart && (
                            <div className="chats-me-page__message-header chats-me-page__message-header--plain">
                                <p className="chats-me-page__message-author">{getMessageAuthorName(message, profilesByUserUuid)}</p>
                                <div className="chats-me-page__message-header-meta">
                                    <time
                                        className="chats-me-page__message-time"
                                        dateTime={message.createdAt}
                                    >
                                        {formatMessageTime(message.createdAt)}
                                    </time>
                                    {isOwnConfirmedMessage(message, authenticatedUserUuid) && (
                                        <span className="material-symbols-rounded chats-me-page__message-delivery-icon" aria-label="Delivered">done</span>
                                    )}
                                </div>
                            </div>
                        )}
                        <div className="chats-me-page__message-body chats-me-page__message-body--plain">
                            <MessageText message={message} showEditedInline />
                            <MessageAttachments attachments={message.attachments} />
                        </div>
                    </>
                )}
            </div>
        </article>
    );
}
