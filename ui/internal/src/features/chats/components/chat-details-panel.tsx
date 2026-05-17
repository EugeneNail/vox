import { type PublicProfile } from "../../../profiles/profile-cache";
import {
    Chat,
    UserAvatar,
    getChatAvatarUrl,
    getChatMemberAvatarLabel,
    getChatMemberAvatarUrl,
    getChatMemberName,
} from "./chats-ui";

type ChatDetailsPanelProps = {
    authenticatedUserUuid: string | null;
    isKickingMember: boolean;
    isSelectedChatGroup: boolean;
    isSelectedChatOwnedByAuthenticatedUser: boolean;
    onKickMember: (memberUuid: string) => void;
    onOpenAddMembersModal: () => void;
    profilesByUserUuid: Record<string, PublicProfile>;
    selectedChat: Chat | null;
};

export function ChatDetailsPanel(props: ChatDetailsPanelProps) {
    const {
        authenticatedUserUuid,
        isKickingMember,
        isSelectedChatGroup,
        isSelectedChatOwnedByAuthenticatedUser,
        onKickMember,
        onOpenAddMembersModal,
        profilesByUserUuid,
        selectedChat,
    } = props;

    return (
        <aside className="chats-me-page__details" aria-label="Chat details">
            {selectedChat ? (
                <div className="chats-me-page__details-shell">
                    <div className="chats-me-page__details-topbar">
                        <button className="chats-me-page__details-icon-button" type="button" aria-label="Close chat info">
                            <span className="material-symbols-rounded" aria-hidden="true">close</span>
                        </button>
                        <h2 className="chats-me-page__details-title">Chat Info</h2>
                        <button className="chats-me-page__details-icon-button" type="button" aria-label="Edit chat info">
                            <span className="material-symbols-rounded" aria-hidden="true">edit</span>
                        </button>
                    </div>

                    <div className="chats-me-page__details-profile">
                        <UserAvatar
                            className="chats-me-page__details-chat-avatar"
                            src={getChatAvatarUrl(selectedChat, authenticatedUserUuid, profilesByUserUuid)}
                            label={selectedChat.uuid}
                        />
                        <h3 className="chats-me-page__details-chat-name">{selectedChat.uuid}</h3>
                        <p className="chats-me-page__details-chat-meta">{selectedChat.memberUuids.length} members</p>
                    </div>

                    <div className="chats-me-page__members-panel">
                    <div className="chats-me-page__members" aria-label="Chat members">
                        {selectedChat.memberUuids.map((memberUuid) => (
                            <div
                                className={
                                    [
                                        "chats-me-page__member",
                                        selectedChat.createdByUserUuid === memberUuid ? "chats-me-page__member--owner" : "",
                                        isSelectedChatGroup && isSelectedChatOwnedByAuthenticatedUser && memberUuid !== authenticatedUserUuid ? "chats-me-page__member--removable" : "",
                                    ].filter(Boolean).join(" ")
                                }
                                key={memberUuid}
                            >
                                <UserAvatar
                                    className="chats-me-page__member-avatar"
                                    src={getChatMemberAvatarUrl(memberUuid, profilesByUserUuid)}
                                    label={getChatMemberAvatarLabel(memberUuid, profilesByUserUuid)}
                                />
                                <span className="chats-me-page__member-name">
                                    {getChatMemberName(memberUuid, profilesByUserUuid)}
                                </span>
                                {selectedChat.createdByUserUuid === memberUuid && (
                                    <span className="chats-me-page__member-role">owner</span>
                                )}
                                {isSelectedChatGroup && isSelectedChatOwnedByAuthenticatedUser && memberUuid !== authenticatedUserUuid && (
                                    <button
                                        className="chats-me-page__member-remove-button"
                                        type="button"
                                        disabled={isKickingMember}
                                        title="Remove member"
                                        aria-label={`Remove member ${getChatMemberName(memberUuid, profilesByUserUuid)}`}
                                        onClick={() => onKickMember(memberUuid)}
                                    >
                                        <span className="material-symbols-rounded" aria-hidden="true">person_remove</span>
                                    </button>
                                )}
                            </div>
                        ))}
                        </div>
                    </div>

                    {selectedChat.chatType === 1 && (
                        <button className="chats-me-page__details-floating-action" type="button" onClick={onOpenAddMembersModal} aria-label="Add members">
                            <span className="material-symbols-rounded" aria-hidden="true">person_add</span>
                        </button>
                    )}
                </div>
            ) : (
                <div className="chats-me-page__details-shell chats-me-page__details-shell--empty">
                    <p className="chats-me-page__state">Choose a chat to see members.</p>
                </div>
            )}
        </aside>
    );
}
