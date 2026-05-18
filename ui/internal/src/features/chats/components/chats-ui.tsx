import { type ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { type Chat as ApiChat, type ChatMessage as ApiChatMessage } from "../chat-api";
import { buildAttachmentUrl, isImageAttachmentName, type MessageAttachment } from "../../../messages/message-attachments";
import { type PublicProfile } from "../../../profiles/profile-cache";
import { buildAvatarPlaceholder } from "../../../lib/avatar-placeholder";

export type Chat = ApiChat;
export type ChatMessage = ApiChatMessage;

const linkPattern = /(https?:\/\/[^\s]+)/g;
const fullLinkPattern = /^https?:\/\/[^\s]+$/;
const messageThreadGapMs = 10 * 60 * 1000;

export function UserAvatar({ className, src, label }: { className: string; src: string | null; label: string; }) {
    if (src) {
        return (
            <img
                className={`${className} ${className}--image`}
                src={src}
                alt=""
                aria-hidden="true"
                loading="lazy"
                decoding="async"
            />
        );
    }

    const placeholder = buildAvatarPlaceholder(label);

    return (
        <span
            className={`${className} ${className}--placeholder`}
            style={placeholder.style}
            aria-hidden="true"
        >
            {placeholder.initials}
        </span>
    );
}

export function getProfileAvatarUrl(profile: PublicProfile | null | undefined) {
    if (!profile?.avatar) {
        return null;
    }

    return buildAttachmentUrl(profile.avatar);
}

export function getChatTitle(chat: Chat, authenticatedUserUuid: string | null, profilesByUserUuid: Record<string, PublicProfile>) {
    if (chat.name) {
        return chat.name;
    }

    if (chat.chatType !== 0) {
        return chat.uuid;
    }

    const companionUuid = chat.memberUuids.find((memberUuid) => memberUuid !== authenticatedUserUuid);
    if (!companionUuid) {
        return chat.uuid;
    }

    return profilesByUserUuid[companionUuid]?.name ?? companionUuid;
}

export function getChatAvatarUrl(chat: Chat, authenticatedUserUuid: string | null, profilesByUserUuid: Record<string, PublicProfile>) {
    if (chat.chatType === 0) {
        const companionUuid = chat.memberUuids.find((memberUuid) => memberUuid !== authenticatedUserUuid);
        if (!companionUuid) {
            return null;
        }

        return getProfileAvatarUrl(profilesByUserUuid[companionUuid] ?? null);
    }

    if (!chat.avatar) {
        return null;
    }

    return buildAttachmentUrl(chat.avatar);
}

export function getChatAvatarLabel(chat: Chat, authenticatedUserUuid: string | null, profilesByUserUuid: Record<string, PublicProfile>) {
    if (chat.chatType === 0) {
        const companionUuid = chat.memberUuids.find((memberUuid) => memberUuid !== authenticatedUserUuid);
        if (!companionUuid) {
            return chat.uuid;
        }

        return profilesByUserUuid[companionUuid]?.name ?? companionUuid;
    }

    return getChatTitle(chat, authenticatedUserUuid, profilesByUserUuid);
}

export function getChatMemberName(memberUuid: string, profilesByUserUuid: Record<string, PublicProfile>) {
    return profilesByUserUuid[memberUuid]?.name ?? memberUuid;
}

export function getChatMemberAvatarUrl(memberUuid: string, profilesByUserUuid: Record<string, PublicProfile>) {
    return getProfileAvatarUrl(profilesByUserUuid[memberUuid] ?? null);
}

export function getChatMemberAvatarLabel(memberUuid: string, profilesByUserUuid: Record<string, PublicProfile>) {
    return getChatMemberName(memberUuid, profilesByUserUuid);
}

export function getMessageAuthorName(message: ChatMessage, profilesByUserUuid: Record<string, PublicProfile>) {
    return profilesByUserUuid[message.userUuid]?.name ?? message.userUuid;
}

export function getMessageAuthorFirstName(message: ChatMessage, profilesByUserUuid: Record<string, PublicProfile>) {
    const fullName = getMessageAuthorName(message, profilesByUserUuid).trim();
    const [firstWord] = fullName.split(/\s+/);
    return firstWord || fullName;
}

export function getChatLastMessagePreview(chat: Chat, profilesByUserUuid: Record<string, PublicProfile>) {
    const message = chat.lastMessage;
    if (!message) {
        return null;
    }

    if (hasRenderableMessageText(message.text)) {
        const authorFirstName = getMessageAuthorFirstName(message, profilesByUserUuid);
        return `${authorFirstName}: ${truncatePreviewText(message.text, 60)}`;
    }

    if (message.attachments.length > 0) {
        return `${message.attachments.length} attachments`;
    }

    return null;
}

export function getChatLastMessageTime(chat: Chat) {
    return chat.lastMessage ? formatMessageTime(chat.lastMessage.createdAt) : null;
}

export function getMessageAuthorAvatarUrl(message: ChatMessage, profilesByUserUuid: Record<string, PublicProfile>) {
    return getProfileAvatarUrl(profilesByUserUuid[message.userUuid] ?? null);
}

export function getMessageAuthorAvatarLabel(message: ChatMessage, profilesByUserUuid: Record<string, PublicProfile>) {
    return getMessageAuthorName(message, profilesByUserUuid);
}

export function getUnreadRevisionCount(chat: Chat, localLastSeenRevisionByChatUuid: Record<string, number>) {
    const localLastSeenRevision = Math.max(
        localLastSeenRevisionByChatUuid[chat.uuid] ?? 0,
        chat.currentUserLastSeenRevision,
    );

    return Math.max(0, chat.revision - localLastSeenRevision);
}

export function getTargetLastSeenRevision(
    chat: Chat,
    chatUuid: string,
    selectedChatUuid: string | null,
    isSelectedChatScrolledToBottom: boolean,
    localLastSeenRevisionByChatUuid: Record<string, number>,
) {
    const localLastSeenRevision = Math.max(
        localLastSeenRevisionByChatUuid[chatUuid] ?? 0,
        chat.currentUserLastSeenRevision,
    );
    if (chatUuid === selectedChatUuid && isSelectedChatScrolledToBottom) {
        return Math.max(localLastSeenRevision, chat.revision);
    }

    return localLastSeenRevision;
}

export function formatMessageTime(date: string) {
    return new Intl.DateTimeFormat(undefined, {
        hour: "2-digit",
        hour12: false,
        hourCycle: "h23",
        minute: "2-digit",
    }).format(new Date(date));
}

export function isMessageThreadStart(message: ChatMessage, previousMessage?: ChatMessage) {
    if (!previousMessage) {
        return true;
    }

    if (previousMessage.userUuid !== message.userUuid) {
        return true;
    }

    return new Date(message.createdAt).getTime() - new Date(previousMessage.createdAt).getTime() >= messageThreadGapMs;
}

export function isMessageEdited(message: ChatMessage) {
    return new Date(message.createdAt).getTime() !== new Date(message.updatedAt).getTime();
}

export function isOwnConfirmedMessage(message: ChatMessage, authenticatedUserUuid: string | null) {
    return !message.isPending && message.userUuid === authenticatedUserUuid;
}

export function hasRenderableMessageText(text: string) {
    return text.trim().length > 0;
}

function truncatePreviewText(text: string, maxLength: number) {
    const characters = Array.from(text.trim());
    if (characters.length <= maxLength) {
        return text.trim();
    }

    return `${characters.slice(0, maxLength).join("")}...`;
}

export function renderMessageText(text: string) {
    return text.split(linkPattern).map((part, index) => {
        if (!fullLinkPattern.test(part)) {
            return part;
        }

        return (
            <a className="chats-me-page__message-link" href={part} key={`${part}-${index}`} target="_blank" rel="noreferrer">
                {part}
            </a>
        );
    });
}

export function MessageText({
    message,
    showEditedInline = true,
    trailingMeta = null,
}: {
    message: ChatMessage;
    showEditedInline?: boolean;
    trailingMeta?: ReactNode;
}) {
    if (!hasRenderableMessageText(message.text)) {
        return null;
    }

    return (
        <p className={trailingMeta ? "chats-me-page__message-text chats-me-page__message-text--with-meta" : "chats-me-page__message-text"}>
            {renderMessageText(message.text)}
            {trailingMeta}
            {showEditedInline && isMessageEdited(message) && (
                <span className="chats-me-page__message-edited"> (edited)</span>
            )}
        </p>
    );
}

export function MessageAttachments({ attachments }: { attachments: MessageAttachment[]; }) {
    if (attachments.length === 0) {
        return null;
    }

    return (
        <div className="chats-me-page__message-attachments" aria-label="Message attachments">
            {attachments.map((attachment) => (
                <MessageAttachmentView attachment={attachment} key={attachment.uuid} />
            ))}
        </div>
    );
}

export function MessageAttachmentView({ attachment }: { attachment: MessageAttachment; }) {
    if (isImageAttachmentName(attachment.name)) {
        const url = buildAttachmentUrl(attachment.name);
        return (
            <a className="chats-me-page__message-attachment chats-me-page__message-attachment--image" href={url} target="_blank" rel="noreferrer">
                <img className="chats-me-page__message-attachment-image" src={url} alt={attachment.name} />
            </a>
        );
    }

    return (
        <a className="chats-me-page__message-attachment chats-me-page__message-attachment--document" href={buildAttachmentUrl(attachment.name)} target="_blank" rel="noreferrer">
            <span className="material-symbols-rounded chats-me-page__message-attachment-icon" aria-hidden="true">description</span>
            <span className="chats-me-page__message-attachment-name">{attachment.name}</span>
        </a>
    );
}

export function ModalBackdrop({
    children,
    onClick,
}: {
    children: ReactNode;
    onClick?: () => void;
}) {
    return (
        <div className="chats-me-page__modal-backdrop" role="presentation" onClick={onClick}>
            {children}
        </div>
    );
}

export function profilesByUserUuidFromProfiles(profiles: PublicProfile[]) {
    return profiles.reduce<Record<string, PublicProfile>>((accumulator, profile) => {
        accumulator[profile.userUuid] = profile;
        return accumulator;
    }, {});
}

export function collectReferencedUserUuids(chats: Chat[], messages: ChatMessage[], searchProfiles: PublicProfile[]) {
    return Array.from(new Set([
        ...chats.flatMap((chat) => chat.memberUuids),
        ...chats.flatMap((chat) => chat.lastMessage ? [chat.lastMessage.userUuid] : []),
        ...messages.map((message) => message.userUuid),
        ...searchProfiles.map((profile) => profile.userUuid),
    ]));
}

export function SearchResultButton({
    disabled,
    onClick,
    profile,
}: {
    disabled?: boolean;
    onClick?: () => void;
    profile: PublicProfile;
}) {
    return (
        <button
            className="chats-me-page__chat-button chats-me-page__chat-button--search-result"
            disabled={disabled}
            type="button"
            onClick={onClick}
        >
            <UserAvatar
                className="chats-me-page__avatar"
                src={getProfileAvatarUrl(profile)}
                label={profile.name}
            />
            <span className="chats-me-page__chat-preview">
                <span className="chats-me-page__chat-name">{profile.name}</span>
            </span>
        </button>
    );
}

export function ChatListLink({
    authenticatedUserUuid,
    chat,
    localLastSeenRevisionByChatUuid,
    profilesByUserUuid,
}: {
    authenticatedUserUuid: string | null;
    chat: Chat;
    localLastSeenRevisionByChatUuid: Record<string, number>;
    profilesByUserUuid: Record<string, PublicProfile>;
}) {
    const unreadRevisionCount = getUnreadRevisionCount(chat, localLastSeenRevisionByChatUuid);
    const chatPreview = getChatLastMessagePreview(chat, profilesByUserUuid);
    const chatTime = getChatLastMessageTime(chat);

    return (
        <NavLink
            className={
                ({ isActive }) => (
                    isActive
                        ? "chats-me-page__chat-button chats-me-page__chat-button--active"
                        : "chats-me-page__chat-button"
                )
            }
            to={`/chats/@me/${chat.uuid}`}
        >
            <UserAvatar
                className="chats-me-page__avatar"
                src={getChatAvatarUrl(chat, authenticatedUserUuid, profilesByUserUuid)}
                label={getChatAvatarLabel(chat, authenticatedUserUuid, profilesByUserUuid)}
            />
            <span className="chats-me-page__chat-preview">
                <span className="chats-me-page__chat-row">
                    <span className="chats-me-page__chat-name">{getChatTitle(chat, authenticatedUserUuid, profilesByUserUuid)}</span>
                    <span className="chats-me-page__chat-meta">
                        {chatTime && <span className="chats-me-page__chat-time">{chatTime}</span>}
                        {unreadRevisionCount > 0 && (
                            <span className="chats-me-page__chat-badge" aria-label={`${unreadRevisionCount} unread events`}>
                                {unreadRevisionCount}
                            </span>
                        )}
                    </span>
                </span>
                {chatPreview && (
                    <span className="chats-me-page__chat-message-piece">
                        {chatPreview}
                    </span>
                )}
            </span>
        </NavLink>
    );
}
