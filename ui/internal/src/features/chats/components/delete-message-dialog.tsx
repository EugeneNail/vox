import { type PublicProfile } from "../../../profiles/profile-cache";
import {
    ChatMessage,
    MessageAttachments,
    MessageText,
    ModalBackdrop,
    UserAvatar,
    formatMessageTime,
    getMessageAuthorAvatarLabel,
    getMessageAuthorAvatarUrl,
    getMessageAuthorName,
} from "./chats-ui";

type DeleteMessageDialogProps = {
    isDeletingMessage: boolean;
    message: ChatMessage | null;
    onCancel: () => void;
    onConfirm: (message: ChatMessage) => Promise<void>;
    profilesByUserUuid: Record<string, PublicProfile>;
};

export function DeleteMessageDialog(props: DeleteMessageDialogProps) {
    const {
        isDeletingMessage,
        message,
        onCancel,
        onConfirm,
        profilesByUserUuid,
    } = props;

    if (!message) {
        return null;
    }

    return (
        <ModalBackdrop>
            <section className="chats-me-page__delete-dialog" role="dialog" aria-modal="true" aria-labelledby="delete-message-title">
                <div className="chats-me-page__delete-dialog-header">
                    <h2 className="chats-me-page__delete-dialog-title" id="delete-message-title">Delete Message</h2>
                    <p className="chats-me-page__delete-dialog-text">Are you sure you want to delete this message?</p>
                </div>

                <article className="chats-me-page__delete-message-preview">
                    <div className="chats-me-page__message-avatar-cell">
                        <UserAvatar
                            className="chats-me-page__message-avatar"
                            src={getMessageAuthorAvatarUrl(message, profilesByUserUuid)}
                            label={getMessageAuthorAvatarLabel(message, profilesByUserUuid)}
                        />
                    </div>
                    <div className="chats-me-page__message-body">
                        <div className="chats-me-page__message-header">
                            <p className="chats-me-page__message-author">{getMessageAuthorName(message, profilesByUserUuid)}</p>
                        </div>
                        <MessageText message={message} />
                        <MessageAttachments attachments={message.attachments} />
                    </div>
                    <div className="chats-me-page__message-status">
                        <time
                            className="chats-me-page__message-time"
                            dateTime={message.createdAt}
                        >
                            {formatMessageTime(message.createdAt)}
                        </time>
                    </div>
                </article>

                <div className="chats-me-page__delete-dialog-actions">
                    <button className="chats-me-page__delete-dialog-button" type="button" disabled={isDeletingMessage} onClick={onCancel}>
                        Cancel
                    </button>
                    <button className="chats-me-page__delete-dialog-button chats-me-page__delete-dialog-button--danger" type="button" disabled={isDeletingMessage} onClick={() => void onConfirm(message)}>
                        Delete
                    </button>
                </div>
            </section>
        </ModalBackdrop>
    );
}
