import { type PublicProfile } from "../../../profiles/profile-cache";
import {
    Chat,
    UserAvatar,
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
                    <div className="chats-me-page__details-header">
                        <p className="chats-me-page__eyebrow">Members</p>
                        {selectedChat.chatType === 1 && (
                            <button className="chats-me-page__details-action-button" type="button" onClick={onOpenAddMembersModal}>
                                <span className="material-symbols-rounded chats-me-page__details-action-button-icon" aria-hidden="true">person_add</span>
                                Add members
                            </button>
                        )}
                    </div>
                    <div className="chats-me-page__members" aria-label="Chat members">
                        {selectedChat.memberUuids.map((memberUuid) => (
                            <div className="chats-me-page__member" key={memberUuid}>
                                <UserAvatar
                                    className="chats-me-page__member-avatar"
                                    src={getChatMemberAvatarUrl(memberUuid, profilesByUserUuid)}
                                    label={getChatMemberAvatarLabel(memberUuid, profilesByUserUuid)}
                                />
                                <span className="chats-me-page__member-name">
                                    {getChatMemberName(memberUuid, profilesByUserUuid)}
                                </span>
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
            ) : (
                <div className="chats-me-page__details-shell chats-me-page__details-shell--empty">
                    <p className="chats-me-page__eyebrow">Members</p>
                    <p className="chats-me-page__state">Choose a chat to see members.</p>
                </div>
            )}
        </aside>
    );
}
